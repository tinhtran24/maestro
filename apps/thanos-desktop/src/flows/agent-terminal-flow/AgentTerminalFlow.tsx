import { TerminalSquare } from "lucide-react";
import type { Task } from "../../domain/models";
import { Dialog } from "../../shared/ui/Dialog";
import { terminalLabel } from "../../features/terminal/mockRuntime";
import { XtermPanel } from "../AgentSessionFlow";
import { AgentSessionToolbar } from "./AgentSessionToolbar";
import { TerminalTabs } from "./TerminalTabs";
import { useAgentTerminalFlow } from "./useAgentTerminalFlow";

export function AgentTerminalFlow({ task }: { task: Task | null }) {
  const flow = useAgentTerminalFlow(task);
  if (!task) {
    return <div className="grid h-full place-items-center rounded-lg border border-slate-800 bg-bg-card text-sm text-text-muted">No task terminal selected.</div>;
  }

  return (
    <section className="grid h-full min-h-0 grid-rows-[auto_auto_minmax(0,1fr)_auto] rounded-lg border border-slate-800 bg-bg-card">
      <AgentSessionToolbar
        onStart={flow.start}
        onStop={() => flow.stop(flow.active)}
        onRestart={() => flow.restart(flow.active)}
        onTranscript={() => flow.active && flow.openTranscript(flow.active.id)}
        codingDisabled={flow.codingDisabled}
        active={flow.active}
      />
      <TerminalTabs
        sessions={flow.sessions}
        activeId={flow.active?.id}
        onSelect={flow.select}
        onPin={flow.pin}
        onClose={flow.close}
        isPinned={flow.isPinned}
      />

      {flow.active ? (
        <XtermPanel output={flow.active.output} sessionId={flow.active.id} />
      ) : (
        <div className="grid place-items-center p-6 text-center">
          <div className="flex flex-col items-center gap-2 text-text-muted">
            <TerminalSquare size={22} />
            <p className="text-sm font-medium text-text-main">No terminal sessions yet</p>
            <p className="text-xs">Start a Planning, Coding, Review, or Tests session from the toolbar.</p>
          </div>
        </div>
      )}

      {flow.active && (
        <footer className="flex flex-wrap items-center gap-x-4 gap-y-1 border-t border-slate-800 px-3 py-1.5 text-[11px] text-text-muted">
          <span className="inline-flex items-center gap-1 font-medium text-text-main">{terminalLabel(flow.active.step, flow.active.provider)}</span>
          <span>status: {flow.active.status}</span>
          <span className="font-mono">{flow.active.command}</span>
          {flow.active.cwd && <span className="font-mono">cwd: {flow.active.cwd}</span>}
          {flow.active.transcriptPath && <span className="ml-auto font-mono">{flow.active.transcriptPath}</span>}
        </footer>
      )}

      <Dialog
        open={Boolean(flow.transcript)}
        onClose={flow.closeTranscript}
        title={flow.transcript ? `Transcript · ${terminalLabel(flow.transcript.step, flow.transcript.provider)}` : "Transcript"}
        description={flow.transcript?.transcriptPath}
        footer={<button onClick={flow.closeTranscript} className="rounded-lg border border-slate-800 px-3 py-2 text-sm text-text-muted hover:text-text-main">Close</button>}
      >
        <pre className="max-h-[55vh] overflow-auto whitespace-pre-wrap rounded-lg bg-slate-950 p-3 font-mono text-xs text-text-muted">
          {flow.transcript?.output.join("\n") || "No output captured."}
        </pre>
      </Dialog>
    </section>
  );
}
