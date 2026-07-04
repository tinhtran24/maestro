import { create } from "zustand";
import type {
    AgentSession,
    AgentProvider,
    BottomTab,
    ExecutionPlan,
    Feature,
    GitDiff,
    InspectorTab,
    MemoryNode,
    PlanningQuestion,
    Project,
    Review,
    Skill,
    SkillRun,
    Task,
    TaskEvent,
    TaskEventType,
    TaskStatus,
    TestRun,
    WorkflowStepConfig,
    WorkflowStepId,
} from "../domain/models";
import type { WorkbenchSnapshot } from "../services/nativeBackend";
import {
    applyProviderOverrides,
    loadProviderOverrides,
    loadWorkflowSteps,
    saveProviderOverride,
    saveWorkflowSteps,
} from "../features/agents/api/agentConfig";
import { transitionEventType, transitionTask } from "./taskMachine";

// --- Phase 3 — Workflow Engine helpers -----------------------------------

function nowIso() {
    return new Date().toISOString();
}

function slug(value: string) {
    return value.toLowerCase().replace(/[^a-z0-9]+/g, "-").replace(/^-|-$/g, "");
}

function makeEvent(taskId: string, type: TaskEventType, patch: Partial<TaskEvent> = {}): TaskEvent {
    const rand = Math.random().toString(36).slice(2, 6);
    return { id: `evt-${taskId}-${Date.now().toString(36)}-${rand}`, taskId, type, at: nowIso(), ...patch };
}

function pushEvent(history: Record<string, TaskEvent[]>, event: TaskEvent): Record<string, TaskEvent[]> {
    return { ...history, [event.taskId]: [...(history[event.taskId] ?? []), event] };
}

// Mock worktree assignment (no Git). Satisfies the `running` gate in Phase 3.
function ensureWorktree(task: Task): Task {
    if (task.worktreePath && task.branchName) return task;
    const branchName = task.branchName || `thanos/${task.id.toLowerCase()}-${slug(task.title)}`;
    const worktreePath = task.worktreePath || `.thanos/worktrees/${task.id.toLowerCase()}`;
    return { ...task, branchName, worktreePath };
}

type TransitionOutcome = { tasks: Task[]; event: TaskEvent } | { error: string };

// Validates and applies a single transition through the state machine. Assigns a
// mock worktree before `running` so the isolation gate holds. Returns an error
// string when the transition is invalid — invalid transitions never mutate.
function applyTransition(tasks: Task[], taskId: string, to: TaskStatus, note?: string): TransitionOutcome {
    const index = tasks.findIndex((task) => task.id === taskId);
    if (index < 0) return { error: "Task not found." };
    const current = tasks[index];
    const prepared = to === "running" ? ensureWorktree(current) : current;
    const result = transitionTask(prepared, to);
    if (!result.ok) return { error: result.reason };
    const next = tasks.slice();
    next[index] = result.task;
    return { tasks: next, event: makeEvent(taskId, transitionEventType(to), { from: current.status, to, note }) };
}


