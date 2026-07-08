package app

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode/utf8"
)

const maxFileExplorerReadBytes = 1024 * 1024

type ListWorkspaceFilesRequest struct {
	Root     string `json:"root"`
	Path     string `json:"path"`
	MaxDepth int    `json:"maxDepth"`
}

type ReadWorkspaceFileRequest struct {
	Root string `json:"root"`
	Path string `json:"path"`
}

type WriteWorkspaceFileRequest struct {
	Root    string `json:"root"`
	Path    string `json:"path"`
	Content string `json:"content"`
}

type PreviewTaskDiffRequest struct {
	Root   string `json:"root"`
	TaskID string `json:"taskId"`
}

type FileExplorerInfo struct {
	Root    string          `json:"root"`
	Base    string          `json:"base"`
	Entries []FileTreeEntry `json:"entries"`
}

type FileTreeEntry struct {
	Name       string          `json:"name"`
	Path       string          `json:"path"`
	Kind       string          `json:"kind"`
	Size       int64           `json:"size"`
	ModifiedAt string          `json:"modifiedAt,omitempty"`
	ReadOnly   bool            `json:"readOnly"`
	Children   []FileTreeEntry `json:"children,omitempty"`
}

type WorkspaceFileInfo struct {
	Path       string `json:"path"`
	Name       string `json:"name"`
	Content    string `json:"content"`
	Encoding   string `json:"encoding"`
	Size       int64  `json:"size"`
	ModifiedAt string `json:"modifiedAt,omitempty"`
	ReadOnly   bool   `json:"readOnly"`
	Virtual    bool   `json:"virtual"`
}

type TaskDiffPreviewInfo struct {
	TaskID   string `json:"taskId"`
	Worktree string `json:"worktree"`
	Diff     string `json:"diff"`
	DiffStat string `json:"diffStat"`
}

func (p *RealProvider) ListWorkspaceFiles(req ListWorkspaceFilesRequest) (*FileExplorerInfo, error) {
	root, err := normalizeWorkspaceRoot(req.Root)
	if err != nil {
		return nil, err
	}
	base, rel, err := resolveWorkspacePath(root, req.Path)
	if err != nil {
		return nil, err
	}
	stat, err := os.Stat(base)
	if err != nil {
		return nil, err
	}
	if !stat.IsDir() {
		return nil, fmt.Errorf("file explorer path is not a directory: %s", rel)
	}
	maxDepth := req.MaxDepth
	if maxDepth <= 0 {
		maxDepth = 2
	}
	entries, err := listFileTree(root, base, maxDepth)
	if err != nil {
		return nil, err
	}
	if rel == "." || rel == "" {
		entries = append(entries, taskPromptVirtualEntries(root)...)
		sortFileEntries(entries)
	}
	return &FileExplorerInfo{Root: root, Base: rel, Entries: entries}, nil
}

