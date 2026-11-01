# Security policy

## Supported versions

Only the latest minor release receives security fixes.

## Reporting a vulnerability

Please **do not** open a public issue for security problems. Use GitHub's private vulnerability reporting
("Security" tab → "Report a vulnerability") on this repository. You will get an acknowledgement within a few
days and a fix or mitigation plan as soon as the report is confirmed.

## Design notes

- API keys are stored AES-256-GCM encrypted in the local SQLite database; the master key lives in
  `<data-dir>/.secret` (mode 0600). Keys are never returned by the API or shown in the UI/logs (only a hint).
- The HTTP API binds to `127.0.0.1` by default. When exposing it on a network, start the server with
  `--access-token` and put it behind TLS.
- Load tests require explicit confirmation (`--confirm`, `confirm: true` in the API/MCP) because they send
  sustained traffic to a third-party endpoint.
- The AI skill reads keys from environment variables only and never writes them to disk.
