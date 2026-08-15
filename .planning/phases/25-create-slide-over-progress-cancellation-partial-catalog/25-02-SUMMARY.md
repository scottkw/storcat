---
phase: 25-create-slide-over-progress-cancellation-partial-catalog
plan: 02
subsystem: catalog
tags: [go, atomic-write, error-classification, partial-catalog, compat]

# Dependency graph
requires:
  - phase: 25-01
    provides: "internal/catalog.CreateCatalogWithContext(ctx, ...)/WriteCatalogFrom shared write path, Options struct, CLI-wrapper-untouched pattern this plan extends"
provides:
  - "internal/catalog.WriteFileAtomic(path, data, perm) error -- crash-safe temp-then-rename write, exported for Phase 27's rename/duplicate/delete reuse"
  - "internal/catalog.SourceUnavailableError/PartialScan/ReadErrorEntry -- typed terminal-source-loss signal carrying everything walked before the loss"
  - "Options.HaltOnSourceLoss -- CLI-safe-by-default terminal-classification toggle for Phase 25's remaining GUI plans (CRT-10/CRT-11 error/partial-catalog UI)"
  - "CatalogItem.Unreadable/ReadError -- the on-disk partial-catalog marker shape (option-a, user-decided at this plan's blocking checkpoint)"
affects: [25-03, 25-04, 25-05, 25-06, 25-07, 27]

# Actuals (#2632)
actuals:
  tokens: 7494
  tasks: 2
  commits: 2

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Atomic write (temp file in destination dir + os.Rename) generalized from internal/config/counts_cache.go into a package-level, exported primitive -- reused by writeJSONFile/writeHTMLFile this plan, and by Phase 27's rename/duplicate/delete next"
    - "Root re-probe classification: any read failure triggers a cheap os.Stat on the scan root (never the failing subdirectory) to distinguish one bad entry from a vanished source -- the single heuristic that both the directory-ReadDir-failure site and the child-iteration site share"
    - "Errors propagate as a typed *SourceUnavailableError checked via errors.As, not a package-level sentinel -- context.Canceled already covers the cancellation case, so no second synonym was added"

key-files:
  created:
    - internal/catalog/atomicwrite.go
    - internal/catalog/atomicwrite_test.go
    - internal/catalog/errors.go
  modified:
    - internal/catalog/service.go
    - internal/catalog/service_test.go
    - internal/catalog/options.go
    - internal/search/service_test.go
    - pkg/models/catalog.go
    - frontend/wailsjs/go/models.ts

key-decisions:
  - "Marker shape resolved to option-a (two flat omitempty scalars: Unreadable bool, ReadError string) at this plan's blocking checkpoint:decision task -- a USER decision at a one-way-door gate, not a Claude's-discretion pick. The user was presented three options (25-RESEARCH.md's flat-scalars recommendation, a nested-object growable alternative, and a single presence-implies-unreadable string) and selected option-a directly after reviewing the tradeoffs. Rejected: option-b (nested `unreadable:{reason,at}` object) -- speculative structure for a second attribute nothing in this plan or 25-RESEARCH.md names a need for; option-c (bare `readError` string, presence implies unreadability) -- disqualified outright, not just costlier, because an unreadable node with an empty reason string becomes byte-identical to a clean node, which is exactly the invisible-data-loss failure the marker exists to prevent."
  - "Errors propagate internally as *SourceUnavailableError with Partial left nil until the outermost CreateCatalogWithContext populates it once -- avoids rebuilding PartialScan at every recursion level while still giving every level's caller an errors.As-matchable signal to stop descending"
  - "Terminal detection sets Unreadable/ReadError ONLY on the origin node (where classify() first returns true); every ancestor the error propagates through returns its own accumulated node unmarked, per the plan's explicit 'and only on that node' instruction"

requirements-completed: [CRT-09, CRT-10, CRT-11, COMPAT-02, COMPAT-03]

