---
phase: 23-rail-virtualized-tree
plan: 05
subsystem: workspace-tree
tags: [react, css]

requires:
  - phase: 23-01
    provides: "FlatCatalog's own fileCount/totalBytes (state.tree, 'ready' status) and the TreeState union"
  - phase: 23-02
    provides: "CatalogMetadata.parseError (byte-offset-formatted syntax errors) and .hasHtml"
  - phase: 23-03
    provides: "formatBytes/formatCount/formatDate and TreePane's existing state routing"
  - phase: 23-04
    provides: "The rail's red status dot keyed on CatalogMetadata.parseError"
provides:
  - "TreeHeader.tsx -- catalog title, companion-file chips, metadata line, mounted only in the tree's ready state"
  - "UnreadableCatalogPanel.tsx -- STATE-02's inline, undismissable diagnostic (filename, byte offset, reason, parser, verbatim raw error)"
  - "TreePane's five-way state routing (empty library / unreadable / loading / empty catalog / rows), each mutually exclusive"
  - "wailsAPI.ts's extractErrorMessage() -- every catch block now reads the real Go error instead of silently downgrading to 'Unknown error'"
affects: [23-06]

actuals:
  tokens: 4068
  tasks: 2
  commits: 2

tech-stack:
  added: []
  patterns:
    - "A catalog can arrive broken by either of two independent routes -- the rail listing's own parseError (has a byte offset) or the tree load's own error (doesn't) -- and the richer one wins when both exist, never silently prefer whichever resolved last"
    - "Route to an error surface using data already known at selection time (parseError from the rail listing) rather than waiting for an async call to fail, so a known-broken catalog never flashes a loading state first"

key-files:
  created:
    - frontend/src/components/workspace/TreeHeader.tsx
    - frontend/src/components/workspace/UnreadableCatalogPanel.tsx
  modified:
    - frontend/src/components/workspace/TreePane.tsx
    - frontend/src/services/wailsAPI.ts

key-decisions:
  - "Title stays the base sans font (not mono) -- despite the UI-SPEC's top-level Typography table saying 'every new value this phase renders... is IBM Plex Mono', the per-section Catalog Header contract only marks the chips and metadata line mono, the prototype's own title span has no font-family override, and 23-04's rail row title (the closest precedent) is already non-mono. The general summary row is read as an overstatement; the specific section and established precedent govern."
  - "Header renders in BOTH ready branches (empty-catalog and rows), not just rows -- the task's own acceptance criteria explicitly require a zero-file catalog to render '0 files' through the formatter rather than being suppressed, so gating the header on nodes.length>0 would have been wrong despite 23-03's precedent of only mounting BreadcrumbBar in the rows branch"
  - "isUnreadable is computed from selectedCatalog.parseError OR tree.status==='error', checked immediately after the empty-library branch and before the loading branch -- a catalog already known broken from the rail listing routes straight to the panel without ever showing 'Reading catalog-broken…' first"
  - "The panel prefers the rail listing's parseError over the tree load's error message when both are present, because only the listing's detectParseError-produced string carries a byte offset; the load's wrapped Go error does not"

patterns-established:
  - "extractErrorMessage(error) in wailsAPI.ts -- Wails' generated bindings reject a Go error as a plain string, not an Error instance, so every wailsAPI catch block must check typeof error === 'string' before falling back to error?.message"

requirements-completed: [TREE-04, STATE-02]

