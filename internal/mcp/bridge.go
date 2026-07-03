package mcp

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/tinhtran/thanos/internal/model"
	"github.com/tinhtran/thanos/internal/workspace"
)

type ToolDescriptor struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Inputs      []string `json:"inputs"`
}

type Bridge struct {
	Workspace *workspace.Workspace
	Now       func() time.Time
}

type CreateSubtaskRequest struct {
	ParentTaskID string `json:"parent_task_id"`
	Title        string `json:"title"`
	Description  string `json:"description,omitempty"`
	Priority     string `json:"priority,omitempty"`
	Agent        string `json:"agent,omitempty"`
}

type TaskMessageRequest struct {
	SourceTaskID string `json:"source_task_id"`
	TargetTaskID string `json:"target_task_id"`
	Content      string `json:"content"`
}

type TaskMessage struct {
	ID           string    `json:"id"`
	SourceTaskID string    `json:"source_task_id"`
	TargetTaskID string    `json:"target_task_id"`
	Content      string    `json:"content"`
	CreatedAt    time.Time `json:"created_at"`
}

type RelatedWorkRequest struct {
	TaskID string `json:"task_id,omitempty"`
	Query  string `json:"query"`
	Limit  int    `json:"limit,omitempty"`
}

type RelatedWork struct {
	Type    string `json:"type"`
	ID      string `json:"id"`
	Title   string `json:"title"`
	Summary string `json:"summary"`
	Path    string `json:"path,omitempty"`
}

type AttachBranchRequest struct {
	TaskID       string `json:"task_id"`
	BranchName   string `json:"branch_name"`
	WorktreePath string `json:"worktree_path,omitempty"`
}

type BranchAttachment struct {
	TaskID       string    `json:"task_id"`
	BranchName   string    `json:"branch_name"`
	WorktreePath string    `json:"worktree_path"`
	AttachedAt   time.Time `json:"attached_at"`
}

type ReviewRequest struct {
	TaskID      string `json:"task_id"`
	RequestedBy string `json:"requested_by,omitempty"`
	Notes       string `json:"notes,omitempty"`
}

