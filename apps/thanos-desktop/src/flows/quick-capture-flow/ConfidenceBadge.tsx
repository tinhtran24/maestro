// Per-field confidence indicator for mock AI extraction.
export function ConfidenceBadge({ value }: { value: number }) {
  const tone = value >= 75 ? "text-green-success border-green-success/30 bg-green-success/10" : value >= 50 ? "text-blue-info border-blue-info/30 bg-blue-info/10" : "text-yellow-warning border-yellow-warning/30 bg-yellow-warning/10";
  return <span className={`shrink-0 rounded-md border px-1.5 py-0.5 text-[10px] font-medium ${tone}`}>{value}% sure</span>;
}
