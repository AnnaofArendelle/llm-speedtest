# HTTP API and MCP tools

The dashboard server (`llm-speedtest serve`, default `http://127.0.0.1:8787`) exposes a JSON API. When started with `--access-token`, send `Authorization: Bearer <token>` (or `?token=` for the SSE stream).

## HTTP API

| Method | Path | Purpose |
|---|---|---|
| GET | `/healthz` | liveness |
| GET | `/api/version` | build info |
| GET | `/api/reasons` | failure taxonomy with descriptions |
| GET | `/api/dialects`, `/api/probes`, `/api/presets`, `/api/questions` | reference data |
| GET/POST | `/api/endpoints` | list / create endpoints (`{name, base_url, api_key, dialect, headers, insecure, proxy}`) |
| GET/PUT/DELETE | `/api/endpoints/{id}` | read / update / delete |
| POST | `/api/endpoints/{id}/check` | connectivity check |
| POST | `/api/endpoints/{id}/discover` | `{filter}` → model preview |
| GET | `/api/endpoints/{id}/overrides` | provider and family overrides |
| PUT | `/api/endpoints/{id}/overrides/provider` | `{raw_key, display_name, merged_into}` |
| PUT | `/api/endpoints/{id}/overrides/family` | `{model_id, family_key, canon}` |
| POST | `/api/check`, `/api/discover` | ad-hoc (`{endpoint:{base_url,…}, api_key, filter}`) |
| POST | `/api/runs` | `{kind: suite|scan|latency|load|capability|quality, endpoint_id | endpoint+api_key, profile, suite, models, confirm}` → `202 {id}` |
| GET | `/api/runs?kind=&limit=&offset=` | list |
| GET | `/api/runs/{id}` | run, children, running flag, latest progress |
| POST | `/api/runs/{id}/cancel` | cancel |
| DELETE | `/api/runs/{id}` | delete |
| GET | `/api/runs/{id}/events` | server-sent events: `snapshot`, `round`, `attempt`, `progress`, `sample`, `cooldown`, `log`, `done`, `error` |
| GET | `/api/runs/{id}/summary` | headline numbers, top families/models, providers |
| GET | `/api/runs/{id}/families?order=composite|stability` | family rankings and unavailable families |
| GET | `/api/runs/{id}/families/{key}` | one family with members |
| GET | `/api/runs/{id}/providers` | provider table with evidence |
| GET | `/api/runs/{id}/models?status=&provider=&reason=&tier=&family=&q=&sort=&order=&page=&size=` | model table |
| GET | `/api/runs/{id}/models/{model}` | one model with its attempt history |
| GET | `/api/runs/{id}/failures` | failures grouped by reason → provider |
| GET | `/api/runs/{id}/latency`, `/load`, `/capabilities`, `/quality` | stage results |
| GET | `/api/runs/{id}/report.md?lang=zh|en&download=1` | Markdown report |
| GET | `/api/runs/{id}/export.json`, `/export.csv` | exports |
| GET | `/api/compare?a=&b=` | diff of two scan runs |

## MCP tools (`llm-speedtest mcp`)

| Tool | Arguments | Returns |
|---|---|---|
| `list_endpoints` | — | saved endpoints (keys masked) |
| `upsert_endpoint` | `id?, name, base_url, api_key?, dialect?` | endpoint |
| `check_endpoint` | `endpoint_id` or `base_url` + `api_key` | reachability, auth, models, dialect |
| `discover_models` | endpoint + `include_owners?, exclude_owners?, id_regex?, limit?` | counts, providers, models |
| `start_run` | `kind, endpoint…, profile?, models?, confirm?, note?` | `run_id` |
| `get_run` | `run_id` | status, children, progress |
| `wait_run` | `run_id, timeout_s?` | final run (polls) |
| `cancel_run` | `run_id` | — |
| `list_runs` | `kind?, limit?` | runs |
| `get_ranking` | `run_id, order?, limit?` | families |
| `get_models` | `run_id, status?, provider?, q?, page?, size?` | models |
| `get_failures` | `run_id, limit_per_group?` | grouped failures |
| `get_providers` | `run_id` | providers |
| `get_latency`, `get_load`, `get_capabilities`, `get_quality` | `run_id` | stage results |
| `export_report` | `run_id, format?, lang?, out_path?` | report text or written path |
| `compare_runs` | `a, b` | diff |

Resources: `speedtest://metrics-glossary`, `speedtest://runs/{id}/summary`. Prompt: `summarize_run(run_id)`.
