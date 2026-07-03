import { Archive, ArchiveRestore, Pencil, Play, X } from "lucide-react";
import type { Project } from "../../domain/models";

export function ProjectDetails({
  project,
  active,
  onOpen,
  onEdit,
  onArchive,
  onClose,
}: {
  project: Project;
  active: boolean;
  onOpen: () => void;
  onEdit: () => void;
  onArchive: (archived: boolean) => void;
  onClose: () => void;
}) {
  const envEntries = Object.entries(project.envVars ?? {});
  const rows: Array<{ label: string; value: string }> = [
    { label: "Local folder", value: project.rootPath || "—" },
    { label: "Git remote", value: project.gitRemoteUrl || "Not linked" },
    { label: "Default branch", value: project.defaultBranch || "main" },
    { label: "Worktree root", value: project.worktreeRoot || ".thanos/worktrees" },
    { label: "Package manager", value: project.packageManager || "npm" },
    { label: "Dev command", value: project.devCommand || "—" },
    { label: "Test command", value: project.testCommand || "—" },
    { label: "Created", value: project.createdAt ? new Date(project.createdAt).toLocaleString() : "—" },
    { label: "Updated", value: project.updatedAt ? new Date(project.updatedAt).toLocaleString() : "—" },
    { label: "Last opened", value: project.lastOpenedAt ? new Date(project.lastOpenedAt).toLocaleString() : "Never" },
  ];

  return (
    <aside className="flex h-full min-h-0 flex-col rounded-xl border border-slate-800 bg-slate-900/70">
      <header className="flex items-start justify-between gap-2 border-b border-slate-800 p-4">
        <div className="min-w-0">
          <h2 className="truncate text-base font-semibold text-text-main">{project.name}</h2>
          <p className="mt-0.5 text-xs text-text-muted">{active ? "Active project" : project.archived ? "Archived" : "Project details"}</p>
        </div>
        <button onClick={onClose} className="rounded-md p-1 text-text-muted hover:bg-slate-800 hover:text-text-main" aria-label="Close details"><X size={16} /></button>
      </header>

      <div className="min-h-0 flex-1 overflow-y-auto p-4">
        <dl className="grid gap-2.5">
          {rows.map((row) => (
            <div key={row.label} className="grid gap-0.5">
              <dt className="text-xs text-text-muted">{row.label}</dt>
              <dd className="truncate text-sm text-text-main">{row.value}</dd>
            </div>
          ))}
        </dl>

        <div className="mt-4">
          <p className="mb-2 text-xs font-semibold uppercase tracking-wide text-text-muted">Environment variables</p>
          {envEntries.length === 0 ? (
            <p className="rounded-lg border border-dashed border-slate-800 bg-slate-950/40 px-3 py-2 text-xs text-text-muted">None.</p>
          ) : (
            <ul className="grid gap-1">
              {envEntries.map(([key, value]) => (
                <li key={key} className="flex items-center justify-between gap-2 rounded-lg border border-slate-800 bg-slate-950/60 px-3 py-1.5 font-mono text-xs">
                  <span className="truncate text-text-main">{key}</span>
                  <span className="truncate text-text-muted">{value}</span>
                </li>
              ))}
            </ul>
          )}
        </div>
      </div>

      <footer className="flex items-center gap-2 border-t border-slate-800 p-4">
        <button onClick={onOpen} className="inline-flex flex-1 items-center justify-center gap-2 rounded-lg bg-purple-primary px-3 py-2 text-sm font-medium text-white hover:bg-purple-hover">
          <Play size={14} /> Open
        </button>
        <button onClick={onEdit} className="inline-flex items-center gap-2 rounded-lg border border-slate-800 px-3 py-2 text-sm text-text-muted hover:text-text-main">
          <Pencil size={14} /> Edit
        </button>
        <button onClick={() => onArchive(!project.archived)} className="inline-flex items-center gap-2 rounded-lg border border-slate-800 px-3 py-2 text-sm text-text-muted hover:text-text-main">
          {project.archived ? <ArchiveRestore size={14} /> : <Archive size={14} />}
          {project.archived ? "Unarchive" : "Archive"}
        </button>
      </footer>
    </aside>
  );
}
