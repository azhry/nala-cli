# CLI authentication

`nala login` uses the Nala Labs CLI authentication boundary:

1. `GET /api/auth/cli/authorize` receives the CLI loopback `redirect_uri` and
   random `state`. Nala Labs validates the loopback target and returns `302`
   to the provider authorization URL.
2. The browser authenticates at Nala Labs. Nala Labs handles the provider
   callback and returns `302` to the loopback URL with a one-time code and the
   original CLI state. A denied login returns an error and state instead.
3. `POST /api/auth/cli/exchange` consumes the one-time code and returns the
   Nala Labs session response.
4. The CLI stores the session token in the local session file. `nala user info`
   sends it as `Authorization: Bearer ...` to `GET /api/auth/session`.

Only `http://localhost`, `http://127.0.0.1`, and `http://[::1]` callbacks with
an explicit port are accepted. Query strings and fragments are rejected. The
callback state is checked before accepting a code, and the local callback
server shuts down after the login attempt.

Live browser verification requires a running Nala Labs backend, its configured
Casdoor application, and a real test account. Local unit tests do not replace
that boundary check.
