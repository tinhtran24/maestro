# to start

Fetch (if needed) and open the Maestro desktop app. The desktop app owns the daemon, state, and updates. `maestro start` no longer runs a daemon: it resolves the installed app (or downloads the latest release), opens it, and exits.

## Syntax

```
maestro start [flags]
```

## Flags

| Flag | Meaning | Default / Required |
|---|---|---|
| `--json` | Output start result as JSON | - |

## Examples

```bash
# Open the Maestro desktop app
maestro start
```

```bash
# Open the app and get the result as JSON
maestro start --json
```
