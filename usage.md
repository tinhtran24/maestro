# Thanos Usage

Thanos is a local-first AI development workbench for planning, running,
verifying, and reviewing coding tasks across local agent providers.

## Workspace Setup

1. Open the desktop app.
2. Select a local repository folder.
3. Create or activate a workspace.
4. Confirm provider detection in Agent Graph.

Workspace state is stored under `.thanos/` in the selected repository. Task
history, routines, automation settings, events, terminal transcripts, and test
outputs remain local.

## Task Flow

Tasks move through these production states:

```text
backlog -> in_progress -> waiting -> committing -> done
```

Failure and cancellation states are available when a run cannot continue:

```text
failed
cancelled
```

Common workflow:

1. Create a task from the board.
2. Select a flow and agent role.
3. Run implementation in the task worktree.
4. Review provider output and turn logs.
5. Run verification.
6. Move the task to `committing` after verification passes.
7. Move the task to `done` after required gates pass.

## Agents And Flows

Built-in agents and flows are read-only. User-authored definitions can be added
as JSON files:

```text
.thanos/agents/*.json
.thanos/flows/*.json
```

Example agent:

```json
{
  "id": "security-review",
  "role": "Security Review",
  "harness": "Codex",
  "model": "gpt-5",
  "capabilities": ["diff.read", "risk.review"]
}
```

Example flow:

```json
{
  "id": "security-pass",
  "name": "Security Pass",
  "steps": ["Implementation", "Security Review", "Testing"],
  "parallelGroups": [["Security Review", "Testing"]]
}
```

Changes are loaded from disk when the workspace reloads.

## Verification

Verification runs a shell command inside the task worktree and stores output
separately from implementation logs:

```text
.thanos/tests/{task-id}/{test-id}.log
```

Pass and fail patterns can override the default verdict parser. When automation
requires tests, `done` is blocked until the latest verification passes.

## Commit Pipeline

Task completion requires committed worktree output. The commit pipeline prepares
a diff and commit message preview before any commit is created. The actual
commit requires explicit approval and runs inside the task worktree.

Commit metadata is stored on the task record:

```text
commit.hash
commit.summary
commit.message
commit.diffStat
```

Merge and cherry-pick operations remain separate from task completion and
require a later explicit approval flow.

## Local Files

Important workspace paths:

```text
.thanos/tasks/          task records
.thanos/logs/           implementation turn logs
.thanos/tests/          verification output
.thanos/worktrees/      task worktrees
.thanos/events.jsonl    workspace timeline
```
