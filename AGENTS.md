# Project agent instructions

## Required first steps

- No AI/Codex faking: never mock, stub, or fake Codex responses, LLM completions, or core agent client methods inside unit tests.
- If using a worktree, clean up the task-owned worktree and its related caches and artifacts after delivery; preserve user-owned files and worktrees.
- Before any task action, load only the immediately required allowlisted key from `.agents/config.md` through a non-printing loader. Never call a generic read tool on this file or render, print, commit, or transmit its contents.
- Read the relevant files under `.agents/knowledge/` before making assumptions about runtime accounts, providers, roles, tiers, endpoints, or test fixtures. Keep secret values out of transcripts and handoff records.
- Read [.agents/knowledge/index.md](.agents/knowledge/index.md) to know the project context.
- Use authenticated connectors or stored CLI credentials without loading optional project secrets. If an immediate operation needs a secret, follow the allowlisted loading rules; never copy another service's private config into this repository.
- Config access must use a non-printing allowlisted loader. Never use `Get-Content`, `Select-String`, `rg`, `type`, or any generic reader on `.agents/config.md`, even when assigning the result to `$null`. Load only the single required `KEY=value` entry into the current process.
- Do not search user-profile or global Codex state (`$env:USERPROFILE`, `.codex`, or `.agents` outside the active repository) for credentials, connector configuration, or session data. Use the documented connector or a repository-scoped allowlisted loader; if neither is available, record the exact blocker and stop.
- Make sure all the necessary tools and credentials work before taking task actions.
- Do one task at a time. A task is complete only after implementation, verification, commit, push, PR handoff, and relevant tracker update are complete.
- Preserve unrelated dirty files. Never stage, modify, discard, or overwrite another person's work.
- Write verification scripts and verification command blocks in Bash (`bash`/`sh`) by default, including on Windows. Use Git Bash or WSL when available. Use PowerShell only when the user explicitly requests it or when the verification cannot run in Bash; document that exception. Manual verification must not use `set -o pipefail`, `set -e`, `set -Eeuo pipefail`, or another fail-fast wrapper that can terminate the interactive shell. Run steps independently, capture and print each target command's immediate exit status, and do not treat a wrapper's final exit code as evidence for an earlier command. Leave the terminal open after failures.
- Store agent-facing project knowledge under `.agents/knowledge/`. Reserve `docs/` for user- or developer-facing product documentation; do not create an agent-knowledge artifact under `docs/`.

## Execution discipline

- Treat the user's requested outcome and scope as the contract. Complete required workflow gates, but do not add unrelated review, refactoring, cleanup, generated artifacts, or product changes. When the user has approved the full task, proceed end-to-end without repeating optional questions; ask only when a missing decision would materially change scope or a safety boundary requires it.
- Keep updates outcome-first, concise, and plain-language. Report only task-relevant progress, blockers, decisions, and evidence; avoid jargon and long narration.
- Write manual verification directly in the issue or PR description as short, independently pasteable Bash steps. For HTTP behavior, use one simple request per step and a concise expected response. Keep setup separate, do not create or hand off a PowerShell/Bash verification script, helper script, or bulk runner, and never present an unrun response as observed. Use `.agents/skills/manual-test-verification/SKILL.md` and its linked reference for the full context-first, sanitization, immediate-status, and classification contract.
- Do not add helper scripts or generated diagnostics to a PR unless the task explicitly requires a maintained artifact. Keep temporary screenshots, logs, traces, and captures outside the repository or in ignored temporary storage, and remove task-owned leftovers.
- Before changing behavior that may cross a repository boundary, read this repository's relevant `.agents/knowledge/` and the related repository knowledge. Nala Labs owns Casdoor/session JWT issuance; nala-svc consumes that JWT and does not mint a second login token; NalaGrow and Nala Trace use their documented shared-platform boundaries. Do not add duplicate authentication/provider behavior unless the issue explicitly requires it and the knowledge/code contract is updated.
- Use session-audit or evaluation skills only when the user explicitly requests an audit/evaluation or the issue explicitly requires it. A bug fix, review, or delivery task alone is not an audit request.
- Treat unit tests, mocks, fakes, stub servers, isolated protocol checks, and local fixtures as regression evidence only. They do not prove live API, authentication, persistence, Vault, or cross-service behavior. For live acceptance, use the actual documented boundary; if it is unavailable, record the exact blocker and do not call it passed.
- Before any Vault mutation, show the human the exact intended diff, affected path/keys, impact, and reason, then wait for explicit confirmation for that Vault change. Keep credentials out of the diff and records.
- Use project-documented run commands and environment names unchanged in manual verification unless the task explicitly changes them. Verify the actual running URL and port from authoritative project knowledge or the process before writing a command; do not guess or silently substitute ports.
- Do not change an existing PR from ready for review back to draft. Use draft only when creating a new PR or when the user explicitly requests draft.

