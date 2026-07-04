import { Sparkles } from "lucide-react";
import type { CapturedAttachment } from "../../features/tasks/quickCapture";
import { EXAMPLE_PROMPTS } from "../../features/tasks/quickCapture";
import { AttachmentDropzone } from "./AttachmentDropzone";
import { ImagePreviewGrid } from "./ImagePreviewGrid";

// Step 1 — Quick Capture. Paste anything: text, images (paste/drop/upload), or
// links. Example chips seed the input.
export function QuickCaptureEditor({
  input,
  onInput,
  onExample,
  attachments,
  onFiles,
  onUrl,
  onRemove,
  onMove,
}: {
  input: string;
  onInput: (value: string) => void;
  onExample: (text: string) => void;
  attachments: CapturedAttachment[];
  onFiles: (files: FileList | File[]) => void;
  onUrl: (url: string) => void;
  onRemove: (id: string) => void;
  onMove: (id: string, direction: -1 | 1) => void;
}) {
  return (
    <div className="grid gap-4">
      <div>
        <p className="mb-2 inline-flex items-center gap-2 text-sm text-text-muted">
          <Sparkles size={15} className="text-purple-hover" /> Paste anything — requirements, a screenshot, a Figma or GitHub link. AI will structure it.
        </p>
        <textarea
          value={input}
          onChange={(event) => onInput(event.target.value)}
          onPaste={(event) => {
            const files = Array.from(event.clipboardData.files).filter((file) => file.type.startsWith("image/"));
            if (files.length) { event.preventDefault(); onFiles(files); }
          }}
          rows={7}
          autoFocus
          placeholder={"Paste anything...\n\n• User requirements\n• Screenshots\n• Figma links\n• GitHub Issues\n• Markdown\n• Product specs"}
          className="w-full resize-y rounded-lg border border-slate-800 bg-slate-950/60 px-3 py-2 text-sm text-text-main placeholder:text-text-muted focus:border-slate-600 focus:outline-none"
        />
      </div>

      <div className="flex flex-wrap gap-1.5">
        {EXAMPLE_PROMPTS.map((prompt) => (
          <button key={prompt} onClick={() => onExample(prompt)} className="rounded-lg border border-slate-800 bg-slate-900/70 px-2.5 py-1 text-xs text-text-muted hover:border-slate-600 hover:text-text-main">
            {prompt}
          </button>
        ))}
      </div>

      <AttachmentDropzone onFiles={onFiles} onUrl={onUrl} />
      <ImagePreviewGrid attachments={attachments} onRemove={onRemove} onMove={onMove} />
    </div>
  );
}
