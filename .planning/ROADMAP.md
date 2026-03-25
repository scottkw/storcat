# Roadmap: StorCat v2.0.0 — Go/Wails Migration

## Overview

This milestone fixes correctness gaps between the Electron v1.2.3 backend and the Go/Wails backend already in `feature/go-refactor-2.0.0-clean`. Work flows strictly bottom-up through the dependency graph: data models first, then services, then config, then the App IPC surface (including lifecycle wiring), then the TypeScript compatibility shim, then platform-specific integration, and finally cross-platform verification and a clean merge to main. Each phase fully completes its layer before the next begins, preventing rework.

## Phases

**Phase Numbering:**
- Integer phases (1, 2, 3): Planned milestone work
- Decimal phases (2.1, 2.2): Urgent insertions (marked with INSERTED)

Decimal phases appear between their surrounding integers in numeric order.

- [x] **Phase 1: Data Models + Catalog Service** - Fix Go structs and JSON output to match Electron's format (completed 2026-03-24)
- [x] **Phase 2: Search Service + Browse Metadata** - Fix browse metadata fields and timestamp semantics (completed 2026-03-25)
- [ ] **Phase 3: Config Manager** - Add window state persistence fields and methods to config
- [ ] **Phase 4: App Layer + Lifecycle** - Complete IPC surface, add LoadCatalog, wire OnDomReady/OnBeforeClose
- [ ] **Phase 5: Frontend Shim** - Update wailsAPI.ts to use new bindings and construct missing envelopes
- [ ] **Phase 6: Platform Integration** - Fix macOS drag region and version sourcing
- [ ] **Phase 7: Verification + Merge** - Cross-platform verification and clean merge to main

## Phase Details

### Phase 1: Data Models + Catalog Service
**Goal**: Go data structures and catalog output match Electron's format exactly, so all downstream layers build on a correct foundation
**Depends on**: Nothing (first phase)
**Requirements**: DATA-01, DATA-02, DATA-03, DATA-04, DATA-05, CATL-02, CATL-03, CATL-04
**Success Criteria** (what must be TRUE):
  1. A generated catalog JSON file opens as a bare object `{...}`, not an array `[{...}]`, and is readable by the v1.0 frontend
  2. An empty directory in the catalog serializes its `contents` field as `[]`, not `null` or absent, so frontend `.map()` calls do not panic
  3. The HTML catalog output renders the root node with `└──` connector and size bracket, matching Electron's tree appearance
  4. `CreateCatalog` returns `fileCount`, `totalSize`, and output paths so the frontend can display post-creation metadata
  5. Traversing a directory containing symlinks produces the same entry count as Electron (symlinks followed, not skipped)
**Plans:** 2/2 plans complete
Plans:
- [x] 01-01-PLAN.md — Fix catalog models, JSON output, traversal, HTML root, CreateCatalog return type, and Wails bindings
- [x] 01-02-PLAN.md — Add browse metadata fields (size, RFC3339 dates, birth time) to CatalogMetadata

### Phase 2: Search Service + Browse Metadata
**Goal**: The browse path returns complete, correctly-typed metadata so file listings display size and dates without type errors
**Depends on**: Phase 1
**Requirements**: CATL-01, DATA-03, DATA-04, DATA-05
**Success Criteria** (what must be TRUE):
  1. The Browse tab displays a `size` column with byte values for each catalog entry
  2. The `modified` field in browse results is a Date-compatible RFC3339 string, not an opaque Go time string
  3. `LoadCatalog` (new Go method) reads and parses a catalog JSON file and returns its contents to the frontend
**Plans:** 1/1 plans complete
Plans:
- [x] 02-01-PLAN.md — LoadCatalog method (TDD), wailsAPI wrapper, Browse tab Size column

**Note on DATA-03/DATA-04/DATA-05 dual coverage:** These model fields are defined in Phase 1 structs but their correctness is observable only through browse output, so verification belongs in Phase 2.

### Phase 3: Config Manager
**Goal**: Window state persistence has a complete config foundation so the App layer can read and write size, position, and toggle without stubs
**Depends on**: Phase 1
**Requirements**: WIN-01, WIN-02, WIN-03
**Success Criteria** (what must be TRUE):
  1. Config file on disk contains `windowWidth`, `windowHeight`, `windowX`, `windowY`, and `windowPersistenceEnabled` fields after a settings change
  2. `GetWindowPersistence` and `SetWindowPersistence` config methods exist and round-trip correctly (set true → read back true)
  3. Window persistence toggle in settings writes to the config file (not a no-op stub)
