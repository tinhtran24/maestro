package mcp

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/tinhtran/thanos/internal/model"
	"github.com/tinhtran/thanos/internal/workspace"
)

func TestToolsExposePhaseSixBridgeCapabilities(t *testing.T) {
	tools := Tools()
	names := make(map[string]bool)
	for _, tool := range tools {
		names[tool.Name] = true
	}
	for _, name := range []string{
		"thanos.task.create_subtask",
		"thanos.task.message_sibling",
		"thanos.memory.inspect_related_work",
		"thanos.task.attach_branch",
		"thanos.review.request_user_review",
	} {
		if !names[name] {
			t.Fatalf("missing tool %s", name)
		}
	}
}

func TestCreateSubtaskPersistsChildAndUpdatesParent(t *testing.T) {
	bridge, ws := testBridge(t)
	parent, err := ws.NewTask("Parent task", "parent", "P1", "planner", "")
	if err != nil {
		t.Fatal(err)
	}
	if err := ws.SaveTask(parent); err != nil {
		t.Fatal(err)
	}

	child, err := bridge.CreateSubtask(CreateSubtaskRequest{ParentTaskID: parent.ID, Title: "Child task", Description: "child"})
	if err != nil {
		t.Fatal(err)
	}
	if child.ParentTaskID != parent.ID {
		t.Fatalf("child parent = %q, want %q", child.ParentTaskID, parent.ID)
	}
	got, err := ws.LoadTask(parent.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Subtasks) != 1 || got.Subtasks[0] != child.ID {
		t.Fatalf("parent subtasks = %#v", got.Subtasks)
	}
}

func TestMessageSiblingRejectsUnrelatedTasks(t *testing.T) {
	bridge, ws := testBridge(t)
	a, _ := ws.NewTask("A", "", "P2", "", "")
	b, _ := ws.NewTask("B", "", "P2", "", "")
	if err := ws.SaveTask(a); err != nil {
		t.Fatal(err)
	}
	if err := ws.SaveTask(b); err != nil {
		t.Fatal(err)
	}
	_, err := bridge.MessageSibling(TaskMessageRequest{SourceTaskID: a.ID, TargetTaskID: b.ID, Content: "hello"})
	if err == nil || !strings.Contains(err.Error(), "not siblings") {
		t.Fatalf("expected sibling error, got %v", err)
	}
}

func TestAttachBranchRefusesProtectedBranches(t *testing.T) {
	bridge, ws := testBridge(t)
	task, _ := ws.NewTask("Task", "", "P2", "", "")
	if err := ws.SaveTask(task); err != nil {
		t.Fatal(err)
	}
	_, err := bridge.AttachBranch(AttachBranchRequest{TaskID: task.ID, BranchName: "main"})
	if err == nil || !strings.Contains(err.Error(), "protected branch") {
		t.Fatalf("expected protected branch error, got %v", err)
	}
}

func TestInspectRelatedWorkReadsTasksAndMemory(t *testing.T) {
	bridge, ws := testBridge(t)
	task, _ := ws.NewTask("Cart sync", "Persist cart sessions", "P2", "", "")
	if err := ws.SaveTask(task); err != nil {
		t.Fatal(err)
	}
	memoryDir := filepath.Join(ws.DotDir(), "memory")
	if err := os.MkdirAll(memoryDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(memoryDir, "feature-graph.md"), []byte("# Cart memory\nCart sync uses local-first storage."), 0o644); err != nil {
		t.Fatal(err)
	}
	related, err := bridge.InspectRelatedWork(RelatedWorkRequest{Query: "cart", Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(related) < 2 {
		t.Fatalf("related = %#v", related)
	}
}

func TestRequestUserReviewWritesPendingReviewRequest(t *testing.T) {
	bridge, ws := testBridge(t)
	task, _ := ws.NewTask("Review me", "", "P2", "", "")
	if err := ws.SaveTask(task); err != nil {
		t.Fatal(err)
	}
	request, err := bridge.RequestUserReview(ReviewRequest{TaskID: task.ID, RequestedBy: "coder", Notes: "Ready for review"})
	if err != nil {
		t.Fatal(err)
	}
	if request.Status != "pending_user_review" {
		t.Fatalf("status = %s", request.Status)
	}
	if _, err := os.Stat(filepath.Join(ws.DotDir(), "review-requests", request.ID+".json")); err != nil {
		t.Fatal(err)
	}
}

func testBridge(t *testing.T) (Bridge, *workspace.Workspace) {
	t.Helper()
	ws := workspace.Open(t.TempDir())
	if err := ws.Init(model.Config{Project: model.Project{Name: "test"}}); err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 7, 3, 2, 10, 0, 0, time.UTC)
	return Bridge{Workspace: ws, Now: func() time.Time { return now }}, ws
}
