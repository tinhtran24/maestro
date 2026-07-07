# Task Plan: Convert Wallfacer Production Behavior Into Thanos

## Objective

Build Thanos into a Wails-native, local-first AI development workbench with the
production behaviors represented in the `wallfacer/` reference project, while
keeping Thanos implementation independent and Thanos-native.

Wallfacer is a reference for product behavior and architecture shape only. Do
not copy implementation code.

## Product Scope

Thanos must support these production surfaces:

- Workspace management
- Board and task lifecycle
- Plan/spec mode
- Agent graph and flows
- Native terminal/runtime execution
- Per-task git worktrees
- Live logs and event timelines
- Oversight summaries
- Test verification
- Routines and automation
- Mission Control graph
- Usage and cost analytics
- Whiteboard
- Local docs
- File explorer
- Configuration and provider routing

## Architecture Target

```text
app/
  app/                Wails shell, RealProvider bindings, native dialogs/events
  frontend/           React workbench UI

internal/
  realprovider/       workspace loader, stores, provider registry
  store/              filesystem-first task/spec/routine/event persistence
  harness/            provider command adapters
  terminal/           native PTY/session manager
  runner/             task turn loop, worktree execution, logs
  flow/               agent role and flow registry/executor
  specs/              spec tree, lifecycle, dispatch
  automation/         autoimplement/test/submit/retry/routines
  oversight/          summaries, risk and diff review
  mission/            unified spec/task graph
```

No Thanos CLI is required for production behavior. Wails owns the app boundary.

## Milestones

### 1. Filesystem Store Foundation

Status: Done.

Goal: make `.thanos/` the durable local data root.

Tasks:

- Define schema versions for task, spec, routine, event, automation, provider,
  agent, flow, usage, and terminal-session records.
- Write atomic JSON helpers using temp file + rename.
- Add append-only JSONL event log.
- Add store migration and validation.
- Add workspace-scoped locking.

Acceptance:

- Loading a workspace never requires a server or CLI.
- Corrupt records are reported as diagnostics without crashing the app.
- Unit tests cover round-trip read/write and migration.

### 2. Workspace Management

Status: Done.

Goal: support multiple local workspaces like Wallfacer workspaces.

Tasks:

- Persist workspace records under user config.
- Support create, update, delete, activate workspace.
- Support multiple folders per workspace.
- Add folder browser and native folder picker.
- Track stable workspace data key separate from folder path.

Acceptance:

- Switching workspaces reloads board/specs/providers without app restart.
- Renaming/moving a workspace does not lose task history.

### 3. Task Board Production Lifecycle

Status: Done.

Goal: replace simple task JSON with production task lifecycle.

Tasks:

- Implement statuses: `backlog`, `in_progress`, `waiting`, `committing`,
  `done`, `failed`, `cancelled`.
- Add archived and deleted/tombstone flags.
- Add dependencies and blocked-state detection.
- Add prompt history, feedback, retry history, failure category.
- Add batch task creation and dependency wiring.
- Add task search.

Acceptance:

- Invalid transitions are rejected in backend logic.
- Task timeline records every state change.
- Board can show active, archived, deleted, dependency-blocked tasks.

### 4. Native Terminal Runtime

Status: Done.

Goal: run provider commands through native terminal/PTY, not auth-token model
runtime.

Tasks:

- Replace current stdout-pipe runner with PTY-backed sessions.
- Support start, stop, resize, write input.
- Stream stdout/stderr to Wails events.
- Persist transcript per session.
- Link sessions to task, flow step, provider, cwd.
- Strip auth token env vars by default; allow explicit user-approved env later.

Acceptance:

- Provider output streams live in UI.
- User can stop a process.
- Session transcript survives reload.
- No API/OAuth token is injected by Thanos.

### 5. Harness/Provider Registry

Status: Done.

Goal: Wallfacer-style harness routing with Thanos-native adapters.

Tasks:

