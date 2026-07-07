package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

type RealProvider struct {
	now func() time.Time
}

func NewRealProvider() *RealProvider {
	return &RealProvider{now: time.Now}
}

type WorkspaceInfo struct {
	WorkspaceID   string                `json:"workspaceId"`
	DataKey       string                `json:"dataKey"`
	Name          string                `json:"name"`
	Path          string                `json:"path"`
	Folders       []WorkspaceFolderInfo `json:"folders"`
	DefaultBranch string                `json:"defaultBranch"`
	Tasks         []TaskInfo            `json:"tasks"`
	Specs         []SpecNodeInfo        `json:"specs"`
	Routines      []RoutineInfo         `json:"routines"`
	Automation    AutomationInfo        `json:"automation"`
	Agents        []AgentRoleInfo       `json:"agents"`
	Flows         []FlowInfo            `json:"flows"`
	Events        []EventInfo           `json:"events"`
	Providers     []ProviderInfo        `json:"providers"`
	Diagnostics   []DiagnosticInfo      `json:"diagnostics"`
}

type TaskInfo struct {
	SchemaVersion   int              `json:"schema_version"`
	ID              string           `json:"id"`
	Title           string           `json:"title"`
	Prompt          string           `json:"prompt"`
	Status          string           `json:"status"`
	Flow            string           `json:"flow"`
	Agent           string           `json:"agent"`
	Branch          string           `json:"branch"`
	Worktree        string           `json:"worktree"`
	UpdatedAt       string           `json:"updatedAt"`
	UsageUSD        float64          `json:"usageUsd"`
	Archived        bool             `json:"archived"`
	Deleted         bool             `json:"deleted"`
	Tombstone       bool             `json:"tombstone"`
	Dependencies    []string         `json:"dependencies"`
	Blocked         bool             `json:"blocked"`
	PromptHistory   []PromptRecord   `json:"promptHistory"`
	FeedbackHistory []FeedbackRecord `json:"feedbackHistory"`
	RetryHistory    []RetryRecord    `json:"retryHistory"`
	Turns           []TaskTurnInfo   `json:"turns"`
	LastTurn        *TaskTurnInfo    `json:"lastTurn,omitempty"`
	LastOutput      string           `json:"lastOutput,omitempty"`
	TestsPassed     bool             `json:"testsPassed"`
	LastTestResult  *TestResultInfo  `json:"lastTestResult,omitempty"`
	Commit          *CommitInfo      `json:"commit,omitempty"`
	Oversight       *OversightInfo   `json:"oversight,omitempty"`
	FailureCategory string           `json:"failureCategory"`
	CreatedAt       string           `json:"createdAt"`
}

type PromptRecord struct {
	At     string `json:"at"`
	Prompt string `json:"prompt"`
}

type FeedbackRecord struct {
	At      string `json:"at"`
	Message string `json:"message"`
}

type RetryRecord struct {
	At     string `json:"at"`
	Reason string `json:"reason"`
}

type TaskTurnInfo struct {
	ID              string  `json:"id"`
	TaskID          string  `json:"taskId"`
	Step            string  `json:"step"`
	ProviderID      string  `json:"providerId"`
	SessionID       string  `json:"sessionId,omitempty"`
	Status          string  `json:"status"`
	Worktree        string  `json:"worktree"`
	StartedAt       string  `json:"startedAt"`
	EndedAt         string  `json:"endedAt,omitempty"`
	StdoutPath      string  `json:"stdoutPath,omitempty"`
	StderrPath      string  `json:"stderrPath,omitempty"`
	TranscriptPath  string  `json:"transcriptPath,omitempty"`
	StopReason      string  `json:"stopReason,omitempty"`
	UsageUSD        float64 `json:"usageUsd"`
	FailureCategory string  `json:"failureCategory,omitempty"`
	AutoContinue    bool    `json:"autoContinue"`
}

type TestResultInfo struct {
	ID          string `json:"id"`
	TaskID      string `json:"taskId"`
	ProviderID  string `json:"providerId,omitempty"`
	Command     string `json:"command"`
	Status      string `json:"status"`
	Passed      bool   `json:"passed"`
	OutputPath  string `json:"outputPath"`
	Output      string `json:"output"`
	ExitCode    int    `json:"exitCode"`
	PassPattern string `json:"passPattern,omitempty"`
	FailPattern string `json:"failPattern,omitempty"`
	StartedAt   string `json:"startedAt"`
	EndedAt     string `json:"endedAt"`
}

type CommitInfo struct {
	Hash        string `json:"hash,omitempty"`
	Summary     string `json:"summary"`
	Message     string `json:"message"`
	Diff        string `json:"diff"`
	DiffStat    string `json:"diffStat"`
	Approved    bool   `json:"approved"`
	Committed   bool   `json:"committed"`
	CommittedAt string `json:"committedAt,omitempty"`
}

type OversightInfo struct {
	SchemaVersion int      `json:"schema_version"`
	ID            string   `json:"id"`
	TaskID        string   `json:"taskId"`
	Status        string   `json:"status"`
	Summary       string   `json:"summary"`
	Phases        []string `json:"phases"`
	Risks         []string `json:"risks"`
	ChangedFiles  []string `json:"changedFiles"`
	Commands      []string `json:"commands"`
	TestResult    string   `json:"testResult"`
	UsageUSD      float64  `json:"usageUsd"`
	GeneratedAt   string   `json:"generatedAt"`
	Path          string   `json:"path"`
	TestPath      string   `json:"testPath,omitempty"`
}

type CreateTaskRequest struct {
	Root         string   `json:"root"`
	Title        string   `json:"title"`
	Prompt       string   `json:"prompt"`
	Flow         string   `json:"flow"`
	Agent        string   `json:"agent"`
	Status       string   `json:"status"`
	Dependencies []string `json:"dependencies"`
}

type UpdateTaskStatusRequest struct {
	Root            string `json:"root"`
	TaskID          string `json:"taskId"`
	Status          string `json:"status"`
	Feedback        string `json:"feedback"`
	FailureCategory string `json:"failureCategory"`
}

type StartTaskTurnRequest struct {
	Root           string `json:"root"`
	TaskID         string `json:"taskId"`
	Step           string `json:"step"`
	ProviderID     string `json:"providerId"`
	SessionID      string `json:"sessionId"`
	TranscriptPath string `json:"transcriptPath"`
}

type FinishTaskTurnRequest struct {
	Root         string  `json:"root"`
	TaskID       string  `json:"taskId"`
	TurnID       string  `json:"turnId"`
	Status       string  `json:"status"`
	Stdout       string  `json:"stdout"`
	Stderr       string  `json:"stderr"`
	StopReason   string  `json:"stopReason"`
	UsageUSD     float64 `json:"usageUsd"`
	ExitCode     int     `json:"exitCode"`
	AutoContinue bool    `json:"autoContinue"`
}

type ResumeTaskTurnRequest struct {
	Root     string `json:"root"`
	TaskID   string `json:"taskId"`
	Feedback string `json:"feedback"`
}

type RunTaskVerificationRequest struct {
	Root        string `json:"root"`
	TaskID      string `json:"taskId"`
	Command     string `json:"command"`
	ProviderID  string `json:"providerId"`
	PassPattern string `json:"passPattern"`
	FailPattern string `json:"failPattern"`
}

type PrepareTaskCommitRequest struct {
	Root    string `json:"root"`
	TaskID  string `json:"taskId"`
	Message string `json:"message"`
}

type CommitTaskChangesRequest struct {
	Root     string `json:"root"`
	TaskID   string `json:"taskId"`
	Message  string `json:"message"`
	Approved bool   `json:"approved"`
}

type RegenerateOversightRequest struct {
	Root   string `json:"root"`
	TaskID string `json:"taskId"`
}

type BatchCreateTasksRequest struct {
	Root  string              `json:"root"`
	Tasks []CreateTaskRequest `json:"tasks"`
}

type SearchTasksRequest struct {
	Root            string `json:"root"`
	Query           string `json:"query"`
	IncludeArchived bool   `json:"includeArchived"`
	IncludeDeleted  bool   `json:"includeDeleted"`
}

type UpdateTaskFlagsRequest struct {
	Root      string `json:"root"`
	TaskID    string `json:"taskId"`
	Archived  bool   `json:"archived"`
	Deleted   bool   `json:"deleted"`
	Tombstone bool   `json:"tombstone"`
}

type SpecNodeInfo struct {
	ID        string         `json:"id"`
	Title     string         `json:"title"`
	State     string         `json:"state"`
	Path      string         `json:"path"`
	Body      string         `json:"body"`
	UpdatedAt string         `json:"updatedAt"`
	Children  []SpecNodeInfo `json:"children"`
}

type CreateSpecRequest struct {
	Root       string `json:"root"`
	Title      string `json:"title"`
	Body       string `json:"body"`
	State      string `json:"state"`
	ParentPath string `json:"parentPath"`
}

type UpdateSpecRequest struct {
	Root  string `json:"root"`
	Path  string `json:"path"`
	Title string `json:"title"`
	Body  string `json:"body"`
	State string `json:"state"`
}

type DispatchSpecsRequest struct {
	Root string `json:"root"`
	Path string `json:"path"`
}

type UndoPlanningChangeRequest struct {
	Root string `json:"root"`
}

type RoutineInfo struct {
	SchemaVersion int    `json:"schema_version"`
	ID            string `json:"id"`
	Name          string `json:"name"`
	Prompt        string `json:"prompt"`
	Flow          string `json:"flow"`
	Schedule      string `json:"schedule"`
	Enabled       bool   `json:"enabled"`
	UpdatedAt     string `json:"updatedAt"`
}

type UpsertRoutineRequest struct {
	Root     string `json:"root"`
	ID       string `json:"id"`
	Name     string `json:"name"`
	Prompt   string `json:"prompt"`
	Flow     string `json:"flow"`
	Schedule string `json:"schedule"`
	Enabled  bool   `json:"enabled"`
}

type AutomationInfo struct {
	SchemaVersion int  `json:"schema_version"`
	AutoImplement bool `json:"autoImplement"`
	AutoTest      bool `json:"autoTest"`
	AutoSubmit    bool `json:"autoSubmit"`
	AutoRetry     bool `json:"autoRetry"`
}

type SaveAutomationRequest struct {
	Root       string         `json:"root"`
	Automation AutomationInfo `json:"automation"`
}

type AgentRoleInfo struct {
	ID           string   `json:"id"`
	Role         string   `json:"role"`
	Harness      string   `json:"harness"`
	Model        string   `json:"model"`
	Capabilities []string `json:"capabilities"`
	ReadOnly     bool     `json:"readOnly"`
	Source       string   `json:"source,omitempty"`
}

type FlowInfo struct {
	ID             string     `json:"id"`
	Name           string     `json:"name"`
	Steps          []string   `json:"steps"`
	ParallelGroups [][]string `json:"parallelGroups,omitempty"`
	ReadOnly       bool       `json:"readOnly"`
	Source         string     `json:"source,omitempty"`
}

type EventInfo struct {
	SchemaVersion int    `json:"schema_version"`
	ID            string `json:"id"`
	At            string `json:"at"`
	Kind          string `json:"kind"`
	Message       string `json:"message"`
}

type ProviderInfo struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Command     string  `json:"command"`
	Status      string  `json:"status"`
	Path        *string `json:"path,omitempty"`
	Version     *string `json:"version,omitempty"`
	Type        string  `json:"type"`
	SetupHint   string  `json:"setupHint"`
	SupportsRun bool    `json:"supportsRun"`
}

type AgentCandidateInfo = ProviderInfo

type DiagnosticInfo struct {
	Kind    string `json:"kind"`
	Message string `json:"message"`
}

func (p *RealProvider) CurrentWorkspaceFolder() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	root, err := gitTopLevel(wd)
	if err == nil && root != "" {
		return root, nil
	}
	return wd, nil
}

