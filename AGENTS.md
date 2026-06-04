# AGENTS.md — apple-compose

Context for AI coding agents working in this repository.

## What this project is

`apple-compose` is a Docker Compose-compatible CLI orchestrator for [Apple's native container CLI](https://github.com/apple/container). It parses `docker-compose.yml` using the official `compose-go` library and maps each service to `container run` shell invocations.

**Target:** macOS 15+, Apple Silicon only. No daemon. No Docker socket.

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
| `cmd/root.go` | Global flags, `loadProject()` helper |
| `cmd/*.go` | One file per command |
| `cmd/util.go` | Shared helpers: `serviceNotFound`, `topologicalOrder`, `serviceTargets` |
| `internal/compose/loader.go` | Loads compose file via compose-go |
| `internal/compose/order.go` | Topological sort for depends_on |
| `internal/backend/apple.go` | Builds `container` CLI args, wraps all container operations |

## Conventions

- **Container names:** `{project}-{service}` e.g. `myapp-web`
- **Network name:** `{project}_default`
- **Named volume path:** `~/.apple-compose/volumes/{project}/{volume}/`
- **Labels on every container:**
  - `com.apple-compose.project={project}`
  - `com.apple-compose.service={service}`
- **Backend binary:** always the string `"container"` (Apple's CLI)
- **Project loading:** always use `loadProject()` in `cmd/` — never call `compose.Load()` directly from commands

## Adding a new command

1. Create `cmd/<name>.go`
2. Define a `var <name>Cmd = &cobra.Command{...}` 
3. Use `loadProject()` to load the compose file
4. Use `serviceTargets(order, args)` for optional per-service filtering
5. Use `serviceNotFound(name, project)` for missing service errors
6. Register in `cmd/root.go` `init()` → `rootCmd.AddCommand(...)`
7. Add tests if the command has non-trivial logic

## Key constraints

- **v0.1 is pull-only** — detect `build:` key and return an error from `RunArgs()`, not silently skip
- **No `--restart` flag** — Apple container CLI does not support it; warn to stderr, do not pass the flag
- **Network commands** — `container network create/delete` only exist on macOS 26+; wrap in graceful fallback (check error, warn, continue)
- **Named volumes** — must call `os.MkdirAll` on the host path before passing `--volume` to container
- **`container list --format json`** — use for all programmatic queries; fall back to plain `container list` if JSON fails

## What Apple container CLI supports

Commands we use: `run`, `stop`, `start`, `delete`, `list`, `logs`, `exec`, `copy`, `kill`, `stats`, `image pull`, `network create/delete/list`, `registry login/logout`

Commands that do NOT exist in Apple container CLI (do not attempt to call them):
- `container pause` / `container unpause`
- `container wait`
- `container commit`
- `container image push`
- `container top` (use `exec <name> ps aux` instead)
- `container restart` (use stop + start)

## Running tests

```sh
go test ./...
```

Tests live in:
- `internal/backend/apple_test.go` — RunArgs, ContainerName, NetworkName, volumes
- `internal/backend/volumes_test.go` — named and anonymous volumes
- `internal/compose/order_test.go` — topological ordering, cycle detection
- `internal/compose/loader_test.go` — compose file parsing, env override

All tests are unit tests — no Apple container CLI required to run them.

## Building

```sh
make build    # produces ./apple-compose binary
make test     # go test ./...
make install  # copies binary to /usr/local/bin/
```

## Sample compose file

`testdata/docker-compose.yml` exercises: named volumes, bind mounts, env vars, ports, depends_on, restart warnings. `testdata/nginx.conf` is the bind-mounted config for the web service.

```sh
./apple-compose up --dry-run   # verify without running anything
```
