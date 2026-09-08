#!/bin/sh
set -eu

required='go.mod go.sum Dockerfile compose.yaml docs/application-contract.md internal/config/config.go internal/commands/web.go internal/commands/worker.go internal/commands/worker_test.go internal/commands/admin.go internal/domain/order_test.go internal/httpapi/api_test.go internal/migrations/contract_test.go internal/queue/queue_test.go internal/telemetry/log.go tests/integration/services_test.go .agents/skills/twelve-factor-app/SKILL.md .claude/skills/twelve-factor-app/SKILL.md'
for path in $required; do
  if [ ! -f "$path" ]; then
    echo "missing required file: $path" >&2
    exit 1
  fi
done

cmp .agents/skills/twelve-factor-app/SKILL.md .claude/skills/twelve-factor-app/SKILL.md
cmp .agents/skills/twelve-factor-app/references/factor-checklist.md .claude/skills/twelve-factor-app/references/factor-checklist.md
cmp db/migrations/001_orders.sql internal/migrations/001_orders.sql
./scripts/check-secrets.sh

repository_root=$(pwd)
export GOPATH="$repository_root/.cache/gopath"
export GOMODCACHE="$repository_root/.cache/gomod"
export GOCACHE="$repository_root/.cache/gobuild"
mkdir -p "$GOPATH" "$GOMODCACHE" "$GOCACHE" bin

go mod download
go mod verify
go vet ./...
go test -race ./...
CGO_ENABLED=0 go build -trimpath -o bin/orders ./cmd/orders
echo "Twelve-Factor verification passed."
