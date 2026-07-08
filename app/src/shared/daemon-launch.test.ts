import { describe, expect, it } from "vitest";
import { resolveDaemonLaunch } from "./daemon-launch";

describe("resolveDaemonLaunch", () => {
	it("uses THANOS_DAEMON_COMMAND when configured", () => {
		expect(resolveDaemonLaunch({ THANOS_DAEMON_COMMAND: "/tmp/to daemon" }, false, "/resources", "/app", "darwin")).toEqual(
			{
				command: "/tmp/to daemon",
				args: [],
				cwd: "/app",
				shell: true,
				source: "configured",
			},
		);
	});

	it("runs the backend daemon from source in dev without an explicit command", () => {
		expect(resolveDaemonLaunch({}, false, "/resources", "/repo/frontend", "darwin")).toEqual({
			command: "go",
			args: ["run", "./cmd/to", "daemon"],
			cwd: "/repo/frontend/../backend",
			shell: false,
			source: "dev",
		});
	});

	it("uses the bundled daemon binary for packaged macOS/Linux builds", () => {
		expect(
			resolveDaemonLaunch({}, true, "/Applications/Thanos.app/Contents/Resources", "/app", "darwin"),
		).toEqual({
			command: "/Applications/Thanos.app/Contents/Resources/daemon/ao",
			args: ["daemon"],
			cwd: "/Applications/Thanos.app/Contents/Resources",
			shell: false,
			source: "bundled",
		});
	});

	it("uses the bundled daemon exe for packaged Windows builds", () => {
		expect(
			resolveDaemonLaunch(
				{},
				true,
				"C:\\Program Files\\Thanos\\resources",
				"C:\\Program Files\\Thanos\\resources\\app.asar",
				"win32",
			),
		).toEqual({
			command: "C:\\Program Files\\Thanos\\resources/daemon/ao.exe",
			args: ["daemon"],
			cwd: "C:\\Program Files\\Thanos\\resources",
			shell: false,
			source: "bundled",
		});
	});
});
