# Thanos Rebuild Plan

> Master implementation roadmap for Thanos.
>
> This document defines the implementation order.
> Always follow this roadmap.
> Never skip phases.
> Never implement future phases unless explicitly requested.

---

# Vision

Thanos is a local-first AI Development Workbench.

It is **not** an AI CLI.

It coordinates:

- Human
- Local AI Agents
- Git Worktrees
- Project Memory
- Kanban
- Review Gates
- Native Agent Terminals

Everything must be transparent.

Everything must be reviewable.

Everything important must be visible.

---

# Product Principles

1. Human stays in control.
2. AI never silently modifies code.
3. Every important workflow requires approval.
4. Native AI terminals are first-class.
5. Project memory persists decisions.
6. Every task belongs to a project.
7. Git is the source of truth.
8. UI comes first.
9. Event-driven architecture.
10. Configuration over hardcoding.

---

# Current Progress

| Phase | Status |
|--------|--------|
| UI Redesign | ✅ Completed |
| Project Management | ✅ Completed |
| Workflow Engine | ✅ Completed |
| Agent Runtime | ✅ Completed |
| Planner Flow | ✅ Completed |
| Coding Flow | ✅ Completed |
| Review Flow | ✅ Completed |
| Memory | ✅ Completed |
| Skills | ⏳ Next |
| MCP | Pending |

**Current Phase: Phase 10 — Skills** (not yet started — awaiting approval).

---

# Phase 1 — AI Workbench UI ✅

Goal

Create a production-ready desktop workbench.

Completed

- App Shell
- Left Sidebar
- Top Bar
- Kanban Board
- Right Context Sidebar
- Bottom Panel
- TailwindCSS
- lucide-react
- Flow Component Pattern

Acceptance

- Responsive
- Clean
- Modern
- No backend dependency

---

# Phase 2 — Project Management ✅

Status: Completed. Local-only (localStorage) project management with mock data;
no Git operations. Projects Page, Project/Import dialogs, Project Details,
topbar switcher, and recent projects implemented.

Goal

Projects become the root entity.

Implement

Projects

Repositories

Workspace

Project Settings

Recent Projects

Import Existing Repository

Create Project

Archive Project

Project Switcher

Branch Selector

Worktree Root

Dev Command

Test Command

Environment Variables

Deliverables

Projects Page

Project Dialog

Import Dialog

Project Details

Topbar Project Switcher

Mock data only.

No Git operations.

Exit Criteria

- Multiple projects supported
- Project switcher works
- Recent projects visible
- Settings stored locally

---

# Phase 3 — Workflow Engine ✅

Status: Completed. Validated task state machine (single source of truth for
legal transitions + gates), per-task timeline events / history, mock transition
controls, and Timeline wired to real history. State-machine unit tests added
(vitest). Mock only — no real agent execution or Git operations.

Goal

Replace static task status with validated workflow.

Task Status

Backlog
↓
Planning
↓
Waiting Approval
↓
Ready
↓
Running
↓
Review
↓
Testing
↓
Done

Side Paths

Blocked

Failed

Rules

Planning requires approval.

Running requires Ready.

Done requires

Review Approved

Tests Passed

Deliverables

State Machine

Transition Validation

Timeline Events

Task History

Mock transitions

Exit Criteria

No invalid transition possible.

---

# Phase 4 — Agent Configuration ✅

Status: Completed. Mock installed-agent detection (8 supported providers with
Installed / Not Found / Needs Setup, path, version), configurable workflow steps
(provider, command, arguments, environment, working dir, timeout, approval,
auto-start, permissions), enable/disable, "use for step" assignment, mock Test
Run. Config + provider overrides persist locally. No hardcoded step providers;
no real detection or execution. Catalog/override unit tests added.

Goal

Users configure local AI agents.

Workflow Steps

Planning

Coding

Review

Testing

Documentation

Memory Update

Supported Providers

Claude Code

Codex

Gemini CLI

OpenCode

Cursor

Aider

Goose

Custom

Each Step Stores

Provider

Command

Arguments

Environment

Timeout

Approval

Auto Start

Deliverables

Agent Settings

Workflow Settings

Installed Agent Detection

Mock detection only.

Exit Criteria

Workflow is configurable.

---

# Phase 5 — Native Terminal Runtime ✅

Status: Completed. First-class terminal UI (mock runtime, no real execution):
multiple sessions per task with tabs (Planning · Claude, Coding · Codex,
Review · Claude, Tests), real-time streamed output, status lifecycle,
restart/stop, transcript viewer, and pinned sessions. Provider/command come from
the Phase 4 workflow-step config. Runtime-helper unit tests added.

Goal

Native terminals become first-class.

