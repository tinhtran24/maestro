---
name: debugging
description: Investigate failures with reproducible evidence.
applies_to:
  - failed
  - blocked
  - debugging
agents:
  - coder
  - tester
required_evidence:
  - reproduction
  - root_cause
  - fix_validation
---

# Skill: Debugging

## When to use
Use when tests fail or a task is blocked by a defect.

## Workflow
1. Reproduce the failure.
2. Identify the smallest root cause.
3. Fix within scope.
4. Validate with the failing test or smoke path.

## Exit criteria
- Reproduction is captured.
- Root cause is recorded.
- Fix validation is attached.
