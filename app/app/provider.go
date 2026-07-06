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
	SchemaVersion int     `json:"schema_version"`
	ID            string  `json:"id"`
	Title         string  `json:"title"`
	Prompt        string  `json:"prompt"`
	Status        string  `json:"status"`
	Flow          string  `json:"flow"`
	Agent         string  `json:"agent"`
	Branch        string  `json:"branch"`
	Worktree      string  `json:"worktree"`
	UpdatedAt     string  `json:"updatedAt"`
	UsageUSD      float64 `json:"usageUsd"`
}

type CreateTaskRequest struct {
	Root   string `json:"root"`
	Title  string `json:"title"`
	Prompt string `json:"prompt"`
	Flow   string `json:"flow"`
	Agent  string `json:"agent"`
}

type UpdateTaskStatusRequest struct {
	Root   string `json:"root"`
	TaskID string `json:"taskId"`
	Status string `json:"status"`
}

type SpecNodeInfo struct {
	ID       string         `json:"id"`
	Title    string         `json:"title"`
	State    string         `json:"state"`
	Path     string         `json:"path"`
	Children []SpecNodeInfo `json:"children"`
}

type CreateSpecRequest struct {
	Root  string `json:"root"`
	Title string `json:"title"`
	Body  string `json:"body"`
	State string `json:"state"`
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
}

