---
phase: 28-re-scan-diff
plan: 01
subsystem: catalog-scan
tags: [go, wails, react, typescript, diff-algorithm, re-scan]

requires:
  - phase: 25-create-catalog
    provides: traverseDirectory/walkState, ProgressCallback, scan:progress event pipeline, ScanningBody/ErrorBody, state.scan slice, useModalBehavior
  - phase: 27-catalog-actions-watch
    provides: DetailsPanel Footer stub, catalog-actions containment-check pattern (DeleteCatalog), --danger token
provides:
  - "CatalogItem.ModTime captured at scan time, omitempty, byte-parity preserved for existing catalogs"
  - "Service.Walk -- the promoted, primary tree-building operation; CreateCatalogWithContext reduced to Walk + WriteCatalogFrom"
  - "catalog.ComputeDiff -- pure size+mtime diff over two trees, four of five DiffState categories shipped"
  - "App.RescanCatalog Wails binding -- walk + diff, zero writes"
  - "state.rescan AppContext slice, RescanDialog shell (steps 1-3, tracer scope), five-tile diff render"
affects: [28-02, 28-03, 28-04]

actuals:
  tokens: 14898
  tasks: 2
  commits: 2

tech-stack:
  added: []
  patterns:
    - "Walk/write split: Walk is the primary exported tree-building operation, CreateCatalogWithContext is a two-line composition over it -- a future phase must never reintroduce a second traversal implementation"
    - "state.rescan is a NEW, separate AppContext slice beside state.scan -- live progress stays on the shared scan slice (same scan:progress subscription, no second listener), only the terminal diff outcome is rescan-owned"
    - "Additive optional props on an already-tested component (ScanningBody's onRunInBackground) rather than a fork, for a second consumer with slightly different needs"

key-files:
  created:
    - internal/catalog/walk.go
    - internal/catalog/diff.go
    - internal/catalog/diff_test.go
    - frontend/src/types/rescan.ts
    - frontend/src/components/workspace/rescan/RescanDialog.tsx
  modified:
    - pkg/models/catalog.go
    - internal/catalog/service.go
    - app.go
    - app_test.go
    - frontend/src/contexts/AppContext.tsx
    - frontend/src/components/workspace/create/ScanningBody.tsx
    - frontend/src/components/workspace/DetailsPanel.tsx
    - frontend/src/workspace.css
    - frontend/src/services/wailsAPI.ts

key-decisions:
  - "Walk extraction is behavior-preserving by construction (verbatim move, not re-derivation) -- TestCreateCatalog_JSONShapeUnchanged and both source-loss tests pass unmodified, confirmed by an empty git diff on service_test.go"
  - "RescanCatalog never touches a.lastPartial/a.lastPartialResult/a.lastScanReq on any path -- given its own reviewable commit (Task 2) per 28-RESEARCH.md's flagged highest-risk regression"
  - "Directories are diffed for existence only (added/removed/unchanged), never 'changed' -- driving it from a directory's own mtime would double-count every file add/remove alongside that file's own row"
  - "RescanDialog's frontend-only error handling (this tracer) dispatches back to step 1 with the failure surfaced through DetailsPanel Footer's existing error slot, rather than building the full Retry/Close error step -- that is explicitly plan 28-02+ scope per the plan's own action text"

requirements-completed: [ACT-06, ACT-08, STATE-03]

coverage:
  - id: D1
    description: "A user clicks 'Re-scan volume...' on a real catalog from the details-panel footer, picks a real mounted volume, and sees five diff stat tiles render with counts matching a staged fixture"
    requirement: "ACT-06"
    verification:
      - kind: e2e
        ref: "dev-browser live session against wails dev :34115 -- staged fixture (added=1, removed=1, changed=1, unchanged=1, unreadable=0) round-tripped through window.go.main.App.RescanCatalog with exact matching counts, then the same flow driven through the real UI against a mounted volume rendered 'Re-scan changed 3 entries' with correct tiles"
        status: pass
    human_judgment: false
  - id: D2
    description: "The volume picker is presented on every re-scan with nothing pre-selected -- no persistence of the previously chosen volume"
    requirement: "ACT-08"
    verification:
      - kind: unit
        ref: "code review: RescanDialog's selectedSource is component-local useState, reset on every fresh mount (dialog is conditionally mounted, never always-on); no localStorage/config write anywhere in the re-scan path"
        status: pass
    human_judgment: false
  - id: D3
    description: "Walk extraction preserves Create's exact behavior -- byte-identical JSON output, unmodified source-loss classification"
    requirement: null
    verification:
      - kind: unit
        ref: "internal/catalog/service_test.go#TestCreateCatalog_JSONShapeUnchanged, #TestCreateCatalogWithContext_SourceLossWritesNothing, #TestCreateCatalogWithContext_RootVanishesBeforeAnyProgress -- all pass UNMODIFIED (git diff on the test file is empty)"
        status: pass
    human_judgment: false
  - id: D4
    description: "A failed re-scan (SourceUnavailableError) never leaves a tree reachable through WritePartialCatalog"
    requirement: "STATE-03"
    verification:
      - kind: unit
        ref: "app_test.go#TestRescan_DoesNotRetainPartialForWritePartialCatalog (both subtests)"
        status: pass
    human_judgment: false
  - id: D5
    description: "ComputeDiff correctly categorizes added/removed/changed/unchanged, with the ModTime==0 size-only fallback for pre-Phase-28 catalogs, and never reports a directory as changed"
    requirement: null
    verification:
      - kind: unit
        ref: "internal/catalog/diff_test.go#TestDiff_AddedRemovedChangedUnchanged, #TestDiff_SameSizeMtimeChange, #TestDiff_MissingOldModTimeFallsBackToSizeOnly, #TestDiff_DirectoryNeverReportsChanged, #TestDiff_NilOldTreeReportsAllAdded"
        status: pass
    human_judgment: false

