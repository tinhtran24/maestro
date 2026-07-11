import { type ReactNode, useState } from "react";
import { cn } from "../../lib/utils";

// AttachmentDropzone wraps its children and captures files dropped anywhere
// inside it, highlighting the drop target. Paste handling lives on the textarea.
export function AttachmentDropzone({
	onFiles,
	children,
	className,
}: {
	onFiles: (files: File[]) => void;
	children: ReactNode;
	className?: string;
}) {
	const [dragging, setDragging] = useState(false);

	return (
		<div
			className={cn("relative rounded-lg transition", dragging && "ring-2 ring-violet-500/60", className)}
			onDragOver={(e) => {
				e.preventDefault();
				if (!dragging) setDragging(true);
			}}
			onDragLeave={(e) => {
				e.preventDefault();
				if (e.currentTarget === e.target) setDragging(false);
			}}
			onDrop={(e) => {
				e.preventDefault();
				setDragging(false);
				const files = Array.from(e.dataTransfer.files ?? []);
				if (files.length) onFiles(files);
			}}
		>
			{children}
			{dragging ? (
				<div className="pointer-events-none absolute inset-0 z-10 grid place-items-center rounded-lg bg-violet-500/10 text-[13px] font-medium text-violet-200">
					Drop files to attach
				</div>
			) : null}
		</div>
	);
}
