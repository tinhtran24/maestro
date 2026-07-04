import { ArrowRight, Check, Play, RotateCcw, X } from "lucide-react";

// Review decision actions. Approve/Reject/Request Changes are always available;
// Finish is gated on the review being approved and tests passing (the state
// machine enforces the same gate — a task cannot reach Done without approval).
export function ReviewActions({
  approved,
  finishable,
  onRun,
  onApprove,
  onReject,
  onRequestChanges,
  onFinish,
}: {
  approved: boolean;
  finishable: boolean;
  onRun: () => void;
  onApprove: () => void;
  onReject: () => void;
  onRequestChanges: () => void;
  onFinish: () => void;
}) {
  return (
    <div className="flex flex-wrap items-center gap-2">
      <button onClick={onRun} className="inline-flex items-center gap-2 rounded-lg border border-slate-800 bg-slate-900/80 px-3 py-2 text-sm text-text-main hover:border-slate-700">
        <Play size={15} /> Run Tests
      </button>
      <button onClick={onRequestChanges} className="inline-flex items-center gap-2 rounded-lg border border-slate-700 px-3 py-2 text-sm text-text-main hover:border-slate-500">
        <RotateCcw size={15} /> Request Changes
      </button>
      <button onClick={onReject} className="inline-flex items-center gap-2 rounded-lg border border-red-danger/40 px-3 py-2 text-sm text-red-danger hover:border-red-danger">
        <X size={15} /> Reject
      </button>
      <button onClick={onApprove} className={`inline-flex items-center gap-2 rounded-lg px-3 py-2 text-sm font-medium ${approved ? "border border-green-success/40 text-green-success" : "bg-purple-primary text-white hover:bg-purple-hover"}`}>
        <Check size={15} /> {approved ? "Approved" : "Approve Review"}
      </button>
      <button
        onClick={onFinish}
        disabled={!finishable}
        title={finishable ? "Finish the task (move to Done)" : "Approve the review and pass tests first"}
        className="inline-flex items-center gap-2 rounded-lg bg-green-success/90 px-3 py-2 text-sm font-medium text-white enabled:hover:bg-green-success disabled:opacity-40"
      >
        Finish Task <ArrowRight size={15} />
      </button>
    </div>
  );
}
