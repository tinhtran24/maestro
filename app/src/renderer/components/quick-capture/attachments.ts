import type { Attachment, AttachmentKind } from "./types";

let counter = 0;
function nextId(): string {
	counter += 1;
	return `att-${counter}-${Math.round(performance.now())}`;
}

export function formatBytes(bytes: number): string {
	if (bytes < 1024) return `${bytes} B`;
	if (bytes < 1024 * 1024) return `${Math.round(bytes / 1024)} KB`;
	return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}

function kindForFile(file: File): AttachmentKind {
	if (file.type.startsWith("image/")) return "image";
	if (file.type === "application/pdf" || file.name.endsWith(".pdf")) return "pdf";
	if (file.name.endsWith(".md") || file.name.endsWith(".markdown")) return "markdown";
	if (file.type.startsWith("text/")) return "text";
	return "file";
}

// fileToAttachment turns a dropped/pasted/uploaded File into an Attachment,
// generating an object URL preview for images.
export function fileToAttachment(file: File): Attachment {
	const kind = kindForFile(file);
	return {
		id: nextId(),
		kind,
		name: file.name || `${kind}-${Date.now()}`,
		previewUrl: kind === "image" ? URL.createObjectURL(file) : undefined,
		sizeLabel: formatBytes(file.size),
	};
}

// urlToAttachment turns a pasted URL into a link attachment.
export function urlToAttachment(url: string, kind: AttachmentKind): Attachment {
	let name = url;
	try {
		const u = new URL(url);
		name = `${u.hostname}${u.pathname}`.replace(/\/$/, "");
	} catch {
		/* keep raw url as name */
	}
	return { id: nextId(), kind, name, url };
}
