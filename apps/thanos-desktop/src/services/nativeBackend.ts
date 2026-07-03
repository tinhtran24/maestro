import { invoke } from "@tauri-apps/api/core";
import { listen } from "@tauri-apps/api/event";
import type {
    AgentSession,
    ExecutionPlan,
    Feature,
    GitDiff,
    MemoryNode,
    Project,
    Review,
    Skill,
    SkillRun,
    Task,
    TestRun,
} from "../domain/models";

type WorktreeInfo = {
  task_id: string;
  branch_name: string;
  worktree_path: string;
  created: boolean;
};

type AgentSessionInfo = {
  id: string;
  task_id: string;
  agent_type: AgentSession["agentType"];
  provider: string;
  command: string;
  args: string[];
  status: AgentSession["status"];
  pty_session_id: string;
  conversation_log_path: string;
  worktree_path: string;
};

type OutputPayload = {
  session_id: string;
  task_id: string;
  data: string;
};

type ExitPayload = {
  session_id: string;
  task_id: string;
  code: number;
};

type ExecutionPlanInfo = {
  id: string;
  task_id: string;
  summary: string;
  steps: Array<{ id: string; title: string; description: string; status: string }>;
  risks: string[];
  files_to_touch: string[];
  test_strategy: string[];
  approval_status: ExecutionPlan["approvalStatus"];
};

type GitDiffInfo = {
    task_id: string;
    summary: string;
    changed_files: Array<{ path: string; status: string }>;
    patch: string;
};

type TestRunInfo = {
    task_id: string;
    command: string;
    status: "passed" | "failed";
    stdout: string;
    stderr: string;
    code: number | null;
};

type ReviewInfo = {
    id: string;
    task_id: string;
    diff_summary: string;
    changed_files: string[];
    test_results: TestRunInfo[];
    reviewer_notes: string;
    status: Review["status"];
};

type MemoryNodeInfo = {
    id: string;
    project_id: string;
    node_type: MemoryNode["type"];
    title: string;
    content: string;
    links: string[];
    created_at: number;
};

export type AgentBridgeTool = {
    name: string;
    description: string;
    inputs: string[];
};

export type BridgeTaskMessage = {
    id: string;
    sourceTaskId: string;
    targetTaskId: string;
    content: string;
    createdAt: string;
};

export type BridgeRelatedWork = {
    type: string;
    id: string;
    title: string;
    summary: string;
    path: string;
};

export type BridgeBranchAttachment = {
    taskId: string;
    branchName: string;
    worktreePath: string;
    attachedAt: string;
};

export type BridgeUserReviewRequest = {
    id: string;
    taskId: string;
    requestedBy: string;
    notes: string;
    status: "pending_user_review";
    createdAt: string;
};

type BridgeTaskMessageInfo = {
    id: string;
    source_task_id: string;
    target_task_id: string;
    content: string;
    created_at: number;
};

type BridgeRelatedWorkInfo = {
    item_type: string;
    id: string;
    title: string;
    summary: string;
    path: string;
};

type BridgeBranchAttachmentInfo = {
    task_id: string;
    branch_name: string;
    worktree_path: string;
    attached_at: number;
};

type BridgeUserReviewRequestInfo = {
    id: string;
    task_id: string;
    requested_by: string;
    notes: string;
    status: "pending_user_review";
    created_at: number;
};

type ProjectInfo = {
    id: string;
    name: string;
    root_path: string;
    repos: string[];
    settings: Record<string, string>;
};

type FeatureInfo = {
    id: string;
    project_id: string;
    title: string;
    description: string;
    status: Feature["status"];
    plan_graph_id?: string | null;
    created_at: string;
};

type TaskInfo = {
    id: string;
    feature_id: string;
    parent_task_id?: string | null;
    title: string;
    description: string;
    status: Task["status"];
    priority: Task["priority"];
    assigned_agent: string;
    executor_profile: string;
    worktree_path: string;
    branch_name: string;
    review_approved: boolean;
    tests_passed: boolean;
    updated_at: string;
    tags: string[];
    progress: number;
};

