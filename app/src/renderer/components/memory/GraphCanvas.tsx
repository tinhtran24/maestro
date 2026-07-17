import { Background, BackgroundVariant, Controls, MiniMap, ReactFlow, type Edge, type Node } from "@xyflow/react";
import "@xyflow/react/dist/style.css";

// GraphCanvas is the shared React Flow surface for the memory graphs (task graph
// and code co-change graph): dark canvas, dotted background, minimap, and
// read-only interaction (pan/zoom/drag, no editing).
export function GraphCanvas({
	nodes,
	edges,
	nodeColor,
}: {
	nodes: Node[];
	edges: Edge[];
	nodeColor: (node: Node) => string;
}) {
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
				<MiniMap pannable zoomable nodeColor={nodeColor} />
				<Controls showInteractive={false} />
			</ReactFlow>
		</div>
	);
}

// layoutCircle places nodes evenly on a circle so a graph is readable before the
// user drags anything. Deterministic so renders stay stable.
export function layoutCircle(count: number, index: number): { x: number; y: number } {
	const radius = Math.max(180, count * 34);
	if (count <= 1) return { x: radius, y: radius };
	const angle = (2 * Math.PI * index) / count - Math.PI / 2;
	return { x: Math.cos(angle) * radius + radius, y: Math.sin(angle) * radius + radius };
}
