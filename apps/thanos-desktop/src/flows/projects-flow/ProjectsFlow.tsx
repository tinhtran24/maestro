import { FolderGit2, FolderKanban, Import, Plus } from "lucide-react";
import { ImportProjectDialog } from "./ImportProjectDialog";
import { ProjectCard } from "./ProjectCard";
import { ProjectDetails } from "./ProjectDetails";
import { ProjectDialog } from "./ProjectDialog";
import { useProjectsFlow } from "./useProjectsFlow";

export function ProjectsFlow() {
  const flow = useProjectsFlow();

  return (
    <section className="grid h-full min-h-0 grid-rows-[auto_minmax(0,1fr)] bg-bg-app">
      <header className="flex flex-wrap items-end justify-between gap-3 border-b border-slate-800 px-6 py-4">
        <div>
          <h1 className="text-xl font-semibold">Projects</h1>
          <p className="mt-0.5 text-sm text-text-muted">Every task belongs to a project. Create, import, and switch between local projects.</p>
        </div>
        <div className="flex items-center gap-2">
          <label className="inline-flex cursor-pointer items-center gap-2 rounded-lg border border-slate-800 bg-slate-900/70 px-3 py-2 text-sm text-text-muted">
            <input type="checkbox" checked={flow.showArchived} onChange={(event) => flow.setShowArchived(event.target.checked)} className="accent-purple-primary" />
            Show archived
          </label>
          <button onClick={flow.openImport} className="inline-flex items-center gap-2 rounded-lg border border-slate-800 bg-slate-900/70 px-3 py-2 text-sm hover:border-slate-700">
            <Import size={16} /> Import Repo
          </button>
          <button onClick={flow.openCreate} className="inline-flex items-center gap-2 rounded-lg bg-purple-primary px-3 py-2 text-sm font-medium text-white hover:bg-purple-hover">
            <Plus size={16} /> New Project
          </button>
        </div>
      </header>

      <div className="grid min-h-0 grid-cols-1 gap-4 overflow-hidden p-6 xl:grid-cols-[minmax(0,1fr)_360px]">
        <div className="min-h-0 overflow-y-auto">
          {flow.projects.length === 0 ? (
            <EmptyProjects onCreate={flow.openCreate} onImport={flow.openImport} />
          ) : (
            <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 2xl:grid-cols-3">
              {flow.projects.map((project) => (
                <ProjectCard
                  key={project.id}
                  project={project}
                  active={project.id === flow.activeProjectId}
                  selected={project.id === flow.detailsProject?.id}
                  onOpen={() => flow.openProject(project)}
                  onEdit={() => flow.openEdit(project.id)}
                  onArchive={(archived) => flow.archive(project.id, archived)}
                  onSelect={() => flow.selectDetails(project.id)}
                />
              ))}
            </div>
          )}
        </div>

        <div className="hidden min-h-0 xl:block">
          {flow.detailsProject ? (
            <ProjectDetails
              project={flow.detailsProject}
              active={flow.detailsProject.id === flow.activeProjectId}
              onOpen={() => flow.openProject(flow.detailsProject!)}
              onEdit={() => flow.openEdit(flow.detailsProject!.id)}
              onArchive={(archived) => flow.archive(flow.detailsProject!.id, archived)}
              onClose={flow.clearDetails}
            />
          ) : (
            <div className="grid h-full place-items-center rounded-xl border border-dashed border-slate-800 bg-slate-950/40 p-6 text-center text-sm text-text-muted">
              <div className="flex flex-col items-center gap-2">
                <FolderKanban size={22} />
                Select a project to view its settings.
              </div>
            </div>
          )}
        </div>
      </div>

      {(flow.dialog?.kind === "create" || flow.dialog?.kind === "edit") && (
        <ProjectDialog
          open
          mode={flow.dialog.kind}
          initial={flow.editingProject}
          onClose={flow.closeDialog}
          onSubmit={flow.submitForm}
          onPickFolder={flow.pickFolder}
        />
      )}
      {flow.dialog?.kind === "import" && (
        <ImportProjectDialog open onClose={flow.closeDialog} onSubmit={flow.submitForm} onPickFolder={flow.pickFolder} />
      )}
    </section>
  );
}

function EmptyProjects({ onCreate, onImport }: { onCreate: () => void; onImport: () => void }) {
  return (
    <div className="grid h-full place-items-center">
      <div className="w-full max-w-md rounded-xl border border-dashed border-slate-700 bg-slate-950/40 p-8 text-center">
        <span className="mx-auto grid h-12 w-12 place-items-center rounded-full border border-slate-800 bg-slate-900 text-text-muted"><FolderKanban size={22} /></span>
        <p className="mt-3 text-base font-semibold text-text-main">No project loaded</p>
        <p className="mt-1 text-sm text-text-muted">Create a new project or import an existing repository.</p>
        <div className="mt-5 flex items-center justify-center gap-2">
          <button onClick={onImport} className="inline-flex items-center gap-2 rounded-lg border border-slate-800 px-3 py-2 text-sm hover:border-slate-700">
            <FolderGit2 size={16} /> Import Repo
          </button>
          <button onClick={onCreate} className="inline-flex items-center gap-2 rounded-lg bg-purple-primary px-3 py-2 text-sm font-medium text-white hover:bg-purple-hover">
            <Plus size={16} /> New Project
          </button>
        </div>
      </div>
    </div>
  );
}