func (p *RealProvider) LoadWorkspace(root string) (*WorkspaceInfo, error) {
	if strings.TrimSpace(root) == "" {
		var err error
		root, err = p.CurrentWorkspaceFolder()
		if err != nil {
			return nil, err
		}
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	if stat, err := os.Stat(abs); err != nil {
		return nil, err
	} else if !stat.IsDir() {
		return nil, fmt.Errorf("workspace path is not a directory: %s", abs)
	}

	gitRoot, err := gitTopLevel(abs)
	if err == nil && gitRoot != "" {
		abs = gitRoot
	}

	providers, providerErr := p.DetectAgentCLIs(context.Background())
	diagnostics := make([]DiagnosticInfo, 0)
	if providerErr != nil {
		diagnostics = append(diagnostics, DiagnosticInfo{Kind: "providers", Message: providerErr.Error()})
	}
	store := NewWorkspaceStore(abs)
	_ = store.WithLock(func() error {
		diagnostics = append(diagnostics, store.MigrateAndValidate()...)
		return nil
	})

	specs, err := loadSpecs(abs)
	if err != nil {
		diagnostics = append(diagnostics, DiagnosticInfo{Kind: "specs", Message: err.Error()})
	}
	tasks, err := loadTasks(abs)
	if err != nil {
		diagnostics = append(diagnostics, DiagnosticInfo{Kind: "tasks", Message: err.Error()})
	}
	events, err := loadEvents(abs)
	if err != nil {
		diagnostics = append(diagnostics, DiagnosticInfo{Kind: "events", Message: err.Error()})
	}
	routines, err := loadRoutines(abs)
	if err != nil {
		diagnostics = append(diagnostics, DiagnosticInfo{Kind: "routines", Message: err.Error()})
	}
	automation, err := loadAutomation(abs)
	if err != nil {
		diagnostics = append(diagnostics, DiagnosticInfo{Kind: "automation", Message: err.Error()})
	}
	agents, err := loadAgentRoles(abs, providers)
	if err != nil {
		diagnostics = append(diagnostics, DiagnosticInfo{Kind: "agents", Message: err.Error()})
		agents = builtinAgentRoles(providers)
	}
	flows, err := loadFlows(abs)
	if err != nil {
		diagnostics = append(diagnostics, DiagnosticInfo{Kind: "flows", Message: err.Error()})
		flows = builtinFlows()
	}
	tasks = normalizeTaskFlows(tasks, flows)
	if len(events) == 0 {
		events = append(events, EventInfo{
			ID:      "workspace-loaded",
			At:      p.now().Format("15:04"),
			Kind:    "WorkspaceLoaded",
			Message: "Loaded real workspace metadata from local filesystem.",
		})
	}

	return &WorkspaceInfo{
		Name:          workspaceName(abs),
		Path:          abs,
		Folders:       []WorkspaceFolderInfo{{ID: "folder-" + stableID(abs), Path: abs, Label: workspaceName(abs)}},
		DefaultBranch: currentBranch(abs),
		Tasks:         tasks,
		Specs:         specs,
		Routines:      routines,
		Automation:    automation,
		Agents:        agents,
		Flows:         flows,
		Events:        events,
		Providers:     providers,
		Diagnostics:   diagnostics,
	}, nil
}

func (p *RealProvider) CreateTask(req CreateTaskRequest) (*TaskInfo, error) {
	root, err := normalizeWorkspaceRoot(req.Root)
	if err != nil {
		return nil, err
	}
	title := strings.TrimSpace(req.Title)
	prompt := strings.TrimSpace(req.Prompt)
	if title == "" {
		title = firstLine(prompt)
	}
	if title == "" {
		return nil, fmt.Errorf("task title or prompt is required")
	}
	now := p.now().UTC()
	id := fmt.Sprintf("task-%s", stableID(fmt.Sprintf("%d-%s", now.UnixNano(), title)))
	flows, _ := loadFlows(root)
	task := TaskInfo{
		SchemaVersion: SchemaVersionTask,
		ID:            id,
		Title:         title,
		Prompt:        prompt,
		Status:        normalizeTaskStatus(req.Status),
		Flow:          normalizeFlowID(req.Flow, flows),
		Agent:         fallback(req.Agent, "unassigned"),
		Branch:        fmt.Sprintf("task/%s", id),
		Worktree:      filepath.ToSlash(filepath.Join(".thanos", "worktrees", id)),
		UpdatedAt:     now.Format(time.RFC3339),
		CreatedAt:     now.Format(time.RFC3339),
		Dependencies:  normalizeTaskDependencies(req.Dependencies),
		PromptHistory: []PromptRecord{{At: now.Format(time.RFC3339), Prompt: prompt}},
	}
	store := NewWorkspaceStore(root)
	if err := store.WithLock(func() error {
		if err := writeTask(store, task); err != nil {
			return err
		}
		return store.AppendEvent(EventInfo{ID: "event-" + id, At: now.Format(time.RFC3339), Kind: "TaskCreated", Message: title})
	}); err != nil {
		return nil, err
	}
	return &task, nil
}

func (p *RealProvider) BatchCreateTasks(req BatchCreateTasksRequest) ([]TaskInfo, error) {
	root, err := normalizeWorkspaceRoot(req.Root)
	if err != nil {
		return nil, err
	}
	if len(req.Tasks) == 0 {
		return []TaskInfo{}, nil
	}
	created := make([]TaskInfo, 0, len(req.Tasks))
	for index, item := range req.Tasks {
		item.Root = root
		if len(item.Dependencies) == 0 && index > 0 {
			item.Dependencies = []string{created[index-1].ID}
		}
		task, err := p.CreateTask(item)
		if err != nil {
			return nil, err
		}
		created = append(created, *task)
	}
	return created, nil
}

func (p *RealProvider) SearchTasks(req SearchTasksRequest) ([]TaskInfo, error) {
	root, err := normalizeWorkspaceRoot(req.Root)
	if err != nil {
		return nil, err
	}
	tasks, err := loadTasks(root)
	if err != nil {
		return nil, err
	}
	query := strings.ToLower(strings.TrimSpace(req.Query))
	out := make([]TaskInfo, 0, len(tasks))
	for _, task := range markBlockedTasks(tasks) {
		if task.Archived && !req.IncludeArchived {
			continue
		}
		if (task.Deleted || task.Tombstone) && !req.IncludeDeleted {
			continue
		}
		if query == "" || strings.Contains(strings.ToLower(task.Title), query) || strings.Contains(strings.ToLower(task.Prompt), query) || strings.Contains(strings.ToLower(task.ID), query) {
			out = append(out, task)
		}
	}
	return out, nil
}

func (p *RealProvider) UpdateTaskFlags(req UpdateTaskFlagsRequest) (*TaskInfo, error) {
	root, err := normalizeWorkspaceRoot(req.Root)
	if err != nil {
		return nil, err
	}
	var updated *TaskInfo
	store := NewWorkspaceStore(root)
	err = store.WithLock(func() error {
		tasks, err := loadTasks(root)
		if err != nil {
			return err
		}
		for _, task := range tasks {
			if task.ID != req.TaskID {
				continue
			}
			task = hydrateTask(task)
			task.Archived = req.Archived
			task.Deleted = req.Deleted
			task.Tombstone = req.Tombstone
			task.UpdatedAt = p.now().UTC().Format(time.RFC3339)
			if task.Tombstone {
				task.Status = "cancelled"
			}
			if err := writeTask(store, task); err != nil {
				return err
			}
			if err := store.AppendEvent(EventInfo{ID: "event-flags-" + stableID(task.ID+"-"+task.UpdatedAt), At: task.UpdatedAt, Kind: "TaskFlagsUpdated", Message: fmt.Sprintf("%s archived=%t deleted=%t tombstone=%t", task.Title, task.Archived, task.Deleted, task.Tombstone)}); err != nil {
				return err
			}
			copied := task
			updated = &copied
			return nil
		}
		return fmt.Errorf("task not found: %s", req.TaskID)
	})
	if err != nil {
		return nil, err
	}
	return updated, nil
}

func (p *RealProvider) UpdateTaskStatus(req UpdateTaskStatusRequest) (*TaskInfo, error) {
	root, err := normalizeWorkspaceRoot(req.Root)
	if err != nil {
		return nil, err
	}
	nextStatus := normalizeTaskStatus(req.Status)
	var updated *TaskInfo
	store := NewWorkspaceStore(root)
	err = store.WithLock(func() error {
		tasks, err := loadTasks(root)
		if err != nil {
			return err
		}
		for _, task := range tasks {
			if task.ID != req.TaskID {
				continue
			}
			task = hydrateTask(task)
			if task.Blocked && nextStatus == "in_progress" {
				return fmt.Errorf("task is blocked by unfinished dependencies: %s", task.ID)
			}
			if err := validateTaskTransition(task.Status, nextStatus); err != nil {
				return err
			}
			if nextStatus == "done" {
				automation, err := loadAutomation(root)
				if err != nil {
					return err
				}
				if automation.AutoTest && !task.TestsPassed {
					return fmt.Errorf("task %s cannot be done before passing verification", task.ID)
				}
				if task.Commit == nil || !task.Commit.Committed || task.Commit.Hash == "" {
					return fmt.Errorf("task %s cannot be done before committing task worktree changes", task.ID)
				}
			}
			task.SchemaVersion = SchemaVersionTask
			previousStatus := task.Status
			task.Status = nextStatus
			task.UpdatedAt = p.now().UTC().Format(time.RFC3339)
			if strings.TrimSpace(req.Feedback) != "" {
				task.FeedbackHistory = append(task.FeedbackHistory, FeedbackRecord{At: task.UpdatedAt, Message: strings.TrimSpace(req.Feedback)})
			}
			if previousStatus == "failed" && nextStatus == "in_progress" {
				task.RetryHistory = append(task.RetryHistory, RetryRecord{At: task.UpdatedAt, Reason: fallback(req.Feedback, "manual retry")})
			}
			if nextStatus == "failed" {
				task.FailureCategory = fallback(req.FailureCategory, "unknown")
			}
			if nextStatus == "cancelled" {
				task.Tombstone = true
			}
			if shouldGenerateOversight(task.Status) {
				oversight, err := writeOversightArtifacts(root, task, p.now().UTC())
				if err != nil {
					return err
				}
				task.Oversight = &oversight
			}
			if err := writeTask(store, task); err != nil {
				return err
			}
			if err := store.AppendEvent(EventInfo{ID: "event-" + stableID(task.ID+"-"+task.Status+"-"+task.UpdatedAt), At: task.UpdatedAt, Kind: "TaskMoved", Message: fmt.Sprintf("%s -> %s", task.Title, task.Status)}); err != nil {
				return err
			}
			copied := task
			updated = &copied
			return nil
		}
		return fmt.Errorf("task not found: %s", req.TaskID)
	})
	if err != nil {
		return nil, err
	}
	return updated, nil
}

func (p *RealProvider) StartTaskTurn(req StartTaskTurnRequest) (*TaskInfo, error) {
	root, err := normalizeWorkspaceRoot(req.Root)
	if err != nil {
		return nil, err
	}
	var updated *TaskInfo
	now := p.now().UTC().Format(time.RFC3339)
	store := NewWorkspaceStore(root)
	err = store.WithLock(func() error {
		tasks, err := loadTasks(root)
		if err != nil {
			return err
		}
		for _, task := range tasks {
			if task.ID != req.TaskID {
				continue
			}
			task = hydrateTask(task)
			if task.Blocked {
				return fmt.Errorf("task is blocked by unfinished dependencies: %s", task.ID)
			}
			if task.Status == "done" || task.Status == "cancelled" {
				return fmt.Errorf("task %s cannot start a turn from %s", task.ID, task.Status)
			}
			if strings.TrimSpace(task.Worktree) == "" {
				return fmt.Errorf("task %s cannot start without an isolated worktree", task.ID)
			}
			if task.Status != "in_progress" {
				if err := validateTaskTransition(task.Status, "in_progress"); err != nil {
					return err
				}
			}
			if task.Status == "failed" {
				task.RetryHistory = append(task.RetryHistory, RetryRecord{At: now, Reason: "start new turn"})
			}
			turnID := fmt.Sprintf("turn-%03d", len(task.Turns)+1)
			turn := TaskTurnInfo{
				ID:             turnID,
				TaskID:         task.ID,
				Step:           fallback(req.Step, "Implementation"),
				ProviderID:     fallback(req.ProviderID, task.Agent),
				SessionID:      strings.TrimSpace(req.SessionID),
				Status:         "running",
				Worktree:       task.Worktree,
				StartedAt:      now,
				TranscriptPath: strings.TrimSpace(req.TranscriptPath),
			}
			task.Turns = append(task.Turns, turn)
			task.LastTurn = &task.Turns[len(task.Turns)-1]
			task.Status = "in_progress"
			task.UpdatedAt = now
			if err := writeTask(store, task); err != nil {
				return err
			}
			if err := store.AppendEvent(EventInfo{ID: "event-turn-start-" + stableID(task.ID+"-"+turnID), At: now, Kind: "TaskTurnStarted", Message: fmt.Sprintf("%s %s started", task.Title, turnID)}); err != nil {
				return err
			}
			copied := task
			updated = &copied
			return nil
		}
		return fmt.Errorf("task not found: %s", req.TaskID)
	})
	if err != nil {
		return nil, err
	}
	return updated, nil
}

func (p *RealProvider) FinishTaskTurn(req FinishTaskTurnRequest) (*TaskInfo, error) {
	root, err := normalizeWorkspaceRoot(req.Root)
	if err != nil {
		return nil, err
	}
	var updated *TaskInfo
	now := p.now().UTC().Format(time.RFC3339)
	store := NewWorkspaceStore(root)
	err = store.WithLock(func() error {
		tasks, err := loadTasks(root)
		if err != nil {
			return err
		}
		for _, task := range tasks {
			if task.ID != req.TaskID {
				continue
			}
			task = hydrateTask(task)
			index := findTurnIndex(task.Turns, req.TurnID)
			if index < 0 {
				return fmt.Errorf("turn not found: %s", req.TurnID)
			}
			turn := task.Turns[index]
			turn.Status = normalizeTurnStatus(req.Status, req.ExitCode)
			turn.EndedAt = now
			turn.StopReason = normalizeStopReason(firstNonEmpty(req.StopReason, parseStopReason(req.Stdout), parseStopReason(req.Stderr)))
			turn.UsageUSD = req.UsageUSD
			turn.AutoContinue = req.AutoContinue
			stdoutPath, stderrPath, err := writeTurnOutput(root, task.ID, turn.ID, req.Stdout, req.Stderr)
			if err != nil {
				return err
			}
			turn.StdoutPath = stdoutPath
			turn.StderrPath = stderrPath
			task.LastOutput = compactTurnOutput(req.Stdout, req.Stderr)
			if turn.Status == "failed" {
				turn.FailureCategory = classifyFailure(req.Stderr, req.Stdout, req.ExitCode)
				task.FailureCategory = turn.FailureCategory
				task.Status = "failed"
			} else if req.AutoContinue && isContinuableStop(turn.StopReason) {
				task.Status = "in_progress"
			} else {
				task.Status = "waiting"
			}
			task.UsageUSD += req.UsageUSD
			task.Turns[index] = turn
			task.LastTurn = &task.Turns[index]
			task.UpdatedAt = now
			if shouldGenerateOversight(task.Status) {
				generatedAt, err := time.Parse(time.RFC3339, now)
				if err != nil {
					generatedAt = p.now().UTC()
				}
				oversight, err := writeOversightArtifacts(root, task, generatedAt)
				if err != nil {
					return err
				}
				task.Oversight = &oversight
			}
			if err := writeTask(store, task); err != nil {
				return err
			}
			if err := store.AppendEvent(EventInfo{ID: "event-turn-finish-" + stableID(task.ID+"-"+turn.ID+"-"+now), At: now, Kind: "TaskTurnFinished", Message: fmt.Sprintf("%s %s finished: %s", task.Title, turn.ID, task.Status)}); err != nil {
				return err
			}
			copied := task
			updated = &copied
			return nil
		}
		return fmt.Errorf("task not found: %s", req.TaskID)
	})
	if err != nil {
		return nil, err
	}
	return updated, nil
}

func (p *RealProvider) ResumeTaskTurn(req ResumeTaskTurnRequest) (*TaskInfo, error) {
	root, err := normalizeWorkspaceRoot(req.Root)
	if err != nil {
		return nil, err
	}
	feedback := strings.TrimSpace(req.Feedback)
	if feedback == "" {
		return nil, fmt.Errorf("feedback is required to resume a task")
	}
	var updated *TaskInfo
	now := p.now().UTC().Format(time.RFC3339)
	store := NewWorkspaceStore(root)
	err = store.WithLock(func() error {
		tasks, err := loadTasks(root)
		if err != nil {
			return err
		}
		for _, task := range tasks {
			if task.ID != req.TaskID {
				continue
			}
			task = hydrateTask(task)
			if task.Status != "waiting" {
				return fmt.Errorf("task %s must be waiting to resume, got %s", task.ID, task.Status)
			}
			task.FeedbackHistory = append(task.FeedbackHistory, FeedbackRecord{At: now, Message: feedback})
			task.Status = "in_progress"
			task.UpdatedAt = now
			if err := writeTask(store, task); err != nil {
				return err
			}
			if err := store.AppendEvent(EventInfo{ID: "event-turn-resume-" + stableID(task.ID+"-"+now), At: now, Kind: "TaskTurnResumed", Message: fmt.Sprintf("%s resumed with feedback", task.Title)}); err != nil {
				return err
			}
			copied := task
			updated = &copied
			return nil
		}
		return fmt.Errorf("task not found: %s", req.TaskID)
	})
	if err != nil {
		return nil, err
	}
	return updated, nil
}

func (p *RealProvider) RunTaskVerification(ctx context.Context, req RunTaskVerificationRequest) (*TaskInfo, error) {
	root, err := normalizeWorkspaceRoot(req.Root)
	if err != nil {
		return nil, err
	}
	command := strings.TrimSpace(req.Command)
	if command == "" {
		return nil, fmt.Errorf("verification command is required")
	}
	var updated *TaskInfo
	startedAt := p.now().UTC().Format(time.RFC3339)
	store := NewWorkspaceStore(root)
	err = store.WithLock(func() error {
		tasks, err := loadTasks(root)
		if err != nil {
			return err
		}
		for _, task := range tasks {
			if task.ID != req.TaskID {
				continue
			}
			task = hydrateTask(task)
			if task.Status != "waiting" && task.Status != "in_progress" && task.Status != "committing" {
				return fmt.Errorf("task %s cannot run verification from %s", task.ID, task.Status)
			}
			worktree := resolveTaskWorktree(root, task)
			if worktree == "" {
				return fmt.Errorf("task %s cannot run verification without an isolated worktree", task.ID)
			}
			output, exitCode := runVerificationCommand(ctx, worktree, command)
			endedAt := p.now().UTC().Format(time.RFC3339)
			passed, verdictErr := verificationPassed(output, exitCode, req.PassPattern, req.FailPattern)
			if verdictErr != nil {
				return verdictErr
			}
			resultID := fmt.Sprintf("test-%03d", countExistingTests(task)+1)
			outputPath, err := writeTestOutput(root, task.ID, resultID, output)
			if err != nil {
				return err
			}
			result := TestResultInfo{
				ID:          resultID,
				TaskID:      task.ID,
				ProviderID:  fallback(req.ProviderID, "shell"),
				Command:     command,
				Status:      mapTestStatus(passed),
				Passed:      passed,
				OutputPath:  outputPath,
				Output:      compactTurnOutput(output, ""),
				ExitCode:    exitCode,
				PassPattern: strings.TrimSpace(req.PassPattern),
				FailPattern: strings.TrimSpace(req.FailPattern),
				StartedAt:   startedAt,
				EndedAt:     endedAt,
			}
			task.LastTestResult = &result
			task.TestsPassed = passed
			task.UpdatedAt = endedAt
			if !passed {
				task.Status = "waiting"
				task.FailureCategory = "test_failed"
			} else if task.FailureCategory == "test_failed" {
				task.FailureCategory = ""
			}
			if shouldGenerateOversight(task.Status) {
				generatedAt, err := time.Parse(time.RFC3339, endedAt)
				if err != nil {
					generatedAt = p.now().UTC()
				}
				oversight, err := writeOversightArtifacts(root, task, generatedAt)
				if err != nil {
					return err
				}
				task.Oversight = &oversight
			}
			if err := writeTask(store, task); err != nil {
				return err
			}
			if err := store.AppendEvent(EventInfo{ID: "event-test-" + stableID(task.ID+"-"+result.ID+"-"+endedAt), At: endedAt, Kind: "TaskVerificationFinished", Message: fmt.Sprintf("%s verification %s", task.Title, result.Status)}); err != nil {
				return err
			}
			copied := task
			updated = &copied
			return nil
		}
		return fmt.Errorf("task not found: %s", req.TaskID)
	})
	if err != nil {
		return nil, err
	}
	return updated, nil
}

func (p *RealProvider) PrepareTaskCommit(ctx context.Context, req PrepareTaskCommitRequest) (*TaskInfo, error) {
	root, err := normalizeWorkspaceRoot(req.Root)
	if err != nil {
		return nil, err
	}
	var updated *TaskInfo
	store := NewWorkspaceStore(root)
	err = store.WithLock(func() error {
		tasks, err := loadTasks(root)
		if err != nil {
			return err
		}
		for _, task := range tasks {
			if task.ID != req.TaskID {
				continue
			}
			task = hydrateTask(task)
			worktree := resolveTaskWorktree(root, task)
			if worktree == "" {
				return fmt.Errorf("task %s cannot prepare commit without an isolated worktree", task.ID)
			}
			diff, diffStat, err := taskGitDiff(ctx, worktree)
			if err != nil {
				return err
			}
			message := normalizeCommitMessage(req.Message, task)
			task.Commit = &CommitInfo{
				Summary:  firstLine(message),
				Message:  message,
				Diff:     diff,
				DiffStat: diffStat,
				Approved: false,
			}
			task.UpdatedAt = p.now().UTC().Format(time.RFC3339)
			if err := writeTask(store, task); err != nil {
				return err
			}
			if err := store.AppendEvent(EventInfo{ID: "event-commit-preview-" + stableID(task.ID+"-"+task.UpdatedAt), At: task.UpdatedAt, Kind: "TaskCommitPrepared", Message: task.Commit.Summary}); err != nil {
				return err
			}
			copied := task
			updated = &copied
			return nil
		}
		return fmt.Errorf("task not found: %s", req.TaskID)
	})
	if err != nil {
		return nil, err
	}
	return updated, nil
}

func (p *RealProvider) CommitTaskChanges(ctx context.Context, req CommitTaskChangesRequest) (*TaskInfo, error) {
	root, err := normalizeWorkspaceRoot(req.Root)
	if err != nil {
		return nil, err
	}
	if !req.Approved {
		return nil, fmt.Errorf("explicit approval is required before committing task changes")
	}
	var updated *TaskInfo
	store := NewWorkspaceStore(root)
	err = store.WithLock(func() error {
		tasks, err := loadTasks(root)
		if err != nil {
			return err
		}
		for _, task := range tasks {
			if task.ID != req.TaskID {
				continue
			}
			task = hydrateTask(task)
			worktree := resolveTaskWorktree(root, task)
			if worktree == "" {
				return fmt.Errorf("task %s cannot commit without an isolated worktree", task.ID)
			}
			diff, diffStat, err := taskGitDiff(ctx, worktree)
			if err != nil {
				return err
			}
			if strings.TrimSpace(diff) == "" && strings.TrimSpace(diffStat) == "" {
				return fmt.Errorf("task %s has no worktree changes to commit", task.ID)
			}
			message := normalizeCommitMessage(req.Message, task)
			if _, err := gitOutput(ctx, worktree, "add", "-A"); err != nil {
				return err
			}
			if _, err := gitOutput(ctx, worktree, "commit", "-m", message); err != nil {
				return err
			}
			hash, err := gitOutput(ctx, worktree, "rev-parse", "HEAD")
			if err != nil {
				return err
			}
			now := p.now().UTC().Format(time.RFC3339)
			task.Commit = &CommitInfo{
				Hash:        strings.TrimSpace(hash),
				Summary:     firstLine(message),
				Message:     message,
				Diff:        diff,
				DiffStat:    diffStat,
				Approved:    true,
				Committed:   true,
				CommittedAt: now,
			}
			task.UpdatedAt = now
			if err := writeTask(store, task); err != nil {
				return err
			}
			if err := store.AppendEvent(EventInfo{ID: "event-commit-" + stableID(task.ID+"-"+task.Commit.Hash), At: now, Kind: "TaskCommitted", Message: fmt.Sprintf("%s committed %s", task.Title, shortHash(task.Commit.Hash))}); err != nil {
				return err
			}
			copied := task
			updated = &copied
			return nil
		}
		return fmt.Errorf("task not found: %s", req.TaskID)
	})
	if err != nil {
		return nil, err
	}
	return updated, nil
}

func (p *RealProvider) RegenerateOversight(req RegenerateOversightRequest) (*TaskInfo, error) {
	root, err := normalizeWorkspaceRoot(req.Root)
	if err != nil {
		return nil, err
	}
	var updated *TaskInfo
	now := p.now().UTC()
	store := NewWorkspaceStore(root)
	err = store.WithLock(func() error {
		tasks, err := loadTasks(root)
		if err != nil {
			return err
		}
		for _, task := range tasks {
			if task.ID != req.TaskID {
				continue
			}
			task = hydrateTask(task)
			oversight, err := writeOversightArtifacts(root, task, now)
			if err != nil {
				return err
			}
			task.Oversight = &oversight
			task.UpdatedAt = now.Format(time.RFC3339)
			if err := writeTask(store, task); err != nil {
				return err
			}
			if err := store.AppendEvent(EventInfo{ID: "event-oversight-" + stableID(task.ID+"-"+task.Oversight.GeneratedAt), At: task.Oversight.GeneratedAt, Kind: "TaskOversightGenerated", Message: task.Oversight.Summary}); err != nil {
				return err
			}
			copied := task
			updated = &copied
			return nil
		}
		return fmt.Errorf("task not found: %s", req.TaskID)
	})
	if err != nil {
		return nil, err
	}
	return updated, nil
}

func (p *RealProvider) CreateSpec(req CreateSpecRequest) (*SpecNodeInfo, error) {
	root, err := normalizeWorkspaceRoot(req.Root)
	if err != nil {
		return nil, err
	}
	title := strings.TrimSpace(req.Title)
	if title == "" {
		return nil, fmt.Errorf("spec title is required")
	}
	state := normalizeSpecState(fallback(req.State, "drafted"))
	body := strings.TrimSpace(req.Body)
	if body == "" {
		body = "Describe the goal, constraints, acceptance criteria, and dispatch plan."
	}
	rel := specCreatePath(req.ParentPath, title)
	path := filepath.Join(root, rel)
	content := renderSpecMarkdown(title, state, body)
	store := NewWorkspaceStore(root)
	if err := store.WithLock(func() error {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return err
		}
		if err := store.WriteJSON(path+".meta.json", map[string]any{"schema_version": 1, "path": filepath.ToSlash(rel), "title": title, "state": state}); err != nil {
			return err
		}
		tmp := path + ".tmp"
		if err := os.WriteFile(tmp, []byte(content), 0o644); err != nil {
			return err
		}
		if err := os.Rename(tmp, path); err != nil {
			return err
		}
		return store.AppendEvent(EventInfo{ID: "event-spec-" + stableID(rel), At: p.now().UTC().Format(time.RFC3339), Kind: "SpecCreated", Message: title})
	}); err != nil {
		return nil, err
	}
	node := SpecNodeInfo{ID: stableID(rel), Title: title, State: state, Path: filepath.ToSlash(rel), Body: body, UpdatedAt: p.now().UTC().Format(time.RFC3339), Children: []SpecNodeInfo{}}
	return &node, nil
}

