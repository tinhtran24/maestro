import { Archive, ArchiveRestore, CheckCircle2, GitBranch, Pencil, Play, Boxes } from "lucide-react";
import type { Project } from "../../domain/models";

export function ProjectCard({
  project,
  active,
  selected,
  onOpen,
  onEdit,
  onArchive,
  onSelect,
}: {
  project: Project;
  active: boolean;
  selected: boolean;
  onOpen: () => void;
  onEdit: () => void;
  onArchive: (archived: boolean) => void;
  onSelect: () => void;
}) {
  return (
    <article
      onClick={onSelect}
      className={`flex cursor-pointer flex-col rounded-xl border bg-slate-900/70 p-4 transition hover:border-slate-700 ${selected ? "border-purple-primary" : "border-slate-800"}`}
    >
      <div className="flex items-start justify-between gap-2">
        <div className="flex min-w-0 items-center gap-2">
          <span className="grid h-9 w-9 shrink-0 place-items-center rounded-lg border border-slate-700 bg-slate-950 text-text-muted"><Boxes size={18} /></span>
          <div className="min-w-0">
            <h3 className="truncate text-sm font-semibold text-text-main">{project.name}</h3>
            <p className="truncate text-xs text-text-muted">{project.rootPath || "No local folder"}</p>
          </div>
        </div>
        {active && <span className="inline-flex shrink-0 items-center gap-1 rounded-md border border-green-success/30 bg-green-success/10 px-2 py-0.5 text-[11px] text-green-success"><CheckCircle2 size={11} /> Active</span>}
        {project.archived && <span className="inline-flex shrink-0 items-center gap-1 rounded-md border border-slate-700 bg-slate-800/60 px-2 py-0.5 text-[11px] text-text-muted">Archived</span>}
      </div>

      <div className="mt-3 flex flex-wrap items-center gap-2 text-xs text-text-muted">
        <span className="inline-flex items-center gap-1 rounded-md border border-slate-800 bg-slate-950/60 px-2 py-1"><GitBranch size={12} /> {project.defaultBranch || "main"}</span>
        <span className="rounded-md border border-slate-800 bg-slate-950/60 px-2 py-1">{project.packageManager || "npm"}</span>
      </div>

      <div className="mt-4 flex items-center gap-2" onClick={(event) => event.stopPropagation()}>
        <button onClick={onOpen} className="inline-flex flex-1 items-center justify-center gap-2 rounded-lg bg-purple-primary px-3 py-2 text-xs font-medium text-white hover:bg-purple-hover">
          <Play size={13} /> Open
        </button>
        <button onClick={onEdit} className="inline-flex items-center gap-2 rounded-lg border border-slate-800 px-3 py-2 text-xs text-text-muted hover:text-text-main" aria-label="Edit project">
          <Pencil size={13} />
        </button>
        <button onClick={() => onArchive(!project.archived)} className="inline-flex items-center gap-2 rounded-lg border border-slate-800 px-3 py-2 text-xs text-text-muted hover:text-text-main" aria-label={project.archived ? "Unarchive project" : "Archive project"}>
          {project.archived ? <ArchiveRestore size={13} /> : <Archive size={13} />}
        </button>
      </div>
    </article>
  );
}
