import { Check, RotateCcw } from "lucide-react";

export function PlanApprovalGate({ onApprove, onReject }: { onApprove: () => void; onReject: () => void }) {
  return (
    <div className="flex flex-wrap items-center justify-between gap-3 rounded-lg border border-yellow-warning/30 bg-yellow-warning/10 p-3">
      <p className="text-sm text-yellow-warning">Review the plan. Coding cannot start until the plan is approved.</p>
      <div className="flex items-center gap-2">
        <button onClick={onReject} className="inline-flex items-center gap-2 rounded-lg border border-slate-700 px-3 py-2 text-sm text-text-main hover:border-slate-500">
          <RotateCcw size={15} /> Request Changes
        </button>
        <button onClick={onApprove} className="inline-flex items-center gap-2 rounded-lg bg-purple-primary px-3 py-2 text-sm font-medium text-white hover:bg-purple-hover">
          <Check size={15} /> Approve Plan
        </button>
      </div>
    </div>
  );
}
