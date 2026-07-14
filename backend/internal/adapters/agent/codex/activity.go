package codex

import "github.com/tinhtran24/maestro/backend/internal/domain"

// DeriveActivityState maps a Codex hook event onto an Maestro activity state. The
// bool is false when the event carries no activity signal.
//
// event is the Maestro hook sub-command name installed in codexManagedHooks
// ("user-prompt-submit", "permission-request", "stop", ...), not the native
// Codex event name. Codex currently has no SessionEnd/Notification equivalent
// in the adapter, so runtime exit still falls back to the reaper.
func DeriveActivityState(event string, _ []byte) (domain.ActivityState, bool) {
	switch event {
	case "user-prompt-submit":
		return domain.ActivityActive, true
	case "permission-request":
		return domain.ActivityWaitingInput, true
	case "stop":
		return domain.ActivityIdle, true
	default:
		return "", false
	}
}
