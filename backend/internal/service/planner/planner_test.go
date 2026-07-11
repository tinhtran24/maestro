package planner

import (
	"context"
	"strings"
	"testing"
)

type fakeRunner struct {
	reply     string
	err       error
	gotAgent  string
	gotPrompt string
	available bool
}

func (f *fakeRunner) Run(_ context.Context, agent, prompt string) (string, error) {
	f.gotAgent = agent
	f.gotPrompt = prompt
	return f.reply, f.err
}
func (f *fakeRunner) Available(string) bool { return f.available }

func TestPlan_ParsesStructuredJSON(t *testing.T) {
	fr := &fakeRunner{available: true, reply: `Sure, here is the task:
` + "```json" + `
{"title":"Implement Shopping Cart","priority":"P1","labels":["frontend"],
"acceptanceCriteria":["Add product","Checkout"],"plan":["Design","Build"],
"missingInformation":["Payment provider?"],
"confidence":{"overall":95,"title":98,"priority":90,"acceptanceCriteria":88}}
` + "```"}
	s := New(Options{Runner: fr})

	draft, err := s.Plan(context.Background(), "build a cart", []string{"mock.png"}, "claude")
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}
	if draft.Title != "Implement Shopping Cart" {
		t.Fatalf("title = %q", draft.Title)
	}
	if draft.Priority != "P1" {
		t.Fatalf("priority = %q", draft.Priority)
	}
	if len(draft.AcceptanceCriteria) != 2 || draft.Confidence.Overall != 95 {
		t.Fatalf("criteria/confidence = %+v", draft)
	}
	if fr.gotAgent != "claude" {
		t.Fatalf("agent passed to runner = %q", fr.gotAgent)
	}
	if !strings.Contains(fr.gotPrompt, "mock.png") {
		t.Fatalf("attachment not included in prompt")
	}
}

func TestPlan_NormalizesBadPriorityAndClampsConfidence(t *testing.T) {
	fr := &fakeRunner{available: true, reply: `{"title":"X","priority":"urgent","confidence":{"overall":150,"title":-5}}`}
	s := New(Options{Runner: fr})
	draft, err := s.Plan(context.Background(), "x", nil, "claude")
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}
	if draft.Priority != "P2" {
		t.Fatalf("priority not normalized: %q", draft.Priority)
	}
	if draft.Confidence.Overall != 100 || draft.Confidence.Title != 0 {
		t.Fatalf("confidence not clamped: %+v", draft.Confidence)
	}
}

func TestPlan_EmptyInputRejected(t *testing.T) {
	s := New(Options{Runner: &fakeRunner{available: true}})
	if _, err := s.Plan(context.Background(), "   ", nil, "claude"); err == nil {
		t.Fatal("expected error for empty input")
	}
}

func TestPlan_NoJSONInReply(t *testing.T) {
	s := New(Options{Runner: &fakeRunner{available: true, reply: "I cannot help with that."}})
	if _, err := s.Plan(context.Background(), "x", nil, "claude"); err == nil {
		t.Fatal("expected error when reply has no JSON object")
	}
}
