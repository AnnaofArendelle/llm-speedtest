# Changelog

All notable changes to this project are documented here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and the project adheres to
[Semantic Versioning](https://semver.org/).

## [Unreleased]

## [0.1.0] - 2026-11-02

### Added
- Availability scan with multi-round probing (full → transient retry → re-sampling → final confirmation),
  lock-aware cooldowns, resumable runs and a 25-code failure taxonomy derived from real gateway errors.
- Data-driven provider identification (owned_by / id prefix / connection ids) and family grouping
  (tokeniser rules + alias learning from served model names + per-endpoint overrides) — no built-in alias tables.
- Composite family ranking (stability × provider breadth × speed) and stability ranking with A/B/C tiers.
- Streaming latency lab: TTFT, time-to-second-token, TPOT, ITL, E2E, decode tok/s (N−1), input tok/s, cold start.
- Load test: closed-loop concurrency sweeps and open-loop Poisson rate mode, goodput under SLO.
- Capability probes (streaming, usage in stream, tools, JSON mode / schema, vision, system prompt, max_tokens,
  stop, determinism, reasoning parameters, long context tiers, embeddings) and a deterministic quality spot-check.
- One-shot `suite` command, interactive web dashboard (React + MUI + ECharts, dual theme, zh/en) with live SSE
  progress, Markdown / JSON / CSV export preserving the detailed report structure, run comparison.
- Gateway dialects: generic, 9router, one-api/new-api, OpenRouter, LiteLLM, Ollama, vLLM (structural only).
- Agent Skills package (`llm-speedtest skill install`) and MCP server (`llm-speedtest mcp`).
- Demo mode (`serve --demo`) seeded from an embedded, anonymised reference scan of 2 306 models.
- GitHub Actions CI (web + Go on Linux/macOS/Windows, race detector, coverage gate, golangci-lint, govulncheck,
  CodeQL, actionlint, shellcheck, Docker build) and release automation (GoReleaser, GHCR multi-arch image).

[Unreleased]: https://github.com/AnnaofArendelle/llm-speedtest/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/AnnaofArendelle/llm-speedtest/releases/tag/v0.1.0
