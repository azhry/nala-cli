# Backend PR description template

Use this shared backend template for Go CLI, API-client, authentication, or local-persistence work. Replace every bracketed placeholder. Remove a conditional section only when it genuinely does not apply.

The PR description is the review and handoff record. Do not claim a check passed unless the exact command exited 0. Put blockers and pre-existing failures in their own section rather than presenting partial execution as success.

## Linked work

- Linear issue: [AZH-000 and URL]
- Parent/source issue: [issue ID and URL, or "None"]
- Depends on: [issue/PR and delivered artifact, or "None"]
- Unblocks: [issue IDs, or "None"]

## Review and merge order

- Delivery shape: [Single focused PR | Stacked PR | Parallel PR group]
- PR/MR link: [Direct URL for this pull request]
- This PR's review position: [Standalone | PR 1 of N | PR N of N | Parallel member A/B]
- Base branch: [main or predecessor branch]
- Depends on: [PR/commit and the exact delivered behavior, or "None"]
- Review order: [Exact order, or "Any order within <parallel group>"]
- Merge order and conditions: [Exact merge sequence and prerequisite checks, or "Any order; all required checks green"]
- Parallel group: [Group name and independent members, or "None"]
- Human-verification focus: [The one behavior and manual check a reviewer should prioritize]

## Summary

- [Observable backend outcome.]
- [API, security, persistence, or migration outcome.]
- [Important compatibility outcome.]

## Scope

### Included

- [Implemented behavior and affected package/path.]
- [Additional in-scope behavior.]

### Excluded

- [Nearby work intentionally not changed.]

## Verification

### Automated code checks (supporting only)

Record exact unfiltered commands and exit statuses here. Unit tests, integration tests, lint, builds, mocks, and local protocol checks are regression evidence at a code seam; they do not prove live API, authentication, PostgreSQL, Vault, or ownership behavior.

### Manual request/response sequence

Write this as a sequence of small, copy-pasteable Bash steps in the PR description.
Do not attach a script file or combine the whole verification into one bulk script.
Use real staging data and the documented staging fixture for the issue; do not use
placeholder values, fake records, or mocks. Never paste credentials, tokens, or
API keys into the PR. Do not use `set -o pipefail`, `set -e`, `set -Eeuo pipefail`,
or another fail-fast wrapper: keep each step independently runnable, print its
exit status, and leave the interactive terminal open when a request fails.
Do not use `exit`, `exit 1`, or cleanup traps that call `exit`. Keep setup in a
separate short block, use at most one network request per numbered step, avoid
helper functions and bulk scripts, and wrap long commands with continuations.

#### Step 0 — Load the verified environment

Verify the actual Nala Labs URL from project knowledge and the running process before publishing a completed handoff. Replace the example below if the documented target differs. The CLI default is a development URL, not evidence that a backend is running. Do not source private environment files or print session data.

```bash
export NALA_API_BASE_URL='http://127.0.0.1:8080'
status=$?
printf 'environment setup exit status: %s\n' "$status"
```

Expected: exit 0. Record the verified non-secret target and environment separately.

#### Step 1 — Authenticate when the flow requires it

For authentication changes, use the CLI's real browser login with the documented account; do not copy a server password-login request or extract the stored token. Respect the user's browser restrictions and record the flow as not run when it is not authorized.

```bash
go run ./cmd/nala login
status=$?
printf 'login exit status: %s\n' "$status"
```

Expected: exit 0 after successful browser authentication and a signed-in confirmation without a token. Record the actual sanitized result separately; this is not an observed pass.

#### Step 2 — Exercise the changed behavior

```bash
go run ./cmd/nala user info
status=$?
printf 'user info exit status: %s\n' "$status"
```

Expected: exit 0 and the authenticated account and entitlements, without a token. Replace this example for non-authentication tasks with the task's real command and observable contract.

#### Step 3 — Exercise the required regression or error case

```bash
[one Bash command for the real CLI regression/error/ownership scenario]
status=$?
printf 'regression exit status: %s\n' "$status"
```

Replace the command before publication. State the expected exit status and sanitized output for the real boundary case, then record the observed result or exact reason it was not run.

## Known limitations and pre-existing failures

- [Exact command, exit status, affected path, and why it is unrelated; or "None known".]

## Reviewer focus

- [Highest-risk contract, migration, authorization, or concurrency decision.]
- [Specific file or behavior that merits close review.]

## Completion self-audit

- [ ] Every issue requirement is mapped to an implemented outcome or explicit blocker.
- [ ] Every changed operation lists its inputs, outputs, errors, and authorization behavior.
- [ ] Persistence claims cross a new request/client/process boundary rather than reuse an in-memory object.
- [ ] Migration up/down and existing-data behavior are documented and tested where applicable.
- [ ] Success, validation, unauthenticated, cross-user, absent-resource, and persistence-error paths are covered where applicable.
- [ ] Exact unfiltered commands and exit statuses are recorded.
- [ ] Generated coverage, logs, dumps, credentials, and unrelated files are absent from the diff.
- [ ] Automated checks are clearly separated from manual acceptance evidence and are not presented as proof that the live work is complete.
- [ ] For API work, the PR body includes a copy-paste manual request/response sequence using real configured fixtures; automated tests alone do not satisfy this check.
- [ ] The linked Linear issue and dependencies reflect the actual handoff state.
