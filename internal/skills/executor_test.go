package skills

import (
	"strings"
	"testing"
	"time"
)

func TestExecutorCompletionRequiresEvidence(t *testing.T) {
	now := time.Date(2026, 7, 3, 1, 0, 0, 0, time.UTC)
	executor := Executor{Now: func() time.Time { return now }}
	skill := testSkill("feature-spec", SourceProject, "feature specs")

	run, err := executor.Start("T-1", skill, "session-1")
	if err != nil {
		t.Fatal(err)
	}
	if run.Status != RunActivated || run.StartedAt != now {
		t.Fatalf("unexpected started run: %#v", run)
	}

	run, err = executor.Complete(run, skill)
	if err == nil || !strings.Contains(err.Error(), "notes") {
		t.Fatalf("expected missing evidence error, got %v", err)
	}
	if run.Status != RunEvidencePending {
		t.Fatalf("run status = %s, want %s", run.Status, RunEvidencePending)
	}

	run, err = executor.AddEvidence(run, Evidence{ID: "E-1", TaskID: "T-1", Type: "notes", Content: "Acceptance criteria captured."})
	if err != nil {
		t.Fatal(err)
	}
	run, err = executor.Complete(run, skill)
	if err != nil {
		t.Fatal(err)
	}
	if run.Status != RunCompleted || run.CompletedAt != now {
		t.Fatalf("unexpected completed run: %#v", run)
	}
}
