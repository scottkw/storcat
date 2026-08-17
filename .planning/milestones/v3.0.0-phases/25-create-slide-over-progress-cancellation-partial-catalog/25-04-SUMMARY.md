---
phase: 25-create-slide-over-progress-cancellation-partial-catalog
plan: 04
subsystem: catalog
tags: [go, wails, stdlib-syscall, volume-enumeration, progress-denominator, windows-ledger]

# Dependency graph
requires:
  - phase: 25-01
    provides: "App.StartScan/startScan split, ScanOptions, throttledProgress, the sourceTotalBytes seam this plan fills"
  - phase: 25-03
    provides: "App-held cancel handle and retained partial-scan tree, WritePartialCatalog, beforeClose's CRT-13 branch"
provides:
  - "internal/volumes: Volume{Name, MountPath, TotalBytes, FreeBytes, Readable}, List(), per-OS mountPoints()/diskUsage() for darwin/linux/windows, all stdlib-only (no golang.org/x/sys)"
  - "internal/catalog.MeasureTree(ctx, root, opts, onProgress) -- count-only pre-pass mirroring traverseDirectory's hidden-file rule and single-entry tolerance"
  - "App.ListVolumes() Wails binding; App.resolveScanTotal fills the sourceTotalBytes seam (hint-or-measure)"
  - "ScanOptions.TotalBytesHint; frontend wailsAPI.listVolumes; regenerated Volume/ScanOptions.totalBytesHint bindings"
  - "WINDOWS.md entries 4 (Windows disk-space/drive-letter path) and 5 (Linux enumeration heuristic, first non-Windows entry)"
affects: [25-05, 25-06, 25-07]

# Actuals (#2632)
actuals:
  tokens: 10775
  tasks: 3
  commits: 3

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Per-OS build-tagged files (internal/volumes) reading stdlib syscall.Statfs on darwin/linux and stdlib syscall.NewLazyDLL/LazyProc against kernel32.dll on windows -- no golang.org/x/sys import anywhere, a deliberate divergence from 25-RESEARCH.md's/25-PATTERNS.md's golang.org/x/sys recommendation, mandated by this plan's phase-specific stdlib-only constraint"
    - "Portable orchestration file (volumes.go, no build tag) calling per-OS primitives (mountPoints()/diskUsage()) defined in build-tagged sibling files within the same package -- the compiler only ever links the one implementation matching GOOS"
    - "buildVolumeList(candidates []string) factored out of List() so the zero-candidates/one-bad-candidate contract is unit-testable without depending on this machine's real mount namespace"
    - "resolveScanTotal(ctx, sourcePath, opts, hint, testHook) factored out of startScan so the hint-vs-pre-pass branch is directly unit-testable, matching 25-03's established pattern of extracting a core function around a.throttledProgress's a.ctx==nil headless-testability guard"

key-files:
  created:
    - internal/volumes/volumes.go
    - internal/volumes/volumes_darwin.go
    - internal/volumes/volumes_linux.go
    - internal/volumes/volumes_windows.go
    - internal/volumes/volumes_test.go
    - internal/volumes/volumes_darwin_test.go
    - internal/catalog/measure.go
    - internal/catalog/measure_test.go
  modified:
    - app.go
    - app_test.go
    - frontend/src/services/wailsAPI.ts
    - frontend/wailsjs/go/main/App.d.ts
    - frontend/wailsjs/go/main/App.js
    - frontend/wailsjs/go/models.ts
    - .planning/WINDOWS.md

