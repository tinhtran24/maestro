package agents

type WorkflowStepConfig struct {
	ID                            string            `json:"id"`
	Enabled                       bool              `json:"enabled"`
	Provider                      string            `json:"provider"`
	Command                       string            `json:"command"`
	WorkingDirectoryMode          string            `json:"working_directory_mode"`
	AutoStartTerminal             bool              `json:"auto_start_terminal"`
	RequireApprovalBeforeNextStep bool              `json:"approval_required"`
	Environment                   map[string]string `json:"env,omitempty"`
	Timeout                       string            `json:"timeout,omitempty"`
	Permissions                   []string          `json:"permissions,omitempty"`
}
