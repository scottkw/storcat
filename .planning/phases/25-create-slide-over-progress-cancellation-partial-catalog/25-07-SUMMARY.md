---
phase: 25-create-slide-over-progress-cancellation-partial-catalog
plan: 07
subsystem: ui
tags: [react, wails, go, error-state, done-state, entry-points, root-vanish-classification]

# Dependency graph
requires:
  - phase: 25-01
    provides: "CreateSlideOver.tsx shell, lifted AppContext.scan state machine, classifyScanFailure/ScanFailure, the minimal 25-01 done body this plan replaces"
  - phase: 25-02
    provides: "catalog.SourceUnavailableError/PartialScan/ReadErrorEntry, the on-disk Unreadable/ReadError marker shape, WriteFileAtomic"
  - phase: 25-03
    provides: "App.CancelScan/WritePartialCatalog bindings, the retained-partial-tree mechanism this plan's write-partial action calls"
  - phase: 25-06
    provides: "ScanningBody.tsx, handleCloseRequest(reason), the status-bar background-scan segment this plan's disabled entry points defer to"
provides:
  - "ErrorBody.tsx -- the full CRT-10/CRT-11 error state: stop-point headline, mount path, files-walked count, an honest aggregate read-error count line, the verbatim explanation, and three working recoveries"
  - "DoneBody.tsx -- the full CRT-12 done state in both flavours, with real per-output-file byte sizes and a real elapsed duration"
  - "CreateCatalogResult.JsonSize/HtmlSize/CopyJsonSize/CopyHtmlSize (pkg/models/catalog.go) -- additive real on-disk size fields, populated at write time from the bytes actually written/copied"
  - "CreateCatalogWithContext's root-vanish classification fix -- the scan root's own initial os.Stat failure now correctly produces *SourceUnavailableError under HaltOnSourceLoss, instead of silently falling through as a generic error"
  - "Four wired create entry points (rail pill, rail empty-state link, tree pane primary/secondary buttons), each opening at the form step and disabled while a scan runs"
  - "AppContext.createFolderPickerIntent -- additive flag landing the tree pane's secondary entry point directly on the folder dialog"
  - "The phase's completed verification matrix (25-VALIDATION.md), including two manual-only checks performed live this session"
affects: []

# Actuals (#2632)
actuals:
  tokens: 22818
  tasks: 3
  commits: 3

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Aggregate-count rendering instead of fabricated per-item detail when the wire payload only carries a count -- ErrorBody's read-error line renders '{N} read errors recorded', never invented per-path text, because ScanProgress never transports individual read-error paths/reasons across the Wails bridge"
    - "Atomic directory rename (not recursive rm -rf) as the live source-loss reproduction technique -- a rename makes the root path unreachable in one syscall, unlike rm -rf, which leaves the root directory entry reachable until its very last internal syscall and was found live to produce a false 'done' instead of a source-loss error"
    - "window.runtime.Quit() called directly in the live webview as the CRT-13 force-quit reproduction -- exercises the real OnBeforeClose -> beforeClose Go hook with zero host-OS automation, the same sanctioned in-webview-binding-call technique prior plans used for native dialogs"
    - "Shared filesFromResult() helper (CreateSlideOver.tsx) so the complete-scan and write-partial success paths build the done state's file-row array from one place, never two independently-drifting literals"

key-files:
  created:
    - frontend/src/components/workspace/create/ErrorBody.tsx
    - frontend/src/components/workspace/create/DoneBody.tsx
  modified:
    - frontend/src/components/workspace/CreateSlideOver.tsx
    - frontend/src/components/workspace/CatalogRail.tsx
    - frontend/src/components/workspace/TreePane.tsx
    - frontend/src/contexts/AppContext.tsx
    - frontend/src/types/scan.ts
    - frontend/src/workspace.css
    - internal/catalog/service.go
    - internal/catalog/service_test.go
    - pkg/models/catalog.go
    - frontend/wailsjs/go/models.ts

