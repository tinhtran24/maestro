package runtime

type EventType string

const (
	EventAgentSessionStarted   EventType = "AgentSessionStarted"
	EventAgentSessionOutput    EventType = "AgentSessionOutput"
	EventAgentSessionExited    EventType = "AgentSessionExited"
	EventRuntimeHandleAttached EventType = "RuntimeHandleAttached"
	EventRuntimeHandleDetached EventType = "RuntimeHandleDetached"
	EventWorktreePrepared      EventType = "WorktreePrepared"
	EventWorkflowGateBlocked   EventType = "WorkflowGateBlocked"
)

type RuntimeEvent struct {
	Type      EventType `json:"type"`
	ProjectID string    `json:"project_id,omitempty"`
	TaskID    string    `json:"task_id,omitempty"`
	SessionID string    `json:"session_id,omitempty"`
	Payload   any       `json:"payload,omitempty"`
}
