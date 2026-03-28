# Pitfalls Research

**Domain:** macOS/Windows code signing, notarization, and package manager CLI PATH setup for Go/Wails desktop app
**Project:** StorCat v2.3.0
**Researched:** 2026-03-27
**Confidence:** HIGH (verified against Apple official docs, Wails signing guide, GitHub Actions docs, Homebrew Cask Cookbook, multiple community post-mortems)

---

## Critical Pitfalls

---

### Pitfall 1: macOS Codesign Hangs in Headless CI Without Keychain ACL Setup

**What goes wrong:**
`codesign` blocks waiting for a UI password dialog that never appears in GitHub Actions. The CI job hangs at the signing step until the job timeout kills it. There is no error — the step just never completes. This is the single most common macOS signing failure mode.

**Why it happens:**
macOS requires user confirmation before a private key can be used for signing. In a headless environment with no UI, the dialog never fires, and `codesign` waits indefinitely. Simply importing the certificate into a keychain is insufficient — the keychain must also be unlocked and the codesign tool must be explicitly added to the key partition list.

**How to avoid:**
After importing the certificate, run all three of these commands in sequence:
```bash
security create-keychain -p "$CI_KEYCHAIN_PWD" build.keychain
security default-keychain -s build.keychain
security unlock-keychain -p "$CI_KEYCHAIN_PWD" build.keychain
security import certificate.p12 -k build.keychain -P "$MACOS_CERT_PWD" -T /usr/bin/codesign
security set-key-partition-list -S apple-tool:,apple:,codesign: -s -k "$CI_KEYCHAIN_PWD" build.keychain
```

The `set-key-partition-list` step is the one most often omitted. It grants codesign access without a UI prompt. Use `apple-actions/import-codesign-certs` action as a verified shortcut for this sequence.

**Warning signs:**
- Signing step runs for 10+ minutes with no output
- Job times out at the codesign step
- No error message — just silence, then timeout
- `security find-identity` step would show the cert, but codesign still hangs

**Phase to address:** Phase 1 (macOS signing setup). This must be the very first thing verified before any other signing steps are attempted.

---

### Pitfall 2: Wrong Signing Order — Sign App Bundle Before Creating DMG, Not After

**What goes wrong:**
The DMG is created first, then the code signing step tries to sign the `.app` inside the already-mounted or packaged DMG. The app is signed but the signature is immediately invalidated when the DMG is re-assembled. Notarization then fails with "The signature is invalid" or "The bundle's signature indicates modified/obsolete code."

**Why it happens:**
Developers copy the existing DMG creation step from the current workflow (which runs after build) and insert signing before it, but after build artifacts are already staged. Any modification to the app bundle after signing — including copying it into a DMG — breaks the signature.

The correct sequence is: sign app bundle → create DMG from signed app → notarize the DMG → staple the notarization ticket to the DMG.

**How to avoid:**
Structure the macOS build job steps in this exact order:
1. `wails build -platform darwin/universal`
2. `codesign` the `.app` bundle (with `--options runtime --entitlements`)
3. `hdiutil create` the DMG from the signed `.app`
4. `xcrun notarytool submit` the DMG (not the `.app`)
5. `xcrun stapler staple` the DMG

Never staple or notarize the `.app` directly — only ZIP, PKG, and DMG containers are accepted by the notarization service.

**Warning signs:**
- Notarization fails with "invalid signature" after signing was reported as successful
- DMG creation step appears before codesign step in workflow YAML
- Signing step targets the binary inside the bundle (`Contents/MacOS/StorCat`) rather than the `.app` bundle root

**Phase to address:** Phase 1 (macOS signing setup). Step ordering must be locked in before the first signing run.

---

### Pitfall 3: Missing Entitlements Break Wails Webview at Runtime

**What goes wrong:**
The app is signed with `--options runtime` (hardened runtime, required for notarization) but without entitlements. At runtime, the Wails WebView fails to load — either a blank window, crash on launch, or network requests blocked. The app passes notarization but is broken for users.

**Why it happens:**
Apple's hardened runtime disables JIT compilation, dynamic code loading, and certain memory features by default. The Wails WebView (based on WKWebView) requires `com.apple.security.cs.allow-jit` and often `com.apple.security.cs.allow-unsigned-executable-memory` to function. Without these entitlements in the codesign invocation, the hardened runtime blocks the WebView at launch.

