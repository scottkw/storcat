# Project Retrospective

*A living document updated after each milestone. Lessons feed forward into future planning.*

## Milestone: v2.0.0 — Go/Wails Migration

**Shipped:** 2026-03-26
**Phases:** 7 | **Plans:** 11 | **Tasks:** 15

### What Was Built
- Complete Go/Wails backend replacing Electron/Node.js — 6 Go source files, Wails bindings, config manager
- Full IPC surface with 17 wailsAPI wrappers maintaining `window.electronAPI` contract
- Window state persistence (size, position, toggle) via Go config lifecycle hooks
- macOS platform integration (drag region, build-time version injection)
- Cross-platform build verification and clean merge to main

### What Worked
- Bottom-up phase ordering (data models → services → config → app layer → shim → platform → verify) prevented rework
- TDD approach in Phase 2 (LoadCatalog) and Phase 3 (Config) caught issues early
- Strict `{success,...}` envelope pattern made IPC debugging straightforward
- 3-day execution timeline for 7 phases — fast cadence maintained throughout
- Milestone audit before completion caught 11 tech debt items cleanly

### What Was Inefficient
- SUMMARY.md `requirements_completed` frontmatter was empty across all 11 summaries — template gap never caught during execution
- Phase 7 Nyquist VALIDATION.md left in draft — validation workflow wasn't enforced for the final verification phase
- Some one-liner fields in SUMMARY.md were empty, degrading automated accomplishment extraction

### Patterns Established
- `{success,...}` envelope pattern for all IPC/API boundaries
- Dual-format parsing (v1 array + v2 bare object) for backward compatibility
- OnDomReady (not OnStartup) for window state restoration in Wails
- `//go:embed wails.json` as simpler alternative to ldflags for version injection
- `--wails-draggable: drag` CSS property for macOS titlebar

### Key Lessons
1. Plan every phase with observable success criteria — the "what must be TRUE" format caught ambiguity early
2. Template gaps (empty frontmatter fields) compound silently — validate template compliance at phase completion
3. Advisory tech debt items are fine to ship with — the audit surfaced them, and none blocked the release

### Cost Observations
- Model mix: Primarily Opus for planning/execution, Sonnet for subagents
- Sessions: ~5 sessions across 3 days
- Notable: Phase 5 (Frontend Shim) was single-plan, single-file — minimal overhead

---

## Milestone: v2.1.0 — CLI Commands

**Shipped:** 2026-03-26
**Phases:** 4 | **Plans:** 7

### What Was Built
- Full CLI subcommand framework in `cli/` package with stdlib flag.FlagSet dispatch
- 6 subcommands: create, search, list, show, open, version — all with --help and proper exit codes
- Table output (tablewriter) and --json flag for machine-readable output on data commands
- Colorized tree rendering with --depth N, --no-color, and NO_COLOR env var support
- Cross-platform browser launch for `storcat open`
- Platform compatibility: Windows console attachment, macOS -psn_* filtering, install script

### What Worked
- Phase 8 stub pattern: create dispatch skeleton with stubs, then replace each stub in subsequent phases — clean incremental approach
- Milestone audit between Phase 10 and Phase 11 caught 6 tech debt items before shipping
- Phase 11 as explicit tech debt cleanup phase resolved 5/6 items with surgical commits
- All 23 requirements verified via 3-source cross-reference (VERIFICATION.md, REQUIREMENTS.md traceability, code inspection)
- Same-day execution for all 4 phases — minimal overhead, focused scope

### What Was Inefficient
- SUMMARY.md one_liner field still not populated in Phase 10/11 summaries — same gap as v2.0.0
- Phase 11 Nyquist VALIDATION.md missing — same gap pattern as v2.0.0 Phase 7 (cleanup/verification phases get skipped)
- Some decisions in STATE.md accumulated as "pending todos" that were already resolved — stale context

