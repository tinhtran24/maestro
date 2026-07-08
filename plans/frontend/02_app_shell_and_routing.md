# App Shell & Routing

Clone the Electron main process and renderer shell/routing from
`thanos-main/frontend/src`. This doc lists the modules to port and
their responsibilities.

## Electron main process (`src/main/`)

Port these modules verbatim (each has a `.test.ts` alongside — port those too):

| Module                 | Responsibility                                                   |
| ---------------------- | --------------------------------------------------------------- |
| `app-state.ts`         | window/app lifecycle state, single-instance lock                |
| `daemon-owner.ts`      | decide whether this app owns/spawns the daemon                  |
| `browser-view-host.ts` | embedded `BrowserView` for the preview / inspector browser tab  |
| `supervisor-link.ts`   | IPC bridge wiring between main and renderer (via preload)       |
| `auto-updater.ts`      | electron-updater integration                                    |
| `update-settings.ts`   | update channel/prefs persistence                                |

### Shared daemon plumbing (`src/shared/`)

Port: `daemon-discovery.ts`, `daemon-launch.ts`, `daemon-attach.ts`,
`daemon-takeover.ts`, `daemon-status.ts`, `shell-env.ts`, `telemetry.ts`,
`posthog-config.ts`. These implement: find a running daemon → attach, else launch
the bundled `ao` binary, else take over a stale one.

> The daemon binary is built by `scripts/build-daemon.mjs` and bundled. Point it at
> `thanos/backend` (`go build ./cmd/ao`). Update the script's source path.

### Preload bridge (`src/preload.ts`, `src/renderer/lib/bridge.ts`)

Exposes a typed, minimal surface to the renderer (daemon status, browser-view
control, update actions). No `nodeIntegration` in the renderer.

### State-dir pinning

In `main.ts`, pin `app.setPath('userData', '~/.thanos/electron')` before app ready.
Mirror Thanos's hard rule — all state under `~/.thanos`.

## Renderer routing (`src/renderer/routes/`)

File-based TanStack Router. Port this exact tree (`routeTree.gen.ts` is generated):

```
__root.tsx                                              # root shell, providers
_shell.tsx                                              # app chrome: sidebar + topbar + panels
  _shell.index.tsx                                      # dashboard (sessions board)
  _shell.projects.$projectId.tsx                        # project view
  _shell.projects.$projectId_.settings.tsx             # project settings
  _shell.projects.$projectId_.sessions.$sessionId.tsx  # session view (project-scoped)
  _shell.sessions.$sessionId.tsx                        # session view (global)
  _shell.prs.tsx                                         # pull requests page
  _shell.settings.tsx                                    # global settings
```

### Providers (in `__root.tsx` / `main.tsx`)

- `QueryClientProvider` (`lib/query-client.ts`)
- Router provider
- `TelemetryBoundary` (posthog, opt-out aware)
- Tooltip/Toast providers from Radix

## Layout (`_shell.tsx`)

Three-region chrome, resizable via `react-resizable-panels` (`useResizable`):

```
┌────────────┬──────────────────────────┬─────────────────────┐
│  Sidebar   │  CenterPane              │  SessionInspector   │
│  (projects,│  (board / session view / │  (details / browser │
│   nav)     │   PRs / settings)        │   / terminal tabs)  │
└────────────┴──────────────────────────┴─────────────────────┘
  ShellTopbar across the top; TitlebarNav for window controls
```

## Client state (`src/renderer/stores/ui-store.ts`)

Zustand store for UI-only state: panel sizes, selected session, inspector tab,
sidebar collapse, theme. Server state stays in TanStack Query — do not duplicate.

## Styling

- Tailwind CSS 4 via `@tailwindcss/vite`; global `styles.css`.
- shadcn components under `components/ui/` (`components.json` config).
- Terminal keeps its own palette (`lib/terminal-themes.ts`); refined-blue accent
  elsewhere, per Thanos DESIGN.md.
