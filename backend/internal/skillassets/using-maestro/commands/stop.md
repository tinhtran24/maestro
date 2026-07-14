# to stop

Stop the Maestro daemon.

## Syntax

```
maestro stop [flags]
```

## Flags

| Flag | Meaning | Default / Required |
|---|---|---|
| `--json` | Output stop result as JSON | - |
| `--timeout duration` | How long to wait for daemon shutdown | `10s` |

## Examples

```bash
# Stop the daemon
maestro stop
```

```bash
# Stop with a longer timeout
maestro stop --timeout 30s
```
