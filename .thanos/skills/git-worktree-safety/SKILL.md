---
name: git-worktree-safety
description: Protect main by requiring isolated branches and worktrees.
applies_to:
  - running
  - coding
  - git
agents:
  - coder
  - reviewer
required_evidence:
  - branch_name
  - worktree_path
  - diff_summary
---

# Skill: Git Worktree Safety

## When to use
Use before a task enters running or review.

## Workflow
1. Confirm the task has an isolated branch.
2. Confirm the task has an isolated worktree.
3. Show changed files and diff before approval.
4. Never auto-merge.

## Exit criteria
- Branch name is recorded.
- Worktree path is recorded.
- Diff summary is visible.
