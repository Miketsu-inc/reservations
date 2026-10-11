# Use linked user names instead of copying them on invitation acceptance

Local branch: `fix/backend-linked-employee-profile-source`
Review base: `65b122fa`
Original review: `fix/backend-review` at `90541a2b`
Prerequisites: None; this fix is based directly on `master` at `65b122fa`.

This is a local review branch. It has not been pushed, merged, or applied to a running application. Compare it with the exact review base below; comparing a dependent fix directly with master also includes its prerequisites.

## Problem and impact

Accepting an invitation copied the user name into Employee, although employee reads prefer non-null Employee names over the linked User. Later user-profile changes could therefore be hidden by the stale copy.

## How to reproduce

Use the review-base revision for the before behavior and this branch for the after behavior, with the same inputs. Use a disposable development database/provider account when the scenario writes records. Keep existing project configuration and authorization requirements; do not run data-changing reproduction against production.

1. Accept an invitation with a development user and inspect Employee first_name/last_name.
2. Change the linked User name, then read employee/team details through queries that coalesce Employee and User names.
3. Compare a linked employee with an unlinked employee whose own names are intentional.

For concurrency scenarios, send the competing requests/jobs from separate clients/workers against the same record. The race is timing dependent; a single successful attempt does not disprove it. The code paths below show the decision/write boundaries to inspect when arranging overlapping requests.

## What changed

Invitation acceptance links the user without populating duplicate first/last name fields. Existing coalesce-based employee reads resolve the current user profile.

The production diff for this chunk is:

```text
backend/internal/service/team/invitation.go | 2 --
 1 file changed, 2 deletions(-)
```

Relevant implementation points:

- [invitation.go: Service:AcceptInvitation](../internal/service/team/invitation.go#L291)

## Why this fix

The linked User already owns this profile data. Avoiding a duplicate at creation removes an unnecessary stale snapshot without changing intentionally supplied names of unlinked employees. Existing linked rows with historical copies are not rewritten, and user-deletion behavior remains outside this scoped change.

## Verification and limits

The isolated branch passed `go build ./backend/...`, `go build -tags=prod ./backend/...`, and `git diff --check`. Changed Go files were formatted with goimports. These are compilation and source checks: they do not prove runtime behavior or race freedom. The reproduction steps above are review instructions and were not executed against a live database, Redis instance, or Calendar account while splitting this branch. No new tests or regression tests were added, and the original review's integration-test results are not presented as results for this isolated branch.

Ignored generated email/frontend assets were copied into the separate worktree to satisfy existing Go embed requirements. Those assets are not committed. Your original checkout and the combined review branch were preserved.

## Review commands

```bash
git switch fix/backend-linked-employee-profile-source
git diff 65b122fa..fix/backend-linked-employee-profile-source -- backend ':!backend/reviews'
git log --oneline 65b122fa..fix/backend-linked-employee-profile-source
```

The original exclusions still apply: intentional localhost addresses, existing TODO items, and intentionally ignored repository query errors immediately followed by pgx collection helpers are not treated as defects.
