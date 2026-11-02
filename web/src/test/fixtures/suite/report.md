# mock — LLM API Speed Test Report

- Generated: 2026-11-02 21:32:27
- Test window: 2026-11-02 13:32:27 ～ 2026-11-02 13:32:27
- Endpoint: `http://127.0.0.1:46183/v1`
- Scope: **3** models, 2 providers, grouped into **3** families
- Rounds: round 1: 3 requests (2 ok)
- Tool: llm-speedtest test

## 1. Summary

| Metric | Value |
|---|---|
| Models tested | 3 |
| Available (at least one success) | **2**（66.7%） |
| Unavailable | 1（33.3%） |
| …of which provider/account-level (credentials, auth, balance, quota, plan) | 0（0.0%） |
| Families | 3（with an available member 2，entirely unavailable 1） |
| Stability of available models (A all rounds ok, B ≥ 2/3, C unstable) | A 2，B 0，C 0 |
| Stability of available families | A 2，B 0，C 0 |
| Latency median / P90 / min / max | 0.03s / 0.04s / 0.01s / 0.04s |

Latency is the end-to-end time of a non-streaming `chat/completions` request with a tiny `max_tokens`, i.e. roughly time-to-first-token plus routing overhead as perceived by the client, not the provider's pure inference latency.

### 1.1 Top 20 most stable and fastest models

| # | Model | Provider | Family | Stability (ok/attempts) | Median | Min | Max |
|---|---|---|---|---|---|---|---|
| 1 | `a/fast` | a | fast | A stable（1/1） | 0.01s | 0.01s | 0.01s |
| 2 | `a/slow` | a | slow | A stable（1/1） | 0.04s | 0.04s | 0.04s |

### 1.2 Latency distribution of available models

| Median latency | Models | Share |
|---|---|---|
| < 1 s | 2 | 100.0% |
| 1 – 2 s | 0 | 0.0% |
| 2 – 5 s | 0 | 0.0% |
| 5 – 10 s | 0 | 0.0% |
| 10 – 30 s | 0 | 0.0% |
| ≥ 30 s | 0 | 0.0% |

### 1.3 Failure reasons

| Reason | Models | Share | Level |
|---|---|---|---|
| Model not found / unavailable (404/400) | 1 | 100.0% | model / network |

### 1.4 Top 15 families by composite score

Composite = stability × 0.4 + provider breadth × 0.3 + speed × 0.3; see section 4.

| # | Family | Score | Providers | Available/total | Stability | Median | Best member @ provider |
|---|---|---|---|---|---|---|---|
| 1 | **fast** | 100.0 | 1（stable 1） | 1/1 | A stable（1/1） | 0.01s | `a/fast` @ a：0.01s |
| 2 | **slow** | 100.0 | 1（stable 1） | 1/1 | A stable（1/1） | 0.04s | `a/slow` @ a：0.04s |

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

Provider field for this endpoint: `owned_by`（owned_by discriminates between models (2 distinct values)）

## 3. Providers

| Provider | Models | Families | Available | Rate | Median latency | Top failure reason | Diagnosis |
|---|---|---|---|---|---|---|---|
| a | 2 | 2 | 2 | 100.0% | 0.03s | - | all available |
| b | 1 | 1 | 0 | 0.0% | - | Model not found / unavailable (404/400) | Model not found / unavailable (404/400): 1/1 failed models |

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
| 1 | **fast** | **100.0** | 100 | 100 | 100 | 1（1） | 1/1 | A stable（1/1） | 0.01s | 0.01s | 0.01s | `a/fast` @ a：0.01s（1/1） |
| 2 | **slow** | **100.0** | 100 | 100 | 100 | 1（1） | 1/1 | A stable（1/1） | 0.04s | 0.04s | 0.04s | `a/slow` @ a：0.04s（1/1） |

### 4.2 Families by stability tier and latency

| # | Family | Stability | Members | Available | Median | Available providers | Top failure reasons |
|---|---|---|---|---|---|---|---|
| 1 | **fast** | A stable（1/1） | 1 | 1 | 0.01s | a(1) | - |
| 2 | **slow** | A stable（1/1） | 1 | 1 | 0.04s | a(1) | - |

### 4.3 Entirely unavailable families

| Family | Members | Top reasons | Providers |
|---|---|---|---|
| **missing** | 1 | Model not found / unavailable (404/400)(1) | b(1) |

### 4.4 All available models (stability, then speed)

| # | Model | Provider | Family | Stability | Median | Min | Max | Jitter (max/min) | Served model | Notes |
|---|---|---|---|---|---|---|---|---|---|---|
| 1 | `a/fast` | a | fast | A stable（1/1） | 0.01s | 0.01s | 0.01s | 1.0× | fast |  |
| 2 | `a/slow` | a | slow | A stable（1/1） | 0.04s | 0.04s | 0.04s | 1.0× | slow |  |

## 5. Unavailable models

### 5.1 Failure categories

| Category | Meaning | Suggested action |
|---|---|---|
| **Model not found / unavailable (404/400)** | The upstream does not know the model, the account cannot access it, or it is disabled; includes models the gateway rejects locally. Model-level. | Remove it from the list or switch to a currently valid model name. |

