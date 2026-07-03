# QA Plan: UI Menu, Task Removal, Provider Dropdown, Terminal Empty State

Created: 2026-07-03T02:27:27Z

## Scope

- Add remove task action.
- Fix duplicate active sidebar states for Workbench/Memory and Settings/Workflow Steps.
- Make shell/topbar/detail menus responsive on small screens.
- Use provider dropdown in Workflow Steps.
- Replace repeated terminal placeholder text with a terminal-like empty state.

## Acceptance Criteria

- Only one sidebar item is active at a time.
- Horizontal menus scroll or wrap safely on small screens.
- A task can be removed from the workbench state.
- Workflow step provider is selected from installed/default providers.
- Empty terminal looks like an interactive OS terminal prompt and does not repeat placeholder text.

## Risks

- Remove task is currently local UI state only unless backed by a future delete command.
- Responsive shell changes must not break desktop dense layout.