type WorkbenchState = {
    activeView: "workbench" | "projects" | "memory" | "executors" | "workflow_steps" | "settings";
    loadError: string;
    project: Project;
    features: Feature[];
    tasks: Task[];
    plans: ExecutionPlan[];
    reviews: Review[];
    memoryNodes: MemoryNode[];
    sessions: AgentSession[];
    pinnedSessions: string[];
    activeSessionByTask: Record<string, string>;
    skills: Skill[];
    skillRuns: SkillRun[];
    agentProviders: AgentProvider[];
    workflowSteps: WorkflowStepConfig[];
    diffs: Record<string, GitDiff>;
    testRuns: Record<string, TestRun>;
    taskHistory: Record<string, TaskEvent[]>;
    planningByTask: Record<string, { questions: PlanningQuestion[] }>;
    selectedTaskId: string;
    boardFilter: string;
    inspectorTab: InspectorTab;
    bottomTab: BottomTab;
    rightCollapsed: boolean;
    taskDialog: { mode: "create" | "edit"; taskId?: string } | null;
    setActiveView(view: WorkbenchState["activeView"]): void;
    toggleRightCollapsed(): void;
    setLoadError(value: string): void;
    hydrate(snapshot: WorkbenchSnapshot): void;
    setProject(project: Project): void;
    setAgentProviders(providers: AgentProvider[]): void;
    updateAgentProvider(id: string, patch: Partial<AgentProvider>): void;
    setWorkflowSteps(steps: WorkflowStepConfig[]): void;
    updateWorkflowStep(id: WorkflowStepId, patch: Partial<WorkflowStepConfig>): void;
    selectTask(taskId: string): void;
    setBoardFilter(value: string): void;
    setInspectorTab(tab: InspectorTab): void;
    setBottomTab(tab: BottomTab): void;
    openCreateTask(): void;
    openEditTask(taskId: string): void;
    closeTaskDialog(): void;
    createTask(input: { title: string; description: string; priority: Task["priority"]; assignedAgent: string; tags?: string[] }): void;
    editTask(taskId: string, input: { title: string; description: string; priority: Task["priority"]; assignedAgent: string }): void;
    removeTask(taskId: string): void;
    moveTask(taskId: string, status: TaskStatus): void;
    advanceTask(taskId: string, status: TaskStatus, note?: string): void;
    approvePlan(taskId: string): void;
    persistPlan(plan: ExecutionPlan): void;
    requestChanges(taskId: string): void;
    approveReview(taskId: string): void;
    rejectReview(taskId: string): void;
    requestReviewChanges(taskId: string): void;
    appendMemory(node: MemoryNode): void;
    runTests(taskId: string): void;
    approveMerge(taskId: string): void;
    persistDiff(diff: GitDiff): void;
    persistTestRun(test: TestRun): void;
    persistReview(review: Review): void;
    persistMemory(nodes: MemoryNode[]): void;
    setPlanningQuestions(taskId: string, questions: PlanningQuestion[]): void;
    answerPlanningQuestion(taskId: string, questionId: string, answer: string): void;
    upsertSession(session: AgentSession): void;
    appendSessionOutput(
        taskId: string,
        sessionId: string,
        output: string,
    ): void;
    patchSession(sessionId: string, patch: Partial<AgentSession>): void;
    removeSession(sessionId: string): void;
    togglePinnedSession(sessionId: string): void;
    setActiveSession(taskId: string, sessionId: string): void;
};

const emptyProject: Project = {
    id: "local",
    name: "Thanos",
    rootPath: "",
    repos: [],
    settings: {},
};

export const defaultWorkflowSteps: WorkflowStepConfig[] = [
    step("planning", "Planning", "Claude Code", "claude", "project", true, true, ["read", "write-plans"]),
    step("coding", "Coding", "Codex", "codex", "worktree", true, true, ["read", "write-code", "run-tests"]),
    step("review", "Review", "Claude Code", "claude", "worktree", true, true, ["read", "inspect-diff"]),
    step("testing", "Testing", "Shell", "npm test", "worktree", true, false, ["run-tests"]),
    step("debugging", "Debugging", "Codex", "codex", "worktree", false, true, ["read", "write-code", "run-tests"]),
    step("documentation", "Documentation", "Codex", "codex", "worktree", false, true, ["read", "write-docs"]),
    step("memory_update", "Memory Update", "Shell", "thanos memory update", "project", false, false, ["write-memory"]),
];

