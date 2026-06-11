# Container Benchmark: OrbStack vs Apple Container

**Runs per test:** 1
**Image:** `alpine:3.20`
**Date:** 2026-06-05
**Host:** arm64 — Apple M4 Pro

| Test | OrbStack | Apple Container | Winner |
|------|----------|-----------------|--------|
| Startup time (s, lower=better) | 0.2250 | 0.7390 | OrbStack |
| CPU single-thread (ev/s) | 11318.1500 | 10485.8400 | OrbStack |
| CPU multi-thread (ev/s) | 87295.7400 | 39969.5300 | OrbStack |
| Memory throughput (MiB/s) | 87536.0200 | 62978.5500 | OrbStack |
| Disk seq-read (MiB/s) | 6400.00 | 4357.45 | OrbStack |
| Small files workflow (s, lower=better) | 0.9790 | 2.0780 | OrbStack |
