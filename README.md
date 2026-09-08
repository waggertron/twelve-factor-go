# Twelve-Factor Go reference app

A runnable order service that applies the Twelve-Factor App method with current Go libraries and executable evidence. The same `orders` binary provides `web`, `worker`, and `admin` process types. The application contract is version 1.0.1.

## Stack

- chi 5.3.2 for HTTP routing
- caarlos0/env 11.4.1 for configuration
- pgx 5.11.0 for PostgreSQL
- Asynq 0.26.0 for Redis-backed jobs
- OpenTelemetry 1.46.0 and `slog` for correlated JSON events
- Cobra 1.10.2 for process commands
- Testcontainers for Go 0.44.0 for PostgreSQL 18 and Redis 8 integration tests

## Run locally

Requirements are Go 1.25 or newer and Docker.

```bash
docker compose -p tf_go --profile admin run --rm migrate
docker compose -p tf_go up --build web worker
```

In another terminal:

```bash
curl http://localhost:3103/health/ready
curl -H 'Content-Type: application/json' \
  -H 'Idempotency-Key: sample-order-1' \
  -d '{"customerId":"customer-001","amountCents":59}' \
  http://localhost:3103/v1/orders
```

`amountCents` is an integer from 1 through 99. An amount of 100 is rejected with HTTP 400.

## Commands

```bash
go mod download
go test ./internal/...
go test ./tests/integration/...
./scripts/check-twelve-factor.sh
docker build -t twelve-factor-go .
./scripts/cleanup.sh
```

The verification script uses repository-local Go caches, checks the migration copies and agent harnesses for drift, scans source for credential-shaped content, runs `go vet`, runs all tests with the race detector, and builds the binary.

## Process types

- `orders web` serves HTTP and removes readiness before bounded shutdown.
- `orders worker` processes `orders.v1.complete` jobs with bounded concurrency, stops intake, then drains active work.
- `orders admin migrate --target 001` runs the packaged migration through the normal config and database modules.

Required deployment configuration is `DATABASE_URL` for every process and `REDIS_URL` for web and worker. Optional values are `PORT`, `WORKER_CONCURRENCY`, `SHUTDOWN_GRACE_MS`, `RELEASE_ID`, and `TELEMETRY_MODE`. Invalid configuration produces categories without echoing values.

## Architecture and contract

The API writes the order and a queue intent in one PostgreSQL transaction, then enqueues a versioned job. A unique idempotency key prevents duplicate orders. The worker conditionally moves `accepted` to `processing` and `processing` to `completed`, so repeated delivery cannot repeat the completion effect. See [the complete application contract](docs/application-contract.md).

## The twelve factors

1. Codebase: one repository and one tracked codebase.
2. Dependencies: Go modules and committed `go.sum` declare the full graph.
3. Config: caarlos0/env reads deploy-varying values from the environment.
4. Backing services: PostgreSQL and Redis are attached by URLs.
5. Build, release, run: one compiled binary is built once and configured at run time.
6. Processes: web and worker keep durable state in backing services.
7. Port binding: web exports HTTP on `PORT`.
8. Concurrency: separate web and worker processes scale independently; worker concurrency is bounded.
9. Disposability: readiness drops before bounded HTTP or job drain.
10. Dev/prod parity: local, CI, and production-shaped tests use PostgreSQL and Redis through the same clients.
11. Logs: JSON events go to stdout with service, process, release, and trace context.
12. Admin processes: migrations run from the same binary and configuration as the app.

## Agent harnesses

Codex and Claude project skills in `.agents/skills/twelve-factor-app` and `.claude/skills/twelve-factor-app` turn these rules into a review workflow. They guide changes but cannot replace the deterministic verification command. Evaluation evidence is in [docs/agent-evaluation.md](docs/agent-evaluation.md).

## Cleanup

`./scripts/cleanup.sh` removes only the `tf_go` Compose project, its volumes, and repository-local caches and binaries. It does not prune global Docker or Go state.

## Contributing

Read [CONTRIBUTING.md](CONTRIBUTING.md) before changing the contract, dependencies, or process lifecycle.
