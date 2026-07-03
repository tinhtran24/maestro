# AGENTS.md

## Role

You are the implementation agent for Thanos.

Thanos is being rebuilt as a local-first AI development workbench. The product
is inspired by AI Kanban, native agent terminals, task graphs, project memory,
review gates, workflow skills, and MCP/ACP-style coordination. Do not copy
implementation code from external repositories.

Thanos is not an AI CLI. Thanos is an AI Development Workbench.

The workbench coordinates humans, local AI agents, git worktrees, project
memory, review gates, and execution workflows in one visual interface.

Every action must be transparent. Every decision must be reviewable. Every AI
action must be reproducible.

## Product Principles

1. The human stays in control.
2. Important steps require review or approval.
3. Agents work in parallel only through isolated git worktrees.
4. Planner, Coder, Reviewer, and Tester are separate roles.
5. The user can chat at any time.
6. The user can inspect plan, files, terminal, browser preview, git diff, and
   task state in one workspace.
7. Native agent CLIs are first-class. Do not hide Codex, Claude Code, Gemini
   CLI, OpenCode, or custom commands behind a heavy CLI workflow.
8. The UI is the primary product. The CLI is only a thin launcher or execution
   backend.
9. Every agent session must be visible inside Thanos.
10. Native CLI output must stream directly into the Workbench.

## Design Principles

Always optimize for:

1. Transparency
2. Human control
3. Low token usage
4. Local-first execution
5. Explainability
6. Reproducible workflows
7. Small independent modules
8. Event-driven architecture
9. Configuration over hardcoding
10. AI-provider agnostic behavior

## New Architecture

```text
apps/
  thanos-desktop/   Tauri desktop shell and workbench UI
  thanos-web/       optional browser UI
  thanos-cli/       thin launcher only

internal/
  orchestrator/     workflow engine, transitions, gates, event dispatch
  board/            kanban columns and task tree
  agents/           provider profiles, discovery, adapters, workflow config
  executor/         process/docker/ssh runtime profiles
  workspace/        repo, branch, worktree management
  memory/           SQLite + FTS project memory
  review/           diff parser, tests, approve/reject flow
  terminal/         PTY sessions and conversation capture
  mcp/              tools exposed to agents
  events/           compact JSON events and WebSocket streams
  workbench/        Phase 1 domain models and SQLite schema
  skills/           project-level workflow skills and evidence gates
  projects/         project onboarding, repository attachment, config
```

## Phase Discipline

Implement only the requested phase. The authoritative phase pointer is the
Current Progress table in `docs/PLAN.md`.

Completed phases:

- Phase 1 — AI Workbench UI: app shell, sidebar, top bar, board, right context
  sidebar, bottom panel (Flow Component Pattern, Tailwind, lucide-react).
- Phase 2 — Project Management: local-only (localStorage) multi-project support
  with mock data — Projects Page, Project/Import dialogs, Project Details,
  topbar project switcher, recent projects, per-project settings. No Git
  operations.
- Phase 3 — Workflow Engine: validated task state machine (single source of
  truth for legal transitions + gates), per-task timeline events / history,
  mock transition controls, Timeline wired to real history, state-machine unit
  tests (vitest). Mock transitions only; no real execution or Git.
- Phase 4 — Agent Configuration: mock installed-agent detection (supported
  catalog with statuses/path/version), configurable workflow steps (provider,
  command, args, env, working dir, timeout, approval, auto-start, permissions),
  enable/disable, per-step provider assignment, mock Test Run, local persistence
  of config + provider overrides. No hardcoded step providers; mock detection
  only, no execution.

- Phase 5 — Native Terminal Runtime: first-class terminal UI with a mock runtime
  — multiple sessions per task, tabs, real-time streamed output, status
  lifecycle, restart/stop, transcript viewer, pinned sessions; provider/command
  sourced from workflow-step config. Terminal UI only — no real execution.

- Phase 6 — Planner Workflow: mock planning — Start Planning launches the
  planning terminal, planner poses clarifying questions, user answers, an
  execution plan is generated, and an approval gate moves the task Ready
  (Approve) or back to Planning (Request Changes). No coding before approval.

