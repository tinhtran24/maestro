// Phase 2 — Project Management.
// Flow controller: owns dialog/selection state and bridges project actions into
// the workbench (active project + workspace path). No Git operations.

import { useQueryClient } from "@tanstack/react-query";
import { useMemo, useState } from "react";
import type { Project } from "../../domain/models";
import { useProjects } from "../../features/projects/hooks/useProjects";
import { NativeBackend } from "../../services/nativeBackend";
import { useWorkbenchStore } from "../../state/workbenchStore";

const backend = new NativeBackend();

export type ProjectFormValues = {
  name: string;
  rootPath: string;
  gitRemoteUrl: string;
  defaultBranch: string;
  worktreeRoot: string;
  packageManager: string;
  devCommand: string;
  testCommand: string;
  envVars: Record<string, string>;
};

type DialogState = { kind: "create" | "edit" | "import"; projectId?: string } | null;

export function useProjectsFlow() {
  const projects = useProjects((state) => state.projects);
  const activeProjectId = useProjects((state) => state.activeProjectId);
  const createProject = useProjects((state) => state.createProject);
  const saveProject = useProjects((state) => state.saveProject);
  const archive = useProjects((state) => state.archive);
  const setActive = useProjects((state) => state.setActive);

  const setProject = useWorkbenchStore((state) => state.setProject);
  const setActiveView = useWorkbenchStore((state) => state.setActiveView);
  const queryClient = useQueryClient();

  const [showArchived, setShowArchived] = useState(false);
  const [dialog, setDialog] = useState<DialogState>(null);
  const [detailsId, setDetailsId] = useState<string>("");

  const visible = useMemo(
    () => projects.filter((project) => (showArchived ? true : !project.archived)),
    [projects, showArchived],
  );
  const editingProject = dialog?.projectId ? projects.find((p) => p.id === dialog.projectId) : undefined;
  const detailsProject = detailsId ? projects.find((p) => p.id === detailsId) : undefined;

  function openProject(project: Project) {
    setActive(project.id);
    setProject(project);
    backend.setWorkspacePath(project.rootPath);
    queryClient.invalidateQueries({ queryKey: ["workbench"] });
    setActiveView("workbench");
  }

  function submitForm(values: ProjectFormValues) {
    if (dialog?.kind === "edit" && editingProject) {
      saveProject({ ...editingProject, ...values });
    } else {
      const created = createProject(values);
      setDetailsId(created.id);
    }
    setDialog(null);
  }

  async function pickFolder(): Promise<string | null> {
    return backend.selectWorkspaceFolder();
  }

  return {
    projects: visible,
    activeProjectId,
    showArchived,
    setShowArchived,
    dialog,
    editingProject,
    detailsProject,
    openCreate: () => setDialog({ kind: "create" }),
    openImport: () => setDialog({ kind: "import" }),
    openEdit: (id: string) => setDialog({ kind: "edit", projectId: id }),
    closeDialog: () => setDialog(null),
    selectDetails: (id: string) => setDetailsId(id),
    clearDetails: () => setDetailsId(""),
    submitForm,
    pickFolder,
    archive,
    openProject,
  };
}
