# LLM API Speed Test Report

- Generated: 2026-11-01 22:57:58
- Test window: 2026-11-01 22:12:58 ～ 2026-11-01 22:41:58
- Endpoint: `http://localhost:20128/v1` (detected as 9router)
- Scope: **163** models, 10 providers, grouped into **78** families（excluded 14 combo models）
- Rounds: round 1: 100 requests (20 ok); round 2: 10 requests (2 ok)
- Tool: llm-speedtest test

## 1. Summary

| Metric | Value |
|---|---|
| Models tested | 163 |
| Available (at least one success) | **34**（20.9%） |
| Unavailable | 129（79.1%） |
| …of which provider/account-level (credentials, auth, balance, quota, plan) | 78（60.5%） |
| Families | 78（with an available member 21，entirely unavailable 57） |
| Stability of available models (A all rounds ok, B ≥ 2/3, C unstable) | A 27，B 4，C 3 |
| Stability of available families | A 15，B 6，C 0 |
| Latency median / P90 / min / max | 1.05s / 10.06s / 0.21s / 20.04s |

Latency is the end-to-end time of a non-streaming `chat/completions` request with a tiny `max_tokens`, i.e. roughly time-to-first-token plus routing overhead as perceived by the client, not the provider's pure inference latency.

### 1.1 Top 20 most stable and fastest models

| # | Model | Provider | Family | Stability (ok/attempts) | Median | Min | Max |
|---|---|---|---|---|---|---|---|
| 1 | `amd/MiniCPM-V-4.7-1B` | amd | minicpm-vision-4 | A stable（3/3） | 0.21s | 0.21s | 0.29s |
| 2 | `cf/@cf/meta/llama-3.2-1b-instruct` | cf | llama-3 | A stable（3/3） | 0.41s | 0.40s | 0.42s |
| 3 | `groq/allam-2-7b` | groq | allam-2 | A stable（3/3） | 0.43s | 0.42s | 1.12s |
| 4 | `amd/MiniCPM5-2B` | amd | minicpm5 | A stable（3/3） | 0.62s | 0.16s | 8.70s |
| 5 | `groq/qwen/qwen3.8-27b` | groq | qwen3 | A stable（3/3） | 0.85s | 0.24s | 0.93s |
| 6 | `groq/meta-llama/llama-prompt-guard-2-86m` | groq | llama-prompt-guard-2 | A stable（3/3） | 0.89s | 0.26s | 1.00s |
| 7 | `groq/meta-llama/llama-prompt-guard-2-22m` | groq | llama-prompt-guard-2 | A stable（3/3） | 0.94s | 0.89s | 0.99s |
| 8 | `cf/@cf/meta/llama-3.3-70b-instruct-fp8-fast` | cf | llama-3 | A stable（3/3） | 0.98s | 0.34s | 1.01s |
| 9 | `groq/openai/gpt-oss-safeguard-20b` | groq | gpt-oss-guard | A stable（3/3） | 0.99s | 0.95s | 1.03s |
| 10 | `cf/@cf/zai-org/glm-4.7-flash` | cf | glm-4-flash | A stable（3/3） | 1.00s | 0.76s | 1.12s |
| 11 | `cf/@cf/meta/llama-3.1-8b-instruct-fp8-fast` | cf | llama-3 | A stable（3/3） | 1.01s | 0.98s | 1.29s |
| 12 | `amd/Qwen3.8-Flash-Next` | amd | qwen3-flash | A stable（3/3） | 1.02s | 0.85s | 1.43s |
| 13 | `groq/openai/gpt-oss-120b` | groq | gpt-oss | A stable（3/3） | 1.03s | 1.00s | 1.14s |
| 14 | `cf/@cf/meta/llama-3.2-3b-instruct` | cf | llama-3 | A stable（3/3） | 1.04s | 0.95s | 1.06s |
| 15 | `groq/openai/gpt-oss-20b` | groq | gpt-oss | A stable（3/3） | 1.05s | 0.39s | 1.23s |
| 16 | `cf/@cf/qwen/qwq-32b` | cf | qwq | A stable（3/3） | 1.05s | 0.95s | 1.11s |
| 17 | `cf/@cf/qwen/qwen2.5-coder-32b-instruct` | cf | qwen2-coder | A stable（3/3） | 1.08s | 0.43s | 1.81s |
| 18 | `cf/@cf/deepseek-ai/deepseek-r1-distill-qwen-32b` | cf | deepseek-r1 | A stable（3/3） | 1.17s | 1.08s | 1.24s |
| 19 | `cf/@cf/mistralai/mistral-small-3.1-24b-instruct` | cf | mistral-small-3 | A stable（3/3） | 1.18s | 0.97s | 1.34s |
| 20 | `amd/Qwen3.8-27B` | amd | qwen3 | A stable（3/3） | 1.30s | 0.68s | 2.34s |

### 1.2 Latency distribution of available models

| Median latency | Models | Share |
|---|---|---|
| < 1 s | 12 | 35.3% |
| 1 – 2 s | 12 | 35.3% |
| 2 – 5 s | 4 | 11.8% |
| 5 – 10 s | 2 | 5.9% |
| 10 – 30 s | 4 | 11.8% |
| ≥ 30 s | 0 | 0.0% |

### 1.3 Failure reasons

| Reason | Models | Share | Level |
|---|---|---|---|
| Insufficient balance | 71 | 55.0% | provider / account |
| Free quota exhausted | 5 | 3.9% | provider / account |
| Terms acceptance required | 2 | 1.6% | provider / account |
| Rate limited (429) | 2 | 1.6% | model / network |
| Model not found / unavailable (404/400) | 19 | 14.7% | model / network |
| Model retired (410) | 6 | 4.7% | model / network |
| Not a chat model | 7 | 5.4% | model / network |
| Endpoint not supported (405) | 3 | 2.3% | model / network |
| Empty upstream response | 14 | 10.9% | model / network |

### 1.4 Top 15 families by composite score

Composite = stability × 0.4 + provider breadth × 0.3 + speed × 0.3; see section 4.

| # | Family | Score | Providers | Available/total | Stability | Median | Best member @ provider |
|---|---|---|---|---|---|---|---|
| 1 | **qwen3** | 84.4 | 2（stable 2） | 2/6 | A stable（6/6） | 1.08s | `groq/qwen/qwen3.8-27b` @ groq：0.85s |
| 2 | **allam-2** | 80.0 | 1（stable 1） | 1/1 | A stable（3/3） | 0.43s | `groq/allam-2-7b` @ groq：0.43s |
| 3 | **minicpm5** | 78.4 | 1（stable 1） | 1/1 | A stable（3/3） | 0.62s | `amd/MiniCPM5-2B` @ amd：0.62s |
| 4 | **llama-prompt-guard-2** | 75.6 | 1（stable 1） | 2/2 | A stable（6/6） | 0.91s | `groq/meta-llama/llama-prompt-guard-2-86m` @ groq：0.89s |
| 5 | **gpt-oss-guard** | 75.0 | 1（stable 1） | 1/1 | A stable（3/3） | 0.99s | `groq/openai/gpt-oss-safeguard-20b` @ groq：0.99s |
| 6 | **glm-4-flash** | 74.9 | 1（stable 1） | 1/1 | A stable（3/3） | 1.00s | `cf/@cf/zai-org/glm-4.7-flash` @ cf：1.00s |
| 7 | **qwen3-flash** | 74.7 | 1（stable 1） | 1/4 | A stable（3/3） | 1.02s | `amd/Qwen3.8-Flash-Next` @ amd：1.02s |
| 8 | **mimo-v2-flash** | 74.7 | 2（stable 2） | 2/4 | A stable（6/6） | 4.05s | `amd/MiMo-V2.6-Flash` @ amd：1.63s |
| 9 | **gpt-oss** | 74.6 | 1（stable 1） | 2/2 | A stable（6/6） | 1.04s | `groq/openai/gpt-oss-120b` @ groq：1.03s |
| 10 | **qwq** | 74.5 | 1（stable 1） | 1/1 | A stable（3/3） | 1.05s | `cf/@cf/qwen/qwq-32b` @ cf：1.05s |
| 11 | **qwen2-coder** | 74.4 | 1（stable 1） | 1/1 | A stable（3/3） | 1.08s | `cf/@cf/qwen/qwen2.5-coder-32b-instruct` @ cf：1.08s |
| 12 | **minicpm-vision-4** | 74.3 | 1（stable 1） | 2/2 | B mostly stable（6/7） | 0.29s | `amd/MiniCPM-V-4.7-1B` @ amd：0.21s |
| 13 | **deepseek-r1** | 73.8 | 1（stable 1） | 1/2 | A stable（3/3） | 1.17s | `cf/@cf/deepseek-ai/deepseek-r1-distill-qwen-32b` @ cf：1.17s |
| 14 | **mistral-small-3** | 73.7 | 1（stable 1） | 1/1 | A stable（3/3） | 1.18s | `cf/@cf/mistralai/mistral-small-3.1-24b-instruct` @ cf：1.18s |
| 15 | **deepseek-v4-flash** | 72.3 | 3（stable 2） | 5/13 | B mostly stable（11/16） | 4.00s | `thai/deepseek-v4.1-flash:free` @ thai：3.64s |

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

Provider field for this endpoint: `owned_by`（owned_by discriminates between models (10 distinct values); consistent with observed connections for 100% of 88 models）

## 3. Providers

