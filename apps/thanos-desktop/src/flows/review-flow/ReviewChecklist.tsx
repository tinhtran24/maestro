import { Check, Circle } from "lucide-react";
import type { ChecklistItem } from "../../features/reviewer/reviewChecklist";

export function ReviewChecklist({ items }: { items: ChecklistItem[] }) {
  return (
    <ul className="grid gap-1.5">
      {items.map((item) => (
        <li key={item.id} className="flex items-center gap-2 text-sm">
          <span className={`grid h-4 w-4 place-items-center rounded-full ${item.done ? "bg-green-success/20 text-green-success" : "text-slate-500"}`}>
            {item.done ? <Check size={12} /> : <Circle size={9} />}
          </span>
          <span className={item.done ? "text-text-main" : "text-text-muted"}>{item.label}</span>
        </li>
      ))}
    </ul>
  );
}
