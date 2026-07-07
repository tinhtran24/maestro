import type { Workspace } from "../app/types";
import { regenerateOversight } from "../services/wails";
import { Panel } from "../shared/Panel";

export function MissionView({ workspace, onReload }: { workspace: Workspace; onReload: () => Promise<void> }) {
  const oversightTasks = workspace.tasks.filter((task) => task.oversight);

  async function handleRegenerate(taskId: string) {
    if (!workspace.path) return;
    await regenerateOversight({ root: workspace.path, taskId });
    await onReload();
  }

  return (
    <div className="mission">
      <Panel title="Pipeline Map" meta="specs to tasks">
        <div className="pipeline">
          <span>Idea</span>
          <span>Spec</span>
          <span>Task</span>
          <span>Agent Run</span>
          <span>Review</span>
          <span>Merge</span>
        </div>
      </Panel>
      <Panel title="Oversight Summaries" meta={`${oversightTasks.length} artifacts`}>
        <div className="oversight-list">
          {oversightTasks.map((task) => (
            <article key={task.id}>
              <div>
                <strong>{task.title}</strong>
                <span>{task.oversight?.generatedAt}</span>
              </div>
              <p>{task.oversight?.summary}</p>
              <dl>
                <div><dt>Status</dt><dd>{task.oversight?.status}</dd></div>
                <div><dt>Tests</dt><dd>{task.oversight?.testResult}</dd></div>
                <div><dt>Usage</dt><dd>${(task.oversight?.usageUsd ?? 0).toFixed(2)}</dd></div>
                <div><dt>Risks</dt><dd>{task.oversight?.risks.length ?? 0}</dd></div>
              </dl>
              {task.oversight?.changedFiles.length ? <small>{task.oversight.changedFiles.join(", ")}</small> : null}
              <button onClick={() => handleRegenerate(task.id)} type="button">Regenerate Oversight</button>
            </article>
          ))}
          {oversightTasks.length === 0 ? <p className="empty">No oversight artifacts yet.</p> : null}
        </div>
      </Panel>
      <Panel title="Oversight Timeline" meta="audit trail">
        <div className="timeline">
          {workspace.events.map((event) => (
            <article key={event.id}>
              <time>{event.at}</time>
              <strong>{event.kind}</strong>
              <p>{event.message}</p>
            </article>
          ))}
        </div>
      </Panel>
    </div>
  );
}
