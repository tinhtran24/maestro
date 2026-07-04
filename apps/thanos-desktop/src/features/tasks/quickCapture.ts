// Feature — AI Quick Capture.
// Pure, deterministic mock "AI extraction". Quick Capture is UI-first: there is no
// real LLM. Given pasted text + attachments, it infers a structured engineering
// task (title, priority, labels, acceptance criteria, likely files, risks, open
// questions) with per-field confidence. Deterministic so the UI and tests are
// stable — matches the planGenerator.ts style. No network, no execution.

import type { Priority } from "../../domain/models";

export type AttachmentKind = "image" | "figma" | "github" | "jira" | "link";

export type CapturedAttachment = {
  id: string;
  kind: AttachmentKind;
  label: string;
  url: string;
  previewUrl?: string;
};

export type ExtractedField<T> = { value: T; confidence: number };

export type ExtractedTask = {
  title: ExtractedField<string>;
  description: ExtractedField<string>;
  feature: ExtractedField<string>;
  priority: ExtractedField<Priority>;
  labels: ExtractedField<string[]>;
  assignedAgent: string;
  acceptanceCriteria: string[];
  technicalNotes: string;
  likelyFiles: string[];
  estimatedScope: string;
  risks: string[];
  openQuestions: string[];
  planSummary: string;
};

export const EXAMPLE_PROMPTS = [
  "Build shopping cart",
  "Fix login bug",
  "Dark mode",
  "Search feature",
  "Payment integration",
];

// Keyword → template mapping. First match wins; falls back to a generic template.
type Template = {
  match: RegExp;
  title: string;
  feature: string;
  priority: Priority;
  labels: string[];
  acceptanceCriteria: string[];
  likelyFiles: string[];
  risks: string[];
};

const TEMPLATES: Template[] = [
  {
    match: /\bcart|checkout|basket\b/i,
    title: "Implement Shopping Cart",
    feature: "E-commerce Core",
    priority: "P1",
    labels: ["cart", "checkout", "frontend"],
    acceptanceCriteria: ["Add product to cart", "Update quantity", "Remove product", "Checkout"],
    likelyFiles: ["components/CartSidebar.tsx", "context/CartContext.tsx", "api/cart/route.ts"],
    risks: ["Cart/backend sync conflicts", "Persistence across sessions"],
  },
  {
    match: /\blogin|auth|oauth|sign[\s-]?in|session\b/i,
    title: "Fix Authentication Flow",
    feature: "Auth & Security",
    priority: "P0",
    labels: ["auth", "security", "bug"],
    acceptanceCriteria: ["User can sign in", "OAuth redirect resolves correctly", "Session persists"],
    likelyFiles: ["auth/oauth.ts", "routes/session.ts", "middleware/auth.ts"],
    risks: ["Token/redirect misconfiguration", "Session fixation"],
  },
  {
    match: /\bdark[\s-]?mode|theme|appearance\b/i,
    title: "Add Dark Mode",
    feature: "Design System",
    priority: "P2",
    labels: ["theming", "frontend", "ui"],
    acceptanceCriteria: ["Toggle light/dark", "Preference persists", "Respects system setting"],
    likelyFiles: ["theme.css", "shared/ui/ThemeToggle.tsx"],
    risks: ["Contrast/accessibility regressions"],
  },
  {
    match: /\bpayment|stripe|billing|invoice\b/i,
    title: "Payment Gateway Integration",
    feature: "Payments",
    priority: "P1",
    labels: ["payment", "stripe", "backend"],
    acceptanceCriteria: ["Create payment intent", "Handle success + failure", "Record transaction"],
    likelyFiles: ["api/payments/route.ts", "services/paymentProvider.ts"],
    risks: ["PCI compliance", "Failure/refund edge cases"],
  },
  {
    match: /\bsearch|filter|query|index\b/i,
    title: "Add Search Feature",
    feature: "Discovery",
    priority: "P2",
    labels: ["search", "frontend", "backend"],
    acceptanceCriteria: ["Query returns relevant results", "Empty + error states", "Debounced input"],
    likelyFiles: ["components/SearchBar.tsx", "api/search/route.ts"],
    risks: ["Result relevance", "Performance at scale"],
  },
];

const GENERIC: Template = {
  match: /.*/,
  title: "New Task",
  feature: "General",
  priority: "P2",
  labels: ["new"],
  acceptanceCriteria: ["Behavior is implemented", "Covered by tests"],
  likelyFiles: ["src/"],
  risks: ["Requirements need clarification"],
};

