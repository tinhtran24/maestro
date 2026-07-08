package daemon

import (
	"os"
	"runtime"
	"strings"
	"testing"
)

func TestAugmentToolPath(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("PATH augmentation is a no-op on Windows")
	}

	t.Run("appends missing common tool dirs, preserving existing order", func(t *testing.T) {
		t.Setenv("PATH", "/usr/bin:/bin")
		t.Setenv("THANOS_TMUX_BIN", "")
		augmentToolPath()
		got := os.Getenv("PATH")
		if !strings.HasPrefix(got, "/usr/bin:/bin") {
			t.Fatalf("existing PATH not preserved at front: %q", got)
		}
		if !strings.Contains(got, "/opt/homebrew/bin") {
			t.Fatalf("PATH missing Homebrew dir: %q", got)
		}
	})

	t.Run("does not duplicate dirs already present", func(t *testing.T) {
		t.Setenv("PATH", "/opt/homebrew/bin:/usr/bin")
		t.Setenv("THANOS_TMUX_BIN", "")
		augmentToolPath()
		got := os.Getenv("PATH")
		if strings.Count(got, "/opt/homebrew/bin") != 1 {
			t.Fatalf("duplicated existing dir: %q", got)
		}
	})

	t.Run("adds THANOS_TMUX_BIN directory", func(t *testing.T) {
		t.Setenv("PATH", "/usr/bin")
		t.Setenv("THANOS_TMUX_BIN", "/custom/prefix/bin/tmux")
		augmentToolPath()
		got := os.Getenv("PATH")
		if !strings.Contains(got, "/custom/prefix/bin") {
			t.Fatalf("PATH missing THANOS_TMUX_BIN dir: %q", got)
		}
	})
}
