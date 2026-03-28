# Feature Research

**Domain:** Code signing (macOS + Windows) and package manager CLI PATH setup for a Go/Wails desktop + CLI app
**Researched:** 2026-03-27
**Confidence:** HIGH (macOS signing/notarization), HIGH (Windows Authenticode), HIGH (Homebrew binary stanza), MEDIUM (WinGet PATH/NSIS integration)

---

## Context

StorCat v2.3.0 is adding code signing and CLI PATH setup to an already-shipping app with
automated CI/CD. The existing pipeline (v2.2.x) produces unsigned macOS DMGs and unsigned
Windows NSIS installers. Homebrew installs the `.app` bundle but does not put `storcat` on
PATH. WinGet installs the app but PATH depends on the NSIS installer.

This document covers **only the new features** needed for v2.3.0.

---

## Feature Landscape

### Table Stakes (Users Expect These)

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| macOS signed + notarized DMG | macOS Sequoia removed the Control-click Gatekeeper bypass. Users now need System Settings > Privacy & Security to run unsigned apps — a multi-step process most users abandon. Notarized apps open with a single first-launch prompt. | HIGH | Requires Apple Developer Program ($99/yr). Certificate export, base64 GitHub secret, temporary CI keychain, `codesign --sign --options runtime`, `xcrun notarytool submit`, `xcrun stapler staple`. |
| Windows signed NSIS installer | Unsigned installers trigger "Windows protected your PC" (SmartScreen). Users must click "More info" then "Run anyway" — a flow that looks malicious. Many users stop. Building reputation organically requires ~15,000 safe downloads (as of 2025). | HIGH | Requires code signing certificate. Two viable options: Azure Artifact Signing ($9.99/month, instant SmartScreen reputation) or traditional OV certificate ($200-300/yr, slow reputation build). Sign with `signtool.exe` or `AzureSignTool`. |
| Homebrew `brew install storcat` puts `storcat` on PATH | Users install via Homebrew expecting `storcat create`, `storcat search`, etc. to work immediately in Terminal without reading docs. If CLI doesn't work after install, it feels broken. | LOW | Homebrew `binary` stanza in the cask creates a symlink from `$(brew --prefix)/bin/storcat` to the binary inside the `.app` bundle. One line in the cask formula. |
| WinGet `winget install storcat` puts `storcat` on PATH | Users installing on Windows expect `storcat` to be accessible from Command Prompt or PowerShell. | MEDIUM | NSIS installer must use `EnvVarUpdate` or `EnVar` plugin to add install dir to system PATH. Alternatively, place binary in a directory already on PATH. Requires NSIS script modification. |
| Signing credentials in GitHub Actions secrets (not hardcoded) | Any CI/CD credential management tutorial shows secrets as the baseline. Checking in certs or passwords = security incident. | LOW | GitHub repository secrets + environment protection rules. macOS: 7 secrets. Windows Azure: 6 secrets. |
| Secrets scoped to release environment (not all workflows) | Credentials should not be accessible to every workflow job (e.g., PR builds should never access signing certs). | LOW | GitHub Actions deployment environments allow scoping secrets to specific environments with required reviewer approval as a gate. |

### Differentiators (Competitive Advantage)

| Feature | Value Proposition | Complexity | Notes |
|---------|-------------------|------------|-------|
| macOS notarization + stapling (not just signing) | Signed-but-not-notarized apps still trigger Gatekeeper warnings. Stapling embeds the notarization ticket into the DMG so Gatekeeper verification works offline (no network call required at install time). | MEDIUM | Stapling is the step most tutorials skip. Without it, users on restricted networks may see unnecessary warnings. `xcrun stapler staple` runs after `xcrun notarytool submit` completes. |
| Azure Artifact Signing for Windows (vs. traditional OV cert) | Azure Trusted Signing provides instant SmartScreen reputation (tied to verified identity, not download count). No HSM hardware required. Keys stored in FIPS 140-2 Level 3 HSMs with 3-day cert rotation — much harder to steal than a static .pfx file. | MEDIUM | Restricted to US/Canadian businesses since April 2025. If eligible: `AzureSignTool` integrates with GitHub Actions. If not eligible: fall back to traditional OV cert via DigiCert/Sectigo. |
| Hardened Runtime entitlements file | Wails apps require `com.apple.security.cs.allow-unsigned-executable-memory` or similar entitlements to run with Hardened Runtime (`--options runtime`). Without this, signing succeeds but the app crashes at launch. | LOW | Already solved in the Electron version (`build/entitlements.mac.plist`). The same `.plist` is reusable for the Wails build. Document the entitlements required for Wails specifically. |
| `storcat` binary symlink via Homebrew `binary` stanza references path inside `.app` | The Wails app embeds the Go binary inside the `.app` bundle at `StorCat.app/Contents/MacOS/StorCat`. The `binary` stanza symlinks this exact path into `$(brew --prefix)/bin`. Works because the binary IS the app — no separate CLI binary needed. | LOW | `binary "#{appdir}/StorCat.app/Contents/MacOS/StorCat", target: "storcat"` — one line. The unified binary handles both GUI (when launched directly) and CLI (when launched with subcommands). |
| CI keychain isolated + cleaned up after signing | Using the system keychain in CI leaks credentials between jobs and across workflow runs. An isolated temporary keychain scoped to the job prevents credential accumulation on shared macOS runners. | LOW | `security create-keychain -p "$KEYCHAIN_PASSWORD" build.keychain` + cleanup step using `security delete-keychain`. GitHub's official macOS signing documentation shows this exact pattern. |

