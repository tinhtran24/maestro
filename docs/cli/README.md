# Thanos CLI

The `to` CLI is a thin Go/Cobra client for the local Thanos daemon.
It starts, discovers, inspects, and stops the daemon through the loopback HTTP
surface and the `running.json` handshake. It must not open SQLite directly or
call runtime, workspace, tracker, or agent adapters in-process.

When using the CLI directly from a shell, make sure the daemon is running first
with `to start` or by opening the desktop app. Product commands such as
`to agent ls` and `to spawn` call the loopback daemon and will fail with a
"daemon is not running" error if no `running.json` points at a live process. From
a source checkout, build and run the local binary explicitly, for example:

```bash
cd backend
go build -o ./bin/to ./cmd/to
./bin/to agent ls
```

## Current commands

Every product command resolves to a daemon HTTP route. Run `to <command>
--help` for the authoritative flag shape.

### Daemon control

| Command                       | Purpose                                                                                                                           |
| ----------------------------- | --------------------------------------------------------------------------------------------------------------------------------- |
| `to start`                    | Start the daemon in the background and wait for `/readyz`.                                                                        |
| `to stop`                     | Gracefully stop the daemon via loopback `POST /shutdown` after verifying daemon identity.                                         |
| `to status` / `--json`        | Report daemon state from `running.json`, process liveness, `/healthz`, and `/readyz`.                                             |
| `to doctor` / `--json`        | Check config, data directory, DB-file presence, daemon state, `git`, and (on Darwin/Linux) `tmux`; on Windows conpty is built in. |
| `to completion <shell>`       | Generate completions for `bash`, `zsh`, `fish`, or `powershell`.                                                                  |
| `to version` / `to --version` | Print build metadata.                                                                                                             |
| `to daemon`                   | Hidden internal daemon entrypoint used by `to start`.                                                                             |

### Product commands

| Command                             | Daemon route                                   |
| ----------------------------------- | ---------------------------------------------- |
| `to project add`                    | `POST /api/v1/projects`                        |
| `to project ls`                     | `GET /api/v1/projects`                         |
| `to project get <id>`               | `GET /api/v1/projects/{id}`                    |
| `to project set-config <id>`        | `PUT /api/v1/projects/{id}/config`             |
| `to project rm <id>`                | `DELETE /api/v1/projects/{id}`                 |
| `to agent ls`                       | `GET /api/v1/agents`                           |
| `to agent ls --refresh`             | `POST /api/v1/agents/refresh`                  |
| `to spawn`                          | `POST /api/v1/sessions`                        |
| `to session ls`                     | `GET /api/v1/sessions`                         |
| `to session get <id>`               | `GET /api/v1/sessions/{id}`                    |
| `to session kill <id>`              | `POST /api/v1/sessions/{id}/kill`              |
| `to session restore <id>`           | `POST /api/v1/sessions/{id}/restore`           |
| `to session rename <id> <name>`     | `PATCH /api/v1/sessions/{id}`                  |
| `to session cleanup`                | `POST /api/v1/sessions/cleanup`                |
| `to session claim-pr <id> <pr-ref>` | `POST /api/v1/sessions/{id}/pr/claim`          |
| `to session complete [id]`          | `POST /api/v1/sessions/{id}/complete`          |
| `to orchestrator ls`                | `GET /api/v1/orchestrators`                    |
| `to orchestrator finalize <worker-id> --state <state>` | `POST /api/v1/orchestrators/{id}/finalizations/{worker-id}` |
| `to task done <worker-id>`          | Final ordered orchestrator transition to `done` |

Workers never transition tasks to Review Pending or Done. They report only
`session complete`. The active orchestrator owns the durable ordered workflow:
git verification, tests, optional commit/push, PR creation or claim, metadata
persistence, runtime cleanup, Review Pending, and finally `to task done`.
Repeated completion reports and repeated acknowledgements of the current step
are idempotent; skipped steps and non-orchestrator callers are rejected.
| `to send`                           | `POST /api/v1/sessions/{id}/send`              |
| `to preview [url]`                  | `POST /api/v1/sessions/{id}/preview`           |
| `to hooks <agent> <event>`          | `POST /api/v1/sessions/{id}/activity` (hidden) |

`to agent ls` prints the daemon-supported agent catalog with local install/auth
readiness. Use `--refresh` to rerun the bounded local probes and `--json` to
print the raw inventory response.

`to spawn` resolves project context in this order: explicit `--project`,
`THANOS_PROJECT_ID`, `THANOS_SESSION_ID` (by fetching the current session from the
daemon), then the current working directory matched against registered project
paths. If `THANOS_SESSION_ID` is set but the session cannot be fetched, pass
`--project` explicitly.

If `--agent` / `--harness` is omitted, `to spawn` uses the resolved project's
`worker.agent` config. Before spawning, the CLI refreshes the advisory agent
catalog and fails early when the selected agent is unsupported, not installed,
or unauthorized. It warns-but-continues when auth remains unknown because daemon
spawn remains the authoritative runtime validation point. Use
`--skip-agent-check` to bypass only this CLI-side preflight.

Task creation also generates branch, commit-subject, and PR-title suggestions
before the worker starts. The desktop New Task flow lets users edit those
suggestions; API callers can pass `branch`, `commitMessage`, and `prTitle` on
`POST /api/v1/sessions`. The orchestrator uses the persisted metadata during
finalization instead of relying on the worker's completion text.

`to preview` resolves its session from the `THANOS_SESSION_ID` environment variable
(it is meant to run inside a session), not a flag. With no argument it
autodetects an `index.html` in the session workspace; with a URL argument it
opens that URL verbatim (`file://`, `http`, `https`).

`go run .` in `backend/` remains a compatibility wrapper around the daemon.

PR and review actions (merge, resolve-comments, review execute/send) are
HTTP-only today and driven by the frontend; there are no `to pr` / `to review`
commands yet.

## Configuration

The CLI and daemon share the same environment-driven config:

| Var                   | Default              | Purpose                |
| --------------------- | -------------------- | ---------------------- |
| `THANOS_PORT`             | `3001`               | Loopback daemon port.  |
| `THANOS_RUN_FILE`         | `~/.thanos/running.json` | PID/port handshake.    |
| `THANOS_DATA_DIR`         | `~/.thanos/data`         | SQLite data directory. |
| `THANOS_REQUEST_TIMEOUT`  | `60s`                | REST request timeout.  |
| `THANOS_SHUTDOWN_TIMEOUT` | `10s`                | Graceful shutdown cap. |

The daemon always binds `127.0.0.1`.

## Manual smoke test

```bash
cd backend
go build -o /tmp/to ./cmd/to

tmp=$(mktemp -d)
export THANOS_RUN_FILE="$tmp/running.json"
export THANOS_DATA_DIR="$tmp/data"
export THANOS_PORT=3037

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
`to events ...` (over the CDC/SSE endpoint) and CLI parity for PR/review
actions.

Do not port old in-process TypeScript CLI behavior that mixed command handling
with storage and runtime implementation details.