StorCat already has entitlements from its Electron era. These must be ported to the new Go/Wails signing step.

**How to avoid:**
Create an `entitlements.plist` with the minimum required permissions for Wails:
```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>com.apple.security.cs.allow-jit</key>
    <true/>
    <key>com.apple.security.cs.allow-unsigned-executable-memory</key>
    <true/>
    <key>com.apple.security.network.client</key>
    <true/>
</dict>
</plist>
```

Pass to codesign: `codesign --deep --force --options runtime --entitlements entitlements.plist --sign "Developer ID Application: ..." StorCat.app`

**Warning signs:**
- App is notarized but shows blank window or crashes on first launch
- Console.app shows `code signing` or `AMFI` denials at launch
- Signing works but Wails runtime fails to initialize WebView
- Testing is done with ad-hoc signing (`--sign -`) which bypasses the hardened runtime

**Phase to address:** Phase 1 (macOS signing setup). Test with hardened runtime signing locally before CI; don't discover this in production.

---

### Pitfall 4: Notarization Timeout — Using `--wait` Indefinitely

**What goes wrong:**
`xcrun notarytool submit --wait` hangs for 30-90 minutes (or indefinitely) in CI. Apple's notarization service is occasionally slow. The GitHub Actions job timeout kills it and the release fails. The submission may have actually succeeded — there's just no way to know from the timed-out output.

**Why it happens:**
`--wait` polls until Apple responds. Apple's service SLA is undefined and can spike during high-load periods. Large DMGs (StorCat universal DMG includes both arm64 and x86_64 binaries plus the Wails frontend) take longer to process. Combining large file size with `--wait` in a 6-hour CI job creates a silent reliability problem.

**How to avoid:**
Use `--wait` with a reasonable timeout (`--timeout 30m`) to get a clean failure rather than a silent hang, then poll for the result:
```bash
xcrun notarytool submit StorCat.dmg \
  --apple-id "$APPLE_ID" \
  --team-id "$APPLE_TEAM_ID" \
  --password "$APPLE_APP_PASSWORD" \
  --wait \
  --timeout 30m
```

Alternatively, use the App Store Connect API key approach (`.p8` key file) instead of Apple ID + app-specific password — the API key method is more reliable in CI and avoids issues with Apple ID 2FA.

**Warning signs:**
- Release workflow consistently takes 45+ minutes on the macOS job
- Job sometimes succeeds and sometimes fails with timeout
- No explicit `--timeout` parameter on notarytool submit

**Phase to address:** Phase 1 (macOS signing setup). Test notarization timing in a dry run before wiring to the release workflow.

---

### Pitfall 5: Apple Developer ID Certificate Stored as P12 Requires Password — Both Must Be in Secrets

**What goes wrong:**
The P12 certificate is exported and base64-encoded into a GitHub Secret (`MACOS_CERTIFICATE`), but the import password is not separately stored, or is stored under the wrong name. The `security import` command fails silently or with "incorrect password" and the keychain contains no valid signing identity. Subsequent `codesign` runs use an ad-hoc signature and the issue goes unnoticed until notarization rejects the submission.

**Why it happens:**
When exporting a Developer ID Application certificate from Keychain Access, you must set an export password. This password is separate from the macOS login password. Teams forget to record it or assume any password works. In CI, the import step uses the password from the secret — if it doesn't match the export password, import silently succeeds (the P12 is decoded) but the private key is not unlocked.

**How to avoid:**
Store three distinct secrets:
- `MACOS_CERTIFICATE` — base64-encoded P12 file (`base64 -i certificate.p12 | pbcopy`)
- `MACOS_CERTIFICATE_PWD` — the export password set during P12 export
- `MACOS_CI_KEYCHAIN_PWD` — a random password for the ephemeral CI keychain

After import, validate the identity is present before proceeding:
```bash
security find-identity -v -p codesigning build.keychain | grep "Developer ID"
```
If this outputs zero identities, fail fast rather than producing an unsigned artifact.

**Warning signs:**
- `security import` exits 0 but `security find-identity` returns no Developer ID certificate
- Notarization fails with "No signing certificate found"
- `codesign -vvv` shows `Signature=adhoc` instead of a certificate identity

**Phase to address:** Phase 1 (macOS signing setup). Certificate validation step must be a hard gate in the workflow.

---

