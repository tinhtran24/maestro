# Plan: Clone the Maestro Frontend into `app/` — Shell Choice (Wails vs. Electron)

> **DECISION (locked): Option B — Wails v3 (Go-native shell).** The app in `app/`
> reuses Maestro's React renderer verbatim but rebuilds the shell in Go with Wails; the
> preview browser and auto-updater are reimplemented natively. See Option B below
> for the responsibility mapping and `04_milestones.md` (shell phases now target
> Wails, not Electron).

> Task: clone `maestro-main/frontend/` into `maestro/app/`, driving the
> cloned `maestro` daemon (`maestro/backend`). Resolved question: **build the app with
> Wails** (Go-native), not Electron.

## Key insight that makes both viable

The cloned backend is a **standalone local daemon** exposing HTTP + SSE + WebSocket
at `/api/v1`. The Maestro renderer is a **plain React SPA** that talks to that daemon
over the network — it does **not** depend on Electron for data. Electron is only
the _shell_: it (a) owns daemon lifecycle, (b) hosts an embedded BrowserView for
`maestro preview`, (c) does auto-update, (d) exposes a small IPC bridge.

So the renderer (routes, components, hooks, API client, terminal, stores — ~90% of
the frontend code) is **portable to either shell unchanged**. The decision only
affects the ~10% "shell" layer (`src/main/*`, `preload.ts`, `shared/daemon-*`).

## Option A — Electron (clone Maestro tech structure verbatim) ★ fidelity

Port `maestro-main/frontend/` into `app/` as-is (Electron 33 + Forge +
Vite + React 19). See `00`–`04` in this folder for the full breakdown.

- **Pros**: highest fidelity, fastest to working parity, every `src/main/*`
  module (daemon-owner, browser-view-host, auto-updater, supervisor-link) and its
  tests port 1:1. Embedded preview BrowserView and electron-updater work out of the
  box. Matches Maestro's DESIGN "clone verbatim" rule exactly.
- **Cons**: ships Chromium + Node (~big binary); a second toolchain (npm/Node)
  alongside Go; diverges from Maestro's original Wails choice.
- **Effort**: lowest — it's a port, following milestones `04_milestones.md`.

## Option B — Wails v3 (Go-native shell, reuse the React renderer) ★ native/light

Keep the **same renderer** (React SPA via Vite) but replace the Electron shell with
a Wails app: a Go `main` that opens a webview window loading the built SPA, plus a
thin Go bridge for the shell responsibilities.

Mapping of Electron shell → Wails equivalent:

| Electron responsibility            | Wails v3 equivalent                                                                                     |
| ---------------------------------- | ------------------------------------------------------------------------------------------------------- |
| `main.ts` window/app lifecycle     | Go `application.New(...)` + `Window` options                                                            |
| `daemon-owner`/`daemon-launch`     | Go: reuse `backend`'s daemon launch/attach directly (in-proc or exec `maestro daemon`)                  |
| `preload.ts` + IPC bridge          | Wails bindings (Go methods) + `wails/runtime` events                                                    |
| `browser-view-host` (preview)      | Reimplement: second webview / OS webview child, or an in-page `<iframe>`/`<webview>` to the preview URL |
| `auto-updater` (electron-updater)  | Reimplement with a Go updater (e.g. self-update / goreleaser assets)                                    |
| `userData` → `~/.maestro/electron` | pin app data under `~/.maestro` in Go                                                                   |

- **Pros**: single Go toolchain shared with the backend; much smaller binary; can
  call the cloned daemon package **directly in-process** (no bundled child binary
  to build/ship); consistent with Maestro's original direction.
- **Cons**: the shell layer is a **reimplementation, not a port** — `src/main/*`
  tests don't carry over. The two heaviest shell features need real work:
  - **Embedded preview browser**: Electron's `BrowserView` has no 1:1 Wails
    equivalent; use a nested OS webview or an `<iframe>` to the daemon-served
    preview URL (`GET /sessions/{id}/preview`).
  - **Auto-update**: build a Go updater.
- **Effort**: higher — renderer ports cleanly, shell must be rebuilt in Go.

## Recommendation

**Two-stage:** start with **Option A (Electron)** to reach full feature parity fast
and lock the renderer↔daemon contract, _then_ optionally migrate the shell to
**Option B (Wails)** once parity is proven — because the renderer is shell-agnostic,
that migration is a contained swap of the `src/main/*` layer, not a rewrite.

If a lightweight, Go-only single-toolchain app is a hard requirement up front, go
**straight to Wails (B)** and budget for reimplementing preview + updater.

Either way: put the app in `app/`, keep the renderer identical to Maestro, and generate
`app/src/api/schema.ts` from `backend/internal/httpd/apispec/openapi.yaml`.

## Decision needed

Pick the shell before Phase 0 of `04_milestones.md`:

- **A) Electron** — port verbatim, fastest parity (recommended first stage).
- **B) Wails v3** — Go-native shell, reuse renderer, reimplement preview + updater.

The rest of the plans (`01`–`04`) apply to the renderer under **both** options; only
`02_app_shell_and_routing.md`'s "Electron main process" section changes for B.
