# Requirements: StorCat

**Defined:** 2026-03-27
**Core Value:** Fast, lightweight directory catalog management — Go/Wails delivers 93% smaller binaries and 5x faster search, with full feature parity.

## v2.3.0 Requirements

Requirements for Code Signing & Package Manager CLI milestone. Each maps to roadmap phases.

### Credentials & Secrets

- [ ] **CRED-01**: User can locate or renew existing Apple Developer ID Application certificate
- [ ] **CRED-02**: User can export Developer ID cert as .p12 and store base64-encoded in GitHub secret
- [ ] **CRED-03**: User can locate or acquire Windows OV code signing certificate (RSA, not ECDSA)
- [ ] **CRED-04**: User can store Windows cert as base64-encoded PFX in GitHub secret
- [x] **CRED-05**: All 9 signing secrets stored in GitHub Actions `release` environment with protection rules
- [x] **CRED-06**: Credential rotation runbook documents what to do when certs expire

### macOS Signing

- [x] **SIGN-01**: release.yml `build-macos` job signs .app bundle with `codesign --sign --options runtime` and entitlements
- [x] **SIGN-02**: release.yml `build-macos` job submits DMG to Apple notarization service via `xcrun notarytool`
- [x] **SIGN-03**: release.yml `build-macos` job staples notarization ticket to DMG via `xcrun stapler`
- [x] **SIGN-04**: CI uses isolated temporary keychain, cleaned up after signing
- [x] **SIGN-05**: Entitlements plist ported from Electron era and verified for Wails runtime requirements
- [x] **SIGN-06**: `spctl --assess` verification step confirms signed .app is accepted by Gatekeeper

### Windows Signing

- [ ] **WSIGN-01**: release.yml `build-windows` job signs NSIS installer with `signtool.exe` (or Azure Trusted Signing)
- [ ] **WSIGN-02**: release.yml `build-windows` job signs portable .exe with same certificate
- [ ] **WSIGN-03**: Signing occurs before `upload-artifact` so WinGet SHA256 is computed from signed binary
- [ ] **WSIGN-04**: `signtool verify` step confirms valid signature before upload

### Package Manager CLI

- [ ] **PKG-01**: Homebrew cask `binary` stanza puts `storcat` on PATH after `brew install --cask storcat`
- [ ] **PKG-02**: NSIS installer adds install directory to system PATH via `EnvVarUpdate`
- [ ] **PKG-03**: `storcat version` works from any new terminal after Homebrew install (macOS)
- [ ] **PKG-04**: `storcat version` works from any new terminal after WinGet install (Windows)

## Future Requirements

### Post-Validation (v2.3.x)

- **SIGN-07**: App Store Connect API key auth for notarytool (more robust than Apple ID + app-specific password)
- **WSIGN-05**: Migrate to Azure Trusted Signing if business eligibility confirmed
- **SIGN-08**: Notarization retry + timeout handling for Apple service slowdowns

### Future Milestone (v3+)

- **SIGN-09**: Code signing certificate expiry monitoring via scheduled GitHub Actions
- **SIGN-10**: Automated certificate renewal runbook via CI

## Out of Scope

| Feature | Reason |
|---------|--------|
| Mac App Store distribution | Requires sandboxing incompatible with Wails filesystem access |
| EV certificate for Windows | No advantage over OV since August 2024; requires hardware token incompatible with CI |
| `gon` for macOS notarization | Deprecated — Apple decommissioned `altool` |
| GoReleaser for signing | Cannot drive `wails build`; Pro feature for notarization |
| Separate CLI binary | Breaks unified binary architecture from v2.1.0 |

## Traceability

Which phases cover which requirements. Updated during roadmap creation.

| Requirement | Phase | Status |
|-------------|-------|--------|
| CRED-01 | Phase 16 | Pending |
| CRED-02 | Phase 16 | Pending |
| CRED-03 | Phase 16 | Pending |
| CRED-04 | Phase 16 | Pending |
| CRED-05 | Phase 16 | Complete |
| CRED-06 | Phase 16 | Complete |
| SIGN-01 | Phase 17 | Complete |
| SIGN-02 | Phase 17 | Complete |
| SIGN-03 | Phase 17 | Complete |
| SIGN-04 | Phase 17 | Complete |
| SIGN-05 | Phase 17 | Complete |
| SIGN-06 | Phase 17 | Complete |
| WSIGN-01 | Phase 18 | Pending |
| WSIGN-02 | Phase 18 | Pending |
| WSIGN-03 | Phase 18 | Pending |
| WSIGN-04 | Phase 18 | Pending |
| PKG-01 | Phase 19 | Pending |
| PKG-03 | Phase 19 | Pending |
| PKG-02 | Phase 20 | Pending |
| PKG-04 | Phase 20 | Pending |

**Coverage:**
- v2.3.0 requirements: 20 total
- Mapped to phases: 20
- Unmapped: 0

---
*Requirements defined: 2026-03-27*
*Last updated: 2026-03-27 after roadmap creation*
