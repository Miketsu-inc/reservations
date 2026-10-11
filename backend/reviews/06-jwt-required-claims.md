# Require JWT expiry and valid UUID subjects

Local branch: `fix/backend-jwt-required-claims`  
Review base: `65b122fa`  
Original review: `fix/backend-review` at `90541a2b`  
Prerequisites: None; this fix is based directly on `master` at `65b122fa`.

This is a local review branch. It has not been pushed, merged, or applied to a running application. Compare it with the exact review base below; comparing a dependent fix directly with master also includes its prerequisites.

## Problem and impact

Signed JWTs with no expiry were not required to expire. A malformed subject could travel beyond token verification and become an internal error. NaN was not explicitly rejected during refresh-version conversion.

## How to reproduce

Use the review-base revision for the before behavior and this branch for the after behavior, with the same inputs. Use a disposable development database/provider account when the scenario writes records. Keep existing project configuration and authorization requirements; do not run data-changing reproduction against production.

1. Using a development signing secret, construct an HS256 access token with a UUID subject but no exp.
2. Repeat with exp present and a missing or malformed sub.
3. Inspect refresh-version conversion with NaN; this is a conversion boundary case, not a JSON token payload that can normally encode NaN.

## What changed

Token parsing explicitly permits HS256 and requires expiry. UUID subjects are checked before claims are returned; bad access/refresh subjects map to the matching invalid-token error. NaN versions are rejected.

The production diff for this chunk is:

```text
backend/internal/api/middleware/auth.go | 12 ++++++++++--
 1 file changed, 10 insertions(+), 2 deletions(-)
```

Relevant implementation points:

- [auth.go: getRefreshVersionFromClaims](../internal/api/middleware/auth.go#L155)
- [auth.go: verifyToken](../internal/api/middleware/auth.go#L183)

## Why this fix

A signature only proves who produced a token; claims still need their required shape. Validating that shape once at verification produces authentication failures consistently before request handlers access the subject.

## Verification and limits

The isolated branch passed `go build ./backend/...`, `go build -tags=prod ./backend/...`, and `git diff --check`. Changed Go files were formatted with goimports. These are compilation and source checks: they do not prove runtime behavior or race freedom. The reproduction steps above are review instructions and were not executed against a live database, Redis instance, or Calendar account while splitting this branch. No new tests or regression tests were added, and the original review's integration-test results are not presented as results for this isolated branch.

Ignored generated email/frontend assets were copied into the separate worktree to satisfy existing Go embed requirements. Those assets are not committed. Your original checkout and the combined review branch were preserved.

## Review commands

```bash
git switch fix/backend-jwt-required-claims
git diff 65b122fa..fix/backend-jwt-required-claims -- backend ':!backend/reviews'
git log --oneline 65b122fa..fix/backend-jwt-required-claims
```

The original exclusions still apply: intentional localhost addresses, existing TODO items, and intentionally ignored repository query errors immediately followed by pgx collection helpers are not treated as defects.
