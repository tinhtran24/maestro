---
name: dev-lifecycle
description: Full-cycle delivery method for Maestro sessions — Analysis and Plan when a task is created, then Development and Testing, then a conventionally-named branch and Conventional Commit. Use when planning a task draft or driving a worker task from start to pull request.
trigger: Planning a task (Analysis + Plan) or executing one (Development + Testing + branch/commit) inside a Maestro session.
---

# Development lifecycle

This is the delivery method every Maestro task follows. It distills a full
software lifecycle (Define → Plan → Build → Verify → Review → Ship) into the two
moments Maestro cares about: **creating a task** (Analysis + Plan) and **doing the
task** (Development + Testing, then branch + commit). Method credit:
[addyosmani/agent-skills](https://github.com/addyosmani/agent-skills).

## When a task is created: Analysis, then Plan

Do these in order and write them down in the task draft.

1. **Analysis** — understand before proposing. Capture:
   - the problem and the user story ("As a … I want … so that …");
   - explicit **acceptance criteria** (observable, testable statements);
   - **scope** (XS/S/M/L/XL) and the **risks**, dependencies, and open questions;
   - the **likely files/areas** the change touches.
   Surface missing information instead of guessing it.
2. **Plan** — decompose the analysis into small, verifiable steps ordered by
   dependency. Each step should be independently checkable. Prefer thin vertical
   slices over broad horizontal refactors.

A good task draft answers "what does done look like?" (acceptance criteria) and
"what is the smallest sequence of verifiable steps to get there?" (plan) before
any code is written.

## When a task is executed: Development, then Testing

The orchestrator coordinates workers through these phases and surfaces each step
so the human can see progress.

1. **Development** — implement one plan step at a time as a thin slice. Keep
   changes surgical and tied to the task; avoid drive-by refactors.
2. **Testing** — verify each slice: add or update tests at the boundary being
   changed, run the project's relevant checks, and exercise the real behavior
   (not just types/compile). Do not report a step complete while its checks fail.

Only after Development and Testing pass do you commit and open the pull request.

## Branch naming and commits

Create the branch **before** the first commit. Use a prefix that matches the
task's nature, then a short kebab-case description:

- `feature/<short-description>`
- `bugfix/<short-description>`
- `hotfix/<short-description>`
- `refactor/<short-description>`
- `chore/<short-description>`
- `docs/<short-description>`
- `test/<short-description>`

Commit messages **must** follow [Conventional Commits](https://www.conventionalcommits.org):
`<type>(<optional-scope>): <imperative summary>`. Use a standard type — `feat`,
`fix`, `refactor`, `docs`, `test`, `chore`, `build`, `ci`, or `perf`. The branch
prefix maps to the commit type: `feature`→`feat`, `bugfix`/`hotfix`→`fix`, and
`refactor`/`chore`/`docs`/`test` keep their names.

Examples:

- `feat(task): add task approval`
- `fix(ui): preserve sidebar width`
- `refactor(workflow): simplify transition logic`

Keep commits small and logically grouped; one concern per branch and pull
request.
