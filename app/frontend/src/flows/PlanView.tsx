import { GitPullRequest, Send } from "lucide-react";
import { useState } from "react";
import type { SpecNode, Workspace } from "../app/types";
import { createSpec } from "../services/wails";
import { Panel } from "../shared/Panel";

export function PlanView({ workspace, onReload }: { workspace: Workspace; onReload: () => Promise<void> }) {
  const [title, setTitle] = useState("");
  const [body, setBody] = useState("");

  async function submitSpec(event: React.FormEvent) {
    event.preventDefault();
    if (!workspace.path || !title.trim()) return;
    await createSpec({ root: workspace.path, title, body, state: "drafted" });
    setTitle("");
    setBody("");
    await onReload();
  }

  return (
    <div className="three-pane">
      <Panel title="Spec Explorer" meta="recursive">
        <div className="spec-tree">
          {workspace.specs.length === 0 ? <p className="empty">No specs directory found in this workspace.</p> : null}
          {workspace.specs.map((node) => <SpecItem key={node.id} node={node} />)}
        </div>
      </Panel>
      <Panel title="Focused Spec" meta="versioned markdown">
        <article className="document">
          <h2>Local AI development workbench</h2>
          <p>
            Thanos turns conversations into specs, specs into tasks, and task output into reviewable evidence.
            Specs are intentionally inspectable before any agent writes code.
          </p>
          <h3>Exit Criteria</h3>
          <ul>
            <li>Human can inspect task state, worktree, logs, diff, and usage.</li>
            <li>Agents run through named flows instead of hidden automation.</li>
            <li>Review remains a hard pause before merge.</li>
          </ul>
        </article>
      </Panel>
      <Panel title="Planning Chat" meta="slash commands">
        <div className="chat-log">
          <p><strong>Planner</strong> Use /create to turn this conversation into a spec node.</p>
          <p><strong>User</strong> Keep it local and Wails based.</p>
          <p><strong>Planner</strong> I will dispatch only leaf specs after validation.</p>
        </div>
        <form className="composer" onSubmit={submitSpec}>
          <input value={title} onChange={(event) => setTitle(event.target.value)} placeholder="/create spec title..." />
          <button disabled={!workspace.path} type="submit"><Send size={15} /></button>
        </form>
        <textarea className="spec-body" value={body} onChange={(event) => setBody(event.target.value)} placeholder="Spec body, acceptance criteria, dependencies..." />
        <button className="wide-action"><GitPullRequest size={16} /> Dispatch leaf specs to board</button>
      </Panel>
    </div>
  );
}

function SpecItem({ node }: { node: SpecNode }) {
  return (
    <div className="spec-item">
      <div>
        <strong>{node.title}</strong>
        <span>{node.path}</span>
      </div>
      <em>{node.state}</em>
      {node.children.length ? <div className="spec-children">{node.children.map((child) => <SpecItem key={child.id} node={child} />)}</div> : null}
    </div>
  );
}
