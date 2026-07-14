package planner

import (
	"strings"
	"testing"
)

func TestNormalizeBranch(t *testing.T) {
	cases := []struct {
		name      string
		suggested string
		title     string
		labels    []string
		want      string
	}{
		{"valid passthrough", "feature/task-approval", "Add task approval", nil, "feature/task-approval"},
		{"sanitizes slug", "bugfix/Fix Sidebar Width!!", "x", nil, "bugfix/fix-sidebar-width"},
		{"unknown prefix falls back to inference", "wip/stuff", "Fix broken login regression", nil, "bugfix/fix-broken-login-regression"},
		{"derive from title default feature", "", "Add native model selection", nil, "feature/add-native-model-selection"},
		{"label drives prefix", "", "Improve docs page", []string{"documentation"}, "docs/improve-docs-page"},
		{"empty everything", "", "", nil, "feature/task"},
		{"long title truncated", "", "Add a very long descriptive feature name that keeps going and going", nil, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := normalizeBranch(tc.suggested, tc.title, tc.labels)
			if tc.want != "" && got != tc.want {
				t.Fatalf("normalizeBranch = %q, want %q", got, tc.want)
			}
			prefix, slug, ok := splitValidBranch(got)
			if !ok {
				t.Fatalf("normalizeBranch produced invalid branch %q", got)
			}
			if !isKnownPrefix(prefix) {
				t.Fatalf("branch %q has unknown prefix", got)
			}
			if len(slug) > slugMaxLen {
				t.Fatalf("slug %q exceeds max len %d", slug, slugMaxLen)
			}
		})
	}
}

func TestNormalizeCommit(t *testing.T) {
	cases := []struct {
		name      string
		suggested string
		branch    string
		title     string
		wantType  string
		wantExact string
	}{
		{"valid passthrough", "feat(task): add task approval", "feature/x", "t", "feat", "feat(task): add task approval"},
		{"scopeless valid", "fix: preserve sidebar width", "bugfix/x", "t", "fix", "fix: preserve sidebar width"},
		{"invalid type derives from branch", "add: a thing", "refactor/x", "Simplify transition logic", "refactor", ""},
		{"empty derives from bugfix branch", "", "bugfix/login", "Login is broken", "fix", ""},
		{"empty derives from feature branch", "", "feature/models", "Add model selection", "feat", ""},
		{"multiline suggested keeps first line", "feat(ui): add panel\n\nbody text", "feature/x", "t", "feat", "feat(ui): add panel"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := normalizeCommit(tc.suggested, tc.branch, tc.title)
			if tc.wantExact != "" && got != tc.wantExact {
				t.Fatalf("normalizeCommit = %q, want %q", got, tc.wantExact)
			}
			if !strings.HasPrefix(got, tc.wantType+"(") && !strings.HasPrefix(got, tc.wantType+":") {
				t.Fatalf("normalizeCommit = %q, want type %q", got, tc.wantType)
			}
			if _, ok := validCommit(got); !ok {
				t.Fatalf("normalizeCommit produced non-conventional commit %q", got)
			}
		})
	}
}

func TestCommitTypeForBranch(t *testing.T) {
	want := map[string]string{
		"feature/x": "feat", "bugfix/x": "fix", "hotfix/x": "fix",
		"refactor/x": "refactor", "chore/x": "chore", "docs/x": "docs",
		"test/x": "test", "unknown/x": "feat", "": "feat",
	}
	for branch, typ := range want {
		if got := commitTypeForBranch(branch); got != typ {
			t.Errorf("commitTypeForBranch(%q) = %q, want %q", branch, got, typ)
		}
	}
}
