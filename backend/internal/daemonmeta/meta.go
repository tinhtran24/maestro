package daemonmeta

// ServiceName identifies the Thanos daemon in loopback health/readiness probes.
// The CLI uses it with the reported PID to avoid signaling an unrelated process
// when a stale run-file's PID has been reused.
const ServiceName = "thanos-daemon"