type FlowInfo struct {
	ID    string   `json:"id"`
	Name  string   `json:"name"`
	Steps []string `json:"steps"`
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
		Agents:        builtinAgentRoles(providers),
		Flows:         builtinFlows(),
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
	task := TaskInfo{
		SchemaVersion: SchemaVersionTask,
		ID:            id,
		Title:         title,
		Prompt:        prompt,
		Status:        "backlog",
		Flow:          fallback(req.Flow, "implement"),
		Agent:         fallback(req.Agent, "unassigned"),
		Branch:        fmt.Sprintf("task/%s", id),
		Worktree:      filepath.ToSlash(filepath.Join(".thanos", "worktrees", id)),
		UpdatedAt:     now.Format(time.RFC3339),
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
			task.SchemaVersion = SchemaVersionTask
			task.Status = nextStatus
			task.UpdatedAt = p.now().UTC().Format(time.RFC3339)
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

func (p *RealProvider) CreateSpec(req CreateSpecRequest) (*SpecNodeInfo, error) {
	root, err := normalizeWorkspaceRoot(req.Root)
	if err != nil {
		return nil, err
	}
	title := strings.TrimSpace(req.Title)
	if title == "" {
		return nil, fmt.Errorf("spec title is required")
	}
	state := fallback(req.State, "drafted")
	body := strings.TrimSpace(req.Body)
	if body == "" {
		body = "Describe the goal, constraints, acceptance criteria, and dispatch plan."
	}
	rel := filepath.Join("specs", stableID(title)+".md")
	path := filepath.Join(root, rel)
	content := fmt.Sprintf("# %s\n\nstate: %s\n\n%s\n", title, state, body)
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
	node := SpecNodeInfo{ID: stableID(rel), Title: title, State: state, Path: filepath.ToSlash(rel), Children: []SpecNodeInfo{}}
	return &node, nil
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
	catalog := []ProviderInfo{
		{ID: "claude", Name: "Claude Code", Command: "claude", Type: "cli", SetupHint: "Install Claude Code and authenticate it before assigning tasks.", SupportsRun: true},
		{ID: "codex", Name: "Codex", Command: "codex", Type: "cli", SetupHint: "Install Codex and ensure `codex` is on PATH.", SupportsRun: true},
		{ID: "gemini", Name: "Gemini CLI", Command: "gemini", Type: "cli", SetupHint: "Install Gemini CLI and ensure `gemini` is on PATH.", SupportsRun: true},
		{ID: "opencode", Name: "OpenCode", Command: "opencode", Type: "cli", SetupHint: "Install OpenCode and finish setup before assigning it.", SupportsRun: true},
		{ID: "cursor", Name: "Cursor Agent", Command: "cursor-agent", Type: "cli", SetupHint: "Install Cursor Agent and ensure `cursor-agent` is on PATH.", SupportsRun: true},
		{ID: "aider", Name: "Aider", Command: "aider", Type: "cli", SetupHint: "Install Aider and ensure `aider` is on PATH.", SupportsRun: true},
		{ID: "goose", Name: "Goose", Command: "goose", Type: "cli", SetupHint: "Install Goose and ensure `goose` is on PATH.", SupportsRun: true},
		{ID: "shell", Name: "Shell", Command: "sh", Type: "shell", Status: "installed", SetupHint: "System shell used for explicit user-approved commands.", SupportsRun: false},
	}

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
		title := markdownTitle(path)
		if title == "" {
			title = strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
		}
		nodes = append(nodes, SpecNodeInfo{
			ID:       stableID(rel),
			Title:    title,
			State:    inferSpecState(path),
			Path:     filepath.ToSlash(rel),
			Children: []SpecNodeInfo{},
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(nodes, func(i, j int) bool { return nodes[i].Path < nodes[j].Path })
	return nodes, nil
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
	data, err := os.ReadFile(path)
	if err != nil {
		return "drafted"
	}
	text := strings.ToLower(string(data))
	switch {
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
	return TaskInfo{
		ID:        id,
		Title:     title,
		Prompt:    prompt,
		Status:    status,
		Flow:      fallback(stringFrom(raw, "flow", "Flow"), "implement"),
		Agent:     fallback(stringFrom(raw, "agent", "Agent", "harness", "Harness"), "unassigned"),
		Branch:    stringFrom(raw, "branch", "Branch", "branch_name", "BranchName"),
		Worktree:  stringFrom(raw, "worktree", "Worktree", "worktree_path", "WorktreePath"),
		UpdatedAt: fallback(stringFrom(raw, "updated_at", "UpdatedAt", "updatedAt"), "unknown"),
		UsageUSD:  floatFrom(raw, "usage_usd", "usageUsd", "cost_usd", "CostUSD"),
	}, true
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
	defaultHarness := firstInstalled(providers, "codex", "claude", "opencode", "cursor")
	if defaultHarness == "" {
		defaultHarness = "codex"
	}
	return []AgentRoleInfo{
		{ID: "impl", Role: "Implementation", Harness: displayHarness(defaultHarness), Model: "provider default", Capabilities: []string{"workspace.read", "workspace.write", "board.context"}},
		{ID: "test", Role: "Testing", Harness: displayHarness(firstNonEmpty(firstInstalled(providers, "claude", "codex"), defaultHarness)), Model: "provider default", Capabilities: []string{"workspace.read", "commands.run"}},
		{ID: "oversight", Role: "Oversight", Harness: displayHarness(firstNonEmpty(firstInstalled(providers, "claude", "codex"), defaultHarness)), Model: "provider default", Capabilities: []string{"diff.read", "timeline.read", "risk.review"}},
		{ID: "title", Role: "Title", Harness: "Shell", Model: "none", Capabilities: []string{"metadata.write"}},
		{ID: "commit-msg", Role: "Commit Message", Harness: displayHarness(defaultHarness), Model: "provider default", Capabilities: []string{"diff.read", "metadata.write"}},
	}
}

func builtinFlows() []FlowInfo {
	return []FlowInfo{
		{ID: "implement", Name: "Implement", Steps: []string{"Implementation", "Testing", "Commit Message", "Title", "Oversight"}},
		{ID: "plan-first", Name: "Plan First", Steps: []string{"Planning", "Implementation", "Testing", "Oversight"}},
		{ID: "oversight-only", Name: "Oversight Only", Steps: []string{"Oversight"}},
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
	case "claude":
		return "Claude"
	case "codex":
		return "Codex"
	case "cursor":
		return "Cursor"
	case "opencode":
		return "OpenCode"
	case "gemini":
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

func normalizeTaskStatus(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "in_progress", "running", "committing":
		return "running"
	case "waiting", "waiting_user", "waiting_approval":
		return "waiting"
	case "in_review", "review":
		return "review"
	case "done", "complete", "completed":
		return "done"
	case "failed", "cancelled":
		return "failed"
	default:
		return "backlog"
	}
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
