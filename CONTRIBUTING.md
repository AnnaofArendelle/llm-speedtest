# Contributing

Thanks for helping make LLM benchmarking more honest. This document covers the local workflow and the few rules
that keep the project coherent.

## Development setup

```sh
git clone https://github.com/AnnaofArendelle/llm-speedtest
cd llm-speedtest
make web-install   # npm ci (Node ≥ 20)
make web           # build the dashboard into web/dist
make build         # bin/llm-speedtest (Go ≥ 1.26, GOTOOLCHAIN=auto downloads it)
./bin/llm-speedtest serve --demo
```

`make check` (or `./scripts/ci-check.sh`) runs exactly what CI runs: web lint/format/typecheck/tests/build,
`go vet`, golangci-lint, tests with the race detector and the coverage gate, govulncheck, shellcheck,
actionlint and the skill package validation.

Useful targets: `make test`, `make race`, `make cover`, `make lint`, `make snapshot` (GoReleaser dry run),
`make screenshots` (regenerates `docs/screenshots` from `serve --demo`).

## Ground rules

1. **No hard-coded provider or alias tables.** Provider names and model families must be inferred from the
   endpoint's own data (`owned_by`, id prefixes, connection ids, served model names) plus user overrides.
   If a heuristic is wrong, fix the heuristic or add evidence — do not add a lookup table.
2. **Every failure class is backed by a real sample.** New classifier rules go together with an entry in
   `internal/classify/testdata/real_errors.json` (redacted) so parity stays testable.
3. **Metrics follow the documented definitions** in `docs/metrics.md`. Changing a definition is a breaking
   change for reports and needs a CHANGELOG entry.
4. **Tests over screenshots.** Engines are tested against `internal/testutil/mockllm`; the HTTP API has
   end-to-end tests; the report has golden files (`go test ./internal/report -update`).
5. **Conventional Commits** (`feat:`, `fix:`, `docs:`, `ci:`, `refactor:`, `test:`) — they drive the release notes.

## Pull requests

- Keep PRs focused; describe the behaviour change and how you verified it.
- Run `make check` before pushing. CI must be green on Linux, macOS and Windows.
- Update `docs/` and the READMEs when user-visible behaviour changes.

## Reporting security issues

See [SECURITY.md](SECURITY.md).
