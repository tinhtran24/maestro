# Update Current Thanos Layout — AI Workbench Redesign

## Goal

Redesign the existing Thanos desktop layout to match the new AI Development Workbench design.

This update is **UI only**. Do not implement backend behavior yet.

The result should feel closer to **Linear + Cursor + Claude Code + Kandev**, while preserving the Thanos identity.

---

# Core Design Principles

- Local-first AI Workbench
- Minimal clicks
- Context-aware UI
- Information density without clutter
- Everything important visible
- Human always in control
- Native AI terminal is first-class
- TailwindCSS + lucide-react only
- Flow Component Design Pattern

---

# Layout

Replace the current layout with a 4-region workbench.

```
┌──────────────┬──────────────────────────────────────────────┬────────────────────┐
│              │ Top Bar                                      │                    │
│ Left Sidebar ├──────────────────────────────────────────────┤ Right Context      │
│              │                                              │ Sidebar            │
│              │ Kanban Board / Workbench                     │                    │
│              │                                              │                    │
├──────────────┼──────────────────────────────────────────────┤                    │
│              │ Bottom Panel                                 │                    │
└──────────────┴──────────────────────────────────────────────┴────────────────────┘
```

No dead space.

Panels must feel connected.

---

# Left Sidebar

Width

260px

Dark background.

Rounded right border.

Contains

## Logo

Thanos logo

Subtitle

AI Development Workbench

---

## Main Navigation

Workbench

Board

Projects

Memory

Agents

Executors

Workflow Steps

Settings

Use lucide-react icons.

---

## Recent Projects

Card

Recent Projects

Example

- Ecommerce Platform
- AI Chat App
- Design System

Footer

View all projects →

---

## User Card

Bottom fixed.

Avatar

Thanos Dev

Local-first workspace

More menu

---

# Top Bar

Height

64px

Glass effect.

Left

Project selector

↓

Current Project

Branch selector

↓

main

↓

.thanos/worktrees

Center

Agent status

🟢

0 Agents Online

Right

Search

Notifications

New Task

Purple gradient button.

---

# Board Header

Title

Board

Subtitle

If empty

No tasks yet.
Use New Task to add the first item.

Otherwise

Track AI workflow tasks across planning, coding, review and testing.

Right

Search Tasks

Filters

---

# Kanban Board

Columns

Backlog

Planning

Waiting Approval

In Progress

In Review

Done

Each column contains

Icon

Title

Count badge

Empty drop zone

Add Task button

Empty drop zone

Dashed border

Package icon

Title

No Tasks

Subtitle

Drag tasks here or create new

Use compact cards.

Horizontal scrolling when needed.

---

# Center Empty State

When no task selected.

Show centered card.

Icon

Package

Title

No task selected

Description

Select a task from the board to view details.

Avoid large blank space.

---

# Right Sidebar

Remove the current vertical tab navigation completely.

The sidebar becomes a contextual dashboard.

Width

400px

Scrollable

Sticky header

Resizable later

---

## Header

Task Details

Collapse button

---

## Empty State

Card

Package icon

Title

No task selected

Description

Select a task from the board to view details.

---

## Quick Actions

2×2 Grid

New Task

Import Tasks

Link Repository

View Memory

Each card

Icon

Title

Small description

---

## Active Workflow

Card

Vertical list

Planning

Coding

Review

Testing

Memory Update

Each row

Icon

Status

Pending badge

---

## Project Information

Card

Repository

Branch

Worktree Root

Last Updated

Edit button

---

# When Task Selected

Replace the empty cards.

Show

---

## Task Header

Task ID

Task Title

Priority

Status

Workflow Step

---

## Primary Actions

Approve Plan

Request Changes

Start Agent

Open Terminal

Run Tests

Open Git Diff

Buttons remain visible.

---

## Workflow Timeline

Vertical stepper

✓ Ticket Created

✓ Planning

● Waiting Approval

○ Coding

○ Review

○ Testing

○ Done

Each step expands.

---

## Active Agent

Card

Agent Name

Provider

Command

Runtime

Status

Buttons

Open Terminal

Restart

Stop

View Transcript

---

## Plan Summary

Summary

Acceptance Criteria

Risks

Files To Modify

Expand button

---

## Related Memory

Architecture

Previous Decisions

Similar Features

Known Bugs

