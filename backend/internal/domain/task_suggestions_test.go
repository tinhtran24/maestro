package domain

import "testing"

func TestSuggestTaskForCredentialDiagnostics(t *testing.T) {
	got := SuggestTask("Fix issue: tracker intake credential diagnostics", "")
	if got.Branch != "bugfix/tracker-intake-credential-diagnostics" {
		t.Fatalf("branch = %q", got.Branch)
	}
	if got.CommitMessage != "fix(tracker): tracker intake credential diagnostics" {
		t.Fatalf("commit = %q", got.CommitMessage)
	}
	if got.PRTitle != got.CommitMessage {
		t.Fatalf("pr title = %q, want commit", got.PRTitle)
	}
}
