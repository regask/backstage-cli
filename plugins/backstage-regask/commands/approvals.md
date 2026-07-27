---
description: Query a Backstage approval, then approve or reject it (with confirmation).
argument-hint: "[approval-url-or-id]"
allowed-tools: >-
  Bash(backstage-regask query-approval:*)
---

Use the `backstage-regask` skill. First run
`backstage-regask query-approval <url-or-id>` and show the details (including the
release link).

If the user wants to approve or reject, follow the skill's safety rule: print the
exact command and get an explicit "yes" first, then run
`echo y | backstage-regask approve <url-or-id>` (or add `--reject`).
Approve/reject is intentionally NOT pre-authorized, so it will prompt for
permission — that is expected.

Request: $ARGUMENTS
