package scm

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/tinhtran/thanos/backend/internal/domain"
	"github.com/tinhtran/thanos/backend/internal/ports"
)

// stubProvider is a minimal Provider that tags every result with its own name
// and records the calls it received, so routing can be asserted. It
// deliberately does NOT implement credentialChecker; gatedStub adds that.
type stubProvider struct {
	name   domain.SCMProvider
	prefix string // ParseRepository matches remotes shaped "<prefix>/<name>"

	repoGuardCalls int
	listCalls      int
	commitCalls    int
	logCalls       int
	reviewCalls    int
	fetchBatches   [][]ports.SCMPRRef
}

func (s *stubProvider) ParseRepository(remote string) (ports.SCMRepo, bool) {
	parts := strings.Split(strings.Trim(remote, "/"), "/")
	if len(parts) != 2 || parts[0] != s.prefix {
		return ports.SCMRepo{}, false
	}
	return ports.SCMRepo{Provider: string(s.name), Owner: parts[0], Name: parts[1], Repo: remote}, true
}

func (s *stubProvider) RepoPRListGuard(context.Context, ports.SCMRepo, string) (ports.SCMGuardResult, error) {
	s.repoGuardCalls++
	return ports.SCMGuardResult{ETag: string(s.name)}, nil
}

func (s *stubProvider) ListOpenPRsByRepo(context.Context, ports.SCMRepo) ([]ports.SCMPRObservation, error) {
	s.listCalls++
	return []ports.SCMPRObservation{{URL: string(s.name)}}, nil
}

func (s *stubProvider) CommitChecksGuard(context.Context, ports.SCMRepo, string, string) (ports.SCMGuardResult, error) {
	s.commitCalls++
	return ports.SCMGuardResult{ETag: string(s.name)}, nil
}

func (s *stubProvider) FetchPullRequests(_ context.Context, refs []ports.SCMPRRef) ([]ports.SCMObservation, error) {
	s.fetchBatches = append(s.fetchBatches, append([]ports.SCMPRRef(nil), refs...))
	out := make([]ports.SCMObservation, 0, len(refs))
	for _, ref := range refs {
		out = append(out, ports.SCMObservation{Provider: string(s.name), PR: ports.SCMPRObservation{URL: ref.URL}})
	}
	return out, nil
}

func (s *stubProvider) FetchFailedCheckLogTail(context.Context, ports.SCMRepo, ports.SCMCheckObservation) (string, error) {
	s.logCalls++
	return string(s.name), nil
}

func (s *stubProvider) FetchReviewThreads(context.Context, ports.SCMPRRef) (ports.SCMReviewObservation, error) {
	s.reviewCalls++
	return ports.SCMReviewObservation{Decision: string(s.name)}, nil
}

// gatedStub adds a credential gate to a stubProvider.
type gatedStub struct {
	*stubProvider
	ok  bool
	err error
}

func (g gatedStub) SCMCredentialsAvailable(context.Context) (bool, error) { return g.ok, g.err }

func repoFor(provider domain.SCMProvider) ports.SCMRepo {
	return ports.SCMRepo{Provider: string(provider), Owner: "o", Name: "r", Repo: "o/r"}
}

func TestMultiProviderParseRepositoryFirstMatch(t *testing.T) {
	a := &stubProvider{name: domain.SCMProviderGitHub, prefix: "gh"}
	b := &stubProvider{name: domain.SCMProviderGitLab, prefix: "gl"}
	mp := NewMultiProvider(
		RegisteredProvider{Name: a.name, Provider: a},
		RegisteredProvider{Name: b.name, Provider: b},
	)

	repo, ok := mp.ParseRepository("gh/thing")
	if !ok || repo.Provider != string(domain.SCMProviderGitHub) {
		t.Fatalf("ParseRepository(gh) = %+v, ok=%v", repo, ok)
	}
	repo, ok = mp.ParseRepository("gl/thing")
	if !ok || repo.Provider != string(domain.SCMProviderGitLab) {
		t.Fatalf("ParseRepository(gl) = %+v, ok=%v", repo, ok)
	}
	if _, ok := mp.ParseRepository("bb/thing"); ok {
		t.Fatalf("ParseRepository(bb) unexpectedly matched")
	}
}

func TestMultiProviderRoutesRepoScopedCalls(t *testing.T) {
	gh := &stubProvider{name: domain.SCMProviderGitHub, prefix: "gh"}
	gl := &stubProvider{name: domain.SCMProviderGitLab, prefix: "gl"}
	mp := NewMultiProvider(
		RegisteredProvider{Name: gh.name, Provider: gh},
		RegisteredProvider{Name: gl.name, Provider: gl},
	)
	ctx := context.Background()

	if res, _ := mp.RepoPRListGuard(ctx, repoFor(domain.SCMProviderGitLab), ""); res.ETag != string(domain.SCMProviderGitLab) {
		t.Fatalf("RepoPRListGuard routed to %q, want gitlab", res.ETag)
	}
	if res, _ := mp.CommitChecksGuard(ctx, repoFor(domain.SCMProviderGitHub), "sha", ""); res.ETag != string(domain.SCMProviderGitHub) {
		t.Fatalf("CommitChecksGuard routed to %q, want github", res.ETag)
	}
	if tail, _ := mp.FetchFailedCheckLogTail(ctx, repoFor(domain.SCMProviderGitLab), ports.SCMCheckObservation{}); tail != string(domain.SCMProviderGitLab) {
		t.Fatalf("FetchFailedCheckLogTail routed to %q, want gitlab", tail)
	}
	if rev, _ := mp.FetchReviewThreads(ctx, ports.SCMPRRef{Repo: repoFor(domain.SCMProviderGitHub)}); rev.Decision != string(domain.SCMProviderGitHub) {
		t.Fatalf("FetchReviewThreads routed to %q, want github", rev.Decision)
	}

	if gh.commitCalls != 1 || gh.reviewCalls != 1 || gh.repoGuardCalls != 0 {
		t.Fatalf("github got unexpected calls: %+v", gh)
	}
	if gl.repoGuardCalls != 1 || gl.logCalls != 1 || gl.commitCalls != 0 {
		t.Fatalf("gitlab got unexpected calls: %+v", gl)
	}
}

