// Package memory bridges the append-only events.jsonl log (the source of truth)
// and the disposable memorydb projection. The Projector folds events into the
// projection idempotently, catches a stale projection up to the log head, and
// rebuilds the projection from scratch when it is missing or suspect.
package memory

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/tinhtran24/maestro/backend/internal/adapters/memoryevents"
	"github.com/tinhtran24/maestro/backend/internal/storage/memorydb"
)

// Projector applies memory events into a project's projection database. It holds
// the shared events-log reader; each call receives the target project's own
// memorydb store, so one Projector serves every project.
type Projector struct {
	events *memoryevents.Store
	now    func() time.Time
	log    *slog.Logger
}

// NewProjector returns a Projector reading from the given events log.
func NewProjector(events *memoryevents.Store, log *slog.Logger) *Projector {
	if log == nil {
		log = slog.Default()
	}
	return &Projector{events: events, now: time.Now, log: log}
}

// Apply folds a single event into the projection and advances the projection
// offset to that event. A task.completed event upserts the task and replaces its
// file/test edges atomically; any other event type only advances the offset, so
// an older projection built by a newer daemon skips events it does not
// understand instead of failing. Apply is idempotent: applying the same event
// twice yields the identical projection.
func (p *Projector) Apply(ctx context.Context, store *memorydb.Store, ev memoryevents.Event) error {
	if ev.ID == "" {
		return errors.New("memory: event has no id")
	}
	if ev.Type != memoryevents.TypeTaskCompleted {
		return store.SetMeta(ctx, ev.ID, p.now())
	}
	task, files, tests, err := toProjection(ev)
	if err != nil {
		return err
	}
	return store.ApplyTask(ctx, task, files, tests, ev.ID, p.now())
}

// CatchUp applies every event appended after the projection's recorded offset,
// leaving already-folded events untouched. It returns the number of events
// applied. Use it on boot or before serving a read to fold in tasks completed
// while the projection was not being written.
func (p *Projector) CatchUp(ctx context.Context, store *memorydb.Store, projectPath string) (int, error) {
	meta, err := store.GetMeta(ctx)
	if err != nil {
		return 0, err
	}
	applied := 0
	err = p.events.Iterate(ctx, projectPath, meta.LastEventID, func(ev memoryevents.Event) error {
		if aerr := p.Apply(ctx, store, ev); aerr != nil {
			return aerr
		}
		applied++
		return nil
	})
	if err != nil {
		return applied, err
	}
	return applied, nil
}

// Rebuild discards the whole projection and replays the entire event log,
// producing the canonical projection regardless of the projection's prior state.
// It is the recovery path when memory.db is missing, stale, or corrupt.
func (p *Projector) Rebuild(ctx context.Context, store *memorydb.Store, projectPath string) error {
	if err := store.Reset(ctx); err != nil {
		return fmt.Errorf("memory: reset projection: %w", err)
	}
	return p.events.Iterate(ctx, projectPath, "", func(ev memoryevents.Event) error {
		return p.Apply(ctx, store, ev)
	})
}

// toProjection converts a task.completed event into the projection row plus its
// file and test edges. Empty PR/decision lists serialise to JSON empty arrays so
// the stored blob is always valid JSON.
func toProjection(ev memoryevents.Event) (memorydb.Task, []string, []string, error) {
	if ev.Task.ID == "" {
		return memorydb.Task{}, nil, nil, fmt.Errorf("memory: task.completed event %s has no task id", ev.ID)
	}
	prsJSON, err := marshalArray(ev.Task.PRs)
	if err != nil {
		return memorydb.Task{}, nil, nil, fmt.Errorf("memory: marshal prs for %s: %w", ev.ID, err)
	}
	decisionsJSON, err := marshalArray(ev.Task.Decisions)
	if err != nil {
		return memorydb.Task{}, nil, nil, fmt.Errorf("memory: marshal decisions for %s: %w", ev.ID, err)
	}
	task := memorydb.Task{
		ID:            ev.Task.ID,
		EventID:       ev.ID,
		SessionID:     ev.Task.SessionID,
		ProjectID:     ev.Task.ProjectID,
		Kind:          ev.Task.Kind,
		Harness:       ev.Task.Harness,
		Intent:        ev.Task.Intent,
		TaskType:      ev.Task.TaskType,
		Branch:        ev.Task.Branch,
		BaseSHA:       ev.Task.BaseSHA,
		HeadSHA:       ev.Task.HeadSHA,
		OccurredAt:    ev.OccurredAt,
		PRsJSON:       prsJSON,
		DecisionsJSON: decisionsJSON,
	}
	return task, ev.Task.ChangedFiles, ev.Task.ChangedTests, nil
}

// marshalArray marshals v to JSON, mapping an empty/nil slice to "[]" rather than
// "null" so the projection never stores a null blob.
func marshalArray[T any](v []T) (string, error) {
	if len(v) == 0 {
		return "[]", nil
	}
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
