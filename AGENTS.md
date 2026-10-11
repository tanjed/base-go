# AGENTS.md: ticket-base-go

Boilerplate for Go services (module `github.com/tanjed/base-go`). See README.md.

## Commands

| Command | Purpose |
|---|---|
| `make tools` | Install pinned buf and protoc plugins into `./bin` |
| `make generate` | Regenerate `api/` from `proto/` (commit the result) |
| `make test` | Go tests (Postgres via testcontainers: needs Docker) |
| `make verify` | lint, test, template |
| `make run` | docker compose: service, Postgres, Redis |

## Stack

uber **fx**: each package owns `var Module = fx.Module(...)` and its constructor `New` in `module.go`; `internal/ioc` only lists the modules (in start order, `router` last). Every implementation has a compile-time assertion (`var _ Iface = (*Impl)(nil)`); constructors take interfaces, declared by the consumer (`user.Users`) or the repository (`userrepo.Repo`), bound in the consumer's module with a private provider (`func(s *usersvc.Service) Users { return s }`). **Connect + Vanguard** on one port (h2c), **pgx**, **goose**, **go-redis**, **franz-go**.

## Conventions (keep them)

1. **The API is defined in proto.** Add an endpoint in `proto/`, `make generate`, implement it in its handler package. No hand-written REST handlers. A handler module declares its service in one line, `router.Serve(basev1connect.NewXServiceHandler)`, with the implementation provided as the generated handler interface (`fx.As`); fx collects services in the `vanguard_services` group.
2. **Three layers.** Handlers (`internal/handler/<name>`) only translate proto; services (`internal/service/<name>`) hold the rules; repositories (`internal/repository/<name>`) hold the SQL and own their row types. Services import repositories, never the reverse.
3. **Unit of work.** Writes run in `uow.UnitOfWork(ctx, timeout, fn)`: one transaction on the primary, carried in the context; a nested call joins the outer one. Repositories query on `uow.Q(ctx)`: the transaction inside a unit of work, the **read replica** outside one (lag accepted). A write outside a unit of work fails (the replica is read-only, and so is the tests' read pool, `testdb.ReadOnly`).
4. **Errors are formatted in one place.** Repositories return `apperror.ErrNotFound` / `ErrDuplicate` (`apperror.FromPG`) or raw errors; services return `apperror.Fail(kind, code, msg)`; handlers return errors untouched. `apperror.Interceptor` maps them to Connect codes with the code as `ErrorInfo.reason` (anything else is logged and becomes `Internal`); `apperror.REST` shapes REST bodies as `{"error": reason, "message"}`. Don't build Connect errors anywhere else.
5. **After commit, best effort:** events (`events.Publisher`) and cache writes never fail a request; failures are logged.
6. **Migrations run at boot**, before anything is served; a migration must be safe while older replicas still serve (expand, then contract).
7. **The chart never creates secrets.**
