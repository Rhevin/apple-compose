---
name: release
description: >
  Cut a new release of apple-compose. Use when the user says "release",
  "cut a release", "tag a release", "release new version", or "bump version".
  Determines the next version, generates the tag message from merged PRs,
  tags, and pushes to trigger GoReleaser.
---

Release apple-compose. Execute all steps — do not just show instructions.

## Version convention

Semantic versioning: `vMAJOR.MINOR.PATCH`

| Change | Bump |
|---|---|
| Breaking change | MAJOR |
| New command or feature | MINOR |
| Bug fix, docs, chore | PATCH |

## Steps

### 1. Ensure master is clean and up to date

```sh
git checkout master
git pull origin master
git status  # must be clean
```

### 2. Find the last release tag

```sh
git tag --sort=-version:refname | head -5
# e.g. v0.3.0
```

### 3. List commits since last tag

```sh
git log v0.3.0..HEAD --oneline
```

Use this to determine the next version and write the tag message.

### 4. Determine next version

- Any new command → bump MINOR (e.g. v0.3.0 → v0.4.0)
- Only fixes/chores → bump PATCH (e.g. v0.3.0 → v0.3.1)
- Breaking change → bump MAJOR

### 5. Create annotated tag

```sh
git tag -a vX.Y.Z -m "vX.Y.Z

- <bullet from commit messages — feat/fix/chore grouped>
- ..."
```

### 6. Push tag — triggers GoReleaser release workflow

```sh
git push origin vX.Y.Z
```

### 7. Verify release

```sh
gh release view vX.Y.Z --repo Rhevin/apple-compose
```

Watch the release workflow: `https://github.com/Rhevin/apple-compose/actions`

## Rules

- Only tag from `master` — never from a feature branch
- Only tag after all PRs for this release are merged
- Always use annotated tags (`-a`) not lightweight tags
- Tag message must list what changed — used by GoReleaser changelog
- Run `make test` before tagging to confirm tests pass

## Example

```sh
git checkout master && git pull
git log v0.2.0..HEAD --oneline
# 3 fix commits → bump PATCH → v0.3.0
make test
git tag -a v0.3.0 -m "v0.3.0

- Fix exec and run allow unknown flags to pass through
- Fix Apple container JSON schema for ps, ls, wait-healthy
- Fix top output format aligned with docker compose top"
git push origin v0.3.0
```
