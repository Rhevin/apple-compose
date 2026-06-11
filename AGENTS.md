# AGENTS.md — apple-compose

Docker Compose CLI for Apple container CLI. macOS 26+, arm64. No daemon. v0.4.0. Requires container CLI 1.0.0+.

## Architecture
```
cmd/ → internal/compose → internal/backend/apple.go → container CLI
```

## Key files
| File | Purpose |
|---|---|
| `cmd/root.go` | Global flags, `loadProject()`, `resolveProjectName()`, `resolveTargets()` |
| `cmd/util.go` | `serviceNotFound`, `topologicalOrder`, `serviceTargets`, `resolveContainerName` |
| `internal/backend/apple.go` | All container ops, `appleContainer` JSON struct |

## Conventions
- Container: `{project}-{service}` · Network: `{project}_default`
- Volumes: `~/.apple-compose/volumes/{project}/{volume}/`
- Labels: `com.apple-compose.project`, `com.apple-compose.service`
- Always `loadProject()` in cmd/ — never `compose.Load()` directly
- Commands not needing compose file: `resolveProjectName()` + `resolveTargets()`

## Adding a command
1. Create `cmd/<name>.go`
2. Needs compose → `loadProject()` · Only project name → `resolveProjectName()`
3. Pass-through flags → `FParseErrWhitelist{UnknownFlags: true}`
4. Register in `cmd/root.go` `init()`
5. Add tests

## Constraints
- `build:` key → error from `RunArgs()`, warn + skip
- No `--restart` → warn to stderr, drop flag
- Named volumes + virtiofs → `chown` fails → `PGDATA=/tmp/pgdata` workaround
- `up` idempotent: skip running, restart stopped, create new
- DNS: IP works macOS 26+, hostname broken (vmnet no DNS, no `--hostname` flag)
- JSON (container 1.0.0+): `status.state`, `configuration.id`, `configuration.labels` (map); pre-1.0 used string `status`
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
