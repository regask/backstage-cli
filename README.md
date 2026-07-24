# backstage-regask (`bsr`)

Regask Backstage from your terminal — deploy status, environment config, secret
refs, ticket lookups, approvals, and the promote / release / cherry-pick
workflows, authenticated as your Backstage user.

## Install

```bash
brew install regask/tap/backstage-regask
```

Update: `brew upgrade backstage-regask`.

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
