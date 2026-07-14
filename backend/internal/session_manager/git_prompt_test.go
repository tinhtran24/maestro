package sessionmanager

import (
	"strings"
	"testing"

	"github.com/tinhtran/thanos/backend/internal/domain"
)

func TestGitWorkflowPromptProvider(t *testing.T) {
	const commitMarker = "git push -u origin HEAD"

	cases := []struct {
		name     string
		provider domain.SCMProvider
		wantOpen string
	}{
		{"empty defaults to github", "", "gh pr create"},
		{"github", domain.SCMProviderGitHub, "gh pr create"},
		{"gitlab", domain.SCMProviderGitLab, "glab mr create"},
		{"bitbucket", domain.SCMProviderBitbucket, "Bitbucket web UI"},
		{"bitbucket server", domain.SCMProviderBitbucketServer, "Bitbucket web UI"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := gitWorkflowPrompt(tc.provider)
			if !strings.Contains(got, commitMarker) {
				t.Errorf("prompt missing shared commit/push step:\n%s", got)
			}
			if !strings.Contains(got, tc.wantOpen) {
				t.Errorf("prompt for %q missing %q:\n%s", tc.provider, tc.wantOpen, got)
			}
			// Every provider's prompt sequences Development then Testing.
			if !strings.Contains(got, "Development") || !strings.Contains(got, "Testing") {
				t.Errorf("prompt for %q missing Development/Testing lifecycle:\n%s", tc.provider, got)
			}
		})
	}

	// gitlab must not fall back to the github command.
	if strings.Contains(gitWorkflowPrompt(domain.SCMProviderGitLab), "gh pr create") {
		t.Error("gitlab prompt leaked the github command")
	}
}
