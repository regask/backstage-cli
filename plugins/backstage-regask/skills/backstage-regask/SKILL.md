---
name: backstage-regask
description: Use when the user wants Regask Backstage deploy status, environment
  or secret lookups, ticket deploy state, approvals, or promote/release/cherry-pick
  workflows. Drives the `bsr` (backstage-regask) CLI from the terminal.
---

# Driving the Regask Backstage CLI (`bsr`)

You operate the `bsr` command-line tool on the user's behalf. Pick the right
command, run it via Bash, and summarize the result. `bsr` talks to the Regask
Backstage backend as the signed-in user.

## Prerequisites

- **Binary:** prefer `bsr`. If `bsr` is not found, use `backstage-regask` (same
  tool; `bsr` is a Homebrew symlink). Do not rely on a shell alias — the Bash
  tool is non-interactive.
- **Auth:** if any command fails with an unauthorized / 401 error, tell the user
  to run `bsr login` themselves (it opens a browser and waits up to 5 minutes).
  Never run `bsr login` yourself — it is interactive and will hang.
- Confirm the signed-in user any time it matters with `bsr whoami`.

## Output convention

- Add `--json` when you need to parse the result, then summarize it for the
  human in plain language. `--json` works on every command.
- Add `--fresh` when the user wants to bypass the server cache. `--fresh` exists
  only on data commands: `check-deploy`, `check-environment`, `check-secrets`,
  `find-ticket`.
- A non-zero exit means failure (failed task, rejected/aborted approval, no
  match). Surface it as a failure — don't claim success.

## Read-only commands (safe to run directly)

- **Deploy status:** `bsr check-deploy <service> [--env <env>]` — what version is
  deployed where (version + ArgoCD sync/health). `<service>` accepts a bare name
  (`alert-service`) or a full ref (`component:default/alert-service`).
- **Environment variables:** `bsr check-environment <service> --env <env>` —
  effective env vars for the service in that environment. `--env` is required.
- **Secret refs:** `bsr check-secrets <service> --env <env>` — secret keys and
  their vault keys, values masked (`********`). `--env` is required.
  - Only add `--reveal` when the user explicitly asks to see secret values. It
    resolves values via `az` and needs `az login`. `--vault <name>` overrides
    the per-env default vault. Never reveal secrets unprompted.
- **Ticket state:** `bsr find-ticket <TICKET> [TICKET...]` — which environments a
  ticket (or several) is deployed to.
- **Identity:** `bsr whoami` — the signed-in user.
- **Approval details:** `bsr query-approval <url-or-id>` — approval details plus
  the release link. Read-only; use this before any approve/reject decision.

## Mutating commands (CONFIRM FIRST — always)

Before running ANY of these: print the exact command you are about to run and
get the user's explicit "yes". Never auto-fire a release, promotion, approval,
or cherry-pick. These launch the existing scaffolder templates — never
reimplement git/gitops.

- **Approve / reject an approval:** the CLI prompts `[y/N]` on stdin, which you
  cannot answer interactively, so pipe the answer after the user confirms:
  - Approve: `echo y | bsr approve <url-or-id>`
  - Reject:  `echo y | bsr approve <url-or-id> --reject`
  Run `bsr query-approval <url-or-id>` first and show the details before asking.
- **Promote:** `bsr promote --to-env <env> --service <a,b>` — one or more
  services (repeatable or comma-separated). Creates a draft release +
  release-publish approval.
- **Release an environment:** `bsr release --env <env>` (all services). Narrow
  with `--include-services <a,b>` OR `--exclude-services <a,b>` (mutually
  exclusive). `--version <v>` overrides the shared-actions release version (prod
  only).
- **Cherry-pick:** `bsr cherry-pick --tag <TICKET> --branch <branch>` where
  `<branch>` is `release/preprod` or `release/prod` (the only accepted values).

promote / release / cherry-pick stream task logs and do not prompt — run them
directly after the user confirms, and report the final task status.

## Maintenance

- **Update the CLI:** `bsr update` (runs `brew update` + `brew upgrade`). Safe to
  run; report the outcome.
