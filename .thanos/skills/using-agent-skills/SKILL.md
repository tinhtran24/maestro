---
name: using-agent-skills
description: Route each task to the smallest relevant skill set.
applies_to:
  - always
  - planning
agents:
  - planner
  - coder
  - reviewer
  - tester
required_evidence:
  - selected_skills
  - routing_reason
---

# Skill: Using Agent Skills

## When to use
Use before loading task-specific skills.

## Workflow
1. Read the task goal, status, agent role, tags, and planned files.
2. Match by workflow stage, role, task type, and files.
3. Load only the smallest set needed for the next action.
4. Mark external or global skills as untrusted until reviewed.
5. Record the selected skills and why they apply.

## Exit criteria
- Selected skills are listed.
- Routing reason is recorded.
- No unrelated skill content is loaded.
