package bridge

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/tinhtran/thanos/internal/agents"
	_ "modernc.org/sqlite"
)

type Service struct {
	detect func(context.Context) ([]agents.Provider, error)
}

func NewService() *Service {
	registry := agents.DefaultAdapterRegistry()
	return &Service{
		detect: func(ctx context.Context) ([]agents.Provider, error) {
			return registry.DetectAll(ctx, agents.DefaultDetectionContext())
		},
	}
}

func NewServiceWithDetector(detector func(context.Context) ([]agents.Provider, error)) *Service {
	return &Service{detect: detector}
}

func (s *Service) DetectAgentCLIs(ctx context.Context) ([]AgentCandidateInfo, error) {
	if s.detect == nil {
		s = NewService()
	}
	providers, err := s.detect(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]AgentCandidateInfo, 0, len(providers))
	for _, provider := range providers {
		path := optionalString(provider.DetectedPath)
		version := optionalString(provider.Version)
		out = append(out, AgentCandidateInfo{
			ID:        provider.ID,
			Name:      provider.Name,
			Command:   provider.Command,
			Installed: provider.Status == agents.StatusInstalled,
			Path:      path,
			Status:    string(provider.Status),
			Version:   version,
			AgentType: provider.Type,
			Enabled:   provider.Enabled,
			SetupHint: provider.SetupHint,
		})
	}
	return out, nil
}

func (s *Service) LoadWorkbenchState(workspace string) (*WorkbenchSnapshot, error) {
	root, err := validateWorkspace(workspace)
	if err != nil {
		return nil, err
	}
	repo := repository{workspace: root}
	return repo.load()
}

func (s *Service) ReadExecutionPlan(request TaskPlanRequest) (*ExecutionPlanInfo, error) {
	root, err := validateWorkspace(request.Workspace)
	if err != nil {
		return nil, err
	}
	taskID, err := sanitizeID(request.TaskID)
	if err != nil {
		return nil, err
	}
	repo := repository{workspace: root}
	return repo.readExecutionPlan(taskID)
}

func (s *Service) SearchMemory(request MemorySearchRequest) ([]MemoryNodeInfo, error) {
	root, err := validateWorkspace(request.Workspace)
	if err != nil {
		return nil, err
	}
	query := strings.TrimSpace(request.Query)
	if query == "" {
		return []MemoryNodeInfo{}, nil
	}
	db, err := openExistingDB(root)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []MemoryNodeInfo{}, nil
		}
		return nil, err
	}
	defer db.Close()
	return searchMemory(db, query)
}

func (s *Service) CreateOrImportProject(request ProjectSetupRequest) (*ProjectInfo, error) {
	root, err := validateWorkspace(request.RootPath)
	if err != nil {
		return nil, err
	}
	if err := writeProjectFiles(root, request); err != nil {
		return nil, err
	}
	repo := repository{workspace: root}
	return repo.loadProject()
}

type repository struct {
	workspace string
}

func (r repository) load() (*WorkbenchSnapshot, error) {
	project, err := r.loadProject()
	if err != nil {
		return nil, err
	}
	plans, err := r.loadPlans()
	if err != nil {
		return nil, err
	}
	reviews, err := r.loadReviews()
	if err != nil {
		return nil, err
	}
	sessions, err := readJSONFiles[AgentSessionInfo](filepath.Join(r.workspace, ".thanos", "logs", "sessions"))
	if err != nil {
		return nil, err
	}
	sort.Slice(sessions, func(i, j int) bool { return sessions[i].StartedAt > sessions[j].StartedAt })
	memoryNodes, err := r.loadMemoryNodes()
	if err != nil {
		return nil, err
	}
	tasks, err := r.loadTasks(project, plans, reviews)
	if err != nil {
		return nil, err
	}
	return &WorkbenchSnapshot{
		Project:     project,
		Features:    localFeatures(project),
		Tasks:       tasks,
		Plans:       plans,
		Sessions:    sessions,
		Reviews:     reviews,
		MemoryNodes: memoryNodes,
		Skills:      []SkillInfo{},
		SkillRuns:   []SkillRunInfo{},
	}, nil
}

