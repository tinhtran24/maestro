import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useState } from "react";
import { RefreshCw } from "lucide-react";
import {
	type MemoryContext,
	type MemoryContextRole,
	type MemoryTask,
	fetchMemoryContext,
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
import { Input } from "../ui/input";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "../ui/table";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "../ui/tabs";
import { cn } from "../../lib/utils";
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

const ROLES: MemoryContextRole[] = ["planner", "coder", "reviewer", "tester"];

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
					<TabsTrigger value="graph">Graph</TabsTrigger>
					<TabsTrigger value="context">Context</TabsTrigger>
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

				<TabsContent value="context" className="min-h-0 flex-1 overflow-y-auto p-[18px]">
					<ContextTab projectId={projectId} tasks={tasks} />
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

function ContextTab({ projectId, tasks }: { projectId: string; tasks: MemoryTask[] }) {
	const [role, setRole] = useState<MemoryContextRole>("coder");
	const [files, setFiles] = useState("");
	const build = useMutation<MemoryContext, unknown, void>({
		mutationFn: () =>
			fetchMemoryContext(
				projectId,
				role,
				files
					.split(",")
					.map((f) => f.trim())
					.filter(Boolean),
			),
	});

	const suggestedFiles = Array.from(new Set(tasks.flatMap((t) => t.changedFiles))).slice(0, 6);
	const pack = build.data;

	return (
		<div className="mx-auto max-w-3xl space-y-4">
			<div className="rounded-lg border border-border bg-surface/40 p-4">
				<div className="mb-3 flex flex-wrap items-center gap-2">
					<span className="text-[12px] text-passive">Role</span>
					{ROLES.map((r) => (
						<button
							key={r}
							type="button"
							onClick={() => setRole(r)}
							className={cn(
								"rounded-full border px-3 py-1 text-[12px] capitalize transition-colors",
								role === r
									? "border-accent bg-accent/15 text-foreground"
									: "border-border text-passive hover:text-foreground",
							)}
						>
							{r}
						</button>
					))}
				</div>
				<div className="flex flex-col gap-2 sm:flex-row">
					<Input
						value={files}
						onChange={(e) => setFiles(e.target.value)}
						placeholder="Changed files to match, comma-separated (e.g. internal/httpd/router.go)"
						className="flex-1 text-[12px]"
					/>
					<Button size="sm" onClick={() => build.mutate()} disabled={build.isPending}>
						{build.isPending ? "Building…" : "Build pack"}
					</Button>
				</div>
				{suggestedFiles.length > 0 && (
					<div className="mt-2 flex flex-wrap items-center gap-1.5">
						<span className="text-[11px] text-passive">Try:</span>
						{suggestedFiles.map((f) => (
							<button
								key={f}
								type="button"
								onClick={() => setFiles((prev) => (prev ? `${prev}, ${f}` : f))}
								className="rounded border border-border px-1.5 py-0.5 font-mono text-[10.5px] text-passive hover:text-foreground"
							>
								{f.split("/").pop()}
							</button>
						))}
					</div>
				)}
			</div>

			{build.error ? <CenterNote>{apiErrorMessage(build.error)}</CenterNote> : null}

			{pack ? (
				<div className="space-y-4">
					<div className="flex items-center gap-3 text-[12px] text-passive">
						<span className="capitalize text-foreground">{pack.role} pack</span>
						<span>~{pack.estimatedTokens} tokens</span>
						{pack.dropped.length > 0 && <span>{pack.dropped.length} dropped for budget</span>}
					</div>

					<PackSection title={`Related tasks (${pack.relatedTasks.length})`}>
						{pack.relatedTasks.length === 0 ? (
							<p className="text-[12px] text-passive">No related prior tasks for these paths.</p>
						) : (
							<ul className="space-y-1.5">
								{pack.relatedTasks.map((t) => (
									<li key={t.taskId} className="flex items-center gap-2 text-[12px]">
										<span className="font-mono">{t.taskId}</span>
										<TypeTag taskType={t.taskType ?? ""} />
										<span className="font-mono text-[11px] text-passive">
											{(t.sharedFiles ?? 0) + (t.sharedTests ?? 0)} shared
										</span>
										<span className="truncate text-passive" title={t.intent}>
											{t.intent}
										</span>
									</li>
								))}
							</ul>
						)}
					</PackSection>

					{pack.relevantFiles.length > 0 && (
						<PackSection title={`Relevant files (${pack.relevantFiles.length})`}>
							<FileList paths={pack.relevantFiles} />
						</PackSection>
					)}
					{pack.relevantTests.length > 0 && (
						<PackSection title={`Relevant tests (${pack.relevantTests.length})`}>
							<FileList paths={pack.relevantTests} />
						</PackSection>
					)}
					{pack.decisions.length > 0 && (
						<PackSection title="Protected decisions">
							<ul className="list-disc space-y-1 pl-5 text-[12px]">
								{pack.decisions.map((d) => (
									<li key={d}>{d}</li>
								))}
							</ul>
						</PackSection>
					)}
				</div>
			) : (
				!build.isPending && (
					<p className="text-center text-[12px] text-passive">
						Pick a role and some files, then build the token-budgeted pack an agent stage would receive.
					</p>
				)
			)}
		</div>
	);
}

function PackSection({ title, children }: { title: string; children: React.ReactNode }) {
	return (
		<div className="rounded-lg border border-border bg-surface/30 p-3">
			<h3 className="mb-2 text-[12px] font-semibold text-foreground">{title}</h3>
			{children}
		</div>
	);
}

function FileList({ paths }: { paths: string[] }) {
	return (
		<ul className="space-y-0.5 font-mono text-[11px] text-passive">
			{paths.map((p) => (
				<li key={p} className="truncate" title={p}>
					{p}
				</li>
			))}
		</ul>
	);
}
