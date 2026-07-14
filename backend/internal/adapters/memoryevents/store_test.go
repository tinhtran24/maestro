package memoryevents

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"
)

func sampleEvent(id string) Event {
	return Event{
		V:          SchemaVersion,
		ID:         id,
		Type:       TypeTaskCompleted,
		OccurredAt: time.Date(2026, 7, 14, 12, 0, 0, 0, time.UTC),
		Task: Task{
			ID:           id + "-task",
			SessionID:    "acme-7",
			ProjectID:    "acme",
			Kind:         "worker",
			Intent:       "do the thing",
			ChangedFiles: []string{"a.go"},
			ChangedTests: []string{"a_test.go"},
		},
	}
}

func collect(t *testing.T, s *Store, dir, after string) []Event {
	t.Helper()
	var got []Event
	if err := s.Iterate(context.Background(), dir, after, func(ev Event) error {
		got = append(got, ev)
		return nil
	}); err != nil {
		t.Fatalf("Iterate: %v", err)
	}
	return got
}

func TestAppendAndIterateRoundTrip(t *testing.T) {
	dir := t.TempDir()
	s := New()
	ctx := context.Background()

	ids := []string{"EVT-1", "EVT-2", "EVT-3"}
	for _, id := range ids {
		if err := s.Append(ctx, dir, sampleEvent(id)); err != nil {
			t.Fatalf("Append %s: %v", id, err)
		}
	}

	got := collect(t, s, dir, "")
	if len(got) != len(ids) {
		t.Fatalf("got %d events, want %d", len(got), len(ids))
	}
	for i, id := range ids {
		if got[i].ID != id {
			t.Errorf("event %d id = %q, want %q", i, got[i].ID, id)
		}
	}
	// Round-trip a representative payload field to prove full fidelity.
	if got[0].Task.ChangedTests[0] != "a_test.go" {
		t.Errorf("changed tests not preserved: %+v", got[0].Task)
	}
}

func TestAppendIsAppendOnly(t *testing.T) {
	dir := t.TempDir()
	s := New()
	ctx := context.Background()

	if err := s.Append(ctx, dir, sampleEvent("EVT-1")); err != nil {
		t.Fatalf("Append: %v", err)
	}
	before, err := os.ReadFile(EventsPath(dir))
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if err := s.Append(ctx, dir, sampleEvent("EVT-2")); err != nil {
		t.Fatalf("Append: %v", err)
	}
	after, err := os.ReadFile(EventsPath(dir))
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	// The second append must leave the first line's bytes untouched and only
	// grow the file.
	if !strings.HasPrefix(string(after), string(before)) {
		t.Fatalf("append rewrote earlier content:\nbefore=%q\nafter=%q", before, after)
	}
	if len(after) <= len(before) {
		t.Fatalf("append did not grow the log: before=%d after=%d", len(before), len(after))
	}
	if lines := strings.Count(strings.TrimRight(string(after), "\n"), "\n") + 1; lines != 2 {
		t.Fatalf("expected 2 lines, got %d: %q", lines, after)
	}
}

func TestIterateSurfacesMalformedTrailingLine(t *testing.T) {
	dir := t.TempDir()
	s := New()
	ctx := context.Background()

	if err := s.Append(ctx, dir, sampleEvent("EVT-1")); err != nil {
		t.Fatalf("Append: %v", err)
	}
	// Simulate a crash mid-append: a non-empty, unterminated JSON fragment.
	f, err := os.OpenFile(EventsPath(dir), os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if _, err := f.WriteString(`{"v":1,"id":"EVT-2"`); err != nil {
		t.Fatalf("write fragment: %v", err)
	}
	f.Close()

	var got []Event
	err = s.Iterate(ctx, dir, "", func(ev Event) error {
		got = append(got, ev)
		return nil
	})
	if err == nil {
		t.Fatal("expected an error on the malformed trailing line, got nil")
	}
	if !strings.Contains(err.Error(), "parse") {
		t.Errorf("error should identify a parse failure, got: %v", err)
	}
	// The valid event preceding the fragment must still have been delivered.
	if len(got) != 1 || got[0].ID != "EVT-1" {
		t.Errorf("expected the valid EVT-1 before the error, got %+v", got)
	}
}

func TestIterateMissingLogIsNotAnError(t *testing.T) {
	dir := t.TempDir()
	s := New()
	calls := 0
	err := s.Iterate(context.Background(), dir, "", func(Event) error {
		calls++
		return nil
	})
	if err != nil {
		t.Fatalf("expected nil for a missing log, got %v", err)
	}
	if calls != 0 {
		t.Fatalf("expected no callbacks, got %d", calls)
	}
}

func TestIterateResumesAfterOffset(t *testing.T) {
	dir := t.TempDir()
	s := New()
	ctx := context.Background()
	for _, id := range []string{"EVT-1", "EVT-2", "EVT-3"} {
		if err := s.Append(ctx, dir, sampleEvent(id)); err != nil {
			t.Fatalf("Append %s: %v", id, err)
		}
	}
	got := collect(t, s, dir, "EVT-2")
	if len(got) != 1 || got[0].ID != "EVT-3" {
		t.Fatalf("resume after EVT-2 should yield only EVT-3, got %+v", got)
	}
}

func TestAppendRequiresID(t *testing.T) {
	dir := t.TempDir()
	s := New()
	ev := sampleEvent("EVT-1")
	ev.ID = ""
	if err := s.Append(context.Background(), dir, ev); err == nil {
		t.Fatal("expected an error when event id is empty")
	}
}

func TestIteratePropagatesCallbackError(t *testing.T) {
	dir := t.TempDir()
	s := New()
	ctx := context.Background()
	for _, id := range []string{"EVT-1", "EVT-2"} {
		if err := s.Append(ctx, dir, sampleEvent(id)); err != nil {
			t.Fatalf("Append %s: %v", id, err)
		}
	}
	sentinel := errors.New("stop")
	err := s.Iterate(ctx, dir, "", func(Event) error { return sentinel })
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected the callback error to propagate, got %v", err)
	}
}
