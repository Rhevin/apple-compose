# apple-compose

[![CI](https://github.com/Rhevin/apple-compose/actions/workflows/ci.yml/badge.svg)](https://github.com/Rhevin/apple-compose/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/Rhevin/apple-compose)](https://github.com/Rhevin/apple-compose/releases/latest)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/Rhevin/apple-compose)](https://goreportcard.com/report/github.com/Rhevin/apple-compose)
[![Platform](https://img.shields.io/badge/platform-macOS%20arm64-lightgrey)](https://github.com/Rhevin/apple-compose#requirements)

Docker Compose-compatible CLI for [Apple Containers](https://github.com/apple/container). Same model as `docker compose` — a CLI that orchestrates a container runtime, not a daemon itself. Uses Apple's runtime instead of Docker Desktop / `dockerd`.

## Requirements

- macOS 26+ (Tahoe), Apple Silicon (arm64)
- [Apple container CLI](https://github.com/apple/container) 1.0.0+, running (`container system start`)

## Install

```sh
curl -fsSL https://raw.githubusercontent.com/Rhevin/apple-compose/main/install.sh | sh
```

Or [download a release](https://github.com/Rhevin/apple-compose/releases/latest), or `make build` from source.

## Quick start

```sh
container system start          # one-time
apple-compose up
apple-compose ps
apple-compose logs web --follow
apple-compose down
```

## Commands

`up` · `down` · `ps` · `logs` · `pull` · `exec` · `run` · `stop` · `start` · `restart` · `kill` · `rm` · `cp` · `top` · `stats` · `images` · `port` · `config` · `ls` · `prune` · `login` · `logout` · `completion`

| Flag | Description |
|---|---|
| `-f, --file` | Compose file(s), merge left-to-right (repeatable) |
| `-p, --project-name` | Override project name |
| `--profile` | Enable service profiles |
| `--dry-run` | Print commands without running (`up`) |
| `--wait` | Wait for `depends_on` conditions and health (`up`, default 30s) |
| `--no-deps` | Skip dependency services (`up`) |
| `--force-recreate` | Recreate containers even if unchanged (`up`) |
| `--remove-orphans` | Remove containers for services no longer in the compose file (`up`) |
| `-d, --detach` | Accepted for script compatibility (`up`, always detached) |
| `-v, --volumes` | Remove named volume data for this project (`down`) |

`config` also supports `--services`, `--quiet`, and `--format json`.

Not implemented: `pause`, `unpause`, `events`, `wait`, `watch`, `scale`, `commit`, `push`.

## Compose file support

**Supported:** `image`, `ports`, `environment`, `env_file`, `volumes`, `depends_on` (`service_started`, `service_healthy`, `service_completed_successfully`), `healthcheck`, `command`, `entrypoint`, `user`, `working_dir`, `cap_add`, `cap_drop`, `tmpfs`, `read_only`, `ulimits`, `init`, `extra_hosts`, `deploy.resources.limits`, `shm_size`, `stop_signal`, `stop_grace_period`, `profiles`

**Warned + skipped:** `build` (pull pre-built images instead), `restart` (no `--restart` in Apple CLI)

**Not implemented:** `secrets`, `configs`, `extends`, custom `networks:`

Unsupported keys in a compose file trigger warnings on `up` and `config`.

## Limitations

- **Pull-only** — no `build:` support; pre-build with `docker build` or `container build`
- **No restart policy** — services won't auto-restart on crash
- **Named volumes** persist at `~/.apple-compose/volumes/<project>/` after `down` (use `down -v` to remove)
- **virtiofs** — `chown`/`chmod` on mounts fails; postgres named volumes auto-set `PGDATA=/tmp/pgdata`
- **Service discovery** — peer IPs are injected into `/etc/hosts`; running peers are recreated when a new service joins the stack
- **No automatic `compose.override.yml`** — pass override files explicitly with `-f`
- **Config drift** — compose field changes tracked via config-hash trigger recreation; use `--force-recreate` to override

### Networking

Creates a per-project network (`<project>_default`). Containers get peer service names via `/etc/hosts` (and `extra_hosts`). `ps` shows each container's project-network IP in the ADDRESS column.

### Named volumes

Host path: `~/.apple-compose/volumes/<project>/<volume>/`. Clean up with `apple-compose down -v` or `rm -rf ~/.apple-compose/volumes/<project>/`.

### Registry auth

```sh
apple-compose login registry.example.com
apple-compose logout ghcr.io
```

## Benchmarks

Apple Container vs OrbStack on M-series — see [benchmark/README.md](benchmark/README.md). Run locally: `./benchmark/benchmark.sh`.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md). Quick checks: `make fmt && make test && make lint`.

## License

MIT
