---
name: testing
description: Define and execute evidence cases before task completion.
applies_to:
  - testing
  - in_review
agents:
  - tester
required_evidence:
  - evidence_cases
  - smoke_results
  - test_commands
---

# Skill: Testing

## When to use
Use after review approval and before completion.

## Workflow
1. Map evidence cases to acceptance criteria.
2. Run focused automated tests.
3. Run adjacent smoke checks.
4. Record failures with bug links or reopen action.

## Exit criteria
- Evidence cases are recorded.
- Smoke results are attached.
- Test commands and results are captured.