Current / next phase:

- Phase 7 — Coding Workflow: approved → (mock) worktree → coding terminal →
  logs → task enters Review. Mock worktree; no real Git or execution.

Do not implement later-phase behavior unless explicitly requested:

- real worktree creation
- PTY terminal manager
- real agent start/stop
- diff parser
- merge flow
- MCP/ACP bridge

Phases may include UI placeholders and typed service boundaries for later
phases, but must not silently execute agents or mutate repositories.

## Workflow Rules

Task statuses:

```text
backlog -> planning -> waiting_approval -> ready -> running -> in_review -> done
```

Allowed side paths:

```text
blocked
failed
request changes -> ready or planning
```

Hard gates:

- `ready` from `waiting_approval` requires plan approval.
- `running` requires an isolated worktree and branch.
- `done` requires `review_approved = true` and `tests_passed = true`.
- Review is a hard pause.
- No auto-merge.
- Never let an agent modify main directly.
- Always capture agent conversation.
- Always show changed files and diff before merge.
- Always store important decisions into memory.

## Human Visibility Rule

The user must never wonder: "What is the AI doing?"

Every AI action must be visible through:

- active workflow step
- live terminal
- timeline
- agent messages
- changed files
- execution logs

If an action cannot be explained visually, reconsider the implementation.

## Projects

Projects are first-class objects.

A project owns:

- repositories
- branches
- worktrees
- settings
- workflow
- memory
- agent configuration
- task board

Users can:

- Create Project
- Import Existing Repository
- Attach Multiple Repositories
- Clone Repository
- Archive Project

Recent projects must appear on startup.

Project actions:

- New Project
- Import Existing Repo
- Select Local Folder
- Attach Git Repository
- Set Default Branch
- Set Worktree Root
- Set Package Manager
- Set Default Test Command
- Set Default Dev Command

Project model:

```text
Project
  id
  name
  root_path
  git_remote_url
  default_branch
  worktree_root
  package_manager
  dev_command
  test_command
  created_at
  updated_at
```

Workbench empty state:

- If no project is loaded, show `ProjectOnboardingFlow`.
- Do not show only `Unable to load workbench state`.
- Show an onboarding card with:
  - `No project loaded`
  - `Create a new project or import an existing repository`
  - `Import Repo`
  - `New Project`
  - `Open Recent`
- If project exists but no board exists, show `CreateFirstTaskFlow`.
- If load failed, show `Retry` and `View Logs`.

## Workflow Agent Settings

Every workflow step uses a configurable local agent.

Workflow steps:

- Planning
- Coding
- Review
- Testing
- Debugging
- Documentation
- Memory Update

Each step has:

- Enabled / Disabled
- Agent provider
- Command
- Working directory mode
- Auto-start terminal
- Require user approval before next step
- Environment variables
- Timeout
- Permissions

Example config:

```yaml
Planning:
  provider: Claude Code
  command: claude
  auto_start_terminal: true
  approval_required: true

Coding:
  provider: Codex
  command: codex
  auto_start_terminal: true
  approval_required: true

Review:
  provider: Claude Code
  command: claude
  auto_start_terminal: true
  approval_required: true

Testing:
  provider: Shell
  command: pnpm test
  auto_start_terminal: true
  approval_required: false
```

Never hardcode providers. Users choose which installed agent runs each workflow
step.

## Installed Agents

Thanos automatically discovers installed local agents.

Supported:

- Claude Code
- Codex
- Gemini CLI
- OpenCode
- Cursor Agent
- Aider
- Goose
- Custom Commands

Detection behavior:

- Use local PATH lookup such as `which <command>` or platform equivalent.
- Show status: `Installed`, `Not Found`, or `Needs Setup`.
- Show executable path and version when available.
- Do not install agents automatically.
- Provide setup hints only.

Agent provider model:

```text
AgentProvider
  id
  name
  command
  detected_path
  status
  version
  type: cli | mcp | acp | shell
```

Settings UI:

