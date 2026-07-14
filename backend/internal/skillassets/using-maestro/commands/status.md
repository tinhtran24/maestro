# to status

Show Maestro daemon status. Use this to verify the daemon is up and check which port it is bound to.

## Syntax

```
maestro status [flags]
```

## Flags

| Flag | Meaning | Default / Required |
|---|---|---|
| `--json` | Output status as JSON | - |

## Examples

```bash
# Check daemon status
maestro status
```

```bash
# Get status as JSON (e.g. to check port programmatically)
maestro status --json
```
