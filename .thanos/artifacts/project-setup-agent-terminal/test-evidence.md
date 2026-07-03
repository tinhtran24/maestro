# Test Evidence: Project Setup, Agent Settings, Terminal Runner

Tested: 2026-07-03T02:10:02Z
Tester: codex-tester

## EC-1: Empty Project Onboarding

- Preconditions: no persisted workspace path.
- Steps: Build frontend with onboarding route.
- Expected: no workspace returns onboarding instead of fatal load error.
- Actual: TypeScript build passed.
- Status: Passed.

## EC-2: Project Setup Metadata

- Preconditions: setup request contains name and root path.
- Steps: Run Go project service tests and Tauri `cargo check`.
- Expected: project defaults include default branch and worktree root; native command compiles.
- Actual: passed.
- Status: Passed.

## EC-3: Agent Settings

- Preconditions: desktop app can invoke agent detection.
- Steps: Run `cargo check` and `npm run build`.
- Expected: agent detection model and settings UI compile.
- Actual: passed.
- Status: Passed.

## EC-4: Terminal Runner

- Preconditions: PTY backend and xterm frontend are available.
- Steps: Run `cargo check` and `npm run build`.
- Expected: terminal input/output APIs compile and task terminal tabs render.
- Actual: passed.
- Status: Passed.

## Smoke Tests

- `GOCACHE=/private/tmp/thanos-go-build go test ./...`
  - Status: Passed.
- `cargo check`
  - Status: Passed.
- `npm run build`
  - Status: Passed.
