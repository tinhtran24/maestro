# to import

Import reads the legacy Thanos flat-file store (`~/.thanos`) read-only and ports its projects and per-project settings into the rewrite database. Legacy files are never modified, and a re-run skips rows that already exist, so it is safe to run more than once. The daemon must be stopped before running: it is the sole writer of the database.

## Syntax

```
to import [flags]
```

## Flags

| Flag | Meaning | Default / Required |
|---|---|---|
| `--dry-run` | Parse and report the planned import without writing | - |
| `--from string` | Legacy Thanos root to read | `~/.thanos` |
| `--json` | Output the import report as JSON | - |
| `-y, --yes` | Skip the confirmation prompt (for non-interactive use) | - |

## Examples

```bash
# Preview what would be imported without writing anything
to import --dry-run
```

```bash
# Run the import non-interactively
to import -y
```

```bash
# Import from a custom legacy path
to import --from /tmp/old-thanos -y
```