type SkillInfo = {
    id: string;
    project_id?: string | null;
    name: string;
    path: string;
    description: string;
    applies_to: string[];
    agents: Skill["agents"];
    version?: string | null;
    source: Skill["source"];
    required_evidence: string[];
    exit_criteria: string[];
    enabled: boolean;
    trusted: boolean;
};

type SkillRunInfo = {
    id: string;
    task_id: string;
    skill_id: string;
    agent_session_id?: string | null;
    status: SkillRun["status"];
    evidence_json: Record<string, string>;
    started_at: string;
    completed_at?: string | null;
};

export type WorkbenchSnapshot = {
    project: Project;
    features: Feature[];
    tasks: Task[];
    plans: ExecutionPlan[];
    sessions: AgentSession[];
    reviews: Review[];
    memoryNodes: MemoryNode[];
    skills: Skill[];
    skillRuns: SkillRun[];
};

type WorkbenchSnapshotInfo = {
    project: ProjectInfo;
    features: FeatureInfo[];
    tasks: TaskInfo[];
    plans: ExecutionPlanInfo[];
    sessions: AgentSessionInfo[];
    reviews: ReviewInfo[];
    memory_nodes: MemoryNodeInfo[];
    skills?: SkillInfo[];
    skill_runs?: SkillRunInfo[];
};

export class NativeBackend {
  private readonly workspace = "/Users/tinh.tran/Tool/thanos";

  async loadWorkbenchState() {
    const info = await this.tryInvoke<WorkbenchSnapshotInfo>("load_workbench_state", {
      workspace: this.workspace,
    });
    return info ? fromWorkbenchSnapshotInfo(info) : null;
  }

  async prepareWorktree(task: Task) {
    const branchName = task.branchName || `thanos/${task.id.toLowerCase()}-${slug(task.title)}`;
    const info = await this.tryInvoke<WorktreeInfo>("prepare_task_worktree", {
      request: {
        workspace: this.workspace,
        task_id: task.id,
        branch_name: branchName,
        base_ref: "HEAD",
      },
    });
    if (!info) return null;
    return {
      branchName: info.branch_name,
      worktreePath: info.worktree_path,
      created: info.created,
    };
  }

  async startAgent(task: Task) {
    return this.startAgentRole(task, "coder");
  }

  async startAgentRole(task: Task, agentType: AgentSession["agentType"]) {
    const planner = agentType === "planner";
    const command = planner ? "claude" : task.executorProfile.includes("claude") ? "claude" : "codex";
    const info = await this.tryInvoke<AgentSessionInfo>("start_agent_session", {
      request: {
        workspace: this.workspace,
        task_id: task.id,
        agent_type: agentType,
        provider: planner ? "claude-code" : task.executorProfile.includes("claude") ? "claude-code" : "codex",
        command,
        args: [],
        worktree_path: planner ? "." : task.worktreePath,
      },
    });
    return info ? mapSession(info) : null;
  }

  async saveExecutionPlan(plan: ExecutionPlan) {
    const info = await this.tryInvoke<ExecutionPlanInfo>("save_execution_plan", {
      request: {
        workspace: this.workspace,
        plan: toPlanInfo(plan),
      },
    });
    return info ? fromPlanInfo(info) : null;
  }

  async approveExecutionPlan(taskId: string) {
    const info = await this.tryInvoke<ExecutionPlanInfo>("approve_execution_plan", {
      request: {
        workspace: this.workspace,
        task_id: taskId,
      },
    });
    return info ? fromPlanInfo(info) : null;
  }

  async readExecutionPlan(taskId: string) {
    const info = await this.tryInvoke<ExecutionPlanInfo>("read_execution_plan", {
      request: {
        workspace: this.workspace,
        task_id: taskId,
      },
    });
    return info ? fromPlanInfo(info) : null;
  }

  async collectGitDiff(task: Task) {
    const info = await this.tryInvoke<GitDiffInfo>("collect_git_diff", {
      request: {
        workspace: this.workspace,
        task_id: task.id,
        worktree_path: task.worktreePath,
      },
    });
    return info ? fromDiffInfo(info) : null;
  }

