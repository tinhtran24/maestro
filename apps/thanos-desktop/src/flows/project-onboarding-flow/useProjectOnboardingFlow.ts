import { useState } from "react";
import { useProjects } from "../../features/projects/hooks/useProjects";
import { NativeBackend } from "../../services/nativeBackend";
import { useWorkbenchStore } from "../../state/workbenchStore";

const backend = new NativeBackend();

export type ProjectDraft = {
  name: string;
  rootPath: string;
  gitRemoteUrl: string;
  defaultBranch: string;
  worktreeRoot: string;
  packageManager: string;
  devCommand: string;
  testCommand: string;
};

export function useProjectOnboardingFlow(onLoaded?: () => void) {
  const setProject = useWorkbenchStore((state) => state.setProject);
  const setLoadError = useWorkbenchStore((state) => state.setLoadError);
  const setActiveView = useWorkbenchStore((state) => state.setActiveView);
  const saveProjectRecord = useProjects((state) => state.saveProject);
  const setActiveProject = useProjects((state) => state.setActive);
  const projectRecords = useProjects((state) => state.projects);
  const [recentProjects, setRecentProjects] = useState(() => backend.getRecentProjects());
  const [draft, setDraft] = useState<ProjectDraft>({
    name: "",
    rootPath: backend.getWorkspacePath(),
    gitRemoteUrl: "",
    defaultBranch: "main",
    worktreeRoot: ".thanos/worktrees",
    packageManager: "npm",
    devCommand: "npm run dev",
    testCommand: "npm test",
  });

  const update = (patch: Partial<ProjectDraft>) => setDraft((current) => ({ ...current, ...patch }));

  async function selectFolder() {
    const path = await backend.selectWorkspaceFolder();
    if (path) {
      const parts = path.split("/").filter(Boolean);
      update({ rootPath: path, name: draft.name || parts[parts.length - 1] || "New Project" });
    }
    return path;
  }

  async function save(rootPathOverride?: string) {
    const rootPath = rootPathOverride || draft.rootPath;
    const parts = rootPath.split("/").filter(Boolean);
    const project = await backend.createOrImportProject({
      root_path: rootPath,
      name: draft.name || parts[parts.length - 1] || "New Project",
      git_remote_url: draft.gitRemoteUrl,
      default_branch: draft.defaultBranch,
      worktree_root: draft.worktreeRoot,
      package_manager: draft.packageManager,
      dev_command: draft.devCommand,
      test_command: draft.testCommand,
    });
    if (!project) {
      setLoadError("Project setup failed. Check that the folder exists and is writable.");
      return null;
    }
    const record = saveProjectRecord(project);
    setActiveProject(record.id);
    setProject(record);
    setRecentProjects(backend.getRecentProjects());
    setLoadError("");
    setActiveView("workbench");
    onLoaded?.();
    return record;
  }

  async function openRecent(rootPath?: string) {
    const selected = rootPath || backend.getWorkspacePath();
    if (!selected) return null;
    backend.setWorkspacePath(selected);
    onLoaded?.();
    return selected;
  }

  const mergedRecents = projectRecords.length ? projectRecords.filter((item) => !item.archived) : recentProjects;

  return { draft, update, selectFolder, save, openRecent, recentProjects: mergedRecents };
}
