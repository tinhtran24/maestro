import { Brain, FolderGit2, Plus, Upload } from "lucide-react";
import type { LucideIcon } from "lucide-react";
import { useWorkbenchStore } from "../../state/workbenchStore";

export function QuickActionsCard() {
  const openCreateTask = useWorkbenchStore((state) => state.openCreateTask);
  const setActiveView = useWorkbenchStore((state) => state.setActiveView);

  const actions: Array<{ icon: LucideIcon; title: string; desc: string; onClick: () => void }> = [
    { icon: Plus, title: "New Task", desc: "Create new task", onClick: () => { setActiveView("workbench"); openCreateTask(); } },
    { icon: Upload, title: "Import Tasks", desc: "Import from file", onClick: () => { setActiveView("workbench"); openCreateTask(); } },
    { icon: FolderGit2, title: "Link Git Repo", desc: "Connect repository", onClick: () => setActiveView("projects") },
    { icon: Brain, title: "View Memory", desc: "Open memory", onClick: () => setActiveView("memory") },
  ];

  return (
    <section>
      <h3 className="mb-2 text-sm font-semibold text-text-main">Quick Actions</h3>
      <div className="grid grid-cols-2 gap-2">
        {actions.map(({ icon: Icon, title, desc, onClick }) => (
          <button key={title} onClick={onClick} className="group flex flex-col gap-2 rounded-xl border border-slate-800 bg-slate-900/70 p-3 text-left transition hover:border-purple-primary/60 hover:bg-slate-900">
            <Icon size={18} className="text-purple-hover" />
            <span className="min-w-0">
              <span className="block truncate text-sm font-medium text-text-main">{title}</span>
              <span className="block truncate text-xs text-text-muted">{desc}</span>
            </span>
          </button>
        ))}
      </div>
    </section>
  );
}
