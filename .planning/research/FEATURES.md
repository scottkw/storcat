# Feature Research

**Domain:** CI/CD release pipeline + repo consolidation for a Go/Wails cross-platform desktop app
**Researched:** 2026-03-27
**Confidence:** HIGH (core pipeline patterns), MEDIUM (Linux .deb packaging specifics)

---

## Context

StorCat v2.2.0 is adding automated build + release infrastructure to an existing Go/Wails desktop
app. The existing `build.yml` does CI checks (build on push/PR) but no packaging or release
automation. Three repos currently exist: `scottkw/storcat` (source), `scottkw/homebrew-storcat`
(tap), `scottkw/winget-storcat` (manifests). All three must be consolidated to two (thin homebrew
satellite stays, winget manifests move into main), and the manual release process must be
automated.

This document focuses **only on the new features** needed for v2.2.0.

---

## Feature Landscape

### Table Stakes (Users Expect These)

Features the pipeline must have to be considered complete. Missing any of these means the release
process still requires manual intervention.

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| Tag/release-triggered release workflow | Industry standard: GH release published triggers full build + packaging | LOW | `on: release: types: [published]` in `release.yml`; separate from existing `build.yml` |
| macOS DMG packaging | macOS users expect a .dmg for drag-to-Applications install | MEDIUM | Post-process `.app` with `create-dmg` or `hdiutil`; no code signing in this milestone |
| Windows NSIS installer | Windows users expect an .exe installer, not a bare binary | LOW | Wails native support: `wails build -nsis`; generates `build/bin/StorCat-*-installer.exe` |
| Linux AppImage | Most portable Linux format; runs on Ubuntu, Fedora, Arch, etc. without changes | MEDIUM | Wails v2 generates AppImage with `wails build -platform linux/amd64`; arm64 needs QEMU or native runner |
| Linux .deb package | Expected by Ubuntu/Debian users (dominant desktop Linux install base) | MEDIUM | Wails v2 does NOT natively produce .deb; requires `dpkg-deb` post-build step or `fpm` tool |
| Multi-platform matrix build | All platforms must build in one workflow, in parallel | LOW | `strategy.matrix` with `macos-latest`, `windows-latest`, `ubuntu-latest` runners |
| SHA256 checksums as release asset | Required by Homebrew and WinGet for artifact verification; users also rely on these | LOW | `shasum -a 256 dist/*` before upload; include `checksums.txt` in release assets |
| All artifacts attached to GitHub Release | Release assets available for download; Homebrew/WinGet consume URLs from the release | LOW | `softprops/action-gh-release` or `actions/upload-release-asset` |
| Homebrew cask auto-update on release | Mac users install via `brew install storcat`; cask must be updated on every release | MEDIUM | Render `storcat.rb` from template in main repo; push to `scottkw/homebrew-storcat` via PAT |
| WinGet manifest update on release | Windows users install via `winget install scottkw.StorCat`; manifests must stay current | MEDIUM | Commit updated YAML manifests under `packaging/winget/manifests/`; optionally PR `microsoft/winget-pkgs` |
| `packaging/homebrew/storcat.rb.template` in main repo | Homebrew template lives with the code that produces the binary it describes | LOW | Copy from `homebrew-storcat`; parameterize `version`, `url`, `sha256` |
| WinGet manifests migrated to `packaging/winget/` | Three YAML files per version live alongside source; `winget-storcat` repo becomes redundant | LOW | Copy from `winget-storcat`; update paths in update script |
| `winget-storcat` repo archived | Eliminates stale satellite repo once automation is proven | LOW | GitHub repo Settings → Archive; one-time action after first automated release succeeds |

### Differentiators (Competitive Advantage)

Features beyond the minimum that make this pipeline better than a simple manual process.

