# Test Evidence: UI Menu, Task Removal, Provider Dropdown, Terminal Empty State

Tested: 2026-07-03T02:27:27Z
Tester: codex-tester

## EC-1: Navigation Active State

- Steps: Build desktop frontend.
- Expected: unique view ids compile and route correctly.
- Actual: passed.
- Status: Passed.

## EC-1B: Empty Imported Repo Board

- Steps: Build desktop frontend after removing first-task route replacement.
- Expected: imported project with no tasks still renders board columns and no selected task state.
- Actual: passed.
- Status: Passed.

## EC-1C: Task Add/Edit Popup

- Steps: Build desktop frontend with task dialog and card hover edit action.
- Expected: New Task opens popup; task card hover exposes edit; types compile.
- Actual: passed.
- Status: Passed.

## EC-2: Provider Dropdown

- Steps: Build desktop frontend.
- Expected: workflow step provider dropdown compiles against detected provider state.
- Actual: passed.
- Status: Passed.

## EC-3: Terminal Empty State

- Steps: Build desktop frontend.
- Expected: terminal placeholder is stable prompt text, not repeated log output.
- Actual: passed.
- Status: Passed.

## Smoke Tests

- `npm run build`
  - Status: Passed.
- `GOCACHE=/private/tmp/thanos-go-build go test ./internal/workbench ./internal/orchestrator`
  - Status: Passed.
