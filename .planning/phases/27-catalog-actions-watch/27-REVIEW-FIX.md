---
phase: 27-catalog-actions-watch
fixed_at: 2026-08-16T15:40:00Z
review_path: .planning/phases/27-catalog-actions-watch/27-REVIEW.md
iteration: 2
findings_in_scope: 1
fixed: 1
skipped: 0
status: all_fixed
---

# Phase 27: Code Review Fix Report

**Fixed at:** 2026-08-16T15:40:00Z
**Source review:** .planning/phases/27-catalog-actions-watch/27-REVIEW.md
**Iteration:** 2

**Summary (this iteration):**
- Findings in scope (critical + warning): 1 (WR-01)
- Fixed: 1
- Skipped: 0

**Cumulative picture across both iterations:**
- Iteration 1 fixed CR-01, WR-01, WR-02, WR-03 (against `27-REVIEW.iter2.md`'s findings) and skipped WR-04
  (verified live, did not reproduce). See that iteration's detail below, preserved from the prior report.
- Iteration 2 (this run) fixes the one new warning the re-review surfaced while re-verifying iteration 1's
  WR-03 fix: iteration-1's WR-03 fix (containment-checking the `.html` sibling) made a pre-existing
  `RenameCatalog` ordering bug newly reachable via its own rejection path.
- Net result after both iterations: 5 of 5 warning findings raised across both reviews are fixed; 1 (WR-04)
  is documented as not reproducing; 0 findings remain open.

**Verification environment:** `go build ./... && go vet ./... && go test ./... -race -count=1` and
`(cd frontend && npx tsc --noEmit && npm run build)` both ran in the main checkout
(`workflow.use_worktrees: false`, so no isolated worktree was created for this run) and are reproducible
from the current tree.

## Fixed Issues (this iteration)

### WR-01: `RenameCatalog` mutates the JSON title before validating the `.html` sibling, so a rejected rename leaves the JSON silently renamed anyway

**Files modified:** `internal/catalog/rename.go`, `internal/catalog/rename_test.go`
**Commit:** `de491fae`
**Applied fix:** Reordered `RenameCatalog` so every step that can fail runs before anything is mutated:
resolve + containment-check + read the `.html` sibling (via the existing `resolveContainedSibling` from the
iteration-1 WR-03 fix) and pre-compute the patched HTML bytes first; only then write the JSON, and only then
write the HTML. A missing sibling (`os.IsNotExist`) still takes the pre-existing "rename with no `.html`" path,
now expressed as `hasHTML = false` rather than an early return, so the JSON write happens exactly once on every
code path. The reviewer's suggested reordering applied cleanly to the current tree with no adaptation needed.

**Regression test added:** `TestRenameCatalog_RejectedHTMLStepLeavesJSONTitleUnchanged` in `rename_test.go`,
table-driven over the two ways the sibling step can be rejected (symlink escaping the catalog directory, and
an unreadable `.html` file), asserting the JSON file's bytes are byte-identical before and after a rejected
rename. Confirmed failing against the pre-fix code via a `git stash` round-trip of `rename.go` alone (both
subtests failed, showing the title mutated to "Photos 2024" despite the returned error — reproducing the
reviewer's manual repro exactly) and passing against the fix. This is the first durable/committed coverage
for this bug; the reviewer's own diagnostic test was deliberately not committed.

**DuplicateCatalog — same hazard class? Considered, and no:** `DuplicateCatalog` writes its destination
`<newRoot>.json` before resolving/reading the *source* `.html` sibling, which looks superficially like the
same ordering. It is not the same hazard, for two reasons specific to that function: (1) the destination JSON
it writes is always a brand-new file — `nextCopyRoot`/`isCandidateRootFree` guarantee `<newRoot>.json` and
`<newRoot>.html` don't exist yet — so a later failure never mutates a pre-existing, already-trusted catalog
file the way `RenameCatalog`'s in-place JSON write did; the worst case is an orphaned new file, not a
corrupted old one. (2) `DuplicateCatalog`'s own doc comment and its return signature (`(string, error)`)
deliberately report "exactly what succeeded" on a later HTML-copy failure by returning the new JSON path
alongside the error, rather than silently discarding that fact — the exact opposite of the silent-mutation
problem WR-01 flagged for `RenameCatalog` (which returns `error` only, with nothing for a caller to inspect).
Left unchanged this iteration. One caveat worth recording for a future pass, not fixed here since it's a
pre-existing, separately-scoped concern rather part of this finding: per WR-01's own text, Wails discards a
Go binding's non-error return value on the JS side when the promise rejects, so `DuplicateCatalog`'s
"report what succeeded" path-return doesn't currently reach the frontend either — the backend's intent is
sound, the wire contract just doesn't carry it through yet.

**Residual gap — explicitly not claiming full atomicity:** This fix closes the entire
containment/read-failure class (the dominant one, and the one iteration 1's WR-03 fix added a new trigger
for) by moving every fallible step ahead of the first write. It does **not** make the two-file update
atomic: `WriteFileAtomic` makes each individual file write crash-safe, but if the process is killed or the
disk fails in the narrow window *between* the JSON write succeeding and the HTML write starting/completing,
the JSON and HTML titles can still end up out of sync. Achieving true all-or-nothing atomicity across two
independent files would require additional machinery (e.g. a write-ahead journal, or renaming both from
staged temp files under a directory-level lock) that this phase's scope does not call for and that isn't
present anywhere else in this codebase's file-write patterns. This is documented in a code comment at the
second `WriteFileAtomic` call in `rename.go` so the residual is visible in place, not just in this report.

## Skipped Issues (this iteration)

None — the single in-scope finding was fixed.

---

## Iteration 1 detail (preserved for history)

**Summary:**
- Findings in scope (critical + warning): 5 (CR-01, WR-01, WR-02, WR-03, WR-04 — against `27-REVIEW.iter2.md`)
- Fixed: 4
- Skipped: 1 (WR-04 — verified live, does not reproduce)

Note: IN-01 (info-level, a previously accepted trade-off) was out of `fix_scope: critical_warning` and not
touched.

### CR-01: Menu.tsx click-outside close loses focus restore to `<body>`
**Files modified:** `frontend/src/components/workspace/Menu.tsx`
**Commit:** `12c32dd3`
**Applied fix:** Added `event.preventDefault()` in `handlePointerDown`'s outside-click branch, before
`onClose()`, so the browser's compatibility `mousedown`/`click` dispatch (and its default focus-follows-click
action) never fires for that gesture, leaving `useModalBehavior`'s `restoreTarget.focus()` as the only focus
mutation. Live-verified via dev-browser against `wails dev` with real CDP-trusted mouse input.

### WR-01 (iteration 1 numbering — internal/watch): error-path callback invocation contradicted its documented threading contract
**Files modified:** `internal/watch/watcher.go`
**Commit:** `993bd266`
**Applied fix:** Changed `w.c.fireNow()` to `go w.c.fireNow()` in `loop()`'s `Errors` branch, matching
`trigger()`'s existing off-loop delivery. Verified with `go test ./internal/watch/... -race -count=1`.

### WR-02: `WriteFileAtomic`'s best-effort directory-sync failure would log on every write on Windows
**Files modified:** `internal/catalog/atomicwrite.go`, `.planning/WINDOWS.md`
**Commit:** `ed2a8d1a`
**Applied fix:** Added a package-level `sync.Once` guarding the log call, so the failure logs at most once
per process lifetime instead of on every write.

### WR-03: `RenameCatalog`/`DuplicateCatalog`'s derived `.html` sibling bypassed `osutil.ContainsPath`
**Files modified:** `internal/catalog/rename.go`, `internal/catalog/duplicate.go`,
`internal/catalog/rename_test.go`, `internal/catalog/duplicate_test.go`
**Commits:** `f143b229` (fix), `767d2dfe` (regression tests)
**Applied fix:** Added `resolveContainedSibling` — resolves symlinks and containment-checks the derived
`.html` sibling against its own parent directory via `osutil.ContainsPath`, the same function
`TrashPaths`/`GetCatalogHtmlPath`/`RenameCatalog`'s primary-path check already use. This is the helper the
iteration-2 WR-01 fix (above) reuses without modification.

### WR-04: `useModalBehavior`'s focus-restore might also lose to `<body>` on `DialogShell`'s own close-button click
**Reason skipped:** Speculative per the review itself; verified live via dev-browser using real CDP-trusted
mouse clicks on both the "×" close button and the footer cancel button — neither reproduced the speculated
failure mode. `useModalBehavior.ts` left untouched (shared by every overlay in the app).

---

_Fixed: 2026-08-16T15:40:00Z_
_Fixer: Claude (gsd-code-fixer)_
_Iteration: 2_
