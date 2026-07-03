import { CheckCircle2, Loader2, Terminal, XCircle } from "lucide-react";
import type { useAgentSettingsFlow } from "./useAgentSettingsFlow";

export function AgentCommandTestPanel({ flow }: { flow: ReturnType<typeof useAgentSettingsFlow> }) {
  const installed = flow.providers.filter((provider) => provider.status === "installed").length;
  const needsSetup = flow.providers.filter((provider) => provider.status === "needs_setup").length;
  const notFound = flow.providers.filter((provider) => provider.status === "not_found").length;
  const result = flow.testResult;

  return (
    <aside className="min-h-0 overflow-y-auto rounded-lg border border-slate-800 bg-bg-card p-4">
      <div className="flex items-center gap-2">
        <Terminal size={18} className="text-blue-info" />
        <h2 className="font-semibold">Command Test</h2>
      </div>
      <p className="mt-2 text-sm text-text-muted">Test Run inspects the local command status only (mock). It does not install, configure, or execute agents.</p>

      <dl className="mt-4 grid gap-2 text-sm">
        <Row label="Installed" value={installed} />
        <Row label="Needs setup" value={needsSetup} />
        <Row label="Not found" value={notFound} />
        <Row label="Configured steps" value={flow.workflowSteps.length} />
      </dl>

      <div className="mt-4 rounded-lg border border-slate-800 bg-slate-950/60 p-3">
        <p className="text-xs font-semibold uppercase tracking-wide text-text-muted">Last test</p>
        {flow.testing ? (
          <p className="mt-2 flex items-center gap-2 text-sm text-blue-info"><Loader2 size={14} className="animate-spin" /> Testing {flow.testing}…</p>
        ) : result ? (
          <div className="mt-2">
            <p className={`flex items-center gap-2 text-sm ${result.ok ? "text-green-success" : "text-red-danger"}`}>
              {result.ok ? <CheckCircle2 size={14} /> : <XCircle size={14} />} {result.name}
            </p>
            <pre className="mt-2 overflow-auto whitespace-pre-wrap rounded-md bg-slate-950 p-2 font-mono text-[11px] text-text-muted">{result.output}</pre>
          </div>
        ) : (
          <p className="mt-2 text-sm text-text-muted">Idle — run a test from the agents list.</p>
        )}
      </div>
    </aside>
  );
}

function Row({ label, value }: { label: string; value: number }) {
  return (
    <div className="flex justify-between">
      <dt className="text-text-muted">{label}</dt>
      <dd>{value}</dd>
    </div>
  );
}