duration: 55min
completed: 2026-08-16
status: complete
---

# Phase 28 Plan 01: Re-scan Tracer Summary

**End-to-end re-scan-and-diff tracer: details-panel footer button → volume picker → live scan → five diff stat tiles, proven live against a mounted volume with Create's output byte-identical throughout**

## Performance

- **Duration:** 55 min
- **Started:** 2026-08-16T18:26:00Z (approx, from prior plan commit)
- **Completed:** 2026-08-16T19:21:22Z
- **Tasks:** 2
- **Files modified:** 17 (excluding `.planning/`)

## Accomplishments

- `Service.Walk` extracted verbatim from `CreateCatalogWithContext`, promoted to the primary tree-building operation; `CreateCatalogWithContext` reduced to a two-line `Walk` + `WriteCatalogFrom` composition with zero behavior change (proven by an unmodified `service_test.go` and all its byte-parity/source-loss tests passing unedited)
- `CatalogItem.ModTime` (Unix seconds, `omitempty`) captured at scan time with zero extra syscalls, in both the file and directory node-construction sites
- `catalog.ComputeDiff` — a pure, size+mtime diff over two trees, with the `old.ModTime == 0` size-only fallback for pre-Phase-28 catalogs, and directories deliberately never diffed as "changed" (existence only)
- `App.RescanCatalog` Wails binding: validates `jsonPath` against the configured catalog directory (same containment sequence `DeleteCatalog` uses), reuses the `scanMu` one-scan-at-a-time guard, calls `Walk` directly (never `startScan`/`CreateCatalogWithContext`), and never touches the Create-only retained-partial fields on any path — verified by its own dedicated test and its own reviewable commit
- `state.rescan`, a new AppContext slice separate from `state.scan`; `RescanDialog` (its own `.ws-rescan-*` class family, 620px) drives steps 1–3 of the tracer, reusing `VolumePicker` and `ScanningBody` (the latter's `onRunInBackground` made optional) and the existing `scan:progress` subscription verbatim — confirmed exactly one such subscription exists in the whole frontend
- Live-verified end-to-end against a running `wails dev` instance: `RescanCatalog` is present in `window.go.main.App`; a staged fixture round-tripped through the binding with exact expected counts; the full UI flow against a real mounted volume rendered "Re-scan changed 3 entries" with correct tiles; "Discard scan and close" left the catalog file byte-identical (MD5-verified before/after)

## Task Commits

1. **Task 1: End-to-end "re-scan this catalog and show me what changed"** - `9d3589ae` (feat)
2. **Task 2: Re-scan failure must never populate Create's retained-partial state** - `36ede3c2` (fix)

## Files Created/Modified

- `pkg/models/catalog.go` - `CatalogItem.ModTime`; `DiffState`/`DiffEntry`/`DiffResult` wire types
- `internal/catalog/walk.go` (new) - `Service.Walk`, extracted verbatim
- `internal/catalog/service.go` - `CreateCatalogWithContext` reduced to `Walk` + `WriteCatalogFrom`; mtime capture in `traverseDirectory`
- `internal/catalog/diff.go` (new) - `ComputeDiff`, `flatten`, `fileChanged`
- `internal/catalog/diff_test.go` (new) - five diff tests
- `app.go` - `RescanCatalog` binding; the retained-partial omission comment (Task 2)
- `app_test.go` - `TestRescan_DoesNotRetainPartialForWritePartialCatalog`
- `frontend/src/types/rescan.ts` (new) - `DiffState`/`DiffEntry`/`DiffResult`/`RescanState`/`RescanStep`
- `frontend/src/contexts/AppContext.tsx` - `state.rescan` slice; `RESCAN_OPENED`/`RESCAN_STARTED`/`RESCAN_DIFFED`/`RESCAN_CLOSED`
- `frontend/src/components/workspace/create/ScanningBody.tsx` - `onRunInBackground` made optional
- `frontend/src/components/workspace/rescan/RescanDialog.tsx` (new) - the 620px dialog, steps 1-3
- `frontend/src/components/workspace/DetailsPanel.tsx` - Footer's third button, "Re-scan volume…"
- `frontend/src/workspace.css` - `.ws-rescan-*` class family
- `frontend/src/services/wailsAPI.ts` - `rescanCatalog` wrapper (deviation, see below)
- `frontend/wailsjs/go/main/App.{d.ts,js}`, `frontend/wailsjs/go/models.ts` - regenerated via `wails generate module`

## Decisions Made

- Walk extraction kept byte-for-byte behavior-preserving by moving the existing block verbatim rather than re-deriving it — the empty `service_test.go` diff is the load-bearing proof, not a claim
- `RescanCatalog`'s never-touch-lastPartial guarantee was implemented correctly from the start (Task 1), then given its own dedicated Task 2 commit (test + a targeted comment at the omission site) per the plan's explicit "own reviewable diff hunk" instruction — the code didn't need fixing, but the review-visibility and regression-test obligation still needed its own commit
- `state.rescan` uses `RESCAN_OPENED` (re-dispatched on failure) to reset the dialog back to step 1 rather than inventing a fourth action type for "return to pick-volume with an error" — a minimal reuse of an already-planned action
- The tracer's frontend error handling (a failed `RescanCatalog` call) surfaces the message through `DetailsPanel`'s existing shared error slot and returns to step 1, rather than building 28-UI-SPEC's full Retry/Close error step — the plan's own action text explicitly scopes the error step, diff row list, similarity warning, and write resolutions to plans 28-02/03/04

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Added `wailsAPI.rescanCatalog` wrapper**
- **Found during:** Task 1 (frontend wiring)
- **Issue:** The plan's file list didn't name `frontend/src/services/wailsAPI.ts`, but every other Wails binding in this codebase is reached through this wrapper layer (never called raw from a component) — `RescanDialog` had no way to invoke `RescanCatalog` without it.
- **Fix:** Added a `rescanCatalog` wrapper following the existing `{success, ...} `/`wailsError` convention every other wrapper in the file uses.
- **Files modified:** `frontend/src/services/wailsAPI.ts`
- **Verification:** `tsc --noEmit` and `npm run build` both pass; live-verified via dev-browser (see Accomplishments).
- **Committed in:** `9d3589ae` (Task 1 commit)

---

**Total deviations:** 1 auto-fixed (1 blocking)
**Impact on plan:** Necessary to complete Task 1's stated goal (a working `RescanDialog` calling the real binding) — no scope creep beyond what the plan's own architecture required.

## Known Stubs

None that block this plan's own goal. Explicitly out of scope per the plan's own text (not stubs, deliberately deferred to 28-02/03/04):
- Diff row list, similarity warning banner, and the Overwrite/Keep-both write resolutions — Step 3's footer carries only "Discard scan and close" this plan.
- The full Retry/Close error/interrupted step — a failed re-scan currently returns to step 1 with the message in `DetailsPanel`'s error slot.
- Catalog-actions menu entry point and `UnreadableCatalogPanel`'s action trio — only the details-panel footer entry point is wired this plan.
- `Options.MarkUnreadableOnSkip` and the `unreadable` diff category are not implemented — `DiffResult.Unreadable` is always `0` this plan, exactly as the plan specifies.

## Issues Encountered

- Initial live-fixture verification used flat display-path names (`./keep.txt`) matching the hand-built trees in `service_test.go`'s unit tests, which don't match the real `Walk`'s actual convention (children are named `./{sourceDirBasename}/...`, only the root itself is bare `./`). Corrected the fixture and re-verified — the diff algorithm itself was correct throughout; the mismatch was in the fixture's construction, not the code under test.
- The first live click-through accidentally selected an already-broken test catalog (`after-toggle-off.json`, a pre-existing corrupt-JSON fixture from an earlier phase) as the re-scan target, which correctly surfaced a `LoadCatalog` failure through the new error-handling path — a useful accidental proof of that path, then re-verified against a genuinely valid catalog (`burst-1.json`) for the intended happy-path proof.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- `Walk`, `ComputeDiff`, `RescanCatalog`, `state.rescan`, and `RescanDialog`'s shell are all in place for plan 28-02 to build on: the fourth diff state (`unreadable`, gated by a new `Options.MarkUnreadableOnSkip`), the diff row list, the similarity warning, and the two write resolutions (Overwrite/Keep-both via `ResolveRescan`).
- No blockers. The `service_test.go` byte-parity guarantee and the `lastPartial` isolation guarantee are both machine-verified (tests), not just reasoned about, so later plans extending `RescanCatalog`'s error/write paths have a firm regression backstop to build against.

---
*Phase: 28-re-scan-diff*
*Completed: 2026-08-16*

## Self-Check: PASSED

All created files verified present on disk; both task commit hashes (`9d3589ae`, `36ede3c2`) verified present in git history.
