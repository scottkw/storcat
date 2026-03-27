# Stack Research

**Domain:** CI/CD release pipeline for cross-platform Go/Wails desktop app
**Researched:** 2026-03-27
**Confidence:** HIGH (GitHub Actions toolchain well-documented; versions verified via direct fetch and official sources)

---

## Scope Note

This document covers ONLY the new stack additions for the v2.2.0 milestone. The existing validated stack (Go 1.23, Wails v2.10.2, React 18, TypeScript 5, Ant Design 5, `wails build` with ldflags) is not re-researched. Focus: GitHub Actions release automation, installer packaging (DMG, NSIS, AppImage, deb), Homebrew tap automation, WinGet manifest generation.

---

## Recommended Stack — New Additions Only

### Core GitHub Actions (Required Version Upgrades)

| Technology | Version | Purpose | Why Recommended |
|------------|---------|---------|-----------------|
| `actions/checkout` | v4 | Repo checkout in all jobs | v3 deprecated Jan 2025; v4 is current standard — already used in existing build.yml |
| `actions/setup-go` | v5 | Go toolchain + module cache | v5 has default module caching built-in; no separate `actions/cache` step needed; already used in build.yml |
| `actions/setup-node` | v4 | Node.js for Wails frontend build | Matches existing build.yml; v4 is current |
| `actions/upload-artifact` | v4 | Upload build artifacts between jobs | v3 deprecated Jan 2025, stopped working; v4 required; up to 98% faster upload |
| `actions/download-artifact` | v4 | Download artifacts in the publish job | Must match upload-artifact version — v3 and v4 artifacts are not cross-compatible |
| `softprops/action-gh-release` | v2 | Create GitHub release and upload all assets | Industry standard; v2.5.0 (Dec 2024); supports draft-until-complete pattern; `GITHUB_TOKEN` is sufficient for asset upload |

### macOS DMG Packaging

| Technology | Version | Purpose | Why Recommended |
|------------|---------|---------|-----------------|
| `create-dmg` (shell script) | v1.2.3 (Nov 2025) | Wrap `.app` bundle in a DMG with drag-to-Applications UI | `brew install create-dmg` on macOS runner; pure shell, no Node.js dependency; handles volume name, window layout, icon positioning, Applications symlink; v1.2.3 is the latest stable as of Nov 2025; actively maintained |

**Install in CI:** `brew install create-dmg` on `macos-latest` runner. Takes ~10 seconds.

**Key command pattern:**
```bash
create-dmg \
  --volname "StorCat" \
  --app-drop-link 400 200 \
  "StorCat-${VERSION}-macOS-universal.dmg" \
  "build/bin/StorCat.app"
```

### Windows NSIS Installer

| Technology | Version | Purpose | Why Recommended |
|------------|---------|---------|-----------------|
| NSIS via `wails build -nsis` | Bundled with Wails v2 | Generate a Windows `.exe` installer | Wails has native NSIS support; produces `build/bin/StorCat-amd64-installer.exe` alongside the bare `.exe`; no external toolchain to install; config lives in `build/windows/installer/` |

**Key command:** `wails build -clean -platform windows/amd64 -nsis`

**Critical runner requirement:** Must run on `windows-latest`. The existing `build.yml` runs Windows builds on `macos-latest` (cross-compile) — NSIS installer generation requires a Windows environment. This is a required fix for the release workflow.

### Linux AppImage Packaging

| Technology | Version | Purpose | Why Recommended |
|------------|---------|---------|-----------------|
| `linuxdeploy` + `linuxdeploy-plugin-appimage` | `continuous` (rolling release) | Bundle binary + libs into AppImage | Standard AppImage toolchain; `--output appimage` flag on `linuxdeploy` automatically invokes the plugin; CI-friendly (reads environment variables from GitHub Actions); simpler than `appimage-builder` for a Go binary with no heavy framework deps |

