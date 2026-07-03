# QA Plan: Agent Skills System

Created: 2026-07-03T01:31:39Z

## Scope

Implement a Phase 1 Thanos-native skill system from `docs/agent-skill.md`:

- Project-level `.thanos/skills/**/SKILL.md` discovery.
- Skill frontmatter parsing and validation.
- Matching by agent role, workflow stage, task type, and changed/planned files.
- Skill run lifecycle and evidence validation models.
- SQLite schema support for skills, skill runs, and evidence.
- Active skills visible in the task detail sidebar.
- Default Thanos skills seeded as project skill contracts.

## Subtasks

1. Add skill domain package
   - Objective: create `internal/skills` registry, loader, matcher, validator, and executor files.
   - Acceptance: valid skills load from `.thanos/skills`; invalid skills report actionable errors; no shell scripts are executed.
   - Dependencies: existing Go model conventions.
   - Priority: P0
   - Complexity: Medium
   - Owner: coder
   - Affected files: `internal/skills/*`
   - Risks: over-parsing Markdown or introducing remote execution behavior outside Phase 1.

2. Extend workbench persistence models
   - Objective: add `Skill`, `SkillRun`, and `Evidence` model/schema types.
   - Acceptance: schema includes requested tables, lifecycle checks, source checks, and foreign keys.
   - Dependencies: current `internal/workbench` schema.
   - Priority: P0
   - Complexity: Low
   - Owner: coder
   - Affected files: `internal/workbench/model.go`, `internal/workbench/schema.sql`
   - Risks: migration compatibility for existing local SQLite files is out of scope for embedded schema tests.

3. Seed default project skills
   - Objective: add concise workflow-contract `SKILL.md` files under `.thanos/skills`.
   - Acceptance: defaults include all requested skill names and each has exit criteria/evidence.
   - Dependencies: loader format.
   - Priority: P1
   - Complexity: Low
   - Owner: coder
   - Affected files: `.thanos/skills/**/SKILL.md`
   - Risks: skills becoming documentation dumps instead of executable contracts.

4. Surface active skills in desktop UI
   - Objective: update domain types, fixtures, and sidebar presentation.
   - Acceptance: selected task shows active skills with name, agent, status, required evidence, and exit criteria.
   - Dependencies: frontend domain model.
   - Priority: P1
   - Complexity: Low
   - Owner: coder
   - Affected files: `apps/thanos-desktop/src/domain/models.ts`, fixtures, `TaskWorkbenchFlow.tsx`
   - Risks: UI-only gating is not sufficient; domain validation remains in Go.

5. Add automated tests and evidence
   - Objective: cover loader, matcher, validator, executor, and schema smoke behavior.
   - Acceptance: `go test ./internal/skills ./internal/workbench` passes; broader `go test ./...` attempted.
   - Dependencies: implementation.
   - Priority: P0
   - Complexity: Medium
   - Owner: tester
   - Affected files: `internal/skills/*_test.go`, `internal/workbench/*_test.go`
   - Risks: existing unrelated tests may fail; record exact evidence.
