import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { QuickCaptureProviderSelectors } from "./CreateTaskWizard";

describe("QuickCaptureProviderSelectors", () => {
	it("offers Codex and Claude and waits for an agent before enabling models", () => {
		render(<QuickCaptureProviderSelectors agent="" model="" models={[]} modelsLoading={false} modelsError={false} onAgentChange={vi.fn()} onModelChange={vi.fn()} />);
		expect(screen.getByRole("option", { name: "Codex" })).toBeInTheDocument();
		expect(screen.getByRole("option", { name: "Claude" })).toBeInTheDocument();
		expect(screen.getByLabelText("Models")).toBeDisabled();
		expect(screen.getByRole("option", { name: "Select an agent first" })).toBeInTheDocument();
	});

	it("reports a model selected for the active provider", async () => {
		const onModelChange = vi.fn();
		render(<QuickCaptureProviderSelectors agent="claude-code" model="" models={["claude-opus-4-5", "claude-sonnet-4-5"]} modelsLoading={false} modelsError={false} onAgentChange={vi.fn()} onModelChange={onModelChange} />);
		await userEvent.selectOptions(screen.getByLabelText("Models"), "claude-sonnet-4-5");
		expect(onModelChange).toHaveBeenCalledWith("claude-sonnet-4-5");
	});

	it("shows a model retrieval error state", () => {
		render(<QuickCaptureProviderSelectors agent="codex" model="" models={[]} modelsLoading={false} modelsError onAgentChange={vi.fn()} onModelChange={vi.fn()} />);
		expect(screen.getByLabelText("Models")).toBeDisabled();
		expect(screen.getByRole("option", { name: "Models unavailable" })).toBeInTheDocument();
	});
});
