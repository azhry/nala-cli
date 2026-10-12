# nala-cli

`nala-cli` is the Go command-line client for the Nala platform. It is intended
to expose command groups for the documented endpoints of both Nala Labs and
`nala-svc`, using the shared authenticated session boundary.

The CLI supports browser-based login, current-user inspection, and the
authenticated app/deployment lifecycle:

```text
nala login
nala user info
nala config show
nala config set --api-url URL --svc-url URL
nala app list [--page N --page-size N]
nala app get --id N
nala app deploy --id N --source-ref REF --idempotency-key KEY
nala app monitor --deployment-id N [--cursor N --follow]
nala app delete --id N
```

`nala login` starts a short-lived loopback callback server, opens the first-party
Nala Labs `/login` page in the browser, validates the returned state, exchanges
the one-time callback code, and stores the Nala Labs session locally. It confirms the
signed-in user without printing the session token. `nala user info` reads that
local session and calls the Nala Labs session endpoint, printing the
authenticated user and entitlements.

The `app` commands use the same stored Nala Labs session bearer. Listing,
details, and deletion call Nala Labs; deployment creation, snapshots, and
optional event streaming call `nala-svc`. `nala app monitor` prints a snapshot
by default and newline-delimited event JSON when `--follow` is supplied.

## Requirements

- Go 1.26 or newer
- Nala Labs and `nala-svc` environments for the command groups being used
- A Nala Labs account for live login verification

The CLI stores non-secret service endpoints in `config.json` beside the local
session file. The documented local defaults work without configuration. To
persist those defaults explicitly, or to replace them for another environment:

```bash
nala config set --api-url 'http://127.0.0.1:8080' --svc-url 'http://127.0.0.1:8081'
nala config show
```

The defaults are `http://127.0.0.1:8080` for Nala Labs and
`http://127.0.0.1:8081` for `nala-svc`. `NALA_API_BASE_URL` and
`NALA_SVC_BASE_URL` remain supported as process-local overrides for CI and
one-off checks, and take precedence over the config file.

`NALA_FRONTEND_URL` is not a CLI setting and is not read by `nala`. The Nala
Labs backend uses its own `FRONTEND_URL` process setting when it starts.

Configuration and session data are stored at the platform user-config
directory under `nala/config.json` and `nala/session.json` with restrictive
local permissions. Set `NALA_CONFIG_DIR` to override the parent directory
during development or testing. Do not commit those files or copy session
contents into logs, tickets, or pull requests.

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
