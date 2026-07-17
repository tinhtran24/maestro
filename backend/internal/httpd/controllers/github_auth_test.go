package controllers

import (
	"context"
	"errors"
	"testing"
)

type fakeGitHubAuthRunner struct {
	path   string
	token  string
	stderr string
	err    error
}

func (f fakeGitHubAuthRunner) LookPath(string) (string, error) {
	if f.path == "" {
		return "", errors.New("missing")
	}
	return f.path, nil
}

func (f fakeGitHubAuthRunner) Token(context.Context) (string, string, error) {
	return f.token, f.stderr, f.err
}

func TestProbeGitHubAuth(t *testing.T) {
	t.Setenv("MAESTRO_GITHUB_TOKEN", "")
	t.Setenv("GITHUB_TOKEN", "")

	cases := []struct {
		name          string
		runner        fakeGitHubAuthRunner
		wantInstalled bool
		wantAuth      bool
		wantSource    string
		wantMessage   string
	}{
		{
			name:          "missing gh",
			runner:        fakeGitHubAuthRunner{},
			wantInstalled: false,
			wantAuth:      false,
			wantSource:    "none",
			wantMessage:   "GitHub CLI is not installed.",
		},
		{
			name:          "gh token failure",
			runner:        fakeGitHubAuthRunner{path: "/usr/bin/gh", stderr: "run gh auth login", err: errors.New("exit 1")},
			wantInstalled: true,
			wantAuth:      false,
			wantSource:    "gh",
			wantMessage:   "run gh auth login",
		},
		{
			name:          "gh token ok",
			runner:        fakeGitHubAuthRunner{path: "/usr/bin/gh", token: "token\n"},
			wantInstalled: true,
			wantAuth:      true,
			wantSource:    "gh",
			wantMessage:   "GitHub CLI authentication is available.",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := probeGitHubAuth(context.Background(), tc.runner)
			if got.Installed != tc.wantInstalled || got.Authenticated != tc.wantAuth || got.Source != tc.wantSource || got.Message != tc.wantMessage {
				t.Fatalf("status = %#v", got)
			}
		})
	}
}

func TestProbeGitHubAuthUsesEnvToken(t *testing.T) {
	t.Setenv("MAESTRO_GITHUB_TOKEN", "token")
	t.Setenv("GITHUB_TOKEN", "")
	got := probeGitHubAuth(context.Background(), fakeGitHubAuthRunner{})
	if !got.Authenticated || got.Source != "MAESTRO_GITHUB_TOKEN" {
		t.Fatalf("status = %#v", got)
	}
}