func (p *RealProvider) UpdateSpec(req UpdateSpecRequest) (*SpecNodeInfo, error) {
	root, err := normalizeWorkspaceRoot(req.Root)
	if err != nil {
		return nil, err
	}
	rel, path, err := resolveSpecPath(root, req.Path)
	if err != nil {
		return nil, err
	}
	title := strings.TrimSpace(req.Title)
	if title == "" {
		title = markdownTitle(path)
	}
	if title == "" {
		title = strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	}
	state := normalizeSpecState(req.State)
	if state == "" {
		state = inferSpecState(path)
	}
	body := strings.TrimSpace(req.Body)
	if body == "" {
		body = specBody(path)
	}
	now := p.now().UTC().Format(time.RFC3339)
	store := NewWorkspaceStore(root)
	if err := store.WithLock(func() error {
		if err := snapshotSpec(root, rel, p.now().UTC()); err != nil {
			return err
		}
		if err := os.WriteFile(path+".tmp", []byte(renderSpecMarkdown(title, state, body)), 0o644); err != nil {
			return err
		}
		if err := os.Rename(path+".tmp", path); err != nil {
			return err
		}
		return store.AppendEvent(EventInfo{ID: "event-spec-update-" + stableID(rel+"-"+now), At: now, Kind: "SpecUpdated", Message: fmt.Sprintf("%s -> %s", title, state)})
	}); err != nil {
		return nil, err
	}
	return &SpecNodeInfo{ID: stableID(rel), Title: title, State: state, Path: filepath.ToSlash(rel), Body: body, UpdatedAt: now, Children: []SpecNodeInfo{}}, nil
}

