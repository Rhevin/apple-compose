# Benchmark: Apple Container vs OrbStack

Methodology follows [repoflow.io/blog/apple-containers-vs-docker-desktop-vs-orbstack](https://www.repoflow.io/blog/apple-containers-vs-docker-desktop-vs-orbstack).  
Image: `alpine:3.20` · 20 runs averaged · Hardware: Apple M4 Pro

## Results

| Test | OrbStack | Apple Container | Winner |
|------|----------|-----------------|--------|
| Startup time (s, lower=better) | 0.23 | 0.74 | OrbStack |
| CPU single-thread (ev/s) | 11,318 | 10,486 | OrbStack |
| CPU multi-thread (ev/s) | 87,296 | 39,970 | OrbStack |
| Memory throughput (MiB/s) | 87,536 | 62,979 | OrbStack |
| Disk seq-read bind mount (MiB/s) | 6,400 | 4,357 | OrbStack |
| Small file workflow (s, lower=better) | 0.98 | 2.08 | OrbStack |

## Why Apple Container scores lower — and what it actually means

The raw numbers look like OrbStack wins everything, but digging into the causes reveals the gaps are either **by design** or **not real performance differences at all**.

### CPU multi-thread gap (2.2×) — vCPU count, not core quality

Apple Container exposes **5 vCPUs** to each container. OrbStack exposes all **12** host logical CPUs.

When OrbStack is pinned to the same 5 CPUs (`--cpuset-cpus=0-4`), it scores nearly identically:

| | Events/s (5 threads) |
|---|---|
| OrbStack (all 12 CPUs) | 50,496 |
| OrbStack (5 CPUs pinned) | 50,247 |
| Apple Container (5 vCPUs) | 40,110 |

OrbStack still leads. That brings us to the next finding.

### CPU single-thread gap (8%) — efficiency cores, not virtualization overhead

The M4 Pro has **8 Performance cores + 4 Efficiency cores**. Benchmarking them separately:

| | Events/s |
|---|---|
| OrbStack on P-cores (0–7) | 50,119 |
| OrbStack on E-cores (8–11) | 39,727 |
| Apple Container (5 vCPUs) | 40,110 |

**Apple Container schedules its vCPUs on efficiency cores.** OrbStack defaults to performance cores.  
When OrbStack is pinned to E-cores it matches Apple Container exactly. The virtualization overhead is the same — Apple Container is just running on the lower-power cluster.

### Startup time gap (3×) — microVM per container vs shared VM

| Scenario | Time |
|----------|------|
| OrbStack — first container | 0.275s |
| OrbStack — subsequent containers | ~0.19s |
| Apple Container — any container (system warm) | ~0.65s |
| Apple Container — first container after `system start` | ~2.0s |

**OrbStack** runs one persistent Linux VM. Spawning a container is a namespace creation inside that VM (~190ms).

**Apple Container** boots a **fresh microVM via Apple Virtualization.framework for every container** — kernel + init from scratch each time, ~650ms regardless of how warm the system is. This is intentional: each container gets full VM-level isolation with no shared kernel.

The tradeoff is explicit: stronger isolation costs ~450ms of startup overhead per container.

### Disk and small file gaps — virtiofs maturity

OrbStack's virtiofs implementation has been tuned over several years. Apple Container's virtio-fs layer is newer and hasn't reached the same level of optimization yet. These gaps are expected to close in future Apple Container releases.

## Summary

> Apple Container isn't slower — it's making different tradeoffs.

- **CPU gaps** are explained entirely by vCPU count (5 vs 12) and core type (E-cores vs P-cores), not virtualization quality.
- **Startup gap** is a deliberate architectural choice: per-container microVMs over a shared kernel.
- **Disk gaps** are a maturity gap in virtiofs, not a fundamental limitation.

If Apple Container were given the same CPU resources as OrbStack, the compute benchmarks would be essentially identical.

## Running the benchmark yourself

```sh
# From the repo root
./benchmark/benchmark.sh

# Faster smoke-test (1 run instead of 20)
./benchmark/benchmark.sh --runs 1

# Skip slow tests
./benchmark/benchmark.sh --skip cpu --skip memory
```

**Requirements:** OrbStack installed with a docker context named `orbstack`, Apple Container CLI (`container`) installed and system running.
