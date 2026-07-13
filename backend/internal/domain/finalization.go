package domain

import "time"

// FinalizationState is the durable cursor for orchestrator-owned completion.
type FinalizationState string

const (
	// FinalizationPending records a worker completion waiting for orchestrator work.
	FinalizationPending FinalizationState = "pending"
	// FinalizationVerifyingGit records the git-state verification checkpoint.
	FinalizationVerifyingGit FinalizationState = "verifying_git"
	// FinalizationTesting records the required-test checkpoint.
	FinalizationTesting FinalizationState = "testing"
	// FinalizationCommitting records the optional commit checkpoint.
	FinalizationCommitting FinalizationState = "committing"
	// FinalizationPushing records the optional push checkpoint.
	FinalizationPushing FinalizationState = "pushing"
	// FinalizationClaimingPR records the pull-request claim/create checkpoint.
	FinalizationClaimingPR FinalizationState = "claiming_pr"
	// FinalizationPersistingMetadata records the metadata persistence checkpoint.
	FinalizationPersistingMetadata FinalizationState = "persisting_metadata"
	// FinalizationCleaningRuntime records the tmux/runtime cleanup checkpoint.
	FinalizationCleaningRuntime FinalizationState = "cleaning_runtime"
	// FinalizationReviewPending records the review-ready task transition checkpoint.
	FinalizationReviewPending FinalizationState = "review_pending"
	// FinalizationDone records the terminal task-done checkpoint.
	FinalizationDone FinalizationState = "done"
)

// SessionFinalization is the restart-safe record for one worker completion.
type SessionFinalization struct {
	SessionID      SessionID         `json:"sessionId"`
	OrchestratorID SessionID         `json:"orchestratorId,omitempty"`
	State          FinalizationState `json:"state"`
	RequestedAt    time.Time         `json:"requestedAt"`
	UpdatedAt      time.Time         `json:"updatedAt"`
	CompletedAt    *time.Time        `json:"completedAt,omitempty"`
	LastError      string            `json:"lastError,omitempty"`
}

var finalizationOrder = []FinalizationState{
	FinalizationPending,
	FinalizationVerifyingGit,
	FinalizationTesting,
	FinalizationCommitting,
	FinalizationPushing,
	FinalizationClaimingPR,
	FinalizationPersistingMetadata,
	FinalizationCleaningRuntime,
	FinalizationReviewPending,
	FinalizationDone,
}

// NextFinalizationState validates a single ordered, retry-safe transition.
func NextFinalizationState(current FinalizationState) (FinalizationState, bool) {
	for i := 0; i+1 < len(finalizationOrder); i++ {
		if finalizationOrder[i] == current {
			return finalizationOrder[i+1], true
		}
	}
	return "", false
}