func (p *RealProvider) DispatchSpecs(req DispatchSpecsRequest) ([]TaskInfo, error) {
	root, err := normalizeWorkspaceRoot(req.Root)
	if err != nil {
		return nil, err
	}
	specs, err := loadSpecs(root)
	if err != nil {
		return nil, err
	}
	leaves := collectDispatchableSpecs(specs, filepath.ToSlash(strings.TrimSpace(req.Path)))
	created := make([]TaskInfo, 0, len(leaves))
	for _, spec := range leaves {
		prompt := strings.TrimSpace(spec.Body)
		if prompt == "" {
			prompt = "Implement spec " + spec.Path
		}
		deps := []string{}
		if len(created) > 0 {
			deps = []string{created[len(created)-1].ID}
		}
		task, err := p.CreateTask(CreateTaskRequest{Root: root, Title: spec.Title, Prompt: prompt, Flow: "implement", Agent: "auto", Dependencies: deps})
		if err != nil {
			return nil, err
		}
		created = append(created, *task)
	}
	store := NewWorkspaceStore(root)
	_ = store.WithLock(func() error {
		return store.AppendEvent(EventInfo{ID: "event-spec-dispatch-" + stableID(req.Path+p.now().String()), At: p.now().UTC().Format(time.RFC3339), Kind: "SpecsDispatched", Message: fmt.Sprintf("Dispatched %d leaf specs", len(created))})
	})
	return created, nil
}

func (p *RealProvider) UndoPlanningChange(req UndoPlanningChangeRequest) (*SpecNodeInfo, error) {
	root, err := normalizeWorkspaceRoot(req.Root)
	if err != nil {
		return nil, err
	}
	snapshot, err := latestSpecSnapshot(root)
	if err != nil {
		return nil, err
	}
	_, path, err := resolveSpecPath(root, snapshot.Path)
	if err != nil {
		return nil, err
	}
	store := NewWorkspaceStore(root)
	now := p.now().UTC().Format(time.RFC3339)
	if err := store.WithLock(func() error {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(path+".tmp", []byte(snapshot.Body), 0o644); err != nil {
			return err
		}
		if err := os.Rename(path+".tmp", path); err != nil {
			return err
		}
		if err := os.Remove(snapshot.File); err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
		return store.AppendEvent(EventInfo{ID: "event-spec-undo-" + stableID(snapshot.Path+"-"+now), At: now, Kind: "SpecUndo", Message: "Restored " + snapshot.Path})
	}); err != nil {
		return nil, err
	}
	title, state, body := parseSpecMarkdown(path)
	return &SpecNodeInfo{ID: stableID(snapshot.Path), Title: title, State: state, Path: filepath.ToSlash(snapshot.Path), Body: body, UpdatedAt: now, Children: []SpecNodeInfo{}}, nil
}

func (p *RealProvider) UpsertRoutine(req UpsertRoutineRequest) (*RoutineInfo, error) {
	root, err := normalizeWorkspaceRoot(req.Root)
	if err != nil {
		return nil, err
	}
	routines, err := loadRoutines(root)
	if err != nil {
		return nil, err
	}
	id := strings.TrimSpace(req.ID)
	if id == "" {
		id = "routine-" + stableID(req.Name)
	}
	routine := RoutineInfo{
		SchemaVersion: SchemaVersionRoutine,
		ID:            id,
		Name:          fallback(req.Name, id),
		Prompt:        req.Prompt,
		Flow:          fallback(req.Flow, "implement"),
		Schedule:      fallback(req.Schedule, "manual"),
		Enabled:       req.Enabled,
		UpdatedAt:     p.now().UTC().Format(time.RFC3339),
	}
	store := NewWorkspaceStore(root)
	if err := store.WithLock(func() error {
		replaced := false
		for index := range routines {
			if routines[index].ID == id {
				routines[index] = routine
				replaced = true
				break
			}
		}
		if !replaced {
			routines = append(routines, routine)
		}
		if err := store.WriteJSON(store.Path("routines.json"), routines); err != nil {
			return err
		}
		return store.AppendEvent(EventInfo{ID: "event-routine-" + routine.ID, At: routine.UpdatedAt, Kind: "RoutineSaved", Message: routine.Name})
	}); err != nil {
		return nil, err
	}
	return &routine, nil
}

func (p *RealProvider) SaveAutomation(req SaveAutomationRequest) (*AutomationInfo, error) {
	root, err := normalizeWorkspaceRoot(req.Root)
	if err != nil {
		return nil, err
	}
	automation := req.Automation
	automation.SchemaVersion = SchemaVersionAutomation
	store := NewWorkspaceStore(root)
	if err := store.WithLock(func() error {
		if err := store.WriteJSON(store.Path("automation.json"), automation); err != nil {
			return err
		}
		return store.AppendEvent(EventInfo{ID: "event-automation-" + stableID(p.now().String()), At: p.now().UTC().Format(time.RFC3339), Kind: "AutomationSaved", Message: "Automation toggles updated"})
	}); err != nil {
		return nil, err
	}
	return &automation, nil
}

func (p *RealProvider) DetectAgentCLIs(ctx context.Context) ([]ProviderInfo, error) {
	catalog := providerCatalog()
	out := make([]ProviderInfo, 0, len(catalog))
	for _, item := range catalog {
		provider := item
		if provider.Status == "" {
			provider.Status = "not_found"
		}
		path, err := exec.LookPath(provider.Command)
		if err == nil {
			provider.Status = "installed"
			provider.Path = &path
			if version := commandVersion(ctx, provider.Command); version != "" {
				provider.Version = &version
			}
		}
		out = append(out, provider)
	}
	return out, nil
}

func providerCatalog() []ProviderInfo {
	return []ProviderInfo{
		{ID: "claude-code", Name: "Claude Code", Command: "claude", Type: "cli", SetupHint: "Install Claude Code and authenticate it before assigning tasks.", SupportsRun: true},
		{ID: "codex", Name: "Codex", Command: "codex", Type: "cli", SetupHint: "Install Codex and ensure `codex` is on PATH.", SupportsRun: true},
		{ID: "gemini-cli", Name: "Gemini CLI", Command: "gemini", Type: "cli", SetupHint: "Install Gemini CLI and ensure `gemini` is on PATH.", SupportsRun: true},
		{ID: "opencode", Name: "OpenCode", Command: "opencode", Type: "cli", SetupHint: "Install OpenCode and finish setup before assigning it.", SupportsRun: true},
		{ID: "cursor-agent", Name: "Cursor Agent", Command: "cursor-agent", Type: "cli", SetupHint: "Install Cursor Agent and ensure `cursor-agent` is on PATH.", SupportsRun: true},
		{ID: "aider", Name: "Aider", Command: "aider", Type: "cli", SetupHint: "Install Aider and ensure `aider` is on PATH.", SupportsRun: true},
		{ID: "goose", Name: "Goose", Command: "goose", Type: "cli", SetupHint: "Install Goose and ensure `goose` is on PATH.", SupportsRun: true},
		{ID: "shell", Name: "Shell", Command: "sh", Type: "shell", Status: "installed", SetupHint: "System shell used for explicit user-approved commands.", SupportsRun: false},
	}
}

func commandVersion(parent context.Context, command string) string {
	ctx, cancel := context.WithTimeout(parent, 2*time.Second)
	defer cancel()
	output, err := exec.CommandContext(ctx, command, "--version").CombinedOutput()
	if err != nil {
		return ""
	}
	line := strings.TrimSpace(string(output))
	if idx := strings.IndexByte(line, '\n'); idx >= 0 {
		line = line[:idx]
	}
	return line
}

