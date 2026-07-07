import type { LucideIcon } from "lucide-react";

export type ViewId =
  | "board"
  | "plan"
  | "mission"
  | "agent-graph"
  | "chat"
  | "routines"
  | "whiteboard"
  | "analytics"
  | "settings"
  | "docs";

export type TaskStatus = "backlog" | "in_progress" | "waiting" | "committing" | "done" | "failed" | "cancelled";

export type NavItem = {
  id: ViewId;
  label: string;
  icon: LucideIcon;
};

export type Task = {
  schema_version: number;
  id: string;
  title: string;
  prompt: string;
  status: TaskStatus;
  flow: string;
  agent: string;
  branch: string;
  worktree: string;
  updatedAt: string;
  usageUsd: number;
  archived: boolean;
  deleted: boolean;
  tombstone: boolean;
  dependencies: string[];
  blocked: boolean;
  promptHistory: PromptRecord[];
  feedbackHistory: FeedbackRecord[];
  retryHistory: RetryRecord[];
  turns: TaskTurn[];
  lastTurn?: TaskTurn;
  lastOutput?: string;
  testsPassed: boolean;
  lastTestResult?: TestResultInfo;
  failureCategory: string;
  createdAt: string;
};

export type PromptRecord = {
  at: string;
  prompt: string;
};

export type FeedbackRecord = {
  at: string;
  message: string;
};

export type RetryRecord = {
  at: string;
  reason: string;
};

export type TaskTurn = {
  id: string;
  taskId: string;
  step: string;
  providerId: string;
  sessionId?: string;
  status: "running" | "completed" | "failed" | "stopped" | string;
  worktree: string;
  startedAt: string;
  endedAt?: string;
  stdoutPath?: string;
  stderrPath?: string;
  transcriptPath?: string;
  stopReason?: string;
  usageUsd: number;
  failureCategory?: string;
  autoContinue: boolean;
};

export type TestResultInfo = {
  id: string;
  taskId: string;
  providerId?: string;
  command: string;
  status: "passed" | "failed" | string;
  passed: boolean;
  outputPath: string;
  output: string;
  exitCode: number;
  passPattern?: string;
  failPattern?: string;
  startedAt: string;
  endedAt: string;
};

export type SpecNode = {
  id: string;
  title: string;
  state: "vague" | "drafted" | "validated" | "testing" | "complete" | "stale";
  path: string;
  children: SpecNode[];
};

export type AgentNode = {
  id: string;
  role: string;
  harness: "Claude" | "Codex" | "Cursor" | "OpenCode" | "Gemini" | "Shell";
  model: string;
  capabilities: string[];
  readOnly: boolean;
  source?: string;
};

export type FlowNode = {
  id: string;
  name: string;
  steps: string[];
  parallelGroups?: string[][];
  readOnly: boolean;
  source?: string;
};

export type EventRecord = {
  id: string;
  at: string;
  kind: string;
  message: string;
};

export type RoutineInfo = {
  id: string;
  name: string;
  prompt: string;
  flow: string;
  schedule: string;
  enabled: boolean;
  updatedAt: string;
};

export type AutomationInfo = {
  autoImplement: boolean;
  autoTest: boolean;
  autoSubmit: boolean;
  autoRetry: boolean;
};

export type CreateTaskRequest = {
  root: string;
  title: string;
  prompt: string;
  flow: string;
  agent: string;
  dependencies?: string[];
};

export type UpdateTaskStatusRequest = {
  root: string;
  taskId: string;
  status: TaskStatus;
  feedback?: string;
  failureCategory?: string;
};

export type StartTaskTurnRequest = {
  root: string;
  taskId: string;
  step?: string;
  providerId?: string;
  sessionId?: string;
  transcriptPath?: string;
};

export type FinishTaskTurnRequest = {
  root: string;
  taskId: string;
  turnId: string;
  status: string;
  stdout?: string;
  stderr?: string;
  stopReason?: string;
  usageUsd?: number;
  exitCode?: number;
  autoContinue?: boolean;
};

export type ResumeTaskTurnRequest = {
  root: string;
  taskId: string;
  feedback: string;
};

export type RunTaskVerificationRequest = {
  root: string;
  taskId: string;
  command: string;
  providerId?: string;
  passPattern?: string;
  failPattern?: string;
};

export type BatchCreateTasksRequest = {
  root: string;
  tasks: CreateTaskRequest[];
};

export type SearchTasksRequest = {
  root: string;
  query: string;
  includeArchived: boolean;
  includeDeleted: boolean;
};

export type UpdateTaskFlagsRequest = {
  root: string;
  taskId: string;
  archived: boolean;
  deleted: boolean;
  tombstone: boolean;
};

export type CreateSpecRequest = {
  root: string;
  title: string;
  body: string;
  state: SpecNode["state"];
};

export type UpsertRoutineRequest = {
  root: string;
  id?: string;
  name: string;
  prompt: string;
  flow: string;
  schedule: string;
  enabled: boolean;
};

export type SaveAutomationRequest = {
  root: string;
  automation: AutomationInfo;
};

export type ProviderInfo = {
  id: string;
  name: string;
  command: string;
  status: "installed" | "not_found" | "needs_setup" | string;
  path?: string;
  version?: string;
  type: "cli" | "shell" | "mcp" | "acp" | string;
  setupHint: string;
  supportsRun: boolean;
};

export type NativeTerminalRequest = {
  providerId?: string;
  taskId?: string;
  step?: string;
  command: string;
  args: string[];
  cwd: string;
  label: string;
  rows?: number;
  cols?: number;
};

export type NativeTerminalInputRequest = {
  sessionId: string;
  data: string;
};

export type NativeTerminalResizeRequest = {
  sessionId: string;
  rows: number;
  cols: number;
};

export type NativeTerminalSession = {
  id: string;
  label: string;
  providerId: string;
  taskId: string;
  step: string;
  command: string;
  args: string[];
  cwd: string;
  status: "running" | "completed" | "failed" | "stopped" | string;
  ptyId: string;
  transcriptPath: string;
  startedAt: string;
  endedAt?: string;
};

export type DiagnosticInfo = {
  kind: string;
  message: string;
};

export type Workspace = {
  workspaceId: string;
  dataKey: string;
  name: string;
  path: string;
  folders: WorkspaceFolderInfo[];
  defaultBranch: string;
  tasks: Task[];
  specs: SpecNode[];
  routines: RoutineInfo[];
  automation: AutomationInfo;
  agents: AgentNode[];
  flows: FlowNode[];
  events: EventRecord[];
  providers: ProviderInfo[];
  diagnostics: DiagnosticInfo[];
};

export type WorkspaceFolderInfo = {
  id: string;
  path: string;
  label: string;
};

export type WorkspaceRecord = {
  schema_version: number;
  id: string;
  name: string;
  dataKey: string;
  folders: WorkspaceFolderInfo[];
  activeFolderId: string;
  createdAt: string;
  updatedAt: string;
};

export type WorkspaceRegistry = {
  schema_version: number;
  activeWorkspaceId: string;
  workspaces: WorkspaceRecord[];
};

export type CreateWorkspaceRequest = {
  name?: string;
  path: string;
  folders?: WorkspaceFolderInfo[];
};

export type UpdateWorkspaceRequest = {
  id: string;
  name?: string;
  folders?: WorkspaceFolderInfo[];
  activeFolderId?: string;
};

export type ActivateWorkspaceRequest = {
  id: string;
};

export type DeleteWorkspaceRequest = {
  id: string;
};
