# Architecture

`cmd/nala` is the command dispatcher for the CLI's service command groups.
`internal/auth` owns the HTTP client, callback server, state validation, browser
launch, code exchange, and session endpoint calls. `internal/config` owns the
local session file. As endpoint coverage expands, Nala Labs and `nala-svc`
operations should remain behind service-specific clients or packages and
reuse the shared authentication/configuration boundary.

The login path is deliberately short-lived:

1. Listen on a loopback port and create random CLI state.
2. Ask Nala Labs for a provider authorization URL and open it.
3. Accept only a callback with the matching state and a code.
4. Exchange the code with Nala Labs and save the returned session locally.

The CLI never logs callback query values or prints the stored token. The
server-side provider exchange and session JWT issuance remain in Nala Labs;
`nala-svc` remains a consumer of that authenticated boundary. The login path
described above is the current foundation, not the limit of the command set.
