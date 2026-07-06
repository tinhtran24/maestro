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
	export class CreateTaskRequest {
	    root: string;
	    title: string;
	    prompt: string;
	    flow: string;
	    agent: string;
	
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
	export class FlowInfo {
	    id: string;
	    name: string;
	    steps: string[];
	
	    static createFrom(source: any = {}) {
	        return new FlowInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.steps = source["steps"];
	    }
	}
	export class NativeTerminalRequest {
	    providerId: string;
	    command: string;
	    args: string[];
	    cwd: string;
	    label: string;
	
	    static createFrom(source: any = {}) {
	        return new NativeTerminalRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.providerId = source["providerId"];
	        this.command = source["command"];
	        this.args = source["args"];
	        this.cwd = source["cwd"];
	        this.label = source["label"];
	    }
	}
	export class NativeTerminalSessionInfo {
	    id: string;
	    label: string;
	    command: string;
	    args: string[];
	    cwd: string;
	    status: string;
	    startedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new NativeTerminalSessionInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.label = source["label"];
	        this.command = source["command"];
	        this.args = source["args"];
	        this.cwd = source["cwd"];
	        this.status = source["status"];
	        this.startedAt = source["startedAt"];
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
	    }
	}
	export class UpdateTaskStatusRequest {
	    root: string;
	    taskId: string;
	    status: string;
	
	    static createFrom(source: any = {}) {
	        return new UpdateTaskStatusRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.root = source["root"];
	        this.taskId = source["taskId"];
	        this.status = source["status"];
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

