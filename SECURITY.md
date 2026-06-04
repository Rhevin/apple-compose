# Security Policy

## Supported versions

| Version | Supported |
|---|---|
| latest (`main`) | ✅ |
| older releases | ❌ — please upgrade |

## Reporting a vulnerability

**Do not open a public GitHub issue for security vulnerabilities.**

Email: **security@rhevin.dev** (or open a [GitHub private security advisory](https://github.com/rhevin/apple-compose/security/advisories/new))

Include:
- Description of the vulnerability
- Steps to reproduce
- Potential impact
- Suggested fix if you have one

You will receive an acknowledgement within **48 hours** and a resolution timeline within **7 days**.

## Scope

### In scope

- Command injection via compose file values passed unsanitised to `exec.Command`
- Path traversal in volume mount or `cp` handling
- Credential leakage (registry passwords, env vars) in logs or error output
- Privilege escalation via container flags

### Out of scope

- Vulnerabilities in Apple's container CLI itself — report those to [Apple](https://support.apple.com/en-us/102549)
- Vulnerabilities in `compose-go` — report those to the [compose-spec project](https://github.com/compose-spec/compose-go/security)
- Social engineering
- Denial-of-service against the local machine

## Security model

apple-compose is a local CLI tool. It:

- Shells out to `container` (Apple's CLI) — it does **not** run as root or request elevated privileges
- Reads `docker-compose.yml` from the current directory — treat untrusted compose files the same as untrusted shell scripts
- Stores registry credentials via Apple's container CLI credential store — apple-compose itself never writes credentials to disk
- Does not open network ports or run a daemon

### Compose file trust

A `docker-compose.yml` can specify:
- Arbitrary `command` values executed inside containers
- Volume mounts that expose host filesystem paths to containers
- Environment variables including secrets

**Only run `apple-compose up` on compose files you trust**, the same as you would with `docker compose up`.

### Environment variables

Secrets passed via `environment:` in a compose file are forwarded as `--env KEY=VALUE` arguments to `container run`. They appear in the process argument list and may be visible to other processes on the same machine via `ps`. Use `.env` files with restricted permissions (`chmod 600 .env`) for sensitive values.

## Dependency security

Dependencies are pinned in `go.sum`. To audit:

```sh
go list -m all
govulncheck ./...    # requires: go install golang.org/x/vuln/cmd/govulncheck@latest
```
