---
description: Run a Backstage promote / release / cherry-pick workflow (with confirmation).
argument-hint: "[promote|release|cherry-pick] ..."
---

Use the `backstage-regask` skill. These are mutating workflows that launch
scaffolder templates:

- `backstage-regask promote --to-env <env> --service <a,b>`
- `backstage-regask release --env <pre-prod|production>` (optionally `--exclude-services <a,b>`, `--version <n>` for production)
- `backstage-regask cherry-pick --tag <TICKET[,TICKET]> --branch <release/preprod|release/prod>`

Resolve the environment and service names BEFORE you print a command to confirm,
so the confirmation shows exactly what will run:

- Environments are exactly `development`, `staging`, `pre-prod`, `production`
  (never `prod`/`dev`/`stg`). `cherry-pick --branch` instead takes
  `release/preprod` or `release/prod`.
- Service names are bare catalog names with no `-service` suffix (`alert`,
  `task-v2`). Confirm each one resolves with a read-only
  `backstage-regask check-deploy <name>` first — that failing is cheap, a
  mis-targeted release is not.

Gather the required arguments, then follow the skill's safety rule: print the
exact command and get an explicit "yes" before running it. Report the final task
status. Never reimplement git/gitops — only launch these templates.

Request: $ARGUMENTS
