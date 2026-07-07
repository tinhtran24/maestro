package app

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestRealProviderLoadWorkspaceReadsRealSpecsAndTasks(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".thanos", "specs", "local"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".thanos", "specs", "local", "workbench.md"), []byte("# Workbench\n\nstate: validated\n"), 0o644); err != nil {
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

func TestPlanSpecModeParsesDispatchesAndUndoes(t *testing.T) {
	root := t.TempDir()
	provider := NewRealProvider()
	provider.now = func() time.Time { return time.Date(2026, 7, 8, 9, 0, 0, 0, time.UTC) }
	parent, err := provider.CreateSpec(CreateSpecRequest{Root: root, Title: "Checkout", Body: "Parent body", State: "validated"})
	if err != nil {
		t.Fatalf("CreateSpec parent returned error: %v", err)
	}
	child, err := provider.CreateSpec(CreateSpecRequest{Root: root, Title: "Payment Form", Body: "Build payment form", State: "validated", ParentPath: parent.Path})
	if err != nil {
		t.Fatalf("CreateSpec child returned error: %v", err)
	}
	if child.Path != ".thanos/specs/checkout/payment-form.md" {
		t.Fatalf("child path = %q", child.Path)
	}

	workspace, err := provider.LoadWorkspace(root)
	if err != nil {
		t.Fatalf("LoadWorkspace returned error: %v", err)
	}
	if len(workspace.Specs) != 1 || len(workspace.Specs[0].Children) != 1 || workspace.Specs[0].Children[0].State != "validated" {
		t.Fatalf("recursive specs = %#v", workspace.Specs)
	}

	provider.now = func() time.Time { return time.Date(2026, 7, 8, 9, 5, 0, 0, time.UTC) }
	updated, err := provider.UpdateSpec(UpdateSpecRequest{Root: root, Path: child.Path, Title: child.Title, Body: "Refined body", State: "testing"})
	if err != nil {
		t.Fatalf("UpdateSpec returned error: %v", err)
	}
	if updated.State != "testing" || updated.Body != "Refined body" {
		t.Fatalf("updated spec = %#v", updated)
	}
	dispatched, err := provider.DispatchSpecs(DispatchSpecsRequest{Root: root, Path: parent.Path})
	if err != nil {
		t.Fatalf("DispatchSpecs returned error: %v", err)
	}
	if len(dispatched) != 1 || dispatched[0].Title != "Payment Form" || !strings.Contains(dispatched[0].Prompt, "Refined body") {
		t.Fatalf("dispatched tasks = %#v", dispatched)
	}

	restored, err := provider.UndoPlanningChange(UndoPlanningChangeRequest{Root: root})
	if err != nil {
		t.Fatalf("UndoPlanningChange returned error: %v", err)
	}
	if restored.State != "validated" || !strings.Contains(restored.Body, "Build payment form") {
		t.Fatalf("restored spec = %#v", restored)
	}
}

func TestTaskTurnLoopPersistsOutputUsageAndResume(t *testing.T) {
	root := t.TempDir()
	provider := NewRealProvider()
	provider.now = func() time.Time { return time.Date(2026, 7, 7, 9, 0, 0, 0, time.UTC) }
	task, err := provider.CreateTask(CreateTaskRequest{Root: root, Title: "Run agent", Prompt: "Implement", Agent: "codex"})
	if err != nil {
		t.Fatalf("CreateTask returned error: %v", err)
	}
	started, err := provider.StartTaskTurn(StartTaskTurnRequest{
		Root:           root,
		TaskID:         task.ID,
		Step:           "Implementation",
		ProviderID:     "codex",
		SessionID:      "native-1",
		TranscriptPath: ".thanos/terminal/sessions/native-1.log",
	})
	if err != nil {
		t.Fatalf("StartTaskTurn returned error: %v", err)
	}
	if started.Status != "in_progress" || len(started.Turns) != 1 || started.LastTurn == nil || started.LastTurn.Status != "running" {
		t.Fatalf("started task = %#v", started)
	}

	provider.now = func() time.Time { return time.Date(2026, 7, 7, 9, 5, 0, 0, time.UTC) }
	finished, err := provider.FinishTaskTurn(FinishTaskTurnRequest{
		Root:       root,
		TaskID:     task.ID,
		TurnID:     "turn-001",
		Status:     "completed",
		Stdout:     "implemented changes\nSTOP_REASON=waiting_user\n",
		Stderr:     "",
		StopReason: "waiting_user",
		UsageUSD:   1.25,
	})
	if err != nil {
		t.Fatalf("FinishTaskTurn returned error: %v", err)
	}
	if finished.Status != "waiting" || finished.UsageUSD != 1.25 || finished.LastTurn.StopReason != "waiting_user" || !strings.Contains(finished.LastOutput, "implemented changes") {
		t.Fatalf("finished task = %#v", finished)
	}
	stdoutData, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(finished.LastTurn.StdoutPath)))
	if err != nil {
		t.Fatalf("stdout log missing: %v", err)
	}
	if !strings.Contains(string(stdoutData), "implemented changes") {
		t.Fatalf("stdout log = %q", string(stdoutData))
	}

	provider.now = func() time.Time { return time.Date(2026, 7, 7, 9, 10, 0, 0, time.UTC) }
	resumed, err := provider.ResumeTaskTurn(ResumeTaskTurnRequest{Root: root, TaskID: task.ID, Feedback: "Continue with tests"})
	if err != nil {
		t.Fatalf("ResumeTaskTurn returned error: %v", err)
	}
	if resumed.Status != "in_progress" || len(resumed.FeedbackHistory) != 1 {
		t.Fatalf("resumed task = %#v", resumed)
	}
}