coverage:
  - id: D1
    description: "Every catalog JSON/HTML write goes through a temp file created in the destination directory, then renamed -- a crash mid-write can never leave a truncated file at the destination path"
    verification:
      - kind: unit
        ref: "internal/catalog/atomicwrite_test.go#TestWriteFileAtomic_CreatesFileWithContent"
        status: pass
      - kind: unit
        ref: "internal/catalog/atomicwrite_test.go#TestWriteFileAtomic_LeavesNoTempResidue"
        status: pass
      - kind: unit
        ref: "internal/catalog/atomicwrite_test.go#TestWriteFileAtomic_RemovesTempOnFailure"
        status: pass
      - kind: unit
        ref: "internal/catalog/atomicwrite_test.go#TestWriteFileAtomic_TempIsCreatedInDestinationDirectory"
        status: pass
      - kind: unit
        ref: "internal/catalog/service_test.go#TestWriteJSONFile_BareObject"
        status: pass
    human_judgment: false
  - id: D2
    description: "A single unreadable entry is still skipped and the walk continues, byte-for-byte unchanged from v2.3.0 -- only the scan root vanishing stops the walk"
    requirement: CRT-10
    verification:
      - kind: unit
        ref: "internal/catalog/service_test.go#TestTraverseDirectory_SingleEntryErrorSkipsAndContinues"
        status: pass
      - kind: unit
        ref: "internal/catalog/service_test.go#TestCreateCatalog_WrapperDoesNotHaltOnSourceLoss"
        status: pass
    human_judgment: false
  - id: D3
    description: "A scan-root loss mid-walk is classified terminal (root re-probe), stops descending, keeps everything already walked, and returns a typed error carrying the partial tree"
    requirement: CRT-10
    verification:
      - kind: unit
        ref: "internal/catalog/service_test.go#TestTraverseDirectory_TerminalSourceLossStopsWalk"
        status: pass
    human_judgment: false
  - id: D4
    description: "A source-loss error writes nothing -- the write path is never reached (CRT-09/CRT-11 adjacency)"
    requirement: CRT-11
    verification:
      - kind: unit
        ref: "internal/catalog/service_test.go#TestCreateCatalogWithContext_SourceLossWritesNothing"
        status: pass
    human_judgment: false
  - id: D5
    description: "A partial catalog's JSON carries the marker keys on exactly the affected node and no other"
    requirement: CRT-11
    verification:
      - kind: unit
        ref: "internal/catalog/service_test.go#TestWritePartialCatalog_Marker"
        status: pass
    human_judgment: false
  - id: D6
    description: "A clean scan's JSON is byte-for-byte the v2.3.0 shape -- the two marker keys never appear"
    requirement: COMPAT-02
    verification:
      - kind: unit
        ref: "internal/catalog/service_test.go#TestCreateCatalog_JSONShapeUnchanged"
        status: pass
      - kind: other
        ref: "grep -c 'unreadable' on a real `storcat create` output -- 0"
        status: pass
    human_judgment: false
  - id: D7
    description: "Every existing reader (LoadCatalog, LoadCatalogFlat, cli show's consumer) tolerates the marker fields with zero code change, since no reader rejects unrecognized JSON keys"
    requirement: COMPAT-02
    verification:
      - kind: unit
        ref: "internal/search/service_test.go#TestLoadCatalog_ToleratesMarkerFields"
        status: pass
      - kind: unit
        ref: "internal/search/service_test.go#TestLoadCatalogFlat_ToleratesMarkerFields"
        status: pass
    human_judgment: false
  - id: D8
    description: "cli/create.go is provably unedited and the CLI wrapper never sets HaltOnSourceLoss -- CLI behavior is unchanged (COMPAT-03)"
    requirement: COMPAT-03
    verification:
      - kind: other
        ref: "git diff --stat -- cli/create.go -- empty"
        status: pass
      - kind: other
        ref: "grep -n HaltOnSourceLoss internal/catalog/service.go -- absent from the wrapper's Options literal"
        status: pass
    human_judgment: false

duration: 16min
completed: 2026-08-14
status: complete
---

# Phase 25 Plan 2: Atomic Writes + Terminal-Source-Loss Classification Summary

**Crash-safe catalog writes via temp-file-then-rename, plus a real walk error contract that classifies a vanished scan root (terminal) from a single bad entry (skip-and-continue, unchanged) -- with the on-disk partial-catalog marker shape decided by the user at this plan's blocking checkpoint.**

## Performance

