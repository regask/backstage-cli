---
description: Check Backstage sign-in status, guide login, or update the bsr CLI.
allowed-tools: >-
  Bash(bsr whoami:*), Bash(bsr update:*),
  Bash(backstage-regask whoami:*), Bash(backstage-regask update:*)
---

Use the `backstage-regask` skill. Check sign-in status with `bsr whoami`.

If the user is not signed in, tell them to run `bsr login` themselves (it opens a
browser and waits) — do NOT run `bsr login` yourself; it is interactive and will
hang. To update the CLI, run `bsr update` (Homebrew).

Request: $ARGUMENTS
