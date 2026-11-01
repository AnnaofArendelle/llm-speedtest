# Configuration

Precedence: **command-line flags → environment variables → config file → built-in profile defaults**.

```sh
llm-speedtest config example > speedtest.yaml
llm-speedtest suite --config speedtest.yaml --profile deep --out report.md
```

Environment: `LLM_SPEEDTEST_KEY` (API key), `LLM_SPEEDTEST_DATA` (data directory, default `./data`).

## Profiles

| Profile | Scan rounds | Latency samples | Load | Use |
|---|---|---|---|---|
| `quick` | 1 full round, no re-sampling | few | off | first look, CI smoke tests |
| `standard` | first → retry → 2 resamples → final | 5 (+1 warm-up) | off | reports |
| `deep` | more resamples, longer timeouts, more long-context tiers | more | off | audits |

## File reference

See [`examples/speedtest.yaml`](../examples/speedtest.yaml) for the complete annotated file. Key sections:

- `endpoint` — `name`, `base_url` (must end with `/v1`), `api_key`, `dialect`, `headers`, `insecure`, `proxy`.
- `suite` — which stages run (`scan`, `latency`, `capability`, `quality`, `load`) and the top-N taken from the
  ranking for each stage; `filter` (`exclude_owners`, `include_owners`, `id_regex`, `exclude_regex`, `limit`,
  `models`); `lang`.
- `suite.specs.scan` — per-round concurrency / per-provider concurrency / timeout, `resample`, `max_tokens`,
  `prompt`, `max_cooldown_wait`, `max_backoff`.
- `suite.specs.latency` — `samples`, `warmup`, `input_tokens`, `output_tokens`, `concurrency`, `per_provider`,
  `timeout`, `stall_timeout`, `temperature`, `seed`.
- `suite.specs.load` — `mode` (`concurrency` | `rate`), `levels`, `rate`, `duration`, `requests`, `ramp_up`,
  `input_tokens`, `output_tokens`, `stream`, `max_inflight`, `slo` (`ttft`, `tpot`, `e2e`).
- `suite.specs.capability` — `probes`, `long_context_tiers`, `concurrency`, `timeout`.
- `suite.specs.quality` — `trials`, `questions` (path to a YAML set), `max_tokens`.
- `suite.specs.scoring` — `w_stability`, `w_breadth`, `w_speed`, `speed_lo_s`, `speed_hi_s`, tier thresholds.

Durations use Go syntax (`90s`, `2m`). Every spec field has a flag equivalent on the `suite`/`scan`/… commands
(`llm-speedtest scan -h`).
