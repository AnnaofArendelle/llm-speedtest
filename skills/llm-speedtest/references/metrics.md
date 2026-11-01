# Metric definitions

## Availability scan

- **Attempt** — one non-streaming `chat/completions` request with `max_tokens` = 1. Success = HTTP 200 with a `choices` array.
- **Rounds** — 1 full pass; 2 retry of *transient* failures (timeout, rate limit, upstream 5xx, no channel/capacity, empty response, network) with lower concurrency; 3…N re-sampling of every model that succeeded; final confirmation of remaining transient failures. Deterministic failures (auth, credentials, balance, plan, missing/retired model) are never retried.
- **Latency (scan)** — end-to-end time of the attempt ≈ time-to-first-token plus gateway overhead. Models use the **median** of their successful attempts; min/max and jitter (max/min) are shown.
- **Stability tier (model)** — A: every attempt succeeded; B: success rate ≥ 2/3; C: below. Rate = successes / attempts of that model.
- **Family** — models grouped by normalised name: provider/org prefix and `:free`-style suffixes removed; parameter sizes (7b, a3b), dates, quantisation markers and qualifiers (preview, instruct, thinking, fast, …) dropped; only the major version kept (v4.1 → v4); tier/speciality words kept (flash, pro, opus, mini, vl, coder …). Aliases are learned from the model name the upstream reports in successful responses and from abbreviations that match the endpoint's own vocabulary; users can override.
- **Family stability** — successes / attempts over the family's *available* members only (provider-level failures excluded). Tiers: A ≥ 95 %, B ≥ 2/3, C below.
- **Composite score** — `0.4 × stability% + 0.3 × breadth + 0.3 × speed` where breadth = providers with an available member ÷ the maximum across families × 100 and speed maps the family median latency on a log scale (≤ 0.5 s → 100, ≥ 30 s → 0).
- **Provider** — chosen per endpoint from the field that best separates models and agrees with connection identifiers seen in error messages (`owned_by` → id prefix → single provider). No built-in name table.

## Latency lab (streaming)

- **TTFT** — first generated token minus request sent.
- **TTST** — second token minus first token.
- **TPOT** — (done − first token) / (N − 1), N = output tokens.
- **ITL** — gaps between streamed chunks, apportioned by tokens per chunk.
- **E2E** — done minus sent.
- **tok/s** — (N − 1) / (done − first token). Token counts come from `usage` when the server reports it in the stream (`stream_options.include_usage`); otherwise they are estimated (CJK ≈ 1 token/char, else 4 chars/token) and flagged.
- **Cold / warm** — the first (warm-up) sample is reported separately; statistics use warm samples.
- Percentiles: p50 (median), p75, p90, p95, p99, min, max, mean.

## Load test

- **Closed loop** — N concurrent workers issue requests back-to-back for the duration; levels form a sweep.
- **Open loop** — Poisson arrivals at the requested rate; `queue_wait` records dispatcher delay when the in-flight cap is hit.
- **req/s** — successful requests / wall time; **out tok/s** — output tokens / wall time.
- **Goodput** — share of successful requests meeting every configured SLO (TTFT, TPOT, E2E).

## Capability probes

Each probe sends one targeted request and reports `supported`, `unsupported` (the endpoint rejected or ignored the feature) or `error` (unrelated failure), with evidence.

## Quality spot-check

20 deterministic questions (arithmetic, string manipulation, extraction, instruction following, knowledge) checked by exact/number/regex matching; score = median over trials of the weighted pass rate (0–100); consistency = share of questions answered identically across trials; identity = the model's own answer to "which model are you" (informational only).
