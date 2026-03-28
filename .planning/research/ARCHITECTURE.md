# Architecture Research

**Domain:** macOS/Windows code signing + notarization + CLI PATH setup for Go/Wails desktop app
**Researched:** 2026-03-27
**Confidence:** HIGH — based on official GitHub Docs (Apple certificate import), Apple notarytool docs, Federico Terzi's signing guide, Melatonin's CI guide, official Homebrew Cask Cookbook, NSIS Path Manipulation docs

---

## Scope

This document covers v2.3.0: how macOS Developer ID signing + notarization and Windows Authenticode signing integrate into the existing `release.yml` pipeline, and how Homebrew/WinGet installations expose `storcat` on PATH.

The existing v2.2.0 pipeline, source, and CLI architecture are unchanged — this milestone is entirely additive steps within existing jobs.

---

## Existing Pipeline (v2.2.0 Baseline)

```
tag push (v*.*.*)
    │
    ├─► build-macos (macos-14)
    │       wails build → codesign (NONE) → hdiutil → upload DMG
    │
    ├─► build-windows (windows-2022)
    │       wails build -nsis → upload .exe + installer.exe (UNSIGNED)
    │
    ├─► build-linux-amd64 (ubuntu-22.04)
    │       wails build → AppImage + .deb → upload
    │
    ├─► build-linux-arm64 (ubuntu-22.04-arm)
    │       wails build → .deb → upload
    │
    └─► release (fan-in)
            create draft release, upload all artifacts

    [on: release:published]
    distribute.yml
        update-homebrew → push cask to tap (app stanza only — no CLI symlink)
        update-winget   → submit PR to microsoft/winget-pkgs
```

**Current problems this milestone solves:**
- macOS Gatekeeper blocks unsigned DMG with "Apple cannot check it for malicious software"
- Windows SmartScreen blocks unsigned installer with "Windows protected your PC"
- `brew install --cask storcat` installs the GUI app but `storcat` CLI is not on PATH
- `winget install scottkw.StorCat` installs to Program Files but `storcat` CLI is not on PATH

---

## Target Architecture (v2.3.0)

### System Overview

```
tag push (v*.*.*)
    │
    ├─► build-macos (macos-14)                          [MODIFIED]
    │       ├── wails build -platform darwin/universal
    │       ├── Import p12 cert → ephemeral keychain     [NEW]
    │       ├── codesign --deep --options=runtime        [NEW]
    │       ├── xcrun notarytool submit --wait            [NEW]
    │       ├── xcrun stapler staple StorCat.app         [NEW]
    │       ├── hdiutil create StorCat-*.dmg
    │       └── upload-artifact (signed + notarized DMG)
    │
    ├─► build-windows (windows-2022)                    [MODIFIED]
    │       ├── wails build -platform windows/amd64 -windowsconsole -nsis
    │       ├── signtool sign StorCat-*.exe              [NEW]
    │       ├── signtool sign StorCat-*-installer.exe    [NEW]
    │       └── upload-artifact (signed binaries)
    │
    ├─► build-linux-amd64 (ubuntu-22.04)                [UNCHANGED]
    ├─► build-linux-arm64 (ubuntu-22.04-arm)            [UNCHANGED]
    └─► release (fan-in)                                [UNCHANGED]

    [on: release:published]
    distribute.yml                                      [UNCHANGED]
        update-homebrew → cask has binary stanza now    [template MODIFIED]
        update-winget   → NSIS installer has PATH reg   [installer.nsi NEW]
```

### Component Responsibilities

| Component | v2.2.0 State | v2.3.0 Change | Files |
|-----------|-------------|---------------|-------|
| `build-macos` job | Builds + packages unsigned DMG | Add 4 signing/notarization steps | `release.yml` |
| `build-windows` job | Builds + packages unsigned installer | Add signtool step after build | `release.yml` |
| `release` fan-in job | Creates draft release | No change | `release.yml` |
| `distribute.yml` update-homebrew | Pushes cask update | Template already correct — cask template adds `binary` stanza | `storcat.rb.template` |
| `distribute.yml` update-winget | Submits WinGet PR | No workflow change — PATH comes from NSIS installer script | `build/windows/installer.nsi` |
| GitHub Actions Secrets | HOMEBREW_TAP_TOKEN, WINGET_TOKEN | Add 9 new signing secrets | GitHub Settings |

