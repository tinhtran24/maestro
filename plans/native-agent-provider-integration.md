# Native Agent Provider Integration

## Objective

Implement a provider-agnostic agent system inside Thanos.

Thanos is responsible for:

* Task orchestration
* Session lifecycle
* Prompt injection @internal/prompts folder
* Approval workflow
* Context management
* Terminal rendering
* Conversation persistence

External agents are responsible for:

* Planning
* Asking questions
* Coding
* Debugging
* Refactoring
* Git operations

Supported initial providers:

* Claude Code
* Codex CLI
* Gemini CLI
* Crush
* Custom local agent

---

# Core Rule

Thanos must never be tightly coupled to Claude.

Claude is only one provider.

Every workflow must run through a common `AgentProvider` interface.

---

# Provider Interface

```go
type AgentProvider interface {
    ID() string
    Name() string
    IsInstalled(ctx context.Context) bool
    Start(ctx context.Context, req StartAgentRequest) (*AgentSession, error)
    Send(ctx context.Context, sessionID string, input string) error
    Stream(ctx context.Context, sessionID string) (<-chan AgentEvent, error)
    Resize(ctx context.Context, sessionID string, cols int, rows int) error
    Interrupt(ctx context.Context, sessionID string) error
    Stop(ctx context.Context, sessionID string) error
}
```

---

# Provider Examples

```go
ClaudeProvider
CodexProvider
GeminiProvider
CrushProvider
CustomCommandProvider
```

Each provider defines:

```go
command
args
env
workingDirectory
supportsPTY
supportsStreaming
supportsInterrupt
supportsResume
```

---

# Session Model

```go
type AgentSession struct {
    ID          string
    ProviderID  string
    ProjectID   string
    TaskID      string
    Mode        AgentMode
    Status      AgentStatus
    Workdir     string
    CreatedAt   time.Time
    UpdatedAt   time.Time
}
```

Modes:

```go
planner
coding
review
debug
research
```

Statuses:

```go
idle
starting
running
waiting_user
completed
failed
stopped
archived
```

---

# Workflow

```text
Task
  ↓
Select Agent Provider
  ↓
Planner Session
  ↓
Agent asks questions
  ↓
User answers in Thanos
  ↓
Plan completed
  ↓
User approves
  ↓
Coding Session
  ↓
Agent executes
  ↓
Review
  ↓
Done
```

---

# Provider Selection

User can choose default agents per project:

```text
Planner Agent: Claude Code
Coding Agent: Codex CLI
Review Agent: Gemini CLI
Debug Agent: Claude Code
```

Fallback rule:

If selected provider is unavailable, Thanos shows:

```text
Provider not installed.
Install command:
brew install ...
```

Do not silently switch provider.

---

# Prompt Contract

Thanos generates provider-neutral prompts.

Bad:

```text
Claude, please implement this...
```

Good:

Read `internal/prompts/*` for each step plan input 

Prompt sections:

```text
Task
Context
Acceptance Criteria
Constraints
Allowed Files
Previous Plan
Expected Output
```

---

# Event System

```go
type AgentEvent struct {
    SessionID string
    Type      AgentEventType
    Payload   string
    Time      time.Time
}
```

Event types:

```text
session_started
stdout
stderr
prompt_sent
waiting_user
plan_ready
approval_required
execution_started
execution_done
error
session_stopped
```

---

# UI Requirements

Use design skill to design follow current UI/UX

---

# Persistence

Persist:

* Sessions
* Provider ID
* Mode
* Prompt history
* Conversation
* Terminal output
* Approved plan
* Execution result

A session must be reopenable after app restart.

---

# Non Goals

Do not implement:

* Claude-only workflow
* Provider-specific UI
* Fake terminal output
* Clipboard automation
* Custom AI reasoning inside Thanos
* Automatic provider switching

---

# Success Criteria

Thanos can run the same workflow with different agents:

```text
Plan with Claude Code
Code with Codex CLI
Review with Gemini CLI
Debug with Crush
```

Adding a new agent should only require a new provider adapter, not changes to workflow, UI, persistence, or task logic.
