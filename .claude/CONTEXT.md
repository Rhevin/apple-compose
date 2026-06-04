# apple-compose — Project Context

## What this is

A Docker Compose-compatible CLI orchestrator for Apple's native `container` CLI.
Replaces Docker Desktop / OrbStack for Apple Silicon Mac developers.
Open source, no daemon, lightweight. Current release: **v0.3.0**

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

## Current status (v0.3.0)

### Commands implemented

- [x] `up` — start all services, idempotent (skip running, restart stopped)
- [x] `down` — stop and remove in reverse dependency order
- [x] `ps` — list containers, works without compose file (-p flag)
- [x] `logs` — tail logs, accepts service name or full container name
- [x] `pull` — pull images without starting
- [x] `exec` — exec into running container (flags pass-through)
- [x] `run` — one-off command (flags pass-through)
- [x] `stop` / `start` / `restart` / `kill` / `rm`
- [x] `cp` — copy files between container and host
- [x] `top` — docker compose top format, works without compose file
- [x] `stats` — live resource usage
- [x] `images` / `port` / `config` / `ls`
- [x] `prune` — remove stopped containers, images, networks, volumes
- [x] `login` / `logout` — registry auth

### Known limitations

- **No DNS between services** — IP connectivity works on macOS 26+ (shared network), but hostname resolution doesn't. Apple's vmnet has no DNS server and no `--hostname` flag.
- **No restart policy** — Apple container CLI has no `--restart` flag yet
- **Named volumes + virtiofs** — images that `chown` data dirs crash. Use `PGDATA=/tmp/pgdata` for postgres
- **No build support** — `build:` key detected and warned, service skipped

## Project structure

```
apple-compose/
  main.go
  cmd/
    root.go           -- cobra root, global flags, loadProject(), resolveProjectName()
    util.go           -- serviceNotFound, topologicalOrder, serviceTargets, resolveContainerName
    up.go / down.go / ps.go / logs.go / ...
    unsupported.go    -- stubs for pause, unpause, events, wait, scale
  internal/
    compose/
      loader.go       -- Load() parses compose file
      order.go        -- TopologicalOrder() dependency sort
    backend/
      apple.go        -- RunArgs(), Up(), Down(), PS(), ContainerStatus(), ListContainersForProject()
  testdata/
    docker-compose.yml  -- nginx + postgres + redis test stack
    nginx.conf
  .claude/
    skills/           -- run-tests, create-pr, release
  .github/
    workflows/        -- ci, lint, release, pr-title
    PULL_REQUEST_TEMPLATE.md
```

## Container naming

- Container: `{project}-{service}` e.g. `testdata-web`
- Network: `{project}_default` e.g. `testdata_default`
- Volumes: `~/.apple-compose/volumes/{project}/{volume}/`
- Labels: `com.apple-compose.project`, `com.apple-compose.service`