- **Duration:** 16 min (active task execution; excludes the checkpoint deliberation, which paused for the user's decision)
- **Started:** 2026-08-14T18:52:39-05:00 (Task 1 commit)
- **Completed:** 2026-08-14T19:09:07-05:00 (Task 2 commit)
- **Tasks:** 2 (plus the plan's blocking checkpoint:decision task, resolved by the user before Task 1/2 executed)
- **Files modified:** 9

## Accomplishments
- `WriteFileAtomic` (new, exported): every catalog JSON/HTML write now goes through a temp file in the destination directory + `os.Rename`, generalized from `internal/config/counts_cache.go`'s proven pattern, ready for Phase 27's rename/duplicate/delete-to-Trash to reuse without retrofitting
- `traverseDirectory` now distinguishes "one bad file" from "the volume is gone": a re-probe of the scan root (never the failing subdirectory) classifies any read failure, and only a root-unreachable failure stops the walk -- single-entry failures keep v2.3.0's skip-and-continue behavior byte-for-byte
- `CreateCatalogWithContext` now branches on three outcomes before ever touching the write path: cancelled (write nothing), source-loss (write nothing, return the typed error with its populated partial tree), or success (proceed to write)
- `CatalogItem.Unreadable`/`ReadError` -- the on-disk partial-catalog marker, shape decided by the user at the plan's blocking checkpoint -- mark only the origin directory node, are `omitempty`, and leave every clean-scan catalog byte-identical to v2.3.0 (COMPAT-02)
- `cli/create.go` remains provably unedited; the CLI wrapper's `Options` literal never sets `HaltOnSourceLoss`, so the CLI's observable behavior is unchanged (COMPAT-03)

## Task Commits

Each task was committed atomically:

1. **Task 1: Atomic catalog writes -- temp file in the destination directory, then rename** - `16769a65` (feat)
2. **Task 2: Classify read failures and mark the unreadable subtree on disk** - `42b6b745` (feat)

_Note: this plan's checkpoint:decision task (the marker-shape choice) preceded Task 1/2 and produced no commit of its own -- it was resolved by the user's explicit selection, recorded above under Decisions Made._

## Files Created/Modified
- `internal/catalog/atomicwrite.go` - `WriteFileAtomic(path, data, perm) error`, temp-in-destination-dir + rename, exported for Phase 27 reuse
- `internal/catalog/atomicwrite_test.go` - content correctness, no temp residue on success/failure, temp-in-destination-dir assertion
- `internal/catalog/errors.go` - `ReadErrorEntry`, `PartialScan`, `SourceUnavailableError` (errors.As-matchable; no second cancellation sentinel)
- `internal/catalog/options.go` - `Options.HaltOnSourceLoss` (default false; CLI wrapper never sets it)
- `internal/catalog/service.go` - `walkState.readErrorEntries`/`terminal`/`recordReadError()`/`classify()` (capped at 50 entries), terminal-aware `traverseDirectory`, three-outcome `CreateCatalogWithContext`, `writeJSONFile`/`writeHTMLFile` delegating to `WriteFileAtomic`
- `internal/catalog/service_test.go` - 6 new tests: single-entry skip, terminal classification, source-loss-writes-nothing, wrapper-never-halts, partial-catalog marker shape
- `pkg/models/catalog.go` - `CatalogItem.Unreadable`/`ReadError`, both `omitempty`, appended after `Contents`
- `internal/search/service_test.go` - `TestLoadCatalog_ToleratesMarkerFields`, `TestLoadCatalogFlat_ToleratesMarkerFields`
- `frontend/wailsjs/go/models.ts` - regenerated via `wails generate module`; `CatalogItem` gains the two optional fields

## Decisions Made

- **Marker shape: option-a, two flat omitempty scalars** (`Unreadable bool`, `ReadError string`) -- decided by the **user** at this plan's `checkpoint:decision gate="blocking"` task, a one-way-door choice (once v3.0.0 ships, partial catalogs carrying these keys exist on disk and renaming them later is a migration, not a refactor). The executor presented three options with tradeoffs and a recommendation; the user reviewed and selected option-a directly.
  - **Rejected: option-b** (nested `unreadable: {reason, at}` object) -- speculative structure for a second attribute (e.g. a timestamp) that nothing in this plan, 25-RESEARCH.md, or the requirements list (CRT-10/CRT-11) actually names a need for.
  - **Rejected: option-c** (bare `readError` string, presence implies unreadability) -- disqualified outright rather than merely costlier: an unreadable node with an empty reason string becomes byte-identical to a clean node, exactly the invisible-data-loss failure this marker exists to prevent.
- Errors propagate internally as `*SourceUnavailableError` with `Partial` left `nil` at every intermediate recursion level, populated once by the outermost `CreateCatalogWithContext` after the walk returns -- avoids rebuilding `PartialScan` at each level while every intermediate caller still gets an `errors.As`-matchable signal telling it to stop descending and propagate.
- Terminal detection marks `Unreadable`/`ReadError` only on the origin node (where `classify()` first returns true); every ancestor the error propagates through returns its own accumulated-so-far node unmarked, per the plan's explicit "and only on that node" instruction.
- `maxReadErrorEntries = 50` caps recorded entries (not the counter) so a device failing on every read can't grow an unbounded slice -- matches the plan's `T-25-10` mitigation.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Doc comments accidentally matched the plan's own literal grep acceptance criteria**
- **Found during:** Task 1 and Task 2, running the plan's own acceptance-criteria greps
- **Issue:** `atomicwrite.go`'s doc comment mentioned `os.TempDir()` and `os.CreateTemp` in prose (to explain the pattern), which made `grep -c 'os.TempDir'` return 1 instead of the required 0, and `grep -c 'os.CreateTemp'` return 2 instead of the required 1. Separately, `pkg/models/catalog.go`'s doc comment used the word "omitempty" in prose, pushing `grep -c 'omitempty'` to 5 instead of the required 4 (baseline 2 + 2 new). A `service_test.go` doc comment for the new reader-tolerance test also literally said `DisallowUnknownFields`, tripping `grep -c 'DisallowUnknownFields' internal/ cli/ -r` to 1 instead of 0.
- **Fix:** Reworded all three comments to describe the same behavior without using the literal grep-target strings (e.g. "the shared system temp directory" instead of naming `os.TempDir()`; "absent-when-zero" instead of "omitempty"; "rejects unrecognized JSON keys" instead of naming `DisallowUnknownFields`).
- **Files modified:** internal/catalog/atomicwrite.go, pkg/models/catalog.go, internal/search/service_test.go
- **Verification:** Re-ran every acceptance-criteria grep from the plan; all now match the plan's exact expected counts.
- **Committed in:** `16769a65` (Task 1) and `42b6b745` (Task 2) -- fixed before each task's commit, not as a follow-up.

---

**Total deviations:** 1 auto-fixed (comment wording only; zero behavior change).
**Impact on plan:** No impact on architecture or scope -- purely doc-comment wording adjustments to satisfy the plan's own literal acceptance-criteria greps.

## Issues Encountered

None beyond the deviation above. The `wails` CLI was available and `wails generate module` regenerated `frontend/wailsjs/go/models.ts` cleanly with no running dev server required (it's a static generation over the Go source, not a runtime call) -- consistent with the coordinator's note that this plan needed no browser verification.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- `internal/catalog.WriteFileAtomic` is exported and ready for Phase 27's rename/duplicate/delete-to-Trash operations to reuse without retrofitting crash-safety.
- `Options.HaltOnSourceLoss`, `SourceUnavailableError`, and `PartialScan` are in place for the remaining Phase 25 GUI plans (CRT-10's error state, CRT-11's write-partial-catalog/retry/cancel actions) to bind directly -- `App.StartScan` can set `HaltOnSourceLoss: true` and `errors.As` the returned error to route to the error/partial-catalog UI.
- The marker shape (`CatalogItem.Unreadable`/`ReadError`) is now the locked, on-disk one-way door for every future partial catalog this app writes; any future phase adding a third partial-catalog attribute must add a third top-level `omitempty` key, not restructure these two.
- No volume detection, options toggles, WILL WRITE preview, or the actual error/partial-catalog UI exist yet -- this plan's scope was the backend error contract and write-safety only, per its own `files_modified` list.

---
*Phase: 25-create-slide-over-progress-cancellation-partial-catalog*
*Completed: 2026-08-14*

## Self-Check: PASSED

All 9 created/modified files verified present on disk; both task commit hashes (`16769a65`, `42b6b745`) verified present in git history.
