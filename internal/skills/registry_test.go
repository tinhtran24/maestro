package skills

import "testing"

func TestRegistryProjectSkillOverridesGlobalSkill(t *testing.T) {
	global := testSkill("feature-spec", SourceGlobal, "global")
	project := testSkill("feature-spec", SourceProject, "project")

	registry, err := NewRegistry(global, project)
	if err != nil {
		t.Fatal(err)
	}
	got, ok := registry.Get("feature-spec")
	if !ok {
		t.Fatal("skill not found")
	}
	if got.Source != SourceProject || got.Description != "project" {
		t.Fatalf("project skill did not override global skill: %#v", got)
	}
}

func TestMatchSkillsUsesAgentStageAndFileSignals(t *testing.T) {
	registry, err := NewRegistry(
		Skill{
			ID:               "project:frontend-flow-component",
			Name:             "frontend-flow-component",
			Description:      "Keep frontend work in flow component boundaries.",
			AppliesTo:        []string{"frontend", "tsx", "planning"},
			Agents:           []string{"coder"},
			Source:           SourceProject,
			RequiredEvidence: []string{"flow_files"},
			ExitCriteria:     []string{"Screens compose flows only."},
			Trusted:          true,
		},
		testSkill("testing", SourceProject, "testing"),
	)
	if err != nil {
		t.Fatal(err)
	}

	matches := MatchSkills(registry, MatchContext{
		TaskType:      "frontend",
		WorkflowStage: "planning",
		AgentRole:     "coder",
		Files:         []string{"app/frontend/src/flows/TaskWorkbenchFlow.tsx"},
	})
	if len(matches) == 0 {
		t.Fatal("expected matches")
	}
	if got, want := matches[0].Skill.Name, "frontend-flow-component"; got != want {
		t.Fatalf("first match = %s, want %s", got, want)
	}
}

func testSkill(name string, source Source, description string) Skill {
	return Skill{
		ID:               idFor(source, name),
		Name:             name,
		Description:      description,
		AppliesTo:        []string{"planning"},
		Agents:           []string{"planner"},
		Source:           source,
		RequiredEvidence: []string{"notes"},
		ExitCriteria:     []string{"Notes exist."},
		Trusted:          source != SourceGlobal,
	}
}
