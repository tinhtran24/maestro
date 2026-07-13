-- +goose Up
CREATE TABLE session_finalization (
    session_id       TEXT PRIMARY KEY REFERENCES sessions (id) ON DELETE CASCADE,
    orchestrator_id  TEXT REFERENCES sessions (id),
    state            TEXT NOT NULL DEFAULT 'pending'
        CHECK (state IN ('pending', 'verifying_git', 'testing', 'committing', 'pushing', 'claiming_pr', 'persisting_metadata', 'cleaning_runtime', 'review_pending', 'done')),
    requested_at     TIMESTAMP NOT NULL,
    updated_at       TIMESTAMP NOT NULL,
    completed_at     TIMESTAMP,
    last_error       TEXT NOT NULL DEFAULT ''
);

-- +goose Down
DROP TABLE session_finalization;
