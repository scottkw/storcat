---
phase: 18-windows-authenticode-signing
plan: "01"
subsystem: ci-cd
tags: [windows, authenticode, code-signing, github-actions, esigner]
dependency_graph:
  requires: [phase-16-secrets]
  provides: [windows-authenticode-signing]
  affects: [.github/workflows/release.yml]
tech_stack:
  added: [SSLcom/esigner-codesign@cf5f6c1d38ad10f47e3ed9aca873f429b1a8d85b]
  patterns: [environment-secrets, pre-upload-signing, signtool-gate]
key_files:
  modified: [.github/workflows/release.yml]
decisions:
  - "Use SSLcom/esigner-codesign SHA-pinned action (cf5f6c1d) for cloud HSM signing — no local certificate storage required"
  - "malware_block: false to avoid paid SSL.com add-on failure"
  - "Sign before rename so eSigner targets Wails canonical filenames (StorCat.exe, StorCat-amd64-installer.exe)"
  - "shell: cmd for signtool verify step — bash does not have signtool on PATH on Windows runners"
metrics:
  duration: "36s"
  completed: "2026-03-28T05:24:28Z"
  tasks_completed: 1
  tasks_total: 2
  files_modified: 1
---

# Phase 18 Plan 01: Windows Authenticode Signing Summary

**One-liner:** Windows Authenticode signing via SSL.com eSigner cloud HSM, with signtool CI gate, in the build-windows job of release.yml.

## Tasks Completed

| Task | Name | Commit | Files |
|------|------|--------|-------|
| 2 | Add Authenticode signing and verification to build-windows job | 78da4d73 | .github/workflows/release.yml |

## Tasks Deferred (Checkpoint)

| Task | Name | Type | Status |
|------|------|------|--------|
| 1 | Store SSL.com eSigner secrets in GitHub release environment | checkpoint:human-action | Awaiting user action |

Task 1 is a human-action checkpoint — user must purchase SSL.com OV code signing certificate, enroll in eSigner, and store 4 secrets (`ES_USERNAME`, `ES_PASSWORD`, `CREDENTIAL_ID`, `ES_TOTP_SECRET`) in the GitHub `release` environment. The workflow code changes (Task 2) are complete and committed; CI will fail on the signing steps until secrets are stored.

## What Was Built

Modified `build-windows` job in `.github/workflows/release.yml` with three changes:

1. **`environment: release`** added to job definition — exposes ES_USERNAME, ES_PASSWORD, CREDENTIAL_ID, ES_TOTP_SECRET to the job
2. **Sign portable exe** step — `SSLcom/esigner-codesign@cf5f6c1d38ad10f47e3ed9aca873f429b1a8d85b` signs `build\bin\StorCat.exe` before rename
3. **Sign NSIS installer** step — same action signs `build\bin\StorCat-amd64-installer.exe` before rename
4. **Verify Authenticode signatures** step — `signtool verify /pa /v` gates both files; CI fails if either signature is invalid

Final step order in build-windows: checkout → setup-go → setup-node → install-wails → get-version → **build** → **sign portable** → **sign NSIS** → **verify** → rename portable → rename installer → upload.

## Deviations from Plan

None — plan executed exactly as written. Task 1 is a checkpoint (human-action), not a deviation.

## Known Stubs

None. The workflow code is complete. The signing steps will be skipped with empty secrets until Task 1 is completed by the user, but that is the expected behavior for a blocking human-action checkpoint.

## Self-Check: PASSED

- FOUND: `.github/workflows/release.yml` (modified)
- FOUND: `18-01-SUMMARY.md` (created)
- FOUND commit `78da4d73` — feat(18-01): add Windows Authenticode signing to build-windows job
