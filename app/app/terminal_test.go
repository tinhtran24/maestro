package app

import (
	"strings"
	"testing"
)

func TestScrubAuthEnvRemovesTokenVariables(t *testing.T) {
	env := scrubAuthEnv([]string{
		"PATH=/bin",
		"ANTHROPIC_API_KEY=secret",
		"CLAUDE_CODE_OAUTH_TOKEN=secret",
		"OPENAI_API_KEY=secret",
		"KEEP=value",
	})
	joined := strings.Join(env, "\n")
	for _, blocked := range []string{"ANTHROPIC_API_KEY", "CLAUDE_CODE_OAUTH_TOKEN", "OPENAI_API_KEY"} {
		if strings.Contains(joined, blocked) {
			t.Fatalf("env still contains %s: %v", blocked, env)
		}
	}
	if !strings.Contains(joined, "KEEP=value") || !strings.Contains(joined, "PATH=/bin") {
		t.Fatalf("expected ordinary env vars to remain: %v", env)
	}
}
