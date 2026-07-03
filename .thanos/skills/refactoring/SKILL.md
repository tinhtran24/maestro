---
name: refactoring
description: Keep refactors scoped and behavior-preserving.
applies_to:
  - refactor
  - coding
agents:
  - coder
  - reviewer
required_evidence:
  - behavior_baseline
  - changed_files
  - regression_tests
---

# Skill: Refactoring

## When to use
Use when changing structure without changing intended behavior.

## Workflow
1. Capture the behavior baseline.
2. Refactor in small, reviewable steps.
3. Keep public contracts stable unless approved.
4. Run regression tests.

## Exit criteria
- Behavior baseline is recorded.
- Changed files are listed.
- Regression test evidence exists.
