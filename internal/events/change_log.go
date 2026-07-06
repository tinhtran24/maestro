package events

import (
	"encoding/json"
	"fmt"
	"slices"
	"sync"
	"time"
)

type AggregateType string

const (
	AggregateTask      AggregateType = "task"
	AggregateGate      AggregateType = "gate"
	AggregateRuntime   AggregateType = "runtime"
	AggregateSkillRun  AggregateType = "skill_run"
	AggregateReadModel AggregateType = "read_model"
)

type ChangeType string

const (
	ChangeTaskStateChanged              ChangeType = "task_state_changed"
	ChangeGateStateChanged              ChangeType = "gate_state_changed"
	ChangeRuntimeFactChanged            ChangeType = "runtime_fact_changed"
	ChangeSkillRunChanged               ChangeType = "skill_run_changed"
	ChangeDerivedStatusRefreshRequested ChangeType = "derived_status_refresh_requested"
)

type Change struct {
	ID            int64          `json:"id"`
	ProjectID     string         `json:"project_id,omitempty"`
	TaskID        string         `json:"task_id,omitempty"`
	SessionID     string         `json:"session_id,omitempty"`
	AggregateType AggregateType  `json:"aggregate_type"`
	AggregateID   string         `json:"aggregate_id"`
	ChangeType    ChangeType     `json:"change_type"`
	Payload       map[string]any `json:"payload,omitempty"`
	DerivedStatus string         `json:"derived_status,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
}

type ChangeInput struct {
	ProjectID     string
	TaskID        string
	SessionID     string
	AggregateType AggregateType
	AggregateID   string
	ChangeType    ChangeType
	Payload       map[string]any
	DerivedStatus string
	CreatedAt     time.Time
}

type ChangeLog interface {
	Append(change ChangeInput) (Change, error)
	ListSince(afterID int64, limit int) ([]Change, error)
}

type MemoryChangeLog struct {
	mu      sync.RWMutex
	nextID  int64
	changes []Change
	clock   func() time.Time
}

func NewMemoryChangeLog() *MemoryChangeLog {
	return &MemoryChangeLog{clock: func() time.Time { return time.Now().UTC() }}
}

func NewMemoryChangeLogWithClock(clock func() time.Time) *MemoryChangeLog {
	log := NewMemoryChangeLog()
	if clock != nil {
		log.clock = clock
	}
	return log
}

func (l *MemoryChangeLog) Append(input ChangeInput) (Change, error) {
	if input.AggregateType == "" {
		return Change{}, fmt.Errorf("aggregate type is required")
	}
	if input.AggregateID == "" {
		return Change{}, fmt.Errorf("aggregate id is required")
	}
	if input.ChangeType == "" {
		return Change{}, fmt.Errorf("change type is required")
	}
	if input.CreatedAt.IsZero() {
		input.CreatedAt = l.clock()
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	l.nextID++
	change := Change{
		ID:            l.nextID,
		ProjectID:     input.ProjectID,
		TaskID:        input.TaskID,
		SessionID:     input.SessionID,
		AggregateType: input.AggregateType,
		AggregateID:   input.AggregateID,
		ChangeType:    input.ChangeType,
		Payload:       clonePayload(input.Payload),
		DerivedStatus: input.DerivedStatus,
		CreatedAt:     input.CreatedAt,
	}
	l.changes = append(l.changes, change)
	return change, nil
}

func (l *MemoryChangeLog) ListSince(afterID int64, limit int) ([]Change, error) {
	if limit <= 0 {
		limit = 100
	}
	l.mu.RLock()
	defer l.mu.RUnlock()
	var out []Change
	for _, change := range l.changes {
		if change.ID <= afterID {
			continue
		}
		out = append(out, cloneChange(change))
		if len(out) == limit {
			break
		}
	}
	return out, nil
}

type RefreshReason string

const (
	RefreshTasks         RefreshReason = "tasks"
	RefreshGates         RefreshReason = "gates"
	RefreshRuntimeStatus RefreshReason = "runtime_status"
	RefreshSkillRuns     RefreshReason = "skill_runs"
)

type RefreshRequest struct {
	AfterID int64           `json:"after_id"`
	NextID  int64           `json:"next_id"`
	Reasons []RefreshReason `json:"reasons"`
	Changes []Change        `json:"changes"`
}

type ReadModelCursor struct {
	LastChangeID int64 `json:"last_change_id"`
}

func RefreshSince(log ChangeLog, cursor ReadModelCursor, limit int) (RefreshRequest, ReadModelCursor, error) {
	changes, err := log.ListSince(cursor.LastChangeID, limit)
	if err != nil {
		return RefreshRequest{}, cursor, err
	}
	next := cursor
	if len(changes) > 0 {
		next.LastChangeID = changes[len(changes)-1].ID
	}
	request := RefreshRequest{
		AfterID: cursor.LastChangeID,
		NextID:  next.LastChangeID,
		Reasons: refreshReasons(changes),
		Changes: changes,
	}
	return request, next, nil
}

func TaskStateChanged(taskID string, from string, to string) ChangeInput {
	return ChangeInput{
		TaskID:        taskID,
		AggregateType: AggregateTask,
		AggregateID:   taskID,
		ChangeType:    ChangeTaskStateChanged,
		Payload: map[string]any{
			"from": from,
			"to":   to,
		},
	}
}

func GateStateChanged(taskID string, gate string, status string) ChangeInput {
	return ChangeInput{
		TaskID:        taskID,
		AggregateType: AggregateGate,
		AggregateID:   taskID,
		ChangeType:    ChangeGateStateChanged,
		Payload: map[string]any{
			"gate":   gate,
			"status": status,
		},
	}
}

func RuntimeFactChanged(taskID string, sessionID string, derivedStatus string, payload map[string]any) ChangeInput {
	return ChangeInput{
		TaskID:        taskID,
		SessionID:     sessionID,
		AggregateType: AggregateRuntime,
		AggregateID:   sessionID,
		ChangeType:    ChangeRuntimeFactChanged,
		Payload:       payload,
		DerivedStatus: derivedStatus,
	}
}

func SkillRunChanged(taskID string, skillRunID string, status string) ChangeInput {
	return ChangeInput{
		TaskID:        taskID,
		AggregateType: AggregateSkillRun,
		AggregateID:   skillRunID,
		ChangeType:    ChangeSkillRunChanged,
		Payload: map[string]any{
			"status": status,
		},
	}
}

func DerivedStatusRefreshRequested(taskID string, sessionID string, derivedStatus string) ChangeInput {
	return ChangeInput{
		TaskID:        taskID,
		SessionID:     sessionID,
		AggregateType: AggregateReadModel,
		AggregateID:   sessionID,
		ChangeType:    ChangeDerivedStatusRefreshRequested,
		Payload: map[string]any{
			"source": "runtime_facts",
		},
		DerivedStatus: derivedStatus,
	}
}

func MarshalChangePayload(payload map[string]any) (string, error) {
	if payload == nil {
		return "{}", nil
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func refreshReasons(changes []Change) []RefreshReason {
	var reasons []RefreshReason
	add := func(reason RefreshReason) {
		if !slices.Contains(reasons, reason) {
			reasons = append(reasons, reason)
		}
	}
	for _, change := range changes {
		switch change.ChangeType {
		case ChangeTaskStateChanged:
			add(RefreshTasks)
		case ChangeGateStateChanged:
			add(RefreshGates)
		case ChangeRuntimeFactChanged, ChangeDerivedStatusRefreshRequested:
			add(RefreshRuntimeStatus)
		case ChangeSkillRunChanged:
			add(RefreshSkillRuns)
		}
	}
	return reasons
}

func cloneChange(change Change) Change {
	change.Payload = clonePayload(change.Payload)
	return change
}

func clonePayload(payload map[string]any) map[string]any {
	if payload == nil {
		return nil
	}
	clone := make(map[string]any, len(payload))
	for key, value := range payload {
		clone[key] = value
	}
	return clone
}
