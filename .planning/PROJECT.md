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

### Active

- [ ] Merge WinGet manifests into main repo under `packaging/winget/`
- [ ] Move Homebrew template and update script into main repo under `packaging/homebrew/`
- [ ] GitHub Actions release workflow builds all platforms on tag/release
- [ ] Full installer packaging: DMG (macOS), MSI (Windows), AppImage+deb (Linux)
- [ ] Auto-update `homebrew-storcat` repo on release via GitHub Action
- [ ] Auto-generate WinGet manifests on release
- [ ] Archive `winget-storcat` repo after migration

### Out of Scope

- Wails v3 migration — still alpha as of March 2026, premature
- Automated test suite — important, separate milestone (TEST-01 through TEST-03)
- Code signing (macOS notarization + Windows Authenticode) — future milestone, needs credential pipeline
- Performance benchmarking — Go is already faster, no formal benchmarks needed
- Tailwind CSS migration — different design direction, not prioritized

## Current Milestone: v2.2.0 Repo Consolidation & CI/CD

**Goal:** Consolidate three repos into one (+ thin Homebrew satellite), automate builds and packaging via GitHub Actions.

**Target features:**
- Merge WinGet manifests into main repo
- Move Homebrew template + update script into main repo
- GitHub Actions release workflow with full installer packaging
- Auto-update homebrew-storcat and winget manifests on release
- Archive winget-storcat repo

## Current State

**Shipped:** v2.1.0 CLI Commands (2026-03-26)

StorCat is a fully functional cross-platform desktop + CLI application. The unified binary supports both GUI mode (`storcat` with no args) and 6 CLI subcommands: `create`, `search`, `list`, `show`, `open`, `version`. Three repos exist: main source, winget manifests, homebrew cask — to be consolidated.

## Context

Shipped v2.1.0 on 2026-03-26 — added full CLI subcommand support to the Go/Wails binary.
Previously shipped v2.0.0 on 2026-03-26 — complete backend rewrite from Electron/Node.js to Go/Wails.

**Codebase:** ~2,500 LOC Go + ~2,700 LOC TypeScript
**Tech stack:** Go 1.23, Wails v2, React 18, TypeScript 5, Ant Design 5
**Dependencies:** tablewriter v1.1.4, fatih/color v1.18.0, pkg/browser
**Platforms:** macOS (universal), Windows (x64/arm64), Linux (x64/arm64)
**Build:** `wails build` with ldflags version injection

**Known tech debt:** All v2.1.0 audit items resolved. 1 non-blocking procedural item (SUMMARY.md frontmatter convention). Remaining v2.0.0 items tracked in `.planning/milestones/v2.0.0-MILESTONE-AUDIT.md`.

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
*Last updated: 2026-03-27 after v2.2.0 milestone start*
