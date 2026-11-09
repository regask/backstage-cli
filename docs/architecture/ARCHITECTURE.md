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
— the product identity is `backstage-regask`, the code lives under
`backstage-cli`. The short `bsr` is a shell alias
(`alias bsr=backstage-regask`), since the GoReleaser Homebrew **cask** links the
binary by its own name rather than adding a second symlink.

## Key decisions

| # | Decision | Rationale |
|---|----------|-----------|
| D1 | **Standalone repo** (not a package in the backstage monorepo) | Wider distribution to people who don't clone the monorepo. |
| D2 | **Go + cobra**, single static binary | User wants a true standalone binary; Go gives the smallest binary, trivial cross-compile, and the cleanest `brew install` / `brew upgrade`. |
| D3 | **Browser OAuth loopback login** (`az login`-style) via Backstage's built-in OIDC provider, with refresh tokens | The backend gates every route on `httpAuth.credentials(req, {allow:['user']})`, so only a real per-user token works. Rejected static `externalAccess` tokens (shared/opaque, bypasses per-user auth). The OIDC provider (auth-backend config only, no new backend code) issues rotating refresh tokens so users don't re-login every `backstageTokenExpiration` (8h). Falls back to the `/cli-auth` page while the portal hasn't enabled it. |
| D4 | **Vendored contracts** (`internal/contracts`) instead of importing `deploy-management-common` | A standalone Go repo can't import the TS backend package; the JSON structs + the pure parsers are small. Parsers must keep **parity** with the backend — see D7. |
| D5 | **Mutating commands launch existing scaffolder templates** | `promote`/`release`/`cherry-pick` already exist as templates with approval gates; reuse them via the scaffolder API, never reimplement git/gitops. |
| D6 | **Secrets fetched on demand via local `az`, masked by default** | The gitops overlay holds only secret *references* (Key Vault key names), never values. Values are resolved with the user's `az` session, printed only with `--reveal`, and never persisted. |
| D7 | **No new backend code** | Every command maps to an existing route or scaffolder template. |
| D8 | **Distribution: GoReleaser + release-please on merge to `main`** | Merge to main computes semver, tags, cross-compiles, and updates the Homebrew tap automatically. |

## Package layout

```
cmd/                 cobra commands (thin: parse args → client → render)
internal/
  auth/              loopback OIDC login (+ /cli-auth fallback), token refresh, config/token store (~/.config/backstage-regask/config.json, 0600)
  client/            typed HTTP client; injects bearer token; ?refresh=1 on --fresh; 401 → ErrUnauthorized
  contracts/         Go structs mirroring backend JSON + ports of parseEnvFile/mergeMaps/parseSecretRefs (parity with deploy-management-common)
  az/                injectable wrapper around `az keyvault secret show`
  scaffolder/        launch a template task + stream /tasks/:id/events; promote-code params
  render/            --json vs table output helper
```

### TUI (`bsr ui`)

The interactive terminal UI is built with **Bubble Tea** (event-driven TUI framework),
**Lipgloss** (styling), and **Bubbles** (reusable components).

**Architecture:** `internal/tui` holds the root App model, global state, messages,
and data-fetching / action cmds. `internal/tui/ui` is a leaf package containing
`Theme` and `Keys` (to break the `tui` ↔ `views` import cycle). `internal/tui/views`
holds the presentational view models for services and approvals. `cmd/ui.go` launches
the app.

**Reuse:** The TUI reuses `internal/client` (`Matrix`, `GetApproval`/`list`,
`DecideApproval`) and `internal/scaffolder` (`Launch`) — no backend changes.

**V1 views:**
- **Services**: deploy matrix (version, sync/health per environment), filterable.
- **Approvals**: list with filter, drill-down to detail + release link + task backlink.

**V2 deferred:** environment/secrets/tickets views, global search, mouse support.

Design intent: thin commands, focused single-responsibility packages, all
dependencies (HTTP transport, token store, `az`, browser-open) injectable so
commands are unit-tested without side effects.

## Auth flow

`bsr login` first checks whether the portal's auth-backend exposes its
built-in OIDC provider with refresh tokens: `GET /api/auth/.well-known/oauth-client/cli.json`
returns 200 (needs `auth.experimentalClientIdMetadataDocuments.enabled`) and
`/api/auth/.well-known/openid-configuration` lists the `refresh_token` grant
(needs `auth.experimentalRefreshToken.enabled`).

**OIDC (preferred).** Authorization code + PKCE (S256) as the public client
the backend publishes at that `cli.json` URL. Ephemeral `127.0.0.1:0`
loopback (the backend ignores the port for loopback redirects) → browser to
`/api/auth/v1/authorize` with scope `openid offline_access` and a CSRF
`state` → the portal's consent page (`/oauth2/authorize/:id`, from
`@backstage/plugin-auth`) → loopback receives `code` + `state` → CLI verifies
`state` and exchanges the code at `/api/auth/v1/token`. The access token
(~1h), refresh token and expiry are persisted to
`~/.config/backstage-regask/config.json` (mode 0600).

**Refresh.** Before each request the token source refreshes a token within a
minute of expiry (`grant_type=refresh_token`). Backstage rotates refresh
tokens (single-use), so the refresh runs under an exclusive `flock` on
`config.lock`, re-reads the config after acquiring it (another process may
have rotated already), and saves atomically (temp file + rename). A rejected
refresh (expired/revoked, 30-day default lifetime) means `bsr login` again.