---

## Files: New vs Modified

| File | Status | What Changes |
|------|--------|--------------|
| `.github/workflows/release.yml` | MODIFIED | Add signing steps to `build-macos` (4 steps) and `build-windows` (1 step) |
| `packaging/homebrew/storcat.rb.template` | MODIFIED | Add `binary` stanza after `app` stanza |
| `build/windows/installer.nsi` | NEW | Custom NSIS script with PATH registration; Wails v2 respects this path |

No new directories. No changes to source Go or TypeScript code. No changes to `distribute.yml` logic.

---

## Architectural Patterns

### Pattern 1: macOS Keychain-Based Signing in Headless CI

**What:** Import a p12 certificate into a temporary ephemeral keychain, sign with codesign, then destroy the keychain. Required because macOS would otherwise show a UI password dialog that blocks a headless runner.

**When to use:** Any macOS GitHub Actions runner — hosted or self-hosted.

**Sequence in `build-macos` job:**

```yaml
# Step A: Import certificate into ephemeral keychain
- name: Import signing certificate
  env:
    MACOS_CERT_P12_BASE64: ${{ secrets.MACOS_CERT_P12_BASE64 }}
    MACOS_CERT_PASSWORD: ${{ secrets.MACOS_CERT_PASSWORD }}
    MACOS_KEYCHAIN_PASSWORD: ${{ secrets.MACOS_KEYCHAIN_PASSWORD }}
  run: |
    echo -n "$MACOS_CERT_P12_BASE64" | base64 --decode -o /tmp/certificate.p12
    security create-keychain -p "$MACOS_KEYCHAIN_PASSWORD" build.keychain
    security default-keychain -s build.keychain
    security unlock-keychain -p "$MACOS_KEYCHAIN_PASSWORD" build.keychain
    security import /tmp/certificate.p12 -k build.keychain \
      -P "$MACOS_CERT_PASSWORD" -T /usr/bin/codesign
    security set-key-partition-list -S apple-tool:,apple: \
      -s -k "$MACOS_KEYCHAIN_PASSWORD" build.keychain
    rm /tmp/certificate.p12

# Step B: Sign the .app bundle
- name: Sign app bundle
  env:
    MACOS_CERT_NAME: ${{ secrets.MACOS_CERT_NAME }}
  run: |
    codesign --force --deep --strict \
      --options=runtime --timestamp \
      -s "$MACOS_CERT_NAME" \
      build/bin/StorCat.app

# Step C: Submit for notarization (sign must complete first)
- name: Notarize
  env:
    APPLE_ID: ${{ secrets.APPLE_ID }}
    APPLE_APP_PASSWORD: ${{ secrets.APPLE_APP_PASSWORD }}
    APPLE_TEAM_ID: ${{ secrets.APPLE_TEAM_ID }}
  run: |
    ditto -c -k --keepParent build/bin/StorCat.app /tmp/StorCat-notarize.zip
    xcrun notarytool submit /tmp/StorCat-notarize.zip \
      --apple-id "$APPLE_ID" \
      --password "$APPLE_APP_PASSWORD" \
      --team-id "$APPLE_TEAM_ID" \
      --wait --timeout 10m
    rm /tmp/StorCat-notarize.zip

# Step D: Staple notarization ticket (notarize must complete first)
- name: Staple
  run: xcrun stapler staple build/bin/StorCat.app

# Existing Step (UNCHANGED): Create macOS DMG
- name: Create macOS DMG
  run: |
    cd build/bin
    hdiutil create \
      -volname "StorCat" \
      -srcfolder StorCat.app \
      -ov -format UDZO \
      "StorCat-${{ steps.version.outputs.VERSION }}-darwin-universal.dmg"
```