### Anti-Features (Commonly Requested, Often Problematic)

| Feature | Why Requested | Why Problematic | Alternative |
|---------|---------------|-----------------|-------------|
| Mac App Store distribution | "Most users install from the App Store" | Requires sandboxing that breaks Wails filesystem access patterns (directory traversal, arbitrary file reads). Wails v2 is not MAS-compatible without major refactoring. | Developer ID distribution (direct download + Homebrew) is the correct path for this app. |
| Windows EV certificate (hardware token) | EV certs gave instant SmartScreen reputation | Since August 2024, Microsoft treats standard OV and EV certs with the same reputation model. EV no longer provides instant reputation advantage. Hardware tokens (YubiKey etc.) cannot be used in CI without complex workarounds. | Azure Artifact Signing (cloud-based, instant reputation) or OV cert with signtool in CI. |
| `gon` for macOS signing/notarization | Listed in Wails docs; popular Go tool | Apple deprecated `altool` (which `gon` relied on). The tool's maintainer has not kept up with `notarytool`. Using `gon` in 2026 results in "Notarization of macOS applications using altool has been decommissioned" errors. | Use `xcrun notarytool` directly or GoReleaser's native notarization (Pro feature). For this project: direct `xcrun` commands in workflow steps. |
| GoReleaser for signing orchestration | Popular Go release tool | GoReleaser cannot drive `wails build` (Wails wraps frontend compilation). GoReleaser native notarization is a Pro feature. The existing pipeline uses native `wails build` commands. Introducing GoReleaser adds dependency without benefit. | Continue with direct `codesign` + `xcrun notarytool` + `signtool` calls in workflow steps. |
| Separate CLI binary (not embedded in .app) | "Simpler PATH setup" | Would break the existing unified binary design (a key v2.1.0 architectural decision). Creates two artifacts to build, test, and sign. Doubles packaging complexity. | The Homebrew `binary` stanza symlinks INTO the `.app`; NSIS installer adds install dir to PATH. One binary, two access paths. |
| Auto-renewal of Apple certificates in CI | "Certificates expire, automate renewal" | Apple Developer ID certificates have a fixed lifespan (1 year for development, longer for distribution). Renewal requires human identity verification via Apple's developer portal. Auto-renewal is not possible — Apple requires human review. | Calendar reminder + documented renewal runbook. GitHub secret rotation is the mechanical part (update base64-encoded cert secret). |

---

## Feature Dependencies

```
[Apple Developer Program enrollment — human, one-time prerequisite]
    └──enables──> [Developer ID Application certificate export (.p12)]
                      └──base64-encoded──> [GitHub secret: MACOS_CERTIFICATE]
                                              └──used by──> [CI: codesign step]
                                                                └──enables──> [CI: xcrun notarytool submit]
                                                                                  └──enables──> [CI: xcrun stapler staple]
                                                                                                    └──produces──> [Signed + notarized + stapled DMG]

[Apple App Store Connect API key — human, one-time prerequisite]
    └──enables──> [GitHub secrets: APPLE_ID, APPLE_PASSWORD (or API key), TEAM_ID]
                      └──used by──> [CI: xcrun notarytool store-credentials]
                                        └──feeds──> [CI: xcrun notarytool submit]

[Azure Artifact Signing account (OR OV cert purchase) — human, one-time prerequisite]
    └──enables──> [Azure App Registration + Certificate Profile (or .pfx export)]
                      └──base64 / tenant ID──> [GitHub secrets: AZURE_TENANT_ID, AZURE_CLIENT_ID, etc.]
                                                   └──used by──> [CI: AzureSignTool or signtool]
                                                                     └──produces──> [Signed NSIS installer]

[Existing release.yml — builds unsigned DMG + NSIS installer]
    └──sign step added after build──> [Signed DMG, Signed installer]
                                          └──existing upload step──> [GitHub Release assets]

[Homebrew cask storcat.rb — existing in packaging/homebrew/]
    └──binary stanza added──> [storcat symlinked to $(brew --prefix)/bin on install]

[NSIS script — existing Wails output]
    └──EnvVarUpdate added──> [install dir added to PATH on Windows install]
```

