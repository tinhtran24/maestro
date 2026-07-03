package workbench

import (
	"strings"
	"testing"
)

func TestSchemaIncludesSkillTablesAndLifecycleChecks(t *testing.T) {
	required := []string{
		"CREATE TABLE IF NOT EXISTS skills",
		"CREATE TABLE IF NOT EXISTS skill_runs",
		"CREATE TABLE IF NOT EXISTS evidence",
		"'evidence_pending'",
		"source TEXT NOT NULL CHECK (source IN ('project', 'global', 'builtin'))",
	}
	for _, fragment := range required {
		if !strings.Contains(Schema, fragment) {
			t.Fatalf("schema missing %q", fragment)
		}
	}
}
