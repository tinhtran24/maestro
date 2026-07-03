package skills

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

var exitCriteriaPattern = regexp.MustCompile(`(?mi)^##\s+Exit criteria\s*$`)

type frontmatter struct {
	Name             string   `yaml:"name"`
	Description      string   `yaml:"description"`
	AppliesTo        []string `yaml:"applies_to"`
	Agents           []string `yaml:"agents"`
	Version          string   `yaml:"version"`
	RequiredEvidence []string `yaml:"required_evidence"`
}

func LoadProjectSkills(root string) ([]Skill, error) {
	return LoadSkills(filepath.Join(root, ".thanos", "skills"), SourceProject, "")
}

func LoadSkills(dir string, source Source, projectID string) ([]Skill, error) {
	var loaded []Skill
	err := filepath.WalkDir(dir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || entry.Name() != "SKILL.md" {
			return nil
		}
		skill, err := LoadSkillFile(path, source, projectID)
		if err != nil {
			return err
		}
		loaded = append(loaded, skill)
		return nil
	})
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	return loaded, nil
}

func LoadSkillFile(path string, source Source, projectID string) (Skill, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return Skill{}, err
	}
	meta, body, err := parseSkillMarkdown(content)
	if err != nil {
		return Skill{}, fmt.Errorf("%s: %w", path, err)
	}
	skill := Skill{
		ID:               idFor(source, meta.Name),
		ProjectID:        projectID,
		Name:             strings.TrimSpace(meta.Name),
		Path:             path,
		Description:      strings.TrimSpace(meta.Description),
		AppliesTo:        compact(meta.AppliesTo),
		Agents:           compact(meta.Agents),
		Version:          strings.TrimSpace(meta.Version),
		Source:           source,
		RequiredEvidence: compact(meta.RequiredEvidence),
		ExitCriteria:     extractExitCriteria(body),
		Body:             strings.TrimSpace(body),
		Trusted:          source == SourceProject || source == SourceBuiltin,
	}
	return skill, ValidateSkill(skill)
}

func parseSkillMarkdown(content []byte) (frontmatter, string, error) {
	if !bytes.HasPrefix(content, []byte("---\n")) {
		return frontmatter{}, "", fmt.Errorf("missing YAML frontmatter")
	}
	rest := content[len("---\n"):]
	end := bytes.Index(rest, []byte("\n---"))
	if end < 0 {
		return frontmatter{}, "", fmt.Errorf("missing frontmatter terminator")
	}
	rawMeta := rest[:end]
	body := string(bytes.TrimPrefix(rest[end+len("\n---"):], []byte("\n")))
	var meta frontmatter
	if err := yaml.Unmarshal(rawMeta, &meta); err != nil {
		return frontmatter{}, "", err
	}
	return meta, body, nil
}

func extractExitCriteria(body string) []string {
	match := exitCriteriaPattern.FindStringIndex(body)
	if match == nil {
		return nil
	}
	section := body[match[1]:]
	if next := regexp.MustCompile(`(?m)^##\s+`).FindStringIndex(section); next != nil {
		section = section[:next[0]]
	}
	var criteria []string
	for _, line := range strings.Split(section, "\n") {
		line = strings.TrimSpace(line)
		line = strings.TrimPrefix(line, "- ")
		line = strings.TrimSpace(line)
		if line != "" {
			criteria = append(criteria, line)
		}
	}
	return criteria
}

func compact(values []string) []string {
	result := make([]string, 0, len(values))
	seen := make(map[string]bool, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		result = append(result, value)
	}
	return result
}