| Feature | Value Proposition | Complexity | Notes |
|---------|-------------------|------------|-------|
| `vedantmgoyal9/winget-releaser` action | Handles manifest generation + PR submission to `microsoft/winget-pkgs` via Komac; eliminates manual manifest authoring | LOW | Requires classic PAT (`public_repo` scope); prerequisite: package already listed in `winget-pkgs` |
| `homebrew-bump-cask` or `update-homebrew-action` | Tested, maintained action handles SHA calc + formula update; avoids custom push script | LOW | `garden-io/update-homebrew-action` or `NSHipster/update-homebrew-formula-action` are proven alternatives to hand-rolled scripts |
| Separate build jobs from release/packaging job | Build jobs run in parallel; release job depends on all via `needs:`; one platform failure does not block others | MEDIUM | Use `needs: [build-macos, build-windows, build-linux]` + artifact download pattern |
| `workflow_dispatch` trigger | Allows re-running release packaging without cutting a new tag; useful when a packaging step fails after binaries are good | LOW | Add alongside `on: release:` |
| Version consistency check | Fail fast if git tag does not match `wails.json productVersion` before any builds run | LOW | `bash` step: compare `git describe --tags` vs `jq -r .info.productVersion wails.json` |
| Per-platform artifact names with version | Release assets named `StorCat-v2.2.0-macOS.dmg`, `StorCat-v2.2.0-Windows-x64-installer.exe` | LOW | `github.ref_name` gives the tag; embed in artifact name |

### Anti-Features (Commonly Requested, Often Problematic)

| Feature | Why Requested | Why Problematic | Alternative |
|---------|---------------|-----------------|-------------|
| macOS code signing + notarization in v2.2.0 | Required for Gatekeeper-clean installs | Needs Apple Developer Program cert, provisioning profile, App Store Connect API key, `codesign`, `xcrun notarytool` — multi-day setup with Apple account dependencies | Out of scope; deliver unsigned DMG; create dedicated CODE-SIGNING milestone |
| Windows Authenticode signing in v2.2.0 | Removes SmartScreen "unknown publisher" warning | Requires EV cert ($200-500/yr) or Microsoft Partner Center enrollment; own milestone | Out of scope; deliver unsigned NSIS installer |
| GoReleaser for the release pipeline | Popular Go release tool with Homebrew and deb/rpm support | GoReleaser cannot drive `wails build` (Wails wraps frontend compilation; GoReleaser only handles pure `go build`). Incompatible at the build step. | Stick with native `wails build` + per-platform jobs; GoReleaser is only appropriate for the pure CLI binary path |
| Auto-creating GitHub Release from the workflow | One less manual step | Removes human editorial control over release notes; creates a release before notes are written | Trigger on already-published release (`on: release: types: [published]`); developer writes release notes manually, then publishes to trigger pipeline |
| Immediately auto-submitting to `microsoft/winget-pkgs` | One less manual step | First submission requires Microsoft review via PR; auto-submission can fail silently or be rejected without context. WinGet Releaser handles this for subsequent versions, but first version must be manually submitted. | Use `winget-releaser` for updates once package is listed; document first-submission as a one-time manual step |
| Linux RPM packaging in v2.2.0 | Fedora/RHEL user request | Low demand relative to AppImage; adds packaging complexity; AppImage runs on RPM-based distros too | Deliver AppImage (universal) + .deb; RPM deferred to P3/future request |
| Self-hosted runners for Linux arm64 | Faster arm64 Linux builds | Requires persistent infra, runner credentials, maintenance, security surface | Cross-compile with QEMU via `docker/setup-qemu-action` on `ubuntu-latest`; acceptable for release cadence |
| Merging `build.yml` and `release.yml` into one file | Fewer workflow files | Mixing CI checks (runs on every push) with release packaging (runs on release) creates confusion, inflated CI times, and harder debugging | Keep separate: `build.yml` for CI checks, `release.yml` for release packaging |

---

## Feature Dependencies

