# to stop

Stop the Thanos daemon.

## Syntax

```
to stop [flags]
```

## Flags

| Flag | Meaning | Default / Required |
|---|---|---|
| `--json` | Output stop result as JSON | - |
| `--timeout duration` | How long to wait for daemon shutdown | `10s` |

## Examples

```bash
# Stop the daemon
to stop
```

```bash
# Stop with a longer timeout
to stop --timeout 30s
```
