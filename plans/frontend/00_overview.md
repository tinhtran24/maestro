# Frontend App — Overview & Architecture

> Goal: build a desktop frontend on Thanos that is a **verbatim clone** of the
> agent-orchestrator (AO) frontend, driving the AO backend that now lives under
> `thanos/backend/` (module `github.com/tinhtran/thanos/backend`).
>
> The backend has already been cloned in full (474 Go files, all features). These
> plans cover the frontend app that consumes it.

## What we are cloning

The AO frontend is an **Electron desktop app** whose renderer is a React SPA. It
talks to the local Go daemon (`ao daemon`) over HTTP + SSE + WebSocket. The
Electron main process owns daemon lifecycle, an embedded browser panel, and
auto-updates.

Source of truth: `agent-orchestrator-main/frontend/`. Clone its looks, structure,
and behavior verbatim (per AO's `DESIGN.md` "clone verbatim" rule).

## Tech stack (match exactly)

| Layer            | Choice                                                            |
| ---------------- | ---------------------------------------------------------------- |
| Shell            | Electron 33 + Electron Forge 7 (Vite plugin)                     |
| Renderer         | React 19, TypeScript 5.6                                          |
| Routing          | TanStack Router (file-based, `routeTree.gen.ts`)                 |
| Server state     | TanStack Query 5                                                 |
| Client state     | Zustand 5 (`ui-store.ts`)                                        |
| UI kit           | shadcn/ui over Radix UI + Tailwind CSS 4 + lucide-react icons    |
| Terminal         | xterm.js 5 + addons (fit, webgl, canvas, search, web-links, unicode11) |
| API client       | `openapi-fetch` typed from generated `src/api/schema.ts`        |
| Panels           | `react-resizable-panels`                                        |
| Telemetry        | posthog-js (behind a boundary; opt-out friendly)                |
| Build/updates    | app-builder-lib, electron-updater, forge makers (zip/deb/rpm)   |
| Tests            | Vitest + Testing Library (unit), Playwright (e2e)               |

## Process model

```
┌─ Electron main (Node) ────────────────────────────────────────────┐
│  • daemon-owner / daemon-launch / daemon-attach / daemon-takeover  │
│    → discover, spawn, adopt the local `ao` daemon                  │
│  • browser-view-host → embedded BrowserView for `preview`/inspector │
│  • auto-updater → electron-updater                                 │
│  • supervisor-link → renderer ↔ main IPC bridge (preload)          │
│  userData pinned to ~/.thanos/electron (mirror AO's ~/.ao rule)    │
└───────────────────────────────────────────────────────────────────┘
        │ IPC (preload bridge)          │ HTTP/SSE/WS (localhost daemon)
┌───────▼───────────────────────────────▼───────────────────────────┐
│  Renderer (React SPA)                                              │
│  routes → components → hooks → lib(api-client, events, terminal)   │
└───────────────────────────────────────────────────────────────────┘
```

## Backend contract (already available)

The renderer consumes the daemon's REST + streaming API mounted at `/api/v1`
(see `01_backend_contract.md`). The types are generated from the backend's
OpenAPI spec:

```
openapi-typescript backend/internal/httpd/apispec/openapi.yaml -o src/api/schema.ts
```

This is the single source of truth for request/response shapes — **do not
hand-write API types**.

## State-dir rule (mirror AO's hard rule)

All app state must resolve under `~/.thanos` (the backend's data dir; overridable
via the backend's data-dir env). Pin Electron `userData` to `~/.thanos/electron`.
Never write to `~/Library/Application Support` or other OS defaults.

## Plan documents

1. `00_overview.md` — this file.
2. `01_backend_contract.md` — REST/SSE/WS endpoints the renderer consumes.
3. `02_app_shell_and_routing.md` — Electron main process + renderer routing/layout.
4. `03_features.md` — feature-by-feature clone checklist.
5. `04_milestones.md` — phased delivery plan with acceptance criteria.