func TestMultiProviderFetchPullRequestsGroupsByProvider(t *testing.T) {
	gh := &stubProvider{name: domain.SCMProviderGitHub, prefix: "gh"}
	gl := &stubProvider{name: domain.SCMProviderGitLab, prefix: "gl"}
	mp := NewMultiProvider(
		RegisteredProvider{Name: gh.name, Provider: gh},
		RegisteredProvider{Name: gl.name, Provider: gl},
	)

	refs := []ports.SCMPRRef{
		{Repo: repoFor(domain.SCMProviderGitHub), URL: "gh-1"},
		{Repo: repoFor(domain.SCMProviderGitLab), URL: "gl-1"},
		{Repo: repoFor(domain.SCMProviderGitHub), URL: "gh-2"},
	}
	obs, err := mp.FetchPullRequests(context.Background(), refs)
	if err != nil {
		t.Fatalf("FetchPullRequests err: %v", err)
	}
	if len(obs) != 3 {
		t.Fatalf("got %d observations, want 3", len(obs))
	}
	// Each provider saw exactly its own refs in a single batch.
	if len(gh.fetchBatches) != 1 || len(gh.fetchBatches[0]) != 2 {
		t.Fatalf("github batches = %+v, want one batch of 2", gh.fetchBatches)
	}
	if len(gl.fetchBatches) != 1 || len(gl.fetchBatches[0]) != 1 {
		t.Fatalf("gitlab batches = %+v, want one batch of 1", gl.fetchBatches)
	}
	byURL := map[string]string{}
	for _, o := range obs {
		byURL[o.PR.URL] = o.Provider
	}
	if byURL["gh-1"] != string(domain.SCMProviderGitHub) || byURL["gl-1"] != string(domain.SCMProviderGitLab) {
		t.Fatalf("observations mis-attributed: %+v", byURL)
	}
}

func TestMultiProviderUnknownProviderErrors(t *testing.T) {
	gh := &stubProvider{name: domain.SCMProviderGitHub, prefix: "gh"}
	mp := NewMultiProvider(RegisteredProvider{Name: gh.name, Provider: gh})

	if _, err := mp.ListOpenPRsByRepo(context.Background(), repoFor(domain.SCMProviderGitLab)); err == nil {
		t.Fatal("ListOpenPRsByRepo for unregistered provider: want error, got nil")
	}
	if _, err := mp.FetchPullRequests(context.Background(), []ports.SCMPRRef{{Repo: repoFor(domain.SCMProviderBitbucket)}}); err == nil {
		t.Fatal("FetchPullRequests for unregistered provider: want error, got nil")
	}
}

func TestMultiProviderCredentialsAggregate(t *testing.T) {
	sentinel := errors.New("boom")
	ctx := context.Background()

	// Any available wins.
	mp := NewMultiProvider(
		RegisteredProvider{Name: domain.SCMProviderGitHub, Provider: gatedStub{&stubProvider{name: "github"}, false, nil}},
		RegisteredProvider{Name: domain.SCMProviderGitLab, Provider: gatedStub{&stubProvider{name: "gitlab"}, true, nil}},
	)
	if ok, err := mp.SCMCredentialsAvailable(ctx); !ok || err != nil {
		t.Fatalf("any-available: got (%v, %v), want (true, nil)", ok, err)
	}

	// All gated and unavailable -> unavailable, no error.
	mp = NewMultiProvider(
		RegisteredProvider{Name: domain.SCMProviderGitHub, Provider: gatedStub{&stubProvider{name: "github"}, false, nil}},
		RegisteredProvider{Name: domain.SCMProviderGitLab, Provider: gatedStub{&stubProvider{name: "gitlab"}, false, nil}},
	)
	if ok, err := mp.SCMCredentialsAvailable(ctx); ok || err != nil {
		t.Fatalf("all-unavailable: got (%v, %v), want (false, nil)", ok, err)
	}

	// None available, one errored -> surface the error.
	mp = NewMultiProvider(
		RegisteredProvider{Name: domain.SCMProviderGitHub, Provider: gatedStub{&stubProvider{name: "github"}, false, sentinel}},
		RegisteredProvider{Name: domain.SCMProviderGitLab, Provider: gatedStub{&stubProvider{name: "gitlab"}, false, nil}},
	)
	if ok, err := mp.SCMCredentialsAvailable(ctx); ok || !errors.Is(err, sentinel) {
		t.Fatalf("errored: got (%v, %v), want (false, boom)", ok, err)
	}

	// A provider with no credential gate is always available.
	mp = NewMultiProvider(RegisteredProvider{Name: domain.SCMProviderGitHub, Provider: &stubProvider{name: "github"}})
	if ok, err := mp.SCMCredentialsAvailable(ctx); !ok || err != nil {
		t.Fatalf("no-gate: got (%v, %v), want (true, nil)", ok, err)
	}
}
