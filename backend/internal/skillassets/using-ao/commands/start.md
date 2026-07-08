# to start

Fetch (if needed) and open the Thanos desktop app. The desktop app owns the daemon, state, and updates. `to start` no longer runs a daemon: it resolves the installed app (or downloads the latest release), opens it, and exits.

## Syntax

```
to start [flags]
```

## Flags

| Flag | Meaning | Default / Required |
|---|---|---|
| `--json` | Output start result as JSON | - |

## Examples

```bash
# Open the Thanos desktop app
to start
```

```bash
# Open the app and get the result as JSON
to start --json
```
