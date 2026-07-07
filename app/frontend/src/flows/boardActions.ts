import type { TaskStatus } from "../app/types";

export type TaskMoveAction = "start-turn" | "update-status";

export function taskMoveActionForStatus(status: TaskStatus): TaskMoveAction {
  return status === "in_progress" ? "start-turn" : "update-status";
}
