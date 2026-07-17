package controllers

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/tinhtran24/maestro/backend/internal/httpd/apispec"
	"github.com/tinhtran24/maestro/backend/internal/httpd/envelope"
	memorysvc "github.com/tinhtran24/maestro/backend/internal/service/memory"
)

// MemoryService is the project-memory read/maintenance surface the controller
// depends on. Nil keeps routes registered but returns OpenAPI-backed 501s.
type MemoryService interface {
	ListTasks(ctx context.Context, projectID string, limit int) ([]memorysvc.TaskView, error)
	Context(ctx context.Context, projectID string, in memorysvc.ContextInput) (memorysvc.ContextPack, error)
	Graph(ctx context.Context, projectID string) (memorysvc.GraphView, error)
	Rebuild(ctx context.Context, projectID string) (memorysvc.RebuildResult, error)
}

// MemoryController owns the /projects/{id}/memory routes.
type MemoryController struct {
	Svc MemoryService
}

// Register mounts the project-memory routes.
func (c *MemoryController) Register(r chi.Router) {
	r.Get("/projects/{id}/memory/tasks", c.listTasks)
	r.Get("/projects/{id}/memory/context", c.context)
	r.Get("/projects/{id}/memory/graph", c.graph)
	r.Post("/projects/{id}/memory/rebuild", c.rebuild)
}

func (c *MemoryController) listTasks(w http.ResponseWriter, r *http.Request) {
	if c.Svc == nil {
		apispec.NotImplemented(w, r, "GET", "/api/v1/projects/{id}/memory/tasks")
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	tasks, err := c.Svc.ListTasks(r.Context(), chi.URLParam(r, "id"), limit)
	if err != nil {
		envelope.WriteError(w, r, err)
		return
	}
	dtos := make([]MemoryTaskDTO, 0, len(tasks))
	for _, t := range tasks {
		dtos = append(dtos, MemoryTaskDTO{
			ID:           t.ID,
			SessionID:    t.SessionID,
			ProjectID:    t.ProjectID,
			Kind:         t.Kind,
			Harness:      t.Harness,
			Intent:       t.Intent,
			TaskType:     t.TaskType,
			Branch:       t.Branch,
			OccurredAt:   t.OccurredAt,
			ChangedFiles: nonNil(t.ChangedFiles),
			ChangedTests: nonNil(t.ChangedTests),
		})
	}
	envelope.WriteJSON(w, http.StatusOK, ListMemoryTasksResponse{Tasks: dtos})
}

func (c *MemoryController) context(w http.ResponseWriter, r *http.Request) {
	if c.Svc == nil {
		apispec.NotImplemented(w, r, "GET", "/api/v1/projects/{id}/memory/context")
		return
	}
	q := r.URL.Query()
	maxTokens, _ := strconv.Atoi(q.Get("maxTokens"))
	pack, err := c.Svc.Context(r.Context(), chi.URLParam(r, "id"), memorysvc.ContextInput{
		Role:      q.Get("role"),
		Intent:    q.Get("intent"),
		Files:     splitCSV(q.Get("files")),
		Tests:     splitCSV(q.Get("tests")),
		MaxTokens: maxTokens,
	})
	if err != nil {
		envelope.WriteError(w, r, err)
		return
	}
	envelope.WriteJSON(w, http.StatusOK, contextResponse(pack))
}

func (c *MemoryController) graph(w http.ResponseWriter, r *http.Request) {
	if c.Svc == nil {
		apispec.NotImplemented(w, r, "GET", "/api/v1/projects/{id}/memory/graph")
		return
	}
	g, err := c.Svc.Graph(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		envelope.WriteError(w, r, err)
		return
	}
	nodes := make([]MemoryGraphNodeDTO, 0, len(g.Nodes))
	for _, n := range g.Nodes {
		nodes = append(nodes, MemoryGraphNodeDTO{
			TaskID:       n.TaskID,
			Intent:       n.Intent,
			TaskType:     n.TaskType,
			Kind:         n.Kind,
			OccurredAt:   n.OccurredAt,
			ChangedFiles: n.ChangedFiles,
			ChangedTests: n.ChangedTests,
		})
	}
	edges := make([]MemoryGraphEdgeDTO, 0, len(g.Edges))
	for _, e := range g.Edges {
		edges = append(edges, MemoryGraphEdgeDTO{Source: e.Source, Target: e.Target, Relation: e.Relation, Confidence: e.Confidence})
	}
	envelope.WriteJSON(w, http.StatusOK, MemoryGraphResponse{Nodes: nodes, Edges: edges})
}

func (c *MemoryController) rebuild(w http.ResponseWriter, r *http.Request) {
	if c.Svc == nil {
		apispec.NotImplemented(w, r, "POST", "/api/v1/projects/{id}/memory/rebuild")
		return
	}
	res, err := c.Svc.Rebuild(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		envelope.WriteError(w, r, err)
		return
	}
	envelope.WriteJSON(w, http.StatusOK, RebuildMemoryResponse{ProjectID: res.ProjectID, Tasks: res.Tasks})
}

func contextResponse(pack memorysvc.ContextPack) MemoryContextResponse {
	tasks := make([]MemoryPackTaskDTO, 0, len(pack.RelatedTasks))
	for _, t := range pack.RelatedTasks {
		tasks = append(tasks, MemoryPackTaskDTO{
			TaskID:      t.TaskID,
			Intent:      t.Intent,
			TaskType:    t.TaskType,
			Branch:      t.Branch,
			SharedFiles: t.SharedFiles,
			SharedTests: t.SharedTests,
			Files:       t.Files,
			Tests:       t.Tests,
			Decisions:   t.Decisions,
		})
	}
	dropped := make([]MemoryDroppedDTO, 0, len(pack.Dropped))
	for _, d := range pack.Dropped {
		dropped = append(dropped, MemoryDroppedDTO{Kind: d.Kind, Ref: d.Ref, Reason: d.Reason})
	}
	return MemoryContextResponse{
		Role:            string(pack.Role),
		RelatedTasks:    tasks,
		RelevantFiles:   nonNil(pack.RelevantFiles),
		RelevantTests:   nonNil(pack.RelevantTests),
		Decisions:       nonNil(pack.Decisions),
		Dropped:         dropped,
		EstimatedTokens: pack.EstimatedTokens,
	}
}

// splitCSV splits a comma-separated query value into trimmed, non-empty parts.
func splitCSV(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	var out []string
	for _, part := range strings.Split(s, ",") {
		if p := strings.TrimSpace(part); p != "" {
			out = append(out, p)
		}
	}
	return out
}

func nonNil(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}
