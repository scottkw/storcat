---
phase: 25-create-slide-over-progress-cancellation-partial-catalog
fixed_at: 2026-08-14T22:35:00Z
review_path: .planning/phases/25-create-slide-over-progress-cancellation-partial-catalog/25-REVIEW.md
iteration: 1
findings_in_scope: 4
fixed: 4
skipped: 0
status: all_fixed
---

# Phase 25: Code Review Fix Report

**Fixed at:** 2026-08-14
**Source review:** `.planning/phases/25-create-slide-over-progress-cancellation-partial-catalog/25-REVIEW.md`
**Iteration:** 1

**Summary:**
- Findings in scope: 4 (CR-01, CR-02, WR-01, WR-02)
- Fixed: 4
- Skipped: 0
- Status: `all_fixed` — all four findings were fixed and verified by build/vet/test. One sub-part of CR-02 (the generation-counter/clobber-avoidance concurrency guarantee) is additionally flagged **"fixed: requires human verification"** per the verification strategy's logic-bug carve-out — see that finding's entry below for why.

**Note on run mode:** `workflow.use_worktrees` is `false` in `.planning/config.json`, so this run edited and committed directly on `main` in the primary checkout (no worktree, no temp branch) per the documented opt-out.

## Fixed Issues

### CR-01: `copyFile` bypassed `WriteFileAtomic` — not crash-safe

