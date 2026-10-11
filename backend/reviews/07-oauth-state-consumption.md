# Consume OAuth state cookies after validation

Local branch: `fix/backend-oauth-state-consumption`
Review base: `65b122fa`
Original review: `fix/backend-review` at `90541a2b`
Prerequisites: None; this fix is based directly on `master` at `65b122fa`.

This is a local review branch. It has not been pushed, merged, or applied to a running application. Compare it with the exact review base below; comparing a dependent fix directly with master also includes its prerequisites.

## Problem and impact

Authentication callbacks left the browser state cookie in place after a valid callback. Its old value remained available to later callback attempts until cookie expiry.

## How to reproduce

Use the review-base revision for the before behavior and this branch for the after behavior, with the same inputs. Use a disposable development database/provider account when the scenario writes records. Keep existing project configuration and authorization requirements; do not run data-changing reproduction against production.

1. Begin Google or Facebook sign-in in a development browser and note the oauth-state cookie.
2. Complete the callback and inspect response Set-Cookie and browser cookie storage.
3. Compare whether the valid callback consumes the cookie.

## What changed

A shared deletion helper clears the same cookie name/path/security attributes. Both authentication callbacks call it after state validation.

The production diff for this chunk is:

```text
backend/internal/api/handler/auth/handler.go | 2 ++
 backend/pkg/oauthutil/oauthutil.go           | 7 +++++++
 2 files changed, 9 insertions(+)
```

Relevant implementation points:

- [handler.go: Handler:GoogleCallback](../internal/api/handler/auth/handler.go#L234)
- [handler.go: Handler:FacebookCallback](../internal/api/handler/auth/handler.go#L268)
- [oauthutil.go: DeleteOauthStateCookie](../pkg/oauthutil/oauthutil.go#L52)

## Why this fix

The browser state value represents one authorization round trip. Consuming it after validation limits accidental reuse while preserving invalid-callback behavior and the existing OAuth provider flow.

## Verification and limits

The isolated branch passed `go build ./backend/...`, `go build -tags=prod ./backend/...`, and `git diff --check`. Changed Go files were formatted with goimports. These are compilation and source checks: they do not prove runtime behavior or race freedom. The reproduction steps above are review instructions and were not executed against a live database, Redis instance, or Calendar account while splitting this branch. No new tests or regression tests were added, and the original review's integration-test results are not presented as results for this isolated branch.

Ignored generated email/frontend assets were copied into the separate worktree to satisfy existing Go embed requirements. Those assets are not committed. Your original checkout and the combined review branch were preserved.

## Review commands

```bash
git switch fix/backend-oauth-state-consumption
git diff 65b122fa..fix/backend-oauth-state-consumption -- backend ':!backend/reviews'
git log --oneline 65b122fa..fix/backend-oauth-state-consumption
```

The original exclusions still apply: intentional localhost addresses, existing TODO items, and intentionally ignored repository query errors immediately followed by pgx collection helpers are not treated as defects.
