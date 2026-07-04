import { ArrowLeft, ArrowRight, Figma, Github, ImageIcon, Link2, Trello, X } from "lucide-react";
import type { AttachmentKind, CapturedAttachment } from "../../features/tasks/quickCapture";

const KIND_ICON: Record<AttachmentKind, typeof Link2> = {
  image: ImageIcon,
  figma: Figma,
  github: Github,
  jira: Trello,
  link: Link2,
};

// Renders captured attachments: image thumbnails and classified link chips, each
// with remove + reorder controls.
export function ImagePreviewGrid({
  attachments,
  onRemove,
  onMove,
}: {
  attachments: CapturedAttachment[];
  onRemove: (id: string) => void;
  onMove: (id: string, direction: -1 | 1) => void;
}) {
  if (!attachments.length) return null;
  return (
    <ul className="grid grid-cols-2 gap-2 sm:grid-cols-3">
      {attachments.map((attachment, index) => {
        const Icon = KIND_ICON[attachment.kind];
        return (
          <li key={attachment.id} className="group relative overflow-hidden rounded-lg border border-slate-800 bg-slate-950/60">
            {attachment.previewUrl ? (
              <img src={attachment.previewUrl} alt={attachment.label} className="h-24 w-full object-cover" />
            ) : (
              <div className="flex h-24 flex-col items-center justify-center gap-1 p-2 text-center">
                <Icon size={18} className="text-text-muted" />
                <span className="line-clamp-2 break-all text-[11px] text-text-muted">{attachment.label}</span>
              </div>
            )}
            <div className="flex items-center justify-between gap-1 border-t border-slate-800 bg-slate-950/80 px-1.5 py-1">
              <span className="inline-flex items-center gap-1 text-[10px] capitalize text-text-muted"><Icon size={11} /> {attachment.kind}</span>
              <div className="flex items-center gap-0.5">
                <button onClick={() => onMove(attachment.id, -1)} disabled={index === 0} className="rounded p-0.5 text-text-muted hover:text-text-main disabled:opacity-30" aria-label="Move left"><ArrowLeft size={12} /></button>
                <button onClick={() => onMove(attachment.id, 1)} disabled={index === attachments.length - 1} className="rounded p-0.5 text-text-muted hover:text-text-main disabled:opacity-30" aria-label="Move right"><ArrowRight size={12} /></button>
                <button onClick={() => onRemove(attachment.id)} className="rounded p-0.5 text-red-danger hover:text-red-danger" aria-label="Remove"><X size={12} /></button>
              </div>
            </div>
          </li>
        );
      })}
    </ul>
  );
}
