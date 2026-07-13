package domain

import (
	"regexp"
	"strings"
	"unicode"
)

var nonBranchChars = regexp.MustCompile(`[^a-z0-9._/-]+`)

// TaskSuggestions are generated when a task is created and persisted with the
// session so finalization does not depend on an agent inventing handoff text.
type TaskSuggestions struct {
	Branch        string `json:"branch,omitempty"`
	CommitMessage string `json:"commitMessage,omitempty"`
	PRTitle       string `json:"prTitle,omitempty"`
}

// SuggestTask derives branch, commit, and PR-title defaults from the task title
// and prompt. Callers may still let users edit the values before spawning.
func SuggestTask(title, prompt string) TaskSuggestions {
	seed := cleanTaskTitle(title)
	if seed == "" {
		seed = firstPromptLine(prompt)
	}
	if seed == "" {
		seed = "update-task"
	}
	prefix := "feature"
	commitType := "feat"
	if looksLikeFix(seed) || looksLikeFix(prompt) {
		prefix = "bugfix"
		commitType = "fix"
	}
	slug := branchSlug(seed)
	if slug == "" {
		slug = "update-task"
	}
	summary := commitSummary(seed)
	scope := inferScope(seed + " " + prompt)
	commit := commitType + ": " + summary
	if scope != "" {
		commit = commitType + "(" + scope + "): " + summary
	}
	return TaskSuggestions{
		Branch:        prefix + "/" + slug,
		CommitMessage: commit,
		PRTitle:       commit,
	}
}

func cleanTaskTitle(s string) string {
	s = strings.TrimSpace(s)
	if i := strings.Index(s, ":"); i >= 0 {
		head := strings.ToLower(strings.TrimSpace(s[:i]))
		if head == "fix issue" || head == "issue" || head == "task" || strings.HasPrefix(head, "github") {
			s = strings.TrimSpace(s[i+1:])
		}
	}
	s = strings.TrimPrefix(s, "#")
	return strings.TrimSpace(s)
}

func firstPromptLine(prompt string) string {
	for _, line := range strings.Split(prompt, "\n") {
		if trimmed := strings.TrimSpace(line); trimmed != "" {
			return cleanTaskTitle(trimmed)
		}
	}
	return ""
}

func looksLikeFix(s string) bool {
	low := strings.ToLower(s)
	for _, marker := range []string{"bug", "fix", "defect", "regression", "crash", "error", "fail", "failure", "broken", "missing", "credential"} {
		if strings.Contains(low, marker) {
			return true
		}
	}
	return false
}

func branchSlug(s string) string {
	var b strings.Builder
	lastDash := false
	for _, r := range strings.ToLower(s) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			lastDash = false
		case r == '/' || r == '.' || r == '_':
			if !lastDash {
				b.WriteByte('-')
				lastDash = true
			}
		case unicode.IsSpace(r) || unicode.IsPunct(r) || unicode.IsSymbol(r):
			if !lastDash {
				b.WriteByte('-')
				lastDash = true
			}
		}
	}
	out := strings.Trim(b.String(), "-")
	out = nonBranchChars.ReplaceAllString(out, "-")
	out = strings.Trim(out, "-")
	if len(out) > 64 {
		out = strings.Trim(out[:64], "-")
	}
	return out
}

func commitSummary(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimRight(s, ". ")
	if s == "" {
		return "update task"
	}
	runes := []rune(s)
	if len(runes) > 1 && unicode.IsUpper(runes[0]) && unicode.IsUpper(runes[1]) {
		return s
	}
	runes[0] = unicode.ToLower(runes[0])
	out := string(runes)
	if len(out) > 72 {
		out = strings.TrimRight(out[:72], " ")
	}
	return out
}

func inferScope(s string) string {
	low := strings.ToLower(s)
	for _, candidate := range []string{"tracker", "orchestrator", "session", "cli", "github", "auth", "app", "api", "docs"} {
		if strings.Contains(low, candidate) {
			return candidate
		}
	}
	return ""
}
