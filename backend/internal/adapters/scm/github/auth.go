package github

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	aoprocess "github.com/tinhtran24/maestro/backend/internal/process"
)

// TokenSource yields a GitHub bearer token on demand. Production wires this
// to EnvTokenSource or GitCredentialTokenSource; tests inject StaticTokenSource.
type TokenSource interface {
	Token(ctx context.Context) (string, error)
}

// tokenInvalidator is the optional capability of dropping a cached token so
// the next call re-fetches it. The Client invokes this whenever GitHub
// responds with an auth-class failure: the next request will pick up a
// rotated token without restarting the daemon.
type tokenInvalidator interface {
	InvalidateToken()
}

// ErrNoToken is returned when no token source could yield a non-empty token.
var ErrNoToken = errors.New("github scm: no token configured")

// StaticTokenSource is a literal token, typically used in tests.
type StaticTokenSource string

// Token returns the literal token, or ErrNoToken if it is blank.
func (s StaticTokenSource) Token(context.Context) (string, error) {
	t := strings.TrimSpace(string(s))
	if t == "" {
		return "", ErrNoToken
	}
	return t, nil
}

// EnvTokenSource reads the first non-empty value from the listed env vars,
// falling back to GITHUB_TOKEN. Order matters: a project-scoped variable
// (MAESTRO_GITHUB_TOKEN) should win over the global default.
type EnvTokenSource struct {
	EnvVars []string
}

// Token returns the first non-empty env-var value found, or ErrNoToken.
func (s EnvTokenSource) Token(context.Context) (string, error) {
	for _, name := range s.EnvVars {
		if v := strings.TrimSpace(os.Getenv(name)); v != "" {
			return v, nil
		}
	}
	if v := strings.TrimSpace(os.Getenv("GITHUB_TOKEN")); v != "" {
		return v, nil
	}
	return "", ErrNoToken
}

// FallbackTokenSource tries each source in order, returning the first token. A
// source that returns ErrNoToken is skipped; other errors are remembered and
// surfaced if no later source yields a token.
type FallbackTokenSource []TokenSource

// Token returns the first non-empty token from the configured sources.
func (s FallbackTokenSource) Token(ctx context.Context) (string, error) {
	var firstErr error
	for _, src := range s {
		if src == nil {
			continue
		}
		tok, err := src.Token(ctx)
		if err == nil {
			return tok, nil
		}
		if errors.Is(err, ErrNoToken) {
			continue
		}
		if firstErr == nil {
			firstErr = err
		}
	}
	if firstErr != nil {
		return "", firstErr
	}
	return "", ErrNoToken
}

// InvalidateToken forwards cache invalidation to sources that support it.
func (s FallbackTokenSource) InvalidateToken() {
	for _, src := range s {
		if inv, ok := src.(tokenInvalidator); ok {
			inv.InvalidateToken()
		}
	}
}

const defaultGHTokenCacheTTL = 5 * time.Minute

// GitCredentialTokenSource reads a GitHub token from Git's configured
// credential helper. It supports the standard `password=<PAT>` response used
// by macOS Keychain, Git Credential Manager, and other native Git helpers, so
// the daemon does not require the optional gh CLI.
type GitCredentialTokenSource struct {
	// Git is the shell-out hook. Production leaves this nil and invokes
	// `git credential fill`; tests inject it to avoid touching local helpers.
	Git      func(ctx context.Context) (string, error)
	TokenTTL time.Duration
	Clock    func() time.Time

	mu        sync.Mutex
	token     string
	expiresAt time.Time
}

// Token returns a cached credential-helper token or queries Git for one.
func (s *GitCredentialTokenSource) Token(ctx context.Context) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.now()
	if s.token != "" && now.Before(s.expiresAt) {
		return s.token, nil
	}
	run := s.Git
	if run == nil {
		run = gitCredentialToken
	}
	out, err := run(ctx)
	if err != nil {
		return "", err
	}
	for _, line := range strings.Split(out, "\n") {
		key, value, ok := strings.Cut(line, "=")
		if ok && key == "password" && strings.TrimSpace(value) != "" {
			s.token = strings.TrimSpace(value)
			s.expiresAt = now.Add(s.ttl())
			return s.token, nil
		}
	}
	return "", ErrNoToken
}

