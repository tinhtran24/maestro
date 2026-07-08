# to doctor

Run local Thanos health checks. Use this to diagnose setup problems or verify the environment is correctly configured.

## Syntax

```
to doctor [flags]
```

## Flags

| Flag | Meaning | Default / Required |
|---|---|---|
| `--json` | Output health checks as JSON | - |

## Examples

```bash
# Run health checks
to doctor
```

```bash
# Get health check results as JSON
to doctor --json
```
