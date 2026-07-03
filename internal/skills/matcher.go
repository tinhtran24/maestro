package skills

import (
	"path/filepath"
	"slices"
	"strings"
)

type MatchContext struct {
	TaskType      string
	WorkflowStage string
	AgentRole     string
	Files         []string
	Tags          []string
}

type Match struct {
	Skill  Skill  `json:"skill"`
	Reason string `json:"reason"`
	Score  int    `json:"score"`
}

func MatchSkills(registry *Registry, ctx MatchContext) []Match {
	if registry == nil {
		return nil
	}
	var matches []Match
	for _, skill := range registry.All() {
		score, reason := scoreSkill(skill, ctx)
		if score > 0 {
			matches = append(matches, Match{Skill: skill, Reason: reason, Score: score})
		}
	}
	slices.SortFunc(matches, func(a, b Match) int {
		if a.Score != b.Score {
			return b.Score - a.Score
		}
		return strings.Compare(a.Skill.Name, b.Skill.Name)
	})
	return matches
}

func scoreSkill(skill Skill, ctx MatchContext) (int, string) {
	score := 0
	reasons := make([]string, 0, 4)
	if containsFold(skill.Agents, ctx.AgentRole) {
		score += 4
		reasons = append(reasons, "agent")
	}
	for _, applies := range skill.AppliesTo {
		if equalFold(applies, ctx.WorkflowStage) || equalFold(applies, ctx.TaskType) || containsFold(ctx.Tags, applies) {
			score += 3
			reasons = append(reasons, "stage")
			break
		}
	}
	for _, file := range ctx.Files {
		base := strings.ToLower(filepath.Base(file))
		for _, applies := range skill.AppliesTo {
			token := strings.ToLower(applies)
			if token != "" && strings.Contains(base, token) {
				score++
				reasons = append(reasons, "file")
				break
			}
		}
	}
	if score == 0 && containsFold(skill.AppliesTo, "always") {
		score = 1
		reasons = append(reasons, "always")
	}
	return score, strings.Join(unique(reasons), ",")
}

func containsFold(values []string, target string) bool {
	for _, value := range values {
		if equalFold(value, target) {
			return true
		}
	}
	return false
}

func equalFold(a, b string) bool {
	return strings.EqualFold(strings.TrimSpace(a), strings.TrimSpace(b))
}

func unique(values []string) []string {
	seen := make(map[string]bool, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	return out
}
