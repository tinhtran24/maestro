import { Check, Plus, X } from "lucide-react";
import { type ReactNode, useState } from "react";
import { cn } from "../../lib/utils";

export function Field({ label, hint, children }: { label: string; hint?: ReactNode; children: ReactNode }) {
	return (
		<div className="space-y-1.5">
			<div className="flex items-center gap-2">
				<span className="text-[11px] font-medium text-muted-foreground">{label}</span>
				{hint}
			</div>
			{children}
		</div>
	);
}

export function TextField({ value, onChange, placeholder }: { value: string; onChange: (v: string) => void; placeholder?: string }) {
	return (
		<input
			value={value}
			onChange={(e) => onChange(e.target.value)}
			placeholder={placeholder}
			className="h-9 w-full rounded-md border border-border bg-transparent px-2.5 text-[13px] text-foreground outline-none transition focus-visible:border-violet-500/70 focus-visible:ring-2 focus-visible:ring-violet-500/20"
		/>
	);
}

export function TextArea({ value, onChange, placeholder, rows = 3 }: { value: string; onChange: (v: string) => void; placeholder?: string; rows?: number }) {
	return (
		<textarea
			value={value}
			onChange={(e) => onChange(e.target.value)}
			placeholder={placeholder}
			rows={rows}
			className="w-full resize-y rounded-md border border-border bg-transparent px-2.5 py-2 text-[13px] leading-relaxed text-foreground outline-none transition focus-visible:border-violet-500/70 focus-visible:ring-2 focus-visible:ring-violet-500/20"
		/>
	);
}

export function PrioritySelect({ value, onChange }: { value: string; onChange: (v: string) => void }) {
	const opts = [
		{ v: "P0", label: "P0 – Critical" },
		{ v: "P1", label: "P1 – High" },
		{ v: "P2", label: "P2 – Medium" },
		{ v: "P3", label: "P3 – Low" },
	];
	return (
		<select
			value={value || "P2"}
			onChange={(e) => onChange(e.target.value)}
			className="h-9 w-full rounded-md border border-border bg-surface px-2.5 text-[13px] text-foreground outline-none focus-visible:border-violet-500/70"
		>
			{opts.map((o) => (
				<option key={o.v} value={o.v}>
					{o.label}
				</option>
			))}
		</select>
	);
}

// LabelChips edits a string[] as removable chips with an add input.
export function LabelChips({ values, onChange }: { values: string[]; onChange: (v: string[]) => void }) {
	const [draft, setDraft] = useState("");
	const add = () => {
		const v = draft.trim();
		if (v && !values.includes(v)) onChange([...values, v]);
		setDraft("");
	};
	return (
		<div className="flex flex-wrap items-center gap-1.5">
			{values.map((label) => (
				<span key={label} className="inline-flex items-center gap-1 rounded-full border border-border bg-surface px-2 py-0.5 text-[11px] text-foreground">
					{label}
					<button type="button" aria-label={`Remove ${label}`} onClick={() => onChange(values.filter((l) => l !== label))} className="text-muted-foreground hover:text-rose-300">
						<X className="size-3" />
					</button>
				</span>
			))}
			<span className="inline-flex items-center gap-1 rounded-full border border-dashed border-border px-1.5 py-0.5">
				<input
					value={draft}
					onChange={(e) => setDraft(e.target.value)}
					onKeyDown={(e) => e.key === "Enter" && (e.preventDefault(), add())}
					placeholder="add"
					className="w-14 bg-transparent text-[11px] text-foreground outline-none placeholder:text-passive"
				/>
				<button type="button" aria-label="Add label" onClick={add} className="text-muted-foreground hover:text-violet-300">
					<Plus className="size-3" />
				</button>
			</span>
		</div>
	);
}

// StringList edits a string[] as add/remove rows, optionally with a leading
// check (used for acceptance criteria).
export function StringList({
	values,
	onChange,
	placeholder,
	check,
}: {
	values: string[];
	onChange: (v: string[]) => void;
	placeholder?: string;
	check?: boolean;
}) {
	const [draft, setDraft] = useState("");
	const add = () => {
		const v = draft.trim();
		if (v) onChange([...values, v]);
		setDraft("");
	};
	return (
		<div className="space-y-1.5">
			{values.map((item, i) => (
				<div key={`${i}-${item}`} className="group flex items-start gap-2">
					{check ? <Check className="mt-1 size-3.5 shrink-0 text-emerald-400" aria-hidden="true" /> : <span className="mt-1 text-muted-foreground">•</span>}
					<input
						value={item}
						onChange={(e) => onChange(values.map((v, j) => (j === i ? e.target.value : v)))}
						className="flex-1 rounded-md border border-transparent bg-transparent px-1.5 py-1 text-[12px] text-foreground outline-none transition hover:border-border focus-visible:border-violet-500/70"
					/>
					<button
						type="button"
						aria-label="Remove"
						onClick={() => onChange(values.filter((_, j) => j !== i))}
						className="mt-1 shrink-0 text-muted-foreground opacity-0 transition group-hover:opacity-100 hover:text-rose-300"
					>
						<X className="size-3.5" />
					</button>
				</div>
			))}
			<div className="flex items-center gap-2">
				<Plus className="size-3.5 text-muted-foreground" aria-hidden="true" />
				<input
					value={draft}
					onChange={(e) => setDraft(e.target.value)}
					onKeyDown={(e) => e.key === "Enter" && (e.preventDefault(), add())}
					placeholder={placeholder ?? "Add item"}
					className={cn("flex-1 rounded-md border border-dashed border-border bg-transparent px-1.5 py-1 text-[12px] text-foreground outline-none placeholder:text-passive focus-visible:border-violet-500/70")}
				/>
			</div>
		</div>
	);
}
