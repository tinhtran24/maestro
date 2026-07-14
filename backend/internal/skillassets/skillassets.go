// Package skillassets embeds the using-maestro skill (the to CLI catalog) and
// installs it into the Maestro data dir at daemon boot. Worker sessions run in a
// worktree of whatever project they were spawned in, so a repo-relative
// skills/ path only resolves when that project happens to be the Maestro repo
// itself. Installing under the data dir gives every session, in any project, a
// stable absolute path to read.
//
// The embedded copy is the single source of truth. Install clobbers the
// on-disk copy on every boot, so a new daemon build always refreshes it and the
// two can never drift; there is no version marker or hash to keep in sync
// because the daemon binary already is the version.
package skillassets

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"embed"
)

//go:embed using-maestro dev-lifecycle
var files embed.FS

// SkillName is the using-maestro skill's directory name under <dataDir>/skills. It
// stays the exported default because the to-CLI prompt pointer cites Dir().
const SkillName = "using-maestro"

// LifecycleSkillName is the dev-lifecycle skill's directory name. It carries the
// Analysis+Plan / Development+Testing method and the branch/commit conventions
// that the planner and orchestrator prompts point at.
const LifecycleSkillName = "dev-lifecycle"

// skillNames is every embedded skill Install lays down under <dataDir>/skills.
var skillNames = []string{SkillName, LifecycleSkillName}

// Dir returns the absolute directory the using-maestro skill installs into for a
// given data dir. Callers building prompts use this so the path they cite always
// matches where Install writes.
func Dir(dataDir string) string {
	return SkillDir(dataDir, SkillName)
}

// LifecycleDir returns the absolute directory the dev-lifecycle skill installs
// into for a given data dir.
func LifecycleDir(dataDir string) string {
	return SkillDir(dataDir, LifecycleSkillName)
}

// SkillDir returns the absolute install directory for a named embedded skill.
func SkillDir(dataDir, name string) string {
	return filepath.Join(dataDir, "skills", name)
}

// Install writes every embedded skill into <dataDir>/skills/<name>, replacing
// any existing copy. It runs once at daemon boot, before any session spawns, so
// a plain clobber-and-write needs no locking: there are no concurrent readers
// yet. A failure is returned but is non-fatal to boot (the skills enhance agent
// prompts, they are not load-bearing).
func Install(dataDir string) error {
	for _, name := range skillNames {
		if err := installSkill(dataDir, name); err != nil {
			return err
		}
	}
	return nil
}

func installSkill(dataDir, name string) error {
	dest := SkillDir(dataDir, name)
	if err := os.RemoveAll(dest); err != nil {
		return fmt.Errorf("clear skill dir %q: %w", dest, err)
	}
	// embed.FS always uses forward-slash paths rooted at the skill name; map each
	// onto <dataDir>/skills/<same path> with the platform separator.
	return fs.WalkDir(files, name, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		target := filepath.Join(dataDir, "skills", filepath.FromSlash(p))
		if d.IsDir() {
			return os.MkdirAll(target, 0o750)
		}
		b, err := files.ReadFile(p)
		if err != nil {
			return fmt.Errorf("read embedded %q: %w", p, err)
		}
		if err := os.WriteFile(target, b, 0o600); err != nil {
			return fmt.Errorf("write %q: %w", target, err)
		}
		return nil
	})
}
