import { PanelRightOpen } from "lucide-react";
import { Panel, PanelGroup, PanelResizeHandle } from "react-resizable-panels";

export function SplitPane({
  left,
  right,
  bottom,
  rightCollapsed = false,
  onExpandRight,
}: {
  left: React.ReactNode;
  right: React.ReactNode;
  bottom: React.ReactNode;
  rightCollapsed?: boolean;
  onExpandRight?: () => void;
}) {
  // A null/absent right pane means "no right rail at all" — hide the panel, its
  // resize handle, and the collapsed expand button.
  const showRight = right != null;
  return (
    <>
      <div className={`grid h-full min-h-0 lg:hidden ${showRight ? "grid-rows-[minmax(28rem,1fr)_minmax(14rem,28rem)_minmax(12rem,20rem)]" : "grid-rows-[minmax(28rem,1fr)_minmax(12rem,20rem)]"}`}>
        <div className="min-h-0">{left}</div>
        {showRight && <div className="min-h-0">{right}</div>}
        <div className="min-h-0">{bottom}</div>
      </div>
      <div className="hidden min-h-0 lg:flex">
        <PanelGroup direction="horizontal" className="flex min-h-0 flex-1">
          <Panel defaultSize={72} minSize={45}>
            <PanelGroup direction="vertical">
              <Panel defaultSize={72} minSize={40}>
                {left}
              </Panel>
              <PanelResizeHandle className="h-1 bg-slate-800 hover:bg-purple-primary" />
              <Panel defaultSize={28} minSize={16}>
                {bottom}
              </Panel>
            </PanelGroup>
          </Panel>
          {showRight && !rightCollapsed && (
            <>
              <PanelResizeHandle className="w-1 bg-slate-800 hover:bg-purple-primary" />
              <Panel defaultSize={28} minSize={22} maxSize={40}>
                {right}
              </Panel>
            </>
          )}
        </PanelGroup>
        {showRight && rightCollapsed && (
          <button
            onClick={onExpandRight}
            className="flex w-10 shrink-0 flex-col items-center gap-2 border-l border-slate-800 bg-slate-900/70 py-3 text-text-muted hover:text-text-main"
            aria-label="Expand task details"
          >
            <PanelRightOpen size={16} />
            <span className="mt-1 text-[10px] uppercase tracking-wide [writing-mode:vertical-rl]">Task Details</span>
          </button>
        )}
      </div>
    </>
  );
}
