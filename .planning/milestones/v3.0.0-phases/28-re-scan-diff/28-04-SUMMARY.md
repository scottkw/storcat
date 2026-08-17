---
phase: 28-re-scan-diff
plan: 04
subsystem: catalog-scan
tags: [go, react, typescript, rescan, write-path, resolution-footer]

requires:
  - phase: 28-re-scan-diff
    plan: 01
    provides: Walk, ComputeDiff (tracer scope), App.RescanCatalog, state.rescan slice, RescanDialog shell (steps 1-3)
  - phase: 28-re-scan-diff
    plan: 03
    provides: DiffList, RescanDialog Step 3 Variants A/B, resolution caption, error step, catalog-actions menu entry point
provides:
  - "WriteRescanResult (internal/catalog/resolve.go) -- writes an already-walked re-scan tree to disk via overwrite-in-place or keep-both, reusing WriteCatalogFrom (the hardened atomic write) and nextCopyRoot (the shared collision loop) unmodified"
  - "ResolveRescan Wails binding (app.go) -- re-derives and re-validates the write target from the catalog's own on-disk path (filepath.Abs -> filepath.EvalSymlinks -> osutil.ContainsPath), rejects discard as a write mode entirely"
  - "RescanDialog's three-action resolution footer (Overwrite catalog / Keep both / Discard scan and close) with an inline write-failure banner following the Rename/Delete pattern"
  - "Keep both's label makes no filename claim -- fixed post-verification so it can never diverge from the collision loop's actual write target"
affects: [28-05, 28-06]

actuals:
  tokens: 6758
  tasks: 2
  commits: 4

tech-stack:
  added: []
  patterns:
    - "WriteRescanResult only reuses the collision-resolution half of Duplicate (nextCopyRoot) -- it does not call DuplicateCatalog itself, since a re-scan has a freshly walked tree to write rather than existing bytes to copy"
    - "The .html sidecar decision is read from disk state at write time (os.Stat beside the original .json), never trusted from the renderer or from Options as passed in -- this is what makes 'rewritten when present, never created when absent' true for both write modes"
    - "ResolveRescan holds the walked tree from RescanCatalog on the App struct, guarded by the existing scan mutex and cleared on dialog close, kept strictly separate from Create's retained-partial fields"
    - "Resolution footer buttons never assert a fact the write path alone can guarantee -- the Keep both fix generalizes this: a label is not allowed to promise a filename unless it re-runs the exact resolution the backend will use"

key-files:
  created:
    - internal/catalog/resolve.go
    - internal/catalog/resolve_test.go
  modified:
    - app.go
    - app_test.go
    - frontend/src/types/rescan.ts
    - frontend/src/components/workspace/rescan/RescanDialog.tsx
    - frontend/src/workspace.css

key-decisions:
  - "ResolveRescan's mode parameter is a plain string, not catalog.ResolveMode -- ACCEPTED deviation (commit 9fa16794). Wails' codegen doesn't emit a TS type for a defined string-const type in internal/catalog referenced only as a parameter, the same limitation DiffState hit in plan 28-02/03. WriteRescanResult itself still takes the typed catalog.ResolveMode; only the binding's bridge-facing signature is a bare string."
  - "Keep both's label dropped its filename preview entirely rather than resolving the real name live at render time. Verification found the preview ('write {root}-copy.json') was not collision-checked and went stale the moment a -copy file already existed on disk. The smaller fix -- and the one that can never go stale again -- was removing the filename claim, not adding a second binding call to keep it accurate for a promise the button was never required to make."

requirements-completed: [ACT-07]

