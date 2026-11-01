# AI skill (Agent Skills)

The repository embeds an [Agent Skills](https://agentskills.io) package at `skills/llm-speedtest`:

```
SKILL.md                      frontmatter (name, description, compatibility, metadata) + workflow instructions
scripts/install.sh            install the llm-speedtest binary (Linux/macOS)
scripts/install.ps1           install the binary (Windows)
scripts/ensure-server.sh      start `llm-speedtest serve` if it is not running
scripts/speedtest.sh          one-shot scan/suite with JSON output for the agent
references/cli.md             CLI reference
references/api.md             HTTP API reference
references/metrics.md         metric definitions
references/interpreting-results.md  how to read rankings, tiers and failure reasons
assets/report-template.md     summary template the agent fills in
```

## Install

```sh
llm-speedtest skill install --target claude    # ~/.claude/skills/llm-speedtest
llm-speedtest skill install --target codex     # ~/.codex/skills/llm-speedtest
llm-speedtest skill install --target cursor    # ~/.cursor/skills/llm-speedtest
llm-speedtest skill install --target project   # ./.claude/skills/llm-speedtest in the current repository
llm-speedtest skill install --target /any/dir  # custom directory
llm-speedtest skill export llm-speedtest-skill.zip
```

The same zip is attached to every GitHub release. Any agent runtime that understands `SKILL.md` can load it.

## What the agent does with it

1. Checks that the binary is installed (installs it otherwise) and that the endpoint is reachable (`check --json`).
2. Previews the model list (`discover --json`) and confirms filters with the user if the list is large.
3. Runs `scan` or `suite` with `--json-out`, or starts a run through the MCP server and polls `wait_run`.
4. Reads the ranking (`get_ranking` / JSON `families`), the failure groups and the latency stats.
5. Answers using the report template: what works, what is fastest and most stable, what is broken and why, and
   what to fix; exports the Markdown report on request.

Keys are read from `LLM_SPEEDTEST_KEY` or the stored endpoint profile and are never written to disk by the skill.
Load tests require an explicit confirmation from the user before the agent may start them.
