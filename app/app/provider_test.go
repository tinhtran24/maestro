package app

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRealProviderLoadWorkspaceReadsRealSpecsAndTasks(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "specs", "local"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "specs", "local", "workbench.md"), []byte("# Workbench\n\nstate: validated\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, ".thanos", "tasks"), 0o755); err != nil {
		t.Fatal(err)
	}
	taskJSON := `{"id":"T-1","title":"Real task","prompt":"Run real provider","status":"waiting","flow":"implement","agent":"codex","cost_usd":0.25}`
	if err := os.WriteFile(filepath.Join(root, ".thanos", "tasks", "task.json"), []byte(taskJSON), 0o644); err != nil {
		t.Fatal(err)
	}

	provider := NewRealProvider()
	provider.now = func() time.Time { return time.Date(2026, 7, 6, 9, 30, 0, 0, time.UTC) }

	workspace, err := provider.LoadWorkspace(root)
	if err != nil {
		t.Fatalf("LoadWorkspace returned error: %v", err)
	}
	if workspace.Name != filepath.Base(root) {
		t.Fatalf("workspace name = %q, want %q", workspace.Name, filepath.Base(root))
	}
	if len(workspace.Specs) != 1 || workspace.Specs[0].Title != "Workbench" || workspace.Specs[0].State != "validated" {
		t.Fatalf("specs = %#v", workspace.Specs)
	}
	if len(workspace.Tasks) != 1 || workspace.Tasks[0].ID != "T-1" || workspace.Tasks[0].Status != "waiting" {
		t.Fatalf("tasks = %#v", workspace.Tasks)
	}
	if len(workspace.Flows) == 0 || len(workspace.Agents) == 0 {
		t.Fatalf("expected built-in roles and flows")
	}
}

func TestRealProviderPersistsWorkspacePrimitives(t *testing.T) {
	root := t.TempDir()
	provider := NewRealProvider()
	provider.now = func() time.Time { return time.Date(2026, 7, 6, 9, 30, 0, 0, time.UTC) }

	task, err := provider.CreateTask(CreateTaskRequest{Root: root, Title: "Build board", Prompt: "Create a local task", Flow: "implement", Agent: "codex"})
	if err != nil {
		t.Fatalf("CreateTask returned error: %v", err)
	}
	if task.Status != "backlog" || task.ID == "" {
		t.Fatalf("task = %#v", task)
	}
	moved, err := provider.UpdateTaskStatus(UpdateTaskStatusRequest{Root: root, TaskID: task.ID, Status: "waiting"})
	if err != nil {
		t.Fatalf("UpdateTaskStatus returned error: %v", err)
	}
	if moved.Status != "waiting" {
		t.Fatalf("moved status = %q", moved.Status)
	}
	spec, err := provider.CreateSpec(CreateSpecRequest{Root: root, Title: "Plan Mode", Body: "Acceptance criteria", State: "drafted"})
	if err != nil {
		t.Fatalf("CreateSpec returned error: %v", err)
	}
	if spec.Path != "specs/plan-mode.md" {
		t.Fatalf("spec path = %q", spec.Path)
	}
	routine, err := provider.UpsertRoutine(UpsertRoutineRequest{Root: root, Name: "Nightly ideas", Prompt: "Find improvements", Enabled: true})
	if err != nil {
		t.Fatalf("UpsertRoutine returned error: %v", err)
	}
	if routine.ID == "" || !routine.Enabled {
		t.Fatalf("routine = %#v", routine)
	}
	automation, err := provider.SaveAutomation(SaveAutomationRequest{Root: root, Automation: AutomationInfo{AutoImplement: true, AutoTest: true}})
	if err != nil {
		t.Fatalf("SaveAutomation returned error: %v", err)
	}
	if !automation.AutoImplement || !automation.AutoTest {
		t.Fatalf("automation = %#v", automation)
	}

	workspace, err := provider.LoadWorkspace(root)
	if err != nil {
		t.Fatalf("LoadWorkspace returned error: %v", err)
	}
	if len(workspace.Tasks) != 1 || len(workspace.Specs) != 1 || len(workspace.Routines) != 1 {
		t.Fatalf("workspace primitives not loaded: tasks=%d specs=%d routines=%d", len(workspace.Tasks), len(workspace.Specs), len(workspace.Routines))
	}
	if !workspace.Automation.AutoImplement || !workspace.Automation.AutoTest {
		t.Fatalf("workspace automation = %#v", workspace.Automation)
	}
}

func TestWorkspaceRegistryLifecycle(t *testing.T) {
	configDir := t.TempDir()
	t.Setenv("THANOS_CONFIG_DIR", configDir)
	rootA := t.TempDir()
	rootB := t.TempDir()
	provider := NewRealProvider()
	provider.now = func() time.Time { return time.Date(2026, 7, 6, 9, 30, 0, 0, time.UTC) }

	created, err := provider.CreateWorkspace(CreateWorkspaceRequest{Name: "Primary", Path: rootA})
	if err != nil {
		t.Fatalf("CreateWorkspace returned error: %v", err)
	}
	if created.ID == "" || created.DataKey == "" || len(created.Folders) != 1 {
		t.Fatalf("created workspace = %#v", created)
	}
	originalDataKey := created.DataKey

	updated, err := provider.UpdateWorkspace(UpdateWorkspaceRequest{
		ID:   created.ID,
		Name: "Renamed",
		Folders: []WorkspaceFolderInfo{
			{ID: created.Folders[0].ID, Path: rootA, Label: "Repo A"},
			{Path: rootB, Label: "Repo B"},
		},
	})
	if err != nil {
		t.Fatalf("UpdateWorkspace returned error: %v", err)
	}
	if updated.Name != "Renamed" || updated.DataKey != originalDataKey || len(updated.Folders) != 2 {
		t.Fatalf("updated workspace = %#v", updated)
	}

	active, err := provider.ActivateWorkspace(ActivateWorkspaceRequest{ID: created.ID})
	if err != nil {
		t.Fatalf("ActivateWorkspace returned error: %v", err)
	}
	if active.ID != created.ID {
		t.Fatalf("active workspace = %#v", active)
	}

	loaded, err := provider.LoadActiveWorkspace()
	if err != nil {
		t.Fatalf("LoadActiveWorkspace returned error: %v", err)
	}
	if loaded.WorkspaceID != created.ID || loaded.DataKey != originalDataKey || len(loaded.Folders) != 2 {
		t.Fatalf("loaded active workspace = %#v", loaded)
	}

	registry, err := provider.DeleteWorkspace(DeleteWorkspaceRequest{ID: created.ID})
	if err != nil {
		t.Fatalf("DeleteWorkspace returned error: %v", err)
	}
	if registry.ActiveWorkspaceID != "" || len(registry.Workspaces) != 0 {
		t.Fatalf("registry after delete = %#v", registry)
	}
}

func TestNormalizeTaskStatus(t *testing.T) {
	cases := map[string]string{
		"in_progress":      "running",
		"waiting_approval": "waiting",
		"in_review":        "review",
		"complete":         "done",
		"cancelled":        "failed",
		"":                 "backlog",
	}
	for in, want := range cases {
		if got := normalizeTaskStatus(in); got != want {
			t.Fatalf("normalizeTaskStatus(%q) = %q, want %q", in, got, want)
		}
	}
}
