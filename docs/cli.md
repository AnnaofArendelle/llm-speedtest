# CLI reference

All commands accept `-h`. Durations use Go syntax (`30s`, `2m`). Exit codes: `0` ok, `1` error, `2` usage, `3` endpoint unreachable / auth failed / discovery failed, `4` cancelled, `5` `--fail-under` threshold not met.

## Endpoint flags (shared)

| Flag | Meaning |
|---|---|
| `--url` | base URL, e.g. `https://host/v1` (OpenRouter: `https://openrouter.ai/api/v1`) |
| `--key` | API key; falls back to `$LLM_SPEEDTEST_KEY` |
| `--name` | display name used in reports |
| `--dialect` | `9router`, `oneapi`, `openrouter`, `litellm`, `ollama`, `vllm`, `generic`; auto-detected when omitted |
| `--header k=v` | extra header (repeatable) |
| `--insecure` | skip TLS verification |
| `--proxy URL` | explicit proxy (defaults to `HTTPS_PROXY`/`HTTP_PROXY`) |
| `--config file.yaml` | load endpoint and suite settings from YAML (`llm-speedtest config example`) |
| `--data-dir` | results database directory (default `./data` or `$LLM_SPEEDTEST_DATA`) |
| `--no-store` | temporary database, nothing kept |

## Filter flags (scan / suite / discover)

`--include-owner`, `--exclude-owner` (default `combo`), `--id-regex`, `--exclude-regex`, `--limit N`, `--model id` (repeatable; explicit list).

## Output flags

`--out report.md`, `--json-out result.json`, `--csv-out models.csv`, `--lang zh|en`, `--note text`, `--quiet`, `--verbose`.

## Commands

### `check`
Connectivity, authentication, model count, dialect, latency of `/models`. `--json` for machine output.

### `discover`
Lists models after filtering with resolved providers and family count. `--json` returns `models[]` with `provider_key`, `provider_name`, `family_key`, `family_canon`, `family_source`.

### `scan`
Availability scan. `--profile quick|standard|deep` sets rounds (1 / 5 / 7). `--fail-under 0.5` exits 5 when fewer than 50 % of models are available (CI gate).

### `latency`
`--model id` (repeatable), `--samples N`, `--warmup N`, `--input-tokens N`, `--output-tokens N`.

### `load`
One model: `--model id`, `--mode concurrency|rate`, `--levels 1,2,4,8`, `--rate 2` (req/s), `--duration 30s`, `--requests N` (cap per level), `--no-stream`. **Requires `--confirm`.**

### `cap`
`--model id` (repeatable), `--long-context 8000,32000,128000` (tiers in tokens; 128k is expensive).

### `quality`
`--model id` (repeatable), `--trials N`.

### `suite`
Scan → latency → capability → quality on the top models of the scan (`--top N`), optional `--load [--load-model id]` (implies confirm). Disable stages with `--no-latency`, `--no-capability`, `--no-quality`.

### `report`
`--run id` (default latest), `--format md|json|csv`, `--lang`, `--out`.

### `compare`
`--a id --b id [--json]`: models that changed status and headline deltas.

### `runs`
List stored runs (`--limit`, `--json`).

### `serve`
`--listen 127.0.0.1:8787`, `--access-token TOKEN` (required for non-localhost exposure), `--demo` (seed demo data).

### `mcp`
Model Context Protocol server over stdio; `--data-dir` to share results with `serve`.

### `skill`
`skill install --target claude|codex|cursor|<dir> [--force]`, `skill export [file.zip]`, `skill show`.