key-decisions:
  - "internal/volumes uses stdlib syscall (Statfs on darwin/linux, NewLazyDLL/LazyProc against kernel32.dll on windows) instead of golang.org/x/sys, overriding 25-RESEARCH.md's and 25-PATTERNS.md's explicit recommendation to promote golang.org/x/sys from indirect to direct -- this plan's phase-specific operating instructions stated the promotion is NOT permitted and required go.mod/go.sum to stay byte-identical. go.mod/go.sum are untouched; Task 1's own acceptance criterion asking for the '// indirect' comment to disappear is superseded by this instruction and intentionally not satisfied."
  - "kernel32.dll is loaded via the stdlib syscall package's own NewLazyDLL/LazyProc (not golang.org/x/sys/windows's more hardened LazyDLL), acceptable specifically because kernel32 is one of Windows' always-protected 'known DLLs', not an arbitrary DLL name"
  - "MeasureTree deliberately has no terminal-vs-single-entry classification (unlike traverseDirectory/walkState.classify) -- it tolerates every stat/read failure via readErrors++, matching only the ordinary skip-and-continue path; a pre-pass's job is a fast count, not source-loss detection"
  - "resolveScanTotal and buildVolumeList extracted as directly-testable cores, not tested through the full StartScan/List() paths, for the same headless-testability reasons 25-03's SUMMARY documents for every EventsEmit-dependent behavior in this file"
  - "TestStartScan_RetainsPartialOnSourceLoss (25-03) updated to supply a non-zero TotalBytesHint so its mid-walk removal timing stays keyed to the real walk, not the new pre-pass this plan added ahead of it -- without the hint, MeasureTree's own pass over the same tree would trigger the removal first and the test's SourceUnavailableError assertion would fail for the wrong reason"
  - "Ledger IDs landed as #4 (Windows) and #5 (Linux), not the plan's anticipated #3/#4 -- entry #3 was already taken by plan 25-03's CRT-13 gap, per this plan's own phase-specific operating instructions. Task 3's acceptance criterion literally asking for open_count/total_count == 4 is correspondingly 5/5, not 4/4."

patterns-established:
  - "A package needing a real OS-specific syscall (not just an argv-shape difference) uses go:build-tagged per-OS files with a shared portable orchestration file, reusable for any future OS-integration need beyond internal/osutil/reveal.go's argv-dispatch precedent"

requirements-completed: [CRT-02, CRT-03, CRT-07]

