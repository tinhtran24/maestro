package domain

import "strings"

// SCMProvider identifies a source-control host implementation. It is the
// provider-neutral vocabulary shared by the SCM observation adapters, the
// project Git-workflow config, and the agent completion prompt. Adding a value
// here is a port-level decision: every provider-dispatching seam
// (observe/scm.MultiProvider, the auto-PR prompt, config validation) must be
// able to route it, even when the concrete adapter lands in a later change.
type SCMProvider string

// The known source-control providers. github is the only one with a shipped
// SCM observation adapter today; the others are recognised by the config and
// agent-prompt seams so provider-aware auto-PR wording is available before the
// matching observation adapter exists.
const (
	SCMProviderGitHub          SCMProvider = "github"
	SCMProviderGitLab          SCMProvider = "gitlab"
	SCMProviderBitbucket       SCMProvider = "bitbucket"
	SCMProviderBitbucketServer SCMProvider = "bitbucket-server"
)

// IsKnown reports whether p is one of the recognised SCM providers.
func (p SCMProvider) IsKnown() bool {
	switch p {
	case SCMProviderGitHub, SCMProviderGitLab, SCMProviderBitbucket, SCMProviderBitbucketServer:
		return true
	default:
		return false
	}
}

// SCMProviderFromHost maps a git remote host to its provider, or returns an
// empty provider (and false) when the host is not recognised. It only knows the
// well-known cloud hosts; self-hosted GitLab/Bitbucket Server instances live on
// arbitrary hosts and must be set explicitly in project config. GitHub
// Enterprise hosts under *.github.com / *.ghe.io are treated as github.
func SCMProviderFromHost(host string) (SCMProvider, bool) {
	h := strings.ToLower(strings.TrimSpace(host))
	switch {
	case h == "":
		return "", false
	case h == "github.com" || h == "www.github.com" || h == "api.github.com" ||
		strings.HasSuffix(h, ".github.com") || strings.HasSuffix(h, ".ghe.io"):
		return SCMProviderGitHub, true
	case h == "gitlab.com" || strings.HasSuffix(h, ".gitlab.com"):
		return SCMProviderGitLab, true
	case h == "bitbucket.org" || strings.HasSuffix(h, ".bitbucket.org"):
		return SCMProviderBitbucket, true
	default:
		return "", false
	}
}