```
[Developer creates GitHub Release manually — writes release notes, publishes]
    └──triggers──> [release.yml workflow: on: release: types: [published]]
                       │
                       ├──[Pre-flight job]
                       │       └──validates: git tag matches wails.json productVersion
                       │
                       ├──[macOS build job: macos-latest]
                       │       ├──wails build -platform darwin/universal
                       │       ├──create-dmg / hdiutil → StorCat-vX.Y.Z-macOS.dmg
                       │       └──upload artifact: DMG
                       │
                       ├──[Windows build job: windows-latest]
                       │       ├──wails build -platform windows/amd64 -nsis
                       │       ├──produces: StorCat-vX.Y.Z-Windows-x64-installer.exe
                       │       └──upload artifact: NSIS installer
                       │
                       ├──[Linux build job: ubuntu-latest]
                       │       ├──wails build -platform linux/amd64
                       │       ├──produces: AppImage (wails native output)
                       │       ├──dpkg-deb / fpm → .deb package
                       │       └──upload artifacts: AppImage + .deb
                       │
                       └──[Release assembly job: needs all build jobs]
                               ├──download all artifacts
                               ├──compute SHA256 checksums.txt
                               ├──attach all artifacts + checksums to GitHub Release
                               ├──[Homebrew update job]
                               │       ├──requires: DMG URL + SHA256
                               │       ├──renders: storcat.rb from packaging/homebrew/storcat.rb.template
                               │       └──pushes to: scottkw/homebrew-storcat Casks/storcat.rb
                               │               └──requires secret: HOMEBREW_TAP_TOKEN (PAT)
                               └──[WinGet manifest job]
                                       ├──requires: Windows installer URL + SHA256
                                       ├──generates: 3 YAML manifests for new version
                                       ├──commits to: packaging/winget/manifests/s/scottkw/StorCat/vX.Y.Z/
                                       └──optionally PRs: microsoft/winget-pkgs
                                               └──requires secret: WINGET_TOKEN (classic PAT, public_repo scope)

[packaging/homebrew/storcat.rb.template] (migrated from homebrew-storcat)
    └──used by──> Homebrew update job

[packaging/winget/manifests/] (migrated from winget-storcat)
    └──base for──> WinGet manifest job (adds new version directory)

[scottkw/winget-storcat] (archived after first successful automated release)
```

### Dependency Notes

- **Pre-flight validation is a hard gate.** A mismatch between the git tag and `wails.json productVersion` would produce a binary reporting the wrong version. This check must pass before any build starts.
- **All build jobs must complete before release assembly.** The assembly job uses `needs: [build-macos, build-windows, build-linux]` and downloads artifacts using `actions/download-artifact`.
- **Homebrew update requires the DMG to be attached to the release.** The cask formula needs an exact URL and SHA256; this is available only after the release assembly job completes.
- **WinGet manifest job requires the Windows installer to be attached to the release.** Same pattern as Homebrew.
- **`winget-storcat` archive is a one-time post-verification step.** Do not archive until at least one automated release has successfully committed manifests to `packaging/winget/`.
- **Windows NSIS build must run on `windows-latest`.** The existing `build.yml` runs Windows builds on `macos-latest` via cross-compile. NSIS packaging requires a Windows runner.

---

## MVP Definition

This is a subsequent milestone on a shipping product.

### Launch With (v2.2.0)

Minimum viable automated pipeline — zero manual steps after publishing the GitHub release.

- [ ] `release.yml` triggered on `on: release: types: [published]`
- [ ] Pre-flight: version consistency check (wails.json vs git tag)
- [ ] macOS universal build → unsigned DMG attached to release
- [ ] Windows amd64 build → NSIS installer attached to release
- [ ] Linux amd64 build → AppImage attached to release
- [ ] Linux amd64 build → .deb attached to release
- [ ] SHA256 checksums.txt attached to release
- [ ] Homebrew cask auto-updated in `scottkw/homebrew-storcat`
- [ ] WinGet manifests committed to `packaging/winget/` on release
- [ ] `packaging/homebrew/storcat.rb.template` in main repo
- [ ] WinGet manifests migrated from `winget-storcat` to `packaging/winget/`
- [ ] `winget-storcat` repo archived

### Add After Validation (v2.2.x)

- [ ] Auto-PR to `microsoft/winget-pkgs` via `winget-releaser` — trigger once package is confirmed listed in official registry
- [ ] Linux arm64 AppImage + .deb — add QEMU cross-compile or native arm64 runner
- [ ] `workflow_dispatch` trigger for release re-runs

### Future Consideration (v3+, own milestones)

- [ ] macOS code signing + notarization — dedicated CODE-SIGNING milestone; Apple credential pipeline
- [ ] Windows Authenticode signing — dedicated milestone; EV cert procurement
- [ ] Linux RPM packaging — add if user demand materializes
- [ ] Release notes auto-generation from conventional commits

---

## Feature Prioritization Matrix

