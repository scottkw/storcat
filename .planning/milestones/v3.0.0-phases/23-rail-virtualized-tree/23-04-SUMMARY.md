---
phase: 23-rail-virtualized-tree
plan: 04
subsystem: rail-and-status-bar
tags: [react, useDeferredValue, localStorage, css]

requires:
  - phase: 23-01
    provides: "AppContext's catalogs/currentCatalogId/tree reducer state and the atomic SELECT_CATALOG action"
  - phase: 23-02
    provides: "CatalogMetadata.fileCount/.totalBytes/.parseError -- the three fields this plan renders"
  - phase: 23-03
    provides: "frontend/src/lib/format.ts (formatBytes/formatCount/formatGB), the phase's fixed formatter module"
provides:
  - "CatalogRail's populated body -- two-line rows, status dot, filter (never in AppContext), interactive directory chip, zero-match line, empty-library gating"
  - "StatusBar's live three-segment body, derived from the rail array with no binding of its own"
  - "SET_CATALOG_DIR extended to clear currentCatalogId/tree/expanded/selected in one reducer case"
  - "themeTokens.ts' safeGetItem/safeSetItem exported for reuse by any module persisting its own storcat-* key"
affects: [23-05, 23-06]

actuals:
  tokens: 3800
  tasks: 3
  commits: 3

tech-stack:
  added: []
  patterns:
    - "Filter string lives in local useState, read through useDeferredValue for the matching pass -- never dispatched into AppContext, so the tree has no reason to re-render on a keystroke (verified empirically via a MutationObserver on the tree's scroll container, not just by code inspection)"
    - "A CSS class rule (not an inline style) owns a row's hover/selected background, with the class rule declared after :hover in source order so the selected state wins the tie -- same discipline as Phase 22's New-pill fix and 23-03's tree-row rule"
    - "A row's accent left border is reserved at 2px transparent on every row (not added only when selected) so a selection never shifts row content -- the same 'never remove from the DOM, only recolor' discipline the always-present status dot uses"

key-files:
  created: []
  modified:
    - frontend/src/components/workspace/CatalogRail.tsx
    - frontend/src/components/workspace/StatusBar.tsx
    - frontend/src/contexts/AppContext.tsx
    - frontend/src/themeTokens.ts
    - frontend/src/workspace.css

key-decisions:
  - "Directory chip converted from a <div> to a <button> (not just given an onClick) -- the click behavior and the aria-label needed a focusable, keyboard-activatable element, and the plan's 'visual footprint is unchanged' constraint is satisfied by keeping the same inline styles rather than by keeping the same tag"
  - "Directory-chip hover gets its own .ws-dir-chip:hover rule (border-color only) instead of reusing the existing .ws-chip:hover class, because .ws-chip also recolors text -- 23-UI-SPEC explicitly calls for 'hover border only, matching the search-field convention', and .ws-search:hover already establishes that narrower pattern for a different element"
  - "themeTokens.ts' safeGetItem/safeSetItem exported rather than duplicated -- CatalogRail needed the exact same try/catch-wrapped localStorage access themeTokens.ts already established for the other three storcat-* keys"
  - "SET_CATALOG_DIR clears exactly the four fields the plan's action text names (currentCatalogId, tree, expanded, selected) and deliberately does not clear state.catalogs -- the rail's own SET_CATALOGS dispatch (issued right after, in the same handler) is what replaces the list; over-clearing here was not asked for and was not added"

patterns-established: []

requirements-completed: [RAIL-01, RAIL-02, RAIL-03, RAIL-04, RAIL-05, STATE-01, SHELL-06]

