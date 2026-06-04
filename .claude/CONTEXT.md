# apple-compose — Project Context

## What this is
A Docker Compose-compatible CLI orchestrator for Apple's native `container` CLI.
Replaces Docker Desktop / OrbStack for Apple Silicon Mac developers.
Open source, no daemon, lightweight.

## Goals
- Parse `docker-compose.yml` via `compose-go`
- Map each service to `apple container run` CLI calls
- Stay OCI-compliant — image pulling delegated to Apple's container tool
- Target: macOS only, Apple Silicon, macOS 15+

## Architecture
```
apple-compose CLI
    ↓
compose-go (parse docker-compose.yml)
    ↓
OCI runtime interface (swappable backend)
    ↓
Apple container CLI (exec.Command)
```

## Tech stack
- Go 1.22+
- `github.com/compose-spec/compose-go/v2` — Compose spec parsing
- `github.com/spf13/cobra` — CLI
- `exec.Command` to shell out to `container` binary
- No daemon process

## Phase 1 scope (v0.1 — pull-only, no custom builds)
- [x] `apple-compose up` — start all services (with `--dry-run`)
- [x] `apple-compose down` — stop and remove in reverse dependency order
- [x] `apple-compose ps` — list running containers
- [x] `apple-compose logs [service]` — tail logs
- [x] `depends_on` topological ordering
- [ ] Shared networking / service-name DNS (hardest unsolved piece — see below)
- [ ] Basic volume mounts — bind mounts only

## Known constraints
- Apple Containers builder lacks outbound network access → `build:` key in compose
  must warn the user and be skipped in v0.1
- No Docker socket shim needed for v0.1
- No Kubernetes, no GUI

## Open problem: Service-to-service DNS
Each Apple container runs in its own micro-VM. Services need to reach each other
by hostname (e.g. `web` → `db`). Options to explore:
1. Userspace DNS resolver sidecar
2. Network bridge / shared vlan
3. /etc/hosts injection at container start

This is the hardest unsolved piece for v0.2.

## Project structure
```
apple-docker-orchestrator/
  main.go
  cmd/
    root.go       -- cobra root, --file flag
    up.go         -- up command, --dry-run flag
    down.go       -- down command (reverse order teardown)
    ps.go         -- ps command
    logs.go       -- logs command, --follow flag
  internal/
    compose/
      loader.go   -- Load() parses compose file via compose-go
      order.go    -- TopologicalOrder() dependency sort
    backend/
      apple.go    -- RunArgs(), Up(), Down(), PS(), Logs()
  docker-compose.yml  -- sample compose file for testing
  go.mod / go.sum
```

## Container naming convention
`{project_name}-{service_name}` e.g. `myapp-db`, `myapp-web`
Project name defaults to directory name or `COMPOSE_PROJECT_NAME` env var.
