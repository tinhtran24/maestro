package skills

import (
	"fmt"
	"strings"
)

var allowedSources = map[Source]bool{
	SourceProject: true,
	SourceGlobal:  true,
	SourceBuiltin: true,
}

var allowedRunStatuses = map[RunStatus]bool{
	RunDiscovered:      true,
	RunMatched:         true,
	RunActivated:       true,
	RunRunning:         true,
	RunEvidencePending: true,
	RunCompleted:       true,
	RunFailed:          true,
}

func ValidateSkill(skill Skill) error {
	if strings.TrimSpace(skill.Name) == "" {
		return fmt.Errorf("skill name is required")
	}
	if strings.TrimSpace(skill.Description) == "" {
		return fmt.Errorf("skill %s description is required", skill.Name)
	}
	if len(skill.Agents) == 0 {
		return fmt.Errorf("skill %s must declare agents", skill.Name)
	}
	if len(skill.ExitCriteria) == 0 {
		return fmt.Errorf("skill %s must define exit criteria", skill.Name)
	}
	if !allowedSources[skill.Source] {
		return fmt.Errorf("skill %s has invalid source %q", skill.Name, skill.Source)
	}
	return nil
}

func ValidateEvidence(skill Skill, evidence map[string]string) error {
	missing := MissingEvidence(skill, evidence)
	if len(missing) > 0 {
		return fmt.Errorf("skill %s missing required evidence: %s", skill.Name, strings.Join(missing, ", "))
	}
	return nil
}

func MissingEvidence(skill Skill, evidence map[string]string) []string {
	var missing []string
	for _, required := range skill.RequiredEvidence {
		if strings.TrimSpace(evidence[required]) == "" {
			missing = append(missing, required)
		}
	}
	return missing
}

func ValidateRun(run SkillRun) error {
	if strings.TrimSpace(run.ID) == "" {
		return fmt.Errorf("skill run id is required")
	}
	if strings.TrimSpace(run.TaskID) == "" {
		return fmt.Errorf("skill run task id is required")
	}
	if strings.TrimSpace(run.SkillID) == "" {
		return fmt.Errorf("skill run skill id is required")
	}
	if !allowedRunStatuses[run.Status] {
		return fmt.Errorf("skill run %s has invalid status %q", run.ID, run.Status)
	}
	return nil
}
