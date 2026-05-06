# greedy-target-engine

A small Go service that, given an `(app, country, os)` tuple, returns the ad campaigns whose targeting rules allow it. Originally written for the GreedyGame backend take-home; rebuilt here as a portfolio piece.

## What it does

- `GET /v1/delivery?app=...&country=...&os=...` returns matching `ACTIVE` campaigns.
- `GET /v2/delivery?...` is the same behavior, served through a go-kit pipeline (decoder → endpoint → service → store) so the layering is explicit.
- `GET /healthz` liveness, `GET /readyz` checks the DB, `GET /metrics` Prometheus.

Response shape:
```json
[{"cid":"spotify","name":"Spotify - Music for everyone","img":"https://somelink","cta":"Download"}]
```
`204 No Content` on no match, `400` on a missing query param, `500` on a DB error.

## Run it

```bash
docker-compose up -d            # app + postgres + prometheus + grafana
curl 'http://localhost:8080/v1/delivery?app=com.gametion.ludokinggame&country=us&os=android'
```

Local without Docker:
```bash
docker-compose up postgres -d
make run
```

Tests (the integration ones expect Postgres on `localhost:5432`):
```bash
make test
```

## Data model

Two tables:

- `campaigns(cid PK, name, img, cta, status)` where `status` is `ACTIVE` or `INACTIVE`.
- `targeting_rules(id, cid FK, include_country TEXT[], exclude_country TEXT[], include_os TEXT[], exclude_os TEXT[], include_app TEXT[], exclude_app TEXT[])`.

A `NULL` array means "no constraint on this dimension". A non-null array on `include_*` means "only these values"; on `exclude_*` means "everything except these". Both can be set on the same rule — exclude wins.

## How matching works

The whole rule set is held in process memory as a snapshot. Each request walks the snapshot in Go and applies the include/exclude predicates — no SQL on the request path. See `internal/cache/cache.go`.

The snapshot is rebuilt:

- Once at boot.
- Whenever Postgres fires `NOTIFY targeting_changes`. Triggers on `campaigns` and `targeting_rules` send the notification on any insert/update/delete; the listener (`internal/cache/listener.go`) coalesces bursts via a 200ms debounce and reloads. Schema is in `internal/migrate/sql/init.sql`.

This trades a tiny amount of staleness (≤200ms after the writer commits) for sub-millisecond reads and zero query load on the DB. The DB stays the source of truth — restart the service and it rebuilds from scratch.

The earlier SQL-per-request implementation lived in an `internal/campaigns` package; it's gone now.

## Environment

| Variable | Default | Notes |
| --- | --- | --- |
| `DATABASE_URL` (or `DB_URL`) | — | Full Postgres URL. If set, takes precedence over the `DB_*` pieces. |
| `DB_HOST` / `DB_PORT` / `DB_NAME` / `DB_USER` / `DB_PASSWORD` / `DB_SSL_MODE` | `localhost` / `5432` / `targeting_db` / `postgres` / `password` / `disable` | Used when `DATABASE_URL` is unset. |
| `DB_MAX_OPEN_CONNS` | 25 | |
| `DB_MAX_IDLE_CONNS` | 10 | |
| `APPLY_SEED` | `false` | If `true`, the embedded `seed.sql` runs on boot. Idempotent. |
| `PORT` | `8080` | Render and similar PaaS inject this. |

Schema lives in `internal/migrate/sql/` and is embedded into the binary; `init.sql` runs on every boot, `seed.sql` only when `APPLY_SEED=true`.

## Live demo

Deployed split-stack:

- Postgres on [Neon](https://neon.tech) (free tier).
- Backend container on [Render](https://render.com) (free tier — first request after idle takes ~30s while it cold-starts).

Wiring it up yourself:

1. **Neon**: create a project, copy the connection string (the one with `?sslmode=require`).
2. **Render**: New → Web Service → connect this repo. Runtime: Docker. Plan: free. Add env vars: `DATABASE_URL` (paste from Neon), `APPLY_SEED=true` for the first deploy (set to `false` afterwards). Render auto-injects `PORT`.
3. First boot applies the schema and seed; subsequent boots only apply the schema.

## Observability

- `/metrics` exposes `delivery_requests_total{status}`, `delivery_request_duration_seconds{status}`, `db_query_duration_seconds`.
- The compose stack starts Prometheus on `:9090` (scraping `app:8080`) and Grafana on `:3000` (admin/admin). Datasource and dashboards are not provisioned — add them manually.
- Logs are JSON via `slog`.

## What's missing / what i'd do next

- Bench harness with `pgbench` or a Go bench so the perf claims have numbers behind them.
- Auth on `/v1/delivery` (the take-home brief didn't ask, but a real ad-serving endpoint shouldn't be open).
- Multi-instance cache coordination — the LISTEN/NOTIFY model fans out fine, but two replicas reload independently, which is wasteful at scale. A delta channel or a versioned snapshot pull would fix it.
- Provisioned Grafana dashboards.
