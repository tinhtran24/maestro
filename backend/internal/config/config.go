// Package config loads the daemon's runtime configuration. The HTTP daemon is
// a loopback-only sidecar: it binds 127.0.0.1, takes no public traffic, and
// reads everything it needs from the environment with sane defaults so it can
// boot with zero configuration in development.
package config

import (
	"fmt"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	// LoopbackHost is the only host the daemon ever binds. There is deliberately
	// no MAESTRO_HOST env var: the daemon has no auth/CORS/TLS and a stray
	// MAESTRO_HOST=0.0.0.0 would turn it into a public no-auth service. If a
	// non-default loopback (e.g. ::1, 127.0.0.2) is ever needed, add it back with
	// an IsLoopback() validator — not a raw env read.
	LoopbackHost = "127.0.0.1"
	// DefaultPort is the single port for REST, terminal mux, health, and control.
	DefaultPort = 3001
	// DefaultRequestTimeout bounds a single REST request. Long-lived terminal mux
	// connections are mounted outside this timeout.
	DefaultRequestTimeout = 60 * time.Second
	// DefaultShutdownTimeout is the hard cap on graceful shutdown. After this
	// the process exits even if connections are still draining.
	DefaultShutdownTimeout = 10 * time.Second
	// DefaultAgent is the compatibility value used when MAESTRO_AGENT is unset. The
	// daemon validates it at startup, but worker/orchestrator spawns resolve from
	// explicit requests or project role config instead of falling back to it.
	DefaultAgent = "claude-code"
	// DefaultTelemetryPostHogHost is the default PostHog ingestion host when
	// remote telemetry is enabled and MAESTRO_TELEMETRY_POSTHOG_HOST is unset.
	DefaultTelemetryPostHogHost = "https://us.i.posthog.com"
)

// TelemetryRemote selects the remote telemetry exporter.
type TelemetryRemote string

const (
	// TelemetryRemoteOff disables remote telemetry export.
	TelemetryRemoteOff TelemetryRemote = "off"
	// TelemetryRemotePostHog exports allowlisted events to PostHog.
	TelemetryRemotePostHog TelemetryRemote = "posthog"
)

// TelemetryConfig controls local and remote telemetry behavior.
type TelemetryConfig struct {
	Events      bool
	Metrics     bool
	Remote      TelemetryRemote
	PostHogKey  string
	PostHogHost string
}

// DefaultAllowedOrigins are the browser origins the daemon's CORS boundary
// trusts, beyond loopback-served content (which the middleware always trusts —
// local pages can reach the no-auth daemon directly anyway). The daemon has no
// auth, so every entry must be an origin web content cannot present:
// app://renderer is the packaged Electron renderer, served from a custom
// scheme only the desktop app registers — no website can bear it. The opaque
// "null" origin (file:// pages, sandboxed iframes on any website) must never
// be added.
var DefaultAllowedOrigins = []string{
	"app://renderer",
}

// Config is the fully-resolved daemon configuration. It is immutable once
// built by Load.
type Config struct {
	// Host is the bind address. Always loopback — see LoopbackHost.
	Host string
	// Port is the TCP port to bind. The daemon fails fast if it is taken.
	Port int
	// RequestTimeout bounds REST request handling.
	RequestTimeout time.Duration
	// ShutdownTimeout is the hard graceful-shutdown deadline.
	ShutdownTimeout time.Duration
	// RunFilePath is where the PID + port handshake file (running.json) is
	// written so the Electron supervisor can discover and reap the daemon.
	RunFilePath string
	// DataDir is the directory holding durable SQLite state: DB and WAL files.
	// It is created on first use by the storage layer.
	DataDir string
	// Agent is the compatibility agent adapter id selected by MAESTRO_AGENT;
	// startSession fails fast if no adapter with this id is registered.
	Agent string
	// PlannerAgent is the default agent used by the Quick Capture "AI Structure"
	// step (MAESTRO_PLANNER_AGENT). Empty falls back to Agent.
	PlannerAgent string
	// AllowedOrigins are the browser origins granted CORS read access (see
	// DefaultAllowedOrigins). Overridden by MAESTRO_ALLOWED_ORIGINS.
	AllowedOrigins []string
	// Telemetry controls local/remote telemetry sinks.
	Telemetry TelemetryConfig
}

