package runtime

import "time"

type DisplayStatus string

const (
	DisplayIdle             DisplayStatus = "idle"
	DisplayStarting         DisplayStatus = "starting"
	DisplayRunning          DisplayStatus = "running"
	DisplayWaitingUser      DisplayStatus = "waiting_user"
	DisplayCompleted        DisplayStatus = "completed"
	DisplayFailed           DisplayStatus = "failed"
	DisplayStopped          DisplayStatus = "stopped"
	DisplayDisconnected     DisplayStatus = "disconnected"
	DisplayRestoreAvailable DisplayStatus = "restore_available"
	DisplayNoSignal         DisplayStatus = "no_signal"
)

type StatusOptions struct {
	NoSignalAfter time.Duration
}

func DefaultStatusOptions() StatusOptions {
	return StatusOptions{NoSignalAfter: 2 * time.Minute}
}

func DeriveDisplayStatus(facts AgentSessionFact, now time.Time, opts StatusOptions) DisplayStatus {
	if facts.ID == "" {
		return DisplayIdle
	}
	if opts.NoSignalAfter == 0 {
		opts = DefaultStatusOptions()
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}
	if facts.Terminated || facts.Activity == ActivityExited || !facts.EndedAt.IsZero() {
		if facts.ExitCode != nil {
			if *facts.ExitCode == 0 {
				return DisplayCompleted
			}
			return DisplayFailed
		}
		if facts.CanRestore() {
			return DisplayRestoreAvailable
		}
		return DisplayStopped
	}
	if !facts.HasRuntimeHandle() && facts.HasStarted() {
		return DisplayDisconnected
	}
	if facts.Activity == ActivityUnknown && facts.HasStarted() && now.Sub(facts.StartedAt) >= opts.NoSignalAfter {
		return DisplayNoSignal
	}
	switch facts.Activity {
	case ActivityStarting:
		return DisplayStarting
	case ActivityActive:
		return DisplayRunning
	case ActivityWaitingUser:
		return DisplayWaitingUser
	case ActivityIdle:
		return DisplayIdle
	default:
		return DisplayStarting
	}
}

func DeriveSessionDisplayStatus(facts SessionFacts, now time.Time, opts StatusOptions) DisplayStatus {
	return DeriveDisplayStatus(facts.Session, now, opts)
}
