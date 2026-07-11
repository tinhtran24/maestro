# to orchestrator

Manage orchestrator sessions.

## Syntax

```
to orchestrator <subcommand> [flags]
```

## Subcommands

---

### to orchestrator ls

List orchestrator sessions. Aliases: `ls`, `list`.

**Syntax:**
```
to orchestrator ls [flags]
```

**Flags:**

| Flag | Meaning | Default / Required |
|---|---|---|
| `--json` | Output as JSON | - |

## Examples

```bash
# List all orchestrator sessions
to orchestrator ls
```

```bash
# List orchestrator sessions as JSON
to orchestrator ls --json
```
