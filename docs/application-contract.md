# Application contract 1.0.1

The `web`, `worker`, and `admin` process types come from one compiled binary. `web` exposes `GET /health/live`, `GET /health/ready`, `POST /v1/orders`, and `GET /v1/orders/{orderID}`.

Order submissions require an `Idempotency-Key` containing 8 to 64 ASCII letters, digits, or hyphens. `customerId` uses 3 to 64 characters from the same set. `amountCents` is an integer from 1 through 99. Unknown JSON fields are rejected.

The valid shared fixtures are:

- `small_order`: key `fixture-small-01`, customer `customer-001`, amount 59.
- `boundary_order`: key `fixture-boundary-99`, customer `customer-099`, amount 99.

Intentional invalid fixtures cover a missing `DATABASE_URL`, amount 0, amount 100, job schema version 2, migration target 999, and reuse of a key with a different amount.

Jobs use task type `orders.v1.complete` and payload `{ "schemaVersion": 1, "orderId": "<uuid>", "attempt": 1 }`. Attempts are integers from 1 through 3. Asynq is configured for two retries after the initial delivery.

Order states are `accepted`, `processing`, `completed`, and `failed`. Submission creates the order and queue intent in one database transaction. A unique submission key maps to one order. Reuse with the same input returns the existing order; reuse with different input returns HTTP 409. The worker conditionally claims and completes an order, making completion an at-most-once database effect under repeated delivery.

Every process requires PostgreSQL. Web and worker require Redis. Web liveness reflects the process; readiness means required services and migration 001 are available and the process is accepting work. Readiness becomes false before shutdown drain.

JSON events go to stdout with `service`, `processType`, `releaseId`, and `event`. Work events add `orderId`; active traces add `traceId`. Configuration values are never logged.

`orders admin migrate --target 001` is the only supported migration command. It uses the same config, database, telemetry, and compiled artifact as web and worker.

Local and CI verification uses PostgreSQL 18, Redis 8, loopback HTTP, console telemetry, and Docker. It requires no cloud account, cloud credential, or production data.