// Addr returns the host:port the HTTP server binds. It uses net.JoinHostPort so
// the result is correct for IPv6 literals as well as IPv4 / hostnames.
func (c Config) Addr() string {
	return net.JoinHostPort(c.Host, strconv.Itoa(c.Port))
}

// Load resolves configuration from the environment, applying defaults. It
// returns an error only for values that are present but malformed (e.g. a
// non-numeric MAESTRO_PORT); missing values fall back to defaults.
//
// Recognised variables:
//
//	MAESTRO_PORT              bind port           (default 3001)
//	MAESTRO_REQUEST_TIMEOUT   per-request timeout (Go duration > 0, default 60s)
//	MAESTRO_SHUTDOWN_TIMEOUT  shutdown deadline   (Go duration > 0, default 10s)
//	MAESTRO_RUN_FILE          running.json path   (default ~/.maestro/running.json)
//	MAESTRO_DATA_DIR          durable state dir   (default ~/.maestro/data)
//	MAESTRO_AGENT             compatibility agent id (default claude-code)
//	MAESTRO_ALLOWED_ORIGINS   CORS origins, comma-separated (default DefaultAllowedOrigins)
//	MAESTRO_TELEMETRY_EVENTS  local event capture off|on (default off)
//	MAESTRO_TELEMETRY_METRICS local metric capture off|on (default off)
//	MAESTRO_TELEMETRY_REMOTE  remote exporter off|posthog (default off)
//	MAESTRO_TELEMETRY_POSTHOG_KEY   PostHog project key
//	MAESTRO_TELEMETRY_POSTHOG_HOST  PostHog host (default DefaultTelemetryPostHogHost)
//
// The bind host is not configurable: the daemon is loopback-only by design.
func Load() (Config, error) {
	MigrateLegacyEnv()
	cfg := Config{
		Host:            LoopbackHost,
		Port:            DefaultPort,
		RequestTimeout:  DefaultRequestTimeout,
		ShutdownTimeout: DefaultShutdownTimeout,
		Agent:           DefaultAgent,
		AllowedOrigins:  DefaultAllowedOrigins,
		Telemetry: TelemetryConfig{
			Remote:      TelemetryRemoteOff,
			PostHogHost: DefaultTelemetryPostHogHost,
		},
	}

	if raw := os.Getenv("MAESTRO_PORT"); raw != "" {
		port, err := strconv.Atoi(raw)
		if err != nil {
			return Config{}, fmt.Errorf("invalid MAESTRO_PORT %q: %w", raw, err)
		}
		if port < 1 || port > 65535 {
			return Config{}, fmt.Errorf("invalid MAESTRO_PORT %d: out of range 1-65535", port)
		}
		cfg.Port = port
	}

	if raw := os.Getenv("MAESTRO_REQUEST_TIMEOUT"); raw != "" {
		d, err := parsePositiveDuration("MAESTRO_REQUEST_TIMEOUT", raw)
		if err != nil {
			return Config{}, err
		}
		cfg.RequestTimeout = d
	}

	if raw := os.Getenv("MAESTRO_SHUTDOWN_TIMEOUT"); raw != "" {
		d, err := parsePositiveDuration("MAESTRO_SHUTDOWN_TIMEOUT", raw)
		if err != nil {
			return Config{}, err
		}
		cfg.ShutdownTimeout = d
	}

	if raw := os.Getenv("MAESTRO_AGENT"); raw != "" {
		cfg.Agent = raw
	}

	if raw := os.Getenv("MAESTRO_PLANNER_AGENT"); raw != "" {
		cfg.PlannerAgent = raw
	}
	if cfg.PlannerAgent == "" {
		cfg.PlannerAgent = cfg.Agent
	}

	if raw, ok := os.LookupEnv("MAESTRO_ALLOWED_ORIGINS"); ok && raw != "" {
		// Explicit override replaces the defaults entirely so a deployment can
		// also narrow the list. The "null" origin is rejected, never silently
		// dropped: an operator allowing it would open the no-auth daemon to
		// every sandboxed iframe on the web.
		origins := make([]string, 0, 4)
		for _, origin := range strings.Split(raw, ",") {
			origin = strings.TrimSpace(origin)
			if origin == "" {
				continue
			}
			if origin == "null" || origin == "*" {
				return Config{}, fmt.Errorf("invalid MAESTRO_ALLOWED_ORIGINS entry %q: wildcard and null origins are not allowed", origin)
			}
			origins = append(origins, origin)
		}
		cfg.AllowedOrigins = origins
	}

	if raw := os.Getenv("MAESTRO_TELEMETRY_EVENTS"); raw != "" {
		v, err := parseToggleEnv("MAESTRO_TELEMETRY_EVENTS", raw)
		if err != nil {
			return Config{}, err
		}
		cfg.Telemetry.Events = v
	}
	if raw := os.Getenv("MAESTRO_TELEMETRY_METRICS"); raw != "" {
		v, err := parseToggleEnv("MAESTRO_TELEMETRY_METRICS", raw)
		if err != nil {
			return Config{}, err
		}
		cfg.Telemetry.Metrics = v
	}
	if raw := os.Getenv("MAESTRO_TELEMETRY_REMOTE"); raw != "" {
		remote, err := parseTelemetryRemote(raw)
		if err != nil {
			return Config{}, fmt.Errorf("invalid MAESTRO_TELEMETRY_REMOTE %q: %w", raw, err)
		}
		cfg.Telemetry.Remote = remote
	}
	if raw := os.Getenv("MAESTRO_TELEMETRY_POSTHOG_KEY"); raw != "" {
		cfg.Telemetry.PostHogKey = raw
	}
	if raw := os.Getenv("MAESTRO_TELEMETRY_POSTHOG_HOST"); raw != "" {
		cfg.Telemetry.PostHogHost = raw
	}

	runFile, err := resolveRunFilePath()
	if err != nil {
		return Config{}, err
	}
	cfg.RunFilePath = runFile

	dataDir, err := resolveDataDir()
	if err != nil {
		return Config{}, err
	}
	cfg.DataDir = dataDir

	return cfg, nil
}

