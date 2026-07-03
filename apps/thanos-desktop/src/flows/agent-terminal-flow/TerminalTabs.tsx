import { Pin, X } from "lucide-react";
import type { AgentSession } from "../../domain/models";
import { terminalLabel } from "../../features/terminal/mockRuntime";

const statusDot: Record<AgentSession["status"], string> = {
  idle: "bg-slate-600",
  starting: "bg-blue-info animate-pulse",
  running: "bg-green-success animate-pulse",
  waiting_user: "bg-yellow-warning",
  completed: "bg-green-success",
  stopping: "bg-yellow-warning",
  stopped: "bg-yellow-warning",
  failed: "bg-red-danger",
};

export function TerminalTabs({
  sessions,
  activeId,
  onSelect,
  onPin,
  onClose,
  isPinned,
}: {
  sessions: AgentSession[];
  activeId?: string;
  onSelect: (id: string) => void;
  onPin: (id: string) => void;
  onClose: (id: string) => void;
  isPinned: (id: string) => boolean;
}) {
  if (sessions.length === 0) return null;
  return (
    <div className="flex items-center gap-1 overflow-x-auto border-b border-slate-800 p-2">
      {sessions.map((session) => {
        const pinned = isPinned(session.id);
        const active = session.id === activeId;
        return (
          <div
            key={session.id}
            className={`group inline-flex shrink-0 items-center gap-2 rounded-md border px-2 py-1 text-xs ${active ? "border-blue-info/60 bg-blue-info/10 text-text-main" : "border-slate-800 text-text-muted hover:border-slate-700"}`}
          >
            <button onClick={() => onSelect(session.id)} className="inline-flex items-center gap-2">
              <span className={`h-1.5 w-1.5 rounded-full ${statusDot[session.status]}`} />
              {terminalLabel(session.step, session.provider)}
            </button>
            <button
              onClick={() => onPin(session.id)}
              title={pinned ? "Unpin" : "Pin"}
              className={`rounded p-0.5 ${pinned ? "text-blue-info" : "text-text-muted opacity-0 hover:text-text-main group-hover:opacity-100"}`}
            >
              <Pin size={11} className={pinned ? "fill-blue-info" : ""} />
            </button>
            {!pinned && (
              <button onClick={() => onClose(session.id)} title="Close" className="rounded p-0.5 text-text-muted opacity-0 hover:text-red-danger group-hover:opacity-100">
                <X size={11} />
              </button>
            )}
          </div>
        );
      })}
    </div>
  );
}
