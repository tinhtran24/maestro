# to orchestrator

Manage orchestrator sessions.

## Syntax

```
to orchestrator <subcommand> [flags]
```

## Subcommands

---

### to orchestrator ls

List orchestrator sessions. Aliases: `ls`, `list`.

**Syntax:**
```
to orchestrator ls [flags]
```

**Flags:**

| Flag | Meaning | Default / Required |
|---|---|---|
| `--json` | Output as JSON | - |

## Examples

```bash
# List all orchestrator sessions
to orchestrator ls
```

```bash
# List orchestrator sessions as JSON
to orchestrator ls --json
```

### Completion protocol

Workers finish coding with `to session complete`; they never transition a task.
The active orchestrator records each successfully completed checkpoint in order:

```bash
to orchestrator finalize <worker-id> --state verifying_git
to orchestrator finalize <worker-id> --state testing
to orchestrator finalize <worker-id> --state committing
to orchestrator finalize <worker-id> --state pushing
to orchestrator finalize <worker-id> --state claiming_pr
to orchestrator finalize <worker-id> --state persisting_metadata
to orchestrator finalize <worker-id> --state cleaning_runtime
to orchestrator finalize <worker-id> --state review_pending
to task done <worker-id>
```

The daemon rejects skipped steps, duplicate side effects, non-orchestrator callers,
and attempts by another orchestrator to take over an active claim.
