-- name: RequestSessionFinalization :exec
INSERT INTO session_finalization (session_id, state, requested_at, updated_at)
VALUES (?, 'pending', ?, ?)
ON CONFLICT(session_id) DO UPDATE SET
    updated_at = CASE
        WHEN session_finalization.state = 'pending' THEN excluded.updated_at
        ELSE session_finalization.updated_at
    END;

-- name: GetSessionFinalization :one
SELECT session_id, orchestrator_id, state, requested_at, updated_at, completed_at, last_error
FROM session_finalization
WHERE session_id = ?;

-- name: ClaimSessionFinalization :execrows
UPDATE session_finalization
SET orchestrator_id = ?, updated_at = ?, last_error = ''
WHERE session_id = ?
  AND state <> 'done'
  AND (orchestrator_id IS NULL OR orchestrator_id = ?);

-- name: AdvanceSessionFinalization :execrows
UPDATE session_finalization
SET state = ?, updated_at = ?, completed_at = ?, last_error = ''
WHERE session_id = ? AND orchestrator_id = ? AND state = ?;

-- name: FailSessionFinalization :execrows
UPDATE session_finalization
SET last_error = ?, updated_at = ?
WHERE session_id = ? AND orchestrator_id = ? AND state <> 'done';

-- name: ListPendingSessionFinalizations :many
SELECT session_id, orchestrator_id, state, requested_at, updated_at, completed_at, last_error
FROM session_finalization
WHERE state <> 'done'
ORDER BY requested_at, session_id;