  async runTaskTests(task: Task, command = "go test ./...") {
    const info = await this.tryInvoke<TestRunInfo>("run_task_tests", {
      request: {
        workspace: this.workspace,
        task_id: task.id,
        worktree_path: task.worktreePath,
        command,
      },
    });
    return info ? fromTestInfo(info) : null;
  }

  async saveReview(review: Review) {
    const info = await this.tryInvoke<ReviewInfo>("save_review", {
      request: {
        workspace: this.workspace,
        review: toReviewInfo(review),
      },
    });
    return info ? fromReviewInfo(info) : null;
  }

  async approveReview(taskId: string) {
    const info = await this.tryInvoke<ReviewInfo>("approve_review", {
      request: {
        workspace: this.workspace,
        task_id: taskId,
      },
    });
    return info ? fromReviewInfo(info) : null;
  }

  async writeMemoryNode(node: MemoryNode) {
    const info = await this.tryInvoke<MemoryNodeInfo>("write_memory_node", {
      request: {
        workspace: this.workspace,
        node: toMemoryInfo(node),
      },
    });
    return info ? fromMemoryInfo(info) : null;
  }

  async searchMemory(query: string) {
    const info = await this.tryInvoke<MemoryNodeInfo[]>("search_memory", {
      request: {
        workspace: this.workspace,
        query,
      },
    });
    return info ? info.map(fromMemoryInfo) : [];
  }

  async listAgentBridgeTools() {
    return await this.tryInvoke<AgentBridgeTool[]>("list_agent_bridge_tools", {}) ?? [];
  }

  async createAgentSubtask(parentTaskId: string, title: string, description = "", priority = "P2", agent = "") {
    const info = await this.tryInvoke<TaskInfo>("bridge_create_subtask", {
      request: {
        workspace: this.workspace,
        parent_task_id: parentTaskId,
        title,
        description,
        priority,
        agent,
      },
    });
    return info ? fromTaskInfo(info) : null;
  }

  async messageSiblingTask(sourceTaskId: string, targetTaskId: string, content: string) {
    const info = await this.tryInvoke<BridgeTaskMessageInfo>("bridge_message_sibling", {
      request: {
        workspace: this.workspace,
        source_task_id: sourceTaskId,
        target_task_id: targetTaskId,
        content,
      },
    });
    return info ? fromBridgeTaskMessageInfo(info) : null;
  }

  async inspectRelatedWork(query: string, taskId?: string, limit = 10) {
    const info = await this.tryInvoke<BridgeRelatedWorkInfo[]>("bridge_inspect_related_work", {
      request: {
        workspace: this.workspace,
        task_id: taskId,
        query,
        limit,
      },
    });
    return info ? info.map(fromBridgeRelatedWorkInfo) : [];
  }

  async attachTaskBranch(taskId: string, branchName: string, worktreePath?: string) {
    const info = await this.tryInvoke<BridgeBranchAttachmentInfo>("bridge_attach_branch", {
      request: {
        workspace: this.workspace,
        task_id: taskId,
        branch_name: branchName,
        worktree_path: worktreePath,
      },
    });
    return info ? fromBridgeBranchAttachmentInfo(info) : null;
  }

  async requestUserReview(taskId: string, requestedBy = "agent", notes = "") {
    const info = await this.tryInvoke<BridgeUserReviewRequestInfo>("bridge_request_user_review", {
      request: {
        workspace: this.workspace,
        task_id: taskId,
        requested_by: requestedBy,
        notes,
      },
    });
    return info ? fromBridgeUserReviewRequestInfo(info) : null;
  }

  async stopAgent() {
    await this.tryInvoke<void>("stop_agent_session", {});
  }

  async resumeAgent(sessionId: string) {
    const info = await this.tryInvoke<AgentSessionInfo>("resume_agent_session", {
      workspace: this.workspace,
      session_id: sessionId,
    });
    return info ? mapSession(info) : null;
  }

