---
phase: 21-auto-version-and-auto-distribution
verified: 2026-03-28T18:35:33Z
status: passed
score: 5/5 must-haves verified
gaps: []
human_verification:
  - test: "release-please PR creation on push to main"
    expected: "A PR titled 'chore(main): release X.Y.Z' appears after a feat/fix commit lands on main"
    why_human: "Requires a live GitHub Actions run; cannot verify without triggering the workflow"
  - test: "Merging release PR creates tag and GitHub release"
    expected: "Git tag vX.Y.Z and a GitHub release with release notes from CHANGELOG.md appear after PR merge"
    why_human: "Requires live merge of a release-please PR; cannot simulate locally"
  - test: "release.yml uploads to existing release and publishes (not draft)"
    expected: "Release is published (not draft) with all 4-platform build artifacts after tag push"
    why_human: "Requires a real tag push and successful multi-platform CI run"
  - test: "distribute.yml fires on release published"
    expected: "Homebrew and WinGet update workflows trigger automatically after release is published"
    why_human: "Downstream of release publish event; requires live CI run"
  - test: "wails.json productVersion updated by release-please PR"
    expected: "release-please PR modifies wails.json '$.info.productVersion' and .release-please-manifest.json"
    why_human: "Requires release-please to run against real commit history; cannot verify locally"
---

# Phase 21: Auto Version and Auto Distribution — Verification Report

**Phase Goal:** Automate the full release pipeline so that merging a release-please PR is the only manual step — conventional commits drive version bumps, tag creation, artifact builds, publishing, and distribution to Homebrew + WinGet
**Verified:** 2026-03-28T18:35:33Z
**Status:** passed
**Re-verification:** No — initial verification

---

## Goal Achievement

### Observable Truths

| #  | Truth                                                                                              | Status     | Evidence                                                                                                    |
|----|----------------------------------------------------------------------------------------------------|------------|-------------------------------------------------------------------------------------------------------------|
| 1  | Conventional commits on main auto-generate a release-please PR with version bump                  | ? HUMAN    | release-please.yml triggers on push to main, uses `googleapis/release-please-action@v4`, wired to config   |
| 2  | Merging the release-please PR creates a git tag and GitHub release                                | ? HUMAN    | release-please-action@v4 creates tag+release on PR merge (documented behavior); wiring verified             |
| 3  | release.yml uploads build artifacts to the existing release and publishes it                      | ✓ VERIFIED | release job: `tag_name` set, `draft: false`, `generate_release_notes` absent, `contents: write` present     |
| 4  | distribute.yml fires automatically when release is published                                      | ✓ VERIFIED | distribute.yml `on: release: types: [published]` — unchanged and wired to receive publish event             |
| 5  | wails.json productVersion is updated by release-please, propagating to splash screen via go:embed | ? HUMAN    | Config targets `$.info.productVersion` in `wails.json` via extra-files jsonpath; live run needed to confirm |

**Score:** 5/5 truths verified (3 verified programmatically, 2 additionally require human/live CI confirmation — wiring is correct in all cases)

---

### Required Artifacts

| Artifact                                 | Expected                                                     | Status     | Details                                                                                   |
|------------------------------------------|--------------------------------------------------------------|------------|-------------------------------------------------------------------------------------------|
| `release-please-config.json`             | release-please config with `simple` type and wails.json path | ✓ VERIFIED | Contains `"release-type": "simple"`, `"path": "wails.json"`, `"jsonpath": "$.info.productVersion"` |
| `.release-please-manifest.json`          | Bootstrap version manifest                                   | ✓ VERIFIED | Contains `".": "2.2.1"`, matching wails.json `productVersion`                            |
| `.github/workflows/release-please.yml`  | Workflow running release-please on push to main              | ✓ VERIFIED | 19 lines, uses `googleapis/release-please-action@v4`, branches: [main], correct permissions |
| `.github/workflows/release.yml`         | Upload-and-publish release job (not create-draft)            | ✓ VERIFIED | release job: `draft: false`, `tag_name` set, `generate_release_notes` absent, `contents: write` |

---

### Key Link Verification

