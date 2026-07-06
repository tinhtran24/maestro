import type { Workspace } from "../app/types";
import { Panel } from "../shared/Panel";

export function MissionView({ workspace }: { workspace: Workspace }) {
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