### Dependency Notes

- **macOS signing and notarization are sequential.** Sign first, notarize second, staple third. Notarizing an unsigned app fails. Stapling before notarization completes fails.
- **Notarytool credentials must be stored in the CI keychain** before `xcrun notarytool submit` is called. The `xcrun notarytool store-credentials` command populates the profile used by `submit`.
- **Homebrew binary stanza requires the .app to be installed first.** The `binary` artifact is created after `app` artifact installation. The cask must declare the `app` stanza before `binary`.
- **NSIS PATH modification requires the installer to run with admin rights.** System PATH (HKLM) modification requires elevation. NSIS RequestExecutionLevel admin must be set. User PATH (HKCU) does not require elevation but affects only the installing user.
- **Signing does not change the DMG content structure** — existing Homebrew SHA256 calculation and URL patterns are unaffected. The distributable file changes (signed vs. unsigned), so the SHA256 changes, but the update automation handles this already.
- **GitHub environment protection rules are a gate, not an alternative.** They restrict which workflows can access signing secrets but do not replace secrets — both are needed together.

---

## MVP Definition

This is a subsequent milestone on a shipping product with existing CI/CD.

### Launch With (v2.3.0)

Minimum viable signed release — users can install without security warnings.

- [ ] macOS DMG: `codesign --sign --options runtime` applied to `.app` bundle in CI
- [ ] macOS DMG: `xcrun notarytool submit` + wait for approval
- [ ] macOS DMG: `xcrun stapler staple` applied to DMG
- [ ] Windows installer: signed with `signtool.exe` or `AzureSignTool` in CI
- [ ] GitHub Actions `signing` environment scoping secrets for signing steps
- [ ] 7 macOS signing secrets stored in GitHub (cert base64, cert password, keychain password, Apple ID, app-specific password, team ID, cert identity string)
- [ ] Windows signing secrets stored in GitHub (6 for Azure or 3 for OV cert route)
- [ ] Homebrew cask: `binary` stanza added to put `storcat` on PATH
- [ ] NSIS installer: PATH modification step added for Windows CLI access
- [ ] Credential rotation runbook documented (what to do when certs expire)

### Add After Validation (v2.3.x)

- [ ] macOS: App Store Connect API key auth (more robust than Apple ID + app-specific password; not deprecated)
- [ ] Windows: Migrate from OV cert to Azure Artifact Signing if business eligibility confirmed
- [ ] Automated notarization status check with retry + timeout (hardened against Apple notarization service slowdowns)

### Future Consideration (v3+)

- [ ] Code signing certificate expiry monitoring (GitHub Actions scheduled check)
- [ ] Automated certificate renewal runbook via CI (mechanical steps; human approval still required)

---

## Feature Prioritization Matrix

| Feature | User Value | Implementation Cost | Priority |
|---------|------------|---------------------|----------|
| Homebrew `binary` stanza for CLI PATH | HIGH | LOW | P1 — one line, zero friction |
| NSIS installer adds install dir to PATH | HIGH | LOW | P1 — EnvVarUpdate/EnVar plugin |
| macOS DMG signed + notarized + stapled | HIGH | HIGH | P1 — table stakes for macOS users |
| Windows NSIS installer signed | HIGH | HIGH | P1 — table stakes for Windows users |
| GitHub environment secrets scoping | HIGH | LOW | P1 — security baseline |
| Apple App Store Connect API key auth | MEDIUM | LOW | P2 — more robust than app-specific password |
| Azure Artifact Signing (if eligible) | MEDIUM | MEDIUM | P2 — instant SmartScreen reputation |
| Notarization retry/timeout handling | MEDIUM | LOW | P2 — hardening against Apple service incidents |
| Cert expiry monitoring | LOW | MEDIUM | P3 — operational concern |

**Priority key:**
- P1: Must have for v2.3.0 launch
- P2: Should have, add after P1 works
- P3: Nice to have, future milestone

---

## Existing Infrastructure Notes