func (r repository) loadProject() (*ProjectInfo, error) {
	path := filepath.Join(r.workspace, ".thanos", "settings.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", path, err)
	}
	var settingsFile map[string]any
	if err := json.Unmarshal(data, &settingsFile); err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", path, err)
	}
	project, _ := settingsFile["project"].(map[string]any)
	name := firstString(project, "name", "Thanos Workspace")
	defaultBranch := firstString(project, "default_branch", "main")
	worktreeRoot := firstString(project, "worktree_root", ".thanos/worktrees")
	packageManager := stringField(project, "package_manager")
	devCommand := stringField(project, "dev_command")
	testCommand := stringField(project, "test_command")
	createdAt := firstString(project, "created_at", "1970-01-01T00:00:00Z")
	updatedAt := firstString(project, "updated_at", createdAt)
	settings := map[string]string{
		"defaultBranch":  defaultBranch,
		"worktreeRoot":   worktreeRoot,
		"packageManager": packageManager,
		"devCommand":     devCommand,
		"testCommand":    testCommand,
	}
	for source, target := range map[string]string{
		"language":       "language",
		"default_runner": "defaultRunner",
		"locale":         "locale",
	} {
		if value := stringField(project, source); value != "" {
			settings[target] = value
		} else if rootValue := stringField(settingsFile, source); rootValue != "" {
			settings[target] = rootValue
		}
	}
	return &ProjectInfo{
		ID:             slugID(name),
		Name:           name,
		RootPath:       r.workspace,
		GitRemoteURL:   stringField(project, "git_remote_url"),
		DefaultBranch:  defaultBranch,
		WorktreeRoot:   worktreeRoot,
		PackageManager: packageManager,
		DevCommand:     devCommand,
		TestCommand:    testCommand,
		CreatedAt:      createdAt,
		UpdatedAt:      updatedAt,
		Repos:          []string{filepath.Base(r.workspace)},
		Settings:       settings,
	}, nil
}

func (r repository) loadPlans() ([]ExecutionPlanInfo, error) {
	db, err := openExistingDB(r.workspace)
	if errors.Is(err, os.ErrNotExist) {
		return []ExecutionPlanInfo{}, nil
	}
	if err != nil {
		return nil, err
	}
	defer db.Close()
	plans, err := readPlans(db)
	if err != nil {
		return nil, err
	}
	sort.Slice(plans, func(i, j int) bool { return plans[i].TaskID < plans[j].TaskID })
	return plans, nil
}

func (r repository) readExecutionPlan(taskID string) (*ExecutionPlanInfo, error) {
	db, err := openExistingDB(r.workspace)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	plan, err := readPlan(db, taskID)
	if err != nil {
		return nil, err
	}
	return plan, nil
}

func (r repository) loadReviews() ([]ReviewInfo, error) {
	db, err := openExistingDB(r.workspace)
	if errors.Is(err, os.ErrNotExist) {
		return []ReviewInfo{}, nil
	}
	if err != nil {
		return nil, err
	}
	defer db.Close()
	reviews, err := readReviews(db)
	if err != nil {
		return nil, err
	}
	sort.Slice(reviews, func(i, j int) bool { return reviews[i].TaskID < reviews[j].TaskID })
	return reviews, nil
}

func (r repository) loadMemoryNodes() ([]MemoryNodeInfo, error) {
	db, err := openExistingDB(r.workspace)
	if errors.Is(err, os.ErrNotExist) {
		return []MemoryNodeInfo{}, nil
	}
	if err != nil {
		return nil, err
	}
	defer db.Close()
	return readMemoryNodes(db)
}

