PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS projects (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  root_path TEXT NOT NULL,
  git_remote_url TEXT NOT NULL DEFAULT '',
  default_branch TEXT NOT NULL DEFAULT 'main',
  worktree_root TEXT NOT NULL DEFAULT '.thanos/worktrees',
  package_manager TEXT NOT NULL DEFAULT '',
  dev_command TEXT NOT NULL DEFAULT '',
  test_command TEXT NOT NULL DEFAULT '',
  settings_json TEXT NOT NULL DEFAULT '{}',
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS repos (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  path TEXT NOT NULL,
  remote TEXT,
  branch TEXT,
  is_primary INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS features (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  title TEXT NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  status TEXT NOT NULL CHECK (status IN ('backlog', 'active', 'done')),
  plan_graph_id TEXT,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS tasks (
  id TEXT PRIMARY KEY,
  feature_id TEXT REFERENCES features(id) ON DELETE SET NULL,
  parent_task_id TEXT REFERENCES tasks(id) ON DELETE SET NULL,
  title TEXT NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  status TEXT NOT NULL CHECK (status IN (
    'backlog',
    'planning',
    'waiting_approval',
    'ready',
    'running',
    'in_review',
    'waiting_user',
    'blocked',
    'done',
    'failed'
  )),
  priority TEXT NOT NULL CHECK (priority IN ('P0', 'P1', 'P2', 'P3')),
  assigned_agent TEXT,
  executor_profile TEXT,
  worktree_path TEXT,
  branch_name TEXT,
  review_approved INTEGER NOT NULL DEFAULT 0,
  tests_passed INTEGER NOT NULL DEFAULT 0,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_tasks_feature ON tasks(feature_id);
CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);
CREATE INDEX IF NOT EXISTS idx_tasks_parent ON tasks(parent_task_id);

CREATE TABLE IF NOT EXISTS execution_plans (
  id TEXT PRIMARY KEY,
  task_id TEXT NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
  summary TEXT NOT NULL DEFAULT '',
  steps_json TEXT NOT NULL DEFAULT '[]',
  risks_json TEXT NOT NULL DEFAULT '[]',
  files_to_touch_json TEXT NOT NULL DEFAULT '[]',
  test_strategy_json TEXT NOT NULL DEFAULT '[]',
  approval_status TEXT NOT NULL CHECK (approval_status IN ('draft', 'pending', 'approved', 'rejected', 'changes_requested')),
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_execution_plans_task ON execution_plans(task_id);

CREATE TABLE IF NOT EXISTS agent_profiles (
  name TEXT PRIMARY KEY,
  type TEXT NOT NULL CHECK (type IN ('planner', 'coder', 'reviewer', 'tester', 'utility')),
  provider TEXT NOT NULL,
  launch_mode TEXT NOT NULL CHECK (launch_mode IN ('acp', 'mcp', 'cli_passthrough')),
  command TEXT NOT NULL,
  env_json TEXT NOT NULL DEFAULT '{}',
  working_dir TEXT,
  permissions_json TEXT NOT NULL DEFAULT '[]',
  max_runtime TEXT,
  auto_continue INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS agent_sessions (
  id TEXT PRIMARY KEY,
  task_id TEXT NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
  agent_type TEXT NOT NULL,
  provider TEXT NOT NULL,
  command TEXT NOT NULL,
  status TEXT NOT NULL,
  pty_session_id TEXT,
  conversation_log_path TEXT,
  started_at TEXT NOT NULL,
  ended_at TEXT
);

CREATE TABLE IF NOT EXISTS runtime_session_facts (
  session_id TEXT PRIMARY KEY REFERENCES agent_sessions(id) ON DELETE CASCADE,
  project_id TEXT REFERENCES projects(id) ON DELETE CASCADE,
  task_id TEXT NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
  step TEXT NOT NULL,
  provider TEXT NOT NULL DEFAULT '',
  command TEXT NOT NULL DEFAULT '',
  args_json TEXT NOT NULL DEFAULT '[]',
  cwd TEXT NOT NULL DEFAULT '',
  runtime_handle_id TEXT NOT NULL DEFAULT '',
  agent_native_session_id TEXT NOT NULL DEFAULT '',
  transcript_path TEXT NOT NULL DEFAULT '',
  activity_state TEXT NOT NULL DEFAULT '',
  is_terminated INTEGER NOT NULL DEFAULT 0,
  exit_code INTEGER,
  display_status TEXT NOT NULL DEFAULT 'idle',
  started_at TEXT,
  last_output_at TEXT,
  ended_at TEXT,
  updated_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_runtime_session_facts_task ON runtime_session_facts(task_id, step);
CREATE INDEX IF NOT EXISTS idx_runtime_session_facts_status ON runtime_session_facts(display_status);

CREATE TABLE IF NOT EXISTS runtime_workspace_facts (
  id TEXT PRIMARY KEY,
  project_id TEXT REFERENCES projects(id) ON DELETE CASCADE,
  task_id TEXT NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
  session_id TEXT REFERENCES agent_sessions(id) ON DELETE SET NULL,
  step TEXT NOT NULL DEFAULT '',
  branch_name TEXT NOT NULL DEFAULT '',
  worktree_path TEXT NOT NULL DEFAULT '',
  prepared INTEGER NOT NULL DEFAULT 0,
  prepared_at TEXT
);

CREATE INDEX IF NOT EXISTS idx_runtime_workspace_facts_task ON runtime_workspace_facts(task_id, session_id);

CREATE TABLE IF NOT EXISTS runtime_terminal_facts (
  runtime_handle_id TEXT PRIMARY KEY,
  session_id TEXT NOT NULL REFERENCES agent_sessions(id) ON DELETE CASCADE,
  attached INTEGER NOT NULL DEFAULT 0,
  rows INTEGER NOT NULL DEFAULT 0,
  cols INTEGER NOT NULL DEFAULT 0,
  attached_at TEXT,
  detached_at TEXT
);

CREATE INDEX IF NOT EXISTS idx_runtime_terminal_facts_session ON runtime_terminal_facts(session_id);

CREATE TABLE IF NOT EXISTS reviews (
  id TEXT PRIMARY KEY,
  task_id TEXT NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
  diff_summary TEXT NOT NULL DEFAULT '',
  changed_files_json TEXT NOT NULL DEFAULT '[]',
  test_results_json TEXT NOT NULL DEFAULT '[]',
  reviewer_notes TEXT NOT NULL DEFAULT '',
  status TEXT NOT NULL CHECK (status IN ('pending', 'approved', 'rejected', 'changes_requested')),
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS skills (
  id TEXT PRIMARY KEY,
  project_id TEXT REFERENCES projects(id) ON DELETE CASCADE,
  name TEXT NOT NULL,
  path TEXT NOT NULL,
  description TEXT NOT NULL,
  applies_to_json TEXT NOT NULL DEFAULT '[]',
  agents_json TEXT NOT NULL DEFAULT '[]',
  version TEXT,
  source TEXT NOT NULL CHECK (source IN ('project', 'global', 'builtin')),
  required_evidence_json TEXT NOT NULL DEFAULT '[]',
  exit_criteria_json TEXT NOT NULL DEFAULT '[]',
  enabled INTEGER NOT NULL DEFAULT 1,
  trusted INTEGER NOT NULL DEFAULT 0,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  UNIQUE(project_id, name)
);

CREATE INDEX IF NOT EXISTS idx_skills_project ON skills(project_id, enabled);

CREATE TABLE IF NOT EXISTS skill_runs (
  id TEXT PRIMARY KEY,
  task_id TEXT NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
  skill_id TEXT NOT NULL REFERENCES skills(id) ON DELETE CASCADE,
  agent_session_id TEXT REFERENCES agent_sessions(id) ON DELETE SET NULL,
  status TEXT NOT NULL CHECK (status IN (
    'discovered',
    'matched',
    'activated',
    'running',
    'evidence_pending',
    'completed',
    'failed'
  )),
  evidence_json TEXT NOT NULL DEFAULT '{}',
  started_at TEXT NOT NULL,
  completed_at TEXT
);

CREATE INDEX IF NOT EXISTS idx_skill_runs_task ON skill_runs(task_id, status);

CREATE TABLE IF NOT EXISTS evidence (
  id TEXT PRIMARY KEY,
  task_id TEXT NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
  skill_run_id TEXT REFERENCES skill_runs(id) ON DELETE SET NULL,
  type TEXT NOT NULL,
  content TEXT NOT NULL,
  verified_by TEXT,
  created_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_evidence_task_type ON evidence(task_id, type);

CREATE TABLE IF NOT EXISTS memory_nodes (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  type TEXT NOT NULL CHECK (type IN ('feature', 'decision', 'architecture', 'file', 'task', 'bug', 'convention')),
  title TEXT NOT NULL,
  content TEXT NOT NULL,
  links_json TEXT NOT NULL DEFAULT '[]',
  embedding_json TEXT,
  created_at TEXT NOT NULL
);

CREATE VIRTUAL TABLE IF NOT EXISTS memory_nodes_fts USING fts5(
  title,
  content,
  content='memory_nodes',
  content_rowid='rowid'
);

CREATE TRIGGER IF NOT EXISTS memory_nodes_ai AFTER INSERT ON memory_nodes BEGIN
  INSERT INTO memory_nodes_fts(rowid, title, content) VALUES (new.rowid, new.title, new.content);
END;

CREATE TRIGGER IF NOT EXISTS memory_nodes_ad AFTER DELETE ON memory_nodes BEGIN
  INSERT INTO memory_nodes_fts(memory_nodes_fts, rowid, title, content)
  VALUES('delete', old.rowid, old.title, old.content);
END;

CREATE TRIGGER IF NOT EXISTS memory_nodes_au AFTER UPDATE ON memory_nodes BEGIN
  INSERT INTO memory_nodes_fts(memory_nodes_fts, rowid, title, content)
  VALUES('delete', old.rowid, old.title, old.content);
  INSERT INTO memory_nodes_fts(rowid, title, content) VALUES (new.rowid, new.title, new.content);
END;

CREATE TABLE IF NOT EXISTS events (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  project_id TEXT REFERENCES projects(id) ON DELETE CASCADE,
  task_id TEXT REFERENCES tasks(id) ON DELETE CASCADE,
  event TEXT NOT NULL,
  stage TEXT,
  status TEXT,
  artifact TEXT,
  payload_json TEXT NOT NULL DEFAULT '{}',
  created_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_events_task ON events(task_id, id);

CREATE TABLE IF NOT EXISTS change_log (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  project_id TEXT REFERENCES projects(id) ON DELETE CASCADE,
  task_id TEXT REFERENCES tasks(id) ON DELETE CASCADE,
  session_id TEXT REFERENCES agent_sessions(id) ON DELETE CASCADE,
  aggregate_type TEXT NOT NULL CHECK (aggregate_type IN ('task', 'gate', 'runtime', 'skill_run', 'read_model')),
  aggregate_id TEXT NOT NULL,
  change_type TEXT NOT NULL CHECK (change_type IN (
    'task_state_changed',
    'gate_state_changed',
    'runtime_fact_changed',
    'skill_run_changed',
    'derived_status_refresh_requested'
  )),
  payload_json TEXT NOT NULL DEFAULT '{}',
  derived_status TEXT,
  created_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_change_log_project ON change_log(project_id, id);
CREATE INDEX IF NOT EXISTS idx_change_log_task ON change_log(task_id, id);
CREATE INDEX IF NOT EXISTS idx_change_log_session ON change_log(session_id, id);
CREATE INDEX IF NOT EXISTS idx_change_log_type ON change_log(change_type, id);

CREATE TABLE IF NOT EXISTS runtime_events (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  project_id TEXT REFERENCES projects(id) ON DELETE CASCADE,
  task_id TEXT REFERENCES tasks(id) ON DELETE CASCADE,
  session_id TEXT REFERENCES agent_sessions(id) ON DELETE CASCADE,
  event_type TEXT NOT NULL,
  payload_json TEXT NOT NULL DEFAULT '{}',
  created_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_runtime_events_project ON runtime_events(project_id, id);
CREATE INDEX IF NOT EXISTS idx_runtime_events_task ON runtime_events(task_id, id);
CREATE INDEX IF NOT EXISTS idx_runtime_events_session ON runtime_events(session_id, id);

CREATE TRIGGER IF NOT EXISTS tasks_state_change_log_update
AFTER UPDATE OF status ON tasks
WHEN OLD.status <> NEW.status
BEGIN
  INSERT INTO change_log (task_id, aggregate_type, aggregate_id, change_type, payload_json, created_at)
  VALUES (
    NEW.id,
    'task',
    NEW.id,
    'task_state_changed',
    json_object('from', OLD.status, 'to', NEW.status),
    NEW.updated_at
  );
END;

CREATE TRIGGER IF NOT EXISTS tasks_gate_change_log_update
AFTER UPDATE OF review_approved, tests_passed ON tasks
WHEN OLD.review_approved <> NEW.review_approved OR OLD.tests_passed <> NEW.tests_passed
BEGIN
  INSERT INTO change_log (task_id, aggregate_type, aggregate_id, change_type, payload_json, created_at)
  VALUES (
    NEW.id,
    'gate',
    NEW.id,
    'gate_state_changed',
    json_object(
      'review_approved', json(CASE WHEN NEW.review_approved THEN 'true' ELSE 'false' END),
      'tests_passed', json(CASE WHEN NEW.tests_passed THEN 'true' ELSE 'false' END)
    ),
    NEW.updated_at
  );
END;

CREATE TRIGGER IF NOT EXISTS execution_plans_gate_change_log_update
AFTER UPDATE OF approval_status ON execution_plans
WHEN OLD.approval_status <> NEW.approval_status
BEGIN
  INSERT INTO change_log (task_id, aggregate_type, aggregate_id, change_type, payload_json, created_at)
  VALUES (
    NEW.task_id,
    'gate',
    NEW.id,
    'gate_state_changed',
    json_object('gate', 'plan_approval', 'from', OLD.approval_status, 'to', NEW.approval_status),
    NEW.updated_at
  );
END;

CREATE TRIGGER IF NOT EXISTS skill_runs_change_log_insert
AFTER INSERT ON skill_runs
BEGIN
  INSERT INTO change_log (task_id, session_id, aggregate_type, aggregate_id, change_type, payload_json, created_at)
  VALUES (
    NEW.task_id,
    NEW.agent_session_id,
    'skill_run',
    NEW.id,
    'skill_run_changed',
    json_object('status', NEW.status, 'skill_id', NEW.skill_id),
    NEW.started_at
  );
END;

CREATE TRIGGER IF NOT EXISTS skill_runs_change_log_update
AFTER UPDATE OF status, evidence_json ON skill_runs
WHEN OLD.status <> NEW.status OR OLD.evidence_json <> NEW.evidence_json
BEGIN
  INSERT INTO change_log (task_id, session_id, aggregate_type, aggregate_id, change_type, payload_json, created_at)
  VALUES (
    NEW.task_id,
    NEW.agent_session_id,
    'skill_run',
    NEW.id,
    'skill_run_changed',
    json_object('from', OLD.status, 'to', NEW.status, 'skill_id', NEW.skill_id),
    COALESCE(NEW.completed_at, NEW.started_at)
  );
END;

CREATE TRIGGER IF NOT EXISTS runtime_session_facts_change_log_insert
AFTER INSERT ON runtime_session_facts
BEGIN
  INSERT INTO change_log (project_id, task_id, session_id, aggregate_type, aggregate_id, change_type, payload_json, derived_status, created_at)
  VALUES (
    NEW.project_id,
    NEW.task_id,
    NEW.session_id,
    'runtime',
    NEW.session_id,
    'runtime_fact_changed',
    json_object('step', NEW.step, 'activity_state', NEW.activity_state, 'runtime_handle_id', NEW.runtime_handle_id),
    NEW.display_status,
    NEW.updated_at
  );
END;

CREATE TRIGGER IF NOT EXISTS runtime_session_facts_change_log_update
AFTER UPDATE ON runtime_session_facts
WHEN OLD.activity_state <> NEW.activity_state
  OR OLD.is_terminated <> NEW.is_terminated
  OR OLD.exit_code IS NOT NEW.exit_code
  OR OLD.runtime_handle_id <> NEW.runtime_handle_id
  OR OLD.display_status <> NEW.display_status
BEGIN
  INSERT INTO change_log (project_id, task_id, session_id, aggregate_type, aggregate_id, change_type, payload_json, derived_status, created_at)
  VALUES (
    NEW.project_id,
    NEW.task_id,
    NEW.session_id,
    'runtime',
    NEW.session_id,
    'runtime_fact_changed',
    json_object('activity_state', NEW.activity_state, 'is_terminated', json(CASE WHEN NEW.is_terminated THEN 'true' ELSE 'false' END), 'exit_code', NEW.exit_code),
    NEW.display_status,
    NEW.updated_at
  );
END;

CREATE TRIGGER IF NOT EXISTS runtime_session_facts_status_refresh_log_update
AFTER UPDATE OF activity_state, is_terminated, exit_code, runtime_handle_id, started_at, last_output_at, ended_at ON runtime_session_facts
BEGIN
  INSERT INTO change_log (project_id, task_id, session_id, aggregate_type, aggregate_id, change_type, payload_json, derived_status, created_at)
  VALUES (
    NEW.project_id,
    NEW.task_id,
    NEW.session_id,
    'read_model',
    NEW.session_id,
    'derived_status_refresh_requested',
    json_object('source', 'runtime_session_facts'),
    NEW.display_status,
    NEW.updated_at
  );
END;