## Branches and publication

- Start new work from `main` on a `task/<topic>` branch.
- For a fix to an existing PR, branch from that PR's branch and update the existing PR; do not open a duplicate unless asked.
- Commit, push, and open a PR without requesting permission when the repository/remote is in scope. Use a draft PR unless asked for ready review.

## Human reviewability and PR sequencing

- Treat human attention as a finite review budget. Each PR must represent one coherent behavior or one independently verifiable delivery unit that a human can understand, test, and manually verify in one focused review.
- Apply a hard split when a change contains multiple independent outcomes, crosses unrelated product areas, combines separate migration/behavior or infrastructure/application concerns, or cannot be explained and verified as one focused unit. Do not use an arbitrary line-count threshold as a substitute for review judgment.
- Before implementation, write the PR shape: each PR's focused scope, base branch, review position, dependency chain, merge condition, and manual verification boundary.
- Use stacked PRs when a later review unit depends on an earlier one. State the review and merge order explicitly, keep each branch based on its predecessor, and merge from the bottom of the stack upward.
- Use parallel PRs only when the units have no required dependency or conflicting shared change. Give the group a shared label/order and state that its members may be reviewed or merged independently.
- Every PR description must include a `Review and merge order` section identifying this PR's position, base/dependencies, parallel group, merge conditions, and the human-verification focus. Keep unrelated cleanup out of the review unit.

## Routing

- For actual UI/frontend work (not terminal command output), read [frontend workflow](.agents/workflows/frontend.md) in full, then spawn the required frontend implementation subagent.
- For Go CLI (`cmd/`, `internal/`), API-client, authentication, or local-persistence work, read [backend workflow](.agents/workflows/backend.md) in full before implementation.
- For Linear, GitHub, issue, PR, or release work, read [delivery workflow](.agents/workflows/delivery.md) in full.
- For creating or editing an agent skill, read [skill workflow](.agents/workflows/skills.md) in full.

## Task-ID protocol

For a request containing a Linear issue ID such as `AZH-385`:

1. If the immediate operation requires a project credential, load only that allowlisted key from `./.agents/config.md` without printing it. Do not read or create the optional config when a connected tool or stored CLI credential already works.
2. Read [the delivery workflow](.agents/workflows/delivery.md) in full before any task-specific repository search, shell command, connector/API call, or implementation. Before implementation, also read every routing workflow applicable to the target path; changes under `.agents/skills/` require `.agents/workflows/skills.md`.
3. Use the connected Linear tool such as Linear MCP. If it is not immediately visible, discover the available tools first.
4. If the connected Linear tool is genuinely unavailable after discovery, immediately use the documented Linear API fallback. Use the official schema or documentation, load only the required credential without output, and never guess requests or bypass an authorization failure.
5. Read the issue, relations, comments, project, and valid team statuses.
6. If the description is incomplete, analyze it first and update it with the human-readable [Linear issue-description template](.agents/templates/linear-issue-description.md). Keep its top-level structure limited to TL;DR, Process Flow, Before-After, and Implementation Manual Test and Verification. Treat that template as a closed heading schema: copy only its headings in the same order; do not add headings or import sections from another template or repository. Put extra implementation detail in prose or lists under an existing heading, or in the agent-facing comment.
7. Preserve user-supplied reference material (including HTML, screenshots, designs, and examples) verbatim in the appropriate human-description section; never replace, trim, or paraphrase the reference unless the user explicitly asks.
8. After the human description is saved, inspect the relevant code and tests, complete the [Linear agent-comment template](.agents/templates/linear-issue-comment.md), and post it as an agent-facing comment. A mockup, full HTML file, screenshot, or one-line request is reference material, not an implementation-ready contract; the comment must add Category, confirmed code-backed analysis, scope boundaries, implementation plan, Definition of Done, correctness checks, and execution controls.
9. The two Linear updates are a hard readiness gate: private reasoning, a todo list, a chat summary, or a code comment does not satisfy it. Verify both tracker mutations succeeded, then re-read the saved human description and agent comment before creating a branch, editing implementation files, moving the issue active, or delegating implementation.
10. Treat the completed human description and agent comment together as the implementation contract. Only then begin implementation. For frontend work, also follow the frontend workflow and its delegation requirement.

