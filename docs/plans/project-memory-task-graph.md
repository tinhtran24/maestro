# Plan: Project Memory & Task Graph

## Goal

Give agents durable, compact project memory so a new task can answer, before any
code is written:

- Has this been implemented before?
- Which previous tasks touched this area (symbols / files / tests)?
- Which architectural decisions must not be broken?
- Is this task a duplicate, extension, regression, or bug fix?

The LLM never receives the whole memory. Memory is captured automatically when a
session finishes, projected into a queryable store, then retrieved, ranked,
deduplicated, and compressed into a small role-specific context pack
(target 300-2500 tokens) at prompt-build time.

## Terminology: two different `.maestro` directories (read first)

This is the single most important thing to get right. There are **two** stores,
and they must not be confused:

|           | Committed project memory                                                                       | App-state cache                                                       |
| --------- | ---------------------------------------------------------------------------------------------- | --------------------------------------------------------------------- |
| Location  | `<project repo>/.maestro/memory/` **inside the managed user repo**, tracked by that repo's git | `~/.maestro/cache/<project-id>/` (overridable via `MAESTRO_DATA_DIR`) |
| Contents  | `events.jsonl` (source of truth) + `tasks/TASK-*.mdx` (generated, human-readable)              | `memory.db`, `codegraph.db` (derived SQLite projections)              |
| Lifecycle | append-only, immutable, portable across machines/clones                                        | disposable, rebuildable from `events.jsonl`, per-machine              |
| Git       | committed to the user's project (opt-in auto-commit, see decisions)                            | **never** committed; gitignored                                       |

The AGENTS.md hard rule "all app state lives under `~/.maestro` only" governs the
**second** column. The daemon's derived DBs live under `~/.maestro/cache/` and
never escape it. The **first** column is a new, deliberate exception: project
memory is data _about the user's project_ and belongs _in the user's repo_, exactly
like source code. It is not Maestro app state. Whenever this plan says
"`.maestro/memory/`" unqualified it means the in-repo committed dir; the app cache
is always written as "`~/.maestro/cache/`".

## Core architectural decisions (settled)

1. **The memory event log is NOT `change_log`/`cdc`.** The repo already has a
   mature append-only log (`change_log` table, DB-trigger-sourced, tailed by the
   `internal/cdc` poller and streamed over `GET /api/v1/events`). That log is
   **runtime CDC**: machine-local, under the app data dir, for live UI streaming.
   Project memory is a **different artifact**: project-scoped, committed to the
   user's repo, portable, and human-meaningful. Do not overload `change_log`, do
   not add memory `event_type`s to its CHECK enum, and do not violate the hard rule
   "change events come from DB triggers, not app code" by writing memory rows
   through triggers. We **reuse the patterns** (sqlc store, `inTx`, the
   append-only mindset) for the projection DB, not the table.

2. **`events.jsonl` is the source of truth; SQLite is a disposable projection.**
   Every completed task appends exactly one immutable JSON line. Events are never
   edited or deleted. `memory.db` is a pure function of the event log and can be
   dropped and rebuilt at any time (`maestro memory rebuild`). Git owns the code;
   events own the memory; SQLite is a cache; MDX is for humans.

3. **Completion memory is mandatory; planning is optional.** There is no required
   up-front plan document. Memory is produced at the session terminal hook where
   the session's durable facts are already known (`lifecycle.Manager.MarkTerminated`
   / `session_manager.Manager.Kill`): id, kind, harness, branch, workspace path,
   prompt (intent), and attributed PRs. A git diff of the session branch against
   its base yields the changed files and changed tests.

4. **Capture is best-effort and never blocks or fails a lifecycle transition.**
   `MarkTerminated` must still succeed if memory capture errors (diff fails, repo
   detached, disk full). Memory is an observer of terminal facts, not a gate on
   them. Failures are logged, and the missed task can be backfilled from git later.

5. **Symbol/call-graph is a swappable provider, and starts file-level.** The first
   shipped `CodebaseMemory` provider derives changed **files and tests** from the
   git diff only. True symbol extraction, reference search, call graph, and impact
   analysis (proposal Phase 2/3) sit behind the same interface and are deferred to
   a follow-on plan (see Scope). We ship the interface now so the resolver never
   has to change when the provider gets smarter.

6. **The resolver never sends the whole graph.** Retrieval is: extract task
   entities → query projection + graph → rank → dedupe → token-budget → render
   role-specific pack. A hard token cap is enforced in code, and any truncation is
   logged (never silent).

7. **Auto-commit of `events.jsonl` is opt-in and never pushes.** Writing to the
   user's git history uninvited is invasive. Default: append the file and leave it
   for the user to commit (it shows up as a normal working-tree change). A project
   config flag can enable auto-commit onto the default branch. The daemon never
   `git push`.

