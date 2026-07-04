import { GitBranch, FolderGit2 } from "lucide-react";
import type { Task } from "../../domain/models";

// Makes the mock "Create Worktree" step visible: the isolated worktree path and
// branch assigned when the task entered `running`.
export function WorktreeCard({ task }: { task: Task }) {
  return (
    <div className="grid gap-2 rounded-lg border border-slate-800 bg-slate-950/40 p-3 text-xs">
      <div className="flex items-center justify-between gap-3">
        <span className="inline-flex items-center gap-2 text-text-muted"><FolderGit2 size={13} /> Worktree</span>
        <span className="truncate font-mono text-text-main">{task.worktreePath || "—"}</span>
      </div>
      <div className="flex items-center justify-between gap-3">
        <span className="inline-flex items-center gap-2 text-text-muted"><GitBranch size={13} /> Branch</span>
        <span className="truncate font-mono text-text-main">{task.branchName || "—"}</span>
      </div>
    </div>
  );
}
