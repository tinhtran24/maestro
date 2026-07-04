import { HelpCircle } from "lucide-react";

// Surfaces what the mock planner is unsure about. The planner can ask these as
// clarifying questions later during Planning.
export function MissingInformationCard({ questions }: { questions: string[] }) {
  if (!questions.length) return null;
  return (
    <div className="grid gap-2 rounded-lg border border-yellow-warning/30 bg-yellow-warning/10 p-3">
      <h4 className="inline-flex items-center gap-2 text-xs font-semibold uppercase tracking-wide text-yellow-warning">
        <HelpCircle size={13} /> Missing information
      </h4>
      <ul className="grid gap-1 text-xs text-yellow-warning/90">
        {questions.map((question) => <li key={question}>• {question}</li>)}
      </ul>
      <p className="text-[11px] text-text-muted">The planner can ask these during Planning.</p>
    </div>
  );
}
