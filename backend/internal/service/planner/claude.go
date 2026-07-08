package planner

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// CLIRunner runs the planning prompt through the chosen agent's CLI in headless
// print mode. It reuses the agent's existing local login — no API key. Claude
// Code is first-class (its `--output-format json` envelope is parsed); other
// agents are run with a generic `-p <prompt>` and their stdout is returned, so
// the JSON-only instruction in the prompt still yields a parseable object.
type CLIRunner struct {
	// Binary overrides the resolved binary path (mainly for tests).
	Binary string
}

// claudeEnvelope is Claude Code's `--output-format json` result envelope. Only
// the fields we consume are modeled.
type claudeEnvelope struct {
	Type    string `json:"type"`
	IsError bool   `json:"is_error"`
	Result  string `json:"result"`
}

// Available reports whether the given agent's CLI can be resolved.
func (c *CLIRunner) Available(agent string) bool {
	return c.resolve(agent) != ""
}

// Run execs the agent CLI headlessly and returns the model's reply text.
func (c *CLIRunner) Run(ctx context.Context, agent, prompt string) (string, error) {
	bin := c.resolve(agent)
	if bin == "" {
		return "", fmt.Errorf("%w (agent %q)", ErrUnavailable, agent)
	}

	var cmd *exec.Cmd
	claude := isClaude(agent)
	if claude {
		// --output-format json wraps the reply in a machine-readable envelope so
		// we don't scrape a TTY. Print mode (-p) is one-shot.
		cmd = exec.CommandContext(ctx, bin, "-p", prompt, "--output-format", "json")
	} else {
		cmd = exec.CommandContext(ctx, bin, "-p", prompt)
	}

	out, err := cmd.Output()
	if err != nil {
		if ctx.Err() != nil {
			return "", fmt.Errorf("planner: %s timed out: %w", agent, ctx.Err())
		}
		var ee *exec.ExitError
		if errors.As(err, &ee) && len(ee.Stderr) > 0 {
			return "", fmt.Errorf("planner: %s failed: %s", agent, strings.TrimSpace(string(ee.Stderr)))
		}
		return "", fmt.Errorf("planner: %s failed: %w", agent, err)
	}

	if !claude {
		return string(out), nil
	}

	var env claudeEnvelope
	if err := json.Unmarshal(out, &env); err != nil {
		// Not the envelope we expected — return raw output; parseDraft still tries.
		return string(out), nil
	}
	if env.IsError {
		return "", fmt.Errorf("planner: claude reported an error")
	}
	return env.Result, nil
}

func isClaude(agent string) bool {
	a := strings.ToLower(strings.TrimSpace(agent))
	return a == "" || a == "claude" || a == "claudecode" || strings.HasPrefix(a, "claude")
}

// resolve finds the CLI binary for the agent. Explicit override and, for Claude,
// THANOS_CLAUDE_BIN win; then PATH; then common install locations. The agent id
// doubles as the binary name for non-Claude agents (codex, aider, cursor, …).
func (c *CLIRunner) resolve(agent string) string {
	if c.Binary != "" {
		return c.Binary
	}
	if isClaude(agent) {
		if env := strings.TrimSpace(os.Getenv("THANOS_CLAUDE_BIN")); env != "" {
			return env
		}
		if p := lookupBinary("claude"); p != "" {
			return p
		}
		home, _ := os.UserHomeDir()
		for _, rel := range [][]string{
			{".local", "bin", "claude"},
			{".npm", "bin", "claude"},
			{".claude", "local", "claude"},
		} {
			if home != "" {
				p := filepath.Join(append([]string{home}, rel...)...)
				if isExecutable(p) {
					return p
				}
			}
		}
		return ""
	}
	return lookupBinary(strings.TrimSpace(agent))
}

func lookupBinary(name string) string {
	if name == "" {
		return ""
	}
	if p, err := exec.LookPath(name); err == nil {
		return p
	}
	return ""
}

func isExecutable(p string) bool {
	info, err := os.Stat(p)
	return err == nil && !info.IsDir() && info.Mode()&0o111 != 0
}
