---
phase: 12-repo-consolidation
plan: 01
subsystem: packaging
tags: [winget, packaging, manifest, migration]
dependency_graph:
  requires: []
  provides: [packaging/winget/manifests/s/scottkw/StorCat/]
  affects: [phase-15-ci-automation]
tech_stack:
  added: []
  patterns: [WinGet manifest schema 1.6.0, packaging directory convention]
key_files:
  created:
    - packaging/winget/manifests/s/scottkw/StorCat/1.1.1/scottkw.StorCat.yaml
    - packaging/winget/manifests/s/scottkw/StorCat/1.1.1/scottkw.StorCat.installer.yaml
    - packaging/winget/manifests/s/scottkw/StorCat/1.1.1/scottkw.StorCat.locale.en-US.yaml
    - packaging/winget/manifests/s/scottkw/StorCat/1.1.2/scottkw.StorCat.yaml
    - packaging/winget/manifests/s/scottkw/StorCat/1.1.2/scottkw.StorCat.installer.yaml
    - packaging/winget/manifests/s/scottkw/StorCat/1.1.2/scottkw.StorCat.locale.en-US.yaml
    - packaging/winget/manifests/s/scottkw/StorCat/1.2.0/scottkw.StorCat.yaml
    - packaging/winget/manifests/s/scottkw/StorCat/1.2.0/scottkw.StorCat.installer.yaml
    - packaging/winget/manifests/s/scottkw/StorCat/1.2.0/scottkw.StorCat.locale.en-US.yaml
    - packaging/winget/manifests/s/scottkw/StorCat/1.2.1/scottkw.StorCat.yaml
    - packaging/winget/manifests/s/scottkw/StorCat/1.2.1/scottkw.StorCat.installer.yaml
    - packaging/winget/manifests/s/scottkw/StorCat/1.2.1/scottkw.StorCat.locale.en-US.yaml
    - packaging/winget/manifests/s/scottkw/StorCat/1.2.2/scottkw.StorCat.yaml
    - packaging/winget/manifests/s/scottkw/StorCat/1.2.2/scottkw.StorCat.installer.yaml
    - packaging/winget/manifests/s/scottkw/StorCat/1.2.2/scottkw.StorCat.locale.en-US.yaml
    - packaging/winget/manifests/s/scottkw/StorCat/1.2.3/scottkw.StorCat.yaml
    - packaging/winget/manifests/s/scottkw/StorCat/1.2.3/scottkw.StorCat.installer.yaml
    - packaging/winget/manifests/s/scottkw/StorCat/1.2.3/scottkw.StorCat.locale.en-US.yaml
    - packaging/winget/manifests/s/scottkw/StorCat/2.1.0/scottkw.StorCat.yaml
    - packaging/winget/manifests/s/scottkw/StorCat/2.1.0/scottkw.StorCat.installer.yaml
    - packaging/winget/manifests/s/scottkw/StorCat/2.1.0/scottkw.StorCat.locale.en-US.yaml
    - packaging/winget/update-winget.sh
  modified: []
decisions:
  - "Historical WinGet manifests (v1.1.1–v1.2.3) copied verbatim including known defect in v1.2.3 installer manifest (malformed SHA256) — these are reference copies, not submitted to winget-pkgs"
  - "v2.1.0 manifests created as stubs with 64-character zero SHA256 placeholder — real values added in Phase 15 when release assets exist"
metrics:
  duration: 5m
  completed: 2026-03-27
  tasks_completed: 2
  files_created: 22
---

# Phase 12 Plan 01: WinGet Manifest Migration Summary

**One-liner:** Migrated all WinGet manifests (v1.1.1–v1.2.3) from satellite repo into `packaging/winget/` and created v2.1.0 stub manifests with Go/Wails description replacing Electron references.

## What Was Built

All WinGet packaging metadata is now consolidated in the main storcat repository under `packaging/winget/`. This includes:

1. **Historical manifests (v1.1.1–v1.2.3):** 6 version directories, each with 3 YAML files (version, installer, locale), totalling 18 files. Copied verbatim from `scottkw/winget-storcat` satellite repo.

2. **v2.1.0 stub manifests:** 3 YAML files for the current app version. Description updated to reflect Go/Wails backend. InstallerSha256 set to 64-character zero placeholder pending Phase 15 release.

3. **update-winget.sh script:** Copied from satellite repo with execute permission intact.

## Tasks Completed

| Task | Name | Commit | Files |
|------|------|--------|-------|
| 1 | Copy historical manifests from winget-storcat | c74b02ee | 19 files (18 manifests + update-winget.sh) |
| 2 | Create v2.1.0 stub manifests with Go/Wails description | 7af98d6a | 3 files |

## Verification Results

- `ls packaging/winget/manifests/s/scottkw/StorCat/ | sort` → 7 directories (1.1.1, 1.1.2, 1.2.0, 1.2.1, 1.2.2, 1.2.3, 2.1.0)
- `find packaging/winget/manifests -name "*.yaml" | wc -l` → 21 files
- `test -x packaging/winget/update-winget.sh` → exit 0
- v2.1.0 locale contains "Go and Wails", zero instances of "Electron"

## Deviations from Plan

None - plan executed exactly as written.

## Known Stubs

**v2.1.0 installer manifest — placeholder SHA256:**
- File: `packaging/winget/manifests/s/scottkw/StorCat/2.1.0/scottkw.StorCat.installer.yaml`
- Line: `InstallerSha256: 0000000000000000000000000000000000000000000000000000000000000000`
- Reason: v2.1.0 GitHub release assets do not exist yet. Phase 15 will replace this with the real SHA256 when the release is published. This is intentional per REPO-01 design.

## Self-Check: PASSED

Files exist:
- packaging/winget/manifests/s/scottkw/StorCat/ (7 version directories)
- packaging/winget/update-winget.sh (executable)
- packaging/winget/manifests/s/scottkw/StorCat/2.1.0/ (3 files)

Commits exist:
- c74b02ee: chore(12-01): copy historical WinGet manifests from winget-storcat satellite repo
- 7af98d6a: chore(12-01): create v2.1.0 WinGet stub manifests with Go/Wails description
