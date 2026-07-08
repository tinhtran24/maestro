import { ArrowLeft, ArrowRight, FileText, Link as LinkIcon, X } from "lucide-react";
import type { Attachment } from "./types";
import { cn } from "../../lib/utils";

function AttachmentIcon({ kind }: { kind: Attachment["kind"] }) {
	if (kind === "figma" || kind === "github" || kind === "url" || kind === "jira") {
		return <LinkIcon className="size-4" aria-hidden="true" />;
	}
	return <FileText className="size-4" aria-hidden="true" />;
}

// ImagePreviewGrid shows attachment thumbnails/chips with remove + reorder.
export function ImagePreviewGrid({
	attachments,
	onRemove,
	onReorder,
	className,
}: {
	attachments: Attachment[];
	onRemove: (id: string) => void;
	onReorder?: (id: string, dir: -1 | 1) => void;
	className?: string;
}) {
	if (attachments.length === 0) return null;
	return (
		<div className={cn("grid grid-cols-2 gap-2 sm:grid-cols-3", className)}>
			{attachments.map((a, i) => (
				<div
					key={a.id}
					className="group relative overflow-hidden rounded-lg border border-border bg-surface"
				>
					{a.previewUrl ? (
						<img src={a.previewUrl} alt={a.name} className="h-24 w-full object-cover" />
					) : (
						<div className="flex h-24 w-full flex-col items-center justify-center gap-1 px-2 text-center text-muted-foreground">
							<AttachmentIcon kind={a.kind} />
							<span className="line-clamp-2 text-[10px] leading-tight">{a.name}</span>
						</div>
					)}
					<div className="flex items-center justify-between gap-1 border-t border-border px-1.5 py-1">
						<span className="truncate text-[10px] text-muted-foreground" title={a.name}>
							{a.sizeLabel ?? a.kind}
						</span>
						<div className="flex shrink-0 items-center gap-0.5">
							{onReorder && i > 0 ? (
								<button
									type="button"
									aria-label="Move left"
									className="grid size-5 place-items-center rounded text-muted-foreground hover:bg-bg-2 hover:text-foreground"
									onClick={() => onReorder(a.id, -1)}
								>
									<ArrowLeft className="size-3" />
								</button>
							) : null}
							{onReorder && i < attachments.length - 1 ? (
								<button
									type="button"
									aria-label="Move right"
									className="grid size-5 place-items-center rounded text-muted-foreground hover:bg-bg-2 hover:text-foreground"
									onClick={() => onReorder(a.id, 1)}
								>
									<ArrowRight className="size-3" />
								</button>
							) : null}
							<button
								type="button"
								aria-label={`Remove ${a.name}`}
								className="grid size-5 place-items-center rounded text-muted-foreground hover:bg-rose-500/15 hover:text-rose-300"
								onClick={() => onRemove(a.id)}
							>
								<X className="size-3" />
							</button>
						</div>
					</div>
				</div>
			))}
		</div>
	);
}
