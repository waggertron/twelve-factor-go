---
name: twelve-factor-app
description: Review or implement Go service changes that affect configuration, backing services, build and release boundaries, process types, port binding, concurrency, graceful shutdown, dev/prod parity, structured logs, admin processes, or Twelve-Factor compliance. Use for architecture and operational behavior changes. Do not use for unrelated spelling, formatting, or isolated prose edits.
---

# Twelve-Factor App workflow

1. Read `AGENTS.md`, `docs/application-contract.md`, and `references/factor-checklist.md`.
2. Name the factors affected by the request before changing code.
3. Trace the change across source, configuration, database, queue, process commands, container, tests, and docs.
4. Reject hardcoded deploy values, realistic credentials, local-only production branches, and changes that split process types into different artifacts.
5. Preserve contract version 1.0.1, including `amountCents` from 1 through 99.
6. Prefer the existing chi, caarlos0/env, pgx, Asynq, OpenTelemetry, slog, Cobra, and Testcontainers boundaries.
7. Add or update unit and production-shaped integration evidence. An agent review alone is not proof.
8. Run `./scripts/check-twelve-factor.sh`. For container changes, also build the image and exercise Compose.
9. Run `./scripts/cleanup.sh`, then verify only intended files remain.

If the implementation already satisfies the request, report the evidence and make no change.
