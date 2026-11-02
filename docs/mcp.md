# MCP server

`llm-speedtest mcp` runs a [Model Context Protocol](https://modelcontextprotocol.io) server over stdio using the
official Go SDK. It shares the data directory with `serve`, so runs started from the dashboard are visible to agents
and vice versa.

## Client configuration

```json
{
  "mcpServers": {
    "llm-speedtest": {
      "command": "llm-speedtest",
      "args": ["mcp", "--data-dir", "/path/to/data"]
    }
  }
}
```

## Tools

| Tool | Purpose |
|---|---|
| `list_endpoints`, `upsert_endpoint`, `check_endpoint`, `discover_models` | endpoint management and preview |
| `start_run` | start `suite` / `scan` / `latency` / `load` / `capability` / `quality` (load requires `confirm: true`) |
| `get_run`, `wait_run`, `cancel_run`, `skip_model`, `list_runs` | lifecycle (`skip_model` cancels one model's in-flight request — or all with `all: true` — and excludes it from the rest of the run) |
| `get_ranking`, `get_models`, `get_failures`, `get_providers` | scan results |
| `get_latency`, `get_load`, `get_capabilities`, `get_quality` | stage results |
| `export_report` | Markdown report (zh/en) |
| `compare_runs` | diff two scan/suite runs |

## Resources and prompts

- `speedtest://metrics-glossary` — metric definitions
- `speedtest://skill` — the SKILL.md workflow
- `speedtest://runs/{id}/summary` — summary of a run
- prompt `summarize_run` — asks the model to summarise a run using the report template

All results are JSON with the same field names as the HTTP API (see [api.md](api.md)).
