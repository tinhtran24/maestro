package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestApplyLegacyEnv(t *testing.T) {
	set := map[string]string{}
	present := map[string]string{"MAESTRO_PORT": "9999"} // new name already set → wins
	lookup := func(k string) (string, bool) { v, ok := present[k]; return v, ok }
	setenv := func(k, v string) error { set[k] = v; return nil }

	environ := []string{
		"THANOS_DATA_DIR=/legacy/data",
		"THANOS_PORT=1111", // must be skipped: MAESTRO_PORT already set
		"PATH=/usr/bin",    // non-prefixed, ignored
		"THANOS_GITHUB_TOKEN=tok",
		"MALFORMED_NO_EQUALS", // no '=' → skipped by the split
	}

	adopted := applyLegacyEnv(environ, lookup, setenv)
	if !adopted {
		t.Fatal("applyLegacyEnv reported nothing adopted, want true")
	}
	if set["MAESTRO_DATA_DIR"] != "/legacy/data" {
		t.Errorf("MAESTRO_DATA_DIR = %q, want /legacy/data", set["MAESTRO_DATA_DIR"])
	}
	if set["MAESTRO_GITHUB_TOKEN"] != "tok" {
		t.Errorf("MAESTRO_GITHUB_TOKEN = %q, want tok", set["MAESTRO_GITHUB_TOKEN"])
	}
	if _, ok := set["MAESTRO_PORT"]; ok {
		t.Errorf("MAESTRO_PORT was overwritten; the already-set new name must win")
	}

	// No legacy vars → nothing adopted.
	if applyLegacyEnv([]string{"PATH=/usr/bin"}, lookup, setenv) {
		t.Error("applyLegacyEnv reported adoption with no legacy vars")
	}
}

func TestResolveMaestroHome(t *testing.T) {
	t.Run("new home present wins", func(t *testing.T) {
		dir := t.TempDir()
		maestro := filepath.Join(dir, ".maestro")
		legacy := filepath.Join(dir, ".thanos")
		mustMkdir(t, maestro)
		mustMkdir(t, legacy)
		if got := resolveMaestroHome(maestro, legacy); got != maestro {
			t.Fatalf("got %q, want %q", got, maestro)
		}
	})

	t.Run("no legacy → new home", func(t *testing.T) {
		dir := t.TempDir()
		maestro := filepath.Join(dir, ".maestro")
		legacy := filepath.Join(dir, ".thanos")
		if got := resolveMaestroHome(maestro, legacy); got != maestro {
			t.Fatalf("got %q, want %q", got, maestro)
		}
	})

	t.Run("legacy only → migrated to new home", func(t *testing.T) {
		dir := t.TempDir()
		maestro := filepath.Join(dir, ".maestro")
		legacy := filepath.Join(dir, ".thanos")
		mustMkdir(t, legacy)
		if err := os.WriteFile(filepath.Join(legacy, "data.db"), []byte("x"), 0o600); err != nil {
			t.Fatal(err)
		}
		got := resolveMaestroHome(maestro, legacy)
		if got != maestro {
			t.Fatalf("got %q, want migrated %q", got, maestro)
		}
		if _, err := os.Stat(filepath.Join(maestro, "data.db")); err != nil {
			t.Fatalf("migrated data missing: %v", err)
		}
		if _, err := os.Stat(legacy); !os.IsNotExist(err) {
			t.Fatalf("legacy dir should be gone after rename (err=%v)", err)
		}
	})
}

func mustMkdir(t *testing.T, p string) {
	t.Helper()
	if err := os.MkdirAll(p, 0o750); err != nil {
		t.Fatal(err)
	}
}
