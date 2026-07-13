# `to start` Bootstrapper + npm Deprecation: Implementation Spec

> **Status:** ready for build (Track A). Grounded against the real codebase on
> branch `feat/to-start-bootstrapper` (= `upstream/main` + PR #2185) on 2026-06-26.
> Every "current state" claim carries a `file:line` reference.
>
> **This is NOT a new JS launcher package.** The `to` binary that npm ships is the
> existing Go cobra CLI (`backend/cmd/to`). This effort rewrites one subcommand,
> `to start`, to fetch and open the desktop app. Everything else in the CLI is
> already wired and rides along.

---

## 0. Goal

npm `to` is the **legacy on-ramp** for users who already have `to` on their PATH.
We are deprecating npm as an app-distribution path:

- `npm update` swaps in our **new Go `to` binary** (the whole CLI), replacing the
  old one in place. No fresh-install story; the audience is existing users.
- The **`to start`** subcommand is rewritten: instead of starting a daemon, it
  **fetches the desktop app from GitHub Releases and opens it**.
- The **desktop app owns the daemon**, auto-update, relocation, and all state. The
  CLI becomes a thin client of the app-owned daemon.

`to start` is the one-time bridge that moves a CLI user onto the canonical,
auto-updating desktop build. It is dumb about versions: its only job is "is the
app present? if not, fetch it; then open it."

---

## 1. Ground truth (what the code actually is today)

### 1.1 App identity and release target

| Fact                         | Value                                                                  | Source                          |
| ---------------------------- | ---------------------------------------------------------------------- | ------------------------------- |
| Product / bundle name        | **`Thanos.app`** (spaced)                                  | `frontend/forge.config.ts:9,50` |
| Bundle id                    | `dev.thanos.desktop`                                       | `frontend/forge.config.ts:8`    |
| Executable name              | `thanos`                                                   | `frontend/forge.config.ts`      |
| **Release repo (canonical)** | **`tinhtran24/thanos`**                                  | per release owner               |
| Forge publisher repo (TODAY) | `tinhtran/thanos` — **stale, must change to tinhtran24** | `frontend/forge.config.ts:86`   |
| GitHub release mode          | **`draft: true`**, `prerelease: false`                                 | `frontend/forge.config.ts`      |

> `tinhtran/thanos` was the **temporary** home during the rewrite; the
> code is now ported and releases land on **`tinhtran24/thanos`**.
> The forge publisher still points at `tinhtran` and must be corrected (task T3).
> The Go **module path** is also `github.com/tinhtran/thanos`; renaming
> the module is a large, separate change and is **out of scope** here (it does not
> affect the release/download URL).

### 1.2 Release / build pipeline

- Workflow: `.github/workflows/frontend-release.yml`. Triggers: tag `desktop-v*`,
  `workflow_dispatch`. Build: `npm run publish` → `build:daemon` +
  `electron-forge publish`.
- **Matrix: `[macos-latest, windows-latest]` only** (`:28`) — no Linux; deb/rpm
  makers configured but never run (upstream issue tinhtran24/thanos#2191).
- Maker outputs (today): macOS `@electron-forge/maker-zip` → versioned `.zip`
  under `out/make/zip/darwin/<arch>/`; Windows `MakerNSIS` → `Thanos
Setup.exe` (per-user installer); Linux `maker-deb`/`maker-rpm` →
  `thanos-<version>.{deb,rpm}`.
- **No asset-rename step** and **`draft: true`** → a constant
  `releases/latest/download/<stable-name>` URL cannot resolve until both are fixed.

### 1.3 Versioning

- Frontend `frontend/package.json` `version: "0.0.0"`; daemon
  `backend/internal/cli/version.go:12` `Version = "dev"`; `build-daemon.mjs` runs
  `go build ./cmd/to` with **no `-ldflags`**. No real semver anywhere.

### 1.4 Signing / notarization / auto-update

- `osxSign`/`osxNotarize` are gated on secrets (`forge.config.ts:24-40`) that are
  **not set in CI**; the workflow header (`frontend-release.yml:13-15`) says builds
  are **UNSIGNED**.
- **Auto-update is already wired**: `frontend/src/main.ts:14` imports
  `updateElectronApp` from `update-electron-app`; `initAutoUpdates()`
  (`main.ts:817`) runs it when `app.isPackaged`. Inert today because builds are
  unsigned and version is `0.0.0` (its own comment, `main.ts:813-816`).

### 1.5 `~/.thanos` state and app lifecycle

- Canonical home `~/.thanos` (`backend/internal/config/config.go:296`,
  `frontend/src/shared/daemon-discovery.ts:107`); overrides `THANOS_DATA_DIR`/`THANOS_RUN_FILE`.
- `userData` pinned to `~/.thanos/electron` (`main.ts:64`, before `whenReady`; CLAUDE.md
  hard rule).
- `~/.thanos/running.json` is written by the **daemon** (`backend/internal/runfile/runfile.go`
  `Write`, atomic temp+rename), read by the app (`daemon-discovery.ts parseRunFile`).
  Only `running.json` exists in `~/.thanos` today; **`app-state.json` does not exist yet**.
- App startup (`main.ts:822` `whenReady`): `registerRendererProtocol()` →
  `createWindow()` → `void startDaemon()` → `initAutoUpdates()`. The app already
  **spawns and owns the daemon** (`startDaemon`, spawns the bundled `to daemon`).
- `app.moveToApplicationsFolder()` is **not used** anywhere (macOS-only).
- Login-shell env resolved at startup via `zsh -ilc '… env -0'`
  (`frontend/src/shared/shell-env.ts:27`).

### 1.6 npm delivery of the Go binary (the packaging gap)

- The `to` binary is `backend/cmd/to` (`cmd/to/main.go` → `cli.Execute()`); the
  same binary serves as both the CLI and `to daemon`. `build-daemon.mjs` builds it
  to `frontend/daemon/to` and bundles it into the desktop app.
- **This repo has no npm-registry publish path for the `to` binary** (only
  electron-forge → GitHub Releases; no `NPM_TOKEN`, no publish workflow — research
  confirmed). The old Thanos npm package shipped `to` via npm; that delivery mechanism
  must be **ported/rebuilt here** (task T2). To honor "zero install scripts"
  (npm v12, est. July 2026, blocks unapproved install scripts), the Go binary
  should ship via **per-platform `optionalDependencies` packages** (the
  esbuild/turbo model: a tiny JS `bin` shim execs the right prebuilt binary), not
  via a `postinstall` download.

### 1.7 The Go `to` CLI surface (already wired)

`backend/cmd/to/main.go` → `backend/internal/cli`. Cobra root (`root.go:154-202`)
registers **all** of: `daemon` (hidden), **`start`**, `stop`, `status`, `doctor`,
`spawn`, `send`, `preview`, `hooks`, `launch`, `ptyhost`, `import`, `project`,
`session`, `orchestrator`, `review`, `completion`, `version`. These are real
(`doctor.go` is 20KB of health checks; `import.go` imports a legacy Thanos install).
The CLI is a thin client: commands "discover the local daemon, call its loopback
HTTP API, and format output" (`root.go:1-3`).

**Current `to start` (`start.go:54-119`):** starts the daemon (spawns `to daemon`,
waits for ready) and runs a first-boot legacy import (`maybeFirstBootImport`,
`start.go:84`). **This entire behavior is being replaced** (§6).

---

## 2. Decisions locked

1. **Releases land on `tinhtran24/thanos`.** Fix the forge publisher
   to match; the download URL uses it.
2. **`to start` = fetch + open the desktop app.** It no longer starts the daemon;
   the frontend owns the daemon. The current daemon-spawn logic in `start.go` is
   removed.
3. **npm ships the Go `to` binary**; existing users update in place. No JS launcher
   package.
4. **Marker = `~/.thanos/app-state.json`**, written only by the app, every launch.
5. **Scope = Track A only** (de-scope auto-update copy; Track B is separate).
6. **All three platforms; Windows installer is NSIS.**
7. **Two release targets, never conflated:**
   - **Production:** GitHub `tinhtran24/thanos`; npm = the real
     package name (legacy `to`). Cutting a prod release is a deliberate, gated
     step, never part of the dev/test loop.
   - **Test/dev:** GitHub **`harshitsinghbhandari/thanos`** (the fork);
     npm scope **`@theharshitsingh/to`**. All `to start` download/open testing runs
     against fork releases and the test npm scope.
     The download repo and npm scope are **build-time overridable** (§6.3, §8) so a
     test binary fetches from the fork and a prod binary from tinhtran24, with no
     code edit between them.

---

## 3. Scope

**In scope (Track A):**

- Rewrite the Go **`to start`** subcommand: `resolve → fetch → open` the desktop
  app, then print a deprecation notice. (`backend/internal/cli/start.go`.)
- Decide the fate of `to start`'s current first-boot legacy import (§6.4).
- **App-side:** write `~/.thanos/app-state.json` every launch (app is sole writer);
  own `moveToApplicationsFolder()` relocation (macOS).
- **Release wiring:** point forge publisher at `tinhtran24/thanos`,
  add stable version-free asset names, finalize the draft (or Releases-API
  fallback), add Linux to the matrix.
- **npm delivery** of the Go binary (port the old Thanos mechanism; zero install
  scripts via optionalDeps platform packages).
- macOS / Windows (NSIS) / Linux (deb/rpm or AppImage) fetch+open paths.

**Out of scope:**

- Track B: real version stamping, making the wired `update-electron-app` updater
  live, configuring signing/notarization CI secrets, any copy promising
  auto-update.
- Renaming the Go module path off `tinhtran` (separate, large, not needed here).
- The other CLI subcommands (already wired; untouched).

---

## 4. Core invariants (load-bearing)

1. **The npm package runs zero install scripts.** No `preinstall`/`install`/
   `postinstall`, no `binding.gyp`. Ship the Go binary via per-platform
   `optionalDependencies` + a JS `bin` shim, not a `postinstall` download.
2. **Filesystem is the source of truth; `app-state.json` is a fast-path hint.**
   Never trust its recorded path without `stat`-ing it.
3. **The app is the sole writer of `app-state.json`.** `to start` is read-only with
   respect to it. This is what makes the npm and website routes converge without an
   orphaned second copy.
4. **The app owns relocation** (`moveToApplicationsFolder()`), and rewrites the
   marker path afterward. `to start` never moves the app.
5. **`to start` is dumb about versions.** Decision is present-or-absent only; never
   compares versions. Updating an installed app is the app's own updater's job.
6. **Resolution order is fixed:** marker path → `stat` → known-location scan →
   fetch. Fetch only when both miss.
7. **Stable, version-free release asset names** so `to start` uses a constant URL.

---

## 5. The marker contract: `~/.thanos/app-state.json`

New file, **app-written**, mirroring the daemon's proven atomic write
(`backend/internal/runfile/runfile.go`: temp file in same dir → atomic rename).

```json
{
	"schemaVersion": 1,
	"appPath": "/Applications/Thanos.app",
	"version": "0.0.0",
	"installedAt": "2026-06-26T10:00:00Z",
	"lastReconciledAt": "2026-06-26T10:05:00Z",
	"installSource": "npm-bootstrap"
}
```

| Field              | Writer | Meaning                                                                          |
| ------------------ | ------ | -------------------------------------------------------------------------------- |
| `schemaVersion`    | app    | Marker format version.                                                           |
| `appPath`          | app    | Bundle path as of the last launch.                                               |
| `version`          | app    | `app.getVersion()`. For the tour/migration, NOT for `to start` update decisions. |
| `installedAt`      | app    | First marker write.                                                              |
| `lastReconciledAt` | app    | Last launch that touched the marker.                                             |
| `installSource`    | app    | `npm-bootstrap` / `website` / `github` / `unknown`; set only on first creation.  |

**Ownership:** only the app writes it, on **every launch**, self-healing a
stale/missing marker no matter how the app arrived. `to start` only reads it, only
after `stat`-ing the path.

---

## 6. The `to start` subcommand (Go) — the heart of this effort

Rewrite `backend/internal/cli/start.go`. Remove the daemon-spawn path
(`startDaemon`, `waitForReady`); the frontend owns the daemon now.

### 6.1 New algorithm

```
to start:
  app = resolveApp()              # marker → stat → known-location scan
  if app == "":
      app = fetchApp()            # download latest for this platform, place it
  opened = openApp(app)           # launch; pass --installed-via=npm-bootstrap
  printDeprecationNotice()        # the app owns any rich first-run tour
  if !opened: printManualOpen(app)
  return nil                      # never blocks/supervises the app
```

All of this is Go, in the `cli` package, reusing existing deps
(`Deps.CommandOutput`, `Deps.LookPath`, `Deps.Executable`) and the `~/.thanos`
resolution already in `backend/internal/config`.

### 6.2 `resolveApp()` (invariants 2, 5, 6)

1. Read `~/.thanos/app-state.json`; if `appPath` `stat`s as a usable bundle, return it.
2. Else scan known locations per platform (covers website installs / stale marker):
   - macOS: `/Applications/Thanos.app`, `~/Applications/…`
   - Windows: `%LOCALAPPDATA%\Programs\thanos\…`, `C:\Program Files\Thanos\…`
   - Linux: `/opt/Thanos/…`, `~/.local/bin`, `/usr/bin`
3. Else return empty → caller fetches. Never compare versions.

### 6.3 `fetchApp()` + `openApp()` — platform asymmetry (real design point)

Constant URL: `https://github.com/<owner>/<repo>/releases/latest/download/<stable-asset>`
(302 → asset; requires non-draft release + stable names, §8). `<owner>/<repo>` is
**build-time overridable**, not hardcoded: default `tinhtran24/thanos`
(prod), overridden to `harshitsinghbhandari/thanos` for test builds via
a `-ldflags -X …cli.releaseRepo=<owner>/<repo>` injection (mirrors how the daemon
version will be stamped). So the dev loop fetches from the fork; prod fetches from
tinhtran24, with no source edit.

- **macOS:** download `.zip` → unpack with **`ditto -x -k`** (preserves the `.app`
  signature; plain unzip corrupts it) → `open <app> --args --installed-via=npm-bootstrap`.
  The app relocates itself to `/Applications` on first launch.
- **Windows:** the asset is an **NSIS installer `.exe`** (not a runnable bundle).
  `fetch` downloads it; `open` runs the installer (interactive, or `/S` silent),
  then `resolveApp()` finds the installed exe and launches it.
- **Linux:** `.deb`/`.rpm` need privileged install, or switch the Linux artifact to
  an **AppImage** (single executable, no install) — better fit for fetch-and-run
  (decide §11).

### 6.4 The legacy first-boot import

`to start` currently runs `maybeFirstBootImport` (`start.go:84`, imports a legacy
Thanos install before the daemon starts). With the daemon-spawn removed, this must
move. Options (decide §11): (a) the **desktop app** runs the import when it first
boots its daemon; (b) drop it from `to start` and rely on the standalone `to
import` command (still wired). Recommended: (a), so the on-ramp still migrates
existing data.

### 6.5 Other subcommands / bare `to`

Unchanged — they stay wired and talk to the app-owned daemon's loopback API. Add a
one-line deprecation hint to the root long-help noting that npm is now an on-ramp
and the app is the home. Do **not** alter `stop`/`status`/`spawn`/etc. behavior.

---

## 7. App-side responsibilities

### 7.1 Marker write + relocation (new)

Hook into `app.whenReady()` (`main.ts:822`), **before** `createWindow()`, ordered
**relocate → write marker** (the marker must record the post-relocation path):

```ts
app.whenReady().then(async () => {
	if (process.platform === "darwin" && app.isPackaged) {
		try {
			app.moveToApplicationsFolder();
		} catch {
			/* declined / not movable */
		}
		// success restarts the app, so code past here runs only if no move happened
	}
	await writeAppStateMarker(); // atomic temp+rename, mirror runfile.Write
	registerRendererProtocol();
	createWindow();
	void startDaemon();
	initAutoUpdates();
});
```

`writeAppStateMarker()` records `app.getAppPath()`/`app.getVersion()` into
`~/.thanos/app-state.json`. On first creation, capture `installSource` from the
`--installed-via` arg `to start` passes (else `website`/`github`/`unknown`).

### 7.2 Already done — rely on it

Daemon ownership (`main.ts startDaemon` + the #2185 supervisor link,
`main/supervisor-link.ts`), login-shell env (`shell-env.ts:27`), and the `userData`
pin (`main.ts:64`) are in place. Do not re-implement.

---

## 8. Release / build wiring

- **Publisher repo is overridable** (`forge.config.ts:86`): default prod
  `tinhtran24/thanos`, but read from an env var (e.g.
  `THANOS_RELEASE_REPO`) so a fork build publishes to
  `harshitsinghbhandari/thanos`. The dev loop publishes a draft+finalize
  release **on the fork** and points the test binary's `cli.releaseRepo` at the same
  fork. Never publish to tinhtran24 from a test run.
- **Stable asset names:** add a release-workflow step renaming each maker output to
  space-free names (`thanos-darwin-arm64.zip`,
  `thanos-win32-x64.exe`, the Linux artifact per §11) before upload.
- **Finalize the draft:** flip `draft: false` or add a CI publish step; the constant
  URL only resolves for a published release.
- **`.zip` for macOS** unpacked with `ditto`; do not switch to `.tar.gz`.
- **Linux in the matrix:** add `ubuntu-latest` (#2191).
- **One tag drives versions** once Track B lands.

---

## 9. Track B prerequisites (NOT this effort; keeps v1 copy honest)

The `update-electron-app` updater is wired (§1.4) but inert until **both**: real
version stamping (bump `package.json`; inject daemon version via `-ldflags -X
…cli.Version=<tag>` in `build-daemon.mjs`) **and** signed+notarized macOS builds
(`CSC_LINK` + `APPLE_*` in CI). Until then, v1 copy must **not** promise
auto-update; users self-update by re-running `to start` or downloading from the
website.

---

## 10. Acceptance criteria / test matrix

| #   | Scenario                                                  | Expected                                                                                                                                                    |
| --- | --------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1   | `npm i -g @theharshitsingh/to` (test scope)               | Zero `allow-scripts` warning; nothing listed by `npm approve-scripts --allow-scripts-pending`.                                                              |
| 2   | `npm i -g @theharshitsingh/to --ignore-scripts` (v12 sim) | Install succeeds; `to` runs; `to start` works (binary delivered via optionalDeps, not a script).                                                            |
| 3   | Fresh macOS `to start`                                    | Fetches `.zip`, `ditto`-unpacks, opens `Thanos.app`; app relocates to `/Applications`; `~/.thanos/app-state.json` records the `/Applications` path. |
| 4   | Website install first, then `to start`                    | Known-location scan finds it; opens; no second copy fetched.                                                                                                |
| 5   | App trashed (marker stale), then `to start`               | Marker `stat` misses → scan misses → re-fetch.                                                                                                              |
| 6   | App relocated by the app                                  | Marker path rewritten; next `to start` opens the right path; no orphan.                                                                                     |
| 7   | Installed-but-old app, `to start`                         | Opens it and exits; does NOT fetch a newer one.                                                                                                             |
| 8   | Windows `to start`                                        | Downloads NSIS `.exe`, runs installer, resolves + opens installed exe.                                                                                      |
| 9   | Linux `to start`                                          | Fetches chosen artifact and launches.                                                                                                                       |
| 10  | `to stop`/`to status`/`to spawn` after `to start`         | Work against the app-owned daemon (CLI is a client).                                                                                                        |
| 11  | Existing CLI user runs `npm update` then `to start`       | New binary in place; `to start` no longer starts a daemon, it opens the app; their `to import` data migrates (per §6.4).                                    |

> `to start` opens the app through the calling shell's enriched env, so a green
> `to start` proves nothing about the Dock-launch path. Test the Dock path
> separately.

---

## 11. Open decisions (decide before the affected task)

1. **npm delivery mechanism** for the Go binary: per-platform `optionalDependencies`
   packages (recommended, zero-install-script) vs porting whatever the old Thanos
   package did. Test scope is **`@theharshitsingh/to`**; the **prod package name**
   (the legacy `to` users already have) still needs confirming, plus an `NPM_TOKEN`
   - publish workflow for each.
2. **Legacy first-boot import** (§6.4): move into the desktop app, or drop from
   `to start` and rely on `to import`?
3. **Linux artifact form:** `.deb`/`.rpm` (install) vs **AppImage** (fetch-and-run).
4. **Draft release finalization:** `draft: false` vs a CI publish step.
5. **Signing gate:** gate the launcher on signed+notarized builds, ship against
   unsigned (Gatekeeper/SmartScreen warnings), or treat signing as a parallel
   effort meeting at release?
6. **Download integrity:** SHA256 vs HTTPS-only vs `codesign --verify`.
7. **First-run tour + `installSource`:** in-app tour now (no auto-update promise),
   defer tour but keep `installSource`, or neither?
8. **Website URL** for the deprecation notice copy.
9. **Module-path rename** off `tinhtran` — confirm out of scope for this effort.

---

## 12. Task breakdown (for Thanos execution, dependency-ordered)

**Batch 1 — wiring (parallel):**

- **T1. Rewrite `to start` core (Go).** Replace `start.go` daemon-spawn with
  `resolveApp()` + the macOS fetch/open path + deprecation notice; remove
  `waitForReady`/daemon logic; decide §11.2. Check: on a mac with the app present,
  `to start` opens it and writes nothing; with it absent, it fetches+opens.
- **T2. npm delivery of the Go binary.** Per §11.1: optionalDeps platform packages
  - JS `bin` shim, zero install scripts; publish workflow. **Publish to the
    `@theharshitsingh/to` test scope**, not the prod package. Check: `npm i -g
@theharshitsingh/to --ignore-scripts` yields a working `to`.
- **T3. Release repo + asset wiring (override-driven).** Make the forge publisher
  repo + the `to start` download repo build-time overridable (§6.3, §8); add the
  stable-asset rename step; finalize the draft (§11.4); add Linux to the matrix.
  Check: a `workflow_dispatch` **on the fork** produces a published
  `harshitsinghbhandari/thanos` release whose
  `releases/latest/download/<stable-name>` 302-resolves. **No prod (tinhtran24)
  release is cut during development.**

**Batch 2 — app-side + macOS end-to-end (after T1):**

- **T4. App-side marker + relocation** (`main.ts whenReady`, §7.1). Check: a
  packaged launch writes/updates `~/.thanos/app-state.json` with the real bundle path.
- **T5. macOS `to start` end-to-end against the FORK release** (needs T3): build the
  test `to` with `cli.releaseRepo=harshitsinghbhandari/thanos`, install
  it from `@theharshitsingh/to`, run `to start`. Check: acceptance #3–#7 on a mac,
  fetching from the fork.

**Batch 3 — cross-platform + integrity (after T1/T3):**

- **T6. Windows path** (NSIS fetch+install+resolve, §6.3).
- **T7. Linux path** (§11.3).
- **T8. Download integrity** (§11.6).

**Batch 4 — rollout:**

- **T9.** Deprecation notice / optional tour + `installSource` (§11.7); legacy
  import placement (§11.2) if not done in T1.

> Track B (version stamping, signing, making the updater live) is a separate
> effort. Any copy added above must not promise auto-update until it lands.