func gitTopLevel(dir string) (string, error) {
	out, err := exec.Command("git", "-C", dir, "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func currentBranch(dir string) string {
	out, err := exec.Command("git", "-C", dir, "branch", "--show-current").Output()
	if err == nil && strings.TrimSpace(string(out)) != "" {
		return strings.TrimSpace(string(out))
	}
	out, err = exec.Command("git", "-C", dir, "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err == nil && strings.TrimSpace(string(out)) != "" {
		return strings.TrimSpace(string(out))
	}
	return "main"
}

func workspaceName(root string) string {
	base := filepath.Base(root)
	if base == "." || base == string(filepath.Separator) || base == "" {
		return "Workspace"
	}
	return base
}

func loadSpecs(root string) ([]SpecNodeInfo, error) {
	specRoot := filepath.Join(root, "specs")
	if stat, err := os.Stat(specRoot); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []SpecNodeInfo{}, nil
		}
		return nil, err
	} else if !stat.IsDir() {
		return []SpecNodeInfo{}, nil
	}

	nodes := make([]SpecNodeInfo, 0)
	err := filepath.WalkDir(specRoot, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if entry.Name() == "node_modules" || strings.HasPrefix(entry.Name(), ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(entry.Name()) != ".md" {
			return nil
		}
		rel, _ := filepath.Rel(root, path)
		title, state, body := parseSpecMarkdown(path)
		if title == "" {
			title = strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
		}
		updatedAt := ""
		if info, err := entry.Info(); err == nil {
			updatedAt = info.ModTime().UTC().Format(time.RFC3339)
		}
		nodes = append(nodes, SpecNodeInfo{
			ID:        stableID(rel),
			Title:     title,
			State:     state,
			Path:      filepath.ToSlash(rel),
			Body:      body,
			UpdatedAt: updatedAt,
			Children:  []SpecNodeInfo{},
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(nodes, func(i, j int) bool { return nodes[i].Path < nodes[j].Path })
	return buildSpecTree(nodes), nil
}

func markdownTitle(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "# "))
		}
	}
	return ""
}

func inferSpecState(path string) string {
	_, state, _ := parseSpecMarkdown(path)
	return state
}

func parseSpecMarkdown(path string) (string, string, string) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", "drafted", ""
	}
	text := string(data)
	title := ""
	bodyLines := make([]string, 0)
	frontmatter := map[string]string{}
	lines := strings.Split(text, "\n")
	index := 0
	if len(lines) > 0 && strings.TrimSpace(lines[0]) == "---" {
		index = 1
		for index < len(lines) && strings.TrimSpace(lines[index]) != "---" {
			if key, value, ok := strings.Cut(lines[index], ":"); ok {
				frontmatter[strings.ToLower(strings.TrimSpace(key))] = strings.Trim(strings.TrimSpace(value), `"'`)
			}
			index++
		}
		if index < len(lines) {
			index++
		}
	}
	for ; index < len(lines); index++ {
		line := lines[index]
		if title == "" && strings.HasPrefix(strings.TrimSpace(line), "# ") {
			title = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "# "))
			continue
		}
		bodyLines = append(bodyLines, line)
	}
	if title == "" {
		title = frontmatter["title"]
	}
	state := normalizeSpecState(firstNonEmpty(frontmatter["state"], frontmatter["status"], inferSpecStateFromText(text)))
	if state == "" {
		state = "drafted"
	}
	return title, state, strings.TrimSpace(strings.Join(bodyLines, "\n"))
}

func inferSpecStateFromText(value string) string {
	text := strings.ToLower(value)
	switch {
	case strings.Contains(text, "state: archived"), strings.Contains(text, "status: archived"):
		return "archived"
	case strings.Contains(text, "state: complete"), strings.Contains(text, "status: complete"):
		return "complete"
	case strings.Contains(text, "state: validated"), strings.Contains(text, "status: validated"):
		return "validated"
	case strings.Contains(text, "state: testing"), strings.Contains(text, "status: testing"):
		return "testing"
	case strings.Contains(text, "state: stale"), strings.Contains(text, "status: stale"):
		return "stale"
	case strings.Contains(text, "todo"), strings.Contains(text, "draft"):
		return "drafted"
	default:
		return "drafted"
	}
}

func specBody(path string) string {
	_, _, body := parseSpecMarkdown(path)
	return body
}

func renderSpecMarkdown(title, state, body string) string {
	return fmt.Sprintf("---\ntitle: %s\nstate: %s\n---\n\n# %s\n\n%s\n", title, normalizeSpecState(state), title, strings.TrimSpace(body))
}

func normalizeSpecState(state string) string {
	switch strings.ToLower(strings.TrimSpace(state)) {
	case "vague", "drafted", "validated", "testing", "complete", "stale", "archived":
		return strings.ToLower(strings.TrimSpace(state))
	case "done", "completed":
		return "complete"
	case "draft", "todo":
		return "drafted"
	default:
		return ""
	}
}

func specCreatePath(parentPath, title string) string {
	slug := stableID(title) + ".md"
	parentPath = filepath.ToSlash(strings.TrimSpace(parentPath))
	if parentPath == "" {
		return filepath.ToSlash(filepath.Join("specs", slug))
	}
	parentPath = strings.TrimPrefix(parentPath, "/")
	parentPath = strings.TrimSuffix(parentPath, ".md")
	if !strings.HasPrefix(parentPath, "specs/") && parentPath != "specs" {
		parentPath = filepath.ToSlash(filepath.Join("specs", parentPath))
	}
	return filepath.ToSlash(filepath.Join(parentPath, slug))
}

func resolveSpecPath(root, rel string) (string, string, error) {
	rel = filepath.ToSlash(strings.TrimSpace(rel))
	rel = strings.TrimPrefix(rel, "/")
	if rel == "" {
		return "", "", fmt.Errorf("spec path is required")
	}
	if !strings.HasPrefix(rel, "specs/") {
		rel = filepath.ToSlash(filepath.Join("specs", rel))
	}
	path := filepath.Join(root, filepath.FromSlash(rel))
	cleanRoot := filepath.Clean(filepath.Join(root, "specs"))
	cleanPath := filepath.Clean(path)
	if cleanPath != cleanRoot && !strings.HasPrefix(cleanPath, cleanRoot+string(filepath.Separator)) {
		return "", "", fmt.Errorf("spec path must stay under specs: %s", rel)
	}
	return filepath.ToSlash(rel), cleanPath, nil
}

func buildSpecTree(nodes []SpecNodeInfo) []SpecNodeInfo {
	byPath := make(map[string]*SpecNodeInfo, len(nodes))
	for index := range nodes {
		nodes[index].Children = []SpecNodeInfo{}
		byPath[nodes[index].Path] = &nodes[index]
	}
	childrenByParent := make(map[string][]string)
	rootPaths := make([]string, 0)
	for _, original := range nodes {
		node := byPath[original.Path]
		parentPath := nearestSpecParent(node.Path, byPath)
		if parentPath == "" {
			rootPaths = append(rootPaths, node.Path)
			continue
		}
		childrenByParent[parentPath] = append(childrenByParent[parentPath], node.Path)
	}
	roots := make([]SpecNodeInfo, 0, len(rootPaths))
	for _, path := range rootPaths {
		roots = append(roots, buildSpecSubtree(path, byPath, childrenByParent))
	}
	sortSpecTree(roots)
	return roots
}

func buildSpecSubtree(path string, nodes map[string]*SpecNodeInfo, childrenByParent map[string][]string) SpecNodeInfo {
	node := *nodes[path]
	node.Children = make([]SpecNodeInfo, 0, len(childrenByParent[path]))
	sort.Strings(childrenByParent[path])
	for _, childPath := range childrenByParent[path] {
		node.Children = append(node.Children, buildSpecSubtree(childPath, nodes, childrenByParent))
	}
	return node
}

func nearestSpecParent(path string, nodes map[string]*SpecNodeInfo) string {
	best := ""
	for candidate := range nodes {
		if candidate == path {
			continue
		}
		prefix := strings.TrimSuffix(candidate, ".md") + "/"
		if strings.HasPrefix(path, prefix) && len(candidate) > len(best) {
			best = candidate
		}
	}
	return best
}

func sortSpecTree(nodes []SpecNodeInfo) {
	sort.Slice(nodes, func(i, j int) bool { return nodes[i].Path < nodes[j].Path })
	for index := range nodes {
		sortSpecTree(nodes[index].Children)
	}
}

func collectDispatchableSpecs(nodes []SpecNodeInfo, selected string) []SpecNodeInfo {
	out := make([]SpecNodeInfo, 0)
	for _, node := range nodes {
		if selected != "" && node.Path != selected && !strings.HasPrefix(node.Path, strings.TrimSuffix(selected, ".md")+"/") {
			continue
		}
		if len(node.Children) > 0 {
			out = append(out, collectDispatchableSpecs(node.Children, "")...)
			continue
		}
		switch node.State {
		case "validated", "testing", "drafted":
			out = append(out, node)
		}
	}
	return out
}

type specSnapshot struct {
	File string `json:"-"`
	Path string `json:"path"`
	Body string `json:"body"`
	At   string `json:"at"`
}

func snapshotSpec(root, rel string, at time.Time) error {
	_, path, err := resolveSpecPath(root, rel)
	if err != nil {
		return err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	snapshot := specSnapshot{Path: rel, Body: string(data), At: at.UTC().Format(time.RFC3339)}
	name := at.UTC().Format("20060102T150405.000000000") + "-" + stableID(rel) + ".json"
	return writeJSON(filepath.Join(root, ".thanos", "spec-history", name), snapshot)
}

func latestSpecSnapshot(root string) (specSnapshot, error) {
	dir := filepath.Join(root, ".thanos", "spec-history")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return specSnapshot{}, err
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() > entries[j].Name() })
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			return specSnapshot{}, err
		}
		var snapshot specSnapshot
		if err := json.Unmarshal(data, &snapshot); err != nil {
			return specSnapshot{}, err
		}
		snapshot.File = path
		return snapshot, nil
	}
	return specSnapshot{}, fmt.Errorf("no planning change snapshot found")
}

func loadTasks(root string) ([]TaskInfo, error) {
	taskDirs := []string{
		filepath.Join(root, ".thanos", "tasks"),
		filepath.Join(root, "data"),
	}
	tasks := make([]TaskInfo, 0)
	for _, dir := range taskDirs {
		if stat, err := os.Stat(dir); err != nil || !stat.IsDir() {
			continue
		}
		err := filepath.WalkDir(dir, func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() {
				if strings.HasPrefix(entry.Name(), ".") {
					return filepath.SkipDir
				}
				return nil
			}
			if entry.Name() != "task.json" && filepath.Ext(entry.Name()) != ".json" {
				return nil
			}
			task, ok := decodeTask(path, root)
			if ok {
				tasks = append(tasks, task)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	tasks = markBlockedTasks(tasks)
	sort.Slice(tasks, func(i, j int) bool { return tasks[i].ID < tasks[j].ID })
	return tasks, nil
}

func decodeTask(path, root string) (TaskInfo, bool) {
	data, err := os.ReadFile(path)
	if err != nil {
		return TaskInfo{}, false
	}
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return TaskInfo{}, false
	}
	rel, _ := filepath.Rel(root, path)
	id := stringFrom(raw, "id", "ID", "uuid", "UUID")
	if id == "" {
		id = stableID(rel)
	}
	title := stringFrom(raw, "title", "Title", "name", "Name")
	prompt := stringFrom(raw, "prompt", "Prompt", "description", "Description")
	if title == "" && prompt != "" {
		title = firstLine(prompt)
	}
	if title == "" {
		title = id
	}
	status := normalizeTaskStatus(stringFrom(raw, "status", "Status"))
	task := TaskInfo{
		SchemaVersion:   intFrom(raw, "schema_version", "schemaVersion"),
		ID:              id,
		Title:           title,
		Prompt:          prompt,
		Status:          status,
		Flow:            fallback(stringFrom(raw, "flow", "Flow"), "implement"),
		Agent:           fallback(stringFrom(raw, "agent", "Agent", "harness", "Harness"), "unassigned"),
		Branch:          stringFrom(raw, "branch", "Branch", "branch_name", "BranchName"),
		Worktree:        stringFrom(raw, "worktree", "Worktree", "worktree_path", "WorktreePath"),
		UpdatedAt:       fallback(stringFrom(raw, "updated_at", "UpdatedAt", "updatedAt"), "unknown"),
		CreatedAt:       stringFrom(raw, "created_at", "CreatedAt", "createdAt"),
		UsageUSD:        floatFrom(raw, "usage_usd", "usageUsd", "cost_usd", "CostUSD"),
		Archived:        boolFrom(raw, "archived", "Archived"),
		Deleted:         boolFrom(raw, "deleted", "Deleted"),
		Tombstone:       boolFrom(raw, "tombstone", "Tombstone"),
		Dependencies:    stringSliceFrom(raw, "dependencies", "Dependencies"),
		PromptHistory:   promptHistoryFrom(raw, prompt),
		FeedbackHistory: feedbackHistoryFrom(raw),
		RetryHistory:    retryHistoryFrom(raw),
		Turns:           taskTurnsFrom(raw),
		LastOutput:      stringFrom(raw, "lastOutput", "last_output"),
		LastTestResult:  testResultFrom(raw),
		FailureCategory: stringFrom(raw, "failure_category", "failureCategory", "FailureCategory"),
		TestsPassed:     boolFrom(raw, "testsPassed", "tests_passed"),
		Commit:          commitInfoFrom(raw),
		Oversight:       oversightInfoFrom(raw),
	}
	return hydrateTask(task), true
}

func loadEvents(root string) ([]EventInfo, error) {
	candidates := []string{
		filepath.Join(root, ".thanos", "events.jsonl"),
		filepath.Join(root, ".thanos", "events.ndjson"),
	}
	for _, path := range candidates {
		data, err := os.ReadFile(path)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return nil, err
		}
		events := make([]EventInfo, 0)
		for index, line := range strings.Split(string(data), "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			var raw map[string]any
			if err := json.Unmarshal([]byte(line), &raw); err != nil {
				continue
			}
			id := stringFrom(raw, "id", "ID")
			if id == "" {
				id = fmt.Sprintf("event-%d", index+1)
			}
			events = append(events, EventInfo{
				ID:      id,
				At:      fallback(stringFrom(raw, "at", "time", "timestamp"), "unknown"),
				Kind:    fallback(stringFrom(raw, "kind", "type", "event"), "Event"),
				Message: fallback(stringFrom(raw, "message", "summary", "text"), line),
			})
		}
		return events, nil
	}
	return []EventInfo{}, nil
}

func loadRoutines(root string) ([]RoutineInfo, error) {
	path := filepath.Join(root, ".thanos", "routines.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []RoutineInfo{}, nil
		}
		return nil, err
	}
	var routines []RoutineInfo
	if err := json.Unmarshal(data, &routines); err != nil {
		return nil, err
	}
	sort.Slice(routines, func(i, j int) bool { return routines[i].ID < routines[j].ID })
	return routines, nil
}

