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
		"CREATE TABLE IF NOT EXISTS runtime_session_facts",
		"CREATE TABLE IF NOT EXISTS runtime_workspace_facts",
		"CREATE TABLE IF NOT EXISTS runtime_terminal_facts",
		"CREATE TABLE IF NOT EXISTS change_log",
		"CREATE TABLE IF NOT EXISTS runtime_events",
		"tasks_state_change_log_update",
		"tasks_gate_change_log_update",
		"runtime_session_facts_status_refresh_log_update",
		"'evidence_pending'",
		"source TEXT NOT NULL CHECK (source IN ('project', 'global', 'builtin'))",
	}
	for _, fragment := range required {
		if !strings.Contains(Schema, fragment) {
			t.Fatalf("schema missing %q", fragment)
		}
	}
}
