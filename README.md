# nala-cli

`nala-cli` is the Go command-line client for the Nala platform. It is intended
to expose command groups for the documented endpoints of both Nala Labs and
`nala-svc`, using the shared authenticated session boundary.

The first implementation slice establishes browser-based login and
current-user inspection:

```text
nala login
nala user info
```

`nala login` starts a short-lived loopback callback server, opens the Nala Labs
login page, validates the returned state, exchanges the one-time callback code,
and stores the Nala Labs session locally. It confirms the signed-in user
without printing the session token. `nala user info` reads that local session
and calls the Nala Labs session endpoint, printing the authenticated user and
entitlements. These are the currently implemented commands, not the complete
CLI surface.

## Requirements

- Go 1.26 or newer
- Nala Labs and `nala-svc` environments for the command groups being used
- A Nala Labs account for live login verification

The current authentication and session client uses the Nala Labs API base URL,
which defaults to `http://127.0.0.1:8080`. Override it with
`NALA_API_BASE_URL`:

```bash
export NALA_API_BASE_URL='http://127.0.0.1:8080'
```

Session data is stored at the platform user-config directory under
`nala/session.json` with restrictive local permissions. Set
`NALA_CONFIG_DIR` to override the parent directory during development or
testing. Do not commit that file or copy its contents into logs, tickets, or
pull requests.

## Development

```bash
go test ./...
go build ./cmd/nala
```

The current auth slice depends on the Nala Labs contract documented in
`.agents/knowledge/authentication.md` and
`.agents/knowledge/nala-labs-api.md`. Future command groups must document the
target service, endpoint contract, and configured base URL before they are
implemented. Unit tests use local HTTP fixtures for regression coverage; they
do not prove live provider authentication or cross-service behavior.

## Repository guidance

Start with [AGENTS.md](AGENTS.md), then the relevant
[project knowledge](.agents/knowledge/index.md). Shared workflows, issue/PR
templates, and bundled skills follow the other Nala services.
Read the knowledge before changing the service boundaries. Nala Labs owns
provider exchange and session JWT issuance; `nala-svc` consumes that session
boundary, and this CLI calls both services without minting a second login
token.
