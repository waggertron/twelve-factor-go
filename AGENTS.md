# Repository instructions

Use the `twelve-factor-app` skill for architecture, configuration, backing service, process, lifecycle, telemetry, migration, container, or verification changes.

- Preserve application contract version 1.0.1 and the `amountCents` range of 1 through 99.
- Keep web, worker, and admin in the one `orders` artifact.
- Keep deploy-varying configuration in environment variables and never log values.
- Never add realistic credential strings or production endpoints. Use `YOUR_VALUE_HERE` or `<token>`.
- Prefer PostgreSQL 18, Redis 8, loopback HTTP, console telemetry, and strict fakes for local tests.
- Run `./scripts/check-twelve-factor.sh` after changes and `./scripts/cleanup.sh` after local service tests.
