# Quick Reference

Natural-language-to-command mappings for common Thanos tasks.

| You want to... | Command |
|---|---|
| Show me this webpage / open this page | `to preview "<url>"` |
| Spawn a worker on issue N | `to spawn --project <p> --issue N --name "<=20 chars>" --prompt "..."` |
| Message a running agent | `to send --session <id> --message "..."` |
| Kill a session | `to session kill <id>` |
| List sessions | `to session ls` |
| Register a repo as a project | `to project add --path <abs-path> --name <name>` |
| List projects | `to project ls` |
| Rename a session | `to session rename <id> "<name>"` |
| Restore a killed session | `to session restore <id>` |
| Clean up terminated sessions | `to session cleanup` |
| See a session's details | `to session get <id>` |
| Open the desktop app | `to start` |
| Check the daemon is up | `to status` |
| Run health checks | `to doctor` |
| Clear the preview panel | `to preview clear` |
| List orchestrator sessions | `to orchestrator ls` |
| Claim an existing PR for a session | `to session claim-pr <id> <pr-ref>` |
| Submit a code review verdict | `to review submit <session-id> --run <run-id> --verdict approved` |
| Configure a project's default branch or model | `to project set-config <id> --default-branch <branch> --model <model>` |
| Import projects from a legacy Thanos install | `to import --dry-run` (preview), then `to import -y` |
