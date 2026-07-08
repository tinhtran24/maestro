import type { LucideIcon } from "lucide-react";

export type ViewId =
  | "board"
  | "plan"
  | "mission"
  | "agent-graph"
  | "chat"
  | "routines"
  | "whiteboard"
  | "files"
  | "analytics"
  | "settings"
  | "docs";

export type TaskStatus = "backlog" | "in_progress" | "waiting" | "committing" | "done" | "failed" | "cancelled";

export type NavItem = {
  id: ViewId;
  label: string;
  icon: LucideIcon;
};

export type NavGroup = {
  label: string;
  items: NavItem[];
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
  commit?: CommitInfo;
  oversight?: OversightInfo;
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

export type CommitInfo = {
  hash?: string;
  summary: string;
  message: string;
  diff: string;
  diffStat: string;
  approved: boolean;
  committed: boolean;
  committedAt?: string;
};

export type OversightInfo = {
  schema_version: number;
  id: string;
  taskId: string;
  status: string;
  summary: string;
  phases: string[];
  risks: string[];
  changedFiles: string[];
  commands: string[];
  testResult: string;
  usageUsd: number;
  generatedAt: string;
  path: string;
  testPath?: string;
};

export type SpecNode = {
  id: string;
  title: string;
  state: "vague" | "drafted" | "validated" | "testing" | "complete" | "stale" | "archived";
  path: string;
  body: string;
  updatedAt: string;
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
  lastRunAt?: string;
  nextRunAt?: string;
  runCount: number;
  failureCount: number;
  disabledReason?: string;
  updatedAt: string;
};

export type AutomationInfo = {
  autoImplement: boolean;
  autoTest: boolean;
  autoSubmit: boolean;
  autoRetry: boolean;
  maxConcurrentRoutineTasks: number;
  circuitBreakerFailureLimit: number;
};

export type CreateTaskRequest = {
  root: string;
  title: string;
  prompt: string;
  flow: string;
  agent: string;
  status?: TaskStatus;
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

export type PrepareTaskCommitRequest = {
  root: string;
  taskId: string;
  message?: string;
};

export type CommitTaskChangesRequest = {
  root: string;
  taskId: string;
  message?: string;
  approved: boolean;
};

export type RegenerateOversightRequest = {
  root: string;
  taskId: string;
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
  parentPath?: string;
};

export type UpdateSpecRequest = {
  root: string;
  path: string;
  title: string;
  body: string;
  state: SpecNode["state"];
};

export type DispatchSpecsRequest = {
  root: string;
  path?: string;
};

export type UndoPlanningChangeRequest = {
  root: string;
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

export type TriggerRoutineRequest = {
  root: string;
  routineId: string;
};

export type RunRoutineSchedulerRequest = {
  root: string;
};

export type ListWorkspaceFilesRequest = {
  root: string;
  path?: string;
  maxDepth?: number;
};

export type ReadWorkspaceFileRequest = {
  root: string;
  path: string;
};

export type WriteWorkspaceFileRequest = {
  root: string;
  path: string;
  content: string;
};

export type PreviewTaskDiffRequest = {
  root: string;
  taskId: string;
};

export type FileTreeEntry = {
  name: string;
  path: string;
  kind: "directory" | "file" | "virtual" | string;
  size: number;
  modifiedAt?: string;
  readOnly: boolean;
  children?: FileTreeEntry[];
};

export type FileExplorerInfo = {
  root: string;
  base: string;
  entries: FileTreeEntry[];
};

export type WorkspaceFileInfo = {
  path: string;
  name: string;
  content: string;
  encoding: string;
  size: number;
  modifiedAt?: string;
  readOnly: boolean;
  virtual: boolean;
};

export type TaskDiffPreviewInfo = {
  taskId: string;
  worktree: string;
  diff: string;
  diffStat: string;
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

export type AgentMode = "planner" | "coding" | "review" | "debug" | "research";

export type AgentEventRecord = {
  sessionId: string;
  type: string;
  payload: string;
  time: string;
};

export type AgentSession = {
  id: string;
  providerId: string;
  providerName: string;
  projectId: string;
  taskId: string;
  mode: AgentMode;
  status: "idle" | "starting" | "running" | "waiting_user" | "completed" | "failed" | "stopped" | "archived" | string;
  workdir: string;
  prompt: string;
  promptHistory: PromptRecord[];
  conversation: AgentEventRecord[];
  terminalId: string;
  transcriptPath: string;
  approvedPlan?: string;
  result?: string;
  createdAt: string;
  updatedAt: string;
};

export type StartAgentRequest = {
  root: string;
  providerId: string;
  projectId?: string;
  taskId?: string;
  mode: AgentMode;
  prompt: string;
  context?: string;
  acceptanceCriteria?: string;
  constraints?: string;
  allowedFiles?: string[];
  previousPlan?: string;
  expectedOutput?: string;
  customCommand?: string;
  rows?: number;
  cols?: number;
};

export type SendAgentInputRequest = {
  root: string;
  sessionId: string;
  input: string;
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
