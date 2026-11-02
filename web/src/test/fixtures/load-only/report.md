# LLM API Speed Test Report

- Endpoint: `http://127.0.0.1:46183/v1`
- Generated: 2026-11-02 21:32:27

## Load test

Model: `a/fast`

| Level | Mode | Requests | OK | Errors | Req/s | Out tok/s | TTFT p50 | TTFT p99 | E2E p50 | E2E p99 | ITL p50 | Goodput |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| 1 | concurrency | 2 | 2 | 0 | 55.74 | 167.2 | 13 ms | 13 ms | 18 ms | 18 ms | 2 ms | 100% |

