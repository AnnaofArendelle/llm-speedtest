# LLM API Speed Test Report

- Endpoint: `http://127.0.0.1:46183/v1`
- Generated: 2026-11-02 21:32:27

## Latency lab

Streaming time-to-first-token (TTFT), time-to-second-token (TTST), time per output token (TPOT), end-to-end latency (E2E) and decode throughput (tok/s, N−1 definition). Statistics exclude warm-up samples; the cold column is the first warm-up sample.

| # | Model | Samples (ok) | TTFT p50 | TTFT p90 | TTST p50 | TPOT p50 | E2E p50 | tok/s p50 | tok/s p90 | Cold TTFT | Output tokens | Notes |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| 1 | `a/fast` | 1(1) | 13 ms | 13 ms | 2 ms | 2 ms | 18 ms | 432.1 | 432.1 | - | 3 |  |

