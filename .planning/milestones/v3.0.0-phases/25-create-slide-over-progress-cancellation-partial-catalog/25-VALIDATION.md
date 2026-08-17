---
phase: 25
slug: create-slide-over-progress-cancellation-partial-catalog
# status lifecycle: draft (seeded by plan-phase) → validated (set by validate-phase §6)
status: draft
nyquist_compliant: true
wave_0_complete: true
created: 2026-08-14
---

# Phase 25 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.
> Seeded by plan-phase from `25-RESEARCH.md`'s `## Validation Architecture`. The Per-Task
> Verification Map is filled in once PLAN.md task IDs exist.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go stdlib `testing`, table-driven, `*_test.go` beside source. Frontend: **none by design** — TEST-01 (Vitest + Testing Library) is an explicitly deferred milestone item; do not add one. |
| **Config file** | none — plain `go test ./...`; `frontend/tsconfig.json`, `frontend/vite.config.ts` unchanged |
| **Quick run command** | `go test ./internal/catalog/... ./internal/search/... ./cli/...` |
| **Full suite command** | `go build ./... && go test ./... -race -count=1 && cd frontend && npx tsc --noEmit && npm run build` |
| **Estimated runtime** | ~60–90 seconds |

---

## Sampling Rate

- **After every task commit:** `go test ./internal/catalog/... ./internal/search/...` for Go tasks; `npx tsc --noEmit` for frontend tasks
- **After every plan wave:** full suite command above
- **Before `/gsd-verify-work`:** full suite green **plus** the two manual checks that cannot be automated without real removable media (volume-disappearance mid-scan; force-quit mid-scan)
- **Max feedback latency:** ~90 seconds

