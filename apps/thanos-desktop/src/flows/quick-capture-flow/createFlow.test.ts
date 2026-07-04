import { beforeEach, describe, expect, it } from "vitest";
import { extractTask, toCreateInput } from "../../features/tasks/quickCapture";
import { useWorkbenchStore } from "../../state/workbenchStore";

// Exercises the real create-task flow at the store level: capture → extract →
// createTask, matching what QuickCaptureFlow.create() does.
describe("create task flow (real store)", () => {
  beforeEach(() => {
    useWorkbenchStore.setState({ tasks: [], features: [], selectedTaskId: "", taskDialog: null, taskHistory: {} });
  });

  it("opens the create dialog, then creates and selects the task", () => {
    const store = useWorkbenchStore.getState();
    store.openCreateTask();
    expect(useWorkbenchStore.getState().taskDialog).toEqual({ mode: "create" });

    const draft = extractTask("Build a shopping cart with guest checkout");
    store.createTask(toCreateInput(draft));

    const state = useWorkbenchStore.getState();
    expect(state.tasks).toHaveLength(1);
    const [task] = state.tasks;
    expect(task.title).toBe("Implement Shopping Cart");
    expect(task.priority).toBe("P1");
    expect(task.tags).toContain("cart");
    // Newly created tasks land in Backlog and are selected/opened.
    expect(task.status).toBe("backlog");
    expect(state.selectedTaskId).toBe(task.id);
    // A creation event is recorded on the timeline.
    expect(useWorkbenchStore.getState().taskHistory[task.id]?.length).toBeGreaterThan(0);
  });

  it("does not auto-start planning — the task stays in backlog", () => {
    useWorkbenchStore.getState().createTask(toCreateInput(extractTask("Fix the login bug")));
    expect(useWorkbenchStore.getState().tasks[0].status).toBe("backlog");
  });
});
