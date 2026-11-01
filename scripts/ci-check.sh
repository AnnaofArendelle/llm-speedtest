#!/usr/bin/env bash
# Runs the same checks as the CI pipeline on a developer machine.
# Tools that are not installed are reported and skipped.
set -euo pipefail
cd "$(dirname "$0")/.."

step() { printf '\n\033[1;36m==> %s\033[0m\n' "$*"; }
have() { command -v "$1" >/dev/null 2>&1; }

step "web: install, lint, format, typecheck, test, build"
(cd web && npm ci --no-audit --no-fund && npm run lint && npm run format:check && npm run typecheck && npm test && npm run build)

step "go: vet"
go vet ./...

if have golangci-lint; then
  step "go: golangci-lint"
  golangci-lint run ./...
else
  echo "golangci-lint not installed — skipped (go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.15.0)"
fi

step "go: tests with race detector and coverage gate"
./scripts/coverage.sh

if have govulncheck; then
  step "go: govulncheck"
  govulncheck ./...
else
  echo "govulncheck not installed — skipped (go install golang.org/x/vuln/cmd/govulncheck@latest)"
fi

if have shellcheck; then
  step "shellcheck"
  shellcheck -S warning scripts/*.sh skills/llm-speedtest/scripts/*.sh
else
  echo "shellcheck not installed — skipped"
fi

if have actionlint; then
  step "actionlint"
  actionlint
else
  echo "actionlint not installed — skipped (go install github.com/rhysd/actionlint/cmd/actionlint@latest)"
fi

step "skill package"
./scripts/package-skill.sh

if have goreleaser; then
  step "goreleaser check"
  goreleaser check
fi

step "all checks passed"
