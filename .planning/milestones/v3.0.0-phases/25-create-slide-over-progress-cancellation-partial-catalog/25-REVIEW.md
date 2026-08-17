---
phase: 25-create-slide-over-progress-cancellation-partial-catalog
reviewed: 2026-08-14T00:00:00Z
depth: standard
files_reviewed: 24
files_reviewed_list:
  - app.go
  - app_test.go
  - frontend/src/components/workspace/CatalogRail.tsx
  - frontend/src/components/workspace/CreateSlideOver.tsx
  - frontend/src/components/workspace/StatusBar.tsx
  - frontend/src/components/workspace/TreePane.tsx
  - frontend/src/components/workspace/create/CreateForm.tsx
  - frontend/src/components/workspace/create/DoneBody.tsx
  - frontend/src/components/workspace/create/ErrorBody.tsx
  - frontend/src/components/workspace/create/OptionsToggles.tsx
  - frontend/src/components/workspace/create/ScanningBody.tsx
  - frontend/src/components/workspace/create/VolumePicker.tsx
  - frontend/src/contexts/AppContext.tsx
  - frontend/src/lib/scanFormat.ts
  - frontend/src/services/wailsAPI.ts
  - frontend/src/types/scan.ts
  - internal/catalog/atomicwrite.go
  - internal/catalog/errors.go
  - internal/catalog/measure.go
  - internal/catalog/options.go
  - internal/catalog/service.go
  - internal/volumes/volumes.go
  - internal/volumes/volumes_darwin.go
  - internal/volumes/volumes_linux.go
  - internal/volumes/volumes_windows.go
  - pkg/models/catalog.go
findings:
  critical: 2
  warning: 2
  info: 0
  total: 4
status: issues_found
---

# Phase 25: Code Review Report

**Reviewed:** 2026-08-14
**Depth:** standard
**Files Reviewed:** 24 (of 31 changed paths; `frontend/wailsjs/go/**` excluded as generated)
**Status:** issues_found

## Summary

Reviewed the diff from `0268f0cd..HEAD` for phase 25 — the milestone's highest-regression-risk phase, the only one touching existing CLI-shared Go code (`internal/catalog`). I verified, directly against current source, every item in the project-specific checklist:

- `cli/create.go` is byte-unchanged (`git diff --stat` empty) and `CreateCatalog`'s wrapper signature/semantics are preserved.
- `Options{WriteHTML: true}` is explicit in the CLI wrapper; `HaltOnSourceLoss` is never set by it (zero value `false`), and this is asserted by `TestCreateCatalog_WrapperWritesHTML`/`TestCreateCatalog_WrapperDoesNotHaltOnSourceLoss`.
- `outputDir` is a genuinely distinct parameter from `sourcePath`; the CLI wrapper is the only caller that collapses them back together.
- `traverseDirectory`'s single-entry read failures keep the original bare skip-and-continue; only a scan-root re-probe failure (`classify()`) breaks the loop and propagates `*SourceUnavailableError`. Verified the root-vanish classification fix (25-07) is correctly scoped to `HaltOnSourceLoss=true` only.
- The whole tree is built in memory before any write in `CreateCatalogWithContext`; no incremental-write path exists; cancellation is re-checked with `ctx.Err()` immediately before the only call into the write path.
- `omitempty` is present on both `CatalogItem.Unreadable` and `CatalogItem.ReadError` (COMPAT-02 confirmed clean-scan-safe).
- `internal/catalog` and `internal/volumes` import no Wails package; `runtime.EventsEmit` appears only in `app.go`, confined to one closure (`throttledProgress`).
- `internal/volumes` uses stdlib `syscall` only (no `golang.org/x/sys`); `go.mod`/`go.sum` diff is empty; per-OS build tags are correct and match each OS's actual `Statfs_t`/Win32 shape.
- Progress counters are clamped monotonic (`Math.max`) in `AppContext`'s `SCAN_PROGRESS` reducer; the counting sub-state renders no progress bar and no fabricated percentage.

Two genuine defects were found in the write/cancellation surface this phase built, both inside the exact area (`WriteCatalogFrom`'s copy path and `WritePartialCatalog`'s idempotency contract) the phase's own design doc calls out as load-bearing. Neither contradicts any of the SUMMARYs' recorded live-verification evidence — both exercise scenarios that were never exercised live (a mid-copy crash on the secondary-location path; two *different* buttons/calls racing each other rather than one button double-clicked). One frontend correctness bug (stale closure) was also found in the ⌘↵ keyboard path.

## Critical Issues

