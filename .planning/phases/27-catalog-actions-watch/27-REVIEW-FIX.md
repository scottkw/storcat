---
phase: 27-catalog-actions-watch
fixed_at: 2026-08-16T18:40:00Z
review_path: .planning/phases/27-catalog-actions-watch/27-REVIEW.md
iteration: 3
findings_in_scope: 2
fixed: 2
skipped: 0
status: all_fixed
---

# Phase 27: Code Review Fix Report

**Fixed at:** 2026-08-16T18:40:00Z
**Source review:** .planning/phases/27-catalog-actions-watch/27-REVIEW.md
**Iteration:** 3

**Summary (this iteration):**
- Findings in scope (critical + warning): 2 (WR-01, WR-02)
- Fixed: 2
- Skipped: 0

Both findings this iteration are regressions surfaced by the review of *this phase's own earlier
fixes* — iteration 2's `RenameCatalog` reordering (WR-01) and iteration 1's `Menu.tsx`
`preventDefault()` fix (WR-02) — which is why they were fixed despite the loop's normal 3-iteration
cap.

**Cumulative picture across all three iterations:**
- Iteration 1 fixed CR-01, WR-01 (internal/watch), WR-02 (atomicwrite.go), WR-03 (against
  `27-REVIEW.iter2.md`'s findings) and skipped WR-04 (verified live, did not reproduce).
- Iteration 2 fixed one new warning (`RenameCatalog` write-ordering, numbered WR-01 in that
  iteration's review) that iteration 1's WR-03 containment fix made newly reachable via its own
  rejection path.
- Iteration 3 (this run) fixes two regressions the re-review surfaced while re-verifying iterations
  1 and 2's own fixes:
  - WR-01: `RenameCatalog`'s HTML-write failure path (introduced by iteration 2's reordering)
    returned an error indistinguishable from a no-op failure, even though the JSON had already been
    written.
  - WR-02: `Menu.tsx`'s outside-click `preventDefault()` (added in iteration 1 for CR-01) had a
    broader blast radius than the accepted trade-off described — it suppressed focus on *any*
    focusable click target, not just re-clicks on the trigger.
- Net result after three iterations: every warning raised across all three reviews (7 total) is
  fixed; 1 (iteration-1 WR-04) is documented as not reproducing; 0 findings remain open in scope.
  IN-01 (info-level, this iteration's two-write durability residual) was consciously declined — see
  below.

**Verification environment:** `go build ./... && go vet ./... && go test ./... -race -count=1` and
`(cd frontend && npx tsc --noEmit && npm run build)` both ran in the main checkout
(`workflow.use_worktrees: false`, so no isolated worktree was created for this run) and are
reproducible from the current tree. WR-02's browser-observable behavior was additionally verified
live against a running `wails dev` instance on `:34115` using the dev-browser skill with real
CDP-trusted mouse events (Playwright's `.click()`, not synthetic `dispatchEvent`) — a prior attempt
in this phase using synthetic events had produced a false negative, so this iteration deliberately
used the same trusted-input methodology iteration 1 used for CR-01.

## Fixed Issues (this iteration)

### WR-01: `RenameCatalog`'s HTML-write failure doesn't tell the caller the JSON already changed

**Files modified:** `internal/catalog/rename.go`
**Commit:** `3365e798`
**Applied fix:** When the second `WriteFileAtomic` call (the `.html` write) fails after the first
(`.json`) write has already succeeded, the returned error now names both files and states which one
succeeded: `"rename %s: title updated in %s but failed to update %s: %w"`, instead of the previous
plain `"rename %s: %w"` that was indistinguishable from an error raised before any write happened.
This gives the caller (and, through it, the user) a way to tell "nothing happened, retry is a no-op"
apart from "the title was already changed in the JSON, only the HTML view is now stale" — mirroring
`DuplicateCatalog`'s existing partial-success reporting intent (`duplicate.go:73-79`), applied here
as a distinguishing message rather than a changed return signature (this function's `error`-only
signature was left as-is; only the message content changed).

**Verification:** `go build ./...`, `go vet ./...` clean. `go test ./internal/catalog/... -run
TestRename -v` — all 12 existing subtests pass unmodified (no test asserted the old message shape).
Re-read the modified block to confirm the fix text and surrounding code are intact.

### WR-02: `Menu.tsx`'s outside-click `preventDefault()` blocks focus on other page elements, not just re-clicks on the trigger

**Files modified:** `frontend/src/components/workspace/Menu.tsx`
**Commit:** `ea3a51e0`
**Applied fix:** `handlePointerDown`'s outside-click branch now only calls `event.preventDefault()`
when the click target is *not* itself a focusable element (`a[href],button,input,select,textarea,
[tabindex]:not([tabindex="-1"])`). Previously the blanket `preventDefault()` added for CR-01 (in
iteration 1) suppressed the browser's focus-follows-click default action for every outside click,
which meant clicking straight from the open menu onto another input or button closed the menu but
silently ate that click's own focus, forcing the user to click a second time. The narrowed condition
matches the reviewer's suggested fix as-is (it applied cleanly to the current tree with no
adaptation needed).

**Live verification (dev-browser, real CDP-trusted mouse events against `wails dev` on `:34115`):**
Both required behaviors were confirmed with the fix applied, in a single running session against a
real catalog with a "Catalog actions" (`⋯`) trigger:
- **(a) Non-focusable target still closes the menu and restores focus to the trigger** — clicked a
  non-interactive label element (`"Catalog"` text) inside the details panel while the menu was open.
  Result: `[role="menu"]` count went to 0, and `document.activeElement` was the trigger button
  (`aria-label="Catalog actions"`) — the original CR-01 behavior did not regress.
- **(b) Focusable target closes the menu AND receives focus on the first click** — clicked directly
  on the "Filter catalogs" text input while the menu was open. Result: `[role="menu"]` count went to
  0, and `document.activeElement` was that input (`placeholder="Filter catalogs…"`) — confirming the
  fix resolves the regression; no second click was needed.

Both checks used Playwright's `.click()` (CDP `Input.dispatchMouseEvent`, trusted from the browser's
perspective), not `page.evaluate()`-dispatched synthetic events, per this iteration's explicit
guidance to avoid the false-negative failure mode a synthetic-event attempt produced earlier in this
phase.

**WINDOWS.md entry 13:** Left as `status: fixed` — no update needed. That entry's description
covers the original CR-01 `preventDefault()` addition and its own live re-verification; this
iteration's WR-02 fix narrows that same `preventDefault()` call without reopening the CR-01 behavior
entry 13 verified (confirmed by check (a) above), so the entry's "fixed" status remains accurate.
The narrowing itself is new code covered by this report, not by a ledger entry.

## Skipped Issues (this iteration)

None — both in-scope findings were fixed.

## Declined (info-level, out of `fix_scope`)

### IN-01: The two-write residual could be narrowed further by batching both temp-writes before both renames — declined

**File:** `internal/catalog/rename.go:76-86`, `internal/catalog/atomicwrite.go:37-105`
**Decision:** Not implemented, per explicit instruction for this iteration. The suggestion (a
`WriteFilesAtomicPair`-style helper that batches write+fsync for both files before performing both
renames adjacently and a single trailing directory fsync) would shrink the crash window between the
JSON and HTML writes, but it is not a one-line change — it needs a new shared helper in
`atomicwrite.go` used by both of `RenameCatalog`'s writes — and this phase's scope does not call for
that additional machinery. This is a conscious deferral, not an oversight: recorded here so it is a
visible decision, and the existing code comment at the second `WriteFileAtomic` call in `rename.go`
(iteration 2) already documents the residual gap in place.

---

## Iteration 2 detail (preserved for history)

**Summary:**
- Findings in scope (critical + warning): 1 (WR-01, against iteration 2's review)
- Fixed: 1
- Skipped: 0

### WR-01 (iteration 2 numbering): `RenameCatalog` mutates the JSON title before validating the `.html` sibling, so a rejected rename leaves the JSON silently renamed anyway

**Files modified:** `internal/catalog/rename.go`, `internal/catalog/rename_test.go`
**Commit:** `de491fae`
**Applied fix:** Reordered `RenameCatalog` so every step that can fail runs before anything is
mutated: resolve + containment-check + read the `.html` sibling (via the existing
`resolveContainedSibling` from the iteration-1 WR-03 fix) and pre-compute the patched HTML bytes
first; only then write the JSON, and only then write the HTML. A missing sibling (`os.IsNotExist`)
still takes the pre-existing "rename with no `.html`" path, now expressed as `hasHTML = false`
rather than an early return, so the JSON write happens exactly once on every code path.

**Regression test added:** `TestRenameCatalog_RejectedHTMLStepLeavesJSONTitleUnchanged` in
`rename_test.go`, table-driven over the two ways the sibling step can be rejected (symlink escaping
the catalog directory, and an unreadable `.html` file), asserting the JSON file's bytes are
byte-identical before and after a rejected rename. Confirmed failing against the pre-fix code and
passing against the fix.

**DuplicateCatalog — same hazard class? Considered, and no:** `DuplicateCatalog` always writes to a
brand-new filename (`nextCopyRoot` guarantees `<newRoot>.json`/`.html` didn't previously exist), so a
rejected `.html` step there leaves an incomplete new artifact, never a corrupted pre-existing one.
Left unchanged.

**Residual gap — explicitly not claiming full atomicity:** This fix closes the entire
containment/read-failure class by moving every fallible step ahead of the first write. It does not
make the two-file update atomic — the crash window between the JSON write succeeding and the HTML
write starting/completing remains, documented in a code comment at the second `WriteFileAtomic` call
in `rename.go`. (Iteration 3's IN-01 addresses a possible further narrowing of this exact residual
and was declined — see above.)

---

## Iteration 1 detail (preserved for history)

**Summary:**
- Findings in scope (critical + warning): 5 (CR-01, WR-01, WR-02, WR-03, WR-04 — against
  `27-REVIEW.iter2.md`)
- Fixed: 4
- Skipped: 1 (WR-04 — verified live, does not reproduce)

Note: IN-01 (info-level, a previously accepted trade-off) was out of `fix_scope: critical_warning`
and not touched.

### CR-01: Menu.tsx click-outside close loses focus restore to `<body>`
**Files modified:** `frontend/src/components/workspace/Menu.tsx`
**Commit:** `12c32dd3`
**Applied fix:** Added `event.preventDefault()` in `handlePointerDown`'s outside-click branch,
before `onClose()`, so the browser's compatibility `mousedown`/`click` dispatch (and its default
focus-follows-click action) never fires for that gesture, leaving `useModalBehavior`'s
`restoreTarget.focus()` as the only focus mutation. Live-verified via dev-browser against `wails dev`
with real CDP-trusted mouse input.
**Note (iteration 3):** This blanket `preventDefault()` had a broader blast radius than intended —
see WR-02 above, fixed this iteration, which narrows it to non-focusable targets only.

### WR-01 (iteration 1 numbering — internal/watch): error-path callback invocation contradicted its documented threading contract
**Files modified:** `internal/watch/watcher.go`
**Commit:** `993bd266`
**Applied fix:** Changed `w.c.fireNow()` to `go w.c.fireNow()` in `loop()`'s `Errors` branch,
matching `trigger()`'s existing off-loop delivery. Verified with `go test ./internal/watch/...
-race -count=1`.

### WR-02 (iteration 1 numbering — atomicwrite.go): `WriteFileAtomic`'s best-effort directory-sync failure would log on every write on Windows
**Files modified:** `internal/catalog/atomicwrite.go`, `.planning/WINDOWS.md`
**Commit:** `ed2a8d1a`
**Applied fix:** Added a package-level `sync.Once` guarding the log call, so the failure logs at
most once per process lifetime instead of on every write.

### WR-03: `RenameCatalog`/`DuplicateCatalog`'s derived `.html` sibling bypassed `osutil.ContainsPath`
**Files modified:** `internal/catalog/rename.go`, `internal/catalog/duplicate.go`,
`internal/catalog/rename_test.go`, `internal/catalog/duplicate_test.go`
**Commits:** `f143b229` (fix), `767d2dfe` (regression tests)
**Applied fix:** Added `resolveContainedSibling` — resolves symlinks and containment-checks the
derived `.html` sibling against its own parent directory via `osutil.ContainsPath`, the same
function `TrashPaths`/`GetCatalogHtmlPath`/`RenameCatalog`'s primary-path check already use. This is
the helper the iteration-2 and iteration-3 `RenameCatalog` fixes both reuse without modification.

### WR-04: `useModalBehavior`'s focus-restore might also lose to `<body>` on `DialogShell`'s own close-button click
**Reason skipped:** Speculative per the review itself; verified live via dev-browser using real
CDP-trusted mouse clicks on both the "×" close button and the footer cancel button — neither
reproduced the speculated failure mode. `useModalBehavior.ts` left untouched.

---

_Fixed: 2026-08-16T18:40:00Z_
_Fixer: Claude (gsd-code-fixer)_
_Iteration: 3_
