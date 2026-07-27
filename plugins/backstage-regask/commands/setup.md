---
description: Check Backstage sign-in status, guide login, or update the backstage-regask CLI.
allowed-tools: >-
  Bash(backstage-regask whoami:*), Bash(backstage-regask update:*),
  Bash(backstage-regask check-deploy:*)
---

Use the `backstage-regask` skill.

`backstage-regask whoami` reports the cached identity but only decodes the stored
JWT locally — it never calls the server and never checks expiry, so it prints a
user even when the session has expired. To actually verify sign-in, follow it
with a read command such as `backstage-regask check-deploy alert`; an
`unauthorized` error there means the token is stale despite `whoami` succeeding.

If the user is not signed in, tell them to run `backstage-regask login`
themselves (it opens a browser and waits up to 5 minutes) — do NOT run it
yourself; it is interactive and will hang. Note the CLI's own error message says
"run `bsr login`", but `bsr` is only a shell alias that may not exist, so always
quote the command as `backstage-regask login`.

To update the CLI, run `backstage-regask update` (Homebrew).

Request: $ARGUMENTS
