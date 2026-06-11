# Contributing to apple-compose

Thanks for taking the time to contribute.

## Before you start

- Check [open issues](https://github.com/rhevin/apple-compose/issues) to avoid duplicate work
- For large changes, open an issue first to discuss the approach
- All contributions must pass `go test ./...` and `go vet ./...`

## Requirements

- Go 1.22+
- macOS 26+ with Apple Silicon (for integration testing)
- Apple container CLI 1.0.0+ installed (`container system start`)

## Setup

```sh
git clone https://github.com/rhevin/apple-compose
cd apple-compose
go mod download
make build
```

## Making changes

### Adding a new command

1. Create `cmd/<name>.go`
2. Use `loadProject()` to load the compose file — never call `compose.Load()` directly
3. Use `serviceTargets(order, args)` for optional per-service filtering
4. Use `serviceNotFound(name, project)` for missing-service errors
5. Register in `cmd/root.go` inside `rootCmd.AddCommand(...)`
6. Add tests for any non-trivial logic

### Adding backend functionality

- All Apple container CLI calls live in `internal/backend/apple.go`
- Use `exec.Command("container", ...)` — never hardcode a path
- Check the [Apple container CLI docs](https://github.com/apple/container/blob/main/docs/command-reference.md) before assuming a flag exists
- Commands that do not exist in the Apple container CLI: `pause`, `unpause`, `wait`, `commit`, `image push`, `top`, `restart` — do not attempt to call them

### Modifying compose loading

- Compose parsing lives in `internal/compose/loader.go`
- Always go through the `Options` struct — do not add parameters directly to `Load()`
- The loader uses `compose-go` (the official Docker Compose spec library) — do not hand-roll YAML parsing

## Testing

```sh
# Run all tests
go test ./...

# Run with verbose output
go test ./... -v

# Run a specific package
go test ./internal/backend/...
```

Tests must not require the Apple container CLI to run — mock or avoid any calls to `exec.Command("container", ...)` in unit tests. The `RunArgs()` function is the primary unit-testable surface for backend logic.

## Code style

- Run `go fmt ./...` before committing
- Run `go vet ./...` and fix all warnings
- No comments explaining what code does — only comments explaining why (non-obvious constraints, workarounds)
- No error swallowing — log or return every error
- Keep commands thin: logic belongs in `internal/`, not `cmd/`

## Commit messages

Use conventional commits:

```
feat: add `watch` command
fix: correct port mapping for UDP protocols
docs: update limitations table in README
test: add loader test for multi-profile compose file
chore: bump compose-go to v2.12.0
```

Subject line: 72 chars max, imperative mood, no period.

## Pull requests

- One PR per feature or fix
- Include a description of what changed and why
- Link any related issues
- Ensure `go test ./...` passes
- Update `README.md` if you add a command, flag, or change behaviour
- Update `AGENTS.md` if you change architecture, conventions, or constraints

## Reporting bugs

Open an issue with:
- macOS version (`sw_vers`)
- Apple container CLI version (`container --version`)
- apple-compose version (`apple-compose --version`)
- The `docker-compose.yml` that triggered the bug (redact secrets)
- Full output including any error messages
- Output of `apple-compose up --dry-run` if relevant
