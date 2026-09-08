# Application contract 1.0.1

The `web`, `worker`, and `admin` process types come from one compiled binary. `web` exposes `GET /health/live`, `GET /health/ready`, `POST /v1/orders`, and `GET /v1/orders/{orderID}`.

Order submissions require an `Idempotency-Key` containing 8 to 64 ASCII letters, digits, or hyphens. `customerId` uses 3 to 64 characters from the same set. `amountCents` is an integer from 1 through 99. Unknown JSON fields are rejected.

The valid shared fixtures are:

- `small_order`: ID `00000000-0000-4000-8000-000000000001`, key `order-submit-001`, customer `customer-001`, amount 59.
- `boundary_order`: ID `00000000-0000-4000-8000-000000000002`, key `order-submit-002`, customer `c-2`, amount 99.

Intentional invalid fixtures cover a missing `DATABASE_URL`, amount 0, amount 100, job schema version 2, migration target 999, and reuse of a key with a different amount.

Queue `orders.v1` carries `complete-order` jobs with `schemaVersion`, `orderId`, `idempotencyKey`, and `attemptedAt`. Schema version 1 is the only supported version. Asynq records two retries after the initial delivery as queue metadata rather than payload fields.

Order states are `accepted`, `processing`, `completed`, and `failed`. Submission creates the order and queue intent in one database transaction. A unique submission key maps to one order. The first valid submission returns HTTP 202, reuse with the same input returns the existing order with HTTP 200, and reuse with different input returns HTTP 409. Invalid order UUIDs return HTTP 400. The worker conditionally claims and completes an order, making completion an at-most-once database effect under repeated delivery.

Every process requires `DATABASE_URL` and `RELEASE_ID`. Web and worker require `REDIS_URL`. Optional configuration is `APP_HOST`, `PORT`, `WORKER_CONCURRENCY`, `SHUTDOWN_GRACE_MS`, and `TELEMETRY_MODE`. Ports range from 1024 through 65535. Telemetry modes are `console`, `memory`, and `disabled`. Web liveness reflects the process; readiness means required services and migration 001 are available and the process is accepting work. Readiness becomes false before shutdown drain.

JSON events go to stdout with `service`, `processType`, `releaseId`, and `event`. Work events add `orderId`; active traces add `traceId`. Configuration values are never logged.

`orders admin migrate --target 001` is the only supported migration command. It uses the same config, database, telemetry, and compiled artifact as web and worker.

Local and CI verification uses PostgreSQL 18, Redis 8, loopback HTTP, console telemetry, and Docker. It requires no cloud account, cloud credential, or production data.
