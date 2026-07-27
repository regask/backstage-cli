---
description: Read-only Backstage status — deploy versions, env vars, secret refs, ticket deploy state.
argument-hint: "[service] [--env <env>]"
allowed-tools: >-
  Bash(bsr check-deploy:*), Bash(bsr check-environment:*),
  Bash(bsr check-secrets:*), Bash(bsr find-ticket:*), Bash(bsr whoami:*),
  Bash(backstage-regask check-deploy:*), Bash(backstage-regask check-environment:*),
  Bash(backstage-regask check-secrets:*), Bash(backstage-regask find-ticket:*),
  Bash(backstage-regask whoami:*)
---

Use the `backstage-regask` skill. Answer a read-only status question with the
`bsr` CLI: `check-deploy`, `check-environment`, `check-secrets` (masked — do NOT
pass `--reveal` unless the user explicitly asks for secret values), or
`find-ticket`, depending on what the user asked.

Prefer `--json` and summarize. Add `--fresh` if the user wants to bypass the
cache. This command is read-only — do not run any mutating command here.

Request: $ARGUMENTS
