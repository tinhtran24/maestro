import { Clock, FolderGit2, GitBranch, Pencil, Trees } from "lucide-react";
import type { LucideIcon } from "lucide-react";
import type { Project } from "../../domain/models";
import { Card } from "../../shared/ui/Card";
import { useWorkbenchStore } from "../../state/workbenchStore";

export function ProjectInfoCard({ project }: { project: Project }) {
  const setActiveView = useWorkbenchStore((state) => state.setActiveView);
  const branch = project.defaultBranch || project.settings.defaultBranch || "main";
  const worktreeRoot = project.worktreeRoot || project.settings.worktreeRoot || ".thanos/worktrees";
  const rows: Array<{ icon: LucideIcon; label: string; value: string; muted?: boolean }> = [
    { icon: FolderGit2, label: "Repository", value: project.gitRemoteUrl || "Not linked", muted: !project.gitRemoteUrl },
    { icon: GitBranch, label: "Branch", value: branch },
    { icon: Trees, label: "Worktree Root", value: worktreeRoot },
    { icon: Clock, label: "Last Updated", value: project.updatedAt || "-", muted: !project.updatedAt },
  ];
  return (
    <Card title="Project Info" action={<button onClick={() => setActiveView("projects")} className="inline-flex items-center gap-1 text-xs text-text-muted hover:text-text-main"><Pencil size={12} /> Edit</button>}>
      <dl className="grid gap-2.5">
        {rows.map(({ icon: Icon, label, value, muted }) => (
          <div key={label} className="flex items-center justify-between gap-3 text-sm">
            <dt className="inline-flex items-center gap-2 text-text-muted"><Icon size={14} /> {label}</dt>
            <dd className={`min-w-0 truncate text-right ${muted ? "text-text-muted" : "text-text-main"}`}>{value}</dd>
          </div>
        ))}
      </dl>
    </Card>
  );
}
