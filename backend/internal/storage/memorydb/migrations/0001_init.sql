-- +goose Up
-- +goose StatementBegin

-- memory_task holds one row per completed task, projected from a task.completed
-- event in the project's events.jsonl. This database is a disposable cache: it
-- can be dropped and rebuilt from the event log at any time, so it stores no
-- authoritative data and is never committed to git.
CREATE TABLE memory_task (
    id             TEXT PRIMARY KEY,
    event_id       TEXT NOT NULL,
    session_id     TEXT NOT NULL,
    project_id     TEXT NOT NULL,
    kind           TEXT NOT NULL DEFAULT '',
    harness        TEXT NOT NULL DEFAULT '',
    intent         TEXT NOT NULL DEFAULT '',
    task_type      TEXT NOT NULL DEFAULT '',
    branch         TEXT NOT NULL DEFAULT '',
    base_sha       TEXT NOT NULL DEFAULT '',
    head_sha       TEXT NOT NULL DEFAULT '',
    occurred_at    TIMESTAMP NOT NULL,
    prs_json       TEXT NOT NULL DEFAULT '[]',
    decisions_json TEXT NOT NULL DEFAULT '[]'
);

CREATE INDEX idx_memory_task_occurred_at ON memory_task(occurred_at DESC);
CREATE INDEX idx_memory_task_project ON memory_task(project_id, occurred_at DESC);

-- memory_file maps a task to each non-test file it changed. The path index backs
-- the "which prior tasks touched this file" relation query.
CREATE TABLE memory_file (
    task_id TEXT NOT NULL REFERENCES memory_task(id) ON DELETE CASCADE,
    path    TEXT NOT NULL,
    PRIMARY KEY (task_id, path)
);

CREATE INDEX idx_memory_file_path ON memory_file(path);

-- memory_test maps a task to each changed test file, kept separate from
-- memory_file so the resolver can weight test overlap distinctly.
CREATE TABLE memory_test (
    task_id TEXT NOT NULL REFERENCES memory_task(id) ON DELETE CASCADE,
    path    TEXT NOT NULL,
    PRIMARY KEY (task_id, path)
);

CREATE INDEX idx_memory_test_path ON memory_test(path);

-- memory_task_edge records a directed task-to-task relation with a confidence
-- score. The schema lands here; the relation resolver (a later task) populates
-- it. src/dst are ordered so a shared-file link is written once per pair.
CREATE TABLE memory_task_edge (
    src_task_id TEXT NOT NULL REFERENCES memory_task(id) ON DELETE CASCADE,
    dst_task_id TEXT NOT NULL REFERENCES memory_task(id) ON DELETE CASCADE,
    relation    TEXT NOT NULL,
    confidence  REAL NOT NULL DEFAULT 0,
    PRIMARY KEY (src_task_id, dst_task_id, relation)
);

CREATE INDEX idx_memory_task_edge_dst ON memory_task_edge(dst_task_id);

-- memory_meta is a single pinned row tracking the last event folded into this
-- projection, so catch-up applies only new events and rebuild is idempotent.
CREATE TABLE memory_meta (
    id            INTEGER PRIMARY KEY CHECK (id = 1),
    last_event_id TEXT NOT NULL DEFAULT '',
    updated_at    TIMESTAMP NOT NULL
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS memory_meta;
DROP TABLE IF EXISTS memory_task_edge;
DROP TABLE IF EXISTS memory_test;
DROP TABLE IF EXISTS memory_file;
DROP TABLE IF EXISTS memory_task;
-- +goose StatementEnd