func TestOversightArtifactsGeneratedAndReloadable(t *testing.T) {
	root := t.TempDir()
	provider := NewRealProvider()
	provider.now = func() time.Time { return time.Date(2026, 7, 7, 12, 0, 0, 0, time.UTC) }
	task, err := provider.CreateTask(CreateTaskRequest{Root: root, Title: "Review run", Prompt: "Implement and review", Agent: "codex"})
	if err != nil {
		t.Fatalf("CreateTask returned error: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(root, filepath.FromSlash(task.Worktree)), 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := provider.StartTaskTurn(StartTaskTurnRequest{Root: root, TaskID: task.ID, Step: "Implementation", ProviderID: "codex"}); err != nil {
		t.Fatalf("StartTaskTurn returned error: %v", err)
	}

	provider.now = func() time.Time { return time.Date(2026, 7, 7, 12, 5, 0, 0, time.UTC) }
	finished, err := provider.FinishTaskTurn(FinishTaskTurnRequest{
		Root:     root,
		TaskID:   task.ID,
		TurnID:   "turn-001",
		Status:   "completed",
		Stdout:   "implemented changes",
		UsageUSD: 0.75,
	})
	if err != nil {
		t.Fatalf("FinishTaskTurn returned error: %v", err)
	}
	if finished.Oversight == nil || finished.Oversight.Path == "" || !strings.Contains(finished.Oversight.Summary, "Review run") {
		t.Fatalf("missing oversight after turn finish: %#v", finished.Oversight)
	}
	if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(finished.Oversight.Path))); err != nil {
		t.Fatalf("oversight artifact missing: %v", err)
	}

	provider.now = func() time.Time { return time.Date(2026, 7, 7, 12, 10, 0, 0, time.UTC) }
	verified, err := provider.RunTaskVerification(context.Background(), RunTaskVerificationRequest{Root: root, TaskID: task.ID, Command: "printf PASS"})
	if err != nil {
		t.Fatalf("RunTaskVerification returned error: %v", err)
	}
	if verified.Oversight == nil || verified.Oversight.TestPath == "" || verified.Oversight.TestResult != "passed" {
		t.Fatalf("missing test oversight: %#v", verified.Oversight)
	}
	if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(verified.Oversight.TestPath))); err != nil {
		t.Fatalf("oversight test artifact missing: %v", err)
	}

	loaded, err := provider.LoadWorkspace(root)
	if err != nil {
		t.Fatalf("LoadWorkspace returned error: %v", err)
	}
	if len(loaded.Tasks) != 1 || loaded.Tasks[0].Oversight == nil || loaded.Tasks[0].Oversight.TestResult != "passed" {
		t.Fatalf("loaded oversight = %#v", loaded.Tasks)
	}
}

