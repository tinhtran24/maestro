# Review: Project Setup, Agent Settings, Terminal Runner

Reviewed: 2026-07-03T02:10:02Z
Reviewer: codex-reviewer
Decision: Approved for testing

## Findings

- No blocker findings.
- Major follow-up: PTY process management is still single-active-session in the Rust backend. The UI and session model support multiple task-linked terminal tabs, but concurrent PTY execution needs a deeper manager change.
- Minor follow-up: workflow step edits are not yet saved back to `.thanos/config.json`.

## Coverage

- Project onboarding avoids the previous hardcoded path failure.
- Agent detection reports installed/not-found state without installing anything.
- Coding start remains blocked before plan approval.
- Terminal output remains visible and input is sent to the PTY.
