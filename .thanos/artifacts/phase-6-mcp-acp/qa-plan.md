# QA Plan: Phase 6 MCP/ACP Bridge

Created: 2026-07-03T02:10:00Z

## Scope

Implement Phase 6 from `docs/rebuild-plan.md`:

- Expose task and memory tools to agents.
- Allow agents to create subtasks.
- Allow agents to message sibling tasks.
- Allow agents to inspect related work.
- Allow agents to attach branches.
- Allow agents to request user review.

This phase stays local-first and does not add a remote MCP server runtime. It
provides deterministic bridge commands and artifacts that native agent CLIs can
call through the Tauri backend or future MCP adapters.

## Subtasks

1. Add bridge domain package
   - Objective: define tool descriptors, requests, responses, validation, and artifact paths.
   - Acceptance: manifest includes all required Phase 6 tools and rejects invalid task/branch/message input.
   - Dependencies: current workspace task model and artifact layout.
   - Priority: P0
   - Complexity: Medium
   - Owner: coder
   - Affected files: `internal/mcp/*`
   - Risks: accidentally creating an external server or bypassing workflow gates.

2. Implement local tool operations
   - Objective: create subtasks, sibling messages, related-work inspection references, branch attachment metadata, and review requests.
   - Acceptance: operations persist to `.thanos/tasks`, `.thanos/messages`, `.thanos/branches`, and `.thanos/reviews` without shell execution.
   - Dependencies: `internal/workspace` task persistence.
   - Priority: P0
   - Complexity: Medium
   - Owner: coder
   - Affected files: `internal/mcp/*`, `internal/workspace/workspace.go`
   - Risks: malformed IDs or path traversal.

3. Expose Tauri bridge commands
   - Objective: add native commands for listing tools and invoking bridge actions.
   - Acceptance: frontend/backend API can call bridge actions with structured JSON.
   - Dependencies: existing Tauri command style.
   - Priority: P1
   - Complexity: Medium
   - Owner: coder
   - Affected files: `apps/thanos-desktop/src-tauri/src/lib.rs`, `apps/thanos-desktop/src/services/nativeBackend.ts`
   - Risks: Rust and TypeScript model drift.

4. Update docs, tests, and memory
   - Objective: document Phase 6 status and record durable decisions.
   - Acceptance: Go tests, desktop build, and artifacts pass; feature memory records the bridge behavior.
   - Dependencies: implementation.
   - Priority: P1
   - Complexity: Low
   - Owner: tester
   - Affected files: docs, `.thanos/artifacts`, `.thanos/memory`
   - Risks: existing generated dist churn from desktop build.
