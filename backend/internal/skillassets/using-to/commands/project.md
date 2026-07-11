# to project

Manage projects: register repos, inspect, configure per-project settings, and remove.

## Syntax

```
to project <subcommand> [args] [flags]
```

## Subcommands

---

### to project add

Register a local git repo as a project so sessions can be spawned in it. The path must be an existing git repository on disk. With `--as-workspace`, the path may be a parent folder containing direct child git repositories; Thanos initializes/adopts the parent as the root repo and gitignores children.

**Syntax:**
```
to project add [flags]
```

**Flags:**

| Flag | Meaning | Default / Required |
|---|---|---|
| `--as-workspace` | Register a parent folder as a workspace project (root-as-repo plus direct child repos) | - |
| `--id string` | Project id | Derived by the daemon from the path |
| `--name string` | Display name | - |
| `--orchestrator-agent string` | Default orchestrator session agent | - |
| `--path string` | Absolute path to the local git repo | Required |
| `--worker-agent string` | Default worker session agent | - |

**Examples:**

```bash
# Register a repo as a project
to project add --path /Users/harshit/Downloads/side-quests/thanos --name "thanos"
```

```bash
# Register a workspace (parent folder containing multiple repos)
to project add --path /Users/harshit/Downloads/side-quests --as-workspace --name "side-quests"
```

---

### to project ls

List registered projects. Aliases: `ls`, `list`.

**Syntax:**
```
to project ls [flags]
```

**Flags:**

| Flag | Meaning | Default / Required |
|---|---|---|
| `--json` | Output projects as JSON | - |

**Examples:**

```bash
# List all registered projects
to project ls
```

---

### to project get

Fetch one registered project.

**Syntax:**
```
to project get <id> [flags]
```

**Flags:**

| Flag | Meaning | Default / Required |
|---|---|---|
| `--json` | Output project as JSON | - |

**Examples:**

```bash
# Get details for the thanos project
to project get thanos
```

---

### to project rm

Remove a registered project. Aliases: `rm`, `remove`, `delete`.

**Syntax:**
```
to project rm <id> [flags]
```

**Flags:**

| Flag | Meaning | Default / Required |
|---|---|---|
| `--json` | Output removal result as JSON | - |
| `-y, --yes` | Skip confirmation prompt | - |

**Examples:**

```bash
# Remove a project (with confirmation)
to project rm thanos
```

```bash
# Remove without prompt
to project rm thanos -y
```

---

### to project set-config

Replace a project's per-project config (branch, session prefix, env, symlinks, post-create, agent model/permissions, role overrides). The config is resolved when a session spawns. Set fields via flags, pass the whole object with `--config-json`, or `--clear` to remove all config.

**Syntax:**
```
to project set-config <id> [flags]
```

**Flags:**

| Flag | Meaning | Default / Required |
|---|---|---|
| `--clear` | Clear all config | - |
| `--config-json string` | Full config as a JSON object (overrides field flags) | - |
| `--default-branch string` | Base branch new session worktrees are created from | - |
| `--env stringArray` | Env var `KEY=VALUE` forwarded into sessions (repeatable) | - |
| `--json` | Output the updated project as JSON | - |
| `--model string` | Agent model override (e.g. `claude-opus-4-5`) | - |
| `--orchestrator-agent string` | Harness override for orchestrator sessions | - |
| `--permission string` | Permission mode: `default`, `accept-edits`, `auto`, `bypass-permissions` | - |
| `--post-create stringArray` | Command to run after workspace creation (repeatable) | - |
| `--session-prefix string` | Displayed session-id prefix | - |
| `--symlink stringArray` | Repo-relative path to symlink into workspaces (repeatable) | - |
| `--worker-agent string` | Harness override for worker sessions | - |

**Examples:**

```bash
# Set default branch and model for a project
to project set-config thanos --default-branch main --model claude-opus-4-5
```

```bash
# Set an env var and a post-create command
to project set-config thanos --env "NODE_ENV=development" --post-create "npm install"
```
