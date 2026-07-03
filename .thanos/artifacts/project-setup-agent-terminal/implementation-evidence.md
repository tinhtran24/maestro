# Implementation Evidence: Project Setup, Agent Settings, Terminal Runner

Created: 2026-07-03T02:10:02Z

## Changed Files

- `apps/thanos-desktop/src/routes/WorkbenchRoute.tsx`
- `apps/thanos-desktop/src/shared/ui/AppShell.tsx`
- `apps/thanos-desktop/src/shared/ui/SidebarItem.tsx`
- `apps/thanos-desktop/src/domain/models.ts`
- `apps/thanos-desktop/src/services/nativeBackend.ts`
- `apps/thanos-desktop/src/state/workbenchStore.ts`
- `apps/thanos-desktop/src/flows/project-onboarding-flow/*`
- `apps/thanos-desktop/src/flows/agent-settings-flow/*`
- `apps/thanos-desktop/src/flows/agent-terminal-flow/*`
- `apps/thanos-desktop/src/flows/task-workbench-flow/*`
- `apps/thanos-desktop/src-tauri/src/lib.rs`
- `internal/projects/*`
- `internal/agents/*`
- `internal/terminal/*`
- `internal/workflow/*`
- `internal/workbench/*`
- `internal/orchestrator/workbench_workflow.go`

## Behavior Summary

- Replaced hardcoded desktop workspace path with a persisted selected workspace.
- Added onboarding for no project, including New Project, Import Repo, and Open Recent entry points.
- Added project setup persistence for `.thanos/settings.json` and `.thanos/config.json`.
- Added richer project metadata: remote URL, default branch, worktree root, package manager, dev command, and test command.
- Added Settings > Agents and Settings > Workflow Steps screens.
- Added local agent detection for `claude`, `codex`, `gemini`, `opencode`, `cursor`, `aider`, and `goose`.
- Added step configuration for planning, coding, review, testing, debugging, documentation, and memory update.
- Added terminal tabs and task step panel with approval-aware start buttons.
- Enabled xterm input passthrough to the backend PTY.
- Added `waiting_user` task status in frontend and Go workbench lifecycle.

## Commands Run

- `GOCACHE=/private/tmp/thanos-go-build go test ./internal/projects ./internal/agents ./internal/terminal ./internal/workflow ./internal/workbench ./internal/orchestrator`
  - Result: passed.
- `GOCACHE=/private/tmp/thanos-go-build go test ./...`
  - Result: passed.
- `cargo check` in `apps/thanos-desktop/src-tauri`
  - Result: passed.
- `npm run build` in `apps/thanos-desktop`
  - Result: passed.

## Known Limitations

- Backend PTY still runs one active native process at a time, while the UI now tracks multiple task-linked session records as tabs.
- Workflow step settings are available in UI/store and default project config, but full persistent editing back into `.thanos/config.json` is a follow-up.
