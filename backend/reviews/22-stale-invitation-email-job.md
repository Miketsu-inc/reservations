# Skip delayed invitation email jobs carrying an obsolete token

Local branch: `fix/backend-stale-invitation-email-job`  
Review base: `65b122fa`  
Original review: `fix/backend-review` at `90541a2b`  
Prerequisites: None; this fix is based directly on `master` at `65b122fa`.

This is a local review branch. It has not been pushed, merged, or applied to a running application. Compare it with the exact review base below; comparing a dependent fix directly with master also includes its prerequisites.

## Problem and impact

A queued email job could carry the plaintext token from an invitation that had since been resent, rotated, or accepted. Processing it could produce a link inconsistent with the current invitation.

## How to reproduce

Use the review-base revision for the before behavior and this branch for the after behavior, with the same inputs. Use a disposable development database/provider account when the scenario writes records. Keep existing project configuration and authorization requirements; do not run data-changing reproduction against production.

1. Queue an invitation email job in development without processing it immediately.
2. Resend the invitation to rotate its token, or accept/revoke it, then resume the original job.
3. Inspect whether the worker continues through delivery dependencies for the obsolete payload.

## What changed

The worker checks that the invitation is pending and that the SHA-256 hash of the job token matches its stored hash before processing.

The production diff for this chunk is:

```text
backend/internal/jobs/workers/email.go | 3 ++-
 1 file changed, 2 insertions(+), 1 deletion(-)
```

Relevant implementation points:

- [email.go: EmployeeInvitationEmail:Work](../internal/jobs/workers/email.go#L404)

## Why this fix

Background work must validate the state represented by its payload when it runs. Comparing the plaintext token directly with the stored hash would always fail; hashing the job token preserves the storage model and suppresses only obsolete work. The existing invitation mailer TODO is not implemented here.

## Verification and limits

The isolated branch passed `go build ./backend/...`, `go build -tags=prod ./backend/...`, and `git diff --check`. Changed Go files were formatted with goimports. These are compilation and source checks: they do not prove runtime behavior or race freedom. The reproduction steps above are review instructions and were not executed against a live database, Redis instance, or Calendar account while splitting this branch. No new tests or regression tests were added, and the original review's integration-test results are not presented as results for this isolated branch.

Ignored generated email/frontend assets were copied into the separate worktree to satisfy existing Go embed requirements. Those assets are not committed. Your original checkout and the combined review branch were preserved.

## Review commands

```bash
git switch fix/backend-stale-invitation-email-job
git diff 65b122fa..fix/backend-stale-invitation-email-job -- backend ':!backend/reviews'
git log --oneline 65b122fa..fix/backend-stale-invitation-email-job
```

The original exclusions still apply: intentional localhost addresses, existing TODO items, and intentionally ignored repository query errors immediately followed by pgx collection helpers are not treated as defects.
