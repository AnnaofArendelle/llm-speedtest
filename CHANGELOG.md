# Changelog

All notable changes to this project are documented here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and the project adheres to
[Semantic Versioning](https://semver.org/).

## [Unreleased]

## [0.2.0] - 2026-11-02

### Added
- **Skip while running**: the progress card lists every request in flight with its elapsed time; one click
  cancels that model's request and leaves it out of the remaining rounds and stages (`POST /api/runs/{id}/skip`,
  MCP tool `skip_model`, new failure reason `skipped`). SSE now carries `start` events and the snapshot includes
  the in-flight list, so a reloaded page shows what is running.
- **Capability matrix**: model and family drawers open with a radar of eight 0–100 dimensions (stability, speed,
  provider breadth, TTFT, throughput, capability coverage, quality, load goodput) plus a coverage row that says which
  stages measured the model (`GET /api/runs/{id}/profile/model/{model}`, `/profile/family/{key}`).
- Route-level error page and per-section error boundaries; `partial` flag on the summary endpoint with a dedicated
  “scan in progress” state; toasts for run start, cancel and skip; demo mode now also seeds a failed and a cancelled run.
- Frontend test suite: API fixtures recorded by the Go contract test (`make fixtures`), a render matrix that mounts
  every page and tab against every run state, SSE/progress/drawer/page interaction tests, i18n completeness and a
  radius guard, coverage thresholds, and a Playwright end-to-end job (page sweep + live skip flow + failed run).

### Fixed
- Run dashboard crashed (`top_models is null`) while a run was discovering/running or after it failed: Go nil
  slices/maps were serialised as `null`. All collection fields are now always arrays/objects (enforced by
  `internal/server/contract_test.go`) and the TypeScript types are nullable so unguarded access fails type-checking.
- Over-sized corners (60 px) on tables, accordions, dialogs and drawers: numeric `borderRadius` values in `sx` are
  MUI multipliers; the theme now exposes `radius` tokens and a test rejects numeric radii.
- Results computed while a run was in progress could be served as the final result for up to 30 s after completion.
- SSE `done` detection for single-stage runs; stage completion of suites no longer ends the live view early.
- `skill install --target <dir>` ignored flags placed after the sub-command.

### Changed
- Material 3 polish: tonal DataGrid headers, hairline rows with tinted hover, staggered KPI entrance, skeleton
  loaders, illustrated empty states, M3 accordion/dialog/drawer radii.

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

[Unreleased]: https://github.com/AnnaofArendelle/llm-speedtest/compare/v0.2.0...HEAD
[0.2.0]: https://github.com/AnnaofArendelle/llm-speedtest/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/AnnaofArendelle/llm-speedtest/releases/tag/v0.1.0
