package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/tinhtran/thanos/backend/internal/domain"
	"github.com/tinhtran/thanos/backend/internal/storage/sqlite/gen"
)

// RequestSessionFinalization creates the idempotent completion cursor for a worker session.
func (s *Store) RequestSessionFinalization(ctx context.Context, id domain.SessionID, at time.Time) error {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	return s.qw.RequestSessionFinalization(ctx, gen.RequestSessionFinalizationParams{SessionID: string(id), RequestedAt: at, UpdatedAt: at})
}

// GetSessionFinalization loads the completion cursor for a worker session.
func (s *Store) GetSessionFinalization(ctx context.Context, id domain.SessionID) (domain.SessionFinalization, bool, error) {
	row, err := s.qr.GetSessionFinalization(ctx, string(id))
	if errors.Is(err, sql.ErrNoRows) {
		return domain.SessionFinalization{}, false, nil
	}
	if err != nil {
		return domain.SessionFinalization{}, false, fmt.Errorf("get finalization for %s: %w", id, err)
	}
	return finalizationFromRow(row), true, nil
}

// ClaimSessionFinalization assigns a pending completion cursor to an orchestrator.
func (s *Store) ClaimSessionFinalization(ctx context.Context, worker, orchestrator domain.SessionID, at time.Time) (bool, error) {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	actor := sql.NullString{String: string(orchestrator), Valid: true}
	n, err := s.qw.ClaimSessionFinalization(ctx, gen.ClaimSessionFinalizationParams{
		OrchestratorID: actor, UpdatedAt: at, SessionID: string(worker), OrchestratorID_2: actor,
	})
	return n > 0, err
}

// AdvanceSessionFinalization moves a claimed cursor from one state to the next.
func (s *Store) AdvanceSessionFinalization(ctx context.Context, worker, orchestrator domain.SessionID, from, to domain.FinalizationState, at time.Time) (bool, error) {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	completed := sql.NullTime{}
	if to == domain.FinalizationDone {
		completed = sql.NullTime{Time: at, Valid: true}
	}
	n, err := s.qw.AdvanceSessionFinalization(ctx, gen.AdvanceSessionFinalizationParams{
		State: string(to), UpdatedAt: at, CompletedAt: completed, SessionID: string(worker),
		OrchestratorID: sql.NullString{String: string(orchestrator), Valid: true}, State_2: string(from),
	})
	return n > 0, err
}

func finalizationFromRow(row gen.SessionFinalization) domain.SessionFinalization {
	out := domain.SessionFinalization{
		SessionID: domain.SessionID(row.SessionID), OrchestratorID: domain.SessionID(row.OrchestratorID.String),
		State: domain.FinalizationState(row.State), RequestedAt: row.RequestedAt, UpdatedAt: row.UpdatedAt, LastError: row.LastError,
	}
	if row.CompletedAt.Valid {
		completed := row.CompletedAt.Time
		out.CompletedAt = &completed
	}
	return out
}
