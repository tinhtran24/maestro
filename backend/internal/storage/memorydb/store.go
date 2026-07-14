package memorydb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/tinhtran24/maestro/backend/internal/storage/memorydb/gen"
)

// Store is the typed access layer over the memory projection database. Writes
// are serialised by writeMu so multi-statement edge replacements stay atomic and
// concurrent writers do not contend on SQLite locks; reads run without the lock.
type Store struct {
	db      *sql.DB
	q       *gen.Queries
	path    string
	writeMu sync.Mutex
}

// Path returns the absolute path of the backing database file.
func (s *Store) Path() string { return s.path }

// Close closes the underlying connection pool.
func (s *Store) Close() error { return s.db.Close() }

// Task is one projected completed task. PRsJSON and DecisionsJSON are stored as
// opaque JSON blobs so the projection stays decoupled from the event and domain
// types; callers marshal on the way in and unmarshal at the render edge.
type Task struct {
	ID            string
	EventID       string
	SessionID     string
	ProjectID     string
	Kind          string
	Harness       string
	Intent        string
	TaskType      string
	Branch        string
	BaseSHA       string
	HeadSHA       string
	OccurredAt    time.Time
	PRsJSON       string
	DecisionsJSON string
}

// Meta is the projection bookkeeping row.
type Meta struct {
	LastEventID string
	UpdatedAt   time.Time
}

func (s *Store) inTx(ctx context.Context, what string, fn func(*gen.Queries) error) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("memorydb: begin %s: %w", what, err)
	}
	defer func() { _ = tx.Rollback() }()
	if err := fn(s.q.WithTx(tx)); err != nil {
		return fmt.Errorf("memorydb: %s: %w", what, err)
	}
	return tx.Commit()
}

// UpsertTask inserts or replaces a task row keyed on its id.
func (s *Store) UpsertTask(ctx context.Context, t Task) error {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	prs := t.PRsJSON
	if prs == "" {
		prs = "[]"
	}
	decisions := t.DecisionsJSON
	if decisions == "" {
		decisions = "[]"
	}
	return s.q.UpsertMemoryTask(ctx, gen.UpsertMemoryTaskParams{
		ID:            t.ID,
		EventID:       t.EventID,
		SessionID:     t.SessionID,
		ProjectID:     t.ProjectID,
		Kind:          t.Kind,
		Harness:       t.Harness,
		Intent:        t.Intent,
		TaskType:      t.TaskType,
		Branch:        t.Branch,
		BaseSHA:       t.BaseSHA,
		HeadSHA:       t.HeadSHA,
		OccurredAt:    t.OccurredAt.UTC(),
		PrsJson:       prs,
		DecisionsJson: decisions,
	})
}

// GetTask returns the task with the given id. The bool is false when no such
// task exists.
func (s *Store) GetTask(ctx context.Context, id string) (Task, bool, error) {
	row, err := s.q.GetMemoryTask(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Task{}, false, nil
		}
		return Task{}, false, fmt.Errorf("memorydb: get task: %w", err)
	}
	return taskFromRow(row), true, nil
}

// ListTasks returns up to limit tasks, most recent first.
func (s *Store) ListTasks(ctx context.Context, limit int) ([]Task, error) {
	rows, err := s.q.ListMemoryTasks(ctx, int64(limit))
	if err != nil {
		return nil, fmt.Errorf("memorydb: list tasks: %w", err)
	}
	out := make([]Task, 0, len(rows))
	for _, r := range rows {
		out = append(out, taskFromRow(r))
	}
	return out, nil
}

// ReplaceFiles sets the exact set of changed non-test files for a task,
// discarding any prior edges so re-applying an event is idempotent.
func (s *Store) ReplaceFiles(ctx context.Context, taskID string, paths []string) error {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	return s.inTx(ctx, "replace files", func(q *gen.Queries) error {
		if err := q.DeleteMemoryFilesByTask(ctx, taskID); err != nil {
			return err
		}
		for _, p := range paths {
			if err := q.AddMemoryFile(ctx, gen.AddMemoryFileParams{TaskID: taskID, Path: p}); err != nil {
				return err
			}
		}
		return nil
	})
}

// ReplaceTests sets the exact set of changed test files for a task.
func (s *Store) ReplaceTests(ctx context.Context, taskID string, paths []string) error {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	return s.inTx(ctx, "replace tests", func(q *gen.Queries) error {
		if err := q.DeleteMemoryTestsByTask(ctx, taskID); err != nil {
			return err
		}
		for _, p := range paths {
			if err := q.AddMemoryTest(ctx, gen.AddMemoryTestParams{TaskID: taskID, Path: p}); err != nil {
				return err
			}
		}
		return nil
	})
}

// ListFiles returns the changed non-test files recorded for a task, sorted.
func (s *Store) ListFiles(ctx context.Context, taskID string) ([]string, error) {
	paths, err := s.q.ListMemoryFilesByTask(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("memorydb: list files: %w", err)
	}
	return paths, nil
}

// ListTests returns the changed test files recorded for a task, sorted.
func (s *Store) ListTests(ctx context.Context, taskID string) ([]string, error) {
	paths, err := s.q.ListMemoryTestsByTask(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("memorydb: list tests: %w", err)
	}
	return paths, nil
}

// GetMeta returns the projection bookkeeping row. An empty projection returns a
// zero Meta and no error.
func (s *Store) GetMeta(ctx context.Context) (Meta, error) {
	row, err := s.q.GetMemoryMeta(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Meta{}, nil
		}
		return Meta{}, fmt.Errorf("memorydb: get meta: %w", err)
	}
	return Meta{LastEventID: row.LastEventID, UpdatedAt: row.UpdatedAt}, nil
}

// SetMeta records the last event folded into the projection.
func (s *Store) SetMeta(ctx context.Context, lastEventID string, updatedAt time.Time) error {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	return s.q.UpsertMemoryMeta(ctx, gen.UpsertMemoryMetaParams{
		LastEventID: lastEventID,
		UpdatedAt:   updatedAt.UTC(),
	})
}

func taskFromRow(r gen.MemoryTask) Task {
	return Task{
		ID:            r.ID,
		EventID:       r.EventID,
		SessionID:     r.SessionID,
		ProjectID:     r.ProjectID,
		Kind:          r.Kind,
		Harness:       r.Harness,
		Intent:        r.Intent,
		TaskType:      r.TaskType,
		Branch:        r.Branch,
		BaseSHA:       r.BaseSHA,
		HeadSHA:       r.HeadSHA,
		OccurredAt:    r.OccurredAt,
		PRsJSON:       r.PrsJson,
		DecisionsJSON: r.DecisionsJson,
	}
}
