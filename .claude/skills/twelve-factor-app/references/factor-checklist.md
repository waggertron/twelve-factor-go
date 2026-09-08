# Factor checklist

1. Codebase: one repository maps to one application.
2. Dependencies: `go.mod` and `go.sum` are complete and frozen.
3. Config: deploy-varying values come from validated environment input.
4. Backing services: PostgreSQL and Redis are replaceable URL attachments.
5. Build, release, run: one immutable binary is built once and configured at run time.
6. Processes: durable state stays outside web and worker memory.
7. Port binding: web exports HTTP through its own listener on `PORT`.
8. Concurrency: process types scale independently and worker concurrency is bounded.
9. Disposability: readiness drops, intake stops, active work drains, and deadlines are enforced.
10. Dev/prod parity: tests use the same protocols and clients as deployment.
11. Logs: structured events go to stdout with release and trace correlation.
12. Admin processes: migrations run once from the same artifact and config modules.

Cross-cutting checks: idempotent redelivery, transactional queue intent, finite retry, safe config errors, no credential leakage, clean local teardown, and deterministic failure evidence.
