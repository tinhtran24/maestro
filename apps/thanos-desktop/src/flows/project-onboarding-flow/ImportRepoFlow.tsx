import { ArrowLeft, FolderOpen, Save } from "lucide-react";
import { ProjectForm, ProjectFormActions } from "./ProjectForm";
import { useProjectOnboardingFlow } from "./useProjectOnboardingFlow";

export function ImportRepoFlow({ onBack, onLoaded }: { onBack: () => void; onLoaded?: () => void }) {
  const flow = useProjectOnboardingFlow(onLoaded);
  const importRepo = async () => {
    const selected = flow.draft.rootPath || await flow.selectFolder();
    if (selected) await flow.save(selected);
  };
  return (
    <section className="grid h-full place-items-center bg-bg-app p-6">
      <div className="w-full max-w-3xl rounded-lg border border-slate-800 bg-bg-card p-5">
        <header className="mb-5 flex items-center justify-between">
          <div>
            <h1 className="text-lg font-semibold">Import Existing Repo</h1>
            <p className="text-sm text-text-muted">Attach a local folder and store Thanos project settings.</p>
          </div>
          <button onClick={onBack} className="rounded-lg border border-slate-800 p-2"><ArrowLeft size={16} /></button>
        </header>
        <ProjectForm draft={flow.draft} update={flow.update} />
        <ProjectFormActions
          primaryIcon={Save}
          primaryLabel="Import Repo"
          onPrimary={importRepo}
          secondaryIcon={FolderOpen}
          secondaryLabel="Select local folder"
          onSecondary={flow.selectFolder}
        />
      </div>
    </section>
  );
}
