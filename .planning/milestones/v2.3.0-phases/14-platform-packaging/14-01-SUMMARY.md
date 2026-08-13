---
phase: 14-platform-packaging
plan: "01"
subsystem: ci-cd
tags: [packaging, github-actions, dmg, nsis, appimage, deb]
dependency_graph:
  requires: [13-01]
  provides: [PKG-01, PKG-02, PKG-03, PKG-04]
  affects: [.github/workflows/release.yml]
tech_stack:
  added: [create-dmg, linuxdeploy, dpkg-deb]
  patterns: [platform-packaging, freedesktop-desktop-entry]
key_files:
  created:
    - build/linux/storcat.desktop
  modified:
    - .github/workflows/release.yml
decisions:
  - "windows-2022 pinned (NSIS removed from windows-2025 image)"
  - "AppImage declares libwebkit2gtk-4.0-37 as system dependency (WebKit bundling not feasible)"
  - "linuxdeploy --plugin gtk for GTK schemas/resources (not WebKit itself)"
  - "printf over heredoc for DEBIAN/control (avoids YAML multi-line quoting issues)"
metrics:
  duration: "2 minutes"
  completed: "2026-03-27"
  tasks_completed: 2
  tasks_total: 2
  files_changed: 2
---

# Phase 14 Plan 01: Platform Packaging Summary

## One-liner

Extended release workflow with macOS DMG via create-dmg, Windows NSIS installer via wails -nsis on windows-2022, Linux AppImage via linuxdeploy, and .deb packages via dpkg-deb for both amd64 and arm64.

## What Was Built

### Task 1: Linux desktop entry file

Created `build/linux/storcat.desktop` — a freedesktop.org desktop entry file required by both linuxdeploy (AppImage creation) and the .deb package build. The file follows the standard spec format with `Type=Application`, `Exec=StorCat`, `Icon=appicon`, and `Terminal=false`.

### Task 2: Extended release.yml with packaging steps

Updated `.github/workflows/release.yml` to add packaging steps to all four platform build jobs:

**build-macos (PKG-01):**
- Added `npm install --global create-dmg` step
- Added `create-dmg --overwrite --dmg-title "StorCat" --no-code-sign` step after tarball creation
- Updated upload-artifact to glob both `.tar.gz` and `.dmg`

**build-windows (PKG-02):**
- Changed runner from `windows-latest` to `windows-2022` (NSIS removed from windows-2025 image)
- Replaced `Build Windows` step with `Build Windows with NSIS installer` using `-nsis` flag
- Added `Rename NSIS installer` step for versioned naming (`StorCat-VERSION-windows-amd64-installer.exe`)
- Updated upload-artifact to include both standalone exe and installer exe

**build-linux-amd64 (PKG-03 + PKG-04):**
- Added `Install AppImage tools` step: libfuse2, linuxdeploy, linuxdeploy-plugin-appimage, linuxdeploy-plugin-gtk.sh
- Added `Create AppImage` step with `--plugin gtk --output appimage`, then mv to versioned name in build/bin/
- Added `Create .deb package` step using dpkg-deb with DEBIAN/control (auto-detects arch, strips 'v' prefix from version)
- Updated upload-artifact to include tarball, AppImage, and .deb

**build-linux-arm64 (PKG-04):**
- Added identical `Create .deb package` step (dpkg --print-architecture returns arm64 on this runner)
- Updated upload-artifact to include tarball and .deb

## Decisions Made

| Decision | Rationale |
|----------|-----------|
| Pin Windows runner to `windows-2022` | NSIS 3.10 present on 2022; removed from windows-2025 image (Sept 2025 migration) |
| AppImage depends on system `libwebkit2gtk-4.0-37` | WebKit subprocess binaries use hardcoded absolute paths; bundling via linuxdeploy causes runtime failures on non-matching systems |
| `linuxdeploy --plugin gtk` (not WebKit) | GTK schemas/resources can be bundled; WebKit cannot — use system dependency |
| `printf` for DEBIAN/control | Avoids heredoc YAML quoting issues in multi-line GitHub Actions run blocks |
| Build AppImage on amd64 only | No linuxdeploy arm64 equivalent available; arm64 gets .deb only |

## Verification Results

All patterns confirmed present:
- PKG-01 (DMG): `create-dmg` found 3 times in release.yml
- PKG-02 (NSIS): `-nsis` flag found, `windows-2022` runner confirmed, `windows-latest` absent
- PKG-03 (AppImage): `linuxdeploy` found 5 times in release.yml
- PKG-04 (.deb): `dpkg-deb` found 2 times in release.yml
- YAML syntax: valid (Python yaml.safe_load passes)
- Desktop entry: `build/linux/storcat.desktop` exists with correct format

## Deviations from Plan

None - plan executed exactly as written.

## Known Stubs

None — all packaging steps are fully implemented. The AppImage system-WebKit dependency is an intentional architectural decision documented in research, not a stub.

## Self-Check: PASSED

Files verified:
- FOUND: build/linux/storcat.desktop
- FOUND: .github/workflows/release.yml (modified)

Commits verified:
- f37c6c06: chore(14-01): add Linux freedesktop desktop entry file
- 29d57058: feat(14-01): extend release workflow with platform packaging steps