coverage:
  - id: D1
    description: "Mounted volumes enumerate with name, mount path, size, free space and a readable flag on macOS, stdlib-only, with no new Go module"
    requirement: CRT-02
    verification:
      - kind: unit
        ref: "internal/volumes/volumes_test.go#TestSkipMountEntry"
        status: pass
      - kind: unit
        ref: "internal/volumes/volumes_test.go#TestVolumeNameFromMountPath"
        status: pass
      - kind: unit
        ref: "internal/volumes/volumes_test.go#TestList_ReturnsEmptySliceNotErrorWhenNoMounts"
        status: pass
      - kind: unit
        ref: "internal/volumes/volumes_test.go#TestProbeReadable_UnreadableDirectory"
        status: pass
      - kind: integration
        ref: "internal/volumes/volumes_darwin_test.go#TestList_ExcludesBootVolumeSymlink (live /Volumes on this machine)"
        status: pass
    human_judgment: false
  - id: D2
    description: "The boot volume and vendor-reserved /Volumes entries never appear as volume cards, verified against this machine's actual, live mount list"
    requirement: CRT-02
    verification:
      - kind: integration
        ref: "internal/volumes/volumes_darwin_test.go#TestList_ExcludesBootVolumeSymlink; manually cross-checked List() output (pi-downloader, software) against `ls -la /Volumes` this session"
        status: pass
    human_judgment: false
  - id: D3
    description: "Windows and Linux volume-enumeration paths compile under their own build tags (compile-verified only, honestly not claimed as runtime-verified)"
    verification:
      - kind: other
        ref: "GOOS=windows GOARCH=amd64 go build ./internal/volumes/; GOOS=linux GOARCH=amd64 go build ./internal/volumes/"
        status: pass
    human_judgment: false
  - id: D4
    description: "A count-only pre-pass (MeasureTree) produces a real denominator for the plain-folder case, applying the identical hidden-file and unreadable-entry rules the real walk applies, honouring cancellation"
    requirement: CRT-07
    verification:
      - kind: unit
        ref: "internal/catalog/measure_test.go#TestMeasureTree_CountsFilesAndBytes"
        status: pass
      - kind: unit
        ref: "internal/catalog/measure_test.go#TestMeasureTree_RespectsIncludeHidden"
        status: pass
      - kind: unit
        ref: "internal/catalog/measure_test.go#TestMeasureTree_HonoursCancellation"
        status: pass
      - kind: unit
        ref: "internal/catalog/measure_test.go#TestMeasureTree_TolerantOfUnreadableEntries"
        status: pass
    human_judgment: false
  - id: D5
    description: "A volume source uses the picker's already-known total (no pre-pass); a zero hint runs the pre-pass and uses its measured total -- the scan denominator is always real, never invented"
    requirement: CRT-07
    verification:
      - kind: unit
        ref: "app_test.go#TestStartScan_UsesTotalHintWhenSupplied"
        status: pass
      - kind: unit
        ref: "app_test.go#TestStartScan_MeasuresWhenNoHintSupplied"
        status: pass
    human_judgment: true
    rationale: "resolveScanTotal's branch logic is unit-tested directly and passes, but the actual wire-level delivery (ScanProgress.TotalBytes over EventsEmit, and the frontend rendering a live file count vs. a percentage) requires a running wails dev session and cannot be exercised by a headless Go test -- a.throttledProgress's a.ctx==nil guard (25-03's own documented constraint) makes that path untestable outside a real Wails runtime. A human should confirm in wails dev that starting a plain-folder scan shows a live count with no percentage until the pre-pass resolves, then a real percentage."
  - id: D6
    description: "The frontend can list mounted volumes through the standard bridge; the binding never returns a nil slice"
    requirement: CRT-02
    verification:
      - kind: unit
        ref: "app_test.go#TestListVolumes_ReturnsSliceWithoutError"
        status: pass
      - kind: other
        ref: "cd frontend && npx tsc --noEmit -- exit 0"
        status: pass
      - kind: other
        ref: "cd frontend && npm run build -- exit 0"
        status: pass
      - kind: other
        ref: "grep -c 'Volume' frontend/wailsjs/go/models.ts -- 2"
        status: pass
    human_judgment: false
  - id: D7
    description: "Two honest platform-gap ledger entries recorded for the Windows and Linux volume paths, neither claiming the path works"
    verification:
      - kind: other
        ref: "gsd-tools windows status --raw -- open_count 5, total_count 5; grep -c internal/volumes/volumes_windows.go .planning/WINDOWS.md -- 2; grep -ci linux .planning/WINDOWS.md -- 5"
        status: pass
    human_judgment: false

duration: 4min
completed: 2026-08-14
status: complete
---

# Phase 25 Plan 4: Volume Enumeration + Real Progress Denominator Summary

**stdlib-only (no golang.org/x/sys) per-OS volume enumeration in a new `internal/volumes` package, a count-only `MeasureTree` pre-pass, and the `App.ListVolumes`/`ScanOptions.TotalBytesHint` wiring that gives the create flow's progress bar a real, never-fabricated denominator.**

## Performance

- **Duration:** 4 min (active task execution; excludes the file-reading/context-gathering phase before the first commit)
- **Started:** 2026-08-14T19:45:06-05:00 (Task 1 commit)
- **Completed:** 2026-08-14T19:48:46-05:00 (Task 3 commit)
- **Tasks:** 3
- **Files modified:** 15 (8 created, 7 modified)

