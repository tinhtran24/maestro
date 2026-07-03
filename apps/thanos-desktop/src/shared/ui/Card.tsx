import type { LucideIcon } from "lucide-react";

export function Card({ title, icon: Icon, action, children, className = "" }: {
  title?: string;
  icon?: LucideIcon;
  action?: React.ReactNode;
  children: React.ReactNode;
  className?: string;
}) {
  return (
    <section className={`rounded-xl border border-slate-800 bg-slate-900/70 p-4 ${className}`}>
      {(title || action) && (
        <header className="mb-3 flex items-center justify-between gap-2">
          <h3 className="inline-flex items-center gap-2 text-sm font-semibold text-text-main">
            {Icon && <Icon size={15} className="text-text-muted" />}
            {title}
          </h3>
          {action}
        </header>
      )}
      {children}
    </section>
  );
}
