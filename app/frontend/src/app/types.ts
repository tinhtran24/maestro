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

export type TaskStatus = "backlog" | "running" | "waiting" | "review" | "done" | "failed";

export type NavItem = {
  id: ViewId;
  label: string;
  icon: LucideIcon;
};

export type Task = {
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
};

export type FlowNode = {
  id: string;
  name: string;
  steps: string[];
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
};

export type UpdateTaskStatusRequest = {
  root: string;
  taskId: string;
  status: TaskStatus;
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
  command: string;
  args: string[];
  cwd: string;
  label: string;
};

export type NativeTerminalSession = {
  id: string;
  label: string;
  command: string;
  args: string[];
  cwd: string;
  status: "running" | "completed" | "failed" | "stopped" | string;
  startedAt: string;
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
