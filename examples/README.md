# Examples

| File | Purpose |
|---|---|
| `speedtest.yaml` | Full configuration file (output of `llm-speedtest config example`). Pass it with `--config`; flags override file values. |

Provider and family overrides are stored **per endpoint** in the results database rather than in a static file,
so that identification stays data-driven: edit them in the web UI (*Endpoints → Providers & families*, or inline in a
run's *Providers* tab) or through the HTTP API (`PUT /api/endpoints/{id}/overrides/provider|family`).