coverage:
  - id: D1
    description: "A user resolves a diff by overwriting the catalog, keeping both, or discarding -- three actions, exactly as locked, with no per-entry picker and no three-way merge"
    requirement: "ACT-07"
    verification:
      - kind: e2e
        ref: "dev-browser live session against wails dev :34115 (coordinator-resolved checkpoint) -- discard left rescan-fixture.json byte-identical (sha256 e3bb4dbd... unchanged, no confirmation prompt); keep-both created a -copy file and left the original at e3bb4dbd...; overwrite changed the hash to b12a275c... and rewrote the .html sidecar"
        status: pass
    human_judgment: false
  - id: D2
    description: "Keep both resolves its target filename through the SAME -copy/-copy-2 collision loop Duplicate already uses -- invoked, not reimplemented"
    requirement: "ACT-07"
    verification:
      - kind: unit
        ref: "TestWriteRescanResult_KeepBothUsesCopySuffix -- with the first candidate pre-created on disk, the write lands on the SECOND candidate"
        status: pass
      - kind: e2e
        ref: "This plan's own live re-verification (below) -- a second keep-both against an occupied -copy name landed on -copy-2, both via the coordinator's original checkpoint and this plan's own re-run of the collision case after the label fix"
        status: pass
    human_judgment: false
  - id: D3
    description: "The .html sidecar is rewritten when the original catalog had one, and is never created where none existed"
    requirement: "ACT-07"
    verification:
      - kind: unit
        ref: "TestWriteRescanResult_RewritesHtmlWhenPresent, TestWriteRescanResult_DoesNotCreateHtmlWhenAbsent"
        status: pass
      - kind: e2e
        ref: "Live: overwrite on a catalog with an .html rewrote it and produced no -copy file; overwrite on a catalog with no .html created none; this plan's fixture (which had an .html) produced rescan-fixture-copy-2.html alongside .json on keep-both"
        status: pass
    human_judgment: false
  - id: D4
    description: "The overwrite target is re-derived from the catalog's own on-disk path and re-validated with filepath.Abs -> filepath.EvalSymlinks -> osutil.ContainsPath at write time -- never trusted from the renderer's payload"
    requirement: "ACT-07"
    verification:
      - kind: unit
        ref: "TestResolveRescan_RejectsPathOutsideCatalogDir"
        status: pass
    human_judgment: false
  - id: D5
    description: "[UI E4-error] A write failure on Overwrite or Keep both surfaces as a 12px --danger inline banner on the same step, with the buttons re-enabled"
    requirement: "ACT-07"
    verification:
      - kind: e2e
        ref: "Live: chmod 500 on the catalog directory, clicked Overwrite -- .ws-rescan-resolve-error banner appeared with the real Go error, all three buttons stayed disabled:false, dialog stayed on the same step. Permissions restored afterward."
        status: pass
    human_judgment: false
  - id: D6
    description: "The Keep both button never names a file it will not write"
    requirement: "ACT-07"
    verification:
      - kind: e2e
        ref: "This plan's own fix and re-verification -- staged rescan-fixture-copy.json to force a collision, reloaded wails dev (bindings probed fresh via Object.keys(window.go.main.App)), drove the real UI (rail row -> Details -> Re-scan volume... -> stubbed SelectDirectory -> Start re-scan -> diff step). Footer HTML captured the button as plain 'Keep both' (no filename), then the click wrote rescan-fixture-copy-2.json/.html while rescan-fixture.json stayed byte-identical (sha256 d5187622... unchanged, content still only a.txt/b.txt) -- label and actual write target can no longer disagree."
        status: pass
    human_judgment: false

