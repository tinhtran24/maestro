---
name: implementation-plan
description: Break approved specs into ordered implementation chunks.
applies_to:
  - planning
  - implementation
agents:
  - planner
  - coder
required_evidence:
  - ordered_steps
  - dependencies
  - test_strategy
---

# Skill: Implementation Plan

## When to use
Use after a feature spec exists and before code changes start.

## Workflow
1. Convert acceptance criteria into ordered steps.
2. Identify dependencies between steps.
3. Keep chunks small enough for review.
4. Define the focused test strategy.

## Exit criteria
- Ordered implementation steps exist.
- Dependencies are called out.
- Test strategy is attached.
