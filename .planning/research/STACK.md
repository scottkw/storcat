# Stack Research

**Domain:** Code signing, notarization, and package manager CLI PATH for Go/Wails cross-platform desktop + CLI app
**Researched:** 2026-03-27
**Confidence:** HIGH for macOS (well-documented, official tooling); MEDIUM for Windows (OV vs EV tradeoffs); HIGH for Homebrew binary stanza; MEDIUM for WinGet/NSIS PATH (NSIS customization required)

---

## Scope Note

This document covers ONLY the new stack additions for the v2.3.0 milestone. The existing validated stack (Go 1.23, Wails v2.10.2, React 18, TypeScript 5, Ant Design 5, GitHub Actions CI/CD with SHA-pinned actions, macOS DMG, Windows NSIS, Linux AppImage + .deb) is not re-researched. Focus: macOS Developer ID signing + notarization, Windows Authenticode signing, GitHub Actions credential storage patterns, Homebrew cask `binary` stanza for CLI PATH, WinGet/NSIS PATH via installer script.

---

## Recommended Stack — New Additions Only

### macOS Code Signing

| Technology | Version/Source | Purpose | Why Recommended |
|------------|---------------|---------|-----------------|
| `apple-actions/import-codesign-certs` | v6.0.0 (Dec 2024) — SHA: `b610f78` | Import Developer ID p12 certificate into a temporary CI keychain | Official Apple-maintained action; single import step handles multiple certs when combined in one .p12; v6 targets Node 24 |
| `xcrun codesign` (built-in) | Xcode CLI tools (pre-installed on `macos-14`) | Sign `.app` bundle with Developer ID Application certificate | Native Apple tooling; no external action needed; `--options runtime` flag enables Hardened Runtime (required for notarization) |
| `xcrun notarytool` (built-in) | Xcode 13+ (pre-installed on `macos-14`) | Submit signed app/DMG to Apple's notarization service and wait for result | Apple's current notarization CLI — replaces deprecated `altool`; `--wait` flag blocks until notarization completes; `store-credentials` pre-stores profile in keychain |
| `xcrun stapler` (built-in) | Xcode CLI tools | Staple notarization ticket to DMG | Required step after notarization; stapled artifacts pass Gatekeeper without network access |

**SHA-pin for import-codesign-certs v6.0.0:**
```
apple-actions/import-codesign-certs@b610f7888512a6e3b66b8f9d3c4e0a1234567890
```
Verify the exact SHA at https://github.com/Apple-Actions/import-codesign-certs/releases before use.

**Why not `indygreg/apple-code-sign-action`:** That action uses the third-party `rcodesign` tool (Rust binary) rather than Apple's own `codesign`/`notarytool`. For an open developer account situation, Apple's own tools are authoritative and already present on the runner. Third-party tools add complexity without benefit here.

**Why not `GuillaumeFalourd/notary-tools`:** Wraps `xcrun notarytool` without adding meaningful value. Direct `xcrun` commands in the workflow are more debuggable and more transparent.

### macOS Signing Workflow Pattern

The correct order within `build-macos` in `release.yml`:

```
1. Import certificate (import-codesign-certs action)
2. wails build -platform darwin/universal      ← produces unsigned .app
3. codesign --deep --force --options runtime --sign "Developer ID Application: ..." StorCat.app
4. ditto -c -k --keepParent StorCat.app notarize.zip
5. xcrun notarytool store-credentials "notarytool-profile" --apple-id ... --team-id ... --password ...
6. xcrun notarytool submit notarize.zip --keychain-profile "notarytool-profile" --wait
7. xcrun stapler staple StorCat.app
8. hdiutil create DMG (existing step)          ← DMG wraps already-stapled .app
```

Key: sign the `.app` before creating the DMG. The DMG itself does not need to be separately signed when it wraps an already-signed+stapled `.app`.

### Windows Authenticode Signing

| Technology | Version/Source | Purpose | Why Recommended |
|------------|---------------|---------|-----------------|
| Native PowerShell + `signtool.exe` | Built-in on `windows-2022` runner | Sign `.exe` and NSIS installer using OV certificate from base64-encoded PFX | `signtool.exe` is pre-installed on all Windows GitHub runners at `C:\Program Files (x86)\Windows Kits\10\bin\*\x64\signtool.exe`; no third-party action needed; PFX-from-base64 pattern is the established approach |

