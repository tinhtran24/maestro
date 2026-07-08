import { cn } from "../../lib/utils";

// ConfidenceBadge renders a 0–100 planner confidence score with a color that
// reflects certainty (green high, amber medium, red low).
export function ConfidenceBadge({ value, label, className }: { value: number; label?: string; className?: string }) {
	const tone =
		value >= 80
			? "border-emerald-500/30 bg-emerald-500/10 text-emerald-300"
			: value >= 50
				? "border-amber-500/30 bg-amber-500/10 text-amber-300"
				: "border-rose-500/30 bg-rose-500/10 text-rose-300";
	return (
		<span
			className={cn(
				"inline-flex items-center gap-1 rounded-full border px-1.5 py-0.5 text-[10px] font-medium tabular-nums",
				tone,
				className,
			)}
			title={label ? `${label} confidence` : "Confidence"}
		>
			{label ? <span className="opacity-70">{label}</span> : null}
			{value}%
		</span>
	);
}