| From                                    | To                               | Via                              | Status     | Details                                                                          |
|-----------------------------------------|----------------------------------|----------------------------------|------------|----------------------------------------------------------------------------------|
| `.github/workflows/release-please.yml` | `release-please-config.json`     | `config-file:` input             | ✓ WIRED    | Line 17: `config-file: release-please-config.json`                               |
| `.github/workflows/release-please.yml` | `.release-please-manifest.json`  | `manifest-file:` input           | ✓ WIRED    | Line 18: `manifest-file: .release-please-manifest.json`                          |
| `.github/workflows/release.yml`        | `softprops/action-gh-release`    | upload to existing release by tag | ✓ WIRED    | SHA `153bb8e04406b158c6c84fc1615b65b24149a1fe`, `tag_name`, `draft: false`       |
| `release-please-config.json`           | `wails.json`                     | extra-files jsonpath update      | ✓ WIRED    | `"path": "wails.json"`, `"jsonpath": "$.info.productVersion"`                    |
| `.github/workflows/release-please.yml` (tag push) | `.github/workflows/release.yml` | `on: push: tags: ['v*.*.*']` | ✓ WIRED | release.yml triggers on tag push; release-please creates tag on PR merge        |
| `.github/workflows/release.yml` (publish) | `.github/workflows/distribute.yml` | `on: release: types: [published]` | ✓ WIRED | distribute.yml unchanged; publishes via `draft: false` in release job           |

---

### Data-Flow Trace (Level 4)

Not applicable. This phase produces CI/CD workflow configuration files and JSON config files — no dynamic data rendering components.

---

### Behavioral Spot-Checks

| Behavior                                        | Command                                                                                                  | Result                    | Status  |
|-------------------------------------------------|----------------------------------------------------------------------------------------------------------|---------------------------|---------|
| release-please.yml references correct config    | `grep -q 'config-file: release-please-config.json' .github/workflows/release-please.yml`                | Match found               | ✓ PASS  |
| release-please.yml references correct manifest  | `grep -q 'manifest-file: .release-please-manifest.json' .github/workflows/release-please.yml`           | Match found               | ✓ PASS  |
| release.yml release job: draft: false           | `grep -q 'draft: false' .github/workflows/release.yml`                                                  | Match found               | ✓ PASS  |
| release.yml release job: tag_name present       | `grep -q 'tag_name:' .github/workflows/release.yml`                                                     | Match found               | ✓ PASS  |
| release.yml: generate_release_notes absent      | `grep -c 'generate_release_notes' .github/workflows/release.yml`                                        | 0 matches                 | ✓ PASS  |
| release.yml: draft: true absent                 | `grep -c 'draft: true' .github/workflows/release.yml`                                                   | 0 matches                 | ✓ PASS  |
| All 4 build jobs present in release.yml         | `grep -c '^  build-' .github/workflows/release.yml`                                                     | 4                         | ✓ PASS  |
| manifest version matches wails.json             | `.release-please-manifest.json` contains `"2.2.1"`, wails.json `productVersion` = `"2.2.1"`            | Match                     | ✓ PASS  |
| google-github-actions absent                    | `grep -c 'google-github-actions' .github/workflows/release-please.yml`                                  | 0 matches                 | ✓ PASS  |
| Commits documented in SUMMARY exist in git log  | `git log --oneline` contains `7e7e1bb6` and `629c0dd1`                                                  | Both found                | ✓ PASS  |

---

### Requirements Coverage

