package runtime

import (
	"testing"
	"time"

	"github.com/tinhtran/thanos/internal/orchestrator"
	"github.com/tinhtran/thanos/internal/workbench"
)

func TestDeriveDisplayStatus(t *testing.T) {
	now := time.Date(2026, 7, 6, 10, 0, 0, 0, time.UTC)
	zero := 0
	one := 1
	tests := []struct {
		name string
		fact AgentSessionFact
		want DisplayStatus
	}{
		{
			name: "empty facts are idle",
			want: DisplayIdle,
		},
		{
			name: "active facts are running",
			fact: AgentSessionFact{ID: "s1", RuntimeHandleID: "pty-1", Activity: ActivityActive, StartedAt: now.Add(-time.Minute)},
			want: DisplayRunning,
		},
		{
			name: "waiting user facts are waiting user",
			fact: AgentSessionFact{ID: "s1", RuntimeHandleID: "pty-1", Activity: ActivityWaitingUser, StartedAt: now.Add(-time.Minute)},
			want: DisplayWaitingUser,
		},
		{
			name: "zero exit is completed",
			fact: AgentSessionFact{ID: "s1", Activity: ActivityExited, Terminated: true, ExitCode: &zero, StartedAt: now.Add(-time.Minute), EndedAt: now},
			want: DisplayCompleted,
		},
		{
			name: "non zero exit is failed",
			fact: AgentSessionFact{ID: "s1", Activity: ActivityExited, Terminated: true, ExitCode: &one, StartedAt: now.Add(-time.Minute), EndedAt: now},
			want: DisplayFailed,
		},
		{
			name: "terminated restorable facts are restore available",
			fact: AgentSessionFact{ID: "s1", Terminated: true, AgentNativeSessionID: "native-1", StartedAt: now.Add(-time.Minute), EndedAt: now},
			want: DisplayRestoreAvailable,
		},
		{
			name: "started without runtime handle is disconnected",
			fact: AgentSessionFact{ID: "s1", Activity: ActivityActive, StartedAt: now.Add(-time.Minute)},
			want: DisplayDisconnected,
		},
		{
			name: "started without activity after grace is no signal",
			fact: AgentSessionFact{ID: "s1", RuntimeHandleID: "pty-1", StartedAt: now.Add(-3 * time.Minute)},
			want: DisplayNoSignal,
		},
		{
			name: "recent unknown activity is starting",
			fact: AgentSessionFact{ID: "s1", RuntimeHandleID: "pty-1", StartedAt: now.Add(-10 * time.Second)},
			want: DisplayStarting,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DeriveDisplayStatus(tt.fact, now, DefaultStatusOptions())
			if got != tt.want {
				t.Fatalf("DeriveDisplayStatus() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRuntimeStatusDoesNotMoveWorkbenchTask(t *testing.T) {
	task := workbench.Task{
		ID:       "T-100",
		Status:   workbench.TaskWaitingApproval,
		Priority: workbench.PriorityP2,
	}
	status := DeriveDisplayStatus(AgentSessionFact{
		ID:              "session-T-100",
		TaskID:          task.ID,
		Step:            StepPlanning,
		RuntimeHandleID: "pty-100",
		Activity:        ActivityActive,
		StartedAt:       time.Now().UTC(),
	}, time.Now().UTC(), DefaultStatusOptions())
	if status != DisplayRunning {
		t.Fatalf("runtime status = %q, want %q", status, DisplayRunning)
	}
	if task.Status != workbench.TaskWaitingApproval {
		t.Fatalf("runtime status changed task status to %q", task.Status)
	}
	if err := orchestrator.CanMoveWorkbenchTask(task, workbench.TaskReady); err == nil {
		t.Fatal("runtime status must not satisfy plan approval gate")
	}
}

func TestWorkspaceFactIsMetadataOnly(t *testing.T) {
	facts := SessionFacts{
		Session: AgentSessionFact{
			ID:              "session-T-200",
			TaskID:          "T-200",
			Step:            StepCoding,
			RuntimeHandleID: "pty-200",
			Activity:        ActivityActive,
			StartedAt:       time.Now().UTC(),
		},
		Workspace: WorkspaceFact{
			TaskID:       "T-200",
			SessionID:    "session-T-200",
			Step:         StepCoding,
			BranchName:   "thanos/t-200",
			WorktreePath: ".thanos/worktrees/T-200",
			Prepared:     true,
			PreparedAt:   time.Now().UTC(),
		},
	}
	if got := DeriveSessionDisplayStatus(facts, time.Now().UTC(), DefaultStatusOptions()); got != DisplayRunning {
		t.Fatalf("DeriveSessionDisplayStatus() = %q, want %q", got, DisplayRunning)
	}
}
