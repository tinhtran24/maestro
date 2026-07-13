import { describe, expect, it } from "vitest";
import { resolveTmux, tmuxInstallGuidance } from "./tmux-dependency";

describe("resolveTmux", () => {
	it("prefers an explicit binary and returns its canonical path", async () => {
		const seen: string[] = [];
		const result = await resolveTmux({
			env: { THANOS_TMUX_BIN: "/custom/tmux", PATH: "/bin" },
			persistedPath: "/persisted/tmux",
			isExecutable: async (candidate) => {
				seen.push(candidate);
				return candidate === "/custom/tmux";
			},
		});
		expect(result.path).toBe("/custom/tmux");
		expect(seen).toEqual(["/custom/tmux"]);
	});

	it("falls back from a stale persisted path to PATH", async () => {
		const result = await resolveTmux({
			env: { PATH: "/tools:/bin" },
			persistedPath: "/old/tmux",
			isExecutable: async (candidate) => candidate === "/tools/tmux",
		});
		expect(result.path).toBe("/tools/tmux");
	});

	it("returns actionable platform guidance", () => {
		expect(tmuxInstallGuidance("darwin")).toContain("brew install tmux");
		expect(tmuxInstallGuidance("linux")).toContain("apt install tmux");
	});
});
