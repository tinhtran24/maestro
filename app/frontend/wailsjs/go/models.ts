export namespace app {
	
	export class ActivateWorkspaceRequest {
	    id: string;
	
	    static createFrom(source: any = {}) {
	        return new ActivateWorkspaceRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	    }
	}
	export class AgentRoleInfo {
	    id: string;
	    role: string;
	    harness: string;
	    model: string;
	    capabilities: string[];
	    readOnly: boolean;
	    source?: string;
	
	    static createFrom(source: any = {}) {
	        return new AgentRoleInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.role = source["role"];
	        this.harness = source["harness"];
	        this.model = source["model"];
	        this.capabilities = source["capabilities"];
	        this.readOnly = source["readOnly"];
	        this.source = source["source"];
	    }
	}
	export class AutomationInfo {
	    schema_version: number;
	    autoImplement: boolean;
	    autoTest: boolean;
	    autoSubmit: boolean;
	    autoRetry: boolean;
	
	    static createFrom(source: any = {}) {
	        return new AutomationInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.schema_version = source["schema_version"];
	        this.autoImplement = source["autoImplement"];
	        this.autoTest = source["autoTest"];
	        this.autoSubmit = source["autoSubmit"];
	        this.autoRetry = source["autoRetry"];
	    }
	}
	export class CreateTaskRequest {
	    root: string;
	    title: string;
	    prompt: string;
	    flow: string;
	    agent: string;
	    status: string;
	    dependencies: string[];
	
	    static createFrom(source: any = {}) {
	        return new CreateTaskRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.root = source["root"];
	        this.title = source["title"];
	        this.prompt = source["prompt"];
	        this.flow = source["flow"];
	        this.agent = source["agent"];
	        this.status = source["status"];
	        this.dependencies = source["dependencies"];
	    }
	}
	export class BatchCreateTasksRequest {
	    root: string;
	    tasks: CreateTaskRequest[];
	
	    static createFrom(source: any = {}) {
	        return new BatchCreateTasksRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.root = source["root"];
	        this.tasks = this.convertValues(source["tasks"], CreateTaskRequest);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class CommitInfo {
	    hash?: string;
	    summary: string;
	    message: string;
	    diff: string;
	    diffStat: string;
	    approved: boolean;
	    committed: boolean;
	    committedAt?: string;
	
	    static createFrom(source: any = {}) {
	        return new CommitInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.hash = source["hash"];
	        this.summary = source["summary"];
	        this.message = source["message"];
	        this.diff = source["diff"];
	        this.diffStat = source["diffStat"];
	        this.approved = source["approved"];
	        this.committed = source["committed"];
	        this.committedAt = source["committedAt"];
	    }
	}
	export class CommitTaskChangesRequest {
	    root: string;
	    taskId: string;
	    message: string;
	    approved: boolean;
	
	    static createFrom(source: any = {}) {
	        return new CommitTaskChangesRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.root = source["root"];
	        this.taskId = source["taskId"];
	        this.message = source["message"];
	        this.approved = source["approved"];
	    }
	}
	export class CreateSpecRequest {
	    root: string;
	    title: string;
	    body: string;
	    state: string;
	
	    static createFrom(source: any = {}) {
	        return new CreateSpecRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.root = source["root"];
	        this.title = source["title"];
	        this.body = source["body"];
	        this.state = source["state"];
	    }
	}
	
	export class WorkspaceFolderInfo {
	    id: string;
	    path: string;
	    label: string;
	
	    static createFrom(source: any = {}) {
	        return new WorkspaceFolderInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.path = source["path"];
	        this.label = source["label"];
	    }
	}
	export class CreateWorkspaceRequest {
	    name: string;
	    path: string;
	    folders: WorkspaceFolderInfo[];
	
	    static createFrom(source: any = {}) {
	        return new CreateWorkspaceRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.path = source["path"];
	        this.folders = this.convertValues(source["folders"], WorkspaceFolderInfo);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class DeleteWorkspaceRequest {
	    id: string;
	
	    static createFrom(source: any = {}) {
	        return new DeleteWorkspaceRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	    }
	}
	export class DiagnosticInfo {
	    kind: string;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new DiagnosticInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.kind = source["kind"];
	        this.message = source["message"];
	    }
	}
	export class EventInfo {
	    schema_version: number;
	    id: string;
	    at: string;
	    kind: string;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new EventInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.schema_version = source["schema_version"];
	        this.id = source["id"];
	        this.at = source["at"];
	        this.kind = source["kind"];
	        this.message = source["message"];
	    }
	}
	export class FeedbackRecord {
	    at: string;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new FeedbackRecord(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.at = source["at"];
	        this.message = source["message"];
	    }
	}
	export class FinishTaskTurnRequest {
	    root: string;
	    taskId: string;
	    turnId: string;
	    status: string;
	    stdout: string;
	    stderr: string;
	    stopReason: string;
	    usageUsd: number;
	    exitCode: number;
	    autoContinue: boolean;
	
	    static createFrom(source: any = {}) {
	        return new FinishTaskTurnRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.root = source["root"];
	        this.taskId = source["taskId"];
	        this.turnId = source["turnId"];
	        this.status = source["status"];
	        this.stdout = source["stdout"];
	        this.stderr = source["stderr"];
	        this.stopReason = source["stopReason"];
	        this.usageUsd = source["usageUsd"];
	        this.exitCode = source["exitCode"];
	        this.autoContinue = source["autoContinue"];
	    }
	}
	export class FlowInfo {
	    id: string;
	    name: string;
	    steps: string[];
	    parallelGroups?: string[][];
	    readOnly: boolean;
	    source?: string;
	
	    static createFrom(source: any = {}) {
	        return new FlowInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.steps = source["steps"];
	        this.parallelGroups = source["parallelGroups"];
	        this.readOnly = source["readOnly"];
	        this.source = source["source"];
	    }
	}
	export class NativeTerminalInputRequest {
	    sessionId: string;
	    data: string;
	
	    static createFrom(source: any = {}) {
	        return new NativeTerminalInputRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.sessionId = source["sessionId"];
	        this.data = source["data"];
	    }
	}
	export class NativeTerminalRequest {
	    providerId: string;
	    taskId: string;
	    step: string;
	    command: string;
	    args: string[];
	    cwd: string;
	    label: string;
	    rows: number;
	    cols: number;
	
	    static createFrom(source: any = {}) {
	        return new NativeTerminalRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.providerId = source["providerId"];
	        this.taskId = source["taskId"];
	        this.step = source["step"];
	        this.command = source["command"];
	        this.args = source["args"];
	        this.cwd = source["cwd"];
	        this.label = source["label"];
	        this.rows = source["rows"];
	        this.cols = source["cols"];
	    }
	}
	export class NativeTerminalResizeRequest {
	    sessionId: string;
	    rows: number;
	    cols: number;
	
	    static createFrom(source: any = {}) {
	        return new NativeTerminalResizeRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.sessionId = source["sessionId"];
	        this.rows = source["rows"];
	        this.cols = source["cols"];
	    }
	}
	export class NativeTerminalSessionInfo {
	    id: string;
	    label: string;
	    providerId: string;
	    taskId: string;
	    step: string;
	    command: string;
	    args: string[];
	    cwd: string;
	    status: string;
	    ptyId: string;
	    transcriptPath: string;
	    startedAt: string;
	    endedAt?: string;
	
	    static createFrom(source: any = {}) {
	        return new NativeTerminalSessionInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.label = source["label"];
	        this.providerId = source["providerId"];
	        this.taskId = source["taskId"];
	        this.step = source["step"];
	        this.command = source["command"];
	        this.args = source["args"];
	        this.cwd = source["cwd"];
	        this.status = source["status"];
	        this.ptyId = source["ptyId"];
	        this.transcriptPath = source["transcriptPath"];
	        this.startedAt = source["startedAt"];
	        this.endedAt = source["endedAt"];
	    }
	}
	export class PrepareTaskCommitRequest {
	    root: string;
	    taskId: string;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new PrepareTaskCommitRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.root = source["root"];
	        this.taskId = source["taskId"];
	        this.message = source["message"];
	    }
	}
	export class PromptRecord {
	    at: string;
	    prompt: string;
	
	    static createFrom(source: any = {}) {
	        return new PromptRecord(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.at = source["at"];
	        this.prompt = source["prompt"];
	    }
	}
	export class ProviderInfo {
	    id: string;
	    name: string;
	    command: string;
	    status: string;
	    path?: string;
	    version?: string;
	    type: string;
	    setupHint: string;
	    supportsRun: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ProviderInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.command = source["command"];
	        this.status = source["status"];
	        this.path = source["path"];
	        this.version = source["version"];
	        this.type = source["type"];
	        this.setupHint = source["setupHint"];
	        this.supportsRun = source["supportsRun"];
	    }
	}
	export class ResumeTaskTurnRequest {
	    root: string;
	    taskId: string;
	    feedback: string;
	
	    static createFrom(source: any = {}) {
	        return new ResumeTaskTurnRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.root = source["root"];
	        this.taskId = source["taskId"];
	        this.feedback = source["feedback"];
	    }
	}
	export class RetryRecord {
	    at: string;
	    reason: string;
	
	    static createFrom(source: any = {}) {
	        return new RetryRecord(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.at = source["at"];
	        this.reason = source["reason"];
	    }
	}
	export class RoutineInfo {
	    schema_version: number;
	    id: string;
	    name: string;
	    prompt: string;
	    flow: string;
	    schedule: string;
	    enabled: boolean;
	    updatedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new RoutineInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.schema_version = source["schema_version"];
	        this.id = source["id"];
	        this.name = source["name"];
	        this.prompt = source["prompt"];
	        this.flow = source["flow"];
	        this.schedule = source["schedule"];
	        this.enabled = source["enabled"];
	        this.updatedAt = source["updatedAt"];
	    }
	}
	export class RunTaskVerificationRequest {
	    root: string;
	    taskId: string;
	    command: string;
	    providerId: string;
	    passPattern: string;
	    failPattern: string;
	
	    static createFrom(source: any = {}) {
	        return new RunTaskVerificationRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.root = source["root"];
	        this.taskId = source["taskId"];
	        this.command = source["command"];
	        this.providerId = source["providerId"];
	        this.passPattern = source["passPattern"];
	        this.failPattern = source["failPattern"];
	    }
	}
	export class SaveAutomationRequest {
	    root: string;
	    automation: AutomationInfo;
	
	    static createFrom(source: any = {}) {
	        return new SaveAutomationRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.root = source["root"];
	        this.automation = this.convertValues(source["automation"], AutomationInfo);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class SearchTasksRequest {
	    root: string;
	    query: string;
	    includeArchived: boolean;
	    includeDeleted: boolean;
	
	    static createFrom(source: any = {}) {
	        return new SearchTasksRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.root = source["root"];
	        this.query = source["query"];
	        this.includeArchived = source["includeArchived"];
	        this.includeDeleted = source["includeDeleted"];
	    }
	}
	export class SpecNodeInfo {
	    id: string;
	    title: string;
	    state: string;
	    path: string;
	    children: SpecNodeInfo[];
	
	    static createFrom(source: any = {}) {
	        return new SpecNodeInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	        this.state = source["state"];
	        this.path = source["path"];
	        this.children = this.convertValues(source["children"], SpecNodeInfo);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class StartTaskTurnRequest {
	    root: string;
	    taskId: string;
	    step: string;
	    providerId: string;
	    sessionId: string;
	    transcriptPath: string;
	
	    static createFrom(source: any = {}) {
	        return new StartTaskTurnRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.root = source["root"];
	        this.taskId = source["taskId"];
	        this.step = source["step"];
	        this.providerId = source["providerId"];
	        this.sessionId = source["sessionId"];
	        this.transcriptPath = source["transcriptPath"];
	    }
	}
	export class TestResultInfo {
	    id: string;
	    taskId: string;
	    providerId?: string;
	    command: string;
	    status: string;
	    passed: boolean;
	    outputPath: string;
	    output: string;
	    exitCode: number;
	    passPattern?: string;
	    failPattern?: string;
	    startedAt: string;
	    endedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new TestResultInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.taskId = source["taskId"];
	        this.providerId = source["providerId"];
	        this.command = source["command"];
	        this.status = source["status"];
	        this.passed = source["passed"];
	        this.outputPath = source["outputPath"];
	        this.output = source["output"];
	        this.exitCode = source["exitCode"];
	        this.passPattern = source["passPattern"];
	        this.failPattern = source["failPattern"];
	        this.startedAt = source["startedAt"];
	        this.endedAt = source["endedAt"];
	    }
	}
	export class TaskTurnInfo {
	    id: string;
	    taskId: string;
	    step: string;
	    providerId: string;
	    sessionId?: string;
	    status: string;
	    worktree: string;
	    startedAt: string;
	    endedAt?: string;
	    stdoutPath?: string;
	    stderrPath?: string;
	    transcriptPath?: string;
	    stopReason?: string;
	    usageUsd: number;
	    failureCategory?: string;
	    autoContinue: boolean;
	
	    static createFrom(source: any = {}) {
	        return new TaskTurnInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.taskId = source["taskId"];
	        this.step = source["step"];
	        this.providerId = source["providerId"];
	        this.sessionId = source["sessionId"];
	        this.status = source["status"];
	        this.worktree = source["worktree"];
	        this.startedAt = source["startedAt"];
	        this.endedAt = source["endedAt"];
	        this.stdoutPath = source["stdoutPath"];
	        this.stderrPath = source["stderrPath"];
	        this.transcriptPath = source["transcriptPath"];
	        this.stopReason = source["stopReason"];
	        this.usageUsd = source["usageUsd"];
	        this.failureCategory = source["failureCategory"];
	        this.autoContinue = source["autoContinue"];
	    }
	}
	export class TaskInfo {
	    schema_version: number;
	    id: string;
	    title: string;
	    prompt: string;
	    status: string;
	    flow: string;
	    agent: string;
	    branch: string;
	    worktree: string;
	    updatedAt: string;
	    usageUsd: number;
	    archived: boolean;
	    deleted: boolean;
	    tombstone: boolean;
	    dependencies: string[];
	    blocked: boolean;
	    promptHistory: PromptRecord[];
	    feedbackHistory: FeedbackRecord[];
	    retryHistory: RetryRecord[];
	    turns: TaskTurnInfo[];
	    lastTurn?: TaskTurnInfo;
	    lastOutput?: string;
	    testsPassed: boolean;
	    lastTestResult?: TestResultInfo;
	    commit?: CommitInfo;
	    failureCategory: string;
	    createdAt: string;
	
	    static createFrom(source: any = {}) {
	        return new TaskInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.schema_version = source["schema_version"];
	        this.id = source["id"];
	        this.title = source["title"];
	        this.prompt = source["prompt"];
	        this.status = source["status"];
	        this.flow = source["flow"];
	        this.agent = source["agent"];
	        this.branch = source["branch"];
	        this.worktree = source["worktree"];
	        this.updatedAt = source["updatedAt"];
	        this.usageUsd = source["usageUsd"];
	        this.archived = source["archived"];
	        this.deleted = source["deleted"];
	        this.tombstone = source["tombstone"];
	        this.dependencies = source["dependencies"];
	        this.blocked = source["blocked"];
	        this.promptHistory = this.convertValues(source["promptHistory"], PromptRecord);
	        this.feedbackHistory = this.convertValues(source["feedbackHistory"], FeedbackRecord);
	        this.retryHistory = this.convertValues(source["retryHistory"], RetryRecord);
	        this.turns = this.convertValues(source["turns"], TaskTurnInfo);
	        this.lastTurn = this.convertValues(source["lastTurn"], TaskTurnInfo);
	        this.lastOutput = source["lastOutput"];
	        this.testsPassed = source["testsPassed"];
	        this.lastTestResult = this.convertValues(source["lastTestResult"], TestResultInfo);
	        this.commit = this.convertValues(source["commit"], CommitInfo);
	        this.failureCategory = source["failureCategory"];
	        this.createdAt = source["createdAt"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	
	export class UpdateTaskFlagsRequest {
	    root: string;
	    taskId: string;
	    archived: boolean;
	    deleted: boolean;
	    tombstone: boolean;
	
	    static createFrom(source: any = {}) {
	        return new UpdateTaskFlagsRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.root = source["root"];
	        this.taskId = source["taskId"];
	        this.archived = source["archived"];
	        this.deleted = source["deleted"];
	        this.tombstone = source["tombstone"];
	    }
	}
	export class UpdateTaskStatusRequest {
	    root: string;
	    taskId: string;
	    status: string;
	    feedback: string;
	    failureCategory: string;
	
	    static createFrom(source: any = {}) {
	        return new UpdateTaskStatusRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.root = source["root"];
	        this.taskId = source["taskId"];
	        this.status = source["status"];
	        this.feedback = source["feedback"];
	        this.failureCategory = source["failureCategory"];
	    }
	}
	export class UpdateWorkspaceRequest {
	    id: string;
	    name: string;
	    folders: WorkspaceFolderInfo[];
	    activeFolderId: string;
	
	    static createFrom(source: any = {}) {
	        return new UpdateWorkspaceRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.folders = this.convertValues(source["folders"], WorkspaceFolderInfo);
	        this.activeFolderId = source["activeFolderId"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class UpsertRoutineRequest {
	    root: string;
	    id: string;
	    name: string;
	    prompt: string;
	    flow: string;
	    schedule: string;
	    enabled: boolean;
	
	    static createFrom(source: any = {}) {
	        return new UpsertRoutineRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.root = source["root"];
	        this.id = source["id"];
	        this.name = source["name"];
	        this.prompt = source["prompt"];
	        this.flow = source["flow"];
	        this.schedule = source["schedule"];
	        this.enabled = source["enabled"];
	    }
	}
	
	export class WorkspaceInfo {
	    workspaceId: string;
	    dataKey: string;
	    name: string;
	    path: string;
	    folders: WorkspaceFolderInfo[];
	    defaultBranch: string;
	    tasks: TaskInfo[];
	    specs: SpecNodeInfo[];
	    routines: RoutineInfo[];
	    automation: AutomationInfo;
	    agents: AgentRoleInfo[];
	    flows: FlowInfo[];
	    events: EventInfo[];
	    providers: ProviderInfo[];
	    diagnostics: DiagnosticInfo[];
	
	    static createFrom(source: any = {}) {
	        return new WorkspaceInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.workspaceId = source["workspaceId"];
	        this.dataKey = source["dataKey"];
	        this.name = source["name"];
	        this.path = source["path"];
	        this.folders = this.convertValues(source["folders"], WorkspaceFolderInfo);
	        this.defaultBranch = source["defaultBranch"];
	        this.tasks = this.convertValues(source["tasks"], TaskInfo);
	        this.specs = this.convertValues(source["specs"], SpecNodeInfo);
	        this.routines = this.convertValues(source["routines"], RoutineInfo);
	        this.automation = this.convertValues(source["automation"], AutomationInfo);
	        this.agents = this.convertValues(source["agents"], AgentRoleInfo);
	        this.flows = this.convertValues(source["flows"], FlowInfo);
	        this.events = this.convertValues(source["events"], EventInfo);
	        this.providers = this.convertValues(source["providers"], ProviderInfo);
	        this.diagnostics = this.convertValues(source["diagnostics"], DiagnosticInfo);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class WorkspaceRecordInfo {
	    schema_version: number;
	    id: string;
	    name: string;
	    dataKey: string;
	    folders: WorkspaceFolderInfo[];
	    activeFolderId: string;
	    createdAt: string;
	    updatedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new WorkspaceRecordInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.schema_version = source["schema_version"];
	        this.id = source["id"];
	        this.name = source["name"];
	        this.dataKey = source["dataKey"];
	        this.folders = this.convertValues(source["folders"], WorkspaceFolderInfo);
	        this.activeFolderId = source["activeFolderId"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class WorkspaceRegistryInfo {
	    schema_version: number;
	    activeWorkspaceId: string;
	    workspaces: WorkspaceRecordInfo[];
	
	    static createFrom(source: any = {}) {
	        return new WorkspaceRegistryInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.schema_version = source["schema_version"];
	        this.activeWorkspaceId = source["activeWorkspaceId"];
	        this.workspaces = this.convertValues(source["workspaces"], WorkspaceRecordInfo);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}