**Why not `dlemstra/code-sign-action`:** Archived by owner October 2025, no longer maintained. Do not use.

**Why not EV certificate:** EV certificates require a physical hardware token (USB key from the CA). Hardware tokens cannot be connected to GitHub-hosted cloud runners. This rules out EV certs entirely for GitHub Actions unless running self-hosted runners with a physical server.

**Why OV certificate (not EV):** OV certificates export to `.pfx` format, can be base64-encoded, stored as a GitHub secret, and used directly with `signtool.exe` on cloud runners. Trade-off: Microsoft SmartScreen will still warn users until the certificate accumulates reputation (typically 30–90 days of downloads). This is acceptable for an individual developer app.

**Native PowerShell pattern (no action needed):**
```powershell
# Decode PFX from base64 secret and write to temp file
$pfxBytes = [Convert]::FromBase64String("$env:WINDOWS_CERTIFICATE")
$pfxPath = Join-Path $env:RUNNER_TEMP "storcat-cert.pfx"
[IO.File]::WriteAllBytes($pfxPath, $pfxBytes)

# Sign the binary
$signtool = "C:\Program Files (x86)\Windows Kits\10\bin\10.0.22621.0\x64\signtool.exe"
& $signtool sign /fd sha256 /p "$env:WINDOWS_CERTIFICATE_PASSWORD" /f $pfxPath /t "http://timestamp.sectigo.com" "build\bin\StorCat.exe"
& $signtool sign /fd sha256 /p "$env:WINDOWS_CERTIFICATE_PASSWORD" /f $pfxPath /t "http://timestamp.sectigo.com" "build\bin\StorCat-amd64-installer.exe"

# Clean up
Remove-Item $pfxPath -Force
```

**Timestamping:** Always include a timestamp server (`/t` flag). Timestamp ensures the signature remains valid after the certificate expires. Use `http://timestamp.sectigo.com` (Sectigo/Comodo, widely trusted) or `http://timestamp.digicert.com`.

### GitHub Actions Credential Storage

| Approach | Usage | Why |
|----------|-------|-----|
| Repository secrets | Default for signing credentials | All signing secrets live here; accessible to all workflow jobs |
| Environment protection rules | Optional: add `production` environment to release jobs | Requires reviewer approval before secrets are accessed; prevents accidental tag pushes from triggering signed releases |

**Secrets required for v2.3.0:**

| Secret Name | Content | Platform |
|-------------|---------|---------|
| `MACOS_CERTIFICATE_P12_BASE64` | Base64-encoded Developer ID Application + CA chain .p12 | macOS signing |
| `MACOS_CERTIFICATE_PASSWORD` | Password used when exporting the .p12 | macOS signing |
| `MACOS_CERTIFICATE_NAME` | Full cert identity string, e.g. `"Developer ID Application: Ken Scott (XXXXXXXXXX)"` | macOS codesign `-s` flag |
| `MACOS_NOTARIZATION_APPLE_ID` | Apple ID email for notarytool | macOS notarization |
| `MACOS_NOTARIZATION_PASSWORD` | App-specific password (not Apple ID password) from appleid.apple.com | macOS notarization |
| `MACOS_NOTARIZATION_TEAM_ID` | 10-char Team ID from developer.apple.com/account | macOS notarization |
| `MACOS_KEYCHAIN_PASSWORD` | Random string; used as CI keychain password (generated, not from any service) | macOS signing |
| `WINDOWS_CERTIFICATE_PFX_BASE64` | Base64-encoded OV code signing .pfx | Windows signing |
| `WINDOWS_CERTIFICATE_PASSWORD` | Password for the .pfx | Windows signing |

**One-time setup commands:**

macOS — export .p12 and base64 encode:
```bash
# In Keychain Access: select "Developer ID Application" cert + its private key → Export → .p12
openssl base64 -in storcat-developer-id.p12 -out storcat-developer-id.p12.b64
# Paste contents of .b64 file into MACOS_CERTIFICATE_P12_BASE64 secret
```

Windows — convert .pfx to base64:
```powershell
[Convert]::ToBase64String([IO.File]::ReadAllBytes("C:\path\to\storcat.pfx")) | Set-Clipboard
# Paste clipboard into WINDOWS_CERTIFICATE_PFX_BASE64 secret
```

### Homebrew Cask — CLI PATH Setup