Recent Reviews

---

## Files To Change

Compact list

Click opens editor later.

---

## Git Status

Branch

Worktree

Files Changed

Insertions

Deletions

Buttons

View Diff

Open Commit

---

## Recent Activity

Timeline

Planner created plan

User approved

Codex started

Tests passed

Review requested

Newest first.

---

## Active Skills

Feature Spec

Implementation Plan

Testing

Review Checklist

Each shows

Status

Progress

Expand

---

# Bottom Panel

Height

280px

Tabs

Terminal

Timeline

Logs

Chat

Use lucide icons.

Empty state

No task selected

Logs and terminal output will appear after an agent starts.

---

# Colors

Background

bg-slate-950

Sidebar

bg-slate-950

Panels

bg-slate-900/70

Cards

bg-slate-900/80

Borders

border-slate-800

Primary

violet-500

Hover

violet-600

Text

text-slate-100

Muted

text-slate-400

Success

green-500

Warning

amber-500

Danger

red-500

Info

sky-400

---

# Components

AppShell

TopBar

LeftSidebar

RightContextSidebar

BottomPanel

BoardToolbar

BoardColumn

TaskCard

EmptyBoardState

TaskHeaderCard

TaskActionBar

WorkflowTimelineCard

ActiveAgentCard

PlanSummaryCard

FilesCard

GitStatusCard

ActivityCard

SkillsCard

ProjectContextCard

QuickActionsCard

EmptyTaskCard

StatusBadge

PriorityBadge

SearchInput

IconButton

---

# Frontend Architecture

Follow Flow Component Design Pattern.

```
src/
  app/
    shell/
      AppShell.tsx
      TopBar.tsx
      LeftSidebar.tsx
      RightContextSidebar.tsx
      BottomPanel.tsx

  flows/
    board-flow/
      BoardFlow.tsx
      BoardToolbar.tsx
      BoardColumnFlow.tsx
      TaskCardFlow.tsx
      EmptyBoardState.tsx
      useBoardFlow.ts

    task-workbench-flow/
      TaskWorkbenchFlow.tsx
      TaskHeaderFlow.tsx
      TaskActionBar.tsx
      WorkflowTimelineCard.tsx
      ActiveAgentCard.tsx
      PlanSummaryCard.tsx
      FilesCard.tsx
      GitStatusCard.tsx
      SkillsCard.tsx
      ActivityCard.tsx
      ProjectContextCard.tsx
      EmptyTaskCard.tsx
      useTaskWorkbenchFlow.ts

    project-context-flow/
      QuickActionsCard.tsx
      ActiveWorkflowCard.tsx
      ProjectInfoCard.tsx

  shared/
    ui/
      Button.tsx
      Card.tsx
      Badge.tsx
      EmptyState.tsx
      SearchInput.tsx
      StatusBadge.tsx
      PriorityBadge.tsx
      IconButton.tsx
```

Rules

- Routes compose Flows only.
- Flows own business logic.
- Shared UI components remain presentational.
- No API calls inside dumb components.
- Keep JSX small.
- Prefer composition.

---

# Icons

Use **lucide-react** only.

Sidebar

LayoutDashboard

Kanban

FolderKanban

Brain

Bot

Cpu

ListChecks

Settings

Board

Inbox

ListTodo

ShieldCheck

PlayCircle

GitPullRequest

CheckCircle2

Right Sidebar

PanelRight

Workflow

Terminal

GitBranch

FolderGit2

Brain

Clock

MessageSquare

ScrollText

FlaskConical

---

# Implementation Scope

Implement only UI.

Use mock data.

Do not implement

- PTY terminal
- Agent execution
- MCP
- ACP
- Worktree creation
- Git diff parser
- SQLite persistence

---

# Acceptance Criteria

- Current layout visually matches the redesigned workbench.
- Right sidebar is redesigned into contextual cards.
- Vertical tab navigation is removed.
- Empty states are polished.
- Bottom panel includes Terminal, Timeline, Logs, Chat.
- Left sidebar includes Recent Projects.
- Top bar includes Project selector, Branch selector, Agent status, Search, Notifications, New Task.
- Uses TailwindCSS.
- Uses lucide-react.
- Uses Flow Component Design Pattern.
- Responsive from 1440px to 4K.
- Clean, modern, production-ready desktop UI.
```