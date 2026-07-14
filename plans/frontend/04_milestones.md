# Milestones & Delivery Plan

Phased plan to clone the Maestro frontend onto the already-cloned Maestro backend. Each
phase ends with a runnable, testable increment. "Port" = copy from
`maestro-main/frontend/src/**` and adjust names/paths (`maestro` → maestro,
`~/.maestro` → `~/.maestro`, backend module path).

## Phase 0 — Scaffold (foundation)

- [ ] Create `maestro/frontend/` from the Maestro frontend: `package.json`, `tsconfig`,
      `vite.*.config.ts`, `forge.config.ts`, `components.json`, `index.html`,
      `vitest.config.ts`, `playwright.config.ts`, Tailwind setup.
- [ ] Install deps (npm or pnpm — match Maestro's lockfile choice).
- [ ] Generate API types: `npm run api:ts` → `src/api/schema.ts` from
      `backend/internal/httpd/apispec/openapi.yaml`. Fix the script's relative path.
- [ ] Update `scripts/build-daemon.mjs` to build `maestro/backend/cmd/maestro`.
- **Done when**: `npm run dev:web` serves an empty shell; `npm run typecheck` passes.

## Phase 1 — Electron shell + daemon lifecycle

- [ ] Port `src/main/*`, `src/preload.ts`, `src/shared/daemon-*`, `shell-env`.
- [ ] Pin `userData` to `~/.maestro/electron`.
- [ ] Wire discover → attach → launch (bundled `maestro`) → takeover.
- [ ] Port `useDaemonStatus` + `lib/daemon-status`.
- **Done when**: `npm run dev` opens the window, spawns/attaches the daemon, and the
  status indicator shows "connected". Main-process tests pass.

## Phase 2 — App chrome + routing

- [ ] Port routes tree (`routes/*`), `__root`, `_shell`, providers.
- [ ] Port `Sidebar`, `ShellTopbar`, `TitlebarNav`, `CenterPane`,
      `SessionInspector`, `useResizable`, `ui-store`.
- [ ] Wire `QueryClientProvider`, router, `TelemetryBoundary`.
- **Done when**: navigation between dashboard / project / session / PRs / settings
  works with empty states. Shell/topbar/sidebar tests pass.

## Phase 3 — Read path (board + live events)

- [ ] Port `api-client`, `query-client`, `events-connection`, `event-transport`,
      `useEventsConnection`, `useWorkspaceQuery`.
- [ ] Port `SessionsBoard`, `BoardEmptyState`, `DashboardSubhead`.
- [ ] Consume `GET /projects`, `GET /sessions`, `GET /events` (SSE resume).
- **Done when**: real projects/sessions render and live-update from the daemon.

## Phase 4 — Projects & sessions write path

- [ ] Port `CreateProjectFlow`, `CreateProjectAgentSheet`, `IntakeFields`,
      `ProjectSettingsForm`, `NewTaskDialog`, `SessionView`, `rename-session`,
      `useSessionScmSummary`, `useAgentsQuery`, `agent-options`.
- [ ] Wire create/config/delete project; spawn/send/kill/restore/rollback session.
- **Done when**: user can register a project, spawn a worker, send a prompt, and
  kill it end-to-end.

## Phase 5 — Terminal

- [ ] Port `XtermTerminal`, `TerminalPane`, `useTerminalSession`, `terminal-mux`,
      `terminal-themes`.
- [ ] Connect `GET /mux` WebSocket; multiplex all session PTYs.
- **Done when**: a live agent terminal streams and accepts input; mux tests pass.

## Phase 6 — Preview / embedded browser

- [ ] Port `BrowserPanel`, `useBrowserView`, main `browser-view-host`.
- [ ] Wire preview start/stop/status + file serving.
- **Done when**: inspector "Browser" tab renders a session preview URL.

## Phase 7 — PRs, reviews, orchestrators

- [ ] Port `PullRequestsPage`, `PRSummaryDisplay`, `pr-display`,
      orchestrator dialogs + `spawn/restart-orchestrator`.
- [ ] Wire PR claim/merge/resolve-comments, review trigger/submit, orchestrator CRUD.
- **Done when**: PR list + summary render; review can be triggered; orchestrator can
  be spawned/replaced.

## Phase 8 — Notifications, settings, migration, updates

- [ ] Port `NotificationCenter`, `useNotificationsQuery`, `notifications`.
- [ ] Port `GlobalSettingsForm`, `UpdatesSection`, `update-settings`, `auto-updater`.
- [ ] Port `MigrationPopup`, `MigrationSection`, `useMigrationOffer` + import wiring.
- **Done when**: notifications stream and mark-read; settings persist; legacy import
  offer works; update check runs.

## Phase 9 — Telemetry, polish, packaging

- [ ] Port telemetry (`telemetry`, `posthog-config`) behind opt-out boundary.
- [ ] Full Vitest + Playwright suites green at parity with Maestro.
- [ ] `electron-forge make` produces installers (zip/deb/rpm); auto-update channel set.
- **Done when**: `make` builds a distributable that boots against the Maestro daemon.

## Acceptance (whole app)

- Every feature in `03_features.md` works against `maestro/backend`.
- All ported unit + e2e tests pass.
- No writes outside `~/.maestro`.
- `src/api/schema.ts` is generated (not hand-written) and current with the backend spec.

## Suggested execution note

Phases 3–8 are largely independent feature ports and can be fanned out across
parallel agents once Phases 0–2 (scaffold + shell + routing) are in place.
