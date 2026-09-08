# Nala Labs API contract consumed by this CLI

The authoritative server-side contract is maintained in the Nala Labs
repository at `.agents/knowledge/api-contracts.md`. The CLI consumes:

| Method | Endpoint | Purpose |
| --- | --- | --- |
| `GET` | `/api/auth/cli/authorize` | Validate loopback callback and start browser login. |
| `GET` | `/api/auth/callback` | Provider callback; returns to the CLI transaction when applicable. |
| `POST` | `/api/auth/cli/exchange` | Consume a one-time CLI code and return a session. |
| `GET` | `/api/auth/session` | Resolve the stored bearer session and current user. |

The CLI should treat session and provider tokens as secrets. Error messages
must not include response bodies that could contain credentials.
