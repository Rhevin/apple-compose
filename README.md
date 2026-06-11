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

`up` · `down` · `ps` · `logs` · `pull` · `exec` · `run` · `stop` · `start` · `restart` · `kill` · `rm` · `cp` · `top` · `stats` · `images` · `port` · `config` · `ls` · `prune` · `login` · `logout`

| Flag | Description |
|---|---|
| `-f, --file` | Compose file (default: `docker-compose.yml`) |
| `-p, --project-name` | Override project name |
| `--profile` | Enable service profiles |
| `--dry-run` | Print commands without running (`up`) |
| `--wait` | Wait for services to reach running state (`up`, default 30s) |
| `--no-deps` | Skip dependency services (`up`) |

Not implemented: `pause`, `unpause`, `events`, `wait`, `watch`, `scale`, `commit`, `push`.

## Compose file support

**Supported:** `image`, `ports`, `environment`, `volumes`, `depends_on`, `command`, `deploy.resources.limits`, `shm_size`, `stop_signal`, `stop_grace_period`, `profiles`

**Ignored / skipped:** `build` (warn + skip — pull pre-built images instead), `healthcheck` (use `--wait`), `restart` (no `--restart` in Apple CLI), `secrets`, `configs`, `extends`

## Limitations

- **Pull-only** — no `build:` support; pre-build with `docker build` or `container build`
- **No restart policy** — services won't auto-restart on crash
- **Named volumes** persist at `~/.apple-compose/volumes/<project>/` after `down`
- **virtiofs** — `chown`/`chmod` on mounts fails; workaround: `PGDATA=/tmp/pgdata` for postgres
- **DNS** — hostname resolution within project networks is unreliable; use IPs or published ports
- **No multi-file compose** or automatic `docker-compose.override.yml` merging

### Networking

Creates a per-project network (`<project>_default`). On macOS 26+, containers attach to it and can reach each other by service name when DNS works. Otherwise publish ports and use `localhost`.

### Named volumes

Host path: `~/.apple-compose/volumes/<project>/<volume>/`. Clean up with `rm -rf ~/.apple-compose/volumes/<project>/`.

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