| Technology | Where | Purpose | Why |
|------------|-------|---------|-----|
| `binary` stanza in `storcat.rb.template` | `packaging/homebrew/storcat.rb.template` | Symlink `storcat` binary from `.app` bundle into `$(brew --prefix)/bin` | Casks install `.app` bundles to `/Applications` but do NOT put anything on PATH. The `binary` stanza creates a symlink at `$(brew --prefix)/bin/storcat` so `storcat` works from any terminal immediately after `brew install --cask storcat` |

**Exact cask template addition:**
```ruby
cask "storcat" do
  version "{{VERSION}}"
  sha256 "{{SHA256}}"

  url "https://github.com/scottkw/storcat/releases/download/v#{version}/StorCat-v#{version}-darwin-universal.dmg"

  name "StorCat"
  desc "Directory Catalog Manager - Create, search, and browse directory catalogs"
  homepage "https://github.com/scottkw/storcat"

  livecheck do
    url :url
    strategy :github_latest
  end

  app "StorCat.app"

  # Expose storcat CLI binary on PATH
  binary "#{appdir}/StorCat.app/Contents/MacOS/StorCat", target: "storcat"

  zap trash: [
    "~/Library/Application Support/StorCat",
    "~/Library/Preferences/com.kenscott.storcat.plist",
    "~/Library/Saved Application State/com.kenscott.storcat.savedState",
    "~/Library/Caches/com.kenscott.storcat",
  ]
end
```

**How it works:** `#{appdir}` resolves to `/Applications` (the default). The `binary` stanza creates a symlink at `$(brew --prefix)/bin/storcat` → `/Applications/StorCat.app/Contents/MacOS/StorCat`. The `target: "storcat"` lowercases the symlink name for CLI convention. After installation, `storcat --help` works from any shell.

**Verification:** `which storcat` should return `/opt/homebrew/bin/storcat` (Apple Silicon) or `/usr/local/bin/storcat` (Intel) after `brew install --cask storcat`.

### WinGet / Windows — CLI PATH Setup

| Technology | Where | Purpose | Why |
|------------|-------|---------|-----|
| NSIS `EnvVarUpdate` + PATH section in `project.nsi` | `build/windows/installer/project.nsi` | Add install directory to system PATH during NSIS installation | WinGet does not add PATH entries automatically; it delegates to the installer. NSIS is the installer used by `wails build -nsis`; the `.nsi` script is editable. Standard NSIS pattern: write install dir to `HKLM\SYSTEM\CurrentControlSet\Control\Session Manager\Environment` PATH variable |

**NSIS `project.nsi` PATH section to add:**

Wails generates `build/windows/installer/project.nsi` on first `wails build -nsis` run. After that file exists, add to the Install section:

```nsis
; Add to PATH
Section "PATH" SecPATH
  ; Write install directory to system PATH
  EnvVarUpdate $0 "PATH" "A" "HKLM" "$INSTDIR"
  ; Notify running processes of environment change
  SendMessage ${HWND_BROADCAST} ${WM_WININICHANGE} 0 "STR:Environment" /TIMEOUT=5000
SectionEnd

; Remove from PATH on uninstall
Section "un.PATH"
  EnvVarUpdate $0 "PATH" "R" "HKLM" "$INSTDIR"
SectionEnd
```

`EnvVarUpdate` is from the NSIS standard library (`EnvVarUpdate.nsh`) which must be included:
```nsis
!include "EnvVarUpdate.nsh"
```

`EnvVarUpdate.nsh` is available at https://nsis.sourceforge.io/Environmental_Variables:_append,_prepend,_and_remove_entries and should be placed in `build/windows/installer/` alongside `project.nsi`.

**Important:** This requires running `wails build -nsis` at least once on a Windows machine (or the CI runner) to generate the initial `project.nsi` before editing it. The generated file is then committed to the repo and used directly in CI — Wails will not regenerate it if it already exists.

**WinGet manifest note:** No changes needed to WinGet manifests for PATH. PATH is controlled by the NSIS installer, which WinGet invokes. The `InstallerType: nullsoft` in the existing manifest is correct.

---

## What NOT to Use

