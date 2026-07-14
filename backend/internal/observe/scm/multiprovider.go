package scm

import (
	"context"
	"fmt"

	"github.com/tinhtran24/maestro/backend/internal/domain"
	"github.com/tinhtran24/maestro/backend/internal/ports"
)

// RegisteredProvider pairs a provider adapter with the SCM provider name it
// owns. Name must match the Provider field the adapter stamps onto the
// ports.SCMRepo values it returns from ParseRepository, since repo-scoped calls
// route back to the adapter by that name.
type RegisteredProvider struct {
	Name     domain.SCMProvider
	Provider Provider
}

// MultiProvider dispatches provider-neutral SCM calls to the adapter that owns
// each repository. It is the registration seam for multi-host SCM observation:
// the observer keeps taking a single Provider, and MultiProvider fans that out
// across the registered adapters. ParseRepository tries each adapter in
// registration order and the first match wins; every repo-scoped call routes by
// ports.SCMRepo.Provider. With a single registered adapter it behaves
// identically to that adapter.
type MultiProvider struct {
	ordered []RegisteredProvider
	byName  map[domain.SCMProvider]Provider
}

// Compile-time assurance that MultiProvider satisfies the observer contract.
var _ Provider = (*MultiProvider)(nil)

// NewMultiProvider builds a dispatcher over the given adapters, in order. When
// two entries share a Name the later one wins the repo-scoped routing map,
// while both remain in ParseRepository order.
func NewMultiProvider(providers ...RegisteredProvider) *MultiProvider {
	byName := make(map[domain.SCMProvider]Provider, len(providers))
	for _, rp := range providers {
		if rp.Provider == nil {
			continue
		}
		byName[rp.Name] = rp.Provider
	}
	return &MultiProvider{ordered: providers, byName: byName}
}

// ParseRepository returns the first registered adapter's parse of remote.
func (m *MultiProvider) ParseRepository(remote string) (ports.SCMRepo, bool) {
	for _, rp := range m.ordered {
		if rp.Provider == nil {
			continue
		}
		if repo, ok := rp.Provider.ParseRepository(remote); ok {
			return repo, true
		}
	}
	return ports.SCMRepo{}, false
}

// RepoPRListGuard routes to the adapter that owns repo.
func (m *MultiProvider) RepoPRListGuard(ctx context.Context, repo ports.SCMRepo, etag string) (ports.SCMGuardResult, error) {
	p, err := m.providerFor(repo.Provider)
	if err != nil {
		return ports.SCMGuardResult{}, err
	}
	return p.RepoPRListGuard(ctx, repo, etag)
}

// ListOpenPRsByRepo routes to the adapter that owns repo.
func (m *MultiProvider) ListOpenPRsByRepo(ctx context.Context, repo ports.SCMRepo) ([]ports.SCMPRObservation, error) {
	p, err := m.providerFor(repo.Provider)
	if err != nil {
		return nil, err
	}
	return p.ListOpenPRsByRepo(ctx, repo)
}

// CommitChecksGuard routes to the adapter that owns repo.
func (m *MultiProvider) CommitChecksGuard(ctx context.Context, repo ports.SCMRepo, headSHA, etag string) (ports.SCMGuardResult, error) {
	p, err := m.providerFor(repo.Provider)
	if err != nil {
		return ports.SCMGuardResult{}, err
	}
	return p.CommitChecksGuard(ctx, repo, headSHA, etag)
}

// FetchPullRequests groups the refs by owning provider and dispatches one batch
// per provider, preserving first-seen provider order. Observations are matched
// back to refs by URL downstream, so the concatenated order is not significant.
// A batch that fails (or a ref whose provider is not registered) fails the whole
// call, matching the observer's per-chunk failure handling.
func (m *MultiProvider) FetchPullRequests(ctx context.Context, refs []ports.SCMPRRef) ([]ports.SCMObservation, error) {
	if len(refs) == 0 {
		return nil, nil
	}
	type group struct {
		name domain.SCMProvider
		refs []ports.SCMPRRef
	}
	var groups []*group
	index := map[domain.SCMProvider]*group{}
	for _, ref := range refs {
		name := domain.SCMProvider(ref.Repo.Provider)
		g := index[name]
		if g == nil {
			g = &group{name: name}
			index[name] = g
			groups = append(groups, g)
		}
		g.refs = append(g.refs, ref)
	}
	out := make([]ports.SCMObservation, 0, len(refs))
	for _, g := range groups {
		p, err := m.providerFor(string(g.name))
		if err != nil {
			return nil, err
		}
		obs, err := p.FetchPullRequests(ctx, g.refs)
		if err != nil {
			return nil, err
		}
		out = append(out, obs...)
	}
	return out, nil
}

// FetchFailedCheckLogTail routes to the adapter that owns repo.
func (m *MultiProvider) FetchFailedCheckLogTail(ctx context.Context, repo ports.SCMRepo, check ports.SCMCheckObservation) (string, error) {
	p, err := m.providerFor(repo.Provider)
	if err != nil {
		return "", err
	}
	return p.FetchFailedCheckLogTail(ctx, repo, check)
}

// FetchReviewThreads routes to the adapter that owns ref.Repo.
func (m *MultiProvider) FetchReviewThreads(ctx context.Context, ref ports.SCMPRRef) (ports.SCMReviewObservation, error) {
	p, err := m.providerFor(ref.Repo.Provider)
	if err != nil {
		return ports.SCMReviewObservation{}, err
	}
	return p.FetchReviewThreads(ctx, ref)
}

// SCMCredentialsAvailable satisfies the observer's optional credential gate by
// aggregating across registered adapters: credentials are available when any
// adapter reports them available (or has no credential gate at all). It reports
// unavailable only when every gated adapter agrees, surfacing the first error
// seen when none reported available.
func (m *MultiProvider) SCMCredentialsAvailable(ctx context.Context) (bool, error) {
	var firstErr error
	for _, rp := range m.ordered {
		if rp.Provider == nil {
			continue
		}
		checker, ok := rp.Provider.(credentialChecker)
		if !ok {
			// A provider without a credential gate is always available, mirroring
			// the observer's own fallback for a token-less provider.
			return true, nil
		}
		avail, err := checker.SCMCredentialsAvailable(ctx)
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		if avail {
			return true, nil
		}
	}
	if firstErr != nil {
		return false, firstErr
	}
	return false, nil
}

func (m *MultiProvider) providerFor(provider string) (Provider, error) {
	if p, ok := m.byName[domain.SCMProvider(provider)]; ok {
		return p, nil
	}
	return nil, fmt.Errorf("scm: no provider registered for %q", provider)
}
