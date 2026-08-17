---
phase: 27-catalog-actions-watch
plan: 02
subsystem: infra
tags: [go, fsync, atomic-write, crash-safety, subprocess-test, sigkill]

# Dependency graph
requires:
  - phase: 27-catalog-actions-watch
    provides: "27-01's WriteFileAtomic-based RenameCatalog, which this plan's fsync hardening now backs with real crash-safety evidence"
provides:
  - "WriteFileAtomic: File.Sync() on the temp file before close, plus a best-effort syncDir(filepath.Dir(path)) after os.Rename, with its failure logged (never silently discarded)"
  - "internal/catalog's first subprocess-based crash test: a real, separately-launched process is SIGKILLed mid-write and the pre-existing destination is proven byte-identical (sha256) across repeated runs"
  - "the repo's first testdata/ standalone helper binary pattern (killtarget), reusable by any future crash-safety test"
affects: [27-03, 27-04, 27-05, 27-06, 27-07]

# Actuals (#2632)
actuals:
  tokens: 5002
  tasks: 2
  commits: 2

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Marker-file synchronization (fsync'd <dest>.killtarget-ready) instead of polling for a temp file's mere appearance -- eliminates the flakiest variable in a subprocess timing test"
    - "Standalone testdata/ helper binary (package main, no imports from this project) built via `go build -o <tmp> ./testdata/<name>` inside the test, for reproducing a real external process to kill/signal"

key-files:
  created:
    - internal/catalog/atomicwrite_sigkill_test.go
    - internal/catalog/testdata/killtarget/main.go
  modified:
    - internal/catalog/atomicwrite.go
    - internal/catalog/atomicwrite_test.go

key-decisions:
  - "Added the parent-directory fsync (recorded_decision id=parent-directory-fsync in 27-02-PLAN.md), with syncDir's error captured into a named variable and log.Printf'd, never `_ = syncDir(...)` -- a persistently-failing directory sync must be observable, not permanently unobservable behind a successful write"
  - "Documented (not silently forced) a literal acceptance-grep mismatch: the plan's `awk '/os.Rename.../,0' | grep -c os.Remove` expects 0, but the pre-existing rename-failure branch's own `os.Remove(tmpPath)` falls inside that awk range because the range starts at the os.Rename line itself (inclusive). The plan's own <action> text requires 'leaving... every existing failure path intact', so this call was kept; removing it to satisfy the literal grep would delete required correctness behavior. Same class of mismatch 27-01-SUMMARY.md documented for its own literal greps."
  - "killtarget's payload write is split into two halves (matching the plan's <action> literal wording) with the fsync'd ready-marker written after the first half and before the delay -- landed a real kill inside the write window on all 42 iterations across both tests, on every one of several repeated runs, so no fallback synchronization mechanism was needed"

requirements-completed: [ACT-09]

coverage:
  - id: D1
    description: "WriteFileAtomic flushes the temp file's own bytes with File.Sync() before it is closed, so the destination's contents are durable and not merely buffered"
    requirement: "ACT-09"
    verification:
      - kind: unit
        ref: "internal/catalog/atomicwrite_test.go#TestWriteFileAtomic_SyncsBeforeRename"
        status: pass
      - kind: integration
        ref: "internal/catalog/atomicwrite_sigkill_test.go#TestWriteFileAtomic_SurvivesKill"
        status: pass
    human_judgment: false
  - id: D2
    description: "WriteFileAtomic best-effort fsyncs the destination's parent directory after os.Rename, logging (not silently discarding) a sync failure, without ever propagating that failure as a write error"
    requirement: "ACT-09"
    verification:
      - kind: unit
        ref: "internal/catalog/atomicwrite_test.go#TestWriteFileAtomic_DirSyncFailureIsNotFatal"
        status: pass
    human_judgment: false
  - id: D3
    description: "A real SIGKILL delivered mid-write leaves a pre-existing destination byte-identical to its pre-write content, across at least 20 iterations at varied timings, via a genuinely separate OS process"
    requirement: "ACT-09"
    verification:
      - kind: integration
        ref: "internal/catalog/atomicwrite_sigkill_test.go#TestWriteFileAtomic_SurvivesKill"
        status: pass
    human_judgment: false
  - id: D4
    description: "With no pre-existing destination, a mid-write SIGKILL leaves the destination absent, never present-but-truncated"
    requirement: "ACT-09"
    verification:
      - kind: integration
        ref: "internal/catalog/atomicwrite_sigkill_test.go#TestWriteFileAtomic_SurvivesKill_NoPriorFile"
        status: pass
    human_judgment: false
  - id: D5
    description: "Two concurrent WriteFileAtomic calls to the same destination each create their own uniquely-named temp file, exactly one rename wins, and no temp residue remains"
    requirement: "ACT-09"
    verification:
      - kind: unit
        ref: "internal/catalog/atomicwrite_test.go#TestWriteFileAtomic_ConcurrentWritersLeaveNoResidue"
        status: pass
    human_judgment: false

duration: 5min
completed: 2026-08-15
status: complete
---

# Phase 27 Plan 02: Atomic write hardening + real SIGKILL proof Summary

**`WriteFileAtomic` now `Sync()`s the temp file before close and best-effort `fsync`s the parent directory after rename (logging, not swallowing, a sync failure); a new subprocess SIGKILL harness proves the crash-safety claim with a real killed process, not a unit-test assumption -- `.planning/WINDOWS.md` #6 is dischargeable on real evidence.**

## Performance

- **Duration:** ~5 min
- **Started:** 2026-08-15T17:38:22Z
- **Completed:** 2026-08-15T17:43:08Z
- **Tasks:** 2
- **Files modified:** 4 (2 modified, 2 created)

## Accomplishments
- `WriteFileAtomic` gained the two fsyncs POSIX durability actually requires: `tmp.Sync()` before close (the CONTEXT.md-locked half), and a best-effort `syncDir(filepath.Dir(path))` after `os.Rename` (this plan's explicit, recorded discretion decision) -- with the directory-sync failure `log.Printf`'d rather than silently discarded, so a persistently-failing sync stays observable.
- Built the repo's first subprocess-based crash test: `internal/catalog/testdata/killtarget/main.go`, a standalone `go build`-able helper that reproduces the write sequence with an injected mid-write delay and a `Sync()`'d `<dest>.killtarget-ready` marker for deterministic synchronization (not polling-on-hope).
- `TestWriteFileAtomic_SurvivesKill` and `TestWriteFileAtomic_SurvivesKill_NoPriorFile` each launch a real OS process, wait for the marker, `cmd.Process.Kill()` (real `SIGKILL`), and assert survival: 21 iterations each (42 total) at delays cycling through 20/50/120ms. **Every single iteration landed the kill successfully inside the write window** (confirmed by 100% `storcat-*.tmp` residue left behind by the killed process each time -- the killed process never reached its own `os.Remove`), and every iteration's destination was byte-identical (SHA-256-verified) to its pre-write seed, or correctly absent when none existed. Re-run 3+ times with identical results -- not a single-shot fluke.

## Task Commits

1. **Task 1: Harden WriteFileAtomic** - `6be3499f` (feat) -- `tmp.Sync()`, `syncDir`, doc-comment durability contract, 4 new unit tests
2. **Task 2: Prove it with a real SIGKILL** - `8f690569` (test) -- `testdata/killtarget/main.go`, `atomicwrite_sigkill_test.go`

**Plan metadata:** pending (this SUMMARY's own commit)

## Files Created/Modified
- `internal/catalog/atomicwrite.go` - `tmp.Sync()` before close; new unexported `syncDir(dir string) error`; called (error logged, not propagated) after `os.Rename`; doc comment records the full four-step durability contract and the recorded decision
- `internal/catalog/atomicwrite_test.go` - `TestWriteFileAtomic_SyncsBeforeRename`, `_ReplacesExistingFileWholesale`, `_DirSyncFailureIsNotFatal`, `_ConcurrentWritersLeaveNoResidue`
- `internal/catalog/testdata/killtarget/main.go` - standalone `package main` helper, no imports from this project, reproduces the write-temp-then-rename sequence with an injected two-half delay and a marker file; excluded from `go build ./...`/`go vet ./...` by living under `testdata/`
- `internal/catalog/atomicwrite_sigkill_test.go` - `TestWriteFileAtomic_SurvivesKill`, `TestWriteFileAtomic_SurvivesKill_NoPriorFile`, plus shared helpers (`buildKillTargetHelper`, `waitForMarker`, `waitForProcessDeath`, `cleanupIteration`)

## Decisions Made
- **Parent-directory fsync: ADDED**, per this plan's `<recorded_decision id="parent-directory-fsync">`. `syncDir`'s error is captured into a named variable and `log.Printf`'d -- never discarded via `_ = syncDir(...)` -- because a directory sync that fails persistently would otherwise leave exactly the durability hole this fsync exists to close, while every write still reports success.
- **Documented, not forced, a literal-grep/behavior conflict** (same class as 27-01-SUMMARY's precedent): the plan's acceptance criterion `awk '/os.Rename\(tmpPath, path\)/,0' internal/catalog/atomicwrite.go | grep -c 'os.Remove'` expects `0`, but the awk range is inclusive of the `if err := os.Rename(...)` line itself, which is immediately followed by the **pre-existing** rename-failure `os.Remove(tmpPath)` this plan's `<action>` explicitly required to leave intact ("leaving... every existing failure path intact"). The grep returns `1`, not `0`. This is not a functional bug -- the removal is required behavior on a failed rename, and it is not a *new* failure path added by this task, which is what the acceptance criterion's prose actually says it's checking for. Verified behaviorally instead: `TestWriteFileAtomic_RemovesTempOnFailure` (pre-existing, unmodified, still passing) proves that exact path.
- **killtarget's marker-before-delay synchronization worked on the first attempt** and was re-verified reliable across multiple full test runs -- no fallback mechanism (e.g., a named pipe) was needed, contrary to `27-RESEARCH.md`'s own LOW-confidence rating of this harness design.

## Deviations from Plan

### Acceptance-criterion mismatch (not a bug, documented for the record)

**1. Task 1's literal `os.Remove` count-after-rename grep returns 1, not 0**
- **Found during:** Task 1 verification
- **Cause:** See "Decisions Made" above -- the awk range starting at the `os.Rename` match line inclusively captures the pre-existing, required `os.Remove(tmpPath)` in the rename-failure branch, which the plan's own `<action>` text mandates be left intact.
- **Action taken:** Kept the correct, required failure-handling code; did not delete it to force a literal grep to `0`. All other acceptance greps for this task pass exactly as specified (`tmp.Sync()` count 1, `func syncDir` count 1, `syncDir(` count 2, call site after rename, `log.Printf` count 1, blank-identifier discard count 0).
- **Verification:** `go test ./internal/catalog/... -run TestWriteFileAtomic -race -count=1 -v` -- all 8 tests (4 pre-existing + 4 new) pass; `go test ./... -race -count=1` green.

**2. Task 2's helper source-file comment initially tripped its own "no production import" acceptance grep**
- **Found during:** Task 2 verification
- **Cause:** `internal/catalog/testdata/killtarget/main.go`'s doc comment explained the helper "deliberately does not import `storcat-wails/internal/catalog`" -- the literal substring the acceptance grep (`! grep -q 'storcat-wails/internal/catalog' ...`) checks for, matched inside a comment rather than an actual import statement.
- **Action taken:** Reworded the comment to convey the same fact ("imports no package from this project") without the literal forbidden substring. Fixed before commit; no functional code change.
- **Verification:** `! grep -q 'storcat-wails/internal/catalog' internal/catalog/testdata/killtarget/main.go` succeeds; `go build ./... && go vet ./...` still green (the file has never actually imported anything from this project).

---

**Total deviations:** 2, both acceptance-criterion/literal-grep mismatches -- one against pre-existing required code (documented, not forced), one a wording fix applied before commit. No functional bug, no scope creep.
**Impact on plan:** None on correctness. Both are the same class of literal-vs-intent mismatch 27-01-SUMMARY.md already established a precedent for documenting rather than silently claiming compliance.

## Issues Encountered
None. The SIGKILL harness -- rated LOW confidence in `27-RESEARCH.md` and flagged as having no in-repo analog -- worked reliably on the first implementation attempt and stayed reliable across more than 3 full repeated runs (~63 total kill iterations with a 100%-mid-write-landing rate). No fallback synchronization mechanism (e.g., a named pipe) was required.

## WINDOWS.md #6 verdict

**`.planning/WINDOWS.md` #6 is dischargeable on real evidence, not assertion.** Its exact wording claimed the crash-safety guarantee rested on unit tests and `os.Rename`'s OS-level atomicity, "not a live kill test." That is no longer true:

- `TestWriteFileAtomic_SurvivesKill` launched 21 genuinely separate OS processes (via `exec.Command` + `Start()`, never an in-process `panic`/`t.Fatal`/`os.Exit`), waited for each to reach a deterministic mid-write point (the `Sync()`'d `<dest>.killtarget-ready` marker), and delivered a real `cmd.Process.Kill()` (`SIGKILL` on this platform) at three varied timings (20/50/120ms into the injected delay window). Every iteration's pre-existing destination came out byte-identical (verified by both length and SHA-256) to its pre-write content, and still parsed as valid JSON.
- `TestWriteFileAtomic_SurvivesKill_NoPriorFile` ran the identical 21-iteration loop with no pre-existing destination; every iteration left the destination genuinely absent, never present-but-truncated.
- Both tests were re-run more than 3 times end-to-end with identical pass results -- this is not a single lucky timing.
- `waitForProcessDeath` asserts (via the portable `exec.ExitError.ProcessState.Exited() == false`) that the process actually died from a signal, not a graceful exit, closing the exact gap `27-RESEARCH.md`'s Pitfall 8 warned an in-process simulation would leave open.

This plan does **not** itself flip `.planning/WINDOWS.md` #6 to `fixed` -- per `27-02-PLAN.md`'s own "Ledger input (consumed by 27-07)" artifact note, that ledger update belongs to plan `27-07`, which reads this SUMMARY's verdict. The evidence above is what `27-07` needs to mark it fixed with a real `resolved_at`, not a reasoning-only close.

## Next Phase Readiness
- `WriteFileAtomic` is now the fully crash-safe, fsync-hardened primitive that plans `27-03` (duplicate), `27-04`/`27-05` (delete-to-Trash confirmation UI), and `27-06` (watcher) build their own write paths on top of, per this phase's threat register (T-27-06).
- Plan `27-07` should record two things in `.planning/WINDOWS.md`: (1) mark entry #6 fixed, citing this SUMMARY's evidence; (2) log the Windows directory-fsync platform divergence (parent-directory `Sync()` is not supported the same way on Windows; `syncDir`'s error there is logged and ignored, matching this plan's `<recorded_decision>`).
- No new Go dependency was added (`go.mod`/`go.sum` untouched), consistent with this plan's standing constraint that the two new dependencies (`wastebasket`, `fsnotify`) land in `27-03`/`27-06`.

## Self-Check: PASSED

All 4 created/modified files verified present on disk (`internal/catalog/atomicwrite.go`, `internal/catalog/atomicwrite_test.go`, `internal/catalog/atomicwrite_sigkill_test.go`, `internal/catalog/testdata/killtarget/main.go`); both commits (`6be3499f`, `8f690569`) verified present in `git log --oneline --all`.

---
*Phase: 27-catalog-actions-watch*
*Completed: 2026-08-15*