func (r repository) loadTasks(project *ProjectInfo, plans []ExecutionPlanInfo, reviews []ReviewInfo) ([]TaskInfo, error) {
	db, err := openExistingDB(r.workspace)
	if errors.Is(err, os.ErrNotExist) {
		return tasksFromPlans(project, plans, reviews), nil
	}
	if err != nil {
		return nil, err
	}
	defer db.Close()
	values, err := readTaskDocs(db)
	if err != nil {
		return nil, err
	}
	tasks := make([]TaskInfo, 0, len(values)+len(plans))
	seen := map[string]bool{}
	for _, value := range values {
		task, ok := taskFromValue(project, value)
		if !ok {
			continue
		}
		tasks = append(tasks, task)
		seen[task.ID] = true
	}
	for _, plan := range plans {
		if seen[plan.TaskID] {
			continue
		}
		tasks = append(tasks, taskFromPlan(project, plan, findReview(reviews, plan.TaskID)))
	}
	sort.Slice(tasks, func(i, j int) bool { return tasks[i].ID < tasks[j].ID })
	return tasks, nil
}

func openExistingDB(workspace string) (*sql.DB, error) {
	path := filepath.Join(workspace, ".thanos", "memory", "workbench.sqlite")
	if _, err := os.Stat(path); err != nil {
		return nil, err
	}
	return sql.Open("sqlite", path)
}

func readPlans(db *sql.DB) ([]ExecutionPlanInfo, error) {
	if hasColumn(db, "execution_plans", "data_json") {
		values, err := readJSONDocs(db, "execution_plans", "data_json")
		if err != nil {
			return nil, err
		}
		out := make([]ExecutionPlanInfo, 0, len(values))
		for _, value := range values {
			var plan ExecutionPlanInfo
			if err := json.Unmarshal(value, &plan); err != nil {
				return nil, err
			}
			out = append(out, plan)
		}
		return out, nil
	}
	rows, err := db.Query(`SELECT id, task_id, summary, steps_json, risks_json, files_to_touch_json, test_strategy_json, approval_status FROM execution_plans`)
	if err != nil {
		if isNoSuchTable(err) {
			return []ExecutionPlanInfo{}, nil
		}
		return nil, err
	}
	defer rows.Close()
	var out []ExecutionPlanInfo
	for rows.Next() {
		var plan ExecutionPlanInfo
		var steps, risks, files, tests string
		if err := rows.Scan(&plan.ID, &plan.TaskID, &plan.Summary, &steps, &risks, &files, &tests, &plan.ApprovalStatus); err != nil {
			return nil, err
		}
		_ = json.Unmarshal([]byte(steps), &plan.Steps)
		_ = json.Unmarshal([]byte(risks), &plan.Risks)
		_ = json.Unmarshal([]byte(files), &plan.FilesToTouch)
		_ = json.Unmarshal([]byte(tests), &plan.TestStrategy)
		out = append(out, plan)
	}
	return out, rows.Err()
}

func readPlan(db *sql.DB, taskID string) (*ExecutionPlanInfo, error) {
	if hasColumn(db, "execution_plans", "data_json") {
		var raw string
		err := db.QueryRow(`SELECT data_json FROM execution_plans WHERE task_id = ?`, taskID).Scan(&raw)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("execution plan not found: %s", taskID)
		}
		if err != nil {
			return nil, err
		}
		var plan ExecutionPlanInfo
		if err := json.Unmarshal([]byte(raw), &plan); err != nil {
			return nil, err
		}
		return &plan, nil
	}
	var plan ExecutionPlanInfo
	var steps, risks, files, tests string
	err := db.QueryRow(`SELECT id, task_id, summary, steps_json, risks_json, files_to_touch_json, test_strategy_json, approval_status FROM execution_plans WHERE task_id = ?`, taskID).
		Scan(&plan.ID, &plan.TaskID, &plan.Summary, &steps, &risks, &files, &tests, &plan.ApprovalStatus)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("execution plan not found: %s", taskID)
	}
	if err != nil {
		return nil, err
	}
	_ = json.Unmarshal([]byte(steps), &plan.Steps)
	_ = json.Unmarshal([]byte(risks), &plan.Risks)
	_ = json.Unmarshal([]byte(files), &plan.FilesToTouch)
	_ = json.Unmarshal([]byte(tests), &plan.TestStrategy)
	return &plan, nil
}

