---
name: using-maestro
description: Catalog of the Maestro `maestro` CLI: spawning workers, managing sessions and projects, sending messages, previewing pages, and daemon control. Use when using the to CLI, spawning workers, or managing Maestro sessions in an Maestro workspace.
trigger: Using the to CLI in an Maestro workspace: spawning workers, managing sessions/projects, sending messages, previewing pages.
---

# Maestro CLI Catalog

`maestro` is a thin CLI over the local Maestro daemon. Every command is `maestro <command> --help` for the authoritative flag list.

| Command | What it does | When to use | Details |
|---|---|---|---|
| `spawn` | Spawn a worker agent in a fresh git worktree | Starting a new task or issue | [commands/spawn.md](commands/spawn.md) |
| `session` | Manage agent sessions (list, kill, rename, restore, etc.) | Inspecting or controlling running/terminated sessions | [commands/session.md](commands/session.md) |
| `project` | Register, inspect, configure, or remove projects | Setting up or managing repos Maestro knows about | [commands/project.md](commands/project.md) |
| `orchestrator` | List orchestrator sessions | Viewing which sessions are orchestrators | [commands/orchestrator.md](commands/orchestrator.md) |
| `review` | Submit a reviewer result for a worker's PR | Completing a code review loop | [commands/review.md](commands/review.md) |
| `send` | Send a message to a running agent session | Correcting or directing a live agent | [commands/send.md](commands/send.md) |
| `preview` | Open a URL in the desktop browser panel | Demoing a local server or file from inside a session | [commands/preview.md](commands/preview.md) |
| `start` | Fetch (if needed) and open the Maestro desktop app | Launching the app | [commands/start.md](commands/start.md) |
| `stop` | Stop the Maestro daemon | Shutting down Maestro | [commands/stop.md](commands/stop.md) |
| `status` | Show daemon status | Verifying the daemon is up and healthy | [commands/status.md](commands/status.md) |
| `doctor` | Run local health checks | Diagnosing Maestro setup problems | [commands/doctor.md](commands/doctor.md) |
| `import` | Import projects from a legacy Maestro install | Migrating from the old flat-file store | [commands/import.md](commands/import.md) |
| `version` | Print version information | Checking installed version | - |
| `completion` | Generate shell completion scripts | Setting up tab completion | - |

## Conventions

- Most read commands accept `--json` for machine-readable output.
- `-p / --project` scopes session subcommand lookups to one project.
- Session and project ids are shown by `maestro session ls` and `maestro project ls`.
- `--agent` is an alias for `--harness` on `maestro spawn`.
- Every command accepts `-h / --help` for the full flag list.

See [references.md](references.md) for natural-language-to-command mappings.