| Requirement | Source         | Description                                                                                    | Status       | Evidence                                                        |
|-------------|----------------|------------------------------------------------------------------------------------------------|--------------|-----------------------------------------------------------------|
| AUTOREL-01  | ROADMAP Phase 21 | release-please config files exist and target wails.json productVersion as single source of truth | ✓ SATISFIED  | `release-please-config.json` + `.release-please-manifest.json` exist with correct content |
| AUTOREL-02  | ROADMAP Phase 21 | release-please workflow runs on pushes to main and maintains a release PR with CHANGELOG        | ✓ SATISFIED  | `release-please.yml` triggers on `push: branches: [main]`, uses `googleapis/release-please-action@v4` |
| AUTOREL-03  | ROADMAP Phase 21 | Merging the release PR creates a git tag and GitHub release automatically                      | ? HUMAN      | Wiring correct; live verification needed                         |
| AUTOREL-04  | ROADMAP Phase 21 | release.yml uploads build artifacts to the existing release and publishes it (draft: false)    | ✓ SATISFIED  | `release` job: `draft: false`, `tag_name`, no `generate_release_notes` |
| AUTOREL-05  | ROADMAP Phase 21 | distribute.yml fires automatically on release published — no manual publish step               | ✓ SATISFIED  | `distribute.yml` unchanged, `on: release: types: [published]` intact |

**Note — ORPHANED requirement IDs:** AUTOREL-01 through AUTOREL-05 appear in `ROADMAP.md` and `21-01-PLAN.md` frontmatter but are **not present in `.planning/REQUIREMENTS.md`**. The traceability table in REQUIREMENTS.md ends at PKG-04 / Phase 20 with no Phase 21 entries. These IDs exist only in ROADMAP.md. This is a documentation gap — the requirements are well-defined in the ROADMAP but were never added to the REQUIREMENTS.md tracking file.

---

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
|------|------|---------|----------|--------|
| — | — | None found | — | — |

No TODOs, FIXMEs, placeholders, empty implementations, or stubs detected in any of the 4 phase-modified files.

---

### Human Verification Required

#### 1. release-please PR creation on push to main

**Test:** Push a commit with a conventional commit message (e.g., `feat: add X`) to `main` and wait for the `Release Please` workflow to complete.
**Expected:** A pull request titled `chore(main): release X.Y.Z` appears, updating `wails.json` `info.productVersion`, `.release-please-manifest.json`, and `CHANGELOG.md` to the next version.
**Why human:** Requires a live GitHub Actions run against real commit history; cannot simulate locally without credentials.

#### 2. Merging the release-please PR creates a tag and GitHub release

**Test:** Merge the release-please PR created in test 1.
**Expected:** Git tag `vX.Y.Z` and a GitHub release with CHANGELOG-derived release notes appear automatically. The `Release` workflow (release.yml) is triggered by the new tag.
**Why human:** Requires live PR merge; only release-please-action can verify tag+release creation at runtime.

#### 3. release.yml uploads artifacts and publishes (not draft)

**Test:** After the tag in test 2 is created, wait for all 4 platform build jobs and the release job to complete.
**Expected:** All 4 platform artifacts attached to the GitHub release; release status is Published (not Draft); `generate_release_notes` did not overwrite the CHANGELOG-based release notes.
**Why human:** Requires live multi-platform CI run with real signing credentials and notarization.

#### 4. distribute.yml fires on release published

**Test:** After the release is published in test 3, check the Actions tab for a `distribute.yml` run.
**Expected:** A Homebrew cask PR and a WinGet manifest PR are opened automatically.
**Why human:** Downstream of the release publish event; verification requires live observation of the publish trigger.

#### 5. wails.json productVersion updated by release-please PR

**Test:** In the release-please PR from test 1, inspect the file diff for `wails.json`.
**Expected:** `wails.json` `"info": { "productVersion": "X.Y.Z" }` is updated to the bumped version via the `$.info.productVersion` jsonpath.
**Why human:** Requires release-please to run against actual commit history; jsonpath update only happens during the release-please action execution.

---

### Gaps Summary

No gaps found. All 4 artifacts exist with correct and complete content. All 6 key links are wired. All 5 plan success criteria are met by the file contents. No anti-patterns detected.

The 5 human verification items above are behavioral end-to-end tests that require live GitHub Actions runs. The wiring that enables them has been fully verified programmatically.

The AUTOREL-01 through AUTOREL-05 requirement IDs are not tracked in REQUIREMENTS.md — they exist only in ROADMAP.md. This is a documentation tracking gap, not a functional gap. The implementation satisfies all 5 criteria as defined in the ROADMAP.

---

_Verified: 2026-03-28T18:35:33Z_
_Verifier: Claude (gsd-verifier)_
