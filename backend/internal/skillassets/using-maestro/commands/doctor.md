# to doctor

Run local Maestro health checks. Use this to diagnose setup problems or verify the environment is correctly configured.

## Syntax

```
maestro doctor [flags]
```

## Flags

| Flag | Meaning | Default / Required |
|---|---|---|
| `--json` | Output health checks as JSON | - |

## Examples

```bash
# Run health checks
maestro doctor
```

```bash
# Get health check results as JSON
maestro doctor --json
```
