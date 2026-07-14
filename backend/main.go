// Command backend is a compatibility wrapper for the Maestro daemon.
// The user-facing CLI lives at cmd/maestro; keep this wrapper so existing `go run .`
// development workflows continue to start the daemon while scripts migrate.
package main

import (
	"fmt"
	"os"

	"github.com/tinhtran24/maestro/backend/internal/daemon"
)

func main() {
	if err := daemon.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "to backend daemon: "+err.Error())
		os.Exit(1)
	}
}
