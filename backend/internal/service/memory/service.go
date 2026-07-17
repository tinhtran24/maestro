package memory

import (
	"context"
	"log/slog"
	"time"

	"github.com/tinhtran24/maestro/backend/internal/adapters/memoryevents"
	"github.com/tinhtran24/maestro/backend/internal/httpd/apierr"
	"github.com/tinhtran24/maestro/backend/internal/storage/memorydb"
)

const (
	defaultListLimit = 100
	maxListLimit     = 500
)

// TaskView is a completed task hydrated with its changed paths, for read APIs.
type TaskView struct {
	ID           string
	SessionID    string
	ProjectID    string
	Kind         string
	Harness      string
	Intent       string
	TaskType     string
	Branch       string
	OccurredAt   time.Time
	ChangedFiles []string
	ChangedTests []string
	PRs          []memoryevents.PRRef
}

// ContextInput parameterises a context-pack request.
type ContextInput struct {
	Role      string
	Intent    string
	Files     []string
	Tests     []string
	MaxTokens int
}

// RebuildResult reports the outcome of rebuilding a projection.
type RebuildResult struct {
	ProjectID string
	Tasks     int
}

// Service is the read/maintenance surface over project memory: it lists tasks,
// builds context packs, and rebuilds projections. It resolves each project's
// checkout, opens its projection, and catches the projection up to the event log
// before serving a read so results reflect the latest committed memory.
type Service struct {
	projector *Projector
	projects  ProjectLookup
	openStore storeOpener
	log       *slog.Logger
}

// NewService wires the read service. cacheDir is the app-state cache root under
// which each project's projection database lives.
func NewService(projector *Projector, projects ProjectLookup, cacheDir string, log *slog.Logger) *Service {
	if log == nil {
		log = slog.Default()
	}
	return &Service{
		projector: projector,
		projects:  projects,
		openStore: func(projectID string) (*memorydb.Store, error) {
			return memorydb.Open(cacheDir, projectID)
		},
		log: log,
	}
}

// ListTasks returns the most recent tasks for a project, each hydrated with its
// changed files and tests.
func (s *Service) ListTasks(ctx context.Context, projectID string, limit int) ([]TaskView, error) {
	if limit <= 0 {
		limit = defaultListLimit
	}
	if limit > maxListLimit {
		limit = maxListLimit
	}
	store, err := s.open(ctx, projectID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = store.Close() }()

	tasks, err := store.ListTasks(ctx, limit)
	if err != nil {
		return nil, err
	}
	views := make([]TaskView, 0, len(tasks))
	for _, t := range tasks {
		files, err := store.ListFiles(ctx, t.ID)
		if err != nil {
			return nil, err
		}
		tests, err := store.ListTests(ctx, t.ID)
		if err != nil {
			return nil, err
		}
		views = append(views, TaskView{
			ID:           t.ID,
			SessionID:    t.SessionID,
			ProjectID:    t.ProjectID,
			Kind:         t.Kind,
			Harness:      t.Harness,
			Intent:       t.Intent,
			TaskType:     t.TaskType,
			Branch:       t.Branch,
			OccurredAt:   t.OccurredAt,
			ChangedFiles: files,
			ChangedTests: tests,
			PRs:          decodePRs(t.PRsJSON),
		})
	}
	return views, nil
}

// Context builds a role-specific, token-budgeted context pack for the described
// work. An unknown role is a client error.
func (s *Service) Context(ctx context.Context, projectID string, in ContextInput) (ContextPack, error) {
	role := Role(in.Role)
	if _, ok := roleSpecs[role]; !ok {
		return ContextPack{}, apierr.Invalid("INVALID_ROLE", "role must be one of planner, coder, reviewer, tester", nil)
	}
	store, err := s.open(ctx, projectID)
	if err != nil {
		return ContextPack{}, err
	}
	defer func() { _ = store.Close() }()

	cr := NewContextResolver(in.MaxTokens, s.log)
	return cr.Pack(ctx, store, Query{Files: in.Files, Tests: in.Tests}, role)
}

// Rebuild discards and replays the project's projection from its event log,
// returning the resulting task count.
func (s *Service) Rebuild(ctx context.Context, projectID string) (RebuildResult, error) {
	project, ok, err := s.projects.GetProject(ctx, projectID)
	if err != nil {
		return RebuildResult{}, err
	}
	if !ok {
		return RebuildResult{}, apierr.NotFound("PROJECT_NOT_FOUND", "Unknown project")
	}
	store, err := s.openStore(projectID)
	if err != nil {
		return RebuildResult{}, err
	}
	defer func() { _ = store.Close() }()

	if err := s.projector.Rebuild(ctx, store, project.Path); err != nil {
		return RebuildResult{}, err
	}
	count, err := store.CountTasks(ctx)
	if err != nil {
		return RebuildResult{}, err
	}
	return RebuildResult{ProjectID: projectID, Tasks: count}, nil
}

// open resolves the project, opens its projection, and folds any events appended
// since the projection was last written so reads are fresh. A missing project is
// a not-found client error.
func (s *Service) open(ctx context.Context, projectID string) (*memorydb.Store, error) {
	project, ok, err := s.projects.GetProject(ctx, projectID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, apierr.NotFound("PROJECT_NOT_FOUND", "Unknown project")
	}
	store, err := s.openStore(projectID)
	if err != nil {
		return nil, err
	}
	if _, err := s.projector.CatchUp(ctx, store, project.Path); err != nil {
		_ = store.Close()
		return nil, err
	}
	return store, nil
}
