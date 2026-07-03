# Test Evidence: Agent Skills System

Tested: 2026-07-03T01:31:39Z
Tester: codex-tester

## EC-1: Load Project Skills

- Preconditions: `.thanos/skills/**/SKILL.md` exists.
- Steps: Run `GOCACHE=/private/tmp/thanos-go-build go test ./internal/skills`.
- Expected: valid skills parse; missing exit criteria fails validation.
- Actual: passed.
- Status: Passed.

## EC-2: Enforce Skill Evidence

- Preconditions: task has matched skill runs.
- Steps: Run orchestrator skill gate tests.
- Expected: ready/done transitions fail until required skill runs are completed with evidence.
- Actual: passed.
- Status: Passed.

## EC-3: Persist Skill Shape

- Preconditions: embedded workbench schema is available.
- Steps: Run workbench schema tests.
- Expected: schema contains skills, skill runs, evidence, lifecycle, and source checks.
- Actual: passed.
- Status: Passed.

## EC-4: Desktop Active Skills Panel

- Preconditions: desktop dependencies are installed.
- Steps: Run `npm run build` from `apps/thanos-desktop`.
- Expected: TypeScript and Vite production build pass.
- Actual: passed.
- Status: Passed.

## Smoke Tests

- `GOCACHE=/private/tmp/thanos-go-build go test ./...`
  - Result: Passed.
- `npm run build`
  - Result: Passed.
