#!/usr/bin/env bash
#
# Fresh-machine install check. The Dockerfile installs `maestro` on PATH in a clean
# image and runs this; it proves a freshly installed binary actually works on a
# machine with no Go toolchain and no developer state. The COMPREHENSIVE,
# cross-platform behavioural suite lives in Go (backend/internal/cli/e2e_test.go,
# `go test -tags e2e`); this stays deliberately small and linear.

set -euo pipefail

MAESTRO_BIN="${MAESTRO_BIN:-to}"
tmp="$(mktemp -d)"
export MAESTRO_RUN_FILE="$tmp/running.json"
export MAESTRO_DATA_DIR="$tmp/data"
trap 'rm -rf "$tmp"' EXIT

fail() { echo "FAIL: $1" >&2; exit 1; }

echo "to binary : $(command -v "$MAESTRO_BIN")"
"$MAESTRO_BIN" version            >/dev/null || fail "version"
"$MAESTRO_BIN" doctor             >/dev/null || fail "doctor"

# `maestro start` is now the desktop-app launcher: it resolves an installed app or
# fetches the release, then opens it (it no longer runs a daemon). On a fresh
# container there is no installed app, so start reaches the fetch path. The smoke
# binary is built against a release repo with no published assets (see
# Dockerfile), so the fetch deterministically 404s and start must exit non-zero
# with a clear `maestro start:` error (an unreachable/404 download on amd64, or an
# unsupported-arch error on arm64), never a panic or a silent success. The full
# launcher behaviour is covered by the Go e2e suite; this only proves the
# fresh-box path is sane on whatever arch the runner uses.
if err="$("$MAESTRO_BIN" start 2>&1)"; then
  fail "start unexpectedly succeeded on a fresh machine with no installed app"
fi
echo "$err" | grep -qiE "download|to start:" || fail "start did not fail with a clear error; got: $err"

echo "fresh-install check: OK"
