import type { LucideIcon } from "lucide-react";

export function FlowTabs<T extends string>({
  tabs,
  active,
  onChange,
  orientation = "horizontal",
}: {
  tabs: Array<{ id: T; label: string; icon: LucideIcon }>;
  active: T;
  onChange: (id: T) => void;
  orientation?: "horizontal" | "vertical";
}) {
  if (orientation === "vertical") {
    return (
      <div className="flex min-h-0 shrink-0 flex-col gap-1 overflow-y-auto border-r border-slate-800 p-2">
        {tabs.map(({ id, label, icon: Icon }) => (
          <button
            key={id}
            onClick={() => onChange(id)}
            title={label}
            className={`inline-flex shrink-0 items-center gap-2 rounded-lg px-3 py-2 text-left text-sm transition ${
              active === id ? "bg-purple-primary text-white" : "text-text-muted hover:bg-slate-800 hover:text-text-main"
            }`}
          >
            <Icon size={16} className="shrink-0" />
            <span className="truncate">{label}</span>
          </button>
        ))}
      </div>
    );
  }
  return (
    <div className="flex min-w-0 items-center gap-1 overflow-x-auto border-b border-slate-800">
      {tabs.map(({ id, label, icon: Icon }) => (
        <button
          key={id}
          onClick={() => onChange(id)}
          className={`inline-flex shrink-0 items-center gap-2 border-b px-3 py-2 text-sm ${
            active === id ? "border-purple-primary text-text-main" : "border-transparent text-text-muted hover:text-text-main"
          }`}
        >
          <Icon size={16} />
          {label}
        </button>
      ))}
    </div>
  );
}
