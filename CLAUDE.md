# Claude instructions for apple-compose

At the start of every session, read `.claude/CONTEXT.md` for full project context.

## Key conventions
- Container names: `{project}-{service}` (e.g. `myapp-web`)
- Backend binary: always `container` (Apple's CLI)
- v0.1 is pull-only — warn and skip any service with a `build:` key
- Teardown order is reverse of startup (topological) order

## What to avoid
- No daemon process
- No Docker socket shim
- No Kubernetes
- No custom build support in v0.1
