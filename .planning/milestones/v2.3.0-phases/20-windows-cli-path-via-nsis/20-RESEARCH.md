# Phase 20: Windows CLI PATH via NSIS - Research

**Researched:** 2026-03-28
**Domain:** NSIS custom installer scripting, Windows system PATH registration, Wails v2 build system
**Confidence:** HIGH

## Summary

Phase 20 adds Windows system PATH registration to the StorCat NSIS installer. When a user runs `winget install scottkw.StorCat`, the installer must add `$INSTDIR` to the system PATH so that `storcat version` works from any new terminal session without manual PATH configuration.

The primary mechanism is a custom `build/windows/installer/project.nsi` file that Wails v2 respects when it already exists on disk (verified against Wails v2.10.2 source at `pkg/buildassets/buildassets.go:48-66`). The custom script adds two NSIS calls after `wails.writeUninstaller`: one to append `$INSTDIR` to the HKLM system PATH using the EnVar plugin, and one to broadcast `WM_SETTINGCHANGE` so running shells pick up the change immediately. The uninstall section mirrors this with a `DeleteValue` call.

The critical technical choice is the **EnVar plugin** over the older `EnvVarUpdate` macro. The NSIS official documentation explicitly warns that `EnvVarUpdate` will silently corrupt PATH variables longer than `NSIS_MAX_STRLEN` (1024 bytes by default on the standard NSIS build shipped with GitHub Actions windows-2022). The EnVar plugin avoids this by using the native Win32 `RegQueryValueEx()` function directly and is not subject to the NSIS string length limit.

**Primary recommendation:** Create `build/windows/installer/project.nsi` as a copy of the Wails v2.10.2 default template with three additions: (1) an `!addplugindir` pointing to a bundled `build/windows/installer/Plugins/x86-unicode/` directory containing `EnVar.dll`, (2) `EnVar::SetHKLM` + `EnVar::AddValue "PATH" "$INSTDIR"` + result pop in the install section, and (3) `EnVar::SetHKLM` + `EnVar::DeleteValue "PATH" "$INSTDIR"` in the uninstall section, both followed by the `SendMessage` broadcast.

## Project Constraints (from CLAUDE.md)

- pnpm preferred for Node; `uv`/`pip` for Python; `go mod` for Go
- Python: always use venv, never global installs
- GSD workflow: use `/gsd:execute-phase` for planned phase work — do not make direct repo edits outside GSD unless explicitly asked
- No Electron dependencies — all fixes must use Go/Wails patterns
- Chesterton's Fence: articulate why existing code exists before removing it
- StorCat v2.0.0 must have exact feature parity with Electron v1.2.3 — no regressions
- Backward compatibility: v1.0 catalog JSON/HTML must remain readable

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| PKG-02 | NSIS installer adds install directory to system PATH via EnvVarUpdate | Covered: EnVar plugin is the correct modern replacement for EnvVarUpdate; same outcome, safer implementation. |
| PKG-04 | `storcat version` works from any new terminal after WinGet install (Windows) | Covered: EnVar PATH addition + SendMessage broadcast enables this; `storcat` CLI dispatch confirmed in main.go. |
</phase_requirements>

## Standard Stack

### Core
| Tool | Version | Purpose | Why Standard |
|------|---------|---------|--------------|
| Wails v2 custom `project.nsi` | v2.10.2 | Override default NSIS template | Documented pattern: Wails reads local file if it exists, writes template only if missing |
| NSIS (makensis) | 3.10 (pre-installed on windows-2022) | Compiles NSIS installer script | Used by `wails build -nsis`; no separate install step needed in CI |
| EnVar NSIS plugin | Latest (GsNSIS/EnVar) | PATH add/remove without string truncation | Official NSIS wiki recommends EnVar over EnvVarUpdate for PATH modification |
| SendMessage WM_SETTINGCHANGE | N/A (NSIS builtin) | Notifies shells of PATH change immediately | Without broadcast, PATH change only visible in new processes spawned after a logoff |

