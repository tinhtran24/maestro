package skills

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadProjectSkillsParsesFrontmatterAndExitCriteria(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, ".thanos", "skills", "feature-spec")
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}
	writeSkill(t, filepath.Join(path, "SKILL.md"), `---
name: feature-spec
description: Create clear feature specs before implementation.
applies_to:
  - planning
  - feature
agents:
  - planner
required_evidence:
  - acceptance_criteria
  - risk_notes
---

# Skill: Feature Spec

## Workflow
- Read ticket.

## Exit criteria
- Acceptance criteria exist.
- Risks are listed.
`)

	loaded, err := LoadProjectSkills(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded) != 1 {
		t.Fatalf("loaded %d skills, want 1", len(loaded))
	}
	skill := loaded[0]
	if skill.ID != "project:feature-spec" || skill.Source != SourceProject || !skill.Trusted {
		t.Fatalf("unexpected loaded skill metadata: %#v", skill)
	}
	if got, want := len(skill.ExitCriteria), 2; got != want {
		t.Fatalf("exit criteria count = %d, want %d", got, want)
	}
}

func TestLoadSkillFileRejectsMissingExitCriteria(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "SKILL.md")
	writeSkill(t, path, `---
name: bad
description: Missing exit criteria.
agents:
  - coder
---

# Skill: Bad
`)

	_, err := LoadSkillFile(path, SourceProject, "")
	if err == nil || !strings.Contains(err.Error(), "exit criteria") {
		t.Fatalf("expected exit criteria error, got %v", err)
	}
}

func writeSkill(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
