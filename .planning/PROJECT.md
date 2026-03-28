# StorCat

## What This Is

StorCat is a cross-platform desktop application for creating, browsing, and searching directory catalogs. It generates JSON and HTML representations of directory trees. Built with Go/Wails backend and React/TypeScript/Ant Design frontend.

## Core Value

Fast, lightweight directory catalog management — Go/Wails delivers 93% smaller binaries and 5x faster search than the original Electron version, with full feature parity.

## Requirements

### Validated

- ✓ Catalog JSON output: bare object format with v1 backward compatibility — v2.0.0
- ✓ Empty directory contents serialize as `[]`, never `null` — v2.0.0
- ✓ Browse metadata: size (bytes), modified (RFC3339), created (birth time) — v2.0.0
- ✓ LoadCatalog with dual-format parsing (v1 array + v2 bare object) — v2.0.0
- ✓ CreateCatalog returns full metadata (jsonPath, htmlPath, fileCount, totalSize) — v2.0.0
- ✓ Directory traversal follows symlinks (matching Electron behavior) — v2.0.0
- ✓ HTML catalog root node with `└──` connector and size bracket — v2.0.0
- ✓ All 17 wailsAPI wrappers return `{success,...}` envelopes — v2.0.0
- ✓ GetCatalogHtmlPath verifies file existence via os.Stat — v2.0.0
- ✓ ReadHtmlFile returns `{success, content}` envelope — v2.0.0
- ✓ Window size + position persistence via Go config — v2.0.0
- ✓ Window persistence settings toggle (not a stub) — v2.0.0
- ✓ Window state restores via OnDomReady hook — v2.0.0
- ✓ macOS header draggable via `--wails-draggable` — v2.0.0
- ✓ Version sourced at build time via ldflags — v2.0.0
- ✓ Clean merge to main with no bloat — v2.0.0
- ✓ .gitignore covers all build artifacts — v2.0.0
- ✓ 11 themes, collapsible sidebar, 3-tab navigation — existing
- ✓ ModernTable with sort, filter, resize, pagination — existing
- ✓ Cross-platform builds (macOS, Windows, Linux) — existing
- ✓ CLI subcommands in unified binary (`storcat` = GUI, `storcat <cmd>` = CLI) — v2.1.0
- ✓ `storcat create` — create catalog from directory — v2.1.0
- ✓ `storcat search` — search catalogs for a term — v2.1.0
- ✓ `storcat list` — list catalogs with metadata — v2.1.0
- ✓ `storcat show` — display catalog tree structure — v2.1.0
- ✓ `storcat open` — open catalog HTML in default browser — v2.1.0
- ✓ `storcat version` — print version — v2.1.0
- ✓ CLI dispatch, subcommand routing, `--help`, exit codes — v2.1.0
- ✓ macOS `-psn_*` filtering, Windows console output, install script — v2.1.0
- ✓ `--json` flag on list, search, create, show — v2.1.0
- ✓ `--depth N` flag on show — v2.1.0
- ✓ Colorized tree output with `--no-color` / `NO_COLOR` support — v2.1.0
- ✓ Cross-platform `storcat open` (macOS/Linux/Windows) — v2.1.0
- ✓ WinGet manifests consolidated in main repo under `packaging/winget/` — v2.2.0
- ✓ Homebrew cask template and update script in main repo under `packaging/homebrew/` — v2.2.0
- ✓ `winget-storcat` repo archived with redirect to main repo — v2.2.0
- ✓ `homebrew-storcat` marked auto-managed — v2.2.0
- ✓ Tag-triggered release workflow with 4-platform parallel builds and fan-in draft release — v2.2.0
- ✓ macOS DMG, Windows NSIS installer, Linux AppImage + .deb packaging — v2.2.0
- ✓ Homebrew cask auto-updated on release (SHA256-verified) — v2.2.0
- ✓ WinGet manifest auto-submitted on release — v2.2.0
- ✓ All GitHub Actions SHA-pinned — v2.2.0

### Active

- [ ] macOS Developer ID code signing automated in CI
- [ ] macOS notarization + stapling automated in CI
- [ ] Windows Authenticode code signing automated in CI
- [ ] Signing credentials securely stored in GitHub Actions secrets
- [ ] Homebrew installation puts `storcat` on PATH for CLI use
- [ ] WinGet installation puts `storcat` on PATH for CLI use

### Out of Scope

- Wails v3 migration — still alpha as of March 2026, premature
- Automated test suite — important, separate milestone (TEST-01 through TEST-03)
- ~~Code signing~~ — now active in v2.3.0
- Performance benchmarking — Go is already faster, no formal benchmarks needed
- Tailwind CSS migration — different design direction, not prioritized

## Current Milestone: v2.3.0 Code Signing & Package Manager CLI

**Goal:** Automate macOS/Windows code signing with secure credential handling, and make Homebrew/WinGet installations provide working CLI out of the box.

**Target features:**
- Apple Developer ID code signing + notarization + stapling in CI
- Windows Authenticode signing in CI
- Credential recovery/renewal guidance for both platforms
- GitHub Actions secrets with environment protection rules
- Homebrew cask installs `storcat` binary to PATH (CLI works immediately)
- WinGet installation puts `storcat` on PATH (CLI works immediately)

## Current State

**Shipped:** v2.2.1 bugfix (2026-03-27), v2.2.0 Repo Consolidation & CI/CD (2026-03-27)

