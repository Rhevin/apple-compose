# apple-compose

[![CI](https://github.com/Rhevin/apple-compose/actions/workflows/ci.yml/badge.svg)](https://github.com/Rhevin/apple-compose/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/Rhevin/apple-compose)](https://github.com/Rhevin/apple-compose/releases/latest)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/Rhevin/apple-compose)](https://goreportcard.com/report/github.com/Rhevin/apple-compose)
[![Platform](https://img.shields.io/badge/platform-macOS%20arm64-lightgrey)](https://github.com/Rhevin/apple-compose#requirements)

Docker Compose-compatible orchestrator for [Apple Containers](https://github.com/apple/container).

Run your `docker-compose.yml` on Apple Silicon without Docker Desktop or OrbStack — no daemon, no VM overhead beyond Apple's own container runtime.

```
apple-compose up
apple-compose ps
apple-compose logs web --follow
apple-compose down
```

## Requirements

- macOS 15 (Sequoia) or later
- Apple Silicon (arm64)
- [Apple container CLI](https://github.com/apple/container) installed and running

## Install

### curl (quickest)

```sh
curl -fsSL https://raw.githubusercontent.com/Rhevin/apple-compose/main/install.sh | sh
```

Pin a specific version:

```sh
VERSION=v0.1.0 curl -fsSL https://raw.githubusercontent.com/Rhevin/apple-compose/main/install.sh | sh
```

### Build from source

```sh
git clone https://github.com/Rhevin/apple-compose
cd apple-compose
make build
sudo mv apple-compose /usr/local/bin/
```

### Manual download

Download the latest release from [GitHub Releases](https://github.com/Rhevin/apple-compose/releases/latest), then:

```sh
tar -xzf apple-compose_*_darwin_arm64.tar.gz
sudo mv apple-compose /usr/local/bin/
```

## Quick start

```sh
# 1. Start Apple's container system (one-time setup)
container system start

# 2. Run your compose project
apple-compose up

# 3. Check running services
apple-compose ps

# 4. Tail logs from a service
apple-compose logs web --follow

# 5. Tear everything down
apple-compose down
```

## Commands

| Command | Description |
|---|---|
| `up` | Create and start all services |
| `down` | Stop and remove containers + network |
| `ps` | List containers for this project |
| `logs <service>` | Fetch logs from a service |
| `pull [service...]` | Pull service images without starting |
| `exec <service> <cmd>` | Execute a command in a running container |
| `run <service> [cmd]` | Run a one-off command on a service |
| `stop [service...]` | Stop containers without removing |
| `start [service...]` | Start stopped containers |
| `restart [service...]` | Restart service containers |
| `kill [service...]` | Force-stop containers with a signal |
| `rm [service...]` | Remove stopped containers |
| `cp <service:path> <local>` | Copy files between container and host |
| `top [service...]` | Show running processes inside containers |
| `stats [service...]` | Live resource usage (CPU, memory, I/O) |
| `images [service...]` | List images used by the project |
| `port <service> <port>` | Print the public port for a binding |
| `config` | Print resolved compose file |
| `ls` | List all apple-compose projects on this machine |
| `prune` | Remove stopped containers, unused images, networks, volumes |
| `login [registry]` | Log in to a container registry |
| `logout <registry>` | Log out from a container registry |

### Global flags

| Flag | Default | Description |
|---|---|---|
| `-f, --file` | `docker-compose.yml` | Compose file path |
| `-p, --project-name` | directory name | Override project name |
| `--env-file` | | Path to an explicit env file |
| `--profile` | | Enable service profiles (repeatable) |

### `up` flags

| Flag | Default | Description |
|---|---|---|
| `--dry-run` | false | Print commands without executing |
| `--wait` | 30s | Time to wait for each service to reach running state (`0` to disable) |

### `up` additional flags

| Flag | Default | Description |
|---|---|---|
| `--no-deps` | false | Start only named services, skip dependencies |

### `prune` flags

| Flag | Default | Description |
|---|---|---|
| `--force` | false | Skip confirmation prompt |
| `-a, --all` | false | Remove all unused images (also implies `--networks`) |
| `--images` | false | Remove dangling images only |
| `--networks` | false | Remove unused networks |
| `--volumes` | false | Remove named volume data for this project |

> **Docker Compose equivalent:** `docker system prune -f -a` →  `apple-compose prune --force -a --volumes`
> Note: Docker Compose uses `-f` as shorthand for `--force` on prune. In apple-compose, `-f` is already taken by `--file` (global flag), so `--force` has no shorthand.

### `run` / `exec` flags

| Flag | Description |
|---|---|
| `-t, --tty` | Allocate a TTY |
| `-i, --interactive` | Keep STDIN open |
| `--rm` | Remove container after run (default: true for `run`) |
| `-e, --env` | Set additional env vars |
| `--entrypoint` | Override the entrypoint (`run` only) |

## Compose file support

### Supported keys

| Key | Notes |
|---|---|
| `image` | Pull-only — pre-built images only |
| `ports` | Host:container port mapping |
| `environment` | Env vars; `.env` file auto-loaded |
| `volumes` | Bind mounts and named volumes |
| `depends_on` | Startup dependency ordering |
| `command` | Override container command |
| `deploy.resources.limits` | CPU and memory limits |
| `restart` | Parsed but ignored — see limitations |
| `profiles` | Supported via `--profile` flag |

### Unsupported keys (with behaviour)

| Key | Behaviour |
|---|---|
| `build` | **Warning printed, service skipped.** Apple's container builder currently lacks outbound network access — pre-build your images and reference them via `image:` instead. |
| `healthcheck` | Ignored. `--wait` polls for `running` status as a substitute. |
| `restart` | Warning printed, flag dropped. Apple container CLI does not yet expose a `--restart` flag. |
| `secrets` / `configs` | Not implemented. |
| `extends` | Not implemented. |

## Limitations

### Commands not available

These Docker Compose commands have no equivalent in the current Apple container CLI and are not implemented:

| Command | Reason |
|---|---|
| `pause` / `unpause` | Apple container CLI has no pause/unpause support |
| `events` | No real-time event stream in Apple container CLI |
| `wait` | No `container wait` command |
| `watch` | No file-watch/rebuild loop — no build support in v0.1 |
| `scale` | Multiple replicas not yet supported |
| `commit` | No `container commit` command |
| `push` | No `container image push` command |

### Platform limitations

| Limitation | Detail |
|---|---|
| **macOS 26+ required for networking** | `container network` commands only exist on macOS 26 (Tahoe). On macOS 15 (Sequoia), containers start but cannot reach each other by service name. Use published ports as a workaround. |
| **Apple Silicon only** | Apple's container runtime is arm64-only. No x86 support. |
| **No restart policy** | `restart: always` / `on-failure` / `unless-stopped` are silently unsupported by the Apple container CLI — a warning is printed. Services will not auto-restart on crash. |
| **No build support** | `build:` keys are detected and skipped with a warning. Pre-build images with `docker build` or `container build` and push to a registry. |
| **Named volumes not cleaned up by `down`** | Named volumes persist at `~/.apple-compose/volumes/<project>/` after `down`. Run `rm -rf ~/.apple-compose/volumes/<project>/` to clean up manually. |
| **Anonymous volumes not cleaned up** | Apple container CLI does not auto-remove anonymous volumes even with `--rm` (unlike Docker). |
| **No multi-file compose** | `-f file1.yml -f file2.yml` merge is not yet supported. |
| **No `docker-compose.override.yml`** | Automatic override file merging is not implemented. |
| **virtiofs mounts don't support `chown`/`chmod`** | Apple containers use virtiofs for volume mounts. Images that `chown` their data directory on startup (e.g. postgres, mysql) will fail with `Operation not permitted`. Workaround: set `PGDATA=/tmp/pgdata` (or equivalent) to keep data inside the container, or avoid named volumes for these services. |

## Networking and DNS

On **macOS 26 (Tahoe)+**, `apple-compose` creates a shared project network and attaches all containers to it. Services can reach each other by name:

```yaml
services:
  web:
    image: nginx:alpine
    environment:
      DB_HOST: db        # resolves via DNS within the project network
  db:
    image: postgres:16-alpine
```

On **macOS 15 (Sequoia)**, network commands are not available. Containers start successfully but service-name DNS does not work. Use published ports as a workaround:

```yaml
services:
  web:
    image: myapp:latest
    environment:
      DB_HOST: localhost
      DB_PORT: 5432       # map to the host-published port
  db:
    image: postgres:16-alpine
    ports:
      - "5432:5432"       # publish so web can reach it via localhost
```

## Named volumes

Named volumes are stored on the host under:

```
~/.apple-compose/volumes/<project>/<volume-name>/
```

```sh
# Inspect
ls ~/.apple-compose/volumes/myapp/pgdata/

# Clean up after down
rm -rf ~/.apple-compose/volumes/myapp/
```

## Registry authentication

```sh
# Interactive login
apple-compose login registry.example.com

# CI / non-interactive (password from stdin)
echo $TOKEN | apple-compose login ghcr.io -u myuser --password-stdin

# Log out
apple-compose logout ghcr.io
```

Credentials are stored by Apple's container CLI and reused automatically on `pull` and `up`.

## Architecture

```
apple-compose CLI
    ↓
compose-go (parse docker-compose.yml — same library Docker uses)
    ↓
backend/apple.go (builds container CLI arguments)
    ↓
Apple container CLI (exec.Command)
```

No daemon. No socket. Purely CLI invocations against Apple's native container runtime.

## Comparison

| Feature | apple-compose | Container-Compose |
|---|---|---|
| Language | Go | Swift |
| Compose parser | compose-go (official spec) | Hand-rolled YAML structs |
| `ps` | ✅ | ❌ |
| `logs` | ✅ | ❌ |
| `pull` | ✅ | ❌ |
| `exec` / `run` | ✅ | ❌ |
| `stop` / `start` / `restart` | ✅ | ❌ |
| `kill` / `rm` | ✅ | ❌ |
| `cp` / `top` / `stats` | ✅ | ❌ |
| `config` / `ls` / `images` | ✅ | ❌ |
| Registry login/logout | ✅ | ❌ |
| Named volumes | ✅ | ❌ |
| Project networking | ✅ (macOS 26+) | ✅ |
| Global `--project-name` / `--profile` | ✅ | ❌ |
| Tests | ✅ | ✅ |
| Homebrew | ⏳ not yet | ✅ |

## Performance

Benchmarks run on Apple Silicon (M-series) comparing **apple-compose** (Apple Container runtime) vs **OrbStack**, using the same methodology as [repoflow.io/blog/apple-containers-vs-docker-desktop-vs-orbstack](https://www.repoflow.io/blog/apple-containers-vs-docker-desktop-vs-orbstack). Each test averaged over 20 runs on `alpine:3.20`.

### Container startup time _(lower is better)_

```
OrbStack  ████░░░░░░░░░░░░░░░░  0.23s
Apple     █████████████░░░░░░░  0.74s
```

### CPU — single-thread (sysbench events/s) _(higher is better)_

```
OrbStack  ████████████████████  11,318 ev/s
Apple     ██████████████████░░  10,486 ev/s
```

### CPU — multi-thread (sysbench events/s) _(higher is better)_

```
OrbStack  ████████████████████  87,296 ev/s
Apple     █████████░░░░░░░░░░░  39,970 ev/s
```

### Memory throughput (sysbench MiB/s) _(higher is better)_

```
OrbStack  ████████████████████  87,536 MiB/s
Apple     ██████████████░░░░░░  62,979 MiB/s
```

### Disk sequential read — bind mount (fio MiB/s) _(higher is better)_

```
OrbStack  ████████████████████  6,400 MiB/s
Apple     █████████████░░░░░░░  4,357 MiB/s
```

### Small file workflow — 1000 files _(lower is better)_

```
OrbStack  █████████░░░░░░░░░░░  0.98s
Apple     ████████████████████  2.08s
```

> **[Full analysis and findings →](benchmark/README.md)** — explains why the gaps exist and what they actually mean.
>
> Run it yourself: `./benchmark/benchmark.sh` — requires OrbStack and Apple Container installed.

## Roadmap

| Item | Blocked by |
|---|---|
| Homebrew tap | Need stable release + tap repo setup |
| `restart` policy | Apple container CLI `--restart` flag not yet available |
| `build:` key support | Apple container builder lacks outbound network access |
| `healthcheck:` key | No native health check API in Apple container CLI |
| `--scale` / multiple replicas | Not yet supported |
| `pause` / `unpause` | Not supported by Apple container CLI |
| Service-name DNS on macOS 15 | Requires userspace DNS resolver sidecar |

## Contributing

```sh
# Build
make build

# Unit tests (no container CLI required)
make test

# Lint
make lint

# Coverage report
make coverage-html

# Dry-run against the sample compose file
./apple-compose -f testdata/docker-compose.yml up --dry-run

# Test a goreleaser config change locally
make release-dry
```

### Integration tests

Integration tests run real containers and require `container system start` first.

```sh
# Run all integration tests (timeout 20 min)
make test-integration

# Run a specific test
go test -tags integration -v -timeout 20m ./integration/ -run TestLifecycle

# Run without make
go test -tags integration -v -timeout 10m ./integration/
```

Tests skip automatically if the `container` CLI is not available or you are not on Apple Silicon.

PRs welcome. Please add tests for new behaviour.

## License

MIT
