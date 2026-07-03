# Test Evidence: Phase 6 MCP/ACP Bridge

Tested: 2026-07-03T02:10:00Z
Tester: codex-tester

## EC-1: Expose Agent Tools

- Preconditions: `internal/mcp` bridge package exists.
- Steps: Run `GOCACHE=/private/tmp/thanos-go-build go test ./internal/mcp`.
- Expected: manifest includes subtask, sibling message, related work, branch attachment, and review request tools.
- Actual: passed.
- Status: Passed.

## EC-2: Persist Task Bridge Artifacts

- Preconditions: initialized temporary `.thanos` workspace.
- Steps: Run bridge package tests.
- Expected: child tasks, messages, branches, and review requests are persisted as local artifacts.
- Actual: passed.
- Status: Passed.

## EC-3: Preserve Safety Gates

- Preconditions: task exists.
- Steps: Attempt protected branch attachment and request user review.
- Expected: protected branches fail; review request remains pending and does not approve/merge.
- Actual: passed.
- Status: Passed.

## EC-4: Native Command Build

- Preconditions: Tauri backend is available.
- Steps: Run `cargo check` in `apps/thanos-desktop/src-tauri`.
- Expected: bridge commands serialize and register successfully.
- Actual: passed.
- Status: Passed.

## Smoke Tests

- `GOCACHE=/private/tmp/thanos-go-build go test ./...`
  - Status: Passed.
- `npm run build`
  - Status: Passed.
- `cargo check`
  - Status: Passed.