**Pattern on ubuntu runner:**
```bash
wget -q "https://github.com/linuxdeploy/linuxdeploy/releases/download/continuous/linuxdeploy-x86_64.AppImage"
wget -q "https://github.com/linuxdeploy/linuxdeploy-plugin-appimage/releases/download/continuous/linuxdeploy-plugin-appimage-x86_64.AppImage"
chmod +x linuxdeploy*.AppImage

./linuxdeploy-x86_64.AppImage \
  --appdir AppDir \
  --executable build/bin/StorCat \
  --desktop-file packaging/linux/storcat.desktop \
  --icon-file build/appicon.png \
  --output appimage
```

**Requires:** A `storcat.desktop` file and an app icon. Both must be created as part of the `packaging/linux/` directory structure.

### Linux deb Packaging

| Technology | Version | Purpose | Why Recommended |
|------------|---------|---------|-----------------|
| `nfpm` (goreleaser/nfpm) | v2.x (latest) | Generate `.deb` (and optionally `.rpm`) from a single declarative YAML config | Pure Go, zero system dependencies (no `dpkg-buildpackage` setup, no Debian packaging toolchain); supports deb/rpm/apk; configured via `nfpm.yaml`; GitHub Action available; actively maintained by the GoReleaser team |

**Install in CI:**
```bash
curl -sLO "https://github.com/goreleaser/nfpm/releases/latest/download/nfpm_amd64.deb"
sudo dpkg -i nfpm_amd64.deb
```

**Usage:**
```bash
VERSION=${GITHUB_REF_NAME#v} nfpm package --config packaging/linux/nfpm.yaml --packager deb
```

### Homebrew Tap Automation

| Technology | Version | Purpose | Why Recommended |
|------------|---------|---------|-----------------|
| `gh` CLI (pre-installed on GitHub runners) + shell script | Built-in on all runners | Push updated cask to `homebrew-storcat` on release | The existing `update-tap.sh` already handles SHA256 calculation and cask template substitution. Wire it via a PAT secret. No third-party action needed — the script logic is already written and validated. |

**Why not `mislav/bump-homebrew-formula-action@v4.1`:** That action explicitly documents "limited support for Homebrew casks." StorCat distributes a pre-built binary DMG (a cask, not a formula). The action is designed for source-build formulas. Custom script is simpler and already exists.

**Pattern:**
```yaml
- name: Update Homebrew tap
  env:
    HOMEBREW_TAP_TOKEN: ${{ secrets.HOMEBREW_TAP_TOKEN }}
  run: ./packaging/update-tap.sh "${{ github.ref_name }}"
```

The `update-tap.sh` script pushes a commit to `scottkw/homebrew-storcat` using the PAT for authentication.

### WinGet Manifest Automation

| Technology | Version | Purpose | Why Recommended |
|------------|---------|---------|-----------------|
| `vedantmgoyal9/winget-releaser` | v2 | Submit updated manifests to `microsoft/winget-pkgs` via PR | Uses Komac under the hood; runs on Linux/macOS runners (not Windows-only); v2 is current; only requires `identifier` (e.g., `scottkw.StorCat`) and a PAT; 284 stars, 13+ contributors, actively maintained; community standard for WinGet automation |

**Usage:**
```yaml
- uses: vedantmgoyal9/winget-releaser@v2
  with:
    identifier: scottkw.StorCat
    token: ${{ secrets.WINGET_TOKEN }}
```

The action detects version and installer URLs from the GitHub release automatically.

---

## Development Tools — No Changes

| Tool | Notes |
|------|-------|
| `wails dev` | Unchanged. CI pipeline doesn't affect local development workflow. |
| `wails build` | Used as-is. Release workflow adds `-nsis` on Windows and wraps macOS output with `create-dmg`. |

---

## Secrets Required

