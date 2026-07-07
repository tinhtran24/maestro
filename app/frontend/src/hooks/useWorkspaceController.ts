import { useEffect, useMemo, useState } from "react";
import type { Workspace } from "../app/types";
import {
  activateWorkspace,
  createWorkspace,
  deleteWorkspace,
  detectAgents,
  getWorkspaceFolder,
  hasWailsRuntime,
  listWorkspaces,
  loadActiveWorkspace,
  loadWorkspace,
  selectWorkspaceFolder,
} from "../services/wails";
import { emptyRegistry, emptyWorkspace, preserveWorkspaceIdentity } from "../stores/workspaceStore";

export function useWorkspaceController() {
  const [workspace, setWorkspace] = useState<Workspace>(emptyWorkspace);
  const [registry, setRegistry] = useState(emptyRegistry);
  const [loadError, setLoadError] = useState("");
  const runtime = useMemo(() => (hasWailsRuntime() ? "Wails" : "Browser preview"), []);

  useEffect(() => {
    void bootWorkspace();
  }, []);

  async function bootWorkspace() {
    setLoadError("");
    try {
      const records = await listWorkspaces();
      setRegistry(records);
      const loaded = records.activeWorkspaceId ? await loadActiveWorkspace() : null;
      if (loaded) {
        setWorkspace(loaded);
        return;
      }
      const path = await getWorkspaceFolder();
      const fallback = await loadWorkspace(path);
      if (fallback) {
        setWorkspace(fallback);
        return;
      }
      const providers = await detectAgents();
      setWorkspace((current) => ({ ...current, providers }));
    } catch (error) {
      setLoadError(error instanceof Error ? error.message : String(error));
    }
  }

  async function chooseFolder() {
    const path = await selectWorkspaceFolder();
    if (!path) return;
    setLoadError("");
    try {
      const record = await createWorkspace({ path });
      const nextRegistry = await listWorkspaces();
      setRegistry(nextRegistry);
      const loaded = record ? await loadActiveWorkspace() : await loadWorkspace(path);
      setWorkspace(loaded ?? { ...emptyWorkspace, path });
    } catch (error) {
      setLoadError(error instanceof Error ? error.message : String(error));
    }
  }

  async function switchWorkspace(workspaceId: string) {
    if (!workspaceId) return;
    setLoadError("");
    try {
      await activateWorkspace({ id: workspaceId });
      const [nextRegistry, loaded] = await Promise.all([listWorkspaces(), loadActiveWorkspace()]);
      setRegistry(nextRegistry);
      if (loaded) setWorkspace(loaded);
    } catch (error) {
      setLoadError(error instanceof Error ? error.message : String(error));
    }
  }

  async function removeActiveWorkspace() {
    if (!registry.activeWorkspaceId) return;
    setLoadError("");
    try {
      const nextRegistry = await deleteWorkspace({ id: registry.activeWorkspaceId });
      setRegistry(nextRegistry);
      const loaded = nextRegistry.activeWorkspaceId ? await loadActiveWorkspace() : null;
      setWorkspace(loaded ?? emptyWorkspace);
    } catch (error) {
      setLoadError(error instanceof Error ? error.message : String(error));
    }
  }

  async function refreshWorkspace() {
    if (!workspace.path) return;
    const loaded = await loadWorkspace(workspace.path);
    if (loaded) setWorkspace(preserveWorkspaceIdentity(workspace, loaded));
  }

  return {
    workspace,
    registry,
    loadError,
    runtime,
    chooseFolder,
    switchWorkspace,
    removeActiveWorkspace,
    refreshWorkspace,
  };
}