key-decisions:
  - "The error log renders one aggregate read-error count line, not one line per failure -- this plan is frontend-only by its own files_modified list, and the wire payload (ScanProgress, and the StartScan rejection's message) never carries individual read-error paths/reasons, only a count. Rendering per-path text matching the design mockup's illustrative 'read error: {path} -- input/output error' line would be fabrication; an honest aggregate is what the data actually supports. Documented as a resolution built on E6's own flagged-unresolved status, not a quiet mark-resolved."
  - "The error headline omits the percentage clause ('Stopped -- the volume went away') when the failure happens during the counting sub-state, before any total was ever known -- the locked template assumes a percentage is always available, which a genuinely fast total disconnect can violate. Same class of reasoned extension 25-UI-SPEC.md itself already applied to the counting sub-state's status-bar copy."
  - "Fixed a real classification gap in internal/catalog/service.go's CreateCatalogWithContext, found live-verifying this plan's own Task 1: the scan root's own top-of-function os.Stat failure had no parent loop to run it through recordReadError+classify() the way a child's failure already does, so an instant/total disconnect (root gone before a single child was ever read -- the common case for an actually-ejected volume) fell through as a generic 'failed to traverse directory' error instead of *SourceUnavailableError, and the frontend's classifyScanFailure misclassified it as a cancellation, silently resetting to idle instead of surfacing the error state at all. Scoped to HaltOnSourceLoss=true (the GUI path only) via st.classify()'s own existing guard, so the CLI wrapper's behavior on a nonexistent source is provably unchanged (new test: TestCreateCatalog_WrapperRootVanishReturnsPlainError)."
  - "Added CreateCatalogResult.JsonSize/HtmlSize/CopyJsonSize/CopyHtmlSize (pkg/models/catalog.go), populated from len() of the bytes actually written/copied at zero extra I/O cost (writeJSONFile/writeHTMLFile/copyFile already had the byte count in hand) -- not in this plan's original files_modified list, but required for the plan's own must_have truth 'the done state lists every file actually written, with its size', which was otherwise unfulfillable without a Go change. Additive fields only; cli/create.go and go.mod/go.sum untouched."
  - "Every create entry point resets a stale terminal scan state (done/error) to idle before opening, rather than reopening into whatever state.scan happened to hold -- found live-testing Task 1's 'Close without writing' -> '+New' sequence, which reopened directly into the same persisted error state instead of the form step 25-UI-SPEC.md's Entry Points section promises. The always-live status-bar segment (a different, unaffected path) still reopens into the true live scanning state."
  - "Retry scan reuses handleCreate directly rather than a bespoke restart path -- StartScan runs synchronously on the Go side and its deferred cleanup clears the active-scan handle before the rejected promise ever reaches the frontend, so the existing submittingRef guard is sufficient defense against a double-retry race; no additional wait-for-release state was built."

patterns-established:
  - "Aggregate-count rendering as the honest fallback when a wire payload carries a count but not the underlying per-item detail -- reusable anywhere else in this app that later needs to summarize N events without inventing N descriptions."

requirements-completed: [CRT-01, CRT-10, CRT-11, CRT-12]

