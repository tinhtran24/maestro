---
name: code-review
description: Review changes for correctness, risk, and acceptance coverage.
applies_to:
  - review
  - in_review
agents:
  - reviewer
required_evidence:
  - review_decision
  - severity_comments
  - reviewed_revision
---

# Skill: Code Review

## When to use
Use when a coder submits changes for review.

## Workflow
1. Read the ticket, plan, and implementation evidence.
2. Inspect changed files and tests.
3. Classify findings by severity.
4. Approve only when testing may begin.

## Exit criteria
- Review decision is explicit.
- Severity-classified comments are recorded.
- Reviewed revision is captured.
