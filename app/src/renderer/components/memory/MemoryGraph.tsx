import { Background, BackgroundVariant, Controls, MiniMap, ReactFlow, type Edge, type Node } from "@xyflow/react";
import "@xyflow/react/dist/style.css";
import { useMemo } from "react";
import type { MemoryGraph as MemoryGraphData } from "../../hooks/useProjectMemory";

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

// layoutCircle places nodes evenly on a circle so the graph is readable before
// the user drags anything. A deterministic layout keeps renders stable.
function layoutCircle(count: number, index: number): { x: number; y: number } {
	const radius = Math.max(180, count * 34);
	if (count <= 1) return { x: radius, y: radius };
	const angle = (2 * Math.PI * index) / count - Math.PI / 2;
	return { x: Math.cos(angle) * radius + radius, y: Math.sin(angle) * radius + radius };
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
		<div className="h-full w-full">
			<ReactFlow
				colorMode="dark"
				nodes={nodes}
				edges={edges}
				fitView
				proOptions={{ hideAttribution: true }}
				nodesConnectable={false}
				edgesFocusable={false}
				minZoom={0.2}
				maxZoom={2}
			>
				<Background variant={BackgroundVariant.Dots} gap={20} size={1} color="#2a2f3a" />
				<MiniMap pannable zoomable nodeColor={(n) => colorFor((n.data as { taskType?: string })?.taskType ?? "")} />
				<Controls showInteractive={false} />
			</ReactFlow>
		</div>
	);
}
