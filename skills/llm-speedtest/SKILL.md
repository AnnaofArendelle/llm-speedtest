---
name: llm-speedtest
description: Benchmark any OpenAI-compatible LLM API endpoint (gateways, routers, self-hosted servers, cloud providers) for availability, latency (TTFT, tokens/s), load capacity, capabilities and answer quality, then rank models and families and write a report. Use when the user asks which models work on an endpoint, how fast or stable they are, to compare providers, to measure TTFT or throughput, to run a load test, or to produce a speed-test report or dashboard.
compatibility: Requires the llm-speedtest CLI (single static binary for Linux, macOS and Windows; scripts/install.sh fetches it) and network access to the endpoint under test.
metadata:
  version: "0.1.0"
  homepage: https://github.com/AnnaofArendelle/llm-speedtest
  license: AGPL-3.0-only
---

# LLM Speedtest

`llm-speedtest` turns "does this endpoint work and how fast is it?" into a repeatable benchmark:

1. **Scan** — discovers every model behind `/models`, probes each one with a minimal chat completion in up to five rounds (full pass → retry of transient failures → re-sampling of successes → final confirmation), classifies every failure into 20+ reasons (auth, balance, quota, rate limit, model missing, non-chat model, upstream error, timeout …), groups model ids into families (`deepseek-v4.1-flash-1017:free` → `deepseek-v4-flash`), attributes models to providers dynamically from the endpoint's own data, and ranks families by a composite of stability, provider breadth and speed.
2. **Latency lab** — streaming TTFT, time-to-second-token, inter-token latency, time per output token, end-to-end latency and tokens/s with percentiles.
3. **Load** — closed-loop concurrency sweeps or open-loop Poisson request rates with throughput, error rate, latency percentiles and goodput under an SLO. **Costs money on paid endpoints; always confirm with the user first.**
4. **Capabilities** — streaming, usage reporting, tool calling, JSON mode, structured output, vision, system-prompt adherence, max_tokens/stop handling, determinism, reasoning parameters, long context, embeddings.
5. **Quality** — a deterministic 20-question spot check (no judge model) that catches broken or mislabelled deployments.

Everything is stored in a local SQLite database and can be rendered as a Markdown report, JSON or CSV, or browsed in the web dashboard (`llm-speedtest serve`).

## Prerequisites

- Check the binary: `llm-speedtest version`. If missing, run `bash scripts/install.sh` (downloads the latest GitHub release for the current OS/arch and verifies the checksum) or ask the user to install it.
- The API key must come from the user or the environment (`LLM_SPEEDTEST_KEY`). Never paste keys into files or reports; the tool never prints them.
- Results live in `./data` by default (override with `--data-dir` or `LLM_SPEEDTEST_DATA`). Use the same directory for every command so runs stay comparable.

## Workflow

### 1. Verify the endpoint

```bash
llm-speedtest check --url https://host/v1 --key "$LLM_SPEEDTEST_KEY" --json
```

Read `reachable`, `auth_ok`, `models`, `dialect`. A `suggested_base_url` means the user gave the origin without `/v1`; use the suggestion.

### 2. Preview what would be tested

```bash
llm-speedtest discover --url https://host/v1 --json | head -c 4000
```

Tell the user how many models and providers were found and which field identifies the provider (`provider_field` + `provider_rationale`). Apply filters when the list is large or the user cares about a subset: `--include-owner groq --include-owner cf`, `--id-regex '^openai/'`, `--limit 50`, or explicit `--model id` flags.

### 3. Run the benchmark

Full suite (recommended default; `quick` ≈ minutes, `standard` ≈ tens of minutes for thousands of models, `deep` longest):

```bash
llm-speedtest suite --url https://host/v1 --profile quick \
  --out report.md --json-out result.json --lang en
```

Targeted stages when the user asks for one thing:

```bash
llm-speedtest scan    --url … --profile standard --out scan.md
llm-speedtest latency --url … --model a/x --model b/y --samples 5 --out latency.md
llm-speedtest cap     --url … --model a/x --out capabilities.md
llm-speedtest quality --url … --model a/x --trials 3 --out quality.md
llm-speedtest load    --url … --model a/x --levels 1,2,4,8 --duration 30s --confirm --out load.md
```

Rules of thumb:
- Never run `load` (or `suite --load`) without explicit user consent; it needs `--confirm`.
- Prefer `--profile quick` for a first look, `standard` for a report the user will act on.
- Long runs stream progress on stderr; summarise it, don't echo every line.
- The command prints a JSON summary on stdout when no `--out/--json-out` is given.

### 4. Read the results

- `result.json` (`schema_version` 1) contains `summary`, `families` (composite order), `families_by_stability`, `unavailable_families`, `providers`, `models`.
- Stability tiers: **A** = every attempt succeeded, **B** = success rate ≥ 2/3, **C** = below.
- Composite score = 0.4·stability + 0.3·provider breadth + 0.3·speed (log scale, 0.5 s → 100, 30 s → 0).
- Provider/account-level reasons (`no_credentials`, `auth`, `quota_billing`, `quota_exhausted`, `plan_restriction`, `terms_required`) mean the *account* is broken, not the model; say so and suggest fixing the account before re-testing.
- See `references/metrics.md` for definitions and `references/interpreting-results.md` for how to phrase recommendations.

### 5. Report back

Use `assets/report-template.md` for the summary you give the user: headline numbers, top 5 families with the best member and provider, what is broken and why, and the path of the generated report. Offer the dashboard: `llm-speedtest serve` then open http://127.0.0.1:8787.

## MCP alternative

If the agent runtime supports Model Context Protocol servers, `llm-speedtest mcp` exposes the same functionality as tools (`check_endpoint`, `discover_models`, `start_run`, `wait_run`, `get_ranking`, `get_failures`, `export_report`, …) and shares the data directory with `serve`. Example for Claude Code:

```bash
claude mcp add llm-speedtest -- llm-speedtest mcp --data-dir ./data
```

## Files

- `scripts/install.sh`, `scripts/install.ps1` — install the binary from GitHub releases.
- `scripts/ensure-server.sh` — start the dashboard in the background if it is not running.
- `scripts/speedtest.sh` — thin wrapper that always adds `--json-out` and prints the result path.
- `references/cli.md` — every command and flag.
- `references/api.md` — HTTP API and MCP tools.
- `references/metrics.md` — metric definitions.
- `references/interpreting-results.md` — how to turn numbers into recommendations.
- `assets/report-template.md` — summary template.
