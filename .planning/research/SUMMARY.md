# Project Research Summary

**Project:** StorCat v2.2.0 — CI/CD Release Pipeline and Repo Consolidation
**Domain:** GitHub Actions release automation, cross-platform installer packaging, distribution channel automation
**Researched:** 2026-03-27
**Confidence:** HIGH

## Executive Summary

StorCat v2.2.0 is a pure CI/CD and packaging infrastructure milestone — no application source changes. The goal is to replace a fully manual release process (developer runs scripts, uploads files, submits PRs by hand) with a fully automated pipeline triggered by a single `git push --tags`. Research confirms this is a well-trodden problem for Go desktop apps: the toolchain is `wails build` per native runner + platform-specific packagers (create-dmg, NSIS, linuxdeploy, nfpm) + fan-in release assembly + satellite updates (Homebrew tap push, WinGet PR). GoReleaser, the default choice for Go release automation, is explicitly incompatible with Wails and must not be used.

The recommended architecture is two separate workflows: the existing `build.yml` for CI on push/PR (unchanged), and a new `release.yml` triggered by version tags. The release workflow fans out to three native-runner build jobs in parallel, fans in to a single release-assembly job, then fans out again to two satellite update jobs (Homebrew, WinGet). The repo consolidation portion moves packaging metadata (winget manifests, homebrew cask template, update scripts) from two satellite repos into a `packaging/` directory in the main repo, then archives the `winget-storcat` satellite.

The primary risks are runner selection (NSIS requires `windows-latest`, macOS universals require `macos-latest`), timing (Homebrew SHA256 must be computed locally before upload, not re-downloaded from CDN after), secret management (Homebrew and WinGet each need classic PATs — `GITHUB_TOKEN` is insufficient for both), and supply chain security (all third-party actions must be SHA-pinned after the March 2025 incident that hit the Wails project itself). All risks have clear, established mitigations documented in the research.

## Key Findings

### Recommended Stack

The existing app stack (Go 1.23, Wails v2.10.2, React 18, TypeScript 5, Ant Design 5) is unchanged and must not be upgraded — Wails v2.10.0 is known broken and v3 is alpha. The new infrastructure stack consists of: GitHub Actions with native-runner matrix (macos-14, windows-latest, ubuntu-22.04); `create-dmg` v1.2.3 for DMG wrapping; `wails build -nsis` on Windows runner for installer generation; `linuxdeploy` + AppImage plugin for portable Linux packages; `nfpm` for .deb generation without system packaging toolchain; `softprops/action-gh-release@v2` for release asset upload; `vedantmgoyal9/winget-releaser@v2` for WinGet PR automation; and existing `update-tap.sh` (already written and validated) for Homebrew cask updates.

**Core technologies:**
- `wails build` per native runner — required; GoReleaser cannot drive Wails frontend compilation (confirmed incompatible Dec 2025)
- `create-dmg` v1.2.3 (shell script, Homebrew install) — DMG packaging on macOS; no Node.js dependency
- `wails build -nsis` on `windows-latest` — NSIS installer; silently fails on non-Windows runners
- `linuxdeploy` (continuous/rolling) — AppImage creation, CI-friendly, correct complexity level for a Go binary
- `nfpm` v2.x — declarative deb/rpm packaging in pure Go; zero system dependencies on ubuntu runner
- `softprops/action-gh-release@v2` — industry standard for GitHub release asset upload; supports fan-in pattern
- `vedantmgoyal9/winget-releaser@v2` — Komac-backed WinGet PR automation; 284 stars, actively maintained
- Custom `update-tap.sh` (existing) — Homebrew cask push; already validated; no third-party action needed
- `actions/upload-artifact@v4` + `actions/download-artifact@v4` — must match versions; v3 deprecated Jan 2025 and stopped working

### Expected Features

Research confirmed all features required for v2.2.0 (P1), features to add post-validation (P2), and what to explicitly defer.

