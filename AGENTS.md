# AGENTS.md — apple-compose

Context for AI coding agents working in this repository.

## What this project is

`apple-compose` is a Docker Compose-compatible CLI orchestrator for [Apple's native container CLI](https://github.com/apple/container). It parses `docker-compose.yml` using the official `compose-go` library and maps each service to `container run` shell invocations.

**Target:** macOS 15+, Apple Silicon only. No daemon. No Docker socket. Current release: v0.3.0.

## Architecture

```
apple-compose CLI (cobra)
    ↓
internal/compose   — Load(), TopologicalOrder()  (wraps compose-go)
    ↓
internal/backend   — RunArgs(), Up(), Down(), Stop(), Start(), ...
    ↓
Apple container CLI (exec.Command("container", ...))
```

Key files:

| File | Purpose |
|---|---|
| `main.go` | Entry point, injects version |
| `cmd/root.go` | Global flags, `loadProject()`, `resolveProjectName()`, `resolveTargets()` helpers |
| `cmd/*.go` | One file per command |
| `cmd/util.go` | Shared: `serviceNotFound`, `topologicalOrder`, `serviceTargets`, `resolveContainerName` |
| `internal/compose/loader.go` | Loads compose file via compose-go |
| `internal/compose/order.go` | Topological sort for depends_on |
| `internal/backend/apple.go` | Builds `container` CLI args, wraps all container operations, `appleContainer` JSON struct |

## Conventions

- **Container names:** `{project}-{service}` e.g. `myapp-web`
- **Network name:** `{project}_default`
- **Named volume path:** `~/.apple-compose/volumes/{project}/{volume}/`
- **Labels on every container:**
  - `com.apple-compose.project={project}`
  - `com.apple-compose.service={service}`
- **Backend binary:** always the string `"container"` (Apple's CLI)
- **Project loading:** always use `loadProject()` in `cmd/` — never call `compose.Load()` directly
- **Commands that don't need compose file** (`ps`, `top`, `logs`, `stop`, `start`, `stats`): use `resolveProjectName()` + `resolveTargets()` — no hard failure if compose file missing

## Adding a new command

1. Create `cmd/<name>.go`
2. If command needs compose file: use `loadProject()`
3. If command only needs project name: use `resolveProjectName()` + `resolveTargets()`
4. Use `serviceNotFound(name, project)` for missing service errors
5. Use `FParseErrWhitelist{UnknownFlags: true}` if command passes flags through to container
6. Register in `cmd/root.go` `init()` → `rootCmd.AddCommand(...)`
7. Add tests if the command has non-trivial logic

## Key constraints

- **Pull-only** — detect `build:` key and return error from `RunArgs()`, warn and skip
- **No `--restart` flag** — Apple container CLI doesn't support it; warn to stderr, drop the flag
- **Network commands** — `container network create/delete` only on macOS 26+; `EnsureNetwork` handles graceful fallback
- **Named volumes + virtiofs** — `os.MkdirAll` before `--volume`; images that `chown` data dirs (postgres, mysql) fail with `Operation not permitted` — use `PGDATA=/tmp/pgdata` workaround
- **`up` is idempotent** — checks container status before creating; skips running, restarts stopped
- **DNS between services** — IP connectivity works on macOS 26+, but hostname resolution does not (Apple's vmnet has no DNS server). No `--hostname` flag available.
- **`container list --format json`** — actual JSON schema: `status` (lowercase), `configuration.id`, `configuration.labels` (map), `configuration.image.reference`

## What Apple container CLI supports

Commands we use: `run`, `stop`, `start`, `delete`, `list`, `logs`, `exec`, `copy`, `kill`, `stats`, `image pull`, `network create/delete/list`, `registry login/logout`

Commands that do NOT exist (do not attempt to call):
- `container pause` / `container unpause`
- `container wait` / `container commit` / `container image push`
- `container top` — use `exec <name> ps -eo user,pid,ppid,...` instead
- `container restart` — use stop + start
- `container run --restart` — flag does not exist

## Running tests

```sh
make test                  # unit tests, no container CLI needed
make test-integration      # real containers, requires: container system start
make lint                  # golangci-lint
make fmt                   # gofmt + goimports
```

Tests:
- `internal/backend/apple_test.go` + `volumes_test.go` — RunArgs, labels, volumes
- `internal/compose/order_test.go` + `loader_test.go` — ordering, parsing
- `cmd/util_test.go` — serviceTargets, needsQuote
- `integration/` — end-to-end with real containers (build tag: `integration`)

## Building

```sh
make build        # ./apple-compose binary
make install      # copies to /usr/local/bin/
make release-dry  # test goreleaser locally
```

## Sample compose file

`testdata/docker-compose.yml` — nginx + postgres + redis. Uses `PGDATA=/tmp/pgdata` for postgres (virtiofs workaround). No named volumes.

```sh
./apple-compose -f testdata/docker-compose.yml up --dry-run
./apple-compose -f testdata/docker-compose.yml up
./apple-compose -p testdata ps
./apple-compose -p testdata top
./apple-compose -f testdata/docker-compose.yml down
```
