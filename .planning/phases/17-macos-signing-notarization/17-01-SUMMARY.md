---
phase: 17-macos-signing-notarization
plan: "01"
subsystem: infrastructure
tags: [github-actions, code-signing, notarization, macos, gatekeeper, entitlements]
dependency_graph:
  requires: [phase-16-secrets-certificate-procurement]
  provides: [macos-signing-pipeline, wails-entitlements-plist]
  affects: [phase-19-homebrew-path]
tech_stack:
  added:
    - apple-actions/import-codesign-certs@b610f78 (v6.0.0) -- keychain management for CI codesign
    - xcrun notarytool -- Apple notarization submission (replaced altool Nov 2023)
    - xcrun stapler -- notarization ticket stapling to DMG
  patterns:
    - isolated-keychain-ci-signing
    - env-block-secret-mapping
    - pitfall3-notarytool-rejection-detection
key_files:
  created:
    - build/darwin/entitlements.plist
  modified:
    - .github/workflows/release.yml
decisions:
  - "Use env: block for secret mapping in run steps (not direct interpolation) — prevents secret values appearing in expanded shell commands in CI logs"
  - "No --deep flag on codesign — StorCat has lean 3-file .app bundle (no nested frameworks), --deep is not needed and not recommended practice"
  - "APPLE_ID secret (10th secret) required for notarytool --apple-id argument — user must add to release environment before first CI run"
  - "spctl --assess targets .app not DMG — exec assessment operates on the bundle; DMG uses --type open"
  - "Tarball moved AFTER codesign — both tarball and DMG now package a signed .app bundle"
metrics:
  duration_seconds: 300
  completed_date: "2026-03-28"
  tasks_completed: 2
  tasks_total: 3
  files_modified: 2
requirements_satisfied:
  - SIGN-01
  - SIGN-02
  - SIGN-03
  - SIGN-04
  - SIGN-05
  - SIGN-06
---

# Phase 17 Plan 01: macOS Signing, Notarization, and Gatekeeper Verification Summary

**One-liner:** Complete macOS Developer ID signing pipeline — entitlements plist, codesign with hardened runtime, xcrun notarytool DMG notarization, stapling, and spctl Gatekeeper verification in CI.

## What Was Built

Two artifacts implement the full macOS signing pipeline for StorCat CI:

**1. `build/darwin/entitlements.plist` (SIGN-05)**

Wails-native entitlements plist with 4 required entitlements for WebKit/WKWebView and filesystem access. Replaces the non-existent Electron-era plist. Passes `plutil -lint`.

**2. `.github/workflows/release.yml` — `build-macos` job (SIGN-01 through SIGN-06)**

7 new steps inserted into the job with mandatory ordering:

| Step | Purpose | Key Implementation |
|------|---------|-------------------|
| Import Apple certificate | Isolated keychain setup | `apple-actions/import-codesign-certs@b610f78` handles ACL (prevents codesign hang) |
| Sign .app bundle | Developer ID + hardened runtime | `codesign --options runtime --entitlements` — no `--deep`, env block for cert name |
| Verify code signature | Fail-fast before DMG | `codesign --verify --verbose` — surfaces signing errors before expensive DMG/notarization |
| (Tarball moved here) | Tarball now gets signed .app | Existing step relocated after codesign |
| Notarize DMG | Apple notarization | `xcrun notarytool submit --wait` + rejection detection (`grep "status: Accepted"`) |
| Staple notarization ticket | Offline Gatekeeper support | `xcrun stapler staple` on DMG |
| Verify Gatekeeper acceptance | CI gate | `spctl --assess --type exec` after stapling |
| Upload macOS artifact | (existing, unchanged) | Both .tar.gz and .dmg now contain signed .app |
| Clean up keychain | Security hygiene | `if: always()` — runs even on signing failure |

**Critical ordering maintained:**
```
wails build → codesign .app → verify sig → tarball → hdiutil DMG → notarytool → stapler → spctl → upload → cleanup
```

## Deviations from Plan

### Auto-fixed Issues

None.

### Planned Deviations

**1. APPLE_ID env variable named `APPLE_ID_EMAIL`**

The plan specified mapping `secrets.APPLE_ID` to env var `APPLE_ID_EMAIL` (matching the research code example). The final implementation uses exactly this naming to disambiguate from the `APPLE_ID` secret name and make the variable's purpose clear in the run script. This is consistent with the plan's intent.

## Checkpoint Reached (Task 3)

Task 3 is a `checkpoint:human-verify` gate requiring the user to:
1. Add `APPLE_ID` secret (10th secret) to the GitHub `release` environment
2. Verify all 10 secrets are present via `gh secret list --env release`
3. Optionally trigger a test tag push to validate the full pipeline end-to-end

## Known Stubs

None. Both files are complete and wired correctly. The signing pipeline cannot be tested locally (requires CI + Apple credentials), but the YAML is structurally complete and all secret references are correct.

## Verification Results

```
plutil -lint build/darwin/entitlements.plist → OK
grep -c 'com.apple.security' build/darwin/entitlements.plist → 4
All 19 acceptance criteria: PASS
YAML validation (python yaml.safe_load): PASS
Step order verification: 17 steps in correct sequence
```

## Self-Check: PASSED

- `build/darwin/entitlements.plist` — FOUND
- `.github/workflows/release.yml` — FOUND (modified)
- Task 1 commit `1dcae828` — FOUND
- Task 2 commit `4f69dc80` — FOUND