duration: ~45min (40min implementation/tasks + checkpoint resolution, ~15min for this plan's label fix + re-verification + closeout)
completed: 2026-08-16
status: complete
---

# Phase 28 Plan 04: Overwrite / Keep Both / Discard Resolution Summary

**`WriteRescanResult` and the `ResolveRescan` binding turn a completed re-scan diff into a decision -- overwrite in place, keep both via the shared collision loop, or discard with no write at all -- backed by the three-action resolution footer, all live-verified against real files on disk including a post-verification fix to a stale button label**

## Performance

- **Duration:** ~45 min total (implementation + checkpoint resolution ~40 min across `e51ab4f3`/`9fa16794`/`baee213b`; this session's label fix + live re-verification + closeout ~15 min)
- **Started:** 2026-08-16T16:22:53-05:00 (`e51ab4f3`)
- **Completed:** 2026-08-16 (this SUMMARY)
- **Tasks:** 2 (+ Task 3, a blocking human-verify checkpoint, resolved by the coordinator) + one accepted post-verification fix
- **Files modified:** 7 across the whole plan (2 new, 5 modified; excluding `.planning/`)

## Accomplishments

- `internal/catalog/resolve.go` (new): `ResolveMode` (`overwrite`/`keep-both`, discard deliberately excluded -- it has no Go call) and `WriteRescanResult`, which derives `dir`/`root` from the original catalog's `jsonPath` exactly the way `DuplicateCatalog` does, resolves the keep-both target through the existing `nextCopyRoot` collision loop (unmodified), reads `.html`-sidecar presence from disk at write time, and writes through `WriteCatalogFrom` -- the same hardened, SIGKILL-tested atomic-write path Create uses. No second write primitive was introduced.
- `app.go`'s `ResolveRescan` binding: re-derives and re-validates the write target from the catalog's own on-disk path with the identical `filepath.Abs` -> `filepath.EvalSymlinks` -> `osutil.ContainsPath` sequence `DeleteCatalog` uses, failing closed on an empty catalog directory. Holds the walked tree from `RescanCatalog` on the `App` struct behind the existing scan mutex, cleared on dialog close, kept strictly separate from Create's retained-partial fields.
- `RescanDialog`'s resolution footer: three locked actions (Overwrite catalog / Keep both / Discard scan and close), 56px `--p2` background, 1px top border, primary cluster left with the discard action isolated right via `margin-left: auto`. The overwrite button stays accent-filled (not danger-styled) per the UI-SPEC's own choice -- destructiveness is carried by the resolution caption 28-03 already shipped, not the button color. A write failure renders a 12px `--danger` inline banner above the footer, re-enables all three buttons, and keeps the dialog on the same step -- the same shape `RenameDialog`/`DeleteConfirmDialog` already use.
- **Live verification (coordinator-resolved checkpoint):** discard left the target catalog byte-identical (sha256 `e3bb4dbd…` unchanged, no confirmation prompt); keep-both created a `-copy` file and left the original at `e3bb4dbd…`; a second keep-both against an occupied `-copy` name landed on `-copy-2`; overwrite changed the hash to `b12a275c…` and rewrote the `.html` sidecar; overwrite on a catalog with no `.html` created none; a forced write failure showed the `.ws-rescan-resolve-error` banner with the real Go error while all three buttons stayed `disabled:false`.
- **Post-verification fix (this session):** the Keep both button's label previously read `"Keep both (write {root}-copy.json)"` -- a first-candidate preview that was never collision-checked. Verification caught it disagreeing with the real write target once `-copy.json` already existed (the write silently landed on `-copy-2.json` while the label kept promising `-copy.json`). Fixed by dropping the filename from the label entirely (`"Keep both"`), removing the now-unused `copyRootPreview` constant -- the smaller diff of the two options offered, and the one that cannot go stale again since it makes no filename claim at all. `nextCopyRoot` was not touched, extended, or given a second call site.

## Task Commits

1. **Task 1: `WriteRescanResult` and the `ResolveRescan` binding** - `e51ab4f3` (feat)
2. **Deviation: `ResolveRescan`'s `mode` param as plain string** - `9fa16794` (fix, accepted -- see Deviations)
3. **Task 2: Resolution footer -- three buttons, locked copy, inline write-failure banner** - `baee213b` (feat)
4. **Post-verification fix: Keep both label drops its stale filename claim** - `33dcdc9d` (fix)

## Files Created/Modified

- `internal/catalog/resolve.go` (new) - `ResolveMode`, `WriteRescanResult`
- `internal/catalog/resolve_test.go` (new) - four table-driven `TestWriteRescanResult_*` cases
- `app.go` - `ResolveRescan` binding, held-tree field and its clear-on-close path
- `app_test.go` - `TestResolveRescan_RejectsPathOutsideCatalogDir`, `TestResolveRescan_DiscardIsNotAWritePath`
- `frontend/src/types/rescan.ts` - `ResolveMode` mirror type
- `frontend/src/components/workspace/rescan/RescanDialog.tsx` - three-action resolution footer, inline write-failure banner, `handleResolve`; Keep both's label fix
- `frontend/src/workspace.css` - `.ws-rescan-footer-actions`, `.ws-rescan-btn-outline`, `.ws-rescan-resolve-error`

## Decisions Made

- **`ResolveRescan`'s `mode` parameter stays a plain `string` across the Wails bridge**, not `catalog.ResolveMode` -- Wails codegen doesn't emit a TS type for a defined string-const type in `internal/catalog` referenced only as a parameter, the same limitation `DiffState` hit in plan 28-02/03. `WriteRescanResult` itself still takes the typed `catalog.ResolveMode`; only the bridge-facing binding signature is untyped at the boundary.
- **Keep both's label drops the filename entirely rather than resolving it live.** The alternative (a synchronous or async call to resolve the real name at render time) would have needed new binding surface and would still go stale between render and click if the directory changed underneath it. The write itself, via the unmodified `nextCopyRoot` collision loop, stays the sole authority on the actual filename -- the label now makes no promise the write could contradict.

## Deviations from Plan

### Accepted Deviation (from implementation)

**1. [Rule 3 — blocking, Wails codegen limitation] `ResolveRescan`'s `mode` parameter is a plain `string`, not `catalog.ResolveMode`.**

- **Found during:** Task 1, wiring the Wails binding.
- **Issue:** Wails' TS codegen does not emit a type for a Go string-const type defined in `internal/catalog` when it's referenced only as a function parameter -- the same gap `DiffState` hit crossing the bridge in plan 28-02/03.
- **Fix:** `ResolveRescan(jsonPath string, catalogDir string, mode string)` at the binding boundary; `WriteRescanResult` (internal, not bridge-crossing) keeps the typed `catalog.ResolveMode` parameter.
- **Files modified:** `app.go`.
- **Commit:** `9fa16794`.

### Accepted Fix (post-verification, this session)

**2. [Rule 1 — bug, user-reviewed and approved] Keep both's label named a file it would not always write.**

- **Found during:** Live verification of the collision case -- with `photos-a-copy.json` already present, the button still read `"Keep both (write photos-a-copy.json)"` while the write actually landed on `photos-a-copy-2.json`.
- **Fix:** Dropped the filename from the label (`"Keep both"`), removed the unused `copyRootPreview` constant and its explanatory comment. No new binding call, no second naming path -- `nextCopyRoot` remains the sole place collision resolution happens.
- **Files modified:** `frontend/src/components/workspace/rescan/RescanDialog.tsx`.
- **Commit:** `33dcdc9d`.
- **Verification:** Re-ran the exact collision scenario live (see below) -- footer HTML showed plain `"Keep both"` before the click, and the write landed on `rescan-fixture-copy-2.json`/`.html`, confirming the label can no longer disagree with the write.

---

**Total deviations:** 2 (one accepted Wails-codegen limitation carried from Task 1, one accepted bug fix from this session's own live verification). No architectural deviations, no Rule 4 escalations.

## This Session's Live Verification (dev-browser, `wails dev` on :34115)

Fresh `wails dev` started (no stale process was found on :34115 at session start; `lsof`/`ps` both confirmed clear). Bindings probed via `Object.keys(window.go.main.App)` before use -- `ResolveRescan` and `RescanCatalog` both present, confirming fresh bindings. No `osascript`/System Events/host-OS GUI automation was used at any point.

**Fixture setup:** a source folder (`a.txt`, `b.txt`) cataloged via `StartScan` into a separate catalog directory (writing both `.json` and `.html`), a `rescan-fixture-copy.json`/`.html` pair pre-created by hand to force the collision, then a third file (`c.txt`) added to the source before re-scanning.

**Driven through the real UI** (not direct binding calls for the interaction itself): selected the catalog in the rail, opened the Details panel, clicked "Re-scan volume…", stubbed `window.go.main.App.SelectDirectory` in-page (pure browser-sandbox JS, not host GUI automation) to bypass the native folder dialog, clicked "…or choose any folder", clicked "Start re-scan", reached the diff step ("Re-scan changed 1 entries", `ADDED · 1 → ./source/c.txt`).

**Findings:**
- Footer HTML at the diff step: `<button class="ws-rescan-btn-outline">Keep both</button>` -- no filename anywhere in the label, confirming the fix.
- Clicked "Keep both". Result on disk: `rescan-fixture-copy-2.json`/`.html` created (containing the re-scanned tree with all three files); `rescan-fixture-copy.json` (the pre-existing collision file) untouched; `rescan-fixture.json` unchanged at sha256 `d5187622b67a5e745dc1600a82af60e9450fccbc58bbdd8b3b9bb495ed2945e9` both before and after the click, still containing only `a.txt`/`b.txt` -- byte-identical, proving the collision loop (not a naive single-suffix append) resolved the write and the original was never touched.
- Dialog closed on success, rail refreshed to show `rescan-fixture-copy-2` (visible in the rail row list immediately after).

**Cleanup:** fixture directory removed from the scratchpad; `wails dev` and its child process killed (confirmed via `lsof -i :34115` returning empty and `ps` showing no `wails`/`storcat` process); `frontend/wailsjs/runtime/{package.json,runtime.d.ts,runtime.js}` file-mode changes (0755→0644, `wails dev`'s known noise) reverted via `git checkout --`. Working tree confirmed clean before this SUMMARY's commit.

## Known Stubs

None. `WriteRescanResult`, `ResolveRescan`, and the three-action footer are all fully wired and live-verified, including the corrected Keep both label.

## Issues Encountered

None beyond the label staleness itself, which is the subject of this session's fix.

## User Setup Required

None — no external service configuration required.

## Next Phase Readiness

- All three re-scan resolutions (overwrite, keep-both, discard) are implemented, unit-tested, and live-verified against real files, including the corrected collision-case label.
- `internal/catalog/atomicwrite.go` and `internal/catalog/duplicate.go` remain unmodified (`git diff HEAD` on both stays empty) -- the crash-safe write and the collision loop were reused, never re-implemented.
- Plan 28-05 (the `UnreadableCatalogPanel` entry point, `oldTreeAvailable: false`) can build directly on this plan's `ResolveRescan` binding and footer component.
- `.planning/WINDOWS.md`'s open entry #11 (parent-directory fsync unsupported on Windows) gains a new call site through `WriteRescanResult` -- same open status, not closed by this plan; plan 28-06's sweep records it.

---
*Phase: 28-re-scan-diff*
*Completed: 2026-08-16*

## Self-Check: PASSED

All files created/modified verified present on disk (`internal/catalog/resolve.go`, `internal/catalog/resolve_test.go`, `app.go`, `app_test.go`, `frontend/src/types/rescan.ts`, `frontend/src/components/workspace/rescan/RescanDialog.tsx`, `frontend/src/workspace.css`, this SUMMARY); all four commit hashes (`e51ab4f3`, `9fa16794`, `baee213b`, `33dcdc9d`) verified present in `git log`. `go build ./... && go vet ./... && go test ./... -race -count=1` green. `cd frontend && npx tsc --noEmit && npm run build` green. Working tree clean, no process listening on :34115.
