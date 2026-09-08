# nala-cli

`nala-cli` is the Go command-line client for the Nala platform. It is intended
to expose command groups for the documented endpoints of both Nala Labs and
`nala-svc`, using the shared authenticated session boundary.

The CLI supports browser-based login, current-user inspection, and the
authenticated app/deployment lifecycle:

```text
nala login
nala user info
nala app list [--page N --page-size N]
nala app get --id N
nala app deploy --id N --source-ref REF --idempotency-key KEY
nala app monitor --deployment-id N [--cursor N --follow]
nala app delete --id N
```

`nala login` starts a short-lived loopback callback server, opens the Nala Labs
login page, validates the returned state, exchanges the one-time callback code,
and stores the Nala Labs session locally. It confirms the signed-in user
without printing the session token. `nala user info` reads that local session
and calls the Nala Labs session endpoint, printing the authenticated user and
entitlements.

The `app` commands use the same stored Nala Labs session bearer. Listing,
details, and deletion call Nala Labs; deployment creation, snapshots, and
optional event streaming call `nala-svc`. `nala app monitor` prints a snapshot
by default and newline-delimited event JSON when `--follow` is supplied.

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

The `nala-svc` base URL defaults to `http://127.0.0.1:8081`. Override it with
`NALA_SVC_BASE_URL` when using a different service environment:

```bash
export NALA_SVC_BASE_URL='http://127.0.0.1:8081'
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

The command groups depend on the Nala Labs and `nala-svc` contracts documented
in `.agents/knowledge/authentication.md`,
`.agents/knowledge/nala-labs-api.md`, and
`.agents/knowledge/nala-svc-api.md`. Unit tests use local HTTP fixtures for
regression coverage; they do not prove live provider authentication,
cross-service ownership, deployment execution, or persistence behavior.

## Repository guidance

Start with [AGENTS.md](AGENTS.md), then the relevant
[project knowledge](.agents/knowledge/index.md). Shared workflows, issue/PR
templates, and bundled skills follow the other Nala services.
Read the knowledge before changing the service boundaries. Nala Labs owns
provider exchange and session JWT issuance; `nala-svc` consumes that session
boundary, and this CLI calls both services without minting a second login
token.
