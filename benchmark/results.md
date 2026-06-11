# Container Benchmark: OrbStack vs Apple Container

**Runs per test:** 3
**Image:** `alpine:3.20`
**Date:** 2026-06-05
**Host:** arm64 — Apple M4 Pro
**Resource limits:** 4 CPUs, 4g memory (identical for both runtimes)

| Test | OrbStack | Apple Container | Winner |
|------|----------|-----------------|--------|
| Startup time (s, lower=better) | 0.2330 | 0.7963 | OrbStack |
| CPU sha256 2GiB (s, lower=better) | 5.8747 | 6.4540 | OrbStack |
| Memory throughput (MiB/s) | 28535.4667 | 30446.9333 | Apple |
| Disk write (MiB/s) | 3857.0667 | 1706.6667 | OrbStack |
| Disk read (MiB/s) | 5870.9333 | 3584.0000 | OrbStack |
| Small files workflow (s, lower=better) | 0.9610 | 2.2050 | OrbStack |