| Provider | Models | Families | Available | Rate | Median latency | Top failure reason | Diagnosis |
|---|---|---|---|---|---|---|---|
| cf | 11 | 7 | 10 | 90.9% | 1.05s | Model not found / unavailable (404/400) | Model not found / unavailable (404/400): 1/1 failed models |
| amd | 16 | 14 | 9 | 56.2% | 1.30s | Not a chat model | Not a chat model: 5/7 failed models |
| groq | 13 | 10 | 7 | 53.8% | 0.94s | Model not found / unavailable (404/400) | Model not found / unavailable (404/400): 2/6 failed models |
| thai | 75 | 33 | 4 | 5.3% | 5.23s | Insufficient balance | Insufficient balance: 71/71 failed models |
| stkj | 8 | 6 | 4 | 50.0% | 2.22s | Model not found / unavailable (404/400) | Model not found / unavailable (404/400): 3/4 failed models |
| cu | 14 | 8 | 0 | 0.0% | - | Empty upstream response | Empty upstream response: 14/14 failed models |
| oc | 12 | 11 | 0 | 0.0% | - | Model not found / unavailable (404/400) | Model not found / unavailable (404/400): 7/12 failed models |
| nvidia | 9 | 8 | 0 | 0.0% | - | Model retired (410) | Model retired (410): 6/9 failed models |
| cohere | 3 | 3 | 0 | 0.0% | - | Endpoint not supported (405) | Endpoint not supported (405): 3/3 failed models |
| ps | 2 | 1 | 0 | 0.0% | - | Model not found / unavailable (404/400) | Model not found / unavailable (404/400): 2/2 failed models |

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
| 1 | **qwen3** | **84.4** | 100 | 67 | 81 | 2（2） | 2/6 | A stable（6/6） | 1.08s | 0.24s | 2.34s | `groq/qwen/qwen3.8-27b` @ groq：0.85s（3/3）<br>`amd/Qwen3.8-27B` @ amd：1.30s（3/3） |
| 2 | **allam-2** | **80.0** | 100 | 33 | 100 | 1（1） | 1/1 | A stable（3/3） | 0.43s | 0.42s | 1.12s | `groq/allam-2-7b` @ groq：0.43s（3/3） |
| 3 | **minicpm5** | **78.4** | 100 | 33 | 95 | 1（1） | 1/1 | A stable（3/3） | 0.62s | 0.16s | 8.70s | `amd/MiniCPM5-2B` @ amd：0.62s（3/3） |
| 4 | **llama-prompt-guard-2** | **75.6** | 100 | 33 | 85 | 1（1） | 2/2 | A stable（6/6） | 0.91s | 0.26s | 1.00s | `groq/meta-llama/llama-prompt-guard-2-86m` @ groq：0.89s（3/3）<br>`groq/meta-llama/llama-prompt-guard-2-22m` @ groq：0.94s（3/3） |
| 5 | **gpt-oss-guard** | **75.0** | 100 | 33 | 83 | 1（1） | 1/1 | A stable（3/3） | 0.99s | 0.95s | 1.03s | `groq/openai/gpt-oss-safeguard-20b` @ groq：0.99s（3/3） |
| 6 | **glm-4-flash** | **74.9** | 100 | 33 | 83 | 1（1） | 1/1 | A stable（3/3） | 1.00s | 0.76s | 1.12s | `cf/@cf/zai-org/glm-4.7-flash` @ cf：1.00s（3/3） |
| 7 | **qwen3-flash** | **74.7** | 100 | 33 | 82 | 1（1） | 1/4 | A stable（3/3） | 1.02s | 0.85s | 1.43s | `amd/Qwen3.8-Flash-Next` @ amd：1.02s（3/3） |
| 8 | **mimo-v2-flash** | **74.7** | 100 | 67 | 49 | 2（2） | 2/4 | A stable（6/6） | 4.05s | 0.47s | 10.76s | `amd/MiMo-V2.6-Flash` @ amd：1.63s（3/3）<br>`thai/mimo-v2.6-flash:free` @ thai：6.47s（3/3） |
| 9 | **gpt-oss** | **74.6** | 100 | 33 | 82 | 1（1） | 2/2 | A stable（6/6） | 1.04s | 0.39s | 1.23s | `groq/openai/gpt-oss-120b` @ groq：1.03s（3/3）<br>`groq/openai/gpt-oss-20b` @ groq：1.05s（3/3） |
| 10 | **qwq** | **74.5** | 100 | 33 | 82 | 1（1） | 1/1 | A stable（3/3） | 1.05s | 0.95s | 1.11s | `cf/@cf/qwen/qwq-32b` @ cf：1.05s（3/3） |
| 11 | **qwen2-coder** | **74.4** | 100 | 33 | 81 | 1（1） | 1/1 | A stable（3/3） | 1.08s | 0.43s | 1.81s | `cf/@cf/qwen/qwen2.5-coder-32b-instruct` @ cf：1.08s（3/3） |
| 12 | **minicpm-vision-4** | **74.3** | 86 | 33 | 100 | 1（1） | 2/2 | B mostly stable（6/7） | 0.29s | 0.21s | 0.43s | `amd/MiniCPM-V-4.7-1B` @ amd：0.21s（3/3） |
| 13 | **deepseek-r1** | **73.8** | 100 | 33 | 79 | 1（1） | 1/2 | A stable（3/3） | 1.17s | 1.08s | 1.24s | `cf/@cf/deepseek-ai/deepseek-r1-distill-qwen-32b` @ cf：1.17s（3/3） |
| 14 | **mistral-small-3** | **73.7** | 100 | 33 | 79 | 1（1） | 1/1 | A stable（3/3） | 1.18s | 0.97s | 1.34s | `cf/@cf/mistralai/mistral-small-3.1-24b-instruct` @ cf：1.18s（3/3） |
| 15 | **deepseek-v4-flash** | **72.3** | 69 | 100 | 49 | 3（2） | 5/13 | B mostly stable（11/16） | 4.00s | 0.56s | 15.22s | `thai/deepseek-v4.1-flash:free` @ thai：3.64s（3/3）<br>`thai/deepseek-v4-flash:free` @ thai：4.00s（3/3）<br>`amd/DeepSeek-V4-Flash` @ amd：6.04s（3/3） |
| 16 | **sensenova-6-flash-lite** | **69.7** | 100 | 33 | 66 | 1（1） | 1/2 | A stable（3/3） | 2.04s | 1.53s | 2.41s | `stkj/sensenova-6.8-flash-lite` @ stkj：2.04s（3/3） |
| 17 | **llama-3** | **69.5** | 87 | 33 | 83 | 1（1） | 5/5 | B mostly stable（13/15） | 1.01s | 0.34s | 1.29s | `cf/@cf/meta/llama-3.2-1b-instruct` @ cf：0.41s（3/3）<br>`cf/@cf/meta/llama-3.3-70b-instruct-fp8-fast` @ cf：0.98s（3/3）<br>`cf/@cf/meta/llama-3.1-8b-instruct-fp8-fast` @ cf：1.01s（3/3）<br>…2 more in section 6 |
| 18 | **glm-5** | **62.6** | 67 | 33 | 87 | 1（0） | 1/6 | B mostly stable（2/3） | 0.87s | 0.80s | 0.93s | `stkj/glm-5.2` @ stkj：0.87s（2/3） |
| 19 | **mimo-v2** | **55.3** | 100 | 33 | 18 | 1（1） | 1/2 | A stable（3/3） | 14.57s | 1.54s | 41.49s | `thai/mimo-v2.5:free` @ thai：14.57s（3/3） |
| 20 | **deepseek-flash** | **43.9** | 75 | 33 | 13 | 1（0） | 1/1 | B mostly stable（3/4） | 17.72s | 13.69s | 21.05s | `stkj/deepseek-flash` @ stkj：17.72s（3/4） |
| 21 | **gemma-4** | **39.6** | 67 | 33 | 10 | 1（0） | 1/1 | B mostly stable（2/3） | 20.04s | 2.29s | 37.79s | `amd/Gemma-4-31B-IT` @ amd：20.04s（2/3） |

### 4.2 Families by stability tier and latency

| # | Family | Stability | Members | Available | Median | Available providers | Top failure reasons |
|---|---|---|---|---|---|---|---|
| 1 | **allam-2** | A stable（3/3） | 1 | 1 | 0.43s | groq(1) | - |
| 2 | **minicpm5** | A stable（3/3） | 1 | 1 | 0.62s | amd(1) | - |
| 3 | **llama-prompt-guard-2** | A stable（6/6） | 2 | 2 | 0.91s | groq(2) | - |
| 4 | **gpt-oss-guard** | A stable（3/3） | 1 | 1 | 0.99s | groq(1) | - |
| 5 | **glm-4-flash** | A stable（3/3） | 1 | 1 | 1.00s | cf(1) | - |
| 6 | **qwen3-flash** | A stable（3/3） | 4 | 1 | 1.02s | amd(1) | Insufficient balance(3) |
| 7 | **gpt-oss** | A stable（6/6） | 2 | 2 | 1.04s | groq(2) | - |
| 8 | **qwq** | A stable（3/3） | 1 | 1 | 1.05s | cf(1) | - |
| 9 | **qwen3** | A stable（6/6） | 6 | 2 | 1.08s | amd(1), groq(1) | Insufficient balance(3), Model not found / unavailable (404/400)(1) |
| 10 | **qwen2-coder** | A stable（3/3） | 1 | 1 | 1.08s | cf(1) | - |
| 11 | **deepseek-r1** | A stable（3/3） | 2 | 1 | 1.17s | cf(1) | Insufficient balance(1) |
| 12 | **mistral-small-3** | A stable（3/3） | 1 | 1 | 1.18s | cf(1) | - |
| 13 | **sensenova-6-flash-lite** | A stable（3/3） | 2 | 1 | 2.04s | stkj(1) | Model not found / unavailable (404/400)(1) |
| 14 | **mimo-v2-flash** | A stable（6/6） | 4 | 2 | 4.05s | amd(1), thai(1) | Insufficient balance(1), Free quota exhausted(1) |
| 15 | **mimo-v2** | A stable（3/3） | 2 | 1 | 14.57s | thai(1) | Insufficient balance(1) |
| 16 | **minicpm-vision-4** | B mostly stable（6/7） | 2 | 2 | 0.29s | amd(2) | - |
| 17 | **glm-5** | B mostly stable（2/3） | 6 | 1 | 0.87s | stkj(1) | Insufficient balance(4), Model retired (410)(1) |
| 18 | **llama-3** | B mostly stable（13/15） | 5 | 5 | 1.01s | cf(5) | - |
| 19 | **deepseek-v4-flash** | B mostly stable（11/16） | 13 | 5 | 4.00s | amd(2), thai(2), stkj(1) | Insufficient balance(4), Model retired (410)(2), Model not found / unavailable (404/400)(1) |
| 20 | **deepseek-flash** | B mostly stable（3/4） | 1 | 1 | 17.72s | stkj(1) | - |
| 21 | **gemma-4** | B mostly stable（2/3） | 1 | 1 | 20.04s | amd(1) | - |

### 4.3 Entirely unavailable families

| Family | Members | Top reasons | Providers |
|---|---|---|---|
| **gpt-6** | 10 | Insufficient balance(10) | thai(10) |
| **gpt-5** | 7 | Insufficient balance(6), Empty upstream response(1) | thai(6), cu(1) |
| **claude-4-opus** | 5 | Empty upstream response(4), Insufficient balance(1) | cu(4), thai(1) |
| **muse-spark-1** | 5 | Insufficient balance(3), Model not found / unavailable (404/400)(2) | thai(3), oc(2) |
| **claude-4-sonnet** | 4 | Empty upstream response(3), Insufficient balance(1) | cu(3), thai(1) |
| **claude-opus-5** | 4 | Insufficient balance(4) | thai(4) |
| **gemini-3-flash** | 4 | Insufficient balance(3), Empty upstream response(1) | thai(3), cu(1) |
| **grok-4** | 3 | Insufficient balance(3) | thai(3) |
| **kimi-k2** | 3 | Empty upstream response(1), Model not found / unavailable (404/400)(1), Insufficient balance(1) | cu(1), nvidia(1), thai(1) |
| **kimi-k3** | 3 | Insufficient balance(2), Rate limited (429)(1) | thai(2), stkj(1) |
| **claude-fable-5** | 2 | Insufficient balance(2) | thai(2) |
| **claude-sonnet-5** | 2 | Insufficient balance(2) | thai(2) |
| **deepseek-v3** | 2 | Not a chat model(1), Insufficient balance(1) | amd(1), thai(1) |
| **deepseek-v4-pro** | 2 | Model retired (410)(1), Insufficient balance(1) | nvidia(1), thai(1) |
| **glm-5-flash** | 2 | Insufficient balance(1), Rate limited (429)(1) | amd(1), thai(1) |
| **gpt-5-codex** | 2 | Empty upstream response(2) | cu(2) |
| **laguna-2** | 2 | Model not found / unavailable (404/400)(2) | ps(2) |
| **mimo-v2-pro** | 2 | Insufficient balance(2) | thai(2) |
| **qwen3-max** | 2 | Insufficient balance(2) | thai(2) |
| **big-pickle** | 1 | Free quota exhausted(1) | oc(1) |
| **cat-jam** | 1 | Model not found / unavailable (404/400)(1) | oc(1) |
| **claude-4-haiku** | 1 | Empty upstream response(1) | cu(1) |
| **claude-haiku-5** | 1 | Insufficient balance(1) | thai(1) |
| **clef** | 1 | Not a chat model(1) | amd(1) |
| **clef-flash** | 1 | Model not found / unavailable (404/400)(1) | cf(1) |
| **command-a** | 1 | Endpoint not supported (405)(1) | cohere(1) |
| **command-r** | 1 | Endpoint not supported (405)(1) | cohere(1) |
| **command-r-plus** | 1 | Endpoint not supported (405)(1) | cohere(1) |
| **compound-mini** | 1 | Model not found / unavailable (404/400)(1) | groq(1) |
| **d3-flash** | 1 | Not a chat model(1) | amd(1) |
| **deepseek-v4-flash-vision** | 1 | Model not found / unavailable (404/400)(1) | amd(1) |
| **default** | 1 | Empty upstream response(1) | cu(1) |
| **gemini-4** | 1 | Insufficient balance(1) | thai(1) |
| **glm-5-flashx** | 1 | Insufficient balance(1) | thai(1) |
| **jev-1** | 1 | Free quota exhausted(1) | oc(1) |
| **kimi-k3-pro** | 1 | Insufficient balance(1) | thai(1) |
| **ling-3-flash** | 1 | Insufficient balance(1) | thai(1) |
| **longcat-2** | 1 | Model not found / unavailable (404/400)(1) | oc(1) |
| **mineru2-pro** | 1 | Not a chat model(1) | amd(1) |
| **minimax-m2** | 1 | Model retired (410)(1) | nvidia(1) |
| **minimax-m3** | 1 | Model retired (410)(1) | nvidia(1) |
| **minimax-m3-flash** | 1 | Insufficient balance(1) | thai(1) |
| **mistral-large** | 1 | Insufficient balance(1) | thai(1) |
| **mistral-large-4** | 1 | Insufficient balance(1) | thai(1) |
| **muse-wonder-1** | 1 | Insufficient balance(1) | thai(1) |
| **nemotron-3-lightning** | 1 | Free quota exhausted(1) | oc(1) |
| **nemotron-3-ultra** | 1 | Model not found / unavailable (404/400)(1) | nvidia(1) |
| **omnivoice** | 1 | Not a chat model(1) | amd(1) |
| **orpheus-arabic-saudi** | 1 | Terms acceptance required(1) | groq(1) |
| **orpheus-v1** | 1 | Terms acceptance required(1) | groq(1) |
| **parakeet-ctc-asr** | 1 | Model not found / unavailable (404/400)(1) | nvidia(1) |
| **sensenova-u1-lite** | 1 | Model not found / unavailable (404/400)(1) | stkj(1) |
| **space-bunny** | 1 | Model not found / unavailable (404/400)(1) | oc(1) |
| **step-5** | 1 | Model not found / unavailable (404/400)(1) | oc(1) |
| **union** | 1 | Model not found / unavailable (404/400)(1) | oc(1) |
| **whisper-large-v3** | 1 | Not a chat model(1) | groq(1) |
| **whisper-large-v3-turbo** | 1 | Not a chat model(1) | groq(1) |

