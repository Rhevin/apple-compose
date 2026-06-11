# apple-compose context

Docker Compose CLI for Apple container. No daemon. arm64 only. v0.4.0.

## Stack
Go 1.22+, compose-go/v2, cobra. Shells out to `container` binary.

## Architecture
```
cmd/ → internal/compose (Load, TopologicalOrder) → internal/backend/apple.go → container CLI
```

## Key files
- `cmd/root.go` — global flags, `loadProject()`, `resolveProjectName()`, `resolveTargets()`
- `cmd/util.go` — `serviceNotFound`, `topologicalOrder`, `serviceTargets`, `resolveContainerName`
- `internal/backend/apple.go` — `RunArgs`, `Up`, `Down`, `PS`, `ContainerStatus`, `ListContainersForProject`

## Naming
- Container: `{project}-{service}` e.g. `testdata-web`
- Network: `{project}_default`
- Volumes: `~/.apple-compose/volumes/{project}/{volume}/`
- Labels: `com.apple-compose.project`, `com.apple-compose.service`

## Commands (all implemented)
up, down, ps, logs, pull, exec, run, stop, start, restart, kill, rm, cp, top, stats, images, port, config, ls, prune, login, logout

## Constraints
- `build:` key → warn + skip (pull-only)
- No `--restart` flag in Apple CLI → warn + drop
- Named volumes + virtiofs → `chown` fails → use `PGDATA=/tmp/pgdata`
- `up` idempotent: skip running, restart stopped, create new
- DNS: IP works on macOS 26+, hostname resolution broken (vmnet has no DNS)
- JSON schema (container 1.0.0+): `status.state`, `configuration.id`, `configuration.labels` (map), `configuration.image.reference`, `configuration.publishedPorts`
- Compose parity: `shm_size` → `--shm-size`; `stop_signal`/`stop_grace_period` → `container stop --signal`/`--time`
- Commands needing only project name: use `resolveProjectName()` + `resolveTargets()` not `loadProject()`
- `exec`/`run`: use `FParseErrWhitelist{UnknownFlags: true}` + `SetInterspersed(false)` so container flags (e.g. `--max-time`) aren't parsed by cobra
- `run bench sh -c "..."`: works only for simple commands — `&&` in the shell string causes cobra to choke; use `--entrypoint sh` + `-c "..."` as workaround in scripts
- Benchmark: Apple Container containers cannot `apk add` at runtime (no outbound network in containers). Use built-in tools (`dd`, `sha256sum`) only.

## Test sample
`testdata/docker-compose.yml` — nginx + postgres (`PGDATA=/tmp/pgdata`) + redis
```sh
./apple-compose -f testdata/docker-compose.yml up
./apple-compose -p testdata ps
make test-integration
```

## PR rules
1 commit per PR. Branch: `<type>/<desc>`. Title: `<type>: <subject>` ≤72 chars. Run `make fmt` before commit.
