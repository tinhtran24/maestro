#!/usr/bin/env bash
# One-shot rebrand: Agent Orchestrator / AO / ao -> Thanos / to.
# Operates on tracked text files only (skips node_modules, .git, lockfiles, binaries).
set -euo pipefail
cd "$(dirname "$0")/.."

PROG='
  # --- GitHub org/repo (most specific first) ---
  s{aoagents/agent-orchestrator}{tinhtran/thanos}g;
  s{\baoagents\b}{tinhtran}g;

  # --- doctor user-agent / hook asset (before generic slug) ---
  s{ao-agent-orchestrator/doctor}{thanos/doctor}g;
  s{ao-activity}{thanos-activity}g;

  # --- display + slug ---
  s{Agent Orchestrator}{Thanos}g;
  s{agent-orchestrator}{thanos}g;

  # --- env var prefix AO_FOO -> THANOS_FOO (incl. VITE_/__ prefixed) ---
  s{\bAO_([A-Z0-9])}{THANOS_$1}g;
  s{VITE_AO_}{VITE_THANOS_}g;
  s{__AO_}{__THANOS_}g;

  # --- misc branded slugs ---
  s{ao-fresh-install-fixture}{thanos-fresh-install-fixture}g;

  # --- data dir + db + doctor temp ---
  s{~/\.ao\b}{~/.thanos}g;
  s{"\.ao"}{".thanos"}g;
  s{\.ao/electron}{.thanos/electron}g;
  s{(^|[^A-Za-z0-9_])\.ao/}{$1.thanos/}gm;
  s{\bao\.db\b}{thanos.db}g;
  s{\.ao-doctor-write}{.thanos-doctor-write}g;

  # --- hook command + doctor binary check ---
  s{\bao hooks\b}{to hooks}g;
  s{LookPath\("ao"\)}{LookPath("to")}g;
  s{"ao-binary"}{"to-binary"}g;
  s{\bao not found in PATH}{to not found in PATH}g;
  s{\bao in PATH}{to in PATH}g;
  s{foreign ao\b}{foreign to}g;

  # --- CLI command examples: `ao <subcommand>` -> `to <subcommand>` ---
  s{\bao (start|stop|status|daemon|doctor|agent|spawn|send|session|orchestrator|project|review|preview|import|hooks|pty-host|version|completion)\b}{to $1}g;

  # --- logo ---
  s{ao-logo}{thanos-logo}g;

  # --- standalone AO acronym in prose/comments ---
  s{\bAO\b}{Thanos}g;
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