func (p *RealProvider) ReadWorkspaceFile(req ReadWorkspaceFileRequest) (*WorkspaceFileInfo, error) {
	root, err := normalizeWorkspaceRoot(req.Root)
	if err != nil {
		return nil, err
	}
	if prompt, ok, err := readTaskPromptVirtualEntry(root, req.Path); ok || err != nil {
		return prompt, err
	}
	path, rel, err := resolveWorkspacePath(root, req.Path)
	if err != nil {
		return nil, err
	}
	stat, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if stat.IsDir() {
		return nil, fmt.Errorf("workspace path is a directory: %s", rel)
	}
	if stat.Size() > maxFileExplorerReadBytes {
		return nil, fmt.Errorf("file is too large to read in explorer: %s", rel)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if !utf8.Valid(data) || looksBinary(data) {
		return nil, fmt.Errorf("file is not a UTF-8 text file: %s", rel)
	}
	return &WorkspaceFileInfo{
		Path:       rel,
		Name:       filepath.Base(path),
		Content:    string(data),
		Encoding:   "utf-8",
		Size:       stat.Size(),
		ModifiedAt: stat.ModTime().UTC().Format(time.RFC3339),
	}, nil
}

func (p *RealProvider) WriteWorkspaceFile(req WriteWorkspaceFileRequest) (*WorkspaceFileInfo, error) {
	root, err := normalizeWorkspaceRoot(req.Root)
	if err != nil {
		return nil, err
	}
	if isTaskPromptVirtualPath(req.Path) {
		return nil, fmt.Errorf("virtual task prompt entries are read-only")
	}
	path, rel, err := resolveWorkspacePath(root, req.Path)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(rel) == "" || rel == "." {
		return nil, fmt.Errorf("workspace file path is required")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	if err := os.WriteFile(path, []byte(req.Content), 0o644); err != nil {
		return nil, err
	}
	now := p.now().UTC().Format(time.RFC3339)
	store := NewWorkspaceStore(root)
	if err := store.AppendEvent(EventInfo{
		ID:      "event-file-write-" + stableID(rel+"-"+now),
		At:      now,
		Kind:    "WorkspaceFileWritten",
		Message: rel,
	}); err != nil {
		return nil, err
	}
	return p.ReadWorkspaceFile(ReadWorkspaceFileRequest{Root: root, Path: rel})
}

func (p *RealProvider) PreviewTaskDiff(ctx context.Context, req PreviewTaskDiffRequest) (*TaskDiffPreviewInfo, error) {
	root, err := normalizeWorkspaceRoot(req.Root)
	if err != nil {
		return nil, err
	}
	tasks, err := loadTasks(root)
	if err != nil {
		return nil, err
	}
	for _, task := range tasks {
		if task.ID != req.TaskID {
			continue
		}
		worktree := resolveTaskWorktree(root, task)
		if worktree == "" {
			return nil, fmt.Errorf("task worktree is not available: %s", task.ID)
		}
		diff, diffStat, err := taskGitDiff(ctx, worktree)
		if err != nil {
			return nil, err
		}
		return &TaskDiffPreviewInfo{
			TaskID:   task.ID,
			Worktree: filepath.ToSlash(task.Worktree),
			Diff:     diff,
			DiffStat: diffStat,
		}, nil
	}
	return nil, fmt.Errorf("task not found: %s", req.TaskID)
}

func listFileTree(root, dir string, maxDepth int) ([]FileTreeEntry, error) {
	items, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	entries := make([]FileTreeEntry, 0, len(items))
	for _, item := range items {
		path := filepath.Join(dir, item.Name())
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return nil, err
		}
		rel = filepath.ToSlash(rel)
		if shouldIgnoreExplorerPath(rel, item) {
			continue
		}
		info, err := item.Info()
		if err != nil {
			if errors.Is(err, fs.ErrPermission) {
				continue
			}
			return nil, err
		}
		entry := FileTreeEntry{
			Name:       item.Name(),
			Path:       rel,
			Kind:       "file",
			Size:       info.Size(),
			ModifiedAt: info.ModTime().UTC().Format(time.RFC3339),
		}
		if item.IsDir() {
			entry.Kind = "directory"
			entry.Size = 0
			if maxDepth > 1 {
				children, err := listFileTree(root, path, maxDepth-1)
				if err != nil {
					return nil, err
				}
				entry.Children = children
			}
		}
		entries = append(entries, entry)
	}
	sortFileEntries(entries)
	return entries, nil
}

func taskPromptVirtualEntries(root string) []FileTreeEntry {
	tasks, err := loadTasks(root)
	if err != nil || len(tasks) == 0 {
		return nil
	}
	children := make([]FileTreeEntry, 0, len(tasks))
	for _, task := range tasks {
		task = hydrateTask(task)
		name := safeExplorerName(firstNonEmpty(task.Title, task.ID)) + ".prompt.md"
		children = append(children, FileTreeEntry{
			Name:       name,
			Path:       taskPromptVirtualPath(task.ID),
			Kind:       "virtual",
			Size:       int64(len(task.Prompt)),
			ModifiedAt: task.UpdatedAt,
			ReadOnly:   true,
		})
	}
	sortFileEntries(children)
	return []FileTreeEntry{{
		Name:     "Task Prompts",
		Path:     ".thanos/virtual/task-prompts",
		Kind:     "directory",
		ReadOnly: true,
		Children: children,
	}}
}

func readTaskPromptVirtualEntry(root, requestPath string) (*WorkspaceFileInfo, bool, error) {
	if !isTaskPromptVirtualPath(requestPath) {
		return nil, false, nil
	}
	taskID := strings.TrimSuffix(strings.TrimPrefix(filepath.ToSlash(requestPath), ".thanos/virtual/task-prompts/"), ".prompt.md")
	tasks, err := loadTasks(root)
	if err != nil {
		return nil, true, err
	}
	for _, task := range tasks {
		if task.ID != taskID {
			continue
		}
		task = hydrateTask(task)
		content := "# " + firstNonEmpty(task.Title, task.ID) + "\n\n" + task.Prompt
		return &WorkspaceFileInfo{
			Path:       taskPromptVirtualPath(task.ID),
			Name:       safeExplorerName(firstNonEmpty(task.Title, task.ID)) + ".prompt.md",
			Content:    content,
			Encoding:   "utf-8",
			Size:       int64(len(content)),
			ModifiedAt: task.UpdatedAt,
			ReadOnly:   true,
			Virtual:    true,
		}, true, nil
	}
	return nil, true, fmt.Errorf("task prompt entry not found: %s", taskID)
}

func resolveWorkspacePath(root, requestPath string) (string, string, error) {
	cleanRequest := filepath.Clean(strings.TrimSpace(filepath.FromSlash(requestPath)))
	if cleanRequest == "." || cleanRequest == string(filepath.Separator) {
		cleanRequest = ""
	}
	if filepath.IsAbs(cleanRequest) {
		return "", "", fmt.Errorf("workspace file path must be relative")
	}
	path := filepath.Clean(filepath.Join(root, cleanRequest))
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return "", "", err
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", "", fmt.Errorf("workspace file path escapes root")
	}
	return path, filepath.ToSlash(rel), nil
}

