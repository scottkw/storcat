---
phase: 25-create-slide-over-progress-cancellation-partial-catalog
plan: 03
subsystem: catalog
tags: [go, wails, context-cancellation, partial-catalog, beforeClose, wailsjs-bridge]

# Dependency graph
requires:
  - phase: 25-01
    provides: "App.StartScan's mutex-guarded activeScanCancel/scanDone fields, CreateCatalogWithContext's ctx-threaded walk, the cancellation-writes-nothing guarantee (tree built in memory before any write)"
  - phase: 25-02
    provides: "catalog.SourceUnavailableError/PartialScan/ReadErrorEntry, Options.HaltOnSourceLoss, WriteCatalogFrom shared write path, the Unreadable/ReadError on-disk marker shape"
provides:
  - "App.CancelScan() Wails binding -- cancels the in-flight scan via the existing mutex-guarded handle, no-op with no scan running"
  - "App.WritePartialCatalog() Wails binding -- idempotent write of a retained source-loss tree through the shared WriteCatalogFrom path"
  - "App.lastPartial/lastPartialResult/lastScanReq -- App-held retention of a failed scan's walked tree + the parameters needed to write it, cleared atomically at every fresh StartScan"
  - "App.cancelActiveScan()/waitForScanStop() -- shared by CancelScan and beforeClose's new scan-cancellation branch"
  - "beforeClose's CRT-13 scan branch: cancel-then-bounded-wait-then-requery-quit, ahead of the existing window-persistence save"
  - "frontend/src/services/wailsAPI.ts cancelScan/writePartialCatalog bridge entries"
  - "frontend/src/types/scan.ts ScanFailure discriminated union + classifyScanFailure(message), keyed on the 'source unavailable' Go-error-text substring"
affects: [25-04, 25-05, 25-06, 25-07]

# Actuals (#2632)
actuals:
  tokens: 6902
  tasks: 3
  commits: 3

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Thin Wails-bound wrapper + testable parameterized core (App.StartScan / App.startScan) -- mirrors internal/catalog's CreateCatalog/CreateCatalogWithContext split, applied here so a headless test can inject a progress-hook without needing a live Wails runtime context (EventsEmit log.Fatals on an invalid one)"
    - "cancelActiveScan()/waitForScanStop() extracted once, shared by both the renderer-facing CancelScan binding and beforeClose's scan branch -- avoids two independent implementations of the same cancel-and-report-whether-something-was-running logic"
    - "Idempotent write via a cached-result short-circuit under the same mutex that guards the retained tree -- WritePartialCatalog's second call never reaches the filesystem"

key-files:
  created: []
  modified:
    - app.go
    - app_test.go
    - frontend/src/services/wailsAPI.ts
    - frontend/src/types/scan.ts
    - frontend/wailsjs/go/main/App.d.ts
    - frontend/wailsjs/go/main/App.js

key-decisions:
  - "startScan (unexported, parameterized on an optional progress test-hook) split out from the Wails-bound StartScan specifically so app_test.go could deterministically reproduce a mid-walk source loss headlessly -- a.throttledProgress's real path requires a live Wails runtime context, and constructing a fake non-nil context.Context to satisfy it would log.Fatal on the very first progress emission (wails' getEvents has no recoverable failure mode for an invalid context; internal/frontend, which owns the type it looks up, is a Go internal package this module cannot import). This is the same test-injection technique internal/catalog/service_test.go already uses at the service layer, applied one level up."
  - "CancelScan delegates to the new cancelActiveScan() helper (added in Task 2) rather than keeping its own independent mutex-guarded body from Task 1 -- a small in-plan refactor to avoid two copies of the same cancel-and-report logic; both task commits were re-verified green after the consolidation."
  - "The force-quit-mid-scan live verification (CRT-13's cancel-then-wait-then-requery sequence, 25-RESEARCH.md Assumption A2) was NOT performed this session -- wails dev was not running at task start, and per this plan's explicit operating instructions the executor does not start a long-lived dev server itself. This is recorded as an open manual verification, not asserted as proven; see Known Gaps below."

patterns-established:
  - "A bound method that needs headless-test injection without a live Wails runtime splits into an exported thin wrapper (fixed signature, matches the generated binding) and an unexported parameterized core the test calls directly -- reuse for any future StartScan-shaped binding that needs a progress/lifecycle hook."

requirements-completed: [CRT-09, CRT-11, CRT-13, COMPAT-04]

