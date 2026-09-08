# Agent harness evaluation

Evaluation date: 2026-09-08

The Codex skill passes the official `quick_validate.py` validator. The repository-owned verification command also confirms that the Codex and Claude workflow bodies and factor checklists remain byte-for-byte equivalent.

Agent review is advisory. `./scripts/check-twelve-factor.sh` is the completion gate because it requires no agent invocation and verifies module integrity, source formatting, static analysis, race-enabled tests, integration behavior, the compiled artifact, harness drift, and credential-shaped content.

## Deterministic evidence

- Unit tests accept `amountCents: 99` and reject `amountCents: 100` at the domain and HTTP boundaries.
- Integration tests use disposable PostgreSQL 18 and Redis 8 services with the production pgx and Asynq clients.
- Redelivery after completion leaves the persisted result unchanged.
- Removing `go.sum` makes the verification command exit nonzero with `missing required file: go.sum`.
- Docker Compose runs migration, web, and worker from the same image. The live workflow accepted 99, rejected 100 with HTTP 400, completed the job, and shut down web and worker with status 0.
- The cleanup command targets only the `tf_go` Compose project and repository-local generated paths.

These executable checks remain authoritative when an agent review is incomplete or incorrect.
