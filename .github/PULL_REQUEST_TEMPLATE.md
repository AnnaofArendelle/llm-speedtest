## Summary

<!-- What does this change and why? -->

## Checklist

- [ ] `./scripts/ci-check.sh` passes locally (Go vet/lint/tests, web lint/typecheck/tests/build)
- [ ] New behaviour is covered by tests (unit, mockllm integration or golden files)
- [ ] Docs / README / CHANGELOG updated when user-visible behaviour changes
- [ ] No provider or model alias tables were hard-coded (identification must stay data-driven)