### Pitfall 6: Windows RSA vs ECDSA Certificate Type Mismatch

**What goes wrong:**
An ECDSA code signing certificate is purchased (Sectigo and several resellers offer them) and used with SignTool or AzureSignTool. Signing appears to succeed, but SmartScreen rejects the binary because the Microsoft Trusted Root Program does not accept ECDSA for Authenticode code signing. Users see the "Windows protected your PC" SmartScreen warning despite the binary being signed.

**Why it happens:**
ECDSA is a valid cryptographic standard and many CAs offer ECDSA code signing certificates. However, Windows Authenticode and SmartScreen specifically require RSA PKCS#1 v1.5. This restriction is not prominently documented by certificate resellers.

**How to avoid:**
When purchasing or configuring any code signing certificate: explicitly verify RSA key type. For Azure Key Vault HSM: select `RSA-HSM` (not `EC-HSM`) during key creation. For OV/EV certs from DigiCert, Sectigo, etc.: request RSA 2048 or RSA 4096 explicitly.

For Microsoft Trusted Signing (the modern approach): this is handled automatically — the service manages certificate lifecycle and key type.

**Warning signs:**
- Certificate details show `Algorithm: EC` or `EC-HSM` key type
- SmartScreen warning appears despite "signed" binary
- `signtool verify /pa /v executable.exe` shows certificate but SmartScreen still warns

**Phase to address:** Phase 2 (Windows signing setup). Verify certificate type before any signing infrastructure is built.

---

### Pitfall 7: WinGet Manifest SHA256 Will Change After Code Signing — Hashes Computed Before Signing Are Wrong

**What goes wrong:**
The current `distribute.yml` computes the SHA256 of the Windows NSIS installer and commits WinGet manifests with that hash. Once Windows Authenticode signing is added to the build job, the installer binary changes (the signature is embedded). The SHA256 computed by the distribute workflow from the GitHub release asset will no longer match the unsigned binary's SHA256 from any prior step. If the compute-hash step runs before signing completes, the wrong hash is committed to the WinGet manifests.

**Why it happens:**
Signing modifies the PE binary. The Authenticode signature is embedded in the binary's Authenticode signature directory. This changes the binary's SHA256. Any hash computed before signing is invalid for WinGet distribution, which verifies the hash of the downloaded file.

**How to avoid:**
Sign the installer in the build job before uploading the artifact. The artifact uploaded to GitHub releases is the signed binary. The SHA256 computed by `distribute.yml` (which downloads from the release URL) will then be correct because it's computing the hash of the already-signed binary.

Sequence:
1. `wails build -nsis` → unsigned installer
2. Sign installer with AzureSignTool or SignTool
3. `actions/upload-artifact` of the signed installer
4. Fan-in release job uploads the signed installer to GitHub releases
5. `distribute.yml` downloads the signed installer, computes SHA256, commits manifests

Never compute SHA256 of the installer before the signing step completes.

**Warning signs:**
- `winget install scottkw.StorCat` fails with "Installer hash does not match"
- Hash computation step in distribute.yml runs on a different artifact than what was signed
- Signing step added to workflow but its position relative to artifact upload was not reconsidered

**Phase to address:** Phase 2 (Windows signing setup). Review artifact upload ordering when adding signing.

---

### Pitfall 8: Azure Trusted Signing vs OV/EV Certificate — SmartScreen Reputation Delay With OV Certs

**What goes wrong:**
A standard OV (Organization Validation) code signing certificate is purchased. The installer is properly signed. However, SmartScreen shows a warning ("Windows protected your PC") for 2-8 weeks after the first release because the certificate has no accumulated reputation. Users see a scary warning and many abandon the install. This is expected behavior for new OV certs — but teams are surprised by it.

**Why it happens:**
SmartScreen reputation is tied to the certificate identity, not the binary hash. A new OV certificate has zero reputation. Microsoft does not publish a timeline, but the community consensus is 2-8 weeks before the warning disappears, and only if enough users click "Run anyway."

EV (Extended Validation) certificates bypass the initial SmartScreen warning, but as of 2023-2025, EV certificates require a hardware HSM token (a physical USB key), which is incompatible with CI/CD pipelines. There is no software-only EV certificate option.

Microsoft Trusted Signing ($9.99/month) provides instant reputation because the service is already trusted by SmartScreen. This is the recommended approach for CI/CD pipelines in 2025.

