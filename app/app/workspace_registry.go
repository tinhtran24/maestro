package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

const SchemaVersionWorkspaceRegistry = 1

var workspaceRegistryMu sync.Mutex

type WorkspaceFolderInfo struct {
	ID    string `json:"id"`
	Path  string `json:"path"`
	Label string `json:"label"`
}

type WorkspaceRecordInfo struct {
	SchemaVersion  int                   `json:"schema_version"`
	ID             string                `json:"id"`
	Name           string                `json:"name"`
	DataKey        string                `json:"dataKey"`
	Folders        []WorkspaceFolderInfo `json:"folders"`
	ActiveFolderID string                `json:"activeFolderId"`
	CreatedAt      string                `json:"createdAt"`
	UpdatedAt      string                `json:"updatedAt"`
}

type WorkspaceRegistryInfo struct {
	SchemaVersion     int                   `json:"schema_version"`
	ActiveWorkspaceID string                `json:"activeWorkspaceId"`
	Workspaces        []WorkspaceRecordInfo `json:"workspaces"`
}

type CreateWorkspaceRequest struct {
	Name    string                `json:"name"`
	Path    string                `json:"path"`
	Folders []WorkspaceFolderInfo `json:"folders"`
}

type UpdateWorkspaceRequest struct {
	ID             string                `json:"id"`
	Name           string                `json:"name"`
	Folders        []WorkspaceFolderInfo `json:"folders"`
	ActiveFolderID string                `json:"activeFolderId"`
}

type ActivateWorkspaceRequest struct {
	ID string `json:"id"`
}

type DeleteWorkspaceRequest struct {
	ID string `json:"id"`
}

func (p *RealProvider) ListWorkspaces() (*WorkspaceRegistryInfo, error) {
	return loadWorkspaceRegistry()
}

