---
name: manual-test-verification
description: Execute task and pull-request manual verification as context-first, Bash-only evidence collection, reporting each command's immediate status, sanitized response, expected/pass/fail classification, and live-boundary limitations. Use when a task or pull request contains manual test steps, API acceptance commands, verification evidence, or a request to report what commands ran and whether results matched expectations.
---

# Manual Test Verification

Use this skill to produce reviewable evidence for manual test steps without overstating what local fixtures or wrappers prove.

## Quick start

1. Read the task or pull-request manual steps, `AGENTS.md`, and the relevant `.agents/knowledge/` files before choosing commands or assumptions.
2. Verify the repository root, documented environment, running URL/port, and fixture or account identity. For `.agents/.env`, load only the exact documented key needed for the current request with a non-printing allowlisted lookup; never dot-source/source the file, load all variables, print credentials, or transmit secret values.
3. Execute the exact contract in Bash only. Make each numbered verification step one independently pasteable fenced Bash block containing one target request (normally one simple `curl`) plus its immediate status capture. Keep setup, authentication/fixture creation, assertions, and cleanup in separate blocks. Never create, execute, or hand off a helper script, full-flow script, bulk runner, loop, function, or block that runs multiple target requests; never replace the actual command with an endpoint label such as `POST /api/foo`.
4. Capture the immediate exit status of every target command before running another command. A wrapper's final status is not evidence for a nested command.
5. Compare the observed status and sanitized response with the stated expected result. Classify each step as pass, expected result, or fail; preserve unexpected nonzero statuses.
6. Report the command, immediate status, sanitized response, expected outcome, classification, environment/fixture identity, and any live API, authentication, persistence, or external-service limitation.

## Execution gate

Before writing or updating manual evidence in a task or pull request:

- Run the live target first and keep its exact immediate status and sanitized response in an evidence ledger. A planned command, a reviewer instruction, a local fixture, or a prose summary is not observed evidence.
- Never put a fabricated or placeholder positive target in the acceptance flow: no `id 0`, `example.invalid`, invented app name, or hard-coded positive ID. Derive positive app and deployment IDs from the real fixture response. If a negative validation check is useful, label it regression-only and do not present it as live acceptance.
- “A reviewer must run this” is a handoff condition, not completion. When the current agent has the documented boundary and authorization, execute the required live flow itself; otherwise record the exact unavailable service, dependency, account, or fixture.
- Treat an accepted `queued` request as request-level evidence only. End-to-end live acceptance requires the documented terminal event/state. An interrupted follow, wrapper timeout, missing terminal event, or nonzero target status is `fail` or an explicit limitation, never a pass.
- Do not update verification-related tracker/PR state—including pass, acceptance, completion, or ready-for-review markers—until every required target has an evidence-ledger entry and, when required, the fresh-context verifier ledger has been returned and checked. Before then, record only an explicit blocker or limitation; never mark unrun work passed.

For live flows that cross services or mutate external state, use the independent-verifier protocol below. When an isolated subagent facility and a safe documented account/task-owned fixture are available, delegation is a required gate before reporting live acceptance; if either is unavailable, record the exact blocker.

## Audit-derived readiness checks

When the task includes a tracker or pull-request readiness artifact, apply these safeguards before execution:

- F-1/F-2: confirm visual references use the requested UML sequence semantics and the saved human description follows the exact closed heading schema; do not treat attachment counts or top-level counts as sufficient.
- F-4: confirm `Implementation Manual Test and Verification` contains Steps 0–4, separate Bash blocks, per-step statuses, no fake records/placeholders, and an explicit live-boundary limitation.
- F-5: inspect the active repository's `AGENTS.md`; do not assume a sibling repository's guardrails apply locally.
- F-3: capture each target command's immediate status before assertions or wrapper work, even when the expected result is nonzero.

## Guardrails

- Keep each Bash command independently pasteable and leave the terminal open after failures. A verification block must not bundle health, login, create, replay, and cleanup requests into one copy-paste. In the report, reproduce the exact command that ran and its immediate status; do not report a prose endpoint name or a whole-script summary as the command.
- Never expose API keys, JWTs, cookies, Vault values, provider tokens, or other credentials in output or reports.
- Unit tests, local fixtures, stub servers, and protocol checks are regression evidence only; they do not prove live API or authentication behavior.
- If the live boundary was not run, say so explicitly and name the unavailable dependency or required human action. Do not call an unrun live flow passed.
- Preserve the repository's documented run commands and environment names. Do not silently substitute ports, accounts, providers, or fixtures.

## Independent verifier protocol

For deploy, monitor, delete, or other destructive/cross-service acceptance, a fresh-context verifier is required when an isolated subagent facility and a safe documented account/task-owned fixture are available. It must read the current task/PR, `AGENTS.md`, the relevant knowledge, and this skill itself; run the exact independently pasteable Bash steps; and return only the structured evidence ledger. Do not pass raw credentials, tokens, cookies, or opaque IDs to it. The verifier must not edit code, issue/PR text, or delete an existing user-owned application. The main agent remains responsible for validating the ledger, recording limitations, authorized cleanup, and tracker/PR updates. Do not write live acceptance into the task or PR until the verifier ledger has been returned and checked.

If no isolated verifier facility exists, or the verifier lacks a safe documented account and task-owned fixture, run the check in the current context and record that exact limitation. A fresh context reduces context-interference risk; it does not make unrun or failed evidence pass.

For the full command and reporting contract, read [manual-test-contract.md](references/manual-test-contract.md).
