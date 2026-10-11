# Return 403 for authenticated permission denials

Local branch: `fix/backend-authenticated-forbidden-status`
Review base: `65b122fa`
Original review: `fix/backend-review` at `90541a2b`
Prerequisites: None; this fix is based directly on `master` at `65b122fa`.

This is a local review branch. It has not been pushed, merged, or applied to a running application. Compare it with the exact review base below; comparing a dependent fix directly with master also includes its prerequisites.

## Problem and impact

Role and subscription denials returned 401 even though authentication had already succeeded. Clients could interpret a capability denial as a lost login and attempt unnecessary reauthentication.

## How to reproduce

Use the review-base revision for the before behavior and this branch for the after behavior, with the same inputs. Use a disposable development database/provider account when the scenario writes records. Keep existing project configuration and authorization requirements; do not run data-changing reproduction against production.

1. Sign in as a merchant employee whose role cannot access an owner-only operation, and call that operation.
2. Repeat with a feature denied by the current subscription.
3. Compare with an unauthenticated request to ensure its separate authentication behavior is not changed.

## What changed

Role and subscription denials use HTTP 403. Their existing permission checks and error messages remain in place.

The production diff for this chunk is:

```text
backend/internal/api/middleware/rbac.go         | 2 +-
 backend/internal/api/middleware/subscription.go | 2 +-
 2 files changed, 2 insertions(+), 2 deletions(-)
```

Relevant implementation points:

- [rbac.go: Manager:RoleBasedAccessControl](../internal/api/middleware/rbac.go#L16)
- [subscription.go: Manager:Subscription](../internal/api/middleware/subscription.go#L15)

## Why this fix

403 expresses a denied permission for an authenticated principal. Correcting the status requires no change to the role matrix or subscription policy, so this fix does not invent new access rules.

## Verification and limits

The isolated branch passed `go build ./backend/...`, `go build -tags=prod ./backend/...`, and `git diff --check`. Changed Go files were formatted with goimports. These are compilation and source checks: they do not prove runtime behavior or race freedom. The reproduction steps above are review instructions and were not executed against a live database, Redis instance, or Calendar account while splitting this branch. No new tests or regression tests were added, and the original review's integration-test results are not presented as results for this isolated branch.

Ignored generated email/frontend assets were copied into the separate worktree to satisfy existing Go embed requirements. Those assets are not committed. Your original checkout and the combined review branch were preserved.

## Review commands

```bash
git switch fix/backend-authenticated-forbidden-status
git diff 65b122fa..fix/backend-authenticated-forbidden-status -- backend ':!backend/reviews'
git log --oneline 65b122fa..fix/backend-authenticated-forbidden-status
```

The original exclusions still apply: intentional localhost addresses, existing TODO items, and intentionally ignored repository query errors immediately followed by pgx collection helpers are not treated as defects.
