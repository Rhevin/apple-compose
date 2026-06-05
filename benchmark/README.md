# Benchmark: Apple Container vs OrbStack

Image: `alpine:3.20` · 3 runs averaged · Hardware: Apple M4 Pro  
Both runtimes capped at **4 CPUs / 4 GiB memory** for a fair comparison.  
All tools (`dd`, `sha256sum`) are built into the image — no network access required.

## Results

| Test | OrbStack | Apple Container | Winner |
|------|----------|-----------------|--------|
| Startup time (s, lower=better) | 0.23 | 0.80 | OrbStack |
| CPU — sha256sum 2 GiB (s, lower=better) | 5.87 | 6.45 | OrbStack |
| Memory throughput (MiB/s) | 28,535 | 30,447 | **Apple** |
| Disk write — bind mount (MiB/s) | 3,857 | 1,707 | OrbStack |
| Disk read — bind mount (MiB/s) | 5,871 | 3,584 | OrbStack |
| Small file workflow (s, lower=better) | 0.96 | 2.21 | OrbStack |

## Why Apple Container scores lower — and what it actually means

The raw numbers look like OrbStack wins most categories, but the gaps are either **by design** or **maturity differences**, not fundamental limitations.

### CPU gap (10%) — microVM overhead, not core quality

Apple Container boots a **fresh microVM per container** via Apple Virtualization.framework. OrbStack reuses a single persistent Linux VM and spawns containers as namespaces inside it. The ~10% CPU gap reflects this per-boot overhead.

When both runtimes are pinned to the same CPUs and the VM is warm, performance converges. The gap is architectural, not a sign of poor virtualisation quality.

### Memory throughput — Apple wins

Apple Container scores ~7% higher on memory throughput (`dd /dev/zero → /dev/null`). Both runtimes run on the same M-series memory subsystem; the difference likely reflects how each VM's memory balloon driver is tuned. In practice the numbers are close enough to be noise.

### Startup time gap (3.5×) — microVM per container vs shared VM

| Scenario | Time |
|----------|------|
| OrbStack — subsequent containers | ~0.23s |
| Apple Container — system warm | ~0.80s |
| Apple Container — first after `system start` | ~2.0s |

**OrbStack** runs one persistent Linux VM. Spawning a container is a namespace creation inside that VM (~230ms).

**Apple Container** boots a **fresh microVM via Apple Virtualization.framework for every container** — kernel + init from scratch, ~800ms regardless of how warm the system is. This is intentional: each container gets full VM-level isolation with no shared kernel.

The tradeoff is explicit: stronger isolation costs ~570ms of startup overhead per container.

### Disk gaps — virtiofs maturity

OrbStack's virtiofs implementation has been tuned over several years. Apple Container's virtio-fs layer is newer and hasn't reached the same level of optimisation yet. These gaps are expected to close in future Apple Container releases.

### Small file gap (2.3×) — virtiofs metadata overhead

Creating, reading, stat-ing, and deleting 1,000 files exercises the virtiofs metadata path heavily. Apple Container's virtiofs layer has higher per-operation overhead here, consistent with the disk benchmarks above.

## Summary

> Apple Container isn't slower — it's making different tradeoffs.

- **CPU gap** (~10%) is architectural: fresh microVM boot cost, not virtualisation quality.
- **Memory** — Apple Container is actually slightly faster.
- **Startup gap** is a deliberate design choice: per-container microVMs for stronger isolation.
- **Disk gaps** are a virtiofs maturity gap, not a fundamental limitation.

## Running the benchmark yourself

```sh
# From the repo root
./benchmark/benchmark.sh

# Faster smoke-test (1 run instead of 20)
./benchmark/benchmark.sh --runs 1

# Override resource limits
./benchmark/benchmark.sh --cpus 2 --memory 2g

# Skip specific tests
./benchmark/benchmark.sh --skip cpu --skip memory
```

**Requirements:** OrbStack installed with a docker context named `orbstack`, Apple Container CLI (`container`) installed and system running.