| Existing Item | Impact on New Features |
|---------------|------------------------|
| `release.yml` — 4-platform parallel builds, fan-in release | Add signing steps **within** platform build jobs, after binary/DMG creation, before artifact upload. No structural changes to workflow needed. |
| macOS universal build on `macos-latest` | `codesign` and `xcrun notarytool` are available on GitHub-hosted `macos-latest` runners. No runner changes needed. |
| Windows NSIS build on `windows-2022` | `signtool.exe` is available on Windows runner. `AzureSignTool` is a dotnet tool install. No runner changes needed. |
| `packaging/homebrew/storcat.rb.template` | Add `binary` stanza to existing template. One line. The `distribute.yml` workflow renders the template and pushes to `homebrew-storcat` — no changes to that workflow needed. |
| NSIS build via `wails build -nsis` | Wails generates a default NSIS script. PATH modification requires either a custom NSIS script or post-processing. The Wails NSIS output must be examined to determine customization method. |
| macOS `build/entitlements.mac.plist` (from Electron era) | May be directly reusable. The Wails binary requires `com.apple.security.cs.allow-unsigned-executable-memory` for the Wails runtime. Verify entitlements against Wails requirements before signing. |
| SHA256 checksums in release assembly | Unchanged — signed artifacts have different SHA256 values than unsigned, but the calculation step is the same. |

---

## User Experience Reference

### macOS Without Signing (Current — Unsigned)
1. User downloads DMG, opens it, drags app to Applications
2. First launch: "StorCat cannot be opened because Apple cannot check it for malicious software"
3. User must go to System Settings > Privacy & Security > "Open Anyway"
4. Second confirmation dialog before the app opens
5. In macOS Sequoia: Control-click bypass removed; only System Settings path works
6. Many users abandon at step 2

### macOS With Signing + Notarization (v2.3.0 Target)
1. User downloads DMG, opens it, drags app to Applications
2. First launch: Standard "This was downloaded from the internet — are you sure?" prompt
3. User clicks Open — app launches
4. Subsequent launches: no prompts

### Windows Without Signing (Current — Unsigned)
1. User runs NSIS installer
2. SmartScreen: "Windows protected your PC — Microsoft Defender SmartScreen prevented an unrecognized app from starting"
3. User must click "More info" to see "Run anyway" option
4. Reputation builds over time (~15,000 installs to suppress warning automatically)
5. Every new version resets reputation

### Windows With Signing (v2.3.0 Target)
1. User runs NSIS installer
2. With Azure Artifact Signing: no SmartScreen warning (instant reputation via Microsoft identity verification)
3. With OV cert: SmartScreen shows publisher name but may still warn; reputation builds over time
4. With EV cert (legacy): same as OV as of August 2024 — no longer provides instant reputation

---

## Sources

- [Apple Developer ID overview](https://developer.apple.com/developer-id/) — HIGH confidence
- [GitHub Docs: Installing Apple certificate on macOS runners](https://docs.github.com/actions/use-cases-and-examples/deploying/installing-an-apple-certificate-on-macos-runners-for-xcode-development) — HIGH confidence
- [Federico Terzi: Automatic code signing + notarization via GitHub Actions](https://federicoterzi.com/blog/automatic-code-signing-and-notarization-for-macos-apps-using-github-actions/) — HIGH confidence
- [Melatonin: macOS code signing + notarization in CI](https://melatonin.dev/blog/how-to-code-sign-and-notarize-macos-audio-plugins-in-ci/) — HIGH confidence
- [Melatonin: Windows Azure Trusted Signing](https://melatonin.dev/blog/code-signing-on-windows-with-azure-trusted-signing/) — HIGH confidence
- [Microsoft Artifact Signing (formerly Trusted Signing)](https://azure.microsoft.com/en-us/products/artifact-signing) — HIGH confidence
- [Azure Trusted Signing FAQ](https://learn.microsoft.com/en-us/azure/artifact-signing/faq) — HIGH confidence (geographic restriction note)
- [Homebrew Cask Cookbook — binary stanza](https://docs.brew.sh/Cask-Cookbook) — HIGH confidence
- [NSIS Path Manipulation docs](https://nsis.sourceforge.io/Path_Manipulation) — HIGH confidence
- [Apple macOS Sequoia Gatekeeper change](https://www.idownloadblog.com/2024/08/07/apple-macos-sequoia-gatekeeper-change-install-unsigned-apps-mac/) — HIGH confidence
- [WinGet PATH issue thread (microsoft/winget-cli#4008)](https://github.com/microsoft/winget-cli/issues/4008) — MEDIUM confidence (known pain point, workaround via NSIS)
- [GoReleaser native notarization](https://goreleaser.com/customization/notarize/) — MEDIUM confidence (Pro feature; noted as alternative, not recommended path)
- [Wails signing guide](https://wails.io/docs/guides/signing/) — LOW confidence (403 on fetch; mentions `gon` which is deprecated; verify against current Wails docs before implementation)
- [DigiCert Software Trust Manager end-of-service March 2026](https://github.com/marketplace/actions/code-signing-with-software-trust-manager) — HIGH confidence (avoid — migration deadline passed)

---

*Feature research for: StorCat v2.3.0 — code signing and package manager CLI PATH setup*
*Researched: 2026-03-27*