8. **No em dashes anywhere** (prose, comments, commit messages, generated MDX).

## Global constraints (binding; reviewers enforce verbatim)

- Derived DBs resolve under `~/.maestro/cache/` only (`MAESTRO_DATA_DIR`
  overridable). Never `~/Library/Application Support`. `events.jsonl` and
  `tasks/*.mdx` resolve under `<project.Path>/.maestro/memory/` only.
- The daemon owns all memory reads/writes. The CLI stays a thin HTTP client
  (no direct DB, git, or filesystem access to memory). The bind host stays
  loopback-only.
- `events.jsonl` is append-only. No code path edits or deletes a prior line.
- `memory.db` is derivable: a rebuild from `events.jsonl` must reproduce it
  byte-for-byte-equivalent (same query results). It is gitignored and never
  committed.
- Memory capture is best-effort and must never fail or delay `MarkTerminated`,
  `Kill`, or `Cleanup`.
- No new `change_log` `event_type`s, no memory writes through DB triggers, no
  parallel manual CDC into `change_log`.
- The context resolver enforces a hard max-token cap and logs any dropped items.
- Migrations for `memory.db` are additive; do not modify a merged migration.
  (These are a **separate** migration set from the app DB, see Data model.)
- New HTTP routes go through the code-first spec flow (`dto.go` +
  `apispec/specgen/build.go` + `npm run api`); do not hand-edit generated files.

## Data model

### Event log (`events.jsonl`, source of truth)

One JSON object per line, append-only, schema-versioned. Minimum Phase 1 shape:

```
{
  "v": 1,
  "id": "EVT-<ulid>",              // stable event id (ULID passed in; scripts have no clock, see note)
  "type": "task.completed",
  "occurred_at": "2026-07-14T...Z",
  "task": {
    "id": "TASK-001",             // derived from session id + sequence
    "session_id": "acme-7",
    "project_id": "acme",
    "kind": "worker",             // worker | orchestrator
    "harness": "...",
    "intent": "<session prompt>",
    "task_type": "feature",       // feature | bugfix | refactor | ... (classified; heuristic ok in P1)
    "branch": "...",
    "base_sha": "...",
    "head_sha": "...",
    "changed_files": ["path", ...],
    "changed_tests": ["path", ...],
    "changed_symbols": [],        // empty in P1; populated when the symbol provider lands
    "prs": [{"url": "...", "number": 12, "state": "merged"}],
    "decisions": []               // protected architectural decisions, if any captured
  }
}
```

Rules: never edit a line; correction is a new event (`task.superseded`,
`decision.retracted`) that the projection folds in. `v` gates schema evolution.

### Projection (`~/.maestro/cache/<project-id>/memory.db`)

A **separate SQLite database from the app DB** (own file, own goose migration
set, own sqlc `queries`/`gen` under a new package, own store). It never shares a
connection or migration numbering with `backend/internal/storage/sqlite`. Tables
(Phase 1-3): `memory_task` (one row per completed task), `memory_file` and
`memory_test` (task ↔ path edges), `memory_symbol` + `memory_task_symbol` (when
the symbol provider lands), `memory_decision`, and `memory_task_edge` (task→task
relations with a `relation` kind and `confidence` score). A `memory_meta` row
tracks the last-applied event id so rebuild/catch-up is incremental.

## Component → package mapping

Everything new lives under a new top-level service so it is easy to find and
delete if the experiment fails:

- `backend/internal/service/memory/` — the memory service: event writer,
  projection applier, rebuild, and the `CodebaseMemory` + `RelationResolver` +
  `ContextResolver` interfaces and their Phase 1 implementations.
- `backend/internal/storage/memorydb/` — the derived-DB story mirroring
  `storage/sqlite/` layout (`migrations/`, `queries/`, `gen/`, `store/`, `db.go`),
  but a wholly separate database. Its own `sqlc` config stanza.
- `backend/internal/adapters/memoryevents/` — the `events.jsonl` reader/writer
  (append, iterate, opt-in auto-commit via the existing git worktree/command
  plumbing).
- Capture hook: a thin `MemoryRecorder` seam invoked from
  `session_manager/manager.go` right after `MarkTerminated` succeeds.
- Read surface: `backend/internal/httpd/controllers/memory.go` +
  `backend/internal/cli/memory.go`.

## Scope of this plan

**Committed here (build now):** proposal Phase 1 (completion memory + event log +
projection) end to end, plus the thin vertical slice of Phase 3/4 needed to prove
value: file/test-level relations and a minimal token-budgeted context pack.

