-- name: UpsertMemoryTask :exec
INSERT INTO memory_task (
    id, event_id, session_id, project_id, kind, harness, intent, task_type,
    branch, base_sha, head_sha, occurred_at, prs_json, decisions_json
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(id) DO UPDATE SET
    event_id = excluded.event_id,
    session_id = excluded.session_id,
    project_id = excluded.project_id,
    kind = excluded.kind,
    harness = excluded.harness,
    intent = excluded.intent,
    task_type = excluded.task_type,
    branch = excluded.branch,
    base_sha = excluded.base_sha,
    head_sha = excluded.head_sha,
    occurred_at = excluded.occurred_at,
    prs_json = excluded.prs_json,
    decisions_json = excluded.decisions_json;

-- name: GetMemoryTask :one
SELECT * FROM memory_task WHERE id = ?;

-- name: ListMemoryTasks :many
SELECT * FROM memory_task ORDER BY occurred_at DESC LIMIT ?;

-- name: CountMemoryTasks :one
SELECT COUNT(*) FROM memory_task;

-- name: AddMemoryFile :exec
INSERT INTO memory_file (task_id, path) VALUES (?, ?)
ON CONFLICT(task_id, path) DO NOTHING;

-- name: ListMemoryFilesByTask :many
SELECT path FROM memory_file WHERE task_id = ? ORDER BY path;

-- name: DeleteMemoryFilesByTask :exec
DELETE FROM memory_file WHERE task_id = ?;

-- name: AddMemoryTest :exec
INSERT INTO memory_test (task_id, path) VALUES (?, ?)
ON CONFLICT(task_id, path) DO NOTHING;

-- name: ListMemoryTestsByTask :many
SELECT path FROM memory_test WHERE task_id = ? ORDER BY path;

-- name: DeleteMemoryTestsByTask :exec
DELETE FROM memory_test WHERE task_id = ?;

-- name: GetMemoryMeta :one
SELECT last_event_id, updated_at FROM memory_meta WHERE id = 1;

-- name: UpsertMemoryMeta :exec
INSERT INTO memory_meta (id, last_event_id, updated_at) VALUES (1, ?, ?)
ON CONFLICT(id) DO UPDATE SET
    last_event_id = excluded.last_event_id,
    updated_at = excluded.updated_at;

-- name: TasksSharingFiles :many
SELECT t.id, t.occurred_at, COUNT(f.path) AS shared
FROM memory_file f
JOIN memory_task t ON t.id = f.task_id
WHERE f.path IN (sqlc.slice('paths'))
GROUP BY t.id, t.occurred_at
ORDER BY shared DESC, t.occurred_at DESC;

-- name: TasksSharingTests :many
SELECT t.id, t.occurred_at, COUNT(te.path) AS shared
FROM memory_test te
JOIN memory_task t ON t.id = te.task_id
WHERE te.path IN (sqlc.slice('paths'))
GROUP BY t.id, t.occurred_at
ORDER BY shared DESC, t.occurred_at DESC;

-- name: DeleteEdgesFrom :exec
DELETE FROM memory_task_edge WHERE src_task_id = ?;

-- name: AddEdge :exec
INSERT INTO memory_task_edge (src_task_id, dst_task_id, relation, confidence)
VALUES (?, ?, ?, ?)
ON CONFLICT(src_task_id, dst_task_id, relation) DO UPDATE SET confidence = excluded.confidence;

-- name: ListEdgesFrom :many
SELECT dst_task_id, relation, confidence
FROM memory_task_edge
WHERE src_task_id = ?
ORDER BY confidence DESC, dst_task_id;
