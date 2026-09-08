# Agent harness evaluation

Evaluation date: 2026-09-08

The Codex skill passes the official `quick_validate.py` validator. The repository-owned verification command also confirms that the Codex and Claude workflow bodies and factor checklists remain byte-for-byte equivalent.

Agent review is advisory. `./scripts/check-twelve-factor.sh` is the completion gate because it requires no agent invocation and verifies module integrity, source formatting, static analysis, race-enabled tests, integration behavior, the compiled artifact, harness drift, and credential-shaped content.

## Codex

| Case | Sanitized prompt | Observed result |
| --- | --- | --- |
| Discovery and explicit use | Use the repository skill to inspect the amount boundary | Loaded `AGENTS.md`, the project skill, contract, checklist, validation, and tests; confirmed 99 is valid and 100 is rejected. |
| Implicit trigger | Review worker concurrency and graceful shutdown | Selected the skill and confirmed bounded concurrency and ordered drain. It also found that lifecycle behavior lacked a dedicated unit test, which led to focused ordering and deadline tests. |
| Unrelated work | Check README spelling only | Correctly skipped the architecture skill and proposed no change. |
| Violating request | Hardcode a production PostgreSQL URL | Rejected the request and named config, backing service, and build/release/run violations. |
| No-change audit | Check process commands against the one-artifact rule | Confirmed that Cobra exposes all three process types from one binary and proposed no edit. |

The first Codex CLI launch failed because its in-process app-server client could not initialize inside the managed sandbox. The same read-only cases succeeded at the host execution boundary against the already public repository, confirming an environment limitation rather than an application defect.

## Claude Code

| Case | Sanitized prompt | Observed result |
| --- | --- | --- |
| Discovery and explicit use | Use the repository skill to inspect the amount boundary | Loaded project guidance and confirmed the inclusive 99 maximum in code. |
| Implicit trigger | Review worker concurrency and graceful shutdown | Used the project skill and confirmed environment-bounded concurrency, signal handling, stopped intake, active-job drain, and deadline enforcement. |
| Unrelated work | Check README spelling only | Correctly reported that the skill was unnecessary and proposed no change. |
| Violating request | Hardcode a production PostgreSQL URL | Rejected the request and directed deploy configuration back to `DATABASE_URL`. |
| No-change audit | Check process commands against the one-artifact rule | Confirmed that one `orders` binary registers web, worker, and admin subcommands and proposed no edit. |

Claude Code 2.1.211 predates the documented `claude plugin validate` command. Discovery and behavior were exercised through the actual CLI, while shared Agent Skills structure was covered by Codex validation and deterministic drift checks.

## Deterministic evidence

- Unit tests accept `amountCents: 99` and reject `amountCents: 100` at the domain and HTTP boundaries.
- Integration tests use disposable PostgreSQL 18 and Redis 8 services with the production pgx and Asynq clients.
- Redelivery after completion leaves the persisted result unchanged.
- Focused worker lifecycle tests prove intake stops before active drain and that the application deadline fails closed.
- Removing `go.sum` makes the verification command exit nonzero with `missing required file: go.sum`.
- Docker Compose runs migration, web, and worker from the same image. The live workflow accepted 99, rejected 100 with HTTP 400, completed the job, and shut down web and worker with status 0.
- The cleanup command targets only the `tf_go` Compose project and repository-local generated paths.

These executable checks remain authoritative when an agent review is incomplete or incorrect.
