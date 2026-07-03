import { FolderOpen, GitBranch, Import, X } from "lucide-react";
import { useState } from "react";
import { Dialog } from "../../shared/ui/Dialog";
import type { ProjectFormValues } from "./useProjectsFlow";

// Phase 2: import is metadata-only (mock). No clone / no Git operations run.
export function ImportProjectDialog({
  open,
  onClose,
  onSubmit,
  onPickFolder,
}: {
  open: boolean;
  onClose: () => void;
  onSubmit: (values: ProjectFormValues) => void;
  onPickFolder: () => Promise<string | null>;
}) {
  const [name, setName] = useState("");
  const [rootPath, setRootPath] = useState("");
  const [gitRemoteUrl, setGitRemoteUrl] = useState("");
  const [defaultBranch, setDefaultBranch] = useState("main");

  const canImport = name.trim().length > 0 && (rootPath.trim().length > 0 || gitRemoteUrl.trim().length > 0);

  async function pick() {
    const path = await onPickFolder();
    if (path) {
      const parts = path.split("/").filter(Boolean);
      setRootPath(path);
      if (!name) setName(parts[parts.length - 1] || "Imported Project");
    }
  }

  function deriveNameFromUrl(url: string) {
    const cleaned = url.replace(/\.git$/, "").split(/[\/:]/).filter(Boolean);
    return cleaned[cleaned.length - 1] || "";
  }

  function submit() {
    if (!canImport) return;
    onSubmit({
      name: name.trim(),
      rootPath: rootPath.trim(),
      gitRemoteUrl: gitRemoteUrl.trim(),
      defaultBranch: defaultBranch.trim() || "main",
      worktreeRoot: ".thanos/worktrees",
      packageManager: "npm",
      devCommand: "npm run dev",
      testCommand: "npm test",
      envVars: {},
    });
  }

  return (
    <Dialog
      open={open}
      onClose={onClose}
      title="Import existing repository"
      description="Register an existing repository as a project. Mock only — no clone or Git operations run."
      size="md"
      footer={
        <>
          <button onClick={onClose} className="inline-flex items-center gap-2 rounded-lg border border-slate-800 px-3 py-2 text-sm text-text-muted hover:text-text-main">
            <X size={15} /> Cancel
          </button>
          <button onClick={submit} disabled={!canImport} className="inline-flex items-center gap-2 rounded-lg bg-purple-primary px-3 py-2 text-sm font-medium text-white enabled:hover:bg-purple-hover disabled:opacity-50">
            <Import size={15} /> Import project
          </button>
        </>
      }
    >
      <div className="grid gap-3">
        <label className="grid gap-1 text-sm">
          <span className="text-text-muted">Local folder</span>
          <div className="flex items-center gap-2">
            <input value={rootPath} onChange={(event) => setRootPath(event.target.value)} placeholder="/path/to/repo" className="h-9 min-w-0 flex-1 rounded-lg border border-slate-800 bg-slate-950/60 px-3 text-text-main placeholder:text-text-muted" />
            <button onClick={pick} type="button" className="inline-flex h-9 shrink-0 items-center gap-2 rounded-lg border border-slate-800 px-3 text-sm text-text-muted hover:text-text-main">
              <FolderOpen size={15} /> Browse
            </button>
          </div>
        </label>
        <label className="grid gap-1 text-sm">
          <span className="text-text-muted">Git remote URL</span>
          <input
            value={gitRemoteUrl}
            onChange={(event) => {
              setGitRemoteUrl(event.target.value);
              if (!name) setName(deriveNameFromUrl(event.target.value));
            }}
            placeholder="git@github.com:org/repo.git"
            className="h-9 rounded-lg border border-slate-800 bg-slate-950/60 px-3 text-text-main placeholder:text-text-muted"
          />
        </label>
        <div className="grid grid-cols-2 gap-3">
          <label className="grid gap-1 text-sm">
            <span className="text-text-muted">Project name</span>
            <input value={name} onChange={(event) => setName(event.target.value)} placeholder="repo" className="h-9 rounded-lg border border-slate-800 bg-slate-950/60 px-3 text-text-main placeholder:text-text-muted" />
          </label>
          <label className="grid gap-1 text-sm">
            <span className="inline-flex items-center gap-1 text-text-muted"><GitBranch size={13} /> Default branch</span>
            <input value={defaultBranch} onChange={(event) => setDefaultBranch(event.target.value)} className="h-9 rounded-lg border border-slate-800 bg-slate-950/60 px-3 text-text-main placeholder:text-text-muted" />
          </label>
        </div>
      </div>
    </Dialog>
  );
}
