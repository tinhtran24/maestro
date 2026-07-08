# Running Thanos

How to build and run Thanos locally. Thanos has two parts:

- **`backend/`** — the Go daemon and the `to` CLI ("Thanos"). A
  loopback-only HTTP/SSE/WebSocket sidecar that supervises coding-agent sessions.
- **`app/`** — the Electron + Vite + React desktop app that supervises the daemon
  and renders the UI.

> Module path: `github.com/tinhtran/thanos/backend`. CLI command: `to`.

## Prerequisites

| Tool     | Version           | Notes                                                        |
| -------- | ----------------- | ------------------------------------------------------------ |
| Go       | 1.25+             | builds the backend daemon + `to` CLI                         |
| Node.js  | **^20.19 \|\| >=22.12** | runs the Electron app; Vite 8 needs `require(esm)` (older Node fails to load `forge.config`) |
| pnpm     | 10 (preferred)    | app package manager (`npm` also works)                       |
| tmux     | any recent        | **required** runtime — the daemon supervises agent sessions in tmux |
| Git      | any recent        | worktrees per session                                        |
| gh       | any recent        | optional — GitHub CLI for PR/SCM features                    |

Install the runtime tools on macOS with: `brew install tmux gh`.

**Node version**: the repo ships an `.nvmrc` (Node 24) — run `nvm use` to select it.
`make` targets go through `scripts/with-node.sh`, which auto-selects a compatible
Node (preferring the active one, else the newest qualifying nvm install), so
`make dev` works even if your system Node is older.

Optional: **Nix + direnv** — `direnv allow` (the repo ships `.envrc` → `use flake`)
drops you into a shell with Go 1.25, Node 22, pnpm 10, and `just` already on PATH.
Otherwise install the tools above yourself.

## Quick start (desktop app)

From the repo root:

```bash
cd app
pnpm install          # or: npm install
pnpm dev              # or: npm run dev  — launches the Electron app
```

In dev mode the app builds and spawns the daemon for you (it runs
`go run ./cmd/to daemon`), so you only need Go on PATH — no separate daemon step.
Vite hot-reloads the renderer.

To run just the renderer in a browser (no Electron shell):

```bash
cd app
pnpm dev:web          # VITE_NO_ELECTRON=1 vite — UI only, expects a daemon running
```

## Backend (daemon + `to` CLI)

Build and run directly when you want the daemon/CLI without the desktop app.

```bash
cd backend

go build ./...                      # compile everything
go build -o to ./cmd/to             # build the CLI binary

# Run the daemon in the foreground:
go run ./cmd/to daemon              # or: ./to daemon

# In another shell, drive it with the CLI:
./to status                         # daemon status
./to project add <path>             # register a project
./to spawn --project <id> ...       # spawn a worker agent session
./to session ls                     # list sessions
./to --help                         # full command list
```

`go run .` from `backend/` also starts the daemon (a compatibility wrapper around
`to daemon`).

The daemon binds to `127.0.0.1` only and serves the API under `/api/v1`
(health at `/api/v1/healthz`).

## API type generation

The app's TypeScript API types are generated from the backend's OpenAPI spec — do
not hand-edit them. From the repo root:

```bash
pnpm run api          # regenerate spec (go generate) + app/src/api/schema.ts
# or individually:
pnpm run api:spec     # cd backend && go generate ./internal/httpd/apispec/...
pnpm run api:ts       # openapi-typescript backend/.../openapi.yaml -o app/src/api/schema.ts
```

Regenerating the OpenAPI document directly:

```bash
cd backend
go generate ./internal/httpd/apispec/...   # runs `go run ./cmd/gen spec`
```

## Tests & checks

```bash
# Backend
cd backend
go test ./...                # unit + package tests
go test -race ./...          # race detector
go vet ./...

# App
cd app
pnpm typecheck               # tsc --noEmit
pnpm test                    # vitest
pnpm test:e2e                # playwright

# From repo root
pnpm run lint                # backend go test ./... + golangci-lint
pnpm run frontend:typecheck  # app TypeScript check
pnpm run sqlc                # regenerate sqlc code from queries/migrations
```

## Packaging the desktop app

```bash
cd app
pnpm build:daemon     # build the bundled daemon binary into app/daemon/
pnpm package          # electron-forge package (unpacked app)
pnpm make             # electron-forge make (installers: zip/deb/rpm)
```

`build:daemon` compiles `backend/cmd/to` into `app/daemon/` (gitignored); `package`
and `make` run it automatically via the `prepackage`/`premake` hooks.

## App state / data directory

All runtime state lives under **`~/.thanos`** (daemon data, `running.json`, worktrees,
and the Electron `userData`). Override with `THANOS_DATA_DIR` / `THANOS_RUN_FILE`. Nothing
is written to OS-default app-data locations.

## Troubleshooting

- **`to`/daemon can't be built**: ensure Go 1.25+ (`go version`) and run from
  `backend/`.
- **App starts but shows no data**: confirm the daemon is up —
  `curl http://127.0.0.1:<port>/api/v1/healthz` or `./to status`.
- **Stale API types**: rerun `pnpm run api` after changing backend controllers/DTOs.
- **Port/lock conflicts**: a previous daemon may still own `~/.thanos/running.json`;
  `./to stop` (or remove the run file) and retry.
- **`Cannot use 'import.meta' outside a module` / `require() of ES Module ... not
  supported` when starting the app**: your Node is too old for Vite 8. Use Node
  `>=22.12` (`nvm use`), or just run via `make dev` (it auto-selects one).
- **`tmux required ... but not in PATH`** when spawning a session: install tmux
  (`brew install tmux`). It's the session runtime. The daemon augments PATH with
  common tool dirs (`/opt/homebrew/bin`, `/usr/local/bin`, …) so a Finder/Dock-
  launched app finds Homebrew tmux; to use a custom/non-standard tmux, set
  **`THANOS_TMUX_BIN`** to its path (the runtime and the prerequisite check both
  honor it).
- **`gh: executable file not found`**: install the GitHub CLI (`brew install gh`)
  for PR/SCM features, or ignore it if you don't need them.
- **pnpm skips Electron's postinstall** (`Ignored build scripts: electron`): the
  approvals live in `app/pnpm-workspace.yaml` (`onlyBuiltDependencies`); if the
  Electron binary is missing, run `node app/node_modules/electron/install.js`.
