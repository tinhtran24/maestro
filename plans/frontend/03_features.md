# Feature Clone Checklist

Every AO frontend feature, the components/hooks/lib that implement it, and the
backend endpoints it drives. Port components **and their `.test.tsx`** files.

## 1. Dashboard / Sessions Board

- **UI**: `SessionsBoard.tsx`, `BoardEmptyState.tsx`, `DashboardSubhead.tsx`,
  `Sidebar.tsx`, `ShellTopbar.tsx`.
- **Data**: `useWorkspaceQuery.ts` (projects+sessions), `useEventsConnection.ts`
  (live updates), `lib/mock-data.ts` (dev fixtures).
- **Backend**: `GET /sessions`, `GET /projects`, `GET /events` (SSE).
- Board groups sessions by status/project; live-updates from the event bus.

## 2. Projects

- **UI**: `CreateProjectFlow.tsx`, `CreateProjectAgentSheet.tsx`,
  `ProjectSettingsForm.tsx`, `IntakeFields.tsx`.
- **Backend**: `GET/POST /projects`, `GET/DELETE /projects/{id}`,
  `PUT /projects/{id}/config`.
- Create/register a project (repo path + agent defaults), edit config, remove.

## 3. Sessions (worker agents)

- **UI**: `SessionView.tsx`, `CenterPane.tsx`, `SessionInspector.tsx`,
  `NewTaskDialog.tsx`.
- **Data**: `useSessionScmSummary.ts`, `lib/rename-session.ts`.
- **Backend**: `POST /sessions`, `GET/PATCH /sessions/{id}`,
  `POST /sessions/{id}/send|kill|restore|rollback|activity`,
  `POST /sessions/cleanup`.
- Spawn a worker in a project, send prompts, rename, kill, restore, rollback.

## 4. Terminal

- **UI**: `XtermTerminal.tsx`, `TerminalPane.tsx`.
- **Data**: `useTerminalSession.ts`, `lib/terminal-mux.ts`, `lib/terminal-themes.ts`.
- **Backend**: `GET /mux` (WebSocket multiplexer).
- One WS multiplexes all session PTYs; xterm + fit/webgl/search/web-links/unicode11
  addons. Preserve AO's terminal palette.

## 5. Embedded Browser / Preview

- **UI**: `BrowserPanel.tsx`.
- **Data**: `useBrowserView.ts`, main-process `browser-view-host.ts`.
- **Backend**: `POST/GET/DELETE /sessions/{id}/preview`,
  `GET /sessions/{id}/preview/files/*`.
- Inspector "Browser" tab renders a preview URL (or workspace `index.html`) in an
  Electron BrowserView. Mirrors the `ao preview` CLI command.

## 6. Pull Requests & Reviews

- **UI**: `PullRequestsPage.tsx`, `PRSummaryDisplay.tsx`.
- **Data**: `lib/pr-display.ts`.
- **Backend**: `GET /sessions/{id}/pr`, `POST /sessions/{id}/pr/claim`,
  `GET /sessions/{id}/reviews`, `POST /sessions/{id}/reviews/trigger|submit`,
  `POST /prs/{id}/merge`, `POST /prs/{id}/resolve-comments`.
- List PRs, view summary, claim, trigger AO review, submit, merge, resolve comments.

## 7. Orchestrators

- **UI**: `OrchestratorReplacementDialog.tsx`, `RestoreUnavailableDialog.tsx`.
- **Data**: `lib/spawn-orchestrator.ts`, `lib/restart-orchestrator.ts`,
  `lib/orchestrator-replacement-telemetry.ts`.
- **Backend**: `GET/POST /orchestrators`, `GET /orchestrators/{id}`.
- Spawn/replace/restart the orchestrator session; handle the replacement flow.

## 8. Agents catalog

- **Data**: `useAgentsQuery.ts`, `lib/agent-options.ts`.
- **Backend**: `GET /agents`, `POST /agents/refresh`, `POST /agents/{agent}/probe`.
- Populate agent pickers with readiness state; refresh/probe from settings.

## 9. Notifications

- **UI**: `NotificationCenter.tsx`.
- **Data**: `useNotificationsQuery.ts`, `lib/notifications.ts`.
- **Backend**: `GET /notifications`, `GET /notifications/stream` (SSE),
  `PATCH /notifications/{id}`, `POST /notifications/read-all`.

## 10. Settings

- **UI**: `GlobalSettingsForm.tsx`, `ProjectSettingsForm.tsx`, `UpdatesSection.tsx`,
  `MigrationSection.tsx`.
- **Data**: main-process `update-settings.ts`, `auto-updater.ts`.
- Global prefs, per-project config, update channel, migration entry point.

## 11. Legacy import / Migration

- **UI**: `MigrationPopup.tsx`, `MigrationSection.tsx`.
- **Data**: `useMigrationOffer.ts`.
- **Backend**: `GET /import`, `POST /import`.
- Detect a legacy AO install and offer to import projects/sessions.

## 12. Daemon status & connection

- **Data**: `useDaemonStatus.ts`, `lib/daemon-status.ts`, `lib/daemon-telemetry.ts`,
  `shared/daemon-*`.
- **Backend**: `GET /healthz`, `GET /readyz`, `POST /shutdown`.
- Show connection state; drive discover→attach→launch→takeover flow.

## 13. Events / live sync

- **Data**: `lib/events-connection.ts`, `lib/event-transport.ts`,
  `useEventsConnection.ts`.
- **Backend**: `GET /events` (SSE, `Last-Event-ID` resume).
- Single global stream fans out to Query cache invalidations + store updates.

## 14. Telemetry

- **UI/Data**: `TelemetryBoundary.tsx`, `lib/telemetry.ts`, `shared/telemetry.ts`,
  `shared/posthog-config.ts`.
- **Backend**: `POST /internal/telemetry/cli-*`.
- posthog-js behind a boundary; respect opt-out; no PII.

## Cross-cutting

- **API client**: `lib/api-client.ts` (typed `openapi-fetch` over `src/api/schema.ts`).
- **Query client**: `lib/query-client.ts`.
- **Utilities**: `lib/utils.ts`, `lib/format-time.ts`, `hooks/use-mobile.tsx`,
  `hooks/useResizable.ts`.
- **Icons**: `components/icons.tsx` + lucide-react.

## Testing parity

- Unit: Vitest + Testing Library — port every `*.test.tsx`/`*.test.ts`.
- E2E: Playwright specs under `frontend/e2e/`.
- Keep coverage at parity; a feature isn't "cloned" until its tests pass.
