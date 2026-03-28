# Project Research Summary

**Project:** StorCat v2.3.0 — Code Signing, Notarization, and Package Manager CLI PATH
**Domain:** Desktop app distribution hardening — macOS Developer ID signing + notarization, Windows Authenticode signing, Homebrew and WinGet CLI PATH setup
**Researched:** 2026-03-27
**Confidence:** HIGH

## Executive Summary

StorCat v2.3.0 is an additive milestone on an already-shipping Go/Wails desktop app. The existing CI/CD pipeline produces unsigned macOS DMGs and unsigned Windows NSIS installers. On current macOS Sequoia (15+), unsigned apps are effectively uninstallable without a multi-step System Settings workaround that most users abandon. On Windows, unsigned installers trigger a SmartScreen "Windows protected your PC" block that stops the majority of cautious users. Solving both requires obtaining platform-specific code signing certificates, adding four sequential signing steps to the macOS CI job, one signing step to the Windows CI job, and modifying two distribution files (the Homebrew cask template and the NSIS installer script) to expose the `storcat` binary on PATH after package manager installation.

The recommended implementation approach is native tooling over third-party actions. On macOS: `xcrun codesign` + `xcrun notarytool` + `xcrun stapler` (all pre-installed on `macos-14` runners), with certificate import handled via `apple-actions/import-codesign-certs@v6`. On Windows: native PowerShell + `signtool.exe` (pre-installed on `windows-2022` runners) with an OV certificate stored as a base64-encoded PFX GitHub secret. The Homebrew CLI PATH fix is one line (`binary` stanza in `storcat.rb.template`). The Windows CLI PATH fix requires a custom NSIS script with `EnvVarUpdate` or `WriteRegExpandStr` to add the install directory to system PATH. No new directories, no source code changes, and no structural changes to `release.yml` are needed — only additions within existing jobs.

The primary risks are ordering-based and silent: codesign hangs indefinitely in CI without the `security set-key-partition-list` ACL step; notarization of a DMG fails if the `.app` inside was not signed first; and WinGet manifests will silently contain wrong SHA256 hashes if computed before signing. All of these are well-documented patterns with known fixes. The sequencing constraint (codesign → hdiutil → notarytool → stapler) is strict and must be treated as a hard ordering rule in the workflow. The one true decision point is Windows certificate type: Microsoft Azure Trusted Signing ($9.99/month) provides instant SmartScreen reputation; a traditional OV certificate is cheaper annually but subjects users to a 2–8 week SmartScreen warning window. This decision should be made before any Windows signing work begins.

## Key Findings

### Recommended Stack

The milestone requires no new frameworks. All tooling is either pre-installed on GitHub Actions runners or available via single-action imports. The `apple-actions/import-codesign-certs@v6` action handles the complex multi-command macOS keychain setup that most signing failures trace back to. Windows signing uses `signtool.exe` found dynamically (not hardcoded path) to survive runner image updates. The Homebrew `binary` stanza is the correct, supported mechanism for putting a Wails `.app` binary on PATH — it symlinks from `$(brew --prefix)/bin/storcat` into `StorCat.app/Contents/MacOS/StorCat`. The NSIS `EnvVarUpdate.nsh` macro is safer than direct registry writes for Windows PATH management because it handles deduplication and size limits. The deprecated `gon` tool (referenced in old Wails docs) must not be used — Apple sunset `altool` in November 2023.

**Core technologies:**
- `apple-actions/import-codesign-certs@v6` (SHA: `b610f78`): CI keychain setup — official Apple-maintained action; handles ACL configuration that prevents codesign hangs
- `xcrun codesign` (built-in, `macos-14`): App bundle signing — native Apple tooling, no external dependency; `--options runtime` flag required for notarization
- `xcrun notarytool` (built-in, Xcode 13+): Notarization submission — replaces deprecated `altool`; `--wait --timeout 30m` prevents silent job hangs
- `xcrun stapler` (built-in): Notarization ticket embedding — enables offline Gatekeeper verification; most commonly skipped step in tutorials
- `signtool.exe` (built-in, `windows-2022`): Windows Authenticode signing — pre-installed on all Windows runners; no third-party action needed
- `EnvVarUpdate.nsh` (NSIS): Windows PATH registration — must be downloaded to `build/windows/installer/` and included in the NSIS script

**Critical version note:** `apple-actions/import-codesign-certs@v6` targets Node 24 and works on current `macos-14` runners. `dlemstra/code-sign-action` was archived October 2025 — do not use.