func (p *RealProvider) CreateWorkspace(req CreateWorkspaceRequest) (*WorkspaceRecordInfo, error) {
	registry, err := loadWorkspaceRegistry()
	if err != nil {
		return nil, err
	}
	now := p.now().UTC().Format(time.RFC3339)
	folders, err := normalizeWorkspaceFolders(req.Path, req.Folders)
	if err != nil {
		return nil, err
	}
	name := strings.TrimSpace(req.Name)
	if name == "" && len(folders) > 0 {
		name = workspaceName(folders[0].Path)
	}
	if name == "" {
		return nil, fmt.Errorf("workspace name or folder is required")
	}
	id := "workspace-" + stableID(fmt.Sprintf("%s-%s-%s", now, name, firstFolderPath(folders)))
	record := WorkspaceRecordInfo{
		SchemaVersion:  SchemaVersionWorkspaceRegistry,
		ID:             id,
		Name:           name,
		DataKey:        "data-" + stableID(id+"-"+now),
		Folders:        folders,
		ActiveFolderID: folders[0].ID,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	registry.Workspaces = append(registry.Workspaces, record)
	registry.ActiveWorkspaceID = record.ID
	if err := saveWorkspaceRegistry(registry); err != nil {
		return nil, err
	}
	return &record, nil
}

func (p *RealProvider) UpdateWorkspace(req UpdateWorkspaceRequest) (*WorkspaceRecordInfo, error) {
	registry, err := loadWorkspaceRegistry()
	if err != nil {
		return nil, err
	}
	folders, err := normalizeWorkspaceFolders("", req.Folders)
	if err != nil {
		return nil, err
	}
	for index := range registry.Workspaces {
		if registry.Workspaces[index].ID != req.ID {
			continue
		}
		record := registry.Workspaces[index]
		name := strings.TrimSpace(req.Name)
		if name != "" {
			record.Name = name
		}
		if len(folders) > 0 {
			record.Folders = folders
		}
		if len(record.Folders) == 0 {
			return nil, fmt.Errorf("workspace must have at least one folder")
		}
		record.ActiveFolderID = normalizeActiveFolderID(record.Folders, req.ActiveFolderID, record.ActiveFolderID)
		record.UpdatedAt = p.now().UTC().Format(time.RFC3339)
		registry.Workspaces[index] = record
		if err := saveWorkspaceRegistry(registry); err != nil {
			return nil, err
		}
		return &record, nil
	}
	return nil, fmt.Errorf("workspace not found: %s", req.ID)
}

func (p *RealProvider) DeleteWorkspace(req DeleteWorkspaceRequest) (*WorkspaceRegistryInfo, error) {
	registry, err := loadWorkspaceRegistry()
	if err != nil {
		return nil, err
	}
	next := registry.Workspaces[:0]
	removed := false
	for _, record := range registry.Workspaces {
		if record.ID == req.ID {
			removed = true
			continue
		}
		next = append(next, record)
	}
	if !removed {
		return nil, fmt.Errorf("workspace not found: %s", req.ID)
	}
	registry.Workspaces = next
	if registry.ActiveWorkspaceID == req.ID {
		registry.ActiveWorkspaceID = ""
		if len(registry.Workspaces) > 0 {
			registry.ActiveWorkspaceID = registry.Workspaces[0].ID
		}
	}
	if err := saveWorkspaceRegistry(registry); err != nil {
		return nil, err
	}
	return registry, nil
}

func (p *RealProvider) ActivateWorkspace(req ActivateWorkspaceRequest) (*WorkspaceRecordInfo, error) {
	registry, err := loadWorkspaceRegistry()
	if err != nil {
		return nil, err
	}
	for index := range registry.Workspaces {
		if registry.Workspaces[index].ID != req.ID {
			continue
		}
		registry.ActiveWorkspaceID = req.ID
		if err := saveWorkspaceRegistry(registry); err != nil {
			return nil, err
		}
		return &registry.Workspaces[index], nil
	}
	return nil, fmt.Errorf("workspace not found: %s", req.ID)
}

func (p *RealProvider) LoadActiveWorkspace() (*WorkspaceInfo, error) {
	registry, err := loadWorkspaceRegistry()
	if err != nil {
		return nil, err
	}
	record := activeWorkspaceRecord(registry)
	if record == nil {
		return p.LoadWorkspace("")
	}
	folder := activeWorkspaceFolder(*record)
	if folder == nil {
		return nil, fmt.Errorf("active workspace has no folder: %s", record.ID)
	}
	workspace, err := p.LoadWorkspace(folder.Path)
	if err != nil {
		return nil, err
	}
	workspace.WorkspaceID = record.ID
	workspace.DataKey = record.DataKey
	workspace.Folders = record.Folders
	return workspace, nil
}

func loadWorkspaceRegistry() (*WorkspaceRegistryInfo, error) {
	workspaceRegistryMu.Lock()
	defer workspaceRegistryMu.Unlock()
	return loadWorkspaceRegistryUnlocked()
}

func loadWorkspaceRegistryUnlocked() (*WorkspaceRegistryInfo, error) {
	path, err := workspaceRegistryPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &WorkspaceRegistryInfo{SchemaVersion: SchemaVersionWorkspaceRegistry, Workspaces: []WorkspaceRecordInfo{}}, nil
		}
		return nil, err
	}
	var registry WorkspaceRegistryInfo
	if err := json.Unmarshal(data, &registry); err != nil {
		return nil, err
	}
	registry.SchemaVersion = SchemaVersionWorkspaceRegistry
	for index := range registry.Workspaces {
		migrateWorkspaceRecord(&registry.Workspaces[index])
	}
	sort.SliceStable(registry.Workspaces, func(i, j int) bool {
		return registry.Workspaces[i].UpdatedAt > registry.Workspaces[j].UpdatedAt
	})
	return &registry, nil
}

func saveWorkspaceRegistry(registry *WorkspaceRegistryInfo) error {
	workspaceRegistryMu.Lock()
	defer workspaceRegistryMu.Unlock()
	return saveWorkspaceRegistryUnlocked(registry)
}

func saveWorkspaceRegistryUnlocked(registry *WorkspaceRegistryInfo) error {
	path, err := workspaceRegistryPath()
	if err != nil {
		return err
	}
	registry.SchemaVersion = SchemaVersionWorkspaceRegistry
	for index := range registry.Workspaces {
		migrateWorkspaceRecord(&registry.Workspaces[index])
	}
	return atomicWriteJSON(path, registry)
}

