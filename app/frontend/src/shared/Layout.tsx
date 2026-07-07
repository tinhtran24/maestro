import {
  Bot,
  GitBranch,
  GitFork,
} from "lucide-react";
import type { ViewId, Workspace, WorkspaceRegistry } from "../app/types";
import { navGroups, navItems } from "../app/viewRegistry";
import { useT } from "../i18n";

export { navItems };

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
  const t = useT();

  return (
    <div className="shell">
      <aside className="sidebar">
        <div className="brand">
          <div className="brand-mark">
            <img className="brand-logo" src="/favicon/logo_tui.png" alt="" />
          </div>
          <div>
            <strong>{t("app.brand")}</strong>
            <span>{t("app.subtitle")}</span>
          </div>
        </div>
        <nav className="nav">
          {navGroups.map((group) => (
            <div className="sidebar-section" key={group.label}>
              <span className="sidebar-section-label">{group.label}</span>
              {group.items.map((item) => {
                const Icon = item.icon;
                return (
                  <button key={item.id} className={activeView === item.id ? "active" : ""} onClick={() => onChangeView(item.id)}>
                    <Icon size={17} />
                    <span>{item.label}</span>
                  </button>
                );
              })}
            </div>
          ))}
        </nav>
        <div className="workspace-card">
          <span>{t("workspace.label")}</span>
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
          <small>{workspace.path || t("workspace.noFolder")}</small>
          {workspace.dataKey ? <small>{t("workspace.dataKey", { key: workspace.dataKey })}</small> : null}
          {workspace.folders.length > 1 ? <small>{t("workspace.attachedFolders", { count: workspace.folders.length })}</small> : null}
          <button onClick={onSelectFolder}>{t("workspace.selectFolder")}</button>
          {registry.activeWorkspaceId ? <button onClick={onRemoveWorkspace}>{t("workspace.remove")}</button> : null}
        </div>
      </aside>

      <main className="main">
        <header className="topbar">
          <div>
            <h1>{workspace.name}</h1>
            <p>{t("app.topbar.subtitle")}</p>
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
