#!/usr/bin/env bash
# Runs the Playwright end-to-end suite against a freshly started demo server.
# Requires: bin/llm-speedtest (make build) and web dependencies (npm ci).
set -euo pipefail
cd "$(dirname "$0")/.."

PORT="${LST_E2E_PORT:-8790}"
DATA="$(mktemp -d)"
./bin/llm-speedtest serve --demo --listen "127.0.0.1:$PORT" --data-dir "$DATA" > "$DATA/serve.log" 2>&1 &
PID=$!
cleanup() {
  kill "$PID" 2>/dev/null || true
  rm -rf "$DATA"
}
trap cleanup EXIT
for _ in $(seq 1 60); do
  curl -sf "http://127.0.0.1:$PORT/healthz" >/dev/null && break
  sleep 1
done
curl -sf "http://127.0.0.1:$PORT/healthz" >/dev/null || { echo "demo server did not start:"; cat "$DATA/serve.log"; exit 1; }
(cd web && LST_BASE_URL="http://127.0.0.1:$PORT" npx playwright test -c tests/playwright.e2e.config.ts)
