package app

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
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

func TestNativeTerminalPersistsTranscript(t *testing.T) {
	root := t.TempDir()
	manager := NewNativeTerminalManager()
	session, err := manager.Start(context.Background(), NativeTerminalRequest{
		Command: "sh",
		Args:    []string{"-c", "printf hello"},
		CWD:     root,
		Label:   "test",
		Rows:    12,
		Cols:    80,
	})
	if err != nil {
		t.Fatalf("Start returned error: %v", err)
	}
	waitForSessionExit(t, manager, session.ID)
	data, err := os.ReadFile(session.TranscriptPath)
	if err != nil {
		t.Fatalf("ReadFile transcript returned error: %v", err)
	}
	text := string(data)
	if !strings.Contains(text, "hello") || !strings.Contains(text, "[exit] completed") {
		t.Fatalf("transcript missing output or exit marker: %q", text)
	}
	if wantPrefix := filepath.Join(root, ".thanos", "terminal", "sessions"); !strings.HasPrefix(session.TranscriptPath, wantPrefix) {
		t.Fatalf("transcript path = %q, want prefix %q", session.TranscriptPath, wantPrefix)
	}
}

func TestNativeTerminalWriteInput(t *testing.T) {
	root := t.TempDir()
	manager := NewNativeTerminalManager()
	session, err := manager.Start(context.Background(), NativeTerminalRequest{
		Command: "sh",
		Args:    []string{"-c", "read line; printf \"got:%s\" \"$line\""},
		CWD:     root,
		Label:   "input",
	})
	if err != nil {
		t.Fatalf("Start returned error: %v", err)
	}
	if err := manager.Write(NativeTerminalInputRequest{SessionID: session.ID, Data: "value\n"}); err != nil {
		t.Fatalf("Write returned error: %v", err)
	}
	waitForSessionExit(t, manager, session.ID)
	data, err := os.ReadFile(session.TranscriptPath)
	if err != nil {
		t.Fatalf("ReadFile transcript returned error: %v", err)
	}
	if !strings.Contains(string(data), "got:value") {
		t.Fatalf("transcript missing input response: %q", string(data))
	}
}

func TestAgentProviderStartsNeutralPersistedSession(t *testing.T) {
	root := t.TempDir()
	provider := NewRealProvider()
	provider.now = func() time.Time { return time.Date(2026, 7, 8, 15, 0, 0, 0, time.UTC) }
	manager := NewNativeTerminalManager()
	session, err := provider.StartAgentSession(context.Background(), manager, StartAgentRequest{
		Root:           root,
		ProviderID:     "custom-local",
		CustomCommand:  "printf",
		Mode:           AgentModePlanner,
		Prompt:         "Add provider routing",
		Context:        "Use the app provider layer.",
		Acceptance:     "Session is persisted.",
		Constraints:    "Do not mention a specific vendor.",
		AllowedFiles:   []string{"app/app"},
		ExpectedOutput: "Return a plan.",
	})
	if err != nil {
		t.Fatalf("StartAgentSession returned error: %v", err)
	}
	waitForSessionExit(t, manager, session.TerminalID)
	if session.ProviderID != "custom-local" || session.Mode != AgentModePlanner || session.TranscriptPath == "" {
		t.Fatalf("session = %#v", session)
	}
	if strings.Contains(strings.ToLower(session.Prompt), "claude") {
		t.Fatalf("prompt is provider-specific: %q", session.Prompt)
	}
	loaded, err := provider.ListAgentSessions(root)
	if err != nil {
		t.Fatalf("ListAgentSessions returned error: %v", err)
	}
	if len(loaded) != 1 || loaded[0].ID != session.ID || !strings.Contains(loaded[0].Prompt, "Acceptance Criteria") {
		t.Fatalf("loaded sessions = %#v", loaded)
	}
}

func TestAgentProviderDoesNotFallbackWhenSelectedProviderMissing(t *testing.T) {
	root := t.TempDir()
	provider := NewRealProvider()
	_, err := provider.StartAgentSession(context.Background(), NewNativeTerminalManager(), StartAgentRequest{
		Root:          root,
		ProviderID:    "custom-local",
		CustomCommand: "thanos-agent-command-that-does-not-exist",
		Mode:          AgentModeCoding,
		Prompt:        "Implement work",
	})
	if err == nil || !strings.Contains(err.Error(), "provider not installed") {
		t.Fatalf("expected provider-not-installed error, got %v", err)
	}
}

func waitForSessionExit(t *testing.T, manager *NativeTerminalManager, sessionID string) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		manager.mu.Lock()
		_, running := manager.sessions[sessionID]
		manager.mu.Unlock()
		if !running {
			return
		}
		time.Sleep(25 * time.Millisecond)
	}
	t.Fatalf("session %s did not exit", sessionID)
}