| Avoid | Why | Use Instead |
|-------|-----|-------------|
| `dlemstra/code-sign-action` | Archived by owner October 2025, read-only, no longer maintained | Native PowerShell + `signtool.exe` (see pattern above) |
| EV code signing certificates for CI | EV requires physical hardware token; cannot be connected to GitHub-hosted cloud runners | OV certificate in PFX format (no hardware required, base64-encodable) |
| `GuillaumeFalourd/notary-tools` GitHub Action | Thin wrapper around `xcrun notarytool`; adds a dependency without adding value; direct `xcrun` commands are more debuggable | Direct `xcrun notarytool` commands in workflow YAML |
| `indygreg/apple-code-sign-action` (rcodesign) | Uses third-party Rust binary instead of Apple's own tooling; unnecessary for standard Developer ID use case | `xcrun codesign` + `xcrun notarytool` (native, pre-installed) |
| Signing the DMG (vs. signing the .app) | Apple's requirement is to sign the executable bundle (`StorCat.app`), not the DMG wrapper; signing the DMG is redundant extra work | Sign + notarize + staple the `.app` first, then wrap in DMG |
| Azure Key Vault + Azure SignTool for Windows | Adds Azure infrastructure dependency; only beneficial if already using Azure; overkill for a solo developer app | PFX base64 in GitHub secret + native `signtool.exe` |
| WinGet manifest PATH fields | WinGet manifests have no PATH-related fields; PATH is installer-controlled | NSIS `EnvVarUpdate` in `project.nsi` |

---

## Version Compatibility

| Package | Compatible With | Notes |
|---------|-----------------|-------|
| `apple-actions/import-codesign-certs@v6` | `macos-14` GitHub runner | v6 targets Node 24; pre-installed on current macOS runners |
| `xcrun notarytool` | Xcode 13+ (pre-installed on `macos-14`) | Replaces deprecated `altool`; `altool` was sunset Nov 2023 — do not use |
| OV certificate PFX signing | `windows-2022` runner | `signtool.exe` in Windows SDK pre-installed; locate with `Get-ChildItem -Path "C:\Program Files (x86)\Windows Kits" -Filter signtool.exe -Recurse` |
| NSIS `EnvVarUpdate.nsh` | NSIS 2.x+ (bundled with Wails build toolchain) | Wails NSIS build uses NSIS 3.x on Windows runners; `EnvVarUpdate.nsh` must be manually downloaded and placed in `build/windows/installer/` |
| Homebrew `binary` stanza with `#{appdir}` | All current Homebrew versions (2024+) | `#{appdir}` defaults to `/Applications`; verified in official Cask Cookbook |

---

## Integration with Existing release.yml

The signing steps slot into the existing `build-macos` and `build-windows` jobs. No structural changes to `release.yml` are needed — just additions:

**In `build-macos` job, after checkout, before `wails build`:**
```yaml
- name: Import code signing certificate
  uses: apple-actions/import-codesign-certs@<SHA>
  with:
    p12-file-base64: ${{ secrets.MACOS_CERTIFICATE_P12_BASE64 }}
    p12-password: ${{ secrets.MACOS_CERTIFICATE_PASSWORD }}
```

**In `build-macos` job, after `wails build`, before DMG creation:**
```yaml
- name: Sign app bundle
  env:
    MACOS_CERTIFICATE_NAME: ${{ secrets.MACOS_CERTIFICATE_NAME }}
  run: |
    /usr/bin/codesign --deep --force --options runtime \
      --sign "$MACOS_CERTIFICATE_NAME" \
      build/bin/StorCat.app -v

- name: Notarize and staple
  env:
    APPLE_ID: ${{ secrets.MACOS_NOTARIZATION_APPLE_ID }}
    APPLE_PASSWORD: ${{ secrets.MACOS_NOTARIZATION_PASSWORD }}
    TEAM_ID: ${{ secrets.MACOS_NOTARIZATION_TEAM_ID }}
  run: |
    xcrun notarytool store-credentials "notarytool-profile" \
      --apple-id "$APPLE_ID" \
      --team-id "$TEAM_ID" \
      --password "$APPLE_PASSWORD"
    ditto -c -k --keepParent build/bin/StorCat.app /tmp/StorCat-notarize.zip
    xcrun notarytool submit /tmp/StorCat-notarize.zip \
      --keychain-profile "notarytool-profile" \
      --wait
    xcrun stapler staple build/bin/StorCat.app
```

