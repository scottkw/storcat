---
phase: 21-auto-version-and-auto-distribution
plan: "01"
subsystem: ci-cd
tags: [release-please, automation, versioning, github-actions]
dependency_graph:
  requires: []
  provides: [release-please-automation, auto-version-bump, auto-tag-creation]
  affects: [release.yml, distribute.yml]
tech_stack:
  added: [release-please]
  patterns: [conventional-commits, release-please-simple-type, softprops-upload-to-existing]
key_files:
  created:
    - release-please-config.json
    - .release-please-manifest.json
    - .github/workflows/release-please.yml
  modified:
    - .github/workflows/release.yml
decisions:
  - "Use release-please simple release-type (not go) to enable wails.json extra-file jsonpath update"
  - "Use softprops tag_name to upload to existing release-please-created release rather than creating a new draft"
  - "Remove generate_release_notes from release job to prevent overwriting release-please changelog"
metrics:
  duration_seconds: 65
  tasks_completed: 2
  tasks_total: 2
  files_created: 3
  files_modified: 1
  completed_date: "2026-03-28"
---

# Phase 21 Plan 01: Release-Please Automation Summary

Release-please wired to auto-bump wails.json productVersion on conventional commits, create tags and GitHub releases, and trigger the existing build + distribute pipeline.

## What Was Built

Three new files + one modified workflow establish the full automated release chain:

1. **`release-please-config.json`** — Configures release-please with `simple` release-type and an `extra-files` entry targeting `$.info.productVersion` in `wails.json`. This is the single source of truth for version.

2. **`.release-please-manifest.json`** — Bootstrap manifest at `2.2.1`, matching the current `wails.json` productVersion. release-please owns this file after first run.

3. **`.github/workflows/release-please.yml`** — Runs `googleapis/release-please-action@v4` on every push to `main`. Creates/updates a release PR with version bump. When merged, creates the git tag and GitHub release automatically.

4. **`.github/workflows/release.yml` (release job modified)** — Release job now uploads artifacts to the release-please-created release (via `tag_name`) and publishes it (`draft: false`). Publishing triggers `distribute.yml` via its existing `release: types: [published]` trigger. `generate_release_notes` removed to preserve release-please changelog.

## Full Automated Chain

```
push to main (conventional commit)
  -> release-please PR (version bump in wails.json + CHANGELOG.md)
  -> merge PR
  -> tag + GitHub release created by release-please
  -> release.yml triggered by tag push
  -> 4 platform builds in parallel
  -> release job: upload artifacts, publish release (draft: false)
  -> distribute.yml triggered by release published
  -> Homebrew + WinGet updated automatically
```

## Commits

| Task | Commit | Files |
|------|--------|-------|
| Task 1: Create release-please config and workflow | 7e7e1bb6 | release-please-config.json, .release-please-manifest.json, .github/workflows/release-please.yml |
| Task 2: Refactor release job | 629c0dd1 | .github/workflows/release.yml |

## Deviations from Plan

None - plan executed exactly as written.

## Known Stubs

None.

## Self-Check: PASSED

Files verified:
- release-please-config.json: FOUND
- .release-please-manifest.json: FOUND
- .github/workflows/release-please.yml: FOUND
- .github/workflows/release.yml: FOUND (modified)

Commits verified:
- 7e7e1bb6: FOUND
- 629c0dd1: FOUND
