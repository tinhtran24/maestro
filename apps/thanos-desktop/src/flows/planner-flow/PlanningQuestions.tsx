import { MessageCircleQuestion } from "lucide-react";
import type { PlanningQuestion } from "../../domain/models";

export function PlanningQuestions({
  questions,
  onAnswer,
  onSubmit,
  canSubmit,
}: {
  questions: PlanningQuestion[];
  onAnswer: (id: string, value: string) => void;
  onSubmit: () => void;
  canSubmit: boolean;
}) {
  return (
    <div className="grid gap-3">
      <p className="inline-flex items-center gap-2 text-sm text-text-muted">
        <MessageCircleQuestion size={15} className="text-blue-info" />
        The planner needs a few answers before drafting an execution plan.
      </p>
      <ol className="grid gap-3">
        {questions.map((question, index) => (
          <li key={question.id} className="grid gap-1">
            <label className="text-sm font-medium text-text-main">
              {index + 1}. {question.prompt}
            </label>
            <textarea
              value={question.answer}
              onChange={(event) => onAnswer(question.id, event.target.value)}
              rows={2}
              placeholder="Your answer…"
              className="w-full resize-y rounded-lg border border-slate-800 bg-slate-950/60 px-3 py-2 text-sm text-text-main placeholder:text-text-muted focus:border-slate-600 focus:outline-none"
            />
          </li>
        ))}
      </ol>
      <div className="flex items-center justify-end gap-2">
        <button
          onClick={onSubmit}
          disabled={!canSubmit}
          title={canSubmit ? "Generate the execution plan" : "Answer every question first"}
          className="inline-flex items-center gap-2 rounded-lg bg-purple-primary px-3 py-2 text-sm font-medium text-white enabled:hover:bg-purple-hover disabled:opacity-50"
        >
          Generate Execution Plan
        </button>
      </div>
    </div>
  );
}