**Must have (P1 — v2.2.0):**
- Release workflow triggered on version tag push (`on: push: tags: v*.*.*`)
- Pre-flight version consistency check: git tag must match `wails.json productVersion`
- macOS universal build → unsigned DMG attached to release
- Windows amd64 build → NSIS installer + portable .exe attached to release
- Linux amd64 build → AppImage attached to release
- Linux amd64 build → .deb attached to release
- SHA256 checksums.txt attached to release
- Homebrew cask auto-updated in `scottkw/homebrew-storcat` on every release
- WinGet manifests migrated from `winget-storcat` to `packaging/winget/` in main repo
- WinGet manifest update on release (PR to `microsoft/winget-pkgs`)
- `winget-storcat` repo archived after first successful automated release

**Should have (P2 — v2.2.x):**
- Auto-PR to `microsoft/winget-pkgs` via `winget-releaser` — once package is confirmed in registry
- Linux arm64 AppImage + .deb via QEMU cross-compile
- `workflow_dispatch` trigger for release re-runs without cutting a new tag

**Defer to own milestone (P3):**
- macOS code signing + notarization — requires Apple Developer credentials; multi-day setup with Apple account dependencies
- Windows Authenticode signing — requires EV cert ($200-500/yr) or Partner Center enrollment
- Linux RPM packaging — add only if user demand materializes; AppImage covers all distros

### Architecture Approach

The pipeline is a classic fan-out / fan-in DAG. Three platform build jobs run in parallel (macos-14, windows-latest, ubuntu-22.04), each uploading artifacts via `actions/upload-artifact@v4`. A single `create-release` job depends on all three via `needs:`, downloads all artifacts, computes SHA256 checksums, and publishes the GitHub release. Two satellite jobs (`update-homebrew`, `update-winget`) run after `create-release` with external side effects. The repo consolidation is entirely additive — a new `packaging/` directory, a new `release.yml` workflow, and one correction to `build.yml` (move Windows job from `macos-latest` to `windows-latest`).

**Major components:**
1. `.github/workflows/release.yml` (NEW) — full release pipeline; tag-triggered; fan-out build → fan-in release → fan-out satellites
2. `.github/workflows/build.yml` (MODIFIED) — CI checks on push/PR; move Windows job to `windows-latest`; otherwise unchanged
3. `packaging/homebrew/` (NEW) — `storcat.rb.template` with `{{VERSION}}` and `{{SHA256_DMG}}` placeholders; `update-tap.sh` migrated here
4. `packaging/winget/manifests/` (MIGRATED) — versioned YAML manifests moved from `scottkw/winget-storcat`; archived after migration
5. `scottkw/homebrew-storcat` (UNCHANGED REPO) — thin tap; auto-updated by pipeline; never touched manually again

### Critical Pitfalls

The research identified 10 critical and 4 moderate pitfalls. The top 5 that must be addressed before any code is written:

1. **NSIS requires Windows runner** — `wails build -nsis` silently fails on macOS/Linux. The existing `build.yml` incorrectly builds Windows on `macos-latest`. The `release.yml` Windows job must run on `windows-latest` from the start; this is a structural requirement, not a configuration detail.

2. **Parallel release upload race condition** — Three platform jobs each calling `gh release create` produces "release already exists" errors on 2 of 3 jobs. Use fan-in: build jobs upload artifacts only; a single `create-release` job downloads all artifacts and publishes once. This must be the baseline design; it cannot be retrofitted.

3. **Homebrew SHA256 CDN propagation delay** — Re-downloading the DMG from GitHub CDN immediately after release creation can return stale content (30-120 second propagation window), producing a wrong SHA256 that breaks `brew install` for all users. Compute SHA256 locally from the built artifact before upload; pass as a workflow output variable. Never re-download from CDN to get the hash.

4. **GITHUB_TOKEN insufficient for cross-repo operations** — Both Homebrew tap push and WinGet PR submission require explicit classic PATs. `GITHUB_TOKEN` is scoped to the current repo by design. Homebrew needs a PAT with `contents: write` on `homebrew-storcat` stored as `HOMEBREW_TAP_TOKEN`. WinGet needs a classic PAT with `public_repo` scope stored as `WINGET_TOKEN`. Fine-grained PATs do not work for either use case.