coverage:
  - id: D1
    description: "Catalog header renders title (ellipsized, one line), a chip per existing companion file (no chip for a missing HTML companion), and a four-value metadata line using state.tree's exact fileCount/totalBytes -- only in the ready state, never while loading, never for a failed catalog"
    requirement: "TREE-04"
    verification:
      - kind: automated_ui
        ref: "dev-browser at :34115 against 3 fixture catalogs: catalog with HTML companion showed 2 chips (catalog-with-html.json, catalog-with-html.html) and metadata '2 files | 312B | 5M catalogued | modified 8/13/2026'; catalog without HTML showed exactly 1 chip and no second chip element in the DOM, metadata '1 files | 151B | 512B catalogued | modified 8/13/2026'; a zero-file catalog rendered the header alongside the empty-catalog message with metadata '0 files | 59B | 0B catalogued | modified 8/13/2026' -- screenshots t2305-header-html.png, t2305-header-nohtml.png"
        status: pass
      - kind: other
        ref: "cd frontend && npx tsc --noEmit && npm run build; grep -c formatBytes/hasHtml in TreeHeader.tsx both -ge 1"
        status: pass
    human_judgment: false
  - id: D2
    description: "Selecting a catalog with a non-empty parseError, or one whose load fails after selection, renders the unreadable-catalog panel (filename, byte offset or em dash, reason, parser attempted-or-not-reached, and the verbatim raw error) in place of the entire tree pane -- no header, breadcrumb, rows, or action buttons; the rail row keeps its red dot and stays selectable"
    requirement: "STATE-02"
    verification:
      - kind: automated_ui
        ref: "dev-browser: a deliberately truncated catalog (95 bytes, missing closing braces) rendered all four meta rows -- File: catalog-broken.json, Failed at: byte 95, Reason: 'unexpected end of JSON input', Parser: 'v2 object / v1 array' -- and the raw-error box read 'byte 95: unexpected end of JSON input' character-for-character matching the listing's own parseError; DOM contained no header, no .ws-crumb, no .ws-tree-row, and 0 button elements -- screenshot t2305-unreadable.png"
        status: pass
      - kind: automated_ui
        ref: "dev-browser: a catalog listed successfully then deleted from disk before being clicked (simulating deletion between listing and load) rendered Failed at: em dash, Reason: 'failed to read catalog file: open .../catalog-vanish.json: no such file or directory' (naming the missing file, not a syntax error), Parser: 'not reached', raw box showing the full wrapped Go error verbatim -- this path required fixing wailsAPI.ts's error extraction (see Deviations)"
        status: pass
      - kind: automated_ui
        ref: "dev-browser: selecting the broken catalog twice in a row left the rail dot at rgb(229, 83, 75) both times and re-rendered the identical panel text each time"
        status: pass
      - kind: other
        ref: "cd frontend && npx tsc --noEmit && npm run build; grep -c wordBreak UnreadableCatalogPanel.tsx -ge 1; grep -c 'maxHeight\\|overflowY' UnreadableCatalogPanel.tsx -eq 0"
        status: pass
    human_judgment: false

duration: 45min
completed: 2026-08-13
status: complete
---

# Phase 23 Plan 05: Catalog Header and the Unreadable-Catalog Panel Summary

**A loaded catalog announces its title, companion chips (never fabricated for a missing HTML file) and exact metadata drawn from the flat catalog itself, while a broken one gets an inline, undismissable diagnostic with a verbatim Go error -- verified against a real truncated catalog and a real deleted-file race, the latter of which exposed and fixed a silent error-message bug in the Wails bridge**

## Performance

- **Duration:** 45 min
- **Completed:** 2026-08-13
- **Tasks:** 2
- **Files modified:** 4 (2 created, 2 modified)

## Accomplishments

- `TreeHeader.tsx`: 17px ellipsized title, `.json`/`.html` mono chips (HTML chip omitted entirely -- not greyed -- when `hasHtml` is false), and a four-value mono metadata line drawn from `state.tree`'s exact `fileCount`/`totalBytes` (never the rail's cache-backed, possibly-cold `CatalogMetadata` fields); mounted in `TreePane` above the breadcrumb in both ready branches (empty-catalog and rows), never while loading
- `UnreadableCatalogPanel.tsx`: badge + headline, the UI-SPEC's fixed generic explanation, four meta rows (File/Failed at/Reason/Parser) computed from whichever of two sources is richer (the rail listing's byte-offset-bearing `parseError`, or the tree load's own error for the two cases the listing can't catch), and the verbatim raw-error box (`word-break: break-all`, no max-height, no scroll) -- zero action buttons
- `TreePane.tsx`: extended to five mutually exclusive states -- the unreadable check runs immediately after the empty-library gate and before the loading state, so a catalog already known broken from the rail listing routes straight to the panel without ever flashing "Reading catalog…" first
- Found and fixed a real bug in `wailsAPI.ts` during Task 2 verification: every catch block read `error.message`, but Wails' generated bindings reject a Go error as a plain JS string, not an `Error` instance, so `error.message` was always `undefined` and every real Go error (missing file, permission denied, parse failure) silently became the string `'Unknown error'` before reaching any caller

## Task Commits

Each task was committed atomically:

1. **Task 1: The catalog header -- title, companion chips, metadata line** - `e885f32d` (feat)
2. **Task 2: The unreadable-catalog panel -- a diagnostic the user can act on** - `7952c2de` (feat, includes the `wailsAPI.ts` error-extraction fix)

## Files Created/Modified