func readReviews(db *sql.DB) ([]ReviewInfo, error) {
	if hasColumn(db, "reviews", "data_json") {
		values, err := readJSONDocs(db, "reviews", "data_json")
		if err != nil {
			return nil, err
		}
		out := make([]ReviewInfo, 0, len(values))
		for _, value := range values {
			var review ReviewInfo
			if err := json.Unmarshal(value, &review); err != nil {
				return nil, err
			}
			out = append(out, review)
		}
		return out, nil
	}
	rows, err := db.Query(`SELECT id, task_id, diff_summary, changed_files_json, test_results_json, reviewer_notes, status FROM reviews`)
	if err != nil {
		if isNoSuchTable(err) {
			return []ReviewInfo{}, nil
		}
		return nil, err
	}
	defer rows.Close()
	var out []ReviewInfo
	for rows.Next() {
		var review ReviewInfo
		var changed, tests string
		if err := rows.Scan(&review.ID, &review.TaskID, &review.DiffSummary, &changed, &tests, &review.ReviewerNotes, &review.Status); err != nil {
			return nil, err
		}
		_ = json.Unmarshal([]byte(changed), &review.ChangedFiles)
		_ = json.Unmarshal([]byte(tests), &review.TestResults)
		out = append(out, review)
	}
	return out, rows.Err()
}

func readTaskDocs(db *sql.DB) ([]map[string]any, error) {
	if hasColumn(db, "tasks", "data_json") {
		values, err := readJSONDocs(db, "tasks", "data_json")
		if err != nil {
			return nil, err
		}
		out := make([]map[string]any, 0, len(values))
		for _, value := range values {
			var task map[string]any
			if err := json.Unmarshal(value, &task); err != nil {
				return nil, err
			}
			out = append(out, task)
		}
		return out, nil
	}
	rows, err := db.Query(`SELECT id, feature_id, parent_task_id, title, description, status, priority, assigned_agent, executor_profile, worktree_path, branch_name, review_approved, tests_passed, updated_at FROM tasks`)
	if err != nil {
		if isNoSuchTable(err) {
			return []map[string]any{}, nil
		}
		return nil, err
	}
	defer rows.Close()
	var out []map[string]any
	for rows.Next() {
		var featureID, parentID, assigned, executor, worktree, branch sql.NullString
		var reviewApproved, testsPassed int
		task := map[string]any{}
		if err := rows.Scan(
			scanMapString(task, "id"),
			&featureID,
			&parentID,
			scanMapString(task, "title"),
			scanMapString(task, "description"),
			scanMapString(task, "status"),
			scanMapString(task, "priority"),
			&assigned,
			&executor,
			&worktree,
			&branch,
			&reviewApproved,
			&testsPassed,
			scanMapString(task, "updated_at"),
		); err != nil {
			return nil, err
		}
		putNullString(task, "feature_id", featureID)
		putNullString(task, "parent_task_id", parentID)
		putNullString(task, "assigned_agent", assigned)
		putNullString(task, "executor_profile", executor)
		putNullString(task, "worktree_path", worktree)
		putNullString(task, "branch_name", branch)
		task["review_approved"] = reviewApproved != 0
		task["tests_passed"] = testsPassed != 0
		out = append(out, task)
	}
	return out, rows.Err()
}

func readMemoryNodes(db *sql.DB) ([]MemoryNodeInfo, error) {
	typeColumn := memoryTypeColumn(db)
	if typeColumn == "" {
		return []MemoryNodeInfo{}, nil
	}
	rows, err := db.Query(fmt.Sprintf(`SELECT id, project_id, %s, title, content, links_json, created_at FROM memory_nodes ORDER BY created_at DESC LIMIT 50`, typeColumn))
	if err != nil {
		if isNoSuchTable(err) {
			return []MemoryNodeInfo{}, nil
		}
		return nil, err
	}
	defer rows.Close()
	return scanMemoryRows(rows)
}