type UserReviewRequest struct {
	ID          string    `json:"id"`
	TaskID      string    `json:"task_id"`
	RequestedBy string    `json:"requested_by,omitempty"`
	Notes       string    `json:"notes,omitempty"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

func Tools() []ToolDescriptor {
	return []ToolDescriptor{
		{Name: "thanos.task.create_subtask", Description: "Create a child task under an existing task.", Inputs: []string{"parent_task_id", "title", "description", "priority", "agent"}},
		{Name: "thanos.task.message_sibling", Description: "Send a durable message from one task to a sibling task.", Inputs: []string{"source_task_id", "target_task_id", "content"}},
		{Name: "thanos.memory.inspect_related_work", Description: "Inspect related tasks and feature-memory notes.", Inputs: []string{"task_id", "query", "limit"}},
		{Name: "thanos.task.attach_branch", Description: "Attach branch and worktree metadata to a task.", Inputs: []string{"task_id", "branch_name", "worktree_path"}},
		{Name: "thanos.review.request_user_review", Description: "Request human review without approving or merging.", Inputs: []string{"task_id", "requested_by", "notes"}},
	}
}

func (b Bridge) CreateSubtask(request CreateSubtaskRequest) (model.Task, error) {
	ws, err := b.workspace()
	if err != nil {
		return model.Task{}, err
	}
	parentID, err := sanitizeID(request.ParentTaskID)
	if err != nil {
		return model.Task{}, fmt.Errorf("parent task id: %w", err)
	}
	parent, err := ws.LoadTask(parentID)
	if err != nil {
		return model.Task{}, err
	}
	title := strings.TrimSpace(request.Title)
	if title == "" {
		return model.Task{}, fmt.Errorf("title is required")
	}
	priority := strings.TrimSpace(request.Priority)
	if priority == "" {
		priority = parent.Priority
	}
	task, err := ws.NewTask(title, strings.TrimSpace(request.Description), priority, strings.TrimSpace(request.Agent), parent.ID)
	if err != nil {
		return model.Task{}, err
	}
	if err := ws.SaveTask(task); err != nil {
		return model.Task{}, err
	}
	if !contains(parent.Subtasks, task.ID) {
		parent.Subtasks = append(parent.Subtasks, task.ID)
		parent.UpdatedAt = b.now()
		if err := ws.SaveTask(parent); err != nil {
			return model.Task{}, err
		}
	}
	return task, nil
}

func (b Bridge) MessageSibling(request TaskMessageRequest) (TaskMessage, error) {
	ws, err := b.workspace()
	if err != nil {
		return TaskMessage{}, err
	}
	sourceID, err := sanitizeID(request.SourceTaskID)
	if err != nil {
		return TaskMessage{}, fmt.Errorf("source task id: %w", err)
	}
	targetID, err := sanitizeID(request.TargetTaskID)
	if err != nil {
		return TaskMessage{}, fmt.Errorf("target task id: %w", err)
	}
	content := strings.TrimSpace(request.Content)
	if content == "" {
		return TaskMessage{}, fmt.Errorf("message content is required")
	}
	source, err := ws.LoadTask(sourceID)
	if err != nil {
		return TaskMessage{}, err
	}
	target, err := ws.LoadTask(targetID)
	if err != nil {
		return TaskMessage{}, err
	}
	if source.ParentTaskID == "" || source.ParentTaskID != target.ParentTaskID {
		return TaskMessage{}, fmt.Errorf("tasks %s and %s are not siblings", source.ID, target.ID)
	}
	now := b.now()
	message := TaskMessage{
		ID:           fmt.Sprintf("%s-to-%s-%d", source.ID, target.ID, now.Unix()),
		SourceTaskID: source.ID,
		TargetTaskID: target.ID,
		Content:      content,
		CreatedAt:    now,
	}
	if err := writeJSON(filepath.Join(ws.DotDir(), "messages", message.ID+".json"), message); err != nil {
		return TaskMessage{}, err
	}
	return message, nil
}

func (b Bridge) InspectRelatedWork(request RelatedWorkRequest) ([]RelatedWork, error) {
	ws, err := b.workspace()
	if err != nil {
		return nil, err
	}
	query := strings.ToLower(strings.TrimSpace(request.Query))
	if query == "" && strings.TrimSpace(request.TaskID) != "" {
		task, err := ws.LoadTask(request.TaskID)
		if err != nil {
			return nil, err
		}
		query = strings.ToLower(task.Title)
	}
	if query == "" {
		return nil, fmt.Errorf("query is required")
	}
	limit := request.Limit
	if limit <= 0 || limit > 20 {
		limit = 10
	}
	var related []RelatedWork
	tasks, err := ws.ListTasks()
	if err != nil {
		return nil, err
	}
	for _, task := range tasks {
		text := strings.ToLower(task.Title + " " + task.Description + " " + strings.Join(task.Subtasks, " "))
		if strings.Contains(text, query) || strings.Contains(query, strings.ToLower(task.Title)) {
			related = append(related, RelatedWork{Type: "task", ID: task.ID, Title: task.Title, Summary: task.Description, Path: filepath.ToSlash(filepath.Join(".thanos", "tasks", task.ID+".json"))})
		}
	}
	for _, path := range []string{filepath.Join(ws.DotDir(), "memory", "feature-graph.md"), filepath.Join(ws.DotDir(), "codebase", "summary.md")} {
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			continue
		}
		text := string(data)
		if strings.Contains(strings.ToLower(text), query) {
			related = append(related, RelatedWork{Type: "memory", ID: filepath.Base(path), Title: filepath.Base(path), Summary: firstLine(text), Path: filepath.ToSlash(relativeToRoot(ws.Root, path))})
		}
	}
	sort.SliceStable(related, func(i, j int) bool {
		if related[i].Type != related[j].Type {
			return related[i].Type < related[j].Type
		}
		return related[i].ID < related[j].ID
	})
	if len(related) > limit {
		related = related[:limit]
	}
	return related, nil
}

func (b Bridge) AttachBranch(request AttachBranchRequest) (BranchAttachment, error) {
	ws, err := b.workspace()
	if err != nil {
		return BranchAttachment{}, err
	}
	taskID, err := sanitizeID(request.TaskID)
	if err != nil {
		return BranchAttachment{}, err
	}
	branch, err := validateBranch(request.BranchName)
	if err != nil {
		return BranchAttachment{}, err
	}
	task, err := ws.LoadTask(taskID)
	if err != nil {
		return BranchAttachment{}, err
	}
	worktree := strings.TrimSpace(request.WorktreePath)
	if worktree == "" {
		worktree = filepath.ToSlash(filepath.Join(".thanos", "worktrees", task.ID))
	}
	task.BranchName = branch
	task.WorktreePath = filepath.ToSlash(worktree)
	task.UpdatedAt = b.now()
	if err := ws.SaveTask(task); err != nil {
		return BranchAttachment{}, err
	}
	attachment := BranchAttachment{TaskID: task.ID, BranchName: branch, WorktreePath: task.WorktreePath, AttachedAt: task.UpdatedAt}
	if err := writeJSON(filepath.Join(ws.DotDir(), "branches", task.ID+".json"), attachment); err != nil {
		return BranchAttachment{}, err
	}
	return attachment, nil
}

func (b Bridge) RequestUserReview(request ReviewRequest) (UserReviewRequest, error) {
	ws, err := b.workspace()
	if err != nil {
		return UserReviewRequest{}, err
	}
	taskID, err := sanitizeID(request.TaskID)
	if err != nil {
		return UserReviewRequest{}, err
	}
	if _, err := ws.LoadTask(taskID); err != nil {
		return UserReviewRequest{}, err
	}
	now := b.now()
	review := UserReviewRequest{
		ID:          fmt.Sprintf("review-request-%s-%d", taskID, now.Unix()),
		TaskID:      taskID,
		RequestedBy: strings.TrimSpace(request.RequestedBy),
		Notes:       strings.TrimSpace(request.Notes),
		Status:      "pending_user_review",
		CreatedAt:   now,
	}
	if err := writeJSON(filepath.Join(ws.DotDir(), "review-requests", review.ID+".json"), review); err != nil {
		return UserReviewRequest{}, err
	}
	return review, nil
}

func (b Bridge) workspace() (*workspace.Workspace, error) {
	if b.Workspace == nil {
		return nil, fmt.Errorf("workspace is required")
	}
	return b.Workspace, nil
}

func (b Bridge) now() time.Time {
	if b.Now != nil {
		return b.Now()
	}
	return time.Now().UTC()
}

func sanitizeID(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", fmt.Errorf("is required")
	}
	for _, ch := range value {
		if !(ch >= 'a' && ch <= 'z' || ch >= 'A' && ch <= 'Z' || ch >= '0' && ch <= '9' || ch == '-' || ch == '_' || ch == '.') {
			return "", fmt.Errorf("contains unsupported characters: %s", value)
		}
	}
	return value, nil
}

func validateBranch(value string) (string, error) {
	branch := strings.TrimSpace(value)
	if branch == "" {
		return "", fmt.Errorf("branch name is required")
	}
	if branch == "main" || branch == "master" || branch == "trunk" {
		return "", fmt.Errorf("refusing to attach protected branch %q", branch)
	}
	if strings.Contains(branch, "..") || strings.Contains(branch, " ") || strings.HasPrefix(branch, "/") || strings.HasSuffix(branch, "/") {
		return "", fmt.Errorf("invalid branch name: %s", branch)
	}
	return branch, nil
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
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
	return os.WriteFile(path, data, 0o644)
}

func firstLine(value string) string {
	for _, line := range strings.Split(value, "\n") {
		line = strings.TrimSpace(strings.TrimPrefix(line, "#"))
		if line != "" {
			return line
		}
	}
	return ""
}

func relativeToRoot(root, path string) string {
	relative, err := filepath.Rel(root, path)
	if err != nil {
		return path
	}
	return relative
}
