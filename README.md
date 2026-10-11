# ticket-base-go

Boilerplate for Go services: **Connect + Vanguard** (one port serving Connect, gRPC, gRPC-Web and
REST from the proto's `google.api.http` annotations), **uber fx**, **pgx** with a unit of work,
**goose** migrations, **Redis** cache, **Kafka** events, and a **Helm** chart. A `user` example runs
through every layer; copy it for real entities, then delete it.

Module: `github.com/tanjed/base-go`.

## Start a new service from it

1. Copy the repo, then rename the module: `go mod edit -module github.com/<you>/<svc>` and replace
   `github.com/tanjed/base-go` in every `.go` file.
2. Rename the proto package (`proto/base/v1`, `package base.v1`, `go_package`), then `make generate`.
3. Rename the chart (`chart/Chart.yaml` `name`, `values.yaml` secret names, topic) and `APP_NAME`.
4. Add your migrations to `migrations/` (`0001_init.sql`, goose format); the service applies them
   at boot. There are none yet (the user example's table is in `internal/repository/user`).
5. Replace the `user` example (proto, `handler/user`, `service/user`, `repository/user`).

## Layout

| Path | Contents |
|---|---|
| `cmd/server` | Entry point: migrations, then serve |
| `proto/`, `api/` | API definitions; generated Go and OpenAPI (`make generate`, committed) |
| `internal/ioc` | Assembles the app from each package's `fx.Module` |
| `internal/handler/<name>` | Transport only: proto in, service call, proto out; declares its service with `router.Serve` |
| `internal/service/<name>` | Business rules; uses repositories through the unit of work |
| `internal/repository/<name>` | SQL only; owns its row types and its `Repo` interface |
| `internal/db`, `internal/db/uow` | Postgres pools (primary, optional read replica), migrations; the unit of work |
| `internal/apperror` | Every error artifact: application errors, Postgres mapping, the Connect interceptor, REST error bodies |
| `internal/router` | The listener: services collected from an fx group, Vanguard transcoder, Connect options |
| `internal/cache` | Redis client, `Cache` (JSON Get/Set/Delete with TTL), readiness probe |
| `internal/events` | Kafka publisher (`APP_KAFKA_BROKERS`), log-only without brokers |
| `internal/health`, `internal/handler/health` | Readiness probes (each dependency contributes one); `GET /live`, `GET /ready` |
| `internal/config`, `internal/logging` | `APP_*` environment; JSON logger |
| `migrations/` | goose SQL, embedded |
| `chart/`, `helmvars/` | Helm chart (Deployment, Service, Postgres); per-environment values |

## Configuration

| Variable | Meaning |
|---|---|
| `APP_NAME` | Service name: cache key prefix, events' source, default topic (`base`) |
| `APP_ADDR` | Listener (`:8080`) |
| `APP_DATABASE_URL` | Postgres primary: writes and migrations (required) |
| `APP_DATABASE_READ_URL` | Read replica for reads outside a unit of work (defaults to the primary; lag accepted) |
| `APP_REDIS_URL` | Redis (required) |
| `APP_KAFKA_BROKERS`, `APP_KAFKA_TOPIC` | Events (`<APP_NAME>.events`); no brokers: events are only logged |

## Commands

```sh
make tools generate   # buf + plugins into ./bin, regenerate api/
make test             # Go tests (db/uow starts a Postgres container: needs Docker)
make verify           # lint, test, helm template
make run              # docker compose: the service, Postgres, Redis on :8080
```

```sh
curl -s localhost:8080/ready
curl -s -X POST localhost:8080/v1/users -H 'content-type: application/json' -d '{"name":"Ada","email":"ada@example.com"}'
```
