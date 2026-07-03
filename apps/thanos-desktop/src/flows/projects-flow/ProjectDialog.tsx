import { FolderOpen, Plus, Save, Trash2, X } from "lucide-react";
import { useState } from "react";
import type { Project } from "../../domain/models";
import { Dialog } from "../../shared/ui/Dialog";
import type { ProjectFormValues } from "./useProjectsFlow";

function toValues(project?: Project): ProjectFormValues {
  return {
    name: project?.name ?? "",
    rootPath: project?.rootPath ?? "",
    gitRemoteUrl: project?.gitRemoteUrl ?? "",
    defaultBranch: project?.defaultBranch ?? "main",
    worktreeRoot: project?.worktreeRoot ?? ".thanos/worktrees",
    packageManager: project?.packageManager ?? "npm",
    devCommand: project?.devCommand ?? "npm run dev",
    testCommand: project?.testCommand ?? "npm test",
    envVars: project?.envVars ?? {},
  };
}

export function ProjectDialog({
  open,
  mode,
  initial,
  onClose,
  onSubmit,
  onPickFolder,
}: {
  open: boolean;
  mode: "create" | "edit";
  initial?: Project;
  onClose: () => void;
  onSubmit: (values: ProjectFormValues) => void;
  onPickFolder: () => Promise<string | null>;
}) {
  const [values, setValues] = useState<ProjectFormValues>(() => toValues(initial));
  const [envRows, setEnvRows] = useState<Array<{ key: string; value: string }>>(() =>
    Object.entries(initial?.envVars ?? {}).map(([key, value]) => ({ key, value })),
  );
  const update = (patch: Partial<ProjectFormValues>) => setValues((current) => ({ ...current, ...patch }));

  const canSave = values.name.trim().length > 0 && values.rootPath.trim().length > 0;

  async function pick() {
    const path = await onPickFolder();
    if (path) {
      const parts = path.split("/").filter(Boolean);
      update({ rootPath: path, name: values.name || parts[parts.length - 1] || "New Project" });
    }
  }

  function submit() {
    if (!canSave) return;
    const envVars: Record<string, string> = {};
    for (const row of envRows) {
      const key = row.key.trim();
      if (key) envVars[key] = row.value;
    }
    onSubmit({ ...values, envVars });
  }

  return (
    <Dialog
      open={open}
      onClose={onClose}
      title={mode === "edit" ? "Edit project" : "New project"}
      description={mode === "edit" ? "Update project settings. Stored locally." : "Create a project. Mock only — no Git operations run."}
      footer={
        <>
          <button onClick={onClose} className="inline-flex items-center gap-2 rounded-lg border border-slate-800 px-3 py-2 text-sm text-text-muted hover:text-text-main">
            <X size={15} /> Cancel
          </button>
          <button onClick={submit} disabled={!canSave} className="inline-flex items-center gap-2 rounded-lg bg-purple-primary px-3 py-2 text-sm font-medium text-white enabled:hover:bg-purple-hover disabled:opacity-50">
            <Save size={15} /> {mode === "edit" ? "Save changes" : "Create project"}
          </button>
        </>
      }
    >
      <div className="grid gap-4">
        <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
          <Field label="Project name" value={values.name} onChange={(name) => update({ name })} placeholder="My Project" />
          <FolderField label="Local folder" value={values.rootPath} onChange={(rootPath) => update({ rootPath })} onPick={pick} />
          <Field label="Git remote URL" value={values.gitRemoteUrl} onChange={(gitRemoteUrl) => update({ gitRemoteUrl })} placeholder="git@github.com:org/repo.git" />
          <Field label="Default branch" value={values.defaultBranch} onChange={(defaultBranch) => update({ defaultBranch })} />
          <Field label="Worktree root" value={values.worktreeRoot} onChange={(worktreeRoot) => update({ worktreeRoot })} />
          <Field label="Package manager" value={values.packageManager} onChange={(packageManager) => update({ packageManager })} />
          <Field label="Dev command" value={values.devCommand} onChange={(devCommand) => update({ devCommand })} />
          <Field label="Test command" value={values.testCommand} onChange={(testCommand) => update({ testCommand })} />
        </div>

        <div>
          <div className="mb-2 flex items-center justify-between">
            <span className="text-sm font-medium text-text-main">Environment variables</span>
            <button onClick={() => setEnvRows((rows) => [...rows, { key: "", value: "" }])} className="inline-flex items-center gap-1 rounded-md border border-slate-800 px-2 py-1 text-xs text-text-muted hover:text-text-main">
              <Plus size={13} /> Add
            </button>
          </div>
          {envRows.length === 0 ? (
            <p className="rounded-lg border border-dashed border-slate-800 bg-slate-950/40 px-3 py-2 text-xs text-text-muted">No environment variables.</p>
          ) : (
            <div className="grid gap-2">
              {envRows.map((row, index) => (
                <div key={index} className="flex items-center gap-2">
                  <input
                    value={row.key}
                    onChange={(event) => setEnvRows((rows) => rows.map((r, i) => (i === index ? { ...r, key: event.target.value } : r)))}
                    placeholder="KEY"
                    className="h-9 min-w-0 flex-1 rounded-lg border border-slate-800 bg-slate-950/60 px-3 font-mono text-sm text-text-main placeholder:text-text-muted"
                  />
                  <input
                    value={row.value}
                    onChange={(event) => setEnvRows((rows) => rows.map((r, i) => (i === index ? { ...r, value: event.target.value } : r)))}
                    placeholder="value"
                    className="h-9 min-w-0 flex-[2] rounded-lg border border-slate-800 bg-slate-950/60 px-3 font-mono text-sm text-text-main placeholder:text-text-muted"
                  />
                  <button onClick={() => setEnvRows((rows) => rows.filter((_, i) => i !== index))} className="shrink-0 rounded-md border border-slate-800 p-2 text-text-muted hover:border-red-danger/40 hover:text-red-danger" aria-label="Remove variable">
                    <Trash2 size={14} />
                  </button>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </Dialog>
  );
}

function Field({ label, value, onChange, placeholder }: { label: string; value: string; onChange: (value: string) => void; placeholder?: string }) {
  return (
    <label className="grid gap-1 text-sm">
      <span className="text-text-muted">{label}</span>
      <input value={value} placeholder={placeholder} onChange={(event) => onChange(event.target.value)} className="h-9 rounded-lg border border-slate-800 bg-slate-950/60 px-3 text-text-main placeholder:text-text-muted" />
    </label>
  );
}

function FolderField({ label, value, onChange, onPick }: { label: string; value: string; onChange: (value: string) => void; onPick: () => void }) {
  return (
    <label className="grid gap-1 text-sm">
      <span className="text-text-muted">{label}</span>
      <div className="flex items-center gap-2">
        <input value={value} onChange={(event) => onChange(event.target.value)} placeholder="/path/to/project" className="h-9 min-w-0 flex-1 rounded-lg border border-slate-800 bg-slate-950/60 px-3 text-text-main placeholder:text-text-muted" />
        <button onClick={onPick} className="inline-flex h-9 shrink-0 items-center gap-2 rounded-lg border border-slate-800 px-3 text-sm text-text-muted hover:text-text-main" type="button">
          <FolderOpen size={15} /> Browse
        </button>
      </div>
    </label>
  );
}
