import { FolderGit2, FolderOpen, Plus } from "lucide-react";
import { ImportRepoFlow } from "./ImportRepoFlow";
import { NewProjectFlow } from "./NewProjectFlow";
import { useState } from "react";
import { useProjectOnboardingFlow } from "./useProjectOnboardingFlow";

export function ProjectOnboardingFlow({ onLoaded }: { onLoaded?: () => void }) {
  const [mode, setMode] = useState<"intro" | "import" | "new">("intro");
  const flow = useProjectOnboardingFlow(onLoaded);
  if (mode === "import") return <ImportRepoFlow onBack={() => setMode("intro")} onLoaded={onLoaded} />;
  if (mode === "new") return <NewProjectFlow onBack={() => setMode("intro")} onLoaded={onLoaded} />;
  return (
    <section className="grid h-full place-items-center overflow-y-auto bg-bg-app p-6">
      <div className="w-full max-w-3xl rounded-lg border border-slate-800 bg-bg-card p-6">
        <h1 className="text-xl font-semibold">{flow.recentProjects.length ? "Projects" : "No project loaded"}</h1>
        <p className="mt-2 text-sm text-text-muted">Create a new project or import an existing repository.</p>
        <div className="mt-6 grid grid-cols-3 gap-3">
          <Action icon={FolderGit2} label="Import Repo" onClick={() => setMode("import")} />
          <Action icon={Plus} label="New Project" onClick={() => setMode("new")} />
          <Action icon={FolderOpen} label="Open Recent" onClick={() => flow.openRecent()} />
        </div>
        <section className="mt-6">
          <div className="flex items-center justify-between">
            <h2 className="text-sm font-semibold">Projects</h2>
            <span className="text-xs text-text-muted">{flow.recentProjects.length}</span>
          </div>
          <div className="mt-2 grid gap-2">
            {flow.recentProjects.length ? flow.recentProjects.map((project) => (
              <button key={project.rootPath} onClick={() => flow.openRecent(project.rootPath)} className="flex items-center justify-between rounded-lg border border-slate-800 bg-slate-950/50 p-3 text-left hover:border-blue-info/60 hover:bg-slate-900">
                <span>
                  <span className="block text-sm font-medium">{project.name}</span>
                  <span className="block text-xs text-text-muted">{project.rootPath}</span>
                </span>
                <span className="text-xs text-blue-info">{project.defaultBranch || "main"}</span>
              </button>
            )) : (
              <div className="rounded-lg border border-slate-800 bg-slate-950/40 p-3 text-sm text-text-muted">No projects added yet.</div>
            )}
          </div>
        </section>
      </div>
    </section>
  );
}

function Action({ icon: Icon, label, onClick }: { icon: typeof Plus; label: string; onClick: () => void }) {
  return (
    <button onClick={onClick} className="grid min-h-28 place-items-center rounded-lg border border-slate-800 bg-slate-900/80 p-4 text-sm hover:border-blue-info/60 hover:bg-slate-900">
      <Icon size={24} className="text-blue-info" />
      <span>{label}</span>
    </button>
  );
}
