import path from "node:path";

export interface TmuxResolution {
	path: string | null;
	checked: string[];
}

export interface ResolveTmuxOptions {
	env: NodeJS.ProcessEnv;
	persistedPath?: string;
	isExecutable: (candidate: string) => Promise<boolean>;
}

const KNOWN_TMUX_PATHS = ["/opt/homebrew/bin/tmux", "/usr/local/bin/tmux", "/usr/bin/tmux"];

/** Resolve one canonical tmux executable without invoking a shell. */
export async function resolveTmux(options: ResolveTmuxOptions): Promise<TmuxResolution> {
	if (process.platform === "win32") return { path: null, checked: [] };

	const candidates: string[] = [];
	const add = (candidate: string | undefined) => {
		const value = candidate?.trim();
		if (value && !candidates.includes(value)) candidates.push(value);
	};

	add(options.env.THANOS_TMUX_BIN);
	add(options.persistedPath);
	for (const dir of (options.env.PATH ?? "").split(path.delimiter)) {
		if (dir) add(path.join(dir, "tmux"));
	}
	for (const candidate of KNOWN_TMUX_PATHS) add(candidate);

	for (const candidate of candidates) {
		if (await options.isExecutable(candidate)) {
			return { path: path.resolve(candidate), checked: candidates };
		}
	}
	return { path: null, checked: candidates };
}

export function tmuxInstallGuidance(platform: NodeJS.Platform): string {
	if (platform === "darwin") return "Install tmux with `brew install tmux`, then restart Thanos.";
	return "Install tmux with your system package manager (for example `sudo apt install tmux`), then restart Thanos.";
}
