import { useEffect, useRef } from "react";
import { Terminal as XTerm } from "@xterm/xterm";
import { FitAddon } from "@xterm/addon-fit";
import "@xterm/xterm/css/xterm.css";
import type { AgentSession, ExecutionPlan, Task } from "../domain/models";
import { NativeBackend } from "../services/nativeBackend";
import { startAgentStep } from "../features/terminal/agentRuntime";
import type { TerminalStep } from "../features/terminal/mockRuntime";

const backend = new NativeBackend();

const STEP_FOR_AGENT: Record<AgentSession["agentType"], TerminalStep> = {
  planner: "planning",
  coder: "coding",
  reviewer: "review",
  tester: "testing",
  utility: "planning",
};

export function useAgentSessionFlow() {
  // Streaming listeners and session creation live in agentRuntime (shared by the
  // terminal toolbar and the planner/coder cards), so the header button uses the
  // same real path.
  return {
    async start(task: Task, agentType: AgentSession["agentType"] = "coder") {
      await startAgentStep(task, STEP_FOR_AGENT[agentType]);
    },
    async stop() {
      await backend.stopAgent();
    },
    async savePlan(plan: ExecutionPlan) {
      return await backend.saveExecutionPlan(plan);
    },
    async approvePlan(taskId: string) {
      return await backend.approveExecutionPlan(taskId);
    },
    async collectDiff(task: Task) {
      return await backend.collectGitDiff(task);
    },
    async runTests(task: Task) {
      return await backend.runTaskTests(task);
    },
    async saveReview(review: import("../domain/models").Review) {
      return await backend.saveReview(review);
    },
    async approveReview(taskId: string) {
      return await backend.approveReview(taskId);
    },
    async searchMemory(query: string) {
      return await backend.searchMemory(query);
    },
    async writeMemory(node: import("../domain/models").MemoryNode) {
      return await backend.writeMemoryNode(node);
    },
  };
}

// Renders a live terminal. `output` is an append-only list of raw chunks (real
// PTY bytes or simulated lines, each already carrying its own newlines); we write
// only the newly-appended entries so ANSI/cursor sequences stream correctly.
// `sessionId` resets the view when the active tab changes.
export function XtermPanel({ output, sessionId }: { output: string[]; sessionId?: string }) {
  const hostRef = useRef<HTMLDivElement | null>(null);
  const terminalRef = useRef<XTerm | null>(null);
  const fitRef = useRef<FitAddon | null>(null);
  const writtenRef = useRef(0);
  const sessionRef = useRef<string | undefined>(undefined);

  useEffect(() => {
    if (!hostRef.current || terminalRef.current) return;
    const term = new XTerm({
      convertEol: true,
      cursorBlink: true,
      fontFamily: "Menlo, Monaco, Consolas, monospace",
      fontSize: 12,
      theme: { background: "#050505", foreground: "#E5E7EB", cursor: "#E5E7EB", selectionBackground: "#264F78" },
    });
    const fit = new FitAddon();
    term.loadAddon(fit);
    term.open(hostRef.current);
    term.onData((data) => {
      void backend.writeTerminal(data);
    });
    fit.fit();
    void backend.resizeTerminal(term.rows, term.cols);
    terminalRef.current = term;
    fitRef.current = fit;
    const onResize = () => {
      fit.fit();
      void backend.resizeTerminal(term.rows, term.cols);
    };
    window.addEventListener("resize", onResize);
    return () => {
      window.removeEventListener("resize", onResize);
      term.dispose();
    };
  }, []);

  useEffect(() => {
    const terminal = terminalRef.current;
    if (!terminal) return;

    // Switched tabs (or restarted → output truncated): clear and replay.
    if (sessionRef.current !== sessionId || output.length < writtenRef.current) {
      sessionRef.current = sessionId;
      writtenRef.current = 0;
      terminal.clear();
    }

    if (output.length === 0) {
      if (writtenRef.current === 0) {
        terminal.write("\x1b[38;5;244mLast login: local Thanos terminal\x1b[0m\r\n\x1b[32mthanos\x1b[0m:\x1b[34m~\x1b[0m$ ");
        writtenRef.current = -1; // sentinel: idle prompt written
      }
      return;
    }
    if (writtenRef.current < 0) {
      terminal.clear();
      writtenRef.current = 0;
    }
    for (let i = writtenRef.current; i < output.length; i += 1) {
      terminal.write(output[i]);
    }
    writtenRef.current = output.length;
  }, [output, sessionId]);

  return <div ref={hostRef} className="h-full min-h-0 overflow-hidden rounded-lg bg-black" />;
}
