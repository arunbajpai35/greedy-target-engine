# greedy-target-engine

A small Go service that, given an `(app, country, os)` tuple, returns the ad campaigns whose targeting rules allow it.

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
| `API_SECRET` | unset | If set, `/v1/delivery` and `/v2/delivery` require an HMAC-signed request. `/healthz`, `/readyz`, `/metrics` stay open. |

Schema lives in `internal/migrate/sql/` and is embedded into the binary; `init.sql` runs on every boot, `seed.sql` only when `APPLY_SEED=true`.

## Live demo

https://greedy-target-engine.onrender.com — first request after idle takes ~30s while Render cold-starts the free instance.

```bash
curl 'https://greedy-target-engine.onrender.com/v1/delivery?app=com.gametion.ludokinggame&country=us&os=android'
```

Stack:

- Postgres on [Neon](https://neon.tech) (free tier, `ap-southeast-1`).
- Backend container on [Render](https://render.com) (free tier, Singapore).

Wiring it up yourself:

1. **Neon**: create a project, copy the **direct** connection string — *not* the pooled one. The pooled hostname has `-pooler` in it; PgBouncer in transaction mode silently drops `LISTEN/NOTIFY`, so cache invalidation breaks. The pooled URL still works for everything else, which makes it especially nasty to debug.
2. **Render**: New → Web Service → connect this repo. Runtime: Docker. Plan: free. Region: pick the same one as Neon. Env vars: `DATABASE_URL` (the direct URL from step 1), `APPLY_SEED=true` for the first deploy (flip to `false` afterwards). Render auto-injects `PORT`.
3. First boot applies the schema and seed; subsequent boots only apply the schema.

## Auth

The delivery endpoints are open by default so the live demo is easy to poke at. Setting `API_SECRET` flips on HMAC auth.

Request must carry:

- `X-Timestamp: <unix seconds>` — rejected if more than 5 minutes off server time
- `X-Signature: hex(hmac_sha256(secret, payload))`

Where `payload` is `<timestamp>\n<method>\n<path>\n<sorted-query>`. Query is `k=v` pairs joined with `&`, keys sorted alphabetically. Example:

```bash
SECRET=...
TS=$(date +%s)
PAYLOAD=$(printf "%s\nGET\n/v1/delivery\napp=com.gametion.ludokinggame&country=us&os=android" "$TS")
SIG=$(printf "%s" "$PAYLOAD" | openssl dgst -sha256 -hmac "$SECRET" -hex | awk '{print $NF}')
curl -H "X-Timestamp: $TS" -H "X-Signature: $SIG" \
  "http://localhost:8080/v1/delivery?app=com.gametion.ludokinggame&country=us&os=android"
```

The signature is checked with constant-time compare. Replay protection is the timestamp window — a captured request is useless after 5 minutes.

## Observability

- `/metrics` exposes `delivery_requests_total{status}`, `delivery_request_duration_seconds{status}`, `db_query_duration_seconds`.
- The compose stack starts Prometheus on `:9090` (scraping `app:8080`) and Grafana on `:3000`. Grafana ships with the Prometheus datasource and the `targeting engine` dashboard already provisioned (`deploy/grafana/`). Anonymous viewer access is on; admin login is `admin`/`admin`.
- Logs are JSON via `slog`.

## Numbers

In-process matcher benchmark (`go test -bench=. ./internal/cache`, M4 Pro):

| rules in cache | ns/op | allocs |
| ---: | ---: | ---: |
| 100 | 1.2 µs | 2 |
| 1 000 | 13 µs | 5 |
| 10 000 | 128 µs | 8 |
| 100 000 | 2.2 ms | 14 |

End-to-end HTTP (`go run ./cmd/bench -c 50 -d 10s` against the local Docker stack, 3 seeded campaigns):

```
rps      8428
latency  p50=2.03ms  p95=18.4ms  p99=35.3ms
```

The matcher is a linear walk over the snapshot, so cost grows with rule count. For real targeting volumes (10k+ rules) the next step is a per-dimension index inside the snapshot — but that only matters once the rule count justifies the complexity.

## What's missing / what i'd do next

- Multi-instance cache coordination — the LISTEN/NOTIFY model fans out fine, but two replicas reload independently, which is wasteful at scale. A delta channel or a versioned snapshot pull would fix it.