### CR-01: The secondary-location copy write is not crash-safe — bypasses the atomic-write primitive this phase built specifically to prevent this

**File:** `internal/catalog/service.go:605-623` (`copyFile`), called from `internal/catalog/service.go:238` and `:247` (`WriteCatalogFrom`)

**Issue:** Every primary catalog write (`writeJSONFile`, `writeHTMLFile`) is routed through `WriteFileAtomic` (temp file in the destination directory, then `os.Rename`), per `25-CONTEXT.md`'s explicit, locked decision: *"Crash-safe writes (write to temp, then rename) are implemented in this phase, not deferred... reuses the primitive across rename/duplicate/delete."* `copyFile`, however, writes directly to the destination:

```go
func (s *Service) copyFile(src, dst string) (int64, error) {
	sourceFile, err := os.Open(src)
	...
	destFile, err := os.Create(dst)   // truncates dst immediately, in place
	...
	defer destFile.Close()
	return io.Copy(destFile, sourceFile)   // writes directly to dst, no temp+rename
}
```

`copyFile` is the only mechanism behind the "copy both files to secondary location" toggle (CRT-05) — a first-class, user-facing write path, not an internal cache file. `os.Create` truncates the destination the instant it's called, before any bytes are copied. A crash, disk-full, or I/O error partway through `io.Copy` (a real risk for exactly the flaky/removable/network-mounted secondary locations this app targets) leaves a truncated `.json`/`.html` at the destination — including **overwriting and destroying a previously good copy** if the secondary catalog already existed from an earlier run. This is precisely the failure mode `WriteFileAtomic` exists to prevent, on a path the phase's own project-specific check #6 asks reviewers to verify.

**Why this isn't contradicted by recorded live evidence:** 25-07-SUMMARY.md's D2 verification (`ls` diffing before/after a "Write partial catalog" click) exercises `WriteFileAtomic`'s primary-file behavior only; no SUMMARY records a crash-mid-copy test against the secondary-location path, and no test file (`atomicwrite_test.go`, `service_test.go`) references `copyFile` at all — `grep -n "copyFile" internal/catalog/*_test.go` returns nothing.

**Fix:** Route `copyFile` through `WriteFileAtomic` — read `src` into memory (catalog JSON/HTML files are already held in memory once per write in this same function, so this is not a new I/O pattern) and call `WriteFileAtomic(dst, data, 0644)` instead of `os.Create`+`io.Copy`, or extend `WriteFileAtomic` with a streaming variant if avoiding a full in-memory read is a goal. Example:

```go
func (s *Service) copyFile(src, dst string) (int64, error) {
	data, err := os.ReadFile(src)
	if err != nil {
		return 0, err
	}
	if err := WriteFileAtomic(dst, data, 0644); err != nil {
		return 0, err
	}
	return int64(len(data)), nil
}
```

---

### CR-02: `App.WritePartialCatalog` has a TOCTOU race that can duplicate the on-disk write and can clobber a legitimately retained new partial scan

