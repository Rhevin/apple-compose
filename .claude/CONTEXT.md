# apple-compose context

Docker Compose CLI for Apple container. CLI only — shells out to `container` runtime (like docker compose → dockerd). macOS 26+, arm64. Requires `container system start`.

## Stack
Go 1.22+, compose-go/v2, cobra. Shells out to `container` binary (1.0.0+).

## Architecture
```
cmd/ → internal/compose (Load, TopologicalOrder) → internal/backend → container CLI
```

## Key files
- `cmd/root.go` — global flags, `loadProject()`, CLI version check, completions
- `cmd/up.go` — `depends_on` waits, `--force-recreate`, `--remove-orphans`, `RefreshPeerHosts`
- `cmd/util.go` — `topologicalOrder`, `serviceTargets`, `stopOptionsForService`
- `internal/compose/loader.go` — multi-file `-f` merge via compose-go
- `internal/backend/apple.go` — `RunArgs`, `Up`, `Down`, `PS`, container JSON
- `internal/backend/config.go` — config-hash drift detection (all RunArgs fields)
- `internal/backend/healthcheck.go` — `healthcheck` + `depends_on` condition polling
- `internal/backend/hosts.go` — peer `/etc/hosts` injection, hosts-hash, `RefreshPeerHosts`
- `internal/backend/volumes.go` — postgres/mysql virtiofs workarounds (`PrepareService`)
- `internal/backend/unsupported.go` — warn on unsupported compose keys
- `internal/backend/version.go` — warn if container CLI < 1.0.0 or legacy JSON

## Naming
- Container: `{project}-{service}` e.g. `testdata-web`
- Network: `{project}_default`
- Volumes: `~/.apple-compose/volumes/{project}/{volume}/`
- Labels: `com.apple-compose.project`, `com.apple-compose.service`, `com.apple-compose.config-hash`, `com.apple-compose.hosts-hash`

## Commands (all implemented)
up, down, ps, logs, pull, exec, run, stop, start, restart, kill, rm, cp, top, stats, images, port, config, ls, prune, login, logout, completion

Not implemented (stub commands): pause, unpause, events, wait, watch, scale, commit, push

## Constraints
- `build:` → warn + skip (pull-only)
- `restart:` → unsupported key warning (no `--restart` in Apple CLI)
- Named volumes + virtiofs → `chown` fails → `PrepareService` auto-sets `PGDATA=/tmp/pgdata` for postgres
- `up` idempotent: skip running; recreate on config-hash or hosts-hash drift, or `--force-recreate`
- DNS: inject peer IPs into `/etc/hosts`; `RefreshPeerHosts` recreates stale peers after each `up`
- JSON (container 1.0.0+): `status.state`, `status.networks`, `configuration.publishedPorts`, `configuration.labels` (map)
- Compose in `RunArgs`: env, env_file, entrypoint, user, workdir, caps, tmpfs, read_only, ulimits, init, shm_size, deploy.resources.limits, extra_hosts
- Stop: `stop_signal` / `stop_grace_period` → `container stop --signal` / `--time`
- Always `loadProject()` in cmd/ — never `compose.Load()` directly
- Commands needing only project name: `resolveProjectName()` + `resolveTargets()` not `loadProject()`
- `RunArgs` expects `PrepareService()` first for DB volume workarounds
- `exec`/`run`: `FParseErrWhitelist{UnknownFlags: true}` + `SetInterspersed(false)` for container flags
- No auto `compose.override.yml` — pass override files with `-f`
- Network create: suppress stderr, ignore already-exists

## Apple container CLI
**Use:** `run`, `stop`, `start`, `delete`, `list`, `logs`, `exec`, `copy`, `kill`, `stats`, `image pull`, `network create/delete/list`, `registry login/logout`

**Do NOT call:** `pause`, `unpause`, `wait`, `commit`, `image push`, `top`, `restart` (use stop+start), `run --restart`

## Test sample
`testdata/docker-compose.yml` — nginx + postgres + redis
```sh
./apple-compose -f testdata/docker-compose.yml up
./apple-compose -p testdata ps
make test-integration   # needs: container system start
```

## PR rules
1 commit per PR. Branch: `<type>/<desc>`. Title: `<type>: <subject>` ≤72 chars. Run `make fmt` before commit.
