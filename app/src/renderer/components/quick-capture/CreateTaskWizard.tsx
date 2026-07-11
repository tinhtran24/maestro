import * as Dialog from "@radix-ui/react-dialog";
import { useMutation, useQuery } from "@tanstack/react-query";
import { ArrowRight, Bot, Check, ClipboardPlus, Loader2, Sparkles, X } from "lucide-react";
import { useEffect, useMemo, useState } from "react";
import type { components } from "../../../api/schema";
import { apiClient, apiErrorMessage } from "../../lib/api-client";
import { captureRendererEvent } from "../../lib/telemetry";
import { composeSessionPrompt, extractTask } from "../../lib/planner";
import { agentsQueryOptions } from "../../hooks/useAgentsQuery";
import type { AgentProvider } from "../../types/workspace";
import { Button } from "../ui/button";
import { RequiredAgentField } from "../CreateProjectAgentSheet";
import { cn } from "../../lib/utils";
import { AIExtractionCard } from "./AIExtractionCard";
import { QuickCaptureEditor } from "./QuickCaptureEditor";
import { TaskReviewFlow } from "./TaskReviewFlow";
import { ConfidenceBadge } from "./ConfidenceBadge";
import { fileToAttachment, urlToAttachment } from "./attachments";
import { classifyUrl, emptyDraft, type Attachment, type TaskDraft, type WizardStep } from "./types";

type Project = components["schemas"]["Project"];

type Props = {
	open: boolean;
	projectId?: string;
	onCreated: (sessionId: string) => void;
	onOpenChange: (open: boolean) => void;
};

const STEPS: { id: WizardStep; label: string; sub: string; icon: typeof Sparkles }[] = [
	{ id: "capture", label: "Quick Capture", sub: "AI-powered", icon: Sparkles },
	{ id: "structure", label: "AI Structure", sub: "Auto-extracted", icon: Bot },
	{ id: "review", label: "Review & Edit", sub: "Confirm details", icon: ClipboardPlus },
	{ id: "create", label: "Create", sub: "Start planning", icon: Check },
];

