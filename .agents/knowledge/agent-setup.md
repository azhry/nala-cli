# Agent setup baseline

The CLI uses the shared Nala service operating setup, not a shortened standalone rule list. The baseline is `azhry/nala-labs` commit `dd26cf146dc20c4cf88f53d43ee3c76ea6bbf7ed`, compared with the local nala-svc, nala-trace, and nala-grow setups during AZH-509.

## Shared components

- Root [AGENTS.md](../../AGENTS.md): scope, credential safety, issue readiness, branch/delivery preflight, review order, verification evidence, and task-owned artifact cleanup.
- Four [workflows](../workflows/): backend/Go, delivery, frontend, and skills. Frontend applies only to actual UI work, not terminal command output.
- Five [templates](../templates/): backend, frontend, and test PR descriptions; Linear human issue description and agent comment. The Linear description retains the four-section heading schema and Step 0–4 verification structure.
- Five [skill packages](../skills/): GitHub delivery, Linear issue management, manual-test verification, Codex session audit, and KiloCode session audit, including their linked references, helpers, and existing tests. Audit packages are available for explicitly requested audits; ordinary delivery does not invoke them.

## Intentional CLI adaptations

- The Go module lives at the root, with `cmd/nala`, `internal/auth`, and `internal/config`; no `backend/` directory, database migrations, or integration Make target is assumed.
- Nala Labs owns providers, Casdoor, JWT issuance, and server persistence. CLI knowledge remains local to this repository and documents only the client contract, loopback callback, and protected local session file.
- Authentication examples use `go run ./cmd/nala login` and `user info`, not a copied server password-login endpoint. They capture immediate exit statuses, never source private environment files, and distinguish expected results from observed acceptance.
- GitHub checks follow the current host permission policy rather than requiring a particular sandbox override. The GitHub skill description is YAML-quoted so its colon is valid frontmatter. Linear discovers connectors before fallback and loads only a required credential through a non-printing allowlisted mechanism; examples never embed keys.
- Explicit user browser restrictions remain in force. Unrun browser login or visual checks must be recorded as not run.
- Shared private `.agents/config.md`, `.agents/.env`, provider/test credentials, session files, logs, caches, and captured audit data are not copied or committed. Authenticated connectors and stored CLI credentials do not require loading optional project secrets.
- Nala Labs' OpenWiki skill, installed-wiki instructions, and installation metadata are excluded because this CLI has no generated OpenWiki installation. Do not claim the CLI has an index or scheduled wiki workflow that does not exist.

## Maintaining consistency

For future shared-rule updates, compare the matching files against the sibling service baseline and bring over applicable safety and delivery changes as a focused update. Preserve CLI paths, ownership boundaries, and private configuration. Record intentional deviations here rather than silently dropping shared sections. Do not synchronize secrets or generated artifacts.
