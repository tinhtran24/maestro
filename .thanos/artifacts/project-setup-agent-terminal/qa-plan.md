# QA Plan: Project Setup, Agent Settings, Terminal Runner

Created: 2026-07-03T02:10:02Z

## Scope

Implement project onboarding and local-agent workflow configuration so the
desktop workbench no longer fails with only `Unable to load workbench state`.

## Subtasks

1. Project onboarding and project model
   - Objective: support creating/importing a local project from UI and persist
     `.thanos/settings.json` plus `.thanos/config.json`.
   - Acceptance: no workspace shows onboarding actions; invalid load shows
     Retry and View Logs; project metadata includes root, remote, default
     branch, worktree root, package manager, dev command, and test command.
   - Dependencies: Tauri folder picker and workbench loader.
   - Priority: P0
   - Complexity: Medium
   - Owner: coder
   - Affected files: desktop routes, project onboarding flow, native backend.

2. Agent detection and workflow-step settings
   - Objective: detect installed CLI agents and configure each workflow step.
   - Acceptance: Settings > Agents shows installed/not-found state, command
     path, version, test run, enable/disable, and defaults per step.
   - Dependencies: existing `detect_agent_clis` Tauri command.
   - Priority: P0
   - Complexity: Medium
   - Owner: coder
   - Affected files: `src-tauri/lib.rs`, `NativeBackend`, agent settings flow.

3. In-app terminal runner
   - Objective: make terminal sessions visible and step-linked for planning,
     coding, review, and tests.
   - Acceptance: task detail shows current step settings, start buttons honor
     approval gates, terminal tab opens automatically, output streams into
     task sessions.
   - Dependencies: existing PTY manager and xterm.js flow.
   - Priority: P0
   - Complexity: Medium
   - Owner: coder
   - Affected files: agent session flow, task workbench flow, terminal flow.

4. Tests and evidence
   - Objective: verify Go models/schema, TypeScript build, and Tauri check.
   - Acceptance: relevant checks pass or blocked reasons are recorded.
   - Dependencies: implementation.
   - Priority: P1
   - Complexity: Low
   - Owner: tester
   - Affected files: artifacts and memory.
