# Defaults

- Module: `github.com/azhry/nala-cli`
- Language: Go 1.26+
- Commands: `nala login`, `nala user info`
- API base URL: `NALA_API_BASE_URL`, default `http://127.0.0.1:8080`
- Session override: `NALA_CONFIG_DIR`; default platform config directory plus
  `nala/session.json`
- Callback bind address: `127.0.0.1` on an operating-system-selected port

The CLI is a client of Nala Labs. It does not own provider credentials, issue
JWTs, or provision applications, databases, deployments, or Vault records.
