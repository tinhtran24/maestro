package agents

type ProviderStatus string

const (
	StatusInstalled  ProviderStatus = "installed"
	StatusNotFound   ProviderStatus = "not_found"
	StatusNeedsSetup ProviderStatus = "needs_setup"
)

type Provider struct {
	ID           string         `json:"id"`
	Name         string         `json:"name"`
	Command      string         `json:"command"`
	DetectedPath string         `json:"detected_path,omitempty"`
	Status       ProviderStatus `json:"status"`
	Version      string         `json:"version,omitempty"`
	Type         string         `json:"type"`
	Enabled      bool           `json:"enabled"`
	SetupHint    string         `json:"setup_hint,omitempty"`
}

func KnownProviders() []Provider {
	return []Provider{
		{ID: "claude-code", Name: "Claude Code", Command: "claude", Type: "cli"},
		{ID: "codex", Name: "Codex", Command: "codex", Type: "cli"},
		{ID: "gemini-cli", Name: "Gemini CLI", Command: "gemini", Type: "cli"},
		{ID: "opencode", Name: "OpenCode", Command: "opencode", Type: "cli"},
		{ID: "cursor-agent", Name: "Cursor Agent", Command: "cursor-agent", Type: "cli"},
		{ID: "aider", Name: "Aider", Command: "aider", Type: "cli"},
		{ID: "goose", Name: "Goose", Command: "goose", Type: "cli"},
		{ID: "shell", Name: "Shell", Command: "sh", Type: "shell"},
	}
}
