import type {
  AutomationInfo,
  ActivateWorkspaceRequest,
  BatchCreateTasksRequest,
  CommitTaskChangesRequest,
  CreateWorkspaceRequest,
  CreateSpecRequest,
  CreateTaskRequest,
  DeleteWorkspaceRequest,
  NativeTerminalRequest,
  NativeTerminalInputRequest,
  NativeTerminalResizeRequest,
  NativeTerminalSession,
  ProviderInfo,
  PrepareTaskCommitRequest,
  RegenerateOversightRequest,
  FinishTaskTurnRequest,
  ResumeTaskTurnRequest,
  RunTaskVerificationRequest,
  RoutineInfo,
  SaveAutomationRequest,
  SearchTasksRequest,
  SpecNode,
  StartTaskTurnRequest,
  Task,
  UpdateTaskFlagsRequest,
  UpdateWorkspaceRequest,
  UpdateTaskStatusRequest,
  UpsertRoutineRequest,
  Workspace,
  WorkspaceRecord,
  WorkspaceRegistry,
} from "../app/types";

export type AgentCandidate = ProviderInfo;

type WailsApp = {
  Ping?: () => Promise<string>;
  CurrentWorkspaceFolder?: () => Promise<string>;
  SelectWorkspaceFolder?: () => Promise<string | null>;
  DetectAgentCLIs?: () => Promise<AgentCandidate[]>;
  ListWorkspaces?: () => Promise<WorkspaceRegistry>;
  CreateWorkspace?: (request: CreateWorkspaceRequest) => Promise<WorkspaceRecord>;
  UpdateWorkspace?: (request: UpdateWorkspaceRequest) => Promise<WorkspaceRecord>;
  DeleteWorkspace?: (request: DeleteWorkspaceRequest) => Promise<WorkspaceRegistry>;
  ActivateWorkspace?: (request: ActivateWorkspaceRequest) => Promise<WorkspaceRecord>;
  LoadActiveWorkspace?: () => Promise<Workspace>;
  LoadWorkspace?: (root: string) => Promise<Workspace>;
  CreateTask?: (request: CreateTaskRequest) => Promise<Task>;
  BatchCreateTasks?: (request: BatchCreateTasksRequest) => Promise<Task[]>;
  SearchTasks?: (request: SearchTasksRequest) => Promise<Task[]>;
  UpdateTaskFlags?: (request: UpdateTaskFlagsRequest) => Promise<Task>;
  UpdateTaskStatus?: (request: UpdateTaskStatusRequest) => Promise<Task>;
  StartTaskTurn?: (request: StartTaskTurnRequest) => Promise<Task>;
  FinishTaskTurn?: (request: FinishTaskTurnRequest) => Promise<Task>;
  ResumeTaskTurn?: (request: ResumeTaskTurnRequest) => Promise<Task>;
  RunTaskVerification?: (request: RunTaskVerificationRequest) => Promise<Task>;
  PrepareTaskCommit?: (request: PrepareTaskCommitRequest) => Promise<Task>;
  RegenerateOversight?: (request: RegenerateOversightRequest) => Promise<Task>;
  CommitTaskChanges?: (request: CommitTaskChangesRequest) => Promise<Task>;
  CreateSpec?: (request: CreateSpecRequest) => Promise<SpecNode>;
  UpsertRoutine?: (request: UpsertRoutineRequest) => Promise<RoutineInfo>;
  SaveAutomation?: (request: SaveAutomationRequest) => Promise<AutomationInfo>;
  StartNativeTerminal?: (request: NativeTerminalRequest) => Promise<NativeTerminalSession>;
  WriteNativeTerminal?: (request: NativeTerminalInputRequest) => Promise<void>;
  ResizeNativeTerminal?: (request: NativeTerminalResizeRequest) => Promise<void>;
  StopNativeTerminal?: (sessionId: string) => Promise<boolean>;
};

type WailsRuntime = {
  EventsOn?: (name: string, callback: (payload: unknown) => void) => () => void;
};

type WailsWindow = Window & {
  go?: {
    app?: {
      App?: WailsApp;
    };
  };
  runtime?: WailsRuntime;
};

function app(): WailsApp {
  return ((window as WailsWindow).go?.app?.App ?? {}) as WailsApp;
}

export function hasWailsRuntime(): boolean {
  return Boolean((window as WailsWindow).go?.app?.App);
}

export async function getWorkspaceFolder(): Promise<string> {
  return (await app().CurrentWorkspaceFolder?.()) ?? "";
}

export async function selectWorkspaceFolder(): Promise<string | null> {
  return (await app().SelectWorkspaceFolder?.()) ?? null;
}

export async function detectAgents(): Promise<AgentCandidate[]> {
  return (await app().DetectAgentCLIs?.()) ?? [];
}

