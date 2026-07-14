package planner

import (
	"regexp"
	"strings"
)

// branchPrefixes is the allowed branch-name taxonomy. A suggested branch must
// start with one of these, followed by "/<short-kebab-description>".
var branchPrefixes = []string{"feature", "bugfix", "hotfix", "refactor", "chore", "docs", "test"}

// prefixToCommitType maps a branch prefix onto its Conventional Commit type.
// feature→feat and bugfix/hotfix→fix; the rest keep their own name.
var prefixToCommitType = map[string]string{
	"feature":  "feat",
	"bugfix":   "fix",
	"hotfix":   "fix",
	"refactor": "refactor",
	"chore":    "chore",
	"docs":     "docs",
	"test":     "test",
}

// commitTypes is the Conventional Commit type vocabulary accepted in a suggested
// commit's leading `type` token (mirrors the repo delivery protocol).
var commitTypes = map[string]bool{
	"feat": true, "fix": true, "refactor": true, "docs": true,
	"test": true, "chore": true, "build": true, "ci": true, "perf": true,
}

// slugMaxLen bounds the kebab description in a suggested branch so a long title
// does not produce an unwieldy ref.
const slugMaxLen = 40

var (
	nonSlugRun    = regexp.MustCompile(`[^a-z0-9]+`)
	commitSubject = regexp.MustCompile(`^([a-z]+)(\([^)]*\))?(!)?:\s+(.+)$`)
)

// normalizeBranch returns a valid `<prefix>/<slug>` branch name. It trusts a
// model-supplied value only when it already matches the taxonomy; otherwise it
// derives the branch from labels/title, defaulting the prefix to "feature".
func normalizeBranch(suggested, title string, labels []string) string {
	if prefix, slug, ok := splitValidBranch(suggested); ok {
		return prefix + "/" + slug
	}
	// The suggestion did not fit the taxonomy, so its description came with a
	// rejected format — derive the slug from the title instead of trusting it.
	prefix := inferPrefix(suggested, title, labels)
	slug := slugify(title)
	if slug == "" {
		slug = "task"
	}
	return prefix + "/" + slug
}

// normalizeCommit returns a Conventional Commit `type(scope): summary` line,
// keeping a valid model-supplied value and otherwise deriving one from the
// branch prefix and title.
func normalizeCommit(suggested, branch, title string) string {
	if c, ok := validCommit(suggested); ok {
		return c
	}
	typ := commitTypeForBranch(branch)
	summary := firstNonEmpty(commitSummary(suggested), imperativeSummary(title))
	if summary == "" {
		summary = "implement task"
	}
	return typ + ": " + summary
}

func splitValidBranch(s string) (prefix, slug string, ok bool) {
	s = strings.TrimSpace(s)
	i := strings.Index(s, "/")
	if i <= 0 {
		return "", "", false
	}
	prefix = strings.ToLower(s[:i])
	if !isKnownPrefix(prefix) {
		return "", "", false
	}
	slug = slugify(s[i+1:])
	if slug == "" {
		return "", "", false
	}
	return prefix, slug, true
}

func validCommit(s string) (string, bool) {
	s = strings.TrimSpace(firstLine(s))
	m := commitSubject.FindStringSubmatch(s)
	if m == nil {
		return "", false
	}
	if !commitTypes[m[1]] {
		return "", false
	}
	return s, true
}

// inferPrefix picks a branch prefix from a model hint, the labels, or the title.
func inferPrefix(suggested, title string, labels []string) string {
	if p := leadingPrefix(suggested); p != "" {
		return p
	}
	hay := strings.ToLower(strings.Join(labels, " ") + " " + title)
	switch {
	case containsAny(hay, "hotfix", "urgent", "outage"):
		return "hotfix"
	case containsAny(hay, "bug", "fix", "regression", "broken"):
		return "bugfix"
	case containsAny(hay, "refactor", "cleanup", "simplify"):
		return "refactor"
	case containsAny(hay, "doc", "readme"):
		return "docs"
	case containsAny(hay, "test", "coverage"):
		return "test"
	case containsAny(hay, "chore", "bump", "dependency", "dependencies"):
		return "chore"
	default:
		return "feature"
	}
}

func commitTypeForBranch(branch string) string {
	if t, ok := prefixToCommitType[leadingPrefix(branch)]; ok {
		return t
	}
	return "feat"
}

func leadingPrefix(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	if i := strings.Index(s, "/"); i > 0 {
		s = s[:i]
	}
	if isKnownPrefix(s) {
		return s
	}
	return ""
}

func commitSummary(s string) string {
	if m := commitSubject.FindStringSubmatch(strings.TrimSpace(firstLine(s))); m != nil {
		return strings.TrimSpace(m[4])
	}
	return ""
}

// imperativeSummary lowercases the first word of the title so the summary reads
// as an imperative clause, keeping it short.
func imperativeSummary(title string) string {
	t := strings.TrimSpace(strings.TrimRight(title, "."))
	if t == "" {
		return ""
	}
	words := strings.Fields(t)
	words[0] = strings.ToLower(words[0])
	out := strings.Join(words, " ")
	if len(out) > 60 {
		out = strings.TrimSpace(out[:60])
	}
	return out
}

func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = nonSlugRun.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if len(s) > slugMaxLen {
		s = strings.Trim(s[:slugMaxLen], "-")
	}
	return s
}

func isKnownPrefix(p string) bool {
	for _, k := range branchPrefixes {
		if p == k {
			return true
		}
	}
	return false
}

func containsAny(hay string, needles ...string) bool {
	for _, n := range needles {
		if strings.Contains(hay, n) {
			return true
		}
	}
	return false
}

func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i]
	}
	return s
}

func firstNonEmpty(a, b string) string {
	if strings.TrimSpace(a) != "" {
		return a
	}
	return b
}
