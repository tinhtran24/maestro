# to import

Import reads the legacy Maestro flat-file store (`~/.maestro`) read-only and ports its projects and per-project settings into the rewrite database. Legacy files are never modified, and a re-run skips rows that already exist, so it is safe to run more than once. The daemon must be stopped before running: it is the sole writer of the database.

## Syntax

```
maestro import [flags]
```

## Flags

| Flag | Meaning | Default / Required |
|---|---|---|
| `--dry-run` | Parse and report the planned import without writing | - |
| `--from string` | Legacy Maestro root to read | `~/.maestro` |
| `--json` | Output the import report as JSON | - |
| `-y, --yes` | Skip the confirmation prompt (for non-interactive use) | - |

## Examples

```bash
# Preview what would be imported without writing anything
maestro import --dry-run
```

```bash
# Run the import non-interactively
maestro import -y
```

```bash
# Import from a custom legacy path
maestro import --from /tmp/old-maestro -y
```