**Files modified:** `internal/catalog/service.go`, `internal/catalog/service_test.go`
**Commit:** `3ebacfdc`
**Applied fix:** Rewrote `copyFile` to read `src` fully into memory and call `WriteFileAtomic(dst, data, 0644)` (temp file in `dst`'s own directory, then `os.Rename`) instead of `os.Create`+`io.Copy`, exactly as the review's suggested fix specified. Removed the now-unused `"io"` import (the file's only other `io.*` use was inside `copyFile`). Added two new tests: `TestCopyFile_CopiesContent` (functional correctness — byte-identical content, correct returned byte count) and `TestCopyFile_PreservesExistingDestinationOnFailure`, the direct CR-01 regression test — it seeds `dst` with known-good content, forces the copy to fail (chmod the destination directory read-only so `WriteFileAtomic`'s `os.CreateTemp` fails before any write), and asserts `dst`'s original content is completely untouched. With the old `os.Create`+`io.Copy` implementation this same scenario would have truncated `dst` to empty the instant `os.Create` ran, before the permission failure was even reached.

**Verification:** `go build ./...` clean; `go vet ./internal/catalog/...` clean; `go test ./internal/catalog/... -race -count=1` passes (all pre-existing tests plus the two new ones); `git diff --stat -- cli/create.go` empty (untouched); `go list -deps ./internal/catalog/... | grep -i wailsapp` empty.

### CR-02: `App.WritePartialCatalog` TOCTOU race (duplicate write / clobbered retained state)

**Files modified:** `app.go`, `app_test.go`
**Commit:** `b411d2d9`
**Applied fix (backend layer):** Applied both of the review's suggested layers.
1. Added a new `writeMu sync.Mutex` field to `App`, held for `WritePartialCatalog`'s entire check-decide-write-record sequence. This is a *different* mutex from `scanMu` (which still gets released around the actual filesystem write, since that write can be slow and must not block `StartScan`/`CancelScan`), so it fully serializes concurrent calls to `WritePartialCatalog` itself without blocking unrelated operations: the second of two overlapping calls blocks on `writeMu` until the first finishes, then re-checks the now-cached `lastPartialResult` and returns that pointer instead of performing a second filesystem write.
2. Added a `retainedGen int` field, bumped every time `StartScan` clears the retained-partial fields (the same place `T-25-12`'s existing clearing logic lives). `WritePartialCatalog` captures `gen` before releasing `scanMu` for the write and compares it again after re-acquiring `scanMu` post-write; a mismatch means a newer `StartScan` has since superseded the tree this call was writing, so the result is still returned to that caller (the write to disk genuinely happened) but is **not** cached over — and does not clobber — the newer retained state.

**Applied fix (frontend layer, folded into WR-02's commit):** See WR-02 below — disabling "Retry scan" and "Close without writing" while `writingPartial` is true is the UI-side half of this same fix, explicitly called out as "worth applying" alongside the backend change. A defense-in-depth `if (writingPartialRef.current) return;` guard was also added to `handleCreate` in `CreateSlideOver.tsx` (in case the disabled prop is ever bypassed) — this hunk landed inside the WR-01 commit (`057be886`) rather than its own, due to a `git apply --cached` hunk-splitting quirk on my end (the two hunks in that file ended up staged together); it is still present and correct, just misattributed by commit message. Flagging this explicitly for transparency rather than leaving it silently merged.

**Verification — status: fixed, but flagged "requires human verification" for the concurrency-correctness half.** Per the verification strategy's logic-bug carve-out: Tier 1/2 (re-read + `go vet`/`go build`) only confirm the code compiles and reads correctly, not that the mutex/generation interplay is race-free under real concurrent load.
- `TestWritePartialCatalog_ConcurrentCallsWriteOnce` (new) is a **deterministic** regression test for the duplicate-write half of CR-02: it launches 8 goroutines calling `WritePartialCatalog()` simultaneously and asserts all 8 return the *identical* result pointer. Because `writeMu` fully serializes the method, this guarantee holds regardless of goroutine-scheduling order — it is not a timing-sensitive/flaky test. Passes under `go test -race`.
- `TestStartScan_ClearsRetainedPartialOnNewScan` was extended to assert `retainedGen` advances across a `StartScan` call, confirming the counter is wired correctly.
- The specific interleaving from CR-02 item 2 (a `WritePartialCatalog` write still in flight while a *concurrent* `StartScan` supersedes the retained tree, and the write's late completion must not clobber the new state) is **not** independently exercised by an automated test. Producing a fully deterministic test for that exact interleaving would require adding a test-only synchronization hook to `WritePartialCatalog` itself (mirroring the existing `testHook` pattern `startScan` already uses for its own concurrency tests) — I judged that additional production-code surface, added solely for test determinism on a fix already time/scope-boxed to this review's four findings, to be outside the requested scope. The fix itself is a standard, correctly-ordered generation-counter guard (read-gen-before-unlock, write, re-lock, compare-gen-before-mutating) — the same idiom `T-25-12`'s existing clear-on-StartScan logic already establishes elsewhere in this file — and passed `go build`/`go vet`/`go test -race` cleanly, but I'm marking this half **"fixed: requires human verification"** rather than a plain "fixed" so it gets a manual double-check (e.g. a brief live `wails dev` session forcing the exact click sequence CR-02 describes: click "Write partial catalog", then immediately click "Retry scan" before it resolves — note constraint 8 says `wails dev` is not currently running, so this specific live check is deferred to your restart).

### WR-01: Stale closure in the ⌘↵ keydown handler

**Files modified:** `frontend/src/components/workspace/CreateSlideOver.tsx`
**Commit:** `057be886`
**Applied fix:** Used the review's ref-based alternative (rather than widening the effect's dependency array) — added `handleCreateRef`, updated on every render (`handleCreateRef.current = handleCreate`), and changed the keydown listener to call `handleCreateRef.current()` instead of `handleCreate()` directly. This means the listener always invokes the *current* render's closure (current `options`/`secondaryDir`/etc.) regardless of which deps the registration effect re-runs on, so the effect's dependency array could be trimmed to just `[isOpen, scan.status]` (the two things that actually determine whether the listener should be registered at all) — removing the `eslint-disable-next-line react-hooks/exhaustive-deps` comment, since it was suppressing a real, applicable warning. The mouse path (`onClick={handleCreate}` on the Create button) was already correct and is unchanged.

**Verification:** `npx tsc --noEmit` clean; `npm run build` succeeds. Not independently confirmed live in `wails dev` per constraint 8 (server not running this session) — this is a pure-frontend closure-correctness fix with no Go-side surface, so `tsc` + `build` is the available automated proof; a toggle-then-⌘↵ live check is recommended once `wails dev` is restarted.

### WR-02: `ErrorBody`'s "Retry scan"/"Close without writing" not disabled during in-flight partial write

**Files modified:** `frontend/src/components/workspace/create/ErrorBody.tsx`
**Commit:** `141b6e6a`
**Applied fix:** Added `disabled={writingPartial}` to both the "Retry scan" and "Close without writing" buttons (previously only the primary "Write partial catalog" button had this guard). No copy strings were touched — "Retry scan" and "Close without writing" remain exactly as locked in `25-UI-SPEC.md`.

**Verification:** `npx tsc --noEmit` clean; `npm run build` succeeds; re-read confirms both buttons now carry the `disabled` prop and no other JSX/copy changed. Not independently confirmed live (same constraint-8 note as WR-01).

## Skipped Issues

None — all four in-scope findings were fixed.

---

_Fixed: 2026-08-14_
_Fixer: Claude (gsd-code-fixer)_
_Iteration: 1_
