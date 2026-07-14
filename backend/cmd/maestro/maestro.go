// Command maestro is the Maestro CLI. It manages the local daemon that
// supervises parallel coding-agent sessions.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tinhtran24/maestro/backend/internal/cli"
)

func main() {
	warnIfLegacyInvocation()
	if err := cli.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(cli.ExitCode(err))
	}
}

// warnIfLegacyInvocation prints a deprecation notice when the CLI is invoked
// under its old name `to` (via a legacy symlink/shim), so existing muscle
// memory keeps working while nudging users to `maestro`. Maestro was Thanos.
func warnIfLegacyInvocation() {
	name := strings.TrimSuffix(filepath.Base(os.Args[0]), ".exe")
	if name == "to" {
		fmt.Fprintln(os.Stderr, "`to` is deprecated and will be removed; use `maestro` (Thanos is now Maestro).")
	}
}
