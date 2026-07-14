# to session

Manage agent sessions: list, inspect, rename, kill, restore, clean up, and claim PRs.

## Syntax

```
maestro session <subcommand> [args] [flags]
```

## Subcommands

---

### to session ls

List sessions.

**Syntax:**
```
maestro session ls [flags]
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
maestro session ls
```

```bash
# List all sessions including terminated, scoped to one project
maestro session ls --include-terminated -p maestro
```

---

### to session get

Fetch one session.

**Syntax:**
```
maestro session get <id> [flags]
```

**Flags:**

| Flag | Meaning | Default / Required |
|---|---|---|
| `--json` | Output as JSON | - |
| `-p, --project string` | Project id to scope the lookup | - |

**Examples:**

```bash
# Get details for session mer-3
maestro session get mer-3
```

```bash
# Get session details as JSON
maestro session get mer-3 --json
```

---

### to session kill

Terminate a session.

**Syntax:**
```
maestro session kill <id> [flags]
```

**Flags:**

| Flag | Meaning | Default / Required |
|---|---|---|
| `-p, --project string` | Project id to scope the lookup | - |

**Examples:**

```bash
# Kill session mer-3
maestro session kill mer-3
```

---

### to session rename

Rename a session.

**Syntax:**
```
maestro session rename <id> <name> [flags]
```

**Flags:**

| Flag | Meaning | Default / Required |
|---|---|---|
| `-p, --project string` | Project id to scope the lookup | - |

**Examples:**

```bash
# Rename session mer-3 to a new display name
maestro session rename mer-3 "fix-auth-bug"
```

---

### to session restore

Relaunch a terminated session.

**Syntax:**
```
maestro session restore <id> [flags]
```

**Flags:**

| Flag | Meaning | Default / Required |
|---|---|---|
| `-p, --project string` | Project id to scope the lookup | - |

**Examples:**

```bash
# Restore a terminated session
maestro session restore mer-3
```

---

### to session cleanup

Clean up terminated sessions by reclaiming eligible workspaces. Dirty worktrees are skipped by the daemon.

**Syntax:**
```
maestro session cleanup [flags]
```

**Flags:**

| Flag | Meaning | Default / Required |
|---|---|---|
| `-p, --project string` | Filter by project ID | - |
| `-y, --yes` | Skip confirmation prompt | - |

**Examples:**

```bash
# Clean up all terminated sessions (skip prompt)
maestro session cleanup -y
```

```bash
# Clean up terminated sessions for one project
maestro session cleanup -p maestro
```

---

### to session claim-pr

Attach an existing PR to a session.

**Syntax:**
```
maestro session claim-pr <session-id> <pr-ref> [flags]
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
maestro session claim-pr mer-3 88
```

```bash
# Claim PR 88 but refuse if another session already owns it
maestro session claim-pr mer-3 88 --no-takeover
```