export const useWorkbenchStore = create<WorkbenchState>((set, get) => ({
    activeView: "workbench",
    loadError: "",
    project: emptyProject,
    features: [],
    tasks: [],
    plans: [],
    reviews: [],
    memoryNodes: [],
    sessions: [],
    pinnedSessions: [],
    activeSessionByTask: {},
    skills: [],
    skillRuns: [],
    agentProviders: [],
    workflowSteps: loadWorkflowSteps(defaultWorkflowSteps),
    diffs: {},
    testRuns: {},
    taskHistory: {},
    planningByTask: {},
    selectedTaskId: "",
    boardFilter: "",
    inspectorTab: "plan",
    bottomTab: "terminal",
    rightCollapsed: false,
    taskDialog: null,
    setActiveView: (activeView) => set({ activeView }),
    toggleRightCollapsed: () => set((state) => ({ rightCollapsed: !state.rightCollapsed })),
    setLoadError: (loadError) => set({ loadError }),
    hydrate: (snapshot) =>
        set((state) => {
            const selectedExists = snapshot.tasks.some(
                (task) => task.id === state.selectedTaskId,
            );
            // Seed a baseline history entry for any task without one, so the
            // timeline reflects the task's current workflow position.
            const taskHistory = { ...state.taskHistory };
            for (const task of snapshot.tasks) {
                if (!taskHistory[task.id]?.length) {
                    taskHistory[task.id] = [makeEvent(task.id, "created", { to: task.status })];
                }
            }
            return {
                project: snapshot.project,
                features: snapshot.features,
                tasks: snapshot.tasks,
                plans: snapshot.plans,
                reviews: snapshot.reviews,
                memoryNodes: snapshot.memoryNodes,
                sessions: snapshot.sessions,
                skills: snapshot.skills,
                skillRuns: snapshot.skillRuns,
                taskHistory,
                loadError: "",
                selectedTaskId: selectedExists
                    ? state.selectedTaskId
                    : snapshot.tasks[0]?.id ?? "",
            };
        }),
    setProject: (project) => set({ project }),
    setAgentProviders: (agentProviders) =>
        set({ agentProviders: applyProviderOverrides(agentProviders, loadProviderOverrides()) }),
    updateAgentProvider: (id, patch) =>
        set((state) => {
            if (patch.enabled !== undefined) saveProviderOverride(id, { enabled: patch.enabled });
            return {
                agentProviders: state.agentProviders.map((provider) =>
                    provider.id === id ? { ...provider, ...patch } : provider,
                ),
            };
        }),
    setWorkflowSteps: (workflowSteps) => {
        saveWorkflowSteps(workflowSteps);
        set({ workflowSteps });
    },
    updateWorkflowStep: (id, patch) =>
        set((state) => {
            const workflowSteps = state.workflowSteps.map((step) =>
                step.id === id ? { ...step, ...patch } : step,
            );
            saveWorkflowSteps(workflowSteps);
            return { workflowSteps };
        }),
    selectTask: (taskId) => set({ selectedTaskId: taskId }),
    setBoardFilter: (boardFilter) => set({ boardFilter }),
    setInspectorTab: (inspectorTab) => set({ inspectorTab }),
    setBottomTab: (bottomTab) => set({ bottomTab }),
    openCreateTask: () => set({ taskDialog: { mode: "create" } }),
    openEditTask: (taskId) => set({ taskDialog: { mode: "edit", taskId } }),
    closeTaskDialog: () => set({ taskDialog: null }),
    createTask: (input) =>
        set((state) => {
            const id = `T-${100 + state.tasks.length}`;
            const task: Task = {
                id,
                featureId: state.features[0]?.id ?? "local",
                title: input.title,
                description: input.description,
                status: "backlog",
                priority: input.priority,
                assignedAgent: input.assignedAgent || "Unassigned",
                executorProfile: input.assignedAgent.toLowerCase().includes("claude") ? "claude-local" : "codex-local",
                worktreePath: "",
                branchName: "",
                planApproved: false,
                reviewApproved: false,
                testsPassed: false,
                updatedAt: nowIso(),
                tags: input.tags?.length ? input.tags : ["new"],
                progress: 0,
            };
            return {
                tasks: [...state.tasks, task],
                selectedTaskId: id,
                taskDialog: null,
                taskHistory: pushEvent(state.taskHistory, makeEvent(id, "created", { to: "backlog" })),
            };
        }),
    editTask: (taskId, input) =>
        set((state) => ({
            tasks: state.tasks.map((task) =>
                task.id === taskId
                    ? {
                          ...task,
                          title: input.title,
                          description: input.description,
                          priority: input.priority,
                          assignedAgent: input.assignedAgent || "Unassigned",
                          updatedAt: "now",
                      }
                    : task,
            ),
            taskDialog: null,
        })),
    removeTask: (taskId) =>
        set((state) => {
            const tasks = state.tasks.filter((task) => task.id !== taskId);
            const taskHistory = { ...state.taskHistory };
            delete taskHistory[taskId];
            const planningByTask = { ...state.planningByTask };
            delete planningByTask[taskId];
            return {
                tasks,
                plans: state.plans.filter((plan) => plan.taskId !== taskId),
                reviews: state.reviews.filter((review) => review.taskId !== taskId),
                sessions: state.sessions.filter((session) => session.taskId !== taskId),
                skillRuns: state.skillRuns.filter((run) => run.taskId !== taskId),
                taskHistory,
                planningByTask,
                selectedTaskId: state.selectedTaskId === taskId ? tasks[0]?.id ?? "" : state.selectedTaskId,
            };
        }),
    // Board drag/drop. Invalid transitions are silently rejected (no mutation).
    moveTask: (taskId, status) =>
        set((state) => {
            const outcome = applyTransition(state.tasks, taskId, status);
            if ("error" in outcome) return {};
            return { tasks: outcome.tasks, taskHistory: pushEvent(state.taskHistory, outcome.event) };
        }),
    // Explicit mock advance from the workflow controls; carries an optional note.
    advanceTask: (taskId, status, note) =>
        set((state) => {
            const outcome = applyTransition(state.tasks, taskId, status, note);
            if ("error" in outcome) return {};
            return { tasks: outcome.tasks, taskHistory: pushEvent(state.taskHistory, outcome.event) };
        }),
    approvePlan: (taskId) =>
        set((state) => {
            const plans = state.plans.map((plan) =>
                plan.taskId === taskId ? { ...plan, approvalStatus: "approved" as const } : plan,
            );
            // Plan approval is the gate for `ready` — it is distinct from review approval.
            const flagged = state.tasks.map((task) =>
                task.id === taskId ? { ...task, planApproved: true } : task,
            );
            const outcome = applyTransition(flagged, taskId, "ready");
            if ("error" in outcome) return { plans, tasks: flagged };
            return { plans, tasks: outcome.tasks, taskHistory: pushEvent(state.taskHistory, outcome.event) };
        }),
    persistPlan: (plan) =>
        set((state) => ({
            plans: state.plans.some((item) => item.taskId === plan.taskId)
                ? state.plans.map((item) =>
                      item.taskId === plan.taskId ? plan : item,
                  )
                : [...state.plans, plan],
        })),
    requestChanges: (taskId) =>
        set((state) => {
            const outcome = applyTransition(state.tasks, taskId, "planning", "Changes requested");
            if ("error" in outcome) return {};
            const event: TaskEvent = { ...outcome.event, type: "changes_requested" };
            return { tasks: outcome.tasks, taskHistory: pushEvent(state.taskHistory, event) };
        }),
    // Phase 8 — Review. Approving persists an approved review (which flags
    // reviewApproved + records the review_approved event via persistReview). The
    // task is not auto-finished — Done still requires the state-machine gate.
    approveReview: (taskId) => {
        get().persistReview({ ...buildReviewDraft(get(), taskId), status: "approved", reviewerNotes: "Approved from the review panel." });
    },
    // Rejecting clears approval and sends the task to Failed.
    rejectReview: (taskId) => {
        get().persistReview({ ...buildReviewDraft(get(), taskId), status: "rejected", reviewerNotes: "Rejected from the review panel." });
        set((state) => {
            const outcome = applyTransition(state.tasks, taskId, "failed", "Review rejected");
            if ("error" in outcome) return {};
            return { tasks: outcome.tasks, taskHistory: pushEvent(state.taskHistory, outcome.event) };
        });
    },
    // Request Changes clears approval and sends the task back to the coder (running).
    requestReviewChanges: (taskId) => {
        get().persistReview({ ...buildReviewDraft(get(), taskId), status: "changes_requested", reviewerNotes: "Changes requested from the review panel." });
        set((state) => {
            const outcome = applyTransition(state.tasks, taskId, "running", "Review changes requested");
            if ("error" in outcome) return {};
            const event: TaskEvent = { ...outcome.event, type: "changes_requested" };
            return { tasks: outcome.tasks, taskHistory: pushEvent(state.taskHistory, event) };
        });
    },
    // Phase 9 — Memory. Adds (or replaces) a memory node; newest first.
    appendMemory: (node) =>
        set((state) => ({
            memoryNodes: state.memoryNodes.some((item) => item.id === node.id)
                ? state.memoryNodes.map((item) => (item.id === node.id ? node : item))
                : [node, ...state.memoryNodes],
        })),
    runTests: (taskId) =>
        set((state) => {
            if (!state.tasks.some((task) => task.id === taskId)) return {};
            return {
                tasks: state.tasks.map((task) =>
                    task.id === taskId ? { ...task, testsPassed: true, updatedAt: nowIso() } : task,
                ),
                taskHistory: pushEvent(state.taskHistory, makeEvent(taskId, "tests_passed", {})),
            };
        }),
    approveMerge: (taskId) =>
        set((state) => {
            const outcome = applyTransition(state.tasks, taskId, "done");
            if ("error" in outcome) return {};
            return { tasks: outcome.tasks, taskHistory: pushEvent(state.taskHistory, outcome.event) };
        }),
    persistDiff: (diff) =>
        set((state) => ({ diffs: { ...state.diffs, [diff.taskId]: diff } })),
    persistTestRun: (test) =>
        set((state) => ({
            testRuns: { ...state.testRuns, [test.taskId]: test },
            tasks: state.tasks.map((task) =>
                task.id === test.taskId
                    ? { ...task, testsPassed: test.status === "passed" }
                    : task,
            ),
        })),
    persistReview: (review) =>
        set((state) => {
            const wasApproved = state.tasks.find((task) => task.id === review.taskId)?.reviewApproved ?? false;
            const nowApproved = review.status === "approved";
            const taskHistory = !wasApproved && nowApproved
                ? pushEvent(state.taskHistory, makeEvent(review.taskId, "review_approved", {}))
                : state.taskHistory;
            return {
                reviews: state.reviews.some((item) => item.taskId === review.taskId)
                    ? state.reviews.map((item) => (item.taskId === review.taskId ? review : item))
                    : [...state.reviews, review],
                tasks: state.tasks.map((task) =>
                    task.id === review.taskId ? { ...task, reviewApproved: nowApproved } : task,
                ),
                taskHistory,
            };
        }),
    persistMemory: (nodes) => set({ memoryNodes: nodes }),
    setPlanningQuestions: (taskId, questions) =>
        set((state) => ({ planningByTask: { ...state.planningByTask, [taskId]: { questions } } })),
    answerPlanningQuestion: (taskId, questionId, answer) =>
        set((state) => {
            const existing = state.planningByTask[taskId];
            if (!existing) return {};
            return {
                planningByTask: {
                    ...state.planningByTask,
                    [taskId]: {
                        questions: existing.questions.map((question) =>
                            question.id === questionId ? { ...question, answer } : question,
                        ),
                    },
                },
            };
        }),
    // Sessions are keyed by id so a task can own multiple terminal sessions.
    upsertSession: (session) =>
        set((state) => {
            const index = state.sessions.findIndex((item) => item.id === session.id);
            if (index < 0) return { sessions: [...state.sessions, session] };
            const next = [...state.sessions];
            next[index] = session;
            return { sessions: next };
        }),
    appendSessionOutput: (_taskId, sessionId, output) =>
        set((state) => ({
            sessions: state.sessions.map((session) =>
                session.id === sessionId ? { ...session, output: [...session.output, output] } : session,
            ),
        })),
    patchSession: (sessionId, patch) =>
        set((state) => ({
            sessions: state.sessions.map((session) =>
                session.id === sessionId ? { ...session, ...patch } : session,
            ),
        })),
    removeSession: (sessionId) =>
        set((state) => ({
            sessions: state.sessions.filter((session) => session.id !== sessionId),
            pinnedSessions: state.pinnedSessions.filter((id) => id !== sessionId),
        })),
    togglePinnedSession: (sessionId) =>
        set((state) => ({
            pinnedSessions: state.pinnedSessions.includes(sessionId)
                ? state.pinnedSessions.filter((id) => id !== sessionId)
                : [...state.pinnedSessions, sessionId],
        })),
    setActiveSession: (taskId, sessionId) =>
        set((state) => ({ activeSessionByTask: { ...state.activeSessionByTask, [taskId]: sessionId } })),
}));

