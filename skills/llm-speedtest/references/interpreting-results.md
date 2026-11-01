# Interpreting results

## What to recommend

1. Start from `families` (composite order). A family at the top has several working providers, a high multi-round success rate and low latency — recommend its `best_id` / `best` member with the provider shown.
2. If the user needs one exact model id, use `models` ranked by tier then median latency; prefer tier A.
3. Mention the second-best option from a *different provider* for redundancy.

## Explaining failures

| Reason group | What to say | What to do |
|---|---|---|
| `no_credentials` | the gateway has no credential for that provider | add/enable the credential in the gateway |
| `auth` | the provider rejected the token/key (expired or disabled) | re-authenticate; every model of that provider is affected |
| `quota_billing`, `quota_exhausted`, `plan_restriction`, `terms_required` | account-level limits (balance, free quota, plan, terms) | top up / wait / upgrade; the models themselves may be fine |
| `model_unavailable`, `model_retired`, `not_chat_model*` | the model id is wrong, gone, or not a chat model | drop it or use the replacement named in the message |
| `rate_limit`, `upstream_5xx`, `upstream_network`, `no_channel`, `no_capacity`, `timeout`, `empty_response` | transient or capacity problems that persisted through retries | retry later; treat as unstable |

Quote the exact upstream message (it is preserved) when the user wants details.

## Sanity checks before trusting numbers

- Scan latency includes gateway and proxy overhead — compare models *within one run*, not across machines.
- A model with 1–2 samples has a weak median; check the `attempts` column.
- `tokens_exact = false` in the latency lab means token counts were estimated.
- Quality identity answers are self-reported; a mismatch between the requested id and the served model (`served_model` column) plus a low quality score suggests a mislabelled deployment.
- Load tests saturate the endpoint; results depend on the provider's current load and the user's plan limits.

## Phrasing

Lead with the answer ("Use X via provider Y: tier A, median 0.6 s, 3 providers available"), then the caveats, then the report path. Keep tables short; link to the full report or dashboard for everything else.
