package app

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

const (
	SchemaVersionTask       = 1
	SchemaVersionRoutine    = 1
	SchemaVersionAutomation = 1
	SchemaVersionEvent      = 1
)

var workspaceLocks sync.Map

type WorkspaceStore struct {
	root string
	mu   *sync.Mutex
}

func NewWorkspaceStore(root string) *WorkspaceStore {
	clean := filepath.Clean(root)
	value, _ := workspaceLocks.LoadOrStore(clean, &sync.Mutex{})
	return &WorkspaceStore{root: clean, mu: value.(*sync.Mutex)}
}

func (s *WorkspaceStore) WithLock(fn func() error) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return fn()
}

func (s *WorkspaceStore) Path(parts ...string) string {
	all := append([]string{s.root, ".thanos"}, parts...)
	return filepath.Join(all...)
}

func (s *WorkspaceStore) WriteJSON(path string, value any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	tmp, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+".*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	cleanup := true
	defer func() {
		if cleanup {
			_ = os.Remove(tmpName)
		}
	}()
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpName, path); err != nil {
		return err
	}
	cleanup = false
	return nil
}

func (s *WorkspaceStore) AppendEvent(event EventInfo) error {
	path := s.Path("events.jsonl")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	event.SchemaVersion = SchemaVersionEvent
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()
	if _, err := file.Write(append(data, '\n')); err != nil {
		return err
	}
	return file.Sync()
}

func (s *WorkspaceStore) MigrateAndValidate() []DiagnosticInfo {
	var diagnostics []DiagnosticInfo
	diagnostics = append(diagnostics, s.migrateTaskRecords()...)
	diagnostics = append(diagnostics, s.validateJSONLines(s.Path("events.jsonl"), "events")...)
	diagnostics = append(diagnostics, s.validateJSONFile(s.Path("routines.json"), "routines")...)
	diagnostics = append(diagnostics, s.validateJSONFile(s.Path("automation.json"), "automation")...)
	return diagnostics
}

func (s *WorkspaceStore) migrateTaskRecords() []DiagnosticInfo {
	var diagnostics []DiagnosticInfo
	dir := s.Path("tasks")
	if stat, err := os.Stat(dir); err != nil || !stat.IsDir() {
		return diagnostics
	}
	_ = filepath.WalkDir(dir, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			diagnostics = append(diagnostics, DiagnosticInfo{Kind: "store", Message: err.Error()})
			return nil
		}
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			return nil
		}
		var raw map[string]any
		data, err := os.ReadFile(path)
		if err != nil {
			diagnostics = append(diagnostics, DiagnosticInfo{Kind: "store", Message: err.Error()})
			return nil
		}
		if err := json.Unmarshal(data, &raw); err != nil {
			diagnostics = append(diagnostics, DiagnosticInfo{Kind: "store", Message: fmt.Sprintf("%s: %v", path, err)})
			return nil
		}
		if _, ok := raw["schema_version"]; !ok {
			raw["schema_version"] = SchemaVersionTask
			if err := s.WriteJSON(path, raw); err != nil {
				diagnostics = append(diagnostics, DiagnosticInfo{Kind: "migration", Message: err.Error()})
			}
		}
		return nil
	})
	return diagnostics
}

func (s *WorkspaceStore) validateJSONFile(path, kind string) []DiagnosticInfo {
	if _, err := os.Stat(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return []DiagnosticInfo{{Kind: kind, Message: err.Error()}}
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return []DiagnosticInfo{{Kind: kind, Message: err.Error()}}
	}
	var raw any
	if err := json.Unmarshal(data, &raw); err != nil {
		return []DiagnosticInfo{{Kind: kind, Message: fmt.Sprintf("%s: %v", filepath.Base(path), err)}}
	}
	return nil
}

func (s *WorkspaceStore) validateJSONLines(path, kind string) []DiagnosticInfo {
	file, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return []DiagnosticInfo{{Kind: kind, Message: err.Error()}}
	}
	defer file.Close()
	var diagnostics []DiagnosticInfo
	scanner := bufio.NewScanner(file)
	line := 0
	for scanner.Scan() {
		line++
		text := strings.TrimSpace(scanner.Text())
		if text == "" {
			continue
		}
		var raw any
		if err := json.Unmarshal([]byte(text), &raw); err != nil {
			diagnostics = append(diagnostics, DiagnosticInfo{Kind: kind, Message: fmt.Sprintf("%s:%d: %v", filepath.Base(path), line, err)})
		}
	}
	if err := scanner.Err(); err != nil {
		diagnostics = append(diagnostics, DiagnosticInfo{Kind: kind, Message: err.Error()})
	}
	return diagnostics
}
