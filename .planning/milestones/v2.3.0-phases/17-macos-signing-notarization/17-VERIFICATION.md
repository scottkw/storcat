---
phase: 17-macos-signing-notarization
verified: 2026-03-28T04:55:55Z
status: passed
score: 8/8 must-haves verified
gaps: []
human_verification:
  - test: "Full Gatekeeper acceptance on macOS 15+"
    expected: "Downloaded DMG opens without quarantine warning or Gatekeeper prompt on a stock macOS 15 machine"
    why_human: "spctl --assess passes in CI on macos-14; user-facing Gatekeeper behavior on macOS 15+ requires installing the actual release DMG on a physical or VM macOS 15 target"
  - test: "Offline staple verification"
    expected: "xcrun stapler validate on the released DMG succeeds with no network access"
    why_human: "Stapling correctness can only be tested offline on the actual released artifact; CI environment has network access which bypasses the offline path"
---

# Phase 17: macOS Signing & Notarization Verification Report

**Phase Goal:** Sign, notarize, and staple every macOS DMG produced by CI release tags so Gatekeeper accepts them without prompting on macOS 15+.
**Verified:** 2026-03-28T04:55:55Z
**Status:** PASSED
**Re-verification:** No — initial verification

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | `build/darwin/entitlements.plist` exists with 4 Wails-required entitlements | VERIFIED | File exists, passes `plutil -lint`, `grep -c 'com.apple.security'` returns 4 |
| 2 | `release.yml` `build-macos` job declares `environment: release` | VERIFIED | L13: `environment: release` immediately after `runs-on: macos-14` |
| 3 | `release.yml` imports Apple certificate via `apple-actions/import-codesign-certs@v6` | VERIFIED | L41: `uses: apple-actions/import-codesign-certs@b610f78488812c1e56b20e6df63ec42d833f2d14` |
| 4 | `release.yml` signs .app bundle with `--options runtime` and entitlements plist BEFORE tarball and DMG | VERIFIED | Sign step L46 < tarball L62 < DMG L65; `--options runtime` and `--entitlements build/darwin/entitlements.plist` present; no `--deep` flag |
| 5 | `release.yml` submits DMG to Apple notarization via `xcrun notarytool submit --wait` | VERIFIED | L86-89: `xcrun notarytool submit ... --wait`; Pitfall 3 rejection detection at L92 |
| 6 | `release.yml` staples notarization ticket to DMG via `xcrun stapler staple` | VERIFIED | L101-104: `xcrun stapler staple` on DMG path |
| 7 | `release.yml` runs `spctl --assess` as a gate step after stapling | VERIFIED | L107: `spctl --assess --type exec build/bin/StorCat.app \|\| exit 1`; appears after staple (L101) |
| 8 | Keychain cleanup handled (by apple-actions post-step) | VERIFIED | Intentional deviation: explicit `security delete-keychain` step removed (commit `3307491e`) because `apple-actions/import-codesign-certs@v6` performs its own post-job keychain cleanup; duplicate caused job failure (exit code 50) |

**Score:** 8/8 truths verified

### Required Artifacts

| Artifact | Expected | Status | Details |
|----------|----------|--------|---------|
| `build/darwin/entitlements.plist` | Wails WebKit entitlements for hardened runtime | VERIFIED | Exists, valid XML (`plutil -lint` OK), 4 entitlements, no app-sandbox |
| `.github/workflows/release.yml` | macOS signing, notarization, stapling, verification CI steps | VERIFIED | Contains all 7 required new steps; valid YAML |

### Entitlements Detail

All 4 required entitlements confirmed present; forbidden keys absent:

| Key | Present | Required |
|-----|---------|----------|
| `com.apple.security.cs.allow-jit` | YES | YES — WebKit JIT |
| `com.apple.security.cs.allow-unsigned-executable-memory` | YES | YES — WebKit memory |
| `com.apple.security.files.user-selected.read-write` | YES | YES — filesystem access |
| `com.apple.security.network.client` | YES | YES — Gatekeeper validation |
| `com.apple.security.app-sandbox` | NO | MUST NOT be present — breaks filesystem access |
| `com.apple.security.cs.disable-library-validation` | NO | MUST NOT be present — no third-party dylibs |

### Key Link Verification

| From | To | Via | Status | Details |
|------|----|-----|--------|---------|
| `.github/workflows/release.yml` | `build/darwin/entitlements.plist` | `--entitlements build/darwin/entitlements.plist` in codesign step | WIRED | L53: `--entitlements build/darwin/entitlements.plist` |
| `.github/workflows/release.yml` | GitHub `release` environment secrets | `environment: release` on job | WIRED | L13: `environment: release` |
| Sign .app step | Tarball + DMG steps | codesign runs before both packaging steps | WIRED | L46 (sign) < L62 (tarball) < L65 (DMG) |
| notarytool step | stapler step | stapler requires successful notarization | WIRED | L75 (notarize) < L101 (staple) |

### Step Ordering Constraints

All mandatory ordering constraints from PLAN satisfied:

| Constraint | Before (line) | After (line) | Status |
|------------|--------------|-------------|--------|
| Sign .app before Create macOS tarball | L46 | L62 | PASS |
| Create macOS tarball before Create macOS DMG | L62 | L65 | PASS |
| Create macOS DMG before Notarize DMG | L65 | L75 | PASS |
| Notarize DMG before Staple notarization ticket | L75 | L101 | PASS |
| Staple before Verify Gatekeeper acceptance | L101 | L106 | PASS |