coverage:
  - id: D1
    description: "Two-line rail rows (title/JSON-size, filename/file-count) with an always-in-DOM status dot, identical row heights whether or not a catalog is broken, and full click/keyboard selectability on broken rows"
    requirement: "RAIL-01, RAIL-04"
    verification:
      - kind: automated_ui
        ref: "dev-browser at :34115 against 3 fixture catalogs (2 healthy + 1 deliberately corrupted): 3 rendered rows, dot backgroundColor rgb(229,83,75) on the corrupted row vs rgba(0,0,0,0) on healthy rows, all three rows measured height 53.5px; clicking the corrupted row set data-selected=true with --sel background, --ac left border and --ac title color -- screenshots t2304-rail-rows.png, t2304-broken-selected.png"
        status: pass
      - kind: automated_ui
        ref: "Cold sidecar cache: both healthy rows rendered filename alone with no count fragment; after opening each catalog once (warming the cache) and reloading, both rows showed '· 12 files' -- confirms absent-vs-zero and the cache self-heal on next rail load"
        status: pass
    human_judgment: false
  - id: D2
    description: "Filter isolated from AppContext, matched case-insensitively against title+filename, order-preserving, with a distinct zero-match line"
    requirement: "RAIL-02"
    verification:
      - kind: automated_ui
        ref: "dev-browser: MutationObserver attached to .ws-tree .pane-scroll before typing 'documents' into the filter recorded 0 mutations after typing completed; rail narrowed to 1 row (catalog-documents) while the already-loaded tree stayed at 12 rows unchanged; empty and whitespace-only filters both restored all 3 rows; uppercase 'BROKEN' matched catalog-broken case-insensitively; a filter matching nothing rendered exactly 'No catalogs match \"zzzznotfound\".'"
        status: pass
    human_judgment: false
  - id: D3
    description: "Directory chip persists a chosen directory and clears stale selection state; no-directory and unreadable-directory both degrade to the empty-library block with no console error"
    requirement: "RAIL-05, STATE-01"
    verification:
      - kind: automated_ui
        ref: "dev-browser: removing storcat-catalog-directory and reloading rendered the empty-library block (0 rows) with the chip reading its placeholder and no BrowseCatalogs call made; pointing the key at a non-existent path and reloading rendered the same empty-library block with only a pre-existing unrelated antd deprecation warning in the console, no error from this code path"
        status: pass
      - kind: manual
        ref: "The native SelectDirectory dialog itself cannot be driven from a browser session (FPA-23-04-B) -- the persistence+reload half was exercised end-to-end via the same code path (localStorage -> SET_CATALOG_DIR -> BrowseCatalogs) the chip's onClick calls"
        status: pass
    human_judgment: true
  - id: D4
    description: "Status bar's three segments are live, derived from the rail array with no binding of its own"
    requirement: "SHELL-06"
    verification:
      - kind: automated_ui
        ref: "dev-browser: after warming the counts cache, status bar read '3 catalogs 24 files indexed 0.1 GB' -- 3 matches the rail header count, 24 matches the sum of the two visible '12 files' rows, and the broken row's absent count was correctly excluded rather than treated as 0"
        status: pass
      - kind: other
        ref: "node -e against esbuild-transpiled format.ts: formatGB(1024 GB in bytes) = '1.0 TB', formatGB(1023.9 GB in bytes) = '1023.9 GB' (threshold inclusive at 1024); formatCount(0) = '0', formatCount(1500000) = '1.5M'; grep -c wailsjs StatusBar.tsx = 0"
        status: pass
    human_judgment: false

duration: 14min
completed: 2026-08-13
status: complete
---

# Phase 23 Plan 04: Rail Population, Filter, Directory Chip, Status Bar Summary

**The rail's two-line rows, always-present status dot, isolated filter (proven via a live MutationObserver to leave the tree's DOM untouched during typing), an interactive directory chip, and a status bar summing the same array the rows render -- all verified against a real fixture directory with one deliberately corrupted catalog, not asserted from code reading alone**

## Performance

- **Duration:** 14 min
- **Completed:** 2026-08-13
- **Tasks:** 3
- **Files modified:** 5 (0 created, 5 modified)

## Accomplishments

- `CatalogRail.tsx` rows now render the full contract: a 6px status dot always in the DOM (transparent when healthy, `#e5534b` when `parseError` is non-empty), 12.5px title, 10.5px mono JSON size via `formatBytes`, and a second mono line reading `{filename} · {fileCount} files` -- with the count fragment omitted entirely (not zeroed) when the sidecar cache hasn't computed it yet
- The filter is local `useState` read through `useDeferredValue`, matched case-insensitively against title+filename, never dispatched into `AppContext` -- proven, not assumed: a `MutationObserver` attached to the tree's scroll container recorded **zero mutations** while typing into the filter with a catalog already loaded
- The directory chip is now a real `<button>`: clicking it calls `SelectDirectory`, persists the chosen path under `storcat-catalog-directory` via `themeTokens.ts`'s (now-exported) safe localStorage helpers, dispatches `SET_CATALOG_DIR`, and re-runs `BrowseCatalogs`
- `SET_CATALOG_DIR` now clears `currentCatalogId`, resets `tree` to idle, and empties `expanded`/`selected` in the same reducer case, so a directory change can't leave the tree pointing at a catalog no longer in the (not-yet-reloaded) list
- `StatusBar.tsx` derives all three segments (`{n} catalogs`, `{sum} files indexed`, `{sum} GB/TB`) from `state.catalogs` in one `useMemo`, skipping absent counts rather than treating them as zero, with no binding of its own (`grep -c wailsjs` returns 0)