**Deferred to follow-on plans (interfaces shipped, impls stubbed):**

- Real symbol search / reference search / call graph / impact analysis
  (`codegraph.db`) — proposal Phase 2 and the symbol half of Phase 3. Large;
  language-specific; its own plan.
- Semantic-similarity task linking (embeddings) — Phase 3.
- MDX generation for humans — Phase 5. (Event schema already carries everything
  MDX needs; renderer is additive.)
- Desktop UI (task history / related tasks / decisions panel) — Phase 6. New
  frontend work over the read endpoint.

Shipping the interfaces now (decision 5) means these follow-ons add providers, not
rewrites.

## Tasks (smallest coherent diff first; each ends with ONE runnable check)

### Task 1 — Event log adapter (`adapters/memoryevents`)

Append-only reader/writer for `<project.Path>/.maestro/memory/events.jsonl`:
`Append(ctx, project, event) error` (create dir, O_APPEND single line + `\n`,
fsync) and `Iterate(ctx, project, from eventID, fn) error`. No auto-commit yet.
ULIDs are supplied by the caller (scripts/daemon have a clock; the adapter does
not mint time itself). **Check:** Go test that appends three events and iterates
them back in order, asserting append-only bytes and that a malformed trailing
line is surfaced, not silently skipped.

### Task 2 — Projection DB (`storage/memorydb`)

Stand up the separate SQLite DB: `db.go` opening `~/.maestro/cache/<project-id>/memory.db`
with its own embedded goose migrations (`0001_init` creating `memory_task`,
`memory_file`, `memory_test`, `memory_task_edge`, `memory_meta`), its own sqlc
stanza and generated store. **Check:** Go test opening a temp DB, migrating,
inserting a task + file edges, and reading them back; assert the file lives under
the configured cache dir and nowhere near the app DB.

### Task 3 — Projection applier + rebuild

`Apply(ctx, event)` folds one event into `memory.db` idempotently (keyed on event
id in `memory_meta`); `Rebuild(ctx, project)` drops and replays the whole
`events.jsonl`. **Check:** Go test that applies N events, snapshots query results,
drops the DB, rebuilds from the log, and asserts identical results; re-applying
the same event twice is a no-op.

### Task 4 — Completion memory builder + capture hook

- `service/memory.Builder.BuildCompletion(ctx, record) (Event, error)`: from a
  terminal `SessionRecord` gather intent/branch/kind/harness/PRs, run a git diff
  of the session branch vs its base (reuse the gitworktree command plumbing) to
  derive `changed_files` + `changed_tests`, classify `task_type` heuristically
  (branch/commit prefix), leave `changed_symbols` empty.
- Wire a `MemoryRecorder` seam into `session_manager` invoked after
  `MarkTerminated` returns: `Append` then `Apply`, best-effort (log on error,
  never propagate). No kind filter.
- **Check:** Go test with fakes asserting (a) capture runs after MarkTerminated,
  (b) a capture error does not fail the kill/terminate path, (c) both worker and
  orchestrator sessions produce an event, (d) changed tests are split out from
  changed files.

### Task 5 — Relation resolver (file/test level)

`RelationResolver.Related(ctx, task) []RelatedTask`: given a new task's changed
files/tests (or, pre-execution, files inferred from the prompt), query
`memory_task_edge`/`memory_file` for prior tasks sharing paths, rank by overlap +
recency, dedupe, and cap to top-K. Populate `memory_task_edge` at apply time
(Task 3) when a new task shares paths with an existing one, with a `confidence`
score. **Check:** Go test seeding three historical tasks and asserting the
resolver returns the two that share a file, ranked, deduped, K-capped.

### Task 6 — Context resolver + token budget

`ContextResolver.Pack(ctx, task, role) (ContextPack, error)` producing a
role-specific pack (planner / coder / reviewer / tester) from related tasks,
protected decisions, and relevant file/test lists. Enforce a hard token ceiling
(rough token estimate is fine) and **log** any dropped items. **Check:** Go test
asserting the pack stays under the cap for a large history and that dropped items
are reported, not silently omitted; and that each role selects its documented
slice.

### Task 7 — Read surface (HTTP + CLI)

- `GET /api/v1/projects/{id}/memory/tasks` (list) and
  `GET /api/v1/projects/{id}/memory/context?role=&intent=` (context pack), via
  `controllers/memory.go` + a spec-registry entry + `npm run api`.
- `maestro memory list|context|rebuild` cobra commands (thin HTTP client), mirror
  DTOs per house convention.
- **Check:** `cd backend && go test ./internal/httpd/...` (spec-drift + route
  parity green) and a CLI table test for `memory list` happy path + daemon error
  envelope.

