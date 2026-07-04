import { CheckCircle2, ClipboardCheck, FlaskConical, ShieldCheck } from "lucide-react";
import type { Task } from "../../domain/models";
import { ReviewActions } from "./ReviewActions";
import { ReviewChangedFiles } from "./ReviewChangedFiles";
import { ReviewChecklist } from "./ReviewChecklist";
import { useReviewFlow } from "./useReviewFlow";

// Shown at the top of the task workbench while a task is under review.
export function isReviewPhase(task: Task): boolean {
  return task.status === "in_review";
}

export function ReviewFlow({ task }: { task: Task }) {
  const flow = useReviewFlow(task);

  return (
    <section className="rounded-xl border border-slate-800 bg-slate-900/70 p-4 shadow-lg shadow-black/20">
      <header className="mb-3 flex items-center justify-between gap-2">
        <h3 className="inline-flex items-center gap-2 text-base font-semibold">
          <ShieldCheck size={16} className="text-purple-hover" /> Review
        </h3>
        <span className="text-xs text-text-muted">Review · {flow.review.status.replace("_", " ")}</span>
      </header>

      <div className="grid gap-4">
        {flow.approved ? (
          <p className="inline-flex items-center gap-2 rounded-lg border border-green-success/30 bg-green-success/10 p-2 text-sm text-green-success">
            <CheckCircle2 size={15} /> Review approved. {flow.finishable ? "Finish the task to move it to Done." : "Run tests to unlock Finish."}
          </p>
        ) : (
          <p className="rounded-lg border border-yellow-warning/30 bg-yellow-warning/10 p-2 text-sm text-yellow-warning">
            The task cannot be finished until the review is approved and tests pass.
          </p>
        )}

        <div className="grid gap-4 lg:grid-cols-2">
          <Block icon={ClipboardCheck} title="Checklist">
            <ReviewChecklist items={flow.checklist} />
          </Block>
          <Block icon={FlaskConical} title="Tests">
            <p className="text-sm text-text-main">{flow.test ? `${flow.test.command} — ${flow.test.status}` : "Tests have not been run yet."}</p>
            {flow.test && <pre className="mt-2 max-h-28 overflow-auto whitespace-pre-wrap rounded-lg bg-slate-950/60 p-2 text-[11px] text-text-muted">{flow.test.stdout || flow.test.stderr}</pre>}
          </Block>
        </div>

        <Block icon={ClipboardCheck} title="Changed files">
          <ReviewChangedFiles diff={flow.diff} review={flow.review} />
        </Block>

        <ReviewActions
          approved={flow.approved}
          finishable={flow.finishable}
          onRun={flow.run}
          onApprove={flow.approve}
          onReject={flow.reject}
          onRequestChanges={flow.requestChanges}
          onFinish={flow.finish}
        />
      </div>
    </section>
  );
}

function Block({ icon: Icon, title, children }: { icon: typeof ShieldCheck; title: string; children: React.ReactNode }) {
  return (
    <div className="rounded-lg border border-slate-800 bg-slate-950/40 p-3">
      <h4 className="mb-2 inline-flex items-center gap-2 text-xs font-semibold uppercase tracking-wide text-text-muted">
        <Icon size={13} /> {title}
      </h4>
      {children}
    </div>
  );
}