func workspaceRegistryPath() (string, error) {
	if override := strings.TrimSpace(os.Getenv("THANOS_CONFIG_DIR")); override != "" {
		return filepath.Join(override, "workspaces.json"), nil
	}
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "Thanos", "workspaces.json"), nil
}

func atomicWriteJSON(path string, value any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	tmp, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+".*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	cleanup := true
	defer func() {
		if cleanup {
			_ = os.Remove(tmpName)
		}
	}()
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpName, path); err != nil {
		return err
	}
	cleanup = false
	return nil
}

func normalizeWorkspaceFolders(path string, folders []WorkspaceFolderInfo) ([]WorkspaceFolderInfo, error) {
	records := make([]WorkspaceFolderInfo, 0, len(folders)+1)
	if strings.TrimSpace(path) != "" {
		records = append(records, WorkspaceFolderInfo{Path: path})
	}
	records = append(records, folders...)
	seen := make(map[string]bool)
	normalized := make([]WorkspaceFolderInfo, 0, len(records))
	for _, folder := range records {
		abs := strings.TrimSpace(folder.Path)
		if abs == "" {
			continue
		}
		var err error
		abs, err = filepath.Abs(abs)
		if err != nil {
			return nil, err
		}
		if stat, err := os.Stat(abs); err != nil {
			return nil, err
		} else if !stat.IsDir() {
			return nil, fmt.Errorf("workspace folder is not a directory: %s", abs)
		}
		if gitRoot, err := gitTopLevel(abs); err == nil && gitRoot != "" {
			abs = gitRoot
		}
		clean := filepath.Clean(abs)
		if seen[clean] {
			continue
		}
		seen[clean] = true
		id := strings.TrimSpace(folder.ID)
		if id == "" {
			id = "folder-" + stableID(clean)
		}
		label := strings.TrimSpace(folder.Label)
		if label == "" {
			label = workspaceName(clean)
		}
		normalized = append(normalized, WorkspaceFolderInfo{ID: id, Path: clean, Label: label})
	}
	if len(normalized) == 0 {
		return nil, fmt.Errorf("workspace must have at least one folder")
	}
	return normalized, nil
}

func migrateWorkspaceRecord(record *WorkspaceRecordInfo) {
	record.SchemaVersion = SchemaVersionWorkspaceRegistry
	if record.DataKey == "" {
		record.DataKey = "data-" + stableID(record.ID+"-"+record.Name)
	}
	for index := range record.Folders {
		if record.Folders[index].ID == "" {
			record.Folders[index].ID = "folder-" + stableID(record.Folders[index].Path)
		}
		if record.Folders[index].Label == "" {
			record.Folders[index].Label = workspaceName(record.Folders[index].Path)
		}
	}
	record.ActiveFolderID = normalizeActiveFolderID(record.Folders, record.ActiveFolderID, "")
}

func normalizeActiveFolderID(folders []WorkspaceFolderInfo, requested string, fallbackID string) string {
	for _, candidate := range []string{requested, fallbackID} {
		candidate = strings.TrimSpace(candidate)
		if candidate == "" {
			continue
		}
		for _, folder := range folders {
			if folder.ID == candidate {
				return candidate
			}
		}
	}
	if len(folders) > 0 {
		return folders[0].ID
	}
	return ""
}

func activeWorkspaceRecord(registry *WorkspaceRegistryInfo) *WorkspaceRecordInfo {
	if registry == nil {
		return nil
	}
	for index := range registry.Workspaces {
		if registry.Workspaces[index].ID == registry.ActiveWorkspaceID {
			return &registry.Workspaces[index]
		}
	}
	if len(registry.Workspaces) > 0 {
		return &registry.Workspaces[0]
	}
	return nil
}

func activeWorkspaceFolder(record WorkspaceRecordInfo) *WorkspaceFolderInfo {
	for index := range record.Folders {
		if record.Folders[index].ID == record.ActiveFolderID {
			return &record.Folders[index]
		}
	}
	if len(record.Folders) > 0 {
		return &record.Folders[0]
	}
	return nil
}

func firstFolderPath(folders []WorkspaceFolderInfo) string {
	if len(folders) == 0 {
		return ""
	}
	return folders[0].Path
}