func loadAutomation(root string) (AutomationInfo, error) {
	path := filepath.Join(root, ".thanos", "automation.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return AutomationInfo{}, nil
		}
		return AutomationInfo{}, err
	}
	var automation AutomationInfo
	if err := json.Unmarshal(data, &automation); err != nil {
		return AutomationInfo{}, err
	}
	return automation, nil
}

func loadAgentRoles(root string, providers []ProviderInfo) ([]AgentRoleInfo, error) {
	agents := builtinAgentRoles(providers)
	userAgents, err := readAgentRoleFiles(filepath.Join(root, ".thanos", "agents"), root)
	if err != nil {
		return nil, err
	}
	return mergeAgents(agents, userAgents), nil
}

func readAgentRoleFiles(dir, root string) ([]AgentRoleInfo, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	var out []AgentRoleInfo
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		rel, _ := filepath.Rel(root, path)
		agents, err := decodeAgentRoleFile(data, filepath.ToSlash(rel))
		if err != nil {
			return nil, fmt.Errorf("%s: %w", path, err)
		}
		out = append(out, agents...)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

func decodeAgentRoleFile(data []byte, source string) ([]AgentRoleInfo, error) {
	var raw any
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}
	var records []any
	switch value := raw.(type) {
	case []any:
		records = value
	case map[string]any:
		if nested, ok := value["agents"].([]any); ok {
			records = nested
		} else {
			records = []any{value}
		}
	default:
		return nil, fmt.Errorf("agent file must contain an object, array, or agents array")
	}
	out := make([]AgentRoleInfo, 0, len(records))
	for _, record := range records {
		rawAgent, ok := record.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("agent record must be an object")
		}
		agent := AgentRoleInfo{
			ID:           stringFrom(rawAgent, "id", "ID"),
			Role:         stringFrom(rawAgent, "role", "Role", "name", "Name"),
			Harness:      fallback(stringFrom(rawAgent, "harness", "Harness", "provider", "Provider"), "Codex"),
			Model:        fallback(stringFrom(rawAgent, "model", "Model"), "provider default"),
			Capabilities: stringSliceFrom(rawAgent, "capabilities", "Capabilities"),
			Source:       source,
		}
		if agent.ID == "" {
			agent.ID = stableID(agent.Role)
		}
		if agent.Role == "" {
			agent.Role = agent.ID
		}
		out = append(out, agent)
	}
	return out, nil
}

func mergeAgents(builtins, userAgents []AgentRoleInfo) []AgentRoleInfo {
	out := append([]AgentRoleInfo(nil), builtins...)
	builtinIDs := make(map[string]bool, len(builtins))
	for _, agent := range builtins {
		builtinIDs[agent.ID] = true
	}
	for _, agent := range userAgents {
		if builtinIDs[agent.ID] {
			agent.ID = "user-" + agent.ID
		}
		out = append(out, agent)
	}
	return out
}

func loadFlows(root string) ([]FlowInfo, error) {
	flows := builtinFlows()
	userFlows, err := readFlowFiles(filepath.Join(root, ".thanos", "flows"), root)
	if err != nil {
		return nil, err
	}
	return mergeFlows(flows, userFlows), nil
}

func readFlowFiles(dir, root string) ([]FlowInfo, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	var out []FlowInfo
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		rel, _ := filepath.Rel(root, path)
		flows, err := decodeFlowFile(data, filepath.ToSlash(rel))
		if err != nil {
			return nil, fmt.Errorf("%s: %w", path, err)
		}
		out = append(out, flows...)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

func decodeFlowFile(data []byte, source string) ([]FlowInfo, error) {
	var raw any
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}
	var records []any
	switch value := raw.(type) {
	case []any:
		records = value
	case map[string]any:
		if nested, ok := value["flows"].([]any); ok {
			records = nested
		} else {
			records = []any{value}
		}
	default:
		return nil, fmt.Errorf("flow file must contain an object, array, or flows array")
	}
	out := make([]FlowInfo, 0, len(records))
	for _, record := range records {
		rawFlow, ok := record.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("flow record must be an object")
		}
		flow := FlowInfo{
			ID:             stringFrom(rawFlow, "id", "ID"),
			Name:           stringFrom(rawFlow, "name", "Name", "title", "Title"),
			Steps:          stringSliceFrom(rawFlow, "steps", "Steps"),
			ParallelGroups: stringSlicesFrom(rawFlow, "parallelGroups", "parallel_groups", "parallel", "ParallelGroups"),
			Source:         source,
		}
		if flow.ID == "" {
			flow.ID = stableID(flow.Name)
		}
		if flow.Name == "" {
			flow.Name = flow.ID
		}
		if len(flow.Steps) == 0 {
			return nil, fmt.Errorf("flow %q must include at least one step", flow.ID)
		}
		out = append(out, flow)
	}
	return out, nil
}

func mergeFlows(builtins, userFlows []FlowInfo) []FlowInfo {
	out := append([]FlowInfo(nil), builtins...)
	builtinIDs := make(map[string]bool, len(builtins))
	for _, flow := range builtins {
		builtinIDs[flow.ID] = true
	}
	for _, flow := range userFlows {
		if builtinIDs[flow.ID] {
			flow.ID = "user-" + flow.ID
		}
		out = append(out, flow)
	}
	return out
}

func normalizeTaskFlows(tasks []TaskInfo, flows []FlowInfo) []TaskInfo {
	for index := range tasks {
		tasks[index].Flow = normalizeFlowID(tasks[index].Flow, flows)
	}
	return tasks
}

func normalizeFlowID(id string, flows []FlowInfo) string {
	id = strings.TrimSpace(id)
	if id == "" {
		return "implement"
	}
	for _, flow := range flows {
		if flow.ID == id {
			return id
		}
	}
	return "implement"
}