func parseToggleEnv(name, raw string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "on", "true", "1", "yes":
		return true, nil
	case "off", "false", "0", "no":
		return false, nil
	default:
		return false, fmt.Errorf("%s must be off|on", name)
	}
}

func parseTelemetryRemote(raw string) (TelemetryRemote, error) {
	switch TelemetryRemote(strings.ToLower(strings.TrimSpace(raw))) {
	case TelemetryRemoteOff:
		return TelemetryRemoteOff, nil
	case TelemetryRemotePostHog:
		return TelemetryRemotePostHog, nil
	default:
		return "", fmt.Errorf("must be off|posthog")
	}
}

// parsePositiveDuration rejects zero and negative durations: a zero
// RequestTimeout would expire every request instantly, and a non-positive
// ShutdownTimeout would defeat graceful shutdown.
func parsePositiveDuration(name, raw string) (time.Duration, error) {
	d, err := time.ParseDuration(raw)
	if err != nil {
		return 0, fmt.Errorf("invalid %s %q: %w", name, raw, err)
	}
	if d <= 0 {
		return 0, fmt.Errorf("invalid %s %q: must be > 0", name, raw)
	}
	return d, nil
}

// resolveRunFilePath picks where running.json lives. An explicit MAESTRO_RUN_FILE
// wins; otherwise it sits under the canonical Maestro home directory so the CLI and
// Electron supervisor share one handshake location.
func resolveRunFilePath() (string, error) {
	if p, ok := os.LookupEnv("MAESTRO_RUN_FILE"); ok && p != "" {
		return p, nil
	}
	stateDir, err := defaultStateDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(stateDir, "running.json"), nil
}

