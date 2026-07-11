package controllers

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/tinhtran/thanos/backend/internal/domain"
	"github.com/tinhtran/thanos/backend/internal/httpd/apispec"
	"github.com/tinhtran/thanos/backend/internal/httpd/envelope"
	"github.com/tinhtran/thanos/backend/internal/service/planner"
)

// PlanRequest is the POST /api/v1/plan body: raw Quick Capture input plus the
// chosen planner agent. Attachments carry name/type strings for context only.
type PlanRequest struct {
	Input       string   `json:"input"`
	Attachments []string `json:"attachments,omitempty"`
	Agent       string   `json:"agent,omitempty"`
	// Model is an optional override from the selected AI Structure provider.
	// Empty deliberately leaves the provider CLI's configured default in place.
	Model     string `json:"model,omitempty"`
	ProjectID string `json:"projectId,omitempty"`
}

// PlanResponse returns the AI-structured task draft and the agent that produced it.
type PlanResponse struct {
	Status string            `json:"status" enum:"ok"`
	Agent  string            `json:"agent"`
	Draft  planner.TaskDraft `json:"draft"`
}

// PlanService is the controller-facing planner contract.
type PlanService interface {
	Plan(ctx context.Context, input string, attachments []string, agent, model string) (planner.TaskDraft, error)
	Available(agent string) bool
}

// PlanController owns POST /plan — the Quick Capture "AI Structure" step.
type PlanController struct {
	Svc          PlanService
	DefaultAgent string // cfg.PlannerAgent; used when the request omits an agent.
}

// Register mounts the plan route.
func (c *PlanController) Register(r chi.Router) {
	r.Post("/plan", c.plan)
}

func (c *PlanController) plan(w http.ResponseWriter, r *http.Request) {
	if c.Svc == nil {
		apispec.NotImplemented(w, r, "POST", "/api/v1/plan")
		return
	}
	var req PlanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		envelope.WriteAPIError(w, r, http.StatusBadRequest, "bad_request", "INVALID_BODY", "invalid JSON body", nil)
		return
	}
	if strings.TrimSpace(req.Input) == "" && len(req.Attachments) == 0 {
		envelope.WriteAPIError(w, r, http.StatusBadRequest, "bad_request", "EMPTY_INPUT", "input or attachments required", nil)
		return
	}

	agent := strings.TrimSpace(req.Agent)
	if agent == "" {
		agent = c.DefaultAgent
	}
	if !c.Svc.Available(agent) {
		envelope.WriteAPIError(w, r, http.StatusServiceUnavailable, "unavailable", "PLANNER_AGENT_UNAVAILABLE",
			"planner agent CLI not found: "+agent, nil)
		return
	}
	if model := strings.TrimSpace(req.Model); model != "" && !domain.IsSupportedModel(domain.AgentHarness(agent), model) {
		envelope.WriteAPIError(w, r, http.StatusBadRequest, "bad_request", "MODEL_UNSUPPORTED", "model is not supported by the selected AI Structure agent", nil)
		return
	}

	draft, err := c.Svc.Plan(r.Context(), req.Input, req.Attachments, agent, strings.TrimSpace(req.Model))
	if err != nil {
		envelope.WriteError(w, r, err)
		return
	}
	envelope.WriteJSON(w, http.StatusOK, PlanResponse{Status: "ok", Agent: agent, Draft: draft})
}
