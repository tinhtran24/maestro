import path from "node:path";
import { MakerBase } from "@electron-forge/maker-base";

// Electron Forge has no first-party NSIS maker, so we bridge to electron-builder's
// `buildForge`, the same engine recordly's working Windows installer uses. We drop
// Squirrel.Windows (per-user only, no custom install dir, fragile updates) for a
// real NSIS installer: per-user or per-machine, custom install directory, and a
// proper uninstaller.
//
// `buildForge` speaks Forge's legacy v5 function API, which Forge 7's class-based
// maker loader cannot resolve, so this thin MakerBase subclass adapts it.
//
// Config shape (all optional): { appId, productName, icon, nsis }.

export default class MakerNSIS extends MakerBase {
	name = "nsis";
	defaultPlatforms = ["win32"];

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
				win: [`nsis:${targetArch}`],
				config: {
					appId: cfg.appId,
					productName: cfg.productName ?? appName,
					directories: { output },
					// Forge owns publishing (the workflow uploads via `gh release`).
					// `null` stops electron-builder from inferring a GitHub publish
					// target from package.json `repository` and trying to upload,
					// which fails in CI with no GH_TOKEN set.
					publish: null,
					...(cfg.icon ? { win: { icon: cfg.icon } } : {}),
					nsis: {
						// A real installer, not Squirrel's silent per-user drop.
						oneClick: false,
						perMachine: false,
						allowToChangeInstallationDirectory: true,
						createDesktopShortcut: true,
						createStartMenuShortcut: true,
						...cfg.nsis,
					},
				},
			},
		);
	}
}

export { MakerNSIS };
