import { FileCode2, FilePlus2, GitCompare } from "lucide-react";
import type { GitDiff } from "../../domain/models";

// Shows the coder's changeset (from the persisted mock GitDiff): a summary line
// and the changed files, distinguishing added from modified.
export function CodingChangeset({ diff }: { diff?: GitDiff }) {
  if (!diff) {
    return <p className="text-xs text-text-muted">The coder has not produced any changes yet.</p>;
  }
  return (
    <div className="grid gap-3">
      <p className="inline-flex items-center gap-2 text-sm text-text-main">
        <GitCompare size={15} className="text-green-success" /> {diff.summary}
      </p>
      <ul className="grid gap-1">
        {diff.changedFiles.map((file) => (
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
    </div>
  );
}
