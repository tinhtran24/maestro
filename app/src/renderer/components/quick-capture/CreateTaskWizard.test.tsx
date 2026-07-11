import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import type { components } from "../../../api/schema";
import { NativeModelSelectors } from "./CreateTaskWizard";

const agents: components["schemas"]["ListAgentsResponse"] = {
	supported: [
		{ id: "claude-code", label: "Claude Code", models: ["claude-opus-4-5", "claude-sonnet-4-5"] },
		{ id: "opencode", label: "OpenCode" },
	],
	installed: [
		{ id: "claude-code", label: "Claude Code", authStatus: "authorized" },
		{ id: "opencode", label: "OpenCode", authStatus: "authorized" },
	],
	authorized: [
		{ id: "claude-code", label: "Claude Code" },
		{ id: "opencode", label: "OpenCode" },
	],
};

describe("NativeModelSelectors", () => {
	it("renders selectable native models and a disabled unavailable state", () => {
		render(<NativeModelSelectors agents={agents} models={{}} onChange={vi.fn()} />);
		expect(screen.getByLabelText("Claude Code model")).toBeEnabled();
		expect(screen.getByRole("option", { name: "claude-opus-4-5" })).toBeInTheDocument();
		expect(screen.getByLabelText("OpenCode model")).toBeDisabled();
		expect(screen.getByRole("option", { name: "Model selection unavailable" })).toBeInTheDocument();
	});

	it("reports an independently changed CLI model", async () => {
		const onChange = vi.fn();
		render(<NativeModelSelectors agents={agents} models={{}} onChange={onChange} />);
		await userEvent.selectOptions(screen.getByLabelText("Claude Code model"), "claude-sonnet-4-5");
		expect(onChange).toHaveBeenCalledWith("claude-code", "claude-sonnet-4-5");
	});
});
