package projects

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Repository struct {
	Root string
}

func (r Repository) Save(project Project) error {
	path := filepath.Join(r.Root, ".thanos", "config.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(map[string]any{"project": project}, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0o644)
}