// resolveDataDir picks where durable state (the SQLite DB) lives. An explicit
// MAESTRO_DATA_DIR wins; otherwise it defaults under the same canonical Maestro home
// directory as the run-file.
func resolveDataDir() (string, error) {
	if p, ok := os.LookupEnv("MAESTRO_DATA_DIR"); ok && p != "" {
		return p, nil
	}
	stateDir, err := defaultStateDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(stateDir, "data"), nil
}

func defaultStateDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve state dir: %w", err)
	}
	return resolveMaestroHome(filepath.Join(homeDir, ".maestro"), filepath.Join(homeDir, ".thanos")), nil
}

var (
	legacyEnvOnce         sync.Once
	legacyHomeMigrateOnce sync.Once
	legacyHomeKeepOnce    sync.Once
)

// legacyEnvPrefix / envPrefix are the pre- and post-rename environment variable
// namespaces. Maestro was previously named Thanos.
const (
	legacyEnvPrefix = "THANOS_"
	envPrefix       = "MAESTRO_"
)

// MigrateLegacyEnv copies any legacy THANOS_* environment variable into its
// MAESTRO_* equivalent when the new name is unset, so callers that still export
// the old names keep working. The new name always wins. It runs once per process
// and emits a single deprecation warning if any legacy variable was adopted.
// Load() calls it before reading any MAESTRO_* value.
func MigrateLegacyEnv() {
	legacyEnvOnce.Do(func() {
		if applyLegacyEnv(os.Environ(), os.LookupEnv, os.Setenv) {
			slog.Warn("THANOS_* environment variables are deprecated; use MAESTRO_* (Thanos is now Maestro)")
		}
	})
}

// applyLegacyEnv maps THANOS_* names in environ onto their MAESTRO_* equivalents
// via setenv, skipping any whose MAESTRO_* name is already set (the new name
// wins). It returns whether any legacy variable was adopted. Split out from
// MigrateLegacyEnv so it can be tested without the process-wide sync.Once.
func applyLegacyEnv(environ []string, lookup func(string) (string, bool), setenv func(string, string) error) bool {
	adopted := false
	for _, kv := range environ {
		eq := strings.IndexByte(kv, '=')
		if eq < 0 {
			continue
		}
		key := kv[:eq]
		if !strings.HasPrefix(key, legacyEnvPrefix) {
			continue
		}
		newKey := envPrefix + strings.TrimPrefix(key, legacyEnvPrefix)
		if _, ok := lookup(newKey); ok {
			continue // the new name is set and wins
		}
		if err := setenv(newKey, kv[eq+1:]); err == nil {
			adopted = true
		}
	}
	return adopted
}

// resolveMaestroHome returns the canonical Maestro home directory, migrating a
// legacy ~/.thanos home to ~/.maestro when only the legacy one exists. The
// migration is a best-effort atomic rename; if it fails (for example across
// filesystems) the legacy directory is used in place so existing data,
// sessions, and worktrees are never lost. A deprecation notice is emitted once.
func resolveMaestroHome(maestroHome, legacyHome string) string {
	if _, err := os.Stat(maestroHome); err == nil {
		return maestroHome // already migrated / fresh install
	}
	if _, err := os.Stat(legacyHome); err != nil {
		return maestroHome // no legacy home to adopt
	}
	legacyHomeMigrateOnce.Do(func() {
		slog.Warn("legacy ~/.thanos data directory detected; migrating to ~/.maestro (Thanos is now Maestro)")
	})
	if err := os.Rename(legacyHome, maestroHome); err != nil {
		legacyHomeKeepOnce.Do(func() {
			slog.Warn("could not migrate ~/.thanos to ~/.maestro; continuing to use ~/.thanos in place", "err", err)
		})
		return legacyHome
	}
	return maestroHome
}