- Settings > Agents
- Settings > Workflow Steps
- Installed agents list
- Command path
- Version
- Test Run button
- Enable / Disable
- Set as default for step

## Native Agent Terminal

Terminal is a first-class component.

When a workflow step starts:

```text
Planning -> launch Claude terminal
Coding   -> launch Codex terminal
Review   -> launch Review terminal
Testing  -> launch Test terminal
```

Never execute agents silently.

Users must always see:

- command
- output
- questions
- logs
- exit status

Terminal output becomes part of task history.

Terminal requirements:

- Use xterm.js in frontend.
- Use a backend PTY manager in Phase 2.
- Support multiple terminal sessions per task.
- Terminal tabs should support names such as:
  - Planning Claude
  - Coding Codex
  - Review Claude
  - Tests
- User can stop/restart session.
- Output must stream in real time.
- Terminal session must be linked to Task and AgentSession.

Agent session model:

```text
AgentSession
  id
  task_id
  step
  provider
  command
  cwd
  status: idle | starting | running | waiting_user | completed | failed | stopped
  pty_id
  transcript_path
  started_at
  ended_at
```

## Agent Skills

Skills define engineering workflows. Skills are executable guidance, not long
static documentation dumps.

Skills are inspired by project-level agent workflows and quality gates. Do not
copy external skill code directly. Implement a Thanos-native skill system.

Skill lifecycle:

```text
discovered -> matched -> activated -> running -> evidence_pending -> completed
```

Rules:

- Load only relevant skills.
- Never load every skill.
- Skills never replace user approval.
- Skills must define exit criteria.
- Skills must produce visible evidence.
- Agents cannot claim `done` without evidence.
- Treat third-party skills as untrusted until reviewed.

Suggested project structure:

```text
.thanos/
  skills/
    using-agent-skills/
      SKILL.md
    feature-spec/
      SKILL.md
    implementation-plan/
      SKILL.md
    code-review/
      SKILL.md
    testing/
      SKILL.md
    debugging/
      SKILL.md
    refactoring/
      SKILL.md
    frontend-flow-component/
      SKILL.md
    tailwind-ui/
      SKILL.md
    tauri-desktop/
      SKILL.md
```

Skill behavior:

- When a task enters Planning, load only relevant skills.
- Use `using-agent-skills` as the router.
- Match skills by task type, agent role, files, and workflow stage.
- Show active skills in the Task Detail right sidebar.
- Do not allow task transition unless required evidence exists.
- Store skill outputs into project memory.
- Allow custom project skills in `.thanos/skills`.
- Allow global skills, but project skills override global ones.

Security:

- Never auto-run shell scripts from skills.
- Never load remote skills without user approval.
- Mark external skills as untrusted.
- Show diff when installing or updating a skill.
- Skills cannot access secrets unless explicitly allowed.

## Data

Use SQLite for local workbench state. Use Git as the source of truth for code
changes. Use compact JSON for events and machine state. Use Markdown only for
human-readable plans, logs, reviews, and test summaries.

Core models:

```text
Feature
  id
  project_id
  title
  description
  status
  plan_graph_id
  created_at

Task
  id
  feature_id
  parent_task_id
  title
  description
  status: backlog | planning | waiting_approval | ready | running | in_review | blocked | done | failed
  priority: P0 | P1 | P2 | P3
  assigned_agent
  executor_profile
  worktree_path
  branch_name
  created_at
  updated_at

ExecutionPlan
  id
  task_id
  summary
  steps[]
  risks[]
  files_to_touch[]
  test_strategy
  approval_status

Review
  id
  task_id
  diff_summary
  changed_files[]
  test_results[]
  reviewer_notes
  status: pending | approved | rejected | changes_requested

MemoryNode
  id
  project_id
  type: feature | decision | architecture | file | task | bug | convention
  title
  content
  links[]
  embedding_optional
  created_at
```

## Event System

Everything emits events.

Examples:

- ProjectCreated
- ProjectLoaded
- TaskCreated
- TaskMoved
- PlanGenerated
- PlanApproved
- AgentStarted
- AgentStopped
- ReviewApproved
- MemoryUpdated