### Supporting
| Item | Purpose | When to Use |
|------|---------|-------------|
| WinGet manifest `scottkw.StorCat.installer.yaml` | Declares installer type as `nullsoft` | Already done for v2.1.0 — the nullsoft installer type handles silent install switches automatically |
| `distribute.yml` `update-winget-manifests` job | Template-generates new manifests from 2.1.0 stubs | No change needed — WinGet manifest generation is already wired up |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| EnVar plugin | EnvVarUpdate macro | EnvVarUpdate silently corrupts PATH > 1024 chars; EnVar uses Win32 API, no limit |
| EnVar plugin | WriteRegStr + SendMessage (manual registry) | More boilerplate, same risk of truncation; EnVar handles all edge cases |
| Bundling EnVar.dll in repo | CI step to download/install plugin | Bundling is simpler, reproducible, no network dependency during build |

**Installation (EnVar plugin DLL):**
Download from https://github.com/GsNSIS/EnVar/releases, place `EnVar.dll` at:
```
build/windows/installer/Plugins/x86-unicode/EnVar.dll
```
Reference in `project.nsi` with `!addplugindir "Plugins"`.

**Version verification:** NSIS 3.10 is pre-installed on `windows-2022` GitHub Actions runner. No separate install step needed.

## Architecture Patterns

### Recommended File Structure
```
build/
└── windows/
    ├── icon.ico
    ├── info.json
    ├── wails.exe.manifest
    └── installer/           # Created by first wails build -nsis run
        ├── project.nsi      # CUSTOM (checked into repo) — Wails respects this
        ├── wails_tools.nsh  # AUTO-GENERATED by Wails — do not edit
        └── Plugins/
            └── x86-unicode/
                └── EnVar.dll  # Bundled plugin DLL (checked into repo)
```

**Critical:** `wails_tools.nsh` is ALWAYS regenerated by Wails (`ReadOriginalFileWithProjectDataAndSave` — always writes from embedded template). Never put PATH logic in `wails_tools.nsh`. Only `project.nsi` is preserved.

### Pattern 1: Wails Custom project.nsi

**What:** Wails v2 `buildassets.ReadFile()` checks if the local file exists first. If `build/windows/installer/project.nsi` is present in the repo, Wails uses it. If absent, Wails extracts the default template.

**Verified against:** `/Users/ken/go/pkg/mod/github.com/wailsapp/wails/v2@v2.10.2/pkg/buildassets/buildassets.go` lines 48-66.

**When to use:** Any time the default NSIS installer template needs modification (PATH, custom pages, shortcuts, etc.).

**Example — Install section with PATH addition:**
```nsis
; Source: Wails v2.10.2 default project.nsi + EnVar additions
!addplugindir "Plugins"

Section
    !insertmacro wails.setShellContext

    !insertmacro wails.webview2runtime

    SetOutPath $INSTDIR

    !insertmacro wails.files

    CreateShortcut "$SMPROGRAMS\${INFO_PRODUCTNAME}.lnk" "$INSTDIR\${PRODUCT_EXECUTABLE}"
    CreateShortCut "$DESKTOP\${INFO_PRODUCTNAME}.lnk" "$INSTDIR\${PRODUCT_EXECUTABLE}"

    !insertmacro wails.associateFiles
    !insertmacro wails.associateCustomProtocols

    !insertmacro wails.writeUninstaller

    ; Add install directory to system PATH
    EnVar::SetHKLM
    EnVar::AddValue "PATH" "$INSTDIR"
    Pop $0   ; $0 = 0 on success

    ; Broadcast WM_SETTINGCHANGE so open shells refresh their PATH
    SendMessage ${HWND_BROADCAST} ${WM_SETTINGCHANGE} 0 "STR:Environment" /TIMEOUT=5000
SectionEnd
```

**Example — Uninstall section with PATH removal:**
```nsis
Section "uninstall"
    !insertmacro wails.setShellContext

    ; Remove install directory from system PATH before deleting files
    EnVar::SetHKLM
    EnVar::DeleteValue "PATH" "$INSTDIR"
    Pop $0   ; $0 = 0 on success

    SendMessage ${HWND_BROADCAST} ${WM_SETTINGCHANGE} 0 "STR:Environment" /TIMEOUT=5000

    RMDir /r "$AppData\${PRODUCT_EXECUTABLE}"
    RMDir /r $INSTDIR

    Delete "$SMPROGRAMS\${INFO_PRODUCTNAME}.lnk"
    Delete "$DESKTOP\${INFO_PRODUCTNAME}.lnk"

    !insertmacro wails.unassociateFiles
    !insertmacro wails.unassociateCustomProtocols

    !insertmacro wails.deleteUninstaller
SectionEnd
```

### Pattern 2: EnVar Plugin DLL Bundled in Repo