| Secret Name | Value | Used By |
|-------------|-------|---------|
| `GITHUB_TOKEN` | Auto-provided by GitHub Actions | `softprops/action-gh-release` for asset upload |
| `HOMEBREW_TAP_TOKEN` | Classic PAT: `public_repo` + `workflow` scopes on `scottkw/homebrew-storcat` | `update-tap.sh` push to homebrew-storcat repo |
| `WINGET_TOKEN` | Classic PAT: `public_repo` scope on a fork of `microsoft/winget-pkgs` | `vedantmgoyal9/winget-releaser@v2` |

---

## Workflow Architecture

Two workflows total. The existing `build.yml` is updated; a new `release.yml` is added.

**`build.yml` (update existing)** — Triggers on push to main, PRs:
- Keep existing structure
- Fix: Move Windows job to `windows-latest` runner (currently incorrectly on `macos-latest`)
- No installer packaging — CI check only

**`release.yml` (new)** — Triggers on `release: published`:
```
Job: build-macos    (macos-latest)
  - wails build -platform darwin/universal
  - brew install create-dmg
  - create-dmg → StorCat-{version}-macOS-universal.dmg
  - upload-artifact: dmg + .app.zip

Job: build-windows  (windows-latest)       ← must be windows-latest
  - wails build -platform windows/amd64 -nsis
  - upload-artifact: .exe + installer.exe

Job: build-linux    (ubuntu-latest, matrix: amd64/arm64)
  - wails build -platform linux/{arch}
  - linuxdeploy → AppImage
  - nfpm → .deb
  - upload-artifact: binary + AppImage + .deb

Job: publish        (ubuntu-latest, needs: [build-macos, build-windows, build-linux])
  - download-artifact: all build outputs
  - softprops/action-gh-release@v2: upload all assets
  - update-tap.sh: push cask to homebrew-storcat
  - vedantmgoyal9/winget-releaser@v2: submit winget PR
```

---

## What NOT to Use

| Avoid | Why | Use Instead |
|-------|-----|-------------|
| GoReleaser | Cannot build Wails apps — GoReleaser's build hooks skip Darwin cross-compile, and the Wails toolchain requires its own `wails build` command. Tested Dec 2025, confirmed fails with bindings generation errors. | `wails build` per-platform on native runners |
| `dAppServer/wails-build-action` | Unmaintained community fork; README says "USE: host-uk/build@v4" with no clear owner; opaque wrapper hides what wails flags are used | Direct `wails build` commands — transparent and debuggable |
| `actions/upload-artifact@v3` | Deprecated Jan 2025; stopped working | `actions/upload-artifact@v4` |
| `actions/checkout@v3` | Deprecated | `actions/checkout@v4` |
| WiX Toolset (MSI) | Wails uses NSIS natively; WiX requires separate manifest authoring and .wxs XML files | `wails build -nsis` |
| `mislav/bump-homebrew-formula-action` for casks | Action documents "limited support for Homebrew casks"; designed for source-build formulas not pre-built binaries | Custom shell script (already exists as `update-tap.sh`) |
| `appimage-builder` (Python) | Heavier Python-based tool; requires a recipe YAML with apt packages specified; over-engineered for a simple Go binary | `linuxdeploy` + plugin |
| Code signing tools (gon, rcodesign, Authenticode) | Out of scope for this milestone per PROJECT.md — future milestone | N/A |
| Hardcoded PATs in workflow YAML | Security risk; secrets exposed in git history | `${{ secrets.HOMEBREW_TAP_TOKEN }}` etc. |

---

## Version Compatibility

| Package | Compatible With | Notes |
|---------|-----------------|-------|
| `actions/upload-artifact@v4` | `actions/download-artifact@v4` | Must use matching v4; v3 and v4 artifacts are not cross-compatible — mixing versions silently fails |
| `wails build -nsis` | `windows-latest` runner only | NSIS requires Windows; macOS cross-compile to Windows does not support `-nsis` flag |
| `create-dmg v1.2.3` | `macos-latest` (Homebrew available) | `brew install create-dmg` installs cleanly on macOS GitHub runners |
| Wails v2.10.2 | Go 1.23, Node 18+ | Existing validated combination; do not upgrade — v2.10.0 is known broken (search results Nov 2024); v3 is alpha |
| `nfpm` | ubuntu-latest | Pure Go binary, no system dependencies needed |
| `linuxdeploy` (continuous) | ubuntu-latest amd64 | `continuous` tag is the AppImage project's rolling release; use architecture-specific binary (`linuxdeploy-x86_64.AppImage` or `linuxdeploy-aarch64.AppImage`) |

