# Nala Labs API contract used by the current CLI slice

The authoritative server-side contract is maintained in the Nala Labs
repository at `.agents/knowledge/api-contracts.md`. The CLI consumes:

| Method | Endpoint | Purpose |
| --- | --- | --- |
| `GET` | `/api/auth/cli/authorize` | Validate loopback callback and start browser login. |
| `POST` | `/api/auth/cli/complete` | Complete the first-party Nala Labs browser login for a CLI state. |
| `GET` | `/api/auth/callback` | Provider callback; returns to the CLI transaction when applicable. |
| `POST` | `/api/auth/cli/exchange` | Consume a one-time CLI code and return a session. |
| `GET` | `/api/auth/session` | Resolve the stored bearer session and current user. |

This document covers the current authentication and session foundation, not
the complete Nala Labs endpoint surface. As additional command groups are
implemented, document each service contract separately and identify whether
the command targets Nala Labs or `nala-svc`.

The CLI should treat session and provider tokens as secrets. Error messages
must not include response bodies that could contain credentials.
