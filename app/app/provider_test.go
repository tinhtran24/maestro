package app

import (
	"os"
	"path/filepath"
	"reflect"
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

func TestProviderCatalogContainsMilestoneFiveProviders(t *testing.T) {
	catalog := providerCatalog()
	got := make([]string, 0, len(catalog))
	for _, provider := range catalog {
		got = append(got, provider.ID)
		if provider.ID != "shell" && (!provider.SupportsRun || provider.SetupHint == "") {
			t.Fatalf("provider missing run metadata: %#v", provider)
		}
	}
	want := []string{"claude-code", "codex", "gemini-cli", "opencode", "cursor-agent", "aider", "goose", "shell"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("catalog IDs = %#v, want %#v", got, want)
	}
}

func TestLoadWorkspaceMergesUserAgentsAndFlowsFromDisk(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".thanos", "agents"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, ".thanos", "flows"), 0o755); err != nil {
		t.Fatal(err)
	}
	agentJSON := `{"id":"security-review","role":"Security Review","harness":"Codex","model":"gpt-5","capabilities":["diff.read","risk.review"]}`
	if err := os.WriteFile(filepath.Join(root, ".thanos", "agents", "security-review.json"), []byte(agentJSON), 0o644); err != nil {
		t.Fatal(err)
	}
	flowJSON := `{"id":"security-pass","name":"Security Pass","steps":["Implementation","Security Review","Testing"],"parallelGroups":[["Security Review","Testing"]]}`
	if err := os.WriteFile(filepath.Join(root, ".thanos", "flows", "security-pass.json"), []byte(flowJSON), 0o644); err != nil {
		t.Fatal(err)
	}

	provider := NewRealProvider()
	workspace, err := provider.LoadWorkspace(root)
	if err != nil {
		t.Fatalf("LoadWorkspace returned error: %v", err)
	}
	if !hasAgent(workspace.Agents, "impl", true) {
		t.Fatalf("built-in agent missing or writable: %#v", workspace.Agents)
	}
	if !hasAgent(workspace.Agents, "security-review", false) {
		t.Fatalf("user agent missing: %#v", workspace.Agents)
	}
	if !hasFlow(workspace.Flows, "implement", true) {
		t.Fatalf("built-in flow missing or writable: %#v", workspace.Flows)
	}
	if !hasFlow(workspace.Flows, "security-pass", false) {
		t.Fatalf("user flow missing: %#v", workspace.Flows)
	}
	if !hasParallelGroup(workspace.Flows, "security-pass", []string{"Security Review", "Testing"}) {
		t.Fatalf("user flow parallel group missing: %#v", workspace.Flows)
	}

	updatedFlowJSON := `{"id":"security-pass","name":"Security Pass","steps":["Security Review","Oversight"]}`
	if err := os.WriteFile(filepath.Join(root, ".thanos", "flows", "security-pass.json"), []byte(updatedFlowJSON), 0o644); err != nil {
		t.Fatal(err)
	}
	reloaded, err := provider.LoadWorkspace(root)
	if err != nil {
		t.Fatalf("LoadWorkspace after flow edit returned error: %v", err)
	}
	for _, flow := range reloaded.Flows {
		if flow.ID == "security-pass" && !reflect.DeepEqual(flow.Steps, []string{"Security Review", "Oversight"}) {
			t.Fatalf("flow did not hot-reload from disk: %#v", flow)
		}
	}
}