### Task 8 — Opt-in auto-commit + gitignore hygiene

Add the project config flag (default off) that, after a successful append,
commits `.maestro/memory/events.jsonl` onto the default branch via existing git
plumbing (never push). Ensure `~/.maestro/cache/` is never committed and that a
managed repo's `.gitignore` is not needed for it (it lives outside the repo).
**Check:** Go test asserting flag-off leaves an uncommitted working-tree change,
flag-on produces exactly one commit and no push; and that the cache DB path is
outside any repo.

## Edge cases the first version must handle

1. Session with no branch/workspace (incomplete): skip capture, no event.
2. Diff fails / base SHA missing: emit an event with empty file lists rather than
   failing; log the reason.
3. `events.jsonl` present but `memory.db` missing or stale: lazily rebuild on
   first read.
4. Two daemons / concurrent appends: single-writer append with O_APPEND; the
   projection applier is idempotent on event id so a double-apply is safe.
5. Orchestrator vs worker: both captured, no kind filter.
6. Corrupt/partial trailing line in `events.jsonl` (crash mid-append): iteration
   surfaces it; rebuild stops at the last valid line and logs the truncation.
7. Renamed/moved project path: memory follows the repo (it is in-repo); the cache
   DB is keyed by project id and rebuilt if absent.

## Key files (verify with grep/codegraph before editing; signatures may drift)

- `backend/internal/session_manager/manager.go` — `Kill` (~555),
  `SaveAndTeardownAll` (~801), `Cleanup` (~1407); `lifecycle/manager.go`
  `MarkTerminated` (~254) is the terminal hook.
- `backend/internal/domain/session.go` — `SessionRecord`, `SessionMetadata`;
  `domain/pr.go` — `PRFacts`; `domain/task_suggestions.go` — task-type heuristic.
- `backend/internal/adapters/workspace/gitworktree/commands.go` — git arg
  builders to reuse for the diff (`add -A`/`write-tree`/`commit-tree`,
  `rev-parse`, `status --porcelain`); `workspace.go` `StashUncommitted` (~335) as
  the pattern for side-effect-free git plumbing.
- `backend/internal/storage/sqlite/db.go`, `sqlc.yaml` — the layout and config to
  mirror for the **separate** `memorydb` (do not extend the app DB).
- `backend/internal/cdc/*` — read as the reference for "append-only log done
  right," but do NOT reuse `change_log` for memory (decision 1).
- `backend/internal/httpd/controllers/dto.go`,
  `backend/internal/httpd/apispec/specgen/build.go`,
  `backend/internal/httpd/router.go` — endpoint + spec wiring.
- `backend/internal/cli/root.go` (subcommand registration ~154),
  `backend/internal/cli/session.go` — CLI command pattern.

## Build & verify commands (from repo root; see AGENTS.md)

- `npm run lint` — backend `go test ./...` + golangci-lint.
- `cd backend && go build ./... && go test -race ./... && go vet ./...`.
- `npm run sqlc` — after adding `memorydb` queries/migrations (extend `sqlc.yaml`
  with the new package; never hand-edit `gen/`).
- `npm run api` — after adding the memory HTTP routes; commit `openapi.yaml` +
  `app/src/api/schema.ts` with the Go changes.
- `npx prettier@3 --check --ignore-unknown <changed text files>`.

## Open decisions to confirm with the user before starting

1. **Auto-commit default and target branch.** Recommended: off by default; when
   on, commit to the repo's default branch, never push (decision 7). Confirm.
2. **In-repo memory path.** Recommended `.maestro/memory/` inside the managed
   repo. Confirm this does not collide with any convention in the user's own
   projects (the daemon must not clobber a pre-existing `.maestro/`).
3. **Task-type classification in Phase 1.** Heuristic from branch/commit prefix
   (`feat`/`fix`/`refactor`) vs deferring types entirely. Recommended: heuristic,
   refined later.
4. **How much of Phase 3/4 to land now.** This plan commits file/test-level
   relations + a minimal pack (Tasks 5-6) and defers symbols/embeddings/MDX/desktop
   to follow-on plans. Confirm that scope split.

## Starting point for the implementing session

- Baseline: commit this plan on `dev` (the repo's main branch per git config), then
  branch `feature/project-memory-task-graph`.
- Execution order: 1 → 2 → 3 → 4 form the load-bearing spine (each depends on the
  prior). 5 → 6 add retrieval. 7 → 8 add surface + hygiene and can follow. Do not
  start a later task before its predecessor's check is green.
- The file:line references are approximate (`~`); confirm signatures with grep or
  codegraph before editing, since the store is sqlc-generated and the daemon is
  loopback-only.