Use compact JSON. UI updates should flow through events. Avoid polling where a
realtime event stream is available.

## Memory

Memory is project-scoped.

Store:

- Architecture
- Feature decisions
- Code decisions
- Reviews
- Known bugs
- ADRs
- Plans
- Task relationships

Search uses SQLite FTS. Memory updates happen only after approval or explicit
workflow completion.

## Frontend Architecture

Use Flow Component Design Pattern.

Core idea:

- Screens only compose flows.
- Flows own business interaction.
- Components stay dumb/presentational.
- State machine controls task lifecycle.
- API hooks are separated from UI.

Structure:

```text
src/
  app/
    routes/
    providers/
    shell/

  flows/
    project-onboarding-flow/
      ProjectOnboardingFlow.tsx
      ImportRepoFlow.tsx
      NewProjectFlow.tsx
      useProjectOnboardingFlow.ts

    board-flow/
      BoardFlow.tsx
      BoardToolbar.tsx
      BoardColumnFlow.tsx
      TaskCardFlow.tsx
      useBoardFlow.ts
      board-flow.machine.ts

    task-workbench-flow/
      TaskWorkbenchFlow.tsx
      TaskHeaderFlow.tsx
      PlanReviewFlow.tsx
      AgentAssignmentFlow.tsx
      FilesFlow.tsx
      TerminalFlow.tsx
      ChatFlow.tsx
      GitChangesFlow.tsx
      TaskStepPanel.tsx
      StepApprovalGate.tsx
      StartAgentButton.tsx
      useTaskWorkbenchFlow.ts
      task-workbench.machine.ts

    agent-settings-flow/
      AgentSettingsFlow.tsx
      InstalledAgentsPanel.tsx
      WorkflowStepSettingsPanel.tsx
      AgentCommandTestPanel.tsx
      useAgentSettingsFlow.ts

    agent-session-flow/
      AgentSessionFlow.tsx
      AgentTerminalFlow.tsx
      AgentMessageFlow.tsx
      useAgentSessionFlow.ts

    agent-terminal-flow/
      AgentTerminalFlow.tsx
      TerminalTabs.tsx
      AgentSessionToolbar.tsx
      useAgentTerminalFlow.ts

  features/
    tasks/
      api/
      model/
      hooks/
      components/

    agents/
      api/
      model/
      hooks/
      components/

    memory/
      api/
      model/
      hooks/
      components/

    workspace/
      api/
      model/
      hooks/
      components/

  shared/
    ui/
      Button.tsx
      Card.tsx
      Tabs.tsx
      Badge.tsx
      Dialog.tsx
      ScrollArea.tsx
      SplitPane.tsx
      CommandInput.tsx

    layout/
      AppShell.tsx
      LeftSidebar.tsx
      RightSidebar.tsx
      BottomPanel.tsx

    lib/
      cn.ts
      event-bus.ts
      websocket.ts
      shortcuts.ts
```

Frontend rules:

1. Route components must only mount screen-level flows.
2. Flow components own orchestration, state transitions, loading states, and user actions.
3. Presentational components must not call APIs directly.
4. Business logic must live in hooks, machines, or flow controllers.
5. UI components in `shared/ui` must be reusable and dumb.
6. Task lifecycle must be controlled by explicit state machines.
7. BoardFlow controls board filtering, drag/drop, column state, and task selection.
8. TaskWorkbenchFlow controls selected task, tabs, approval actions, agent session,
   terminal, files, git changes, and chat.
9. AgentSessionFlow controls native CLI session lifecycle and output streaming.
10. Keep layout separate from domain logic.
11. Do not put large logic inside JSX.
12. Prefer small composed components over one giant screen file.

## UI Rules

Use:

- Tailwind CSS
- shadcn/ui-style primitives
- lucide-react
- react-resizable-panels
- xterm.js
- dnd-kit for Kanban drag/drop
- TanStack Query for server state
- Zustand or XState for flow/local state

Do not use emoji icons in production UI.

Icon rules:

