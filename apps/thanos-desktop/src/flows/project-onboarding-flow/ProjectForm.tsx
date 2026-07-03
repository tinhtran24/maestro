import type { LucideIcon } from "lucide-react";
import type { ProjectDraft } from "./useProjectOnboardingFlow";

export function ProjectForm({ draft, update }: { draft: ProjectDraft; update: (patch: Partial<ProjectDraft>) => void }) {
  return (
    <div className="grid grid-cols-2 gap-3">
      <Field label="Project name" value={draft.name} onChange={(name) => update({ name })} />
      <Field label="Local folder" value={draft.rootPath} onChange={(rootPath) => update({ rootPath })} />
      <Field label="Git remote URL" value={draft.gitRemoteUrl} onChange={(gitRemoteUrl) => update({ gitRemoteUrl })} />
      <Field label="Default branch" value={draft.defaultBranch} onChange={(defaultBranch) => update({ defaultBranch })} />
      <Field label="Worktree root" value={draft.worktreeRoot} onChange={(worktreeRoot) => update({ worktreeRoot })} />
      <Field label="Package manager" value={draft.packageManager} onChange={(packageManager) => update({ packageManager })} />
      <Field label="Dev command" value={draft.devCommand} onChange={(devCommand) => update({ devCommand })} />
      <Field label="Test command" value={draft.testCommand} onChange={(testCommand) => update({ testCommand })} />
    </div>
  );
}

export function ProjectFormActions({ primaryIcon: PrimaryIcon, primaryLabel, onPrimary, secondaryIcon: SecondaryIcon, secondaryLabel, onSecondary }: { primaryIcon: LucideIcon; primaryLabel: string; onPrimary: () => void; secondaryIcon: LucideIcon; secondaryLabel: string; onSecondary: () => void }) {
  return (
    <div className="mt-5 flex items-center justify-end gap-2">
      <button onClick={onSecondary} className="inline-flex items-center gap-2 rounded-lg border border-slate-800 px-3 py-2 text-sm">
        <SecondaryIcon size={16} />
        {secondaryLabel}
      </button>
      <button onClick={onPrimary} className="inline-flex items-center gap-2 rounded-lg bg-purple-primary px-3 py-2 text-sm font-medium text-white">
        <PrimaryIcon size={16} />
        {primaryLabel}
      </button>
    </div>
  );
}

function Field({ label, value, onChange }: { label: string; value: string; onChange: (value: string) => void }) {
  return (
    <label className="grid gap-1 text-sm">
      <span className="text-text-muted">{label}</span>
      <input value={value} onChange={(event) => onChange(event.target.value)} className="h-9 rounded-lg border border-slate-800 bg-slate-950/60 px-3 text-text-main" />
    </label>
  );
}
