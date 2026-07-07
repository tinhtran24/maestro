import type { Workspace } from "../app/types";
import { regenerateOversight } from "../services/wails";
import { Panel } from "../shared/Panel";
import { buildMissionGraph } from "./missionGraph";

export function MissionView({ workspace, onReload }: { workspace: Workspace; onReload: () => Promise<void> }) {
  const oversightTasks = workspace.tasks.filter((task) => task.oversight);
  const graph = buildMissionGraph(workspace);
  const nodesByKind = graph.nodes.reduce<Record<string, number>>((acc, node) => {
    acc[node.kind] = (acc[node.kind] ?? 0) + 1;
    return acc;
  }, {});

  async function handleRegenerate(taskId: string) {
    if (!workspace.path) return;
    await regenerateOversight({ root: workspace.path, taskId });
    await onReload();
  }

  return (
    <div className="mission">
      <Panel title="Mission Graph" meta={`${graph.nodes.length} nodes · ${graph.edges.length} edges`}>
        <div className="mission-graph-summary">
          {Object.entries(nodesByKind).map(([kind, count]) => (
            <span key={kind}><strong>{count}</strong>{titleCase(kind)}</span>
          ))}
        </div>
        <div className="mission-graph">
          <section>
            <h3>Nodes</h3>
            <div className="mission-node-list">
              {graph.nodes.map((node) => (
                <article key={node.id} className={node.kind}>
                  <strong>{node.title}</strong>
                  <span>{titleCase(node.kind)} · {node.status}</span>
                  <small>{node.detail}</small>
                </article>
              ))}
            </div>
          </section>
          <section>
            <h3>Edges</h3>
            <div className="mission-edge-list">
              {graph.edges.map((edge) => (
                <article key={edge.id}>
                  <strong>{titleCase(edge.kind)}</strong>
                  <span>{edge.from}</span>
                  <span>{edge.to}</span>
                </article>
              ))}
            </div>
          </section>
        </div>
      </Panel>
      <Panel title="Blocked Work" meta={`${graph.blockedTaskIds.length} tasks`}>
        <div className="mission-path">
          {graph.blockedTaskIds.length === 0 ? <p className="empty">No blocked tasks.</p> : null}
          {graph.blockedTaskIds.map((taskId) => {
            const task = workspace.tasks.find((item) => item.id === taskId);
            return (
              <article key={taskId}>
                <strong>{task?.title ?? taskId}</strong>
                <span>{task?.dependencies.length ?? 0} dependencies · {task?.status ?? "unknown"}</span>
              </article>
            );
          })}
        </div>
      </Panel>
      <Panel title="Critical Path" meta={`${graph.criticalPath.length} tasks`}>
        <div className="pipeline">
          {graph.criticalPath.length === 0 ? <span>No Tasks</span> : null}
          {graph.criticalPath.map((taskId) => {
            const task = workspace.tasks.find((item) => item.id === taskId);
            return <span key={taskId}>{task?.title ?? taskId}</span>;
          })}
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

function titleCase(value: string) {
  return value.replace(/_/g, " ").replace(/\b\w/g, (letter) => letter.toUpperCase());
}