**What:** The EnVar plugin DLL must be available to `makensis` at compile time. Bundling it in `build/windows/installer/Plugins/x86-unicode/` avoids a CI download step and keeps the build reproducible.

**When to use:** Any NSIS script that uses EnVar functions.

**Example reference in project.nsi:**
```nsis
; Place at top of project.nsi, after !include "wails_tools.nsh"
!addplugindir "${__FILEDIR__}\Plugins"
```

### Anti-Patterns to Avoid

- **Using EnvVarUpdate macro:** The NSIS wiki explicitly states "Do NOT use this function to update %PATH%, use the EnVar_plug-in instead." Windows users with many PATH entries easily exceed the 1024-byte NSIS_MAX_STRLEN, silently corrupting their PATH.
- **Editing wails_tools.nsh:** Wails always overwrites this file from embedded assets (`ReadOriginalFileWithProjectDataAndSave`). Changes will be lost on every build.
- **Writing PATH via WriteRegStr directly:** Bypasses the EnVar plugin's duplicate detection — if user runs installer twice, PATH would contain duplicate entries.
- **Forgetting SendMessage broadcast:** PATH registry change is written, but running terminal sessions (cmd.exe, PowerShell) won't see it until they restart. The broadcast is the difference between "works in new terminal" and "requires reboot."

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| PATH truncation | Custom registry write code | EnVar plugin | Win32 `RegQueryValueEx()` handles paths of any length; NSIS string vars are capped |
| Duplicate PATH entries | Custom check-before-add logic | EnVar `AddValue` | EnVar's `AddValue` is idempotent — does not add if already present |
| PATH broadcast | Manual WM_SETTINGCHANGE + lParam | `SendMessage` NSIS built-in | NSIS provides the constant; manual casting is error-prone |

**Key insight:** PATH manipulation has hidden complexity — truncation, duplicates, encoding, registry type (REG_SZ vs REG_EXPAND_SZ). EnVar handles all of it.

## Common Pitfalls

### Pitfall 1: PATH Corruption via EnvVarUpdate
**What goes wrong:** NSIS built-in `EnvVarUpdate` reads the PATH registry value into a fixed-length NSIS variable. If PATH exceeds `NSIS_MAX_STRLEN` (default 1024), the value is silently truncated. The truncated value is then written back, permanently destroying entries.
**Why it happens:** `ReadRegStr` returns an empty string for values that are too long — NSIS cannot distinguish "empty" from "too long."
**How to avoid:** Use EnVar plugin exclusively for PATH manipulation.
**Warning signs:** CI builds PATH correctly but user machines with many development tools (JDK, Python, Go, etc.) report PATH corruption after install.

### Pitfall 2: wails_tools.nsh Overwritten Every Build
**What goes wrong:** Developer adds PATH logic to `wails_tools.nsh`. After next `wails build -nsis`, all customizations are gone with no warning.
**Why it happens:** Wails always regenerates `wails_tools.nsh` from embedded assets using `ReadOriginalFileWithProjectDataAndSave`, which overwrites regardless of local changes.
**How to avoid:** Only customize `project.nsi`. Verify by checking source: `buildassets.go` line 87 uses `ReadOriginalFileWithProjectDataAndSave`; line 31 uses `ReadFile` (preserves local) for `project.nsi`.
**Warning signs:** PATH logic disappears silently after any `wails build -nsis` invocation.

### Pitfall 3: installer/ Directory Does Not Exist Before First Build
**What goes wrong:** `project.nsi` and `Plugins/` are committed to repo, but `build/windows/installer/` doesn't exist on the CI runner until `wails build -nsis` creates it.
**Why it happens:** Wails creates the `installer/` directory and writes `wails_tools.nsh` during the build. The directory itself doesn't need to pre-exist for Wails to work.
**How to avoid:** This is handled automatically — `build/windows/installer/` will be created by Wails during build. But the committed `project.nsi` and `Plugins/` must be at the correct path in the repo so they are checked out before `wails build -nsis` runs.
**Warning signs:** CI runner cannot find `project.nsi`; installer built without PATH registration.

### Pitfall 4: PATH Change Not Visible in Current Terminal Session
**What goes wrong:** User installs StorCat, immediately opens a terminal, `storcat version` fails.
**Why it happens:** `SendMessage ${HWND_BROADCAST}` notifies running applications to re-read environment, but some terminals (particularly cmd.exe) opened before the install won't update. Terminals opened AFTER the install will have the new PATH.
**How to avoid:** This is expected Windows behavior. The success criterion specifies "open a new terminal" — this is correct. Document it in install completion notes.
**Warning signs:** Testers reporting `storcat version` fails when they run it in the SAME terminal session that watched the installer run.