coverage:
  - id: D1
    description: "A vanished source produces a distinct error state naming the stop percentage (or an honest fallback when no percentage was ever known), the mount path, and the files-walked count -- never a silently truncated catalog reported as success (CRT-10)"
    requirement: CRT-10
    verification:
      - kind: unit
        ref: "internal/catalog/service_test.go#TestCreateCatalogWithContext_RootVanishesBeforeAnyProgress, #TestTraverseDirectory_TerminalSourceLossStopsWalk"
        status: pass
      - kind: automated_ui
        ref: "dev-browser :34115 -- an atomic-rename source-loss simulation against a real 60k-150k-file scratch fixture produced 'Stopped — the volume went away' (percent honestly omitted, failure landed mid-count) with the real mount path and a real files-walked count (66401 in one run); a second run with two broken symlinks showed the correct mount path/count plus '2 read errors recorded before the stop.'"
        status: pass
    human_judgment: false
  - id: D2
    description: "From the error state, write-partial-catalog, retry-scan, and close-without-writing all behave as specified; the error state itself never writes anything unless the write action was clicked, and a double-click on write-partial produces exactly one catalog (CRT-11, idempotency)"
    requirement: CRT-11
    verification:
      - kind: automated_ui
        ref: "dev-browser :34115 -- two same-tick clicks on 'Write partial catalog' produced exactly one JSON+HTML pair on disk (ls confirmed no duplicate/renamed files); the JSON carried the unreadable/readError marker on the root node; 'Retry scan' restarted on the same (still-gone) source and correctly stayed in the error state on a second failure rather than resetting to idle; 'Close without writing' ran the standard exit animation and left the output directory empty"
        status: pass
    human_judgment: false
  - id: D3
    description: "The done state lists every file actually written, with its real on-disk size, in the order the result provides, keyed on path -- and a source with zero files still reaches done with a JSON row present (CRT-12)"
    requirement: CRT-12
    verification:
      - kind: automated_ui
        ref: "dev-browser :34115 -- an empty source directory produced a done state with a real JSON row (55B) and HTML row (2.2K), no partial tag; a run with write-HTML and copy-to-secondary both enabled produced exactly four rows (primary json/html + secondary json/html), each with a real measured size matching the copied file's actual bytes"
        status: pass
    human_judgment: false
  - id: D4
    description: "A partial write is visually and textually distinguished from a complete one -- the partial tag, the stop-percentage clause replacing duration, and the on-disk marker together (CRT-11/CRT-12 partial)"
    requirement: CRT-12
    verification:
      - kind: automated_ui
        ref: "dev-browser :34115 -- the write-partial-catalog flow (D2 above) transitioned to the done state's partial flavour: the 'partial' tag rendered beside the title, and the doneLine read the stop-point clause rather than a duration"
        status: pass
    human_judgment: false
  - id: D5
    description: "Open in workspace re-fetches the catalog listing, selects the new catalog, and closes via the standard exit (the fifth CRT-01 close path); Catalog another volume resets the form and returns to the form step in the same still-open panel, with no re-entry animation"
    requirement: CRT-01
    verification:
      - kind: automated_ui
        ref: "dev-browser :34115 -- 'Open in workspace' closed the panel and selected the new catalog in the rail (confirmed via the rail row's data-selected attribute and title); 'Catalog another volume' left the panel open, showed the form step, and produced no ws-create-panel-exit class"
        status: pass
    human_judgment: false
  - id: D6
    description: "All four create entry points open the panel at the form step (resetting a stale terminal scan state rather than reopening into it), close the palette if open, and are visibly disabled -- and functionally inert, not just styled -- while a scan is running, with the status-bar segment as the one live path back in"
    requirement: CRT-01
    verification:
      - kind: automated_ui
        ref: "dev-browser :34115 -- all four entry points opened the panel at the form step from a variety of prior states; during a real ~150k-300k-file scan, all four showed aria-disabled=true, the locked tooltip copy, and not-allowed cursor; a click on the disabled +New pill after the exit animation fully completed produced no reopened panel (confirmed panel absent both before and after the click); the status-bar segment still reopened the panel into the live scanning state during the same disabled window"
        status: pass
    human_judgment: false
  - id: D7
    description: "The tree pane's secondary entry point lands directly on the folder dialog rather than the generic volume-card list"
    verification:
      - kind: automated_ui
        ref: "dev-browser :34115 -- clicking 'Choose catalog folder…' with SelectDirectory stubbed in-page invoked the binding exactly once and the returned folder became the selected source, with the form's volume-card list otherwise unused for that flow"
        status: pass
    human_judgment: false
  - id: D8
    description: "Force-quit mid-scan writes nothing (CRT-13) -- open since 25-03-SUMMARY.md's D5, closed this session"
    verification:
      - kind: automated_ui
        ref: "dev-browser :34115 -- window.runtime.Quit() called directly in the live webview mid-walk against a real 180k-file scan (the same Wails runtime binding a real window-close/Cmd+Q ultimately calls, routed through the real beforeClose Go hook, with no host-OS automation involved). The app process exited; the output directory was empty afterward with zero .tmp residue. Side effect: the whole wails dev supervisor session also exited and required a manual restart -- see Issues Encountered."
        status: pass
    human_judgment: false
  - id: D9
    description: "Atomic write survives a real process kill mid-write (CRT-11, and Phase 27's ACT-09 which reuses this primitive)"
    verification: []
    human_judgment: true
    rationale: "The D8 force-quit check exercises the cancel-before-any-write path (StartScan cancels and returns before WriteCatalogFrom is ever reached, since it fired mid-walk) -- it does not exercise a kill landing inside WriteFileAtomic's own few-millisecond temp-then-rename window, which this environment has no reliable way to schedule (no debugger/breakpoint access to pause execution mid-syscall). Recorded honestly as not run, same as this phase's validation strategy anticipated from the start; the guarantee rests on WriteFileAtomic's unit tests and os.Rename's OS-level atomicity within one filesystem, not an empirical kill test. Logged to .planning/WINDOWS.md (entry #6, kind unrun-verify)."
  - id: D10
    description: "Windows disk-space/drive-letter enumeration and the Linux mount-enumeration heuristic remain compile-verified only -- no runtime machine available"
    verification:
      - kind: other
        ref: "GOOS=windows GOARCH=amd64 go build ./internal/volumes/ and GOOS=linux GOARCH=amd64 go build ./internal/volumes/, both re-run clean this session"
        status: pass
    human_judgment: true
    rationale: "No Windows or Linux machine/VM available to this session, same limitation 25-04-SUMMARY.md documented (WINDOWS.md entries #4/#5, both still open). Must not be claimed as runtime-verified."
  - id: D11
    description: "Full-phase regression: go build/test -race across all packages, both platform cross-builds, tsc --noEmit, npm run build, cli/create.go untouched, internal/catalog imports no Wails package, go.mod/go.sum/frontend package files untouched"
    verification:
      - kind: other
        ref: "go build ./... && go test ./... -count=1 -race (all 8 packages pass) && GOOS=windows/linux GOARCH=amd64 go build ./internal/volumes/ (both exit 0) && cd frontend && npx tsc --noEmit && npm run build (both exit 0); git diff --stat -- cli/create.go empty; go list -deps ./internal/catalog/... | grep -i wailsapp empty; git diff --stat -- go.mod go.sum frontend/package.json frontend/package-lock.json empty -- all re-run clean by 25-07-T3"
        status: pass
    human_judgment: false