func normalizeWorkspaceRoot(root string) (string, error) {
	if strings.TrimSpace(root) == "" {
		return "", fmt.Errorf("workspace root is required")
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	if stat, err := os.Stat(abs); err != nil {
		return "", err
	} else if !stat.IsDir() {
		return "", fmt.Errorf("workspace path is not a directory: %s", abs)
	}
	if gitRoot, err := gitTopLevel(abs); err == nil && gitRoot != "" {
		return gitRoot, nil
	}
	return abs, nil
}

func writeTask(store *WorkspaceStore, task TaskInfo) error {
	task.SchemaVersion = SchemaVersionTask
	return store.WriteJSON(store.Path("tasks", task.ID+".json"), task)
}

func writeJSON(path string, value any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func writeTurnOutput(root, taskID, turnID, stdout, stderr string) (string, string, error) {
	dir := filepath.Join(root, ".thanos", "logs", taskID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", "", err
	}
	stdoutRel := filepath.ToSlash(filepath.Join(".thanos", "logs", taskID, turnID+".stdout.log"))
	stderrRel := filepath.ToSlash(filepath.Join(".thanos", "logs", taskID, turnID+".stderr.log"))
	if err := os.WriteFile(filepath.Join(root, filepath.FromSlash(stdoutRel)), []byte(stdout), 0o644); err != nil {
		return "", "", err
	}
	if err := os.WriteFile(filepath.Join(root, filepath.FromSlash(stderrRel)), []byte(stderr), 0o644); err != nil {
		return "", "", err
	}
	return stdoutRel, stderrRel, nil
}

func findTurnIndex(turns []TaskTurnInfo, id string) int {
	id = strings.TrimSpace(id)
	for index, turn := range turns {
		if turn.ID == id {
			return index
		}
	}
	return -1
}

func normalizeTurnStatus(status string, exitCode int) string {
	status = strings.ToLower(strings.TrimSpace(status))
	switch status {
	case "completed", "complete", "done", "success":
		return "completed"
	case "stopped", "cancelled", "canceled":
		return "stopped"
	case "failed", "error":
		return "failed"
	}
	if exitCode != 0 {
		return "failed"
	}
	return "completed"
}

func parseStopReason(output string) string {
	text := strings.ToLower(output)
	switch {
	case strings.Contains(text, "max token") || strings.Contains(text, "context length") || strings.Contains(text, "token limit"):
		return "max_tokens"
	case strings.Contains(text, "waiting for user") || strings.Contains(text, "needs input") || strings.Contains(text, "user input"):
		return "waiting_user"
	case strings.Contains(text, "paused") || strings.Contains(text, "pause"):
		return "paused"
	case strings.Contains(text, "completed") || strings.Contains(text, "done"):
		return "completed"
	default:
		return ""
	}
}

func normalizeStopReason(reason string) string {
	reason = strings.ToLower(strings.TrimSpace(reason))
	reason = strings.ReplaceAll(reason, " ", "_")
	if reason == "" {
		return "completed"
	}
	return reason
}

func isContinuableStop(reason string) bool {
	switch normalizeStopReason(reason) {
	case "max_tokens", "paused":
		return true
	default:
		return false
	}
}

func classifyFailure(stderr, stdout string, exitCode int) string {
	text := strings.ToLower(stderr + "\n" + stdout)
	switch {
	case strings.Contains(text, "permission denied") || strings.Contains(text, "unauthorized") || strings.Contains(text, "forbidden"):
		return "permissions"
	case strings.Contains(text, "timed out") || strings.Contains(text, "timeout") || strings.Contains(text, "deadline exceeded"):
		return "timeout"
	case strings.Contains(text, "not found") || strings.Contains(text, "no such file") || strings.Contains(text, "executable file not found"):
		return "environment"
	case strings.Contains(text, "rate limit") || strings.Contains(text, "quota"):
		return "provider_limit"
	case exitCode != 0:
		return "process_exit"
	default:
		return "unknown"
	}
}

func compactTurnOutput(stdout, stderr string) string {
	output := strings.TrimSpace(strings.TrimSpace(stdout) + "\n" + strings.TrimSpace(stderr))
	if len(output) <= 4000 {
		return output
	}
	return output[len(output)-4000:]
}

func resolveTaskWorktree(root string, task TaskInfo) string {
	worktree := strings.TrimSpace(task.Worktree)
	if worktree == "" {
		return ""
	}
	if filepath.IsAbs(worktree) {
		if stat, err := os.Stat(worktree); err == nil && stat.IsDir() {
			return worktree
		}
		return ""
	}
	path := filepath.Join(root, filepath.FromSlash(worktree))
	if stat, err := os.Stat(path); err == nil && stat.IsDir() {
		return path
	}
	return ""
}

func runVerificationCommand(ctx context.Context, cwd, command string) (string, int) {
	if ctx == nil {
		ctx = context.Background()
	}
	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	cmd.Dir = cwd
	cmd.Env = scrubAuthEnv(os.Environ())
	output, err := cmd.CombinedOutput()
	if err == nil {
		return string(output), 0
	}
	if cmd.ProcessState != nil {
		return string(output), cmd.ProcessState.ExitCode()
	}
	return string(output) + "\n" + err.Error(), 1
}

func verificationPassed(output string, exitCode int, passPattern, failPattern string) (bool, error) {
	if strings.TrimSpace(failPattern) != "" {
		matched, err := regexp.MatchString(failPattern, output)
		if err != nil {
			return false, fmt.Errorf("invalid fail pattern: %w", err)
		}
		if matched {
			return false, nil
		}
	}
	if strings.TrimSpace(passPattern) != "" {
		matched, err := regexp.MatchString(passPattern, output)
		if err != nil {
			return false, fmt.Errorf("invalid pass pattern: %w", err)
		}
		return matched, nil
	}
	upper := strings.ToUpper(output)
	if strings.Contains(upper, "FAIL") || strings.Contains(upper, "FAILED") {
		return false, nil
	}
	if strings.Contains(upper, "PASS") || strings.Contains(upper, "PASSED") || strings.Contains(upper, "OK") {
		return exitCode == 0, nil
	}
	return exitCode == 0, nil
}

func mapTestStatus(passed bool) string {
	if passed {
		return "passed"
	}
	return "failed"
}

func writeTestOutput(root, taskID, resultID, output string) (string, error) {
	rel := filepath.ToSlash(filepath.Join(".thanos", "tests", taskID, resultID+".log"))
	path := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", err
	}
	if err := os.WriteFile(path, []byte(output), 0o644); err != nil {
		return "", err
	}
	return rel, nil
}

func shouldGenerateOversight(status string) bool {
	switch normalizeTaskStatus(status) {
	case "waiting", "done", "failed":
		return true
	default:
		return false
	}
}

func writeOversightArtifacts(root string, task TaskInfo, generatedAt time.Time) (OversightInfo, error) {
	task = hydrateTask(task)
	oversight := buildOversight(task, generatedAt)
	dir := filepath.Join(root, ".thanos", "oversight", task.ID)
	oversight.Path = filepath.ToSlash(filepath.Join(".thanos", "oversight", task.ID, "oversight.json"))
	if task.LastTestResult != nil {
		oversight.TestPath = filepath.ToSlash(filepath.Join(".thanos", "oversight", task.ID, "oversight-test.json"))
	}
	if err := writeJSON(filepath.Join(dir, "oversight.json"), oversight); err != nil {
		return OversightInfo{}, err
	}
	if task.LastTestResult != nil {
		testArtifact := map[string]any{
			"schema_version": 1,
			"id":             oversight.ID + "-test",
			"taskId":         task.ID,
			"status":         task.LastTestResult.Status,
			"passed":         task.LastTestResult.Passed,
			"command":        task.LastTestResult.Command,
			"exitCode":       task.LastTestResult.ExitCode,
			"outputPath":     task.LastTestResult.OutputPath,
			"output":         task.LastTestResult.Output,
			"generatedAt":    oversight.GeneratedAt,
		}
		if err := writeJSON(filepath.Join(dir, "oversight-test.json"), testArtifact); err != nil {
			return OversightInfo{}, err
		}
	}
	return oversight, nil
}

func buildOversight(task TaskInfo, generatedAt time.Time) OversightInfo {
	testResult := "not_run"
	if task.LastTestResult != nil {
		testResult = task.LastTestResult.Status
	}
	risks := oversightRisks(task)
	return OversightInfo{
		SchemaVersion: 1,
		ID:            "oversight-" + task.ID,
		TaskID:        task.ID,
		Status:        task.Status,
		Summary:       oversightSummary(task, testResult, risks),
		Phases:        oversightPhases(task),
		Risks:         risks,
		ChangedFiles:  oversightChangedFiles(task),
		Commands:      oversightCommands(task),
		TestResult:    testResult,
		UsageUSD:      task.UsageUSD,
		GeneratedAt:   generatedAt.UTC().Format(time.RFC3339),
	}
}

func oversightSummary(task TaskInfo, testResult string, risks []string) string {
	title := strings.TrimSpace(task.Title)
	if title == "" {
		title = task.ID
	}
	if len(risks) == 0 {
		return fmt.Sprintf("%s is %s with %s verification and $%.2f usage.", title, task.Status, testResult, task.UsageUSD)
	}
	return fmt.Sprintf("%s is %s with %s verification, $%.2f usage, and %d review risk(s).", title, task.Status, testResult, task.UsageUSD, len(risks))
}

func oversightPhases(task TaskInfo) []string {
	phases := make([]string, 0, len(task.Turns)+2)
	for _, turn := range task.Turns {
		phase := strings.TrimSpace(turn.Step)
		if phase == "" {
			phase = "Implementation"
		}
		parts := []string{phase}
		if turn.ProviderID != "" {
			parts = append(parts, turn.ProviderID)
		}
		if turn.Status != "" {
			parts = append(parts, turn.Status)
		}
		if turn.StopReason != "" {
			parts = append(parts, turn.StopReason)
		}
		phases = append(phases, strings.Join(parts, " / "))
	}
	if task.LastTestResult != nil {
		phases = append(phases, "Testing / "+task.LastTestResult.Status)
	}
	if task.Commit != nil {
		if task.Commit.Committed {
			phases = append(phases, "Commit / committed")
		} else {
			phases = append(phases, "Commit / prepared")
		}
	}
	return uniqueStrings(phases)
}

func oversightCommands(task TaskInfo) []string {
	commands := make([]string, 0, len(task.Turns)+1)
	for _, turn := range task.Turns {
		command := strings.TrimSpace(turn.ProviderID)
		if command == "" {
			command = strings.TrimSpace(task.Agent)
		}
		if command != "" {
			commands = append(commands, command+" "+strings.TrimSpace(turn.Step))
		}
	}
	if task.LastTestResult != nil && strings.TrimSpace(task.LastTestResult.Command) != "" {
		commands = append(commands, task.LastTestResult.Command)
	}
	return uniqueStrings(commands)
}

func oversightChangedFiles(task TaskInfo) []string {
	values := make([]string, 0)
	if task.Commit != nil {
		values = append(values, parseChangedFiles(task.Commit.DiffStat)...)
		values = append(values, parseDiffFiles(task.Commit.Diff)...)
	}
	return uniqueStrings(values)
}

func parseChangedFiles(diffStat string) []string {
	out := make([]string, 0)
	for _, line := range strings.Split(diffStat, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.Contains(line, " file changed") || strings.Contains(line, " files changed") {
			continue
		}
		if strings.HasPrefix(line, "?? ") || strings.HasPrefix(line, "M ") || strings.HasPrefix(line, "A ") || strings.HasPrefix(line, "D ") {
			out = append(out, strings.TrimSpace(line[2:]))
			continue
		}
		if before, _, ok := strings.Cut(line, "|"); ok {
			out = append(out, strings.TrimSpace(before))
		}
	}
	return out
}

func parseDiffFiles(diff string) []string {
	out := make([]string, 0)
	for _, line := range strings.Split(diff, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "diff --git ") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) >= 4 {
			out = append(out, strings.TrimPrefix(fields[3], "b/"))
		}
	}
	return out
}

func oversightRisks(task TaskInfo) []string {
	risks := make([]string, 0)
	if task.Status == "failed" {
		risks = append(risks, "Task is failed.")
	}
	if task.Blocked {
		risks = append(risks, "Task has unfinished dependencies.")
	}
	if strings.TrimSpace(task.FailureCategory) != "" {
		risks = append(risks, "Failure category: "+task.FailureCategory+".")
	}
	if task.LastTurn != nil && task.LastTurn.Status == "failed" {
		risks = append(risks, "Last provider turn failed.")
	}
	if task.LastTestResult == nil {
		risks = append(risks, "No verification result is recorded.")
	} else if !task.LastTestResult.Passed {
		risks = append(risks, "Last verification failed.")
	}
	if task.Status == "done" && (task.Commit == nil || !task.Commit.Committed) {
		risks = append(risks, "Task is done without a committed change artifact.")
	}
	return uniqueStrings(risks)
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]bool, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	return out
}

func countExistingTests(task TaskInfo) int {
	if task.LastTestResult == nil {
		return 0
	}
	return 1
}

func taskGitDiff(ctx context.Context, worktree string) (string, string, error) {
	diff, err := gitOutput(ctx, worktree, "diff", "--")
	if err != nil {
		return "", "", err
	}
	staged, err := gitOutput(ctx, worktree, "diff", "--cached", "--")
	if err != nil {
		return "", "", err
	}
	stat, err := gitOutput(ctx, worktree, "diff", "--stat", "HEAD", "--")
	if err != nil {
		return "", "", err
	}
	status, err := gitOutput(ctx, worktree, "status", "--short")
	if err != nil {
		return "", "", err
	}
	diff = strings.TrimSpace(strings.TrimSpace(diff) + "\n" + strings.TrimSpace(staged))
	diffStat := strings.TrimSpace(stat)
	if strings.TrimSpace(status) != "" {
		diffStat = strings.TrimSpace(strings.TrimSpace(diffStat) + "\n" + strings.TrimSpace(status))
	}
	return diff, diffStat, nil
}

func gitOutput(ctx context.Context, dir string, args ...string) (string, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(string(output)))
	}
	return string(output), nil
}

func normalizeCommitMessage(message string, task TaskInfo) string {
	message = strings.TrimSpace(message)
	if message != "" {
		return message
	}
	title := strings.TrimSpace(task.Title)
	if title == "" {
		title = task.ID
	}
	return "task: " + title
}

func shortHash(hash string) string {
	hash = strings.TrimSpace(hash)
	if len(hash) <= 12 {
		return hash
	}
	return hash[:12]
}

func appendEvent(root string, event EventInfo) error {
	path := filepath.Join(root, ".thanos", "events.jsonl")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()
	if _, err := file.Write(append(data, '\n')); err != nil {
		return err
	}
	return nil
}

func builtinAgentRoles(providers []ProviderInfo) []AgentRoleInfo {
	defaultHarness := firstInstalled(providers, "codex", "claude-code", "claude", "opencode", "cursor-agent", "cursor")
	if defaultHarness == "" {
		defaultHarness = "codex"
	}
	return []AgentRoleInfo{
		{ID: "impl", Role: "Implementation", Harness: displayHarness(defaultHarness), Model: "provider default", Capabilities: []string{"workspace.read", "workspace.write", "board.context"}, ReadOnly: true, Source: "builtin"},
		{ID: "test", Role: "Testing", Harness: displayHarness(firstNonEmpty(firstInstalled(providers, "claude-code", "claude", "codex"), defaultHarness)), Model: "provider default", Capabilities: []string{"workspace.read", "commands.run"}, ReadOnly: true, Source: "builtin"},
		{ID: "oversight", Role: "Oversight", Harness: displayHarness(firstNonEmpty(firstInstalled(providers, "claude-code", "claude", "codex"), defaultHarness)), Model: "provider default", Capabilities: []string{"diff.read", "timeline.read", "risk.review"}, ReadOnly: true, Source: "builtin"},
		{ID: "title", Role: "Title", Harness: "Shell", Model: "none", Capabilities: []string{"metadata.write"}, ReadOnly: true, Source: "builtin"},
		{ID: "commit-msg", Role: "Commit Message", Harness: displayHarness(defaultHarness), Model: "provider default", Capabilities: []string{"diff.read", "metadata.write"}, ReadOnly: true, Source: "builtin"},
	}
}

func builtinFlows() []FlowInfo {
	return []FlowInfo{
		{ID: "implement", Name: "Implement", Steps: []string{"Implementation", "Testing", "Commit Message", "Title", "Oversight"}, ReadOnly: true, Source: "builtin"},
		{ID: "plan-first", Name: "Plan First", Steps: []string{"Planning", "Implementation", "Testing", "Oversight"}, ParallelGroups: [][]string{{"Testing", "Oversight"}}, ReadOnly: true, Source: "builtin"},
		{ID: "oversight-only", Name: "Oversight Only", Steps: []string{"Oversight"}, ReadOnly: true, Source: "builtin"},
	}
}

func firstInstalled(providers []ProviderInfo, ids ...string) string {
	for _, id := range ids {
		for _, provider := range providers {
			if provider.ID == id && provider.Status == "installed" {
				return provider.ID
			}
		}
	}
	return ""
}

func displayHarness(id string) string {
	switch id {
	case "claude", "claude-code":
		return "Claude"
	case "codex":
		return "Codex"
	case "cursor", "cursor-agent":
		return "Cursor"
	case "opencode":
		return "OpenCode"
	case "gemini", "gemini-cli":
		return "Gemini"
	default:
		return "Shell"
	}
}

func stableID(value string) string {
	id := strings.ToLower(value)
	id = strings.ReplaceAll(id, string(filepath.Separator), "-")
	id = strings.ReplaceAll(id, "/", "-")
	id = strings.TrimSuffix(id, filepath.Ext(id))
	id = strings.Map(func(r rune) rune {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' {
			return r
		}
		return '-'
	}, id)
	id = strings.Trim(id, "-")
	if id == "" {
		return "item"
	}
	return id
}

