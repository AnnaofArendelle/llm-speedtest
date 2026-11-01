# Gateway dialects

A dialect tells the client how a particular gateway *wraps* upstream responses. It is deliberately limited to
structure: error-message envelopes, health/version endpoints and extra metadata fields. Dialects never carry
provider or model name tables — those are derived from data (see [providers-and-families.md](providers-and-families.md)).

| Dialect | Auto-detection evidence | Unwrapping / extras |
|---|---|---|
| `generic` | fallback | OpenAI error object (`error.message`, `error.code`) |
| `9router` | `GET /api/version` returns `{currentVersion, latestVersion}`; `/api/health`; errors like `[conn/model] [429]: …` | recursive `[status]:` unwrapping (≤3 levels), connection id captured as provider evidence, `(reset after 1h2m)` lock hints feed the cooldown |
| `oneapi` | one-api / new-api style `(request id: …)` suffixes, `/api/status` | request-id stripping |
| `openrouter` | `error.metadata`, `pricing` / `context_length` in `/models` | metadata passthrough |
| `litellm` | `/health` + LiteLLM headers | — |
| `vllm` | `/health`, `/version` | — |
| `ollama` | `/api/tags`, `/api/version` | — |

Detection scores each dialect from cheap probes made during `check`/`discover`; the winner is stored with the run so
reports stay reproducible. Override with `--dialect <name>` or the endpoint's *Dialect* field in the UI.

## Adding a dialect

1. Implement the structural hooks in `internal/dialect` (detect, unwrap, health).
2. Add real (redacted) error samples to `internal/classify/testdata/real_errors.json`.
3. Add a mockllm behaviour for the wrapper format and a test in `internal/dialect`.
4. Document the dialect here.