export function CreateTaskWizard({ open, projectId, onCreated, onOpenChange }: Props) {
	const [step, setStep] = useState<WizardStep>("capture");
	const [input, setInput] = useState("");
	const [attachments, setAttachments] = useState<Attachment[]>([]);
	const [draft, setDraft] = useState<TaskDraft>(emptyDraft());
	const [extraContext, setExtraContext] = useState("");
	const [agent, setAgent] = useState("");
	const [agentTouched, setAgentTouched] = useState(false);
	const [models, setModels] = useState<Record<string, string>>({});
	const [error, setError] = useState<string | undefined>();
	const [createdId, setCreatedId] = useState<string | undefined>();

	const projectQuery = useQuery({
		queryKey: ["project", projectId],
		enabled: open && Boolean(projectId),
		queryFn: async () => {
			const { data, error: apiError } = await apiClient.GET("/api/v1/projects/{id}", { params: { path: { id: projectId as string } } });
			if (apiError) throw new Error(apiErrorMessage(apiError));
			return data?.project as Project;
		},
	});
	const agentsQuery = useQuery({ ...agentsQueryOptions, enabled: open });
	const agentCatalog = agentsQuery.data;

	// Default the planner agent from project planner → worker config. A saved
	// project default can outlive its local login, so prefer a currently
	// authorized agent rather than submitting an immediately doomed request.
	const defaultAgent = useMemo(() => {
		const cfg = projectQuery.data?.config as { planner?: { agent?: string }; worker?: { agent?: string } } | undefined;
		const configured = cfg?.planner?.agent ?? cfg?.worker?.agent ?? "";
		const installed = agentCatalog?.installed ?? [];
		const configuredStatus = installed.find((item) => item.id === configured)?.authStatus;
		if (configured && configuredStatus !== "unauthorized") return configured;
		return agentCatalog?.authorized?.[0]?.id ?? configured;
	}, [agentCatalog, projectQuery.data]);

	useEffect(() => {
		if (!open) {
			setStep("capture");
			setInput("");
			setAttachments([]);
			setDraft(emptyDraft());
			setExtraContext("");
			setAgent("");
			setAgentTouched(false);
			setModels({});
			setError(undefined);
			setCreatedId(undefined);
		}
	}, [open]);
	useEffect(() => {
		if (open && !agentTouched) setAgent(defaultAgent);
	}, [open, agentTouched, defaultAgent]);

	const addFiles = (files: File[]) => setAttachments((prev) => [...prev, ...files.map(fileToAttachment)]);
	const addUrl = (url: string) => setAttachments((prev) => [...prev, urlToAttachment(url, classifyUrl(url))]);
	const removeAttachment = (id: string) => setAttachments((prev) => prev.filter((a) => a.id !== id));
	const reorderAttachment = (id: string, dir: -1 | 1) =>
		setAttachments((prev) => {
			const i = prev.findIndex((a) => a.id === id);
			const j = i + dir;
			if (i < 0 || j < 0 || j >= prev.length) return prev;
			const next = [...prev];
			[next[i], next[j]] = [next[j], next[i]];
			return next;
		});
	const patchDraft = (patch: Partial<TaskDraft>) => setDraft((d) => ({ ...d, ...patch }));

	// Step 1 → 2: run the planner.
	const extractMutation = useMutation({
		mutationFn: () => extractTask({ input, attachments, agent, projectId }),
		onMutate: () => {
			setError(undefined);
			void captureRendererEvent("thanos.renderer.quick_capture_extract", { project_id: projectId });
		},
		onSuccess: ({ draft: d }) => {
			setDraft({ ...emptyDraft(), ...d });
			setStep("structure");
		},
		onError: (e) => setError(e instanceof Error ? e.message : "AI could not structure the task"),
	});

	// Step 4: create the worker session from the reviewed draft.
	const createMutation = useMutation({
		mutationFn: async () => {
			if (!projectId) throw new Error("No project selected");
			const taskAgent = agent || defaultAgent;
			const taskModel = taskAgent ? models[taskAgent] : undefined;
			const { data, error: apiError } = await apiClient.POST("/api/v1/sessions", {
				body: {
					projectId,
					kind: "worker",
					harness: (agentTouched && agent) || taskModel ? (taskAgent as AgentProvider) : undefined,
					model: taskModel || undefined,
					issueId: draft.title.trim() || "Untitled task",
					prompt: composeSessionPrompt(draft, extraContext) || draft.title,
				},
			});
			if (apiError) throw new Error(apiErrorMessage(apiError, "Unable to create task"));
			if (!data?.session?.id) throw new Error("Task creation returned no session");
			return data.session.id;
		},
		onSuccess: (id) => {
			setCreatedId(id);
			void captureRendererEvent("thanos.renderer.quick_capture_created", { project_id: projectId });
		},
		onError: (e) => setError(e instanceof Error ? e.message : "Unable to create task"),
	});

	const canExtract = (input.trim().length > 0 || attachments.length > 0) && !extractMutation.isPending;
	const stepIndex = STEPS.findIndex((s) => s.id === step);

	return (
		<Dialog.Root open={open} onOpenChange={onOpenChange}>
			<Dialog.Portal>
				<Dialog.Overlay className="fixed inset-0 z-50 bg-black/60 data-[state=open]:animate-overlay-in" />
				<Dialog.Content className="fixed left-1/2 top-1/2 z-50 flex max-h-[calc(100vh-40px)] w-[min(880px,calc(100vw-32px))] -translate-x-1/2 -translate-y-1/2 overflow-hidden rounded-xl border border-border bg-popover text-popover-foreground shadow-2xl data-[state=open]:animate-modal-in">
					{/* Stepper rail */}
					<aside className="hidden w-[200px] shrink-0 flex-col gap-1 border-r border-border bg-surface/40 p-4 sm:flex">
						<Dialog.Title className="mb-3 text-[14px] font-semibold text-foreground">Create New Task</Dialog.Title>
						{STEPS.map((s, i) => {
							const active = s.id === step;
							const done = i < stepIndex || Boolean(createdId);
							return (
								<div key={s.id} className={cn("flex items-center gap-2.5 rounded-lg px-2 py-2", active && "bg-violet-500/10")}>
									<span
										className={cn(
											"grid size-7 shrink-0 place-items-center rounded-full border text-[11px]",
											active ? "border-violet-500 bg-violet-500/20 text-violet-200" : done ? "border-emerald-500/40 bg-emerald-500/15 text-emerald-300" : "border-border text-muted-foreground",
										)}
									>
										{done ? <Check className="size-3.5" /> : <s.icon className="size-3.5" />}
									</span>
									<div className="min-w-0">
										<div className={cn("truncate text-[12px] font-medium", active ? "text-foreground" : "text-muted-foreground")}>{s.label}</div>
										<div className="truncate text-[10px] text-passive">{s.sub}</div>
									</div>
								</div>
							);
						})}
					</aside>

					{/* Body */}
					<div className="flex min-w-0 flex-1 flex-col">
						<div className="flex items-center justify-between border-b border-border px-5 py-3">
							<div className="text-[13px] font-medium text-foreground">
								{createdId ? "Task Created" : STEPS[stepIndex]?.label}
								{step === "structure" && draft.confidence?.overall ? (
									<ConfidenceBadge className="ml-2" value={draft.confidence.overall} label="overall" />
								) : null}
							</div>
							<Dialog.Close asChild>
								<button type="button" aria-label="Close" className="grid size-7 place-items-center rounded-md text-muted-foreground transition hover:bg-surface hover:text-foreground">
									<X className="size-4" />
								</button>
							</Dialog.Close>
						</div>

						<div className="min-h-0 flex-1 overflow-y-auto px-5 py-4">
							{createdId ? (
								<div className="flex flex-col items-center justify-center gap-3 py-10 text-center">
									<span className="grid size-14 place-items-center rounded-full border border-emerald-500/40 bg-emerald-500/10 text-emerald-300">
										<Check className="size-7" />
									</span>
									<div className="text-[15px] font-semibold text-foreground">Task Created</div>
									<p className="max-w-sm text-[12px] text-muted-foreground">Planning will start shortly — {agent || "the agent"} will analyze the task and ask questions if needed.</p>
								</div>
							) : step === "capture" ? (
								<>
									<QuickCaptureEditor
										input={input}
										onInput={setInput}
										attachments={attachments}
										onAddFiles={addFiles}
										onAddUrl={addUrl}
										onRemove={removeAttachment}
										onReorder={reorderAttachment}
									/>
									<NativeModelSelectors
										agents={agentCatalog}
										models={models}
										onChange={(harness, model) => setModels((current) => ({ ...current, [harness]: model }))}
									/>
								</>
							) : step === "structure" ? (
								<AIExtractionCard draft={draft} onChange={patchDraft} agent={agent} />
							) : step === "review" ? (
								<TaskReviewFlow
									draft={draft}
									onChange={patchDraft}
									attachments={attachments}
									onRemove={removeAttachment}
									onReorder={reorderAttachment}
									extraContext={extraContext}
									onExtraContext={setExtraContext}
								/>
							) : (
								<CreateSummary draft={draft} attachments={attachments} agent={agent} />
							)}

							{error ? (
								<div className="mt-4 rounded-md border border-destructive/40 bg-destructive/10 px-3 py-2 text-[12px] text-destructive">
									{error}
								</div>
							) : null}
						</div>

						{/* Footer */}
						<div className="flex items-center justify-between gap-3 border-t border-border px-5 py-3">
							<div className="min-w-0">
								{step === "capture" ? (
									<div className="w-[220px]">
										<RequiredAgentField
											id="quick-capture-agent"
											label="AI Structure agent"
											placeholder="Project default"
											value={agent}
											authorized={agentCatalog?.authorized}
											installed={agentCatalog?.installed}
											supported={agentCatalog?.supported}
											onChange={(v) => {
												setAgent(v);
												setAgentTouched(true);
											}}
										/>
									</div>
								) : null}
							</div>

							{createdId ? (
								<Button
									type="button"
									onClick={() => (onCreated(createdId), onOpenChange(false))}
									className="bg-violet-600 hover:bg-violet-500"
								>
									Open Task
								</Button>
							) : (
								<div className="flex items-center gap-2">
									{step !== "capture" ? (
										<Button type="button" variant="ghost" onClick={() => setStep(STEPS[Math.max(0, stepIndex - 1)].id)}>
											Back
										</Button>
									) : (
										<Dialog.Close asChild>
											<Button type="button" variant="ghost">
												Cancel
											</Button>
										</Dialog.Close>
									)}
									{step === "capture" ? (
										<Button
											type="button"
											disabled={!canExtract}
											onClick={() => extractMutation.mutate()}
											className="bg-violet-600 hover:bg-violet-500"
										>
											{extractMutation.isPending ? (
												<Loader2 className="size-3.5 animate-spin" />
											) : (
												<Sparkles className="size-3.5" />
											)}
											{extractMutation.isPending ? "Structuring…" : "AI Structure"}
											{!extractMutation.isPending ? <ArrowRight className="size-3.5" /> : null}
										</Button>
									) : step === "create" ? (
										<Button
											type="button"
											disabled={createMutation.isPending || !projectId}
											onClick={() => createMutation.mutate()}
											className="bg-violet-600 hover:bg-violet-500"
										>
											{createMutation.isPending ? (
												<Loader2 className="size-3.5 animate-spin" />
											) : (
												<Check className="size-3.5" />
											)}
											{createMutation.isPending ? "Creating…" : "Create Task"}
										</Button>
									) : (
										<Button
											type="button"
											onClick={() => setStep(STEPS[stepIndex + 1].id)}
											className="bg-violet-600 hover:bg-violet-500"
										>
											Next <ArrowRight className="size-3.5" />
										</Button>
									)}
								</div>
							)}
						</div>
					</div>
				</Dialog.Content>
			</Dialog.Portal>
		</Dialog.Root>
	);
}

