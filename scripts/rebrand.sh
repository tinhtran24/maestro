#!/usr/bin/env bash
# One-shot rebrand: Agent Orchestrator / to / to -> Maestro / to.
# Operates on tracked text files only (skips node_modules, .git, lockfiles, binaries).
set -euo pipefail
cd "$(dirname "$0")/.."

PROG='
  # --- GitHub org/repo (most specific first) ---
  s{aoagents/agent-orchestrator}{tinhtran24/maestro}g;
  s{\baoagents\b}{tinhtran}g;

  # --- doctor user-agent / hook asset (before generic slug) ---
  s{to-agent-orchestrator/doctor}{maestro/doctor}g;
  s{to-activity}{maestro-activity}g;

  # --- display + slug ---
  s{Agent Orchestrator}{Maestro}g;
  s{agent-orchestrator}{maestro}g;

  # --- env var prefix AO_FOO -> MAESTRO_FOO (incl. VITE_/__ prefixed) ---
  s{\bAO_([A-Z0-9])}{MAESTRO_$1}g;
  s{VITE_AO_}{VITE_MAESTRO_}g;
  s{__AO_}{__MAESTRO_}g;

  # --- misc branded slugs ---
  s{to-fresh-install-fixture}{maestro-fresh-install-fixture}g;

  # --- data dir + db + doctor temp ---
  s{~/\.to\b}{~/.maestro}g;
  s{"\.to"}{".maestro"}g;
  s{\.to/electron}{.maestro/electron}g;
  s{(^|[^A-Za-z0-9_])\.to/}{$1.maestro/}gm;
  s{\bto\.db\b}{maestro.db}g;
  s{\.to-doctor-write}{.maestro-doctor-write}g;

  # --- hook command + doctor binary check ---
  s{\bto hooks\b}{maestro hooks}g;
  s{LookPath\("to"\)}{LookPath("maestro")}g;
  s{"to-binary"}{"to-binary"}g;
  s{\bto not found in PATH}{to not found in PATH}g;
  s{\bto in PATH}{to in PATH}g;
  s{foreign to\b}{foreign to}g;

  # --- CLI command examples: `maestro <subcommand>` -> `maestro <subcommand>` ---
  s{\bto (start|stop|status|daemon|doctor|agent|spawn|send|session|orchestrator|project|review|preview|import|hooks|pty-host|version|completion)\b}{to $1}g;

  # --- logo ---
  s{to-logo}{maestro-logo}g;

  # --- standalone to acronym in prose/comments ---
  s{\bto\b}{Maestro}g;
'

find . \
  \( -path ./node_modules -o -path ./.git -o -path ./.idea -o -path '*/node_modules' -o -path ./app/daemon \) -prune -o \
  -type f \
  ! -name 'pnpm-lock.yaml' ! -name 'package-lock.json' ! -name 'openapi.yaml' \
  ! -name 'routeTree.gen.ts' ! -name 'rebrand.sh' \
  \( \
    -name '*.go' -o -name '*.ts' -o -name '*.tsx' -o -name '*.js' -o -name '*.mjs' -o -name '*.cjs' \
    -o -name '*.json' -o -name '*.md' -o -name '*.mdx' -o -name '*.yaml' -o -name '*.yml' \
    -o -name '*.sh' -o -name '*.html' -o -name '*.svg' -o -name '*.txt' -o -name '*.toml' -o -name '*.sql' \
    -o -name '*.css' -o -name '*.nix' -o -name 'Makefile' -o -name '.nvmrc' \
    -o -name '.gitignore' -o -name '.prettierrc' -o -name '.prettierignore' -o -name '.envrc' \
    -o -name 'Dockerfile' \
  \) -print0 \
| xargs -0 perl -0777 -i -pe "$PROG"

echo "Rebrand pass complete."
