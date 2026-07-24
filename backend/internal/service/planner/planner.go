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
	Title              string   `json:"title"`
	Description        string   `json:"description"`
	UserStory          string   `json:"userStory"`
	Priority           string   `json:"priority"` // P0 | P1 | P2 | P3
	Labels             []string `json:"labels"`
	AcceptanceCriteria []string `json:"acceptanceCriteria"`
	// Analysis is the planner's synthesized problem understanding and approach
	// rationale — the "Analysis" half of the Analysis+Plan lifecycle step.
	Analysis           string   `json:"analysis"`
	TechnicalNotes     string   `json:"technicalNotes"`
	LikelyFiles        []string `json:"likelyFiles"`
	Risks              []string `json:"risks"`
	Dependencies       []string `json:"dependencies"`
	Scope              string   `json:"scope"`
	Plan               []string `json:"plan"`
	OpenQuestions      []string `json:"openQuestions"`
	MissingInformation []string `json:"missingInformation"`
	SuggestedAgent     string   `json:"suggestedAgent"`
	// SuggestedBranch is a `<prefix>/<short-description>` branch name drawn from
	// the feature/bugfix/hotfix/refactor/chore/docs/test taxonomy.
	SuggestedBranch string `json:"suggestedBranch"`
	// SuggestedCommit is an example Conventional Commit subject
	// (`type(scope): summary`) for the first commit of this task.
	SuggestedCommit string     `json:"suggestedCommit"`
	Confidence      Confidence `json:"confidence"`
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
		return nil //nolint:nilerr // tolerant parse: an unparseable score defaults to 0
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
	b.WriteString("Work the Analysis-then-Plan method: first ANALYZE (understand the problem, the user story, ")
	b.WriteString("observable acceptance criteria, scope, risks, and the areas of code it touches), then PLAN ")
	b.WriteString("(decompose into small, dependency-ordered, independently verifiable steps).\n\n")
	b.WriteString("Respond with ONLY a single minified JSON object — no prose, no markdown, no code fences. ")
	b.WriteString("Use exactly these keys:\n")
	b.WriteString(`title (string), description (string), userStory (string, "As a ... I want ... so that ..."), `)
	b.WriteString("priority (one of \"P0\",\"P1\",\"P2\",\"P3\"), labels (string[]), acceptanceCriteria (string[]), ")
	b.WriteString("analysis (string: a short synthesized problem understanding and approach rationale), ")
	b.WriteString("technicalNotes (string), likelyFiles (string[]), risks (string[]), dependencies (string[]), ")
	b.WriteString("scope (\"XS\"|\"S\"|\"M\"|\"L\"|\"XL\"), plan (string[] of short high-level steps), ")
	b.WriteString("openQuestions (string[]), missingInformation (string[] of specific facts you need but were not given), ")
	b.WriteString("suggestedAgent (string, e.g. \"claude\"), ")
	b.WriteString("suggestedBranch (string \"<prefix>/<short-kebab-description>\" where prefix is one of ")
	b.WriteString("feature|bugfix|hotfix|refactor|chore|docs|test, chosen to match the task's nature), ")
	b.WriteString("suggestedCommit (string: a Conventional Commit subject \"type(scope): imperative summary\" ")
	b.WriteString("where type is one of feat|fix|refactor|docs|test|chore|build|ci|perf; ")
	b.WriteString("feature->feat, bugfix/hotfix->fix, others keep their name), ")
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
		// Agent models occasionally return a JavaScript-style object literal
		// despite the strict-JSON instruction (for example, `{Title: "..."}`).
		// Repair only bare object keys; values and quoted content remain
		// untouched so malformed replies still fail instead of being guessed at.
		repaired := quoteBareObjectKeys(raw)
		if repaired == raw {
			return TaskDraft{}, fmt.Errorf("planner: parse draft: %w", err)
		}
		if repairErr := json.Unmarshal([]byte(repaired), &d); repairErr != nil {
			return TaskDraft{}, fmt.Errorf("planner: parse draft: %w", err)
		}
	}
	return d, nil
}

// quoteBareObjectKeys converts JavaScript-style object keys to JSON keys while
// leaving strings and array values unchanged.
func quoteBareObjectKeys(raw string) string {
	var out strings.Builder
	out.Grow(len(raw) + 16)

	var containers []byte
	inString := false
	escaped := false
	expectKey := false

	for i := 0; i < len(raw); {
		ch := raw[i]
		if inString {
			out.WriteByte(ch)
			i++
			if escaped {
				escaped = false
			} else if ch == '\\' {
				escaped = true
			} else if ch == '"' {
				inString = false
			}
			continue
		}

		switch ch {
		case '"':
			inString = true
			expectKey = false
			out.WriteByte(ch)
			i++
		case '{', '[':
			containers = append(containers, ch)
			expectKey = ch == '{'
			out.WriteByte(ch)
			i++
		case '}', ']':
			if len(containers) > 0 {
				containers = containers[:len(containers)-1]
			}
			expectKey = false
			out.WriteByte(ch)
			i++
		case ',':
			expectKey = len(containers) > 0 && containers[len(containers)-1] == '{'
			out.WriteByte(ch)
			i++
		default:
			if expectKey && isBareKeyStart(ch) {
				end := i + 1
				for end < len(raw) && isBareKeyPart(raw[end]) {
					end++
				}
				colon := end
				for colon < len(raw) && (raw[colon] == ' ' || raw[colon] == '\t' || raw[colon] == '\r' || raw[colon] == '\n') {
					colon++
				}
				if colon < len(raw) && raw[colon] == ':' {
					out.WriteByte('"')
					out.WriteString(raw[i:end])
					out.WriteByte('"')
					out.WriteString(raw[end:colon])
					out.WriteByte(':')
					i = colon + 1
					expectKey = false
					continue
				}
			}
			out.WriteByte(ch)
			i++
		}
	}
	return out.String()
}

func isBareKeyStart(ch byte) bool {
	return ch == '_' || ch >= 'A' && ch <= 'Z' || ch >= 'a' && ch <= 'z'
}

func isBareKeyPart(ch byte) bool {
	return isBareKeyStart(ch) || ch >= '0' && ch <= '9'
}

// extractJSONObject returns the first complete JSON object, stripping markdown
// fences the model may have added. Native agent CLIs can append diagnostics or
// additional JSON after their final answer, so using the last brace would turn
// an otherwise valid draft into invalid JSON.
func extractJSONObject(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "```json")
	s = strings.TrimPrefix(s, "```")
	s = strings.TrimSuffix(s, "```")
	start := strings.Index(s, "{")
	if start < 0 {
		return ""
	}

	depth := 0
	inString := false
	escaped := false
	for i := start; i < len(s); i++ {
		ch := s[i]
		if inString {
			if escaped {
				escaped = false
				continue
			}
			if ch == '\\' {
				escaped = true
				continue
			}
			if ch == '"' {
				inString = false
			}
			continue
		}

		switch ch {
		case '"':
			inString = true
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return s[start : i+1]
			}
		}
	}
	return ""
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

	// Always emit a valid branch/commit suggestion: keep the model's value when
	// it already fits the taxonomy, otherwise derive one so the UI and worker
	// never see an empty or malformed convention.
	d.SuggestedBranch = normalizeBranch(d.SuggestedBranch, d.Title, d.Labels)
	d.SuggestedCommit = normalizeCommit(d.SuggestedCommit, d.SuggestedBranch, d.Title)
	return d
}
