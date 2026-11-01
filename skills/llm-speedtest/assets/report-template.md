# {{endpoint}} — speed test summary ({{date}})

**Headline:** {{available}} of {{models}} models available ({{available_pct}}), {{families_ok}} of {{families}} families have a working member. Median latency {{median}}; {{provider_level_pct}} of failures are account-level (credentials / balance / quota / plan).

## Recommended

| # | Family | Score | Stability | Providers | Median | Best model @ provider |
|---|---|---|---|---|---|---|
| 1 | … | … | A … | … | … | `…` @ … |
| 2 | … | … | … | … | … | … |
| 3 | … | … | … | … | … | … |

## What is broken and why

- **{{provider}}** — {{reason}} ({{count}} models): {{one-line explanation and fix}}
- …

## Notes

- {{caveats: estimated tokens, single samples, proxy overhead, excluded models}}

Full report: `{{report_path}}` · JSON: `{{json_path}}` · Dashboard: `llm-speedtest serve` → http://127.0.0.1:8787