**Critical ordering constraint:** codesign → notarize → staple → hdiutil. These four steps are sequential and cannot be parallelized.

**Why staple before DMG:** Stapling embeds the notarization ticket into the `.app` bundle. When the DMG is created afterward, the stapled `.app` inside it already carries the ticket. Users can open the DMG without internet access and Gatekeeper will accept it.

### Pattern 2: Windows OV Certificate Signing with signtool

**What:** Decode a base64-encoded PFX from a secret, invoke signtool.exe (pre-installed on all Windows runners), then immediately delete the decoded file. Works on GitHub-hosted `windows-2022` runners without a hardware dongle.

**OV vs EV trade-off:** OV (Organization Validation) certificates work on hosted runners and require no HSM/hardware. SmartScreen may still show a warning on initial downloads until reputation builds over time. EV certificates (via Azure Key Vault) eliminate SmartScreen warnings entirely but require Azure setup and cost more. For an open-source project with modest download volume, OV is the practical starting point.

**Sequence in `build-windows` job (after wails build, before rename steps):**

```yaml
- name: Sign Windows binaries
  env:
    WINDOWS_CERT_P12_BASE64: ${{ secrets.WINDOWS_CERT_P12_BASE64 }}
    WINDOWS_CERT_PASSWORD: ${{ secrets.WINDOWS_CERT_PASSWORD }}
  run: |
    $certBytes = [Convert]::FromBase64String($env:WINDOWS_CERT_P12_BASE64)
    $certPath = Join-Path $env:TEMP "codesign.pfx"
    [IO.File]::WriteAllBytes($certPath, $certBytes)
    $signtool = (Get-Command signtool.exe -ErrorAction SilentlyContinue)?.Source
    if (-not $signtool) {
      $signtool = "C:\Program Files (x86)\Windows Kits\10\App Certification Kit\signtool.exe"
    }
    & $signtool sign /fd sha256 /td sha256 `
      /tr http://timestamp.digicert.com `
      /f $certPath /p $env:WINDOWS_CERT_PASSWORD `
      "build\bin\StorCat.exe"
    & $signtool sign /fd sha256 /td sha256 `
      /tr http://timestamp.digicert.com `
      /f $certPath /p $env:WINDOWS_CERT_PASSWORD `
      "build\bin\StorCat-amd64-installer.exe"
    Remove-Item $certPath
  shell: pwsh
```

**Sign both files:** Both `StorCat.exe` (portable binary) and `StorCat-amd64-installer.exe` (NSIS installer) must be signed. The installer contains the portable binary — signing the installer wraps a signed executable, but signing both individually is required for clean verification.

**Insert before rename steps:** The existing rename steps reference `build\bin\StorCat.exe` and `build\bin\StorCat-amd64-installer.exe`. Signing must happen before rename, otherwise paths will not match.

### Pattern 3: Homebrew Cask with app + binary Stanzas

**What:** A single Homebrew cask installs the `.app` bundle to `/Applications` (existing) and creates a lowercase `storcat` symlink in `$(brew --prefix)/bin` (new). The symlink points into the `.app` bundle's embedded binary.

**Binary path inside Wails .app:** `StorCat.app/Contents/MacOS/StorCat`

**Homebrew binary stanza uses `#{appdir}` to reference the installed app location:**

```ruby
# packaging/homebrew/storcat.rb.template — modified
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
  binary "#{appdir}/StorCat.app/Contents/MacOS/StorCat", target: "storcat"   # NEW

  zap trash: [
    "~/Library/Application Support/StorCat",
    "~/Library/Preferences/com.kenscott.storcat.plist",
    "~/Library/Saved Application State/com.kenscott.storcat.savedState",
    "~/Library/Caches/com.kenscott.storcat",
  ]
end
```