func shouldIgnoreExplorerPath(rel string, entry os.DirEntry) bool {
	name := entry.Name()
	if name == ".DS_Store" {
		return true
	}
	if entry.IsDir() {
		switch name {
		case ".git", "node_modules", "dist", "target", ".next", ".vite":
			return true
		}
	}
	if strings.HasPrefix(rel, ".thanos/worktrees/") {
		return true
	}
	return false
}

func sortFileEntries(entries []FileTreeEntry) {
	sort.SliceStable(entries, func(i, j int) bool {
		if entries[i].Kind == "directory" && entries[j].Kind != "directory" {
			return true
		}
		if entries[i].Kind != "directory" && entries[j].Kind == "directory" {
			return false
		}
		return strings.ToLower(entries[i].Name) < strings.ToLower(entries[j].Name)
	})
}

func looksBinary(data []byte) bool {
	for _, b := range data {
		if b == 0 {
			return true
		}
	}
	return false
}

func isTaskPromptVirtualPath(path string) bool {
	path = filepath.ToSlash(strings.TrimSpace(path))
	return strings.HasPrefix(path, ".thanos/virtual/task-prompts/") && strings.HasSuffix(path, ".prompt.md")
}

func taskPromptVirtualPath(taskID string) string {
	return ".thanos/virtual/task-prompts/" + taskID + ".prompt.md"
}

func safeExplorerName(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	replacer := strings.NewReplacer("/", "-", "\\", "-", ":", "-", "*", "-", "?", "-", "\"", "", "<", "-", ">", "-", "|", "-")
	value = replacer.Replace(value)
	value = strings.Join(strings.Fields(value), "-")
	if value == "" {
		return "task"
	}
	return value
}
