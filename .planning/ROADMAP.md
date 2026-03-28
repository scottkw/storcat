# Roadmap: StorCat

## Milestones

- ✅ **v2.0.0 Go/Wails Migration** — Phases 1-7 (shipped 2026-03-26) — [Archive](milestones/v2.0.0-ROADMAP.md)
- ✅ **v2.1.0 CLI Commands** — Phases 8-11 (shipped 2026-03-26) — [Archive](milestones/v2.1.0-ROADMAP.md)
- ✅ **v2.2.0 Repo Consolidation & CI/CD** — Phases 12-15 (shipped 2026-03-27) — [Archive](milestones/v2.2.0-ROADMAP.md)
- 🚧 **v2.3.0 Code Signing & Package Manager CLI** — Phases 16-20 (in progress)

## Phases

<details>
<summary>✅ v2.0.0 Go/Wails Migration (Phases 1-7) — SHIPPED 2026-03-26</summary>

- [x] Phase 1: Data Models + Catalog Service (2/2 plans) — completed 2026-03-24
- [x] Phase 2: Search Service + Browse Metadata (1/1 plans) — completed 2026-03-25
- [x] Phase 3: Config Manager (2/2 plans) — completed 2026-03-25
- [x] Phase 4: App Layer + Lifecycle (2/2 plans) — completed 2026-03-25
- [x] Phase 5: Frontend Shim (1/1 plans) — completed 2026-03-26
- [x] Phase 6: Platform Integration (1/1 plans) — completed 2026-03-26
- [x] Phase 7: Verification + Merge (2/2 plans) — completed 2026-03-26

</details>

<details>
<summary>✅ v2.1.0 CLI Commands (Phases 8-11) — SHIPPED 2026-03-26</summary>

- [x] Phase 8: CLI Foundation and Platform Compatibility (2/2 plans) — completed 2026-03-26
- [x] Phase 9: Core Subcommands — Create, List, Search (2/2 plans) — completed 2026-03-26
- [x] Phase 10: Show, Open, and Output Polish (2/2 plans) — completed 2026-03-26
- [x] Phase 11: Tech Debt Cleanup (1/1 plan) — completed 2026-03-26

</details>

<details>
<summary>✅ v2.2.0 Repo Consolidation & CI/CD (Phases 12-15) — SHIPPED 2026-03-27</summary>

- [x] Phase 12: Repo Consolidation (3/3 plans) — completed 2026-03-27
- [x] Phase 13: CI Scaffold and Multi-Platform Build (1/1 plans) — completed 2026-03-27
- [x] Phase 14: Platform Packaging (1/1 plans) — completed 2026-03-27
- [x] Phase 15: Distribution Channel Automation (2/2 plans) — completed 2026-03-27

</details>

### 🚧 v2.3.0 Code Signing & Package Manager CLI (In Progress)

**Milestone Goal:** Automate macOS/Windows code signing with secure credential handling, and make Homebrew/WinGet installations provide working CLI out of the box.

- [x] **Phase 16: Secrets & Certificate Procurement** - Obtain and configure all signing credentials before any CI automation begins (completed 2026-03-28)
- [x] **Phase 17: macOS Signing & Notarization** - Sign, notarize, and staple every macOS DMG produced by CI (completed 2026-03-28)
- [x] **Phase 18: Windows Authenticode Signing** - Sign Windows NSIS installer and portable .exe before artifact upload (completed 2026-03-28)
- [ ] **Phase 19: Homebrew CLI PATH** - Ensure `brew install --cask storcat` delivers a working `storcat` CLI immediately
- [ ] **Phase 20: Windows CLI PATH via NSIS** - Ensure `winget install scottkw.StorCat` delivers a working `storcat` CLI immediately

## Phase Details

### Phase 16: Secrets & Certificate Procurement
**Goal**: All signing credentials are in hand and configured in GitHub Actions before any signing automation is built
**Depends on**: Nothing (first phase of v2.3.0)
**Requirements**: CRED-01, CRED-02, CRED-03, CRED-04, CRED-05, CRED-06
**Success Criteria** (what must be TRUE):
  1. Developer ID Application certificate is located or renewed and confirmed valid via `security find-identity -v -p codesigning`
  2. Apple certificate is exported as .p12 and base64-encoded value stored as `APPLE_CERTIFICATE` GitHub secret in the `release` environment
  3. Windows OV code signing certificate is obtained with RSA (not ECDSA) confirmed before purchase
  4. Windows Authenticode signing credentials are stored as 4 eSigner API secrets (`ES_USERNAME`, `ES_PASSWORD`, `CREDENTIAL_ID`, `ES_TOTP_SECRET`) in the `release` environment
  5. GitHub `release` environment exists with protection rules and all 9 signing secrets populated
  6. Credential rotation runbook document exists describing what to do when each cert expires
**Plans**: 3 plans
Plans:
- [x] 16-01-PLAN.md — Create GitHub release environment and credential rotation runbook
- [x] 16-02-PLAN.md — Apple certificate verification, export, and secret storage
- [x] 16-03-PLAN.md — Windows signing vendor decision, enrollment, and secret storage