**`target: "storcat"`** creates the symlink as lowercase `storcat` in `$(brew --prefix)/bin`. Without this, the symlink would be `StorCat` — uppercase, inconsistent with CLI conventions, and awkward to type.

**`#{appdir}`** expands to `/Applications` — this is the Homebrew Cask variable for the application directory. The binary stanza path must be relative to this, not relative to the cask staging area.

**Signed binary requirement:** macOS 15+ (Sequoia) blocks unsigned binaries from being symlinked into PATH by Homebrew casks. The signed DMG from Pattern 1 is a prerequisite for this pattern to work without Gatekeeper interference. Phase 4 (Homebrew CLI PATH) must follow Phase 2 (macOS signing).

### Pattern 4: NSIS PATH Registration for Windows CLI

**What:** A custom NSIS installer script adds the installation directory to the system PATH during install and removes it during uninstall. Wails v2 respects `build/windows/installer.nsi` as a custom script override when `-nsis` is used.

**Why custom script:** The Wails-generated NSIS script does not register PATH. Wails v2 generates NSIS scripts from its internal templates — placing `build/windows/installer.nsi` overrides the template.

**PATH registration mechanism:** Direct registry write to `HKLM\SYSTEM\CurrentControlSet\Control\Session Manager\Environment`. The NSIS installer already requests admin elevation, so the HKLM write succeeds. After writing, broadcast `WM_SETTINGCHANGE` so open terminals pick up the change without reboot.

**Key NSIS additions:**

```nsis
; In Section "Install" block — add after CopyFiles
WriteRegExpandStr HKLM \
  "SYSTEM\CurrentControlSet\Control\Session Manager\Environment" \
  "Path" "$INSTDIR;%Path%"
SendMessage ${HWND_BROADCAST} ${WM_SETTINGCHANGE} 0 "STR:Environment" /TIMEOUT=5000

; In Section "Uninstall" block — remove install dir from PATH
; Option A (simple, may leave behind if path changed):
DeleteRegValue HKLM \
  "SYSTEM\CurrentControlSet\Control\Session Manager\Environment" \
  "Path"
; Option B (correct, requires EnVar plugin):
EnVar::SetHKLM
EnVar::DeleteValue "Path" "$INSTDIR"
```

**EnVar plugin trade-off:** The EnVar NSIS plugin correctly handles PATH append/remove without clobbering existing PATH entries. The simple `WriteRegExpandStr` approach prepends `$INSTDIR;` to the current `%Path%` value, which works for install but leaves a stale entry if uninstalled without the plugin. For a v2.3.0 first implementation, the simpler approach is acceptable — add EnVar as a follow-up if needed.

**WinGet manifest impact:** No changes to YAML manifests. PATH registration is an installer behavior, not a WinGet manifest concern.

---

## Data Flow

### macOS Signing Data Flow

```
GitHub Secret: MACOS_CERT_P12_BASE64
    │ base64 --decode
    ▼
/tmp/certificate.p12
    │ security import
    ▼
build.keychain  (ephemeral — lives only for this runner job)
    │ codesign reads cert from keychain
    ▼
StorCat.app  (signed: Developer ID Application)
    │ ditto -c -k → zip
    ▼
/tmp/StorCat-notarize.zip
    │ xcrun notarytool submit --wait
    ▼
Apple Notary Service  (HTTPS outbound from macos-14 runner)
    │ success → staple ticket
    ▼
StorCat.app  (signed + notarized + stapled)
    │ hdiutil create
    ▼
StorCat-vX.Y.Z-darwin-universal.dmg  (uploaded to GitHub release)
```

### Windows Signing Data Flow

```
GitHub Secret: WINDOWS_CERT_P12_BASE64
    │ [Convert]::FromBase64String
    ▼
%TEMP%\codesign.pfx
    │ signtool.exe /fd sha256 /tr timestamp
    ▼
StorCat-vX.Y.Z-windows-amd64.exe           (signed)
StorCat-vX.Y.Z-windows-amd64-installer.exe (signed)
    │ Remove-Item (cert file deleted)
    │ upload-artifact
    ▼
GitHub Release
```

