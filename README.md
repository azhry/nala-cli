# nala-cli

`nala-cli` is the Go command-line client for Nala Labs account authentication.
The first slice supports browser-based login and current-user inspection:

```text
nala login
nala user info
```

`nala login` starts a short-lived loopback callback server, opens the Nala Labs
login page, validates the returned state, exchanges the one-time callback code,
and stores the Nala Labs session locally. It confirms the signed-in user
without printing the session token. `nala user info` reads that local session
and calls the Nala Labs session endpoint, printing the authenticated user and
entitlements.

## Requirements

- Go 1.26 or newer
- A Nala Labs backend with the CLI authentication endpoints enabled
- A Nala Labs account for live login verification

The API base URL defaults to `http://127.0.0.1:8080`. Override it with
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

The CLI depends on the Nala Labs contract documented in
`.agents/knowledge/authentication.md` and
`.agents/knowledge/nala-labs-api.md`. Unit tests use local HTTP fixtures for
regression coverage; they do not prove live provider authentication.

## Repository guidance

Start with [AGENTS.md](AGENTS.md), then the relevant
[project knowledge](.agents/knowledge/index.md). Shared workflows, issue/PR
templates, and bundled skills follow the other Nala services; the
Read the knowledge before changing the authentication boundary. Nala Labs
owns provider exchange and session JWT issuance; this CLI only receives the
one-time exchange result, stores the session locally, and consumes the session
endpoint.
