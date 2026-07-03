---
name: safe-code-change
description: Make scoped code changes with evidence and rollback awareness.
applies_to:
  - coding
  - implementation
agents:
  - coder
required_evidence:
  - changed_files
  - test_results
  - limitations
---

# Skill: Safe Code Change

## When to use
Use for implementation work.

## Workflow
1. Read nearby code and tests before editing.
2. Keep the change within the approved scope.
3. Add or update focused tests.
4. Record changed files and known limitations.

## Exit criteria
- Changed files are listed.
- Relevant tests were run or explicitly blocked.
- Limitations are recorded.
