package events

import (
	"fmt"
	"sync"
	"time"

	runtimefacts "github.com/tinhtran/thanos/internal/runtime"
)

type RuntimeEventInput struct {
	ProjectID string
	TaskID    string
	SessionID string
	EventType runtimefacts.EventType
	Payload   map[string]any
	CreatedAt time.Time
}

type StoredRuntimeEvent struct {
	ID        int64                  `json:"id"`
	ProjectID string                 `json:"project_id,omitempty"`
	TaskID    string                 `json:"task_id,omitempty"`
	SessionID string                 `json:"session_id,omitempty"`
	EventType runtimefacts.EventType `json:"event_type"`
	Payload   map[string]any         `json:"payload,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
}

type RuntimeEventLog interface {
	AppendRuntimeEvent(event RuntimeEventInput) (StoredRuntimeEvent, error)
	ListRuntimeEventsSince(afterID int64, limit int) ([]StoredRuntimeEvent, error)
}

type MemoryRuntimeEventLog struct {
	mu     sync.RWMutex
	nextID int64
	events []StoredRuntimeEvent
	clock  func() time.Time
}

func NewMemoryRuntimeEventLog() *MemoryRuntimeEventLog {
	return &MemoryRuntimeEventLog{clock: func() time.Time { return time.Now().UTC() }}
}

func NewMemoryRuntimeEventLogWithClock(clock func() time.Time) *MemoryRuntimeEventLog {
	log := NewMemoryRuntimeEventLog()
	if clock != nil {
		log.clock = clock
	}
	return log
}

func (l *MemoryRuntimeEventLog) AppendRuntimeEvent(input RuntimeEventInput) (StoredRuntimeEvent, error) {
	if input.EventType == "" {
		return StoredRuntimeEvent{}, fmt.Errorf("runtime event type is required")
	}
	if input.CreatedAt.IsZero() {
		input.CreatedAt = l.clock()
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	l.nextID++
	event := StoredRuntimeEvent{
		ID:        l.nextID,
		ProjectID: input.ProjectID,
		TaskID:    input.TaskID,
		SessionID: input.SessionID,
		EventType: input.EventType,
		Payload:   clonePayload(input.Payload),
		CreatedAt: input.CreatedAt,
	}
	l.events = append(l.events, event)
	return event, nil
}

func (l *MemoryRuntimeEventLog) ListRuntimeEventsSince(afterID int64, limit int) ([]StoredRuntimeEvent, error) {
	if limit <= 0 {
		limit = 100
	}
	l.mu.RLock()
	defer l.mu.RUnlock()
	var out []StoredRuntimeEvent
	for _, event := range l.events {
		if event.ID <= afterID {
			continue
		}
		out = append(out, cloneStoredRuntimeEvent(event))
		if len(out) == limit {
			break
		}
	}
	return out, nil
}

func RuntimeEventStarted(projectID string, taskID string, sessionID string, step runtimefacts.WorkflowStep) RuntimeEventInput {
	return RuntimeEventInput{
		ProjectID: projectID,
		TaskID:    taskID,
		SessionID: sessionID,
		EventType: runtimefacts.EventAgentSessionStarted,
		Payload: map[string]any{
			"step": step,
		},
	}
}

func RuntimeEventOutput(projectID string, taskID string, sessionID string, bytes int) RuntimeEventInput {
	return RuntimeEventInput{
		ProjectID: projectID,
		TaskID:    taskID,
		SessionID: sessionID,
		EventType: runtimefacts.EventAgentSessionOutput,
		Payload: map[string]any{
			"bytes": bytes,
		},
	}
}

func RuntimeEventExited(projectID string, taskID string, sessionID string, exitCode int) RuntimeEventInput {
	return RuntimeEventInput{
		ProjectID: projectID,
		TaskID:    taskID,
		SessionID: sessionID,
		EventType: runtimefacts.EventAgentSessionExited,
		Payload: map[string]any{
			"exit_code": exitCode,
		},
	}
}

func RuntimeEventToTimeline(event StoredRuntimeEvent) Event {
	payload := clonePayload(event.Payload)
	if payload == nil {
		payload = map[string]any{}
	}
	payload["runtime_event_id"] = event.ID
	payload["runtime_event_type"] = string(event.EventType)
	if event.SessionID != "" {
		payload["session_id"] = event.SessionID
	}
	return Event{
		ProjectID: event.ProjectID,
		TaskID:    event.TaskID,
		Event:     "runtime." + string(event.EventType),
		Stage:     "runtime",
		Status:    "fact",
		Payload:   payload,
		CreatedAt: event.CreatedAt,
	}
}

func RuntimeEventsToTimeline(events []StoredRuntimeEvent) []Event {
	timeline := make([]Event, 0, len(events))
	for _, event := range events {
		timeline = append(timeline, RuntimeEventToTimeline(event))
	}
	return timeline
}

func cloneStoredRuntimeEvent(event StoredRuntimeEvent) StoredRuntimeEvent {
	event.Payload = clonePayload(event.Payload)
	return event
}
