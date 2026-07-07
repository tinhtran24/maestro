import { useState } from "react";
import type { ViewId } from "./types";
import { useWorkspaceController } from "../hooks/useWorkspaceController";
import { Layout } from "../shared/Layout";
import {
  AgentGraphView,
  AnalyticsView,
  BoardView,
  ChatView,
  DocsView,
  MissionView,
  PlanView,
  RoutinesView,
  SettingsView,
  WhiteboardView,
} from "../views";

export function App() {
  const [activeView, setActiveView] = useState<ViewId>("board");
  const {
    workspace,
    registry,
    loadError,
    runtime,
    chooseFolder,
    switchWorkspace,
    removeActiveWorkspace,
    refreshWorkspace,
  } = useWorkspaceController();

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
