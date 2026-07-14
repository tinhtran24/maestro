// Package memorydb owns the derived project-memory projection: a disposable
// SQLite database, one per project, that is a pure function of the project's
// events.jsonl log. It can be dropped and rebuilt at any time, is never
// committed to git, and lives under the app-state cache root (below ~/.maestro),
// wholly separate from the daemon's primary maestro.db.
package memorydb

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/pressly/goose/v3"

	"github.com/tinhtran24/maestro/backend/internal/storage/memorydb/gen"

	// modernc.org/sqlite is the pure-Go (CGO-free) SQLite driver, matching the
	// primary store so the daemon still ships as a static binary.
	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// pragmas match the primary store: WAL + NORMAL for concurrent readers,
// busy_timeout to absorb brief write contention, and foreign_keys ON so the
// task -> file/test cascades are enforced.
const pragmas = "?_pragma=journal_mode(WAL)" +
	"&_pragma=busy_timeout(5000)" +
	"&_pragma=foreign_keys(ON)" +
	"&_pragma=synchronous(NORMAL)"

// dbFileName is the projection database file within a project's cache directory.
const dbFileName = "memory.db"

// Open opens (creating if absent) the memory projection database for a project
// at <cacheDir>/<projectID>/memory.db and returns a Store. cacheDir is the
// app-state cache root the caller derives from config; the resulting database is
// disposable and rebuildable from the event log.
func Open(cacheDir, projectID string) (*Store, error) {
	if cacheDir == "" {
		return nil, errors.New("memorydb: empty cache dir")
	}
	if err := validateProjectID(projectID); err != nil {
		return nil, err
	}
	dbDir := filepath.Join(cacheDir, projectID)
	if err := os.MkdirAll(dbDir, 0o750); err != nil {
		return nil, fmt.Errorf("memorydb: create cache dir: %w", err)
	}
	path := filepath.Join(dbDir, dbFileName)

	db, err := sql.Open("sqlite", "file:"+path+pragmas)
	if err != nil {
		return nil, fmt.Errorf("memorydb: open sqlite: %w", err)
	}
	if err := migrate(db); err != nil {
		_ = db.Close()
		return nil, err
	}
	return &Store{db: db, q: gen.New(db), path: path}, nil
}

// validateProjectID rejects ids that would escape the cache directory. Project
// ids are slugs in this system; this is a defensive guard, not sanitisation.
func validateProjectID(projectID string) error {
	if projectID == "" {
		return errors.New("memorydb: empty project id")
	}
	if projectID == "." || projectID == ".." ||
		strings.ContainsRune(projectID, '/') || strings.ContainsRune(projectID, os.PathSeparator) {
		return fmt.Errorf("memorydb: invalid project id %q", projectID)
	}
	return nil
}

// migrate applies the embedded migrations using a goose Provider bound to this
// db and FS. Unlike goose's package-level SetBaseFS/SetDialect globals, the
// Provider carries no shared state, so opening this database concurrently with
// the primary store cannot race on goose internals.
func migrate(db *sql.DB) error {
	sub, err := fs.Sub(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("memorydb: sub migrations fs: %w", err)
	}
	provider, err := goose.NewProvider(goose.DialectSQLite3, db, sub)
	if err != nil {
		return fmt.Errorf("memorydb: new goose provider: %w", err)
	}
	if _, err := provider.Up(context.Background()); err != nil {
		return fmt.Errorf("memorydb: run migrations: %w", err)
	}
	return nil
}
