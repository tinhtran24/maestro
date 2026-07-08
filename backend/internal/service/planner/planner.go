// Package planner turns raw, unstructured task input (pasted text, requirement
// dumps, issue bodies, attachment descriptions) into a structured engineering
// task draft. It drives the locally-authenticated Claude Code CLI in headless
// print mode (`claude -p --output-format json`) — real Claude, no API key.
package planner

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ErrUnavailable is returned when no Claude CLI could be resolved to run the
// planner. Callers surface it as a 501/503 so the UI can fall back gracefully.
var ErrUnavailable = errors.New("planner: claude CLI not available")

// TaskDraft is the structured task the planner infers from raw input. Every
// field is best-effort; the UI lets the user review and edit before creating.
type TaskDraft struct {
	Title              string     `json:"title"`
	Description        string     `json:"description"`
	UserStory          string     `json:"userStory"`
	Priority           string     `json:"priority"` // P0 | P1 | P2 | P3
	Labels             []string   `json:"labels"`
	AcceptanceCriteria []string   `json:"acceptanceCriteria"`
	TechnicalNotes     string     `json:"technicalNotes"`
	LikelyFiles        []string   `json:"likelyFiles"`
	Risks              []string   `json:"risks"`
	Dependencies       []string   `json:"dependencies"`
	Estimate           string     `json:"estimate"`
	Scope              string     `json:"scope"`
	Plan               []string   `json:"plan"`
	OpenQuestions      []string   `json:"openQuestions"`
	MissingInformation []string   `json:"missingInformation"`
	SuggestedAgent     string     `json:"suggestedAgent"`
	Confidence         Confidence `json:"confidence"`
}

// Confidence carries 0–100 confidence scores the UI renders as badges.
type Confidence struct {
	Overall            Score `json:"overall"`
	Title              Score `json:"title"`
	Priority           Score `json:"priority"`
	AcceptanceCriteria Score `json:"acceptanceCriteria"`
}

// Score is a 0–100 confidence value. Models are inconsistent about the shape:
// it may arrive as a number (95), a numeric string ("95"/"95%"), or a
// qualitative word ("high"/"medium"/"low"). Score tolerates all of them.
type Score int

// UnmarshalJSON accepts number, numeric string, or a qualitative word.
func (s *Score) UnmarshalJSON(b []byte) error {
	b = bytes.TrimSpace(b)
	if len(b) == 0 || string(b) == "null" {
		*s = 0
		return nil
	}
	if b[0] != '"' {
		var n float64
		if err := json.Unmarshal(b, &n); err == nil {
			*s = Score(int(n))
			return nil
		}
	}
	var str string
	if err := json.Unmarshal(b, &str); err != nil {
		*s = 0
		return nil
	}
	str = strings.TrimSpace(strings.ToLower(strings.TrimSuffix(strings.TrimSpace(str), "%")))
	switch str {
	case "very high", "high":
		*s = 90
	case "medium", "med", "moderate":
		*s = 70
	case "low":
		*s = 40
	case "very low":
		*s = 20
	default:
		if f, err := strconv.ParseFloat(str, 64); err == nil {
			*s = Score(int(f))
		} else {
			*s = 0
		}
	}
	return nil
}

// Runner runs the planning prompt through the chosen agent's CLI and returns
// the model's raw reply text. The default runner execs the agent CLI headlessly
// (Claude gets its JSON envelope); tests inject a fake.
type Runner interface {
	Run(ctx context.Context, agent, prompt string) (string, error)
	// Available reports whether the given agent can be run headlessly.
	Available(agent string) bool
}

// Service structures raw input into a TaskDraft.
type Service struct {
	runner  Runner
	timeout time.Duration
}

// Options configures the planner. Zero values get sensible defaults.
type Options struct {
	Runner  Runner        // default: the agent CLI runner (CLIRunner)
	Timeout time.Duration // default 110s (headless agents can take tens of seconds)
}

// New builds a planner Service.
func New(opts Options) *Service {
	r := opts.Runner
	if r == nil {
		r = &CLIRunner{}
	}
	t := opts.Timeout
	if t <= 0 {
		t = 110 * time.Second
	}
	return &Service{runner: r, timeout: t}
}

// Available reports whether the planner can run the given agent headlessly.
func (s *Service) Available(agent string) bool {
	return s.runner.Available(agent)
}

