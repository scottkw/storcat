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

### Active

<!-- v2.1.0 CLI Commands -->
- [ ] CLI subcommands in unified binary (`storcat` = GUI, `storcat <cmd>` = CLI)
- ✓ `storcat create` — create catalog from directory — Phase 9
- ✓ `storcat search` — search catalogs for a term — Phase 9
- ✓ `storcat list` — list catalogs with metadata — Phase 9
- [ ] `storcat show` — display catalog tree structure
- [ ] `storcat open` — open catalog HTML in default browser
- [ ] `storcat version` — print version
- ✓ CLI dispatch, subcommand routing, `--help`, exit codes — Phase 8
- ✓ macOS `-psn_*` filtering, Windows console output, install script — Phase 8

### Out of Scope

- Wails v3 migration — still alpha as of March 2026, premature
- Automated test suite — important, separate milestone (TEST-01 through TEST-03)
- macOS code signing/notarization — separate concern, needs Apple credentials
- Performance benchmarking — Go is already faster, no formal benchmarks needed
- Tailwind CSS migration — different design direction, not prioritized

## Current Milestone: v2.1.0 CLI Commands

**Goal:** Add CLI subcommands to the unified StorCat binary so users can create, search, browse, and view catalogs from the command line.

**Target features:**
- `storcat create` — create catalog (JSON + HTML) from a directory
- `storcat search` — search catalogs for a term
- `storcat list` — list catalogs in a directory with metadata
- `storcat show` — display a catalog's tree structure
- `storcat open` — open catalog HTML in default browser
- `storcat version` — print version
- `storcat` (no args) — launch GUI as today

## Context

Shipped v2.0.0 on 2026-03-26 — complete backend rewrite from Electron/Node.js to Go/Wails.

**Codebase:** ~1,700 LOC Go + ~2,700 LOC TypeScript
**Tech stack:** Go 1.23, Wails v2, React 18, TypeScript 5, Ant Design 5
**Platforms:** macOS (universal), Windows (x64/arm64), Linux (x64/arm64)
**Build:** `wails build` with ldflags version injection

**Known tech debt (11 items):**
- testData hardcoded fallbacks in Search/Browse tabs show fake rows on first load
- `if (result)` instead of `if (result.success)` in CreateCatalogTab getVersion
- Phase 7 Nyquist VALIDATION.md in draft status
- 3 human verification items outstanding (smoke test, cross-platform CI, clean build on main)
- See `.planning/milestones/v2.0.0-MILESTONE-AUDIT.md` for full list

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
*Last updated: 2026-03-26 after Phase 9 (core subcommands — create, list, search) complete*
