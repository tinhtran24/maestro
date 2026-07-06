import { useEffect, useMemo, useState } from "react";
import type { ViewId, Workspace, WorkspaceRegistry } from "./types";
import { AgentGraphView } from "../flows/AgentGraphView";
import { BoardView } from "../flows/BoardView";
import { MissionView } from "../flows/MissionView";
import { PlanView } from "../flows/PlanView";
import { AnalyticsView, ChatView, DocsView, RoutinesView, SettingsView, WhiteboardView } from "../flows/UtilityViews";
import {
  activateWorkspace,
  createWorkspace,
  deleteWorkspace,
  detectAgents,
  getWorkspaceFolder,
  hasWailsRuntime,
  listWorkspaces,
  loadActiveWorkspace,
  loadWorkspace,
  selectWorkspaceFolder,
} from "../services/wails";
import { Layout } from "../shared/Layout";

const emptyWorkspace: Workspace = {
  workspaceId: "",
  dataKey: "",
  name: "No Workspace",
  path: "",
  folders: [],
  defaultBranch: "main",
  tasks: [],
  specs: [],
  routines: [],
  automation: { autoImplement: false, autoTest: false, autoSubmit: false, autoRetry: false },
  agents: [],
  flows: [],
  events: [],
  providers: [],
  diagnostics: [],
};

const emptyRegistry: WorkspaceRegistry = {
  schema_version: 1,
  activeWorkspaceId: "",
  workspaces: [],
};

export function App() {
  const [activeView, setActiveView] = useState<ViewId>("board");
  const [workspace, setWorkspace] = useState<Workspace>(emptyWorkspace);
  const [registry, setRegistry] = useState<WorkspaceRegistry>(emptyRegistry);
  const [loadError, setLoadError] = useState("");

  useEffect(() => {
    void bootWorkspace();
  }, []);

  const runtime = useMemo(() => (hasWailsRuntime() ? "Wails" : "Browser preview"), []);

  async function bootWorkspace() {
    setLoadError("");
    try {
      const records = await listWorkspaces();
      setRegistry(records);
      const loaded = records.activeWorkspaceId ? await loadActiveWorkspace() : null;
      if (loaded) {
        setWorkspace(loaded);
        return;
      }
      const path = await getWorkspaceFolder();
      const fallback = await loadWorkspace(path);
      if (fallback) {
        setWorkspace(fallback);
        return;
      }
      const providers = await detectAgents();
      setWorkspace((current) => ({ ...current, providers }));
    } catch (error) {
      setLoadError(error instanceof Error ? error.message : String(error));
    }
  }

  async function chooseFolder() {
    const path = await selectWorkspaceFolder();
    if (!path) return;
    setLoadError("");
    try {
      const record = await createWorkspace({ path });
      const nextRegistry = await listWorkspaces();
      setRegistry(nextRegistry);
      const loaded = record ? await loadActiveWorkspace() : await loadWorkspace(path);
      setWorkspace(loaded ?? { ...emptyWorkspace, path });
    } catch (error) {
      setLoadError(error instanceof Error ? error.message : String(error));
    }
  }

  async function switchWorkspace(workspaceId: string) {
    if (!workspaceId) return;
    setLoadError("");
    try {
      await activateWorkspace({ id: workspaceId });
      const [nextRegistry, loaded] = await Promise.all([listWorkspaces(), loadActiveWorkspace()]);
      setRegistry(nextRegistry);
      if (loaded) setWorkspace(loaded);
    } catch (error) {
      setLoadError(error instanceof Error ? error.message : String(error));
    }
  }

  async function removeActiveWorkspace() {
    if (!registry.activeWorkspaceId) return;
    setLoadError("");
    try {
      const nextRegistry = await deleteWorkspace({ id: registry.activeWorkspaceId });
      setRegistry(nextRegistry);
      const loaded = nextRegistry.activeWorkspaceId ? await loadActiveWorkspace() : null;
      setWorkspace(loaded ?? emptyWorkspace);
    } catch (error) {
      setLoadError(error instanceof Error ? error.message : String(error));
    }
  }

  async function refreshWorkspace() {
    if (!workspace.path) return;
    const loaded = await loadWorkspace(workspace.path);
    if (loaded) {
      setWorkspace({
        ...loaded,
        workspaceId: workspace.workspaceId,
        dataKey: workspace.dataKey,
        folders: workspace.folders.length ? workspace.folders : loaded.folders,
      });
    }
  }

  return (
    <Layout
      activeView={activeView}
      workspace={workspace}
      registry={registry}
      onActivateWorkspace={switchWorkspace}
      onRemoveWorkspace={removeActiveWorkspace}
      onChangeView={setActiveView}
      onSelectFolder={chooseFolder}
    >
      {loadError ? <div className="error-banner">{loadError}</div> : null}
      {activeView === "board" ? <BoardView workspace={workspace} onReload={refreshWorkspace} /> : null}
      {activeView === "plan" ? <PlanView workspace={workspace} onReload={refreshWorkspace} /> : null}
      {activeView === "mission" ? <MissionView workspace={workspace} /> : null}
      {activeView === "agent-graph" ? <AgentGraphView workspace={workspace} /> : null}
      {activeView === "chat" ? <ChatView /> : null}
      {activeView === "routines" ? <RoutinesView workspace={workspace} onReload={refreshWorkspace} /> : null}
      {activeView === "whiteboard" ? <WhiteboardView /> : null}
      {activeView === "analytics" ? <AnalyticsView workspace={workspace} /> : null}
      {activeView === "settings" ? <SettingsView runtime={runtime} workspace={workspace} onReload={refreshWorkspace} /> : null}
      {activeView === "docs" ? <DocsView /> : null}
    </Layout>
  );
}