### 4.4 All available models (stability, then speed)

| # | Model | Provider | Family | Stability | Median | Min | Max | Jitter (max/min) | Served model | Notes |
|---|---|---|---|---|---|---|---|---|---|---|
| 1 | `amd/MiniCPM-V-4.7-1B` | amd | minicpm-vision-4 | A stable（3/3） | 0.21s | 0.21s | 0.29s | 1.4× | self-dploy/MiniCPM-V-4.7-1B |  |
| 2 | `cf/@cf/meta/llama-3.2-1b-instruct` | cf | llama-3 | A stable（3/3） | 0.41s | 0.40s | 0.42s | 1.1× | @cf/meta/llama-3.2-1b-instruct |  |
| 3 | `groq/allam-2-7b` | groq | allam-2 | A stable（3/3） | 0.43s | 0.42s | 1.12s | 2.7× | allam-2-7b |  |
| 4 | `amd/MiniCPM5-2B` | amd | minicpm5 | A stable（3/3） | 0.62s | 0.16s | 8.70s | 54.7× | self-dploy/MiniCPM5-2B |  |
| 5 | `groq/qwen/qwen3.8-27b` | groq | qwen3 | A stable（3/3） | 0.85s | 0.24s | 0.93s | 3.9× | qwen/qwen3.8-27b |  |
| 6 | `groq/meta-llama/llama-prompt-guard-2-86m` | groq | llama-prompt-guard-2 | A stable（3/3） | 0.89s | 0.26s | 1.00s | 3.8× | meta-llama/llama-prompt-guard-2-86m |  |
| 7 | `groq/meta-llama/llama-prompt-guard-2-22m` | groq | llama-prompt-guard-2 | A stable（3/3） | 0.94s | 0.89s | 0.99s | 1.1× | meta-llama/llama-prompt-guard-2-22m |  |
| 8 | `cf/@cf/meta/llama-3.3-70b-instruct-fp8-fast` | cf | llama-3 | A stable（3/3） | 0.98s | 0.34s | 1.01s | 3.0× | @cf/meta/llama-3.3-70b-instruct-fp8-fast |  |
| 9 | `groq/openai/gpt-oss-safeguard-20b` | groq | gpt-oss-guard | A stable（3/3） | 0.99s | 0.95s | 1.03s | 1.1× | openai/gpt-oss-safeguard-20b |  |
| 10 | `cf/@cf/zai-org/glm-4.7-flash` | cf | glm-4-flash | A stable（3/3） | 1.00s | 0.76s | 1.12s | 1.5× | @cf/zai-org/glm-4.7-flash |  |
| 11 | `cf/@cf/meta/llama-3.1-8b-instruct-fp8-fast` | cf | llama-3 | A stable（3/3） | 1.01s | 0.98s | 1.29s | 1.3× | @cf/meta/llama-3.1-8b-instruct-fp8-fast |  |
| 12 | `amd/Qwen3.8-Flash-Next` | amd | qwen3-flash | A stable（3/3） | 1.02s | 0.85s | 1.43s | 1.7× | self-dploy/Qwen3.8-Flash-Next |  |
| 13 | `groq/openai/gpt-oss-120b` | groq | gpt-oss | A stable（3/3） | 1.03s | 1.00s | 1.14s | 1.1× | openai/gpt-oss-120b |  |
| 14 | `cf/@cf/meta/llama-3.2-3b-instruct` | cf | llama-3 | A stable（3/3） | 1.04s | 0.95s | 1.06s | 1.1× | @cf/meta/llama-3.2-3b-instruct |  |
| 15 | `groq/openai/gpt-oss-20b` | groq | gpt-oss | A stable（3/3） | 1.05s | 0.39s | 1.23s | 3.1× | openai/gpt-oss-20b |  |
| 16 | `cf/@cf/qwen/qwq-32b` | cf | qwq | A stable（3/3） | 1.05s | 0.95s | 1.11s | 1.2× | @cf/qwen/qwq-32b |  |
| 17 | `cf/@cf/qwen/qwen2.5-coder-32b-instruct` | cf | qwen2-coder | A stable（3/3） | 1.08s | 0.43s | 1.81s | 4.2× | @cf/qwen/qwen2.5-coder-32b-instruct |  |
| 18 | `cf/@cf/deepseek-ai/deepseek-r1-distill-qwen-32b` | cf | deepseek-r1 | A stable（3/3） | 1.17s | 1.08s | 1.24s | 1.1× | @cf/deepseek-ai/deepseek-r1-distill-qwe… |  |
| 19 | `cf/@cf/mistralai/mistral-small-3.1-24b-instruct` | cf | mistral-small-3 | A stable（3/3） | 1.18s | 0.97s | 1.34s | 1.4× | @cf/mistralai/mistral-small-3.1-24b-ins… |  |
| 20 | `amd/Qwen3.8-27B` | amd | qwen3 | A stable（3/3） | 1.30s | 0.68s | 2.34s | 3.5× | self-dploy/Qwen3.8-27B |  |
| 21 | `amd/MiMo-V2.6-Flash` | amd | mimo-v2-flash | A stable（3/3） | 1.63s | 0.63s | 2.19s | 3.5× | self-dploy/MiMo-V2.6-Flash |  |
| 22 | `stkj/sensenova-6.8-flash-lite` | stkj | sensenova-6-flash-lite | A stable（3/3） | 2.04s | 1.53s | 2.41s | 1.6× | sensenova-6.8-flash-lite |  |
| 23 | `thai/deepseek-v4.1-flash:free` | thai | deepseek-v4-flash | A stable（3/3） | 3.64s | 0.56s | 4.35s | 7.7× | deepseek-v4.1-flash |  |
| 24 | `thai/deepseek-v4-flash:free` | thai | deepseek-v4-flash | A stable（3/3） | 4.00s | 1.58s | 4.09s | 2.6× | deepseek-v4-flash |  |
| 25 | `amd/DeepSeek-V4-Flash` | amd | deepseek-v4-flash | A stable（3/3） | 6.04s | 1.02s | 15.22s | 15.0× | self-dploy/DeepSeek-V4-Flash |  |
| 26 | `thai/mimo-v2.6-flash:free` | thai | mimo-v2-flash | A stable（3/3） | 6.47s | 0.47s | 10.76s | 23.1× | mimo-v2.6-flash |  |
| 27 | `thai/mimo-v2.5:free` | thai | mimo-v2 | A stable（3/3） | 14.57s | 1.54s | 41.49s | 27.0× | mimo-v2.5 |  |
| 28 | `amd/MiniCPM-V-4.7-35B-A3B` | amd | minicpm-vision-4 | B mostly stable（3/4） | 0.37s | 0.29s | 0.43s | 1.5× | self-dploy/MiniCPM-V-4.7-35B-A3B | failed Rate limited (429)×1, succeeded on retry |
| 29 | `stkj/glm-5.2` | stkj | glm-5 | B mostly stable（2/3） | 0.87s | 0.80s | 0.93s | 1.2× | glm-5.2 | failed Rate limited (429)×1, succeeded on retry |
| 30 | `stkj/deepseek-flash` | stkj | deepseek-flash | B mostly stable（3/4） | 17.72s | 13.69s | 21.05s | 1.5× | deepseek-flash | failed Rate limited (429)×1, succeeded on retry |
| 31 | `amd/Gemma-4-31B-IT` | amd | gemma-4 | B mostly stable（2/3） | 20.04s | 2.29s | 37.79s | 16.5× | self-dploy/Gemma-4-31B-IT | failed Timeout×1, succeeded on retry |
| 32 | `cf/@cf/meta/llama-3.1-70b-instruct-fp8-fast` | cf | llama-3 | C unstable（1/3） | 1.25s | 1.25s | 1.25s | 1.0× | @cf/meta/llama-3.1-70b-instruct-fp8-fast | failed Timeout×2, succeeded on retry |
| 33 | `stkj/deepseek-v4-flash` | stkj | deepseek-v4-flash | C unstable（1/3） | 2.40s | 2.40s | 2.40s | 1.0× | deepseek-v4-flash | failed Rate limited (429)×2, succeeded on retry |
| 34 | `amd/DeepSeek-V4.1-Flash` | amd | deepseek-v4-flash | C unstable（1/4） | 11.59s | 11.59s | 11.59s | 1.0× | self-dploy/DeepSeek-V4.1-Flash | failed Rate limited (429)×3, succeeded on retry |

## 5. Unavailable models

### 5.1 Failure categories

| Category | Meaning | Suggested action |
|---|---|---|
| **Insufficient balance** | The upstream account has no balance or credits left; billing pre-authorisation failed. Account-level, unrelated to the model. | Top up the account or switch accounts. |
| **Free quota exhausted** | The free request quota, free-model allowance or gifted tokens for the account are used up (usually 429); it resets with the quota window rather than being a momentary rate limit. | Wait for the quota window to reset, top up or switch accounts. |
| **Terms acceptance required** | The provider requires an organisation admin to accept the model's terms of use in the console first. | Accept the terms in the provider console. |
| **Rate limited (429)** | Momentary RPM/TPM limits, per-minute free-tier caps or 'all channels busy' on an aggregator; still failing after slower retries. | Reduce the request rate and retry later. |
| **Model not found / unavailable (404/400)** | The upstream does not know the model, the account cannot access it, or it is disabled; includes models the gateway rejects locally. Model-level. | Remove it from the list or switch to a currently valid model name. |
| **Model retired (410)** | The upstream states the model reached end of life; most responses name a replacement. | Use the replacement model named in the error. |
| **Not a chat model** | The upstream states the model does not support /v1/chat/completions (image generation, OCR, speech or other dedicated endpoints). | Use the dedicated endpoint for that modality. |
| **Endpoint not supported (405)** | The gateway adapter for this provider does not support the chat completions call. | Use the provider's native API or wait for gateway support. |
| **Empty upstream response** | The upstream connection succeeded but returned no content (also on retry). | Switch models or retry later. |

Note: messages are the upstream responses forwarded by the gateway with the `[connection/model] [status]:` wrapper removed. The HTTP column shows `endpoint / upstream`.

### 5.2 Insufficient balance（71）

#### thai（71）