// InvalidateToken drops the cached credential-helper token.
func (s *GitCredentialTokenSource) InvalidateToken() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.token = ""
	s.expiresAt = time.Time{}
}

func (s *GitCredentialTokenSource) now() time.Time {
	if s.Clock != nil {
		return s.Clock()
	}
	return time.Now()
}

func (s *GitCredentialTokenSource) ttl() time.Duration {
	if s.TokenTTL > 0 {
		return s.TokenTTL
	}
	return defaultGHTokenCacheTTL
}

func gitCredentialToken(ctx context.Context) (string, error) {
	cmd := aoprocess.CommandContext(ctx, "git", "credential", "fill")
	cmd.Stdin = strings.NewReader("protocol=https\nhost=github.com\n\n")
	out, err := cmd.Output()
	if err != nil {
		if ctx.Err() != nil {
			return "", ctx.Err()
		}
		// A missing Git credential helper or Git installation means this optional
		// authentication source has no token. The observer can then disable
		// cleanly instead of retrying a local configuration failure forever.
		return "", ErrNoToken
	}
	return string(out), nil
}

// GHTokenSource shells out to `gh auth token` when env vars are not
// configured. It memoizes the result for TokenTTL so we don't fork-exec on
// every request, but the Client invalidates the cache on auth failures so a
// rotated token is picked up on the next call. Tests inject GH so the gh
// binary is never required.
type GHTokenSource struct {
	// GH is the shell-out hook. Production leaves this nil and falls back
	// to `exec.CommandContext("gh", "auth", "token")`; tests inject a
	// fake to avoid touching the real binary.
	GH func(ctx context.Context) (string, error)
	// TokenTTL is how long a successful read is memoized. Zero means use
	// defaultGHTokenCacheTTL.
	TokenTTL time.Duration
	// Clock allows tests to drive expiration. Zero means time.Now.
	Clock func() time.Time

	mu        sync.Mutex
	token     string
	expiresAt time.Time
}

// Token returns the cached token if still fresh, otherwise re-runs gh.
func (s *GHTokenSource) Token(ctx context.Context) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.now()
	if s.token != "" && now.Before(s.expiresAt) {
		return s.token, nil
	}
	run := s.GH
	if run == nil {
		run = ghAuthToken
	}
	out, err := run(ctx)
	if err != nil {
		return "", err
	}
	token := strings.TrimSpace(out)
	if token == "" {
		return "", ErrNoToken
	}
	s.token = token
	s.expiresAt = now.Add(s.ttl())
	return token, nil
}

// InvalidateToken drops the memoized token so the next Token call shells
// out again. The Client calls this on 401/403-auth responses.
func (s *GHTokenSource) InvalidateToken() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.token = ""
	s.expiresAt = time.Time{}
}

func (s *GHTokenSource) now() time.Time {
	if s.Clock != nil {
		return s.Clock()
	}
	return time.Now()
}

func (s *GHTokenSource) ttl() time.Duration {
	if s.TokenTTL > 0 {
		return s.TokenTTL
	}
	return defaultGHTokenCacheTTL
}

func ghAuthToken(ctx context.Context) (string, error) {
	out, err := aoprocess.CommandContext(ctx, "gh", "auth", "token").Output()
	if err != nil {
		return "", ghTokenError(err)
	}
	return string(out), nil
}

func ghTokenError(err error) error {
	// A missing optional gh CLI is equivalent to having no configured token.
	// This lets the SCM observer disable itself once instead of logging a
	// transient retry warning on every poll.
	if errors.Is(err, exec.ErrNotFound) {
		return ErrNoToken
	}
	return err
}
