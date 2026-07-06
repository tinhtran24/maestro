package bridge

import (
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/tinhtran/thanos/internal/agents"
)

func TestDetectAgentCLIsMapsProviders(t *testing.T) {
	service := NewServiceWithDetector(func(context.Context) ([]agents.Provider, error) {
		return []agents.Provider{
			{
				ID:           "codex",
				Name:         "Codex",
				Command:      "codex",
				DetectedPath: "/usr/local/bin/codex",
				Status:       agents.StatusInstalled,
				Version:      "1.0.0",
				Type:         "cli",
				Enabled:      true,
				SetupHint:    "ready",
			},
		}, nil
	})

	providers, err := service.DetectAgentCLIs(context.Background())
	if err != nil {
		t.Fatalf("DetectAgentCLIs returned error: %v", err)
	}
	if len(providers) != 1 {
		t.Fatalf("expected 1 provider, got %d", len(providers))
	}
	got := providers[0]
	if got.ID != "codex" || !got.Installed || got.Path == nil || *got.Path != "/usr/local/bin/codex" {
		t.Fatalf("provider was not mapped correctly: %+v", got)
	}
}

func TestReadOnlyWorkbenchState(t *testing.T) {
	workspace := testWorkspace(t)
	db := openTestDB(t, workspace)
	defer db.Close()

	plan := ExecutionPlanInfo{
		ID:             "plan-T-1",
		TaskID:         "T-1",
		Summary:        "Implement read-only Wails bridge.",
		Steps:          []PlanStepInfo{{ID: "step-1", Title: "Load", Description: "Read snapshot", Status: "done"}},
		Risks:          []string{"schema drift"},
		FilesToTouch:   []string{"internal/desktop/bridge/service.go"},
		TestStrategy:   []string{"go test"},
		ApprovalStatus: "approved",
	}
	insertJSONDoc(t, db, "execution_plans", "task_id", "T-1", plan)
	insertJSONDoc(t, db, "tasks", "id", "T-1", map[string]any{
		"id":              "T-1",
		"title":           "Bridge reads",
		"status":          "ready",
		"priority":        "P1",
		"assigned_agent":  "Codex",
		"review_approved": false,
		"tests_passed":    false,
		"updated_at":      "2026-07-06T00:00:00Z",
		"tags":            []string{"desktop"},
	})
	insertMemoryNode(t, db, "memory-1", "Wails bridge", "Read-only bridge migration")

	snapshot, err := NewService().LoadWorkbenchState(workspace)
	if err != nil {
		t.Fatalf("LoadWorkbenchState returned error: %v", err)
	}
	if snapshot.Project == nil || snapshot.Project.Name != "Test Project" {
		t.Fatalf("unexpected project: %+v", snapshot.Project)
	}
	if len(snapshot.Plans) != 1 || snapshot.Plans[0].TaskID != "T-1" {
		t.Fatalf("unexpected plans: %+v", snapshot.Plans)
	}
	if len(snapshot.Tasks) != 1 || snapshot.Tasks[0].Status != "ready" {
		t.Fatalf("unexpected tasks: %+v", snapshot.Tasks)
	}
	if len(snapshot.MemoryNodes) != 1 || snapshot.MemoryNodes[0].Title != "Wails bridge" {
		t.Fatalf("unexpected memory nodes: %+v", snapshot.MemoryNodes)
	}
}

func TestReadExecutionPlan(t *testing.T) {
	workspace := testWorkspace(t)
	db := openTestDB(t, workspace)
	defer db.Close()

	insertJSONDoc(t, db, "execution_plans", "task_id", "T-2", ExecutionPlanInfo{
		ID:             "plan-T-2",
		TaskID:         "T-2",
		Summary:        "Read plan",
		ApprovalStatus: "pending",
	})

	plan, err := NewService().ReadExecutionPlan(TaskPlanRequest{Workspace: workspace, TaskID: "T-2"})
	if err != nil {
		t.Fatalf("ReadExecutionPlan returned error: %v", err)
	}
	if plan.Summary != "Read plan" || plan.ApprovalStatus != "pending" {
		t.Fatalf("unexpected plan: %+v", plan)
	}
}

