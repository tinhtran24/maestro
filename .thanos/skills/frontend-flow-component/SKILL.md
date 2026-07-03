---
name: frontend-flow-component
description: Keep frontend work aligned with Flow Component Design Pattern.
applies_to:
  - frontend
  - react
  - tsx
agents:
  - planner
  - coder
  - reviewer
required_evidence:
  - flow_boundary
  - api_hook_boundary
  - component_boundary
---

# Skill: Frontend Flow Component

## When to use
Use for desktop or web UI work.

## Workflow
1. Keep routes/screens as flow composition.
2. Put business interaction in flow modules.
3. Keep feature components presentational.
4. Separate API hooks from UI components.

## Exit criteria
- Flow ownership is identified.
- API hook boundary is recorded.
- Presentational component boundary is preserved.
