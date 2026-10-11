# Validate UUID request fields as typed UUIDs

Local branch: `fix/backend-customer-uuid-validation`  
Review base: `65b122fa`  
Original review: `fix/backend-review` at `90541a2b`  
Prerequisites: None; this fix is based directly on `master` at `65b122fa`.

This is a local review branch. It has not been pushed, merged, or applied to a running application. Compare it with the exact review base below; comparing a dependent fix directly with master also includes its prerequisites.

## Problem and impact

Customer request structs applied the string-oriented uuid validation tag to uuid.UUID fields. JSON decoding already parses the UUID, while required enforces a nonzero value; the extra tag used the wrong representation.

## How to reproduce

Use the review-base revision for the before behavior and this branch for the after behavior, with the same inputs. Use a disposable development database/provider account when the scenario writes records. Keep existing project configuration and authorization requirements; do not run data-changing reproduction against production.

1. Send a customer update or blacklist request with a correctly formatted nonzero UUID.
2. Trace validator handling of uuid.UUID, which is an array type rather than the original JSON string.
3. Repeat with a malformed JSON UUID and the zero UUID to inspect decoding and required validation.

For concurrency scenarios, send the competing requests/jobs from separate clients/workers against the same record. The race is timing dependent; a single successful attempt does not disprove it. The code paths below show the decision/write boundaries to inspect when arranging overlapping requests.

## What changed

The customer update and blacklist UUID fields retain required and remove the incompatible uuid tag.

The production diff for this chunk is:

```text
backend/internal/api/handler/merchants/customers/handler.go | 4 ++--
 1 file changed, 2 insertions(+), 2 deletions(-)
```

Relevant implementation points:

- [handler.go: type:updateReq](../internal/api/handler/merchants/customers/handler.go#L68)
- [handler.go: type:blacklistReq](../internal/api/handler/merchants/customers/handler.go#L173)

## Why this fix

Validation should match the decoded type. The UUID unmarshaler handles string syntax, and required handles the zero value; revalidating an array through a string-specific tag adds no protection.

## Verification and limits

The isolated branch passed `go build ./backend/...`, `go build -tags=prod ./backend/...`, and `git diff --check`. Changed Go files were formatted with goimports. These are compilation and source checks: they do not prove runtime behavior or race freedom. The reproduction steps above are review instructions and were not executed against a live database, Redis instance, or Calendar account while splitting this branch. No new tests or regression tests were added, and the original review's integration-test results are not presented as results for this isolated branch.

Ignored generated email/frontend assets were copied into the separate worktree to satisfy existing Go embed requirements. Those assets are not committed. Your original checkout and the combined review branch were preserved.

## Review commands

```bash
git switch fix/backend-customer-uuid-validation
git diff 65b122fa..fix/backend-customer-uuid-validation -- backend ':!backend/reviews'
git log --oneline 65b122fa..fix/backend-customer-uuid-validation
```

The original exclusions still apply: intentional localhost addresses, existing TODO items, and intentionally ignored repository query errors immediately followed by pgx collection helpers are not treated as defects.
