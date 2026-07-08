package daemon

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// commonToolDirs are locations a working macOS/Linux box keeps CLI tools (tmux,
// git, agent binaries) that a GUI/Finder-launched process often omits from PATH.
// Kept in sync with the desktop app's PATH floor (app/src/shared/shell-env.ts).
var commonToolDirs = []string{
	"/opt/homebrew/bin",
	"/opt/homebrew/sbin",
	"/usr/local/bin",
	"/usr/bin",
	"/bin",
}

// augmentToolPath appends common tool directories (and, when set, the directory
// of THANOS_TMUX_BIN) to PATH so runtime prerequisites like tmux resolve even
// when the daemon was launched with a minimal environment — e.g. a packaged app
// started from Finder/Dock, where PATH lacks Homebrew. It only ever appends
// missing dirs, so an operator's PATH ordering/overrides win. No-op on Windows.
//
// This runs before any exec.LookPath, so both the prerequisite check
// (session_manager) and the tmux runtime adapter see the same, enriched PATH.
func augmentToolPath() {
	if runtime.GOOS == "windows" {
		return
	}

	dirs := make([]string, 0, len(commonToolDirs)+1)
	// THANOS_TMUX_BIN lets an operator point Thanos at a specific tmux binary
	// (custom build, non-standard prefix). Adding its directory keeps LookPath
	// consistent across the runtime and the prerequisite check.
	if bin := strings.TrimSpace(os.Getenv("THANOS_TMUX_BIN")); bin != "" {
		dirs = append(dirs, filepath.Dir(bin))
	}
	dirs = append(dirs, commonToolDirs...)

	current := os.Getenv("PATH")
	have := make(map[string]bool)
	for _, p := range strings.Split(current, string(os.PathListSeparator)) {
		if p != "" {
			have[p] = true
		}
	}

	parts := []string{}
	if current != "" {
		parts = append(parts, current)
	}
	for _, d := range dirs {
		if d != "" && !have[d] {
			parts = append(parts, d)
			have[d] = true
		}
	}
	_ = os.Setenv("PATH", strings.Join(parts, string(os.PathListSeparator)))
}
