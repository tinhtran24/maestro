import { ImagePlus, Link, Upload } from "lucide-react";
import { useRef, useState } from "react";

// Drag-and-drop / upload zone plus a paste-URL input. Client-side only — files
// become object-URL previews via the parent; nothing is uploaded.
export function AttachmentDropzone({
  onFiles,
  onUrl,
}: {
  onFiles: (files: FileList | File[]) => void;
  onUrl: (url: string) => void;
}) {
  const inputRef = useRef<HTMLInputElement | null>(null);
  const [dragOver, setDragOver] = useState(false);
  const [url, setUrl] = useState("");

  function submitUrl() {
    if (!url.trim()) return;
    onUrl(url);
    setUrl("");
  }

  return (
    <div className="grid gap-2">
      <div
        onDragOver={(event) => { event.preventDefault(); setDragOver(true); }}
        onDragLeave={() => setDragOver(false)}
        onDrop={(event) => {
          event.preventDefault();
          setDragOver(false);
          if (event.dataTransfer.files.length) onFiles(event.dataTransfer.files);
        }}
        className={`grid place-items-center rounded-lg border border-dashed p-4 text-center text-xs transition ${dragOver ? "border-purple-primary bg-purple-primary/10 text-purple-hover" : "border-slate-700 bg-slate-950/40 text-text-muted"}`}
      >
        <span className="inline-flex items-center gap-2"><ImagePlus size={15} /> Drag & drop images, or</span>
        <button onClick={() => inputRef.current?.click()} className="mt-2 inline-flex items-center gap-2 rounded-lg border border-slate-800 bg-slate-900/80 px-2.5 py-1.5 text-text-main hover:border-slate-600">
          <Upload size={13} /> Upload image
        </button>
        <input
          ref={inputRef}
          type="file"
          accept="image/*"
          multiple
          className="hidden"
          onChange={(event) => { if (event.target.files) onFiles(event.target.files); event.target.value = ""; }}
        />
      </div>
      <div className="flex items-center gap-2">
        <div className="relative flex-1">
          <Link size={14} className="pointer-events-none absolute left-2.5 top-1/2 -translate-y-1/2 text-text-muted" />
          <input
            value={url}
            onChange={(event) => setUrl(event.target.value)}
            onKeyDown={(event) => { if (event.key === "Enter") { event.preventDefault(); submitUrl(); } }}
            placeholder="Paste a Figma / GitHub / Jira link…"
            className="h-9 w-full rounded-lg border border-slate-800 bg-slate-950/60 pl-8 pr-3 text-sm text-text-main placeholder:text-text-muted focus:border-slate-600 focus:outline-none"
          />
        </div>
        <button onClick={submitUrl} className="rounded-lg border border-slate-800 bg-slate-900/80 px-3 py-2 text-sm text-text-main hover:border-slate-600">Add</button>
      </div>
    </div>
  );
}
