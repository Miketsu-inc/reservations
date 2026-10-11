# Format valid currencies outside the configured locale map

Local branch: `fix/backend-currency-format-fallback`
Review base: `65b122fa`
Original review: `fix/backend-review` at `90541a2b`
Prerequisites: None; this fix is based directly on `master` at `65b122fa`.

This is a local review branch. It has not been pushed, merged, or applied to a running application. Compare it with the exact review base below; comparing a dependent fix directly with master also includes its prerequisites.

## Problem and impact

Formatting a valid currency outside the predefined formatter map could assert and then dereference a missing formatter in production.

## How to reproduce

Use the review-base revision for the before behavior and this branch for the after behavior, with the same inputs. Use a disposable development database/provider account when the scenario writes records. Keep existing project configuration and authorization requirements; do not run data-changing reproduction against production.

1. Inspect the configured formatter map and choose a currency supported by the currency library but absent from that map.
2. Follow Format with a nonzero amount in that currency.
3. Compare development assertion and production nil-formatter behavior with the fallback.

## What changed

A fresh English formatter is selected when no configured formatter exists for the currency.

The production diff for this chunk is:

```text
backend/pkg/currencyx/currencyx.go | 4 +++-
 1 file changed, 3 insertions(+), 1 deletion(-)
```

Relevant implementation points:

- [currencyx.go: Format](../pkg/currencyx/currencyx.go#L43)

## Why this fix

The formatter map is a presentation preference, not a validity check for all currencies the library supports. A local fallback avoids crashes and avoids adding shared-map mutations that would introduce synchronization concerns.

## Verification and limits

The isolated branch passed `go build ./backend/...`, `go build -tags=prod ./backend/...`, and `git diff --check`. Changed Go files were formatted with goimports. These are compilation and source checks: they do not prove runtime behavior or race freedom. The reproduction steps above are review instructions and were not executed against a live database, Redis instance, or Calendar account while splitting this branch. No new tests or regression tests were added, and the original review's integration-test results are not presented as results for this isolated branch.

Ignored generated email/frontend assets were copied into the separate worktree to satisfy existing Go embed requirements. Those assets are not committed. Your original checkout and the combined review branch were preserved.

## Review commands

```bash
git switch fix/backend-currency-format-fallback
git diff 65b122fa..fix/backend-currency-format-fallback -- backend ':!backend/reviews'
git log --oneline 65b122fa..fix/backend-currency-format-fallback
```

The original exclusions still apply: intentional localhost addresses, existing TODO items, and intentionally ignored repository query errors immediately followed by pgx collection helpers are not treated as defects.
