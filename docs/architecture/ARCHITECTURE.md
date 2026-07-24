# Architecture — `backstage-regask` (`bsr`)

This document records the **architecture and the decisions behind it**. It is the
source of truth for *why* the CLI is shaped the way it is. Any change that alters
a decision here MUST update this document in the same change (see
[CLAUDE.md](../../CLAUDE.md)).

## What this is

`backstage-regask` is a standalone Go CLI that lets engineers do everyday Regask
Backstage work from the terminal: check deploy status, inspect a service's
environment variables and secret refs, look up tickets, view/approve approval
requests, and launch the promote / release / cherry-pick scaffolder workflows —
authenticated as the real Backstage user.

**Naming:** the repo/folder and Go module are `backstage-cli`
(`github.com/regask/backstage-cli`), following the Go convention of module name
= directory. The compiled **binary and Homebrew install stay `backstage-regask`**
(the Homebrew formula also installs a `bsr` symlink) — the product identity is
`backstage-regask`, the code lives under `backstage-cli`.

## Key decisions

| # | Decision | Rationale |
|---|----------|-----------|
| D1 | **Standalone repo** (not a package in the backstage monorepo) | Wider distribution to people who don't clone the monorepo. |
| D2 | **Go + cobra**, single static binary | User wants a true standalone binary; Go gives the smallest binary, trivial cross-compile, and the cleanest `brew install` / `brew upgrade`. |
| D3 | **Browser OAuth loopback login** (`az login`-style) | The backend gates every route on `httpAuth.credentials(req, {allow:['user']})`, so only a real per-user token works. Rejected static `externalAccess` tokens (shared/opaque, bypasses per-user auth). |
| D4 | **Vendored contracts** (`internal/contracts`) instead of importing `deploy-management-common` | A standalone Go repo can't import the TS backend package; the JSON structs + the pure parsers are small. Parsers must keep **parity** with the backend — see D7. |
| D5 | **Mutating commands launch existing scaffolder templates** | `promote`/`release`/`cherry-pick` already exist as templates with approval gates; reuse them via the scaffolder API, never reimplement git/gitops. |
| D6 | **Secrets fetched on demand via local `az`, masked by default** | The gitops overlay holds only secret *references* (Key Vault key names), never values. Values are resolved with the user's `az` session, printed only with `--reveal`, and never persisted. |
| D7 | **No new backend code** | Every command maps to an existing route or scaffolder template. |
| D8 | **Distribution: GoReleaser + release-please on merge to `main`** | Merge to main computes semver, tags, cross-compiles, and updates the Homebrew tap automatically. |

## Package layout

```
cmd/                 cobra commands (thin: parse args → client → render)
internal/
  auth/              loopback OAuth login + config/token store (~/.config/backstage-regask/config.json, 0600)
  client/            typed HTTP client; injects bearer token; ?refresh=1 on --fresh; 401 → ErrUnauthorized
  contracts/         Go structs mirroring backend JSON + ports of parseEnvFile/mergeMaps/parseSecretRefs (parity with deploy-management-common)
  az/                injectable wrapper around `az keyvault secret show`
  scaffolder/        launch a template task + stream /tasks/:id/events
  render/            --json vs table output helper
```

Design intent: thin commands, focused single-responsibility packages, all
dependencies (HTTP transport, token store, `az`, browser-open) injectable so
commands are unit-tested without side effects.

## Auth flow

`bsr login` → ephemeral `127.0.0.1:0` loopback server → open browser to the
portal's **`/cli-auth`** handshake page with the loopback as `redirect_uri` plus
a random CSRF `state` → the page (a normal authenticated app route, so it forces
sign-in first if needed) reads the signed-in user's Backstage identity token via
`identityApi.getCredentials()` and redirects back to the loopback with
`token` + `state` → the CLI verifies `state`, captures the token, and persists to
`~/.config/backstage-regask/config.json` (mode 0600). A `401` on any command
tells the user to run `bsr login`.

The `/cli-auth` page lives in the Regask portal (`packages/app/src/modules/cli-auth`
in the backstage repo). The loopback only accepts `http://127.0.0.1|localhost`
redirect targets, so a token is never handed off-machine.

## Command → backend map

| Command | Backend |
|---|---|
| `check-deploy <svc> [--env]` | `GET /deploy-management/matrix`, `/service/releases` |
| `check-environment <svc> --env` | effective env vars from `/deploy-management/service/overlays` (base ⊕ overlay) |
| `check-secrets <svc> --env [--reveal] [--vault]` | secret refs from overlay → local `az keyvault secret show` (masked by default) |
| `find-ticket <TICKET...>` | `POST /deploy-management/ticket-lookup` |
| `query-approval <link-or-id>` | `GET /approvals/requests/:id` — detail + release link (`resultUrl`) + task backlink |
| `approve <link-or-id> [--reject]` | `GET` detail, confirm, `POST .../approve`\|`/reject` |
| `promote --to-env <env> --service <svc…>` / `release --env [--include-services\|--exclude-services]` / `cherry-pick --tag <TICKET> --branch <release/preprod\|release/prod>` | launch scaffolder templates (`regask:github:promote`, `release:*`, `cherry-pick:*`), stream log. promote takes one+ required services; release's include/exclude are mutually exclusive; cherry-pick's `--tag` is a ticket ref (e.g. REG-12345) and `--branch` is restricted to `release/preprod`/`release/prod` |

Cross-cutting: `--json` on every command; `--fresh` bypasses the server TTL
cache; non-zero exit on failure / rejected approval / failed task.

## Deferred backend contracts (confirm against the live backend)

These use documented defaults and are marked `TODO(execution)` in code until
confirmed:

1. ~~Portal CLI login start path~~ — **resolved:** `/cli-auth` page added to the
   portal; CLI sends a CSRF `state` and verifies it on the callback.
2. The `whoami` identity endpoint.
3. Scaffolder task/events endpoints + each template's `templateRef` and parameter names.
4. The runs/task backlink route used in the approval detail view.

Remaining login hardening (login timeout on a stalled browser flow) is still open.

## Distribution

Push to `main` → release-please computes semver + tags + GitHub Release →
GoReleaser cross-compiles (macOS arm64/x64, Linux) and pushes the formula to
`regask/homebrew-tap`. Users install with `brew install regask/tap/backstage-regask`
and update with `brew upgrade`. Requires the `HOMEBREW_TAP_TOKEN` CI secret.