- For visual-reference work, UML sequence diagrams are the default for Process Flow, Before, and After. The readiness re-read must confirm the saved tracker rendering itself, exact inline-asset/source pairing, and that each artifact uses UML sequence notation with actors/participants as lifelines, directional messages, and return/activation markers. Only an explicit source request for another diagram type overrides this default; do not substitute a generic architecture, box, or flowchart diagram. API or text-presence counts alone do not satisfy the gate.
- Before the readiness gate passes, validate the saved description against every heading and instruction in `.agents/templates/linear-issue-description.md`: require exactly the four top-level headings in the template's order, Step 0–4 under Implementation Manual Test and Verification, separate Bash verification blocks with per-step exit statuses, no bracketed placeholders/fake records/mocks, and explicit limitations for unrun live flows. API/text-presence counts alone do not establish template compliance.
- Cross-repository guardrails do not transfer automatically: when a recommendation changes agent behavior, apply the equivalent local rule to this repository's `AGENTS.md` before the next task, or record an explicit exception and verify that the active repository already has an equivalent rule.

## Credentials and delivery preflight

- Never print, echo, commit, or transmit `.agents/config.md` or any secret value.
- Do not use a generic file-read tool that renders `.agents/config.md` into a transcript. Load only the allowlisted key needed for the immediate operation through a non-printing secret-loading mechanism; do not recover or copy a literal credential from prior conversation, tool output, memory, or a previous command.
- Do not dot-source config files. Load only allowlisted `KEY=value` entries into the current process environment without output.
- Reading config does not export values into the current shell environment. Never assume a credential environment variable is available.
- Prefer authenticated connectors for Linear and GitHub. Do not manually inject project secrets into `curl` or other direct HTTP commands.
- Before changing code for a task that requires GitHub delivery:
   1. Resolve the intended GitHub CLI executable with a platform-appropriate path-inspection command. Verify that it is the official GitHub CLI, not an npm package, shell alias, or wrapper.
   2. Run a non-interactive authentication and repository-access check with that resolved executable, without overriding a working stored credential.
   3. Verify the GitHub connector can access the repository if it will be used.
- If authentication or repository access fails, stop before implementation, record the blocker in Linear, and tell the user exactly which credential/integration must be fixed.
- Never invoke interactive `gh auth login` in an unattended agent workflow.
- Do not write, edit, generate, or stage implementation files until the GitHub delivery preflight succeeds and a `task/<topic>` branch has been created from `main`. If unrelated work makes that unsafe, use an isolated worktree or stop and report the blocker.
- Never report a check as passed, a build as successful, or a task as complete unless the recorded command exited successfully. State pre-existing failures separately with the exact command and affected path; do not describe a partial compile or filtered output as a successful build.
- Before declaring Vault, Casdoor, PostgreSQL, or another configured runtime dependency unavailable, inspect the relevant project knowledge and configuration sources and use any available read-only infrastructure capability. Distinguish a rejected credential from an unavailable service and from an undeployed application endpoint.
- Before staging and again before handoff, inspect `git status --short` and preserve unrelated files. Put generated screenshots, browser traces, lint captures, and other diagnostics outside the repository or in an ignored temporary directory; remove only artifacts created by the current task.

## CLI boundary

- Nala Labs owns Casdoor/provider exchange and session JWT issuance. The CLI consumes the one-time exchange result and session endpoint; it does not mint login tokens or own provider credentials.
- The CLI is the unified client for Nala Labs and `nala-svc`. Reuse the Nala Labs-issued session JWT across documented service boundaries; `nala-svc` consumes it and does not mint a second login token.
- Preserve the documented local session contract, restrictive file permissions, state validation, and token-free output. Never print, commit, or transmit session-file contents.
- The Go module is at the repository root. Use the root-level commands in the backend workflow, not a sibling service's `backend/` path or Make targets.
- Respect explicit user restrictions on browser use. If a required live login or visual verification cannot be performed under those restrictions, record it as not run and do not claim live acceptance.