func TestTaskFlowFallsBackToImplementWhenMissing(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".thanos", "tasks"), 0o755); err != nil {
		t.Fatal(err)
	}
	taskJSON := `{"id":"T-unknown-flow","title":"Unknown flow","prompt":"Run","status":"backlog","flow":"deleted-flow"}`
	if err := os.WriteFile(filepath.Join(root, ".thanos", "tasks", "unknown.json"), []byte(taskJSON), 0o644); err != nil {
		t.Fatal(err)
	}
	provider := NewRealProvider()
	workspace, err := provider.LoadWorkspace(root)
	if err != nil {
		t.Fatalf("LoadWorkspace returned error: %v", err)
	}
	if len(workspace.Tasks) != 1 || workspace.Tasks[0].Flow != "implement" {
		t.Fatalf("task flow fallback = %#v", workspace.Tasks)
	}
	created, err := provider.CreateTask(CreateTaskRequest{Root: root, Title: "Create with missing flow", Prompt: "Run", Flow: "missing-flow"})
	if err != nil {
		t.Fatalf("CreateTask returned error: %v", err)
	}
	if created.Flow != "implement" {
		t.Fatalf("created flow = %q", created.Flow)
	}
	if err := os.MkdirAll(filepath.Join(root, ".thanos", "flows"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".thanos", "flows", "custom.json"), []byte(`{"id":"custom-flow","name":"Custom Flow","steps":["Implementation"]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	custom, err := provider.CreateTask(CreateTaskRequest{Root: root, Title: "Create with custom flow", Prompt: "Run", Flow: "custom-flow"})
	if err != nil {
		t.Fatalf("CreateTask with custom flow returned error: %v", err)
	}
	if custom.Flow != "custom-flow" {
		t.Fatalf("custom flow = %q", custom.Flow)
	}
}

func hasAgent(agents []AgentRoleInfo, id string, readOnly bool) bool {
	for _, agent := range agents {
		if agent.ID == id && agent.ReadOnly == readOnly {
			return true
		}
	}
	return false
}

func hasFlow(flows []FlowInfo, id string, readOnly bool) bool {
	for _, flow := range flows {
		if flow.ID == id && flow.ReadOnly == readOnly {
			return true
		}
	}
	return false
}

func hasParallelGroup(flows []FlowInfo, id string, group []string) bool {
	for _, flow := range flows {
		if flow.ID != id {
			continue
		}
		for _, candidate := range flow.ParallelGroups {
			if reflect.DeepEqual(candidate, group) {
				return true
			}
		}
	}
	return false
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
	moved, err := provider.UpdateTaskStatus(UpdateTaskStatusRequest{Root: root, TaskID: task.ID, Status: "in_progress"})
	if err != nil {
		t.Fatalf("UpdateTaskStatus to in_progress returned error: %v", err)
	}
	moved, err = provider.UpdateTaskStatus(UpdateTaskStatusRequest{Root: root, TaskID: task.ID, Status: "waiting", Feedback: "needs input"})
	if err != nil {
		t.Fatalf("UpdateTaskStatus returned error: %v", err)
	}
	if moved.Status != "waiting" {
		t.Fatalf("moved status = %q", moved.Status)
	}
	if len(moved.FeedbackHistory) != 1 {
		t.Fatalf("feedback history = %#v", moved.FeedbackHistory)
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

func TestTaskLifecycleRejectsInvalidTransitionsAndBlocksDependencies(t *testing.T) {
	root := t.TempDir()
	provider := NewRealProvider()
	provider.now = func() time.Time { return time.Date(2026, 7, 6, 10, 0, 0, 0, time.UTC) }

	parent, err := provider.CreateTask(CreateTaskRequest{Root: root, Title: "Parent", Prompt: "first"})
	if err != nil {
		t.Fatal(err)
	}
	child, err := provider.CreateTask(CreateTaskRequest{Root: root, Title: "Child", Prompt: "second", Dependencies: []string{parent.ID}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := provider.UpdateTaskStatus(UpdateTaskStatusRequest{Root: root, TaskID: child.ID, Status: "in_progress"}); err == nil {
		t.Fatal("expected blocked dependency to reject in_progress transition")
	}
	if _, err := provider.UpdateTaskStatus(UpdateTaskStatusRequest{Root: root, TaskID: parent.ID, Status: "done"}); err == nil {
		t.Fatal("expected backlog -> done to be rejected")
	}
	for _, status := range []string{"in_progress", "committing", "done"} {
		if _, err := provider.UpdateTaskStatus(UpdateTaskStatusRequest{Root: root, TaskID: parent.ID, Status: status}); err != nil {
			t.Fatalf("parent transition to %s returned error: %v", status, err)
		}
	}
	updatedChild, err := provider.UpdateTaskStatus(UpdateTaskStatusRequest{Root: root, TaskID: child.ID, Status: "in_progress"})
	if err != nil {
		t.Fatalf("child transition after dependency done returned error: %v", err)
	}
	if updatedChild.Blocked {
		t.Fatalf("updated child should not be blocked: %#v", updatedChild)
	}
}

func TestBatchSearchAndTaskFlags(t *testing.T) {
	root := t.TempDir()
	provider := NewRealProvider()
	provider.now = func() time.Time { return time.Date(2026, 7, 6, 11, 0, 0, 0, time.UTC) }

	created, err := provider.BatchCreateTasks(BatchCreateTasksRequest{
		Root: root,
		Tasks: []CreateTaskRequest{
			{Title: "Build parser", Prompt: "parse task records"},
			{Title: "Render board", Prompt: "show lifecycle"},
		},
	})
	if err != nil {
		t.Fatalf("BatchCreateTasks returned error: %v", err)
	}
	if len(created) != 2 || len(created[1].Dependencies) != 1 || created[1].Dependencies[0] != created[0].ID {
		t.Fatalf("created batch = %#v", created)
	}
	if _, err := provider.UpdateTaskFlags(UpdateTaskFlagsRequest{Root: root, TaskID: created[0].ID, Archived: true}); err != nil {
		t.Fatalf("UpdateTaskFlags archive returned error: %v", err)
	}
	results, err := provider.SearchTasks(SearchTasksRequest{Root: root, Query: "parser"})
	if err != nil {
		t.Fatalf("SearchTasks returned error: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("archived task should be hidden by default: %#v", results)
	}
	results, err = provider.SearchTasks(SearchTasksRequest{Root: root, Query: "parser", IncludeArchived: true})
	if err != nil {
		t.Fatalf("SearchTasks with archived returned error: %v", err)
	}
	if len(results) != 1 || !results[0].Archived {
		t.Fatalf("archived search results = %#v", results)
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
		"in_progress":      "in_progress",
		"running":          "in_progress",
		"waiting_approval": "waiting",
		"commit":           "committing",
		"complete":         "done",
		"cancelled":        "cancelled",
		"":                 "backlog",
	}
	for in, want := range cases {
		if got := normalizeTaskStatus(in); got != want {
			t.Fatalf("normalizeTaskStatus(%q) = %q, want %q", in, got, want)
		}
	}
}
