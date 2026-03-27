# Requirements: StorCat

**Defined:** 2026-03-27
**Core Value:** Fast, lightweight directory catalog management — Go/Wails delivers 93% smaller binaries and 5x faster search, with full feature parity.

## v2.2.0 Requirements

Requirements for repo consolidation and CI/CD pipeline. Each maps to roadmap phases.

### Repo Consolidation

- [ ] **REPO-01**: WinGet manifests moved to `packaging/winget/` in main repo
- [ ] **REPO-02**: Homebrew cask template and update script moved to `packaging/homebrew/` in main repo
- [ ] **REPO-03**: `winget-storcat` repo archived after migration verified
- [ ] **REPO-04**: `homebrew-storcat` README updated to indicate auto-managed status

### CI/CD Pipeline

- [ ] **CICD-01**: Release workflow triggers on GitHub release publish event
- [ ] **CICD-02**: macOS builds on `macos-14` runner producing universal binary
- [ ] **CICD-03**: Windows builds on `windows-latest` runner with `-windowsconsole` preserved
- [ ] **CICD-04**: Linux builds on `ubuntu-22.04` runner for x64 and arm64
- [ ] **CICD-05**: Fan-in release pattern: all platform builds complete before release upload
- [ ] **CICD-06**: All third-party GitHub Actions SHA-pinned (not tag-referenced)

### Installer Packaging

- [ ] **PKG-01**: macOS DMG installer produced via `create-dmg`
- [ ] **PKG-02**: Windows NSIS installer produced via `wails build -nsis`
- [ ] **PKG-03**: Linux AppImage produced for x64
- [ ] **PKG-04**: Linux .deb package produced for x64 and arm64

### Distribution Automation

- [ ] **DIST-01**: Homebrew cask in `homebrew-storcat` auto-updated on release (SHA256 computed locally)
- [ ] **DIST-02**: WinGet manifest auto-submitted to `microsoft/winget-pkgs` on release
- [ ] **DIST-03**: WinGet manifests in main repo auto-updated with new version on release

## Future Requirements

### Code Signing

- **SIGN-01**: macOS binaries codesigned with Apple Developer ID
- **SIGN-02**: macOS DMG notarized via Apple notarization service
- **SIGN-03**: Windows executables Authenticode-signed with PFX certificate
- **SIGN-04**: Signing credentials stored securely in GitHub Secrets with environment protection rules

## Out of Scope

| Feature | Reason |
|---------|--------|
| Code signing (macOS + Windows) | Deferred to future milestone, needs credential pipeline |
| Self-hosted runners | GitHub-hosted runners sufficient for unsigned builds |
| GoReleaser | Incompatible with Wails (bindings generation fails) |
| Linux snap/rpm packages | Low priority, AppImage + deb covers major use cases |
| Automated testing in CI | Separate milestone concern |

## Traceability

Which phases cover which requirements. Updated during roadmap creation.

| Requirement | Phase | Status |
|-------------|-------|--------|
| REPO-01 | — | Pending |
| REPO-02 | — | Pending |
| REPO-03 | — | Pending |
| REPO-04 | — | Pending |
| CICD-01 | — | Pending |
| CICD-02 | — | Pending |
| CICD-03 | — | Pending |
| CICD-04 | — | Pending |
| CICD-05 | — | Pending |
| CICD-06 | — | Pending |
| PKG-01 | — | Pending |
| PKG-02 | — | Pending |
| PKG-03 | — | Pending |
| PKG-04 | — | Pending |
| DIST-01 | — | Pending |
| DIST-02 | — | Pending |
| DIST-03 | — | Pending |

**Coverage:**
- v2.2.0 requirements: 17 total
- Mapped to phases: 0
- Unmapped: 17

---
*Requirements defined: 2026-03-27*
*Last updated: 2026-03-27 after initial definition*