### Expected Features

**Must have (table stakes — v2.3.0):**
- macOS DMG: signed (`--options runtime`) + notarized + stapled — required for Gatekeeper acceptance on macOS Sequoia
- Windows NSIS installer: signed with RSA certificate — required to suppress or reduce SmartScreen blocking
- Homebrew cask: `binary` stanza adding `storcat` to PATH — users expect CLI to work immediately after `brew install --cask storcat`
- WinGet/NSIS: install directory added to system PATH — users expect `storcat` to work in any new terminal after `winget install scottkw.StorCat`
- GitHub Actions signing environment: secrets scoped to `release` environment, not all workflows — security baseline
- 9 new GitHub secrets: 7 macOS (cert p12, cert password, keychain password, Apple ID, app-specific password, team ID, cert identity string) + 2 Windows (PFX base64, PFX password)

**Should have (competitive — v2.3.x after validation):**
- App Store Connect API key auth for notarytool (more robust than Apple ID + app-specific password; not subject to 2FA friction)
- Azure Trusted Signing if US/Canadian business eligibility confirmed (instant SmartScreen reputation vs. 2–8 week wait for OV cert)
- Notarization retry + timeout handling (hardened against Apple notarization service slowdowns)

**Defer (v3+):**
- Code signing certificate expiry monitoring via GitHub Actions scheduled check
- Automated certificate renewal runbook via CI (mechanical steps; human approval still required by Apple)
- Mac App Store distribution (requires sandboxing incompatible with Wails filesystem access patterns)

### Architecture Approach

The v2.3.0 architecture is purely additive CI/CD changes. Three files change: `release.yml` gains 4 steps in `build-macos` and 1 step in `build-windows`; `packaging/homebrew/storcat.rb.template` gains one `binary` stanza line; `build/windows/installer.nsi` is created as a new custom NSIS script overriding Wails' generated default. No Go source, TypeScript source, or distribution workflow logic changes. The four macOS signing steps are strictly sequential (cannot be parallelized) and must run in this exact order: import certificate → codesign → hdiutil → notarytool submit → stapler staple. The Windows signing step must complete before `upload-artifact` runs so that the SHA256 computed by `distribute.yml` (from the GitHub release asset) matches the signed binary.

**Major components:**
1. `build-macos` job in `release.yml` — gains 4 steps: cert import, codesign (with entitlements), notarytool, stapler; ordering is a hard constraint
2. `build-windows` job in `release.yml` — gains 1 PowerShell step: PFX decode, signtool sign both binaries, PFX delete; must precede artifact upload
3. `packaging/homebrew/storcat.rb.template` — gains 1 line: `binary "#{appdir}/StorCat.app/Contents/MacOS/StorCat", target: "storcat"`
4. `build/windows/installer.nsi` — new file: custom NSIS script with `EnvVarUpdate` PATH registration; Wails respects this path over its generated default

### Critical Pitfalls

1. **Codesign hangs silently in CI** — Missing `security set-key-partition-list -S apple-tool:,apple:,codesign: -s -k "$CI_KEYCHAIN_PWD" build.keychain` after certificate import. The step runs indefinitely with no output until the job times out. Use `apple-actions/import-codesign-certs` which handles this correctly, or run all 5 keychain setup commands explicitly. Verify by checking that the codesign step completes in under 2 minutes.

2. **Wrong signing order invalidates signature** — Signing the `.app` after `hdiutil create` or signing the DMG instead of the `.app` inside it causes notarization to reject with "invalid signature." Correct order is immovable: codesign `.app` → `hdiutil create` DMG → `notarytool submit` DMG → `stapler staple` DMG. Add `codesign --verify --deep --strict StorCat.app` as a gate step after signing.

3. **Missing entitlements crash the Wails WebView** — Signing with `--options runtime` (required for notarization) without an entitlements plist causes the hardened runtime to block JIT and dynamic memory needed by WKWebView. The app passes notarization but shows a blank window or crashes on user machines. The existing `build/entitlements.mac.plist` from the Electron era must be ported and passed via `--entitlements` flag. Minimum: `com.apple.security.cs.allow-jit` and `com.apple.security.cs.allow-unsigned-executable-memory`.