func TestTaskTurnLoopClassifiesFailuresAndAutoContinue(t *testing.T) {
	root := t.TempDir()
	provider := NewRealProvider()
	provider.now = func() time.Time { return time.Date(2026, 7, 7, 10, 0, 0, 0, time.UTC) }
	task, err := provider.CreateTask(CreateTaskRequest{Root: root, Title: "Run with limits", Prompt: "Implement", Agent: "codex"})
	if err != nil {
		t.Fatalf("CreateTask returned error: %v", err)
	}
	if _, err := provider.StartTaskTurn(StartTaskTurnRequest{Root: root, TaskID: task.ID, ProviderID: "codex"}); err != nil {
		t.Fatalf("StartTaskTurn returned error: %v", err)
	}
	continued, err := provider.FinishTaskTurn(FinishTaskTurnRequest{
		Root:         root,
		TaskID:       task.ID,
		TurnID:       "turn-001",
		Status:       "completed",
		Stdout:       "hit max token limit",
		AutoContinue: true,
	})
	if err != nil {
		t.Fatalf("FinishTaskTurn auto-continue returned error: %v", err)
	}
	if continued.Status != "in_progress" || continued.LastTurn.StopReason != "max_tokens" || !continued.LastTurn.AutoContinue {
		t.Fatalf("auto-continued task = %#v", continued)
	}
	if _, err := provider.StartTaskTurn(StartTaskTurnRequest{Root: root, TaskID: task.ID, ProviderID: "codex"}); err != nil {
		t.Fatalf("second StartTaskTurn returned error: %v", err)
	}
	failed, err := provider.FinishTaskTurn(FinishTaskTurnRequest{
		Root:     root,
		TaskID:   task.ID,
		TurnID:   "turn-002",
		Status:   "failed",
		Stderr:   "permission denied opening file",
		ExitCode: 1,
	})
	if err != nil {
		t.Fatalf("FinishTaskTurn failure returned error: %v", err)
	}
	if failed.Status != "failed" || failed.FailureCategory != "permissions" || failed.LastTurn.FailureCategory != "permissions" {
		t.Fatalf("failed task = %#v", failed)
	}
}