Note: messages are the upstream responses forwarded by the gateway with the `[connection/model] [status]:` wrapper removed. The HTTP column shows `endpoint / upstream`.

### 5.2 Model not found / unavailable (404/400)（1）

#### b（1）

| Model | HTTP | Elapsed | Message |
|---|---|---|---|
| `b/missing` | 404 | 0.00s | The model 'missing' does not exist |

## 6. Family details (every model with its provider and results)

Same order as the composite ranking in 4.1: available families first, then entirely unavailable ones. Notes show the failure reason and message, or that an available model failed in some rounds.

### 1. fast

Key `fast||` | members 1, available 1 | providers: a(1) | score 100.0 (stability 100 / breadth 100 / speed 100) | 1 available providers, 1 stable | stability A stable（1/1） | latency min 0.01s / median 0.01s / max 0.01s

| Model | Provider | Status | Stability | Median | Min | Max | Served model | Context | Notes |
|---|---|---|---|---|---|---|---|---|---|
| `a/fast` | a | ✅ ok | A stable（1/1） | 0.01s | 0.01s | 0.01s | fast | - |  |

### 2. slow

Key `slow||` | members 1, available 1 | providers: a(1) | score 100.0 (stability 100 / breadth 100 / speed 100) | 1 available providers, 1 stable | stability A stable（1/1） | latency min 0.04s / median 0.04s / max 0.04s

| Model | Provider | Status | Stability | Median | Min | Max | Served model | Context | Notes |
|---|---|---|---|---|---|---|---|---|---|
| `a/slow` | a | ✅ ok | A stable（1/1） | 0.04s | 0.04s | 0.04s | slow | - |  |

### 3. missing

Key `missing||` | members 1, available 0 | providers: b(1)

| Model | Provider | Status | Stability | Median | Min | Max | Served model | Context | Notes |
|---|---|---|---|---|---|---|---|---|---|
| `b/missing` | b | ❌ Model not found / unavailable (404/400) | 0/1 | - | - | - | - | - | HTTP 404：The model 'missing' does not exist |

## Latency lab

Streaming time-to-first-token (TTFT), time-to-second-token (TTST), time per output token (TPOT), end-to-end latency (E2E) and decode throughput (tok/s, N−1 definition). Statistics exclude warm-up samples; the cold column is the first warm-up sample.

| # | Model | Samples (ok) | TTFT p50 | TTFT p90 | TTST p50 | TPOT p50 | E2E p50 | tok/s p50 | tok/s p90 | Cold TTFT | Output tokens | Notes |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| 1 | `a/fast` | 1(1) | 13 ms | 13 ms | 2 ms | 2 ms | 18 ms | 429.4 | 429.4 | - | 3 |  |

## Capability matrix

| Model | Supported | Streaming (SSE) | Usage in stream | Tool calling | JSON mode | Structured output (json_schema) | Vision input | System prompt adherence | Honours max_tokens | Honours stop sequences | Deterministic at temperature 0 | Accepts reasoning_effort | Long context | Embeddings endpoint |
|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|
| `a/fast` | 12/13 | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ❌ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |

Evidence:

- `a/fast`
  - Streaming (SSE): supported — 3 chunks, first token after 13ms
  - Usage in stream: supported — usage.completion_tokens=3
  - Tool calling: supported — called get_weather({"city":"Paris"})
  - JSON mode: supported — valid JSON object: {"answer":"OK"}
  - Structured output (json_schema): supported — schema-conformant JSON: {"answer":"OK","confidence":1}
  - Vision input: supported — answered: one two three
  - System prompt adherence: unsupported — ignored the system instruction: one two three
  - Honours max_tokens: supported — stopped after 3 tokens (finish_reason=stop)
  - Honours stop sequences: supported — stopped before 4: one two three
  - Deterministic at temperature 0: supported — identical outputs: one two three
  - Accepts reasoning_effort: supported — accepted reasoning_effort=low
  - Long context: supported — accepted ~2000 tokens
  - Embeddings endpoint: supported — 4 dimensions

## Quality spot-check

Built-in deterministic question set (no judge model). Score is the median over trials of the weighted pass rate; consistency is the share of questions answered identically across trials; identity is the model's own answer to “which model are you”.

| # | Model | Score | Trials | Consistency | Identity | Category pass rates |
|---|---|---|---|---|---|---|
| 1 | `a/fast` | 0.0 | 1 | 100% | one two three | arithmetic 0%, code 0%, extraction 0%, instruction 0%, knowledge 0%, language 0%, logic 0%, reasoning 0%, string 0% |

## 10. Caveats

- Latency is the end-to-end time of small single requests with limited samples; use it for relative comparisons, not as a throughput benchmark. See the latency lab for tokens/s.
- All traffic passes through the local network and possibly proxies; absolute numbers vary with network, nodes and time of day, so relative order within one run is more meaningful.
- Many failures are **provider/account-level** (missing credentials, expired tokens, exhausted balance or quota, plan restrictions); the models themselves may work once the account is fixed.
- 429-class errors are split into momentary rate limits and exhausted quotas based on the error text, so a few may be misclassified; the original messages are in section 5.
- Family grouping is rule-based plus aliases learned from this endpoint; different providers may serve different backends under the same name (see the served model column).