coverage:
  - id: D1
    description: "Cancelling a scan actually stops the underlying walk: the cancel handle lives on the App, is reachable from a second bound call, and the walk sees the cancellation at the next directory boundary"
    requirement: CRT-09
    verification:
      - kind: unit
        ref: "app_test.go#TestCancelScan_CancelsTheActiveContext"
        status: pass
      - kind: unit
        ref: "app_test.go#TestCancelScan_NoActiveScanIsNoOp"
        status: pass
    human_judgment: false
  - id: D2
    description: "A source-loss StartScan writes nothing into the output directory and retains a non-nil partial tree plus the originating request parameters on the App"
    requirement: CRT-09
    verification:
      - kind: unit
        ref: "app_test.go#TestStartScan_RetainsPartialOnSourceLoss"
        status: pass
      - kind: unit
        ref: "app_test.go#TestStartScan_ClearsRetainedPartialOnNewScan"
        status: pass
    human_judgment: false
  - id: D3
    description: "WritePartialCatalog writes the retained tree exactly once -- a second call returns the cached result without touching the filesystem, and the on-disk marker survives on exactly the node the walk marked"
    requirement: CRT-11
    verification:
      - kind: unit
        ref: "app_test.go#TestWritePartialCatalog_WritesOnce"
        status: pass
      - kind: unit
        ref: "app_test.go#TestWritePartialCatalog_MarkerSurvivesToDisk"
        status: pass
      - kind: unit
        ref: "app_test.go#TestWritePartialCatalog_WithoutRetainedScanErrors"
        status: pass
    human_judgment: false
  - id: D4
    description: "cancelActiveScan/waitForScanStop helpers correctly report whether a scan was running and bound-wait for the scan-done channel; beforeClose allows a normal close when idle"
    verification:
      - kind: unit
        ref: "app_test.go#TestCancelActiveScan_ReportsWhetherAScanWasRunning"
        status: pass
      - kind: unit
        ref: "app_test.go#TestWaitForScanStop_ReturnsWhenChannelCloses"
        status: pass
      - kind: unit
        ref: "app_test.go#TestWaitForScanStop_ReturnsOnDeadline"
        status: pass
      - kind: unit
        ref: "app_test.go#TestBeforeCloseDecision_AllowsCloseWhenIdle"
        status: pass
    human_judgment: false
  - id: D5
    description: "Closing the window mid-scan cancels the walk before the process exits, writing nothing and leaving no temp residue (CRT-13's live cancel-then-wait-then-requery sequence)"
    requirement: CRT-13
    verification: []
    human_judgment: true
    rationale: "This sequence (beforeClose returning prevent:true, cancelling, bounded wait, runtime.Quit re-entering the hook) is synthesised from individually verified Wails primitives, not found documented end-to-end anywhere (25-RESEARCH.md Assumption A2, MEDIUM confidence). It requires a live wails dev instance, which was not running at task start; per this plan's explicit operating instructions the executor does not start a long-lived dev server itself. The unit tests above prove the helpers' logic in isolation (cancelActiveScan reports correctly, waitForScanStop bounds correctly, the idle path is unaffected) but NOT the live OS-close-event -> OnBeforeClose -> runtime.Quit -> re-entry chain. A human (or a future session with wails dev already running) must perform the force-quit-mid-scan manual check this plan's own acceptance criteria names."
  - id: D6
    description: "The renderer can cancel a scan and request the partial write through the standard bridge, and can tell a cancellation apart from a source loss without inspecting raw error prose at each call site"
    requirement: COMPAT-04
    verification:
      - kind: other
        ref: "cd frontend && npx tsc --noEmit -- exit 0"
        status: pass
      - kind: other
        ref: "cd frontend && npm run build -- exit 0"
        status: pass
      - kind: other
        ref: "grep -c 'CancelScan' frontend/wailsjs/go/main/App.d.ts -- 1; grep -c 'WritePartialCatalog' frontend/wailsjs/go/main/App.d.ts -- 1; grep -c 'classifyScanFailure' frontend/src/types/scan.ts -- 2"
        status: pass
    human_judgment: false

duration: 3min
completed: 2026-08-14
status: complete
---

# Phase 25 Plan 3: Cancel, Retained Partial, Window-Close Guard Summary

**App-held cancel handle and retained partial-scan tree so a second bound call (CancelScan) stops an in-flight walk and a source-loss error's tree can be written exactly once (WritePartialCatalog); beforeClose gains a CRT-13 branch that cancels before quitting, and the frontend bridge (cancelScan/writePartialCatalog/classifyScanFailure) is wired for later plans to consume.**

## Performance

- **Duration:** 3 min (active task execution; excludes the file-reading/context-gathering phase before the first commit)
- **Started:** 2026-08-14T19:28:58-05:00 (Task 1 commit)
- **Completed:** 2026-08-14T19:31:28-05:00 (Task 3 commit)
- **Tasks:** 3
- **Files modified:** 6

