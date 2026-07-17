import { useQuery } from "@tanstack/react-query";
import type { components } from "../../api/schema";
import { apiClient } from "../lib/api-client";
import { mockMemoryGraph, mockMemoryTasks } from "../lib/mock-data";

export type MemoryTask = components["schemas"]["MemoryTaskDTO"];
export type MemoryGraph = components["schemas"]["MemoryGraphResponse"];
export type MemoryContext = components["schemas"]["MemoryContextResponse"];
export type MemoryContextRole = "planner" | "coder" | "reviewer" | "tester";

const usePreviewData = import.meta.env.VITE_NO_ELECTRON === "1";

export const memoryTasksQueryKey = (projectId?: string) =>
	projectId ? (["memory-tasks", projectId] as const) : (["memory-tasks"] as const);

export const memoryGraphQueryKey = (projectId?: string) =>
	projectId ? (["memory-graph", projectId] as const) : (["memory-graph"] as const);

export async function fetchMemoryTasks(projectId: string): Promise<MemoryTask[]> {
	const { data, error } = await apiClient.GET("/api/v1/projects/{id}/memory/tasks", {
		params: { path: { id: projectId } },
	});
	if (error) throw error;
	return data?.tasks ?? [];
}

export async function fetchMemoryGraph(projectId: string): Promise<MemoryGraph> {
	const { data, error } = await apiClient.GET("/api/v1/projects/{id}/memory/graph", {
		params: { path: { id: projectId } },
	});
	if (error) throw error;
	return data ?? { nodes: [], edges: [] };
}

export async function fetchMemoryContext(
	projectId: string,
	role: MemoryContextRole,
	files: string[],
): Promise<MemoryContext> {
	const { data, error } = await apiClient.GET("/api/v1/projects/{id}/memory/context", {
		params: {
			path: { id: projectId },
			query: { role, files: files.length > 0 ? files.join(",") : undefined },
		},
	});
	if (error) throw error;
	return (
		data ?? {
			role,
			relatedTasks: [],
			relevantFiles: [],
			relevantTests: [],
			decisions: [],
			dropped: [],
			estimatedTokens: 0,
		}
	);
}

export async function rebuildMemory(projectId: string): Promise<number> {
	const { data, error } = await apiClient.POST("/api/v1/projects/{id}/memory/rebuild", {
		params: { path: { id: projectId } },
	});
	if (error) throw error;
	return data?.tasks ?? 0;
}

export function useMemoryTasks(projectId?: string) {
	return useQuery({
		queryKey: memoryTasksQueryKey(projectId),
		enabled: Boolean(projectId),
		queryFn: () => (usePreviewData ? Promise.resolve(mockMemoryTasks) : fetchMemoryTasks(projectId!)),
		retry: 1,
	});
}

export function useMemoryGraph(projectId?: string) {
	return useQuery({
		queryKey: memoryGraphQueryKey(projectId),
		enabled: Boolean(projectId),
		queryFn: () => (usePreviewData ? Promise.resolve(mockMemoryGraph) : fetchMemoryGraph(projectId!)),
		retry: 1,
	});
}
