package skills

import (
	"fmt"
	"time"
)

type Executor struct {
	Now func() time.Time
}

func (e Executor) Start(taskID string, skill Skill, agentSessionID string) (SkillRun, error) {
	if err := ValidateSkill(skill); err != nil {
		return SkillRun{}, err
	}
	if taskID == "" {
		return SkillRun{}, fmt.Errorf("task id is required")
	}
	now := e.now()
	return SkillRun{
		ID:             fmt.Sprintf("%s:%s", taskID, skill.ID),
		TaskID:         taskID,
		SkillID:        skill.ID,
		AgentSessionID: agentSessionID,
		Status:         RunActivated,
		Evidence:       map[string]string{},
		StartedAt:      now,
	}, nil
}

func (e Executor) AddEvidence(run SkillRun, evidence Evidence) (SkillRun, error) {
	if err := ValidateRun(run); err != nil {
		return run, err
	}
	if evidence.TaskID != run.TaskID {
		return run, fmt.Errorf("evidence task %s does not match skill run task %s", evidence.TaskID, run.TaskID)
	}
	if evidence.Type == "" {
		return run, fmt.Errorf("evidence type is required")
	}
	if evidence.Content == "" {
		return run, fmt.Errorf("evidence content is required")
	}
	if run.Evidence == nil {
		run.Evidence = map[string]string{}
	}
	run.Evidence[evidence.Type] = evidence.Content
	run.Status = RunEvidencePending
	return run, nil
}

func (e Executor) Complete(run SkillRun, skill Skill) (SkillRun, error) {
	if err := ValidateRun(run); err != nil {
		return run, err
	}
	if err := ValidateEvidence(skill, run.Evidence); err != nil {
		run.Status = RunEvidencePending
		return run, err
	}
	run.Status = RunCompleted
	run.CompletedAt = e.now()
	return run, nil
}

func (e Executor) Fail(run SkillRun) (SkillRun, error) {
	if err := ValidateRun(run); err != nil {
		return run, err
	}
	run.Status = RunFailed
	run.CompletedAt = e.now()
	return run, nil
}

func (e Executor) now() time.Time {
	if e.Now != nil {
		return e.Now()
	}
	return time.Now().UTC()
}