function step(
    id: WorkflowStepId,
    label: string,
    provider: string,
    command: string,
    workingDirectoryMode: WorkflowStepConfig["workingDirectoryMode"],
    autoStartTerminal: boolean,
    approvalRequired: boolean,
    permissions: string[],
): WorkflowStepConfig {
    return {
        id,
        label,
        enabled: true,
        provider,
        command,
        args: [],
        workingDirectoryMode,
        autoStartTerminal,
        approvalRequired,
        env: {},
        timeout: "30m",
        permissions,
    };
}

export function taskEventsFor(state: WorkbenchState, taskId: string): TaskEvent[] {
    return state.taskHistory[taskId] ?? [];
}

export function planningFor(state: WorkbenchState, taskId: string): PlanningQuestion[] {
    return state.planningByTask[taskId]?.questions ?? [];
}

export function selectedTask(state: WorkbenchState): Task | null {
    return (
        state.tasks.find((task) => task.id === state.selectedTaskId) ??
        state.tasks[0] ??
        null
    );
}

export function planFor(task: Task, state: WorkbenchState) {
    return (
        state.plans.find((plan) => plan.taskId === task.id) ?? {
            id: `plan-${task.id}`,
            taskId: task.id,
            summary: "No execution plan has been saved for this task.",
            steps: [],
            risks: [],
            filesToTouch: [],
            testStrategy: [],
            approvalStatus: "draft" as const,
        }
    );
}