5. **Third-party action supply chain risk** — The March 2025 `tj-actions/changed-files` compromise directly affected Wails project CI and 23,000+ other repos. All community actions must be pinned to commit SHA, not mutable version tags (`@v1`, `@v2`, `@latest` are all unsafe for third-party actions). `actions/*` first-party actions can use version tags.

Secondary pitfalls:
- Linux WebKit: `ubuntu-latest` now resolves to Ubuntu 24.04 which drops `libwebkit2gtk-4.0-dev`; pin to `ubuntu-22.04` or add `-tags webkit2_41`
- macOS runner drift: `macos-latest` changes silently; pin to `macos-14`; add `lipo -info` to verify universal binary
- WinGet first submission must be manual; `winget-releaser` automation only works after package exists in `winget-pkgs`
- Version validation pre-flight must exit non-zero before any build if git tag != `wails.json productVersion`

## Implications for Roadmap

The dependency graph dictates a strict 4-phase sequence. Packaging infrastructure must exist before automation references it. CI scaffold must be correct before packaging layers are added. Homebrew and WinGet automation have external side effects and must come last, after the release pipeline is proven with a test tag.

### Phase 1: Repo Consolidation and Packaging Infrastructure

**Rationale:** Moving packaging artifacts into the main repo before adding any automation is prerequisite work with zero CI risk. `release.yml` will reference `packaging/homebrew/storcat.rb.template` and `packaging/winget/` — these paths must exist and be validated before any workflow is written. No workflow changes in this phase.

**Delivers:** `packaging/` directory structure committed; `storcat.rb.template` in place; WinGet manifests migrated from `winget-storcat`; update scripts relocated to `packaging/`; local verification that scripts work from new paths.

**Addresses:** Table-stakes features — WinGet manifests in main repo; packaging source of truth consolidated.

**Avoids:** Building automation against directory paths that haven't been validated; keeps `build.yml` green throughout.

**Research flag:** No deeper research needed. Directory layout and file migration are mechanical steps with no technical risk.

### Phase 2: CI Scaffold and Multi-Platform Build

**Rationale:** Runner selection and the fan-in DAG pattern are architectural decisions that affect every subsequent phase. All critical pitfalls related to NSIS, macOS cross-compile, WebKit, parallel race conditions, SHA pinning, and trigger strategy must be locked down here on a minimal workflow. Once the scaffold is correct, packaging layers can be added without restructuring.

**Delivers:** `release.yml` with correct matrix (macos-14, windows-latest, ubuntu-22.04); fan-in `create-release` job pattern; raw binary artifacts uploaded and visible in GitHub release; workflow fires correctly on `git push --tags`. `build.yml` corrected to use `windows-latest` for Windows job.

**Addresses:** Pre-flight version check; release trigger strategy; artifact upload/download v4 version lock; SHA pinning for all community actions.

**Avoids:** NSIS on wrong runner (Pitfall 1); parallel release upload race (Pitfall 4); `ubuntu-latest` webkit breakage (Pitfall 3 in PITFALLS.md); supply chain risk (Pitfall 8).

**Research flag:** Well-documented patterns. Confirm that `macos-14` (arm64) can produce `darwin/universal` fat binaries; add `lipo -info` verification step to surface any runner-arch issues immediately.

### Phase 3: Platform Packaging (DMG, NSIS, AppImage, deb)

**Rationale:** Add packaging steps to each platform job after confirming raw binaries build correctly on the scaffold. Order within the phase: macOS DMG first (simplest, well-documented action), Windows NSIS (native to Wails, already works via `-nsis` flag), Linux AppImage + deb (most moving parts; `linuxdeploy` + `nfpm` need a `.desktop` file and `nfpm.yaml`). Each is independently verifiable before the next is added.

**Delivers:** All 7 release assets (unsigned DMG, NSIS installer, portable .exe, AppImage amd64, deb, SHA256SUMS.txt) attached to the GitHub release automatically on tag push.

