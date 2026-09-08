# Contributing

Keep all three process types in the one `orders` artifact, preserve contract version 1.0.1, and keep `amountCents` between 1 and 99. Add focused unit coverage and production-shaped integration coverage for behavior changes.

Run `./scripts/check-twelve-factor.sh` and `docker build -t twelve-factor-go .` before committing. Run `./scripts/cleanup.sh` when local services are no longer needed. Examples must use placeholders and local-only values, never realistic credentials or production endpoints.