## Accomplishments
- `App.CancelScan()` cancels the in-flight scan via the existing mutex-guarded handle; a cancel with no scan running is a documented no-op, not an error
- `StartScan` now sets `HaltOnSourceLoss:true` (the GUI wants the CRT-10 distinction; the CLI wrapper still never sets it), clears any previously retained partial before every fresh scan (T-25-12), and retains the walked tree + originating request parameters when the walk returns a `*catalog.SourceUnavailableError`
- `App.WritePartialCatalog()` writes the retained tree through the shared `WriteCatalogFrom` path exactly once -- idempotent by design (T-25-13): a second call returns the cached result without touching the filesystem
- `beforeClose` gains a scan-cancellation branch ahead of the existing window-persistence save: cancels an active scan, spawns a goroutine bounded at 3s to wait for it to stop, then re-requests quit -- the hook itself never blocks
- Frontend bridge (`wailsAPI.ts` `cancelScan`/`writePartialCatalog`, `scan.ts`'s `ScanFailure`/`classifyScanFailure`) regenerated and wired, ready for plans 25-06/25-07 to consume without building their own error-classification logic

## Task Commits

Each task was committed atomically:

1. **Task 1: Cancel handle and retained partial scan on the App** - `dfb9fdc7` (feat)
2. **Task 2: Cancel an in-flight scan before the window closes** - `d3cd1bb3` (feat)
3. **Task 3: Frontend bridge for cancel, partial write, and the two failure kinds** - `61794d74` (feat)

## Files Created/Modified
- `app.go` - `App.lastPartial`/`lastPartialResult`/`lastScanReq`, `lastScanRequest` type, `App.CancelScan()`, `App.WritePartialCatalog()`, `App.cancelActiveScan()`/`waitForScanStop()`, `StartScan`/`startScan` split, extended `beforeClose`
- `app_test.go` - 11 new tests across cancel/retention/write-once/beforeClose-helpers
- `frontend/src/services/wailsAPI.ts` - `cancelScan`, `writePartialCatalog`
- `frontend/src/types/scan.ts` - `ScanFailure` union, `classifyScanFailure()`
- `frontend/wailsjs/go/main/App.d.ts` / `App.js` - regenerated (`wails generate module`)

## Decisions Made
- `startScan` (unexported core, optional progress test-hook parameter) split from the Wails-bound `StartScan` specifically to make `TestStartScan_RetainsPartialOnSourceLoss` deterministic headlessly -- see key-decisions in frontmatter for the full reasoning on why a live-context workaround wasn't viable
- `CancelScan` was written standalone in Task 1's commit, then refactored in Task 2's commit to delegate to the new `cancelActiveScan()` helper (avoiding two copies of the same logic); both commits were independently rebuilt and retested green before being made
- The manual force-quit-mid-scan verification (this plan's own acceptance criterion for Task 2, and CRT-13's `25-RESEARCH.md` Assumption A2) was not performed -- `wails dev` was not running at task start, and this plan's operating instructions explicitly direct the executor not to start a long-lived dev server itself. Recorded honestly as an open item (coverage D5, `human_judgment: true`) rather than asserted as proven.

## Deviations from Plan

None beyond the in-plan `CancelScan`/`cancelActiveScan` consolidation noted above (not a deviation from written behavior -- both tasks' acceptance criteria were satisfied at their respective commits; it is a dedup refactor performed within Task 2's own commit, not a correction of a bug).

## Issues Encountered

**Headless-test injection required a small, justified core-extraction.** `throttledProgress`'s closure short-circuits when `a.ctx == nil` (a deliberate Plan 25-01 safety guard so tests never call into a real Wails runtime), which also meant the standard `internal/catalog/service_test.go` technique for simulating a mid-walk source loss (removing the scan root when a specific node's progress callback fires) had no observable effect when driven through `App.StartScan`'s real progress wiring. Setting `a.ctx` to any non-nil placeholder context is not viable either -- Wails' `getEvents(ctx)` calls `log.Fatalf` (not a recoverable panic) on any context lacking its internal `"events"` value, and the type it needs (`internal/frontend.Events`) lives in a Go `internal/` package this module cannot import across the module boundary. Resolved by extracting `startScan` (all of `StartScan`'s real logic, parameterized on an optional `testHook catalog.ProgressCallback`) with `StartScan` itself reduced to a one-line wrapper calling `a.startScan(..., nil)` -- the Wails-bound signature and behavior are unchanged; only `app_test.go` calls the parameterized core directly.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- `App.CancelScan`/`App.WritePartialCatalog` and their frontend bridge entries are ready for plan 25-06 (wiring cancel into the scanning body) and plan 25-07 (wiring the partial-write action into the error body) to consume directly -- neither later plan needs to build its own error-classification or write-path logic
- `ScanFailure`/`classifyScanFailure` give the frontend a single place (the `'source unavailable'` substring) to distinguish a cancellation from a source loss; a future change to `catalog.SourceUnavailableError`'s Go error text has exactly one TypeScript constant to update
- **Open item carried forward, not silently dropped:** the CRT-13 force-quit-mid-scan live verification (beforeClose's cancel-then-wait-then-requery sequence actually working end-to-end against a real OS close event) has not been performed. The next session with `wails dev` running should: start a large-directory scan, close the window mid-walk, and confirm (a) the process exits, (b) no `.json`/`.html` landed at the destination, and (c) no `.tmp` residue remains in the output directory -- exactly this plan's own Task 2 acceptance criterion.
- Pitfall 3's accepted residual limitation (a syscall already blocked inside the OS cannot be interrupted by `ctx.Err()` checks) is documented in `CancelScan`'s doc comment; no code changes were made to work around it, per `25-RESEARCH.md`'s explicit recommendation.

---
*Phase: 25-create-slide-over-progress-cancellation-partial-catalog*
*Completed: 2026-08-14*

## Self-Check: PASSED

All 6 created/modified files verified present on disk; all three task commit hashes (`dfb9fdc7`, `d3cd1bb3`, `61794d74`) verified present in git history.
