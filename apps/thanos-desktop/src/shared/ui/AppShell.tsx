import { useQueryClient } from "@tanstack/react-query";
import { Bell, Bot, Boxes, Brain, Check, ChevronDown, CircleDot, Cpu, FolderGit2, FolderKanban, GitBranch, Kanban, LayoutDashboard, ListChecks, MoreHorizontal, Plus, RotateCw, Search, Settings } from "lucide-react";
import { useState } from "react";
import thanosLogo from "../../../src-tauri/imgs/logo/thanos-logo.png";
import type { Project } from "../../domain/models";
import { useProjects, useRecentProjects } from "../../features/projects/hooks/useProjects";
import { NativeBackend } from "../../services/nativeBackend";
import { useWorkbenchStore } from "../../state/workbenchStore";
import { Dropdown } from "./Dropdown";
import { SidebarItem } from "./SidebarItem";

const backend = new NativeBackend();

const nav = [
  { id: "workbench" as const, icon: LayoutDashboard, label: "Workbench" },
  { id: "projects" as const, icon: FolderKanban, label: "Projects" },
  { id: "memory" as const, icon: Brain, label: "Memory" },
  { id: "executors" as const, icon: Cpu, label: "Executors" },
  { id: "workflow_steps" as const, icon: ListChecks, label: "Workflow Steps" },
];

