package bridge

type WorkbenchSnapshot struct {
	Project     *ProjectInfo        `json:"project,omitempty"`
	Features    []FeatureInfo       `json:"features"`
	Tasks       []TaskInfo          `json:"tasks"`
	Plans       []ExecutionPlanInfo `json:"plans"`
	Sessions    []AgentSessionInfo  `json:"sessions"`
	Reviews     []ReviewInfo        `json:"reviews"`
	MemoryNodes []MemoryNodeInfo    `json:"memory_nodes"`
	Skills      []SkillInfo         `json:"skills"`
	SkillRuns   []SkillRunInfo      `json:"skill_runs"`
}

type ProjectInfo struct {
	ID             string            `json:"id"`
	Name           string            `json:"name"`
	RootPath       string            `json:"root_path"`
	GitRemoteURL   string            `json:"git_remote_url,omitempty"`
	DefaultBranch  string            `json:"default_branch,omitempty"`
	WorktreeRoot   string            `json:"worktree_root,omitempty"`
	PackageManager string            `json:"package_manager,omitempty"`
	DevCommand     string            `json:"dev_command,omitempty"`
	TestCommand    string            `json:"test_command,omitempty"`
	CreatedAt      string            `json:"created_at,omitempty"`
	UpdatedAt      string            `json:"updated_at,omitempty"`
	Repos          []string          `json:"repos"`
	Settings       map[string]string `json:"settings"`
}

type FeatureInfo struct {
	ID          string  `json:"id"`
	ProjectID   string  `json:"project_id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Status      string  `json:"status"`
	PlanGraphID *string `json:"plan_graph_id,omitempty"`
	CreatedAt   string  `json:"created_at"`
}

type TaskInfo struct {
	ID              string   `json:"id"`
	FeatureID       string   `json:"feature_id"`
	ParentTaskID    *string  `json:"parent_task_id,omitempty"`
	Title           string   `json:"title"`
	Description     string   `json:"description"`
	Status          string   `json:"status"`
	Priority        string   `json:"priority"`
	AssignedAgent   string   `json:"assigned_agent"`
	ExecutorProfile string   `json:"executor_profile"`
	WorktreePath    string   `json:"worktree_path"`
	BranchName      string   `json:"branch_name"`
	ReviewApproved  bool     `json:"review_approved"`
	TestsPassed     bool     `json:"tests_passed"`
	UpdatedAt       string   `json:"updated_at"`
	Tags            []string `json:"tags"`
	Progress        uint8    `json:"progress"`
}

type PlanStepInfo struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
}

type ExecutionPlanInfo struct {
	ID             string         `json:"id"`
	TaskID         string         `json:"task_id"`
	Summary        string         `json:"summary"`
	Steps          []PlanStepInfo `json:"steps"`
	Risks          []string       `json:"risks"`
	FilesToTouch   []string       `json:"files_to_touch"`
	TestStrategy   []string       `json:"test_strategy"`
	ApprovalStatus string         `json:"approval_status"`
}

type AgentSessionInfo struct {
	ID                  string   `json:"id"`
	TaskID              string   `json:"task_id"`
	AgentType           string   `json:"agent_type"`
	Provider            string   `json:"provider"`
	Command             string   `json:"command"`
	Args                []string `json:"args"`
	Status              string   `json:"status"`
	PTYSessionID        string   `json:"pty_session_id"`
	ConversationLogPath string   `json:"conversation_log_path"`
	WorktreePath        string   `json:"worktree_path"`
	StartedAt           uint64   `json:"started_at"`
	EndedAt             *uint64  `json:"ended_at,omitempty"`
}

type TestRunInfo struct {
	TaskID  string `json:"task_id"`
	Command string `json:"command"`
	Status  string `json:"status"`
	Stdout  string `json:"stdout"`
	Stderr  string `json:"stderr"`
	Code    *int   `json:"code"`
}

type ReviewInfo struct {
	ID            string        `json:"id"`
	TaskID        string        `json:"task_id"`
	DiffSummary   string        `json:"diff_summary"`
	ChangedFiles  []string      `json:"changed_files"`
	TestResults   []TestRunInfo `json:"test_results"`
	ReviewerNotes string        `json:"reviewer_notes"`
	Status        string        `json:"status"`
}

type MemoryNodeInfo struct {
	ID        string   `json:"id"`
	ProjectID string   `json:"project_id"`
	NodeType  string   `json:"node_type"`
	Title     string   `json:"title"`
	Content   string   `json:"content"`
	Links     []string `json:"links"`
	CreatedAt uint64   `json:"created_at"`
}

type SkillInfo struct {
	ID               string   `json:"id"`
	ProjectID        *string  `json:"project_id,omitempty"`
	Name             string   `json:"name"`
	Path             string   `json:"path"`
	Description      string   `json:"description"`
	AppliesTo        []string `json:"applies_to"`
	Agents           []string `json:"agents"`
	Version          *string  `json:"version,omitempty"`
	Source           string   `json:"source"`
	RequiredEvidence []string `json:"required_evidence"`
	ExitCriteria     []string `json:"exit_criteria"`
	Enabled          bool     `json:"enabled"`
	Trusted          bool     `json:"trusted"`
}

type SkillRunInfo struct {
	ID             string            `json:"id"`
	TaskID         string            `json:"task_id"`
	SkillID        string            `json:"skill_id"`
	AgentSessionID *string           `json:"agent_session_id,omitempty"`
	Status         string            `json:"status"`
	EvidenceJSON   map[string]string `json:"evidence_json"`
	StartedAt      string            `json:"started_at"`
	CompletedAt    *string           `json:"completed_at,omitempty"`
}

type AgentCandidateInfo struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Command   string  `json:"command"`
	Installed bool    `json:"installed"`
	Path      *string `json:"path,omitempty"`
	Status    string  `json:"status"`
	Version   *string `json:"version,omitempty"`
	AgentType string  `json:"agent_type"`
	Enabled   bool    `json:"enabled"`
	SetupHint string  `json:"setup_hint"`
}

type ProjectSetupRequest struct {
	RootPath       string `json:"root_path"`
	Name           string `json:"name"`
	GitRemoteURL   string `json:"git_remote_url,omitempty"`
	DefaultBranch  string `json:"default_branch,omitempty"`
	WorktreeRoot   string `json:"worktree_root,omitempty"`
	PackageManager string `json:"package_manager,omitempty"`
	DevCommand     string `json:"dev_command,omitempty"`
	TestCommand    string `json:"test_command,omitempty"`
}

type TaskPlanRequest struct {
	Workspace string `json:"workspace"`
	TaskID    string `json:"task_id"`
}

type MemorySearchRequest struct {
	Workspace string `json:"workspace"`
	Query     string `json:"query"`
}
