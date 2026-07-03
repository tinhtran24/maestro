import { AlertTriangle, FileCode2, FlaskConical, ListChecks } from "lucide-react";
import type { ExecutionPlan } from "../../domain/models";

export function ExecutionPlanPreview({ plan }: { plan: ExecutionPlan }) {
  return (
    <div className="grid gap-4">
      <p className="text-sm text-text-main">{plan.summary}</p>

      <Section icon={ListChecks} title="Steps">
        <ol className="grid gap-2">
          {plan.steps.map((step) => (
            <li key={step.id} className="rounded-lg border border-slate-800 bg-slate-950/40 p-2">
              <p className="text-sm font-medium text-text-main">{step.title}</p>
              <p className="mt-0.5 text-xs text-text-muted">{step.description}</p>
            </li>
          ))}
        </ol>
      </Section>

      <div className="grid gap-4 sm:grid-cols-2">
        <Section icon={FileCode2} title="Files to touch">
          <ul className="grid gap-1 text-sm text-text-muted">
            {plan.filesToTouch.map((file) => (
              <li key={file} className="font-mono text-xs">{file}</li>
            ))}
          </ul>
        </Section>
        <Section icon={FlaskConical} title="Test strategy">
          <ul className="grid gap-1 text-sm text-text-muted">
            {plan.testStrategy.map((item) => (
              <li key={item} className="text-xs">• {item}</li>
            ))}
          </ul>
        </Section>
      </div>

      <Section icon={AlertTriangle} title="Risks">
        <ul className="grid gap-1 text-sm text-text-muted">
          {plan.risks.map((risk) => (
            <li key={risk} className="text-xs">• {risk}</li>
          ))}
        </ul>
      </Section>
    </div>
  );
}

function Section({ icon: Icon, title, children }: { icon: typeof ListChecks; title: string; children: React.ReactNode }) {
  return (
    <div>
      <h4 className="mb-2 inline-flex items-center gap-2 text-xs font-semibold uppercase tracking-wide text-text-muted">
        <Icon size={13} /> {title}
      </h4>
      {children}
    </div>
  );
}
