# Backend workflow

Apply this shared Go/backend workflow to CLI changes under `cmd/` and `internal/`, including command behavior, HTTP clients, loopback callbacks, authentication, and local session persistence. The Go module lives at the repository root; this repository has no server-side database or migrations.

## Contract and analysis

- Complete the issue-description readiness gate before implementation. Treat the verified description as the contract for inputs, outputs, authorization, persistence, error behavior, and compatibility.
- Inspect the relevant handler, service, middleware, migration, and tests before choosing an implementation seam. Confirm current behavior from code or a reproduction; clearly separate confirmed facts from hypotheses.
- Trace every affected operation end to end: request parsing, authentication, authorization or resource ownership, validation, business logic, persistence, response serialization, and errors.
- Define changed GraphQL or HTTP fields and operations explicitly. Include required versus optional inputs, nullability, defaults, response shape, error cases, and backward-compatibility expectations.
- Keep work within the stated scope. Do not silently add adjacent schema changes, data migrations, refactors, or behavior changes that are not required by the issue contract.

## Implementation

- Preserve established package boundaries and existing response conventions. Keep transport parsing, domain behavior, authentication, and persistence responsibilities at their current seams unless the task explicitly requires an architectural change.
- Validate untrusted input at the service boundary. Return stable client-safe errors and avoid exposing credentials, tokens, SQL details, stack traces, or internal implementation data.
- Enforce authentication and resource ownership on every protected read and mutation, including indirect access through parent resources. A resource that belongs to another user should not become discoverable through differing success or error behavior.
- Keep authentication secure: Nala Labs owns password/provider exchange and session JWT issuance. Reuse the CLI's state, callback, and session helpers; never log secrets or persist provider credentials. The existing local session-file contract is documented in [authentication knowledge](../knowledge/authentication.md); preserve its restrictive permissions and do not introduce a second token issuer.
- Persist data when the issue contract requires durable behavior. Do not substitute process-local or demo state for the documented local session file or server-owned persistence boundary without an explicit scope decision in the issue.
- Make multi-step writes atomic when partial completion would violate an invariant. Pass cancellation-aware contexts through HTTP and callback operations and do not discard returned errors.
- Server schema and database migrations belong to Nala Labs, not this CLI. If a task explicitly requires server changes, read that repository's knowledge and workflow and deliver them there with explicit review/merge dependencies. Keep local session-format changes backward compatible unless the contract authorizes migration.
- Keep API behavior compatible with existing callers unless the issue explicitly authorizes a breaking change. Update all in-repository consumers and fixtures when a contract intentionally changes.

## Tests and verification

- Add regression tests at the narrowest reliable seam, then add operation-level coverage when behavior crosses authentication, authorization, serialization, or persistence boundaries.
- Cover the successful path and relevant failures: missing or malformed input, unauthenticated access, cross-user access, absent resources, persistence errors, and boundary values.
- For local session-format changes, verify compatibility, restrictive file permissions, read/write failures, and behavior across a new process. Server database acceptance belongs to the owning repository; record any unrun cross-service flow explicitly.
- Run focused tests while implementing. Before handing off Go changes, run these commands unfiltered from the repository root and require zero exit status. Capture and print each command's immediate status independently. For guidance-only changes, validate the changed links, skill structure, and bundled helper tests; do not claim runtime acceptance from those checks:

  ```bash
  go build ./...
  go vet ./...
  go test ./... -count=1 -short -v
  go test ./... -count=1 -coverprofile=coverage.out -v
  go tool cover -func=coverage.out
  ```

- Write any temporary verification script for these checks in Bash and run it with `bash`. Do not generate a PowerShell verification script unless Bash is unavailable or the user explicitly requests PowerShell. Do not add `set -o pipefail`, `set -e`, `set -Eeuo pipefail`, or another fail-fast wrapper to manual verification; keep each request/check independently runnable, print its exit status, and do not terminate the interactive terminal when a check fails.

- There is no CLI `make test-integration` target. When the change depends on live authentication or session behavior, use the documented CLI-to-Nala Labs flow with the real configured environment. Do not claim integration verification when only unit tests ran.
- Treat race-sensitive shared-state changes as requiring `go test -race` for the affected packages when the current platform supports it.
- Report exact commands and exit results. Distinguish environment blockers and pre-existing failures from regressions introduced by the task.
- Before staging and handoff, inspect `git status --short`; do not commit `coverage.out`, local environment files, database dumps, logs, credentials, or unrelated user files.

### API acceptance and test-environment lifecycle

- Automated unit, fake, mock, and SQL-mock tests are regression evidence at a seam. They do not substitute for API acceptance or prove real PostgreSQL, Vault, Casdoor, or other external-service behavior.
- For every HTTP/API, authentication, persistence, or migration change, run a documented manual API flow against the actual built or started service at the exact URL and port recorded for handoff. Use real configured dependencies, or explicitly documented process-level configuration when Vault is intentionally disabled. If that flow cannot run, record it as not run or blocked; do not promote an isolated-port, stale-process, mock, or partial harness result to live acceptance.
- Put copy-paste request steps in the issue or PR: working directory and launch command, readiness check, every documented approved fixture source checked before declaring authentication blocked, real request payloads, expected status/body, observed status/body, persistence boundary, fixture cleanup, and the command's exit status. Keep credentials, tokens, cookies, and connection strings out of the record.
- Run API acceptance in a unique temporary workspace outside the repository. Track the launched process tree and command path, clean it up in a `finally`/failure path, and verify the listener and child processes are gone before handoff. Remove task-created logs, response bodies, binaries, and coverage artifacts after the run.
- Do not mutate shared Vault or other runtime configuration without capturing non-secret original endpoint metadata, restoring it in a cleanup path, and verifying the restored state. If restoration cannot be proven, report the external-state risk and leave acceptance incomplete.

## Handoff

Stage only intended CLI, test, and documentation files. Commit, push, and open or update the PR according to the delivery workflow. Summarize API or schema changes, migration impact, compatibility notes, and verification results in the PR and linked work item.
