# Thanos Development

Thanos uses a Go backend with a Wails desktop app and a React frontend.

## Repository Layout

```text
app/app/              Wails application boundary and local provider
app/frontend/         React desktop frontend
internal/agents/      provider registry and adapter logic
internal/workspace/   workspace and worktree primitives
internal/workbench/   workbench domain model
plans/                implementation roadmap
```

## Test Commands

Root Go packages:

```sh
GOCACHE=/private/tmp/thanos-go-build GOMODCACHE=/private/tmp/thanos-go-mod go test ./...
```

Wails app module:

```sh
cd app
GOCACHE=/private/tmp/thanos-go-build GOMODCACHE=/private/tmp/thanos-go-mod go test ./...
```

Frontend build:

```sh
cd app/frontend
npm run build
```

## Desktop App

Frontend development server:

```sh
cd app/frontend
npm install
npm run dev
```

Wails development mode:

```sh
cd app
GOCACHE=/private/tmp/thanos-go-build GOMODCACHE=/private/tmp/thanos-go-mod wails dev
```

## Implementation Rules

Bug fixes require a reproducible test. The test must fail without the fix and
pass with the fix.

Milestone work should update `plans/task-wallfacer-production-parity.md` when
the acceptance criteria are implemented.

Generated Wails bindings live under:

```text
app/frontend/wailsjs/
```

When Go app types or exported Wails methods change, regenerate bindings with
`wails generate` or update the generated metadata consistently in the same
change.

## Local Data Contracts

The desktop app loads durable local state from `.thanos/` in the selected
workspace. New records should remain JSON-compatible and include stable field
names for frontend use.

Task implementation output and verification output are intentionally separated:

```text
.thanos/logs/{task-id}/      implementation turns
.thanos/tests/{task-id}/     verification runs
```

No provider API token should be injected by Thanos. Native terminal and
verification commands scrub common provider token variables by default.