**Fallback: `/cli-auth`.** If OIDC isn't enabled, the browser opens the
portal's **`/cli-auth`** handshake page with the loopback as `redirect_uri`
plus a random CSRF `state` → the page (a normal authenticated app route, so it
forces sign-in first if needed) reads the signed-in user's Backstage identity
token via `identityApi.getCredentials()` and redirects back with `token` +
`state`. This token can't be refreshed and expires after
`backstageTokenExpiration`. The page lives in the Regask portal
(`packages/app/src/modules/cli-auth` in the backstage repo) and only accepts
`http://127.0.0.1|localhost` redirect targets, so a token is never handed
off-machine.

A `401` on any command tells the user to run `bsr login`.

## Command → backend map

| Command | Backend |
|---|---|
| `check-deploy <svc> [--env]` | `GET /deploy-management/matrix`, `/service/releases` |
| `check-environment <svc> --env` | effective env vars from `/deploy-management/service/overlays` (base ⊕ overlay) |
| `check-secrets <svc> --env [--reveal] [--vault]` | secret refs from overlay → local `az keyvault secret show` (masked by default). Vault defaults per env (`SECRET_VAULT_BY_ENV`: dev→regask-k8s-dev, staging→regask-k8s-qa, pre-prod→regask-k8s-pre-prod, production→regask-k8s-prod); `--vault` overrides |
| `find-ticket <TICKET...>` | `POST /deploy-management/ticket-lookup` |
| `query-approval <link-or-id>` | `GET /approvals/requests/:id` — detail + release link (`resultUrl`) + task backlink |
| `approve <link-or-id> [--reject]` | `GET` detail, confirm, `POST .../approve`\|`/reject` |
| `promote --to-env <env> --service <svc…>` | `template:default/promote-code` with `environment` (one of `staging`/`pre-prod`/`production`, validated client-side) and `services` = deduped `owner/repo` slugs read from each entity's `github.com/project-slug` annotation (`GET /catalog/entities/by-name/...`), matching the portal's GitopsServicesPicker. Streams the log; the final status is read from `GET /scaffolder/v2/tasks/:id` because the completion event carries no status |
| `release --env [--include-services\|--exclude-services]` / `cherry-pick --tag <TICKET> --branch <release/preprod\|release/prod>` | launch scaffolder templates (`release:*`, `cherry-pick:*`), stream log. release's include/exclude are mutually exclusive; cherry-pick's `--tag` is a ticket ref (e.g. REG-12345) and `--branch` is restricted to `release/preprod`/`release/prod` |

Cross-cutting: `--json` on every command; `--fresh` bypasses the server TTL
cache; non-zero exit on failure / rejected approval / failed task.

## Deferred backend contracts (confirm against the live backend)

These use documented defaults and are marked `TODO(execution)` in code until
confirmed:

1. ~~Portal CLI login start path~~ — **resolved:** `/cli-auth` page added to the
   portal; CLI sends a CSRF `state` and verifies it on the callback.
2. The `whoami` identity endpoint.
3. Scaffolder task/events endpoints + the `release` and `cherry-pick` templates' parameter names (promote-code is confirmed against `templates/promote-code/template.yaml`).
4. The runs/task backlink route used in the approval detail view.

Remaining login hardening (login timeout on a stalled browser flow) is still open.

Refreshable login needs the portal to set `auth.experimentalClientIdMetadataDocuments.enabled: true`
and `auth.experimentalRefreshToken.enabled: true` (backstage repo `app-config.yaml`);
until then `bsr login` uses the `/cli-auth` fallback.

## Distribution

Push to `main` → release-please computes semver + tags + GitHub Release →
GoReleaser cross-compiles (macOS arm64/x64, Linux) and pushes a Homebrew **cask**
to `regask/homebrew-tap`. Users install with `brew install regask/tap/backstage-regask`
and update with `brew upgrade`. Requires the `HOMEBREW_TAP_TOKEN` CI secret.

### Claude Code plugin (`plugins/backstage-regask/`)

A Claude Code plugin ships from this repo (`plugins/backstage-regask/`,
marketplace `regask` at `.claude-plugin/marketplace.json`) as a new
distribution surface alongside the Homebrew cask above. It is a thin wrapper:
its slash commands and bundled skill shell out to the existing binary
via the Bash tool, so it adds **no new backend contracts** — nothing in the
contracts/auth model above changes. Read-only commands (`status`, querying an
approval) pre-authorize their invocations via `allowed-tools`; mutating
actions (`approve`/`reject`, `promote`/`release`/`cherry-pick`) are left
ungated in the frontmatter and instead gated behind an explicit confirmation
rule in the bundled skill, consistent with D5's reuse of the existing
scaffolder workflows.

The plugin always spells the binary `backstage-regask`, never `bsr`: `bsr` is a
shell alias the user adds to their profile, and aliases don't resolve in the
non-interactive Bash tool. `allowed-tools` patterns therefore only list
`backstage-regask`, keeping the pre-authorized surface to one name.

The skill also carries the CLI's **naming vocabulary**, because neither the CLI
nor the backend normalizes these and both fail quietly:

- Environments are only `development` / `staging` / `pre-prod` / `production`
  (`ENVIRONMENTS` in the backend's `deploy-management-backend/src/types.ts`). An
  unknown `--env` makes `check-deploy` return `[]` with exit 0, and `release`
  forwards it to the scaffolder unvalidated. `promote` accepts only
  `staging`/`pre-prod`/`production`.
- Services are bare catalog names with no `-service` suffix, matched exactly.
  There is no list-services command, so the skill documents reading
  `/api/deploy-management/matrix` directly to enumerate them.
