package domain

import "testing"

func TestSCMProviderIsKnown(t *testing.T) {
	known := []SCMProvider{SCMProviderGitHub, SCMProviderGitLab, SCMProviderBitbucket, SCMProviderBitbucketServer}
	for _, p := range known {
		if !p.IsKnown() {
			t.Errorf("IsKnown(%q) = false, want true", p)
		}
	}
	for _, p := range []SCMProvider{"", "linear", "GitHub", "gitea"} {
		if p.IsKnown() {
			t.Errorf("IsKnown(%q) = true, want false", p)
		}
	}
}

func TestSCMProviderFromHost(t *testing.T) {
	cases := []struct {
		host string
		want SCMProvider
		ok   bool
	}{
		{"github.com", SCMProviderGitHub, true},
		{"WWW.GitHub.com", SCMProviderGitHub, true},
		{"api.github.com", SCMProviderGitHub, true},
		{"acme.github.com", SCMProviderGitHub, true},
		{"acme.ghe.io", SCMProviderGitHub, true},
		{"gitlab.com", SCMProviderGitLab, true},
		{"code.gitlab.com", SCMProviderGitLab, true},
		{"bitbucket.org", SCMProviderBitbucket, true},
		{"", "", false},
		{"git.internal.example", "", false},
		{"example.com", "", false},
	}
	for _, tc := range cases {
		got, ok := SCMProviderFromHost(tc.host)
		if got != tc.want || ok != tc.ok {
			t.Errorf("SCMProviderFromHost(%q) = (%q, %v), want (%q, %v)", tc.host, got, ok, tc.want, tc.ok)
		}
	}
}
