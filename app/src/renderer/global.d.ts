import type { AoBridge } from "../preload";

declare global {
	interface Window {
		to?: AoBridge;
	}

	interface ImportMetaEnv {
		readonly VITE_MAESTRO_POSTHOG_KEY?: string;
		readonly VITE_MAESTRO_POSTHOG_HOST?: string;
	}
}

export {};
