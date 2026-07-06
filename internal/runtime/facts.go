package runtime

import "time"

type WorkflowStep string

const (
	StepPlanning      WorkflowStep = "planning"
	StepCoding        WorkflowStep = "coding"
	StepReview        WorkflowStep = "review"
	StepTesting       WorkflowStep = "testing"
	StepDebugging     WorkflowStep = "debugging"
	StepDocumentation WorkflowStep = "documentation"
	StepMemoryUpdate  WorkflowStep = "memory_update"
)

type ActivityState string

const (
	ActivityUnknown     ActivityState = ""
	ActivityStarting    ActivityState = "starting"
	ActivityActive      ActivityState = "active"
	ActivityIdle        ActivityState = "idle"
	ActivityWaitingUser ActivityState = "waiting_user"
	ActivityExited      ActivityState = "exited"
)

type AgentSessionFact struct {
	ID                   string        `json:"id"`
	ProjectID            string        `json:"project_id,omitempty"`
	TaskID               string        `json:"task_id"`
	Step                 WorkflowStep  `json:"step"`
	Provider             string        `json:"provider"`
	Command              string        `json:"command"`
	Args                 []string      `json:"args,omitempty"`
	CWD                  string        `json:"cwd,omitempty"`
	RuntimeHandleID      string        `json:"runtime_handle_id,omitempty"`
	AgentNativeSessionID string        `json:"agent_native_session_id,omitempty"`
	TranscriptPath       string        `json:"transcript_path,omitempty"`
	Activity             ActivityState `json:"activity_state"`
	Terminated           bool          `json:"is_terminated"`
	ExitCode             *int          `json:"exit_code,omitempty"`
	StartedAt            time.Time     `json:"started_at"`
	LastOutputAt         time.Time     `json:"last_output_at,omitempty"`
	EndedAt              time.Time     `json:"ended_at,omitempty"`
	UpdatedAt            time.Time     `json:"updated_at"`
}

type WorkspaceFact struct {
	ProjectID    string       `json:"project_id,omitempty"`
	TaskID       string       `json:"task_id"`
	SessionID    string       `json:"session_id,omitempty"`
	Step         WorkflowStep `json:"step,omitempty"`
	BranchName   string       `json:"branch_name,omitempty"`
	WorktreePath string       `json:"worktree_path,omitempty"`
	Prepared     bool         `json:"prepared"`
	PreparedAt   time.Time    `json:"prepared_at,omitempty"`
}

type TerminalFact struct {
	SessionID       string    `json:"session_id"`
	RuntimeHandleID string    `json:"runtime_handle_id"`
	Attached        bool      `json:"attached"`
	Rows            uint16    `json:"rows,omitempty"`
	Cols            uint16    `json:"cols,omitempty"`
	AttachedAt      time.Time `json:"attached_at,omitempty"`
	DetachedAt      time.Time `json:"detached_at,omitempty"`
}

type SessionFacts struct {
	Session   AgentSessionFact `json:"session"`
	Workspace WorkspaceFact    `json:"workspace,omitempty"`
	Terminal  TerminalFact     `json:"terminal,omitempty"`
}

func (f AgentSessionFact) CanRestore() bool {
	return f.AgentNativeSessionID != "" || f.TranscriptPath != ""
}

func (f AgentSessionFact) HasRuntimeHandle() bool {
	return f.RuntimeHandleID != ""
}

func (f AgentSessionFact) HasStarted() bool {
	return !f.StartedAt.IsZero()
}