### Pitfall 5: EnVar.dll Not Found at makensis Compile Time
**What goes wrong:** `makensis` errors: "Plugin not found: EnVar" or similar. The installer compiles without PATH support silently if `!addplugindir` points to wrong path.
**Why it happens:** NSIS plugin directories are relative to the script file location, not the working directory. `!addplugindir "Plugins"` resolves relative to `project.nsi`'s location.
**How to avoid:** Use `!addplugindir "${__FILEDIR__}\Plugins"` (absolute path based on script file location) or verify the DLL is in `build/windows/installer/Plugins/x86-unicode/EnVar.dll`.
**Warning signs:** No compile error but installer doesn't add PATH; NSIS compile warnings about unknown symbols.

## Code Examples

### Complete Custom project.nsi
```nsis
; Source: Wails v2.10.2 default template + EnVar PATH additions
; Wails v2 will use this file instead of the embedded default
; because buildassets.ReadFile() preserves local files if they exist.

Unicode true

!include "wails_tools.nsh"

VIProductVersion "${INFO_PRODUCTVERSION}.0"
VIFileVersion    "${INFO_PRODUCTVERSION}.0"
VIAddVersionKey "CompanyName"     "${INFO_COMPANYNAME}"
VIAddVersionKey "FileDescription" "${INFO_PRODUCTNAME} Installer"
VIAddVersionKey "ProductVersion"  "${INFO_PRODUCTVERSION}"
VIAddVersionKey "FileVersion"     "${INFO_PRODUCTVERSION}"
VIAddVersionKey "LegalCopyright"  "${INFO_COPYRIGHT}"
VIAddVersionKey "ProductName"     "${INFO_PRODUCTNAME}"

ManifestDPIAware true

!include "MUI.nsh"

!define MUI_ICON "..\icon.ico"
!define MUI_UNICON "..\icon.ico"
!define MUI_FINISHPAGE_NOAUTOCLOSE
!define MUI_ABORTWARNING

!insertmacro MUI_PAGE_WELCOME
!insertmacro MUI_PAGE_DIRECTORY
!insertmacro MUI_PAGE_INSTFILES
!insertmacro MUI_PAGE_FINISH

!insertmacro MUI_UNPAGE_INSTFILES

!insertmacro MUI_LANGUAGE "English"

Name "${INFO_PRODUCTNAME}"
OutFile "..\..\bin\${INFO_PROJECTNAME}-${ARCH}-installer.exe"
InstallDir "$PROGRAMFILES64\${INFO_COMPANYNAME}\${INFO_PRODUCTNAME}"
ShowInstDetails show

; EnVar plugin directory (relative to this script file)
!addplugindir "${__FILEDIR__}\Plugins"

Function .onInit
   !insertmacro wails.checkArchitecture
FunctionEnd

Section
    !insertmacro wails.setShellContext

    !insertmacro wails.webview2runtime

    SetOutPath $INSTDIR

    !insertmacro wails.files

    CreateShortcut "$SMPROGRAMS\${INFO_PRODUCTNAME}.lnk" "$INSTDIR\${PRODUCT_EXECUTABLE}"
    CreateShortCut "$DESKTOP\${INFO_PRODUCTNAME}.lnk" "$INSTDIR\${PRODUCT_EXECUTABLE}"

    !insertmacro wails.associateFiles
    !insertmacro wails.associateCustomProtocols

    !insertmacro wails.writeUninstaller

    ; Add install directory to system PATH (requires admin, set by wails_tools.nsh)
    EnVar::SetHKLM
    EnVar::AddValue "PATH" "$INSTDIR"
    Pop $0  ; 0=success, non-zero=error (idempotent: no-op if already present)

    ; Notify running processes of environment change
    SendMessage ${HWND_BROADCAST} ${WM_SETTINGCHANGE} 0 "STR:Environment" /TIMEOUT=5000
SectionEnd

Section "uninstall"
    !insertmacro wails.setShellContext

    ; Remove install directory from system PATH
    EnVar::SetHKLM
    EnVar::DeleteValue "PATH" "$INSTDIR"
    Pop $0  ; 0=success

    SendMessage ${HWND_BROADCAST} ${WM_SETTINGCHANGE} 0 "STR:Environment" /TIMEOUT=5000

    RMDir /r "$AppData\${PRODUCT_EXECUTABLE}"
    RMDir /r $INSTDIR

    Delete "$SMPROGRAMS\${INFO_PRODUCTNAME}.lnk"
    Delete "$DESKTOP\${INFO_PRODUCTNAME}.lnk"

    !insertmacro wails.unassociateFiles
    !insertmacro wails.unassociateCustomProtocols

    !insertmacro wails.deleteUninstaller
SectionEnd
```

