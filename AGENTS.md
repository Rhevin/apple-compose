# AGENTS.md — apple-compose

Docker Compose CLI for Apple container CLI. macOS 26+, arm64. Shells out to `container` (like docker compose → dockerd); requires `container system start`.

## Architecture
```
cmd/ → internal/compose (Load, TopologicalOrder) → internal/backend → container CLI
```

## Key files
| File | Purpose |
|---|---|
| `cmd/root.go` | Global flags, `loadProject()`, CLI version check, completions |
| `cmd/up.go` | `depends_on` waits, `--force-recreate`, `--remove-orphans` |
| `cmd/util.go` | `serviceNotFound`, `topologicalOrder`, `stopOptionsForService` |
| `internal/compose/loader.go` | Multi-file `-f` merge via compose-go |
| `internal/backend/apple.go` | `RunArgs`, `Up`, `Down`, `PS`, container JSON |
| `internal/backend/config.go` | Config-hash labels for drift detection |
| `internal/backend/healthcheck.go` | `healthcheck` + `depends_on` condition polling |
| `internal/backend/hosts.go` | Peer `/etc/hosts` injection, hosts-hash drift, `RefreshPeerHosts` |
| `internal/backend/volumes.go` | Postgres/mysql virtiofs workarounds (`PrepareService`) |
| `internal/backend/unsupported.go` | Warn on unsupported compose keys |
| `internal/backend/version.go` | Warn if container CLI < 1.0.0 or legacy JSON |

## Conventions
- Container: `{project}-{service}` · Network: `{project}_default`
- Volumes: `~/.apple-compose/volumes/{project}/{volume}/`
- Labels: `com.apple-compose.project`, `com.apple-compose.service`, `com.apple-compose.config-hash`, `com.apple-compose.hosts-hash`
- Always `loadProject()` in cmd/ — never `compose.Load()` directly
- Commands not needing compose file: `resolveProjectName()` + `resolveTargets()`
- `RunArgs` expects `PrepareService()` first for DB volume workarounds

## Adding a command
1. Create `cmd/<name>.go`
2. Needs compose → `loadProject()` · Only project name → `resolveProjectName()`
3. Pass-through flags → `FParseErrWhitelist{UnknownFlags: true}`
4. Register in `cmd/root.go` `init()`
5. Add tests

## Constraints
- `build:` key → warn + skip (pull-only)
- `restart:` → unsupported key warning (no `--restart` in Apple CLI)
- Named volumes + virtiofs → `chown` fails → `PrepareService` auto-sets `PGDATA` for postgres
- `up` idempotent: skip running; recreate on config-hash or hosts-hash drift, or `--force-recreate`
- DNS: inject peer IPs into `/etc/hosts`; `RefreshPeerHosts` recreates stale peers after each `up`
- JSON (container 1.0.0+): `status.state`, `status.networks`, `configuration.publishedPorts`
- Compose mapped in `RunArgs`: env, env_file, entrypoint, user, workdir, caps, tmpfs, read_only, ulimits, init, shm_size, deploy.resources.limits
- Stop: `stop_signal` / `stop_grace_period` → `container stop --signal` / `--time`
- Network create: suppress stderr, ignore already-exists

## Apple container CLI
**Use:** `run`, `stop`, `start`, `delete`, `list`, `logs`, `exec`, `copy`, `kill`, `stats`, `image pull`, `network create/delete/list`, `registry login/logout`

**Do NOT call:** `pause`, `unpause`, `wait`, `commit`, `image push`, `top`, `restart` (use stop+start), `run --restart`

## Tests
```sh
make test               # unit, no CLI needed
make test-integration   # real containers, needs: container system start
make fmt                # gofmt + goimports — run before every commit
```

## Build
```sh
make build · make install · make release-dry
./apple-compose -f testdata/docker-compose.yml up --dry-run
```