duration: 47min
completed: 2026-08-14
status: complete
---

# Phase 25 Plan 7: The Error State, Done State, and Entry Points Summary

**The interrupted-scan case this whole phase exists for is now honest end to end: a vanished source produces a real error state with three working recoveries, a finished scan lists exactly what landed on disk with real sizes and a real duration, and all four create entry points open at the form step and lock while a scan runs -- closing on a real classification bug (an instant source loss silently resetting to idle) found live-testing the very feature meant to catch it.**

## Performance

- **Duration:** ~47 min (approximate -- `PLAN_START_TIME` was not captured via `date -u` at the literal start of this session; derived from the earliest available evidence and the three task commit timestamps, 21:15:54 through 21:37:38)
- **Started:** ~2026-08-14T20:50:00-05:00 (approximate, immediately following plan 25-06's completion; includes reading all six prior plans' SUMMARYs, the UI-SPEC, PATTERNS, and current source before Task 1's first edit)
- **Completed:** 2026-08-14T21:37:38-05:00 (Task 3 commit)
- **Tasks:** 3
- **Files modified:** 13 (2 created, 11 modified)

## Accomplishments

- `ErrorBody.tsx` renders CRT-10/CRT-11's error state in full: round error badge, stop-point headline (with an honest fallback when no percentage was ever known), mount path + files-walked sub-line, an aggregate read-error count line (omitted entirely when zero, satisfying E6's zero-one-many case), the verbatim explanation paragraph, and three working recoveries -- verified live against real 60k-150k-file scratch fixtures using an atomic-rename source-loss simulation, which proved more deterministic than the `rm -rf` technique tried first (documented as a deviation, not silently discarded)
- Found and fixed a real classification bug in `internal/catalog/service.go`'s `CreateCatalogWithContext` while live-verifying Task 1: the scan root's own initial `os.Stat` failure had no parent loop to classify it as source-loss the way a child's failure already does, so an instant/total disconnect silently fell through as a generic error and the frontend reset straight to idle instead of showing the error state at all -- exactly the case CRT-10's own zero-one-many resolution names. Fixed scoped to the GUI path only (`HaltOnSourceLoss=true`); the CLI wrapper's behavior on a nonexistent source is provably unchanged (new test)
- `DoneBody.tsx` replaces the 25-01 stub done body: one shared layout for both flavours, a doneLine that swaps duration for a stop-percentage clause on a partial, the "partial" tag reusing the volume card's read-errors tag styling verbatim, and written-file rows keyed on path with real per-file sizes -- added `CreateCatalogResult.JsonSize`/`HtmlSize`/`CopyJsonSize`/`CopyHtmlSize` (additive Go fields, zero extra I/O) since the done state's own "with its size" truth was otherwise unfulfillable
- `CreateSlideOver.tsx`'s `handleCreate` now computes a real elapsed duration instead of the previous hardcoded `durationMs: 0`; both the complete and partial `SCAN_DONE` dispatches share one `filesFromResult` helper
- All four create entry points (rail pill, rail empty-state link, tree pane primary/secondary buttons) are wired, disabled (and functionally inert, verified by attempting a click mid-scan) while a scan runs with the locked tooltip copy, and now correctly reset a stale done/error scan state before opening -- a second real bug found live-testing this task ("Close without writing" then "+New" was reopening into the same persisted error state instead of the form step)
- Closed a phase-long open item: the CRT-13 force-quit-mid-scan manual check (open since 25-03-SUMMARY.md's D5) was performed live this session via `window.runtime.Quit()` called directly in the webview -- no host-OS automation, the real `beforeClose` Go hook -- confirming the process exits with nothing written and no `.tmp` residue
- `.planning/phases/.../25-VALIDATION.md`'s full per-task map and manual-only table are populated with observed results, not pending markers; `.planning/WINDOWS.md` gained one new honest not-run entry (atomic-write-survives-a-crash) and had its CRT-13 entry marked fixed

## Task Commits

Each task was committed atomically:

1. **Task 1: The error state -- stop point, read errors, and three recoveries** - `917deb3c` (feat)
2. **Task 2: The done state in both flavours** - `d8ce8f90` (feat)
3. **Task 3: The four entry points, their disabled state, and the phase verification matrix** - `3991a9ff` (feat)

## Files Created/Modified

- `frontend/src/components/workspace/create/ErrorBody.tsx` (new) -- round error badge, headline, sub-line, aggregate read-error log, explanation, three recovery actions
- `frontend/src/components/workspace/create/DoneBody.tsx` (new) -- round success badge, both flavours' shared layout, toggle-gated written-files list, two actions
- `frontend/src/components/workspace/CreateSlideOver.tsx` -- `ErrorBody`/`DoneBody` composed in; `handleWritePartial`/`handleCatalogAnother`; real duration; `filesFromResult`; the folder-picker-intent consumer effect; `SCAN_FAILED` now carries `sourcePath`
- `frontend/src/components/workspace/CatalogRail.tsx` -- the +New pill and the empty-state link both wired to `openCreatePanel` (form-step reset + disabled guard), the link converted from a static span to a real button
- `frontend/src/components/workspace/TreePane.tsx` -- both empty-state buttons wired; the secondary one also sets `createFolderPickerIntent`
- `frontend/src/contexts/AppContext.tsx` -- `createFolderPickerIntent` state + action; `SCAN_FAILED`/`SCAN_DONE` payload extensions; `readErrors` now tracked through the counting sub-state
- `frontend/src/types/scan.ts` -- `error` variant gains `sourcePath`/`filesSeen`/`stopPercent`/`readErrors`; `done` variant gains optional `stopPercent`; `counting` variant gains `readErrors`
- `frontend/src/workspace.css` -- `.ws-create-state-body`, `.ws-create-badge*`, `.ws-create-errlog*`, `.ws-create-explain`, `.ws-create-actions`, `.ws-create-tag-partial`, `.ws-create-file-size`, `.ws-entry-disabled`
- `internal/catalog/service.go` -- the root-vanish classification fix in `CreateCatalogWithContext`; `writeJSONFile`/`writeHTMLFile`/`copyFile` now return their real byte count
- `internal/catalog/service_test.go` -- `TestCreateCatalogWithContext_RootVanishesBeforeAnyProgress`, `TestCreateCatalog_WrapperRootVanishReturnsPlainError`, 3 call-site signature updates
- `pkg/models/catalog.go` -- `CreateCatalogResult.JsonSize`/`HtmlSize`/`CopyJsonSize`/`CopyHtmlSize`
- `frontend/wailsjs/go/models.ts` -- regenerated (`wails generate module`)
- `.planning/phases/25-.../25-VALIDATION.md` -- full per-task map and manual-only table populated

## Decisions Made

See `key-decisions` in the frontmatter for the full reasoning on: the aggregate read-error line (frontend-only plan, no per-path data crosses the bridge), the headline's percent-omitted fallback, the root-vanish classification fix and its CLI-safe scoping, the additive `CreateCatalogResult` size fields, the entry-points form-step reset, and reusing `handleCreate` for retry.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Scan root's initial `os.Stat` failure not classified as source-loss**
- **Found during:** Task 1, live-verifying the error state against a real fixture
- **Issue:** `traverseDirectory`'s top-of-function `os.Stat` check returns a bare error with no classification for the outermost (scan-root) call -- only a child's failure gets run through `recordReadError`+`classify()`, via the parent loop's own error handling. An instant/total disconnect (root gone before any child was ever read) therefore produced a generic `"failed to traverse directory: ..."` error instead of `*SourceUnavailableError`, and the frontend's `classifyScanFailure` (keyed on the `"source unavailable"` substring) silently classified it as a cancellation, resetting the panel straight to idle instead of showing the error state.
- **Fix:** `CreateCatalogWithContext` now checks `st.classify()` itself when the top-level `traverseDirectory` call returns a non-`SourceUnavailableError`, non-cancellation error, constructing a valid marked-unreadable root node (never a nil `Tree`, which would marshal to JSON `null`). `st.classify()`'s own `!HaltOnSourceLoss` short-circuit keeps the CLI wrapper's behavior on a nonexistent source unchanged.
- **Files modified:** `internal/catalog/service.go`, `internal/catalog/service_test.go` (2 new tests)
- **Verification:** `go test ./internal/catalog/... -race`; live-verified against a real scratch fixture with a rename-based root-loss simulation, both with and without prior read errors
- **Committed in:** `917deb3c` (Task 1 commit)

**2. [Rule 2 - Missing critical functionality] `CreateCatalogResult` had no real per-output-file byte size**
- **Found during:** Task 2, implementing the done state's "with its size" truth
- **Issue:** The only size data available was `TotalSize` (the scanned tree's total content size), which 25-01's own decision explicitly rejected using as a per-file size (would misrepresent an individual file's own size). Without a real per-file size, the done state's written-file rows could only ever render sizeless, leaving CRT-12's "lists every file actually written, with its size" truth unfulfillable.
- **Fix:** Added `JsonSize`/`HtmlSize`/`CopyJsonSize`/`CopyHtmlSize` to `CreateCatalogResult` (additive `omitempty` fields matching their path counterparts' convention), populated from `len()` of the bytes `writeJSONFile`/`writeHTMLFile` already had in hand and `io.Copy`'s own return value in `copyFile` -- zero extra I/O.
- **Files modified:** `pkg/models/catalog.go`, `internal/catalog/service.go`, `frontend/wailsjs/go/models.ts` (regenerated)
- **Verification:** `go test ./internal/catalog/... -race`; live-verified real sizes render correctly for 1-row, 2-row, and 4-row done states, matching the actual bytes on disk
- **Committed in:** `d8ce8f90` (Task 2 commit)

**3. [Rule 1 - Bug] Create entry points reopened into a stale terminal scan state**
- **Found during:** Task 3, live-testing the "Close without writing" -> "+New" sequence
- **Issue:** Dispatching `SET_CREATE_OPEN: true` alone (the pre-existing +New pill behavior) reopens the panel showing whatever `state.scan` currently holds. A `done`/`error` state dismissed via Escape/scrim/X (rather than its own action buttons) persists in `AppContext`, so reopening via any entry point landed back on the stale terminal state instead of the form step 25-UI-SPEC.md's Entry Points section promises ("defaulting to the form step").
- **Fix:** Each entry point's handler (`openCreatePanel`) now dispatches `SCAN_RESET` first when `state.scan.status` is `done` or `error`, before opening. The always-live status-bar segment (a structurally different path, only rendered while actually scanning) is unaffected and still reopens into the true live scanning state.
- **Files modified:** `frontend/src/components/workspace/CatalogRail.tsx`, `frontend/src/components/workspace/TreePane.tsx`
- **Verification:** Live-verified: reopening after a dismissed error/done state now lands on the form; reopening the status-bar segment during a real backgrounded scan still lands on the live scanning body
- **Committed in:** `3991a9ff` (Task 3 commit)

---

**Total deviations:** 3 auto-fixed (2 bugs, 1 missing critical functionality). All three were necessary for the plan's own must_have truths to actually hold; none changed architecture or added scope beyond what CRT-10/CRT-11/CRT-12 and the Entry Points contract already require.
**Impact on plan:** The Go-side changes (root-vanish classification, output-file sizes) touch `internal/catalog/service.go` and `pkg/models/catalog.go`, which were not in this plan's original `files_modified` list -- both are additive, `cli/create.go`-safe (verified: `git diff --stat -- cli/create.go` empty), and directly required by this plan's own locked must-have truths, per deviation Rules 1 and 2.

## Issues Encountered

**Calling `window.runtime.Quit()` to reproduce CRT-13 also exited the whole `wails dev` supervisor session, not just the app window.** This was not anticipated going in -- a real `beforeClose`-mediated quit was expected to behave like any other rebuild-triggered relaunch this session had already seen many times. Instead, `wails dev`'s own process logged `Development mode exited` and the entire dev server (Vite + the Go supervisor) terminated, requiring a manual restart (`wails dev` re-launched in the background, ~10s to come back up) before the session's remaining live checks (the `/Volumes` re-confirmation, the four-entry-point disabled-state check) could continue. No data was lost -- `AppContext`'s state lives in the browser tab, which survived and reconnected cleanly to the relaunched app on `:34115`. Recorded here rather than silently worked around, since a future session attempting the same CRT-13 reproduction technique should expect this and budget the restart time.

**The `rm -rf`-based source-loss simulation, tried first, does not faithfully reproduce a vanished volume.** A recursive `rm -rf` removes a directory tree bottom-up, meaning the ROOT directory entry itself remains stat-able until the very last internal syscall of the whole operation -- so a walk racing a concurrent `rm -rf` can (and did, once) complete successfully instead of hitting the terminal classification, because `classify()`'s root re-probe kept succeeding the whole time individual files were vanishing underneath it. Switched to an atomic directory rename (`mv src src.moved`), which makes the root path unreachable in exactly one syscall -- the same effect a real device eject has -- and reliably reproduced the error state on every subsequent attempt.

## User Setup Required

None -- no external service configuration required.

## Next Phase Readiness

- Phase 25 (create slide-over, progress, cancellation, partial-catalog) is functionally complete: all 25 CRT requirements across 7 plans are implemented and live-verified against real data on this machine, including the phase-long open CRT-13 item closed this session
- `.planning/WINDOWS.md` carries 5 open entries at phase close: #1 (Phase 23, Windows reveal argv), #2 (Phase 24, Windows/Linux Ctrl+K), #4/#5 (this phase, Windows/Linux volume enumeration runtime-unverified), #6 (this phase, atomic-write-survives-a-crash not empirically tested) -- all require real non-macOS hardware or a live kill-signal timing setup this environment cannot provide, and must be swept before v3.0.0 ships
- No new frontend test infrastructure was added (TEST-01 stays deferred, per the phase's own validation strategy) -- every check in this plan was either a Go unit test or a live `dev-browser` session
- `cli/create.go` is provably unedited across all 7 plans in this phase; `internal/catalog`/`internal/volumes` import no Wails package; `go.mod`/`go.sum`/`frontend/package.json`/`frontend/package-lock.json` are all untouched from this milestone's start

---
*Phase: 25-create-slide-over-progress-cancellation-partial-catalog*
*Completed: 2026-08-14*

## Self-Check: PASSED

All 14 created/modified files (2 new, 12 modified) verified present on disk; all three task commit hashes (`917deb3c`, `d8ce8f90`, `3991a9ff`) verified present in git history.
