---
name: backstage-regask
description: Use when the user wants Regask Backstage deploy status, environment
  or secret lookups, ticket deploy state, approvals, or promote/release/cherry-pick
  workflows. Drives the `backstage-regask` CLI from the terminal.
---

# Driving the Regask Backstage CLI (`backstage-regask`)

You operate the `backstage-regask` command-line tool on the user's behalf. Pick
the right command, run it via Bash, and summarize the result. It talks to the
Regask Backstage backend as the signed-in user.

## Prerequisites

- **Binary: always invoke `backstage-regask`.** That is the name Homebrew
  installs. Never invoke `bsr` — it is only a shell alias the user may have
  added, and aliases don't resolve in the non-interactive Bash tool. Note the
  CLI's own error text hardcodes `bsr login` — rewrite that to
  `backstage-regask login` when relaying it, or the user hits
  `command not found`.
- **Auth:** `whoami` decodes the locally cached JWT and nothing more — it never
  calls the server and never checks expiry, so **it succeeds with an expired
  token and is not an auth check.** To confirm the session really works, run a
  read command (`check-deploy <known-service>`). If any command fails with
  unauthorized / 401, tell the user to run `login` themselves (it opens a
  browser and waits up to 5 minutes). Never run `login` yourself — it is
  interactive and will hang.

## Vocabulary (get this right before running anything)

Environment and service names are the two biggest sources of confidently wrong
answers. Neither the CLI nor the backend normalizes them.

### Environments — exactly four, spelled out

`development` · `staging` · `pre-prod` · `production`

There are **no aliases**: `prod`, `dev`, `stg`, `preprod` are all invalid.
(Source of truth: `ENVIRONMENTS` in the backend's
`plugins/deploy-management-backend/src/types.ts`.) The promotion chain runs
`development → staging → pre-prod → production`.

When the user says "prod", they mean `production` — pass the canonical name.
A wrong env name fails differently per command, and one of them fails silently:

| Command | A wrong `--env` gives |
|---|---|
| `check-deploy` | **empty result `[]`, exit 0** — indistinguishable from "not deployed". Never report an empty `check-deploy` as "nothing deployed there"; re-check the spelling, or drop `--env` and read the whole matrix. |
| `check-environment` | errors: `no overlay for env "prod"` |
| `check-secrets` | masked run succeeds with no vault; `--reveal` errors |
| `promote --to-env`, `release --env` | **nothing validates it.** The string goes straight to the scaffolder template, so a typo launches a real task with a bad parameter. Only ever pass a canonical name. |

`cherry-pick --branch` takes git branch names, not env names: `release/preprod`
or `release/prod` are the only accepted values.

### Service names — bare catalog names, no `-service` suffix

Services are named after their repo — `alert`, `notification` — **not**
`alert-service`. Accepted forms: the bare name (`alert`), the full entity ref
(`component:default/alert`), or `default/alert`. Matching is exact name, exact
ref, or a `/name` suffix — there is no fuzzy or substring matching, so a
near-miss simply errors.

Two related traps:

- The not-found hint blindly prepends the prefix, so passing a full ref
  suggests the nonsense `component:default/component:default/alert`. Ignore it.
- Any argument containing `/` is passed through **unvalidated** — on mutating
  commands a typo'd full ref reaches the scaffolder template as-is.

#### Naming patterns

~61 services live in the catalog. The suffix encodes the *kind* of workload, and
none of them is the word "service":

| Shape | Kind (`type`) | Examples |
|---|---|---|
| bare domain noun | `services` | `alert`, `comment`, `company`, `identity`, `notification`, `report`, `suggestion` |
| `-v2` suffix | `services` | `document-v2`, `task-v2` — the domain name alone does **not** resolve |
| `-event-handler` | `event-handlers` | `alert-event-handler`, `comment-event-handler`, `identity-event-handler` |
| `-cronjob` (or a bare verb phrase) | `cronjobs` | `suggestion-cronjob`, `chunk-documents`, `delete-orphan-documents` |
| `-integration` | `kservices` | `veeva-integration`, `esko-integration` |
| `usvc-` prefix | `services` / `migrations` | `usvc-source`, `usvc-tag`, `usvc-node-app`, `usvc-migrations` |
| web/app names | `frontends`, `gateways` | `web-frontend`, `regulations-web-app`, `storybook`, `web-api-gateway` |

