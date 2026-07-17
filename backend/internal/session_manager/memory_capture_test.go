package sessionmanager

import (
	"context"
	"fmt"
	"testing"

	"github.com/tinhtran24/maestro/backend/internal/domain"
	"github.com/tinhtran24/maestro/backend/internal/ports"
)

// recordingMemory is a fake MemoryRecorder that records each call and, via st,
// observes whether the session was already terminated when capture ran.
type recordingMemory struct {
	st                   *fakeStore
	calls                int
	terminatedWhenCalled bool
	lastRec              domain.SessionRecord
}

func (r *recordingMemory) RecordCompletion(_ context.Context, rec domain.SessionRecord) {
	r.calls++
	r.lastRec = rec
	if r.st != nil {
		r.terminatedWhenCalled = r.st.sessions[rec.ID].IsTerminated
	}
}

func newManagerWithMemory(mem MemoryRecorder) (*Manager, *fakeStore, *fakeWorkspace) {
	st := newFakeStore()
	st.projects["mer"] = domain.ProjectRecord{ID: "mer", Config: testRoleAgents()}
	ws := &fakeWorkspace{}
	m := New(Deps{
		Runtime:   &fakeRuntime{},
		Agents:    fakeAgents{},
		Workspace: ws,
		Store:     st,
		Messenger: &fakeMessenger{},
		Lifecycle: &fakeLCM{store: st},
		LookPath:  func(string) (string, error) { return "/bin/true", nil },
		Memory:    mem,
	})
	return m, st, ws
}

// TestKill_RecordsCompletionMemoryAfterTerminate covers checks (a) capture runs
// after MarkTerminated and (c) the path is exercised for a real session.
func TestKill_RecordsCompletionMemoryAfterTerminate(t *testing.T) {
	mem := &recordingMemory{}
	m, st, _ := newManagerWithMemory(mem)
	mem.st = st
	st.sessions["mer-1"] = mkLive("mer-1")

	freed, err := m.Kill(ctx, "mer-1")
	if err != nil || !freed {
		t.Fatalf("freed=%v err=%v", freed, err)
	}
	if mem.calls != 1 {
		t.Fatalf("RecordCompletion calls = %d, want 1", mem.calls)
	}
	if !mem.terminatedWhenCalled {
		t.Fatal("capture must run after MarkTerminated: session was not terminated when RecordCompletion ran")
	}
	if mem.lastRec.ID != "mer-1" {
		t.Fatalf("captured record id = %q, want mer-1", mem.lastRec.ID)
	}
}

// TestKill_OrchestratorAlsoRecordsCompletion covers check (c): no kind filter,
// an orchestrator session produces completion memory too.
func TestKill_OrchestratorAlsoRecordsCompletion(t *testing.T) {
	mem := &recordingMemory{}
	m, st, _ := newManagerWithMemory(mem)
	mem.st = st
	rec := mkLive("mer-1")
	rec.Kind = domain.KindOrchestrator
	st.sessions["mer-1"] = rec

	if _, err := m.Kill(ctx, "mer-1"); err != nil {
		t.Fatalf("kill: %v", err)
	}
	if mem.calls != 1 || mem.lastRec.Kind != domain.KindOrchestrator {
		t.Fatalf("orchestrator completion not captured: calls=%d kind=%q", mem.calls, mem.lastRec.Kind)
	}
}

// TestKill_PreservedDirtyWorkspaceRecordsNothing: a session preserved because of
// a dirty worktree is not terminated, so no completion memory is captured.
func TestKill_PreservedDirtyWorkspaceRecordsNothing(t *testing.T) {
	mem := &recordingMemory{}
	m, st, ws := newManagerWithMemory(mem)
	mem.st = st
	st.sessions["mer-1"] = mkLive("mer-1")
	ws.destroyErr = fmt.Errorf("gitworktree: refusing to remove: %w", ports.ErrWorkspaceDirty)

	if _, err := m.Kill(ctx, "mer-1"); err != nil {
		t.Fatalf("kill dirty workspace err = %v, want nil", err)
	}
	if mem.calls != 0 {
		t.Fatalf("no completion memory expected for a preserved session, got %d calls", mem.calls)
	}
}

// TestKill_NilRecorderIsSafe covers check (b) at the wiring level: with no
// recorder configured, Kill behaves exactly as before. The recorder interface
// returns nothing, so a capture failure can never fail the kill path.
func TestKill_NilRecorderIsSafe(t *testing.T) {
	m, st, _, _ := newManager()
	st.sessions["mer-1"] = mkLive("mer-1")
	freed, err := m.Kill(ctx, "mer-1")
	if err != nil || !freed {
		t.Fatalf("kill with nil recorder freed=%v err=%v", freed, err)
	}
}
