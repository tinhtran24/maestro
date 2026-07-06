package events

import (
	"testing"
	"time"

	runtimefacts "github.com/tinhtran/thanos/internal/runtime"
)

func TestMemoryRuntimeEventLogAppendsAndListsRuntimeEvents(t *testing.T) {
	now := time.Date(2026, 7, 6, 13, 0, 0, 0, time.UTC)
	log := NewMemoryRuntimeEventLogWithClock(func() time.Time { return now })

	first, err := log.AppendRuntimeEvent(RuntimeEventStarted("P-1", "T-1", "session-1", runtimefacts.StepPlanning))
	if err != nil {
		t.Fatalf("append start: %v", err)
	}
	second, err := log.AppendRuntimeEvent(RuntimeEventOutput("P-1", "T-1", "session-1", 42))
	if err != nil {
		t.Fatalf("append output: %v", err)
	}
	if first.ID != 1 || second.ID != 2 {
		t.Fatalf("ids = %d, %d; want 1, 2", first.ID, second.ID)
	}

	events, err := log.ListRuntimeEventsSince(1, 10)
	if err != nil {
		t.Fatalf("list runtime events: %v", err)
	}
	if len(events) != 1 || events[0].EventType != runtimefacts.EventAgentSessionOutput {
		t.Fatalf("unexpected events: %#v", events)
	}
	if !events[0].CreatedAt.Equal(now) {
		t.Fatalf("created_at = %s, want %s", events[0].CreatedAt, now)
	}
}

func TestRuntimeEventLogRejectsMissingType(t *testing.T) {
	log := NewMemoryRuntimeEventLog()
	if _, err := log.AppendRuntimeEvent(RuntimeEventInput{TaskID: "T-1"}); err == nil {
		t.Fatal("expected missing runtime event type error")
	}
}

func TestRuntimeEventToTimelineDistinguishesRuntimeFacts(t *testing.T) {
	now := time.Date(2026, 7, 6, 13, 30, 0, 0, time.UTC)
	event := StoredRuntimeEvent{
		ID:        7,
		ProjectID: "P-1",
		TaskID:    "T-1",
		SessionID: "session-1",
		EventType: runtimefacts.EventAgentSessionExited,
		Payload:   map[string]any{"exit_code": 0},
		CreatedAt: now,
	}

	timeline := RuntimeEventToTimeline(event)
	if timeline.Event != "runtime.AgentSessionExited" {
		t.Fatalf("timeline event = %q", timeline.Event)
	}
	if timeline.Stage != "runtime" || timeline.Status != "fact" {
		t.Fatalf("timeline stage/status = %q/%q", timeline.Stage, timeline.Status)
	}
	if timeline.TaskID != "T-1" || timeline.ProjectID != "P-1" {
		t.Fatalf("timeline ids = %#v", timeline)
	}
	if timeline.Payload["runtime_event_id"] != int64(7) {
		t.Fatalf("runtime_event_id payload missing: %#v", timeline.Payload)
	}
	if timeline.Payload["session_id"] != "session-1" {
		t.Fatalf("session_id payload missing: %#v", timeline.Payload)
	}
}

func TestRuntimeEventPayloadIsCopied(t *testing.T) {
	log := NewMemoryRuntimeEventLog()
	payload := map[string]any{"bytes": 9}
	event, err := log.AppendRuntimeEvent(RuntimeEventInput{
		TaskID:    "T-1",
		SessionID: "session-1",
		EventType: runtimefacts.EventAgentSessionOutput,
		Payload:   payload,
	})
	if err != nil {
		t.Fatalf("append: %v", err)
	}
	payload["bytes"] = 99
	events, err := log.ListRuntimeEventsSince(0, 10)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if event.Payload["bytes"] != 9 {
		t.Fatalf("returned payload was not copied: %#v", event.Payload)
	}
	if events[0].Payload["bytes"] != 9 {
		t.Fatalf("listed payload was not copied: %#v", events[0].Payload)
	}
}
