import {
  BarChart3,
  Bot,
  BrainCircuit,
  CalendarClock,
  FileText,
  GitBranch,
  GitFork,
  Kanban,
  Map,
  MessageSquare,
  Settings,
  Workflow,
} from "lucide-react";
import type { NavItem, ViewId, Workspace, WorkspaceRegistry } from "../app/types";

export const navItems: NavItem[] = [
  { id: "board", label: "Board", icon: Kanban },
  { id: "plan", label: "Plan", icon: FileText },
  { id: "mission", label: "Mission", icon: Map },
  { id: "agent-graph", label: "Agent Graph", icon: Workflow },
  { id: "chat", label: "Chat", icon: MessageSquare },
  { id: "routines", label: "Routines", icon: CalendarClock },
  { id: "whiteboard", label: "Whiteboard", icon: BrainCircuit },
  { id: "analytics", label: "Analytics", icon: BarChart3 },
  { id: "settings", label: "Settings", icon: Settings },
];

export function Layout({
  activeView,
  workspace,
  registry,
  onActivateWorkspace,
  onRemoveWorkspace,
  onChangeView,
  onSelectFolder,
  children,
}: {
  activeView: ViewId;
  workspace: Workspace;
  registry: WorkspaceRegistry;
  onActivateWorkspace: (workspaceId: string) => void;
  onRemoveWorkspace: () => void;
  onChangeView: (view: ViewId) => void;
  onSelectFolder: () => void;
  children: React.ReactNode;
}) {
  return (
    <div className="shell">
      <aside className="sidebar">
        <div className="brand">
          <div className="brand-mark">T</div>
          <div>
            <strong>Thanos</strong>
            <span>AI Development Workbench</span>
          </div>
        </div>
        <nav className="nav">
          {navItems.map((item) => {
            const Icon = item.icon;
            return (
              <button key={item.id} className={activeView === item.id ? "active" : ""} onClick={() => onChangeView(item.id)}>
                <Icon size={17} />
                <span>{item.label}</span>
              </button>
            );
          })}
        </nav>
        <div className="workspace-card">
          <span>Workspace</span>
          <strong>{workspace.name}</strong>
          {registry.workspaces.length > 0 ? (
            <select value={registry.activeWorkspaceId} onChange={(event) => onActivateWorkspace(event.target.value)}>
              {registry.workspaces.map((record) => (
                <option key={record.id} value={record.id}>
                  {record.name}
                </option>
              ))}
            </select>
          ) : null}
          <small>{workspace.path || "No folder selected"}</small>
          {workspace.dataKey ? <small>Data key: {workspace.dataKey}</small> : null}
          {workspace.folders.length > 1 ? <small>{workspace.folders.length} folders attached</small> : null}
          <button onClick={onSelectFolder}>Select Folder</button>
          {registry.activeWorkspaceId ? <button onClick={onRemoveWorkspace}>Remove Workspace</button> : null}
        </div>
      </aside>

      <main className="main">
        <header className="topbar">
          <div>
            <h1>{workspace.name}</h1>
            <p>Host-native Wails shell · local state first · human approval before merge</p>
          </div>
          <div className="topbar-meta">
            <span><GitBranch size={15} /> {workspace.defaultBranch}</span>
            <span><Bot size={15} /> {workspace.providers.filter((provider) => provider.status === "installed").length} providers</span>
            <span><GitFork size={15} /> {workspace.flows.length} flows</span>
          </div>
        </header>
        <div className="content">{children}</div>
      </main>
    </div>
  );
}