func searchMemory(db *sql.DB, query string) ([]MemoryNodeInfo, error) {
	typeColumn := memoryTypeColumn(db)
	if typeColumn == "" {
		return []MemoryNodeInfo{}, nil
	}
	rows, err := db.Query(fmt.Sprintf(`
		SELECT m.id, m.project_id, m.%s, m.title, m.content, m.links_json, m.created_at
		FROM memory_nodes_fts f
		JOIN memory_nodes m ON m.rowid = f.rowid
		WHERE memory_nodes_fts MATCH ?
		ORDER BY rank
		LIMIT 20`, typeColumn), query)
	if err != nil {
		if isNoSuchTable(err) {
			return []MemoryNodeInfo{}, nil
		}
		return nil, err
	}
	defer rows.Close()
	return scanMemoryRows(rows)
}

func scanMemoryRows(rows *sql.Rows) ([]MemoryNodeInfo, error) {
	var out []MemoryNodeInfo
	for rows.Next() {
		var node MemoryNodeInfo
		var links string
		var created any
		if err := rows.Scan(&node.ID, &node.ProjectID, &node.NodeType, &node.Title, &node.Content, &links, &created); err != nil {
			return nil, err
		}
		_ = json.Unmarshal([]byte(links), &node.Links)
		node.CreatedAt = epoch(created)
		out = append(out, node)
	}
	return out, rows.Err()
}

func readJSONDocs(db *sql.DB, table, column string) ([][]byte, error) {
	rows, err := db.Query(fmt.Sprintf(`SELECT %s FROM %s`, column, table))
	if err != nil {
		if isNoSuchTable(err) {
			return nil, nil
		}
		return nil, err
	}
	defer rows.Close()
	var out [][]byte
	for rows.Next() {
		var raw string
		if err := rows.Scan(&raw); err != nil {
			return nil, err
		}
		out = append(out, []byte(raw))
	}
	return out, rows.Err()
}

func hasColumn(db *sql.DB, table, column string) bool {
	rows, err := db.Query(`PRAGMA table_info(` + table + `)`)
	if err != nil {
		return false
	}
	defer rows.Close()
	for rows.Next() {
		var cid int
		var name, typ string
		var notNull int
		var defaultValue any
		var pk int
		if err := rows.Scan(&cid, &name, &typ, &notNull, &defaultValue, &pk); err != nil {
			return false
		}
		if name == column {
			return true
		}
	}
	return false
}

func memoryTypeColumn(db *sql.DB) string {
	if hasColumn(db, "memory_nodes", "node_type") {
		return "node_type"
	}
	if hasColumn(db, "memory_nodes", "type") {
		return "type"
	}
	return ""
}

func validateWorkspace(workspace string) (string, error) {
	if strings.TrimSpace(workspace) == "" {
		return "", errors.New("workspace is required")
	}
	root, err := filepath.Abs(workspace)
	if err != nil {
		return "", err
	}
	info, err := os.Stat(root)
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		return "", fmt.Errorf("workspace is not a directory: %s", root)
	}
	return root, nil
}

