// Command to is the Thanos Orchestrator CLI. It manages the local daemon that
// supervises parallel coding-agent sessions.
package main

import (
	"fmt"
	"os"

	"github.com/tinhtran/thanos/backend/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(cli.ExitCode(err))
	}
}
