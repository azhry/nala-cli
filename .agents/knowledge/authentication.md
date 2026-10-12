# CLI authentication

`nala login` establishes the shared authenticated session used by command
groups that call Nala Labs and `nala-svc`:

1. `GET /api/auth/cli/authorize` receives the CLI loopback `redirect_uri` and
   random `state`. Nala Labs validates the loopback target and returns `302`
   to the first-party Nala Labs `/login` page with an opaque CLI state.
2. The browser authenticates through the Nala Labs login page. The page calls
   `POST /api/auth/cli/complete` with the Nala Labs session bearer and opaque
   CLI state; Nala Labs returns a one-time callback URL with a code and the
   original CLI state. A denied login returns an error and state instead.
3. `POST /api/auth/cli/exchange` consumes the one-time code and returns the
   Nala Labs session response.
4. The CLI stores the session token in the local session file. `nala user info`
   sends it as `Authorization: Bearer ...` to `GET /api/auth/session`.

Only `http://localhost`, `http://127.0.0.1`, and `http://[::1]` callbacks with
an explicit port are accepted. Query strings and fragments are rejected. The
callback state is checked before accepting a code, and the local callback
server shuts down after the login attempt.

Live browser verification requires a running Nala Labs backend, its first-party
login page, and a real test account. Local unit tests do not replace that
boundary check.