func TestSearchMemory(t *testing.T) {
	workspace := testWorkspace(t)
	db := openTestDB(t, workspace)
	defer db.Close()
	insertMemoryNode(t, db, "memory-1", "Bridge decision", "Wails reads project memory through SQLite FTS")
	insertMemoryNode(t, db, "memory-2", "Other", "Unrelated content")

	nodes, err := NewService().SearchMemory(MemorySearchRequest{Workspace: workspace, Query: "Wails"})
	if err != nil {
		t.Fatalf("SearchMemory returned error: %v", err)
	}
	if len(nodes) != 1 || nodes[0].ID != "memory-1" {
		t.Fatalf("unexpected search results: %+v", nodes)
	}

	blank, err := NewService().SearchMemory(MemorySearchRequest{Workspace: workspace, Query: " "})
	if err != nil {
		t.Fatalf("blank SearchMemory returned error: %v", err)
	}
	if len(blank) != 0 {
		t.Fatalf("expected no results for blank query, got %+v", blank)
	}
}

func TestCreateOrImportProjectWritesMetadata(t *testing.T) {
	root := t.TempDir()

	project, err := NewService().CreateOrImportProject(ProjectSetupRequest{
		RootPath:       root,
		Name:           "Imported Repo",
		DefaultBranch:  "main",
		WorktreeRoot:   ".thanos/worktrees",
		PackageManager: "npm",
		DevCommand:     "npm run dev",
		TestCommand:    "npm test",
	})
	if err != nil {
		t.Fatalf("CreateOrImportProject returned error: %v", err)
	}
	if project.Name != "Imported Repo" || project.RootPath != root {
		t.Fatalf("unexpected project: %+v", project)
	}
	for _, path := range []string{
		".thanos/settings.json",
		".thanos/config.json",
		".thanos/agents.yaml",
		".thanos/tasks",
		".thanos/memory",
	} {
		if _, err := os.Stat(filepath.Join(root, path)); err != nil {
			t.Fatalf("expected %s to exist: %v", path, err)
		}
	}
}

func testWorkspace(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".thanos", "memory"), 0o755); err != nil {
		t.Fatal(err)
	}
	settings := `{
  "project": {
    "name": "Test Project",
    "default_branch": "main",
    "worktree_root": ".thanos/worktrees",
    "package_manager": "npm",
    "dev_command": "npm run dev",
    "test_command": "npm test",
    "created_at": "2026-07-06T00:00:00Z",
    "updated_at": "2026-07-06T00:00:00Z"
  },
  "default_runner": "codex"
}`
	if err := os.WriteFile(filepath.Join(root, ".thanos", "settings.json"), []byte(settings), 0o644); err != nil {
		t.Fatal(err)
	}
	return root
}

func openTestDB(t *testing.T, workspace string) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", filepath.Join(workspace, ".thanos", "memory", "workbench.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`
CREATE TABLE tasks (id TEXT PRIMARY KEY, data_json TEXT NOT NULL);
CREATE TABLE execution_plans (task_id TEXT PRIMARY KEY, data_json TEXT NOT NULL);
CREATE TABLE reviews (task_id TEXT PRIMARY KEY, data_json TEXT NOT NULL);
CREATE TABLE memory_nodes (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  node_type TEXT NOT NULL,
  title TEXT NOT NULL,
  content TEXT NOT NULL,
  links_json TEXT NOT NULL,
  created_at INTEGER NOT NULL
);
CREATE VIRTUAL TABLE memory_nodes_fts USING fts5(title, content, content='memory_nodes', content_rowid='rowid');
CREATE TRIGGER memory_nodes_ai AFTER INSERT ON memory_nodes BEGIN
  INSERT INTO memory_nodes_fts(rowid, title, content) VALUES (new.rowid, new.title, new.content);
END;`)
	if err != nil {
		t.Fatal(err)
	}
	return db
}

func insertJSONDoc(t *testing.T, db *sql.DB, table string, keyColumn string, key string, value any) {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec("INSERT INTO "+table+" ("+keyColumn+", data_json) VALUES (?, ?)", key, string(data))
	if err != nil {
		t.Fatal(err)
	}
}

func insertMemoryNode(t *testing.T, db *sql.DB, id, title, content string) {
	t.Helper()
	_, err := db.Exec(
		`INSERT INTO memory_nodes (id, project_id, node_type, title, content, links_json, created_at) VALUES (?, 'local', 'decision', ?, ?, '[]', 1783296000)`,
		id,
		title,
		content,
	)
	if err != nil {
		t.Fatal(err)
	}
}