**Uses:** `create-dmg` v1.2.3 on macOS job; `wails build -nsis` on Windows job; `linuxdeploy` + `nfpm` on Linux job; SHA256 computed locally and passed as workflow output before upload.

**Avoids:** Homebrew SHA CDN delay (SHA computed pre-upload and passed as output variable for use in Phase 4).

**Research flag:** Linux AppImage WebKit runtime bundling requires execution-time validation — the AppImage must launch on a clean Ubuntu VM to confirm WebKitNetworkProcess is bundled. This is a known Wails issue (#4313). Treat as a required verification gate before Phase 3 is closed; if AppImage WebKit bundling fails, `.deb` becomes the primary Linux artifact.

### Phase 4: Distribution Channel Automation (Homebrew and WinGet)

**Rationale:** Satellite updates have external side effects (push to another repo, PR to `microsoft/winget-pkgs`). They must come last, after Phase 3 is proven with a test release candidate tag. Secret configuration (PATs) is a prerequisite that can be set up in parallel with Phase 3.

**Delivers:** Homebrew cask auto-updated in `scottkw/homebrew-storcat` on every release; WinGet PR submitted to `microsoft/winget-pkgs`; `winget-storcat` repo archived.

**Uses:** Existing `update-tap.sh` wired to `HOMEBREW_TAP_TOKEN` secret; SHA passed as workflow output from Phase 3 (no CDN re-download); `vedantmgoyal9/winget-releaser@v2` SHA-pinned with `WINGET_TOKEN` classic PAT.

**Avoids:** CDN SHA delay (pass SHA as env var from Phase 3 job output); cross-repo `GITHUB_TOKEN` failure (separate PATs); WinGet automation before manual submission is confirmed.

**Research flag:** Verify whether `scottkw.StorCat` (or `KenScott.StorCat`) already exists in `microsoft/winget-pkgs` before enabling `winget-releaser`. If not listed, manual first-submission is a prerequisite gate that is not part of the automated pipeline — it requires a PR review cycle from Microsoft that can take 24-72 hours.

### Phase Ordering Rationale

- Phase 1 before Phase 2: `release.yml` references `packaging/` paths; those paths must exist before the workflow is written.
- Phase 2 before Phase 3: Runner selection and fan-in structure are architectural; retrofitting them after packaging layers are added is error-prone and risky.
- Phase 3 before Phase 4: Satellite updates depend on artifact URLs and SHA values that are only stable and proven after packaging is confirmed working.
- Homebrew (Phase 4a) before WinGet (Phase 4b): Homebrew is lower risk — you own the tap and can fix mistakes instantly. WinGet requires an external PR review cycle.

### Research Flags

Phases needing deeper investigation during planning or execution:

- **Phase 4 (WinGet identifier):** Whether `scottkw.StorCat` already exists in `winget-pkgs` is unknown. Must verify before writing automation. Check exact identifier capitalization — mismatch causes silent failure.
- **Phase 3 (AppImage WebKit):** Confirmed known issue (#4313); cannot be resolved by research alone. Requires testing the built AppImage on a clean Ubuntu VM as an execution gate.
- **Phase 2 (macos-14 universal binary):** Confirm `darwin/universal` fat binary production on arm64 runner with `lipo -info`; surface any toolchain gaps before packaging layers are added.

Phases with standard patterns (skip research-phase):

- **Phase 1:** Mechanical file migration; no technical ambiguity.
- **Phase 2:** GitHub Actions matrix + fan-in is exhaustively documented with official examples.

## Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| Stack | HIGH | Versions verified via direct fetch; GoReleaser incompatibility confirmed by community blog (Dec 2025); v3/v4 artifact action deprecation confirmed from official GitHub changelog; Wails v2.10.2 pinning confirmed from existing `go.mod` |
| Features | HIGH | Feature set derived directly from existing repo state (`build.yml`, satellite repos, `storcat-repo-consolidation.md`) — all known quantities; P1/P2/P3 boundaries are clearly supported by complexity analysis |
| Architecture | HIGH | Based on direct analysis of existing `build.yml`, `wails.json`, and `scripts/`; fan-in DAG pattern is industry-standard GitHub Actions practice with multiple reference implementations |
| Pitfalls | HIGH | Critical pitfalls verified against official Wails issues (#3581, #1714, #4313), GitHub advisory (CVE-2025-30066), and `softprops/action-gh-release` issue #602; macOS runner and Ubuntu version drift confirmed from runner-images changelog |

**Overall confidence:** HIGH

### Gaps to Address

- **WinGet identifier and existing listing status:** Whether `scottkw.StorCat` exists in `microsoft/winget-pkgs` is unknown. Verify before Phase 4 begins. If not listed, manual first-submission must be tracked as a prerequisite task.
- **AppImage WebKit runtime bundling:** Known Wails issue #4313 confirms the risk; the resolution is an execution-time test on a clean VM, not researchable in advance. Treat as Phase 3 verification gate.
- **macOS arm64 runner universal binary:** `macos-14` is arm64; `darwin/universal` requires Intel cross-compilation toolchain. Research says it is supported, but this must be confirmed with `lipo -info` in Phase 2 before packaging layers are added.

## Sources

### Primary (HIGH confidence)
- `.github/workflows/build.yml` (codebase) — existing CI structure, runner choices, artifact patterns
- `wails.json` (codebase) — confirms Wails v2.10.2, Go 1.23, productVersion field
- `storcat-repo-consolidation.md` (project-internal) — repo layout, Option A decision, existing scripts
- [create-dmg GitHub](https://github.com/create-dmg/create-dmg) — v1.2.3 Nov 2025, fetched directly
- [softprops/action-gh-release](https://github.com/softprops/action-gh-release) — v2.5.0, current standard
- [winget-releaser Marketplace](https://github.com/marketplace/actions/winget-releaser) — v2, fetched directly
- [mislav/bump-homebrew-formula-action](https://github.com/mislav/bump-homebrew-formula-action) — limited cask support documented, fetched directly
- [actions/upload-artifact deprecation notice](https://github.blog/changelog/2024-04-16-deprecation-notice-v3-of-the-artifact-actions/) — official GitHub changelog; v3 stopped working Jan 2025
- [Wails NSIS cross-compile error #1714](https://github.com/wailsapp/wails/issues/1714) — confirmed failure mode
- [tj-actions supply chain CVE-2025-30066](https://github.com/advisories/ghsa-mrrh-fwg8-r2c3) — incident directly affecting Wails project CI
- [softprops/action-gh-release race condition #602](https://github.com/softprops/action-gh-release/issues/602) — fan-in pattern confirmed as solution
- [macOS 26 runner GA](https://github.blog/changelog/2026-02-26-macos-26-is-now-generally-available-for-github-hosted-runners/) — runner landscape

### Secondary (MEDIUM confidence)
- [GoReleaser + Wails incompatibility](https://chriswheeler.dev/posts/cross-compilation-with-wails/) — Dec 2025 community blog confirming failure
- [Wails Ubuntu 24.04 webkit issue #3581](https://github.com/wailsapp/wails/issues/3581) — webkit2_41 workaround confirmed
- [AppImage WebKitNetworkProcess runtime issue #4313](https://github.com/wailsapp/wails/issues/4313) — known Wails AppImage pitfall
- [Homebrew tap automation pattern](https://builtfast.dev/blog/automating-homebrew-tap-updates-with-github-actions/) — community blog
- [nfpm goreleaser](https://github.com/goreleaser/nfpm) — Go-native deb/rpm packager, active 2025
- [linuxdeploy AppImage toolchain](https://docs.appimage.org/packaging-guide/from-source/linuxdeploy-user-guide.html)
- [Wails NSIS installer guide](https://wails.io/docs/guides/windows-installer/) — direct fetch returned 403; confirmed via search results
- [StepSecurity SHA pinning guide](https://www.stepsecurity.io/blog/pinning-github-actions-for-enhanced-security-a-complete-guide)

---
*Research completed: 2026-03-27*
*Ready for roadmap: yes*
