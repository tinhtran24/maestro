# Backend Contract — API the Frontend Consumes

The cloned backend (`maestro/backend`) exposes everything the frontend needs. This
doc enumerates the surface so the renderer can be built against it. All routes are
mounted under `/api/v1` (see `backend/internal/httpd/router.go`). Types come from
`backend/internal/httpd/apispec/openapi.yaml` via `openapi-typescript`.

## Transport types

- **REST/JSON** — standard CRUD, wrapped in the envelope (`httpd/envelope`).
- **SSE** — `GET /api/v1/events` (global event bus) and
  `GET /api/v1/notifications/stream` (notification feed). Consumed via
  `useEventsConnection` / `event-transport.ts` with `Last-Event-ID` resume.
- **WebSocket** — `GET /api/v1/mux` (terminal multiplexer). Consumed via
  `terminal-mux.ts` + xterm. One socket multiplexes all session PTYs.

## Health & lifecycle

| Method | Path            | Purpose                       |
| ------ | --------------- | ----------------------------- |
| GET    | `/healthz`      | liveness                      |
| GET    | `/readyz`       | readiness                     |
| POST   | `/shutdown`     | graceful daemon shutdown      |
| GET    | `/openapi.yaml` | live spec (for codegen/debug) |
| GET    | `/panic`        | dev-only recover test         |

## Projects

| Method | Path                    | Purpose                 |
| ------ | ----------------------- | ----------------------- |
| GET    | `/projects`             | list projects           |
| POST   | `/projects`             | create/register project |
| GET    | `/projects/{id}`        | project detail          |
| DELETE | `/projects/{id}`        | remove project          |
| PUT    | `/projects/{id}/config` | update project config   |

## Sessions (agent worker sessions)

| Method | Path                             | Purpose                              |
| ------ | -------------------------------- | ------------------------------------ |
| GET    | `/sessions`                      | list (filter: project/status/active) |
| POST   | `/sessions`                      | spawn session                        |
| GET    | `/sessions/{sessionId}`          | session detail                       |
| PATCH  | `/sessions/{sessionId}`          | rename / mutate                      |
| POST   | `/sessions/{sessionId}/send`     | send message/prompt to agent         |
| POST   | `/sessions/{sessionId}/activity` | report activity                      |
| POST   | `/sessions/{sessionId}/kill`     | kill running session                 |
| POST   | `/sessions/{sessionId}/restore`  | restore session                      |
| POST   | `/sessions/{sessionId}/rollback` | rollback session state               |
| POST   | `/sessions/cleanup`              | bulk cleanup                         |

## Preview (embedded browser panel)

| Method | Path                                    | Purpose               |
| ------ | --------------------------------------- | --------------------- |
| POST   | `/sessions/{sessionId}/preview`         | start preview server  |
| GET    | `/sessions/{sessionId}/preview`         | preview status/url    |
| DELETE | `/sessions/{sessionId}/preview`         | stop preview          |
| GET    | `/sessions/{sessionId}/preview/files/*` | serve workspace files |

## Reviews & Pull Requests

| Method | Path                                    | Purpose                |
| ------ | --------------------------------------- | ---------------------- |
| GET    | `/sessions/{sessionId}/reviews`         | list reviews           |
| POST   | `/sessions/{sessionId}/reviews/trigger` | trigger Maestro review |
| POST   | `/sessions/{sessionId}/reviews/submit`  | submit review          |
| GET    | `/sessions/{sessionId}/pr`              | PR summary for session |
| POST   | `/sessions/{sessionId}/pr/claim`        | claim PR for session   |
| POST   | `/prs/{id}/merge`                       | merge PR               |
| POST   | `/prs/{id}/resolve-comments`            | resolve PR comments    |

## Orchestrators

| Method | Path                  | Purpose                    |
| ------ | --------------------- | -------------------------- |
| GET    | `/orchestrators`      | list orchestrator sessions |
| POST   | `/orchestrators`      | spawn orchestrator         |
| GET    | `/orchestrators/{id}` | orchestrator detail        |

## Agents (catalog / readiness)

| Method | Path                    | Purpose                   |
| ------ | ----------------------- | ------------------------- |
| GET    | `/agents`               | agent catalog + readiness |
| POST   | `/agents/refresh`       | refresh catalog           |
| POST   | `/agents/{agent}/probe` | probe a specific agent    |

## Notifications

| Method | Path                      | Purpose       |
| ------ | ------------------------- | ------------- |
| GET    | `/notifications`          | list          |
| GET    | `/notifications/stream`   | SSE stream    |
| PATCH  | `/notifications/{id}`     | mark one read |
| POST   | `/notifications/read-all` | mark all read |

## Legacy import

| Method | Path      | Purpose                                   |
| ------ | --------- | ----------------------------------------- |
| GET    | `/import` | discover legacy Maestro install to import |
| POST   | `/import` | run import                                |

## Streaming & telemetry internals

- `GET /events` — global SSE bus (query: `after`, `limit`, `project`, `status`…).
- `GET /mux` — terminal WebSocket multiplexer.
- `POST /internal/telemetry/cli-invoked`, `POST /internal/telemetry/cli-usage-error`
  — telemetry ingestion (renderer mirrors via `daemon-telemetry.ts`).

## Codegen workflow

```bash
# from frontend/
npm run api:ts   # regenerates src/api/schema.ts from backend openapi.yaml
```

Run this whenever backend handlers/DTOs change. CI should fail if `schema.ts` is
stale.
