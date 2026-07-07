import type { FlowNode, ProviderInfo, RoutineInfo, SpecNode, Task, Workspace } from "../app/types";

export type MissionNodeKind = "spec" | "task" | "routine" | "provider" | "flow";
export type MissionEdgeKind = "dispatches" | "depends_on" | "implements" | "blocked_by" | "produced" | "routes_to";

export type MissionNode = {
  id: string;
  kind: MissionNodeKind;
  title: string;
  status: string;
  detail: string;
};

export type MissionEdge = {
  id: string;
  from: string;
  to: string;
  kind: MissionEdgeKind;
};

export type MissionGraph = {
  nodes: MissionNode[];
  edges: MissionEdge[];
  blockedTaskIds: string[];
  criticalPath: string[];
};

export function buildMissionGraph(workspace: Workspace): MissionGraph {
  const specs = flattenSpecs(workspace.specs);
  const nodes: MissionNode[] = [
    ...specs.map(specNode),
    ...workspace.tasks.map(taskNode),
    ...workspace.routines.map(routineNode),
    ...workspace.providers.map(providerNode),
    ...workspace.flows.map(flowNode),
  ];
  const edges: MissionEdge[] = [];

  for (const task of workspace.tasks) {
    const spec = matchSpecForTask(task, specs);
    if (spec) edges.push(edge(specNodeId(spec.path), taskNodeId(task.id), "dispatches"));
    for (const dependency of task.dependencies) {
      edges.push(edge(taskNodeId(task.id), taskNodeId(dependency), "depends_on"));
      if (!isTaskDone(dependency, workspace.tasks)) edges.push(edge(taskNodeId(task.id), taskNodeId(dependency), "blocked_by"));
    }
    if (task.flow) edges.push(edge(taskNodeId(task.id), flowNodeId(task.flow), "implements"));
    if (task.agent) edges.push(edge(taskNodeId(task.id), providerNodeId(task.agent), "routes_to"));
    if (task.commit?.hash) edges.push(edge(taskNodeId(task.id), `commit:${task.commit.hash}`, "produced"));
    if (task.oversight?.id) edges.push(edge(taskNodeId(task.id), `oversight:${task.oversight.id}`, "produced"));
  }

  for (const routine of workspace.routines) {
    if (routine.flow) edges.push(edge(routineNodeId(routine.id), flowNodeId(routine.flow), "implements"));
    for (const task of workspace.tasks) {
      if (task.prompt.includes(`Routine: ${routine.id}`)) edges.push(edge(routineNodeId(routine.id), taskNodeId(task.id), "produced"));
    }
  }

  for (const flow of workspace.flows) {
    for (const provider of workspace.providers) {
      if (flow.steps.some((step) => provider.name.toLowerCase().includes(step.toLowerCase()) || provider.id.toLowerCase().includes(step.toLowerCase()))) {
        edges.push(edge(flowNodeId(flow.id), providerNodeId(provider.id), "routes_to"));
      }
    }
  }

  const blockedTaskIds = workspace.tasks.filter((task) => task.blocked || task.dependencies.some((dependency) => !isTaskDone(dependency, workspace.tasks))).map((task) => task.id);
  return {
    nodes,
    edges: uniqueEdges(edges),
    blockedTaskIds,
    criticalPath: computeCriticalPath(workspace.tasks),
  };
}

function specNode(spec: SpecNode): MissionNode {
  return { id: specNodeId(spec.path), kind: "spec", title: spec.title, status: spec.state, detail: spec.path };
}

function taskNode(task: Task): MissionNode {
  return { id: taskNodeId(task.id), kind: "task", title: task.title, status: task.status, detail: task.id };
}

function routineNode(routine: RoutineInfo): MissionNode {
  return { id: routineNodeId(routine.id), kind: "routine", title: routine.name, status: routine.enabled ? "enabled" : "disabled", detail: routine.schedule };
}

function providerNode(provider: ProviderInfo): MissionNode {
  return { id: providerNodeId(provider.id), kind: "provider", title: provider.name, status: provider.status, detail: provider.command };
}

function flowNode(flow: FlowNode): MissionNode {
  return { id: flowNodeId(flow.id), kind: "flow", title: flow.name, status: flow.readOnly ? "builtin" : "custom", detail: flow.steps.join(" -> ") };
}

function edge(from: string, to: string, kind: MissionEdgeKind): MissionEdge {
  return { id: `${kind}:${from}:${to}`, from, to, kind };
}

function uniqueEdges(edges: MissionEdge[]) {
  return Array.from(new Map(edges.map((item) => [item.id, item])).values());
}

function flattenSpecs(nodes: SpecNode[]): SpecNode[] {
  return nodes.flatMap((node) => [node, ...flattenSpecs(node.children)]);
}

function matchSpecForTask(task: Task, specs: SpecNode[]) {
  const title = task.title.toLowerCase();
  return specs.find((spec) => spec.title.toLowerCase() === title || task.prompt.includes(spec.path));
}

function isTaskDone(taskId: string, tasks: Task[]) {
  return tasks.some((task) => task.id === taskId && task.status === "done");
}

function computeCriticalPath(tasks: Task[]) {
  const byId = new Map(tasks.map((task) => [task.id, task]));
  const memo = new Map<string, string[]>();
  function pathFor(task: Task): string[] {
    if (memo.has(task.id)) return memo.get(task.id) ?? [];
    const dependencyPaths = task.dependencies.map((id) => byId.get(id)).filter((item): item is Task => Boolean(item)).map(pathFor);
    const longest = dependencyPaths.sort((a, b) => b.length - a.length)[0] ?? [];
    const path = [...longest, task.id];
    memo.set(task.id, path);
    return path;
  }
  return tasks.map(pathFor).sort((a, b) => b.length - a.length)[0] ?? [];
}

const specNodeId = (path: string) => `spec:${path}`;
const taskNodeId = (id: string) => `task:${id}`;
const routineNodeId = (id: string) => `routine:${id}`;
const providerNodeId = (id: string) => `provider:${id}`;
const flowNodeId = (id: string) => `flow:${id}`;