function CreateSummary({ draft, attachments, agent }: { draft: TaskDraft; attachments: Attachment[]; agent: string }) {
	return (
		<div className="space-y-3">
			<div className="flex items-center gap-2 text-[13px] text-foreground">
				<Check className="size-4 text-emerald-400" /> Everything is ready.
			</div>
			<div className="rounded-lg border border-border bg-surface/50 p-3 text-[12px]">
				<div className="text-[14px] font-semibold text-foreground">{draft.title || "Untitled task"}</div>
				<div className="mt-1 flex flex-wrap items-center gap-1.5 text-muted-foreground">
					<span className="rounded border border-border px-1.5 py-0.5">{draft.priority || "P2"}</span>
					{draft.labels?.slice(0, 6).map((l) => (
						<span key={l} className="rounded-full border border-border px-1.5 py-0.5">
							{l}
						</span>
					))}
				</div>
				{draft.description ? <p className="mt-2 line-clamp-3 text-muted-foreground">{draft.description}</p> : null}
				<div className="mt-2 text-[11px] text-passive">
					{draft.acceptanceCriteria?.length ?? 0} acceptance criteria · {attachments.length} attachment(s) · agent:{" "}
					{agent || "project default"}
				</div>
			</div>
		</div>
	);
}

export function NativeModelSelectors({
	agents,
	models,
	onChange,
}: {
	agents: components["schemas"]["ListAgentsResponse"] | undefined;
	models: Record<string, string>;
	onChange: (harness: string, model: string) => void;
}) {
	const installed = new Set((agents?.installed ?? []).map((agent) => agent.id));
	const authorized = new Set((agents?.authorized ?? []).map((agent) => agent.id));
	const available = (agents?.supported ?? []).filter(
		(agent) => installed.has(agent.id) && (authorized.has(agent.id) || agent.authStatus !== "unauthorized"),
	);
	if (available.length === 0) return null;

	return (
		<div className="mt-4 rounded-lg border border-border bg-surface/30 p-3" aria-label="Native CLI model configuration">
			<div className="text-[12px] font-medium text-foreground">Native CLI models</div>
			<p className="mt-1 text-[11px] text-muted-foreground">
				Optional task-level overrides. Leave a CLI on its default to preserve its configured model.
			</p>
			<div className="mt-3 grid gap-2 sm:grid-cols-2">
				{available.map((agent) => {
					const options = agent.models ?? [];
					const unsupported = options.length === 0;
					return (
						<label key={agent.id} className="flex flex-col gap-1 text-[11px] text-muted-foreground">
							<span>{agent.label}</span>
							<select
								aria-label={`${agent.label} model`}
								className="h-8 rounded-md border border-border bg-transparent px-2 text-[12px] text-foreground disabled:cursor-not-allowed disabled:opacity-60"
								disabled={unsupported}
								value={models[agent.id] ?? ""}
								onChange={(event) => onChange(agent.id, event.target.value)}
							>
								<option value="">{unsupported ? "Model selection unavailable" : "CLI default"}</option>
								{options.map((model) => (
									<option key={model} value={model}>
										{model}
									</option>
								))}
							</select>
						</label>
					);
				})}
			</div>
		</div>
	);
}
