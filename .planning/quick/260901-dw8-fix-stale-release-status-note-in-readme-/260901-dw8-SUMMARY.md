---
quick_id: 260901-dw8
status: complete
date: 2026-09-01
commit: ee639013
---

# Quick Task 260901-dw8 — Summary

## What changed

`README.md`: deleted the three-line `> **Note:**` block that claimed the latest
published release was v2.3.0 and directed users to build from source for the
workspace redesign. Net −4 lines (note + trailing blank).

## Why deletion, not a rewrite

The README already states the current version in three places that stay correct on
their own — the `# StorCat v3.0.0` title, the `## What's New in v3.0.0` section, and
the Installation/Releases sections which link to the unversioned `/releases` page. A
replacement line asserting "v3.0.0 is the latest release" would need editing again at
v3.1.0; the note only existed to cover the gap between `main` and the last tag, and
that gap is closed.

## Verification

- `gh release list` → `v3.0.0  Latest  2026-08-18` (premise confirmed before editing)
- `grep -n 'published' README.md` → no matches
- `grep -n 'v2\.3\.0' README.md` → only L565-566, the historical "Migration from v1.x"
  changelog entries for code signing and release automation — correct to keep
- `git diff --stat` → `README.md | 4 ----`

## Left alone deliberately

Version *history* is not a release-status claim: the "Why StorCat v2.0.0" migration
rationale, the Electron v1.2.3 vs Wails v2.0.0 performance table, and the (v2.1.0) /
(v2.3.0) / (v3.0.0) changelog tags all remain.

## Commit

`ee639013` docs(readme): drop stale pre-release note — v3.0.0 is the published latest
