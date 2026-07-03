import { FileText, FlaskConical, Play, RotateCw, Square } from "lucide-react";
import type { LucideIcon } from "lucide-react";
import type { AgentSession } from "../../domain/models";
import type { TerminalStep } from "../../features/terminal/mockRuntime";

const starters: Array<{ step: TerminalStep; label: string; icon: LucideIcon }> = [
  { step: "planning", label: "Planning", icon: Play },
  { step: "coding", label: "Coding", icon: Play },
  { step: "review", label: "Review", icon: Play },
  { step: "testing", label: "Tests", icon: FlaskConical },
];

export function AgentSessionToolbar({
  onStart,
  onStop,
  onRestart,
  onTranscript,
  codingDisabled,
  active,
}: {
  onStart: (step: TerminalStep) => void;
  onStop: () => void;
  onRestart: () => void;
  onTranscript: () => void;
  codingDisabled: boolean;
  active: AgentSession | null;
}) {
  const running = active?.status === "running" || active?.status === "starting";
  return (
    <div className="flex flex-wrap items-center gap-2 border-b border-slate-800 p-2">
      {starters.map(({ step, label, icon: Icon }) => {
        const disabled = step === "coding" && codingDisabled;
        return (
          <button
            key={step}
            onClick={() => onStart(step)}
            disabled={disabled}
            title={disabled ? "Approve the execution plan before coding." : `Start ${label} session`}
            className="inline-flex items-center gap-2 rounded-md border border-slate-700 px-2 py-1 text-xs hover:border-slate-500 disabled:cursor-not-allowed disabled:opacity-40"
          >
            <Icon size={13} /> {label}
          </button>
        );
      })}
      <div className="ml-auto flex items-center gap-2">
        <button onClick={onRestart} disabled={!active} className="inline-flex items-center gap-2 rounded-md border border-slate-700 px-2 py-1 text-xs hover:border-slate-500 disabled:opacity-40">
          <RotateCw size={13} /> Restart
        </button>
        <button onClick={onTranscript} disabled={!active} className="inline-flex items-center gap-2 rounded-md border border-slate-700 px-2 py-1 text-xs hover:border-slate-500 disabled:opacity-40">
          <FileText size={13} /> Transcript
        </button>
        <button onClick={onStop} disabled={!running} className="inline-flex items-center gap-2 rounded-md border border-red-danger/40 px-2 py-1 text-xs text-red-danger hover:bg-red-danger/10 disabled:opacity-40">
          <Square size={13} /> Stop
        </button>
      </div>
    </div>
  );
}