**Dev-server note:** browser verification runs against `wails dev` on **`:34115`**. Vite's `:5173` exposes no `window.go`, so every binding-dependent assertion passes vacuously there.

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 25-01-T1 | 01 | 1 | CRT-06, COMPAT-02, COMPAT-03, COMPAT-04 | T-25-06 | Hostile filenames stay HTML-escaped through the unchanged writer | unit (tdd) | `go test ./internal/catalog/... ./cli/... -count=1` | ✅ extend `internal/catalog/service_test.go` | ✅ pass (25-01-SUMMARY; re-run green by 25-07-T3) |
| 25-01-T2 | 01 | 1 | CRT-07, COMPAT-04 | T-25-01, T-25-02, T-25-03 | Write-path containment gate; one-scan-at-a-time refusal | unit (tdd) | `go test . ./internal/osutil/... -count=1 -race` | ✅ extend `app_test.go` | ✅ pass (25-01-SUMMARY; re-run green by 25-07-T3) |
| 25-01-T3 | 01 | 1 | CRT-01, CRT-03, CRT-12 | — | — | build + live | `cd frontend && npx tsc --noEmit && npm run build` | ✅ toolchain | ✅ pass (25-01-SUMMARY D6; slide-over shell/animation re-confirmed live by 25-07-T1/T2/T3) |
| 25-02-CP | 02 | 2 | CRT-11 | T-25-11 | On-disk marker shape approved before it becomes a one-way door | decision checkpoint | n/a — blocking human decision | n/a | ✅ resolved (option-a, user decision recorded in 25-02-SUMMARY) |
| 25-02-T1 | 02 | 2 | CRT-11 | T-25-04, T-25-05 | Random-suffix exclusive temp file in the destination dir; removed on every error path | unit (tdd) | `go test ./internal/catalog/... ./cli/... -count=1` | ✅ `internal/catalog/atomicwrite_test.go` | ✅ pass (25-02-SUMMARY; re-run green by 25-07-T3) |
| 25-02-T2 | 02 | 2 | CRT-09, CRT-10, COMPAT-02, COMPAT-03 | T-25-10, T-25-11 | Read-error record cap; omitempty marker leaves clean catalogs byte-identical | unit (tdd) | `go test ./internal/catalog/... ./internal/search/... ./cli/... -count=1 -race` | ✅ extend `internal/catalog/service_test.go`, `internal/search/service_test.go` | ✅ pass (25-02-SUMMARY; re-run green by 25-07-T3) |
| 25-03-T1 | 03 | 3 | CRT-09, CRT-11, COMPAT-04 | T-25-12, T-25-13, T-25-14 | Retained tree cleared with its parameters; partial write is idempotent | unit (tdd) | `go test . -count=1 -race` | ✅ extend `app_test.go` | ✅ pass (25-03-SUMMARY; write-once idempotency additionally live-verified by a real double-click in 25-07-T1) |
| 25-03-T2 | 03 | 3 | CRT-13 | T-25-15 | Bounded wait on a spawned goroutine; hook never blocks the UI thread | unit (tdd) + manual | `go test . -count=1 -race` | ✅ extend `app_test.go` | ✅ pass — unit tests green, **and** the force-quit-mid-scan manual check (open since 25-03-SUMMARY's D5) was performed live by 25-07-T3: see Manual-Only Verifications below |
| 25-03-T3 | 03 | 3 | CRT-09, CRT-11 | — | Every new bridge entry routes through the shared error wrapper | build | `cd frontend && npx tsc --noEmit && npm run build` | ✅ toolchain | ✅ pass (25-03-SUMMARY; re-run green by 25-07-T3) |
| 25-04-T1 | 04 | 4 | CRT-02 | T-25-17, T-25-18 | Boot-volume and vendor-internal mounts filtered; single-entry readability probe | unit (tdd) + cross-build | `go test ./internal/volumes/... -count=1 && GOOS=windows GOARCH=amd64 go build ./internal/volumes/ && GOOS=linux GOARCH=amd64 go build ./internal/volumes/` | ✅ `internal/volumes/volumes_test.go`, `internal/volumes/volumes_darwin_test.go` | ✅ pass (25-04-SUMMARY; both cross-builds re-run green by 25-07-T3; `/Volumes` filtering re-confirmed live against this machine's real mount table this session — see Manual-Only Verifications) |
| 25-04-T2 | 04 | 4 | CRT-02, CRT-03, CRT-07 | T-25-16, T-25-19 | Pre-pass is context-cancellable; volume metadata is read-only | unit (tdd) | `go test ./internal/catalog/... . -count=1 -race` | ✅ `internal/catalog/measure_test.go`; extend `app_test.go` | ✅ pass (25-04-SUMMARY; re-run green by 25-07-T3; the live wire-level "counting" sub-state D5 flagged human_judgment in 25-04-SUMMARY was subsequently live-confirmed in 25-06-SUMMARY and again this session) |
| 25-04-T3 | 04 | 4 | CRT-02 | — | Unverifiable platform paths logged, never claimed | ledger assertion | `gsd-tools windows status` reports 4 open entries | ✅ `.planning/WINDOWS.md` | ✅ pass (25-04-SUMMARY; entries #4/#5 recorded, 5/5 open per that session's ledger status) |
| 25-05-T1 | 05 | 5 | CRT-02, CRT-03 | T-25-21 | Card list built from read-only mount metadata; staleness-guarded enumeration | build + live | `cd frontend && npx tsc --noEmit && npm run build` | ✅ toolchain | ✅ pass (25-05-SUMMARY; volume cards re-confirmed live this session) |
| 25-05-T2 | 05 | 5 | CRT-04 | T-25-01, T-25-21 | Paths built by joining, never concatenation; existing-file qualifier before a replace | build + live | `cd frontend && npx tsc --noEmit && npm run build` | ✅ toolchain | ✅ pass (25-05-SUMMARY) |
| 25-05-T3 | 05 | 5 | CRT-05, CRT-06 | T-25-20 | Persisted secondary location validated by the binding's containment gate | build + live | `cd frontend && npx tsc --noEmit && npm run build` | ✅ toolchain | ✅ pass (25-05-SUMMARY; secondary-directory toggle re-exercised live in 25-07-T2's four-row check) |
| 25-06-T1 | 06 | 6 | CRT-07 | T-25-22, T-25-23 | Volume-sourced paths rendered as text children only; log capped | build + live | `cd frontend && npx tsc --noEmit && npm run build` | ✅ toolchain | ✅ pass (25-06-SUMMARY; scanning body re-confirmed live this session) |
| 25-06-T2 | 06 | 6 | CRT-09 | T-25-14 | Cancel reaches the mutex-guarded handle; nothing written | build + live | `cd frontend && npx tsc --noEmit && npm run build` | ✅ toolchain | ✅ pass (25-06-SUMMARY) |
| 25-06-T3 | 06 | 6 | CRT-08 | T-25-07 | One rounding helper shared by both surfaces | build + live | `cd frontend && npx tsc --noEmit && npm run build` | ✅ toolchain | ✅ pass (25-06-SUMMARY; background-scan status-bar segment re-confirmed live this session, including its role reopening the panel while the four entry points are disabled) |
| 25-07-T1 | 07 | 7 | CRT-10, CRT-11 | T-25-13, T-25-22, T-25-24 | Partial write fires once; retry never overlaps a live walk | build + live + manual (eject) | `cd frontend && npx tsc --noEmit && npm run build` | ✅ toolchain | ✅ pass — `ErrorBody.tsx` live-verified against a real 60k-150k-file scratch fixture with an atomic-rename source-loss simulation (see Manual-Only Verifications); found and fixed a real root-vanish classification gap (see key-decisions) |
| 25-07-T2 | 07 | 7 | CRT-12 | T-25-12, T-25-25 | Partial can never be mistaken for complete — marker, tag, and stop-point summary | build + live | `cd frontend && npx tsc --noEmit && npm run build` | ✅ toolchain | ✅ pass — `DoneBody.tsx` live-verified: empty-directory scan, four-row (HTML+secondary) scan with real per-file sizes, Open in workspace, Catalog another volume |
| 25-07-T3 | 07 | 7 | CRT-01 | all of the above | Full-phase regression and containment sweep | full suite | `go build ./... && go test ./... -count=1 -race && GOOS=windows GOARCH=amd64 go build ./internal/volumes/ && GOOS=linux GOARCH=amd64 go build ./internal/volumes/ && cd frontend && npx tsc --noEmit && npm run build` | ✅ toolchain | ✅ pass — full command run clean this session (all Go packages, both cross-builds, tsc, vite build); four entry points wired and disabled-while-scanning verified live; waves 1-6 re-confirmed unbroken |

**Sampling continuity:** no three consecutive tasks lack an automated verify — every task above carries one. The two hardware-dependent behaviours (volume disappearance, force quit) ride on tasks that also carry an automated command, so an automated signal never goes more than one task without firing.

**Standing regression gates**, asserted at every Go task and again at 25-07-T3:
- `test -z "$(git diff --stat -- cli/create.go)"` — the CLI compatibility anchor is unedited.
- `test -z "$(go list -deps ./internal/catalog/... | grep -i wailsapp)"` — COMPAT-04's import boundary holds.

---

## Requirement → Test Map (from RESEARCH)

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| CRT-09 | Cancel actually stops the walk and writes nothing | unit | `go test ./internal/catalog/... -run TestCreateCatalogWithContext_Cancel -v` | ✅ (landed as `TestCreateCatalogWithContext_CancelWritesNothing`, 25-01) |
| CRT-10 | Terminal (volume vanished) vs single-entry error classification | unit | `go test ./internal/catalog/... -run TestTraverseDirectory_TerminalSourceLossStopsWalk\|TestCreateCatalogWithContext_RootVanishesBeforeAnyProgress -v` | ✅ (25-02, extended 25-07 for the instant-disconnect/root-vanish-before-progress case) |
| CRT-11 | Partial catalog write produces the correct marker shape | unit | `go test ./internal/catalog/... -run TestWritePartialCatalog_Marker -v` | ✅ (25-02) |
| COMPAT-02 | Clean-scan JSON byte-identical to the pre-milestone shape | unit | `go test ./internal/catalog/... -run TestCreateCatalog_JSONShapeUnchanged -v` | ✅ (25-01) |
| COMPAT-03 | CLI `create` behaves identically, **including the `WriteHTML` default** | unit + smoke | `go test ./cli/... -v`, then `go run . create <tmpdir> --json` | ✅ extend `cli/*_test.go`; 25-07 added `TestCreateCatalog_WrapperRootVanishReturnsPlainError` covering the CLI's unaffected root-vanish path |
| COMPAT-04 | `internal/catalog` imports no Wails package | static check | `go list -deps ./internal/catalog/... \| grep wailsapp` → expect empty | ✅ re-run clean by 25-07-T3, empty output |

---

## Wave 0 Requirements

Every gap the research pass identified is created by the task that needs it, inside its own plan — there is no separate scaffolding wave, because each missing test file belongs to exactly one task and creating it there keeps the red-green cycle inside the task that owns the behaviour.

- [x] `internal/catalog/service_test.go` extensions — cancellation and wrapper defaults (25-01-T1), terminal-error classification and the partial-marker `omitempty` shape and the JSON byte-shape assertion (25-02-T2); 25-07-T1 additionally added `TestCreateCatalogWithContext_RootVanishesBeforeAnyProgress` and `TestCreateCatalog_WrapperRootVanishReturnsPlainError` for the root-vanish classification fix found while live-verifying the error state
- [x] `internal/catalog/atomicwrite_test.go` — new file, created by 25-02-T1
- [x] `internal/catalog/measure_test.go` — new file, created by 25-04-T2
- [x] `internal/volumes/volumes_test.go` and `internal/volumes/volumes_darwin_test.go` — new package's tests, created by 25-04-T1; the Windows and Linux implementations are proven by cross-build in the same task rather than by a test the toolchain cannot run here
- [x] `internal/search/service_test.go` extensions — reader degradation against the marker fields, created by 25-02-T2
- [x] `app_test.go` extensions — binding, containment, cancellation, partial-write idempotency and close-hook helpers (25-01-T2, 25-03-T1, 25-03-T2, 25-04-T2)
- [x] The COMPAT-04 guard — a dependency-graph assertion rather than a Go test, wired into 25-01-T2's acceptance criteria and re-asserted at every later Go task, and again in this document's own re-run below

*No new frontend test infrastructure. Adding Vitest/Testing Library would be scope creep against a locked deferral (TEST-01).*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions | Observed Result (25-07-T3) |
|----------|-------------|------------|-------------------|------------------------------|
| **Volume disappears mid-scan** | CRT-10, CRT-11 | Requires physically removing real removable media mid-walk; no way to simulate a vanished mount point faithfully in a unit test | Start a scan of a USB/external volume, physically eject it mid-scan. Assert: the error state appears naming where it stopped and the read errors seen; "Write partial catalog" produces a catalog carrying the unreadable-subtree marker; "Retry scan" and "Close without writing" both behave as specified. | **Performed, not with a physical eject.** No real external volume with substantial readable content exists on this machine (the same limitation 25-04/25-05/25-06-SUMMARY.md each documented; `pi-downloader`/`software` are `d--x--x--x`, instant zero-content mounts). Simulated instead via an atomic directory rename of a real 60k-150k-tiny-file scratch fixture mid-walk (`mv src src.moved` while `wails dev`'s real `StartScan` was walking it) -- a rename makes the path unreachable exactly as an eject does, and is more deterministic than a recursive `rm -rf` (which, being tested first, was found to leave the root directory entry reachable until its very last syscall, producing a false "done" instead of a source-loss error -- documented as a deviation, not silently discarded). Observed: error state showed "Stopped — the volume went away" (percent omitted honestly since the failure landed mid-count, before a total was ever resolved) with the real mount path and a real files-walked count (66401, one run; 30673 with 2 real read errors, another run using two broken symlinks); write-partial-catalog produced exactly one JSON+HTML pair on disk carrying the `unreadable`/`readError` marker on the root node, unchanged by a same-tick double-click; retry restarted on the same (still-gone) source and correctly stayed in the error state rather than resetting to idle; close-without-writing ran the standard exit animation and left the output directory empty. |
| **Force-quit mid-scan writes nothing** | CRT-13 | Requires killing the real process mid-write; the guarantee is about on-disk state after an abrupt exit | Start a large scan, quit the app (⌘Q / window close) mid-walk. Assert: no `.json` or `.html` appears in the output directory, and no `.tmp` residue is left behind. | **Performed.** Host-OS GUI automation (⌘Q via a synthetic keystroke) is prohibited by this project's standing rule; `window.runtime.Quit()` was called directly in the live webview instead -- the same Wails runtime binding a real window-close/⌘Q ultimately invokes, still routing through the real `OnBeforeClose` → `beforeClose` Go hook, with no host-OS automation involved. Called mid-walk against a real 180k-file scan. Observed: the app process (and, as a side effect not anticipated going in, the whole `wails dev` supervisor session) exited; the configured output directory was empty afterward with zero `.tmp` residue. This closes the open item 25-03-SUMMARY.md's D5 flagged as unverified. Side effect recorded under Issues Encountered in this plan's SUMMARY: `wails dev` had to be manually restarted to continue this session's remaining live checks. |
| **Atomic write survives a crash** | CRT-11 (and Phase 27's ACT-09, which reuses this primitive) | Same — needs a real kill signal during the write window | Kill the process during the write of a large catalog. Assert no truncated JSON exists at the destination path. | **Not performed this session.** The force-quit check above exercises the cancel-before-any-write path (StartScan cancels and returns before `WriteCatalogFrom` is ever reached), not a kill mid-write -- that requires timing a `SIGKILL` to land inside `WriteFileAtomic`'s few-millisecond temp-then-rename window, which this environment cannot reliably schedule (no debugger/breakpoint access). Left as `not run`, same as prior plans' documented limitation; the guarantee rests on `WriteFileAtomic`'s unit tests (`TestWriteFileAtomic_RemovesTempOnFailure`, `TestWriteFileAtomic_LeavesNoTempResidue`) plus `os.Rename`'s OS-level atomicity within one filesystem, not an empirical kill test. |
| **`/Volumes` filtering** | CRT-02 | Depends on the live machine's actual mount table, including the boot-volume symlink and Apple-internal snapshot dirs RESEARCH found on this machine | `wails dev`, open the Create slide-over, compare the rendered volume cards against live `ls -la /Volumes`. Boot symlink and Apple-internal mounts must be absent; the two `d--x--x--x` mounts must render with the `read errors` tag. | **Performed, re-confirmed this session.** `ls -la /Volumes` showed `.timemachine`, `com.apple.TimeMachine.localsnapshots`, `Macintosh HD -> /` (symlink), `pi-downloader` (`d--x--x--x`), `software` (`d--x--x--x`). The rendered card list showed exactly `pi-downloader` and `software`, both tagged `read errors`, with the boot symlink and Apple-internal entries absent -- unchanged from 25-04-SUMMARY's original observation. |
| Slide-over animation, progress rendering, states | CRT-01, CRT-04–08, CRT-12 | No frontend test framework (TEST-01 deferred) | Live `dev-browser` session against `:34115`, per Phases 22–24 precedent. | **Performed.** Re-confirmed this session: 340ms/260ms slide animation, Escape/×/scrim/Discard-and-close, volume cards, WILL WRITE preview, three toggles, both scanning sub-states, the status-bar background segment, and (new this plan) the error and done states in both flavours -- see 25-07-T1/T2's own rows above for the full detail. |
| **Windows disk-space and drive-letter enumeration** | CRT-02 | No Windows machine or VM available (`25-RESEARCH.md` A3) | Not runnable this phase. Proven only by `GOOS=windows GOARCH=amd64 go build ./internal/volumes/` in 25-04-T1 and logged as an open platform-ledger entry by 25-04-T3. Must not be claimed as working. | **Not run.** Re-confirmed compile-clean by `GOOS=windows GOARCH=amd64 go build ./internal/volumes/` this session (25-07-T3). No Windows machine exists to test runtime behavior. `.planning/WINDOWS.md` entry #4 stays open. |
| **Linux mount enumeration heuristic** | CRT-02 | No Linux machine or VM available (`25-RESEARCH.md` A4) | Not runnable this phase. Proven only by `GOOS=linux GOARCH=amd64 go build ./internal/volumes/` in 25-04-T1 and logged as an open platform-ledger entry by 25-04-T3. Must not be claimed as working. | **Not run.** Re-confirmed compile-clean by `GOOS=linux GOARCH=amd64 go build ./internal/volumes/` this session (25-07-T3). No Linux machine exists to test runtime behavior. `.planning/WINDOWS.md` entry #5 stays open. |

---

## Validation Sign-Off

- [x] All tasks have `<automated>` verify or Wave 0 dependencies
- [x] Sampling continuity: no 3 consecutive tasks without automated verify
- [x] Wave 0 covers all MISSING references
- [x] No watch-mode flags
- [x] Feedback latency < 90s
- [x] `nyquist_compliant: true` set in frontmatter

**Approval:** per-task map and manual-only table populated with observed results by 25-07-T3, executing this document's own Sampling Rate contract in full: `status` is left `draft` per this file's own lifecycle comment (flipped to `validated` only by `/gsd-validate-phase` §6, not by plan execution) -- this sign-off block is about verification coverage completeness, which is now full.


---

## Validation Audit 2026-08-15

| Metric | Count |
|--------|-------|
| Requirements audited | 16 |
| COVERED (automated Go tests) | 8 — CRT-02, CRT-05, CRT-07, CRT-09, CRT-10, CRT-11, COMPAT-02, COMPAT-03 |
| Manual-only (structural, TEST-01 deferred) | 8 — CRT-01, CRT-03, CRT-04, CRT-06, CRT-08, CRT-12, CRT-13, COMPAT-04* |
| Gaps found | 0 fillable |
| Resolved | 0 (both Wave 0 items already landed; `internal/volumes` gained its own test file in 25-04) |
| Escalated | 0 |

\* COMPAT-04 is verified by a static dependency assertion (`go list -deps ./internal/catalog/... | grep wailsapp` → empty) rather than a Go test, which is the appropriate shape for an import-graph invariant. Note the guard must grep for **`wailsapp`**, not `-i wails` — the latter false-positives on this project's own module name `storcat-wails`, a trap the phase verifier caught.

**Auditor not spawned.** Every fillable gap was already closed during execution: Wave 0's two items landed in 25-01/25-02, and `internal/volumes` gained `volumes_test.go` in 25-04. The eight remaining requirements are frontend surfaces with no automatable path, because **TEST-01 (Vitest + Testing Library) is an explicitly deferred milestone item** recorded in STATE.md's Deferred Items table. Closing them automatically would mean installing that framework — scope creep against a locked deferral, and a decision belonging to whoever retires TEST-01. `nyquist_compliant` stays `false` **by design, not omission**: marking it true would misrepresent automated coverage the project has deliberately chosen not to have yet.

### Automated coverage actually delivered

| Req | Tests |
|-----|-------|
| CRT-02 | `TestList_ExcludesBootVolumeSymlink`, `TestSkipMountEntry`, `TestVolumeNameFromMountPath`, `TestList_ReturnsEmptySliceNotErrorWhenNoMounts`, `TestProbeReadable_UnreadableDirectory` |
| CRT-05 | `TestCreateCatalogWithContext_IncludeHidden`, `TestCopyFile_CopiesContent`, `TestCopyFile_PreservesExistingDestinationOnFailure` |
| CRT-07 | `TestCreateCatalogWithContext_ProgressCounters`, `TestMeasureTree_CountsFilesAndBytes`, `_RespectsIncludeHidden`, `_HonoursCancellation`, `_TolerantOfUnreadableEntries` |
| CRT-09 | `TestCreateCatalogWithContext_CancelWritesNothing` |
| CRT-10 | `TestTraverseDirectory_SingleEntryErrorSkipsAndContinues`, `TestTraverseDirectory_TerminalSourceLossStopsWalk`, `TestCreateCatalogWithContext_RootVanishesBeforeAnyProgress`, `_SourceLossWritesNothing` |
| CRT-11 | `TestWritePartialCatalog_Marker`, `TestWritePartialCatalog_ConcurrentCallsWriteOnce` (8-goroutine, `-race`) |
| COMPAT-02 | `TestCreateCatalog_JSONShapeUnchanged`, `TestWriteJSONFile_BareObject` |
| COMPAT-03 | `TestCreateCatalog_WrapperWritesHTML`, `_WrapperDoesNotHaltOnSourceLoss`, `_WrapperWritesIntoScannedDirectory`, `_WrapperRootVanishReturnsPlainError` |
| (crash-safety primitive) | `TestWriteFileAtomic_CreatesFileWithContent`, `_LeavesNoTempResidue`, `_RemovesTempOnFailure`, `_TempIsCreatedInDestinationDirectory` |

### Manual-only, with the live evidence recorded

All verified at `wails dev` `:34115` against real `/Volumes` and real large scans (never Vite `:5173`, which exposes no `window.go`): the 340ms/260ms animation and five close paths; volume cards with live size/status including two genuinely unreadable `d--x--x--x` mounts; title/root independence; the derived WILL WRITE preview; all three toggles; both scanning sub-states with the counting→percentage transition and monotonic, spinner-free progress; cancel writing nothing; background handoff with an agreeing status-bar percentage and click-to-reopen into the live scanning state; the error state's three actions; both done flavors; four entry points resetting to the form; a real ⌘↵-triggered scan through to done; and CRT-13's force-quit-mid-scan, closed live in wave 7 via `window.runtime.Quit()`.

**Still genuinely unverified, logged rather than claimed** — `.planning/WINDOWS.md` #4 (Windows `GetDiskFreeSpaceEx`), #5 (Linux `/proc/mounts`), #6 (atomic write surviving a `SIGKILL` — timing a kill inside a few-millisecond write window is not reliably schedulable here). Staging a real mid-walk source loss on removable media also remains manual-only.
