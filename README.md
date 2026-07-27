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

bsr check-deploy <service> --env prod              # what version is deployed where
bsr check-environment <service> --env prod         # effective env vars for the service in an env
bsr check-secrets <service> --env prod             # secret refs, values masked
bsr check-secrets <service> --env prod --reveal    # resolve values via `az` (needs `az login`)

bsr find-ticket ABC-123                            # which envs a ticket is deployed to

bsr query-approval https://backstage.regask.com/approvals/<id>   # details + release link (read-only)
bsr approve https://backstage.regask.com/approvals/<id>          # approve (or --reject)

bsr promote --to-env staging --service a,b         # promote one or more services to an environment
bsr release --env prod                             # release an environment (all services)
bsr release --env prod --include-services a,b      # only these services (XOR --exclude-services)
bsr release --env prod --exclude-services c        # all except these services
bsr cherry-pick --tag REG-12345 --branch release/preprod  # cherry-pick a ticket onto a release branch
                                                          # --branch: release/preprod | release/prod
```

- Add `--json` to any command for machine-readable output.
- Add `--fresh` to a data command to bypass the server cache.

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

Requires `bsr` installed and `bsr login` done. Commands:

- `/backstage-regask:status [service] [--env <env>]` — deploy versions, env vars, secret refs (masked), ticket deploy state (read-only).
- `/backstage-regask:approvals [url]` — query an approval, then approve/reject (asks first).
- `/backstage-regask:release` — promote / release / cherry-pick workflows (asks first).
- `/backstage-regask:setup` — check sign-in, guide `bsr login`, or `bsr update`.

You can also just ask in plain language ("what's on prod for alert-service") —
the bundled skill picks the right command. Mutating actions always confirm first.

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