export async function listWorkspaces(): Promise<WorkspaceRegistry> {
  return (await app().ListWorkspaces?.()) ?? { schema_version: 1, activeWorkspaceId: "", workspaces: [] };
}

export async function createWorkspace(request: CreateWorkspaceRequest): Promise<WorkspaceRecord | null> {
  return (await app().CreateWorkspace?.(request)) ?? null;
}

export async function updateWorkspace(request: UpdateWorkspaceRequest): Promise<WorkspaceRecord | null> {
  return (await app().UpdateWorkspace?.(request)) ?? null;
}

export async function deleteWorkspace(request: DeleteWorkspaceRequest): Promise<WorkspaceRegistry> {
  return (await app().DeleteWorkspace?.(request)) ?? { schema_version: 1, activeWorkspaceId: "", workspaces: [] };
}

export async function activateWorkspace(request: ActivateWorkspaceRequest): Promise<WorkspaceRecord | null> {
  return (await app().ActivateWorkspace?.(request)) ?? null;
}

export async function loadActiveWorkspace(): Promise<Workspace | null> {
  return (await app().LoadActiveWorkspace?.()) ?? null;
}

export async function loadWorkspace(root: string): Promise<Workspace | null> {
  return (await app().LoadWorkspace?.(root)) ?? null;
}

export async function createTask(request: CreateTaskRequest): Promise<Task | null> {
  return (await app().CreateTask?.(request)) ?? null;
}

export async function batchCreateTasks(request: BatchCreateTasksRequest): Promise<Task[]> {
  return (await app().BatchCreateTasks?.(request)) ?? [];
}

export async function searchTasks(request: SearchTasksRequest): Promise<Task[]> {
  return (await app().SearchTasks?.(request)) ?? [];
}

export async function updateTaskFlags(request: UpdateTaskFlagsRequest): Promise<Task | null> {
  return (await app().UpdateTaskFlags?.(request)) ?? null;
}

export async function updateTaskStatus(request: UpdateTaskStatusRequest): Promise<Task | null> {
  return (await app().UpdateTaskStatus?.(request)) ?? null;
}

export async function startTaskTurn(request: StartTaskTurnRequest): Promise<Task | null> {
  return (await app().StartTaskTurn?.(request)) ?? null;
}

export async function finishTaskTurn(request: FinishTaskTurnRequest): Promise<Task | null> {
  return (await app().FinishTaskTurn?.(request)) ?? null;
}

export async function resumeTaskTurn(request: ResumeTaskTurnRequest): Promise<Task | null> {
  return (await app().ResumeTaskTurn?.(request)) ?? null;
}

export async function runTaskVerification(request: RunTaskVerificationRequest): Promise<Task | null> {
  return (await app().RunTaskVerification?.(request)) ?? null;
}

export async function prepareTaskCommit(request: PrepareTaskCommitRequest): Promise<Task | null> {
  return (await app().PrepareTaskCommit?.(request)) ?? null;
}

export async function regenerateOversight(request: RegenerateOversightRequest): Promise<Task | null> {
  return (await app().RegenerateOversight?.(request)) ?? null;
}

export async function commitTaskChanges(request: CommitTaskChangesRequest): Promise<Task | null> {
  return (await app().CommitTaskChanges?.(request)) ?? null;
}

export async function createSpec(request: CreateSpecRequest): Promise<SpecNode | null> {
  return (await app().CreateSpec?.(request)) ?? null;
}

export async function upsertRoutine(request: UpsertRoutineRequest): Promise<RoutineInfo | null> {
  return (await app().UpsertRoutine?.(request)) ?? null;
}

export async function saveAutomation(request: SaveAutomationRequest): Promise<AutomationInfo | null> {
  return (await app().SaveAutomation?.(request)) ?? null;
}

export async function startNativeTerminal(request: NativeTerminalRequest): Promise<NativeTerminalSession | null> {
  return (await app().StartNativeTerminal?.(request)) ?? null;
}

export async function writeNativeTerminal(request: NativeTerminalInputRequest): Promise<void> {
  await app().WriteNativeTerminal?.(request);
}

export async function resizeNativeTerminal(request: NativeTerminalResizeRequest): Promise<void> {
  await app().ResizeNativeTerminal?.(request);
}

export async function stopNativeTerminal(sessionId: string): Promise<boolean> {
  return (await app().StopNativeTerminal?.(sessionId)) ?? false;
}

export function onWailsEvent<T>(name: string, callback: (payload: T) => void): () => void {
  const unsubscribe = (window as WailsWindow).runtime?.EventsOn?.(name, (payload) => callback(payload as T));
  return unsubscribe ?? (() => {});
}
