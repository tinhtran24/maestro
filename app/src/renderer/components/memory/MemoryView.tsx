import { useMutation, useQueryClient } from "@tanstack/react-query";
import { RefreshCw } from "lucide-react";
import {
	type MemoryTask,
	memoryGraphQueryKey,
	memoryTasksQueryKey,
	rebuildMemory,
	useMemoryGraph,
	useMemoryTasks,
} from "../../hooks/useProjectMemory";
import { apiErrorMessage } from "../../lib/api-client";
import { useWorkspaceQuery } from "../../hooks/useWorkspaceQuery";
import { DashboardSubhead } from "../DashboardSubhead";
import { Button } from "../ui/button";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "../ui/table";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "../ui/tabs";
import { cn } from "../../lib/utils";
import { CodeGraph } from "./CodeGraph";
import { MemoryGraph } from "./MemoryGraph";

const typeColor: Record<string, string> = {
	feature: "#4d8dff",
	bugfix: "#f2545b",
	refactor: "#a877ff",
	docs: "#8a94a6",
	test: "#37c98b",
	chore: "#c9a23a",
	perf: "#37b3c9",
};

function TypeTag({ taskType }: { taskType: string }) {
	if (!taskType) return <span className="text-passive">—</span>;
	const color = typeColor[taskType] ?? "#6b7688";
	return (
		<span
			className="inline-flex items-center rounded-full px-2 py-0.5 font-mono text-[10.5px]"
			style={{ color, backgroundColor: `${color}1a`, border: `1px solid ${color}33` }}
		>
			{taskType}
		</span>
	);
}

export function MemoryView({ projectId }: { projectId: string }) {
	const workspaceQuery = useWorkspaceQuery();
	const projectName = (workspaceQuery.data ?? []).find((w) => w.id === projectId)?.name ?? projectId;
	const tasksQuery = useMemoryTasks(projectId);
	const graphQuery = useMemoryGraph(projectId);
	const queryClient = useQueryClient();

	const rebuild = useMutation({
		mutationFn: () => rebuildMemory(projectId),
		onSuccess: () => {
			void queryClient.invalidateQueries({ queryKey: memoryTasksQueryKey(projectId) });
			void queryClient.invalidateQueries({ queryKey: memoryGraphQueryKey(projectId) });
		},
	});

	const tasks = tasksQuery.data ?? [];
	const graph = graphQuery.data ?? { nodes: [], edges: [] };

	return (
		<div className="flex h-full min-h-0 flex-col bg-background text-foreground">
			<DashboardSubhead
				title="Memory"
				subtitle={`Completed-task memory and the shared-path task graph for ${projectName}.`}
				count={tasks.length}
				actions={
					<Button
						size="sm"
						variant="secondary"
						disabled={rebuild.isPending}
						onClick={() => rebuild.mutate()}
						title="Rebuild the projection from the committed event log"
					>
						<RefreshCw aria-hidden="true" className={cn("size-3.5", rebuild.isPending && "animate-spin")} />
						{rebuild.isPending ? "Rebuilding…" : "Rebuild"}
					</Button>
				}
			/>

			<Tabs defaultValue="tasks" className="flex min-h-0 flex-1 flex-col">
				<TabsList className="mx-[18px] mt-3 self-start">
					<TabsTrigger value="tasks">Tasks</TabsTrigger>
					<TabsTrigger value="graph">Task graph</TabsTrigger>
					<TabsTrigger value="code">Code graph</TabsTrigger>
				</TabsList>

				<TabsContent value="tasks" className="min-h-0 flex-1 overflow-y-auto p-[18px]">
					<TasksTab query={tasksQuery} tasks={tasks} />
				</TabsContent>

				<TabsContent value="graph" className="min-h-0 flex-1 border-t border-border">
					{graphQuery.isLoading ? (
						<CenterNote>Loading graph…</CenterNote>
					) : graph.nodes.length === 0 ? (
						<CenterNote>No task graph yet. Complete a few sessions to populate it.</CenterNote>
					) : (
						<MemoryGraph graph={graph} />
					)}
				</TabsContent>

				<TabsContent value="code" className="min-h-0 flex-1 border-t border-border">
					{tasksQuery.isLoading ? (
						<CenterNote>Loading code graph…</CenterNote>
					) : tasks.length === 0 ? (
						<CenterNote>No changed files recorded yet. The code graph fills in as sessions complete.</CenterNote>
					) : (
						<CodeGraph tasks={tasks} />
					)}
				</TabsContent>
			</Tabs>
		</div>
	);
}

function CenterNote({ children }: { children: React.ReactNode }) {
	return (
		<p className="flex h-full items-center justify-center px-6 text-center text-[12px] text-passive">{children}</p>
	);
}

function TasksTab({ query, tasks }: { query: ReturnType<typeof useMemoryTasks>; tasks: MemoryTask[] }) {
	if (query.isLoading) return <CenterNote>Loading task memory…</CenterNote>;
	if (query.error) return <CenterNote>{apiErrorMessage(query.error)}</CenterNote>;
	if (tasks.length === 0) {
		return <CenterNote>No task memory recorded yet. It fills in as sessions complete.</CenterNote>;
	}
	return (
		<Table>
			<TableHeader>
				<TableRow>
					<TableHead className="w-28">Task</TableHead>
					<TableHead className="w-24">Type</TableHead>
					<TableHead className="w-24">Kind</TableHead>
					<TableHead className="w-16 text-right">Files</TableHead>
					<TableHead className="w-16 text-right">Tests</TableHead>
					<TableHead>Intent</TableHead>
				</TableRow>
			</TableHeader>
			<TableBody>
				{tasks.map((t) => (
					<TableRow key={t.id}>
						<TableCell className="font-mono text-[12px]">{t.id}</TableCell>
						<TableCell>
							<TypeTag taskType={t.taskType ?? ""} />
						</TableCell>
						<TableCell className="text-[12px] text-passive">{t.kind}</TableCell>
						<TableCell className="text-right font-mono text-[12px]">{t.changedFiles.length}</TableCell>
						<TableCell className="text-right font-mono text-[12px]">{t.changedTests.length}</TableCell>
						<TableCell className="max-w-0 truncate text-[12px]" title={t.intent}>
							{t.intent || <span className="text-passive">—</span>}
						</TableCell>
					</TableRow>
				))}
			</TableBody>
		</Table>
	);
}
