// Phase 2 — Project Management.
// Reactive store over the local projects persistence layer. UI reads from here;
// every mutation writes through to localStorage via projectsStore.

import { create } from "zustand";
import type { Project } from "../../../domain/models";
import {
  archiveProject as persistArchive,
  getActiveProjectId,
  listProjects,
  makeProjectId,
  recentProjects,
  setActiveProjectId,
  touchProject,
  upsertProject,
} from "../api/projectsStore";

type ProjectsState = {
  projects: Project[];
  activeProjectId: string;
  reload(): void;
  saveProject(project: Project): Project;
  createProject(input: Partial<Project> & { name: string; rootPath: string }): Project;
  archive(id: string, archived: boolean): void;
  setActive(id: string): void;
};

function normalize(input: Partial<Project> & { name: string; rootPath: string }): Project {
  return {
    id: input.id || makeProjectId(),
    name: input.name,
    rootPath: input.rootPath,
    gitRemoteUrl: input.gitRemoteUrl || "",
    defaultBranch: input.defaultBranch || "main",
    worktreeRoot: input.worktreeRoot || ".thanos/worktrees",
    packageManager: input.packageManager || "npm",
    devCommand: input.devCommand || "",
    testCommand: input.testCommand || "",
    envVars: input.envVars || {},
    archived: input.archived ?? false,
    repos: input.repos ?? [],
    settings: input.settings ?? {},
    createdAt: input.createdAt,
    updatedAt: input.updatedAt,
    lastOpenedAt: input.lastOpenedAt,
  };
}

export const useProjects = create<ProjectsState>((set) => ({
  projects: listProjects(),
  activeProjectId: getActiveProjectId(),
  reload: () => set({ projects: listProjects(), activeProjectId: getActiveProjectId() }),
  saveProject: (project) => {
    const projects = upsertProject(project);
    set({ projects });
    return projects.find((item) => item.id === project.id) ?? project;
  },
  createProject: (input) => {
    const project = normalize(input);
    const projects = upsertProject(project);
    set({ projects });
    return projects.find((item) => item.id === project.id) ?? project;
  },
  archive: (id, archived) => set({ projects: persistArchive(id, archived) }),
  setActive: (id) => {
    setActiveProjectId(id);
    const projects = touchProject(id);
    set({ projects, activeProjectId: id });
  },
}));

export function useRecentProjects(limit = 5) {
  // Re-derive from the reactive list so components update on any change.
  const projects = useProjects((state) => state.projects);
  void projects;
  return recentProjects(limit);
}
