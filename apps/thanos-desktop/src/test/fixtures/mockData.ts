import type { AgentSession, ExecutionPlan, Feature, MemoryNode, Project, Review, Skill, SkillRun, Task } from "../../domain/models";

export const project: Project = {
  id: "project-ecommerce",
  name: "E-commerce Platform",
  rootPath: "/workspace/e-commerce",
  repos: ["e-commerce"],
  settings: { defaultPlanner: "Claude Code", defaultCoder: "Codex" },
};

export const features: Feature[] = [
  {
    id: "F-100",
    projectId: project.id,
    title: "E-commerce Core",
    description: "Shopping, catalog, checkout, and payments.",
    status: "active",
    planGraphId: "graph-ecommerce-core",
    createdAt: "2024-05-10T00:00:00Z",
  },
];

export const tasks: Task[] = [
  task("T-106", "Shopping Cart Implementation", "Implement add, update, remove, persistence, API sync, and review-ready tests.", "waiting_approval", "P0", "Planner / Claude Code", 34, ["cart", "checkout"]),
  task("T-100", "Product Catalog API", "Create paginated catalog endpoints and contract tests.", "running", "P0", "Coder / Codex", 72, ["api", "backend"]),
  task("T-099", "Database Schema Design", "Review migrations, indexes, and rollback strategy.", "in_review", "P0", "Reviewer / Codex", 88, ["db", "schema"]),
  task("T-101", "User Authentication System", "Plan OAuth handoff, session storage, and guarded routes.", "planning", "P1", "Planner / Claude Code", 48, ["auth", "security"]),
  task("T-097", "Project Setup", "Initial toolchain, scripts, and baseline app shell.", "done", "P3", "Coder / Codex", 100, ["setup"]),
  task("T-103", "Payment Gateway Integration", "Backlog payment provider adapter and failure modes.", "backlog", "P1", "Unassigned", 0, ["stripe", "payment"]),
];

export const plans: ExecutionPlan[] = [
  {
    id: "plan-T-106",
    taskId: "T-106",
    summary: "Create a cart domain model, backend API, frontend context, persistent state, UI components, and tests.",
    approvalStatus: "pending",
    steps: [
      step("1", "Analyze requirements and data model", "Model Cart, CartItem, Product relationships"),
      step("2", "Create cart schema and migrations", "Add cart tables and indexes"),
      step("3", "Implement backend API endpoints", "CRUD operations for cart"),
      step("4", "Implement frontend state management", "React context for cart state"),
      step("5", "Create cart UI components", "Cart icon, sidebar, item list"),
      step("6", "Add persistence and sync", "localStorage plus backend sync"),
      step("7", "Write tests", "Unit and integration tests"),
    ],
    risks: ["Data consistency issues", "Performance with large carts", "Sync conflicts"],
    filesToTouch: ["api/cart/route.ts", "models/cart.ts", "components/CartSidebar.tsx", "context/CartContext.tsx", "types/cart.ts"],
    testStrategy: ["Unit tests for cart logic", "Integration tests for API", "E2E test for cart flow"],
  },
];

export const sessions: AgentSession[] = [
  {
    id: "session-T-106-planner",
    taskId: "T-106",
    agentType: "planner",
    provider: "claude-code",
    command: "claude",
    status: "stopped",
    conversationLogPath: ".thanos/logs/T-106-planner.md",
    output: ["Plan generated. Waiting for human approval before coder starts."],
  },
];

export const reviews: Review[] = [
  {
    id: "review-T-106",
    taskId: "T-106",
    diffSummary: "No implementation diff yet. Plan approval is pending.",
    changedFiles: ["api/cart/route.ts", "models/cart.ts", "components/CartSidebar.tsx"],
    testResults: [],
    reviewerNotes: "Reviewer starts after coder marks task ready for review.",
    status: "pending",
  },
];

export const skills: Skill[] = [
  skill("project:using-agent-skills", "using-agent-skills", "Route each task to the smallest relevant skill set.", ["planner", "coder", "reviewer", "tester"], ["selected_skills", "routing_reason"], ["Selected skills are listed.", "Routing reason is recorded."]),
  skill("project:feature-spec", "feature-spec", "Create clear feature specs with acceptance criteria before implementation.", ["planner"], ["acceptance_criteria", "risk_notes", "files_to_touch"], ["Acceptance criteria exist.", "Risks are listed.", "Files likely to change are recorded."]),
  skill("project:implementation-plan", "implementation-plan", "Break approved specs into ordered implementation chunks.", ["planner", "coder"], ["ordered_steps", "dependencies", "test_strategy"], ["Ordered implementation steps exist.", "Dependencies are called out.", "Test strategy is attached."]),
];

export const skillRuns: SkillRun[] = [
  skillRun("skillrun-T-106-router", "T-106", "project:using-agent-skills", "completed", { selected_skills: "feature-spec, implementation-plan", routing_reason: "Planner task in waiting approval stage." }),
  skillRun("skillrun-T-106-spec", "T-106", "project:feature-spec", "evidence_pending", { acceptance_criteria: "Cart add, update, remove, persistence, sync, and review-ready tests." }),
  skillRun("skillrun-T-106-plan", "T-106", "project:implementation-plan", "matched", {}),
];

export const memoryNodes: MemoryNode[] = [
  memory("mem-1", "decision", "Cart System Design Decision", "Cart state is frontend-first with backend sync after checkout mutations."),
  memory("mem-2", "architecture", "Database Schema Guidelines", "Prefer indexed foreign keys and explicit rollback migrations."),
  memory("mem-3", "feature", "Shopping Cart Requirements", "Cart supports add, remove, quantity update, and persisted sessions."),
];

function task(id: string, title: string, description: string, status: Task["status"], priority: Task["priority"], assignedAgent: string, progress: number, tags: string[]): Task {
  return {
    id,
    featureId: "F-100",
    title,
    description,
    status,
    priority,
    assignedAgent,
    executorProfile: assignedAgent.includes("Codex") ? "codex-local" : "claude-local",
    branchName: status === "backlog" || status === "planning" ? "" : `thanos/${id.toLowerCase()}-${title.toLowerCase().replace(/\s+/g, "-")}`,
    worktreePath: status === "backlog" || status === "planning" ? "" : `.thanos/worktrees/${id}`,
    reviewApproved: status === "done" || status === "running",
    testsPassed: status === "done" || status === "in_review",
    updatedAt: status === "done" ? "2d ago" : "2:45 PM",
    tags,
    progress,
  };
}

function step(id: string, title: string, description: string) {
  return { id, title, description, status: "pending" };
}

function memory(id: string, type: MemoryNode["type"], title: string, content: string): MemoryNode {
  return { id, projectId: project.id, type, title, content, links: [], createdAt: "2024-05-18T00:00:00Z" };
}

function skill(id: string, name: string, description: string, agents: Skill["agents"], requiredEvidence: string[], exitCriteria: string[]): Skill {
  return {
    id,
    projectId: project.id,
    name,
    path: `.thanos/skills/${name}/SKILL.md`,
    description,
    appliesTo: ["planning"],
    agents,
    source: "project",
    requiredEvidence,
    exitCriteria,
    enabled: true,
    trusted: true,
  };
}

function skillRun(id: string, taskId: string, skillId: string, status: SkillRun["status"], evidence: Record<string, string>): SkillRun {
  return {
    id,
    taskId,
    skillId,
    status,
    evidence,
    startedAt: "2026-07-03T01:31:39Z",
  };
}