func TestTaskVerificationStoresResultAndGatesDone(t *testing.T) {
	root := t.TempDir()
	provider := NewRealProvider()
	provider.now = func() time.Time { return time.Date(2026, 7, 7, 12, 0, 0, 0, time.UTC) }
	task, err := provider.CreateTask(CreateTaskRequest{Root: root, Title: "Verify task", Prompt: "Run tests", Agent: "codex"})
	if err != nil {
		t.Fatalf("CreateTask returned error: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(root, filepath.FromSlash(task.Worktree)), 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := provider.UpdateTaskStatus(UpdateTaskStatusRequest{Root: root, TaskID: task.ID, Status: "in_progress"}); err != nil {
		t.Fatalf("move in_progress: %v", err)
	}
	if _, err := provider.UpdateTaskStatus(UpdateTaskStatusRequest{Root: root, TaskID: task.ID, Status: "waiting"}); err != nil {
		t.Fatalf("move waiting: %v", err)
	}
	if _, err := provider.SaveAutomation(SaveAutomationRequest{Root: root, Automation: AutomationInfo{AutoTest: true}}); err != nil {
		t.Fatalf("SaveAutomation returned error: %v", err)
	}
	if _, err := provider.UpdateTaskStatus(UpdateTaskStatusRequest{Root: root, TaskID: task.ID, Status: "committing"}); err != nil {
		t.Fatalf("move committing before failed verification: %v", err)
	}
	if _, err := provider.UpdateTaskStatus(UpdateTaskStatusRequest{Root: root, TaskID: task.ID, Status: "done"}); err == nil || !strings.Contains(err.Error(), "passing verification") {
		t.Fatalf("expected done gate before verification, got %v", err)
	}
	if _, err := provider.UpdateTaskStatus(UpdateTaskStatusRequest{Root: root, TaskID: task.ID, Status: "failed", FailureCategory: "reset"}); err != nil {
		t.Fatalf("move failed for retry: %v", err)
	}
	if _, err := provider.UpdateTaskStatus(UpdateTaskStatusRequest{Root: root, TaskID: task.ID, Status: "in_progress"}); err != nil {
		t.Fatalf("retry in_progress: %v", err)
	}
	if _, err := provider.UpdateTaskStatus(UpdateTaskStatusRequest{Root: root, TaskID: task.ID, Status: "waiting"}); err != nil {
		t.Fatalf("retry waiting: %v", err)
	}

	failed, err := provider.RunTaskVerification(context.Background(), RunTaskVerificationRequest{
		Root:        root,
		TaskID:      task.ID,
		Command:     "printf 'FAIL unit test'",
		FailPattern: "FAIL",
	})
	if err != nil {
		t.Fatalf("RunTaskVerification failed run returned error: %v", err)
	}
	if failed.TestsPassed || failed.LastTestResult == nil || failed.LastTestResult.Status != "failed" || failed.Status != "waiting" {
		t.Fatalf("failed verification task = %#v", failed)
	}
	if _, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(failed.LastTestResult.OutputPath))); err != nil {
		t.Fatalf("test output was not persisted separately: %v", err)
	}

	passed, err := provider.RunTaskVerification(context.Background(), RunTaskVerificationRequest{
		Root:        root,
		TaskID:      task.ID,
		Command:     "printf 'PASS unit test'",
		PassPattern: "PASS",
	})
	if err != nil {
		t.Fatalf("RunTaskVerification pass returned error: %v", err)
	}
	if !passed.TestsPassed || passed.LastTestResult == nil || passed.LastTestResult.Status != "passed" || passed.FailureCategory != "" {
		t.Fatalf("passed verification task = %#v", passed)
	}
	if _, err := provider.UpdateTaskStatus(UpdateTaskStatusRequest{Root: root, TaskID: task.ID, Status: "committing"}); err != nil {
		t.Fatalf("move committing after pass: %v", err)
	}
	if _, err := provider.UpdateTaskStatus(UpdateTaskStatusRequest{Root: root, TaskID: task.ID, Status: "done"}); err == nil || !strings.Contains(err.Error(), "committing task worktree") {
		t.Fatalf("expected commit gate after passing verification, got %v", err)
	}
}