- `frontend/src/components/workspace/TreeHeader.tsx` - title, companion-file chips, metadata line
- `frontend/src/components/workspace/UnreadableCatalogPanel.tsx` - badge, headline, explanation, four meta rows, raw-error box
- `frontend/src/components/workspace/TreePane.tsx` - five-way state routing, mounts `TreeHeader` and `UnreadableCatalogPanel`
- `frontend/src/services/wailsAPI.ts` - `extractErrorMessage()` helper, applied to all 12 call sites that previously read `error.message` directly against a Wails rejection

## Decisions Made

- Title stays the base sans font, not mono -- the UI-SPEC's top-level Typography table's blanket "every new value... is IBM Plex Mono" is overridden by its own more specific Catalog Header section (which only marks the chips and metadata line mono), the prototype's title span has no font-family override, and 23-04's rail row title (the closest built precedent) is already non-mono
- Header renders in both ready branches (empty-catalog and rows), not gated to rows only -- the plan's own acceptance criteria require a zero-file catalog to show "0 files" through the formatter, which is unreachable if the header is suppressed alongside the empty-catalog message
- `isUnreadable` is checked before the loading branch, not after -- a catalog already known broken from the rail's `parseError` never needs to wait for `LoadCatalogFlat` to also fail before showing the diagnostic
- The panel prefers the rail listing's `parseError` over the tree load's error message when both exist, because only `detectParseError`'s string carries a real byte offset; the load's wrapped Go error never does

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking issue] `wailsAPI.ts` silently discarded every real Go error as `'Unknown error'`**
- **Found during:** Task 2, live verification of the "catalog deleted between listing and click" acceptance criterion
- **Issue:** Every `wailsAPI` method's catch block read `error.message || 'Unknown error'`. Manually invoking `window.go.main.App.LoadCatalogFlat` against a deleted file confirmed Wails rejects with a plain string (`"load catalog for flatten: failed to read catalog file: open ...: no such file or directory"`), not an `Error` instance -- so `error.message` was always `undefined` and the fallback string was returned instead. This made the panel's `Reason` row read "Unknown error" and its `Parser` row incorrectly read "v2 object / v1 array" instead of "not reached", directly failing the plan's own acceptance criterion for this scenario.
- **Fix:** Added `extractErrorMessage(error)` (checks `typeof error === 'string'` first, falls back to `error?.message`) and routed all 12 pre-existing `error.message || 'Unknown error'` call sites through it, including `loadCatalogFlat` and `browseCatalogs` which this plan's diagnostic panel depends on directly.
- **Files modified:** `frontend/src/services/wailsAPI.ts`
- **Verification:** Re-ran the deleted-file scenario via dev-browser after the fix: `Reason` now reads the real wrapped Go error naming the missing file, `Parser` correctly reads "not reached", and the raw-error box shows the full string verbatim.
- **Committed in:** `7952c2de` (part of the Task 2 commit)

---

**Total deviations:** 1 auto-fixed (1 blocking issue)
**Impact on plan:** Necessary for correctness -- without this fix, STATE-02's "missing file" and "permission denied" scenarios (both explicitly named in the plan's must-haves) would always show a generic, un-diagnostic "Unknown error" instead of the real cause. 5 of `wailsAPI.ts`'s ~17 catch blocks (theme/window-size/openExternal setters) were left untouched -- they don't feed this plan's error surfaces and are out of scope.

## Issues Encountered

None beyond the deviation above. `wails dev` was already running; dev-browser's session persisted across both tasks' verification. Fixture catalogs (two healthy -- one with an HTML companion, one without -- and one deliberately truncated at byte 95) were created under `/tmp/storcat-05-fixtures*` for live verification and deleted afterward, not committed to the repo.

## Known Stubs

None. Both `TreeHeader` and `UnreadableCatalogPanel` are fully wired end-to-end and verified live against real fixture catalogs, including the two backstop scenarios (deleted-between-listing-and-load, and re-selecting a broken catalog repeatedly).

## User Setup Required

None -- no external service configuration required.

## Next Phase Readiness

- `TreeHeader.tsx` and `UnreadableCatalogPanel.tsx` are complete surfaces; 23-06 (reveal-in-file-manager and any remaining TREE-07/08 work) can build alongside them without touching this plan's files.
- `wailsAPI.ts`'s `extractErrorMessage()` fix benefits every other binding call in the app, not just this plan's two components -- any future plan reading `result.error` from a wailsAPI call now gets the real Go error text.
- No blockers.

## Self-Check: PASSED

All 4 files claimed as created/modified confirmed present on disk; both task commit hashes (`e885f32d`, `7952c2de`) confirmed in `git log`.

---
*Phase: 23-rail-virtualized-tree*
*Completed: 2026-08-13*
