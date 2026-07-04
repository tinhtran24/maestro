import { invoke } from "@tauri-apps/api/core";
import { listen } from "@tauri-apps/api/event";
import type {
    AgentSession,
    AgentProvider,
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

type AgentCandidateInfo = {
    id: string;
    name: string;
    command: string;
    installed: boolean;
    path?: string | null;
    status: "installed" | "not_found" | "needs_setup";
    version?: string | null;
    agent_type: "cli" | "mcp" | "acp" | "shell";
    enabled: boolean;
    setup_hint: string;
};

type ProjectSetupInfo = {
    root_path: string;
    name: string;
    git_remote_url?: string;
    default_branch?: string;
    worktree_root?: string;
    package_manager?: string;
    dev_command?: string;
    test_command?: string;
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
    git_remote_url?: string;
    default_branch?: string;
    worktree_root?: string;
    package_manager?: string;
    dev_command?: string;
    test_command?: string;
    created_at?: string;
    updated_at?: string;
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
  private readonly workspaceKey = "thanos.workspace";
  private readonly recentProjectsKey = "thanos.recentProjects";

  getWorkspacePath() {
    return localStorage.getItem(this.workspaceKey) || "";
  }

  setWorkspacePath(path: string) {
    localStorage.setItem(this.workspaceKey, path);
  }

  getRecentProjects() {
    try {
      return JSON.parse(localStorage.getItem(this.recentProjectsKey) || "[]") as Project[];
    } catch {
      return [];
    }
  }

  private rememberProject(project: Project) {
    const recent = [project, ...this.getRecentProjects().filter((item) => item.rootPath !== project.rootPath)].slice(0, 8);
    localStorage.setItem(this.recentProjectsKey, JSON.stringify(recent));
  }

  async loadWorkbenchState() {
    const workspace = this.getWorkspacePath();
    if (!workspace) return null;
    const info = await this.tryInvoke<WorkbenchSnapshotInfo>("load_workbench_state", {
      workspace,
    });
    return info ? fromWorkbenchSnapshotInfo(info) : null;
  }

  async selectWorkspaceFolder() {
    return await this.tryInvoke<string | null>("select_workspace_folder", {}) ?? null;
  }

  async createOrImportProject(setup: ProjectSetupInfo) {
    const info = await this.tryInvoke<ProjectInfo>("create_or_import_project", { request: setup });
    if (!info) return null;
    const project = fromProjectInfo(info);
    this.setWorkspacePath(project.rootPath);
    this.rememberProject(project);
    return project;
  }

  async detectAgents() {
    const info = await this.tryInvoke<AgentCandidateInfo[]>("detect_agent_clis", {});
    return info ? info.map(fromAgentCandidateInfo) : [];
  }

  // Persists a task to the workbench store (SQLite) so it survives a reload.
  async saveTask(task: Task) {
    const workspace = this.getWorkspacePath();
    const info = await this.tryInvoke<TaskInfo>("save_task", {
      request: {
        workspace,
        task: {
          id: task.id,
          feature_id: task.featureId,
          title: task.title,
          description: task.description,
          status: task.status,
          priority: task.priority,
          assigned_agent: task.assignedAgent,
          executor_profile: task.executorProfile,
          worktree_path: task.worktreePath,
          branch_name: task.branchName,
          tags: task.tags,
          review_approved: task.reviewApproved,
          tests_passed: task.testsPassed,
          updated_at: task.updatedAt,
        },
      },
    });
    return info ? fromTaskInfo(info) : null;
  }

  async prepareWorktree(task: Task) {
    const workspace = this.getWorkspacePath();
    const branchName = task.branchName || `thanos/${task.id.toLowerCase()}-${slug(task.title)}`;
    const info = await this.tryInvoke<WorktreeInfo>("prepare_task_worktree", {
      request: {
        workspace,
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
    const workspace = this.getWorkspacePath();
    const planner = agentType === "planner";
    const reviewer = agentType === "reviewer";
    const tester = agentType === "tester";
    const command = planner || reviewer ? "claude" : tester ? (task.executorProfile || "npm test") : task.executorProfile.includes("claude") ? "claude" : "codex";
    const provider = planner || reviewer ? "claude-code" : tester ? "shell" : task.executorProfile.includes("claude") ? "claude-code" : "codex";
    const info = await this.tryInvoke<AgentSessionInfo>("start_agent_session", {
      request: {
        workspace,
        task_id: task.id,
        agent_type: agentType,
        provider,
        command,
        args: [],
        worktree_path: planner ? "." : task.worktreePath,
      },
    });
    return info ? mapSession(info) : null;
  }

  async saveExecutionPlan(plan: ExecutionPlan) {
    const workspace = this.getWorkspacePath();
    const info = await this.tryInvoke<ExecutionPlanInfo>("save_execution_plan", {
      request: {
        workspace,
        plan: toPlanInfo(plan),
      },
    });
    return info ? fromPlanInfo(info) : null;
  }

  async approveExecutionPlan(taskId: string) {
    const workspace = this.getWorkspacePath();
    const info = await this.tryInvoke<ExecutionPlanInfo>("approve_execution_plan", {
      request: {
        workspace,
        task_id: taskId,
      },
    });
    return info ? fromPlanInfo(info) : null;
  }

  async readExecutionPlan(taskId: string) {
    const workspace = this.getWorkspacePath();
    const info = await this.tryInvoke<ExecutionPlanInfo>("read_execution_plan", {
      request: {
        workspace,
        task_id: taskId,
      },
    });
    return info ? fromPlanInfo(info) : null;
  }

  async collectGitDiff(task: Task) {
    const workspace = this.getWorkspacePath();
    const info = await this.tryInvoke<GitDiffInfo>("collect_git_diff", {
      request: {
        workspace,
        task_id: task.id,
        worktree_path: task.worktreePath,
      },
    });
    return info ? fromDiffInfo(info) : null;
  }

  async runTaskTests(task: Task, command = "go test ./...") {
    const workspace = this.getWorkspacePath();
    const info = await this.tryInvoke<TestRunInfo>("run_task_tests", {
      request: {
        workspace,
        task_id: task.id,
        worktree_path: task.worktreePath,
        command,
      },
    });
    return info ? fromTestInfo(info) : null;
  }

  async saveReview(review: Review) {
    const workspace = this.getWorkspacePath();
    const info = await this.tryInvoke<ReviewInfo>("save_review", {
      request: {
        workspace,
        review: toReviewInfo(review),
      },
    });
    return info ? fromReviewInfo(info) : null;
  }

  async approveReview(taskId: string) {
    const workspace = this.getWorkspacePath();
    const info = await this.tryInvoke<ReviewInfo>("approve_review", {
      request: {
        workspace,
        task_id: taskId,
      },
    });
    return info ? fromReviewInfo(info) : null;
  }

  async writeMemoryNode(node: MemoryNode) {
    const workspace = this.getWorkspacePath();
    const info = await this.tryInvoke<MemoryNodeInfo>("write_memory_node", {
      request: {
        workspace,
        node: toMemoryInfo(node),
      },
    });
    return info ? fromMemoryInfo(info) : null;
  }

  async searchMemory(query: string) {
    const workspace = this.getWorkspacePath();
    const info = await this.tryInvoke<MemoryNodeInfo[]>("search_memory", {
      request: {
        workspace,
        query,
      },
    });
    return info ? info.map(fromMemoryInfo) : [];
  }

  async listAgentBridgeTools() {
    return await this.tryInvoke<AgentBridgeTool[]>("list_agent_bridge_tools", {}) ?? [];
  }

  async createAgentSubtask(parentTaskId: string, title: string, description = "", priority = "P2", agent = "") {
    const workspace = this.getWorkspacePath();
    const info = await this.tryInvoke<TaskInfo>("bridge_create_subtask", {
      request: {
        workspace,
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
    const workspace = this.getWorkspacePath();
    const info = await this.tryInvoke<BridgeTaskMessageInfo>("bridge_message_sibling", {
      request: {
        workspace,
        source_task_id: sourceTaskId,
        target_task_id: targetTaskId,
        content,
      },
    });
    return info ? fromBridgeTaskMessageInfo(info) : null;
  }

  async inspectRelatedWork(query: string, taskId?: string, limit = 10) {
    const workspace = this.getWorkspacePath();
    const info = await this.tryInvoke<BridgeRelatedWorkInfo[]>("bridge_inspect_related_work", {
      request: {
        workspace,
        task_id: taskId,
        query,
        limit,
      },
    });
    return info ? info.map(fromBridgeRelatedWorkInfo) : [];
  }

  async attachTaskBranch(taskId: string, branchName: string, worktreePath?: string) {
    const workspace = this.getWorkspacePath();
    const info = await this.tryInvoke<BridgeBranchAttachmentInfo>("bridge_attach_branch", {
      request: {
        workspace,
        task_id: taskId,
        branch_name: branchName,
        worktree_path: worktreePath,
      },
    });
    return info ? fromBridgeBranchAttachmentInfo(info) : null;
  }

  async requestUserReview(taskId: string, requestedBy = "agent", notes = "") {
    const workspace = this.getWorkspacePath();
    const info = await this.tryInvoke<BridgeUserReviewRequestInfo>("bridge_request_user_review", {
      request: {
        workspace,
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

  async writeTerminal(data: string) {
    await this.tryInvoke<void>("write_terminal", { data });
  }

  async resumeAgent(sessionId: string) {
    const workspace = this.getWorkspacePath();
    const info = await this.tryInvoke<AgentSessionInfo>("resume_agent_session", {
      workspace,
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
    cwd: info.worktree_path,
    ptySessionId: info.pty_session_id,
    conversationLogPath: info.conversation_log_path,
    transcriptPath: info.conversation_log_path,
    output: [],
  };
}

function fromWorkbenchSnapshotInfo(info: WorkbenchSnapshotInfo): WorkbenchSnapshot {
  return {
    project: fromProjectInfo(info.project),
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
  if (status === "starting" || status === "running" || status === "waiting_user" || status === "completed" || status === "stopping" || status === "stopped" || status === "failed") {
    return status;
  }
  if (status === "exited" || status === "complete" || status === "completed") {
    return "stopped";
  }
  return "idle";
}

function fromProjectInfo(info: ProjectInfo): Project {
  return {
    id: info.id,
    name: info.name,
    rootPath: info.root_path,
    gitRemoteUrl: info.git_remote_url,
    defaultBranch: info.default_branch,
    worktreeRoot: info.worktree_root,
    packageManager: info.package_manager,
    devCommand: info.dev_command,
    testCommand: info.test_command,
    createdAt: info.created_at,
    updatedAt: info.updated_at,
    repos: info.repos,
    settings: info.settings,
  };
}

function fromAgentCandidateInfo(info: AgentCandidateInfo): AgentProvider {
  return {
    id: info.id,
    name: info.name,
    command: info.command,
    detectedPath: info.path ?? undefined,
    status: info.status,
    version: info.version ?? undefined,
    type: info.agent_type,
    enabled: info.enabled,
    setupHint: info.setup_hint,
  };
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
