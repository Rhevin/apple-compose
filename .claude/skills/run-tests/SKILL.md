---
name: run-tests
description: >
  Run the apple-compose test suite. Use when the user says "run tests",
  "test it", "check tests", "run integration tests", or asks to verify
  the project is working. Handles both unit and integration tests with
  correct setup steps.
---

Run apple-compose tests. Two modes: unit (fast, no container CLI needed) and integration (real containers).

## Unit tests

```sh
make test
```

No setup needed. Runs in seconds.

## Integration tests

Require Apple container system running first:

```sh
# 1. Check container system is running
container system start

# 2. Run integration tests (timeout 20 min)
make test-integration
```

## What each test covers

| Test | What it tests |
|---|---|
| `TestConfig_*` | compose file parsing and config output |
| `TestImages` | image list from compose file |
| `TestPort` | port binding resolution |
| `TestExec` | exec into running container |
| `TestRun_OneOff` | one-off container run (note: `run` uses `SetInterspersed(false)` — flags after service name pass through to container) |
| `TestLifecycle` | full up → ps → logs → stop → start → restart → down |
| `TestUp_DryRun` | dry-run output correctness |
| `TestUp_NoDeps` | --no-deps flag |
| `TestPS_NoComposefile` | ps without compose file |
| `TestLS` | project listing |
| `TestPull_*` | image pull commands |

## Common failures

| Failure | Fix |
|---|---|
| `container: command not found` | Install Apple container CLI and run `container system start` |
| `network not found` | Run `apple-compose down` to clean up, then retry |
| `db did not become ready` | Postgres init takes time on first run — re-run, it will pass |
| `no containers found` | Previous test left containers running — run `apple-compose -f testdata/docker-compose.yml down` |

## Cleanup stuck containers

```sh
apple-compose -f testdata/docker-compose.yml down
container prune
```

## Run a single test

```sh
go test -tags integration -v -timeout 20m ./integration/ -run TestLifecycle
```