**How to avoid:**
Use Microsoft Azure Trusted Signing instead of a traditional OV certificate. Benefits:
- Instant SmartScreen reputation (tied to Microsoft's root program)
- No hardware token requirement — fully compatible with GitHub Actions
- $9.99/month vs $200-500/year for EV cert + hardware HSM
- Certificates auto-rotate (3-day validity); no annual renewal

Integration: `azure/trusted-signing-action` GitHub Action with OIDC authentication.

**Warning signs:**
- OV certificate purchased without awareness of reputation delay
- Assuming "signed = no SmartScreen warning"
- EV certificate purchased expecting CI/CD compatibility

**Phase to address:** Phase 2 (Windows signing decision). Decide on signing approach before purchasing any certificate.

---

### Pitfall 9: Homebrew Cask `binary` Stanza Path Must Exactly Match the App Bundle Layout

**What goes wrong:**
The `binary` stanza in `storcat.rb` points to the wrong path inside the `.app` bundle, causing `brew install --cask storcat` to succeed but `storcat` command not found on PATH. Users install successfully but get no CLI functionality.

**Why it happens:**
The binary stanza requires an exact path relative to the cask staging directory. The path `"StorCat.app/Contents/MacOS/StorCat"` is case-sensitive on macOS. If the app binary name or directory case differs between the wails build output and the stanza, the symlink target does not exist and Homebrew silently skips it (in some versions) or fails with "symlink source not found."

Also: Homebrew cask requires the binary to be in the installed application (the `.app` bundle in `/Applications`). The binary stanza creates a symlink from `$(brew --prefix)/bin/storcat` to the full path inside the installed app.

**How to avoid:**
Use the exact case of the binary as produced by `wails build`. Verify with:
```bash
ls -la build/bin/StorCat.app/Contents/MacOS/
```

The cask stanza:
```ruby
binary "#{appdir}/StorCat.app/Contents/MacOS/StorCat", target: "storcat"
```

Note: `appdir` is a Homebrew variable pointing to `/Applications` (or `~/Applications`). The `target:` parameter sets the symlink name in `$(brew --prefix)/bin/`.

Test locally before deploying: `brew install --cask --verbose ./Casks/storcat.rb`

**Warning signs:**
- `brew install --cask storcat` succeeds but `which storcat` returns nothing
- Homebrew verbose output shows no symlink creation step
- Path in stanza uses lowercase `storcat` but binary is `StorCat` (case mismatch)

**Phase to address:** Phase 3 (Homebrew CLI PATH setup). Must be tested in dry-run install before wiring to release.

---

### Pitfall 10: WinGet Installation PATH — NSIS Installer Must Configure PATH During Install

**What goes wrong:**
Users install StorCat via WinGet (`winget install scottkw.StorCat`). The NSIS installer runs, the GUI app is installed, but `storcat` on the command line returns "command not found." The PATH is not updated because the NSIS installer does not include an `EnvVarUpdate` or `${EnvVarUpdate}` step for the install directory.

**Why it happens:**
NSIS installers do not add to PATH by default. Wails generates a basic NSIS installer that installs the binary to `%ProgramFiles%\StorCat\` but does not modify `%PATH%`. WinGet does not add anything to PATH either — it runs the installer and trusts the installer to handle PATH configuration.

**How to avoid:**
Customize the Wails NSIS installer to add the install directory to the system PATH. Wails supports NSIS script customization via a custom `.nsi` file. The standard approach is to use the NSIS `EnvVarUpdate` macro or call `WriteRegExpandStr` to the `Environment` registry key:

```nsis
!include "EnvVarUpdate.nsh"
${EnvVarUpdate} $0 "PATH" "A" "HKLM" "$INSTDIR"
```

Alternatively, use the NSIS `CreateShortCut` approach to add a batch wrapper in a directory already on PATH.

For WinGet manifests, ensure `InstallerType: exe` with the correct `Scope` and that no special handling is needed beyond what the NSIS installer does.

**Warning signs:**
- `winget install scottkw.StorCat` succeeds but `storcat version` fails
- NSIS script has no PATH modification section
- `EnvVarUpdate.nsh` not included in NSIS build
- Testing only verifies GUI launch, not CLI invocation from a new terminal

**Phase to address:** Phase 3 (WinGet CLI PATH setup). Must be verified in a fresh Windows VM after WinGet install — not just a developer machine with a pre-existing PATH.

---

### Pitfall 11: Notarization of DMG Requires the App Inside to Already Be Code-Signed

**What goes wrong:**
The DMG is created from the unsigned `.app`, then the DMG is submitted for notarization. Apple's notarization service rejects it with "The signature of the binary is invalid" because the `.app` inside the DMG has no Developer ID signature.

This is a sequencing variant of Pitfall 2, but specifically about the DMG notarization — teams sometimes sign the DMG itself without signing the `.app` inside it first.

**Why it happens:**
Apple's notarization service requires that every executable inside a notarized container be code-signed with a Developer ID. If you sign the DMG (or submit a DMG without signing it), the outer container may pass but the inner app will be flagged.

**How to avoid:**
Always verify the `.app` is signed before packaging into DMG:
```bash
codesign --verify --deep --strict StorCat.app
```
Only proceed to DMG creation if this exits 0. The DMG does not need its own codesign step — notarization of the DMG covers it. But the `.app` inside must be signed first.

**Warning signs:**
- Notarization log mentions "unsigned binary inside archive"
- `codesign --verify` step is absent from workflow YAML
- DMG creation step comes before codesign step in the workflow

**Phase to address:** Phase 1 (macOS signing setup). Add explicit verification step after codesign.

---

### Pitfall 12: Azure Credential Expiration Silently Breaks Windows Signing After 24 Months

**What goes wrong:**
The Azure service principal client secret configured for AzureSignTool or Azure Trusted Signing expires. The signing step in CI starts returning 401/403 errors. This happens 24 months after setup — well after the initial implementation is forgotten. The release workflow fails silently (or with a cryptic Azure auth error) and unsigned binaries may slip through if the failure isn't caught.

**Why it happens:**
Azure service principal client secrets have a maximum validity of 24 months. There's no default alert or reminder. The GitHub Actions secret doesn't expire — only the Azure credential it contains.

**How to avoid:**
Use OIDC (OpenID Connect) federation instead of client secrets. OIDC eliminates the need for long-lived credentials:
```yaml
- uses: azure/login@v2
  with:
    client-id: ${{ secrets.AZURE_CLIENT_ID }}
    tenant-id: ${{ secrets.AZURE_TENANT_ID }}
    subscription-id: ${{ secrets.AZURE_SUBSCRIPTION_ID }}
```

OIDC tokens are issued per-run and never expire. This is the recommended approach as of 2024-2025 for Azure + GitHub Actions.

If client secrets are used (fallback), set a calendar reminder for 20 months from creation date and rotate before expiration.

**Warning signs:**
- Azure auth step returning 401 errors
- Signing worked previously but now fails with no code changes
- `AZURE_CLIENT_SECRET` stored as a long string in GitHub Secrets (sign of client secret, not OIDC)

**Phase to address:** Phase 2 (Windows signing setup). Choose OIDC from the start to avoid the expiration problem entirely.

---

## Technical Debt Patterns

| Shortcut | Immediate Benefit | Long-term Cost | When Acceptable |
|----------|-------------------|----------------|-----------------|
| Computing SHA256 before signing | Simpler workflow ordering | WinGet installs fail with hash mismatch for every release | Never |
| Signing DMG without signing the `.app` inside | Fewer steps | Notarization rejects the submission; release blocked | Never |
| Omitting `--options runtime` from codesign | Avoids entitlement complexity | Notarization requires hardened runtime; unsigned-equivalent app | Never |
| Using OV certificate expecting no SmartScreen warning | Cheaper than Trusted Signing | 2-8 week SmartScreen window causes user trust issues | Only if reputation delay is acceptable and documented |
| Skipping `security set-key-partition-list` | One less setup step | codesign hangs indefinitely in CI | Never |
| Storing Azure client secret instead of using OIDC | Simpler initial setup | Credential expires after 24 months; silent CI breakage | Only as temporary stopgap with documented expiry date |
| Using `binary` stanza without testing on a fresh install | Works in dev | CLI not on PATH after WinGet/Homebrew install; users can't use CLI | Never |

---

## Integration Gotchas

| Integration | Common Mistake | Correct Approach |
|-------------|----------------|------------------|
| macOS codesign + GitHub Actions | Missing `set-key-partition-list` | Run full 5-command keychain setup sequence; or use `apple-actions/import-codesign-certs` |
| notarytool + DMG | Submitting unsigned app inside DMG | Sign `.app` first, then create DMG, then notarize DMG |
| notarytool + `--wait` | No timeout on `--wait` | Use `--wait --timeout 30m` to get clean failure instead of silent hang |
| Windows Authenticode + AzureSignTool | ECDSA key type | Explicitly choose RSA-HSM in Azure Key Vault; never ECDSA for Authenticode |
| Windows Authenticode + WinGet | Hash computed before signing | Artifact upload must be post-signing; distribute workflow downloads signed binary |
| Azure credentials + GitHub Actions | Client secret with 2-year expiry | Use OIDC federation (`azure/login` with `client-id`/`tenant-id`) |
| Homebrew `binary` stanza | Wrong path or case in stanza | Verify exact binary path with `ls` on build output; test with `brew install --cask --verbose` |
| WinGet + NSIS installer | No PATH configuration in installer | Add `EnvVarUpdate` macro or registry PATH write to NSIS script |
| macOS hardened runtime + Wails WebView | Missing entitlements | Include `allow-jit` and `allow-unsigned-executable-memory` in entitlements.plist |

---

## "Looks Done But Isn't" Checklist

- [ ] **Keychain ACL configured:** `set-key-partition-list` step present in macOS signing workflow; codesign does not hang.
- [ ] **Signing order correct:** codesign step appears before `hdiutil create` step in workflow YAML.
- [ ] **Entitlements plist included:** `--entitlements entitlements.plist` flag on every `codesign` invocation; Wails WebView loads after signing.
- [ ] **Certificate identity verified:** `security find-identity -v -p codesigning` step after import shows Developer ID certificate.
- [ ] **Notarization timeout set:** `xcrun notarytool submit` uses `--wait --timeout 30m` not unbounded `--wait`.
- [ ] **Stapling step present:** `xcrun stapler staple StorCat.dmg` runs after notarization succeeds.
- [ ] **Stapling verified:** `xcrun stapler validate` or `spctl -a -v` confirms staple is attached.
- [ ] **Windows cert type RSA:** Certificate or Azure key type is RSA (not ECDSA).
- [ ] **Signing before upload:** Windows NSIS installer is signed before `actions/upload-artifact` step.
- [ ] **WinGet hash correct:** SHA256 in WinGet manifests matches the signed installer, not the unsigned build artifact.
- [ ] **NSIS PATH config:** NSIS installer includes PATH modification for install directory.
- [ ] **WinGet PATH verified:** Fresh Windows VM install via `winget install scottkw.StorCat` followed by `storcat version` in a new terminal succeeds.
- [ ] **Homebrew binary stanza path:** Path in `storcat.rb` matches the exact case/name of the binary in `StorCat.app/Contents/MacOS/`.
- [ ] **Homebrew CLI verified:** Fresh macOS install via `brew install --cask storcat` followed by `storcat version` in a new terminal succeeds.
- [ ] **Azure OIDC (not client secret):** Azure signing uses OIDC federation, not a client secret with expiry.

---

## Recovery Strategies

| Pitfall | Recovery Cost | Recovery Steps |
|---------|---------------|----------------|
| Signing hangs (missing ACL) | LOW | Add `set-key-partition-list` step; re-run release workflow |
| Wrong signing order (DMG before sign) | LOW | Reorder YAML steps; re-run; re-upload assets to existing release |
| Notarization failed (unsigned app in DMG) | LOW | Fix signing order; re-run; Apple notarization resubmit is fast |
| Entitlements missing (blank WebView) | LOW | Add entitlements.plist; re-sign; re-notarize; re-upload |
| Wrong certificate identity in CI | MEDIUM | Re-export P12 with known password; update both secrets; re-run |
| ECDSA cert purchased (wrong type) | HIGH | Must purchase new RSA certificate; cannot convert ECDSA to RSA |
| WinGet hash mismatch (signed vs unsigned) | MEDIUM | Fix ordering; bump patch version; resubmit WinGet PR with correct hash |
| SmartScreen warning (OV cert, no reputation) | HIGH | Switch to Microsoft Trusted Signing, or wait 2-8 weeks for reputation accumulation |
| Homebrew CLI not on PATH (binary stanza wrong) | LOW | Fix path in template; push tap update; users re-run `brew upgrade --cask storcat` |
| WinGet CLI not on PATH (NSIS missing EnvVarUpdate) | MEDIUM | Fix NSIS script; bump patch version; rebuild; resubmit WinGet PR |
| Azure client secret expired | MEDIUM | Rotate secret in Azure; update GitHub Secret; or migrate to OIDC |

---

## Pitfall-to-Phase Mapping

| Pitfall | Prevention Phase | Verification |
|---------|------------------|--------------|
| Codesign hangs — missing keychain ACL | Phase 1: macOS signing | `codesign` step completes in <2 min; no timeout |
| Wrong signing order (sign after DMG) | Phase 1: macOS signing | YAML step order: codesign → hdiutil → notarytool → stapler |
| Missing entitlements | Phase 1: macOS signing | App launches and WebView renders after notarized install |
| Notarization timeout — no `--timeout` | Phase 1: macOS signing | `--wait --timeout 30m` on notarytool submit |
| P12 password mismatch | Phase 1: macOS signing | `security find-identity` gate step |
| Windows ECDSA cert | Phase 2: Windows signing | Cert details show RSA before any signing work |
| WinGet hash computed before signing | Phase 2: Windows signing | Artifact upload step is after signing step in build job |
| OV cert SmartScreen reputation delay | Phase 2: Windows signing decision | Trusted Signing chosen as approach before phase begins |
| Homebrew binary stanza wrong path | Phase 3: Homebrew CLI PATH | `brew install --cask` dry-run on fresh macOS before release |
| WinGet NSIS missing PATH | Phase 3: WinGet CLI PATH | Fresh Windows VM test: `winget install` + `storcat version` |
| Azure credential expiration | Phase 2: Windows signing | OIDC chosen; no expiry-prone client secret in use |
| Unsigned app inside DMG | Phase 1: macOS signing | `codesign --verify --deep --strict` step before hdiutil |

---

## Sources

- Wails code signing guide: https://wails.io/docs/guides/signing/
- Apple: Installing certificate on macOS runners for Xcode development: https://docs.github.com/en/actions/deployment/deploying-xcode-applications/installing-an-apple-certificate-on-macos-runners-for-xcode-development
- Federico Terzi: Automatic code signing and notarization for macOS with GitHub Actions: https://federicoterzi.com/blog/automatic-code-signing-and-notarization-for-macos-apps-using-github-actions/
- apple-actions/import-codesign-certs: https://github.com/Apple-Actions/import-codesign-certs
- Apple: `security set-key-partition-list` explanation: https://developer.apple.com/forums/thread/666107
- Melatonin: How to code sign Windows installers with EV cert on GitHub Actions: https://melatonin.dev/blog/how-to-code-sign-windows-installers-with-an-ev-cert-on-github-actions/
- Melatonin: Code signing on Windows with Azure Trusted Signing: https://melatonin.dev/blog/code-signing-on-windows-with-azure-trusted-signing/
- Scott Hanselman: Signing a Windows EXE with Azure Trusted Signing and GitHub Actions: https://www.hanselman.com/blog/automatically-signing-a-windows-exe-with-azure-trusted-signing-dotnet-sign-and-github-actions
- Rick Strahl: Fighting through Setting up Microsoft Trusted Signing: https://weblog.west-wind.com/posts/2025/Jul/20/Fighting-through-Setting-up-Microsoft-Trusted-Signing
- AzureSignTool (vcsjones): https://github.com/vcsjones/AzureSignTool
- Azure: artifact-signing-action: https://github.com/Azure/artifact-signing-action
- Homebrew Cask Cookbook — binary stanza: https://docs.brew.sh/Cask-Cookbook
- Notarizing Apps Distributed as DMGs (Decipher Tools): https://deciphertools.com/blog/notarizing-dmg/
- Random Errata: A rough guide to notarizing CLI apps for macOS (2024): https://www.randomerrata.com/articles/2024/notarize/
- WinGet hash verification: https://learn.microsoft.com/en-us/windows/package-manager/winget/hash
- DigiCert: MS SmartScreen and Application Reputation: https://www.digicert.com/blog/ms-smartscreen-application-reputation
- GoReleaser: Homebrew Casks — binary stanza patterns: https://goreleaser.com/customization/homebrew_casks/

---
*Pitfalls research for: macOS/Windows code signing, notarization, and package manager CLI PATH setup for Go/Wails desktop app*
*Researched: 2026-03-27*