func stringFrom(raw map[string]any, keys ...string) string {
	for _, key := range keys {
		if value, ok := raw[key]; ok {
			switch typed := value.(type) {
			case string:
				return strings.TrimSpace(typed)
			case fmt.Stringer:
				return strings.TrimSpace(typed.String())
			}
		}
	}
	return ""
}

func floatFrom(raw map[string]any, keys ...string) float64 {
	for _, key := range keys {
		if value, ok := raw[key]; ok {
			switch typed := value.(type) {
			case float64:
				return typed
			case int:
				return float64(typed)
			case json.Number:
				out, _ := typed.Float64()
				return out
			}
		}
	}
	return 0
}

func intFrom(raw map[string]any, keys ...string) int {
	for _, key := range keys {
		if value, ok := raw[key]; ok {
			switch typed := value.(type) {
			case float64:
				return int(typed)
			case int:
				return typed
			case json.Number:
				out, _ := typed.Int64()
				return int(out)
			}
		}
	}
	return 0
}

func boolFrom(raw map[string]any, keys ...string) bool {
	for _, key := range keys {
		if value, ok := raw[key]; ok {
			switch typed := value.(type) {
			case bool:
				return typed
			case string:
				return typed == "true" || typed == "1"
			}
		}
	}
	return false
}

func stringSliceFrom(raw map[string]any, keys ...string) []string {
	for _, key := range keys {
		value, ok := raw[key]
		if !ok {
			continue
		}
		switch typed := value.(type) {
		case []string:
			return normalizeTaskDependencies(typed)
		case []any:
			out := make([]string, 0, len(typed))
			for _, item := range typed {
				if text, ok := item.(string); ok {
					out = append(out, text)
				}
			}
			return normalizeTaskDependencies(out)
		}
	}
	return []string{}
}

func stringSlicesFrom(raw map[string]any, keys ...string) [][]string {
	for _, key := range keys {
		value, ok := raw[key]
		if !ok {
			continue
		}
		groups, ok := value.([]any)
		if !ok {
			continue
		}
		out := make([][]string, 0, len(groups))
		for _, group := range groups {
			switch typed := group.(type) {
			case []any:
				items := make([]string, 0, len(typed))
				for _, item := range typed {
					if text, ok := item.(string); ok && strings.TrimSpace(text) != "" {
						items = append(items, strings.TrimSpace(text))
					}
				}
				if len(items) > 0 {
					out = append(out, items)
				}
			case string:
				items := stringSliceFrom(map[string]any{"items": strings.Split(typed, "+")}, "items")
				if len(items) > 0 {
					out = append(out, items)
				}
			}
		}
		return out
	}
	return nil
}

func promptHistoryFrom(raw map[string]any, prompt string) []PromptRecord {
	out := make([]PromptRecord, 0)
	if records, ok := raw["promptHistory"].([]any); ok {
		for _, item := range records {
			if record, ok := item.(map[string]any); ok {
				out = append(out, PromptRecord{At: stringFrom(record, "at"), Prompt: stringFrom(record, "prompt")})
			}
		}
	}
	if len(out) == 0 && strings.TrimSpace(prompt) != "" {
		out = append(out, PromptRecord{At: stringFrom(raw, "created_at", "createdAt", "updated_at", "updatedAt"), Prompt: prompt})
	}
	return out
}

func feedbackHistoryFrom(raw map[string]any) []FeedbackRecord {
	out := make([]FeedbackRecord, 0)
	if records, ok := raw["feedbackHistory"].([]any); ok {
		for _, item := range records {
			if record, ok := item.(map[string]any); ok {
				out = append(out, FeedbackRecord{At: stringFrom(record, "at"), Message: stringFrom(record, "message")})
			}
		}
	}
	return out
}

func retryHistoryFrom(raw map[string]any) []RetryRecord {
	out := make([]RetryRecord, 0)
	if records, ok := raw["retryHistory"].([]any); ok {
		for _, item := range records {
			if record, ok := item.(map[string]any); ok {
				out = append(out, RetryRecord{At: stringFrom(record, "at"), Reason: stringFrom(record, "reason")})
			}
		}
	}
	return out
}

func taskTurnsFrom(raw map[string]any) []TaskTurnInfo {
	records, ok := raw["turns"].([]any)
	if !ok {
		return []TaskTurnInfo{}
	}
	out := make([]TaskTurnInfo, 0, len(records))
	for _, item := range records {
		record, ok := item.(map[string]any)
		if !ok {
			continue
		}
		out = append(out, TaskTurnInfo{
			ID:              stringFrom(record, "id"),
			TaskID:          stringFrom(record, "taskId", "task_id"),
			Step:            stringFrom(record, "step"),
			ProviderID:      stringFrom(record, "providerId", "provider_id"),
			SessionID:       stringFrom(record, "sessionId", "session_id"),
			Status:          stringFrom(record, "status"),
			Worktree:        stringFrom(record, "worktree"),
			StartedAt:       stringFrom(record, "startedAt", "started_at"),
			EndedAt:         stringFrom(record, "endedAt", "ended_at"),
			StdoutPath:      stringFrom(record, "stdoutPath", "stdout_path"),
			StderrPath:      stringFrom(record, "stderrPath", "stderr_path"),
			TranscriptPath:  stringFrom(record, "transcriptPath", "transcript_path"),
			StopReason:      stringFrom(record, "stopReason", "stop_reason"),
			UsageUSD:        floatFrom(record, "usageUsd", "usage_usd"),
			FailureCategory: stringFrom(record, "failureCategory", "failure_category"),
			AutoContinue:    boolFrom(record, "autoContinue", "auto_continue"),
		})
	}
	return out
}

func testResultFrom(raw map[string]any) *TestResultInfo {
	record, ok := raw["lastTestResult"].(map[string]any)
	if !ok {
		record, ok = raw["last_test_result"].(map[string]any)
	}
	if !ok {
		return nil
	}
	return &TestResultInfo{
		ID:          stringFrom(record, "id"),
		TaskID:      stringFrom(record, "taskId", "task_id"),
		ProviderID:  stringFrom(record, "providerId", "provider_id"),
		Command:     stringFrom(record, "command"),
		Status:      stringFrom(record, "status"),
		Passed:      boolFrom(record, "passed"),
		OutputPath:  stringFrom(record, "outputPath", "output_path"),
		Output:      stringFrom(record, "output"),
		ExitCode:    intFrom(record, "exitCode", "exit_code"),
		PassPattern: stringFrom(record, "passPattern", "pass_pattern"),
		FailPattern: stringFrom(record, "failPattern", "fail_pattern"),
		StartedAt:   stringFrom(record, "startedAt", "started_at"),
		EndedAt:     stringFrom(record, "endedAt", "ended_at"),
	}
}

func commitInfoFrom(raw map[string]any) *CommitInfo {
	record, ok := raw["commit"].(map[string]any)
	if !ok {
		return nil
	}
	return &CommitInfo{
		Hash:        stringFrom(record, "hash"),
		Summary:     stringFrom(record, "summary"),
		Message:     stringFrom(record, "message"),
		Diff:        stringFrom(record, "diff"),
		DiffStat:    stringFrom(record, "diffStat", "diff_stat"),
		Approved:    boolFrom(record, "approved"),
		Committed:   boolFrom(record, "committed"),
		CommittedAt: stringFrom(record, "committedAt", "committed_at"),
	}
}

func oversightInfoFrom(raw map[string]any) *OversightInfo {
	record, ok := raw["oversight"].(map[string]any)
	if !ok {
		return nil
	}
	return &OversightInfo{
		SchemaVersion: intFrom(record, "schema_version", "schemaVersion"),
		ID:            stringFrom(record, "id"),
		TaskID:        stringFrom(record, "taskId", "task_id"),
		Status:        stringFrom(record, "status"),
		Summary:       stringFrom(record, "summary"),
		Phases:        stringSliceFrom(record, "phases"),
		Risks:         stringSliceFrom(record, "risks"),
		ChangedFiles:  stringSliceFrom(record, "changedFiles", "changed_files"),
		Commands:      stringSliceFrom(record, "commands"),
		TestResult:    stringFrom(record, "testResult", "test_result"),
		UsageUSD:      floatFrom(record, "usageUsd", "usage_usd"),
		GeneratedAt:   stringFrom(record, "generatedAt", "generated_at"),
		Path:          stringFrom(record, "path"),
		TestPath:      stringFrom(record, "testPath", "test_path"),
	}
}

func normalizeTaskStatus(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "in_progress", "running":
		return "in_progress"
	case "waiting", "waiting_user", "waiting_approval":
		return "waiting"
	case "committing", "commit":
		return "committing"
	case "done", "complete", "completed":
		return "done"
	case "failed":
		return "failed"
	case "cancelled", "canceled":
		return "cancelled"
	default:
		return "backlog"
	}
}

func normalizeTaskDependencies(values []string) []string {
	seen := make(map[string]bool)
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	return out
}

func hydrateTask(task TaskInfo) TaskInfo {
	task.SchemaVersion = SchemaVersionTask
	task.Status = normalizeTaskStatus(task.Status)
	task.Dependencies = normalizeTaskDependencies(task.Dependencies)
	if task.CreatedAt == "" {
		task.CreatedAt = task.UpdatedAt
	}
	if task.UpdatedAt == "" {
		task.UpdatedAt = task.CreatedAt
	}
	if task.PromptHistory == nil {
		task.PromptHistory = []PromptRecord{}
	}
	if len(task.PromptHistory) == 0 && strings.TrimSpace(task.Prompt) != "" {
		task.PromptHistory = append(task.PromptHistory, PromptRecord{At: task.CreatedAt, Prompt: task.Prompt})
	}
	if task.FeedbackHistory == nil {
		task.FeedbackHistory = []FeedbackRecord{}
	}
	if task.RetryHistory == nil {
		task.RetryHistory = []RetryRecord{}
	}
	if task.Turns == nil {
		task.Turns = []TaskTurnInfo{}
	}
	for index := range task.Turns {
		if task.Turns[index].TaskID == "" {
			task.Turns[index].TaskID = task.ID
		}
		if task.Turns[index].Worktree == "" {
			task.Turns[index].Worktree = task.Worktree
		}
	}
	if len(task.Turns) > 0 {
		task.LastTurn = &task.Turns[len(task.Turns)-1]
	} else {
		task.LastTurn = nil
	}
	if task.LastTestResult != nil && task.LastTestResult.TaskID == "" {
		task.LastTestResult.TaskID = task.ID
	}
	if task.Oversight != nil {
		if task.Oversight.SchemaVersion == 0 {
			task.Oversight.SchemaVersion = 1
		}
		if task.Oversight.TaskID == "" {
			task.Oversight.TaskID = task.ID
		}
	}
	return task
}

func markBlockedTasks(tasks []TaskInfo) []TaskInfo {
	done := make(map[string]bool)
	for _, task := range tasks {
		if normalizeTaskStatus(task.Status) == "done" {
			done[task.ID] = true
		}
	}
	out := make([]TaskInfo, 0, len(tasks))
	for _, task := range tasks {
		task = hydrateTask(task)
		task.Blocked = false
		for _, dependency := range task.Dependencies {
			if !done[dependency] {
				task.Blocked = true
				break
			}
		}
		out = append(out, task)
	}
	return out
}

func validateTaskTransition(from, to string) error {
	from = normalizeTaskStatus(from)
	to = normalizeTaskStatus(to)
	if from == to {
		return nil
	}
	allowed := map[string]map[string]bool{
		"backlog":     {"in_progress": true, "cancelled": true},
		"in_progress": {"waiting": true, "committing": true, "failed": true, "cancelled": true},
		"waiting":     {"in_progress": true, "committing": true, "failed": true, "cancelled": true},
		"committing":  {"done": true, "failed": true, "cancelled": true},
		"failed":      {"backlog": true, "in_progress": true, "cancelled": true},
		"cancelled":   {"backlog": true},
		"done":        {},
	}
	if allowed[from][to] {
		return nil
	}
	return fmt.Errorf("invalid task transition: %s -> %s", from, to)
}

func firstLine(value string) string {
	line := strings.TrimSpace(value)
	if idx := strings.IndexByte(line, '\n'); idx >= 0 {
		line = line[:idx]
	}
	if len(line) > 80 {
		return strings.TrimSpace(line[:80])
	}
	return line
}

func fallback(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
