# Factor checklist

1. Factor I: Codebase, trace one application release to one revision and build target.
2. Factor II: Dependencies, declare, lock, isolate, and verify every dependency.
3. Factor III: Config, keep deploy values outside code and validate them safely.
4. Factor IV: Backing services, attach PostgreSQL and Redis through config and contracts.
5. Factor V: Build, release, run, promote one immutable artifact without runtime builds.
6. Factor VI: Processes, keep durable correctness outside replaceable process memory.
7. Factor VII: Port binding, let the web process own its configurable listener.
8. Factor VIII: Concurrency, scale process types independently and bound worker concurrency.
9. Factor IX: Disposability, start promptly and remove readiness before bounded drain.
10. Factor X: Dev/prod parity, test production-significant boundaries with faithful local services.
11. Factor XI: Logs, emit structured event streams with release and trace correlation.
12. Factor XII: Admin processes, run bounded migrations from the same release and config modules.

Cross-cutting checks: idempotent redelivery, transactional queue intent, finite retry, safe config errors, no credential leakage, clean local teardown, and deterministic failure evidence.
