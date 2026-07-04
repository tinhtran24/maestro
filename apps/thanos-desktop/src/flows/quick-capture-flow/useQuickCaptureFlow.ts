// Feature — AI Quick Capture (mock).
// Orchestrates the 4-step create-task wizard: Quick Capture → AI Structure →
// Review & Edit → Create. AI extraction is mocked (no LLM). On create, the task is
// added and opened; the planner is NOT auto-started (per feature scope).

import { useRef, useState } from "react";
import type { Priority } from "../../domain/models";
import {
  extractTask,
  classifyUrl,
  toCreateInput,
  type CapturedAttachment,
  type ExtractedTask,
} from "../../features/tasks/quickCapture";
import { NativeBackend } from "../../services/nativeBackend";
import { useWorkbenchStore } from "../../state/workbenchStore";

export type CaptureStep = "capture" | "structure" | "review" | "create";

const ORDER: CaptureStep[] = ["capture", "structure", "review", "create"];
const backend = new NativeBackend();

export function useQuickCaptureFlow() {
  const createTask = useWorkbenchStore((state) => state.createTask);
  const closeTaskDialog = useWorkbenchStore((state) => state.closeTaskDialog);
  const setActiveView = useWorkbenchStore((state) => state.setActiveView);

  const [step, setStep] = useState<CaptureStep>("capture");
  const [input, setInput] = useState("");
  const [attachments, setAttachments] = useState<CapturedAttachment[]>([]);
  const [draft, setDraft] = useState<ExtractedTask | null>(null);
  const counter = useRef(0);

  const canContinue = step === "capture" ? input.trim().length > 0 || attachments.length > 0 : true;

  function nextId() {
    counter.current += 1;
    return `att-${counter.current}`;
  }

  function appendExample(text: string) {
    setInput((current) => (current.trim() ? `${current.trim()}\n${text}` : text));
  }

  function addFiles(files: FileList | File[]) {
    const next: CapturedAttachment[] = [];
    for (const file of Array.from(files)) {
      if (!file.type.startsWith("image/")) continue;
      const url = URL.createObjectURL(file);
      next.push({ id: nextId(), kind: "image", label: file.name, url, previewUrl: url });
    }
    if (next.length) setAttachments((current) => [...current, ...next]);
  }

  function addUrl(raw: string) {
    const url = raw.trim();
    if (!url) return;
    const kind = classifyUrl(url);
    setAttachments((current) => [...current, { id: nextId(), kind, label: url, url }]);
  }

  function removeAttachment(id: string) {
    setAttachments((current) => current.filter((attachment) => attachment.id !== id));
  }

  function moveAttachment(id: string, direction: -1 | 1) {
    setAttachments((current) => {
      const index = current.findIndex((attachment) => attachment.id === id);
      const target = index + direction;
      if (index < 0 || target < 0 || target >= current.length) return current;
      const next = current.slice();
      [next[index], next[target]] = [next[target], next[index]];
      return next;
    });
  }

  function runExtraction() {
    setDraft(extractTask(input, attachments));
  }

  function next() {
    const index = ORDER.indexOf(step);
    if (index >= ORDER.length - 1) return;
    // Leaving Quick Capture runs the mock extraction (or re-runs it if edited).
    if (step === "capture") runExtraction();
    setStep(ORDER[index + 1]);
  }

  function back() {
    const index = ORDER.indexOf(step);
    if (index > 0) setStep(ORDER[index - 1]);
  }

  function goTo(target: CaptureStep) {
    // Only allow jumping to steps that are already reachable (extraction done).
    if (target !== "capture" && !draft) return;
    setStep(target);
  }

  // Field editors for the AI Structure / Review steps (value edits keep confidence).
  function patchTitle(value: string) {
    setDraft((current) => (current ? { ...current, title: { ...current.title, value } } : current));
  }
  function patchDescription(value: string) {
    setDraft((current) => (current ? { ...current, description: { ...current.description, value } } : current));
  }
  function patchPriority(value: Priority) {
    setDraft((current) => (current ? { ...current, priority: { ...current.priority, value } } : current));
  }
  function patchLabels(value: string[]) {
    setDraft((current) => (current ? { ...current, labels: { ...current.labels, value } } : current));
  }
  function patchAcceptance(value: string[]) {
    setDraft((current) => (current ? { ...current, acceptanceCriteria: value } : current));
  }
  function patchTechnicalNotes(value: string) {
    setDraft((current) => (current ? { ...current, technicalNotes: value } : current));
  }

  function create() {
    if (!draft) return;
    const task = createTask(toCreateInput(draft));
    // Persist to the workbench store (SQLite) so the task survives a reload.
    void backend.saveTask(task);
    setActiveView("workbench");
    closeTaskDialog();
  }

  return {
    step,
    input,
    attachments,
    draft,
    canContinue,
    setInput,
    appendExample,
    addFiles,
    addUrl,
    removeAttachment,
    moveAttachment,
    next,
    back,
    goTo,
    patchTitle,
    patchDescription,
    patchPriority,
    patchLabels,
    patchAcceptance,
    patchTechnicalNotes,
    create,
    close: closeTaskDialog,
  };
}
