import { ClipboardPaste, ImagePlus, Link as LinkIcon, MousePointerSquareDashed, Sparkles } from "lucide-react";
import { type ClipboardEvent, useRef, useState } from "react";
import { Button } from "../ui/button";
import { cn } from "../../lib/utils";
import { AttachmentDropzone } from "./AttachmentDropzone";
import { ImagePreviewGrid } from "./ImagePreviewGrid";
import { fileToAttachment, urlToAttachment } from "./attachments";
import { classifyUrl, type Attachment } from "./types";

const EXAMPLE_CHIPS = [
	"Build shopping cart",
	"Fix login bug",
	"Add dark mode",
	"Add search feature",
	"Payment integration",
];

const TABS = [
	{ id: "text", label: "Paste Text", icon: ClipboardPaste },
	{ id: "image", label: "Paste Image", icon: ImagePlus },
	{ id: "drop", label: "Drag & Drop", icon: MousePointerSquareDashed },
	{ id: "link", label: "Add Link", icon: LinkIcon },
] as const;

// QuickCaptureEditor is Step 1: paste anything (text, images, links, files).
export function QuickCaptureEditor({
	input,
	onInput,
	attachments,
	onAddFiles,
	onAddUrl,
	onRemove,
	onReorder,
}: {
	input: string;
	onInput: (value: string) => void;
	attachments: Attachment[];
	onAddFiles: (files: File[]) => void;
	onAddUrl: (url: string) => void;
	onRemove: (id: string) => void;
	onReorder: (id: string, dir: -1 | 1) => void;
}) {
	const [tab, setTab] = useState<(typeof TABS)[number]["id"]>("text");
	const [linkValue, setLinkValue] = useState("");
	const fileInputRef = useRef<HTMLInputElement>(null);

	const handlePaste = (e: ClipboardEvent<HTMLTextAreaElement>) => {
		const imageFiles = Array.from(e.clipboardData.files ?? []).filter((f) => f.type.startsWith("image/"));
		if (imageFiles.length) {
			e.preventDefault();
			onAddFiles(imageFiles);
			return;
		}
		const text = e.clipboardData.getData("text");
		if (text && /^https?:\/\/\S+$/.test(text.trim()) && input.trim() === "") {
			// A bare URL pasted into an empty editor becomes a link attachment.
			e.preventDefault();
			onAddUrl(text.trim());
		}
	};

	const submitLink = () => {
		const v = linkValue.trim();
		if (!v) return;
		onAddUrl(v);
		setLinkValue("");
	};

	return (
		<div className="space-y-3">
			<div className="flex items-center gap-2 text-[13px]">
				<Sparkles className="size-4 text-violet-400" aria-hidden="true" />
				<span className="font-medium text-foreground">Quick Capture</span>
				<span className="rounded-full border border-violet-500/30 bg-violet-500/10 px-1.5 py-0.5 text-[10px] font-medium text-violet-300">
					AI-powered
				</span>
			</div>

			{/* Tab row */}
			<div className="flex flex-wrap gap-1 border-b border-border pb-2">
				{TABS.map(({ id, label, icon: Icon }) => (
					<button
						key={id}
						type="button"
						onClick={() => {
							setTab(id);
							if (id === "image") fileInputRef.current?.click();
						}}
						className={cn(
							"inline-flex items-center gap-1.5 rounded-md px-2 py-1 text-[12px] transition",
							tab === id ? "bg-violet-500/15 text-violet-200" : "text-muted-foreground hover:bg-surface hover:text-foreground",
						)}
					>
						<Icon className="size-3.5" aria-hidden="true" />
						{label}
					</button>
				))}
			</div>

			<AttachmentDropzone onFiles={onAddFiles}>
				<textarea
					autoFocus
					className="min-h-[168px] w-full resize-y rounded-lg border border-border bg-transparent px-3 py-2.5 text-[13px] leading-relaxed text-foreground outline-none transition placeholder:text-passive focus-visible:border-violet-500/70 focus-visible:ring-2 focus-visible:ring-violet-500/25"
					placeholder={
						"Paste anything…\n\n• User requirements\n• Screenshots (⌘V / Ctrl+V)\n• Figma links\n• GitHub Issues\n• Markdown\n• Product specs"
					}
					value={input}
					onChange={(e) => onInput(e.target.value)}
					onPaste={handlePaste}
				/>
			</AttachmentDropzone>

			{tab === "link" ? (
				<div className="flex items-center gap-2">
					<input
						type="url"
						value={linkValue}
						onChange={(e) => setLinkValue(e.target.value)}
						onKeyDown={(e) => e.key === "Enter" && (e.preventDefault(), submitLink())}
						placeholder="Paste a Figma / GitHub / Jira URL…"
						className="h-8 flex-1 rounded-md border border-border bg-transparent px-2.5 text-[12px] text-foreground outline-none focus-visible:border-violet-500/70"
					/>
					<Button type="button" variant="secondary" size="sm" onClick={submitLink}>
						Attach link
					</Button>
				</div>
			) : null}

			<ImagePreviewGrid attachments={attachments} onRemove={onRemove} onReorder={onReorder} />

			<input
				ref={fileInputRef}
				type="file"
				multiple
				accept="image/png,image/jpeg,image/webp,application/pdf,.md,.markdown,text/*"
				className="hidden"
				onChange={(e) => {
					const files = Array.from(e.target.files ?? []);
					if (files.length) onAddFiles(files);
					e.target.value = "";
				}}
			/>

			<div className="space-y-1.5">
				<span className="text-[11px] text-muted-foreground">Examples</span>
				<div className="flex flex-wrap gap-1.5">
					{EXAMPLE_CHIPS.map((chip) => (
						<button
							key={chip}
							type="button"
							onClick={() => onInput(input ? `${input}\n${chip}` : chip)}
							className="rounded-full border border-border bg-surface px-2.5 py-1 text-[11px] text-muted-foreground transition hover:border-violet-500/40 hover:text-foreground"
						>
							{chip}
						</button>
					))}
				</div>
			</div>
		</div>
	);
}

export { classifyUrl, fileToAttachment, urlToAttachment };
