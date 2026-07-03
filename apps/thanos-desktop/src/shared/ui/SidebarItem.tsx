import type { LucideIcon } from "lucide-react";

export function SidebarItem({ icon: Icon, label, active = false, onClick }: { icon: LucideIcon; label: string; active?: boolean; onClick?: () => void }) {
  return (
    <button
      onClick={onClick}
      className={`flex h-9 w-max shrink-0 items-center gap-2 rounded-lg px-3 text-sm transition lg:w-full lg:gap-3 ${
        active ? "bg-purple-primary text-white" : "text-text-muted hover:bg-slate-900 hover:text-text-main"
      }`}
    >
      <Icon size={18} />
      <span>{label}</span>
    </button>
  );
}
