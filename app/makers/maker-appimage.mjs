import path from "node:path";
import { MakerBase } from "@electron-forge/maker-base";

// Electron Forge has no first-party AppImage maker, so we bridge to
// electron-builder's `buildForge`, exactly as makers/maker-nsis.mjs does for the
// Windows NSIS installer. AppImage is the Linux fetch-and-run artifact for the
// `to start` bootstrapper: a single self-contained executable the Go agent can
// download from releases/latest/download and run directly, with no system
// package manager. The deb/rpm makers stay for users who want a system package.
//
// `buildForge` speaks Forge's legacy v5 function API, which Forge 7's class-based
// maker loader cannot resolve, so this thin MakerBase subclass adapts it.
//
// Config shape (all optional): { appId, productName, icon, appImage }.

export default class MakerAppImage extends MakerBase {
	name = "appimage";
	defaultPlatforms = ["linux"];

	isSupportedOnCurrentPlatform() {
		return true;
	}

	async make({ dir, targetArch, appName }) {
		const { buildForge } = await import("app-builder-lib");
		const cfg = this.config ?? {};
		// Mirror buildForge's own output layout (<dir>/../make) so artifacts land
		// where Forge's publisher expects them.
		const output = path.join(path.dirname(path.resolve(dir)), "make");
		return buildForge(
			{ dir },
			{
				linux: [`appImage:${targetArch}`],
				config: {
					appId: cfg.appId,
					productName: cfg.productName ?? appName,
					directories: { output },
					// Forge owns publishing (the workflow uploads via `gh release`).
					// `null` stops electron-builder from inferring a GitHub publish
					// target from package.json `repository` and trying to upload,
					// which fails in CI with no GH_TOKEN set.
					publish: null,
					linux: {
						...(cfg.icon ? { icon: cfg.icon } : {}),
					},
					appImage: {
						...cfg.appImage,
					},
				},
			},
		);
	}
}

export { MakerAppImage };