### Data-Flow Trace (Level 4)

Not applicable — this phase delivers CI pipeline configuration (YAML), not UI components or data-rendering code. No data-flow trace is relevant.

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
|----------|---------|--------|--------|
| entitlements.plist is valid XML plist | `plutil -lint build/darwin/entitlements.plist` | `build/darwin/entitlements.plist: OK` | PASS |
| entitlements.plist has exactly 4 security keys | `grep -c 'com.apple.security' build/darwin/entitlements.plist` | `4` | PASS |
| release.yml contains all required signing steps | grep for 6 required patterns | All 6 found | PASS |
| Commits documented in SUMMARY exist in git log | `git log --oneline \| grep -E "1dcae828\|4f69dc80\|3307491e"` | All 3 found | PASS |
| Step ordering maintained | Line number analysis | All 5 constraints satisfied | PASS |

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
|-------------|------------|-------------|--------|----------|
| SIGN-01 | 17-01-PLAN.md | `release.yml` `build-macos` signs .app with `codesign --sign --options runtime` and entitlements | SATISFIED | L46-57: `codesign --sign "$APPLE_CERT_NAME" --options runtime --entitlements build/darwin/entitlements.plist` |
| SIGN-02 | 17-01-PLAN.md | `release.yml` submits DMG to Apple notarization via `xcrun notarytool` | SATISFIED | L75-99: `xcrun notarytool store-credentials` + `xcrun notarytool submit --wait` with rejection detection |
| SIGN-03 | 17-01-PLAN.md | `release.yml` staples notarization ticket to DMG via `xcrun stapler` | SATISFIED | L101-104: `xcrun stapler staple "build/bin/StorCat-...-darwin-universal.dmg"` |
| SIGN-04 | 17-01-PLAN.md | CI uses isolated temporary keychain, cleaned up after signing | SATISFIED | `apple-actions/import-codesign-certs@v6` creates isolated keychain with ACL and performs post-job cleanup; explicit duplicate step removed per commit `3307491e` |
| SIGN-05 | 17-01-PLAN.md | Entitlements plist ported from Electron era and verified for Wails runtime | SATISFIED | `build/darwin/entitlements.plist` created with 4 Wails/WebKit-appropriate entitlements; passes `plutil -lint` |
| SIGN-06 | 17-01-PLAN.md | `spctl --assess` verification confirms signed .app accepted by Gatekeeper | SATISFIED | L107: `spctl --assess --type exec build/bin/StorCat.app \|\| exit 1` after stapling |

**All 6 requirement IDs (SIGN-01 through SIGN-06) satisfied.**

No orphaned requirements — all 6 IDs declared in PLAN frontmatter are fully mapped and verified.

### Anti-Patterns Found

| File | Pattern | Severity | Impact |
|------|---------|----------|--------|
| None detected | — | — | — |

Checked for: placeholder comments (TODO/FIXME), empty implementations, hardcoded empty values, stub patterns. None found in `build/darwin/entitlements.plist` or `.github/workflows/release.yml`.

### Deviation Analysis: Keychain Cleanup Step

**Truth 8** (keychain cleanup) is VERIFIED with a documented deviation from the PLAN spec.

- **Plan specified:** Explicit `Clean up keychain` step with `security delete-keychain signing_temp.keychain-db` and `if: always()`
- **Actual implementation:** Step removed (commit `3307491e`)
- **Reason:** `apple-actions/import-codesign-certs@v6` contains built-in post-job cleanup. The explicit step deleted the keychain first, then the action's post-step attempted the same deletion and exited with code 50, marking the entire `build-macos` job as failed.
- **Verdict:** The PLAN's intent (keychain cleaned up after signing, even on failure) is fully achieved by the action's own mechanism. This is a correct fix, not a gap. The CI run 23677752782 confirmed the pipeline passes end-to-end with this configuration.

### Human Verification Required

#### 1. Gatekeeper Acceptance on macOS 15+

**Test:** Download a release DMG from GitHub Releases (e.g., `v99.0.1-test` or the next production release). Double-click the DMG on a macOS 15 machine. Attempt to open StorCat.app.
**Expected:** No Gatekeeper warning dialog. App launches immediately.
**Why human:** CI runs `spctl --assess --type exec` on the .app inside `macos-14`. Actual end-user Gatekeeper behavior on macOS 15+ with a downloaded (quarantined) DMG can only be confirmed on real hardware or a macOS 15 VM.

#### 2. Offline Staple Verification

**Test:** Download the DMG, disconnect from the internet, run `xcrun stapler validate StorCat-*-darwin-universal.dmg`.
**Expected:** Returns "The staple and validate action worked!" without network access.
**Why human:** Confirms the staple ticket is embedded in the DMG for offline Gatekeeper support. Cannot be automated without network isolation.

### Gaps Summary

No gaps. All 8 truths verified. All 6 requirements satisfied. Both artifacts exist and are substantive. All key links wired. Step ordering constraints satisfied. CI pipeline confirmed green (run 23677752782, all jobs passed).

The phase goal is achieved: every macOS DMG produced by a CI release tag is signed with Developer ID Application certificate using hardened runtime and Wails-appropriate entitlements, notarized by Apple via `xcrun notarytool`, stapled via `xcrun stapler`, and verified by `spctl --assess` as a CI gate.

---

_Verified: 2026-03-28T04:55:55Z_
_Verifier: Claude (gsd-verifier)_
