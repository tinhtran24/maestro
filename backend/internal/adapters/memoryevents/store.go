// Package memoryevents reads and writes a project's append-only memory log,
// the events.jsonl file that is the source of truth for project memory. The log
// lives inside the managed user repo at <projectPath>/.maestro/memory/events.jsonl
// (committed with that repo), NOT under the ~/.maestro app-state directory. Each
// completed task appends exactly one immutable JSON line; nothing here edits or
// deletes a prior line.
package memoryevents

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
)

const (
	memoryDirName  = ".maestro"
	memorySubdir   = "memory"
	eventsFileName = "events.jsonl"
)

// Store appends to and iterates over per-project events.jsonl logs. A single
// process-wide mutex serialises appends so concurrent completions cannot
// interleave a partial line; O_APPEND plus fsync makes each write durable.
type Store struct {
	mu sync.Mutex
}

// New returns a ready Store.
func New() *Store { return &Store{} }

// EventsPath returns the absolute events.jsonl path for a project checkout.
func EventsPath(projectPath string) string {
	return filepath.Join(projectPath, RelEventsPath())
}

// RelEventsPath returns the events.jsonl path relative to the project checkout
// root, for use as a git pathspec.
func RelEventsPath() string {
	return filepath.Join(memoryDirName, memorySubdir, eventsFileName)
}

// Append writes ev as a single JSON line to the project's events.jsonl, creating
// the .maestro/memory directory if needed. The write is O_APPEND and fsynced so a
// crash leaves either a whole line or nothing after it. ev.ID must be set by the
// caller; Append does not mint identifiers or timestamps.
func (s *Store) Append(ctx context.Context, projectPath string, ev Event) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if projectPath == "" {
		return errors.New("memoryevents: empty project path")
	}
	if ev.ID == "" {
		return errors.New("memoryevents: event id is required")
	}
	line, err := json.Marshal(ev)
	if err != nil {
		return fmt.Errorf("memoryevents: marshal event: %w", err)
	}
	// json.Marshal escapes control characters, so line never contains a raw
	// newline; one event is always exactly one line.
	line = append(line, '\n')

	path := EventsPath(projectPath)

	s.mu.Lock()
	defer s.mu.Unlock()

	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return fmt.Errorf("memoryevents: create memory dir: %w", err)
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("memoryevents: open events log: %w", err)
	}
	defer func() { _ = f.Close() }()
	if _, err := f.Write(line); err != nil {
		return fmt.Errorf("memoryevents: append event: %w", err)
	}
	if err := f.Sync(); err != nil {
		return fmt.Errorf("memoryevents: fsync events log: %w", err)
	}
	return nil
}

// Iterate reads the project's events.jsonl in append order and calls fn for each
// event. When after is non-empty, events up to and including the event whose ID
// equals after are skipped, so callers can resume from a recorded offset; an
// after that is never found yields nothing. A missing log is not an error (no
// memory yet). A malformed line is surfaced as an error and stops iteration
// rather than being silently skipped, so callers decide how to treat a truncated
// tail. If fn returns an error, iteration stops and returns it.
func (s *Store) Iterate(ctx context.Context, projectPath, after string, fn func(Event) error) error {
	if projectPath == "" {
		return errors.New("memoryevents: empty project path")
	}
	path := EventsPath(projectPath)
	f, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("memoryevents: open events log: %w", err)
	}
	defer func() { _ = f.Close() }()

	r := bufio.NewReader(f)
	resuming := after != ""
	lineNo := 0
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		line, readErr := r.ReadBytes('\n')
		trimmed := trimLine(line)
		if len(trimmed) > 0 {
			lineNo++
			var ev Event
			if err := json.Unmarshal(trimmed, &ev); err != nil {
				return fmt.Errorf("memoryevents: parse %s line %d: %w", path, lineNo, err)
			}
			if resuming {
				// Still seeking the resume point; consume this event silently and
				// only start yielding once we pass the matching id.
				if ev.ID == after {
					resuming = false
				}
			} else if err := fn(ev); err != nil {
				return err
			}
		}
		if readErr != nil {
			if errors.Is(readErr, io.EOF) {
				return nil
			}
			return fmt.Errorf("memoryevents: read events log: %w", readErr)
		}
	}
}

// trimLine drops a trailing CR/LF pair and surrounding ASCII whitespace so a
// final newline (or a blank line) reads as empty and is skipped, while a partial
// JSON write survives as non-empty content that fails to parse.
func trimLine(b []byte) []byte {
	for len(b) > 0 {
		switch b[len(b)-1] {
		case '\n', '\r', ' ', '\t':
			b = b[:len(b)-1]
		default:
			return b
		}
	}
	return b
}
