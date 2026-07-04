import { ArrowLeft, ArrowRight, Bot, Check, ClipboardPlus, FileText, Sparkles } from "lucide-react";
import type { LucideIcon } from "lucide-react";
import { Dialog } from "../../shared/ui/Dialog";
import { AIExtractionCard } from "./AIExtractionCard";
import { CreateTaskFlow } from "./CreateTaskFlow";
import { QuickCaptureEditor } from "./QuickCaptureEditor";
import { ReviewTaskFlow } from "./ReviewTaskFlow";
import { useQuickCaptureFlow, type CaptureStep } from "./useQuickCaptureFlow";

const STEPS: Array<{ id: CaptureStep; label: string; icon: LucideIcon }> = [
  { id: "capture", label: "Quick Capture", icon: ClipboardPlus },
  { id: "structure", label: "AI Structure", icon: Bot },
  { id: "review", label: "Review & Edit", icon: FileText },
  { id: "create", label: "Create", icon: Check },
];

// AI-first create-task wizard. Replaces the traditional CRUD create form.
export function QuickCaptureFlow() {
  const flow = useQuickCaptureFlow();
  const activeIndex = STEPS.findIndex((step) => step.id === flow.step);

  return (
    <Dialog
      open
      onClose={flow.close}
      title="Create New Task"
      description="Paste anything — AI structures it into an engineering task."
      size="xl"
      footer={<Footer step={flow.step} canContinue={flow.canContinue} onBack={flow.back} onNext={flow.next} onCreate={flow.create} />}
    >
      <div className="grid gap-5 md:grid-cols-[10rem_minmax(0,1fr)]">
        <nav className="hidden gap-1 md:grid md:content-start">
          {STEPS.map((step, index) => {
            const Icon = step.icon;
            const state = index === activeIndex ? "active" : index < activeIndex ? "done" : "todo";
            return (
              <button
                key={step.id}
                onClick={() => flow.goTo(step.id)}
                disabled={step.id !== "capture" && !flow.draft}
                className={`flex items-center gap-2 rounded-lg border px-2.5 py-2 text-left text-xs transition disabled:opacity-40 ${state === "active" ? "border-purple-primary bg-purple-primary/10 text-purple-hover" : state === "done" ? "border-slate-800 bg-slate-900/70 text-text-main" : "border-slate-800 bg-slate-900/40 text-text-muted"}`}
              >
                <span className={`grid h-5 w-5 shrink-0 place-items-center rounded-full text-[10px] ${state === "done" ? "bg-green-success/20 text-green-success" : "bg-slate-800 text-text-muted"}`}>
                  {state === "done" ? <Check size={11} /> : index + 1}
                </span>
                <Icon size={13} /> {step.label}
              </button>
            );
          })}
        </nav>

        <div className="min-w-0">
          {flow.step === "capture" && (
            <QuickCaptureEditor
              input={flow.input}
              onInput={flow.setInput}
              onExample={flow.appendExample}
              attachments={flow.attachments}
              onFiles={flow.addFiles}
              onUrl={flow.addUrl}
              onRemove={flow.removeAttachment}
              onMove={flow.moveAttachment}
            />
          )}
          {flow.step === "structure" && flow.draft && (
            <AIExtractionCard
              draft={flow.draft}
              onTitle={flow.patchTitle}
              onDescription={flow.patchDescription}
              onPriority={flow.patchPriority}
              onLabels={flow.patchLabels}
              onAcceptance={flow.patchAcceptance}
              onTechnicalNotes={flow.patchTechnicalNotes}
            />
          )}
          {flow.step === "review" && flow.draft && (
            <ReviewTaskFlow draft={flow.draft} attachments={flow.attachments} onRemove={flow.removeAttachment} onMove={flow.moveAttachment} />
          )}
          {flow.step === "create" && flow.draft && <CreateTaskFlow draft={flow.draft} attachments={flow.attachments} />}
        </div>
      </div>
    </Dialog>
  );
}

function Footer({ step, canContinue, onBack, onNext, onCreate }: { step: CaptureStep; canContinue: boolean; onBack: () => void; onNext: () => void; onCreate: () => void }) {
  const nextLabel: Record<CaptureStep, string> = {
    capture: "AI Structure",
    structure: "Review & Edit",
    review: "Continue",
    create: "Create Task",
  };
  return (
    <div className="flex w-full items-center justify-between gap-2">
      <button onClick={onBack} disabled={step === "capture"} className="inline-flex items-center gap-2 rounded-lg border border-slate-800 px-3 py-2 text-sm text-text-muted enabled:hover:border-slate-600 enabled:hover:text-text-main disabled:opacity-40">
        <ArrowLeft size={15} /> Back
      </button>
      {step === "create" ? (
        <button onClick={onCreate} className="inline-flex items-center gap-2 rounded-lg bg-purple-primary px-4 py-2 text-sm font-medium text-white hover:bg-purple-hover">
          <Check size={15} /> Create Task
        </button>
      ) : (
        <button onClick={onNext} disabled={!canContinue} className="inline-flex items-center gap-2 rounded-lg bg-purple-primary px-4 py-2 text-sm font-medium text-white enabled:hover:bg-purple-hover disabled:opacity-50">
          {step === "capture" && <Sparkles size={15} />} {nextLabel[step]} <ArrowRight size={15} />
        </button>
      )}
    </div>
  );
}