| Feature | User Value | Implementation Cost | Priority |
|---------|------------|---------------------|----------|
| release.yml triggered on GH release published | HIGH | LOW | P1 |
| macOS DMG (unsigned) | HIGH | MEDIUM | P1 |
| Windows NSIS installer (unsigned) | HIGH | LOW | P1 |
| Linux AppImage amd64 | HIGH | MEDIUM | P1 |
| Linux .deb amd64 | MEDIUM | MEDIUM | P1 |
| SHA256 checksums.txt | HIGH | LOW | P1 |
| Homebrew cask auto-update | HIGH | MEDIUM | P1 |
| WinGet manifests migrated to main repo | HIGH | LOW | P1 |
| WinGet manifest commit on release | HIGH | LOW | P1 |
| `winget-storcat` repo archived | MEDIUM | LOW | P1 |
| Version consistency pre-flight check | MEDIUM | LOW | P1 |
| Auto-PR to microsoft/winget-pkgs | MEDIUM | LOW | P2 |
| Linux arm64 AppImage + .deb | MEDIUM | MEDIUM | P2 |
| `workflow_dispatch` re-run trigger | LOW | LOW | P2 |
| macOS code signing + notarization | HIGH | HIGH | P3 (own milestone) |
| Windows Authenticode signing | HIGH | HIGH | P3 (own milestone) |
| Linux RPM packaging | LOW | MEDIUM | P3 (on request) |

**Priority key:**
- P1: Must have for v2.2.0 launch
- P2: Should have, add in v2.2.x
- P3: Valuable but own milestone or deferred by demand

---

## Existing Infrastructure Notes

| Existing Item | Impact on New Features |
|---------------|------------------------|
| `build.yml` — CI checks on push/PR | Keep as-is; new `release.yml` is a separate file; do not merge them |
| `wails build` invocations in `scripts/` | Release workflow reuses the same `wails build` commands; no changes to build scripts |
| `//go:embed wails.json` for version | `wails.json productVersion` must be bumped before tagging; pre-flight check enforces this |
| macOS `darwin/universal` build on `macos-latest` | Continue using this runner and platform flag |
| Windows build on `macos-latest` (cross-compile, no NSIS) | Move Windows build to `windows-latest` runner; NSIS tool requires Windows |
| Linux build on `ubuntu-latest` | Continue using; add AppImage + .deb packaging steps after binary build |
| `scottkw/homebrew-storcat` repo | Stays alive; auto-updated by pipeline; README updated to note it is auto-managed |
| `scottkw/winget-storcat` repo | Manifests migrated to main repo; repo archived after first successful automated release |

---

## Sources

- [Wails Build GitHub Action (marketplace)](https://github.com/marketplace/actions/wails-build) — HIGH confidence
- [Wails Action CI/CD (marketplace)](https://github.com/marketplace/actions/wails-action-ci-cd) — HIGH confidence
- [WinGet Releaser action (vedantmgoyal9)](https://github.com/vedantmgoyal9/winget-releaser) — HIGH confidence
- [microsoft/winget-create](https://github.com/microsoft/winget-create) — HIGH confidence
- [Homebrew Bump Cask action (marketplace)](https://github.com/marketplace/actions/homebrew-bump-cask) — HIGH confidence
- [garden-io/update-homebrew-action](https://github.com/garden-io/update-homebrew-action) — MEDIUM confidence
- [NSHipster/update-homebrew-formula-action](https://github.com/NSHipster/update-homebrew-formula-action) — MEDIUM confidence
- [create-dmg/create-dmg](https://github.com/create-dmg/create-dmg) — HIGH confidence
- [AppImage-builder GitHub Actions guide](https://appimage-builder.readthedocs.io/en/latest/hosted-services/github-actions.html) — MEDIUM confidence
- [GoReleaser GitHub Action](https://github.com/goreleaser/goreleaser-action) — HIGH confidence (evaluated and excluded: incompatible with Wails build pipeline)
- [storcat-repo-consolidation.md](/Users/ken/dev/storcat/storcat-repo-consolidation.md) — project-internal context, HIGH confidence

---

*Feature research for: StorCat v2.2.0 CI/CD release pipeline + repo consolidation*
*Researched: 2026-03-27*
