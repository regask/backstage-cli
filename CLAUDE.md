# backstage-regask CLI — Claude guardrails

A standalone Go CLI (`backstage-regask`, alias `bsr`) that drives the Regask
Backstage backend from the terminal. **Read the architecture doc before making
changes:** [docs/architecture/ARCHITECTURE.md](docs/architecture/ARCHITECTURE.md)
— it holds the design and the decisions behind it.

## Keep docs in sync (required)

- **Any change that alters an architectural decision MUST update
  `docs/architecture/ARCHITECTURE.md` in the same change.** New command, new
  dependency, changed auth/distribution model, a decision in the table changing
  — update the doc. Small bug fixes and refactors that don't change a decision
  don't need a doc edit.
- **User-facing changes update `README.md`** (install/usage) in the same change.

## Conventions

- Go module: `github.com/regask/backstage-cli`; Go floor `go 1.22`.
- Binary name: `backstage-regask`; `bsr` is a symlink from the Homebrew formula.
- **Thin commands, focused packages.** `cmd/` only parses args, calls a client
  method, and renders. Logic lives in `internal/{auth,client,contracts,az,scaffolder,render}`.
  Every external dependency (HTTP transport, token store, `az`, browser-open) is
  injectable so commands unit-test without side effects.
- **Every command supports `--json`**; data commands support `--fresh`
  (`?refresh=1`, bypasses the server TTL cache). Non-zero exit on
  failure / rejected approval / failed task.
- **Secrets are never printed without `--reveal` and never persisted.** The
  gitops overlay holds only secret references; values come from local `az`.
- **`internal/contracts` parsers must keep parity with the backend**
  (`plugins/deploy-management-common/src/index.ts` in the backstage repo). Don't
  let the porting logic drift; add parity tests when you touch them.
- Mutating workflows (`promote`/`release`/`cherry-pick`) launch the **existing**
  scaffolder templates — never reimplement git/gitops here.

## Git

- **Local commits only. NEVER `git push`** without an explicit request.
- **NEVER add a Claude/AI co-author or any Claude attribution** to commits or PR
  descriptions.
- Conventional-commit messages.

## Build / test

- Build: `go build ./...`  ·  Vet: `go vet ./...`  ·  Format: `gofmt -l .` (must be empty).
- Test: `go test ./...` (unit tests use injected transports/fakes — no real
  network, browser, or `az`).

## Deferred backend contracts

Remaining endpoint contracts use documented defaults and are marked
`TODO(execution)` until confirmed against the live backend: the `whoami` identity
endpoint, and the scaffolder task/events endpoints + template params. The portal
login handshake is resolved (`/cli-auth` page + CSRF `state`). See the
architecture doc's "Deferred backend contracts" section.
