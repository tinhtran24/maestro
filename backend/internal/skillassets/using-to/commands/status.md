# to status

Show Thanos daemon status. Use this to verify the daemon is up and check which port it is bound to.

## Syntax

```
to status [flags]
```

## Flags

| Flag | Meaning | Default / Required |
|---|---|---|
| `--json` | Output status as JSON | - |

## Examples

```bash
# Check daemon status
to status
```

```bash
# Get status as JSON (e.g. to check port programmatically)
to status --json
```
