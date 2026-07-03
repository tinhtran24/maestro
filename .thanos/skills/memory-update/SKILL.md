---
name: memory-update
description: Store important implementation decisions into project memory.
applies_to:
  - done
  - review
  - memory
agents:
  - planner
  - reviewer
  - tester
required_evidence:
  - memory_node
  - affected_paths
  - decision_summary
---

# Skill: Memory Update

## When to use
Use when behavior, architecture, or workflow decisions change.

## Workflow
1. Identify durable decisions and changed behavior.
2. Link affected paths and task identifiers.
3. Store concise memory entries.
4. Verify memory summaries remain readable.

## Exit criteria
- Memory node or artifact is recorded.
- Affected paths are linked.
- Decision summary is concise.