### WinGet Manifest with Commands Field (Optional Enhancement)
```yaml
# scottkw.StorCat.installer.yaml
PackageIdentifier: scottkw.StorCat
PackageVersion: 2.3.0
InstallerType: nullsoft
Commands:
- storcat
Installers:
- Architecture: x64
  InstallerUrl: https://github.com/scottkw/storcat/releases/download/v2.3.0/StorCat-v2.3.0-windows-amd64-installer.exe
  InstallerSha256: <sha256>
ManifestType: installer
ManifestVersion: 1.6.0
```

Note: The `Commands` field informs WinGet of available CLI commands for discoverability. For `nullsoft` type, PATH is handled by the NSIS installer itself — WinGet does not add nullsoft install locations to PATH automatically. The `Commands` field is metadata only.

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| `EnvVarUpdate` macro | EnVar plugin | NSIS community consensus circa 2015-2020 | Eliminates PATH corruption on machines with long PATH values |
| `WriteEnvStr` / `DeleteEnvStr` | EnVar plugin | Same era | EnVar handles REG_EXPAND_SZ correctly; older macros do not |
| Default Wails installer (no PATH) | Custom `project.nsi` in repo | v2.10.2 supports local override | Wails v2 preserves `project.nsi` if it exists — no patching of Wails source needed |

**Deprecated/outdated:**
- `EnvVarUpdate.nsh` macro: NSIS wiki explicitly says "Do NOT use this function to update %PATH%, use the EnVar_plug-in instead."
- `WriteEnvStr` / `DeleteEnvStr`: Obsoleted by EnVar for path-list variables.

## Open Questions

1. **EnVar.dll binary license and size**
   - What we know: GsNSIS/EnVar is maintained on GitHub, MIT-licensed
   - What's unclear: Exact DLL size (likely < 50KB); whether bundling a binary DLL in git is acceptable for this project
   - Recommendation: Check repo policy; if binary DLLs are unacceptable, add a CI step to download from GitHub releases instead

2. **`__FILEDIR__` availability in NSIS 3.10**
   - What we know: `__FILEDIR__` is a standard NSIS predefined constant available since NSIS 3.x
   - What's unclear: Whether the exact windows-2022 NSIS version (3.10) supports it
   - Recommendation: Use `${__FILEDIR__}` form; if unavailable, fall back to `${NSISDIR}\Plugins\x86-unicode` or install EnVar system-wide in CI

3. **PATH registration visible in System Environment Variables UI**
   - What we know: `EnVar::SetHKLM` writes to `HKLM\SYSTEM\CurrentControlSet\Control\Session Manager\Environment`, which is what the System Properties dialog reads
   - What's unclear: Whether the NSIS admin execution level is definitely triggered when WinGet runs the installer silently
   - Recommendation: WinGet silent installs nullsoft type with `/S` flag; wails_tools.nsh sets `RequestExecutionLevel admin` by default — this should trigger UAC elevation and grant HKLM write access

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|-------------|-----------|---------|----------|
| NSIS (makensis) | `wails build -nsis` | ✓ (CI only) | 3.10 (windows-2022 runner) | None needed — pre-installed |
| Wails v2 | Build | ✓ | v2.10.2 | — |
| Go | Build | ✓ | 1.26.1 | — |
| EnVar.dll | NSIS PATH ops | Needs bundling | Latest GsNSIS/EnVar | CI download step (alternative) |
| makensis (local) | Local dev testing | ✗ (macOS dev machine) | — | Test on CI; or use Parallels/VM |

**Missing dependencies with no fallback:**
- None that block CI execution. NSIS is pre-installed on windows-2022.

