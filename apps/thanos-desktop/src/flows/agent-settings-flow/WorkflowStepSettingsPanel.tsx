import { Plus, Trash2 } from "lucide-react";
import { useState } from "react";
import type { WorkflowStepConfig } from "../../domain/models";
import type { useAgentSettingsFlow } from "./useAgentSettingsFlow";

export function WorkflowStepSettingsPanel({ flow }: { flow: ReturnType<typeof useAgentSettingsFlow> }) {
  return (
    <div className="min-h-0 overflow-y-auto rounded-lg border border-slate-800 bg-bg-card">
      <header className="border-b border-slate-800 p-3">
        <h2 className="font-semibold">Workflow Agents</h2>
        <p className="text-xs text-text-muted">Choose which installed agent runs each step. Configuration is stored locally.</p>
      </header>
      <div className="grid gap-3 p-3">
        {flow.workflowSteps.map((step) => (
          <StepEditor key={step.id} step={step} providers={flow.providers} update={(patch) => flow.updateWorkflowStep(step.id, patch)} />
        ))}
      </div>
    </div>
  );
}

function StepEditor({ step, providers, update }: { step: WorkflowStepConfig; providers: ReturnType<typeof useAgentSettingsFlow>["providers"]; update: (patch: Partial<WorkflowStepConfig>) => void }) {
  const options = [
    { name: "Shell", command: step.command || "npm test" },
    ...providers.map((provider) => ({ name: provider.name, command: provider.command })),
  ];
  return (
    <article className={`grid gap-3 rounded-lg border p-3 ${step.enabled ? "border-slate-800 bg-slate-950/50" : "border-slate-800/60 bg-slate-950/30 opacity-70"}`}>
      <div className="flex items-center justify-between">
        <h3 className="font-medium">{step.label}</h3>
        <label className="flex items-center gap-2 text-xs text-text-muted">
          <input type="checkbox" checked={step.enabled} onChange={(event) => update({ enabled: event.target.checked })} className="accent-purple-primary" />
          Enabled
        </label>
      </div>
      <div className="grid grid-cols-2 gap-2">
        <label className="grid gap-1 text-xs text-text-muted">
          <span>Agent provider</span>
          <select
            value={step.provider}
            onChange={(event) => {
              const selected = options.find((provider) => provider.name === event.target.value);
              update({ provider: event.target.value, command: selected?.command ?? step.command });
            }}
            className="h-9 rounded-lg border border-slate-800 bg-slate-950 px-2 text-text-main"
          >
            {options.some((option) => option.name === step.provider) ? null : <option value={step.provider}>{step.provider}</option>}
            {options.map((provider) => (
              <option key={provider.name} value={provider.name}>{provider.name}</option>
            ))}
          </select>
        </label>
        <Field label="Command" value={step.command} onChange={(command) => update({ command })} mono />
        <Field label="Arguments (space-separated)" value={step.args.join(" ")} onChange={(value) => update({ args: value.split(/\s+/).filter(Boolean) })} mono />
        <label className="grid gap-1 text-xs text-text-muted">
          <span>Working directory</span>
          <select value={step.workingDirectoryMode} onChange={(event) => update({ workingDirectoryMode: event.target.value as WorkflowStepConfig["workingDirectoryMode"] })} className="h-9 rounded-lg border border-slate-800 bg-slate-950 px-2 text-text-main">
            <option value="project">Project</option>
            <option value="worktree">Worktree</option>
            <option value="custom">Custom</option>
          </select>
        </label>
        <Field label="Timeout" value={step.timeout} onChange={(timeout) => update({ timeout })} />
        <Field label="Permissions (comma-separated)" value={step.permissions.join(", ")} onChange={(value) => update({ permissions: value.split(",").map((item) => item.trim()).filter(Boolean) })} />
      </div>

      <EnvEditor env={step.env} onChange={(env) => update({ env })} />

      <div className="flex flex-wrap gap-4 text-xs text-text-muted">
        <label className="flex items-center gap-2"><input type="checkbox" checked={step.autoStartTerminal} onChange={(event) => update({ autoStartTerminal: event.target.checked })} className="accent-purple-primary" />Auto-start terminal</label>
        <label className="flex items-center gap-2"><input type="checkbox" checked={step.approvalRequired} onChange={(event) => update({ approvalRequired: event.target.checked })} className="accent-purple-primary" />Require approval</label>
      </div>
    </article>
  );
}

function EnvEditor({ env, onChange }: { env: Record<string, string>; onChange: (env: Record<string, string>) => void }) {
  const [rows, setRows] = useState<Array<{ key: string; value: string }>>(() => Object.entries(env).map(([key, value]) => ({ key, value })));

  function commit(next: Array<{ key: string; value: string }>) {
    setRows(next);
    const record: Record<string, string> = {};
    for (const row of next) {
      const key = row.key.trim();
      if (key) record[key] = row.value;
    }
    onChange(record);
  }

  return (
    <div>
      <div className="mb-1 flex items-center justify-between">
        <span className="text-xs text-text-muted">Environment variables</span>
        <button onClick={() => commit([...rows, { key: "", value: "" }])} className="inline-flex items-center gap-1 rounded-md border border-slate-800 px-2 py-0.5 text-[11px] text-text-muted hover:text-text-main">
          <Plus size={12} /> Add
        </button>
      </div>
      {rows.length === 0 ? (
        <p className="rounded-md border border-dashed border-slate-800 bg-slate-950/40 px-2 py-1.5 text-[11px] text-text-muted">No environment variables.</p>
      ) : (
        <div className="grid gap-1.5">
          {rows.map((row, index) => (
            <div key={index} className="flex items-center gap-1.5">
              <input value={row.key} onChange={(event) => commit(rows.map((r, i) => (i === index ? { ...r, key: event.target.value } : r)))} placeholder="KEY" className="h-8 min-w-0 flex-1 rounded-md border border-slate-800 bg-slate-950 px-2 font-mono text-xs text-text-main placeholder:text-text-muted" />
              <input value={row.value} onChange={(event) => commit(rows.map((r, i) => (i === index ? { ...r, value: event.target.value } : r)))} placeholder="value" className="h-8 min-w-0 flex-[2] rounded-md border border-slate-800 bg-slate-950 px-2 font-mono text-xs text-text-main placeholder:text-text-muted" />
              <button onClick={() => commit(rows.filter((_, i) => i !== index))} className="shrink-0 rounded-md border border-slate-800 p-1.5 text-text-muted hover:border-red-danger/40 hover:text-red-danger" aria-label="Remove variable">
                <Trash2 size={13} />
              </button>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

function Field({ label, value, onChange, mono }: { label: string; value: string; onChange: (value: string) => void; mono?: boolean }) {
  return (
    <label className="grid gap-1 text-xs text-text-muted">
      <span>{label}</span>
      <input value={value} onChange={(event) => onChange(event.target.value)} className={`h-9 rounded-lg border border-slate-800 bg-slate-950 px-2 text-text-main ${mono ? "font-mono" : ""}`} />
    </label>
  );
}
