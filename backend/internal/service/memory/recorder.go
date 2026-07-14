package memory

import (
	"context"
	"log/slog"

	"github.com/tinhtran24/maestro/backend/internal/adapters/memoryevents"
	"github.com/tinhtran24/maestro/backend/internal/domain"
	"github.com/tinhtran24/maestro/backend/internal/storage/memorydb"
)

// ProjectLookup resolves the on-disk checkout path for a project. It is the one
// piece of daemon state the recorder needs beyond the terminated session record.
type ProjectLookup interface {
	GetProject(ctx context.Context, id string) (domain.ProjectRecord, bool, error)
}

// storeOpener opens the memorydb projection for a project. It is a seam so tests
// can supply a store rooted at a temp directory.
type storeOpener func(projectID string) (*memorydb.Store, error)

// Recorder captures completion memory when a session terminates. It is the
// concrete implementation behind the session manager's MemoryRecorder seam:
// RecordCompletion never returns an error, and it logs and abandons on any
// failure so a capture problem cannot fail or delay a lifecycle transition.
type Recorder struct {
	builder   *Builder
	events    *memoryevents.Store
	projector *Projector
	projects  ProjectLookup
	openStore storeOpener
	log       *slog.Logger
}

// NewRecorder wires a Recorder. cacheDir is the app-state cache root under which
// each project's projection database lives; projects resolves checkout paths.
func NewRecorder(builder *Builder, events *memoryevents.Store, projector *Projector, projects ProjectLookup, cacheDir string, log *slog.Logger) *Recorder {
	if log == nil {
		log = slog.Default()
	}
	return &Recorder{
		builder:   builder,
		events:    events,
		projector: projector,
		projects:  projects,
		openStore: func(projectID string) (*memorydb.Store, error) {
			return memorydb.Open(cacheDir, projectID)
		},
		log: log,
	}
}

// RecordCompletion builds the completion event for a terminated session, appends
// it to the project's events.jsonl, and folds it into the projection. Every
// failure path logs and returns; nothing propagates.
func (r *Recorder) RecordCompletion(ctx context.Context, rec domain.SessionRecord) {
	if rec.ID == "" || rec.ProjectID == "" {
		return
	}
	project, ok, err := r.projects.GetProject(ctx, string(rec.ProjectID))
	if err != nil {
		r.log.Warn("memory: resolve project for completion capture failed", "sessionID", rec.ID, "error", err)
		return
	}
	if !ok || project.Path == "" {
		r.log.Debug("memory: skipping completion capture; no project path", "sessionID", rec.ID)
		return
	}
	ev, err := r.builder.BuildCompletion(ctx, rec, project.Path)
	if err != nil {
		r.log.Warn("memory: build completion event failed", "sessionID", rec.ID, "error", err)
		return
	}
	if err := r.events.Append(ctx, project.Path, ev); err != nil {
		r.log.Warn("memory: append completion event failed", "sessionID", rec.ID, "error", err)
		return
	}
	store, err := r.openStore(string(rec.ProjectID))
	if err != nil {
		r.log.Warn("memory: open projection failed", "sessionID", rec.ID, "error", err)
		return
	}
	defer func() { _ = store.Close() }()
	// CatchUp folds this event plus any the projection missed, self-healing a
	// stale cache rather than blindly jumping the offset.
	if _, err := r.projector.CatchUp(ctx, store, project.Path); err != nil {
		r.log.Warn("memory: fold completion event failed", "sessionID", rec.ID, "error", err)
	}
}