**Plans:** 2 plans
Plans:
- [ ] 03-01-PLAN.md — TDD: Config struct extension with WindowX, WindowY, WindowPersistenceEnabled fields and methods
- [ ] 03-02-PLAN.md — App bound methods, lifecycle hooks (domReady/beforeClose), frontend stub replacement

### Phase 4: App Layer + Lifecycle
**Goal**: The complete IPC surface is exposed and the Wails lifecycle hooks save and restore window state, eliminating all runtime panics and missing-method errors
**Depends on**: Phase 2, Phase 3
**Requirements**: API-01, API-02, API-03, WIN-04
**Success Criteria** (what must be TRUE):
  1. `GetCatalogHtmlPath` checks file existence before returning and always responds with a `{success, htmlPath}` shaped value (never a bare string or Promise rejection on missing file)
  2. `ReadHtmlFile` always responds with a `{success, content}` shaped value; a missing file returns `{success: false}`, not a Promise rejection or `null`
  3. Window size restores to its last-saved value on app relaunch (verified by resizing, quitting, and relaunching)
  4. All IPC wrappers consistently return `{success, ...}` envelopes; no method returns a raw value or throws unhandled rejections
**Plans**: TBD
**UI hint**: yes

### Phase 5: Frontend Shim
**Goal**: The TypeScript compatibility layer reflects all new Go method signatures and constructs the two missing IPC envelopes, so the React components receive the shapes they expect
**Depends on**: Phase 4
**Requirements**: API-01, API-02, API-03
**Success Criteria** (what must be TRUE):
  1. `wailsAPI.createCatalog` forwards `fileCount`, `totalSize`, and path fields from the new `CatalogResult` return value
  2. `wailsAPI.getCatalogHtmlPath` returns `{success: true, htmlPath: "..."}` when the file exists and `{success: false}` when it does not
  3. `wailsAPI.readHtmlFile` returns `{success: true, content: "..."}` for a valid file and `{success: false}` for a missing one
  4. Wails bindings (`wailsjs/`) are regenerated from the updated Go structs before any TypeScript edits are made
**Plans**: TBD
**UI hint**: yes

**Note on API-01/API-02/API-03 dual coverage:** Phase 4 delivers the Go methods; Phase 5 delivers the TypeScript wrappers that make them observable in the browser. The requirements are satisfied only when both layers are complete. Phase 5 is the observable completion point.

### Phase 6: Platform Integration
**Goal**: macOS titlebar is draggable and the displayed version string is sourced from the build, not a hardcoded constant
**Depends on**: Phase 5
**Requirements**: PLAT-01, PLAT-02
**Success Criteria** (what must be TRUE):
  1. On macOS, clicking and dragging the application header moves the window (the drag region is functional)
  2. The version string shown in the app UI matches the version defined in the project's version source file, not a hardcoded Go constant
**Plans**: TBD
**UI hint**: yes

### Phase 7: Verification + Merge
**Goal**: The Go/Wails build is confirmed at feature parity with Electron v1.2.3 on all target platforms and merges cleanly to main with no bloat
**Depends on**: Phase 6
**Requirements**: REL-01, REL-02
**Success Criteria** (what must be TRUE):
  1. App builds and runs on macOS, Windows, and Linux with no crash on startup
  2. All three tabs (Create, Search, Browse) complete their primary operations without errors on macOS
  3. `git log` on main after merge shows no committed `node_modules/`, `build/bin/`, or archive files; `.gitignore` covers all build artifacts
  4. The merged main branch produces a clean `wails build` output with no unexpected warnings
**Plans**: TBD

## Progress

**Execution Order:**
Phases execute in numeric order: 1 → 2 → 3 → 4 → 5 → 6 → 7

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 1. Data Models + Catalog Service | 2/2 | Complete   | 2026-03-24 |
| 2. Search Service + Browse Metadata | 1/1 | Complete   | 2026-03-25 |
| 3. Config Manager | 0/2 | Not started | - |
| 4. App Layer + Lifecycle | 0/? | Not started | - |
| 5. Frontend Shim | 0/? | Not started | - |
| 6. Platform Integration | 0/? | Not started | - |
| 7. Verification + Merge | 0/? | Not started | - |
