# Project agent instructions

- Read `.agents/knowledge/` before changing the CLI or its Nala Labs boundary.
- Never print, commit, or transmit session tokens, provider tokens, passwords,
  API keys, or local session-file contents.
- Preserve unrelated dirty files and use a `task/<topic>` branch from `main`.
- Keep changes focused on the requested behavior. Run `go test ./...` and
  `go build ./cmd/nala` before handoff.
- Use Bash for documented verification commands. Unit tests and local HTTP
  fixtures are regression evidence only; live browser authentication must be
  reported separately.
- Nala Labs owns Casdoor/provider exchange and Nala Labs session JWT issuance.
  Do not add a second login-token issuer to this repository.
