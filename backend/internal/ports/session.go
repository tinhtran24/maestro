package ports

import (
	"errors"

	"github.com/tinhtran/thanos/backend/internal/domain"
)

// ErrSessionNotFound reports an observation for an unknown session id.
var ErrSessionNotFound = errors.New("session not found")

// SpawnConfig is the request to start a new session: which project/issue, which
// agent harness, and the branch/prompt the agent launches with.
type SpawnConfig struct {
	ProjectID domain.ProjectID
	IssueID   domain.IssueID
	Kind      domain.SessionKind
	Harness   domain.AgentHarness
	Branch    string
	Prompt    string
	// CommitMessage and PRTitle are task-creation suggestions persisted for
	// commit and pull-request naming. They are metadata, not agent output.
	CommitMessage string
	PRTitle       string
	// DisplayName is the user-facing sidebar label. Empty falls back to the
	// session id in the read model (e.g. orchestrator sessions).
	DisplayName string
	// Model is a task-level native CLI override. Empty retains the project's
	// configured model (or the CLI's own default).
	Model string
}