- Use lucide-react only.
- Icons must be 16-18px by default.
- Sidebar icons: 18px.
- Button icons: 16px.
- Status icons: 14px.
- Keep icon + label short.

Suggested icons:

```text
Sidebar:
  Workbench: LayoutDashboard
  Board: Kanban
  Projects: FolderKanban
  Memory: Brain
  Agents: Bot
  Executors: Cpu
  Settings: Settings

Top bar:
  Project switcher: Boxes
  Branch: GitBranch
  Online agents: CircleDot
  New task: Plus
  Search: Search
  Notifications: Bell

Board:
  Backlog: Inbox
  Planning: ListTodo
  Waiting Approval: ShieldCheck
  In Progress: PlayCircle
  In Review: GitPullRequest
  Done: CheckCircle2
  Priority: Flag
  Assignee: UserRound
  Agent: Bot

Task detail:
  Overview: LayoutPanelTop
  Plan: ScrollText
  Files: FileTree
  Changes: GitCompare
  Terminal: Terminal
  Browser: Globe
  Tests: FlaskConical
  Timeline: Clock3
  Chat: MessageSquare
  Memory: Brain
  Approve: Check
  Request Changes: RotateCcw
  Stop Agent: Square
  Run Tests: Play
```

Tailwind design tokens:

```text
bg-app: #070A12
bg-sidebar: #0B1020
bg-panel: #0F172A
bg-card: #111827
border-soft: #1F2937
text-main: #E5E7EB
text-muted: #9CA3AF
purple-primary: #7C3AED
purple-hover: #8B5CF6
blue-info: #38BDF8
green-success: #22C55E
yellow-warning: #EAB308
red-danger: #EF4444
orange-review: #F97316
```

Component style:

- Use `rounded-xl` for cards.
- Use `rounded-lg` for buttons/inputs.
- Use `border border-slate-800`.
- Use `bg-slate-950/70` for main background.
- Use `bg-slate-900/80` for panels.
- Use `shadow-lg shadow-black/20` for floating panels.
- Use `backdrop-blur` for top bar.
- Use `text-xs` for metadata.
- Use `text-sm` for normal UI text.
- Use `text-base` or `text-lg` for section titles.
- Use compact spacing: `p-3`, `p-4`, `gap-3`.

Required reusable components:

- AppShell
- SidebarItem
- StatusBadge
- PriorityBadge
- TaskCard
- BoardColumn
- FlowTabs
- SplitPane
- AgentAvatar
- EmptyState

## Workbench UI Requirements

Main layout:

- Left sidebar:
  - Projects
  - Board
  - Workbench
  - Memory
  - Agents
  - Executors
  - Settings
- Top bar:
  - Current project
  - Active branch/worktree
  - Agent status
  - Run/Stop buttons
  - New Task
- Center:
  - Kanban board or selected task workbench
- Right sidebar:
  - Task details
  - Plan
  - Agent messages
  - Related memory
  - Must be scrollable
- Bottom panel:
  - Chat
  - Terminal
  - Timeline
  - Logs
  - Browser Preview
  - Chat supports `/` commands

Workbench screen:

- Task header
- Plan card
- Current step indicator
- File tree
- Code editor / diff viewer
- Terminal with native agent CLI
- Browser preview
- Git changes
- Review actions:
  - Approve plan
  - Request changes
  - Run tests
  - Approve merge
  - Reject

Task detail must include Step Settings panel:

- Current step
- Assigned local agent
- Command
- Terminal status
- Approval required
- Open Terminal button
- Start Planning Agent button
- Start Coding Agent button only after plan approval

## Engineering Rules

- Keep modules small.
- No hardcoded providers.
- All agent commands are configurable.
- All state transitions are validated in Go.
- Add tests for orchestrator state machine changes.
- Add mock executors before testing agent flows.
- Use event-driven architecture.
- Prefer simple implementation first.
- E2E: Playwright.
- No God Objects.
- Interfaces should live near usage.
- No global mutable state.
- No hidden magic.
- Prefer composition.
- Avoid reflection unless necessary.
- Configuration over hardcoded behavior.