- Add adapters for Claude Code, Codex, Gemini CLI, OpenCode, Cursor Agent,
  Aider, Goose, Shell.
- Detect executable path and version.
- Build provider-specific argv for task prompt, model, cwd, permissions.
- Parse known JSON/NDJSON event streams into normalized terminal events.
- Keep unknown output as raw logs.

Acceptance:

- Installed providers are detected from PATH.
- Missing providers show actionable setup hints.
- A task flow can pin different providers per role.

### 6. Git Worktree Execution

Status: Done.

Goal: safe per-task execution isolation.

Tasks:

- Create task branch and worktree on run.
- Record base commit and branch name.
- Write board manifest visible to task runtime.
- Support sync/rebase onto default branch.
- Capture diff against default branch.
- Clean up cancelled worktrees only after user confirmation.

Acceptance:

- No task mutates the main worktree directly.
- Multiple tasks can run concurrently in separate worktrees.
- Diff panel shows task output before completion.

### 7. Agent Roles and Flow Engine

Status: Done.

Goal: implement Wallfacer Agent Graph behavior.

Tasks:

- Define built-in roles: implementation, testing, title, oversight,
  commit-message.
- Persist user-authored agents under `.thanos/agents/`.
- Persist user-authored flows under `.thanos/flows/`.
- Support fixed sequence flows.
- Support parallel fan-out groups.
- Add flow fallback to `implement`.
- Add UI for clone/edit/delete user agents and flows. Deferred to the Agent
  Graph editing surface; backend now loads read-only built-ins and hot-loads
  user-authored definitions from disk.

Acceptance:

- Built-ins are read-only.
- User agents/flows hot-reload after file changes.
- A task selects one flow and execution follows that flow.

### 8. Runner Turn Loop

Status: Done.

Goal: production task execution semantics.

Tasks:

- Start provider process in task worktree.
- Save per-turn stdout/stderr.
- Parse stop reason when available.
- Support waiting state and user feedback resume.
- Support max token/pause auto-continue when provider supports it.
- Track usage/cost when provider reports it.
- Classify failures.

Acceptance:

- Task can move backlog -> in_progress -> waiting.
- User feedback resumes the same task.
- Outputs and usage are visible in task detail.

### 9. Test Verification

Status: Done.

Goal: add test-agent verification behavior.

Tasks:

- Trigger test run from waiting task.
- Run selected test provider or shell command in worktree.
- Parse PASS/FAIL verdict.
- Store last test result and test output.
- Allow custom pass/fail patterns.
- Gate done/submit on passing verification when enabled.

Acceptance:

- Failed test keeps task out of done.
- Test output is separated from implementation output.

### 10. Commit and Submit Pipeline

Status: Done.

Goal: safely finish task output.

Tasks:

- Generate commit message through commit-message role or user input.
- Commit task worktree changes.
- Merge or cherry-pick only after explicit user approval.
- Optional push/PR creation later.
- Never auto-merge without enabled automation and review gates.

Acceptance:

- User sees diff and commit message before finish.
- Done task records commit hash and summary.

### 11. Oversight

Status: Done.

Goal: Wallfacer-style review visibility.

Tasks:

- Generate oversight summary after waiting/done/failed.
- Persist `oversight.json` and `oversight-test.json`.
- Include phases, risks, changed files, commands, test result, usage.
- Render in task detail and Mission Control.

Acceptance:

- Every completed run has a reviewable oversight artifact.
- Oversight can be regenerated.

### 12. Plan/Spec Mode

Status: Done.

Goal: recursive spec tree and planning workflow.

Tasks:

- Parse `specs/**/*.md` with frontmatter/state.
- Implement states: vague, drafted, validated, testing, complete, stale,
  archived.
- Add slash commands: create, refine, validate, break-down, dispatch.
- Dispatch leaf specs to board tasks.
- Add undo for planning changes via git revert or local event rollback.
- Detect stale candidates from changed files/spec dependencies.

Acceptance:

