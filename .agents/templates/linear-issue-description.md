## TL;DR

[In 1–3 sentences, explain what this issue changes, who is affected, and the observable outcome.]

## Process Flow

All three diagrams in this template must be UML sequence diagrams: use participants/lifelines, directional messages, and activation or return semantics. Do not substitute a generic architecture, box, or flowchart diagram unless the issue explicitly requests that type.

Show the actors and services involved, then mark the implementation change with the highlighted region and the IMPLEMENTATION CHANGE annotation. Keep the diagram focused on the normal path.

![Rendered Process Flow](<uploaded Linear SVG asset URL>)

The published issue must show the rendered SVG inline, not only diagram syntax. Keep the exact SVG source used for this image in the native Linear collapsible block below. Do not use a `mermaid` fence for SVG source: Linear renders that fence as a diagram.

+++ Diagram source

~~~xml
<svg>
[paste the exact SVG source used for the inline image]
</svg>
~~~

+++

Legend: the pale-yellow region and annotation identify where the implementation changes. Replace the example actors, services, messages, and change annotation with the real flow before publishing the issue.

## Before-After

### Before

![Rendered Before diagram](<uploaded Linear SVG asset URL>)

+++ Before diagram source

~~~xml
<svg>
[paste the exact SVG source used for the inline image]
</svg>
~~~

+++

### After

![Rendered After diagram](<uploaded Linear SVG asset URL>)

+++ After diagram source

~~~xml
<svg>
[paste the exact SVG source used for the inline image]
</svg>
~~~

+++

Replace both diagrams with the real before and after paths. Keep the same actors and services where they are unchanged, and make the changed sequence visibly distinct.

After saving, reopen the issue in Linear and visually verify that all three images render inline, each source block is collapsed by default, and expanding a source shows literal SVG/XML text rather than a rendered diagram. API text counts alone do not establish this acceptance criterion.

## Implementation Manual Test and Verification

Write the verification directly in this description and in the PR description as separate Bash steps. Do not create a script file or paste one large bulk script. Replace every bracketed value with the real staging target, fixture, path, payload, and observed response before handoff. Never use fake records, mocks, or copied secrets. Do not use `set -o pipefail`, `set -e`, `set -Eeuo pipefail`, or another fail-fast wrapper; keep each step independently runnable, print its exit status, and leave the interactive terminal open when a command fails.

### Step 0 — Load the verified environment

Verify the actual Nala Labs URL from project knowledge and the running process before publishing a completed handoff. Replace the example below if the documented target differs. The CLI default is a development URL, not evidence that a backend is running. Do not source private environment files or print session data.

~~~bash
export NALA_API_BASE_URL='http://127.0.0.1:8080'
status=$?
printf 'environment setup exit status: %s\n' "$status"
~~~

Expected: exit 0. Record the verified non-secret target and environment separately.

### Step 1 — Authenticate when the flow requires it

For authentication changes, use the CLI's real browser login with the documented account; do not copy a server password-login request or extract the stored token. Respect the user's browser restrictions and record the flow as not run when it is not authorized.

~~~bash
go run ./cmd/nala login
status=$?
printf 'login exit status: %s\n' "$status"
~~~

Expected: exit 0 after successful browser authentication and a signed-in confirmation without a token. Record the actual sanitized result separately; this is not an observed pass.

### Step 2 — Exercise the changed behavior

~~~bash
go run ./cmd/nala user info
status=$?
printf 'user info exit status: %s\n' "$status"
~~~

Expected: exit 0 and the authenticated account and entitlements, without a token. Replace this example for non-authentication tasks with the task's real command and observable contract.

### Step 3 — Exercise the required regression, error, or ownership case

~~~bash
[one Bash command for the real CLI regression/error/ownership scenario]
status=$?
printf 'regression exit status: %s\n' "$status"
~~~

Replace the command before publication. State the expected exit status and sanitized output for the real boundary case, then record the observed result or exact reason it was not run.

### Step 4 — Record evidence

For every step, record the exact command, exit status, sanitized response, and environment/fixture identity. Keep credentials, JWTs, API keys, and unrelated diagnostics out of the issue and PR.