export function AppShell({ project, children }: { project: Project; children: React.ReactNode }) {
  const activeView = useWorkbenchStore((state) => state.activeView);
  const setActiveView = useWorkbenchStore((state) => state.setActiveView);
  const setProject = useWorkbenchStore((state) => state.setProject);
  const openCreateTask = useWorkbenchStore((state) => state.openCreateTask);
  const installed = useWorkbenchStore((state) => state.agentProviders.filter((agent) => agent.enabled && agent.status === "installed").length);
  const allProjects = useProjects((state) => state.projects);
  const setActiveProject = useProjects((state) => state.setActive);
  const reloadProjects = useProjects((state) => state.reload);
  const recentProjects = useRecentProjects(4);
  const queryClient = useQueryClient();

  const branch = project.defaultBranch || project.settings.defaultBranch || "main";
  const worktreeRoot = project.worktreeRoot || project.settings.worktreeRoot || ".thanos/worktrees";

  function switchProject(next: Project) {
    setActiveProject(next.id);
    backend.setWorkspacePath(next.rootPath);
    setProject(next);
    queryClient.invalidateQueries({ queryKey: ["workbench"] });
  }

  function changeBranch(nextBranch: string) {
    const value = nextBranch.trim();
    if (!value) return;
    setProject({ ...project, defaultBranch: value, settings: { ...project.settings, defaultBranch: value } });
  }

  const switchable = allProjects.filter((item) => !item.archived && item.id !== project.id && item.rootPath !== project.rootPath);
  const branchSuggestions = Array.from(new Set([branch, "main", "master", "develop"]));

  return (
    <div className="grid h-screen grid-rows-[auto_minmax(0,1fr)] overflow-hidden bg-bg-app text-text-main lg:grid-cols-[260px_minmax(0,1fr)] lg:grid-rows-1">
      <aside className="flex min-w-0 flex-col gap-4 border-b border-slate-800 bg-bg-sidebar p-4 lg:min-h-0 lg:border-b-0 lg:border-r">
        <div className="flex items-center gap-3">
          <img src={thanosLogo} alt="Thanos" className="h-9 w-32 shrink-0 object-contain object-left" />
        </div>
        <p className="-mt-2 text-[11px] font-medium uppercase tracking-[0.18em] text-text-muted">AI Development Workbench</p>

        <nav className="flex min-w-0 gap-1 overflow-x-auto pb-1 lg:grid lg:gap-1 lg:overflow-visible lg:pb-0">
          {nav.map((item) => (
            <SidebarItem key={item.id} icon={item.icon} label={item.label} active={activeView === item.id} onClick={() => setActiveView(item.id)} />
          ))}
        </nav>

        <div className="hidden min-h-0 flex-1 flex-col lg:flex">
          <div className="rounded-xl border border-slate-800 bg-slate-900/60 p-3">
            <div className="flex items-center justify-between">
              <span className="text-xs font-semibold uppercase tracking-wide text-text-muted">Recent Projects</span>
              <div className="flex items-center gap-1 text-text-muted">
                <button onClick={() => { reloadProjects(); queryClient.invalidateQueries({ queryKey: ["workbench"] }); }} className="rounded-md p-1 hover:bg-slate-800 hover:text-text-main" aria-label="Refresh"><RotateCw size={13} /></button>
                <button onClick={() => setActiveView("projects")} className="rounded-md p-1 hover:bg-slate-800 hover:text-text-main" aria-label="Add project"><Plus size={13} /></button>
              </div>
            </div>
            <ul className="mt-2 grid gap-1">
              {recentProjects.length === 0 ? (
                <li className="rounded-lg px-2 py-2 text-xs text-text-muted">No recent projects yet.</li>
              ) : (
                recentProjects.slice(0, 4).map((item) => {
                  const active = item.rootPath === project.rootPath;
                  return (
                    <li key={item.rootPath}>
                      <button onClick={() => switchProject(item)} className={`flex w-full min-w-0 items-center gap-2 rounded-lg px-2 py-1.5 text-left transition ${active ? "bg-slate-800/80" : "hover:bg-slate-800/60"}`}>
                        <span className="grid h-7 w-7 shrink-0 place-items-center rounded-md border border-slate-700 bg-slate-950 text-text-muted"><FolderGit2 size={14} /></span>
                        <span className="min-w-0">
                          <span className="block truncate text-sm text-text-main">{item.name}</span>
                          <span className="block truncate text-[11px] text-text-muted">{item.rootPath}</span>
                        </span>
                      </button>
                    </li>
                  );
                })
              )}
            </ul>
            <button onClick={() => setActiveView("projects")} className="mt-2 inline-flex items-center gap-1 text-xs font-medium text-purple-hover hover:text-purple-primary">
              View all projects →
            </button>
          </div>

          <div className="mt-auto pt-3">
            <div className="flex items-center gap-2 rounded-xl border border-slate-800 bg-slate-900/80 p-3">
              <span className="grid h-9 w-9 shrink-0 place-items-center rounded-full bg-purple-primary/20 text-purple-hover"><Bot size={18} /></span>
              <span className="min-w-0 flex-1">
                <span className="block truncate text-sm font-medium">Thanos Dev</span>
                <span className="block truncate text-xs text-text-muted">Local-first workspace</span>
              </span>
              <button className="rounded-md p-1 text-text-muted hover:bg-slate-800 hover:text-text-main" aria-label="More"><MoreHorizontal size={16} /></button>
            </div>
          </div>
        </div>
      </aside>

      <main className="grid min-h-0 grid-rows-[auto_minmax(0,1fr)]">
        <header className="flex h-16 min-w-0 flex-wrap items-center justify-between gap-3 border-b border-slate-800 bg-slate-950/60 px-4 backdrop-blur-xl">
          <div className="flex min-w-0 flex-wrap items-center gap-2">
            <Dropdown
              width="w-72"
              trigger={
                <span className="inline-flex max-w-full items-center gap-2 rounded-lg border border-slate-800 bg-slate-900/80 px-3 py-2 text-sm hover:border-slate-700">
                  <Boxes size={16} className="shrink-0" />
                  <span className="truncate">{project.name || "Thanos Cli"}</span>
                  <ChevronDown size={14} className="shrink-0 text-text-muted" />
                </span>
              }
            >
              {(close) => (
                <div className="grid gap-1">
                  <p className="px-2 py-1 text-[11px] uppercase tracking-wide text-text-muted">Current project</p>
                  <div className="flex items-center gap-2 rounded-md bg-slate-800/60 px-2 py-2 text-sm">
                    <Check size={14} className="shrink-0 text-green-success" />
                    <span className="min-w-0">
                      <span className="block truncate font-medium">{project.name || "No project"}</span>
                      <span className="block truncate text-xs text-text-muted">{project.rootPath || "—"}</span>
                    </span>
                  </div>
                  {switchable.length > 0 && (
                    <>
                      <p className="px-2 pt-2 text-[11px] uppercase tracking-wide text-text-muted">Switch to</p>
                      {switchable.map((item) => (
                        <button key={item.id} onClick={() => { switchProject(item); close(); }} className="flex min-w-0 items-center gap-2 rounded-md px-2 py-2 text-left text-sm hover:bg-slate-800">
                          <Boxes size={14} className="shrink-0 text-text-muted" />
                          <span className="min-w-0">
                            <span className="block truncate font-medium">{item.name}</span>
                            <span className="block truncate text-xs text-text-muted">{item.rootPath}</span>
                          </span>
                        </button>
                      ))}
                    </>
                  )}
                  <button onClick={() => { setActiveView("projects"); close(); }} className="mt-1 flex items-center gap-2 rounded-md border-t border-slate-800 px-2 py-2 text-left text-sm text-text-muted hover:bg-slate-800 hover:text-text-main">
                    <FolderKanban size={14} />
                    Manage projects…
                  </button>
                </div>
              )}
            </Dropdown>
            <Dropdown
              width="w-72"
              trigger={
                <span className="inline-flex min-w-0 max-w-full items-center gap-2 rounded-lg border border-slate-800 bg-slate-900/80 px-3 py-2 text-sm text-blue-info hover:border-slate-700">
                  <GitBranch size={16} className="shrink-0" />
                  <span className="truncate">{branch} · {worktreeRoot}</span>
                  <ChevronDown size={14} className="shrink-0 text-text-muted" />
                </span>
              }
            >
              {(close) => <BranchMenu current={branch} worktreeRoot={worktreeRoot} suggestions={branchSuggestions} onSelect={(value) => { changeBranch(value); close(); }} />}
            </Dropdown>
          </div>

          <div className="hidden items-center gap-2 text-sm text-text-muted md:flex">
            <CircleDot size={14} className="text-green-success" />
            {installed} Agents Online
          </div>

          <div className="flex min-w-0 items-center gap-2 overflow-x-auto">
            <button className="inline-flex items-center gap-2 rounded-lg border border-slate-800 bg-slate-900/80 px-3 py-2 text-sm text-text-muted hover:border-slate-700">
              <Search size={16} />
              <span className="hidden sm:inline">Search</span>
            </button>
            <button className="rounded-lg border border-slate-800 bg-slate-900/80 p-2 text-text-muted hover:border-slate-700">
              <Bell size={16} />
            </button>
            <div className="inline-flex items-center overflow-hidden rounded-lg bg-gradient-to-r from-purple-primary to-purple-hover text-white shadow-lg shadow-purple-primary/20">
              <button onClick={() => { setActiveView("workbench"); openCreateTask(); }} className="inline-flex items-center gap-2 px-3 py-2 text-sm font-medium hover:bg-white/10">
                <Plus size={16} />
                <span className="hidden sm:inline">New Task</span>
              </button>
              <span className="h-6 w-px bg-white/20" />
              <button onClick={() => { setActiveView("workbench"); openCreateTask(); }} className="px-1.5 py-2 hover:bg-white/10" aria-label="New task options">
                <ChevronDown size={14} />
              </button>
            </div>
          </div>
        </header>
        {children}
      </main>
    </div>
  );
}