**Missing dependencies with fallback:**
- `EnVar.dll`: If bundling binary in git is rejected, add a CI step before `wails build -nsis` to download it from `https://github.com/GsNSIS/EnVar/releases`.
- Local NSIS testing: Not available on macOS dev machine. Testing requires CI run or Windows VM.

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go `testing` package (`go test ./...`) |
| Config file | none (standard Go test discovery) |
| Quick run command | `go test ./... -count=1` |
| Full suite command | `go test ./... -count=1 -race` |

### Phase Requirements → Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| PKG-02 | NSIS script adds $INSTDIR to HKLM PATH | Manual/CI only | `wails build -nsis` then inspect installer | ❌ Wave 0 — NSIS script validation only possible on Windows |
| PKG-04 | `storcat version` works after WinGet install | Manual/E2E | Human UAT on Windows | ❌ Wave 0 — Windows-only E2E |

**Note on automation:** There is no practical automated test for NSIS PATH registration on macOS. Validation is:
1. **Structural**: Verify `project.nsi` contains the correct EnVar calls (grep/diff in CI)
2. **Build verification**: CI `wails build -nsis` succeeds without errors
3. **Manual UAT**: Human verifies PATH on a Windows machine after install

### Sampling Rate
- **Per task commit:** `go test ./... -count=1` (ensures no Go regressions from any Go changes in this phase)
- **Per wave merge:** `go test ./... -count=1 -race`
- **Phase gate:** CI build succeeds + human UAT on Windows machine confirms `storcat version` works in new terminal

### Wave 0 Gaps
- [ ] No new Go test files needed — this phase is NSIS script + binary file only
- [ ] Windows manual UAT checklist: run installer, open new terminal, `storcat version`, check System Environment Variables dialog

## Sources

### Primary (HIGH confidence)
- Wails v2.10.2 source: `/Users/ken/go/pkg/mod/github.com/wailsapp/wails/v2@v2.10.2/pkg/buildassets/buildassets.go` — confirmed `ReadFile` preserves local `project.nsi`
- Wails v2.10.2 source: `/Users/ken/go/pkg/mod/github.com/wailsapp/wails/v2@v2.10.2/pkg/buildassets/build/windows/installer/project.nsi` — default template verified
- Wails v2.10.2 source: `/Users/ken/go/pkg/mod/github.com/wailsapp/wails/v2@v2.10.2/pkg/buildassets/build/windows/installer/wails_tools.nsh` — confirmed always regenerated
- Wails v2.10.2 source: `/Users/ken/go/pkg/mod/github.com/wailsapp/wails/v2@v2.10.2/pkg/commands/build/nsis_installer.go` — confirmed build flow
- StorCat `release.yml` — confirmed `wails build -clean -platform windows/amd64 -windowsconsole -nsis` is the exact build command
- StorCat `main.go` — confirmed `storcat version` subcommand dispatch exists

### Secondary (MEDIUM confidence)
- [NSIS EnVar plug-in](https://nsis.sourceforge.io/EnVar_plug-in) — official NSIS wiki recommendation to use EnVar over EnvVarUpdate for PATH
- [GsNSIS/EnVar GitHub](https://github.com/GsNSIS/EnVar) — plugin source and releases
- [NSIS Path Manipulation](https://nsis.sourceforge.io/Path_Manipulation) — explicit "Do NOT use for %PATH%" warning on EnvVarUpdate
- [actions/runner-images #7740](https://github.com/actions/runner-images/issues/7740) — NSIS large-strings request; confirmed standard NSIS (1024-byte limit) is on runners
- [WM_SETTINGCHANGE Microsoft Docs](https://learn.microsoft.com/en-us/windows/win32/winmsg/wm-settingchange) — broadcast requirement after registry PATH change

### Tertiary (LOW confidence)
- WebSearch results indicating NSIS 3.10 on windows-2022 runner — could not verify exact version from official runner image spec; version matters only for `__FILEDIR__` support which is NSIS 3.x standard

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — verified against Wails v2.10.2 source code directly on local machine
- Architecture: HIGH — `buildassets.go` source code proves the local-file-override behavior
- Pitfalls: HIGH — EnVar vs EnvVarUpdate based on NSIS official documentation; wails_tools.nsh overwrite verified in source
- WinGet interaction: MEDIUM — WinGet's handling of nullsoft PATH is based on web research; silent install behavior verified by reading wails_tools.nsh `RequestExecutionLevel admin`

**Research date:** 2026-03-28
**Valid until:** 2026-06-28 (stable ecosystem — Wails v2 NSIS, NSIS plugin API; changes unlikely)
