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

// PathOverlap reports how many of a queried path set a task shares, split by
// files vs tests, along with the task's recency for ranking.
type PathOverlap struct {
	TaskID      string
	OccurredAt  time.Time
	SharedFiles int
	SharedTests int
}

// Edge is a directed task-to-task relation emanating from a source task.
type Edge struct {
	DstTaskID  string
	Relation   string
	Confidence float64
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
	return s.q.UpsertMemoryTask(ctx, upsertTaskParams(t))
}

// upsertTaskParams builds the generated insert params, defaulting empty JSON
// blobs to empty arrays so a row always holds valid JSON.
func upsertTaskParams(t Task) gen.UpsertMemoryTaskParams {
	prs := t.PRsJSON
	if prs == "" {
		prs = "[]"
	}
	decisions := t.DecisionsJSON
	if decisions == "" {
		decisions = "[]"
	}
	return gen.UpsertMemoryTaskParams{
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
	}
}

// ApplyTask folds one completed task into the projection atomically: it upserts
// the task row, replaces the task's file and test edges, and advances the
// projection offset (memory_meta) to eventID, all in a single transaction. A
// reader therefore never sees a half-applied task, and a crash leaves the offset
// consistent with the rows. It is idempotent: re-applying the same event
// reproduces the identical projection.
func (s *Store) ApplyTask(ctx context.Context, t Task, files, tests []string, eventID string, at time.Time) error {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	return s.inTx(ctx, "apply task", func(q *gen.Queries) error {
		if err := q.UpsertMemoryTask(ctx, upsertTaskParams(t)); err != nil {
			return err
		}
		if err := q.DeleteMemoryFilesByTask(ctx, t.ID); err != nil {
			return err
		}
		for _, p := range files {
			if err := q.AddMemoryFile(ctx, gen.AddMemoryFileParams{TaskID: t.ID, Path: p}); err != nil {
				return err
			}
		}
		if err := q.DeleteMemoryTestsByTask(ctx, t.ID); err != nil {
			return err
		}
		for _, p := range tests {
			if err := q.AddMemoryTest(ctx, gen.AddMemoryTestParams{TaskID: t.ID, Path: p}); err != nil {
				return err
			}
		}
		return q.UpsertMemoryMeta(ctx, gen.UpsertMemoryMetaParams{LastEventID: eventID, UpdatedAt: at.UTC()})
	})
}

// Reset clears every projected row, returning the database to its post-migration
// empty state so Rebuild can replay the event log from scratch. It does not drop
// the schema or the file.
func (s *Store) Reset(ctx context.Context) error {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("memorydb: begin reset: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	for _, stmt := range []string{
		"DELETE FROM memory_task_edge",
		"DELETE FROM memory_test",
		"DELETE FROM memory_file",
		"DELETE FROM memory_task",
		"DELETE FROM memory_meta",
	} {
		if _, err := tx.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("memorydb: reset %q: %w", stmt, err)
		}
	}
	return tx.Commit()
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

// CountTasks returns the number of tasks in the projection.
func (s *Store) CountTasks(ctx context.Context) (int, error) {
	n, err := s.q.CountMemoryTasks(ctx)
	if err != nil {
		return 0, fmt.Errorf("memorydb: count tasks: %w", err)
	}
	return int(n), nil
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

// TasksSharingPaths returns every task (including, if present, the querying task
// itself) that shares at least one of the given file or test paths, with the
// per-task shared counts. Callers exclude the querying task and rank the results.
// Empty path sets are skipped rather than issuing an invalid `IN ()`.
func (s *Store) TasksSharingPaths(ctx context.Context, files, tests []string) ([]PathOverlap, error) {
	byTask := make(map[string]*PathOverlap)
	if len(files) > 0 {
		rows, err := s.q.TasksSharingFiles(ctx, files)
		if err != nil {
			return nil, fmt.Errorf("memorydb: tasks sharing files: %w", err)
		}
		for _, r := range rows {
			o := overlapFor(byTask, r.ID, r.OccurredAt)
			o.SharedFiles = int(r.Shared)
		}
	}
	if len(tests) > 0 {
		rows, err := s.q.TasksSharingTests(ctx, tests)
		if err != nil {
			return nil, fmt.Errorf("memorydb: tasks sharing tests: %w", err)
		}
		for _, r := range rows {
			o := overlapFor(byTask, r.ID, r.OccurredAt)
			o.SharedTests = int(r.Shared)
		}
	}
	out := make([]PathOverlap, 0, len(byTask))
	for _, o := range byTask {
		out = append(out, *o)
	}
	return out, nil
}

func overlapFor(byTask map[string]*PathOverlap, id string, occurredAt time.Time) *PathOverlap {
	o, ok := byTask[id]
	if !ok {
		o = &PathOverlap{TaskID: id, OccurredAt: occurredAt}
		byTask[id] = o
	}
	return o
}

// ReplaceEdgesFrom sets the exact set of outgoing edges for a source task,
// discarding any prior edges so re-applying a task's relations is idempotent.
func (s *Store) ReplaceEdgesFrom(ctx context.Context, srcTaskID string, edges []Edge) error {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	return s.inTx(ctx, "replace edges", func(q *gen.Queries) error {
		if err := q.DeleteEdgesFrom(ctx, srcTaskID); err != nil {
			return err
		}
		for _, e := range edges {
			if err := q.AddEdge(ctx, gen.AddEdgeParams{
				SrcTaskID:  srcTaskID,
				DstTaskID:  e.DstTaskID,
				Relation:   e.Relation,
				Confidence: e.Confidence,
			}); err != nil {
				return err
			}
		}
		return nil
	})
}

// FullEdge is a directed task-to-task relation with both endpoints, for graph
// export.
type FullEdge struct {
	SrcTaskID  string
	DstTaskID  string
	Relation   string
	Confidence float64
}

// ListAllEdges returns every edge in the projection, for rendering the task
// graph.
func (s *Store) ListAllEdges(ctx context.Context) ([]FullEdge, error) {
	rows, err := s.q.ListAllEdges(ctx)
	if err != nil {
		return nil, fmt.Errorf("memorydb: list all edges: %w", err)
	}
	out := make([]FullEdge, 0, len(rows))
	for _, r := range rows {
		out = append(out, FullEdge{SrcTaskID: r.SrcTaskID, DstTaskID: r.DstTaskID, Relation: r.Relation, Confidence: r.Confidence})
	}
	return out, nil
}

// ListEdgesFrom returns the outgoing edges of a task, strongest confidence first.
func (s *Store) ListEdgesFrom(ctx context.Context, srcTaskID string) ([]Edge, error) {
	rows, err := s.q.ListEdgesFrom(ctx, srcTaskID)
	if err != nil {
		return nil, fmt.Errorf("memorydb: list edges: %w", err)
	}
	out := make([]Edge, 0, len(rows))
	for _, r := range rows {
		out = append(out, Edge{DstTaskID: r.DstTaskID, Relation: r.Relation, Confidence: r.Confidence})
	}
	return out, nil
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
