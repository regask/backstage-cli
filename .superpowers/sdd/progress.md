# backstage-regask CLI — SDD progress ledger

Mode: standalone repo github.com/regask/backstage-regask-cli. COMMITS: local per-task OK; NO push; NO Claude co-author/attribution; PR only when user asks.
Plan: <backstage>/docs/superpowers/plans/2026-07-24-backstage-regask-cli.md
Spec: <backstage>/docs/superpowers/specs/2026-07-24-backstage-regask-cli-design.md

## Tasks
- [x] Task 1: Repo scaffold + cobra root
- [x] Task 2: Config/token store
- [x] Task 3: Typed HTTP client
- [x] Task 4: Contracts + parser ports
- [x] Task 5: Loopback login + logout + whoami
- [ ] Task 6: check-deploy + check-environment
- [ ] Task 7: check-secrets + az
- [ ] Task 8: find-ticket
- [ ] Task 9: query-approval + approve
- [ ] Task 10: Scaffolder runner + promote/release/cherry-pick
- [ ] Task 11: Distribution CI

## Log
BASE for Task 1 = empty tree (4b825dc)
Task 1: complete (empty..6113511, review clean, no attribution). BASE for Task 2 = 6113511
  MINOR (final review): render.Output JSON branch writes os.Stdout directly (brief-mandated signature; --json output not buffer-injectable in tests); table() invoked without nil-check.
Task 2: complete (6113511..99940ae incl. fix, review clean after fix). BASE for Task 3 = 99940ae
  Fix: Save now os.Chmod dir 0700 + file 0600 to enforce on pre-existing paths (+regression test).
  MINOR (final review): Expiry omitempty is no-op on struct; store_test uses string concat not filepath.Join; no Load-missing/dir-mode tests.
Task 3: complete (99940ae..f7753c0, review clean, no attribution). BASE for Task 4 = f7753c0
  MINOR (final review): GetJSON mutates caller-supplied url.Values in place (latent aliasing; current callers build fresh map each call — worth a defensive clone); io.ReadAll err swallowed in non-2xx path; no PostJSON/BaseURL/error-body tests.
Task 4: complete (f7753c0..32f8668 incl. fix, review clean after fix). BASE for Task 5 = 32f8668
  Fix: ParseSecretRefs remoteRef detection anchored to ^remoteRef:$ on trimmed line (parity w/ backend); secretKey regex widened to ^\s*-?\s* (correct fix for brief snippet bug: TS trims first) + why-comment; +2 parity tests.
Task 5: complete (32f8668..HEAD, review clean, no attribution). BASE for Task 6 = HEAD
  OPEN (intentional, per brief): TODO(execution) markers in internal/auth/login.go (portal start-path shape) and cmd/whoami.go (identity endpoint) await the real backstage auth-config contract.
