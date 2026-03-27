# Roadmap: StorCat

## Milestones

- ✅ **v2.0.0 Go/Wails Migration** — Phases 1-7 (shipped 2026-03-26) — [Archive](milestones/v2.0.0-ROADMAP.md)
- ✅ **v2.1.0 CLI Commands** — Phases 8-11 (shipped 2026-03-26) — [Archive](milestones/v2.1.0-ROADMAP.md)
- 🔄 **v2.2.0 Repo Consolidation & CI/CD** — Phases 12-15 (active)

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

### v2.2.0 Repo Consolidation & CI/CD (Phases 12-15)

- [x] **Phase 12: Repo Consolidation** — Move WinGet manifests and Homebrew files into main repo; archive satellite (completed 2026-03-27)
- [ ] **Phase 13: CI Scaffold and Multi-Platform Build** — Release workflow with fan-in DAG, correct runners, SHA pinning
- [ ] **Phase 14: Platform Packaging** — DMG, NSIS installer, AppImage, and .deb produced and attached to release
- [ ] **Phase 15: Distribution Channel Automation** — Homebrew tap and WinGet manifests auto-updated on release

## Phase Details

### Phase 12: Repo Consolidation
**Goal**: All packaging metadata lives in the main repo with verified local scripts
**Depends on**: Nothing (first phase of milestone)
**Requirements**: REPO-01, REPO-02, REPO-03, REPO-04
**Success Criteria** (what must be TRUE):
  1. `packaging/winget/manifests/` directory exists in main repo with current version manifests committed
  2. `packaging/homebrew/storcat.rb.template` and `update-tap.sh` exist in main repo and execute correctly from their new paths
  3. `winget-storcat` repo is archived and its README links to `packaging/winget/` in main repo
  4. `homebrew-storcat` README states the tap is auto-managed and links to main repo for source of truth
**Plans**: 3 plans
Plans:
- [x] 12-01-PLAN.md — Migrate WinGet manifests and create v2.1.0 stubs (complete 2026-03-27)
- [x] 12-02-PLAN.md — Migrate Homebrew cask template and update script (complete 2026-03-27)
- [x] 12-03-PLAN.md — Archive winget-storcat and update homebrew-storcat README

### Phase 13: CI Scaffold and Multi-Platform Build
**Goal**: Release workflow fires on tag push and produces raw binaries on correct runners with fan-in release assembly
**Depends on**: Phase 12
**Requirements**: CICD-01, CICD-02, CICD-03, CICD-04, CICD-05, CICD-06
**Success Criteria** (what must be TRUE):
  1. Pushing a `v*.*.*` tag triggers `release.yml` and a GitHub release draft is created automatically
  2. macOS binary is a confirmed universal fat binary (verified via `lipo -info`) built on `macos-14`
  3. Windows binary preserves `-windowsconsole` and is built on `windows-latest`
  4. Linux binary builds on `ubuntu-22.04` for x64 and arm64 without WebKit dependency errors
  5. Release assets are only uploaded once (fan-in job completes after all platform builds); all third-party actions are SHA-pinned
**Plans**: TBD

### Phase 14: Platform Packaging
**Goal**: Every supported platform produces an installable package attached to the GitHub release automatically
**Depends on**: Phase 13
**Requirements**: PKG-01, PKG-02, PKG-03, PKG-04
**Success Criteria** (what must be TRUE):
  1. macOS release asset is a signed-layout DMG (unsigned binary inside) installable by drag-and-drop
  2. Windows release asset is an NSIS installer `.exe` that installs StorCat to Program Files on Windows
  3. Linux release includes an AppImage for x64 that launches without system WebKit dependency
  4. Linux release includes a `.deb` package installable via `dpkg -i` on x64 and arm64
**Plans**: TBD

### Phase 15: Distribution Channel Automation
**Goal**: Homebrew and WinGet package indexes update themselves on every release with no manual steps
**Depends on**: Phase 14
**Requirements**: DIST-01, DIST-02, DIST-03
**Success Criteria** (what must be TRUE):
  1. After a release tag is pushed, `homebrew-storcat` cask is updated with new version and locally-computed SHA256 (no CDN re-download) within minutes, and `brew upgrade storcat` installs the new version
  2. After a release tag is pushed, a PR is auto-submitted to `microsoft/winget-pkgs` with the new version manifest
  3. `packaging/winget/` manifests in main repo are auto-updated with the new version on release
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
| 12. Repo Consolidation | v2.2.0 | 3/3 | Complete   | 2026-03-27 |
| 13. CI Scaffold and Multi-Platform Build | v2.2.0 | 0/? | Not started | - |
| 14. Platform Packaging | v2.2.0 | 0/? | Not started | - |
| 15. Distribution Channel Automation | v2.2.0 | 0/? | Not started | - |
