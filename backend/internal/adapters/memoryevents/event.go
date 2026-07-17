package memoryevents

import "time"

// SchemaVersion is the current on-disk event schema version. It gates schema
// evolution: readers dispatch on Event.V, and a breaking field change bumps this
// rather than rewriting prior lines (the log is append-only and immutable).
const SchemaVersion = 1

// Event is one immutable line in a project's events.jsonl memory log. Exactly one
// Event is appended per completed task. IDs and timestamps are supplied by the
// caller; this package never mints them (it has no clock of its own).
type Event struct {
	V          int       `json:"v"`
	ID         string    `json:"id"`
	Type       string    `json:"type"`
	OccurredAt time.Time `json:"occurred_at"`
	Task       Task      `json:"task"`
}

// Task is the completion memory recorded for a single finished session.
type Task struct {
	ID             string   `json:"id"`
	SessionID      string   `json:"session_id"`
	ProjectID      string   `json:"project_id"`
	Kind           string   `json:"kind"`
	Harness        string   `json:"harness,omitempty"`
	Intent         string   `json:"intent,omitempty"`
	TaskType       string   `json:"task_type,omitempty"`
	Branch         string   `json:"branch,omitempty"`
	BaseSHA        string   `json:"base_sha,omitempty"`
	HeadSHA        string   `json:"head_sha,omitempty"`
	ChangedFiles   []string `json:"changed_files,omitempty"`
	ChangedTests   []string `json:"changed_tests,omitempty"`
	ChangedSymbols []string `json:"changed_symbols,omitempty"`
	PRs            []PRRef  `json:"prs,omitempty"`
	// Decisions holds protected architectural decisions surfaced by the task. It
	// is empty in the file-level (Phase 1) implementation; richer structured
	// decisions may arrive under a bumped SchemaVersion.
	Decisions []string `json:"decisions,omitempty"`
}

// PRRef is a compact reference to a pull request attributed to the task.
type PRRef struct {
	URL    string `json:"url"`
	Number int    `json:"number,omitempty"`
	State  string `json:"state,omitempty"`
}

// Event type constants for Event.Type.
const (
	// TypeTaskCompleted records a finished session's completion memory.
	TypeTaskCompleted = "task.completed"
)