function clampConfidence(value: number): number {
  return Math.max(0, Math.min(100, Math.round(value)));
}

function firstSentence(input: string): string {
  const trimmed = input.trim();
  if (!trimmed) return "";
  const sentence = trimmed.split(/(?<=[.!?])\s|\n/)[0];
  return sentence.length > 120 ? `${sentence.slice(0, 117)}…` : sentence;
}

function titleCaseFallback(input: string): string {
  const words = input.trim().split(/\s+/).slice(0, 6).join(" ");
  return words ? words.charAt(0).toUpperCase() + words.slice(1) : "New Task";
}

// Classifies a pasted URL into an attachment kind.
export function classifyUrl(url: string): AttachmentKind {
  if (/figma\.com/i.test(url)) return "figma";
  if (/github\.com/i.test(url)) return "github";
  if (/atlassian\.net|jira/i.test(url)) return "jira";
  return "link";
}

// Synthesizes a structured task from pasted input + attachments. Deterministic.
export function extractTask(input: string, attachments: CapturedAttachment[] = []): ExtractedTask {
  const text = input.trim();
  const template = TEMPLATES.find((item) => item.match.test(text)) ?? GENERIC;
  const matched = template !== GENERIC;

  // Confidence scales with input richness: keyword match, length, attachments.
  const lengthBoost = Math.min(30, Math.floor(text.length / 8));
  const attachmentBoost = Math.min(15, attachments.length * 5);
  const base = matched ? 60 : 35;
  const confidence = clampConfidence(base + lengthBoost + attachmentBoost);

  const title = matched ? template.title : text ? titleCaseFallback(text) : "New Task";
  const description = firstSentence(text) || "Captured from Quick Capture. Add detail before implementation.";

  const openQuestions: string[] = [];
  if (!matched) openQuestions.push("What is the primary user goal?");
  if (text.length < 24) openQuestions.push("Can you add more detail or acceptance criteria?");
  if (attachments.length === 0) openQuestions.push("Any screenshots, designs, or links to attach?");

  const attachmentLabels = attachments.map((attachment) => attachment.kind);

  return {
    title: { value: title, confidence },
    description: { value: description, confidence: clampConfidence(confidence - 10) },
    feature: { value: template.feature, confidence: clampConfidence(confidence - 5) },
    priority: { value: template.priority, confidence: matched ? clampConfidence(confidence - 5) : 40 },
    labels: { value: dedupe([...template.labels, ...attachmentLabels]), confidence: clampConfidence(confidence - 8) },
    assignedAgent: "Planner / Claude Code",
    acceptanceCriteria: template.acceptanceCriteria,
    technicalNotes: matched
      ? `Follow existing conventions in ${template.likelyFiles[0] ?? "the affected modules"}.`
      : "Clarify scope, then implement in the affected modules.",
    likelyFiles: template.likelyFiles,
    estimatedScope: matched ? (template.priority === "P0" ? "Medium–Large" : "Small–Medium") : "Unknown",
    risks: template.risks,
    openQuestions,
    planSummary: `Plan for ${title}. ${description}`,
  };
}

function dedupe(values: string[]): string[] {
  return Array.from(new Set(values.filter(Boolean)));
}

// Maps the (possibly edited) extracted task onto the fields the Task model supports.
// Acceptance criteria + technical notes are folded into the description as markdown
// so no captured detail is lost without changing the domain model.
export function toCreateInput(extracted: ExtractedTask): {
  title: string;
  description: string;
  priority: Priority;
  assignedAgent: string;
  tags: string[];
} {
  const parts = [extracted.description.value.trim()];
  if (extracted.acceptanceCriteria.length) {
    parts.push(["", "**Acceptance Criteria**", ...extracted.acceptanceCriteria.map((item) => `- ${item}`)].join("\n"));
  }
  if (extracted.technicalNotes.trim()) {
    parts.push(["", "**Technical Notes**", extracted.technicalNotes.trim()].join("\n"));
  }
  return {
    title: extracted.title.value.trim() || "Untitled task",
    description: parts.filter(Boolean).join("\n"),
    priority: extracted.priority.value,
    assignedAgent: extracted.assignedAgent.trim(),
    tags: extracted.labels.value.length ? extracted.labels.value : ["new"],
  };
}
