# Quick Reference

Natural-language-to-command mappings for common Maestro tasks.

| You want to... | Command |
|---|---|
| Show me this webpage / open this page | `maestro preview "<url>"` |
| Spawn a worker on issue N | `maestro spawn --project <p> --issue N --name "<=20 chars>" --prompt "..."` |
| Message a running agent | `maestro send --session <id> --message "..."` |
| Kill a session | `maestro session kill <id>` |
| List sessions | `maestro session ls` |
| Register a repo as a project | `maestro project add --path <abs-path> --name <name>` |
| List projects | `maestro project ls` |
| Rename a session | `maestro session rename <id> "<name>"` |
| Restore a killed session | `maestro session restore <id>` |
| Clean up terminated sessions | `maestro session cleanup` |
| See a session's details | `maestro session get <id>` |
| Open the desktop app | `maestro start` |
| Check the daemon is up | `maestro status` |
| Run health checks | `maestro doctor` |
| Clear the preview panel | `maestro preview clear` |
| List orchestrator sessions | `maestro orchestrator ls` |
| Claim an existing PR for a session | `maestro session claim-pr <id> <pr-ref>` |
| Submit a code review verdict | `maestro review submit <session-id> --run <run-id> --verdict approved` |
| Configure a project's default branch or model | `maestro project set-config <id> --default-branch <branch> --model <model>` |
| Import projects from a legacy Maestro install | `maestro import --dry-run` (preview), then `maestro import -y` |
