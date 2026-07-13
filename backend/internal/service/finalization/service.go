// Package finalization owns the durable orchestrator-only completion cursor.
package finalization

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/tinhtran/thanos/backend/internal/domain"
)

var (
	// ErrSessionNotFound means the requested worker or finalization row does not exist.
	ErrSessionNotFound = errors.New("finalization: session not found")
	// ErrWorkerRequired means only worker sessions may report completion.
	ErrWorkerRequired = errors.New("finalization: completion can only be reported for a worker")
	// ErrOrchestratorRequired means only a live orchestrator may advance finalization.
	ErrOrchestratorRequired = errors.New("finalization: only an orchestrator may finalize a task")
	// ErrAlreadyClaimed means another orchestrator owns the finalization cursor.
	ErrAlreadyClaimed = errors.New("finalization: claimed by another orchestrator")
	// ErrInvalidTransition means the requested state is not the next durable checkpoint.
	ErrInvalidTransition = errors.New("finalization: invalid transition")
)

// Store is the persistence boundary required by the finalization service.
type Store interface {
	GetSession(context.Context, domain.SessionID) (domain.SessionRecord, bool, error)
	RequestSessionFinalization(context.Context, domain.SessionID, time.Time) error
	GetSessionFinalization(context.Context, domain.SessionID) (domain.SessionFinalization, bool, error)
	ClaimSessionFinalization(context.Context, domain.SessionID, domain.SessionID, time.Time) (bool, error)
	AdvanceSessionFinalization(context.Context, domain.SessionID, domain.SessionID, domain.FinalizationState, domain.FinalizationState, time.Time) (bool, error)
}

// Service coordinates worker completion reports and orchestrator-owned finalization.
type Service struct {
	store Store
	now   func() time.Time
}

// New creates a finalization service.
func New(store Store, now func() time.Time) *Service {
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	return &Service{store: store, now: now}
}

// ReportComplete records that a worker session has completed its coding work.
func (s *Service) ReportComplete(ctx context.Context, workerID domain.SessionID) (domain.SessionFinalization, error) {
	worker, ok, err := s.store.GetSession(ctx, workerID)
	if err != nil {
		return domain.SessionFinalization{}, err
	}
	if !ok {
		return domain.SessionFinalization{}, ErrSessionNotFound
	}
	if worker.Kind != domain.KindWorker {
		return domain.SessionFinalization{}, ErrWorkerRequired
	}
	if err := s.store.RequestSessionFinalization(ctx, workerID, s.now()); err != nil {
		return domain.SessionFinalization{}, fmt.Errorf("request completion: %w", err)
	}
	out, _, err := s.store.GetSessionFinalization(ctx, workerID)
	return out, err
}

// Advance moves a worker finalization cursor by one orchestrator-owned step.
func (s *Service) Advance(ctx context.Context, orchestratorID, workerID domain.SessionID, requested domain.FinalizationState) (domain.SessionFinalization, error) {
	actor, ok, err := s.store.GetSession(ctx, orchestratorID)
	if err != nil {
		return domain.SessionFinalization{}, err
	}
	if !ok || actor.Kind != domain.KindOrchestrator || actor.IsTerminated {
		return domain.SessionFinalization{}, ErrOrchestratorRequired
	}
	current, ok, err := s.store.GetSessionFinalization(ctx, workerID)
	if err != nil {
		return domain.SessionFinalization{}, err
	}
	if !ok {
		return domain.SessionFinalization{}, ErrSessionNotFound
	}
	if current.OrchestratorID != "" && current.OrchestratorID != orchestratorID {
		return domain.SessionFinalization{}, ErrAlreadyClaimed
	}
	if current.OrchestratorID == "" {
		claimed, err := s.store.ClaimSessionFinalization(ctx, workerID, orchestratorID, s.now())
		if err != nil {
			return domain.SessionFinalization{}, err
		}
		if !claimed {
			return domain.SessionFinalization{}, ErrAlreadyClaimed
		}
		current.OrchestratorID = orchestratorID
	}
	if current.State == requested {
		return current, nil
	}
	next, ok := domain.NextFinalizationState(current.State)
	if !ok || next != requested {
		return domain.SessionFinalization{}, fmt.Errorf("%w: %s -> %s", ErrInvalidTransition, current.State, requested)
	}
	advanced, err := s.store.AdvanceSessionFinalization(ctx, workerID, orchestratorID, current.State, next, s.now())
	if err != nil {
		return domain.SessionFinalization{}, err
	}
	if !advanced {
		// A concurrent identical request may have won. Re-read and accept it.
		latest, exists, readErr := s.store.GetSessionFinalization(ctx, workerID)
		if readErr != nil {
			return domain.SessionFinalization{}, readErr
		}
		if exists && latest.OrchestratorID == orchestratorID && latest.State == requested {
			return latest, nil
		}
		return domain.SessionFinalization{}, ErrInvalidTransition
	}
	out, _, err := s.store.GetSessionFinalization(ctx, workerID)
	return out, err
}
