---
name: create-pr
description: >
  Create a pull request following apple-compose PR conventions.
  Use when the user says "create a PR", "open a PR", "make a pull request",
  or "submit my changes". Enforces conventional commit title format,
  correct branch naming, 1 commit per PR rule, and required checklist.
---

Create PR for apple-compose. Follow these rules exactly.

## Branch naming

Format: `<type>/<short-description>`

| Type | When |
|---|---|
| `feat/` | New feature |
| `fix/` | Bug fix |
| `chore/` | Maintenance, deps, config |
| `docs/` | Documentation only |
| `test/` | Tests only |
| `ci/` | CI/workflow changes |
| `refactor/` | Code refactor, no behaviour change |

Examples: `feat/watch-command`, `fix/exec-flag-parsing`, `ci/add-dependabot`

## PR title format

```
<type>: <subject>
```

Rules:
- Types: `feat`, `fix`, `docs`, `test`, `chore`, `ci`, `refactor`, `perf`
- Subject: imperative mood, max 72 chars, no period
- No scope required

Good: `feat: add watch command`
Bad: `Add watch command`, `feat(cmd): Add watch command.`

## 1 commit per PR rule

Each PR must have exactly 1 commit. Squash before pushing if needed:

```sh
# Check how many commits ahead of master
git log origin/master..HEAD --oneline
# Must show exactly 1 line

# If multiple commits, squash them
git rebase -i origin/master
# In editor: keep first as 'pick', change rest to 'squash'
# Write a single conventional commit message

# Force push after squash
git push origin <branch> --force-with-lease
```

## PR description template

```markdown
## Summary

- <what changed and why — bullet points>
- <second change if applicable>

## Test plan

- [ ] `make test` passes
- [ ] `make test-integration` passes (if touching runtime code)
- [ ] `make lint` passes
- [ ] `make fmt` run — no formatting issues
- [ ] Dry-run verified: `./apple-compose -f testdata/docker-compose.yml up --dry-run`
```

## Steps

1. Check branch and commit count:
```sh
git branch --show-current          # must NOT be master
git log origin/master..HEAD --oneline  # must be exactly 1 line
```

2. If on master, create a branch first:
```sh
git checkout -b <type>/<description>
```

3. If more than 1 commit, squash:
```sh
git rebase -i origin/master
git push origin <branch> --force-with-lease
```

4. Run checks:
```sh
make fmt
make test
make lint
```

5. Push and create PR:
```sh
git push origin <branch-name>

gh pr create \
  --title "<type>: <subject>" \
  --body "$(cat <<'EOF'
## Summary

- <what changed and why>

## Test plan

- [ ] `make test` passes
- [ ] `make test-integration` passes (if touching runtime code)
- [ ] `make lint` passes
- [ ] `make fmt` run — no formatting issues
- [ ] Dry-run verified: `./apple-compose -f testdata/docker-compose.yml up --dry-run`
EOF
)" \
  --base master
```

## Checklist before creating PR

- [ ] Exactly 1 commit on the branch
- [ ] Branch name follows `<type>/<description>` convention
- [ ] PR title follows `<type>: <subject>` format (max 72 chars)
- [ ] PR description has Summary and Test plan sections filled in
- [ ] `make fmt` run — no gofmt issues
- [ ] `make test` passes
- [ ] `make lint` passes
- [ ] No binaries committed (`apple-compose`, `apple-compose-test`)
- [ ] No secrets or `.env` files staged

## Common mistakes to avoid

- Do NOT push directly to `master` — branch protection will reject it
- Do NOT have multiple commits — squash to 1 before opening PR
- Do NOT skip `make fmt` — CI will fail on gofmt issues
- Do NOT commit `apple-compose` or `apple-compose-test` binaries
- Always branch from latest master: `git checkout master && git pull && git checkout -b <branch>`
