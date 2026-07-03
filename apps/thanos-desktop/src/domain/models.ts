export type TaskStatus =
    | "backlog"
    | "planning"
    | "waiting_approval"
    | "ready"
    | "running"
    | "in_review"
    | "waiting_user"
    | "blocked"
    | "done"
    | "failed";

export type Priority = "P0" | "P1" | "P2" | "P3";
export type InspectorTab = "overview" | "plan" | "files" | "changes" | "browser" | "tests" | "memory";
export type BottomTab = "terminal" | "timeline" | "logs" | "chat";

export type Project = {
    id: string;
    name: string;
    rootPath: string;
    gitRemoteUrl?: string;
    defaultBranch?: string;
    worktreeRoot?: string;
    packageManager?: string;
    devCommand?: string;
    testCommand?: string;
    createdAt?: string;
    updatedAt?: string;
    lastOpenedAt?: string;
    archived?: boolean;
    envVars?: Record<string, string>;
    repos: string[];
    settings: Record<string, string>;
};

export type Feature = {
    id: string;
    projectId: string;
    title: string;
    description: string;
    status: "backlog" | "active" | "done";
    planGraphId?: string;
    createdAt: string;
};

export type Task = {
    id: string;
    featureId: string;
    parentTaskId?: string;
    title: string;
    description: string;
    status: TaskStatus;
    priority: Priority;
    assignedAgent: string;
    executorProfile: string;
    worktreePath: string;
    branchName: string;
    planApproved?: boolean;
    reviewApproved: boolean;
    testsPassed: boolean;
    updatedAt: string;
    tags: string[];
    progress: number;
};

// Phase 3 — Workflow Engine. A recorded workflow transition / lifecycle event.
export type TaskEventType =
    | "created"
    | "moved"
    | "plan_approved"
    | "changes_requested"
    | "agent_started"
    | "tests_passed"
    | "review_approved"
    | "blocked"
    | "failed"
    | "done";

export type TaskEvent = {
    id: string;
    taskId: string;
    type: TaskEventType;
    from?: TaskStatus;
    to?: TaskStatus;
    note?: string;
    at: string;
};

// Phase 6 — Planner Workflow. A clarifying question posed by the planner and
// the user's answer, captured before an execution plan is generated.
export type PlanningQuestion = {
    id: string;
    prompt: string;
    answer: string;
};

export type ExecutionPlan = {
    id: string;
    taskId: string;
    summary: string;
    steps: Array<{
        id: string;
        title: string;
        description: string;
        status: string;
    }>;
    risks: string[];
    filesToTouch: string[];
    testStrategy: string[];
    approvalStatus:
        "draft" | "pending" | "approved" | "rejected" | "changes_requested";
};

export type AgentSession = {
    id: string;
    taskId: string;
    step?: WorkflowStepId;
    agentType: "planner" | "coder" | "reviewer" | "tester" | "utility";
    provider: string;
    command: string;
    cwd?: string;
    status: "idle" | "starting" | "running" | "waiting_user" | "completed" | "stopping" | "stopped" | "failed";
    ptySessionId?: string;
    conversationLogPath?: string;
    transcriptPath?: string;
    startedAt?: string;
    endedAt?: string;
    output: string[];
};

export type AgentProvider = {
    id: string;
    name: string;
    command: string;
    detectedPath?: string;
    status: "installed" | "not_found" | "needs_setup";
    version?: string;
    type: "cli" | "mcp" | "acp" | "shell";
    enabled: boolean;
    setupHint?: string;
};

export type WorkflowStepId =
    | "planning"
    | "coding"
    | "review"
    | "testing"
    | "debugging"
    | "documentation"
    | "memory_update";

export type WorkflowStepConfig = {
    id: WorkflowStepId;
    label: string;
    enabled: boolean;
    provider: string;
    command: string;
    args: string[];
    workingDirectoryMode: "project" | "worktree" | "custom";
    autoStartTerminal: boolean;
    approvalRequired: boolean;
    env: Record<string, string>;
    timeout: string;
    permissions: string[];
};

export type SkillRunStatus =
    | "discovered"
    | "matched"
    | "activated"
    | "running"
    | "evidence_pending"
    | "completed"
    | "failed";

export type Skill = {
    id: string;
    projectId?: string;
    name: string;
    path: string;
    description: string;
    appliesTo: string[];
    agents: Array<AgentSession["agentType"]>;
    version?: string;
    source: "project" | "global" | "builtin";
    requiredEvidence: string[];
    exitCriteria: string[];
    enabled: boolean;
    trusted: boolean;
};

export type SkillRun = {
    id: string;
    taskId: string;
    skillId: string;
    agentSessionId?: string;
    status: SkillRunStatus;
    evidence: Record<string, string>;
    startedAt: string;
    completedAt?: string;
};

export type Review = {
    id: string;
    taskId: string;
    diffSummary: string;
    changedFiles: string[];
    testResults: Array<{ command: string; status: string; output: string }>;
    reviewerNotes: string;
    status: "pending" | "approved" | "rejected" | "changes_requested";
};

export type GitDiff = {
    taskId: string;
    summary: string;
    changedFiles: Array<{ path: string; status: string }>;
    patch: string;
};

export type TestRun = {
    taskId: string;
    command: string;
    status: "passed" | "failed";
    stdout: string;
    stderr: string;
    code: number | null;
};

export type MemoryNode = {
    id: string;
    projectId: string;
    type:
        | "feature"
        | "decision"
        | "architecture"
        | "file"
        | "task"
        | "bug"
        | "convention";
    title: string;
    content: string;
    links: string[];
    createdAt: string;
};

export type WorkbenchEvent =
    | { type: "onCreateTask"; title: string }
    | { type: "onSelectTask"; taskId: string }
    | { type: "onApprovePlan"; taskId: string }
    | { type: "onRequestChanges"; taskId: string; notes?: string }
    | { type: "onStartAgent"; taskId: string }
    | { type: "onStopAgent"; taskId: string }
    | { type: "onRunTests"; taskId: string }
    | { type: "onApproveMerge"; taskId: string };

export type FlowEvents = {
    onCreateTask(title: string): void;
    onSelectTask(taskId: string): void;
    onApprovePlan(taskId: string): void;
    onRequestChanges(taskId: string, notes?: string): void;
    onStartAgent(taskId: string): void;
    onStopAgent(taskId: string): void;
    onRunTests(taskId: string): void;
    onApproveMerge(taskId: string): void;
};
