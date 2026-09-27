# Defaults

- Module: `github.com/azhry/nala-cli`
- Language: Go 1.26+
- Current commands: `nala login`, `nala user info`, `nala config show`, `nala config set`
- Service targets: Nala Labs and `nala-svc`; endpoint-specific command groups are added as their contracts are implemented.
- Current auth API base URL: config file `apiBaseURL`, then `NALA_API_BASE_URL`, default `http://127.0.0.1:8080`
- Current `nala-svc` base URL: config file `svcBaseURL`, then `NALA_SVC_BASE_URL`, default `http://127.0.0.1:8081`
- Configuration override: `NALA_CONFIG_DIR`; default platform config directory plus
  `nala/config.json` and `nala/session.json`
- Callback bind address: `127.0.0.1` on an operating-system-selected port

The CLI is a client of Nala Labs and `nala-svc`. Nala Labs owns provider
credentials, provider exchange, and session JWT issuance; `nala-svc` consumes
the documented session boundary. The CLI does not issue JWTs or provision
applications, databases, deployments, or Vault records.
