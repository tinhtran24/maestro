# to session

Manage agent sessions: list, inspect, rename, kill, restore, clean up, and claim PRs.

## Syntax

```
to session <subcommand> [args] [flags]
```

## Subcommands

---

### to session ls

List sessions.

**Syntax:**
```
to session ls [flags]
```

**Flags:**

| Flag | Meaning | Default / Required |
|---|---|---|
| `-a, --all` | Include orchestrator sessions | - |
| `--include-terminated` | Include terminated sessions | - |
| `--json` | Output as JSON | - |
| `-p, --project string` | Filter by project ID | - |

**Examples:**

```bash
# List all active worker sessions
to session ls
```

```bash
# List all sessions including terminated, scoped to one project
to session ls --include-terminated -p thanos
```

---

### to session get

Fetch one session.

**Syntax:**
```
to session get <id> [flags]
```

**Flags:**

| Flag | Meaning | Default / Required |
|---|---|---|
| `--json` | Output as JSON | - |
| `-p, --project string` | Project id to scope the lookup | - |

**Examples:**

```bash
# Get details for session mer-3
to session get mer-3
```

```bash
# Get session details as JSON
to session get mer-3 --json
```

---

### to session kill

Terminate a session.

**Syntax:**
```
to session kill <id> [flags]
```

**Flags:**

| Flag | Meaning | Default / Required |
|---|---|---|
| `-p, --project string` | Project id to scope the lookup | - |

**Examples:**

```bash
# Kill session mer-3
to session kill mer-3
```

---

### to session rename

Rename a session.

**Syntax:**
```
to session rename <id> <name> [flags]
```

**Flags:**

| Flag | Meaning | Default / Required |
|---|---|---|
| `-p, --project string` | Project id to scope the lookup | - |

**Examples:**

```bash
# Rename session mer-3 to a new display name
to session rename mer-3 "fix-auth-bug"
```

---

### to session restore

Relaunch a terminated session.

**Syntax:**
```
to session restore <id> [flags]
```

**Flags:**

| Flag | Meaning | Default / Required |
|---|---|---|
| `-p, --project string` | Project id to scope the lookup | - |

**Examples:**

```bash
# Restore a terminated session
to session restore mer-3
```

---

### to session cleanup

Clean up terminated sessions by reclaiming eligible workspaces. Dirty worktrees are skipped by the daemon.

**Syntax:**
```
to session cleanup [flags]
```

**Flags:**

| Flag | Meaning | Default / Required |
|---|---|---|
| `-p, --project string` | Filter by project ID | - |
| `-y, --yes` | Skip confirmation prompt | - |

**Examples:**

```bash
# Clean up all terminated sessions (skip prompt)
to session cleanup -y
```

```bash
# Clean up terminated sessions for one project
to session cleanup -p thanos
```

---

### to session claim-pr

Attach an existing PR to a session.

**Syntax:**
```
to session claim-pr <session-id> <pr-ref> [flags]
```

**Flags:**

| Flag | Meaning | Default / Required |
|---|---|---|
| `--json` | Output as JSON | - |
| `--no-takeover` | Refuse if another active session owns the PR | - |
| `-p, --project string` | Project id to scope the lookup | - |

**Examples:**

```bash
# Attach PR 88 to session mer-3
to session claim-pr mer-3 88
```

```bash
# Claim PR 88 but refuse if another session already owns it
to session claim-pr mer-3 88 --no-takeover
```