---

## Alternatives Considered

| Recommended | Alternative | When to Use Alternative |
|-------------|-------------|-------------------------|
| `wails build -nsis` | WiX Toolset MSI | When enterprise deployment requires `.msi` format specifically (Group Policy, SCCM); consumer app NSIS `.exe` installer is sufficient |
| `create-dmg` shell tool | `sindresorhus/create-dmg` (Node.js) | Node.js version is also viable; shell version preferred — no Node.js dependency, directly installable via Homebrew |
| `vedantmgoyal9/winget-releaser@v2` | `microsoft/wingetcreate` CLI directly | wingetcreate gives more control over manifest YAML structure; releaser action is simpler for standard cases and handles URL/SHA256 detection automatically |
| Custom shell script for Homebrew | `mislav/bump-homebrew-formula-action@v4.1` | Use the action if distributing a Homebrew formula (source build); for casks (pre-built binary DMG), the action's cask support is limited per its own docs |
| `nfpm` for deb | `dpkg-deb` directly | `dpkg-deb` is fine but requires manual `DEBIAN/control` file construction; nfpm is declarative YAML and handles both deb and rpm from one config |
| `linuxdeploy` for AppImage | `appimage-builder` | `appimage-builder` handles complex apps with Python/Qt deps; for a simple Go binary with GTK from Wails, `linuxdeploy` is the right complexity level |
| Two separate workflows | Single workflow with all jobs | Splitting keeps CI fast (build.yml is quick feedback) and release clean (release.yml only runs on tags) |

---

## Sources

- [create-dmg GitHub](https://github.com/create-dmg/create-dmg) — v1.2.3 Nov 2025 (HIGH — fetched directly)
- [softprops/action-gh-release](https://github.com/softprops/action-gh-release) — v2.5.0, current standard (HIGH — multiple search sources)
- [winget-releaser action](https://github.com/marketplace/actions/winget-releaser) — v2 with Komac, Linux-compatible (HIGH — fetched directly from Marketplace)
- [mislav/bump-homebrew-formula-action](https://github.com/mislav/bump-homebrew-formula-action) — v4.1 Mar 2026, limited cask support documented (HIGH — fetched directly)
- [actions/upload-artifact deprecation](https://github.blog/changelog/2024-04-16-deprecation-notice-v3-of-the-artifact-actions/) — v3 deprecated Jan 2025 (HIGH — official GitHub changelog)
- [Wails NSIS docs](https://wails.io/docs/guides/windows-installer/) — `-nsis` flag confirmation (MEDIUM — search results; direct fetch returned 403)
- [nfpm goreleaser](https://github.com/goreleaser/nfpm) — Go-native deb/rpm packager, active 2025 (MEDIUM — search results)
- [linuxdeploy AppImage toolchain](https://docs.appimage.org/packaging-guide/from-source/linuxdeploy-user-guide.html) — standard AppImage pipeline (MEDIUM — search results)
- [GoReleaser + Wails cross-compile failure](https://chriswheeler.dev/posts/cross-compilation-with-wails/) — Dec 2025 confirmed incompatibility (MEDIUM — community blog, recent date)
- [storcat-repo-consolidation.md](../storcat-repo-consolidation.md) — existing scripts, Option A decision (HIGH — local file, already validated plan)
- Existing `.github/workflows/build.yml` — current workflow structure and runner choices (HIGH — codebase)
- `go.mod` — confirms Wails v2.10.2 in use (HIGH — codebase)

---
*Stack research for: StorCat v2.2.0 Repo Consolidation and CI/CD Release Pipeline*
*Researched: 2026-03-27*
