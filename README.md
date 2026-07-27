# backstage-regask (`bsr`)

Regask Backstage from your terminal — deploy status, environment config, secret
refs, ticket lookups, approvals, and the promote / release / cherry-pick
workflows, authenticated as your Backstage user.

## Install

```bash
brew install regask/tap/backstage-regask
```

Update: run **`bsr update`** (wraps `brew update` + `brew upgrade backstage-regask`),
or `brew upgrade backstage-regask` directly. The command is `backstage-regask`; for
the short form add `alias bsr=backstage-regask` to your shell profile.

## Usage

```bash
bsr login                                          # browser sign-in, token cached in ~/.config/backstage-regask
bsr whoami

bsr check-deploy alert                             # what version is deployed in every env
bsr check-deploy alert --env production            # narrow to one environment
bsr check-environment alert --env production       # effective env vars for the service in an env
bsr check-secrets alert --env production           # secret refs, values masked
bsr check-secrets alert --env production --reveal   # resolve values via `az` (needs `az login`)

bsr find-ticket ABC-123                            # which envs a ticket is deployed to

bsr query-approval https://backstage.regask.com/approvals/<id>   # details + release link (read-only)
bsr approve https://backstage.regask.com/approvals/<id>          # approve (or --reject)

bsr promote --to-env staging --service a,b         # promote one or more services to an environment
bsr release --env production                       # release an environment (all services)
bsr release --env production --include-services a,b # only these services (XOR --exclude-services)
bsr release --env production --exclude-services c  # all except these services
bsr cherry-pick --tag REG-12345 --branch release/preprod  # cherry-pick a ticket onto a release branch
                                                          # --branch: release/preprod | release/prod
```

- Add `--json` to any command for machine-readable output.
- Add `--fresh` to a data command to bypass the server cache.

### Naming

- **Environments** are exactly `development`, `staging`, `pre-prod`,
  `production` — there are no short aliases, so `--env prod` is not valid.
  Nothing normalizes the value: `check-deploy --env prod` returns an empty
  result rather than an error, and `promote`/`release` pass it straight to the
  scaffolder template.
- **Services** use their bare catalog name, with no `-service` suffix — `alert`,
  `task-v2`, `alert-event-handler`. `check-deploy <service>` also accepts the
  full ref (`component:default/alert`). Omit `--env` to see all environments.

## TUI

```bash
bsr ui
```

Opens an interactive k9s-style terminal UI (requires `bsr login` first).

**V1 views:**
- **Services**: deploy matrix (version + ArgoCD sync/health by environment), filterable by service name.
- **Approvals**: approval requests list → detail with release link and task backlink.

**Keys:**
- `Tab` switch between services and approvals
- `:` command bar (`services`, `approvals`, `quit`)
- `/` filter (current view)
- `↑/↓` or `j/k` navigate
- `Enter` open (detail/drill-down)
- `p` promote (services) / `a` approve (approvals)
- `r` release (services) / `x` reject (approvals)
- `Ctrl-R` refresh (hard-bypass caches)
- `?` help
- `q` quit

## Claude Code plugin

Drive `bsr` from inside Claude Code. Install once:

```
/plugin marketplace add regask/backstage-cli
/plugin install backstage-regask@regask
```

Requires the CLI installed and `bsr login` done. The plugin always invokes
`backstage-regask` (the Homebrew binary name) rather than the `bsr` alias, since
shell aliases don't resolve in Claude Code's non-interactive shell. Commands:

- `/backstage-regask:status [service] [--env <env>]` — deploy versions, env vars, secret refs (masked), ticket deploy state (read-only).
- `/backstage-regask:approvals [url]` — query an approval, then approve/reject (asks first).
- `/backstage-regask:release` — promote / release / cherry-pick workflows (asks first).
- `/backstage-regask:setup` — check sign-in, guide `bsr login`, or `bsr update`.

You can also just ask in plain language ("what's on prod for alert") — the
bundled skill picks the right command and maps everyday wording onto the CLI's
canonical environment and service names. Mutating actions always confirm first.

## Documentation

Architecture and design decisions: [docs/architecture/ARCHITECTURE.md](docs/architecture/ARCHITECTURE.md).

## Development

```bash
go build ./...
go test ./...
gofmt -l .   # must print nothing
```

Contributor guardrails and the "keep docs in sync" rule live in
[CLAUDE.md](CLAUDE.md).