### Phase 17: macOS Signing & Notarization
**Goal**: Every macOS DMG produced by a CI release tag is signed with Developer ID, notarized by Apple, and stapled — Gatekeeper accepts it without prompting on macOS 15+
**Depends on**: Phase 16
**Requirements**: SIGN-01, SIGN-02, SIGN-03, SIGN-04, SIGN-05, SIGN-06
**Success Criteria** (what must be TRUE):
  1. `release.yml` `build-macos` job imports certificate into an isolated temporary keychain and cleans it up after signing completes
  2. The StorCat .app bundle is signed with `--options runtime` and the ported entitlements plist before DMG creation
  3. The DMG is submitted to Apple notarization via `xcrun notarytool` and returns a "Accepted" status
  4. Notarization ticket is stapled to the DMG via `xcrun stapler` so Gatekeeper works offline
  5. `spctl --assess --type exec StorCat.app` returns "accepted" as a CI gate step confirming Gatekeeper acceptance
**Plans**: 1 plan
Plans:
- [x] 17-01-PLAN.md — Create entitlements plist and add signing/notarization/verification pipeline to release.yml

### Phase 18: Windows Authenticode Signing
**Goal**: Every Windows NSIS installer and portable .exe produced by CI is signed with Authenticode before upload, suppressing or reducing SmartScreen blocking
**Depends on**: Phase 16
**Requirements**: WSIGN-01, WSIGN-02, WSIGN-03, WSIGN-04
**Success Criteria** (what must be TRUE):
  1. `release.yml` `build-windows` job signs both the NSIS installer and portable .exe via SSL.com eSigner before `upload-artifact`
  2. `signtool verify /v /pa` confirms a valid Authenticode signature on both Windows binaries as a CI gate step
  3. WinGet manifests in `distribute.yml` compute SHA256 from the signed binaries (signing occurs before artifact upload)
**Plans**: 1 plan
Plans:
- [x] 18-01-PLAN.md — Store eSigner secrets and add signing/verification steps to build-windows job

### Phase 19: Homebrew CLI PATH
**Goal**: Users who run `brew install --cask storcat` get a `storcat` command immediately available in any new terminal session
**Depends on**: Phase 17
**Requirements**: PKG-01, PKG-03
**Success Criteria** (what must be TRUE):
  1. `packaging/homebrew/storcat.rb.template` contains a `binary` stanza that symlinks the StorCat binary into `$(brew --prefix)/bin/storcat`
  2. A user on a fresh macOS machine can run `brew install --cask storcat` and then `storcat version` in a new terminal without any additional PATH configuration
  3. The `storcat` symlink points to the correct binary path inside StorCat.app
**Plans**: TBD
**UI hint**: no

### Phase 20: Windows CLI PATH via NSIS
**Goal**: Users who install StorCat via WinGet get a `storcat` command immediately available in any new Command Prompt or PowerShell session
**Depends on**: Phase 18
**Requirements**: PKG-02, PKG-04
**Success Criteria** (what must be TRUE):
  1. `build/windows/installer.nsi` custom NSIS script adds the StorCat install directory to system PATH via `EnvVarUpdate` macro on install and removes it on uninstall
  2. A user on a fresh Windows machine can run `winget install scottkw.StorCat` and then open a new terminal and run `storcat version` without any additional PATH configuration
  3. PATH registration is visible in System Environment Variables after installation
**Plans**: TBD

## Progress

| Phase | Milestone | Plans Complete | Status | Completed |
|-------|-----------|----------------|--------|-----------|
| 1. Data Models + Catalog Service | v2.0.0 | 2/2 | Complete | 2026-03-24 |
| 2. Search Service + Browse Metadata | v2.0.0 | 1/1 | Complete | 2026-03-25 |
| 3. Config Manager | v2.0.0 | 2/2 | Complete | 2026-03-25 |
| 4. App Layer + Lifecycle | v2.0.0 | 2/2 | Complete | 2026-03-25 |
| 5. Frontend Shim | v2.0.0 | 1/1 | Complete | 2026-03-26 |
| 6. Platform Integration | v2.0.0 | 1/1 | Complete | 2026-03-26 |
| 7. Verification + Merge | v2.0.0 | 2/2 | Complete | 2026-03-26 |
| 8. CLI Foundation and Platform Compatibility | v2.1.0 | 2/2 | Complete | 2026-03-26 |
| 9. Core Subcommands — Create, List, Search | v2.1.0 | 2/2 | Complete | 2026-03-26 |
| 10. Show, Open, and Output Polish | v2.1.0 | 2/2 | Complete | 2026-03-26 |
| 11. Tech Debt Cleanup | v2.1.0 | 1/1 | Complete | 2026-03-26 |
| 12. Repo Consolidation | v2.2.0 | 3/3 | Complete | 2026-03-27 |
| 13. CI Scaffold and Multi-Platform Build | v2.2.0 | 1/1 | Complete | 2026-03-27 |
| 14. Platform Packaging | v2.2.0 | 1/1 | Complete | 2026-03-27 |
| 15. Distribution Channel Automation | v2.2.0 | 2/2 | Complete | 2026-03-27 |
| 16. Secrets & Certificate Procurement | v2.3.0 | 3/3 | Complete    | 2026-03-28 |
| 17. macOS Signing & Notarization | v2.3.0 | 1/1 | Complete    | 2026-03-28 |
| 18. Windows Authenticode Signing | v2.3.0 | 1/1 | Complete    | 2026-03-28 |
| 19. Homebrew CLI PATH | v2.3.0 | 0/TBD | Not started | - |
| 20. Windows CLI PATH via NSIS | v2.3.0 | 0/TBD | Not started | - |