func TestCommitPipelineRequiresApprovalAndRecordsCommit(t *testing.T) {
	root := t.TempDir()
	provider := NewRealProvider()
	provider.now = func() time.Time { return time.Date(2026, 7, 7, 13, 0, 0, 0, time.UTC) }
	task, err := provider.CreateTask(CreateTaskRequest{Root: root, Title: "Commit task", Prompt: "Change file", Agent: "codex"})
	if err != nil {
		t.Fatalf("CreateTask returned error: %v", err)
	}
	worktree := filepath.Join(root, filepath.FromSlash(task.Worktree))
	initGitRepo(t, worktree)
	if err := os.WriteFile(filepath.Join(worktree, "app.txt"), []byte("after\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(worktree, "new.txt"), []byte("new\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, status := range []string{"in_progress", "waiting", "committing"} {
		if _, err := provider.UpdateTaskStatus(UpdateTaskStatusRequest{Root: root, TaskID: task.ID, Status: status}); err != nil {
			t.Fatalf("move %s: %v", status, err)
		}
	}
	if _, err := provider.UpdateTaskStatus(UpdateTaskStatusRequest{Root: root, TaskID: task.ID, Status: "done"}); err == nil || !strings.Contains(err.Error(), "committing task worktree") {
		t.Fatalf("expected done gate before commit, got %v", err)
	}

	preview, err := provider.PrepareTaskCommit(context.Background(), PrepareTaskCommitRequest{
		Root:    root,
		TaskID:  task.ID,
		Message: "task: commit worktree output",
	})
	if err != nil {
		t.Fatalf("PrepareTaskCommit returned error: %v", err)
	}
	if preview.Commit == nil || preview.Commit.Approved || preview.Commit.Committed || !strings.Contains(preview.Commit.Diff, "after") || !strings.Contains(preview.Commit.DiffStat, "new.txt") {
		t.Fatalf("preview commit = %#v", preview.Commit)
	}
	if _, err := provider.CommitTaskChanges(context.Background(), CommitTaskChangesRequest{Root: root, TaskID: task.ID, Message: preview.Commit.Message}); err == nil || !strings.Contains(err.Error(), "approval") {
		t.Fatalf("expected approval error, got %v", err)
	}

	committed, err := provider.CommitTaskChanges(context.Background(), CommitTaskChangesRequest{
		Root:     root,
		TaskID:   task.ID,
		Message:  preview.Commit.Message,
		Approved: true,
	})
	if err != nil {
		t.Fatalf("CommitTaskChanges returned error: %v", err)
	}
	if committed.Commit == nil || !committed.Commit.Committed || !committed.Commit.Approved || committed.Commit.Hash == "" || committed.Commit.Summary != "task: commit worktree output" {
		t.Fatalf("committed task = %#v", committed.Commit)
	}
	done, err := provider.UpdateTaskStatus(UpdateTaskStatusRequest{Root: root, TaskID: task.ID, Status: "done"})
	if err != nil {
		t.Fatalf("done after commit returned error: %v", err)
	}
	if done.Status != "done" || done.Commit == nil || done.Commit.Hash == "" {
		t.Fatalf("done task = %#v", done)
	}
}

func initGitRepo(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	runGit(t, dir, "init")
	runGit(t, dir, "config", "user.email", "test@example.com")
	runGit(t, dir, "config", "user.name", "Test User")
	if err := os.WriteFile(filepath.Join(dir, "app.txt"), []byte("before\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, dir, "add", "app.txt")
	runGit(t, dir, "commit", "-m", "initial")
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s failed: %v\n%s", strings.Join(args, " "), err, string(output))
	}
}

func markTaskCommitted(t *testing.T, root, taskID string) {
	t.Helper()
	path := filepath.Join(root, ".thanos", "tasks", taskID+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	raw["commit"] = map[string]any{
		"hash":      "abc123",
		"summary":   "task: test commit",
		"message":   "task: test commit",
		"approved":  true,
		"committed": true,
	}
	if err := writeJSON(path, raw); err != nil {
		t.Fatal(err)
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

	task, err := provider.CreateTask(CreateTaskRequest{Root: root, Title: "Build board", Prompt: "Create a local task", Flow: "implement", Agent: "codex", Status: "waiting"})
	if err != nil {
		t.Fatalf("CreateTask returned error: %v", err)
	}
	if task.Status != "waiting" || task.ID == "" {
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
	if spec.Path != ".thanos/specs/plan-mode.md" {
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

func TestRoutineTriggerSchedulerAndCircuitBreaker(t *testing.T) {
	root := t.TempDir()
	provider := NewRealProvider()
	provider.now = func() time.Time { return time.Date(2026, 7, 8, 10, 0, 0, 0, time.UTC) }

	routine, err := provider.UpsertRoutine(UpsertRoutineRequest{Root: root, Name: "Morning Review", Prompt: "Find stale specs", Schedule: "now", Flow: "implement", Enabled: true})
	if err != nil {
		t.Fatalf("UpsertRoutine returned error: %v", err)
	}
	task, err := provider.TriggerRoutine(TriggerRoutineRequest{Root: root, RoutineID: routine.ID})
	if err != nil {
		t.Fatalf("TriggerRoutine returned error: %v", err)
	}
	if task.Title != "Morning Review" || !strings.Contains(task.Prompt, "Routine: "+routine.ID) {
		t.Fatalf("triggered task = %#v", task)
	}
	loaded, err := provider.LoadWorkspace(root)
	if err != nil {
		t.Fatalf("LoadWorkspace returned error: %v", err)
	}
	if len(loaded.Routines) != 1 || loaded.Routines[0].RunCount != 1 || loaded.Routines[0].LastRunAt == "" {
		t.Fatalf("routine run state = %#v", loaded.Routines)
	}

	provider.now = func() time.Time { return time.Date(2026, 7, 8, 10, 5, 0, 0, time.UTC) }
	due, err := provider.UpsertRoutine(UpsertRoutineRequest{Root: root, Name: "Due Routine", Prompt: "Spawn from scheduler", Schedule: "now", Flow: "implement", Enabled: true})
	if err != nil {
		t.Fatalf("UpsertRoutine due returned error: %v", err)
	}
	scheduled, err := provider.RunRoutineScheduler(RunRoutineSchedulerRequest{Root: root})
	if err != nil {
		t.Fatalf("RunRoutineScheduler returned error: %v", err)
	}
	if len(scheduled) != 1 || scheduled[0].Title != due.Name {
		t.Fatalf("scheduled tasks = %#v", scheduled)
	}

	if _, err := provider.SaveAutomation(SaveAutomationRequest{Root: root, Automation: AutomationInfo{MaxConcurrentRoutineTasks: 1, CircuitBreakerFailureLimit: 2}}); err != nil {
		t.Fatalf("SaveAutomation returned error: %v", err)
	}
	provider.now = func() time.Time { return time.Date(2026, 7, 8, 10, 10, 0, 0, time.UTC) }
	blocked, err := provider.UpsertRoutine(UpsertRoutineRequest{Root: root, Name: "Blocked Routine", Prompt: "Should trip", Schedule: "now", Flow: "implement", Enabled: true})
	if err != nil {
		t.Fatalf("UpsertRoutine blocked returned error: %v", err)
	}
	if _, err := provider.RunRoutineScheduler(RunRoutineSchedulerRequest{Root: root}); err != nil {
		t.Fatalf("RunRoutineScheduler first blocked returned error: %v", err)
	}
	if _, err := provider.RunRoutineScheduler(RunRoutineSchedulerRequest{Root: root}); err != nil {
		t.Fatalf("RunRoutineScheduler second blocked returned error: %v", err)
	}
	loaded, err = provider.LoadWorkspace(root)
	if err != nil {
		t.Fatalf("LoadWorkspace after circuit breaker returned error: %v", err)
	}
	for _, item := range loaded.Routines {
		if item.ID == blocked.ID && (item.Enabled || !strings.Contains(item.DisabledReason, "circuit breaker")) {
			t.Fatalf("circuit breaker routine = %#v", item)
		}
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
	for _, status := range []string{"in_progress", "committing"} {
		if _, err := provider.UpdateTaskStatus(UpdateTaskStatusRequest{Root: root, TaskID: parent.ID, Status: status}); err != nil {
			t.Fatalf("parent transition to %s returned error: %v", status, err)
		}
	}
	markTaskCommitted(t, root, parent.ID)
	if _, err := provider.UpdateTaskStatus(UpdateTaskStatusRequest{Root: root, TaskID: parent.ID, Status: "done"}); err != nil {
		t.Fatalf("parent transition to done returned error: %v", err)
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