### CLI PATH Setup Data Flow (macOS via Homebrew)

```
User: brew install --cask scottkw/storcat/storcat
    │
    ▼ Homebrew downloads signed DMG
    ▼ Mounts DMG, reads cask definition
    │
    ├─► app stanza → /Applications/StorCat.app
    │
    └─► binary stanza → symlink created:
            $(brew --prefix)/bin/storcat
            → /Applications/StorCat.app/Contents/MacOS/StorCat
    │
    ▼ $(brew --prefix)/bin is already in PATH
    ▼
$ storcat --help  (works immediately in any terminal)
```

### CLI PATH Setup Data Flow (Windows via WinGet)

```
User: winget install scottkw.StorCat
    │
    ▼ WinGet downloads signed NSIS installer
    ▼ Runs installer (admin elevation requested)
    │
    ├─► Installs StorCat.exe to C:\Program Files\StorCat\
    │
    └─► NSIS PATH registration:
            WriteRegExpandStr HKLM\...\Environment "Path" "$INSTDIR;%Path%"
            SendMessage WM_SETTINGCHANGE
    │
    ▼ New terminals see updated PATH
    ▼
> storcat --help  (works in any new terminal session)
```

---

## Integration Points

### New GitHub Actions Secrets Required

| Secret Name | Purpose | How to Obtain |
|-------------|---------|---------------|
| `MACOS_CERT_P12_BASE64` | Developer ID Application certificate | Export from Keychain as .p12, `base64 -i cert.p12` (macOS) |
| `MACOS_CERT_PASSWORD` | p12 export password | Set when exporting from Keychain |
| `MACOS_CERT_NAME` | Full cert identifier string | `security find-identity -v -p codesigning` output |
| `MACOS_KEYCHAIN_PASSWORD` | Ephemeral CI keychain password | Any strong random string (e.g. `openssl rand -hex 16`) |
| `APPLE_ID` | Apple Developer account email | developer.apple.com account |
| `APPLE_APP_PASSWORD` | App-specific password for notarytool | appleid.apple.com > Sign-In and Security > App-Specific Passwords |
| `APPLE_TEAM_ID` | Developer team identifier | developer.apple.com > Membership page |
| `WINDOWS_CERT_P12_BASE64` | OV/EV code signing certificate | CA provider (DigiCert, Sectigo, Comodo) |
| `WINDOWS_CERT_PASSWORD` | PFX password | Set when obtaining certificate from CA |

**Total: 9 new secrets** (HOMEBREW_TAP_TOKEN and WINGET_TOKEN from v2.2.0 are unchanged)

**Environment protection recommendation:** Create a `release` environment in GitHub Settings > Environments. Move all signing secrets (these 9) to that environment. Set required reviewers or branch protection so signing credentials are only accessible to the release workflow.

### Internal Boundaries

| Boundary | Communication | Constraint |
|----------|---------------|------------|
| `build-macos` → Apple Notary Service | `xcrun notarytool submit` outbound HTTPS | Requires network from macos-14; timeouts rare but add `--timeout 10m` |
| `build-macos` → GitHub Secrets | `env:` block injection | Secrets never appear in logs; use only in `env:` not `run:` interpolation |
| `build-windows` → Windows cert store | PowerShell + signtool.exe | PFX written to `$env:TEMP` and deleted in same step |
| Homebrew binary stanza → `$(brew --prefix)/bin` | Symlink created at cask install | `brew --prefix` = `/usr/local` on Intel, `/opt/homebrew` on Apple Silicon — both are in PATH |
| NSIS installer → Windows PATH | Registry write to HKLM | Requires elevation — NSIS installer already requests admin |
| Wails `-nsis` flag → `build/windows/installer.nsi` | File override — Wails checks this path | Must be present and valid NSIS before `-nsis` build runs |