// Builds a review draft for a task from the existing review plus any collected
// diff/test artifacts, so an approve/reject action carries real context.
function buildReviewDraft(state: WorkbenchState, taskId: string): Review {
    const existing = state.reviews.find((review) => review.taskId === taskId);
    const diff = state.diffs[taskId];
    const test = state.testRuns[taskId];
    return {
        id: existing?.id ?? `review-${taskId}`,
        taskId,
        diffSummary: existing?.diffSummary || diff?.summary || "",
        changedFiles: existing?.changedFiles.length ? existing.changedFiles : diff?.changedFiles.map((file) => file.path) ?? [],
        testResults: existing?.testResults.length
            ? existing.testResults
            : test
              ? [{ command: test.command, status: test.status, output: test.stdout || test.stderr }]
              : [],
        reviewerNotes: existing?.reviewerNotes ?? "",
        status: existing?.status ?? "pending",
    };
}

export function reviewFor(task: Task, state: WorkbenchState) {
    return (
        state.reviews.find((review) => review.taskId === task.id) ?? {
            id: `review-${task.id}`,
            taskId: task.id,
            diffSummary: "",
            changedFiles: [],
            testResults: [],
            reviewerNotes: "",
            status: "pending" as const,
        }
    );
}

export function sessionFor(task: Task, state: WorkbenchState): AgentSession {
    return (
        state.sessions.find((session) => session.taskId === task.id) ?? {
            id: `session-${task.id}`,
            taskId: task.id,
            agentType: "coder",
            provider: "codex",
            command: "codex",
            status: "idle",
            output: [],
        }
    );
}

export function workflowStepFor(state: WorkbenchState, id: WorkflowStepId) {
    return state.workflowSteps.find((step) => step.id === id) ?? defaultWorkflowSteps.find((step) => step.id === id)!;
}

export function currentWorkflowStep(task: Task): WorkflowStepId {
    switch (task.status) {
        case "backlog":
        case "planning":
        case "waiting_approval":
            return "planning";
        case "ready":
        case "running":
            return "coding";
        case "in_review":
            return "review";
        case "failed":
        case "blocked":
        case "waiting_user":
            return "debugging";
        case "done":
            return "memory_update";
    }
}

export function activeSkillsFor(task: Task, state: WorkbenchState) {
    return state.skillRuns
        .filter((run) => run.taskId === task.id)
        .map((run) => ({
            run,
            skill: state.skills.find((skill) => skill.id === run.skillId),
        }))
        .filter((item): item is { run: SkillRun; skill: Skill } =>
            Boolean(item.skill),
        );
}
