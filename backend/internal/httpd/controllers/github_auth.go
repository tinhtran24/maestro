package controllers

import (
	"context"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/tinhtran/thanos/backend/internal/httpd/envelope"
	aoprocess "github.com/tinhtran/thanos/backend/internal/process"
)

const githubAuthProbeTimeout = 5 * time.Second

// GitHubAuthRunner runs the gh auth probe. Tests inject a fake.
type GitHubAuthRunner interface {
	LookPath(file string) (string, error)
	Token(ctx context.Context) (string, string, error)
}

// GitHubAuthController owns global GitHub credential diagnostics.
type GitHubAuthController struct {
	Runner GitHubAuthRunner
}

// Register mounts GitHub credential diagnostic routes.
func (c *GitHubAuthController) Register(r chi.Router) {
	r.Get("/github/auth", c.status)
}

func (c *GitHubAuthController) status(w http.ResponseWriter, r *http.Request) {
	if c.Runner == nil {
		c.Runner = shellGitHubAuthRunner{}
	}
	status := probeGitHubAuth(r.Context(), c.Runner)
	envelope.WriteJSON(w, 200, GitHubAuthStatusResponse{Status: status})
}

func probeGitHubAuth(ctx context.Context, runner GitHubAuthRunner) GitHubAuthStatus {
	if tokenSource := envGitHubTokenSource(); tokenSource != "" {
		return GitHubAuthStatus{Installed: true, Authenticated: true, Source: tokenSource, Message: "GitHub token is configured in the environment."}
	}
	path, err := runner.LookPath("gh")
	if err != nil {
		return GitHubAuthStatus{Installed: false, Authenticated: false, Source: "none", InstallCommand: "brew install gh", LoginCommand: "gh auth login", Message: "GitHub CLI is not installed."}
	}
	probeCtx, cancel := context.WithTimeout(ctx, githubAuthProbeTimeout)
	defer cancel()
	out, stderr, err := runner.Token(probeCtx)
	if err != nil {
		return GitHubAuthStatus{Installed: true, Authenticated: false, Source: "gh", BinaryPath: path, LoginCommand: "gh auth login", Message: sanitizeAuthOutput(stderr, err.Error())}
	}
	if strings.TrimSpace(out) == "" {
		return GitHubAuthStatus{Installed: true, Authenticated: false, Source: "gh", BinaryPath: path, LoginCommand: "gh auth login", Message: "GitHub CLI returned an empty auth token."}
	}
	return GitHubAuthStatus{Installed: true, Authenticated: true, Source: "gh", BinaryPath: path, Message: "GitHub CLI authentication is available."}
}

func envGitHubTokenSource() string {
	for _, name := range []string{"THANOS_GITHUB_TOKEN", "GITHUB_TOKEN"} {
		if strings.TrimSpace(os.Getenv(name)) != "" {
			return name
		}
	}
	return ""
}

func sanitizeAuthOutput(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return "GitHub authentication is not available."
}

type shellGitHubAuthRunner struct{}

func (shellGitHubAuthRunner) LookPath(file string) (string, error) {
	return exec.LookPath(file)
}

func (shellGitHubAuthRunner) Token(ctx context.Context) (string, string, error) {
	cmd := aoprocess.CommandContext(ctx, "gh", "auth", "token")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", string(out), err
	}
	return string(out), "", nil
}
