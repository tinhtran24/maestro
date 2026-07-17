import { type Edge, type Node } from "@xyflow/react";
import { useMemo } from "react";
import type { MemoryGraph as MemoryGraphData } from "../../hooks/useProjectMemory";
import { GraphCanvas, layoutCircle } from "./GraphCanvas";

// taskTypeColor maps a coarse task type to a node accent that reads on the dark
// canvas. Unknown types fall back to slate.
const taskTypeColor: Record<string, string> = {
	feature: "#4d8dff",
	bugfix: "#f2545b",
	refactor: "#a877ff",
	docs: "#8a94a6",
	test: "#37c98b",
	chore: "#c9a23a",
	perf: "#37b3c9",
};

function colorFor(taskType: string): string {
	return taskTypeColor[taskType] ?? "#6b7688";
}

export function MemoryGraph({ graph }: { graph: MemoryGraphData }) {
	const nodes = useMemo<Node[]>(() => {
		const count = graph.nodes.length;
		return graph.nodes.map((n, i) => {
			const accent = colorFor(n.taskType ?? "");
			return {
				id: n.taskId,
				position: layoutCircle(count, i),
				data: {
					taskType: n.taskType ?? "",
					label: (
						<div className="flex flex-col items-center gap-0.5 px-1 py-0.5">
							<span className="font-mono text-[11px] font-semibold" style={{ color: accent }}>
								{n.taskId}
							</span>
							<span className="max-w-[150px] truncate text-[10px] text-passive" title={n.intent}>
								{n.intent || n.taskType || "task"}
							</span>
							<span className="text-[9px] text-passive/70">
								{n.changedFiles} file{n.changedFiles === 1 ? "" : "s"}
								{n.changedTests > 0 ? ` · ${n.changedTests} test${n.changedTests === 1 ? "" : "s"}` : ""}
							</span>
						</div>
					),
				},
				style: {
					borderRadius: 10,
					border: `1px solid ${accent}55`,
					background: "var(--surface, #14161c)",
					boxShadow: `0 0 0 1px ${accent}22`,
					width: 172,
					fontSize: 11,
				},
			};
		});
	}, [graph.nodes]);

	const edges = useMemo<Edge[]>(
		() =>
			graph.edges.map((e, i) => ({
				id: `${e.source}-${e.target}-${e.relation}-${i}`,
				source: e.source,
				target: e.target,
				animated: false,
				style: { stroke: "#4d8dff66", strokeWidth: Math.min(1 + e.confidence, 4) },
				label: e.confidence > 1 ? String(Math.round(e.confidence)) : undefined,
				labelStyle: { fill: "var(--passive, #8a94a6)", fontSize: 9 },
			})),
		[graph.edges],
	);

	return (
		<GraphCanvas
			nodes={nodes}
			edges={edges}
			nodeColor={(n) => colorFor((n.data as { taskType?: string })?.taskType ?? "")}
		/>
	);
}