- Spec tree updates live.
- Leaf specs can dispatch tasks with dependencies.
- Plan changes are inspectable and reversible.

### 13. Chat Sessions

Status: Pending.

Goal: dedicated workspace chat and task-mode chat.

Tasks:

- Persist chat sessions and messages.
- Support multiple named threads.
- Support visible/archived/active states.
- Support @file and @task context mentions.
- Route messages to native provider terminal/session.
- Support interrupt and clear history.

Acceptance:

- Chat survives reload.
- Task-mode chat can update a task prompt with reviewable events.

### 14. Routines and Automation

Status: Next.

Goal: scheduled recurring task generation and guarded automation.

Tasks:

- Persist routines with schedule, prompt, flow, enabled flag.
- Add manual trigger.
- Add scheduler loop.
- Implement autoimplement, autotest, autosubmit, autoretry.
- Add circuit breakers and concurrency limits.
- Respect dependencies and scheduled time.

Acceptance:

- Enabled routine can spawn task instances.
- Automation toggles are visible and persisted.
- Circuit breaker stops runaway loops.

### 15. Mission Control

Goal: unified spec/task graph.

Tasks:

- Build graph nodes for specs, tasks, routines, providers, flows.
- Add typed edges: dispatches, depends_on, implements, blocked_by, produced.
- Compute critical path and blocked set.
- Render actionable graph UI.

Acceptance:

- User can inspect where work is stuck.
- Graph links to task/spec details.

### 16. Usage and Analytics

Goal: track cost, time, and provider activity.

Tasks:

- Persist per-turn usage JSONL.
- Aggregate by task, provider, role, flow, day.
- Show usage dashboard.
- Track terminal/runtime durations and exit codes.

Acceptance:

- Dashboard reflects real task artifacts, not mock values.
- Missing provider usage is shown as unknown, not zero.

### 17. File Explorer

Goal: inspect and edit workspace files.

Tasks:

- Add tree listing with ignore rules.
- Read/write file endpoint.
- Add file watcher events.
- Add task prompt virtual entries.
- Add diff preview for task worktree.

Acceptance:

- User can inspect files used by tasks.
- Writes are explicit and evented.

### 18. Whiteboard

Goal: per-workspace persisted whiteboard.

Tasks:

- Store whiteboard document under `.thanos/whiteboard.json`.
- Add save/load Wails methods.
- Add image/file references later.

Acceptance:

- Whiteboard survives reload and workspace switch.

### 19. Local Docs

Goal: in-app documentation for Thanos behavior.

Tasks:

- Add docs index.
- Render local markdown docs.
- Document providers, workflows, automation, safety gates.

Acceptance:

- Docs are available without network.

### 20. Production Hardening

Goal: make the system dependable.

Tasks:

- Add crash recovery for in-progress tasks.
- Reconcile orphaned processes/worktrees.
- Add log rotation and artifact retention.
- Add store validation command in UI.
- Add broad unit/integration tests.
- Add Playwright smoke tests for major surfaces.

Acceptance:

- Restarting the app does not lose running/failed task evidence.
- Recovery actions are visible and reviewable.

## Suggested Implementation Order

1. Store foundation
2. Workspace manager
3. Task lifecycle and events
4. Native PTY terminal
5. Provider/harness adapters
6. Git worktrees
7. Flow engine
8. Runner turn loop
9. Test verification
10. Oversight
11. Plan/spec mode
12. Routines and automation
13. Mission Control
14. Analytics
15. File explorer
16. Whiteboard/docs
17. Production recovery and E2E

## Definition of Done

Thanos reaches production parity when a user can:

1. Create or select a workspace.
2. Write specs and dispatch them to tasks.
3. Compose agent flows.
4. Run tasks through installed local providers in isolated worktrees.
5. Watch native terminal output live.
6. Review logs, diff, tests, usage, and oversight.
7. Approve or reject completion.
8. Run routines and guarded automation.
9. Recover/review every action from local artifacts.
