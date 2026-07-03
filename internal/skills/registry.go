package skills

import (
	"fmt"
	"slices"
	"strings"
	"time"
)

type Source string

const (
	SourceProject Source = "project"
	SourceGlobal  Source = "global"
	SourceBuiltin Source = "builtin"
)

type RunStatus string

const (
	RunDiscovered      RunStatus = "discovered"
	RunMatched         RunStatus = "matched"
	RunActivated       RunStatus = "activated"
	RunRunning         RunStatus = "running"
	RunEvidencePending RunStatus = "evidence_pending"
	RunCompleted       RunStatus = "completed"
	RunFailed          RunStatus = "failed"
)

type Skill struct {
	ID               string   `json:"id"`
	ProjectID        string   `json:"project_id,omitempty"`
	Name             string   `json:"name"`
	Path             string   `json:"path"`
	Description      string   `json:"description"`
	AppliesTo        []string `json:"applies_to,omitempty"`
	Agents           []string `json:"agents,omitempty"`
	Version          string   `json:"version,omitempty"`
	Source           Source   `json:"source"`
	RequiredEvidence []string `json:"required_evidence,omitempty"`
	ExitCriteria     []string `json:"exit_criteria,omitempty"`
	Body             string   `json:"body,omitempty"`
	Trusted          bool     `json:"trusted"`
}

type Registry struct {
	skills map[string]Skill
	order  []string
}

func NewRegistry(skills ...Skill) (*Registry, error) {
	registry := &Registry{skills: make(map[string]Skill)}
	for _, skill := range skills {
		if err := registry.Add(skill); err != nil {
			return nil, err
		}
	}
	return registry, nil
}

func (r *Registry) Add(skill Skill) error {
	if r.skills == nil {
		r.skills = make(map[string]Skill)
	}
	if err := ValidateSkill(skill); err != nil {
		return err
	}
	key := strings.ToLower(skill.Name)
	existing, exists := r.skills[key]
	if !exists {
		r.order = append(r.order, key)
		r.skills[key] = skill
		return nil
	}
	if sourceRank(skill.Source) < sourceRank(existing.Source) {
		r.skills[key] = skill
	}
	return nil
}

func (r *Registry) Get(name string) (Skill, bool) {
	if r == nil {
		return Skill{}, false
	}
	skill, ok := r.skills[strings.ToLower(name)]
	return skill, ok
}

func (r *Registry) All() []Skill {
	if r == nil {
		return nil
	}
	all := make([]Skill, 0, len(r.order))
	for _, key := range r.order {
		all = append(all, r.skills[key])
	}
	slices.SortFunc(all, func(a, b Skill) int {
		if sourceRank(a.Source) != sourceRank(b.Source) {
			return sourceRank(a.Source) - sourceRank(b.Source)
		}
		return strings.Compare(a.Name, b.Name)
	})
	return all
}

type SkillRun struct {
	ID             string            `json:"id"`
	TaskID         string            `json:"task_id"`
	SkillID        string            `json:"skill_id"`
	AgentSessionID string            `json:"agent_session_id,omitempty"`
	Status         RunStatus         `json:"status"`
	Evidence       map[string]string `json:"evidence_json,omitempty"`
	StartedAt      time.Time         `json:"started_at"`
	CompletedAt    time.Time         `json:"completed_at,omitempty"`
}

type Evidence struct {
	ID         string    `json:"id"`
	TaskID     string    `json:"task_id"`
	SkillRunID string    `json:"skill_run_id,omitempty"`
	Type       string    `json:"type"`
	Content    string    `json:"content"`
	VerifiedBy string    `json:"verified_by,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

func sourceRank(source Source) int {
	switch source {
	case SourceProject:
		return 0
	case SourceGlobal:
		return 1
	case SourceBuiltin:
		return 2
	default:
		return 3
	}
}

func idFor(source Source, name string) string {
	return fmt.Sprintf("%s:%s", source, strings.ToLower(strings.TrimSpace(name)))
}