**File:** `app.go:415-444` (`WritePartialCatalog`); enabled by `frontend/src/components/workspace/create/ErrorBody.tsx:101,108` (`Retry scan` / `Close without writing` have no `writingPartial`-based disabled guard) and `frontend/src/components/workspace/CreateSlideOver.tsx:181-246,254-293` (`handleCreate`'s guard is independent of `writingPartialRef`)

**Issue:** `WritePartialCatalog`'s idempotency is documented as: *"once a write succeeds, the cached result is returned on every later call without touching the filesystem again."* The implementation is check-then-act, not atomic under the lock:

```go
func (a *App) WritePartialCatalog() (*models.CreateCatalogResult, error) {
	a.scanMu.Lock()
	if a.lastPartialResult != nil {
		result := a.lastPartialResult
		a.scanMu.Unlock()
		return result, nil
	}
	if a.lastPartial == nil || a.lastScanReq == nil { ... }
	partial := a.lastPartial
	req := a.lastScanReq
	a.scanMu.Unlock()                      // <-- lock released before the write

	result, err := a.catalogService.WriteCatalogFrom(...)   // <-- actual I/O, unguarded
	...
	a.scanMu.Lock()
	a.lastPartialResult = result           // <-- unconditional overwrite
	a.lastPartial = nil
	a.lastScanReq = nil
	a.scanMu.Unlock()
	...
}
```

Two concrete, UI-reachable ways this breaks:

1. **Duplicate write.** If `WritePartialCatalog` is invoked twice before either call re-acquires the lock (e.g. two near-simultaneous frontend calls, or a future/alternate frontend path that calls the raw binding twice), both calls read `lastPartialResult == nil`, both proceed to `WriteCatalogFrom`, and both perform the actual filesystem write — contradicting the documented "second call never touches the filesystem" guarantee. Given CR-01, the secondary-copy half of that duplicate write is also non-atomic, so two concurrent writes to the same `dst` via `io.Copy` (`os.Create` truncates on each) can genuinely interleave and corrupt the file, not just waste work.

2. **Clobbering a new, legitimate partial scan.** `ErrorBody`'s "Retry scan" button (`onRetry`, wired to `handleCreate`) and "Close without writing" button carry **no** `disabled={writingPartial}` guard — only "Write partial catalog" itself does (`ErrorBody.tsx:88`). `handleCreate`'s own guard (`CreateSlideOver.tsx:187`, `scan.status !== 'idle' && scan.status !== 'error'`) does not check `writingPartial` either, and `scan.status` stays `'error'` for the whole duration of an in-flight "Write partial catalog" call (it only changes on the dispatched `SCAN_DONE` at the end). So a user can click "Write partial catalog" (kicking off a — per CR-01, non-atomic and possibly slow — Go-side write) and then immediately click "Retry scan" before that write finishes. `StartScan`'s own setup path (`app.go:294-299`) clears `a.lastPartial`/`a.lastScanReq` unconditionally the moment the retry begins. If the retry itself then also fails with a new source loss, it repopulates `a.lastPartial`/`a.lastScanReq` with the *new* failure's tree. When the original, still-in-flight `WritePartialCatalog` call finally finishes, it re-acquires the lock and unconditionally sets `a.lastPartialResult = result; a.lastPartial = nil; a.lastScanReq = nil` — silently destroying the new failure's legitimately retained partial tree and replacing the retained state with the stale first write's result. A subsequent "Write partial catalog" click for what the user believes is the *new* failure will short-circuit on the stale cached `lastPartialResult` and never write the new tree at all.

**Why this isn't contradicted by recorded live evidence:** 25-07-SUMMARY.md's D2 check ("two same-tick clicks on 'Write partial catalog' produced exactly one JSON+HTML pair") exercises two clicks on the *same* button, which the frontend's synchronous `writingPartialRef` guard (`CreateSlideOver.tsx:255`) correctly serializes before either call's `await` resolves — that guard is real and does work for that specific case. The race identified here is different: it requires the *other* two error-state buttons, which are unguarded, to be exercised concurrently with an in-flight partial write. No SUMMARY records this interleaving being tested, and `app_test.go` has no test exercising concurrent `WritePartialCatalog`/`StartScan` calls (`grep -n "func Test" app_test.go` shows sequential-only partial/write tests).

**Fix (two independent layers, both worth applying):**
- Backend: hold `scanMu` for the whole check-decide-write-record sequence, or use a `sync.Once`-per-retained-scan / generation-counter so a write started against generation N can detect that generation has since moved on and abort/re-check before mutating shared state:
```go
func (a *App) WritePartialCatalog() (*models.CreateCatalogResult, error) {
	a.scanMu.Lock()
	if a.lastPartialResult != nil { ... }
	if a.lastPartial == nil { ... }
	gen := a.retainedGen // bump on every StartScan/clear
	partial, req := a.lastPartial, a.lastScanReq
	a.scanMu.Unlock()

	result, err := a.catalogService.WriteCatalogFrom(...)

	a.scanMu.Lock()
	defer a.scanMu.Unlock()
	if err != nil || gen != a.retainedGen {
		return result, err // stale generation: don't clobber newer retained state
	}
	a.lastPartialResult = result
	a.lastPartial, a.lastScanReq = nil, nil
	return result, nil
}
```
- Frontend: disable "Retry scan" and "Close without writing" while `writingPartial` is true (pass the same prop already threaded to the primary button), so the race requires bypassing the UI entirely rather than being reachable from two adjacent buttons.

## Warnings

### WR-01: The ⌘↵ keyboard shortcut can start a scan with stale option values after a toggle change

**File:** `frontend/src/components/workspace/CreateSlideOver.tsx:328-340`

**Issue:** The keydown-listener effect's dependency array omits `options` and `secondaryDir`:

```tsx
useEffect(() => {
  if (!isOpen) return;
  if (scan.status !== 'idle' && scan.status !== 'error') return;
  const onKeyDown = (event: KeyboardEvent) => {
    if ((event.metaKey || event.ctrlKey) && event.key === 'Enter') {
      event.preventDefault();
      handleCreate();
    }
  };
  window.addEventListener('keydown', onKeyDown);
  return () => window.removeEventListener('keydown', onKeyDown);
  // eslint-disable-next-line react-hooks/exhaustive-deps
}, [isOpen, selectedSource, title, root, scan.status, state.catalogDir]);
```

`handleCreate` is a fresh closure every render (not memoized), and it reads `options.writeHTML`/`options.includeHidden`/`options.copyToSecondary`/`secondaryDir` directly from render-scope state (`CreateSlideOver.tsx:209-213`). Because the effect only re-subscribes when one of the *listed* deps changes, toggling any of the three create options (or changing the secondary-copy directory) without also touching `title`, `root`, or `selectedSource` leaves the registered `keydown` listener bound to the *previous* render's `handleCreate`, closing over the *old* `options`/`secondaryDir` values. The mouse path (`onClick={handleCreate}` on the Create button, `CreateSlideOver.tsx:451`) is unaffected — JSX event handlers always use the current render's closure.

**Why this matters:** the WILL WRITE preview and the toggle UI are the documented contract for "what will actually be written" (`25-UI-SPEC.md`'s CRT-04 "recomputed live... never stale"). A user who flips "Also write HTML catalog" off and immediately presses ⌘↵ (a very plausible sequence — toggle, then use the shortcut instead of reaching for the mouse) will silently start a scan using the *previous* toggle state, writing files the UI just told them would not be written (or vice versa).

**Why this isn't contradicted by recorded live evidence:** 25-05-SUMMARY.md's D6 verification ("pressed Cmd+Enter twice in immediate succession... exactly one scan ran") tests double-invocation of the same shortcut, not a toggle-then-shortcut sequence — that interleaving was never exercised.

**Fix:** Add `options` and `secondaryDir` to the effect's dependency array (removing the `eslint-disable` comment, which is currently suppressing a real, applicable warning rather than a false positive), or hold `handleCreate` in a ref updated every render so the listener always calls the latest version regardless of which deps changed:

```tsx
const handleCreateRef = useRef(handleCreate);
handleCreateRef.current = handleCreate;
// ...
const onKeyDown = (event: KeyboardEvent) => {
  if ((event.metaKey || event.ctrlKey) && event.key === 'Enter') {
    event.preventDefault();
    handleCreateRef.current();
  }
};
```

### WR-02: `ErrorBody`'s "Retry scan" and "Close without writing" are not disabled during an in-flight "Write partial catalog" write

**File:** `frontend/src/components/workspace/create/ErrorBody.tsx:101,108`

**Issue:** Documented separately from CR-02 (which covers the backend consequence) because it is also independently worth fixing as a UX/guard issue: only the primary button (`disabled={writingPartial}`, line 88) is gated on the in-flight write. A user can click "Retry scan" or "Close without writing" while a partial-catalog write is still running, producing a race with no visible feedback that anything unusual is happening (no "writing…" indicator on the other two actions, no confirmation that the in-flight write's outcome will or won't be reflected in what they see next).

**Fix:** Pass `writingPartial` down and disable/`aria-disabled` all three action buttons while `true`, not just the one that initiated the write — see the combined fix under CR-02.

## Not Fully Investigated

Two items from the coordinator's follow-up were folded into the findings above rather than left open, but are called out explicitly per the request for honesty about verification depth:

- **`WriteCatalogFrom`'s copy path vs. atomic writes** — fully investigated; this is CR-01 above. Confirmed by direct source read (`internal/catalog/service.go:605-623`, `:236-253`) and by the absence of any test referencing `copyFile`.
- **Whether a `handleWritePartial`/Retry race guard already exists** — fully investigated; no such guard exists at either layer. Confirmed by direct source read of `ErrorBody.tsx`, `CreateSlideOver.tsx`, and `app.go`'s `WritePartialCatalog`, and by the absence of any concurrency test in `app_test.go` covering this interleaving (only sequential `TestWritePartialCatalog_WritesOnce` exists). This is CR-02 above — the 25-07 plan's stated intent ("Write partial catalog" idempotent, "Retry scan" must not start a second walk over a still-winding-down one) is correct in describing the *intended* contract, but the implementation does not fully enforce it under concurrent invocation from the two different UI actions.

Everything else in the project-specific checklist (items 1, 2, 3, 4, 5, 7, 9, 10, 12, and the cosmetic-lint note) was checked directly against current source and found to hold; see the Summary above for the specific evidence for each.

---

_Reviewed: 2026-08-14_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_
