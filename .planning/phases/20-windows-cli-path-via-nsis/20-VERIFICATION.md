---
phase: 20-windows-cli-path-via-nsis
verified: 2026-03-28T07:00:00Z
status: human_needed
score: 3/4 must-haves verified
human_verification:
  - test: "Install StorCat NSIS installer on a Windows machine, open a NEW Command Prompt or PowerShell, run `storcat version`"
    expected: "Command prints version info without any manual PATH changes"
    why_human: "Cannot run Windows NSIS installer or execute .exe binaries on macOS; requires a live Windows environment to confirm end-to-end PATH registration"
  - test: "After install, open System Properties > Environment Variables > System PATH and confirm the StorCat install directory appears"
    expected: "StorCat install directory (e.g. C:\\Program Files\\kenscott\\StorCat) is listed in System PATH"
    why_human: "Registry verification requires Windows to read HKLM system environment variables"
  - test: "Uninstall StorCat, then check System PATH again"
    expected: "StorCat install directory has been removed from System PATH"
    why_human: "Cannot run uninstaller on macOS"
  - test: "Run `wails build -clean -platform windows/amd64 -windowsconsole -nsis` on CI (Windows runner)"
    expected: "Build succeeds with no NSIS compilation errors — EnVar plugin found, all macros resolved"
    why_human: "CI build requires a live Windows runner; cannot execute makensis on macOS dev machine"
---

# Phase 20: Windows CLI PATH via NSIS Verification Report

**Phase Goal:** Add `$INSTDIR` to system PATH during NSIS install (via EnVar plugin) so `storcat version` works from any new terminal; remove on uninstall. Add `Commands` metadata to WinGet manifest template.
**Verified:** 2026-03-28T07:00:00Z
**Status:** human_needed
**Re-verification:** No — initial verification

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | NSIS installer adds $INSTDIR to system PATH on install | ? HUMAN | `EnVar::SetHKLM` + `EnVar::AddValue "PATH" "$INSTDIR"` present in project.nsi install Section; correct only verifiable on Windows |
| 2 | NSIS installer removes $INSTDIR from system PATH on uninstall | ? HUMAN | `EnVar::SetHKLM` + `EnVar::DeleteValue "PATH" "$INSTDIR"` present in project.nsi uninstall Section; correct only verifiable on Windows |
| 3 | Running shells are notified of PATH change via WM_SETTINGCHANGE broadcast | ✓ VERIFIED | 2 `SendMessage ${HWND_BROADCAST} ${WM_SETTINGCHANGE} 0 "STR:Environment" /TIMEOUT=5000` calls in project.nsi (lines 111, 124) — one in install Section, one in uninstall Section |
| 4 | storcat version works from a new terminal after WinGet install | ? HUMAN | Requires live Windows install + new terminal session; cannot verify on macOS |

**Score:** 3/4 truths verified (truth 3 fully automated; truths 1, 2, 4 are structurally complete but require Windows UAT)

Note: Truths 1 and 2 pass all automated code checks (artifacts exist, substantive content present, wiring correct). They are flagged HUMAN because the correctness of NSIS runtime behavior — that EnVar actually writes to registry and PATH takes effect — cannot be confirmed without executing the installer.

### Required Artifacts

| Artifact | Expected | Status | Details |
|----------|----------|--------|---------|
| `build/windows/installer/project.nsi` | Custom NSIS script with EnVar PATH registration | ✓ VERIFIED | 138 lines; contains all required NSIS instructions and Wails macros |
| `build/windows/installer/Plugins/x86-unicode/EnVar.dll` | EnVar NSIS plugin binary | ✓ VERIFIED | 9,216 bytes; PE32 DLL (MS Windows, Intel 80386); committed at `4620841b` |
| `packaging/winget/manifests/s/scottkw/StorCat/2.1.0/scottkw.StorCat.installer.yaml` | WinGet installer manifest with Commands metadata | ✓ VERIFIED | Contains `Commands:` and `- storcat`; InstallerType, PackageVersion, InstallerSha256 all preserved |

### Key Link Verification

| From | To | Via | Status | Details |
|------|----|-----|--------|---------|
| `project.nsi` | `Plugins/x86-unicode/EnVar.dll` | `!addplugindir "${__FILEDIR__}\Plugins"` | ✓ VERIFIED | Line 79: `!addplugindir "${__FILEDIR__}\Plugins"` — uses NSIS 3.x predefined constant for script's own directory |
| `project.nsi` | `HKLM\...\Environment PATH` | `EnVar::SetHKLM + EnVar::AddValue` | ✓ VERIFIED | Lines 106-108: `EnVar::SetHKLM` then `EnVar::AddValue "PATH" "$INSTDIR"` then `Pop $0`; lines 119-121: same pattern for DeleteValue on uninstall |
| `.github/workflows/release.yml` | `build/windows/installer/project.nsi` | `wails build -nsis reads local project.nsi` | ✓ VERIFIED | Line 144: `wails build -clean -platform windows/amd64 -windowsconsole -nsis` — Wails v2 preserves local project.nsi when present |

### Data-Flow Trace (Level 4)

