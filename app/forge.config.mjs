import { VitePlugin } from "@electron-forge/plugin-vite";
import MakerNSIS from "./makers/maker-nsis.mjs";
import MakerAppImage from "./makers/maker-appimage.mjs";
import { writeFileSync } from "node:fs";

// This config is ESM (.mjs) on purpose: Electron Forge loads .ts/.cts/.mts configs
// through jiti, which fails to transform Vite 8's `import.meta.require` and throws
// "Cannot use 'import.meta' outside a module". A native ESM .mjs config is loaded
// with real `import()`, where Vite's import.meta is valid. Keep it .mjs.

// Default GitHub release target (production).
const DEFAULT_RELEASE_REPO = "tinhtran24/maestro";

// parseReleaseRepo turns an "owner/repo" string (from TO_RELEASE_REPO) into the
// publisher-github { owner, name } shape, falling back to the production default
// when unset or malformed.
function parseReleaseRepo(value) {
	const [owner, name] = (value || DEFAULT_RELEASE_REPO).split("/");
	if (!owner || !name) {
		const [defOwner, defName] = DEFAULT_RELEASE_REPO.split("/");
		return { owner: defOwner, name: defName };
	}
	return { owner, name };
}

/** @type {import("@electron-forge/shared-types").ForgeConfig} */
const config = {
	packagerConfig: {
		asar: true,
		appBundleId: "dev.maestro.desktop",
		name: "Maestro",
		executableName: "maestro",
		appCategoryType: "public.app-category.developer-tools",
		// App icon. electron-packager appends the per-platform extension
		// (.icns on macOS, .ico on Windows); Linux menu icons come from the
		// deb/rpm makers below, and the runtime window icon from src/main.ts.
		icon: "assets/icon",
		extraResource: ["daemon", "assets/icon.png", "app-update.yml"],
		// Notarization. Two paths:
		//  - CI: an App Store Connect API key. APPLE_API_KEY is a PATH to the .p8,
		//    plus the key id + issuer uuid.
		//  - Local: TO_NOTARY_PROFILE, a notarytool keychain profile created with
		//    `notarytool store-credentials`.
		osxSign: process.env.APPLE_SIGNING_IDENTITY
			? { identity: process.env.APPLE_SIGNING_IDENTITY }
			: process.env.CSC_LINK
				? {}
				: undefined,
		osxNotarize: process.env.TO_NOTARY_PROFILE
			? { keychainProfile: process.env.TO_NOTARY_PROFILE }
			: process.env.APPLE_API_KEY
				? {
						appleApiKey: process.env.APPLE_API_KEY,
						appleApiKeyId: process.env.APPLE_API_KEY_ID,
						appleApiIssuer: process.env.APPLE_API_ISSUER,
					}
				: undefined,
	},
	hooks: {
		// electron-forge does not generate app-update.yml (electron-builder does);
		// electron-updater reads it from the app's Resources dir at runtime to know
		// which GitHub repo to pull from, else it throws ENOENT during download.
		// Generate it in prePackage (BEFORE osxSign) and ship it via extraResource
		// above, so it is copied into the bundle and SIGNED as part of the seal.
		prePackage: async () => {
			const { owner, name } = parseReleaseRepo(process.env.TO_RELEASE_REPO);
			const yml = [
				"provider: github",
				`owner: ${owner}`,
				`repo: ${name}`,
				"updaterCacheDirName: maestro-updater",
				"",
			].join("\n");
			writeFileSync("app-update.yml", yml);
		},
	},
	rebuildConfig: {},
	makers: [
		// Windows installer: NSIS via electron-builder (see makers/maker-nsis.mjs).
		new MakerNSIS(
			{
				appId: "dev.maestro.desktop",
				productName: "Maestro",
				icon: "assets/icon.ico",
			},
			["win32"],
		),
		{ name: "@electron-forge/maker-zip", platforms: ["darwin"], config: {} },
		// Linux fetch-and-run artifact for `to start`: a single self-contained
		// AppImage the Go bootstrapper downloads and runs directly (see
		// makers/maker-appimage.mjs). The deb/rpm makers below stay for users who
		// prefer a system package.
		new MakerAppImage(
			{
				appId: "dev.maestro.desktop",
				productName: "Maestro",
				icon: "assets/icon.png",
			},
			["linux"],
		),
		{
			name: "@electron-forge/maker-deb",
			config: {
				options: {
					// Must match packagerConfig.executableName, or the deb maker
					// looks for the package name and fails with "could not find
					// the Electron app binary". (Both are "maestro".)
					bin: "maestro",
					icon: "assets/icon.png",
					maintainer: "Maestro",
					homepage: "https://github.com/tinhtran24/maestro",
				},
			},
		},
		{
			name: "@electron-forge/maker-rpm",
			config: {
				options: {
					icon: "assets/icon.png",
					// rpmbuild rejects a spec with an empty License field.
					license: "MIT",
					homepage: "https://github.com/tinhtran24/maestro",
				},
			},
		},
	],
	publishers: [
		{
			name: "@electron-forge/publisher-github",
			// Release target is build-time overridable so a fork run publishes to the
			// fork without a source edit. TO_RELEASE_REPO is "owner/repo"; it defaults
			// to the production target.
			config: {
				repository: parseReleaseRepo(process.env.TO_RELEASE_REPO),
				prerelease: process.env.TO_RELEASE_PRERELEASE === "true",
				draft: false,
			},
		},
	],
	plugins: [
		new VitePlugin({
			build: [
				{ entry: "src/main.ts", config: "vite.main.config.ts", target: "main" },
				{ entry: "src/preload.ts", config: "vite.preload.config.ts", target: "preload" },
				{ entry: "src/annotate-preload.ts", config: "vite.preload.config.ts", target: "preload" },
			],
			renderer: [{ name: "main_window", config: "vite.renderer.config.ts" }],
		}),
	],
};

export default config;
