#!/usr/bin/env bash
# Cross-compile the Go `to` binary (backend/cmd/to) for every supported
# platform and drop each into the matching platform package's bin/ dir.
#
# Run this from any cwd before `npm publish`. It is the ONLY way the binaries
# get into the platform packages; they are gitignored and produced here, then
# shipped in each npm tarball via that package's `files` entry.
#
# CGO-free build (modernc.org/sqlite driver) so cross-compilation needs no C
# toolchain. Prod build: no -ldflags, so cli.releaseRepo keeps its default
# (AgentWrapper/thanos).
set -euo pipefail

# Repo layout: this script lives at <repo>/packages/build-binaries.sh.
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
BACKEND_DIR="${REPO_ROOT}/backend"

# pkg_dir : npm_os : npm_arch : GOOS : GOARCH : bin_name
TARGETS=(
  "thanos-darwin-arm64:darwin:arm64:darwin:arm64:to"
  "thanos-darwin-x64:darwin:x64:darwin:amd64:to"
  "thanos-win32-x64:win32:x64:windows:amd64:to.exe"
  "thanos-linux-x64:linux:x64:linux:amd64:to"
)

echo "Building to binaries from ${BACKEND_DIR}/cmd/to"
for t in "${TARGETS[@]}"; do
  IFS=":" read -r pkg npm_os npm_arch goos goarch bin <<<"$t"
  out="${SCRIPT_DIR}/${pkg}/bin/${bin}"
  mkdir -p "${SCRIPT_DIR}/${pkg}/bin"
  echo "  -> ${pkg} (GOOS=${goos} GOARCH=${goarch}) -> bin/${bin}"
  (cd "${BACKEND_DIR}" && CGO_ENABLED=0 GOOS="${goos}" GOARCH="${goarch}" \
    go build -o "${out}" ./cmd/to)
  chmod 0755 "${out}"
done

echo "Done. Built binaries:"
for t in "${TARGETS[@]}"; do
  IFS=":" read -r pkg _ _ _ _ bin <<<"$t"
  file "${SCRIPT_DIR}/${pkg}/bin/${bin}"
done
