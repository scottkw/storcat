# Phase 14: Platform Packaging - Research

**Researched:** 2026-03-27
**Domain:** Cross-platform installer packaging (macOS DMG, Windows NSIS, Linux AppImage + .deb)
**Confidence:** HIGH for macOS/Windows/deb; MEDIUM for AppImage (WebKit bundling is a known hard problem)

---

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| PKG-01 | macOS DMG installer produced via `create-dmg` | sindresorhus/create-dmg npm package; --no-code-sign flag; brew install on macos-14 runner |
| PKG-02 | Windows NSIS installer produced via `wails build -nsis` | Wails v2 built-in NSIS support; NSIS 3.10 on windows-2022; must NOT use windows-latest |
| PKG-03 | Linux AppImage produced for x64 | linuxdeploy + linuxdeploy-plugin-appimage; significant WebKit caveats (see Pitfalls) |
| PKG-04 | Linux .deb package produced for x64 and arm64 | dpkg-deb; manual DEBIAN/control; one .deb per runner arch (no cross-compile needed) |
</phase_requirements>

---

## Summary

Phase 14 extends the existing four-runner release workflow to also produce installable packages alongside the raw binaries. Each package type has a well-understood primary tool, but two platform-specific pitfalls require immediate design decisions.

**Critical finding for PKG-02 (Windows NSIS):** `windows-latest` was migrated to Windows Server 2025 in September 2025. NSIS 3.10 is present on `windows-2022` but was removed from `windows-2025`. The Windows build job MUST change from `windows-latest` to `windows-2022`. If it stays on `windows-latest`, `wails build -nsis` will fail at the `makensis` call.

**Critical finding for PKG-03 (AppImage):** Bundling WebKit2GTK inside an AppImage is unreliable. WebKit subprocess binaries (`WebKitNetworkProcess`, `WebKitGPUProcess`, `WebKitWebProcess`) hardcode absolute system paths, causing launch failures on distributions with different WebKit layouts. The practical approach for a Wails v2 app on ubuntu-22.04 is to produce an AppImage that declares WebKit as a system dependency — the AppImage launches only on systems with `libwebkit2gtk-4.0-37` installed. This covers Ubuntu 20.04, 22.04, Debian 11/12, and Linux Mint 20/21. Ubuntu 24.04+ users will need the .deb package (which can target webkit2gtk 4.1) or a Flatpak (out of scope).

**Primary recommendation:** Extend each platform's build job with a packaging step after the binary build, rename the runner for Windows from `windows-latest` to `windows-2022`, and accept the system-WebKit limitation for AppImage.

---

## Standard Stack

### Core

| Tool | Version | Purpose | Why Standard |
|------|---------|---------|--------------|
| `create-dmg` (sindresorhus) | 8.x (npm) | macOS DMG with drag-to-Applications layout | npm package, Node 20 available on macos-14, actively maintained |
| `wails build -nsis` | v2.10.2 (project version) | Windows NSIS installer | Built into Wails CLI; uses pre-installed `makensis` on the runner |
| `linuxdeploy` | latest AppImage (x86_64) | Linux AppImage creation | Official AppImage toolchain; bundles shared libs |
| `linuxdeploy-plugin-appimage` | latest | Converts AppDir to AppImage | Required output plugin for linuxdeploy |
| `linuxdeploy-plugin-gtk` | latest | Bundles GTK resources, schemas | Required for GTK3-based apps (Wails uses GTK3 on Linux) |
| `dpkg-deb` | system (ubuntu-22.04) | Creates .deb binary package | Standard Debian packaging tool; pre-installed on all Ubuntu runners |

### Supporting

| Tool | Version | Purpose | When to Use |
|------|---------|---------|-------------|
| `brew install create-dmg` | — | Install create-dmg on macOS runner | Alternative to npm install; both work on macos-14 |
| `APPIMAGETOOL_ARCH` env var | — | Controls AppImage architecture | Set before linuxdeploy download to target arm64 |

### Alternatives Considered

| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| sindresorhus/create-dmg (npm) | create-dmg/create-dmg (shell script, brew) | Shell script requires more flags for drag layout; npm version simpler API |
| wails -nsis | Inno Setup (also on windows-2022) | Inno Setup has more customization but no Wails integration; NSIS is native Wails |
| linuxdeploy AppImage | appimage-builder | appimage-builder has Docker dependency; linuxdeploy is simpler for CI |
| manual dpkg-deb | nfpm, go-bin-deb | nfpm is cleaner but an extra dependency; dpkg-deb is pre-installed |

