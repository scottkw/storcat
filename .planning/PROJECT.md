# StorCat v2.0.0 — Go/Wails Migration

## What This Is

StorCat is a cross-platform desktop application for creating, browsing, and searching directory catalogs. It generates JSON and HTML representations of directory trees. v2.0.0 is a complete backend rewrite from Electron/Node.js to Go/Wails, keeping the same React/TypeScript/Ant Design frontend.

## Core Value

The Go/Wails v2.0.0 release must have exact feature parity with the Electron v1.2.3 release — no regressions, no missing functionality. Users switching from v1 to v2 should notice only improvements (speed, size), never missing features.

## Requirements

### Validated

Existing capabilities confirmed working in the Go/Wails branch:

- ✓ Create catalog (JSON + HTML) from directory tree — existing
- ✓ Search across multiple catalogs (case-insensitive, substring) — existing
- ✓ Browse catalog files with metadata — existing
- ✓ HTML catalog modal viewer with dark mode injection — existing
- ✓ 11 themes (StorCat Light/Dark, Dracula, Solarized, Nord, One Dark, Monokai, GitHub, Gruvbox) — existing
- ✓ Collapsible sidebar with left/right positioning — existing
- ✓ 3-tab navigation (Create, Search, Browse) — existing
- ✓ ModernTable with sort, filter, resize, pagination — existing
- ✓ AppContext state management via useReducer — existing
- ✓ localStorage persistence of last-used directories and settings — existing
- ✓ v1.0 catalog backward compatibility (both JSON formats) — existing
- ✓ Wails API wrapper maintaining window.electronAPI interface — existing
- ✓ Cross-platform builds (macOS, Windows, Linux) — existing
- ✓ LoadCatalog method with dual-format parsing (v1 array + v2 bare object) — Phase 2
- ✓ Browse tab Size column with human-readable byte formatting — Phase 2
- ✓ Browse metadata fields (size, modified, created) verified — Phase 1+2
- ✓ createCatalog returns fileCount, totalSize, and all path fields via shim wrapper — Phase 5
- ✓ getCatalogHtmlPath returns `{success, htmlPath}` envelope with file existence check — Phase 4+5
- ✓ readHtmlFile returns `{success, content}` envelope — Phase 4+5

### Active

- [ ] Fix symlink handling — Go skips symlinks via `os.Lstat`, must follow them like Electron's `fs.stat`
- [x] Implement full window state persistence — size + position save/restore + settings toggle — Phase 3
- [ ] Fix header drag region — `WebkitAppRegion: 'drag'` missing for macOS titlebar
- [ ] Fix HTML root node rendering — match Electron's tree format
- [ ] Fix version source — read from config/package.json instead of hardcoded constant
- [ ] Clean merge to main — verified, no bloat, proper .gitignore

### Out of Scope

- New features beyond Electron parity — future milestones
- Tailwind CSS migration — different design direction, not this milestone
- macOS code signing/notarization setup — separate concern, may need Apple credentials
- Automated test suite — important but separate milestone
- Performance benchmarking — Go is already faster, no need to formalize

## Context

- **Current state:** Clean branch `feature/go-refactor-2.0.0-clean` created from main with Go/Wails code, Electron files removed, bloat stripped (node_modules, binaries, archive)
- **Original Go branch:** `origin/feature/go-refactor-2.0.0` has the source code but with 415MB of committed bloat
- **Codebase map:** `.planning/codebase/` contains 7 analysis documents from the Electron version
- **Comparison research:** Full API, frontend component, and backend logic comparisons completed in this session
- **Go backend:** 6 Go source files in `main.go`, `app.go`, `internal/`, `pkg/models/`
- **Frontend:** React/TypeScript under `frontend/src/` with Wails bindings in `frontend/wailsjs/`

## Constraints

- **Backward compatibility**: v1.0 catalog JSON/HTML files must remain readable
- **API surface**: `window.electronAPI` interface must be maintained — frontend components depend on it
- **Branch**: Work on `feature/go-refactor-2.0.0-clean`, merge to `main` when complete
- **No Electron dependencies**: All fixes must use Go/Wails patterns, not Electron APIs

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Go/Wails over Electron | 93% smaller binaries, 5x faster search, no V8/ARM64 issues | — Pending |
| JSON format: bare object `{...}` | Backward compatibility with existing catalogs and external tools | — Pending |
| Full window persistence via Go config | Feature parity with Electron version, user expectation | — Pending |
| Clean branch (not filter-branch) | Simpler, avoids rewriting git history | ✓ Good |
| Strip node_modules/binaries/archive | 415MB of bloat prevented from entering main | ✓ Good |

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
*Last updated: 2026-03-26 after Phase 5 completion — Frontend Shim verified (all wailsAPI wrappers return correct envelopes with full field forwarding)*
