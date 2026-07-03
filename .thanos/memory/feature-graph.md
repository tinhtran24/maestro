# Thanos Feature Memory

Generated: 2026-07-03T01:31:39Z

- Features: 4
- Relationships: 0

## Agent Skills System

Thanos supports project-level skill contracts loaded from `.thanos/skills`,
matched by role, workflow stage, task type, tags, and files, validated for exit
criteria, tracked through skill runs and evidence, gated in orchestrator
helpers, and surfaced in the desktop task sidebar.

Affected paths:

- `internal/skills`
- `internal/workbench/model.go`
- `internal/workbench/schema.sql`
- `internal/orchestrator/skill_gate.go`
- `apps/thanos-desktop/src/flows/TaskWorkbenchFlow.tsx`
- `.thanos/skills`

Decisions:

- Skills are local workflow contracts, not executable scripts.
- Project skills override global and builtin skills by registry precedence.
- Phase 1 models execution as skill-run lifecycle and evidence validation without remote installs or script execution.

Verification:

- `GOCACHE=/private/tmp/thanos-go-build go test ./...`
- `npm run build`

## Phase 6 MCP/ACP Bridge

Thanos exposes local agent bridge tools for task creation, sibling messaging,
related-work inspection, branch attachment, and user review requests through
`internal/mcp` and Tauri commands.

Affected paths:

- `internal/mcp`
- `apps/thanos-desktop/src-tauri/src/lib.rs`
- `apps/thanos-desktop/src/services/nativeBackend.ts`
- `.thanos/messages`
- `.thanos/branches`
- `.thanos/review-requests`

Decisions:

- Phase 6 is implemented as a local artifact-backed bridge, not a remote MCP server runtime.
- Review requests are stored separately from approved review artifacts to avoid bypassing review gates.
- Bridge tools never run shell commands and never auto-merge.

Verification:

- `GOCACHE=/private/tmp/thanos-go-build go test ./...`
- `cargo check`
- `npm run build`

## UI Menu, Task Removal, Provider Dropdown, Terminal Empty State

Workbench navigation uses unique active states, shell/detail menus are
responsive, task add/edit uses a popup with hover edit actions, imported
projects with no tasks still show the empty board, project onboarding lists
recent projects, workflow step providers use dropdown selection, and empty
terminals render a stable native-style prompt.

Affected paths:

- `apps/thanos-desktop/src/shared/ui/AppShell.tsx`
- `apps/thanos-desktop/src/shared/ui/SplitPane.tsx`
- `apps/thanos-desktop/src/flows/TaskWorkbenchFlow.tsx`
- `apps/thanos-desktop/src/flows/AgentSessionFlow.tsx`
- `apps/thanos-desktop/src/flows/project-onboarding-flow`
- `apps/thanos-desktop/src/flows/task-workbench-flow/TaskDialog.tsx`
- `apps/thanos-desktop/src/flows/agent-settings-flow/WorkflowStepSettingsPanel.tsx`
- `apps/thanos-desktop/src/state/workbenchStore.ts`

Decisions:

- Each sidebar label has a unique view id so only one menu item is active.
- Remove task is local UI state until a persistent backend delete command exists.
- Imported repos with no tasks should render the board empty state.
- Project switching uses local recent-project metadata.
- Empty terminal renders a prompt rather than log-placeholder text.

Verification:

- `npm run build`
- `GOCACHE=/private/tmp/thanos-go-build go test ./internal/workbench ./internal/orchestrator`

## Project Setup, Agent Settings, and Terminal Runner

Thanos desktop now distinguishes no-project onboarding from load failures,
persists project setup metadata, detects local CLI agents, configures workflow
step agents, and surfaces task-linked terminal sessions with xterm input
passthrough.

Affected paths:

- `apps/thanos-desktop/src/flows/project-onboarding-flow`
- `apps/thanos-desktop/src/flows/agent-settings-flow`
- `apps/thanos-desktop/src/flows/agent-terminal-flow`
- `apps/thanos-desktop/src/flows/task-workbench-flow`
- `apps/thanos-desktop/src-tauri/src/lib.rs`
- `internal/projects`
- `internal/agents`
- `internal/terminal`
- `internal/workflow`

Decisions:

- No workspace selected is an onboarding state, not an error.
- Local agents are detected but never installed automatically.
- Coding remains blocked until plan approval.
- Terminal execution stays visible and accepts user input through xterm.

Verification:

- `GOCACHE=/private/tmp/thanos-go-build go test ./...`
- `cargo check`
- `npm run build`
