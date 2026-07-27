---
description: Run a Backstage promote / release / cherry-pick workflow (with confirmation).
argument-hint: "[promote|release|cherry-pick] ..."
---

Use the `backstage-regask` skill. These are mutating workflows that launch
scaffolder templates:

- `bsr promote --to-env <env> --service <a,b>`
- `bsr release --env <env>` (optionally `--include-services` XOR `--exclude-services`, `--version`)
- `bsr cherry-pick --tag <TICKET> --branch <release/preprod|release/prod>`

Gather the required arguments, then follow the skill's safety rule: print the
exact command and get an explicit "yes" before running it. Report the final task
status. Never reimplement git/gitops — only launch these templates.

Request: $ARGUMENTS
