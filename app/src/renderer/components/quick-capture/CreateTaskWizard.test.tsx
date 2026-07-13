import { describe, expect, it } from "vitest";
import { taskCreationBody } from "./CreateTaskWizard";

describe("taskCreationBody", () => {
	it("omits native CLI model overrides while preserving the selected execution agent", () => {
		const body = taskCreationBody({
			projectId: "proj-1",
			agent: "codex",
			agentTouched: true,
			issueId: "Implement task creation",
			prompt: "Implement task creation",
		});

		expect(body).toMatchObject({ projectId: "proj-1", kind: "worker", harness: "codex" });
		expect(body).not.toHaveProperty("model");
	});
});