  async onAgentOutput(handler: (taskId: string, sessionId: string, data: string) => void) {
    return listen<OutputPayload>("agent-session-output", (event) => {
      handler(event.payload.task_id, event.payload.session_id, event.payload.data);
    });
  }

  async onAgentExit(handler: (taskId: string, sessionId: string, code: number) => void) {
    return listen<ExitPayload>("agent-session-exit", (event) => {
      handler(event.payload.task_id, event.payload.session_id, event.payload.code);
    });
  }

  private async tryInvoke<T>(command: string, payload: Record<string, unknown>) {
    try {
      return await invoke<T>(command, payload);
    } catch {
      return null;
    }
  }
}

function mapSession(info: AgentSessionInfo): AgentSession {
  return {
    id: info.id,
    taskId: info.task_id,
    agentType: info.agent_type,
    provider: info.provider,
    command: [info.command, ...info.args].join(" "),
    status: mapSessionStatus(info.status),
    ptySessionId: info.pty_session_id,
    conversationLogPath: info.conversation_log_path,
    output: [],
  };
}

function fromWorkbenchSnapshotInfo(info: WorkbenchSnapshotInfo): WorkbenchSnapshot {
  return {
    project: {
      id: info.project.id,
      name: info.project.name,
      rootPath: info.project.root_path,
      repos: info.project.repos,
      settings: info.project.settings,
    },
    features: info.features.map((feature) => ({
      id: feature.id,
      projectId: feature.project_id,
      title: feature.title,
      description: feature.description,
      status: feature.status,
      planGraphId: feature.plan_graph_id ?? undefined,
      createdAt: feature.created_at,
    })),
    tasks: info.tasks.map(fromTaskInfo),
    plans: info.plans.map(fromPlanInfo),
    sessions: info.sessions.map(mapSession),
    reviews: info.reviews.map(fromReviewInfo),
    memoryNodes: info.memory_nodes.map(fromMemoryInfo),
    skills: (info.skills ?? []).map(fromSkillInfo),
    skillRuns: (info.skill_runs ?? []).map(fromSkillRunInfo),
  };
}

function fromTaskInfo(info: TaskInfo): Task {
  return {
    id: info.id,
    featureId: info.feature_id,
    parentTaskId: info.parent_task_id ?? undefined,
    title: info.title,
    description: info.description,
    status: info.status,
    priority: info.priority,
    assignedAgent: info.assigned_agent,
    executorProfile: info.executor_profile,
    worktreePath: info.worktree_path,
    branchName: info.branch_name,
    reviewApproved: info.review_approved,
    testsPassed: info.tests_passed,
    updatedAt: info.updated_at,
    tags: info.tags,
    progress: info.progress,
  };
}

function mapSessionStatus(status: string): AgentSession["status"] {
  if (status === "starting" || status === "running" || status === "stopping" || status === "stopped" || status === "failed") {
    return status;
  }
  if (status === "exited" || status === "complete" || status === "completed") {
    return "stopped";
  }
  return "idle";
}

function toPlanInfo(plan: ExecutionPlan): ExecutionPlanInfo {
  return {
    id: plan.id,
    task_id: plan.taskId,
    summary: plan.summary,
    steps: plan.steps,
    risks: plan.risks,
    files_to_touch: plan.filesToTouch,
    test_strategy: plan.testStrategy,
    approval_status: plan.approvalStatus,
  };
}

function fromPlanInfo(info: ExecutionPlanInfo): ExecutionPlan {
  return {
    id: info.id,
    taskId: info.task_id,
    summary: info.summary,
    steps: info.steps,
    risks: info.risks,
    filesToTouch: info.files_to_touch,
    testStrategy: info.test_strategy,
    approvalStatus: info.approval_status,
  };
}

function fromDiffInfo(info: GitDiffInfo): GitDiff {
  return {
    taskId: info.task_id,
    summary: info.summary,
    changedFiles: info.changed_files,
    patch: info.patch,
  };
}