## Accomplishments
- `internal/volumes` enumerates mounted volumes on darwin (live-verified against this machine's real `/Volumes`: `pi-downloader` and `software` list with `Readable: false`, matching their `d--x--x--x` permissions; `Macintosh HD`'s boot-volume symlink and the hidden/vendor-reserved entries are excluded) and cross-builds cleanly for `GOOS=windows`/`GOOS=linux`
- Zero new Go dependency: every per-OS primitive (`syscall.Statfs` on darwin/linux, `syscall.NewLazyDLL`/`LazyProc.Call` against `kernel32.dll` on Windows) is stdlib -- `go.mod`/`go.sum` are byte-identical to before this plan, a deliberate divergence from 25-RESEARCH.md's/25-PATTERNS.md's `golang.org/x/sys` recommendation, per this plan's own phase-specific operating instructions
- `internal/catalog.MeasureTree` gives the plain-folder case (no volume total available) a real, cancellable, hidden-file-rule-matching file/byte count without doubling I/O for every scan
- `App.resolveScanTotal` fills the `sourceTotalBytes` seam left by plan 25-01: a volume's hint is used as-is (no pre-pass); a zero hint runs `MeasureTree` first, with its own progress carrying a zero total (the "denominator unknown" wire signal) before the real walk starts with the measured total
- `App.ListVolumes()` binding and the regenerated frontend bridge (`wailsAPI.listVolumes`, the `volumes.Volume` model, `ScanOptions.totalBytesHint`) are ready for the volume-card picker UI in a later plan
- Two honest ledger entries (`.planning/WINDOWS.md` #4 Windows, #5 Linux) record exactly what compiled and what was not run, per this plan's explicit instruction not to claim either platform works

## Task Commits

Each task was committed atomically:

1. **Task 1: The internal/volumes package with per-operating-system enumeration** - `1ac9a91f` (feat)
2. **Task 2: The ListVolumes binding and a real progress denominator** - `914c81f7` (feat)
3. **Task 3: Log the two unverifiable platform paths to the ledger** - `f004d266` (docs)

## Files Created/Modified
- `internal/volumes/volumes.go` - `Volume`, `List()`, `buildVolumeList()`, `volumeNameFromMountPath()`, `skipMountEntry()`, `probeReadable()`
- `internal/volumes/volumes_darwin.go` - `mountPoints()` over `/Volumes` excluding the boot-volume symlink, `diskUsage()` via stdlib `syscall.Statfs`
- `internal/volumes/volumes_linux.go` - `mountPoints()` over `/mnt` + `/media` (incl. per-user second level) cross-checked against `/proc/mounts`, `diskUsage()` via stdlib `syscall.Statfs`
- `internal/volumes/volumes_windows.go` - `mountPoints()` from `GetLogicalDrives`' bitmask, `diskUsage()` via `GetDiskFreeSpaceExW`, both through stdlib `syscall.NewLazyDLL`/`LazyProc`
- `internal/volumes/volumes_test.go` / `volumes_darwin_test.go` - portable table-driven tests plus the darwin-only live-mount-namespace assertion
- `internal/catalog/measure.go` - `MeasureTree(ctx, root, opts, onProgress) (files int, bytes int64, err error)`
- `internal/catalog/measure_test.go` - four behavior tests including a cross-check against `traverseDirectory`'s own counts
- `app.go` - `ScanOptions.TotalBytesHint`, `App.resolveScanTotal`, `App.ListVolumes`, `sourceTotalBytes` removed
- `app_test.go` - three new tests plus one existing test (`TestStartScan_RetainsPartialOnSourceLoss`) updated to supply a hint
- `frontend/src/services/wailsAPI.ts` - `listVolumes`
- `frontend/wailsjs/go/main/App.d.ts` / `App.js` / `models.ts` - regenerated (`wails generate module`)
- `.planning/WINDOWS.md` - entries 4 and 5

## Decisions Made
- Used stdlib `syscall` instead of `golang.org/x/sys` throughout `internal/volumes`, per this plan's phase-specific instructions -- this directly contradicts 25-RESEARCH.md's and 25-PATTERNS.md's written recommendation to promote `golang.org/x/sys` to a direct dependency, and Task 1's own acceptance criterion asking for that promotion. `go.mod`/`go.sum` are untouched, matching this plan's overriding success criterion.
- `kernel32.dll` loaded via stdlib `syscall.NewLazyDLL`, not `golang.org/x/sys/windows`'s more hardened `LazyDLL` -- acceptable specifically because kernel32 is a Windows "known DLL," always resolved from the protected System32 path.
- `MeasureTree` has no terminal-vs-single-entry classification (unlike `traverseDirectory`) -- every stat/read failure is tolerated via a `readErrors` counter, matching only the ordinary skip-and-continue behavior; classifying a vanished source is out of scope for a count-only pre-pass.
- `resolveScanTotal`/`buildVolumeList` extracted as directly-testable core functions, following 25-03's established pattern around `a.throttledProgress`'s `a.ctx == nil` headless-testability guard.
- `TestStartScan_RetainsPartialOnSourceLoss` (from plan 25-03) updated to supply `TotalBytesHint: 4096` so its mid-walk-removal timing stays keyed to the real walk rather than the new pre-pass this plan added ahead of it.
- Ledger entries landed as #4/#5 (not the plan's anticipated #3/#4) since entry #3 was already taken by plan 25-03's CRT-13 gap -- per this plan's own phase-specific operating instructions, IDs were left to auto-assign rather than hand-forced.

## Deviations from Plan

### Auto-fixed Issues

**1. [Instruction override, documented not auto-decided] stdlib `syscall` instead of `golang.org/x/sys`; `go.mod`/`go.sum` untouched**
- **Found during:** Task 1
- **Issue:** 25-RESEARCH.md, 25-PATTERNS.md, and this plan's own Task 1 action text and acceptance criteria all specify promoting `golang.org/x/sys` from an indirect to a direct dependency for `unix.Statfs`/`windows.GetDiskFreeSpaceEx`. This plan's phase-specific operating instructions explicitly state the opposite: "`golang.org/x/sys` is already an indirect dependency; if you need it directly, that is a new direct dependency and is NOT permitted -- use `syscall` from stdlib instead," and the plan's own success criteria require `git diff --stat -- go.mod go.sum` to be empty.
- **Fix:** Implemented all three per-OS files (`volumes_darwin.go`, `volumes_linux.go`, `volumes_windows.go`) using only the standard library's `syscall` package -- `syscall.Statfs`/`syscall.Statfs_t` on darwin/linux (present in stdlib, not just `golang.org/x/sys/unix`), and `syscall.NewLazyDLL("kernel32.dll")`/`LazyProc.Call` to reach `GetDiskFreeSpaceExW` and `GetLogicalDrives` on Windows (present in stdlib `syscall`, unlike `golang.org/x/sys/windows`'s already-wrapped versions).
- **Files modified:** `internal/volumes/volumes_darwin.go`, `internal/volumes/volumes_linux.go`, `internal/volumes/volumes_windows.go`
- **Verification:** `go build ./...`, `GOOS=windows/linux GOARCH=amd64 go build ./internal/volumes/`, `git diff --stat -- go.mod go.sum` empty -- all pass.
- **Committed in:** `1ac9a91f` (Task 1 commit)
- **Consequence not fixed:** Task 1's literal acceptance criterion `grep -c '// indirect' go.mod for golang.org/x/sys returns 0` is now unmet by design -- `golang.org/x/sys` remains `// indirect`, unchanged, because it is genuinely unused by this package. This is the intended outcome of the overriding instruction, not an oversight.