func writeProjectFiles(root string, request ProjectSetupRequest) error {
	dot := filepath.Join(root, ".thanos")
	for _, dir := range []string{"tasks", "plans", "logs", "reviews", "tests", "worktrees", "memory", "events"} {
		if err := os.MkdirAll(filepath.Join(dot, dir), 0o755); err != nil {
			return fmt.Errorf("failed to create %s: %w", filepath.Join(dot, dir), err)
		}
	}
	now := time.Now().UTC().Format(time.RFC3339)
	name := strings.TrimSpace(request.Name)
	if name == "" {
		name = filepath.Base(root)
	}
	if name == "." || name == string(filepath.Separator) || name == "" {
		name = "Thanos Project"
	}
	defaultBranch := strings.TrimSpace(request.DefaultBranch)
	if defaultBranch == "" {
		defaultBranch = "main"
	}
	worktreeRoot := strings.TrimSpace(request.WorktreeRoot)
	if worktreeRoot == "" {
		worktreeRoot = ".thanos/worktrees"
	}
	settings := map[string]any{
		"project": map[string]any{
			"name":            name,
			"language":        "",
			"framework":       "",
			"git_remote_url":  strings.TrimSpace(request.GitRemoteURL),
			"default_branch":  defaultBranch,
			"worktree_root":   worktreeRoot,
			"package_manager": strings.TrimSpace(request.PackageManager),
			"dev_command":     strings.TrimSpace(request.DevCommand),
			"test_command":    strings.TrimSpace(request.TestCommand),
			"created_at":      now,
			"updated_at":      now,
		},
		"default_runner": "codex",
		"locale":         "en",
	}
	if err := writeJSONFile(filepath.Join(dot, "settings.json"), settings); err != nil {
		return err
	}
	config := map[string]any{
		"project":         settings["project"],
		"workflow_agents": defaultWorkflowAgentConfig(),
		"updated_at":      now,
	}
	if err := writeJSONFile(filepath.Join(dot, "config.json"), config); err != nil {
		return err
	}
	agentsPath := filepath.Join(dot, "agents.yaml")
	if _, err := os.Stat(agentsPath); errors.Is(err, os.ErrNotExist) {
		data := "agents:\n  - name: codex\n    command: codex\n    args: []\n    env:\n    role: implementation\n    allowed_steps: [plan, execute]\n"
		if err := os.WriteFile(agentsPath, []byte(data), 0o644); err != nil {
			return fmt.Errorf("failed to write %s: %w", agentsPath, err)
		}
	} else if err != nil {
		return fmt.Errorf("failed to inspect %s: %w", agentsPath, err)
	}
	return nil
}

func writeJSONFile(path string, value any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("failed to create %s: %w", filepath.Dir(path), err)
	}
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("failed to write %s: %w", path, err)
	}
	return nil
}

func defaultWorkflowAgentConfig() map[string]any {
	return map[string]any{
		"planning": map[string]any{
			"enabled": true, "provider": "Claude Code", "command": "claude", "working_directory_mode": "project",
			"auto_start_terminal": true, "approval_required": true, "env": map[string]string{}, "timeout": "30m",
			"permissions": []string{"read", "write-plans"},
		},
		"coding": map[string]any{
			"enabled": true, "provider": "Codex", "command": "codex", "working_directory_mode": "worktree",
			"auto_start_terminal": true, "approval_required": true, "env": map[string]string{}, "timeout": "30m",
			"permissions": []string{"read", "write-code", "run-tests"},
		},
		"review": map[string]any{
			"enabled": true, "provider": "Claude Code", "command": "claude", "working_directory_mode": "worktree",
			"auto_start_terminal": true, "approval_required": true, "env": map[string]string{}, "timeout": "30m",
			"permissions": []string{"read", "inspect-diff"},
		},
		"testing": map[string]any{
			"enabled": true, "provider": "Shell", "command": "npm test", "working_directory_mode": "worktree",
			"auto_start_terminal": true, "approval_required": false, "env": map[string]string{}, "timeout": "20m",
			"permissions": []string{"run-tests"},
		},
	}
}

func sanitizeID(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", errors.New("id is required")
	}
	for _, ch := range value {
		if !(ch == '-' || ch == '_' || ch == '.' || ch == '/' || ch >= '0' && ch <= '9' || ch >= 'a' && ch <= 'z' || ch >= 'A' && ch <= 'Z') {
			return "", fmt.Errorf("invalid id: %s", value)
		}
	}
	return value, nil
}

func localFeatures(project *ProjectInfo) []FeatureInfo {
	planGraphID := "local-plan-graph"
	return []FeatureInfo{{
		ID:          "local",
		ProjectID:   project.ID,
		Title:       project.Name,
		Description: "Local Thanos workspace",
		Status:      "active",
		PlanGraphID: &planGraphID,
		CreatedAt:   "1970-01-01T00:00:00Z",
	}}
}

