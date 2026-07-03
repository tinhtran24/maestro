import { useEffect, useRef, useState } from "react";

export function Dropdown({
  trigger,
  children,
  align = "left",
  width = "w-72",
}: {
  trigger: React.ReactNode;
  children: (close: () => void) => React.ReactNode;
  align?: "left" | "right";
  width?: string;
}) {
  const [open, setOpen] = useState(false);
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!open) return;
    function onDoc(event: MouseEvent) {
      if (ref.current && !ref.current.contains(event.target as Node)) setOpen(false);
    }
    function onKey(event: KeyboardEvent) {
      if (event.key === "Escape") setOpen(false);
    }
    document.addEventListener("mousedown", onDoc);
    document.addEventListener("keydown", onKey);
    return () => {
      document.removeEventListener("mousedown", onDoc);
      document.removeEventListener("keydown", onKey);
    };
  }, [open]);

  return (
    <div className="relative min-w-0" ref={ref}>
      <button type="button" onClick={() => setOpen((value) => !value)} className="min-w-0 max-w-full">
        {trigger}
      </button>
      {open && (
        <div
          className={`absolute z-50 mt-1 ${align === "right" ? "right-0" : "left-0"} ${width} max-w-[calc(100vw-1.5rem)] rounded-lg border border-slate-800 bg-slate-900 p-1 shadow-xl shadow-black/40`}
        >
          {children(() => setOpen(false))}
        </div>
      )}
    </div>
  );
}
