import { useEffect } from "react";
import { WorkbenchEventStream } from "../events/eventStream";
import { AgentSettingsFlow } from "../flows/agent-settings-flow/AgentSettingsFlow";
import { BoardFlow } from "../flows/BoardFlow";
import { ProjectOnboardingFlow } from "../flows/project-onboarding-flow/ProjectOnboardingFlow";
import { ProjectsFlow } from "../flows/projects-flow/ProjectsFlow";
import { RightContextSidebar } from "../flows/project-context-flow/RightContextSidebar";
import { TaskBottomPanel, TaskWorkbenchMain } from "../flows/TaskWorkbenchFlow";
import { TaskDialog } from "../flows/task-workbench-flow/TaskDialog";
import { useWorkbenchQuery } from "../queries/useWorkbenchQuery";
import { AppShell } from "../shared/ui/AppShell";
import { EmptyState } from "../shared/ui/EmptyState";
import { SplitPane } from "../shared/ui/SplitPane";
import { useWorkbenchStore } from "../state/workbenchStore";

const eventStream = new WorkbenchEventStream("ws://127.0.0.1:1421/events");

export function WorkbenchRoute() {
  const activeView = useWorkbenchStore((state) => state.activeView);
  const project = useWorkbenchStore((state) => state.project);
  const rightCollapsed = useWorkbenchStore((state) => state.rightCollapsed);
  const toggleRightCollapsed = useWorkbenchStore((state) => state.toggleRightCollapsed);
  const hydrate = useWorkbenchStore((state) => state.hydrate);
  const selectTask = useWorkbenchStore((state) => state.selectTask);
  const setLoadError = useWorkbenchStore((state) => state.setLoadError);
  const workbenchQuery = useWorkbenchQuery();

  useEffect(() => {
    if (workbenchQuery.data) {
      hydrate(workbenchQuery.data);
    }
  }, [hydrate, workbenchQuery.data]);

  useEffect(() => {
    if (workbenchQuery.error) setLoadError(String(workbenchQuery.error));
  }, [setLoadError, workbenchQuery.error]);

  useEffect(() => {
    eventStream.connect();
    return eventStream.subscribe((event) => {
      if (event.type === "onSelectTask") selectTask(event.taskId);
    });
  }, [selectTask]);

  if (workbenchQuery.isLoading) {
    return (
      <AppShell project={project}>
        <div className="grid h-full place-items-center bg-bg-app p-4">
          <EmptyState label="Loading workbench" />
        </div>
      </AppShell>
    );
  }

  if (workbenchQuery.isError) {
    return (
      <AppShell project={project}>
        <div className="grid h-full place-items-center bg-bg-app p-4">
          <div className="w-full max-w-lg rounded-lg border border-slate-800 bg-bg-card p-5">
            <h1 className="text-lg font-semibold">Unable to load workbench state</h1>
            <p className="mt-2 text-sm text-text-muted">Retry loading the current workspace or open setup to select a project folder.</p>
            <div className="mt-4 flex gap-2">
              <button onClick={() => workbenchQuery.refetch()} className="rounded-lg border border-slate-800 px-3 py-2 text-sm">Retry</button>
              <button onClick={() => console.log(workbenchQuery.error)} className="rounded-lg border border-slate-800 px-3 py-2 text-sm">View Logs</button>
            </div>
          </div>
        </div>
      </AppShell>
    );
  }

  if (!workbenchQuery.data) {
    return (
      <AppShell project={project}>
        <ProjectOnboardingFlow onLoaded={() => workbenchQuery.refetch()} />
      </AppShell>
    );
  }

  if (activeView === "projects") {
    return (
      <AppShell project={project}>
        <ProjectsFlow />
      </AppShell>
    );
  }

  if (activeView === "workflow_steps" || activeView === "settings" || activeView === "executors") {
    return (
      <AppShell project={project}>
        <AgentSettingsFlow tab={activeView === "workflow_steps" || activeView === "settings" ? "workflow" : "agents"} />
      </AppShell>
    );
  }

  if (activeView === "memory") {
    return (
      <AppShell project={project}>
        <div className="grid h-full place-items-center bg-bg-app p-6">
          <div className="w-full max-w-lg rounded-lg border border-slate-800 bg-bg-card p-5">
            <h1 className="text-lg font-semibold">Memory</h1>
            <p className="mt-2 text-sm text-text-muted">Project memory nodes appear in the task inspector and search results.</p>
          </div>
        </div>
      </AppShell>
    );
  }

  return (
    <AppShell project={project}>
      <>
        <SplitPane
          left={<div className="grid h-full min-h-0 grid-rows-[minmax(16rem,40%)_minmax(0,1fr)]"><BoardFlow /><TaskWorkbenchMain /></div>}
          right={<RightContextSidebar />}
          bottom={<TaskBottomPanel />}
          rightCollapsed={rightCollapsed}
          onExpandRight={toggleRightCollapsed}
        />
        <TaskDialog />
      </>
    </AppShell>
  );
}