**Installation (macOS job):**
```bash
npm install --global create-dmg
```

**Installation (Linux amd64 job — AppImage):**
```bash
wget -O linuxdeploy-x86_64.AppImage https://github.com/linuxdeploy/linuxdeploy/releases/download/continuous/linuxdeploy-x86_64.AppImage
wget -O linuxdeploy-plugin-appimage-x86_64.AppImage https://github.com/linuxdeploy/linuxdeploy-plugin-appimage/releases/download/continuous/linuxdeploy-plugin-appimage-x86_64.AppImage
wget -O linuxdeploy-plugin-gtk.sh https://raw.githubusercontent.com/linuxdeploy/linuxdeploy-plugin-gtk/master/linuxdeploy-plugin-gtk.sh
chmod +x linuxdeploy-x86_64.AppImage linuxdeploy-plugin-appimage-x86_64.AppImage linuxdeploy-plugin-gtk.sh
sudo apt-get install -y libfuse2
```

**Version verification:**
```bash
npm view create-dmg version   # Confirmed 8.1.0 as of research date
```

---

## Architecture Patterns

### PKG-01: macOS DMG (macos-14 job)

After `wails build -clean -platform darwin/universal`:

```bash
# Install create-dmg
npm install --global create-dmg

# Create DMG (unsigned — code signing is out of scope)
cd build/bin
create-dmg \
  --overwrite \
  --dmg-title "StorCat" \
  --no-code-sign \
  "StorCat-${VERSION}-darwin-universal.dmg" \
  "StorCat.app"
```

Output file: `build/bin/StorCat-${VERSION}-darwin-universal.dmg`

The DMG automatically gets a drag-to-Applications layout. The `--no-code-sign` flag prevents exit-code failure when no Apple Developer certificate is present (which is the case in unsigned CI). The `--overwrite` flag is required when re-running the step.

Upload both the tarball AND the DMG as release assets from the same job. The artifact upload step needs to glob both:
```yaml
path: |
  build/bin/StorCat-*-darwin-universal.tar.gz
  build/bin/StorCat-*-darwin-universal.dmg
```

### PKG-02: Windows NSIS Installer (windows-2022 job)

**Runner change: `windows-latest` → `windows-2022`**

Wails generates NSIS template files at `build/windows/installer/` on the first `-nsis` build run. These are committed to the repo after first generation. On subsequent runs they are already present.

```yaml
- name: Build Windows with NSIS installer
  run: wails build -clean -platform windows/amd64 -windowsconsole -nsis
```

