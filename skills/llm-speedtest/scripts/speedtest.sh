#!/usr/bin/env bash
# Run a benchmark stage and always produce machine-readable output.
# Usage: scripts/speedtest.sh <suite|scan|latency|cap|quality|load> --url … [flags]
# Writes report.md and result.json into $LLM_SPEEDTEST_OUT (default ./speedtest-out) and prints their paths.
set -euo pipefail
stage="${1:?stage required: suite|scan|latency|cap|quality|load}"
shift
out="${LLM_SPEEDTEST_OUT:-./speedtest-out}"
mkdir -p "$out"
stamp="$(date +%Y%m%d-%H%M%S)"
md="$out/$stage-$stamp.md"
json="$out/$stage-$stamp.json"
llm-speedtest "$stage" "$@" --out "$md" --json-out "$json"
echo "report=$md"
echo "json=$json"
