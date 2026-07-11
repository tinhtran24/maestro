package controllers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/tinhtran/thanos/backend/internal/service/planner"
)

type fakePlanService struct {
	agent string
	model string
}

func (s *fakePlanService) Available(string) bool { return true }

func (s *fakePlanService) Plan(_ context.Context, _ string, _ []string, agent, model string) (planner.TaskDraft, error) {
	s.agent = agent
	s.model = model
	return planner.TaskDraft{Title: "Structured task"}, nil
}

func TestPlanAcceptsModelForSelectedAgent(t *testing.T) {
	svc := &fakePlanService{}
	router := chi.NewRouter()
	(&PlanController{Svc: svc, DefaultAgent: "claude-code"}).Register(router)

	req := httptest.NewRequest(http.MethodPost, "/plan", strings.NewReader(`{"input":"build a cart","agent":"claude-code","model":"claude-opus-4-5"}`))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s, want 200", res.Code, res.Body.String())
	}
	if svc.agent != "claude-code" || svc.model != "claude-opus-4-5" {
		t.Fatalf("planner received agent=%q model=%q", svc.agent, svc.model)
	}
}

func TestPlanRejectsModelForDifferentAgent(t *testing.T) {
	svc := &fakePlanService{}
	router := chi.NewRouter()
	(&PlanController{Svc: svc, DefaultAgent: "codex"}).Register(router)

	req := httptest.NewRequest(http.MethodPost, "/plan", strings.NewReader(`{"input":"build a cart","agent":"codex","model":"claude-opus-4-5"}`))
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf("status = %d body=%s, want 400", res.Code, res.Body.String())
	}
	if svc.model != "" {
		t.Fatalf("unsupported model reached planner: %q", svc.model)
	}
}