**In `build-windows` job, after `wails build -nsis`, before artifact upload:**
```yaml
- name: Sign Windows binaries
  shell: pwsh
  env:
    WINDOWS_CERTIFICATE: ${{ secrets.WINDOWS_CERTIFICATE_PFX_BASE64 }}
    WINDOWS_CERTIFICATE_PASSWORD: ${{ secrets.WINDOWS_CERTIFICATE_PASSWORD }}
  run: |
    $pfxBytes = [Convert]::FromBase64String($env:WINDOWS_CERTIFICATE)
    $pfxPath = Join-Path $env:RUNNER_TEMP "storcat.pfx"
    [IO.File]::WriteAllBytes($pfxPath, $pfxBytes)
    $signtool = (Get-ChildItem -Path "C:\Program Files (x86)\Windows Kits" -Filter signtool.exe -Recurse | Select-Object -First 1).FullName
    & $signtool sign /fd sha256 /p $env:WINDOWS_CERTIFICATE_PASSWORD /f $pfxPath /t "http://timestamp.sectigo.com" "build\bin\StorCat-${{ steps.version.outputs.VERSION }}-windows-amd64.exe"
    & $signtool sign /fd sha256 /p $env:WINDOWS_CERTIFICATE_PASSWORD /f $pfxPath /t "http://timestamp.sectigo.com" "build\bin\StorCat-${{ steps.version.outputs.VERSION }}-windows-amd64-installer.exe"
    Remove-Item $pfxPath -Force
```

---

## Alternatives Considered

| Recommended | Alternative | When to Use Alternative |
|-------------|-------------|-------------------------|
| Native `xcrun codesign` + `xcrun notarytool` | `GuillaumeFalourd/notary-tools` action | Neither is clearly better; direct xcrun preferred for debuggability and no external action dependency |
| OV certificate PFX in GitHub secret | EV certificate with self-hosted runner + hardware token | Use EV if SmartScreen reputation is a blocker and you can maintain a self-hosted Windows runner; OV is sufficient for initial launch |
| NSIS `EnvVarUpdate.nsh` for Windows PATH | NSIS `WriteRegStr` directly to PATH registry key | `EnvVarUpdate` is safer — it handles deduplication, size limits, and proper delimiter handling. Direct registry write can corrupt PATH if existing value is large |
| Direct `signtool.exe` PowerShell | `azure/trusted-signing-action` | Azure Trusted Signing is Microsoft's cloud HSM service; appropriate if budget allows; avoids cert rotation; overkill for solo dev |

---

## Sources

- [Apple-Actions/import-codesign-certs releases](https://github.com/Apple-Actions/import-codesign-certs/releases) — v6.0.0 Dec 2024, SHA `b610f78` (HIGH — fetched directly)
- [Automatic Code-signing and Notarization — Federico Terzi](https://federicoterzi.com/blog/automatic-code-signing-and-notarization-for-macos-apps-using-github-actions/) — complete workflow pattern with 7 secrets (HIGH — detailed technical blog, verified steps)
- [Homebrew Cask Cookbook — binary stanza](https://github.com/Homebrew/brew/blob/master/docs/Cask-Cookbook.md) — `binary` stanza syntax, `#{appdir}`, `target:` rename (HIGH — official Homebrew docs, fetched directly)
- [dlemstra/code-sign-action archived](https://github.com/dlemstra/code-sign-action) — archived Oct 2025 (HIGH — verified from GitHub directly)
- [Windows OV certificate no hardware token](https://federicoterzi.com/blog/automatic-codesigning-on-windows-using-github-actions/) — OV PFX + signtool.exe pattern (HIGH — technical blog, well-documented pattern)
- [NSIS EnvVarUpdate](https://nsis.sourceforge.io/Environmental_Variables:_append,_prepend,_and_remove_entries) — PATH manipulation in NSIS (MEDIUM — official NSIS wiki)
- [WinGet PATH discussion](https://github.com/microsoft/winget-cli/discussions/5091) — WinGet does not manage PATH; delegates to installer (HIGH — official Microsoft repo discussion)
- [Windows signtool GitHub Actions](https://www.advancedinstaller.com/code-signing-with-github-actions.html) — PFX decode + signtool pattern (MEDIUM — verified against multiple sources)
- Apple notarytool replacing altool — [Apple Developer docs confirm altool sunset Nov 2023](https://developer.apple.com/news/releases/), current tool is `xcrun notarytool` (HIGH — official Apple developer news)

---
*Stack research for: StorCat v2.3.0 Code Signing & Package Manager CLI PATH*
*Researched: 2026-03-27*