func taskFromValue(project *ProjectInfo, value map[string]any) (TaskInfo, bool) {
	id := stringField(value, "id")
	if id == "" {
		return TaskInfo{}, false
	}
	status := mapTaskStatus(firstString(value, "status", "backlog"))
	reviewApproved := boolField(value, "review_approved")
	testsPassed := boolField(value, "tests_passed")
	return TaskInfo{
		ID:              id,
		FeatureID:       firstString(value, "feature_id", "local"),
		ParentTaskID:    optionalString(stringField(value, "parent_task_id")),
		Title:           firstString(value, "title", id),
		Description:     stringField(value, "description"),
		Status:          status,
		Priority:        normalizePriority(stringField(value, "priority")),
		AssignedAgent:   firstString(value, "assigned_agent", "Unassigned"),
		ExecutorProfile: firstString(value, "executor_profile", firstSetting(project.Settings, "defaultRunner", "codex")),
		WorktreePath:    stringField(value, "worktree_path"),
		BranchName:      stringField(value, "branch_name"),
		ReviewApproved:  reviewApproved,
		TestsPassed:     testsPassed,
		UpdatedAt:       firstString(value, "updated_at", "unknown"),
		Tags:            stringArrayField(value, "tags"),
		Progress:        progressForStatus(status, reviewApproved, testsPassed),
	}, true
}

func taskFromPlan(project *ProjectInfo, plan ExecutionPlanInfo, review *ReviewInfo) TaskInfo {
	approved := plan.ApprovalStatus == "approved"
	reviewApproved := review != nil && review.Status == "approved"
	status := "waiting_approval"
	if reviewApproved {
		status = "done"
	} else if approved {
		status = "ready"
	}
	return TaskInfo{
		ID:              plan.TaskID,
		FeatureID:       "local",
		Title:           titleFromTaskID(plan.TaskID),
		Description:     plan.Summary,
		Status:          status,
		Priority:        "P2",
		AssignedAgent:   "Planner",
		ExecutorProfile: firstSetting(project.Settings, "defaultRunner", "codex"),
		WorktreePath:    ".thanos/worktrees/" + plan.TaskID,
		BranchName:      "thanos/" + strings.ToLower(plan.TaskID),
		ReviewApproved:  reviewApproved,
		TestsPassed:     false,
		UpdatedAt:       "from plan artifact",
		Tags:            tagsFromFiles(plan.FilesToTouch),
		Progress:        progressForStatus(status, reviewApproved, false),
	}
}

func tasksFromPlans(project *ProjectInfo, plans []ExecutionPlanInfo, reviews []ReviewInfo) []TaskInfo {
	tasks := make([]TaskInfo, 0, len(plans))
	for _, plan := range plans {
		tasks = append(tasks, taskFromPlan(project, plan, findReview(reviews, plan.TaskID)))
	}
	return tasks
}

func findReview(reviews []ReviewInfo, taskID string) *ReviewInfo {
	for i := range reviews {
		if reviews[i].TaskID == taskID {
			return &reviews[i]
		}
	}
	return nil
}

func readJSONFiles[T any](dir string) ([]T, error) {
	entries, err := os.ReadDir(dir)
	if errors.Is(err, os.ErrNotExist) {
		return []T{}, nil
	}
	if err != nil {
		return nil, err
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })
	out := []T{}
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			return nil, err
		}
		var item T
		if err := json.Unmarshal(data, &item); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, nil
}

type stringScanner struct {
	target map[string]any
	key    string
}

func scanMapString(target map[string]any, key string) *stringScanner {
	return &stringScanner{target: target, key: key}
}

func (s *stringScanner) Scan(value any) error {
	if value == nil {
		s.target[s.key] = ""
		return nil
	}
	switch typed := value.(type) {
	case string:
		s.target[s.key] = typed
	case []byte:
		s.target[s.key] = string(typed)
	default:
		s.target[s.key] = fmt.Sprint(typed)
	}
	return nil
}

