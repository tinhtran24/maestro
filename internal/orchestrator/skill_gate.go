package orchestrator

import (
	"fmt"
	"strings"

	"github.com/tinhtran/thanos/internal/workbench"
)

func CanMoveWorkbenchTaskWithSkills(task workbench.Task, to workbench.TaskStatus, skills []workbench.Skill, runs []workbench.SkillRun) error {
	if err := CanMoveWorkbenchTask(task, to); err != nil {
		return err
	}
	required := requiredSkillRunsForTransition(task, to, skills, runs)
	for _, run := range required {
		if run.Status != workbench.SkillRunCompleted {
			return fmt.Errorf("task %s cannot transition %s -> %s before skill %s is completed", task.ID, task.Status, to, run.SkillID)
		}
		if len(run.Evidence) == 0 {
			return fmt.Errorf("task %s cannot transition %s -> %s before skill %s has evidence", task.ID, task.Status, to, run.SkillID)
		}
	}
	return nil
}

func MoveWorkbenchTaskWithSkills(task workbench.Task, to workbench.TaskStatus, reason string, skills []workbench.Skill, runs []workbench.SkillRun) (workbench.Task, WorkbenchTransition, error) {
	if err := CanMoveWorkbenchTaskWithSkills(task, to, skills, runs); err != nil {
		return task, WorkbenchTransition{}, err
	}
	return MoveWorkbenchTask(task, to, reason)
}

func requiredSkillRunsForTransition(task workbench.Task, to workbench.TaskStatus, skills []workbench.Skill, runs []workbench.SkillRun) []workbench.SkillRun {
	byID := make(map[string]workbench.Skill, len(skills))
	for _, skill := range skills {
		if skill.Enabled {
			byID[skill.ID] = skill
		}
	}
	var required []workbench.SkillRun
	for _, run := range runs {
		if run.TaskID != task.ID {
			continue
		}
		skill, ok := byID[run.SkillID]
		if !ok {
			continue
		}
		if transitionNeedsSkill(task, to, skill) {
			required = append(required, run)
		}
	}
	return required
}

func transitionNeedsSkill(task workbench.Task, to workbench.TaskStatus, skill workbench.Skill) bool {
	if to == workbench.TaskReady && task.Status == workbench.TaskWaitingApproval {
		return hasAny(skill.AppliesTo, "planning", "feature")
	}
	if to == workbench.TaskDone {
		return hasAny(skill.AppliesTo, "review", "testing", "done")
	}
	return false
}

func hasAny(values []string, targets ...string) bool {
	for _, value := range values {
		for _, target := range targets {
			if strings.EqualFold(strings.TrimSpace(value), target) {
				return true
			}
		}
	}
	return false
}