---

## Build Order for Implementation Phases

Dependencies flow: secrets must exist before signing steps can run; signing must work before cask PATH can be validated.

### Phase 1: GitHub Secrets Setup (manual, prerequisite)

No code changes. Pure configuration. Must complete before any signing workflow changes.

```
1a. Obtain/locate Developer ID Application certificate → export as p12
1b. Create app-specific password at appleid.apple.com
1c. Obtain/purchase Windows OV certificate → export as PFX
1d. Add all 9 signing secrets to GitHub repo (or release environment)
1e. Set up release environment with reviewer protection (optional but recommended)
```

Validation: GitHub Settings > Secrets > confirm all 9 names present.

### Phase 2: macOS Signing + Notarization (modify `release.yml`)

Depends on Phase 1. Can be developed and tested independently from Windows signing.

```
2a. Add 4 signing steps to build-macos job in release.yml
2b. Tag a test release (v2.3.0-rc1), observe workflow
2c. Download DMG, run: spctl -vvv --assess --type exec /Volumes/StorCat/StorCat.app
    Expected: accepted
2d. Verify: codesign -dv --verbose=4 StorCat.app shows Developer ID
```

### Phase 3: Windows Signing (modify `release.yml`)

Depends on Phase 1. Independent of Phase 2 — can run in parallel with Phase 2.

```
3a. Add signtool step to build-windows job in release.yml
3b. Tag a test release, download installer
3c. Verify: signtool verify /v /pa StorCat-*-installer.exe
    Expected: "Successfully verified"
```

### Phase 4: Homebrew CLI PATH (modify `storcat.rb.template`)

Depends on Phase 2. Signed DMG is required for macOS 15+ Gatekeeper to allow symlink creation.

```
4a. Add binary stanza to storcat.rb.template
4b. Publish a release (must be real publish to trigger distribute.yml)
4c. brew install --cask scottkw/storcat/storcat
4d. Verify: which storcat → /opt/homebrew/bin/storcat (or /usr/local/bin/storcat)
4e. Verify: storcat version → vX.Y.Z
```

### Phase 5: WinGet CLI PATH (new `build/windows/installer.nsi`)

Depends on Phase 3. Signed installer should be in place before new NSIS script adds PATH.

```
5a. Create build/windows/installer.nsi with PATH registration
    (base on Wails-generated template + add WriteRegExpandStr)
5b. Build locally with wails build -nsis, test installer in VM
5c. Tag a release, verify installed CLI:
    winget install scottkw.StorCat
    (new terminal) storcat version
```

---

## Anti-Patterns

### Anti-Pattern 1: Sign the DMG Instead of the .app

**What people do:** Run `codesign` on the `.dmg` file after `hdiutil create`.

**Why it's wrong:** Gatekeeper checks the `.app` bundle, not the DMG container. Signing the DMG does nothing for Gatekeeper acceptance of the enclosed app.

**Do this instead:** Sign `.app` → notarize → staple `.app` → then create DMG. The DMG wraps an already-signed and stapled app.

### Anti-Pattern 2: Notarize Without `--options=runtime`

**What people do:** Omit `--options=runtime` from the codesign invocation.

**Why it's wrong:** Apple's notarization service requires hardened runtime. Without it, notarization will be rejected: "The executable does not have the hardened runtime enabled."

**Do this instead:** Always include `--options=runtime --timestamp` in the codesign command.

### Anti-Pattern 3: Use Main Apple ID Password for notarytool

**What people do:** Store the main Apple Developer account password as a secret.

**Why it's wrong:** `xcrun notarytool` in CI requires an app-specific password, not the main account password. The main password may also require 2FA. Using the main password gives notarization access the same credential as full account management.

**Do this instead:** Create an app-specific password at appleid.apple.com > Sign-In and Security > App-Specific Passwords. Store as `APPLE_APP_PASSWORD`. It is notarization-scoped and revocable independently.

