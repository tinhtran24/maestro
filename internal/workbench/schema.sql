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
