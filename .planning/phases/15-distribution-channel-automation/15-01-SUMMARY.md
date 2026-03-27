---
phase: 15-distribution-channel-automation
plan: 01
subsystem: infra
tags: [github-actions, homebrew, winget, distribution, release-automation, cicd]

# Dependency graph
requires:
  - phase: 14-platform-packaging
    provides: release.yml with NSIS installer and darwin-universal DMG artifacts
provides:
  - distribute.yml workflow triggered on release:published for automated Homebrew and WinGet distribution
  - Corrected Homebrew cask template URL matching actual release artifact filenames
  - Corrected WinGet 2.1.0 installer manifest with nullsoft type and correct URL pattern
affects:
  - any future release process (run distribute.yml on publish)
  - winget submission workflow (needs HOMEBREW_TAP_TOKEN and WINGET_TOKEN secrets)

# Tech tracking
tech-stack:
  added: [vedantmgoyal9/winget-releaser@4ffc7888 (v2)]
  patterns: [release:published trigger for post-release distribution, sha256sum for SHA256 on ubuntu runners, curl --retry for CDN propagation resilience, sed template substitution for cask/manifest generation]

key-files:
  created:
    - .github/workflows/distribute.yml
  modified:
    - packaging/homebrew/storcat.rb.template
    - packaging/winget/manifests/s/scottkw/StorCat/2.1.0/scottkw.StorCat.installer.yaml
    - packaging/winget/manifests/s/scottkw/StorCat/2.1.0/scottkw.StorCat.locale.en-US.yaml

key-decisions:
  - "Use release:published trigger (not push:tags) so distribution fires after draft is manually published"
  - "Three independent parallel jobs: update-homebrew, update-winget, update-winget-manifests"
  - "sha256sum (not shasum -a 256) for SHA256 on ubuntu-22.04 runners"
  - "Download DMG/installer to temp file before hashing (not pipe) so curl exit code catches partial downloads"
  - "winget-releaser SHA-pinned to 4ffc7888 (v2), verified against actual repo tag"
  - "WinGet InstallerType changed from portable to nullsoft (NSIS installer)"
  - "VERSION_CLEAN strips v-prefix everywhere; TAG retains v-prefix for download URLs"

patterns-established:
  - "SHA-pin all third-party actions to 40-char commit SHAs with version comment"
  - "Use 2.1.0 manifests as template baseline for sed-based generation of new version manifests"
  - "git diff --cached --quiet guard before commit to make steps idempotent"

requirements-completed: [DIST-01, DIST-02, DIST-03]

# Metrics
duration: 2min
completed: 2026-03-27
---

# Phase 15 Plan 01: Distribution Channel Automation Summary

**distribute.yml workflow with Homebrew cask auto-update and WinGet PR submission on release:published, plus corrected filename patterns in packaging templates**

## Performance

- **Duration:** ~2 min
- **Started:** 2026-03-27T19:57:51Z
- **Completed:** 2026-03-27T19:59:28Z
- **Tasks:** 2
- **Files modified:** 4

## Accomplishments

- Fixed critical filename mismatches in Homebrew cask template (darwin-universal.dmg) and WinGet installer manifest (nullsoft type, NSIS URL pattern)
- Created distribute.yml with three parallel jobs covering all three DIST requirements
- Verified winget-releaser SHA against actual repo (4ffc7888, not placeholder)
- WinGet 2.1.0 stubs now serve as correct template baseline for sed-based generation of new version manifests

## Task Commits

Each task was committed atomically:

1. **Task 1: Fix filename mismatches in Homebrew template and WinGet 2.1.0 stubs** - `9b8d3d82` (fix)
2. **Task 2: Create distribute.yml workflow with Homebrew, WinGet, and manifest update jobs** - `465443e3` (feat)

**Plan metadata:** (see final commit below)

## Files Created/Modified

- `.github/workflows/distribute.yml` - Distribution workflow: three parallel jobs triggered on release:published
- `packaging/homebrew/storcat.rb.template` - Fixed URL to use v-prefixed tag and darwin-universal.dmg filename
- `packaging/winget/manifests/s/scottkw/StorCat/2.1.0/scottkw.StorCat.installer.yaml` - Changed InstallerType to nullsoft, URL to NSIS installer pattern, removed Commands block
- `packaging/winget/manifests/s/scottkw/StorCat/2.1.0/scottkw.StorCat.locale.en-US.yaml` - Fixed ReleaseNotesUrl to include v-prefix on tag

## Decisions Made

- Used `release:published` trigger so distribution fires only after the draft is manually promoted, not on tag push
- Three independent jobs with no `needs:` so Homebrew and WinGet can run in parallel
- Used `sha256sum` (Linux) instead of `shasum -a 256` (macOS) since jobs run on ubuntu-22.04
- Downloaded to temp file before hashing (not piped) to ensure partial download failures are caught
- Verified winget-releaser v2 tag SHA live from GitHub API: `4ffc7888bffd451b357355dc214d43bb9f23917e`
- Homebrew template uses `v#{version}` in both the download path and filename to match actual release asset names

## Deviations from Plan

None - plan executed exactly as written.

## User Setup Required

Two GitHub repository secrets must be added before distribute.yml can run successfully:

| Secret | Source | Scope |
|--------|--------|-------|
| `HOMEBREW_TAP_TOKEN` | GitHub Settings -> Developer Settings -> Personal access tokens -> Fine-grained or Classic PAT with contents:write on scottkw/homebrew-storcat | Cross-repo push to homebrew-storcat tap |
| `WINGET_TOKEN` | GitHub Settings -> Developer Settings -> Personal access tokens -> Tokens (classic) -> Generate new token -> scope: public_repo | PR creation on microsoft/winget-pkgs via winget-releaser |

Add both secrets at: GitHub -> scottkw/storcat -> Settings -> Secrets and variables -> Actions -> New repository secret

Additionally, `scottkw.StorCat` must be manually submitted to microsoft/winget-pkgs (initial PR with 2.1.0 manifests) before winget-releaser automation can work. The 2.1.0 manifests in `packaging/winget/manifests/s/scottkw/StorCat/2.1.0/` are now correct templates (nullsoft type, correct URL pattern) but have a placeholder SHA256 that must be replaced with the real v2.1.0 installer hash for the initial submission.

## Known Stubs

The 2.1.0 WinGet installer manifest (`packaging/winget/manifests/s/scottkw/StorCat/2.1.0/scottkw.StorCat.installer.yaml`) retains `InstallerSha256: 0000000000000000000000000000000000000000000000000000000000000000` (placeholder). This is intentional: the 2.1.0 stubs serve as sed templates for new version generation by distribute.yml. The workflow replaces the zeroed SHA256 with the computed real value for new releases. The 2.1.0 stubs themselves would need the real SHA256 only if manually submitted to winget-pkgs — this is a prerequisite human step documented in User Setup Required above.

## Next Phase Readiness

- distribute.yml is ready to fire on the next published release
- User must add HOMEBREW_TAP_TOKEN and WINGET_TOKEN secrets to repo settings
- User must submit initial v2.1.0 (or v2.2.0) WinGet PR manually to microsoft/winget-pkgs to unblock winget-releaser automation
- Once those prerequisites are met, every future release publish will auto-update both distribution channels

---
*Phase: 15-distribution-channel-automation*
*Completed: 2026-03-27*
