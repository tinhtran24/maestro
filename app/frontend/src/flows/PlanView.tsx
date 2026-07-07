import { GitPullRequest, RotateCcw, Save, Send } from "lucide-react";
import { type FormEvent, useEffect, useMemo, useState } from "react";
import type { SpecNode, Workspace } from "../app/types";
import { createSpec, dispatchSpecs, undoPlanningChange, updateSpec } from "../services/wails";
import { Panel } from "../shared/Panel";

const specStates: SpecNode["state"][] = ["vague", "drafted", "validated", "testing", "complete", "stale", "archived"];

export function PlanView({ workspace, onReload }: { workspace: Workspace; onReload: () => Promise<void> }) {
  const flatSpecs = useMemo(() => flattenSpecs(workspace.specs), [workspace.specs]);
  const [selectedPath, setSelectedPath] = useState("");
  const selected = flatSpecs.find((spec) => spec.path === selectedPath) ?? flatSpecs[0];
  const [title, setTitle] = useState("");
  const [body, setBody] = useState("");
  const [state, setState] = useState<SpecNode["state"]>("drafted");
  const [command, setCommand] = useState("/create ");

  useEffect(() => {
    if (!selected) return;
    setSelectedPath(selected.path);
    setTitle(selected.title);
    setBody(selected.body);
    setState(selected.state);
  }, [selected?.path]);

  async function submitSpec(event: FormEvent) {
    event.preventDefault();
    if (!workspace.path || !title.trim()) return;
    await createSpec({ root: workspace.path, title, body, state, parentPath: selectedPath });
    setCommand("/create ");
    await onReload();
  }

  async function saveSpec(nextState = state) {
    if (!workspace.path || !selected) return;
    await updateSpec({ root: workspace.path, path: selected.path, title, body, state: nextState });
    setState(nextState);
    await onReload();
  }

  async function dispatchSelected() {
    if (!workspace.path) return;
    await dispatchSpecs({ root: workspace.path, path: selected?.path });
    await onReload();
  }

  async function undoLastChange() {
    if (!workspace.path) return;
    await undoPlanningChange({ root: workspace.path });
    await onReload();
  }

  async function runCommand(event: FormEvent) {
    event.preventDefault();
    const trimmed = command.trim();
    if (trimmed.startsWith("/create")) {
      const nextTitle = trimmed.replace(/^\/create\s*/, "").trim();
      if (nextTitle) {
        await createSpec({ root: workspace.path, title: nextTitle, body: "", state: "drafted", parentPath: selectedPath });
        setCommand("/create ");
        await onReload();
      }
      return;
    }
    if (trimmed === "/validate") await saveSpec("validated");
    if (trimmed === "/refine") await saveSpec("drafted");
    if (trimmed === "/break-down") await createSpec({ root: workspace.path, title: `${title || "Spec"} Follow Up`, body: "Break this spec into a dispatchable leaf task.", state: "drafted", parentPath: selected?.path });
    if (trimmed === "/dispatch") await dispatchSelected();
  }

  return (
    <div className="three-pane">
      <Panel title="Spec Explorer" meta={`${flatSpecs.length} specs`}>
        <div className="spec-tree">
          {workspace.specs.length === 0 ? <p className="empty">No specs directory found in this workspace.</p> : null}
          {workspace.specs.map((node) => <SpecItem key={node.id} node={node} selectedPath={selected?.path ?? ""} onSelect={setSelectedPath} />)}
        </div>
      </Panel>
      <Panel title="Focused Spec" meta={selected?.path ?? "No Spec"}>
        {selected ? (
          <div className="spec-editor">
            <input value={title} onChange={(event) => setTitle(event.target.value)} placeholder="Spec Title" />
            <select value={state} onChange={(event) => setState(event.target.value as SpecNode["state"])}>
              {specStates.map((item) => <option key={item} value={item}>{titleCase(item)}</option>)}
            </select>
            <textarea value={body} onChange={(event) => setBody(event.target.value)} placeholder="Spec body, acceptance criteria, dependencies..." />
            <div className="task-actions">
              <button onClick={() => saveSpec()} type="button"><Save size={15} /> Save Spec</button>
              <button onClick={() => saveSpec("validated")} type="button">Validate</button>
              <button onClick={() => saveSpec("stale")} type="button">Mark Stale</button>
              <button onClick={() => saveSpec("archived")} type="button">Archive</button>
            </div>
          </div>
        ) : (
          <article className="document">
            <h2>Plan Mode</h2>
            <p>Create a spec to start planning from local markdown.</p>
          </article>
        )}
      </Panel>
      <Panel title="Planning Commands" meta="slash commands">
        <div className="chat-log">
          <p><strong>Planner</strong> /create, /refine, /validate, /break-down, and /dispatch update local spec files.</p>
          <p><strong>Planner</strong> Leaf specs dispatch to board tasks with dependency wiring.</p>
          <p><strong>Planner</strong> Undo restores the latest saved spec snapshot.</p>
        </div>
        <form className="composer" onSubmit={runCommand}>
          <input value={command} onChange={(event) => setCommand(event.target.value)} placeholder="/create Spec Title" />
          <button disabled={!workspace.path} type="submit"><Send size={15} /></button>
        </form>
        <form className="routine-form plan-create-form" onSubmit={submitSpec}>
          <input value={title} onChange={(event) => setTitle(event.target.value)} placeholder="New Spec Title" />
          <select value={state} onChange={(event) => setState(event.target.value as SpecNode["state"])}>
            {specStates.map((item) => <option key={item} value={item}>{titleCase(item)}</option>)}
          </select>
          <button disabled={!workspace.path || !title.trim()} type="submit">Create Spec</button>
          <textarea value={body} onChange={(event) => setBody(event.target.value)} placeholder="Spec body, acceptance criteria, dependencies..." />
        </form>
        <button className="wide-action" onClick={dispatchSelected} type="button"><GitPullRequest size={16} /> Dispatch Leaf Specs To Board</button>
        <button className="wide-action secondary-action" onClick={undoLastChange} type="button"><RotateCcw size={16} /> Undo Planning Change</button>
      </Panel>
    </div>
  );
}

function SpecItem({ node, selectedPath, onSelect }: { node: SpecNode; selectedPath: string; onSelect: (path: string) => void }) {
  return (
    <div className={`spec-item ${node.path === selectedPath ? "active" : ""}`}>
      <button onClick={() => onSelect(node.path)} type="button">
        <strong>{node.title}</strong>
        <span>{node.path}</span>
        <em>{titleCase(node.state)}</em>
      </button>
      {node.children.length ? <div className="spec-children">{node.children.map((child) => <SpecItem key={child.id} node={child} selectedPath={selectedPath} onSelect={onSelect} />)}</div> : null}
    </div>
  );
}

function flattenSpecs(nodes: SpecNode[]): SpecNode[] {
  return nodes.flatMap((node) => [node, ...flattenSpecs(node.children)]);
}

function titleCase(value: string) {
  return value.replace(/_/g, " ").replace(/\b\w/g, (letter) => letter.toUpperCase());
}
