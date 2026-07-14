# Maestro CLI

The `maestro` CLI is a thin Go/Cobra client for the local Maestro daemon.
It starts, discovers, inspects, and stops the daemon through the loopback HTTP
surface and the `running.json` handshake. It must not open SQLite directly or
call runtime, workspace, tracker, or agent adapters in-process.

When using the CLI directly from a shell, make sure the daemon is running first
with `maestro start` or by opening the desktop app. Product commands such as
`maestro agent ls` and `maestro spawn` call the loopback daemon and will fail with a
"daemon is not running" error if no `running.json` points at a live process. From
a source checkout, build and run the local binary explicitly, for example:

```bash
cd backend
go build -o ./bin/to ./cmd/maestro
./bin/to agent ls
```

## Current commands

Every product command resolves to a daemon HTTP route. Run `maestro <command>
--help` for the authoritative flag shape.

### Daemon control

| Command                                 | Purpose                                                                                                                           |
| --------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------- |
| `maestro start`                         | Start the daemon in the background and wait for `/readyz`.                                                                        |
| `maestro stop`                          | Gracefully stop the daemon via loopback `POST /shutdown` after verifying daemon identity.                                         |
| `maestro status` / `--json`             | Report daemon state from `running.json`, process liveness, `/healthz`, and `/readyz`.                                             |
| `maestro doctor` / `--json`             | Check config, data directory, DB-file presence, daemon state, `git`, and (on Darwin/Linux) `tmux`; on Windows conpty is built in. |
| `maestro completion <shell>`            | Generate completions for `bash`, `zsh`, `fish`, or `powershell`.                                                                  |
| `maestro version` / `maestro --version` | Print build metadata.                                                                                                             |
| `maestro daemon`                        | Hidden internal daemon entrypoint used by `maestro start`.                                                                        |

### Product commands

| Command                                  | Daemon route                                   |
| ---------------------------------------- | ---------------------------------------------- |
| `maestro project add`                    | `POST /api/v1/projects`                        |
| `maestro project ls`                     | `GET /api/v1/projects`                         |
| `maestro project get <id>`               | `GET /api/v1/projects/{id}`                    |
| `maestro project set-config <id>`        | `PUT /api/v1/projects/{id}/config`             |
| `maestro project rm <id>`                | `DELETE /api/v1/projects/{id}`                 |
| `maestro agent ls`                       | `GET /api/v1/agents`                           |
| `maestro agent ls --refresh`             | `POST /api/v1/agents/refresh`                  |
| `maestro spawn`                          | `POST /api/v1/sessions`                        |
| `maestro session ls`                     | `GET /api/v1/sessions`                         |
| `maestro session get <id>`               | `GET /api/v1/sessions/{id}`                    |
| `maestro session kill <id>`              | `POST /api/v1/sessions/{id}/kill`              |
| `maestro session restore <id>`           | `POST /api/v1/sessions/{id}/restore`           |
| `maestro session rename <id> <name>`     | `PATCH /api/v1/sessions/{id}`                  |
| `maestro session cleanup`                | `POST /api/v1/sessions/cleanup`                |
| `maestro session claim-pr <id> <pr-ref>` | `POST /api/v1/sessions/{id}/pr/claim`          |
| `maestro orchestrator ls`                | `GET /api/v1/orchestrators`                    |
| `maestro send`                           | `POST /api/v1/sessions/{id}/send`              |
| `maestro preview [url]`                  | `POST /api/v1/sessions/{id}/preview`           |
| `maestro hooks <agent> <event>`          | `POST /api/v1/sessions/{id}/activity` (hidden) |

`maestro agent ls` prints the daemon-supported agent catalog with local install/auth
readiness. Use `--refresh` to rerun the bounded local probes and `--json` to
print the raw inventory response.

`maestro spawn` resolves project context in this order: explicit `--project`,
`MAESTRO_PROJECT_ID`, `MAESTRO_SESSION_ID` (by fetching the current session from the
daemon), then the current working directory matched against registered project
paths. If `MAESTRO_SESSION_ID` is set but the session cannot be fetched, pass
`--project` explicitly.

If `--agent` / `--harness` is omitted, `maestro spawn` uses the resolved project's
`worker.agent` config. Before spawning, the CLI refreshes the advisory agent
catalog and fails early when the selected agent is unsupported, not installed,
or unauthorized. It warns-but-continues when auth remains unknown because daemon
spawn remains the authoritative runtime validation point. Use
`--skip-agent-check` to bypass only this CLI-side preflight.

`maestro preview` resolves its session from the `MAESTRO_SESSION_ID` environment variable
(it is meant to run inside a session), not a flag. With no argument it
autodetects an `index.html` in the session workspace; with a URL argument it
opens that URL verbatim (`file://`, `http`, `https`).

`go run .` in `backend/` remains a compatibility wrapper around the daemon.

PR and review actions (merge, resolve-comments, review execute/send) are
HTTP-only today and driven by the frontend; there are no `maestro pr` / `maestro review`
commands yet.

## Configuration

The CLI and daemon share the same environment-driven config:

| Var                        | Default                   | Purpose                |
| -------------------------- | ------------------------- | ---------------------- |
| `MAESTRO_PORT`             | `3001`                    | Loopback daemon port.  |
| `MAESTRO_RUN_FILE`         | `~/.maestro/running.json` | PID/port handshake.    |
| `MAESTRO_DATA_DIR`         | `~/.maestro/data`         | SQLite data directory. |
| `MAESTRO_REQUEST_TIMEOUT`  | `60s`                     | REST request timeout.  |
| `MAESTRO_SHUTDOWN_TIMEOUT` | `10s`                     | Graceful shutdown cap. |

The daemon always binds `127.0.0.1`.

## Manual smoke test

```bash
cd backend
go build -o /tmp/to ./cmd/maestro

tmp=$(mktemp -d)
export MAESTRO_RUN_FILE="$tmp/running.json"
export MAESTRO_DATA_DIR="$tmp/data"
export MAESTRO_PORT=3037

/tmp/to status --json
/tmp/to doctor
/tmp/to start
/tmp/to status --json
/tmp/to stop
/tmp/to status --json
rm -rf "$tmp"
```

## Adding new commands

Add a product command only when a daemon HTTP route owns the corresponding
mutation/read; the CLI must call that route rather than reimplementing daemon
behavior. Commands not yet exposed but with backend routes in place include
`maestro events ...` (over the CDC/SSE endpoint) and CLI parity for PR/review
actions.

Do not port old in-process TypeScript CLI behavior that mixed command handling
with storage and runtime implementation details.
