import { createFileRoute } from "@tanstack/react-router";
import { MemoryView } from "../components/memory/MemoryView";

export const Route = createFileRoute("/_shell/projects/$projectId_/memory")({
	component: ProjectMemoryRoute,
});

function ProjectMemoryRoute() {
	const { projectId } = Route.useParams();
	return <MemoryView projectId={projectId} />;
}
