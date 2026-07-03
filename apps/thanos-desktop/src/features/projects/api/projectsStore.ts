// Phase 2 — Project Management.
// Local-only project persistence. Mock data only: this layer never performs
// Git or filesystem operations. Projects are stored in localStorage so they
// survive reloads and support multiple projects + switching.

import type { Project } from "../../../domain/models";

const PROJECTS_KEY = "thanos.projects.v1";
const ACTIVE_KEY = "thanos.projects.activeId";
const LEGACY_RECENT_KEY = "thanos.recentProjects";

function nowIso() {
  return new Date().toISOString();
}

export function makeProjectId() {
  const rand = Math.random().toString(36).slice(2, 8);
  return `proj-${Date.now().toString(36)}-${rand}`;
}

function read(): Project[] {
  try {
    const raw = localStorage.getItem(PROJECTS_KEY);
    if (raw) return JSON.parse(raw) as Project[];
  } catch {
    // fall through to migration / empty
  }
  const seeded = migrateLegacy();
  if (seeded.length) write(seeded);
  return seeded;
}

function write(projects: Project[]) {
  localStorage.setItem(PROJECTS_KEY, JSON.stringify(projects));
}

// Seed the new store from the legacy `thanos.recentProjects` list (written by
// the onboarding flow) so existing users keep their projects.
function migrateLegacy(): Project[] {
  try {
    const raw = localStorage.getItem(LEGACY_RECENT_KEY);
    if (!raw) return [];
    const legacy = JSON.parse(raw) as Project[];
    return legacy
      .filter((item) => item && item.rootPath)
      .map((item) => ({
        ...item,
        id: item.id || makeProjectId(),
        repos: item.repos ?? [],
        settings: item.settings ?? {},
        archived: item.archived ?? false,
        createdAt: item.createdAt ?? nowIso(),
        updatedAt: item.updatedAt ?? nowIso(),
      }));
  } catch {
    return [];
  }
}

export function listProjects(includeArchived = true): Project[] {
  const projects = read();
  return includeArchived ? projects : projects.filter((project) => !project.archived);
}

export function getProject(id: string): Project | undefined {
  return read().find((project) => project.id === id);
}

export function getActiveProjectId(): string {
  return localStorage.getItem(ACTIVE_KEY) || "";
}

export function setActiveProjectId(id: string) {
  localStorage.setItem(ACTIVE_KEY, id);
}

export function upsertProject(project: Project): Project[] {
  const projects = read();
  const index = projects.findIndex((item) => item.id === project.id);
  const stamped: Project = { ...project, updatedAt: nowIso() };
  if (index < 0) {
    projects.push({ ...stamped, createdAt: stamped.createdAt || nowIso() });
  } else {
    projects[index] = { ...projects[index], ...stamped };
  }
  write(projects);
  return projects;
}

export function archiveProject(id: string, archived: boolean): Project[] {
  const projects = read().map((project) =>
    project.id === id ? { ...project, archived, updatedAt: nowIso() } : project,
  );
  write(projects);
  return projects;
}

export function touchProject(id: string): Project[] {
  const projects = read().map((project) =>
    project.id === id ? { ...project, lastOpenedAt: nowIso() } : project,
  );
  write(projects);
  return projects;
}

// Recent = non-archived projects ordered by last opened (falling back to
// created time), most recent first.
export function recentProjects(limit = 5): Project[] {
  return listProjects(false)
    .slice()
    .sort((a, b) => (b.lastOpenedAt || b.createdAt || "").localeCompare(a.lastOpenedAt || a.createdAt || ""))
    .slice(0, limit);
}
