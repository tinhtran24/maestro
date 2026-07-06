# Thanos

Thanos is being rebuilt as a local-first AI development workbench for planning,
running, reviewing, and approving AI coding tasks across local agent providers.

The human stays in control. Plans, code execution, review, tests, and merge
approval are explicit gates. Agents can work in parallel only through isolated
git worktrees.

## Product Shape

- Wails desktop workbench.
- Board, Plan, Mission Control, Agent Graph, Chat, Routines, Whiteboard,
  Analytics, Settings, and local Docs surfaces.
- Real provider discovery for local agents such as Codex, Claude Code, Gemini
  CLI, OpenCode, Cursor Agent, Aider, and Goose.
- Git remains the source of truth for code changes.
- Local-first state loaded from the selected workspace, not from a Thanos CLI.

## Workflow

```text
Backlog
  -> Planning
  -> Waiting Approval
  -> Ready
  -> Running
  -> In Review
  -> Done
```

Rules:

- Planner writes an execution plan.
- User approves the plan before coder starts.
- Coder runs only in an isolated task worktree and branch.
- Reviewer inspects diff and test results.
- User approves merge or requests changes.
- Done requires approved review and passing tests.
- No auto-merge.

## Current Status

The old CLI-centered Thanos app has been removed from the active desktop path.
The Wails app now owns the shell and loads real workspace metadata through a
Wails-native provider.

## Development

Run Go tests:

```sh
cd app
GOCACHE=/private/tmp/thanos-go-build GOMODCACHE=/private/tmp/thanos-go-mod go test ./...
```

Run the desktop UI:

```sh
cd app/frontend
npm install
npm run dev
```

Build the desktop frontend:

```sh
cd app/frontend
npm run build
```
