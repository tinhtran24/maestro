import { describe, expect, it } from "vitest";
import { resolveDaemonLaunch } from "./daemon-launch";

describe("resolveDaemonLaunch", () => {
	it("uses MAESTRO_DAEMON_COMMAND when configured", () => {
		expect(
			resolveDaemonLaunch({ MAESTRO_DAEMON_COMMAND: "/tmp/to daemon" }, false, "/resources", "/app", "darwin"),
		).toEqual({
			command: "/tmp/to daemon",
			args: [],
			cwd: "/app",
			shell: true,
			source: "configured",
		});
	});

	it("runs the backend daemon from source in dev without an explicit command", () => {
		expect(resolveDaemonLaunch({}, false, "/resources", "/repo/frontend", "darwin")).toEqual({
			command: "go",
			args: ["run", "./cmd/maestro", "daemon"],
			cwd: "/repo/frontend/../backend",
			shell: false,
			source: "dev",
		});
	});

	it("uses the bundled daemon binary for packaged macOS/Linux builds", () => {
		expect(resolveDaemonLaunch({}, true, "/Applications/Maestro.app/Contents/Resources", "/app", "darwin")).toEqual({
			command: "/Applications/Maestro.app/Contents/Resources/daemon/maestro",
			args: ["daemon"],
			cwd: "/Applications/Maestro.app/Contents/Resources",
			shell: false,
			source: "bundled",
		});
	});

	it("uses the bundled daemon exe for packaged Windows builds", () => {
		expect(
			resolveDaemonLaunch(
				{},
				true,
				"C:\\Program Files\\Maestro\\resources",
				"C:\\Program Files\\Maestro\\resources\\app.asar",
				"win32",
			),
		).toEqual({
			command: "C:\\Program Files\\Maestro\\resources/daemon/maestro.exe",
			args: ["daemon"],
			cwd: "C:\\Program Files\\Maestro\\resources",
			shell: false,
			source: "bundled",
		});
	});
});
