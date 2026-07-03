package orchestrator

import (
	"strings"
	"testing"

	"github.com/tinhtran/thanos/internal/workbench"
)

func TestWorkbenchTaskSkillGateRequiresCompletedPlanningSkillBeforeReady(t *testing.T) {
	task := workbench.Task{ID: "T-1", Status: workbench.TaskWaitingApproval, ReviewApproved: true}
	skill := workbench.Skill{ID: "project:feature-spec", AppliesTo: []string{"planning"}, Enabled: true}
	run := workbench.SkillRun{
		ID:       "SR-1",
		TaskID:   task.ID,
		SkillID:  skill.ID,
		Status:   workbench.SkillRunEvidencePending,
		Evidence: map[string]any{"acceptance_criteria": "exists"},
	}

	err := CanMoveWorkbenchTaskWithSkills(task, workbench.TaskReady, []workbench.Skill{skill}, []workbench.SkillRun{run})
	if err == nil || !strings.Contains(err.Error(), "feature-spec") {
		t.Fatalf("expected incomplete skill gate, got %v", err)
	}

	run.Status = workbench.SkillRunCompleted
	if err := CanMoveWorkbenchTaskWithSkills(task, workbench.TaskReady, []workbench.Skill{skill}, []workbench.SkillRun{run}); err != nil {
		t.Fatal(err)
	}
}

func TestWorkbenchTaskSkillGateRequiresEvidenceBeforeDone(t *testing.T) {
	task := workbench.Task{ID: "T-1", Status: workbench.TaskInReview, ReviewApproved: true, TestsPassed: true}
	skill := workbench.Skill{ID: "project:testing", AppliesTo: []string{"testing"}, Enabled: true}
	run := workbench.SkillRun{ID: "SR-1", TaskID: task.ID, SkillID: skill.ID, Status: workbench.SkillRunCompleted}

	err := CanMoveWorkbenchTaskWithSkills(task, workbench.TaskDone, []workbench.Skill{skill}, []workbench.SkillRun{run})
	if err == nil || !strings.Contains(err.Error(), "has evidence") {
		t.Fatalf("expected evidence gate, got %v", err)
	}
}