4. **WinGet SHA256 computed before signing** — If the distribute workflow downloads a pre-signing artifact or the hash step runs before signing completes, every WinGet install will fail with "Installer hash does not match." Sign in the build job before `upload-artifact`. The distribute workflow then downloads the already-signed binary from the GitHub release, and its SHA256 is correct.

5. **Windows ECDSA certificate rejected by SmartScreen** — ECDSA code signing certificates are cryptographically valid but Windows Authenticode requires RSA PKCS#1 v1.5. The binary will appear signed but SmartScreen still blocks it. When purchasing or configuring any Windows signing certificate, explicitly select RSA-2048 or RSA-4096 (or `RSA-HSM` in Azure Key Vault — never `EC-HSM`). Verify before any signing infrastructure work begins.

## Implications for Roadmap

The architecture identifies a clear sequential dependency chain with one independent parallel track. macOS signing must precede Homebrew CLI PATH (signed binary is required for macOS 15+ symlink creation). Windows signing must precede WinGet PATH (signed binary should be in place before new NSIS script). Windows signing and macOS signing are independent of each other and can proceed in parallel. All phases depend on secrets and certificates being configured first.

### Phase 1: Secrets and Certificate Procurement

**Rationale:** No signing work can begin without certificates and GitHub secrets. This is a manual human-performed phase with no code changes. It must complete before any other phase starts. Certificate procurement (especially Windows OV cert from a CA) can take 1–3 business days for verification. The Windows certificate type decision (Azure Trusted Signing vs. OV cert) must be made here before any infrastructure is built.
**Delivers:** All 9 GitHub secrets configured; `release` environment created with reviewer protection; signing certificates in hand and validated locally (`security find-identity` confirms Developer ID on macOS; `signtool verify` confirms RSA cert on Windows); Windows certificate type decision locked in
**Addresses:** "Signing credentials in GitHub Actions secrets" and "secrets scoped to release environment" table stakes features
**Avoids:** P12 password mismatch pitfall (verify export password matches secret at import time); ECDSA cert pitfall (Windows cert type confirmed RSA before any infrastructure work)

### Phase 2: macOS Signing and Notarization

**Rationale:** macOS has the most complex signing chain (4 sequential steps) and the most pitfalls. Getting it working first, independently from Windows, reduces debugging surface area. The entitlements plist port from Electron is a prerequisite only for this phase.
**Delivers:** Signed, notarized, and stapled macOS DMG produced by CI on every release tag; macOS users install without Gatekeeper blocking; `spctl -vvv --assess --type exec StorCat.app` returns "accepted"
**Uses:** `apple-actions/import-codesign-certs@v6`, `xcrun codesign`, `xcrun notarytool --wait --timeout 30m`, `xcrun stapler`; `build/entitlements.mac.plist` ported from Electron era
**Avoids:** All macOS-specific pitfalls — keychain ACL (Pitfall 1), signing order (Pitfall 2), missing entitlements (Pitfall 3), notarization timeout (Pitfall 4), P12 password mismatch (Pitfall 5)
**Research flag:** Well-documented patterns from official Apple docs and multiple high-quality implementation blogs. No additional research needed. The Wails signing guide returned 403 during research — verify current Wails entitlements requirements against Wails GitHub repo before signing.

### Phase 3: Windows Signing

**Rationale:** Independent of Phase 2 and can be developed in parallel, but sequenced after because the Windows cert type decision (made in Phase 1) must be confirmed. Simpler than macOS signing (one PowerShell step). Critical: signing must happen in the build job before `upload-artifact` to ensure WinGet manifests receive the correct SHA256.
**Delivers:** Signed Windows NSIS installer and portable `.exe` produced by CI; reduced or eliminated SmartScreen blocking; `signtool verify /v /pa StorCat-*-installer.exe` returns "Successfully verified"; WinGet SHA256 hashes correct
**Uses:** Native PowerShell + `signtool.exe` dynamically located on `windows-2022` runner; OV certificate PFX or Azure Trusted Signing action; timestamp server (Sectigo or DigiCert)
**Avoids:** ECDSA cert pitfall (Pitfall 6 — confirmed RSA in Phase 1), WinGet SHA256 mismatch (Pitfall 7 — sign before upload-artifact), Azure credential expiration (Pitfall 12 — use OIDC if Azure route chosen)
**Research flag:** Well-documented. If Azure Trusted Signing chosen: `azure/trusted-signing-action` with OIDC integration is the pattern. Confirm US/Canada eligibility in Phase 1 before choosing this path.

### Phase 4: Homebrew CLI PATH