| Model | HTTP | Elapsed | Message |
|---|---|---|---|
| `thai/claude-fable-5` | 503/402 | 1.26s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 1s) |
| `thai/claude-fable-5.1` | 503/402 | 2.34s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 11s) |
| `thai/claude-haiku-5.5` | 503/402 | 0.50s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 24s) |
| `thai/claude-opus-4.8` | 503/402 | 1.37s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 1m 2s) |
| `thai/claude-opus-5` | 503/402 | 1.09s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 42s) |
| `thai/claude-opus-5-fast` | 503/402 | 1.08s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 16s) |
| `thai/claude-opus-5.5` | 503/402 | 0.56s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 11s) |
| `thai/claude-opus-5.5-fast` | 503/402 | 1.25s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 16s) |
| `thai/claude-sonnet-4.6` | 503/402 | 1.14s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 9s) |
| `thai/claude-sonnet-5` | 503/402 | 1.17s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 3s) |
| `thai/claude-sonnet-5.5` | 503/402 | 1.21s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 12s) |
| `thai/deepseek-r1-0528` | 503/402 | 1.03s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 49s) |
| `thai/deepseek-v3.2` | 503/402 | 0.39s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 2s) |
| `thai/deepseek-v4-flash` | 503/402 | 0.61s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 16s) |
| `thai/deepseek-v4-pro` | 503/402 | 0.37s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 6s) |
| `thai/deepseek-v4.1-flash` | 503/402 | 1.17s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 2s) |
| `thai/deepseek-v4.1-flash-1017` | 503/402 | 1.36s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 3s) |
| `thai/deepseek-v4.1-flash-1017-fast` | 503/402 | 0.49s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 6s) |
| `thai/gemini-3.6-flash` | 503/402 | 1.08s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 2s) |
| `thai/gemini-3.7-flash` | 503/402 | 0.65s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 59s) |
| `thai/gemini-3.8-flash` | 503/402 | 1.23s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 6s) |
| `thai/gemini-4-argon` | 503/402 | 0.98s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 28s) |
| `thai/glm-5.1` | 503/402 | 1.30s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 10s) |
| `thai/glm-5.2` | 503/402 | 1.09s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 2s) |
| `thai/glm-5.3` | 503/402 | 1.12s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 1m 41s) |
| `thai/glm-5.3-fast` | 503/402 | 1.10s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 5s) |
| `thai/glm-5.3-flash` | 503/402 | 0.44s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 11s) |
| `thai/glm-5.3-flashx` | 503/402 | 0.69s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 2s) |
| `thai/gpt-5.6-luna` | 503/402 | 0.35s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 2s) |
| `thai/gpt-5.6-luna-fast` | 503/402 | 0.53s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 16s) |
| `thai/gpt-5.6-sol` | 503/402 | 1.63s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 4s) |
| `thai/gpt-5.6-sol-fast` | 503/402 | 0.43s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 9s) |
| `thai/gpt-5.6-terra` | 503/402 | 1.05s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 27s) |
| `thai/gpt-5.6-terra-fast` | 503/402 | 0.99s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 9s) |
| `thai/gpt-6-astra` | 503/402 | 1.30s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 23s) |
| `thai/gpt-6-astra-fast` | 503/402 | 1.10s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 2s) |
| `thai/gpt-6-luna` | 503/402 | 1.47s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 15s) |
| `thai/gpt-6-luna-fast` | 503/402 | 1.16s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 7s) |
| `thai/gpt-6-sol` | 503/402 | 0.69s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 26s) |
| `thai/gpt-6-sol-fast` | 503/402 | 0.41s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 9s) |
| `thai/gpt-6.1-sol` | 503/402 | 2.60s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 2s) |
| `thai/gpt-6.1-sol-fast` | 503/402 | 2.47s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 12s) |
| `thai/gpt-6.1-terra` | 503/402 | 1.17s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 9s) |
| `thai/gpt-6.1-terra-fast` | 503/402 | 1.53s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 24s) |
| `thai/grok-4.5` | 503/402 | 1.15s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 1s) |
| `thai/grok-4.6` | 503/402 | 1.22s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 29s) |
| `thai/grok-4.7` | 503/402 | 1.12s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 2s) |
| `thai/kimi-k2.6` | 503/402 | 1.09s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 48s) |
| `thai/kimi-k3` | 503/402 | 0.63s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 18s) |
| `thai/kimi-k3-fast` | 503/402 | 1.30s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 2s) |
| `thai/kimi-k3.1-pro` | 503/402 | 0.48s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 11s) |
| `thai/ling-3.1-flash` | 503/402 | 1.06s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 28s) |
| `thai/mimo-v2.5` | 503/402 | 1.12s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 5s) |
| `thai/mimo-v2.5-pro` | 503/402 | 1.61s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 2m) |
| `thai/mimo-v2.6-flash` | 503/402 | 1.18s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 4s) |
| `thai/mimo-v2.6-pro` | 503/402 | 2.25s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 20s) |
| `thai/minimax-m3.1-flash` | 503/402 | 1.10s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 4s) |
| `thai/mistral-large-2512` | 503/402 | 1.07s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 3s) |
| `thai/mistral-large-4` | 503/402 | 1.54s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 20s) |
| `thai/muse-spark-1-3` | 503/402 | 1.02s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 1m 41s) |
| `thai/muse-spark-1.1` | 503/402 | 1.08s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 33s) |
| `thai/muse-spark-1.2` | 503/402 | 1.11s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 1s) |
| `thai/muse-wonder-1.0` | 503/402 | 1.38s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 16s) |
| `thai/qwen3.5-27b` | 503/402 | 1.32s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 14s) |
| `thai/qwen3.6-27b` | 503/402 | 0.43s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 1m 59s) |
| `thai/qwen3.6-flash` | 503/402 | 2.27s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 15s) |
| `thai/qwen3.7-flash` | 503/402 | 1.22s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 12s) |
| `thai/qwen3.7-max` | 503/402 | 0.51s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 23s) |
| `thai/qwen3.8-27b` | 503/402 | 1.30s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 12s) |
| `thai/qwen3.8-flash` | 503/402 | 0.39s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 10s) |
| `thai/qwen3.8-max` | 503/402 | 1.05s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 2s) |

### 5.3 Free quota exhausted（5）

#### oc（5）

All 5 models report the same error: `Free usage limit reached. Free models work in OpenCode (https://opencode.ai); subscribe to OpenCode Go to keep going: https://opencode.ai/go`