function fromTestInfo(info: TestRunInfo): TestRun {
  return {
    taskId: info.task_id,
    command: info.command,
    status: info.status,
    stdout: info.stdout,
    stderr: info.stderr,
    code: info.code,
  };
}

function toReviewInfo(review: Review): ReviewInfo {
  return {
    id: review.id,
    task_id: review.taskId,
    diff_summary: review.diffSummary,
    changed_files: review.changedFiles,
    test_results: review.testResults.map((item) => ({
      task_id: review.taskId,
      command: item.command,
      status: item.status === "passed" ? "passed" : "failed",
      stdout: item.output,
      stderr: "",
      code: item.status === "passed" ? 0 : 1,
    })),
    reviewer_notes: review.reviewerNotes,
    status: review.status,
  };
}

function fromReviewInfo(info: ReviewInfo): Review {
  return {
    id: info.id,
    taskId: info.task_id,
    diffSummary: info.diff_summary,
    changedFiles: info.changed_files,
    testResults: info.test_results.map((item) => ({ command: item.command, status: item.status, output: item.stdout || item.stderr })),
    reviewerNotes: info.reviewer_notes,
    status: info.status,
  };
}

function toMemoryInfo(node: MemoryNode): MemoryNodeInfo {
  return {
    id: node.id,
    project_id: node.projectId,
    node_type: node.type,
    title: node.title,
    content: node.content,
    links: node.links,
    created_at: Date.parse(node.createdAt) || Math.floor(Date.now() / 1000),
  };
}

function fromMemoryInfo(info: MemoryNodeInfo): MemoryNode {
  return {
    id: info.id,
    projectId: info.project_id,
    type: info.node_type,
    title: info.title,
    content: info.content,
    links: info.links,
    createdAt: new Date(info.created_at * 1000).toISOString(),
  };
}

function fromBridgeTaskMessageInfo(info: BridgeTaskMessageInfo): BridgeTaskMessage {
  return {
    id: info.id,
    sourceTaskId: info.source_task_id,
    targetTaskId: info.target_task_id,
    content: info.content,
    createdAt: new Date(info.created_at * 1000).toISOString(),
  };
}

function fromBridgeRelatedWorkInfo(info: BridgeRelatedWorkInfo): BridgeRelatedWork {
  return {
    type: info.item_type,
    id: info.id,
    title: info.title,
    summary: info.summary,
    path: info.path,
  };
}

function fromBridgeBranchAttachmentInfo(info: BridgeBranchAttachmentInfo): BridgeBranchAttachment {
  return {
    taskId: info.task_id,
    branchName: info.branch_name,
    worktreePath: info.worktree_path,
    attachedAt: new Date(info.attached_at * 1000).toISOString(),
  };
}

function fromBridgeUserReviewRequestInfo(info: BridgeUserReviewRequestInfo): BridgeUserReviewRequest {
  return {
    id: info.id,
    taskId: info.task_id,
    requestedBy: info.requested_by,
    notes: info.notes,
    status: info.status,
    createdAt: new Date(info.created_at * 1000).toISOString(),
  };
}

function fromSkillInfo(info: SkillInfo): Skill {
  return {
    id: info.id,
    projectId: info.project_id ?? undefined,
    name: info.name,
    path: info.path,
    description: info.description,
    appliesTo: info.applies_to,
    agents: info.agents,
    version: info.version ?? undefined,
    source: info.source,
    requiredEvidence: info.required_evidence,
    exitCriteria: info.exit_criteria,
    enabled: info.enabled,
    trusted: info.trusted,
  };
}

function fromSkillRunInfo(info: SkillRunInfo): SkillRun {
  return {
    id: info.id,
    taskId: info.task_id,
    skillId: info.skill_id,
    agentSessionId: info.agent_session_id ?? undefined,
    status: info.status,
    evidence: info.evidence_json,
    startedAt: info.started_at,
    completedAt: info.completed_at ?? undefined,
  };
}

function slug(value: string) {
  return value.toLowerCase().replace(/[^a-z0-9]+/g, "-").replace(/^-|-$/g, "");
}