// Plan structures raw input (plus optional attachment descriptions) into a
// TaskDraft using the chosen agent. It is safe to pass a large blob of text.
func (s *Service) Plan(ctx context.Context, input string, attachments []string, agent string) (TaskDraft, error) {
	input = strings.TrimSpace(input)
	if input == "" && len(attachments) == 0 {
		return TaskDraft{}, errors.New("planner: empty input")
	}

	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	reply, err := s.runner.Run(ctx, agent, buildPrompt(input, attachments))
	if err != nil {
		return TaskDraft{}, err
	}

	draft, err := parseDraft(reply)
	if err != nil {
		return TaskDraft{}, err
	}
	return normalize(draft), nil
}

// buildPrompt asks Claude for a single JSON object, no prose, matching TaskDraft.
func buildPrompt(input string, attachments []string) string {
	var b strings.Builder
	b.WriteString("You are a senior engineering planner for an AI development workbench. ")
	b.WriteString("Convert the raw input below into ONE structured engineering task.\n\n")
	b.WriteString("Respond with ONLY a single minified JSON object — no prose, no markdown, no code fences. ")
	b.WriteString("Use exactly these keys:\n")
	b.WriteString(`title (string), description (string), userStory (string, "As a ... I want ... so that ..."), `)
	b.WriteString("priority (one of \"P0\",\"P1\",\"P2\",\"P3\"), labels (string[]), acceptanceCriteria (string[]), ")
	b.WriteString("technicalNotes (string), likelyFiles (string[]), risks (string[]), dependencies (string[]), ")
	b.WriteString("estimate (short string like \"4h\" or \"2d\"), scope (\"XS\"|\"S\"|\"M\"|\"L\"|\"XL\"), plan (string[] of short high-level steps), ")
	b.WriteString("openQuestions (string[]), missingInformation (string[] of specific facts you need but were not given), ")
	b.WriteString("suggestedAgent (string, e.g. \"claude\"), ")
	b.WriteString("confidence (object with integer 0-100 fields: overall, title, priority, acceptanceCriteria).\n")
	b.WriteString("Infer priority, labels, acceptance criteria and the rest yourself; do not ask the user to fill them in.\n\n")
	if len(attachments) > 0 {
		b.WriteString("Attachments provided (names/types only): ")
		b.WriteString(strings.Join(attachments, ", "))
		b.WriteString("\n\n")
	}
	b.WriteString("Raw input:\n")
	b.WriteString(input)
	return b.String()
}

// parseDraft extracts the JSON object from the model reply and unmarshals it,
// tolerating stray prose or code fences around the object.
func parseDraft(reply string) (TaskDraft, error) {
	raw := extractJSONObject(reply)
	if raw == "" {
		return TaskDraft{}, fmt.Errorf("planner: no JSON object in model reply")
	}
	var d TaskDraft
	if err := json.Unmarshal([]byte(raw), &d); err != nil {
		return TaskDraft{}, fmt.Errorf("planner: parse draft: %w", err)
	}
	return d, nil
}

// extractJSONObject returns the substring from the first '{' to the matching
// last '}', stripping markdown fences the model may have added.
func extractJSONObject(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "```json")
	s = strings.TrimPrefix(s, "```")
	s = strings.TrimSuffix(s, "```")
	start := strings.Index(s, "{")
	end := strings.LastIndex(s, "}")
	if start < 0 || end <= start {
		return ""
	}
	return s[start : end+1]
}

// normalize clamps confidence to 0–100 and defaults priority when the model
// returned something unexpected, so the UI always has usable values.
func normalize(d TaskDraft) TaskDraft {
	switch strings.ToUpper(strings.TrimSpace(d.Priority)) {
	case "P0", "P1", "P2", "P3":
		d.Priority = strings.ToUpper(strings.TrimSpace(d.Priority))
	default:
		d.Priority = "P2"
	}
	clamp := func(n Score) Score {
		if n < 0 {
			return 0
		}
		if n > 100 {
			return 100
		}
		return n
	}
	d.Confidence.Overall = clamp(d.Confidence.Overall)
	d.Confidence.Title = clamp(d.Confidence.Title)
	d.Confidence.Priority = clamp(d.Confidence.Priority)
	d.Confidence.AcceptanceCriteria = clamp(d.Confidence.AcceptanceCriteria)
	return d
}