| Model | HTTP | Elapsed | Message |
|---|---|---|---|
| `oc/big-pickle` | 429 | 1.10s | Free usage limit reached. Free models work in OpenCode (https://opencode.ai); subscribe to OpenCode Go to keep going: https://opencode.ai/go |
| `oc/deepseek-v4.1-flash-free` | 429 | 1.05s | Free usage limit reached. Free models work in OpenCode (https://opencode.ai); subscribe to OpenCode Go to keep going: https://opencode.ai/go |
| `oc/jev-1.13-free` | 429 | 1.12s | Free usage limit reached. Free models work in OpenCode (https://opencode.ai); subscribe to OpenCode Go to keep going: https://opencode.ai/go |
| `oc/mimo-v2.6-flash-free` | 429 | 1.12s | Free usage limit reached. Free models work in OpenCode (https://opencode.ai); subscribe to OpenCode Go to keep going: https://opencode.ai/go |
| `oc/nemotron-3.5-lightning-free` | 429 | 1.23s | Free usage limit reached. Free models work in OpenCode (https://opencode.ai); subscribe to OpenCode Go to keep going: https://opencode.ai/go |

### 5.4 Terms acceptance required（2）

#### groq（2）

| Model | HTTP | Elapsed | Message |
|---|---|---|---|
| `groq/canopylabs/orpheus-arabic-saudi` | 400 | 0.99s | The model `canopylabs/orpheus-arabic-saudi` requires terms acceptance. Please have the org admin accept the terms at https://console.groq.com/playground?model=… |
| `groq/canopylabs/orpheus-v1-english` | 400 | 0.85s | The model `canopylabs/orpheus-v1-english` requires terms acceptance. Please have the org admin accept the terms at https://console.groq.com/playground?model=ca… |

### 5.5 Rate limited (429)（2）

#### amd（1）

| Model | HTTP | Elapsed | Message |
|---|---|---|---|
| `amd/GLM-5.3-Flash` | 503/429 | 0.14s | {'error': {'message': "Model 'GLM-5.3-Flash' is at its concurrency limit (32); please retry later or use another model", 'type': 'rate_limit_error', 'code': 'm… |

#### stkj（1）

| Model | HTTP | Elapsed | Message |
|---|---|---|---|
| `stkj/kimi-k3` | 503/429 | 0.19s | inference exceeds tpm/rpm limit (reset after 8s) |

### 5.6 Model not found / unavailable (404/400)（19）

#### oc（7）

All 7 models report the same error: `Upstream request failed: Model is unavailable.`

| Model | HTTP | Elapsed | Message |
|---|---|---|---|
| `oc/cat-jam-free` | 400 | 0.97s | Upstream request failed: Model is unavailable. |
| `oc/longcat-2.5-preview-free` | 400 | 1.03s | Upstream request failed: Model is unavailable. |
| `oc/muse-spark-1.2-contributor-free` | 400 | 0.95s | Upstream request failed: Model is unavailable. |
| `oc/muse-spark-1.3-contributor-free` | 400 | 0.87s | Upstream request failed: Model is unavailable. |
| `oc/space-bunny-free` | 400 | 0.89s | Upstream request failed: Model is unavailable. |
| `oc/step-5-preview-free` | 400 | 0.89s | Upstream request failed: Model is unavailable. |
| `oc/union-alpha` | 400 | 0.99s | Upstream request failed: Model is unavailable. |

#### nvidia（3）

| Model | HTTP | Elapsed | Message |
|---|---|---|---|
| `nvidia/moonshotai/kimi-k2.6` | 503/404 | 1.01s | Function '23d4f03a-b8a6-4adb-a183-7daa083a09cc': Not found for account 'W2B7Rw3HvwisCg4qUJPJzj1U1KQv4lpkBSYeuVEg4UM' (reset after 2m) |
| `nvidia/nemotron-3-ultra-550b-a55b` | 503/404 | 1.02s | 404 page not found (reset after 1m 25s) |
| `nvidia/parakeet-ctc-1.1b-asr` | 503/404 | 0.88s | 404 page not found (reset after 2m) |

#### stkj（3）

| Model | HTTP | Elapsed | Message |
|---|---|---|---|
| `stkj/deepseek-v4.1-flash` | 503/403 | 3.14s | model is not available in the current token plan (reset after 2m) |
| `stkj/sensenova-6.7-flash-lite` | 503/404 | 0.10s | model is not found (reset after 1m 18s) |
| `stkj/sensenova-u1.5-lite` | 503/404 | 0.13s | model is not found (reset after 2m) |

#### groq（2）

| Model | HTTP | Elapsed | Message |
|---|---|---|---|
| `groq/groq/compound-mini` | 503/404 | 1.00s | The model `groq/compound-mini` does not exist or you do not have access to it. (reset after 2m) |
| `groq/qwen/qwen3.6-27b` | 503/404 | 0.81s | The model `qwen/qwen3.6-27b` does not exist or you do not have access to it. (reset after 2m) |

#### ps（2）

| Model | HTTP | Elapsed | Message |
|---|---|---|---|
| `ps/laguna-s-2.1` | 503/404 | 1.24s | please check the model you provided (reset after 2m) |
| `ps/laguna-xs-2.1` | 503/404 | 1.06s | please check the model you provided (reset after 1m 25s) |

#### amd（1）

| Model | HTTP | Elapsed | Message |
|---|---|---|---|
| `amd/DeepSeek-V4-Flash-Vision-Exp` | 400 | 0.14s | Requested model DeepSeek-V4-Flash-Vision-Exp not supported |

#### cf（1）

| Model | HTTP | Elapsed | Message |
|---|---|---|---|
| `cf/@cf/cloudflare/clef-flash` | 400 | 1.60s | {"errors":[{"message":"AiError: No such model: No such model clef-flash or task (8f4d7a58-6cf3-4253-933b-39a8d38270ef)","code":5007}],"success":false,"result":… |

### 5.7 Model retired (410)（6）

#### nvidia（6）

| Model | HTTP | Elapsed | Message |
|---|---|---|---|
| `nvidia/deepseek-ai/deepseek-v4-flash` | 410 | 0.88s | The model 'deepseek-ai/deepseek-v4-flash' has reached its end of life on 2026-08-07T09:00:00Z and is no longer available. |
| `nvidia/deepseek-ai/deepseek-v4-flash-0731` | 410 | 0.86s | The model 'deepseek-ai/deepseek-v4-flash-0731' has reached its end of life on 2026-09-21T08:00:00Z and is no longer available. |
| `nvidia/deepseek-ai/deepseek-v4-pro` | 410 | 0.98s | The model 'deepseek-ai/deepseek-v4-pro' has reached its end of life on 2026-08-07T09:00:00Z and is no longer available. |
| `nvidia/minimaxai/minimax-m2.7` | 410 | 0.19s | The model 'minimaxai/minimax-m2.7' has reached its end of life on 2026-07-27T00:00:00Z and is no longer available. |
| `nvidia/minimaxai/minimax-m3` | 410 | 0.96s | The model 'minimaxai/minimax-m3' has reached its end of life on 2026-09-09T09:00:00Z and is no longer available. |
| `nvidia/z-ai/glm-5.2` | 410 | 0.85s | The model 'z-ai/glm-5.2' has reached its end of life on 2026-08-21T09:00:00Z and is no longer available. |

### 5.8 Not a chat model（7）

#### amd（5）

| Model | HTTP | Elapsed | Message |
|---|---|---|---|
| `amd/Clef` | 400 | 0.14s | Model Clef is not a chat model; call /v1/systemone instead of /v1/chat/completions |
| `amd/d3` | 400 | 0.15s | Model d3 is not a chat model; call /v1/systemone instead of /v1/chat/completions |
| `amd/d3-flash` | 400 | 0.13s | Model d3-flash is not a chat model; call /v1/systemone instead of /v1/chat/completions |
| `amd/MinerU2.5-Pro` | 400 | 0.17s | Model MinerU2.5-Pro is not a chat model; call /v1/ocr instead of /v1/chat/completions |
| `amd/OmniVoice` | 400 | 0.17s | Model OmniVoice is not a chat model; call /v1/audio/speech instead of /v1/chat/completions |

#### groq（2）

| Model | HTTP | Elapsed | Message |
|---|---|---|---|
| `groq/whisper-large-v3` | 400 | 0.92s | The model `whisper-large-v3` does not support chat completions |
| `groq/whisper-large-v3-turbo` | 400 | 1.01s | The model `whisper-large-v3-turbo` does not support chat completions |

### 5.9 Endpoint not supported (405)（3）

#### cohere（3）

All 3 models report the same error: `HTTP 405`

| Model | HTTP | Elapsed | Message |
|---|---|---|---|
| `cohere/command-a-03-2025` | 405 | 0.99s | HTTP 405 |
| `cohere/command-r-08-2024` | 405 | 0.87s | HTTP 405 |
| `cohere/command-r-plus-08-2024` | 405 | 2.05s | HTTP 405 |

### 5.10 Empty upstream response（14）

#### cu（14）

All 14 models report the same error: `Cursor AgentService returned an empty turn`

| Model | HTTP | Elapsed | Message |
|---|---|---|---|
| `cu/claude-4.5-haiku` | 400 | 0.95s | Cursor AgentService returned an empty turn |
| `cu/claude-4.5-opus` | 400 | 1.07s | Cursor AgentService returned an empty turn |
| `cu/claude-4.5-opus-high` | 400 | 1.06s | Cursor AgentService returned an empty turn |
| `cu/claude-4.5-opus-high-thinking` | 400 | 0.96s | Cursor AgentService returned an empty turn |
| `cu/claude-4.5-sonnet` | 400 | 1.07s | Cursor AgentService returned an empty turn |
| `cu/claude-4.5-sonnet-thinking` | 400 | 1.28s | Cursor AgentService returned an empty turn |
| `cu/claude-4.6-opus-max` | 400 | 1.14s | Cursor AgentService returned an empty turn |
| `cu/claude-4.6-sonnet-medium-thinking` | 400 | 1.05s | Cursor AgentService returned an empty turn |
| `cu/default` | 400 | 2.04s | Cursor AgentService returned an empty turn |
| `cu/gemini-3-flash-preview` | 400 | 1.01s | Cursor AgentService returned an empty turn |
| `cu/gpt-5.2` | 400 | 1.02s | Cursor AgentService returned an empty turn |
| `cu/gpt-5.2-codex` | 400 | 1.07s | Cursor AgentService returned an empty turn |
| `cu/gpt-5.3-codex` | 400 | 0.96s | Cursor AgentService returned an empty turn |
| `cu/kimi-k2.5` | 400 | 0.95s | Cursor AgentService returned an empty turn |

## 6. Family details (every model with its provider and results)

Same order as the composite ranking in 4.1: available families first, then entirely unavailable ones. Notes show the failure reason and message, or that an available model failed in some rounds.

### 1. qwen3

Key `qwen||3` | members 6, available 2 | providers: thai(3), groq(2), amd(1) | score 84.4 (stability 100 / breadth 67 / speed 81) | 2 available providers, 2 stable | stability A stable（6/6） | latency min 0.24s / median 1.08s / max 2.34s

| Model | Provider | Status | Stability | Median | Min | Max | Served model | Context | Notes |
|---|---|---|---|---|---|---|---|---|---|
| `groq/qwen/qwen3.8-27b` | groq | ✅ ok | A stable（3/3） | 0.85s | 0.24s | 0.93s | qwen/qwen3.8-27b | - |  |
| `amd/Qwen3.8-27B` | amd | ✅ ok | A stable（3/3） | 1.30s | 0.68s | 2.34s | self-dploy/Qwen3.8-27B | - |  |
| `groq/qwen/qwen3.6-27b` | groq | ❌ Model not found / unavailable (404/400) | 0/1 | - | - | - | - | - | HTTP 503/404：The model `qwen/qwen3.6-27b` does not exist or you do not have access to it. (reset after 2m) |
| `thai/qwen3.5-27b` | thai | ❌ Insufficient balance | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |
| `thai/qwen3.6-27b` | thai | ❌ Insufficient balance | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |
| `thai/qwen3.8-27b` | thai | ❌ Insufficient balance | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |

### 2. allam-2

Key `allam||2` | members 1, available 1 | providers: groq(1) | score 80.0 (stability 100 / breadth 33 / speed 100) | 1 available providers, 1 stable | stability A stable（3/3） | latency min 0.42s / median 0.43s / max 1.12s

| Model | Provider | Status | Stability | Median | Min | Max | Served model | Context | Notes |
|---|---|---|---|---|---|---|---|---|---|
| `groq/allam-2-7b` | groq | ✅ ok | A stable（3/3） | 0.43s | 0.42s | 1.12s | allam-2-7b | - |  |

### 3. minicpm5

Key `minicpm||5` | members 1, available 1 | providers: amd(1) | score 78.4 (stability 100 / breadth 33 / speed 95) | 1 available providers, 1 stable | stability A stable（3/3） | latency min 0.16s / median 0.62s / max 8.70s

| Model | Provider | Status | Stability | Median | Min | Max | Served model | Context | Notes |
|---|---|---|---|---|---|---|---|---|---|
| `amd/MiniCPM5-2B` | amd | ✅ ok | A stable（3/3） | 0.62s | 0.16s | 8.70s | self-dploy/MiniCPM5-2B | - |  |

### 4. llama-prompt-guard-2

Key `llama-prompt|guard|2` | members 2, available 2 | providers: groq(2) | score 75.6 (stability 100 / breadth 33 / speed 85) | 1 available providers, 1 stable | stability A stable（6/6） | latency min 0.26s / median 0.91s / max 1.00s

| Model | Provider | Status | Stability | Median | Min | Max | Served model | Context | Notes |
|---|---|---|---|---|---|---|---|---|---|
| `groq/meta-llama/llama-prompt-guard-2-86m` | groq | ✅ ok | A stable（3/3） | 0.89s | 0.26s | 1.00s | meta-llama/llama-prompt-guard-2-86m | - |  |
| `groq/meta-llama/llama-prompt-guard-2-22m` | groq | ✅ ok | A stable（3/3） | 0.94s | 0.89s | 0.99s | meta-llama/llama-prompt-guard-2-22m | - |  |

### 5. gpt-oss-guard

Key `gpt-oss|guard|` | members 1, available 1 | providers: groq(1) | score 75.0 (stability 100 / breadth 33 / speed 83) | 1 available providers, 1 stable | stability A stable（3/3） | latency min 0.95s / median 0.99s / max 1.03s

| Model | Provider | Status | Stability | Median | Min | Max | Served model | Context | Notes |
|---|---|---|---|---|---|---|---|---|---|
| `groq/openai/gpt-oss-safeguard-20b` | groq | ✅ ok | A stable（3/3） | 0.99s | 0.95s | 1.03s | openai/gpt-oss-safeguard-20b | - |  |

### 6. glm-4-flash

Key `glm|flash|4` | members 1, available 1 | providers: cf(1) | score 74.9 (stability 100 / breadth 33 / speed 83) | 1 available providers, 1 stable | stability A stable（3/3） | latency min 0.76s / median 1.00s / max 1.12s

| Model | Provider | Status | Stability | Median | Min | Max | Served model | Context | Notes |
|---|---|---|---|---|---|---|---|---|---|
| `cf/@cf/zai-org/glm-4.7-flash` | cf | ✅ ok | A stable（3/3） | 1.00s | 0.76s | 1.12s | @cf/zai-org/glm-4.7-flash | - |  |

### 7. qwen3-flash

Key `qwen|flash|3` | members 4, available 1 | providers: thai(3), amd(1) | score 74.7 (stability 100 / breadth 33 / speed 82) | 1 available providers, 1 stable | stability A stable（3/3） | latency min 0.85s / median 1.02s / max 1.43s

| Model | Provider | Status | Stability | Median | Min | Max | Served model | Context | Notes |
|---|---|---|---|---|---|---|---|---|---|
| `amd/Qwen3.8-Flash-Next` | amd | ✅ ok | A stable（3/3） | 1.02s | 0.85s | 1.43s | self-dploy/Qwen3.8-Flash-Next | - |  |
| `thai/qwen3.6-flash` | thai | ❌ Insufficient balance | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |
| `thai/qwen3.7-flash` | thai | ❌ Insufficient balance | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |
| `thai/qwen3.8-flash` | thai | ❌ Insufficient balance | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |

### 8. mimo-v2-flash

Key `mimo|flash|2` | members 4, available 2 | providers: thai(2), amd(1), oc(1) | score 74.7 (stability 100 / breadth 67 / speed 49) | 2 available providers, 2 stable | stability A stable（6/6） | latency min 0.47s / median 4.05s / max 10.76s

| Model | Provider | Status | Stability | Median | Min | Max | Served model | Context | Notes |
|---|---|---|---|---|---|---|---|---|---|
| `amd/MiMo-V2.6-Flash` | amd | ✅ ok | A stable（3/3） | 1.63s | 0.63s | 2.19s | self-dploy/MiMo-V2.6-Flash | - |  |
| `thai/mimo-v2.6-flash:free` | thai | ✅ ok | A stable（3/3） | 6.47s | 0.47s | 10.76s | mimo-v2.6-flash | - |  |
| `oc/mimo-v2.6-flash-free` | oc | ❌ Free quota exhausted | 0/2 | - | - | - | - | - | HTTP 429：Free usage limit reached. Free models work in OpenCode (https://opencode.ai); subscribe to OpenCode Go to keep going: h… |
| `thai/mimo-v2.6-flash` | thai | ❌ Insufficient balance | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |

### 9. gpt-oss

Key `gpt-oss||` | members 2, available 2 | providers: groq(2) | score 74.6 (stability 100 / breadth 33 / speed 82) | 1 available providers, 1 stable | stability A stable（6/6） | latency min 0.39s / median 1.04s / max 1.23s

| Model | Provider | Status | Stability | Median | Min | Max | Served model | Context | Notes |
|---|---|---|---|---|---|---|---|---|---|
| `groq/openai/gpt-oss-120b` | groq | ✅ ok | A stable（3/3） | 1.03s | 1.00s | 1.14s | openai/gpt-oss-120b | - |  |
| `groq/openai/gpt-oss-20b` | groq | ✅ ok | A stable（3/3） | 1.05s | 0.39s | 1.23s | openai/gpt-oss-20b | - |  |

### 10. qwq

Key `qwq||` | members 1, available 1 | providers: cf(1) | score 74.5 (stability 100 / breadth 33 / speed 82) | 1 available providers, 1 stable | stability A stable（3/3） | latency min 0.95s / median 1.05s / max 1.11s

| Model | Provider | Status | Stability | Median | Min | Max | Served model | Context | Notes |
|---|---|---|---|---|---|---|---|---|---|
| `cf/@cf/qwen/qwq-32b` | cf | ✅ ok | A stable（3/3） | 1.05s | 0.95s | 1.11s | @cf/qwen/qwq-32b | - |  |

### 11. qwen2-coder

Key `qwen|coder|2` | members 1, available 1 | providers: cf(1) | score 74.4 (stability 100 / breadth 33 / speed 81) | 1 available providers, 1 stable | stability A stable（3/3） | latency min 0.43s / median 1.08s / max 1.81s

| Model | Provider | Status | Stability | Median | Min | Max | Served model | Context | Notes |
|---|---|---|---|---|---|---|---|---|---|
| `cf/@cf/qwen/qwen2.5-coder-32b-instruct` | cf | ✅ ok | A stable（3/3） | 1.08s | 0.43s | 1.81s | @cf/qwen/qwen2.5-coder-32b-instruct | - |  |

### 12. minicpm-vision-4

Key `minicpm|vision|4` | members 2, available 2 | providers: amd(2) | score 74.3 (stability 86 / breadth 33 / speed 100) | 1 available providers, 1 stable | stability B mostly stable（6/7） | latency min 0.21s / median 0.29s / max 0.43s

| Model | Provider | Status | Stability | Median | Min | Max | Served model | Context | Notes |
|---|---|---|---|---|---|---|---|---|---|
| `amd/MiniCPM-V-4.7-1B` | amd | ✅ ok | A stable（3/3） | 0.21s | 0.21s | 0.29s | self-dploy/MiniCPM-V-4.7-1B | - |  |
| `amd/MiniCPM-V-4.7-35B-A3B` | amd | ✅ ok | B mostly stable（3/4） | 0.37s | 0.29s | 0.43s | self-dploy/MiniCPM-V-4.7-35B-A3B | - | failed Rate limited (429)×1, succeeded on retry |

### 13. deepseek-r1

Key `deepseek||1` | members 2, available 1 | providers: cf(1), thai(1) | score 73.8 (stability 100 / breadth 33 / speed 79) | 1 available providers, 1 stable | stability A stable（3/3） | latency min 1.08s / median 1.17s / max 1.24s

| Model | Provider | Status | Stability | Median | Min | Max | Served model | Context | Notes |
|---|---|---|---|---|---|---|---|---|---|
| `cf/@cf/deepseek-ai/deepseek-r1-distill-qwen-32b` | cf | ✅ ok | A stable（3/3） | 1.17s | 1.08s | 1.24s | @cf/deepseek-ai/deepseek-r1-distill-qwe… | - |  |
| `thai/deepseek-r1-0528` | thai | ❌ Insufficient balance | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |

### 14. mistral-small-3

Key `mistral|small|3` | members 1, available 1 | providers: cf(1) | score 73.7 (stability 100 / breadth 33 / speed 79) | 1 available providers, 1 stable | stability A stable（3/3） | latency min 0.97s / median 1.18s / max 1.34s

| Model | Provider | Status | Stability | Median | Min | Max | Served model | Context | Notes |
|---|---|---|---|---|---|---|---|---|---|
| `cf/@cf/mistralai/mistral-small-3.1-24b-instruct` | cf | ✅ ok | A stable（3/3） | 1.18s | 0.97s | 1.34s | @cf/mistralai/mistral-small-3.1-24b-ins… | - |  |

### 15. deepseek-v4-flash

Key `deepseek|flash|4` | members 13, available 5 | providers: thai(6), amd(2), nvidia(2), stkj(2), oc(1) | score 72.3 (stability 69 / breadth 100 / speed 49) | 3 available providers, 2 stable | stability B mostly stable（11/16） | latency min 0.56s / median 4.00s / max 15.22s

| Model | Provider | Status | Stability | Median | Min | Max | Served model | Context | Notes |
|---|---|---|---|---|---|---|---|---|---|
| `thai/deepseek-v4.1-flash:free` | thai | ✅ ok | A stable（3/3） | 3.64s | 0.56s | 4.35s | deepseek-v4.1-flash | - |  |
| `thai/deepseek-v4-flash:free` | thai | ✅ ok | A stable（3/3） | 4.00s | 1.58s | 4.09s | deepseek-v4-flash | - |  |
| `amd/DeepSeek-V4-Flash` | amd | ✅ ok | A stable（3/3） | 6.04s | 1.02s | 15.22s | self-dploy/DeepSeek-V4-Flash | - |  |
| `stkj/deepseek-v4-flash` | stkj | ✅ ok | C unstable（1/3） | 2.40s | 2.40s | 2.40s | deepseek-v4-flash | - | failed Rate limited (429)×2, succeeded on retry |
| `amd/DeepSeek-V4.1-Flash` | amd | ✅ ok | C unstable（1/4） | 11.59s | 11.59s | 11.59s | self-dploy/DeepSeek-V4.1-Flash | - | failed Rate limited (429)×3, succeeded on retry |
| `nvidia/deepseek-ai/deepseek-v4-flash` | nvidia | ❌ Model retired (410) | 0/1 | - | - | - | - | - | HTTP 410：The model 'deepseek-ai/deepseek-v4-flash' has reached its end of life on 2026-08-07T09:00:00Z and is no longer availabl… |
| `nvidia/deepseek-ai/deepseek-v4-flash-0731` | nvidia | ❌ Model retired (410) | 0/1 | - | - | - | - | - | HTTP 410：The model 'deepseek-ai/deepseek-v4-flash-0731' has reached its end of life on 2026-09-21T08:00:00Z and is no longer ava… |
| `oc/deepseek-v4.1-flash-free` | oc | ❌ Free quota exhausted | 0/2 | - | - | - | - | - | HTTP 429：Free usage limit reached. Free models work in OpenCode (https://opencode.ai); subscribe to OpenCode Go to keep going: h… |
| `stkj/deepseek-v4.1-flash` | stkj | ❌ Model not found / unavailable (404/400) | 0/1 | - | - | - | - | - | HTTP 503/403：model is not available in the current token plan (reset after 2m) |
| `thai/deepseek-v4-flash` | thai | ❌ Insufficient balance | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |
| `thai/deepseek-v4.1-flash` | thai | ❌ Insufficient balance | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |
| `thai/deepseek-v4.1-flash-1017` | thai | ❌ Insufficient balance | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |
| `thai/deepseek-v4.1-flash-1017-fast` | thai | ❌ Insufficient balance | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |

### 16. sensenova-6-flash-lite

Key `sensenova|flash+lite|6` | members 2, available 1 | providers: stkj(2) | score 69.7 (stability 100 / breadth 33 / speed 66) | 1 available providers, 1 stable | stability A stable（3/3） | latency min 1.53s / median 2.04s / max 2.41s

| Model | Provider | Status | Stability | Median | Min | Max | Served model | Context | Notes |
|---|---|---|---|---|---|---|---|---|---|
| `stkj/sensenova-6.8-flash-lite` | stkj | ✅ ok | A stable（3/3） | 2.04s | 1.53s | 2.41s | sensenova-6.8-flash-lite | - |  |
| `stkj/sensenova-6.7-flash-lite` | stkj | ❌ Model not found / unavailable (404/400) | 0/1 | - | - | - | - | - | HTTP 503/404：model is not found (reset after 1m 18s) |

### 17. llama-3

Key `llama||3` | members 5, available 5 | providers: cf(5) | score 69.5 (stability 87 / breadth 33 / speed 83) | 1 available providers, 1 stable | stability B mostly stable（13/15） | latency min 0.34s / median 1.01s / max 1.29s

| Model | Provider | Status | Stability | Median | Min | Max | Served model | Context | Notes |
|---|---|---|---|---|---|---|---|---|---|
| `cf/@cf/meta/llama-3.2-1b-instruct` | cf | ✅ ok | A stable（3/3） | 0.41s | 0.40s | 0.42s | @cf/meta/llama-3.2-1b-instruct | - |  |
| `cf/@cf/meta/llama-3.3-70b-instruct-fp8-fast` | cf | ✅ ok | A stable（3/3） | 0.98s | 0.34s | 1.01s | @cf/meta/llama-3.3-70b-instruct-fp8-fast | - |  |
| `cf/@cf/meta/llama-3.1-8b-instruct-fp8-fast` | cf | ✅ ok | A stable（3/3） | 1.01s | 0.98s | 1.29s | @cf/meta/llama-3.1-8b-instruct-fp8-fast | - |  |
| `cf/@cf/meta/llama-3.2-3b-instruct` | cf | ✅ ok | A stable（3/3） | 1.04s | 0.95s | 1.06s | @cf/meta/llama-3.2-3b-instruct | - |  |
| `cf/@cf/meta/llama-3.1-70b-instruct-fp8-fast` | cf | ✅ ok | C unstable（1/3） | 1.25s | 1.25s | 1.25s | @cf/meta/llama-3.1-70b-instruct-fp8-fast | - | failed Timeout×2, succeeded on retry |

### 18. glm-5

Key `glm||5` | members 6, available 1 | providers: thai(4), nvidia(1), stkj(1) | score 62.6 (stability 67 / breadth 33 / speed 87) | 1 available providers, 0 stable | stability B mostly stable（2/3） | latency min 0.80s / median 0.87s / max 0.93s

| Model | Provider | Status | Stability | Median | Min | Max | Served model | Context | Notes |
|---|---|---|---|---|---|---|---|---|---|
| `stkj/glm-5.2` | stkj | ✅ ok | B mostly stable（2/3） | 0.87s | 0.80s | 0.93s | glm-5.2 | - | failed Rate limited (429)×1, succeeded on retry |
| `nvidia/z-ai/glm-5.2` | nvidia | ❌ Model retired (410) | 0/1 | - | - | - | - | - | HTTP 410：The model 'z-ai/glm-5.2' has reached its end of life on 2026-08-21T09:00:00Z and is no longer available. |
| `thai/glm-5.1` | thai | ❌ Insufficient balance | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |
| `thai/glm-5.2` | thai | ❌ Insufficient balance | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |
| `thai/glm-5.3` | thai | ❌ Insufficient balance | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |
| `thai/glm-5.3-fast` | thai | ❌ Insufficient balance | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |

### 19. mimo-v2

Key `mimo||2` | members 2, available 1 | providers: thai(2) | score 55.3 (stability 100 / breadth 33 / speed 18) | 1 available providers, 1 stable | stability A stable（3/3） | latency min 1.54s / median 14.57s / max 41.49s

| Model | Provider | Status | Stability | Median | Min | Max | Served model | Context | Notes |
|---|---|---|---|---|---|---|---|---|---|
| `thai/mimo-v2.5:free` | thai | ✅ ok | A stable（3/3） | 14.57s | 1.54s | 41.49s | mimo-v2.5 | - |  |
| `thai/mimo-v2.5` | thai | ❌ Insufficient balance | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |

### 20. deepseek-flash

Key `deepseek|flash|` | members 1, available 1 | providers: stkj(1) | score 43.9 (stability 75 / breadth 33 / speed 13) | 1 available providers, 0 stable | stability B mostly stable（3/4） | latency min 13.69s / median 17.72s / max 21.05s

| Model | Provider | Status | Stability | Median | Min | Max | Served model | Context | Notes |
|---|---|---|---|---|---|---|---|---|---|
| `stkj/deepseek-flash` | stkj | ✅ ok | B mostly stable（3/4） | 17.72s | 13.69s | 21.05s | deepseek-flash | - | failed Rate limited (429)×1, succeeded on retry |

### 21. gemma-4

Key `gemma||4` | members 1, available 1 | providers: amd(1) | score 39.6 (stability 67 / breadth 33 / speed 10) | 1 available providers, 0 stable | stability B mostly stable（2/3） | latency min 2.29s / median 20.04s / max 37.79s

| Model | Provider | Status | Stability | Median | Min | Max | Served model | Context | Notes |
|---|---|---|---|---|---|---|---|---|---|
| `amd/Gemma-4-31B-IT` | amd | ✅ ok | B mostly stable（2/3） | 20.04s | 2.29s | 37.79s | self-dploy/Gemma-4-31B-IT | - | failed Timeout×1, succeeded on retry |

### 22. gpt-6

Key `gpt||6` | members 10, available 0 | providers: thai(10)

| Model | Provider | Status | Stability | Median | Min | Max | Served model | Context | Notes |
|---|---|---|---|---|---|---|---|---|---|
| `thai/gpt-6-astra` | thai | ❌ Insufficient balance | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |
| `thai/gpt-6-astra-fast` | thai | ❌ Insufficient balance | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |
| `thai/gpt-6-luna` | thai | ❌ Insufficient balance | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |
| `thai/gpt-6-luna-fast` | thai | ❌ Insufficient balance | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |
| `thai/gpt-6-sol` | thai | ❌ Insufficient balance | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |
| `thai/gpt-6-sol-fast` | thai | ❌ Insufficient balance | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |
| `thai/gpt-6.1-sol` | thai | ❌ Insufficient balance | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |
| `thai/gpt-6.1-sol-fast` | thai | ❌ Insufficient balance | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |
| `thai/gpt-6.1-terra` | thai | ❌ Insufficient balance | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |
| `thai/gpt-6.1-terra-fast` | thai | ❌ Insufficient balance | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |

### 23. gpt-5

Key `gpt||5` | members 7, available 0 | providers: thai(6), cu(1)

| Model | Provider | Status | Stability | Median | Min | Max | Served model | Context | Notes |
|---|---|---|---|---|---|---|---|---|---|
| `cu/gpt-5.2` | cu | ❌ Empty upstream response | 0/3 | - | - | - | - | - | HTTP 400：Cursor AgentService returned an empty turn |
| `thai/gpt-5.6-luna` | thai | ❌ Insufficient balance | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |
| `thai/gpt-5.6-luna-fast` | thai | ❌ Insufficient balance | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |
| `thai/gpt-5.6-sol` | thai | ❌ Insufficient balance | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |
| `thai/gpt-5.6-sol-fast` | thai | ❌ Insufficient balance | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |
| `thai/gpt-5.6-terra` | thai | ❌ Insufficient balance | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |
| `thai/gpt-5.6-terra-fast` | thai | ❌ Insufficient balance | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |

### 24. claude-4-opus

Key `claude|opus|4` | members 5, available 0 | providers: cu(4), thai(1)

| Model | Provider | Status | Stability | Median | Min | Max | Served model | Context | Notes |
|---|---|---|---|---|---|---|---|---|---|
| `cu/claude-4.5-opus` | cu | ❌ Empty upstream response | 0/3 | - | - | - | - | - | HTTP 400：Cursor AgentService returned an empty turn |
| `cu/claude-4.5-opus-high` | cu | ❌ Empty upstream response | 0/3 | - | - | - | - | - | HTTP 400：Cursor AgentService returned an empty turn |
| `cu/claude-4.5-opus-high-thinking` | cu | ❌ Empty upstream response | 0/3 | - | - | - | - | - | HTTP 400：Cursor AgentService returned an empty turn |
| `cu/claude-4.6-opus-max` | cu | ❌ Empty upstream response | 0/3 | - | - | - | - | - | HTTP 400：Cursor AgentService returned an empty turn |
| `thai/claude-opus-4.8` | thai | ❌ Insufficient balance | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |

### 25. muse-spark-1

Key `muse-spark||1` | members 5, available 0 | providers: thai(3), oc(2)

| Model | Provider | Status | Stability | Median | Min | Max | Served model | Context | Notes |
|---|---|---|---|---|---|---|---|---|---|
| `oc/muse-spark-1.2-contributor-free` | oc | ❌ Model not found / unavailable (404/400) | 0/1 | - | - | - | - | - | HTTP 400：Upstream request failed: Model is unavailable. |
| `oc/muse-spark-1.3-contributor-free` | oc | ❌ Model not found / unavailable (404/400) | 0/1 | - | - | - | - | - | HTTP 400：Upstream request failed: Model is unavailable. |
| `thai/muse-spark-1-3` | thai | ❌ Insufficient balance | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |
| `thai/muse-spark-1.1` | thai | ❌ Insufficient balance | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |
| `thai/muse-spark-1.2` | thai | ❌ Insufficient balance | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |

### 26. claude-4-sonnet

Key `claude|sonnet|4` | members 4, available 0 | providers: cu(3), thai(1)

| Model | Provider | Status | Stability | Median | Min | Max | Served model | Context | Notes |
|---|---|---|---|---|---|---|---|---|---|
| `cu/claude-4.5-sonnet` | cu | ❌ Empty upstream response | 0/3 | - | - | - | - | - | HTTP 400：Cursor AgentService returned an empty turn |
| `cu/claude-4.5-sonnet-thinking` | cu | ❌ Empty upstream response | 0/3 | - | - | - | - | - | HTTP 400：Cursor AgentService returned an empty turn |
| `cu/claude-4.6-sonnet-medium-thinking` | cu | ❌ Empty upstream response | 0/3 | - | - | - | - | - | HTTP 400：Cursor AgentService returned an empty turn |
| `thai/claude-sonnet-4.6` | thai | ❌ Insufficient balance | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |

### 27. claude-opus-5

Key `claude|opus|5` | members 4, available 0 | providers: thai(4)

| Model | Provider | Status | Stability | Median | Min | Max | Served model | Context | Notes |
|---|---|---|---|---|---|---|---|---|---|
| `thai/claude-opus-5` | thai | ❌ Insufficient balance | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |
| `thai/claude-opus-5-fast` | thai | ❌ Insufficient balance | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |
| `thai/claude-opus-5.5` | thai | ❌ Insufficient balance | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |
| `thai/claude-opus-5.5-fast` | thai | ❌ Insufficient balance | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |

### 28. gemini-3-flash

Key `gemini|flash|3` | members 4, available 0 | providers: thai(3), cu(1)

| Model | Provider | Status | Stability | Median | Min | Max | Served model | Context | Notes |
|---|---|---|---|---|---|---|---|---|---|
| `cu/gemini-3-flash-preview` | cu | ❌ Empty upstream response | 0/3 | - | - | - | - | - | HTTP 400：Cursor AgentService returned an empty turn |
| `thai/gemini-3.6-flash` | thai | ❌ Insufficient balance | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |
| `thai/gemini-3.7-flash` | thai | ❌ Insufficient balance | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |
| `thai/gemini-3.8-flash` | thai | ❌ Insufficient balance | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |

### 29. grok-4

Key `grok||4` | members 3, available 0 | providers: thai(3)

| Model | Provider | Status | Stability | Median | Min | Max | Served model | Context | Notes |
|---|---|---|---|---|---|---|---|---|---|
| `thai/grok-4.5` | thai | ❌ Insufficient balance | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |
| `thai/grok-4.6` | thai | ❌ Insufficient balance | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |
| `thai/grok-4.7` | thai | ❌ Insufficient balance | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |

### 30. kimi-k2

Key `kimi||2` | members 3, available 0 | providers: cu(1), nvidia(1), thai(1)

| Model | Provider | Status | Stability | Median | Min | Max | Served model | Context | Notes |
|---|---|---|---|---|---|---|---|---|---|
| `cu/kimi-k2.5` | cu | ❌ Empty upstream response | 0/3 | - | - | - | - | - | HTTP 400：Cursor AgentService returned an empty turn |
| `nvidia/moonshotai/kimi-k2.6` | nvidia | ❌ Model not found / unavailable (404/400) | 0/1 | - | - | - | - | - | HTTP 503/404：Function '23d4f03a-b8a6-4adb-a183-7daa083a09cc': Not found for account 'W2B7Rw3HvwisCg4qUJPJzj1U1KQv4lpkBSYeuVEg4UM' (r… |
| `thai/kimi-k2.6` | thai | ❌ Insufficient balance | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |

### 31. kimi-k3

Key `kimi||3` | members 3, available 0 | providers: thai(2), stkj(1)

| Model | Provider | Status | Stability | Median | Min | Max | Served model | Context | Notes |
|---|---|---|---|---|---|---|---|---|---|
| `stkj/kimi-k3` | stkj | ❌ Rate limited (429) | 0/3 | - | - | - | - | - | HTTP 503/429：inference exceeds tpm/rpm limit (reset after 8s) |
| `thai/kimi-k3` | thai | ❌ Insufficient balance | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |
| `thai/kimi-k3-fast` | thai | ❌ Insufficient balance | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |

### 32. claude-fable-5

Key `claude|fable|5` | members 2, available 0 | providers: thai(2)

| Model | Provider | Status | Stability | Median | Min | Max | Served model | Context | Notes |
|---|---|---|---|---|---|---|---|---|---|
| `thai/claude-fable-5` | thai | ❌ Insufficient balance | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |
| `thai/claude-fable-5.1` | thai | ❌ Insufficient balance | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |

### 33. claude-sonnet-5

Key `claude|sonnet|5` | members 2, available 0 | providers: thai(2)

| Model | Provider | Status | Stability | Median | Min | Max | Served model | Context | Notes |
|---|---|---|---|---|---|---|---|---|---|
| `thai/claude-sonnet-5` | thai | ❌ Insufficient balance | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |
| `thai/claude-sonnet-5.5` | thai | ❌ Insufficient balance | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |

### 34. deepseek-v3

Key `deepseek||3` | members 2, available 0 | providers: amd(1), thai(1) | 1 members joined via alias learning/overrides

| Model | Provider | Status | Stability | Median | Min | Max | Served model | Context | Notes |
|---|---|---|---|---|---|---|---|---|---|
| `amd/d3` | amd | ❌ Not a chat model | 0/1 | - | - | - | - | - | HTTP 400：Model d3 is not a chat model; call /v1/systemone instead of /v1/chat/completions |
| `thai/deepseek-v3.2` | thai | ❌ Insufficient balance | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |

### 35. deepseek-v4-pro

Key `deepseek|pro|4` | members 2, available 0 | providers: nvidia(1), thai(1)

| Model | Provider | Status | Stability | Median | Min | Max | Served model | Context | Notes |
|---|---|---|---|---|---|---|---|---|---|
| `nvidia/deepseek-ai/deepseek-v4-pro` | nvidia | ❌ Model retired (410) | 0/1 | - | - | - | - | - | HTTP 410：The model 'deepseek-ai/deepseek-v4-pro' has reached its end of life on 2026-08-07T09:00:00Z and is no longer available. |
| `thai/deepseek-v4-pro` | thai | ❌ Insufficient balance | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |

### 36. glm-5-flash

Key `glm|flash|5` | members 2, available 0 | providers: amd(1), thai(1)

| Model | Provider | Status | Stability | Median | Min | Max | Served model | Context | Notes |
|---|---|---|---|---|---|---|---|---|---|
| `amd/GLM-5.3-Flash` | amd | ❌ Rate limited (429) | 0/3 | - | - | - | - | - | HTTP 503/429：{'error': {'message': "Model 'GLM-5.3-Flash' is at its concurrency limit (32); please retry later or use another model"… |
| `thai/glm-5.3-flash` | thai | ❌ Insufficient balance | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |

### 37. gpt-5-codex

Key `gpt|codex|5` | members 2, available 0 | providers: cu(2)

| Model | Provider | Status | Stability | Median | Min | Max | Served model | Context | Notes |
|---|---|---|---|---|---|---|---|---|---|
| `cu/gpt-5.2-codex` | cu | ❌ Empty upstream response | 0/3 | - | - | - | - | - | HTTP 400：Cursor AgentService returned an empty turn |
| `cu/gpt-5.3-codex` | cu | ❌ Empty upstream response | 0/3 | - | - | - | - | - | HTTP 400：Cursor AgentService returned an empty turn |

### 38. laguna-2

Key `laguna||2` | members 2, available 0 | providers: ps(2)

| Model | Provider | Status | Stability | Median | Min | Max | Served model | Context | Notes |
|---|---|---|---|---|---|---|---|---|---|
| `ps/laguna-s-2.1` | ps | ❌ Model not found / unavailable (404/400) | 0/1 | - | - | - | - | - | HTTP 503/404：please check the model you provided (reset after 2m) |
| `ps/laguna-xs-2.1` | ps | ❌ Model not found / unavailable (404/400) | 0/1 | - | - | - | - | - | HTTP 503/404：please check the model you provided (reset after 1m 25s) |

### 39. mimo-v2-pro

Key `mimo|pro|2` | members 2, available 0 | providers: thai(2)

| Model | Provider | Status | Stability | Median | Min | Max | Served model | Context | Notes |
|---|---|---|---|---|---|---|---|---|---|
| `thai/mimo-v2.5-pro` | thai | ❌ Insufficient balance | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |
| `thai/mimo-v2.6-pro` | thai | ❌ Insufficient balance | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |

### 40. qwen3-max

Key `qwen|max|3` | members 2, available 0 | providers: thai(2)

| Model | Provider | Status | Stability | Median | Min | Max | Served model | Context | Notes |
|---|---|---|---|---|---|---|---|---|---|
| `thai/qwen3.7-max` | thai | ❌ Insufficient balance | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |
| `thai/qwen3.8-max` | thai | ❌ Insufficient balance | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |

### 41. big-pickle

Key `big-pickle||` | members 1, available 0 | providers: oc(1)

| Model | Provider | Status | Stability | Median | Min | Max | Served model | Context | Notes |
|---|---|---|---|---|---|---|---|---|---|
| `oc/big-pickle` | oc | ❌ Free quota exhausted | 0/2 | - | - | - | - | - | HTTP 429：Free usage limit reached. Free models work in OpenCode (https://opencode.ai); subscribe to OpenCode Go to keep going: h… |

### 42. cat-jam

Key `cat-jam||` | members 1, available 0 | providers: oc(1)

| Model | Provider | Status | Stability | Median | Min | Max | Served model | Context | Notes |
|---|---|---|---|---|---|---|---|---|---|
| `oc/cat-jam-free` | oc | ❌ Model not found / unavailable (404/400) | 0/1 | - | - | - | - | - | HTTP 400：Upstream request failed: Model is unavailable. |

### 43. claude-4-haiku

Key `claude|haiku|4` | members 1, available 0 | providers: cu(1)

| Model | Provider | Status | Stability | Median | Min | Max | Served model | Context | Notes |
|---|---|---|---|---|---|---|---|---|---|
| `cu/claude-4.5-haiku` | cu | ❌ Empty upstream response | 0/3 | - | - | - | - | - | HTTP 400：Cursor AgentService returned an empty turn |

### 44. claude-haiku-5

Key `claude|haiku|5` | members 1, available 0 | providers: thai(1)

| Model | Provider | Status | Stability | Median | Min | Max | Served model | Context | Notes |
|---|---|---|---|---|---|---|---|---|---|
| `thai/claude-haiku-5.5` | thai | ❌ Insufficient balance | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |

### 45. clef

Key `clef||` | members 1, available 0 | providers: amd(1)

| Model | Provider | Status | Stability | Median | Min | Max | Served model | Context | Notes |
|---|---|---|---|---|---|---|---|---|---|
| `amd/Clef` | amd | ❌ Not a chat model | 0/1 | - | - | - | - | - | HTTP 400：Model Clef is not a chat model; call /v1/systemone instead of /v1/chat/completions |

### 46. clef-flash

Key `clef|flash|` | members 1, available 0 | providers: cf(1)

| Model | Provider | Status | Stability | Median | Min | Max | Served model | Context | Notes |
|---|---|---|---|---|---|---|---|---|---|
| `cf/@cf/cloudflare/clef-flash` | cf | ❌ Model not found / unavailable (404/400) | 0/1 | - | - | - | - | - | HTTP 400：{"errors":[{"message":"AiError: No such model: No such model clef-flash or task (8f4d7a58-6cf3-4253-933b-39a8d38270ef)"… |

### 47. command-a

Key `command-a||` | members 1, available 0 | providers: cohere(1)

| Model | Provider | Status | Stability | Median | Min | Max | Served model | Context | Notes |
|---|---|---|---|---|---|---|---|---|---|
| `cohere/command-a-03-2025` | cohere | ❌ Endpoint not supported (405) | 0/1 | - | - | - | - | - | HTTP 405：HTTP 405 |

### 48. command-r

Key `command-r||` | members 1, available 0 | providers: cohere(1)

| Model | Provider | Status | Stability | Median | Min | Max | Served model | Context | Notes |
|---|---|---|---|---|---|---|---|---|---|
| `cohere/command-r-08-2024` | cohere | ❌ Endpoint not supported (405) | 0/1 | - | - | - | - | - | HTTP 405：HTTP 405 |

### 49. command-r-plus

Key `command-r|plus|` | members 1, available 0 | providers: cohere(1)

| Model | Provider | Status | Stability | Median | Min | Max | Served model | Context | Notes |
|---|---|---|---|---|---|---|---|---|---|
| `cohere/command-r-plus-08-2024` | cohere | ❌ Endpoint not supported (405) | 0/1 | - | - | - | - | - | HTTP 405：HTTP 405 |

### 50. compound-mini

Key `compound|mini|` | members 1, available 0 | providers: groq(1)

| Model | Provider | Status | Stability | Median | Min | Max | Served model | Context | Notes |
|---|---|---|---|---|---|---|---|---|---|
| `groq/groq/compound-mini` | groq | ❌ Model not found / unavailable (404/400) | 0/1 | - | - | - | - | - | HTTP 503/404：The model `groq/compound-mini` does not exist or you do not have access to it. (reset after 2m) |

### 51. d3-flash

Key `d|flash|3` | members 1, available 0 | providers: amd(1)

| Model | Provider | Status | Stability | Median | Min | Max | Served model | Context | Notes |
|---|---|---|---|---|---|---|---|---|---|
| `amd/d3-flash` | amd | ❌ Not a chat model | 0/1 | - | - | - | - | - | HTTP 400：Model d3-flash is not a chat model; call /v1/systemone instead of /v1/chat/completions |

### 52. deepseek-v4-flash-vision

Key `deepseek|flash+vision|4` | members 1, available 0 | providers: amd(1)

| Model | Provider | Status | Stability | Median | Min | Max | Served model | Context | Notes |
|---|---|---|---|---|---|---|---|---|---|
| `amd/DeepSeek-V4-Flash-Vision-Exp` | amd | ❌ Model not found / unavailable (404/400) | 0/1 | - | - | - | - | - | HTTP 400：Requested model DeepSeek-V4-Flash-Vision-Exp not supported |

### 53. default

Key `default||` | members 1, available 0 | providers: cu(1)

| Model | Provider | Status | Stability | Median | Min | Max | Served model | Context | Notes |
|---|---|---|---|---|---|---|---|---|---|
| `cu/default` | cu | ❌ Empty upstream response | 0/3 | - | - | - | - | - | HTTP 400：Cursor AgentService returned an empty turn |

### 54. gemini-4

Key `gemini||4` | members 1, available 0 | providers: thai(1)

| Model | Provider | Status | Stability | Median | Min | Max | Served model | Context | Notes |
|---|---|---|---|---|---|---|---|---|---|
| `thai/gemini-4-argon` | thai | ❌ Insufficient balance | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |

### 55. glm-5-flashx

Key `glm|flashx|5` | members 1, available 0 | providers: thai(1)

| Model | Provider | Status | Stability | Median | Min | Max | Served model | Context | Notes |
|---|---|---|---|---|---|---|---|---|---|
| `thai/glm-5.3-flashx` | thai | ❌ Insufficient balance | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |

### 56. jev-1

Key `jev||1` | members 1, available 0 | providers: oc(1)

| Model | Provider | Status | Stability | Median | Min | Max | Served model | Context | Notes |
|---|---|---|---|---|---|---|---|---|---|
| `oc/jev-1.13-free` | oc | ❌ Free quota exhausted | 0/2 | - | - | - | - | - | HTTP 429：Free usage limit reached. Free models work in OpenCode (https://opencode.ai); subscribe to OpenCode Go to keep going: h… |

### 57. kimi-k3-pro

Key `kimi|pro|3` | members 1, available 0 | providers: thai(1)

| Model | Provider | Status | Stability | Median | Min | Max | Served model | Context | Notes |
|---|---|---|---|---|---|---|---|---|---|
| `thai/kimi-k3.1-pro` | thai | ❌ Insufficient balance | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |

### 58. ling-3-flash

Key `ling|flash|3` | members 1, available 0 | providers: thai(1)

| Model | Provider | Status | Stability | Median | Min | Max | Served model | Context | Notes |
|---|---|---|---|---|---|---|---|---|---|
| `thai/ling-3.1-flash` | thai | ❌ Insufficient balance | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |

### 59. longcat-2

Key `longcat||2` | members 1, available 0 | providers: oc(1)

| Model | Provider | Status | Stability | Median | Min | Max | Served model | Context | Notes |
|---|---|---|---|---|---|---|---|---|---|
| `oc/longcat-2.5-preview-free` | oc | ❌ Model not found / unavailable (404/400) | 0/1 | - | - | - | - | - | HTTP 400：Upstream request failed: Model is unavailable. |

### 60. mineru2-pro

Key `mineru|pro|2` | members 1, available 0 | providers: amd(1)

| Model | Provider | Status | Stability | Median | Min | Max | Served model | Context | Notes |
|---|---|---|---|---|---|---|---|---|---|
| `amd/MinerU2.5-Pro` | amd | ❌ Not a chat model | 0/1 | - | - | - | - | - | HTTP 400：Model MinerU2.5-Pro is not a chat model; call /v1/ocr instead of /v1/chat/completions |

### 61. minimax-m2

Key `minimax||2` | members 1, available 0 | providers: nvidia(1)

| Model | Provider | Status | Stability | Median | Min | Max | Served model | Context | Notes |
|---|---|---|---|---|---|---|---|---|---|
| `nvidia/minimaxai/minimax-m2.7` | nvidia | ❌ Model retired (410) | 0/1 | - | - | - | - | - | HTTP 410：The model 'minimaxai/minimax-m2.7' has reached its end of life on 2026-07-27T00:00:00Z and is no longer available. |

### 62. minimax-m3

Key `minimax||3` | members 1, available 0 | providers: nvidia(1)

| Model | Provider | Status | Stability | Median | Min | Max | Served model | Context | Notes |
|---|---|---|---|---|---|---|---|---|---|
| `nvidia/minimaxai/minimax-m3` | nvidia | ❌ Model retired (410) | 0/1 | - | - | - | - | - | HTTP 410：The model 'minimaxai/minimax-m3' has reached its end of life on 2026-09-09T09:00:00Z and is no longer available. |

### 63. minimax-m3-flash

Key `minimax|flash|3` | members 1, available 0 | providers: thai(1)

| Model | Provider | Status | Stability | Median | Min | Max | Served model | Context | Notes |
|---|---|---|---|---|---|---|---|---|---|
| `thai/minimax-m3.1-flash` | thai | ❌ Insufficient balance | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |

### 64. mistral-large

Key `mistral|large|` | members 1, available 0 | providers: thai(1)

| Model | Provider | Status | Stability | Median | Min | Max | Served model | Context | Notes |
|---|---|---|---|---|---|---|---|---|---|
| `thai/mistral-large-2512` | thai | ❌ Insufficient balance | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |

### 65. mistral-large-4

Key `mistral|large|4` | members 1, available 0 | providers: thai(1)

| Model | Provider | Status | Stability | Median | Min | Max | Served model | Context | Notes |
|---|---|---|---|---|---|---|---|---|---|
| `thai/mistral-large-4` | thai | ❌ Insufficient balance | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |

### 66. muse-wonder-1

Key `muse-wonder||1` | members 1, available 0 | providers: thai(1)

| Model | Provider | Status | Stability | Median | Min | Max | Served model | Context | Notes |
|---|---|---|---|---|---|---|---|---|---|
| `thai/muse-wonder-1.0` | thai | ❌ Insufficient balance | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |

### 67. nemotron-3-lightning

Key `nemotron|lightning|3` | members 1, available 0 | providers: oc(1)

| Model | Provider | Status | Stability | Median | Min | Max | Served model | Context | Notes |
|---|---|---|---|---|---|---|---|---|---|
| `oc/nemotron-3.5-lightning-free` | oc | ❌ Free quota exhausted | 0/2 | - | - | - | - | - | HTTP 429：Free usage limit reached. Free models work in OpenCode (https://opencode.ai); subscribe to OpenCode Go to keep going: h… |

### 68. nemotron-3-ultra

Key `nemotron|ultra|3` | members 1, available 0 | providers: nvidia(1)

| Model | Provider | Status | Stability | Median | Min | Max | Served model | Context | Notes |
|---|---|---|---|---|---|---|---|---|---|
| `nvidia/nemotron-3-ultra-550b-a55b` | nvidia | ❌ Model not found / unavailable (404/400) | 0/1 | - | - | - | - | - | HTTP 503/404：404 page not found (reset after 1m 25s) |

### 69. omnivoice

Key `omnivoice||` | members 1, available 0 | providers: amd(1)

| Model | Provider | Status | Stability | Median | Min | Max | Served model | Context | Notes |
|---|---|---|---|---|---|---|---|---|---|
| `amd/OmniVoice` | amd | ❌ Not a chat model | 0/1 | - | - | - | - | - | HTTP 400：Model OmniVoice is not a chat model; call /v1/audio/speech instead of /v1/chat/completions |

### 70. orpheus-arabic-saudi

Key `orpheus-arabic-saudi||` | members 1, available 0 | providers: groq(1)

| Model | Provider | Status | Stability | Median | Min | Max | Served model | Context | Notes |
|---|---|---|---|---|---|---|---|---|---|
| `groq/canopylabs/orpheus-arabic-saudi` | groq | ❌ Terms acceptance required | 0/2 | - | - | - | - | - | HTTP 400：The model `canopylabs/orpheus-arabic-saudi` requires terms acceptance. Please have the org admin accept the terms at ht… |

### 71. orpheus-v1

Key `orpheus||1` | members 1, available 0 | providers: groq(1)

| Model | Provider | Status | Stability | Median | Min | Max | Served model | Context | Notes |
|---|---|---|---|---|---|---|---|---|---|
| `groq/canopylabs/orpheus-v1-english` | groq | ❌ Terms acceptance required | 0/2 | - | - | - | - | - | HTTP 400：The model `canopylabs/orpheus-v1-english` requires terms acceptance. Please have the org admin accept the terms at http… |

### 72. parakeet-ctc-asr

Key `parakeet-ctc|asr|` | members 1, available 0 | providers: nvidia(1)

| Model | Provider | Status | Stability | Median | Min | Max | Served model | Context | Notes |
|---|---|---|---|---|---|---|---|---|---|
| `nvidia/parakeet-ctc-1.1b-asr` | nvidia | ❌ Model not found / unavailable (404/400) | 0/1 | - | - | - | - | - | HTTP 503/404：404 page not found (reset after 2m) |

### 73. sensenova-u1-lite

Key `sensenova-u|lite|1` | members 1, available 0 | providers: stkj(1)

| Model | Provider | Status | Stability | Median | Min | Max | Served model | Context | Notes |
|---|---|---|---|---|---|---|---|---|---|
| `stkj/sensenova-u1.5-lite` | stkj | ❌ Model not found / unavailable (404/400) | 0/1 | - | - | - | - | - | HTTP 503/404：model is not found (reset after 2m) |

### 74. space-bunny

Key `space-bunny||` | members 1, available 0 | providers: oc(1)

| Model | Provider | Status | Stability | Median | Min | Max | Served model | Context | Notes |
|---|---|---|---|---|---|---|---|---|---|
| `oc/space-bunny-free` | oc | ❌ Model not found / unavailable (404/400) | 0/1 | - | - | - | - | - | HTTP 400：Upstream request failed: Model is unavailable. |

### 75. step-5

Key `step||5` | members 1, available 0 | providers: oc(1)

| Model | Provider | Status | Stability | Median | Min | Max | Served model | Context | Notes |
|---|---|---|---|---|---|---|---|---|---|
| `oc/step-5-preview-free` | oc | ❌ Model not found / unavailable (404/400) | 0/1 | - | - | - | - | - | HTTP 400：Upstream request failed: Model is unavailable. |

### 76. union

Key `union||` | members 1, available 0 | providers: oc(1)

| Model | Provider | Status | Stability | Median | Min | Max | Served model | Context | Notes |
|---|---|---|---|---|---|---|---|---|---|
| `oc/union-alpha` | oc | ❌ Model not found / unavailable (404/400) | 0/1 | - | - | - | - | - | HTTP 400：Upstream request failed: Model is unavailable. |

### 77. whisper-large-v3

Key `whisper|large|3` | members 1, available 0 | providers: groq(1)

| Model | Provider | Status | Stability | Median | Min | Max | Served model | Context | Notes |
|---|---|---|---|---|---|---|---|---|---|
| `groq/whisper-large-v3` | groq | ❌ Not a chat model | 0/2 | - | - | - | - | - | HTTP 400：The model `whisper-large-v3` does not support chat completions |

### 78. whisper-large-v3-turbo

Key `whisper|large+turbo|3` | members 1, available 0 | providers: groq(1)

| Model | Provider | Status | Stability | Median | Min | Max | Served model | Context | Notes |
|---|---|---|---|---|---|---|---|---|---|
| `groq/whisper-large-v3-turbo` | groq | ❌ Not a chat model | 0/2 | - | - | - | - | - | HTTP 400：The model `whisper-large-v3-turbo` does not support chat completions |

## 7. Caveats

- Latency is the end-to-end time of small single requests with limited samples; use it for relative comparisons, not as a throughput benchmark. See the latency lab for tokens/s.
- All traffic passes through the local network and possibly proxies; absolute numbers vary with network, nodes and time of day, so relative order within one run is more meaningful.
- Many failures are **provider/account-level** (missing credentials, expired tokens, exhausted balance or quota, plan restrictions); the models themselves may work once the account is fixed.
- 429-class errors are split into momentary rate limits and exhausted quotas based on the error text, so a few may be misclassified; the original messages are in section 5.
- Family grouping is rule-based plus aliases learned from this endpoint; different providers may serve different backends under the same name (see the served model column).
