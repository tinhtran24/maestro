package daemon

// This file wires the provider-neutral SCM observer into daemon startup using
// the GitHub provider for v1. It keeps provider setup non-blocking for readiness
// by resolving tokens lazily inside the background observer path.

import (
	"context"
	"errors"
	"log/slog"

	scmgithub "github.com/tinhtran24/maestro/backend/internal/adapters/scm/github"
	"github.com/tinhtran24/maestro/backend/internal/domain"
	"github.com/tinhtran24/maestro/backend/internal/lifecycle"
	scmobserve "github.com/tinhtran24/maestro/backend/internal/observe/scm"
	"github.com/tinhtran24/maestro/backend/internal/storage/sqlite"
)

// startSCMObserver wires the provider-neutral SCM observer with the GitHub
// provider used by v1. Missing credentials do not fail daemon startup; the
// observer performs a lazy credential check in its background goroutine, logs
// one warning, and disables itself before any provider API calls.
func startSCMObserver(ctx context.Context, store *sqlite.Store, lcm *lifecycle.Manager, logger *slog.Logger) <-chan struct{} {
	github, err := newGitHubSCMProvider(logger)
	if err != nil {
		logSCMProviderDisabled(logger, err)
		return closedDone()
	}
	// Register every SCM adapter behind the dispatcher. Today only GitHub ships
	// an observation adapter, so this behaves identically to a bare GitHub
	// provider; adding GitLab/Bitbucket is a one-line registration here plus the
	// adapter, with no observer-level change.
	provider := scmobserve.NewMultiProvider(
		scmobserve.RegisteredProvider{Name: domain.SCMProviderGitHub, Provider: github},
	)
	observer := scmobserve.New(provider, store, lcm, scmobserve.Config{Logger: logger})
	return observer.Start(ctx)
}

func newGitHubSCMProvider(logger *slog.Logger) (*scmgithub.Provider, error) {
	tokens := scmgithub.FallbackTokenSource{
		scmgithub.EnvTokenSource{EnvVars: []string{"MAESTRO_GITHUB_TOKEN"}},
		&scmgithub.GitCredentialTokenSource{},
	}
	// Avoid token preflight on daemon startup and session service construction.
	// Git credential helpers may prompt or be slow; provider calls resolve
	// credentials lazily when claim-pr or the background observer needs GitHub.
	return scmgithub.NewProvider(scmgithub.ProviderOptions{Token: tokens, SkipTokenPreflight: true, Logger: logger})
}

func logSCMProviderDisabled(logger *slog.Logger, err error) {
	if errors.Is(err, scmgithub.ErrNoToken) || errors.Is(err, scmgithub.ErrAuthFailed) {
		logger.Warn("scm observer disabled: no usable GitHub token", "err", err)
	} else {
		logger.Warn("scm observer disabled: GitHub provider setup failed", "err", err)
	}
}

func closedDone() <-chan struct{} {
	done := make(chan struct{})
	close(done)
	return done
}