**2. [Rule 1 - Bug, self-caused by Task 2's own change] `TestStartScan_RetainsPartialOnSourceLoss` updated to supply a total-bytes hint**
- **Found during:** Task 2, while wiring `resolveScanTotal` into `startScan`
- **Issue:** With a zero `TotalBytesHint` (the test's original `ScanOptions{WriteHTML: true}`), the new pre-pass (`MeasureTree`) now runs over the same source tree *before* the real walk, driving the test's `testHook` during the pre-pass instead of the real walk -- the source-root removal would fire early, and the real walk's very first `os.Stat` on the (now-missing) root would return a plain error rather than a `*catalog.SourceUnavailableError`, breaking the test's `errors.As` assertion.
- **Fix:** Added `TotalBytesHint: 4096` to the test's `ScanOptions`, which skips the pre-pass entirely (per `resolveScanTotal`'s documented contract) and restores the test's original intent: the mid-walk removal happens during the real walk, exactly as plan 25-03 wrote it.
- **Files modified:** `app_test.go`
- **Verification:** `go test ./... -race` -- all tests, including this one, pass.
- **Committed in:** `914c81f7` (Task 2 commit)

**3. [Instruction override, documented not auto-decided] Ledger entry IDs land as #4/#5, not #3/#4**
- **Found during:** Task 3
- **Issue:** The plan text anticipated ledger entries #3 and #4 for the Windows and Linux gaps, but entry #3 was already recorded by plan 25-03's CRT-13 gap.
- **Fix:** Appended both entries via `gsd-tools windows append` with auto-assigned IDs (per this plan's explicit phase-specific instruction), landing as #4 (Windows) and #5 (Linux). Task 3's literal acceptance criterion ("open count is 4 and total count is 4") is correspondingly 5/5.
- **Files modified:** `.planning/WINDOWS.md`
- **Verification:** `gsd-tools windows status --raw` -- `open_count: 5, total_count: 5`.
- **Committed in:** `f004d266` (Task 3 commit)

---

**Total deviations:** 3, all instructed overrides of the plan's literal text (not independently decided) plus one self-caused test fix (Rule 1) required by an in-plan behavior change.
**Impact on plan:** No scope creep. The stdlib-vs-x/sys and ledger-ID deviations were mandated by this plan's own phase-specific operating instructions layered over the plan text; the test fix was necessary to keep an existing regression test's original scenario intact after this plan's own pre-pass addition.

## Issues Encountered

None beyond the deviations documented above. `wails generate module` ran cleanly with no running dev server required (static generation over the Go source).

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- `App.ListVolumes` and the frontend bridge (`wailsAPI.listVolumes`, `models.Volume`) are ready for a later plan's volume-card picker UI to consume directly.
- `ScanOptions.TotalBytesHint` is ready for that same UI to populate from a selected volume card's already-known `TotalBytes` -- no further backend wiring needed.
- **Open items carried forward, not silently dropped, in `.planning/WINDOWS.md`:**
  - Entry 4: the Windows disk-space/drive-letter path (`GetDiskFreeSpaceExW`, `GetLogicalDrives` via stdlib `syscall`) is compile-verified only; a Windows machine must run it before v3.0.0 ships.
  - Entry 5: the Linux `/mnt`+`/media`(+per-user)/`proc/mounts` enumeration heuristic is compile-verified and statically reasoned about only; a Linux machine must run it before v3.0.0 ships.
- **Coverage D5 needs human judgment**, not because the branch logic is unproven (it is, via `TestStartScan_UsesTotalHintWhenSupplied`/`_MeasuresWhenNoHintSupplied`), but because the actual wire-level delivery of `ScanProgress.TotalBytes` and the frontend's zero-total "counting" sub-state cannot be exercised by a headless Go test (same `a.ctx == nil` limitation 25-03's SUMMARY documents for `EventsEmit`-dependent behavior). A future session with `wails dev` running should start a plain-folder scan and confirm a live file count with no percentage appears until the pre-pass resolves, then a real percentage.
- Live verification this session (recorded, not asserted from memory): `ls -la /Volumes` on this exact machine showed `.timemachine`, `com.apple.TimeMachine.localsnapshots`, `Macintosh HD -> /`, `pi-downloader` (`d--x--x--x`), `software` (`d--x--x--x`); `internal/volumes.List()` returned exactly `pi-downloader` and `software`, both `Readable: false` -- matching 25-RESEARCH.md's Pitfall 4 observation exactly, with no drift.

---
*Phase: 25-create-slide-over-progress-cancellation-partial-catalog*
*Completed: 2026-08-14*

## Self-Check: PASSED

All 15 created/modified files verified present on disk; all three task commit hashes (`1ac9a91f`, `914c81f7`, `f004d266`) verified present in git history.