### Anti-Pattern 4: Homebrew binary Stanza Without `target:`

**What people do:** Write `binary "#{appdir}/StorCat.app/Contents/MacOS/StorCat"` without `target:`.

**Why it's wrong:** Creates symlink as `StorCat` (capital S) in `$(brew --prefix)/bin`. Convention for CLI tools is lowercase. Users typing `storcat` get "command not found."

**Do this instead:** Always specify `target: "storcat"` to create a lowercase symlink name.

### Anti-Pattern 5: Hardcode signtool.exe Path

**What people do:** Hardcode `"C:\Program Files (x86)\Windows Kits\10\App Certification Kit\signtool.exe"`.

**Why it's wrong:** The Windows SDK path varies between runner image versions and breaks silently when the runner image updates.

**Do this instead:** Resolve with `Get-Command signtool.exe` first, fall back to known path only if not found. Or use a marketplace action (`dlemstra/code-sign-action`) that handles resolution internally.

### Anti-Pattern 6: Phase 4 Before Phase 2

**What people do:** Add the Homebrew binary stanza before signing is in place.

**Why it's wrong:** On macOS 15+, Gatekeeper prevents creation of PATH symlinks from unsigned binaries inside casks. The `brew install` will fail or warn depending on security settings.

**Do this instead:** Phases 2 → 4 in order. Signed DMG must exist before Homebrew binary stanza is added to the cask.

---

## Scalability Considerations

This is CI/CD infrastructure. Scale concerns are operational, not user-facing.

| Concern | Impact | Mitigation |
|---------|--------|------------|
| Notarization timeout | Rare; typically 10–120 seconds | Add `--timeout 10m` to notarytool submit |
| Certificate expiry | Apple: 1 year; Windows CA: 1–3 years | Set calendar reminder; renewal = update 2–3 secrets, no workflow changes |
| SmartScreen reputation (OV cert) | Users see warning initially; builds over time | Document in README; consider EV cert if reputation growth is too slow |
| Secrets rotation | Manual — update GitHub secret value | Document renewal process in a RUNBOOK; no code changes required |
| NSIS script maintenance | Must be updated if Wails changes its template | Re-check `build/windows/installer.nsi` when upgrading Wails version |

---

## Sources

- [Installing an Apple certificate on macOS runners — GitHub Docs](https://docs.github.com/en/actions/use-cases-and-examples/deploying/installing-an-apple-certificate-on-macos-runners-for-xcode-development)
- [Automatic Code-signing and Notarization for macOS with GitHub Actions — Federico Terzi](https://federicoterzi.com/blog/automatic-code-signing-and-notarization-for-macos-apps-using-github-actions/)
- [How to code sign and notarize macOS audio plugins in CI — Melatonin](https://melatonin.dev/blog/how-to-code-sign-and-notarize-macos-audio-plugins-in-ci/)
- [How to code sign Windows installers with an EV cert on GitHub Actions — Melatonin](https://melatonin.dev/blog/how-to-code-sign-windows-installers-with-an-ev-cert-on-github-actions/)
- [Homebrew Cask Cookbook — Official Docs](https://docs.brew.sh/Cask-Cookbook)
- [NSIS Path Manipulation — Official Docs](https://nsis.sourceforge.io/Path_Manipulation)
- [NSIS Setting Environment Variables — Official Docs](https://nsis.sourceforge.io/Setting_Environment_Variables)
- [Windows Code Signing — Tauri v1 (reference pattern for OV via PFX)](https://v1.tauri.app/v1/guides/distribution/sign-windows/)
- Direct analysis: `.github/workflows/release.yml`, `.github/workflows/distribute.yml`, `packaging/homebrew/storcat.rb.template`, `packaging/winget/manifests/s/scottkw/StorCat/2.1.0/scottkw.StorCat.installer.yaml`

---

*Architecture research for: StorCat v2.3.0 code signing + notarization + CLI PATH setup*
*Researched: 2026-03-27*
