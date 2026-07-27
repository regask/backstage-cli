---
description: Read-only Backstage status — deploy versions, env vars, secret refs, ticket deploy state.
argument-hint: "[service] [--env <env>]"
allowed-tools: >-
  Bash(backstage-regask check-deploy:*), Bash(backstage-regask check-environment:*),
  Bash(backstage-regask check-secrets:*), Bash(backstage-regask find-ticket:*),
  Bash(backstage-regask whoami:*)
---

Use the `backstage-regask` skill. Answer a read-only status question with the
`backstage-regask` CLI: `check-deploy`, `check-environment`, `check-secrets`
(masked — do NOT pass `--reveal` unless the user explicitly asks for secret
values), or `find-ticket`, depending on what the user asked.

Translate the user's wording into the CLI's vocabulary before running anything —
see the skill's Vocabulary section:

- Environments are only `development`, `staging`, `pre-prod`, `production`.
  "prod" means `production`, "dev" means `development`. An unrecognized `--env`
  makes `check-deploy` return an empty list with exit 0, which is NOT the same
  as "nothing deployed there".
- Services use bare catalog names with no `-service` suffix (`alert`, not
  `alert-service`). Matching is exact — if a name doesn't resolve, try the bare
  repo name once, then ask the user instead of guessing.

For a "what's deployed" question, `check-deploy <service>` without `--env` is the
safest start: it returns all four environments, so you can answer about the one
asked and give the rest as context.

Prefer `--json` and summarize. Add `--fresh` if the user wants to bypass the
cache. This command is read-only — do not run any mutating command here.

Request: $ARGUMENTS
