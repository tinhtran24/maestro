# Implementation Evidence: UI Menu, Task Removal, Provider Dropdown, Terminal Empty State

Created: 2026-07-03T02:27:27Z

## Changed Files

- `apps/thanos-desktop/src/shared/ui/AppShell.tsx`
- `apps/thanos-desktop/src/shared/ui/SidebarItem.tsx`
- `apps/thanos-desktop/src/shared/ui/FlowTabs.tsx`
- `apps/thanos-desktop/src/shared/ui/SplitPane.tsx`
- `apps/thanos-desktop/src/flows/TaskWorkbenchFlow.tsx`
- `apps/thanos-desktop/src/flows/AgentSessionFlow.tsx`
- `apps/thanos-desktop/src/flows/agent-settings-flow/WorkflowStepSettingsPanel.tsx`
- `apps/thanos-desktop/src/routes/WorkbenchRoute.tsx`
- `apps/thanos-desktop/src/state/workbenchStore.ts`
- `apps/thanos-desktop/src/shared/ui/TaskCard.tsx`
- `apps/thanos-desktop/src/flows/project-onboarding-flow/*`
- `apps/thanos-desktop/src/flows/task-workbench-flow/TaskDialog.tsx`

## Behavior Summary

- Added local remove-task action from task detail.
- Added Add/Edit Task popup and card hover edit action.
- Changed imported repos with zero tasks to show the empty board instead of replacing it with a first-task-only screen.
- Added recent project list and project switching in the Projects screen.
- Made Import Repo select and persist the chosen folder immediately when no path is set.
- Fixed sidebar active-state collisions by giving Board, Memory, Settings, Workflow Steps, and Executors distinct view ids.
- Made sidebar, topbar, detail tabs, and split panes responsive for small screens.
- Changed Workflow Steps provider control from text input to a dropdown using detected/default providers.
- Replaced repeated terminal placeholder text with a stable terminal prompt styled like a native terminal.

## Commands Run

- `npm run build`
  - Result: passed.
- `GOCACHE=/private/tmp/thanos-go-build go test ./internal/workbench ./internal/orchestrator`
  - Result: passed.
