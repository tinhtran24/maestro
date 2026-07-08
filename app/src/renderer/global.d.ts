import type { AoBridge } from "../preload";

declare global {
	interface Window {
		ao?: AoBridge;
	}

	interface ImportMetaEnv {
		readonly VITE_THANOS_POSTHOG_KEY?: string;
		readonly VITE_THANOS_POSTHOG_HOST?: string;
	}
}

export {};
