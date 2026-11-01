#!/usr/bin/env bash
# Runs the Go tests with the race detector and enforces a coverage threshold.
# Test utilities, embedded data packages and the thin main package are excluded
# from the gate; everything else must stay above COVERAGE_THRESHOLD (default 75%).
set -euo pipefail
cd "$(dirname "$0")/.."

THRESHOLD="${COVERAGE_THRESHOLD:-75}"
mapfile -t PKGS < <(go list ./... | grep -v -e '/web/node_modules/' -e '/internal/testutil/' -e '/internal/refdata$' -e '/cmd/')

go test -race -count=1 -covermode=atomic -coverprofile=coverage.out "${PKGS[@]}"

total=$(go tool cover -func=coverage.out | awk '/^total:/ { gsub("%", "", $3); print $3 }')
echo "total coverage: ${total}% (threshold ${THRESHOLD}%)"
if ! awk -v t="$total" -v th="$THRESHOLD" 'BEGIN { exit (t + 0 >= th + 0) ? 0 : 1 }'; then
  echo "coverage ${total}% is below the ${THRESHOLD}% threshold" >&2
  exit 1
fi
