import type { Workspace, WorkspaceRegistry } from "../app/types";

export const emptyWorkspace: Workspace = {
  workspaceId: "",
  dataKey: "",
  name: "No Workspace",
  path: "",
  folders: [],
  defaultBranch: "main",
  tasks: [],
  specs: [],
  routines: [],
  automation: {
    autoImplement: false,
    autoTest: false,
    autoSubmit: false,
    autoRetry: false,
    maxConcurrentRoutineTasks: 3,
    circuitBreakerFailureLimit: 3,
  },
  agents: [],
  flows: [],
  events: [],
  providers: [],
  diagnostics: [],
};

export const emptyRegistry: WorkspaceRegistry = {
  schema_version: 1,
  activeWorkspaceId: "",
  workspaces: [],
};

export function preserveWorkspaceIdentity(previous: Workspace, loaded: Workspace): Workspace {
  return {
    ...loaded,
    workspaceId: previous.workspaceId,
    dataKey: previous.dataKey,
    folders: previous.folders.length ? previous.folders : loaded.folders,
  };
}
