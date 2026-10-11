# Allow changing the capitalization of your own email address

Local branch: `fix/backend-user-email-case`
Review base: `65b122fa`
Original review: `fix/backend-review` at `90541a2b`
Prerequisites: None; this fix is based directly on `master` at `65b122fa`.

This is a local review branch. It has not been pushed, merged, or applied to a running application. Compare it with the exact review base below; comparing a dependent fix directly with master also includes its prerequisites.

## Problem and impact

Profile editing compared email strings case-sensitively before checking uniqueness. An email differing only in case could be rejected as already used by the same user even though the database models email identity case-insensitively.

## How to reproduce

Use the review-base revision for the before behavior and this branch for the after behavior, with the same inputs. Use a disposable development database/provider account when the scenario writes records. Keep existing project configuration and authorization requirements; do not run data-changing reproduction against production.

1. Sign in as a user with person@example.com.
2. Submit an otherwise-valid profile update using Person@example.com.
3. Compare the self-conflict with a genuinely different address already owned by another user.

## What changed

The decision to perform uniqueness checking uses strings.EqualFold against the current email.

The production diff for this chunk is:

```text
backend/internal/service/user/user.go | 3 ++-
 1 file changed, 2 insertions(+), 1 deletion(-)
```

Relevant implementation points:

- [user.go: Service:Edit](../internal/service/user/user.go#L30)

## Why this fix

The application check should agree with the lower(email) unique index. A case-only update is the same identity; the existing database constraint still protects genuinely conflicting identities.

## Verification and limits

The isolated branch passed `go build ./backend/...`, `go build -tags=prod ./backend/...`, and `git diff --check`. Changed Go files were formatted with goimports. These are compilation and source checks: they do not prove runtime behavior or race freedom. The reproduction steps above are review instructions and were not executed against a live database, Redis instance, or Calendar account while splitting this branch. No new tests or regression tests were added, and the original review's integration-test results are not presented as results for this isolated branch.

Ignored generated email/frontend assets were copied into the separate worktree to satisfy existing Go embed requirements. Those assets are not committed. Your original checkout and the combined review branch were preserved.

## Review commands

```bash
git switch fix/backend-user-email-case
git diff 65b122fa..fix/backend-user-email-case -- backend ':!backend/reviews'
git log --oneline 65b122fa..fix/backend-user-email-case
```

The original exclusions still apply: intentional localhost addresses, existing TODO items, and intentionally ignored repository query errors immediately followed by pgx collection helpers are not treated as defects.