func putNullString(target map[string]any, key string, value sql.NullString) {
	if value.Valid {
		target[key] = value.String
	}
}

func optionalString(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func stringField(value map[string]any, key string) string {
	if value == nil {
		return ""
	}
	raw, ok := value[key]
	if !ok || raw == nil {
		return ""
	}
	switch typed := raw.(type) {
	case string:
		return typed
	case fmt.Stringer:
		return typed.String()
	default:
		return fmt.Sprint(typed)
	}
}

func firstString(value map[string]any, key, fallback string) string {
	if out := stringField(value, key); out != "" {
		return out
	}
	return fallback
}

func firstSetting(value map[string]string, key, fallback string) string {
	if out := value[key]; out != "" {
		return out
	}
	return fallback
}

func boolField(value map[string]any, key string) bool {
	raw, ok := value[key]
	if !ok {
		return false
	}
	switch typed := raw.(type) {
	case bool:
		return typed
	case float64:
		return typed != 0
	case int:
		return typed != 0
	default:
		return false
	}
}

func stringArrayField(value map[string]any, key string) []string {
	raw, ok := value[key].([]any)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(raw))
	for _, item := range raw {
		if text, ok := item.(string); ok {
			out = append(out, text)
		}
	}
	return out
}

func mapTaskStatus(status string) string {
	switch status {
	case "plan":
		return "planning"
	case "execute":
		return "running"
	case "verify":
		return "in_review"
	case "done", "blocked", "failed", "waiting_user", "waiting_approval", "ready", "running", "in_review", "planning":
		return status
	default:
		return "backlog"
	}
}

func normalizePriority(value string) string {
	switch value {
	case "P0", "P1", "P2", "P3":
		return value
	case "high":
		return "P1"
	case "low":
		return "P3"
	default:
		return "P2"
	}
}

func progressForStatus(status string, reviewApproved, testsPassed bool) uint8 {
	switch status {
	case "done":
		return 100
	case "in_review":
		if reviewApproved && testsPassed {
			return 95
		}
		return 80
	case "running":
		return 65
	case "waiting_user":
		return 60
	case "ready":
		return 45
	case "waiting_approval":
		return 30
	case "planning":
		return 20
	default:
		return 0
	}
}

func titleFromTaskID(taskID string) string {
	return strings.NewReplacer("-", " ", "_", " ").Replace(taskID)
}

func tagsFromFiles(files []string) []string {
	limit := len(files)
	if limit > 2 {
		limit = 2
	}
	out := make([]string, 0, limit)
	for _, file := range files[:limit] {
		if idx := strings.IndexByte(file, '/'); idx >= 0 {
			out = append(out, file[:idx])
		} else {
			out = append(out, file)
		}
	}
	return out
}

func slugID(value string) string {
	var out strings.Builder
	lastDash := false
	for _, ch := range strings.ToLower(value) {
		if ch >= 'a' && ch <= 'z' || ch >= '0' && ch <= '9' {
			out.WriteRune(ch)
			lastDash = false
			continue
		}
		if !lastDash {
			out.WriteByte('-')
			lastDash = true
		}
	}
	return strings.Trim(out.String(), "-")
}

func epoch(value any) uint64 {
	switch typed := value.(type) {
	case int64:
		return uint64(typed)
	case int:
		return uint64(typed)
	case float64:
		return uint64(typed)
	case []byte:
		return epochString(string(typed))
	case string:
		return epochString(typed)
	case time.Time:
		return uint64(typed.Unix())
	default:
		return 0
	}
}

func epochString(value string) uint64 {
	if parsed, err := strconv.ParseUint(value, 10, 64); err == nil {
		return parsed
	}
	if parsed, err := time.Parse(time.RFC3339, value); err == nil {
		return uint64(parsed.Unix())
	}
	return 0
}

func isNoSuchTable(err error) bool {
	return strings.Contains(err.Error(), "no such table")
}
