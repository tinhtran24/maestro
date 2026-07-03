# Thanos Feature Memory

Generated: 2026-07-03T01:31:39Z

- Features: 2
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
