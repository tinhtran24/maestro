# Feature Plan — Progressive AI Review

## Purpose

Build a Thanos-native review system that lets the user approve AI code changes progressively:

1. Semantic intent
2. File group
3. Function/component
4. Diff hunk
5. Line-level advanced review
6. Tests
7. Memory update

This should feel stronger than Cursor because Thanos reviews engineering intent first, then implementation details.

---

## Scope

Implement frontend-only UI and state for the Progressive Review flow.

Use mock data only.

Do not implement real git parsing, patch apply/revert, backend persistence, PTY, agent execution, or MCP/ACP.

---

## Product Goals

- Help users understand what changed before reading raw diffs.
- Allow approval at the right level of detail.
- Make high-risk changes visible.
- Prevent final approval until required review items are complete.
- Keep human approval as a hard gate.

---

## Review Levels

### Level 1 — Plan Review

Before coding starts.

Show:

- Execution plan
- Acceptance criteria
- Files likely to touch
- Risks
- Test strategy

Actions:

- Approve Plan
- Request Changes
- Reject

Coding cannot start before plan approval.

---

### Level 2 — Semantic Review

Group changes by engineering intent.

Example groups:

- Shopping Cart UI
- Cart State Management
- Checkout API
- Tests
- Styling
- Refactor

Each group shows:

- Summary
- Files changed
- Functions/components changed
- Risk level
- Test coverage
- AI explanation

Actions:

- Approve Group
- Request Changes
- Open Details

---

### Level 3 — File Review

Each changed file shows:

- File path
- Change type: added / modified / deleted / renamed
- Insertions
- Deletions
- AI summary
- Risk
- Related task requirement

Actions:

- Approve File
- Reject File
- Open Diff
- Open in Editor

---

### Level 4 — Function / Component Review

Main differentiator.

Each modified function/component shows:

- Function or component name
- File path
- Change summary
- Why it changed
- Before behavior
- After behavior
- Risk level
- Dependencies
- Suggested tests

Actions:

- Approve Function
- Request Changes
- Open Diff
- Explain More

---

### Level 5 — Hunk Review

Cursor-like raw diff hunk review.

Each hunk shows:

- Diff
- AI explanation
- Related function
- Related requirement

Actions:

- Accept Hunk
- Reject Hunk
- Ask AI to revise

---

### Level 6 — Line Review

Advanced mode only.

The user can approve/reject individual lines after expanding a hunk.

Actions:

- Accept Line
- Reject Line
- Comment

Line approval must not be the default UX.

---

### Level 7 — Test Review

Before final approval.

Show:

- Test command
- Passed tests
- Failed tests
- Coverage if available
- New tests added
- Missing test warnings

Actions:

- Approve Tests
- Run Again
- Request Test Fixes

Task cannot become Done unless tests_passed = true.

---

### Level 8 — Memory Review

Before storing project knowledge.

Show suggested memory updates:

- Architecture decision
- Feature decision
- Known bug
- File relationship
- Review note
- Convention

Actions:

- Save Memory
- Edit Memory
- Reject Memory

Do not silently write memory.

---

## Data Types

```ts
type ReviewLevel =
  | "plan"
  | "semantic_group"
  | "file"
  | "function"
  | "hunk"
  | "line"
  | "tests"
  | "memory";

type ReviewStatus =
  | "pending"
  | "approved"
  | "rejected"
  | "changes_requested";

type RiskLevel = "low" | "medium" | "high";

type ReviewGroup = {
  id: string;
  taskId: string;
  title: string;
  summary: string;
  intent: string;
  risk: RiskLevel;
  files: ReviewFile[];
  status: ReviewStatus;
};

type ReviewFile = {
  id: string;
  path: string;
  changeType: "added" | "modified" | "deleted" | "renamed";
  insertions: number;
  deletions: number;
  summary: string;
  risk: RiskLevel;
  functions: ReviewFunction[];
  hunks: ReviewHunk[];
  status: ReviewStatus;
};

type ReviewFunction = {
  id: string;
  filePath: string;
  name: string;
  kind: "function" | "component" | "method" | "class" | "hook" | "unknown";
  summary: string;
  whyChanged: string;
  beforeBehavior?: string;
  afterBehavior?: string;
  risk: RiskLevel;
  dependencies: string[];
  suggestedTests: string[];
  status: ReviewStatus;
};

type ReviewHunk = {
  id: string;
  filePath: string;
  header: string;
  diff: string;
  explanation: string;
  relatedFunctionId?: string;
  status: ReviewStatus;
};

type ReviewLine = {
  id: string;
  hunkId: string;
  lineNumber: number;
  content: string;
  type: "added" | "removed" | "context";
  status: ReviewStatus;
};
```

---

## Frontend Structure

```txt
src/
  flows/
    progressive-review-flow/
      ProgressiveReviewFlow.tsx
      ReviewLevelNav.tsx
      SemanticReviewGroupCard.tsx
      ReviewFileCard.tsx
      ReviewFunctionCard.tsx
      ReviewHunkViewer.tsx
      ReviewLineControls.tsx
      TestReviewCard.tsx
      MemoryReviewCard.tsx
      useProgressiveReviewFlow.ts

  features/
    review/
      model/
        review.types.ts
        review.mock.ts
      components/
        RiskBadge.tsx
        ReviewStatusBadge.tsx
        DiffBlock.tsx
        ApprovalButtons.tsx
```

---

## Main Workbench Review View

Add a dedicated Review view inside the workbench center.

Layout:

```txt
┌─────────────────────┬──────────────────────────────┬─────────────────────┐
│ Review levels       │ Selected review item          │ AI explanation      │
│ Semantic groups     │ Function / hunk / file view   │ Risk / tests/actions│
└─────────────────────┴──────────────────────────────┴─────────────────────┘
```

Bottom panel remains available for Terminal, Timeline, Logs, and Chat.

---

## Right Sidebar Integration

When task status is `in_review`, the right sidebar should prioritize review.

Show:

1. Review Summary
2. Risk Overview
3. Approval Progress
4. Changed Groups
5. Tests
6. Memory Updates

Example progress:

- Plan approved
- 3/4 groups approved
- 8/10 functions approved
- Tests passed
- Memory pending

Primary button:

`Approve Final Review`

Disabled until required items are approved.

---

## Mock Data Requirements

Create mock review data with:

- 3 semantic groups
- 5 changed files
- 8 changed functions/components
- 4 diff hunks
- 2 test results
- 2 suggested memory updates
- At least 1 high-risk function
- At least 1 rejected/request-changes example state

---

## Approval Rules

Final task approval is disabled until:

- All required semantic groups are approved
- All high-risk functions are approved
- Tests are approved
- Memory updates are either approved or rejected

Line approval is optional.

---

## UI Requirements

Use:

- TailwindCSS
- lucide-react
- Flow Component Design Pattern
- Existing Thanos dark design system

Design style:

- Rounded cards
- Compact density
- Purple accents for active actions
- Risk badges: low / medium / high
- Clear approval progress
- No emoji icons

---

## Implementation Boundaries

Do not implement:

- Real diff parsing
- Git apply/revert
- Real line patching
- Agent review execution
- Backend persistence
- PTY terminal
- MCP/ACP

---

## Acceptance Criteria

- Review flow supports semantic group review.
- Review flow supports file review.
- Review flow supports function/component review.
- Review flow supports hunk review.
- Line review exists as advanced expand mode.
- Right sidebar shows review progress when task is in review.
- Final approval button respects approval requirements.
- Uses mock data only.
- Uses TailwindCSS.
- Uses lucide-react.
- Follows Flow Component Design Pattern.
- Does not implement future phases.