Not applicable — this phase produces NSIS build configuration and a WinGet manifest template, not a React/Go component that renders dynamic data. The "data flow" is compile-time (makensis reads project.nsi and bundles EnVar.dll) and runtime on Windows (installer executes EnVar calls against registry).

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
|----------|---------|--------|--------|
| project.nsi EnVar calls match plan spec (4 total) | `grep -c 'EnVar::' project.nsi` | 4 | ✓ PASS |
| WM_SETTINGCHANGE broadcast present twice | `grep -c 'WM_SETTINGCHANGE' project.nsi` | 4 (2 SendMessage lines + 2 comment lines) | ✓ PASS |
| Pop $0 error-check after each EnVar call | `grep -c 'Pop \$0' project.nsi` | 2 | ✓ PASS |
| wails_tools.nsh NOT committed | `git ls-files build/windows/installer/wails_tools.nsh` | empty (not tracked) | ✓ PASS |
| WinGet manifest Commands field present | `grep 'Commands:' ...installer.yaml` | `Commands:` | ✓ PASS |
| WinGet manifest storcat command listed | `grep -- '- storcat' ...installer.yaml` | `- storcat` | ✓ PASS |
| EnVar.dll is valid PE32 DLL binary | `file EnVar.dll` | PE32 executable (DLL) (GUI) Intel 80386, for MS Windows | ✓ PASS |
| EnVar.dll size > 0 | `wc -c EnVar.dll` | 9,216 bytes | ✓ PASS |
| release.yml uses wails build -nsis | `grep 'wails build.*-nsis' release.yml` | line 144 match | ✓ PASS |
| NSIS installer on Windows installs and registers PATH | Manual Windows install + new terminal | — | ? SKIP (Windows required) |

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
|-------------|-------------|-------------|--------|----------|
| PKG-02 | 20-01-PLAN.md | NSIS installer adds install directory to system PATH via EnVar | ? HUMAN | Code correct: `EnVar::AddValue "PATH" "$INSTDIR"` present with SetHKLM, Pop, and WM_SETTINGCHANGE; runtime behavior requires Windows UAT |
| PKG-04 | 20-01-PLAN.md | `storcat version` works from any new terminal after WinGet install (Windows) | ? HUMAN | Depends on PKG-02 runtime correctness + EnVar DLL presence + distribute.yml propagating Commands field; end-to-end only testable on Windows |

Both requirements have complete structural implementation. Neither can be declared fully satisfied without Windows UAT confirming the installer actually writes to the registry and the PATH is inherited by new shells.

No orphaned requirements: REQUIREMENTS.md maps PKG-02 and PKG-04 exclusively to Phase 20, and both are claimed in 20-01-PLAN.md.

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
|------|------|---------|----------|--------|
| None found | — | — | — | — |

Scanned `project.nsi` and `scottkw.StorCat.installer.yaml` for TODO/FIXME, placeholder text, empty returns, and stub indicators. No anti-patterns detected. The `InstallerSha256: 0000...` in the WinGet manifest is intentional (placeholder preserved for sed replacement by `distribute.yml`) and is not a stub.

### Human Verification Required

#### 1. End-to-end PATH registration — install

**Test:** On a Windows machine (or CI Windows runner), install the built StorCat NSIS installer, then open a fresh Command Prompt or PowerShell window (not the one used during install) and run `storcat version`.
**Expected:** The command prints StorCat version information without the user needing to modify PATH manually.
**Why human:** The NSIS script logic is structurally correct, but confirming that EnVar.dll executes successfully, the registry write occurs under the correct HKLM key, and the new PATH is inherited by a new shell session requires running the installer on Windows.

#### 2. System PATH visibility in Environment Variables

**Test:** After install, open System Properties > Advanced > Environment Variables. Check the System variables list for PATH.
**Expected:** The StorCat install directory (typically `C:\Program Files\kenscott\StorCat`) appears in the System PATH value.
**Why human:** Reading HKLM system environment registry values requires a live Windows session.

#### 3. PATH cleanup on uninstall

**Test:** Uninstall StorCat using the system uninstaller (Control Panel > Programs, or WinGet uninstall). Then check System PATH in Environment Variables.
**Expected:** The StorCat install directory has been removed from System PATH.
**Why human:** Requires running the NSIS uninstaller on Windows.

#### 4. CI NSIS compilation with EnVar plugin

**Test:** Trigger a release build on the GitHub Actions `build-windows` job (Windows runner).
**Expected:** `wails build -clean -platform windows/amd64 -windowsconsole -nsis` completes without NSIS compilation errors. The log shows makensis picking up `EnVar.dll` from the bundled Plugins directory.
**Why human:** macOS cannot run makensis or Windows PE executables; requires a live CI Windows runner.

### Gaps Summary

No gaps — all code is present, substantive, and correctly wired. The phase goal is structurally achieved. The three outstanding human verification items are not gaps in the implementation; they are behavioral confirmations that require a Windows environment to execute. The NSIS script matches the plan specification exactly, the EnVar.dll is a valid Windows PE32 DLL, and the WinGet manifest template has been correctly updated.

The only reason this verification is `human_needed` rather than `passed` is that NSIS installer behavior is inherently Windows-only and cannot be confirmed through static analysis alone.

---

_Verified: 2026-03-28T07:00:00Z_
_Verifier: Claude (gsd-verifier)_