### Patterns Established
- Interspersed flag parsing: pre-separate positional args from flags before flag.Parse
- `cli/output.go` shared helpers (printJSON, formatBytes) across all data commands
- TDD red-green for each command: failing tests first, then implementation
- fs.Usage override pattern for custom --help output on flag.FlagSet

### Key Lessons
1. Stub pattern for CLI commands is excellent — provides compile-time safety while enabling incremental delivery
2. Audit-then-cleanup as a two-step flow works well — audit identifies, dedicated phase resolves
3. SUMMARY.md template compliance remains the biggest recurring gap — needs enforcement in execution workflow

### Cost Observations
- Model mix: Primarily Opus for planning/execution, Sonnet for subagents
- Sessions: ~3 sessions in 1 day
- Notable: 4 phases in a single day — CLI commands were well-scoped with clear boundaries

---

## Milestone: v2.2.0 — Repo Consolidation & CI/CD

**Shipped:** 2026-03-27
**Phases:** 4 | **Plans:** 7 | **Tasks:** 11

### What Was Built
- Repo consolidation: WinGet manifests and Homebrew cask template moved from satellite repos into main repo `packaging/`
- `winget-storcat` archived; `homebrew-storcat` marked auto-managed
- `release.yml`: tag-triggered workflow with 4 parallel platform builds (macOS universal, Windows, Linux x64+arm64) and fan-in draft release
- Platform packaging: macOS DMG (create-dmg), Windows NSIS installer, Linux AppImage + .deb
- `distribute.yml`: auto-updates Homebrew cask (SHA256-verified) and submits WinGet PR on release publish

### What Worked
- Research-first approach for each phase identified critical gotchas before planning (NSIS runner requirements, WebKit bundling limitations, fan-in race condition)
- Linear phase dependency chain (12→13→14→15) naturally built on previous work without rework
- SHA-pinning all actions upfront avoided supply chain concerns
- Nyquist validation achieved COMPLIANT status across all 4 phases — improvement over prior milestones
- Milestone audit with `tech_debt` status correctly identified all 7 informational items without blocking the release

### What Was Inefficient
- Some SUMMARY.md one-liner fields still empty (12-01, 14-01, 15-02) — same recurring gap from v2.0.0 and v2.1.0
- Stale checkboxes in REQUIREMENTS.md (REPO-01 through REPO-04) only caught during audit — should be updated during execution
- Several decisions accumulated in STATE.md that were resolved but not cleared

### Patterns Established
- `release:published` trigger for distribution (not tag push) — ensures draft review before public distribution
- SHA-pinned actions with comment annotation for human readability
- Fan-in pattern: platform builds upload artifacts → single release job downloads and attaches all
- Template-based packaging: `storcat.rb.template` with sed substitution for version/SHA256

### Key Lessons
1. CI/CD workflows can't be tested locally — plan for human verification items and document them upfront
2. Runner image deprecation (NSIS on Windows 2025, WebKit on Ubuntu 24.04) makes pinning runner versions essential
3. Cross-repo automation requires classic PATs — `GITHUB_TOKEN` is scoped to the current repo only
4. First WinGet submission must be manual — automation only works after the package exists in the registry

### Cost Observations
- Model mix: Primarily Opus for planning/execution, Sonnet for subagents
- Sessions: ~2 sessions in 1 day
- Notable: All 4 phases completed in a single day — CI/CD work is mostly workflow YAML, minimal Go code changes

---

## Milestone: v2.3.0 — Code Signing & Package Manager CLI

**Shipped:** 2026-03-28
**Phases:** 6 | **Plans:** 8 | **Tasks:** 12

### What Was Built
- macOS Developer ID code signing + notarization + stapling automated in CI (Gatekeeper-verified end-to-end)
- Windows Authenticode signing pipeline with SSL.com eSigner cloud HSM integration (code complete, awaiting credentials)
- GitHub `release` environment with `v*.*.*` tag protection rules and 6 Apple signing secrets
- Credential rotation runbook for all signing certificates
- Homebrew cask `binary` stanza for immediate CLI availability after `brew install --cask storcat`
- Custom NSIS installer with EnVar plugin PATH registration for Windows CLI
- release-please automation: conventional commits → version bumps → tags → builds → publish → Homebrew/WinGet distribution