**Rationale:** Depends on Phase 2 — macOS 15+ Gatekeeper blocks symlink creation from unsigned binaries in Homebrew casks. Must follow Phase 2 not just in planning but in actual release timing. One-line code change but requires a real published release to validate end-to-end.
**Delivers:** `brew install --cask storcat` results in `storcat` being immediately available in PATH; `storcat version` works from any new terminal after install
**Uses:** Homebrew `binary` stanza in `storcat.rb.template`; `#{appdir}` variable expanding to `/Applications`; `target: "storcat"` lowercase rename
**Avoids:** Binary stanza wrong path pitfall (Pitfall 9 — verify exact binary path via `ls build/bin/StorCat.app/Contents/MacOS/` before committing stanza); Phase ordering violation (binary stanza must not go live until signed DMG is in production)
**Research flag:** Homebrew Cask Cookbook is the authoritative source. Pattern fully documented. No research needed. Verification requires a fresh macOS install — do not test on a developer machine with existing PATH entries.

### Phase 5: Windows CLI PATH via NSIS

**Rationale:** Depends on Phase 3 — signed installer should be in place before PATH registration is added. Also requires a real Windows test environment (fresh VM or CI runner) to validate that PATH is set correctly in a new terminal session. Highest-friction verification in this milestone.
**Delivers:** `winget install scottkw.StorCat` results in `storcat` being available in any new terminal session; `storcat version` works from Command Prompt and PowerShell
**Uses:** Custom `build/windows/installer.nsi` with `EnvVarUpdate.nsh` macro; `EnvVarUpdate.nsh` downloaded to `build/windows/installer/`; HKLM PATH registration with `WM_SETTINGCHANGE` broadcast
**Avoids:** NSIS missing PATH configuration (Pitfall 10); stale PATH verification on developer machine (must test on fresh Windows VM)
**Research flag:** NSIS path manipulation is well-documented. `EnvVarUpdate.nsh` must be manually downloaded from nsis.sourceforge.io and committed to `build/windows/installer/`. Verify Wails v2.10.2 respects `build/windows/installer.nsi` as the NSIS script override path before building the custom script.

### Phase Ordering Rationale

- Phase 1 is a hard prerequisite for all other phases — no code can run without certificates and secrets
- Phases 2 and 3 are independent of each other and can be developed in parallel if two people are working; Phase 2 is more complex with more failure modes so it gets priority attention in single-developer execution
- Phase 4 must follow Phase 2 because macOS 15+ Gatekeeper blocks cask binary symlinks from unsigned apps — this is not just a best-practice ordering, it is a technical requirement
- Phase 5 should follow Phase 3 so the PATH-enabled installer is also signed, but the NSIS script change is technically independent
- The WinGet SHA256 constraint makes the signing position within the `build-windows` job a strict rule: sign before `upload-artifact`, always

### Research Flags

Phases with standard, well-documented patterns — research-phase not needed:
- **Phase 2 (macOS signing):** Apple official docs, GitHub official docs, and multiple high-quality implementation posts all agree on the exact sequence
- **Phase 3 (Windows signing):** OV cert + signtool pattern is thoroughly documented; Azure Trusted Signing path equally so
- **Phase 4 (Homebrew PATH):** Single official source (Homebrew Cask Cookbook) is definitive
- **Phase 5 (WinGet NSIS PATH):** NSIS `EnvVarUpdate` is well-documented; only mechanical gap is downloading `EnvVarUpdate.nsh`

One decision (not research gap) to resolve before implementation:
- **Windows signing approach (Phase 1):** Azure Trusted Signing (instant SmartScreen reputation, $9.99/month, US/Canada businesses only as of April 2025) vs. traditional OV cert (2–8 week SmartScreen warning window, ~$200–300/year, available everywhere). This is a business and budget decision, not a technical research gap.

## Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| Stack | HIGH | Official Apple docs, official GitHub Actions docs, official Homebrew Cask Cookbook; deprecated tools clearly identified (`gon`, `dlemstra/code-sign-action`) |
| Features | HIGH | Apple Gatekeeper Sequoia changes well-documented; SmartScreen reputation behavior confirmed by DigiCert and Microsoft; Homebrew binary stanza from official cookbook |
| Architecture | HIGH | Derived directly from analysis of existing `release.yml`, `distribute.yml`, and `storcat.rb.template`; patterns from GitHub official docs and multiple implementation references |
| Pitfalls | HIGH | 12 specific pitfalls with exact error messages, root causes, and prevention steps; sourced from official docs and multiple implementation post-mortems |

