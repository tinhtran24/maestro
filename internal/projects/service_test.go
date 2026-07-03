package projects

import (
	"testing"
	"time"
)

func TestNewProjectDefaultsLocalWorkbenchFields(t *testing.T) {
	now := time.Date(2026, 7, 3, 0, 0, 0, 0, time.UTC)
	project, err := NewProject(SetupRequest{Name: "My App", RootPath: "/tmp/my-app"}, now)
	if err != nil {
		t.Fatal(err)
	}
	if project.ID != "my-app" || project.DefaultBranch != "main" || project.WorktreeRoot != ".thanos/worktrees" {
		t.Fatalf("unexpected project defaults: %#v", project)
	}
}