StorCat is a fully functional cross-platform desktop + CLI application with automated CI/CD. The unified Go/Wails binary supports GUI mode (`storcat`) and 6 CLI subcommands (`create`, `search`, `list`, `show`, `open`, `version`). Three repos consolidated into one (main repo + thin `homebrew-storcat` satellite). Full CI/CD pipeline: `release.yml` triggers on `v*.*.*` tag push, builds on 4 platform runners (macOS universal, Windows, Linux x64+arm64) with fan-in draft release. Platform packaging: macOS DMG, Windows NSIS installer, Linux AppImage + .deb. Distribution automation: `distribute.yml` auto-updates Homebrew cask and submits WinGet PR on release publish.

## Context

Shipped v2.2.0 on 2026-03-27 — repo consolidation, CI/CD pipeline, platform packaging, and distribution automation.
Previously shipped v2.1.0 on 2026-03-26 — full CLI subcommand support.
Previously shipped v2.0.0 on 2026-03-26 — complete backend rewrite from Electron/Node.js to Go/Wails.

**Codebase:** ~2,500 LOC Go + ~2,700 LOC TypeScript + ~800 LOC YAML (CI/CD workflows)
**Tech stack:** Go 1.23, Wails v2, React 18, TypeScript 5, Ant Design 5
**Dependencies:** tablewriter v1.1.4, fatih/color v1.18.0, pkg/browser
**Platforms:** macOS (universal), Windows (x64/arm64), Linux (x64/arm64)
**Build:** `wails build` with ldflags version injection
**CI/CD:** GitHub Actions — `build.yml` (CI), `release.yml` (release), `distribute.yml` (distribution)

**Known tech debt:** 7 informational items from v2.2.0 audit (0 critical). See `.planning/milestones/v2.2.0-MILESTONE-AUDIT.md`. First WinGet submission to microsoft/winget-pkgs is a manual prerequisite.

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Go/Wails over Electron | 93% smaller binaries, 5x faster search, no V8/ARM64 issues | ✓ Good |
| JSON format: bare object `{...}` | Backward compatibility with existing catalogs and external tools | ✓ Good |
| Full window persistence via Go config | Feature parity with Electron version, user expectation | ✓ Good |
| Clean branch (not filter-branch) | Simpler, avoids rewriting git history | ✓ Good |
| Strip node_modules/binaries/archive | 415MB of bloat prevented from entering main | ✓ Good |
| OnDomReady for window restore (not OnStartup) | Window not yet rendered at OnStartup | ✓ Good |
| ldflags for version injection | Standard Go pattern, overridable at build time | ✓ Good |
| `//go:embed wails.json` for version | Simpler than ldflags, automatic from wails.json | ✓ Good (diverged from plan, functionally correct) |
| stdlib flag.FlagSet for CLI (no Cobra) | Zero deps, project decision CLIP-03 | ✓ Good |
| CLI dispatch before wails.Run() | Known subcommands exit early, unknown args fall through to GUI | ✓ Good |
| Dual-format LoadCatalog (array + object) | v1 backward compatibility with zero user friction | ✓ Good |
| `{success,...}` envelope pattern | Consistent error handling, matches Electron contract | ✓ Good |
| stdlib flag.FlagSet for CLI (no Cobra) | Zero deps, 6 subcommands don't justify ~2MB | ✓ Good |
| CLI dispatch before wails.Run() | Known subcommands exit early, unknown fall through to GUI | ✓ Good |
| tablewriter v1.1.4 for table output | Builder API, lightweight, good alignment | ✓ Good |
| fatih/color for tree rendering | stdlib-compatible, respects NO_COLOR | ✓ Good |
| pkg/browser for cross-platform open | Proven package, handles macOS/Linux/Windows | ✓ Good |
| -windowsconsole for Windows CLI | Simpler than AttachConsole, no runtime complexity | ✓ Good |
| Universal -psn_* filtering | One function on all platforms, no build tags | ✓ Good |
| Tag-push trigger (not release-publish) for builds | Creates release draft automatically, avoids chicken-and-egg | ✓ Good |
| SHA-pin all third-party actions | Supply chain security — no tag-only references | ✓ Good |
| Fan-in release pattern | Prevents race conditions on parallel release uploads | ✓ Good |
| windows-2022 for NSIS builds | NSIS removed from Windows Server 2025 image | ✓ Good |
| AppImage with system WebKit dependency | WebKit subprocess paths hardcoded, bundling not feasible for Wails | ✓ Good |
| release:published trigger for distribute.yml | Distribution fires only after manual draft promotion | ✓ Good |
| winget-releaser for WinGet submission | Maintained action, SHA-pinned to v2 | ✓ Good |

## Constraints

- **Backward compatibility**: v1.0 catalog JSON/HTML files must remain readable
- **API surface**: `window.electronAPI` interface maintained via wailsAPI shim
- **No Electron dependencies**: All functionality uses Go/Wails patterns

## Evolution

This document evolves at phase transitions and milestone boundaries.

**After each phase transition** (via `/gsd:transition`):
1. Requirements invalidated? → Move to Out of Scope with reason
2. Requirements validated? → Move to Validated with phase reference
3. New requirements emerged? → Add to Active
4. Decisions to log? → Add to Key Decisions
5. "What This Is" still accurate? → Update if drifted

**After each milestone** (via `/gsd:complete-milestone`):
1. Full review of all sections
2. Core Value check — still the right priority?
3. Audit Out of Scope — reasons still valid?
4. Update Context with current state

---
*Last updated: 2026-03-27 after v2.3.0 milestone start*
