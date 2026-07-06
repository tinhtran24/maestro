package events

import (
	"slices"
	"testing"
	"time"
)

func TestMemoryChangeLogAppendsAndListsChanges(t *testing.T) {
	now := time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)
	log := NewMemoryChangeLogWithClock(func() time.Time { return now })

	first, err := log.Append(TaskStateChanged("T-1", "planning", "waiting_approval"))
	if err != nil {
		t.Fatalf("append first: %v", err)
	}
	second, err := log.Append(GateStateChanged("T-1", "plan_approval", "approved"))
	if err != nil {
		t.Fatalf("append second: %v", err)
	}
	if first.ID != 1 || second.ID != 2 {
		t.Fatalf("ids = %d, %d; want 1, 2", first.ID, second.ID)
	}

	changes, err := log.ListSince(1, 10)
	if err != nil {
		t.Fatalf("list since: %v", err)
	}
	if len(changes) != 1 || changes[0].ChangeType != ChangeGateStateChanged {
		t.Fatalf("unexpected changes: %#v", changes)
	}
	if !changes[0].CreatedAt.Equal(now) {
		t.Fatalf("created_at = %s, want %s", changes[0].CreatedAt, now)
	}
}

func TestChangeLogRejectsIncompleteChanges(t *testing.T) {
	log := NewMemoryChangeLog()
	if _, err := log.Append(ChangeInput{AggregateID: "T-1", ChangeType: ChangeTaskStateChanged}); err == nil {
		t.Fatal("expected missing aggregate type error")
	}
	if _, err := log.Append(ChangeInput{AggregateType: AggregateTask, ChangeType: ChangeTaskStateChanged}); err == nil {
		t.Fatal("expected missing aggregate id error")
	}
	if _, err := log.Append(ChangeInput{AggregateType: AggregateTask, AggregateID: "T-1"}); err == nil {
		t.Fatal("expected missing change type error")
	}
}

func TestRefreshSinceDerivesReadModelReasonsAndAdvancesCursor(t *testing.T) {
	log := NewMemoryChangeLog()
	inputs := []ChangeInput{
		TaskStateChanged("T-1", "ready", "running"),
		GateStateChanged("T-1", "review", "pending"),
		RuntimeFactChanged("T-1", "session-1", "running", map[string]any{"activity_state": "active"}),
		SkillRunChanged("T-1", "skillrun-1", "evidence_pending"),
		DerivedStatusRefreshRequested("T-1", "session-1", "running"),
	}
	for _, input := range inputs {
		if _, err := log.Append(input); err != nil {
			t.Fatalf("append %s: %v", input.ChangeType, err)
		}
	}

	request, cursor, err := RefreshSince(log, ReadModelCursor{}, 100)
	if err != nil {
		t.Fatalf("refresh: %v", err)
	}
	if cursor.LastChangeID != 5 || request.NextID != 5 {
		t.Fatalf("cursor = %#v request.NextID=%d, want 5", cursor, request.NextID)
	}
	for _, reason := range []RefreshReason{RefreshTasks, RefreshGates, RefreshRuntimeStatus, RefreshSkillRuns} {
		if !slices.Contains(request.Reasons, reason) {
			t.Fatalf("refresh reasons missing %q: %#v", reason, request.Reasons)
		}
	}

	empty, next, err := RefreshSince(log, cursor, 100)
	if err != nil {
		t.Fatalf("second refresh: %v", err)
	}
	if len(empty.Changes) != 0 || len(empty.Reasons) != 0 {
		t.Fatalf("second refresh should be empty: %#v", empty)
	}
	if next.LastChangeID != cursor.LastChangeID {
		t.Fatalf("cursor advanced without changes: got %d want %d", next.LastChangeID, cursor.LastChangeID)
	}
}

func TestPayloadIsCopied(t *testing.T) {
	log := NewMemoryChangeLog()
	payload := map[string]any{"activity_state": "active"}
	change, err := log.Append(RuntimeFactChanged("T-1", "session-1", "running", payload))
	if err != nil {
		t.Fatalf("append: %v", err)
	}
	payload["activity_state"] = "mutated"
	changes, err := log.ListSince(0, 10)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if change.Payload["activity_state"] != "active" {
		t.Fatalf("returned payload was not copied: %#v", change.Payload)
	}
	if changes[0].Payload["activity_state"] != "active" {
		t.Fatalf("listed payload was not copied: %#v", changes[0].Payload)
	}
}
