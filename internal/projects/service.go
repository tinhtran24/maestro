package projects

import (
	"fmt"
	"strings"
	"time"
)

type Project struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	RootPath       string    `json:"root_path"`
	GitRemoteURL   string    `json:"git_remote_url,omitempty"`
	DefaultBranch  string    `json:"default_branch"`
	WorktreeRoot   string    `json:"worktree_root"`
	PackageManager string    `json:"package_manager,omitempty"`
	DevCommand     string    `json:"dev_command,omitempty"`
	TestCommand    string    `json:"test_command,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type SetupRequest struct {
	Name           string
	RootPath       string
	GitRemoteURL   string
	DefaultBranch  string
	WorktreeRoot   string
	PackageManager string
	DevCommand     string
	TestCommand    string
}

func NewProject(request SetupRequest, now time.Time) (Project, error) {
	if strings.TrimSpace(request.RootPath) == "" {
		return Project{}, fmt.Errorf("root path is required")
	}
	name := strings.TrimSpace(request.Name)
	if name == "" {
		return Project{}, fmt.Errorf("project name is required")
	}
	if request.DefaultBranch == "" {
		request.DefaultBranch = "main"
	}
	if request.WorktreeRoot == "" {
		request.WorktreeRoot = ".thanos/worktrees"
	}
	return Project{
		ID:             slug(name),
		Name:           name,
		RootPath:       strings.TrimSpace(request.RootPath),
		GitRemoteURL:   strings.TrimSpace(request.GitRemoteURL),
		DefaultBranch:  strings.TrimSpace(request.DefaultBranch),
		WorktreeRoot:   strings.TrimSpace(request.WorktreeRoot),
		PackageManager: strings.TrimSpace(request.PackageManager),
		DevCommand:     strings.TrimSpace(request.DevCommand),
		TestCommand:    strings.TrimSpace(request.TestCommand),
		CreatedAt:      now,
		UpdatedAt:      now,
	}, nil
}

func slug(value string) string {
	value = strings.ToLower(value)
	var out strings.Builder
	dash := false
	for _, ch := range value {
		if ch >= 'a' && ch <= 'z' || ch >= '0' && ch <= '9' {
			out.WriteRune(ch)
			dash = false
		} else if !dash && out.Len() > 0 {
			out.WriteByte('-')
			dash = true
		}
	}
	return strings.Trim(out.String(), "-")
}
