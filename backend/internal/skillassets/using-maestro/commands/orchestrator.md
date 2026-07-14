# to orchestrator

Manage orchestrator sessions.

## Syntax

```
maestro orchestrator <subcommand> [flags]
```

## Subcommands

---

### to orchestrator ls

List orchestrator sessions. Aliases: `ls`, `list`.

**Syntax:**
```
maestro orchestrator ls [flags]
```

**Flags:**

| Flag | Meaning | Default / Required |
|---|---|---|
| `--json` | Output as JSON | - |

## Examples

```bash
# List all orchestrator sessions
maestro orchestrator ls
```

```bash
# List orchestrator sessions as JSON
maestro orchestrator ls --json
```
