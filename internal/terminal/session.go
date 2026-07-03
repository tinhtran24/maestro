package terminal

import "time"

type SessionStatus string

const (
	SessionIdle        SessionStatus = "idle"
	SessionStarting    SessionStatus = "starting"
	SessionRunning     SessionStatus = "running"
	SessionWaitingUser SessionStatus = "waiting_user"
	SessionCompleted   SessionStatus = "completed"
	SessionFailed      SessionStatus = "failed"
	SessionStopped     SessionStatus = "stopped"
)

type Session struct {
	ID             string        `json:"id"`
	TaskID         string        `json:"task_id"`
	Step           string        `json:"step"`
	Provider       string        `json:"provider"`
	Command        string        `json:"command"`
	CWD            string        `json:"cwd"`
	Status         SessionStatus `json:"status"`
	PTYID          string        `json:"pty_id,omitempty"`
	TranscriptPath string        `json:"transcript_path,omitempty"`
	StartedAt      time.Time     `json:"started_at"`
	EndedAt        time.Time     `json:"ended_at,omitempty"`
}
