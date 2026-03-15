---
name: release-github
description: Use this skill when publishing a small GitHub release for AiP2P News. It covers version bumps, tests, fresh-clone verification, safe push flow, tag creation, and GitHub release creation without breaking the split-repo layout.
---

# GitHub Release

Use this skill when you need to publish a small version for:

- `AiP2P News`

This workspace layout is unusual:

- the local workspace has one root Git repository
- GitHub has a separate downstream repository:
  - `AiP2P/AiP2P-News`

Do not push the `aip2p-news/` subdirectory from the root repo directly.

## Required Outcome

For every release:

1. bump the version strings in source and docs
2. update the draft release note file
3. run tests locally
4. verify with a fresh clone
5. fresh-clone `https://github.com/AiP2P/AiP2P-News.git` into a temp directory
6. copy only the local `aip2p-news/` subtree into that temp clone
7. commit with GitHub `noreply` email
8. push `main`
9. create and push the new tag
10. create the GitHub release from the release note file

## Version Rules

Use small increments only.

Typical pattern:

- `AiP2P News`: `v0.2.X-demo`

Update these places at minimum:

- `cmd/latest/main.go`
- `README.md`
- `docs/install.md`
- `docs/release.md`
- `skills/bootstrap-aip2p-news/SKILL.md`
- `docs/github-release-v*.md`

## Test Before Push

Run:

```bash
go test ./...
go -C ./aip2p test ./...
```

If the change affects runtime behavior, also do a fresh install style check from a fresh clone.

## Safe Push Flow

Do not publish from the root workspace checkout.

Instead:

1. create a temp directory
2. clone `https://github.com/AiP2P/AiP2P-News.git`
3. copy local `aip2p-news/` into the fresh clone
4. exclude `.git`
5. do not commit built binaries unless the repo already intentionally tracks them

Recommended copy pattern:

```bash
rsync -a --delete --exclude '.git' /local/path/aip2p-news/ /tmp/push-aip2p-news/
```

## Commit Identity

GitHub may reject pushes that expose a private email address.

Before commit:

```bash
git config user.name AiP2P
git config user.email 112829784+AiP2P@users.noreply.github.com
```

## Tag And Release

Pattern:

```bash
git push origin main
git tag -f <tag>
git push -f origin <tag>
gh release create <tag> --repo AiP2P/AiP2P-News --title "<title>" --notes-file <release-notes-file>
```

## Release Notes

Keep one release note file per version:

- `docs/github-release-vX.Y.Z-demo.md`

Each note should include:

- release title
- one-sentence summary
- highlights
- install or upgrade reminder
