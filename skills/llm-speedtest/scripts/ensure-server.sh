#!/usr/bin/env bash
# Start the llm-speedtest dashboard in the background if it is not already running.
# Usage: scripts/ensure-server.sh [listen-address] [data-dir]
set -euo pipefail
LISTEN="${1:-127.0.0.1:8787}"
DATA_DIR="${2:-${LLM_SPEEDTEST_DATA:-./data}}"
if curl -fsS "http://$LISTEN/healthz" >/dev/null 2>&1; then
  echo "llm-speedtest already running at http://$LISTEN"
  exit 0
fi
mkdir -p "$DATA_DIR"
nohup llm-speedtest serve --listen "$LISTEN" --data-dir "$DATA_DIR" >"$DATA_DIR/serve.log" 2>&1 &
for _ in $(seq 1 30); do
  if curl -fsS "http://$LISTEN/healthz" >/dev/null 2>&1; then
    echo "llm-speedtest started at http://$LISTEN (log: $DATA_DIR/serve.log)"
    exit 0
  fi
  sleep 0.5
done
echo "server did not become healthy; see $DATA_DIR/serve.log" >&2
exit 1
