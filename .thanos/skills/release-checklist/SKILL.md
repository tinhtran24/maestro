---
name: release-checklist
description: Verify release readiness with approvals and test evidence.
applies_to:
  - release
  - done
agents:
  - reviewer
  - tester
required_evidence:
  - approval_summary
  - test_summary
  - known_risks
---

# Skill: Release Checklist

## When to use
Use before release or final completion milestones.

## Workflow
1. Confirm review approval.
2. Confirm tests passed.
3. List known risks and deferred work.
4. Confirm no required gate is bypassed.

## Exit criteria
- Approval summary is present.
- Test summary is present.
- Known risks are recorded.
