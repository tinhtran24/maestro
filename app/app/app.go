package app

import (
	"context"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx      context.Context
	provider *RealProvider
	terminal *NativeTerminalManager
}

func New() *App {
	return &App{
		provider: NewRealProvider(),
		terminal: NewNativeTerminalManager(),
	}
}

func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) Ping() string {
	return "thanos-app"
}

func (a *App) CurrentWorkspaceFolder() (string, error) {
	return a.provider.CurrentWorkspaceFolder()
}

func (a *App) SelectWorkspaceFolder() (*string, error) {
	if a.ctx == nil {
		return nil, nil
	}
	path, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select local repository folder",
	})
	if err != nil {
		return nil, err
	}
	if path == "" {
		return nil, nil
	}
	return &path, nil
}

func (a *App) ListWorkspaces() (*WorkspaceRegistryInfo, error) {
	return a.provider.ListWorkspaces()
}

func (a *App) CreateWorkspace(request CreateWorkspaceRequest) (*WorkspaceRecordInfo, error) {
	return a.provider.CreateWorkspace(request)
}

func (a *App) UpdateWorkspace(request UpdateWorkspaceRequest) (*WorkspaceRecordInfo, error) {
	return a.provider.UpdateWorkspace(request)
}

func (a *App) DeleteWorkspace(request DeleteWorkspaceRequest) (*WorkspaceRegistryInfo, error) {
	return a.provider.DeleteWorkspace(request)
}

func (a *App) ActivateWorkspace(request ActivateWorkspaceRequest) (*WorkspaceRecordInfo, error) {
	return a.provider.ActivateWorkspace(request)
}

func (a *App) LoadActiveWorkspace() (*WorkspaceInfo, error) {
	return a.provider.LoadActiveWorkspace()
}

func (a *App) LoadWorkspace(root string) (*WorkspaceInfo, error) {
	return a.provider.LoadWorkspace(root)
}

func (a *App) CreateTask(request CreateTaskRequest) (*TaskInfo, error) {
	return a.provider.CreateTask(request)
}

func (a *App) BatchCreateTasks(request BatchCreateTasksRequest) ([]TaskInfo, error) {
	return a.provider.BatchCreateTasks(request)
}

func (a *App) SearchTasks(request SearchTasksRequest) ([]TaskInfo, error) {
	return a.provider.SearchTasks(request)
}

func (a *App) UpdateTaskFlags(request UpdateTaskFlagsRequest) (*TaskInfo, error) {
	return a.provider.UpdateTaskFlags(request)
}

func (a *App) UpdateTaskStatus(request UpdateTaskStatusRequest) (*TaskInfo, error) {
	return a.provider.UpdateTaskStatus(request)
}

func (a *App) StartTaskTurn(request StartTaskTurnRequest) (*TaskInfo, error) {
	return a.provider.StartTaskTurn(request)
}

func (a *App) FinishTaskTurn(request FinishTaskTurnRequest) (*TaskInfo, error) {
	return a.provider.FinishTaskTurn(request)
}

func (a *App) ResumeTaskTurn(request ResumeTaskTurnRequest) (*TaskInfo, error) {
	return a.provider.ResumeTaskTurn(request)
}

func (a *App) RunTaskVerification(request RunTaskVerificationRequest) (*TaskInfo, error) {
	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	return a.provider.RunTaskVerification(ctx, request)
}

func (a *App) PrepareTaskCommit(request PrepareTaskCommitRequest) (*TaskInfo, error) {
	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	return a.provider.PrepareTaskCommit(ctx, request)
}

func (a *App) CommitTaskChanges(request CommitTaskChangesRequest) (*TaskInfo, error) {
	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	return a.provider.CommitTaskChanges(ctx, request)
}

func (a *App) RegenerateOversight(request RegenerateOversightRequest) (*TaskInfo, error) {
	return a.provider.RegenerateOversight(request)
}

func (a *App) CreateSpec(request CreateSpecRequest) (*SpecNodeInfo, error) {
	return a.provider.CreateSpec(request)
}

func (a *App) UpdateSpec(request UpdateSpecRequest) (*SpecNodeInfo, error) {
	return a.provider.UpdateSpec(request)
}

func (a *App) DispatchSpecs(request DispatchSpecsRequest) ([]TaskInfo, error) {
	return a.provider.DispatchSpecs(request)
}

func (a *App) UndoPlanningChange(request UndoPlanningChangeRequest) (*SpecNodeInfo, error) {
	return a.provider.UndoPlanningChange(request)
}

func (a *App) UpsertRoutine(request UpsertRoutineRequest) (*RoutineInfo, error) {
	return a.provider.UpsertRoutine(request)
}

func (a *App) TriggerRoutine(request TriggerRoutineRequest) (*TaskInfo, error) {
	return a.provider.TriggerRoutine(request)
}

func (a *App) RunRoutineScheduler(request RunRoutineSchedulerRequest) ([]TaskInfo, error) {
	return a.provider.RunRoutineScheduler(request)
}

func (a *App) SaveAutomation(request SaveAutomationRequest) (*AutomationInfo, error) {
	return a.provider.SaveAutomation(request)
}

func (a *App) DetectAgentCLIs() ([]AgentCandidateInfo, error) {
	return a.provider.DetectAgentCLIs(context.Background())
}

func (a *App) StartNativeTerminal(request NativeTerminalRequest) (*NativeTerminalSessionInfo, error) {
	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	return a.terminal.Start(ctx, request)
}

func (a *App) WriteNativeTerminal(request NativeTerminalInputRequest) error {
	return a.terminal.Write(request)
}

func (a *App) ResizeNativeTerminal(request NativeTerminalResizeRequest) error {
	return a.terminal.Resize(request)
}

func (a *App) StopNativeTerminal(sessionID string) bool {
	return a.terminal.Stop(sessionID)
}