### What Worked
- Phase 16 as hard prerequisite (credentials before code) prevented wasted CI iteration on signing phases
- Phases 17 and 18 independent of each other — could have been parallelized (were sequential but fast)
- Research-first approach again proved essential for unfamiliar domains (code signing, notarization, NSIS scripting)
- Phase 21 (release-please) added mid-milestone — flexible roadmap evolution worked well
- Milestone audit caught AUTOREL orphaned requirements and Windows pipeline cascade issue before shipping
- All signing code verified in isolation — clean separation of "code complete" from "credentials provisioned"

### What Was Inefficient
- SUMMARY.md one-liner fields still not populated for some phases (17-01, 18-01) — same recurring gap
- Phase 21 requirements (AUTOREL-01 through AUTOREL-05) never added to REQUIREMENTS.md traceability table after phase was added to roadmap
- release-please-action@v4 uses tag reference instead of SHA pin — inconsistent with security posture of all other actions
- credential-rotation.md missing APPLE_ID (10th secret added during Phase 17, not backported to runbook)

### Patterns Established
- `env:` block for secret mapping in GitHub Actions run steps (prevents log leakage)
- `apple-actions/import-codesign-certs@v6` for keychain ACL setup (prevents codesign hangs)
- EnVar plugin over EnvVarUpdate for NSIS PATH (no 1024-byte truncation)
- release-please with `simple` release-type + `extra-files` jsonpath for non-npm version sources
- softprops/action-gh-release `tag_name` trick to upload to existing release-please release

### Key Lessons
1. Code signing credentials are ops work, not code work — separating "pipeline code complete" from "credentials provisioned" is the right abstraction
2. When adding a phase mid-milestone, update REQUIREMENTS.md traceability immediately — orphaned requirements compound
3. Windows signing pipeline cascade (one build failure blocks all distribution) needs mitigation — consider conditional release job
4. Bundling small binaries in-repo (EnVar.dll, 9KB) is preferable to network-dependent CI downloads

### Cost Observations
- Model mix: Primarily Opus for planning/execution, Sonnet for subagents
- Sessions: ~3 sessions in 1 day
- Notable: 6 phases completed in a single day — CI/CD and signing work is mostly YAML and config

---

## Cross-Milestone Trends

### Process Evolution

| Metric | v2.0.0 | v2.1.0 | v2.2.0 | v2.3.0 |
|--------|--------|--------|--------|--------|
| Phases | 7 | 4 | 4 | 6 |
| Plans | 11 | 7 | 7 | 8 |
| Tasks | 15 | 9 | 11 | 12 |
| Timeline | 3 days | 1 day | 1 day | 1 day |
| Requirements | 20/20 | 23/23 | 17/17 | 17/20 (3 deferred) |
| Tech Debt Items | 11 | 6 (5 resolved) | 7 (0 critical) | 4 (all credential-related) |
| Nyquist | Partial | Partial | COMPLIANT | Partial (3 phases) |

### Recurring Themes
- Bottom-up dependency ordering is effective for migration work
- Nyquist validation regressed in v2.3.0 (partial) after v2.2.0 achieved full compliance — credential-heavy milestones with human-action checkpoints are harder to validate
- SUMMARY.md one_liner/requirements_completed fields still inconsistently populated — template gap persists across all 4 milestones
- Audit-then-cleanup two-step is a reliable pattern for shipping quality
- Research-first approach for unfamiliar domains (CI/CD, packaging, code signing) prevents costly rework
- Same-day milestone execution is achievable with well-scoped phases and clear success criteria
- Mid-milestone phase additions (Phase 21) work well but require immediate REQUIREMENTS.md updates to avoid orphaned requirements