Planning

↓

Claude Terminal

Coding

↓

Codex Terminal

Review

↓

Claude Terminal

Testing

↓

Shell Terminal

Requirements

Multiple Tabs

Realtime Logs

Restart

Stop

Transcript

Pinned Sessions

No hidden execution.

Everything visible.

Exit Criteria

Terminal UI complete.

No actual execution required.

---

# Phase 6 — Planner Workflow ✅

Status: Completed (mock). Start Planning launches the planning terminal, the
planner poses clarifying questions, the user answers, an execution plan is
generated (summary, steps, files, risks, test strategy), and an approval gate
moves the task Ready (Approve) or back to Planning (Request Changes). Coding
stays blocked until approval. Plan-generator unit tests added.

Goal

Implement Planning.

Flow

Create Task

↓

Start Planning

↓

Open Claude

↓

Ask Questions

↓

User Answers

↓

Execution Plan

↓

Approval

↓

Ready

Deliverables

Planning Flow

Question UI

Execution Plan

Approval Gate

Exit Criteria

No Coding before approval.

---

# Phase 7 — Coding Workflow ✅

Status: Completed (mock). An approved task exposes the Coder flow: "Create
Worktree & Start Coding" mock-assigns an isolated worktree/branch (recorded as an
`agent_started` timeline event), opens the coding (Codex) terminal streaming
mock output, and synthesizes a deterministic changeset (changed files + diff
summary) into the Git Changes panel. A review gate — disabled until the coding
session completes — moves the task into Review. Changeset-generator unit tests
added. Mock worktree only; no real Git or agent execution.

Goal

Implement Coding.

Flow

Approved

↓

Create Worktree

↓

Open Codex

↓

Coding

↓

Logs

↓

Review

Deliverables

Coding Flow

Task Timeline

Terminal Session

Mock Worktree

Exit Criteria

Task enters Review.

---

# Phase 8 — Review ✅

Status: Completed (mock). An in-review task exposes the Review panel: changed
files + diff summary, a derived checklist (diff collected, tests pass, matches
plan, no secrets, conventions), and the test run. Approve / Reject / Request
Changes / Run Tests / Finish actions route through the state machine — Finish is
gated on review approval + passing tests, so a task cannot reach Done without
approval. Approving synthesizes a memory node (see Phase 9). Reviewer-helper unit
tests added. No real diff parser or merge.

Goal

Review before Done.

Review shows

Changed Files

Git Diff

Checklist

Tests

Approve

Reject

Request Changes

Deliverables

Review Panel

Review Timeline

Review Actions

Exit Criteria

Cannot finish without approval.

---

# Phase 9 — Memory ✅

Status: Completed (mock). A full-screen Memory browser with mock full-text search,
type filtering, a node list, and a detail view showing related memory (derived
from explicit links + shared type). Approving a review synthesizes a linked
decision node — "memory updates after approval". Backed by the in-memory store;
memory search/synthesis unit tests added. No real SQLite/FTS yet.

Goal

Persistent project memory.

Store

Architecture

Feature Decisions

Bugs

Plans

Reviews

ADR

Relationships

SQLite FTS

Deliverables

Memory UI

Memory Search

Related Memory

Exit Criteria

Memory updates after approval.

---

# Phase 10 — Skills

Inspired by

addyosmani/agent-skills

Goal

Workflow guidance.

Skills

Feature Spec

Implementation Plan

Testing

Review

Frontend

Tailwind

Debugging

Memory Update

Skill Lifecycle

Discover

↓

Match

↓

Activate

↓

Evidence

↓

Complete

Deliverables

Skill Registry

Skill Loader

Skill Runner

Skill Evidence

Exit Criteria

Only relevant skills loaded.

---

# Phase 11 — MCP / ACP

Goal

External agent interoperability.

Support

MCP

ACP

External Tool Calls

Agent Messages

Deliverables

Bridge

Task API

Memory API

Workflow API

Exit Criteria

Agents communicate through standard interfaces.

---

# Future

Distributed Agents

Cloud Executors

Docker Executors

SSH Executors

Remote Teams

AI Pair Programming

Visual Plan Graph

Knowledge Graph

Plugin Marketplace

---

# Development Rules

Always

- Small PRs
- One phase at a time
- Mock before backend
- UI before logic
- Event driven
- Testable
- Reusable components

Never

- Skip roadmap
- Implement future phases
- Hardcode providers
- Hide AI execution
- Bypass approval

---

# Definition of Done

A phase is complete only if:

- Acceptance criteria are satisfied
- UI is polished
- Components are reusable
- Code is tested
- Documentation updated
- AGENTS.md still reflects the implementation
- No unfinished placeholder code remains