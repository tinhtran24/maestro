package finalization

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/tinhtran/thanos/backend/internal/domain"
)

type fakeStore struct {
	sessions map[domain.SessionID]domain.SessionRecord
	flow     domain.SessionFinalization
}

func (f *fakeStore) GetSession(_ context.Context, id domain.SessionID) (domain.SessionRecord, bool, error) {
	r, ok := f.sessions[id]
	return r, ok, nil
}
func (f *fakeStore) RequestSessionFinalization(_ context.Context, id domain.SessionID, at time.Time) error {
	if f.flow.SessionID == "" {
		f.flow = domain.SessionFinalization{SessionID: id, State: domain.FinalizationPending, RequestedAt: at, UpdatedAt: at}
	}
	return nil
}
func (f *fakeStore) GetSessionFinalization(context.Context, domain.SessionID) (domain.SessionFinalization, bool, error) {
	return f.flow, f.flow.SessionID != "", nil
}
func (f *fakeStore) ClaimSessionFinalization(_ context.Context, worker, actor domain.SessionID, at time.Time) (bool, error) {
	if f.flow.SessionID != worker || (f.flow.OrchestratorID != "" && f.flow.OrchestratorID != actor) {
		return false, nil
	}
	f.flow.OrchestratorID, f.flow.UpdatedAt = actor, at
	return true, nil
}
func (f *fakeStore) AdvanceSessionFinalization(_ context.Context, worker, actor domain.SessionID, from, to domain.FinalizationState, at time.Time) (bool, error) {
	if f.flow.SessionID != worker || f.flow.OrchestratorID != actor || f.flow.State != from {
		return false, nil
	}
	f.flow.State, f.flow.UpdatedAt = to, at
	return true, nil
}

func TestCompletionFlowIsOrderedAndIdempotent(t *testing.T) {
	now := time.Date(2026, 7, 13, 1, 2, 3, 0, time.UTC)
	store := &fakeStore{sessions: map[domain.SessionID]domain.SessionRecord{
		"worker": {ID: "worker", Kind: domain.KindWorker},
		"orch":   {ID: "orch", Kind: domain.KindOrchestrator},
	}}
	svc := New(store, func() time.Time { return now })
	ctx := context.Background()
	if _, err := svc.ReportComplete(ctx, "worker"); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.ReportComplete(ctx, "worker"); err != nil {
		t.Fatalf("duplicate report: %v", err)
	}
	got, err := svc.Advance(ctx, "orch", "worker", domain.FinalizationVerifyingGit)
	if err != nil || got.State != domain.FinalizationVerifyingGit {
		t.Fatalf("advance = %#v, %v", got, err)
	}
	got, err = svc.Advance(ctx, "orch", "worker", domain.FinalizationVerifyingGit)
	if err != nil || got.State != domain.FinalizationVerifyingGit {
		t.Fatalf("duplicate advance = %#v, %v", got, err)
	}
	if _, err := svc.Advance(ctx, "orch", "worker", domain.FinalizationPushing); err == nil {
		t.Fatal("expected skipped-step error")
	}
}

func TestOnlyLiveOrchestratorCanAdvance(t *testing.T) {
	store := &fakeStore{sessions: map[domain.SessionID]domain.SessionRecord{"worker": {ID: "worker", Kind: domain.KindWorker}}, flow: domain.SessionFinalization{SessionID: "worker", State: domain.FinalizationPending}}
	if _, err := New(store, nil).Advance(context.Background(), "worker", "worker", domain.FinalizationVerifyingGit); !errors.Is(err, ErrOrchestratorRequired) {
		t.Fatalf("err = %v", err)
	}
}
