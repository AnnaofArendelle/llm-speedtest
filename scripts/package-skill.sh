#!/usr/bin/env bash
# Validates the embedded Agent Skills package and writes build/llm-speedtest-skill.zip.
set -euo pipefail
cd "$(dirname "$0")/.."
mkdir -p build
go run ./cmd/llm-speedtest skill export build/llm-speedtest-skill.zip
ls -l build/llm-speedtest-skill.zip
