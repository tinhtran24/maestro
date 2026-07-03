import { Play } from "lucide-react";
import type { WorkflowStepId } from "../../domain/models";

export function StartAgentButton({ label, step, disabled, onStart }: { label: string; step: WorkflowStepId; disabled?: boolean; onStart: (step: WorkflowStepId) => void }) {
  return (
    <button disabled={disabled} onClick={() => onStart(step)} className="inline-flex items-center gap-2 rounded-lg border border-slate-800 bg-slate-900/80 px-3 py-2 text-sm disabled:opacity-40">
      <Play size={16} />
      {label}
    </button>
  );
}
