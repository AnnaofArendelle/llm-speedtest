# mock — LLM API Speed Test Report

- Generated: 2026-11-02 21:32:27
- Test window: 2026-11-02 13:32:27 ～ 0001-01-01 00:00:00
- Endpoint: `http://127.0.0.1:46183/v1`
- Scope: **0** models, 0 providers, grouped into **0** families
- Tool: llm-speedtest test

## 1. Summary

| Metric | Value |
|---|---|
| Models tested | 0 |
| Available (at least one success) | **0**（-） |
| Unavailable | 0（-） |
| …of which provider/account-level (credentials, auth, balance, quota, plan) | 0（-） |
| Families | 0（with an available member 0，entirely unavailable 0） |
| Stability of available models (A all rounds ok, B ≥ 2/3, C unstable) | A 0，B 0，C 0 |
| Stability of available families | A 0，B 0，C 0 |

Latency is the end-to-end time of a non-streaming `chat/completions` request with a tiny `max_tokens`, i.e. roughly time-to-first-token plus routing overhead as perceived by the client, not the provider's pure inference latency.

### 1.3 Failure reasons

| Reason | Models | Share | Level |
|---|---|---|---|

## 2. Method

1. Fetch the complete model list with `GET /models` (including `owned_by` and context metadata).
2. **Round 1**: one minimal `chat/completions` request per model, recording HTTP status, nested upstream status, error text, latency and the served model name; concurrency is bounded globally and per provider.
3. **Retry round**: transient failures (timeouts, rate limits, upstream 5xx, no channel/capacity, empty responses, network errors) are retried with lower concurrency and a longer timeout, honouring `reset after`/`retry-after` hints. Deterministic failures (auth, credentials, balance, missing model) are not retried.
4. **Resample rounds**: every model that succeeded is probed again; rankings use the **median** latency with min/max shown.
5. **Final round**: remaining transient failures get one last attempt before being declared unavailable.
6. Success means HTTP 200 with a `choices` array. Failures are classified status-code first, keywords second; the original message is always kept.
7. **Provider attribution** uses no built-in table: the field that best discriminates models and agrees with connection identifiers seen in errors (`owned_by` / id prefix) is chosen per endpoint; names can be edited in the UI.
8. **Family grouping** strips provider/org prefixes, `:free`-style suffixes, parameter sizes, date stamps, quantisation markers and qualifiers, keeps only the major version and the tier/speciality words that distinguish product lines; aliases are learned from served model names and the endpoint's own vocabulary, never from a built-in dictionary.
9. **Ranking**: stability tier first (A = every attempt succeeded, B = success rate ≥ 2/3, C = below), then median latency; families additionally get a composite score = stability × w + provider breadth × w + log-scaled speed × w.

Provider field for this endpoint: `endpoint`（no field discriminates between models; treating the endpoint as a single provider）

## 3. Providers

| Provider | Models | Families | Available | Rate | Median latency | Top failure reason | Diagnosis |
|---|---|---|---|---|---|---|---|

## 4. Family rankings

**4.1 Composite ranking**: from most providers / fastest / most stable to fewest / slowest / least stable.

| Component | Definition | Weight |
|---|---|---|
| Stability | success rate over all attempts of the family's available members × 100 | 0.40 |
| Provider breadth | providers with an available member ÷ the maximum across families × 100 | 0.30 |
| Speed | median latency on a log scale: ≤ 0.5 s scores 100, ≥ 30 s scores 0 | 0.30 |

Stability counts every attempt of members that succeeded at least once; provider/account-level failures say nothing about the model and are excluded from family stability but shown in the member counts and failure reasons.

### 4.1 Composite ranking

| # | Family | Score | Stab. | Breadth | Speed | Providers (stable) | Avail./total | Stability | Median | Min | Max | Stable members (model @ provider: median, ok/attempts) |
|---|---|---|---|---|---|---|---|---|---|---|---|---|

### 4.2 Families by stability tier and latency

| # | Family | Stability | Members | Available | Median | Available providers | Top failure reasons |
|---|---|---|---|---|---|---|---|

### 4.3 Entirely unavailable families

| Family | Members | Top reasons | Providers |
|---|---|---|---|

### 4.4 All available models (stability, then speed)

| # | Model | Provider | Family | Stability | Median | Min | Max | Jitter (max/min) | Served model | Notes |
|---|---|---|---|---|---|---|---|---|---|---|

## 5. Unavailable models

### 5.1 Failure categories

| Category | Meaning | Suggested action |
|---|---|---|

Note: messages are the upstream responses forwarded by the gateway with the `[connection/model] [status]:` wrapper removed. The HTTP column shows `endpoint / upstream`.

## 6. Family details (every model with its provider and results)

Same order as the composite ranking in 4.1: available families first, then entirely unavailable ones. Notes show the failure reason and message, or that an available model failed in some rounds.

## 7. Caveats

- Latency is the end-to-end time of small single requests with limited samples; use it for relative comparisons, not as a throughput benchmark. See the latency lab for tokens/s.
- All traffic passes through the local network and possibly proxies; absolute numbers vary with network, nodes and time of day, so relative order within one run is more meaningful.
- Many failures are **provider/account-level** (missing credentials, expired tokens, exhausted balance or quota, plan restrictions); the models themselves may work once the account is fixed.
- 429-class errors are split into momentary rate limits and exhausted quotas based on the error text, so a few may be misclassified; the original messages are in section 5.
- Family grouping is rule-based plus aliases learned from this endpoint; different providers may serve different backends under the same name (see the served model column).
