#!/usr/bin/env bash
# Run a command with a Node >= 22.12 (or 20.19+) on PATH.
#
# Vite 8 (used by the Electron app) requires Node ^20.19 || >=22.12 — older Node
# can't `require()` Vite's ESM and electron-forge fails to load forge.config.
# This picks the active node if it qualifies, else the newest qualifying nvm
# install, and execs the command with it on PATH. Keeps `make dev` working even
# when the system node is too old.
set -euo pipefail

qualifies() {
	"$1" -e 'const v=process.versions.node.split(".").map(Number);process.exit((v[0]>22||(v[0]===22&&v[1]>=12)||(v[0]===20&&v[1]>=19))?0:1)' 2>/dev/null
}

NODE_BIN=""
if command -v node >/dev/null 2>&1 && qualifies "$(command -v node)"; then
	NODE_BIN="$(command -v node)"
elif [ -d "$HOME/.nvm/versions/node" ]; then
	for d in $(ls -d "$HOME"/.nvm/versions/node/* 2>/dev/null | sort -V -r); do
		if qualifies "$d/bin/node"; then NODE_BIN="$d/bin/node"; break; fi
	done
fi

if [ -z "$NODE_BIN" ]; then
	echo "error: Node >=22.12 (or 20.19+) is required for Vite 8, but none was found." >&2
	echo "       Install one, e.g.:  nvm install 24   (or upgrade your system Node)." >&2
	exit 1
fi

export PATH="$(dirname "$NODE_BIN"):$PATH"
exec "$@"
