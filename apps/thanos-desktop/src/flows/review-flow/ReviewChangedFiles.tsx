import { FileCode2, FilePlus2, GitCompare } from "lucide-react";
import type { GitDiff, Review } from "../../domain/models";

// Changed files + diff summary for the review panel. Prefers the collected GitDiff
// (per-file add/modify status); falls back to the review's recorded file paths.
export function ReviewChangedFiles({ diff, review }: { diff?: GitDiff; review: Review }) {
  const files = diff?.changedFiles ?? review.changedFiles.map((path) => ({ path, status: "modified" as const }));
  const summary = diff?.summary || review.diffSummary || "No diff collected yet.";
  return (
    <div className="grid gap-3">
      <p className="inline-flex items-center gap-2 text-sm text-text-main">
        <GitCompare size={15} className="text-blue-info" /> {summary}
      </p>
      {files.length ? (
        <ul className="grid gap-1">
          {files.map((file) => (
            <li key={file.path} className="flex items-center gap-2 text-xs">
              {file.status === "added" ? (
                <FilePlus2 size={13} className="text-green-success" />
              ) : (
                <FileCode2 size={13} className="text-blue-info" />
              )}
              <span className="font-mono text-text-main">{file.path}</span>
              <span className="text-text-muted">{file.status}</span>
            </li>
          ))}
        </ul>
      ) : (
        <p className="text-xs text-text-muted">No changed files recorded.</p>
      )}
    </div>
  );
}
