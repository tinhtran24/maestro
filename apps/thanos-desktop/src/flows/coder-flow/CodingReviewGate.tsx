import { ArrowRight, Loader2 } from "lucide-react";

// Gate that moves the task into Review. Disabled until the coding session
// completes — coding must finish before the task can enter Review.
export function CodingReviewGate({ ready, onSendToReview }: { ready: boolean; onSendToReview: () => void }) {
  return (
    <div className="flex flex-wrap items-center justify-between gap-3 rounded-lg border border-slate-800 bg-slate-950/40 p-3">
      <p className="inline-flex items-center gap-2 text-sm text-text-muted">
        {ready ? (
          "Coding complete. Send the changes to review."
        ) : (
          <>
            <Loader2 size={15} className="animate-spin text-blue-info" /> Coding in progress — waiting for the coder to finish…
          </>
        )}
      </p>
      <button
        onClick={onSendToReview}
        disabled={!ready}
        title={ready ? "Move the task into review" : "Wait for coding to complete"}
        className="inline-flex items-center gap-2 rounded-lg bg-purple-primary px-3 py-2 text-sm font-medium text-white enabled:hover:bg-purple-hover disabled:opacity-50"
      >
        Send to Review <ArrowRight size={15} />
      </button>
    </div>
  );
}