## Task Commits

Each task was committed atomically:

1. **Task 1: Populated rail rows -- two lines, a status dot that never moves the layout, and selection** - `49e45acf` (feat)
2. **Task 2: A filter that leaves the tree alone, the directory chip, and the two empty states** - `b3f24cb1` (feat)
3. **Task 3: The status bar goes live** - `ed1d3abd` (feat)

## Files Created/Modified

- `frontend/src/components/workspace/CatalogRail.tsx` - full row markup, filter wiring, interactive directory chip, zero-match/empty-library states
- `frontend/src/components/workspace/StatusBar.tsx` - three derived segments in one `useMemo`
- `frontend/src/contexts/AppContext.tsx` - `SET_CATALOG_DIR` extended to clear catalog/tree/expansion/selection together
- `frontend/src/themeTokens.ts` - `safeGetItem`/`safeSetItem` exported (were module-private) for reuse by `CatalogRail.tsx`
- `frontend/src/workspace.css` - `.ws-rail-row` full row/hover/selected/dot/title/size/meta rules, `.ws-dir-chip:hover` (border-only)

## Decisions Made

- Directory chip converted from a `<div>` to a `<button>` -- needed for real click handling, keyboard activation and an accessible name, while keeping the exact same inline visual footprint Phase 22 established
- `.ws-dir-chip:hover` is its own rule (border-color only), not a reuse of `.ws-chip:hover` (which also recolors text) -- 23-UI-SPEC explicitly specifies "hover border only, matching the search-field convention"
- Exported `themeTokens.ts`'s existing `safeGetItem`/`safeSetItem` rather than writing a second try/catch wrapper in `CatalogRail.tsx`
- `SET_CATALOG_DIR` clears exactly the four fields the plan's action text names and deliberately leaves `state.catalogs` untouched -- the rail's own `SET_CATALOGS` dispatch, issued immediately after in the same handler, is what replaces the list

## Deviations from Plan

None -- plan executed exactly as written. All three tasks, their file lists, and their acceptance criteria matched the plan.

## Issues Encountered

None. `wails dev` was already running; dev-browser's session persisted across all verification passes. A synthetic fixture directory (`/tmp/storcat-rail-fixtures`, 2 healthy catalogs copied from the 23-01 flat-fixture generator plus 1 hand-corrupted JSON file) was used for live verification and deleted afterward -- not committed to the repo.

## Known Stubs

None new. The directory chip's native `SelectDirectory` dialog itself cannot be driven from a browser automation session (FPA-23-04-B, recorded in the plan's own frontmatter) -- the persistence-and-reload half of that flow was verified end-to-end via the identical code path the dialog's callback uses (`localStorage` write -> `SET_CATALOG_DIR` -> `BrowseCatalogs`), which is the same resolution the plan itself anticipated, not a gap introduced here.

## User Setup Required

None -- no external service configuration required.

## Next Phase Readiness

- The rail's red dot and `CatalogMetadata.parseError` are ready for 23-05's unreadable-catalog panel to key off directly -- selecting a broken row already works today (this plan confirmed it dispatches `SELECT_CATALOG` like any other row); 23-05 adds the panel that replaces the tree pane on that selection.
- `StatusBar.tsx`'s live totals and the rail's populated rows give 23-06 (reveal-in-file-manager) and 23-05 (details panel) real data to read from `state.catalogs`/`state.tree`, unchanged.
- No blockers.

## Self-Check: PASSED

All 5 files claimed as modified confirmed present on disk with the expected content; all 3 task commit hashes (`49e45acf`, `b3f24cb1`, `ed1d3abd`) confirmed in `git log`.