Mapping what a user says to a real name:

- Drop any `-service` / `-svc` the user added: "alert-service" → `alert`.
- A domain word may name several distinct services. "alert" alone matches the
  `alert` service exactly, but `alert-chronology`, `alert-vectorized`,
  `alert-event-handler`, and `alert-integration-start` are separate deployables —
  if the question is about behaviour rather than a version, confirm which one.
- For a domain with a `-v2`, prefer the `-v2` name (`task-v2`, `document-v2`).

#### Enumerating services

There is **no list-services command** — `check-deploy` requires a name and
matches exactly. When a name won't resolve, don't brute-force guesses; read the
whole matrix (read-only, same endpoint `check-deploy` uses, authenticated with
the already-stored token):

```bash
TOK=$(python3 -c "import json,os;print(json.load(open(os.path.expanduser('~/.config/backstage-regask/config.json')))['token'])")
curl -s -H "Authorization: Bearer $TOK" \
  https://backstage.regask.com/api/deploy-management/matrix |
  python3 -c "import json,sys;[print(r['type'],r['serviceName']) for r in json.load(sys.stdin)]"
```

Grep that for the user's word, pick the match, then go back to the normal
commands. If several plausibly fit, ask rather than guessing.

## Output convention

- Add `--json` when you need to parse the result, then summarize it for the
  human in plain language. `--json` works on every command.
- Add `--fresh` when the user wants to bypass the server cache. `--fresh` exists
  only on data commands: `check-deploy`, `check-environment`, `check-secrets`,
  `find-ticket`.
- A non-zero exit means failure (failed task, rejected/aborted approval, no
  match). Surface it as a failure — don't claim success.

## Read-only commands (safe to run directly)

- **Deploy status:** `backstage-regask check-deploy <service> [--env <env>]` —
  what version is deployed where (version + ArgoCD sync/health). Omit `--env` to
  see all four environments at once; that is usually the better opening move,
  since it also shows you the exact env spellings.
- **Environment variables:**
  `backstage-regask check-environment <service> --env <env>` — effective env vars
  for the service in that environment. `--env` is required.
- **Secret refs:** `backstage-regask check-secrets <service> --env <env>` —
  secret keys and their vault keys, values masked (`********`). `--env` is
  required.
  - Only add `--reveal` when the user explicitly asks to see secret values. It
    resolves values via `az` and needs `az login`. `--vault <name>` overrides
    the per-env default vault. Never reveal secrets unprompted.
- **Ticket state:** `backstage-regask find-ticket <TICKET> [TICKET...]` — which
  environments a ticket (or several) is deployed to.
- **Identity:** `backstage-regask whoami` — the cached user (see the auth caveat
  above; this does not prove the token is still valid).
- **Approval details:** `backstage-regask query-approval <url-or-id>` — approval
  details plus the release link. Read-only; use this before any approve/reject
  decision.

## Mutating commands (CONFIRM FIRST — always)

Before running ANY of these: print the exact command you are about to run and
get the user's explicit "yes". Never auto-fire a release, promotion, approval,
or cherry-pick. These launch the existing scaffolder templates — never
reimplement git/gitops.

- **Approve / reject an approval:** the CLI prompts `[y/N]` on stdin, which you
  cannot answer interactively, so pipe the answer after the user confirms:
  - Approve: `echo y | backstage-regask approve <url-or-id>`
  - Reject:  `echo y | backstage-regask approve <url-or-id> --reject`
  Run `backstage-regask query-approval <url-or-id>` first and show the details
  before asking.
- **Promote:** `backstage-regask promote --to-env <env> --service <a,b>` — one or
  more services (repeatable or comma-separated). Creates a draft release +
  release-publish approval.
- **Release an environment:** `backstage-regask release --env <env>` (all
  services). Narrow with `--include-services <a,b>` OR
  `--exclude-services <a,b>` (mutually exclusive). `--version <v>` overrides the
  shared-actions release version (`production` only).
- **Cherry-pick:** `backstage-regask cherry-pick --tag <TICKET> --branch <branch>`
  where `<branch>` is `release/preprod` or `release/prod` (the only accepted
  values).

promote / release / cherry-pick stream task logs and do not prompt — run them
directly after the user confirms, and report the final task status.

## Maintenance

- **Update the CLI:** `backstage-regask update` (runs `brew update` +
  `brew upgrade`). Safe to run; report the outcome.