Wails internally calls `makensis` (pre-installed at `C:\Program Files (x86)\NSIS\`) on the generated `build/windows/installer/project.nsi` script. The `-nsis` and `-windowsconsole` flags are orthogonal and can be combined.

Output files in `build/bin/`:
- `StorCat-amd64-installer.exe` — NSIS installer (installs to Program Files, adds uninstaller, creates Start Menu entry)
- `StorCat.exe` — raw standalone binary (also present)

The installer filename follows the pattern `{outputfilename}-{arch}-installer.exe` from `wails.json`. For StorCat: `StorCat-amd64-installer.exe`.

The NSIS installer produced by Wails:
- Installs to `%PROGRAMFILES%\StorCat\` by default
- Registers an uninstaller in Add/Remove Programs
- Creates a Start Menu shortcut
- Does NOT create a desktop shortcut by default (configurable in the .nsi script)

Upload both the raw `.exe` and the installer `.exe` as release assets.

### PKG-03: Linux AppImage (build-linux-amd64 job only)

The AppImage is x64 only (ubuntu-22.04 runner). The arm64 runner produces only the .deb + tarball.

AppImage WebKit constraint: the AppImage will depend on the system's `libwebkit2gtk-4.0-37`. This is present on Ubuntu 20.04, 22.04, Debian 11/12, and Linux Mint 20/21 by default. It is **not** present on Ubuntu 24.04+. This is an explicit design decision — see Pitfalls.

Required assets to create before linuxdeploy:
1. `StorCat.desktop` — freedesktop desktop entry file
2. `appicon.png` — 256x256 icon (already exists at `build/appicon.png`)

**Desktop entry** (`build/linux/storcat.desktop`):
```ini
[Desktop Entry]
Type=Application
Name=StorCat
Comment=Storage Media Cataloging Tool
Exec=StorCat
Icon=appicon
Categories=Utility;FileManager;
Terminal=false
```

**Full AppImage build sequence:**
```bash
# Install tools
wget -q -O linuxdeploy https://github.com/linuxdeploy/linuxdeploy/releases/download/continuous/linuxdeploy-x86_64.AppImage
wget -q -O linuxdeploy-plugin-appimage https://github.com/linuxdeploy/linuxdeploy-plugin-appimage/releases/download/continuous/linuxdeploy-plugin-appimage-x86_64.AppImage
wget -q -O linuxdeploy-plugin-gtk.sh https://raw.githubusercontent.com/linuxdeploy/linuxdeploy-plugin-gtk/master/linuxdeploy-plugin-gtk.sh
chmod +x linuxdeploy linuxdeploy-plugin-appimage linuxdeploy-plugin-gtk.sh

# libfuse2 required to run AppImages during build on ubuntu-22.04
sudo apt-get install -y libfuse2

# Build the AppImage
export ARCH=x86_64
export OUTPUT="StorCat-${VERSION}-linux-x86_64.AppImage"
./linuxdeploy \
  --appdir AppDir \
  --executable build/bin/StorCat \
  --icon-file build/appicon.png \
  --desktop-file build/linux/storcat.desktop \
  --plugin gtk \
  --output appimage

# Rename to versioned name
mv StorCat-x86_64.AppImage "StorCat-${VERSION}-linux-x86_64.AppImage"
```

Output: `StorCat-${VERSION}-linux-x86_64.AppImage` in the workspace root (linuxdeploy writes AppImages to the current directory).

### PKG-04: Linux .deb Package (both linux jobs)

Each Linux runner (amd64 and arm64) builds its own .deb from the binary it just built. No cross-compilation needed — each runner builds natively for its arch.

**Directory structure:**
```
deb-pkg/
├── DEBIAN/
│   └── control
└── usr/
    ├── bin/
    │   └── StorCat
    └── share/
        ├── applications/
        │   └── storcat.desktop
        └── pixmaps/
            └── storcat.png
```

**DEBIAN/control file:**
```
Package: storcat
Version: 2.2.0
Architecture: amd64
Maintainer: Ken Scott <kenscott@gmail.com>
Description: Storage Media Cataloging Tool
 StorCat creates, browses, and searches directory catalogs.
 Generates JSON and HTML representations of directory trees.
Depends: libgtk-3-0, libwebkit2gtk-4.0-37
Section: utils
Priority: optional
```

For arm64, `Architecture: arm64` — determined at runtime from `dpkg --print-architecture`.

**Build script:**
```bash
ARCH=$(dpkg --print-architecture)
VERSION_CLEAN="${VERSION#v}"   # strip leading 'v' from tag
PKG_DIR="deb-pkg"

mkdir -p "$PKG_DIR/DEBIAN"
mkdir -p "$PKG_DIR/usr/bin"
mkdir -p "$PKG_DIR/usr/share/applications"
mkdir -p "$PKG_DIR/usr/share/pixmaps"

cp build/bin/StorCat "$PKG_DIR/usr/bin/StorCat"
cp build/linux/storcat.desktop "$PKG_DIR/usr/share/applications/storcat.desktop"
cp build/appicon.png "$PKG_DIR/usr/share/pixmaps/storcat.png"
chmod 755 "$PKG_DIR/usr/bin/StorCat"

cat > "$PKG_DIR/DEBIAN/control" << EOF
Package: storcat
Version: ${VERSION_CLEAN}
Architecture: ${ARCH}
Maintainer: Ken Scott <kenscott@gmail.com>
Description: Storage Media Cataloging Tool
 StorCat creates, browses, and searches directory catalogs.
 Generates JSON and HTML representations of directory trees.
Depends: libgtk-3-0, libwebkit2gtk-4.0-37
Section: utils
Priority: optional
EOF

dpkg-deb --root-owner-group --build "$PKG_DIR" \
  "build/bin/storcat_${VERSION_CLEAN}_${ARCH}.deb"
```

Output: `build/bin/storcat_2.2.0_amd64.deb` and `build/bin/storcat_2.2.0_arm64.deb`

### Anti-Patterns to Avoid

- **Using `windows-latest` for NSIS builds:** As of September 2025, `windows-latest` = Windows Server 2025 which does NOT have NSIS. Use `windows-2022`.
- **Trying to bundle WebKit into AppImage:** WebKit subprocess binaries use hardcoded absolute paths. Bundling them with linuxdeploy causes runtime failures on non-Debian systems. Accept the system dependency.
- **Cross-compiling .deb packages:** dpkg-deb can build for foreign arches but requires multiarch apt setup. The arm64 runner builds the arm64 .deb natively — no cross-compilation needed.
- **Using version tag directly in deb Version field:** Debian version strings cannot start with `v`. Strip the `v` prefix: `${VERSION#v}`.

---

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| macOS DMG layout | Custom `hdiutil` + AppleScript | `create-dmg` (npm) | Applications symlink, window sizing, Finder prettification handled automatically |
| Windows NSIS script | Custom `.nsi` from scratch | `wails build -nsis` | Wails generates and maintains the `.nsi` template; integrates with `wails.json` metadata |
| AppImage creation | Custom AppDir assembly | `linuxdeploy` | Correctly handles shared library resolution, rpath, GTK schema bundling |
| .deb file structure | Manual tar/ar packaging | `dpkg-deb --build` | Handles md5sums generation, control format validation, correct ar archive layout |

---

## Runtime State Inventory

Not applicable — this is a CI/CD packaging phase with no runtime state, databases, or stored data.

---

## Common Pitfalls

### Pitfall 1: NSIS missing on windows-latest (BLOCKING)
**What goes wrong:** `wails build -nsis` exits with "makensis: command not found" or similar error
**Why it happens:** `windows-latest` resolved to Windows Server 2025 after September 2025. NSIS was removed from the 2025 image.
**How to avoid:** Set `runs-on: windows-2022` on the Windows build job. NSIS 3.10 is pre-installed there.
**Warning signs:** Any build log mentioning `makensis not found` or `cannot build nsis installer`

### Pitfall 2: AppImage WebKit subprocess failure at runtime
**What goes wrong:** AppImage launches but immediately crashes with "Failed to spawn child process '/usr/lib/x86_64-linux-gnu/webkit2gtk-4.0/WebKitNetworkProcess'"
**Why it happens:** WebKit subprocess binaries reference absolute paths. Even when bundled, they look for libraries at the host system's webkit2gtk path, not the AppImage's bundled path.
**How to avoid:** Do NOT try to bundle WebKit. Let the AppImage depend on the system's `libwebkit2gtk-4.0-37`. Use linuxdeploy `--plugin gtk` only (which bundles GTK schemas/resources, not WebKit itself). Document in release notes that the AppImage requires `libwebkit2gtk-4.0-37` on the host.
**Warning signs:** AppImage starts but shows blank window, or any `WebKitNetworkProcess` error in stderr

### Pitfall 3: libfuse2 not installed on ubuntu-22.04 build runner
**What goes wrong:** linuxdeploy itself (an AppImage) cannot execute: "fuse: device not found, try 'modprobe fuse' first" or FUSE errors
**Why it happens:** Ubuntu 22.04 does not install `libfuse2` by default; AppImages require FUSE to mount
**How to avoid:** Add `sudo apt-get install -y libfuse2` to the Linux amd64 build job before running linuxdeploy
**Warning signs:** Any FUSE error when trying to run the downloaded linuxdeploy AppImage

### Pitfall 4: create-dmg fails with non-zero exit code when no signing identity
**What goes wrong:** `create-dmg` exits with error code even though DMG was created
**Why it happens:** create-dmg attempts to sign by default; fails if no Apple Developer certificate
**How to avoid:** Always pass `--no-code-sign` flag. Without this flag, CI exits with failure even though the file is present.
**Warning signs:** Step marked as failed but DMG file exists

### Pitfall 5: Debian version string starts with 'v'
**What goes wrong:** `dpkg-deb` rejects the package or installs with bad version string
**Why it happens:** Git tags use `v2.2.0` format; Debian version fields cannot start with a letter
**How to avoid:** Strip the `v` prefix: `VERSION_CLEAN="${VERSION#v}"` before writing the control file
**Warning signs:** `dpkg: error processing package` or version shows as "v2.2.0" in dpkg -l

### Pitfall 6: NSIS installer template files not in repo
**What goes wrong:** `wails build -nsis` succeeds locally but `build/windows/installer/` doesn't exist in CI
**Why it happens:** Wails auto-generates `build/windows/installer/` on first `-nsis` run; if those files are gitignored or not committed, CI doesn't have them
**How to avoid:** Run `wails build -nsis` once locally, commit the generated `build/windows/installer/` directory. Verify it is not in `.gitignore`.
**Warning signs:** Empty `build/windows/installer/` in git, or NSIS "script not found" errors

### Pitfall 7: linuxdeploy writes AppImage to current directory, not build/bin
**What goes wrong:** Upload step can't find the AppImage
**Why it happens:** linuxdeploy outputs the AppImage to the current working directory, named `{AppName}-{ARCH}.AppImage`
**How to avoid:** After linuxdeploy completes, rename/move the output: `mv StorCat-x86_64.AppImage build/bin/StorCat-${VERSION}-linux-x86_64.AppImage`
**Warning signs:** Upload artifact step finds no files matching the expected pattern

---

## Code Examples

### PKG-01: macOS DMG step (GitHub Actions YAML)
```yaml
- name: Install create-dmg
  run: npm install --global create-dmg

- name: Create macOS DMG
  run: |
    cd build/bin
    create-dmg \
      --overwrite \
      --dmg-title "StorCat" \
      --no-code-sign \
      "StorCat-${{ steps.version.outputs.VERSION }}-darwin-universal.dmg" \
      "StorCat.app"
```

### PKG-02: Windows NSIS installer step (GitHub Actions YAML)
```yaml
# NOTE: runner must be windows-2022, not windows-latest
- name: Build Windows with NSIS installer
  run: wails build -clean -platform windows/amd64 -windowsconsole -nsis
```

### PKG-03: Linux AppImage step (GitHub Actions YAML)
```yaml
- name: Install AppImage tools
  run: |
    sudo apt-get install -y libfuse2
    wget -q -O linuxdeploy https://github.com/linuxdeploy/linuxdeploy/releases/download/continuous/linuxdeploy-x86_64.AppImage
    wget -q -O linuxdeploy-plugin-appimage https://github.com/linuxdeploy/linuxdeploy-plugin-appimage/releases/download/continuous/linuxdeploy-plugin-appimage-x86_64.AppImage
    wget -q -O linuxdeploy-plugin-gtk.sh https://raw.githubusercontent.com/linuxdeploy/linuxdeploy-plugin-gtk/master/linuxdeploy-plugin-gtk.sh
    chmod +x linuxdeploy linuxdeploy-plugin-appimage linuxdeploy-plugin-gtk.sh

- name: Create AppImage
  run: |
    export ARCH=x86_64
    ./linuxdeploy \
      --appdir AppDir \
      --executable build/bin/StorCat \
      --icon-file build/appicon.png \
      --desktop-file build/linux/storcat.desktop \
      --plugin gtk \
      --output appimage
    mv StorCat-x86_64.AppImage \
      build/bin/StorCat-${{ steps.version.outputs.VERSION }}-linux-x86_64.AppImage
```

### PKG-04: Linux .deb build script (inline GitHub Actions YAML)
```yaml
- name: Create .deb package
  run: |
    ARCH=$(dpkg --print-architecture)
    VERSION_CLEAN="${{ steps.version.outputs.VERSION }}"
    VERSION_CLEAN="${VERSION_CLEAN#v}"
    PKG_DIR="deb-pkg"
    mkdir -p "$PKG_DIR/DEBIAN" "$PKG_DIR/usr/bin" \
             "$PKG_DIR/usr/share/applications" "$PKG_DIR/usr/share/pixmaps"
    cp build/bin/StorCat "$PKG_DIR/usr/bin/StorCat"
    cp build/linux/storcat.desktop "$PKG_DIR/usr/share/applications/storcat.desktop"
    cp build/appicon.png "$PKG_DIR/usr/share/pixmaps/storcat.png"
    chmod 755 "$PKG_DIR/usr/bin/StorCat"
    printf "Package: storcat\nVersion: %s\nArchitecture: %s\nMaintainer: Ken Scott <kenscott@gmail.com>\nDescription: Storage Media Cataloging Tool\n StorCat creates, browses, and searches directory catalogs.\nDepends: libgtk-3-0, libwebkit2gtk-4.0-37\nSection: utils\nPriority: optional\n" \
      "$VERSION_CLEAN" "$ARCH" > "$PKG_DIR/DEBIAN/control"
    dpkg-deb --root-owner-group --build "$PKG_DIR" \
      "build/bin/storcat_${VERSION_CLEAN}_${ARCH}.deb"
```

### Desktop entry file (to create at `build/linux/storcat.desktop`)
```ini
[Desktop Entry]
Type=Application
Name=StorCat
Comment=Storage Media Cataloging Tool
Exec=StorCat
Icon=appicon
Categories=Utility;FileManager;
Terminal=false
```

---

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| `windows-latest` had NSIS | NSIS removed from windows-2025 image | Sept 2025 | Must pin to `windows-2022` for NSIS builds |
| webkit2gtk-4.0 universal on Ubuntu | Ubuntu 24.04+ ships webkit2gtk-4.1 only | Ubuntu 24.04 (Apr 2024) | AppImage + deb targeting 4.0 won't install on Ubuntu 24.04+ |

**Deprecated/outdated:**
- Using `windows-latest` for NSIS: No longer valid as of Sept 2025
- `libfuse` package on Ubuntu 22.04: Package is `libfuse2`, NOT `fuse` or `libfuse`

---

## Open Questions

1. **Does `build/windows/installer/` exist in the repo?**
   - What we know: Wails auto-generates this directory on first `-nsis` build; the directory is not present in the current repo
   - What's unclear: Whether the gitignore excludes it, and whether the generated `.nsi` script is correct out-of-the-box for the wails.json metadata
   - Recommendation: Run `wails build -nsis` locally once, inspect the generated `project.nsi`, commit the `build/windows/installer/` directory; plan should include this step

2. **AppImage compatibility goal: Ubuntu 22.04 only, or also 20.04?**
   - What we know: Building on ubuntu-22.04 means glibc 2.35; Ubuntu 20.04 has glibc 2.31 (incompatible)
   - What's unclear: Whether the release notes should state "Ubuntu 22.04+" as minimum or if 20.04 is a target
   - Recommendation: Accept ubuntu-22.04 as minimum for AppImage; document this clearly in release notes

3. **Should the .deb Depends target webkit2gtk-4.0 or 4.1, or both?**
   - What we know: Built on ubuntu-22.04 which has webkit2gtk-4.0; the binary links against 4.0; 4.1 is a parallel package
   - What's unclear: Whether `Depends: libwebkit2gtk-4.0-37 | libwebkit2gtk-4.1-0` would work correctly
   - Recommendation: Start with `libwebkit2gtk-4.0-37` only (matches the build system); a separate arm64 .deb targeting ubuntu-22.04-arm will have the same library available

---

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| `brew` | macOS DMG (create-dmg install) | On macos-14 runner | pre-installed | `npm install -g create-dmg` (also works) |
| Node.js 20 | create-dmg npm install | On macos-14 runner (workflow already sets up Node 20) | 20.x | — |
| NSIS (`makensis`) | Windows NSIS build | On `windows-2022` runner | 3.10 | NOT available on `windows-latest`/2025 |
| `dpkg-deb` | .deb build | On all ubuntu runners | system | — |
| `libfuse2` | linuxdeploy (AppImage) execution | NOT installed by default on ubuntu-22.04 | — | `sudo apt-get install -y libfuse2` |
| `linuxdeploy` | AppImage creation | Downloaded at build time | continuous | — |
| `wget` | Download linuxdeploy | Pre-installed on ubuntu runners | system | `curl -Lo` equivalent |

**Missing dependencies with no fallback:**
- NSIS on `windows-latest` (Windows Server 2025) — must use `windows-2022` runner

**Missing dependencies with fallback:**
- `libfuse2` on ubuntu-22.04 — install via apt in the build step

---

## Validation Architecture

### Test Framework

| Property | Value |
|----------|-------|
| Framework | None detected in project |
| Config file | N/A |
| Quick run command | N/A |
| Full suite command | N/A |

This is a CI/CD-only phase — all validation is smoke-test and artifact-inspection based, not automated unit tests.

### Phase Requirements → Test Map

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| PKG-01 | macOS DMG present and mountable | smoke | `test -f build/bin/StorCat-*-darwin-universal.dmg` | N/A (CI only) |
| PKG-02 | Windows NSIS installer present | smoke | `Test-Path build\bin\StorCat-amd64-installer.exe` (PowerShell) | N/A (CI only) |
| PKG-03 | Linux AppImage present and executable | smoke | `test -x StorCat-*-linux-x86_64.AppImage` | N/A (CI only) |
| PKG-04 | .deb packages present for both arches | smoke | `test -f build/bin/storcat_*_amd64.deb && test -f build/bin/storcat_*_arm64.deb` | N/A (CI only) |

### Wave 0 Gaps

- [ ] `build/linux/storcat.desktop` — required by linuxdeploy and .deb packaging; does not exist yet
- [ ] `build/windows/installer/` — required by `wails build -nsis`; run locally once and commit
- [ ] Windows job runner pin: `windows-latest` → `windows-2022` in release.yml

*(No test framework gaps — this phase has no unit tests; verification is CI artifact inspection)*

---

## Sources

### Primary (HIGH confidence)
- [sindresorhus/create-dmg README](https://github.com/sindresorhus/create-dmg/blob/main/readme.md) — `--no-code-sign` flag, `--overwrite`, basic usage
- [create-dmg/create-dmg](https://github.com/create-dmg/create-dmg) — `--app-drop-link`, `--sandbox-safe`, full flag list
- [wailsapp/wails NSIS installer source](https://github.com/wailsapp/wails/blob/master/v2/pkg/commands/build/nsis_installer.go) — confirms `build/windows/installer/project.nsi`, output naming
- [GitHub runner-images Windows2022-Readme](https://github.com/actions/runner-images/blob/main/images/windows/Windows2022-Readme.md) — NSIS 3.10 confirmed present
- [GitHub runner-images Windows2025-Readme](https://github.com/actions/runner-images/blob/main/images/windows/Windows2025-Readme.md) — NSIS confirmed absent
- [GitHub runner-images issue #12677](https://github.com/actions/runner-images/issues/12677) — windows-latest migration to Server 2025 completed Sept 2025
- [linuxdeploy user guide](https://docs.appimage.org/packaging-guide/from-source/linuxdeploy-user-guide.html) — `--executable`, `--plugin gtk`, `--output appimage`
- [Wails issue #3141](https://github.com/wailsapp/wails/issues/3141) — `libwebkit2gtk-4.0-37` confirmed correct runtime dep for .deb

### Secondary (MEDIUM confidence)
- [Wails issue #4313](https://github.com/wailsapp/wails/issues/4313) — WebKit subprocess hardcoded paths (confirmed bundling does not work reliably)
- [Tauri AppImage distribution docs](https://v2.tauri.app/distribute/appimage/) — linuxdeploy confirmed as standard AppImage toolchain; WebKit as system dependency
- [internalpointers.com .deb guide](https://www.internalpointers.com/post/build-binary-deb-package-practical-guide) — DEBIAN/control file structure, `dpkg-deb --root-owner-group`
- [npm create-dmg package](https://www.npmjs.com/package/create-dmg) — version 8.1.0 confirmed, Node 20+ required

### Tertiary (LOW confidence — needs CI validation)
- NSIS installer output filename pattern `StorCat-amd64-installer.exe` — inferred from Wails source; must be confirmed on first CI run
- linuxdeploy output filename `StorCat-x86_64.AppImage` — inferred from linuxdeploy behavior; must be confirmed on first CI run

---

## Project Constraints (from CLAUDE.md)

The project has no CONTEXT.md for this phase. Constraints from CLAUDE.md:

- **No Electron dependencies**: All solutions must use Go/Wails patterns
- **Branch strategy**: `none` (direct main branch commits per config.json)
- **pnpm preferred for Node**: Node 20 already configured in release.yml; `npm install -g create-dmg` is acceptable for a one-off global tool install on the runner
- **GitHub Actions all SHA-pinned**: Any new third-party actions must be SHA-pinned (established convention from Phase 13); the linuxdeploy downloads use direct URLs (not GitHub Actions) so SHA pinning does not apply, but checksum verification should be considered
- **Out of scope (from REQUIREMENTS.md)**: Code signing, GoReleaser, snap/rpm

---

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — tools verified against official docs and runner image manifests
- Architecture/commands: HIGH for macOS/Windows/deb; MEDIUM for AppImage (WebKit constraint well-documented but runtime behavior needs CI confirmation)
- Pitfalls: HIGH — NSIS/windows-latest issue confirmed via official GitHub issue; WebKit bundling issue confirmed via multiple wails+tauri issues
- AppImage WebKit claim: MEDIUM — confirmed impossible to fully bundle, but system-dependency approach not CI-tested for this specific project

**Research date:** 2026-03-27
**Valid until:** 2026-06-27 (90 days — tooling is stable; re-verify if ubuntu-22.04-arm runner changes or NSIS added back to windows-latest)