function BranchMenu({ current, worktreeRoot, suggestions, onSelect }: { current: string; worktreeRoot: string; suggestions: string[]; onSelect: (value: string) => void }) {
  const [value, setValue] = useState(current);
  return (
    <div className="grid gap-1">
      <p className="px-2 py-1 text-[11px] uppercase tracking-wide text-text-muted">Base branch</p>
      <form onSubmit={(event) => { event.preventDefault(); onSelect(value); }} className="flex items-center gap-1 px-1">
        <input value={value} autoFocus onChange={(event) => setValue(event.target.value)} placeholder="branch name" className="h-8 min-w-0 flex-1 rounded-md border border-slate-800 bg-slate-950 px-2 text-sm text-text-main placeholder:text-text-muted" />
        <button type="submit" className="shrink-0 rounded-md bg-purple-primary px-2 py-1.5 text-xs font-medium text-white hover:bg-purple-hover">Set</button>
      </form>
      <div className="mt-1 grid gap-0.5">
        {suggestions.map((item) => (
          <button key={item} onClick={() => onSelect(item)} className="flex items-center justify-between rounded-md px-2 py-1.5 text-left text-sm hover:bg-slate-800">
            <span className="inline-flex items-center gap-2"><GitBranch size={13} className="text-text-muted" />{item}</span>
            {item === current && <Check size={13} className="text-green-success" />}
          </button>
        ))}
      </div>
      <p className="mt-1 flex items-center gap-2 border-t border-slate-800 px-2 pt-2 text-xs text-text-muted">
        <Settings size={12} /> Worktrees · {worktreeRoot}
      </p>
    </div>
  );
}