**Overall confidence:** HIGH

### Gaps to Address

- **Wails entitlements plist validation:** The existing `build/entitlements.mac.plist` from the Electron era needs to be verified against Wails-specific runtime requirements before signing. The Wails signing guide returned a 403 during research. Before Phase 2 begins, confirm current Wails entitlements requirements against the Wails GitHub repo. The minimum set (`allow-jit`, `allow-unsigned-executable-memory`, `network.client`) is established from community sources and is very likely correct.

- **NSIS script Wails override path:** Wails v2 respects `build/windows/installer.nsi` as a custom script override when `-nsis` is used. Verify this against Wails v2.10.2 specifically before Phase 5, since Wails versions occasionally change internal template paths.

- **Azure Trusted Signing eligibility:** Geographic restriction to US and Canadian businesses as of April 2025. Confirm eligibility before planning Phase 3 around Azure; if ineligible, use OV cert route from the start.

- **Notarization timing baseline:** Apple notarization typically takes 10–120 seconds but can spike. The `--timeout 30m` flag is a reasonable ceiling. Monitor the first several production notarizations to confirm typical timing before relying on the release pipeline for time-sensitive releases.

## Sources

### Primary (HIGH confidence)
- [Apple-Actions/import-codesign-certs releases](https://github.com/Apple-Actions/import-codesign-certs/releases) — v6.0.0, SHA `b610f78`, Dec 2024
- [GitHub Docs: Installing Apple certificate on macOS runners](https://docs.github.com/actions/use-cases-and-examples/deploying/installing-an-apple-certificate-on-macos-runners-for-xcode-development) — keychain setup sequence
- [Apple Developer: Developer ID overview](https://developer.apple.com/developer-id/) — signing requirements
- [Homebrew Cask Cookbook](https://docs.brew.sh/Cask-Cookbook) — `binary` stanza syntax, `#{appdir}`, `target:`
- [Federico Terzi: macOS signing + notarization via GitHub Actions](https://federicoterzi.com/blog/automatic-code-signing-and-notarization-for-macos-apps-using-github-actions/) — complete workflow with 7 secrets
- [Melatonin: macOS code signing + notarization in CI](https://melatonin.dev/blog/how-to-code-sign-and-notarize-macos-audio-plugins-in-ci/) — implementation patterns
- [Melatonin: Windows Azure Trusted Signing](https://melatonin.dev/blog/code-signing-on-windows-with-azure-trusted-signing/) — Azure approach
- [Microsoft Azure Artifact Signing](https://azure.microsoft.com/en-us/products/artifact-signing) — product page, geographic restrictions
- [NSIS Path Manipulation](https://nsis.sourceforge.io/Path_Manipulation) — `EnvVarUpdate` macro
- [WinGet PATH discussion](https://github.com/microsoft/winget-cli/discussions/5091) — confirmed WinGet does not manage PATH; delegates to installer
- [dlemstra/code-sign-action archived](https://github.com/dlemstra/code-sign-action) — confirmed archived Oct 2025; do not use

### Secondary (MEDIUM confidence)
- [Federico Terzi: Windows Authenticode via GitHub Actions](https://federicoterzi.com/blog/automatic-codesigning-on-windows-using-github-actions/) — OV PFX + signtool pattern
- [NSIS EnvVarUpdate](https://nsis.sourceforge.io/Environmental_Variables:_append,_prepend,_and_remove_entries) — PATH append/remove entries
- [Windows signtool GitHub Actions (Advanced Installer)](https://www.advancedinstaller.com/code-signing-with-github-actions.html) — PFX decode + signtool pattern
- [Notarizing Apps Distributed as DMGs (Decipher Tools)](https://deciphertools.com/blog/notarizing-dmg/) — DMG-specific notarization sequence
- [Scott Hanselman: Azure Trusted Signing + GitHub Actions](https://www.hanselman.com/blog/automatically-signing-a-windows-exe-with-azure-trusted-signing-dotnet-sign-and-github-actions) — AzureSignTool integration

### Tertiary (LOW confidence)
- [Wails signing guide](https://wails.io/docs/guides/signing/) — returned 403 during research; mentions deprecated `gon`; verify against current Wails docs before implementation

---
*Research completed: 2026-03-27*
*Ready for roadmap: yes*
