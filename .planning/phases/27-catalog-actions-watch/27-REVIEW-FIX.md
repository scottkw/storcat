---
phase: 27-catalog-actions-watch
fixed_at: 2026-08-16T14:30:00Z
review_path: .planning/phases/27-catalog-actions-watch/27-REVIEW.md
iteration: 1
findings_in_scope: 5
fixed: 4
skipped: 1
status: partial
---

# Phase 27: Code Review Fix Report

**Fixed at:** 2026-08-16T14:30:00Z
**Source review:** .planning/phases/27-catalog-actions-watch/27-REVIEW.md
**Iteration:** 1

**Summary:**
- Findings in scope (critical + warning): 5 (CR-01, WR-01, WR-02, WR-03, WR-04)
- Fixed: 4
- Skipped: 1 (WR-04 — verified live, does not reproduce)

Note: IN-01 (info-level, a previously accepted trade-off) was out of `fix_scope: critical_warning` and not touched.

**Verification environment:** all live browser verification ran against the pre-existing `wails dev` process on `:34115` (fresh `window.go.main.App` bindings confirmed, 35 methods) via the dev-browser skill, using real CDP-level mouse input (not synthetic `dispatchEvent`) — synthetic events do not trigger the browser's native focus-follows-click default action, so an earlier synthetic-event attempt gave a false negative before switching to `page.click()`/`page.mouse.click()`. Go verification (`go build ./... && go vet ./... && go test ./... -race -count=1`) and frontend verification (`npx tsc --noEmit && npm run build`) both ran in the main checkout (`workflow.use_worktrees: false`, so no isolated worktree was created for this run) and are reproducible from the current tree.

## Fixed Issues

### CR-01: Menu.tsx click-outside close loses focus restore to `<body>`

**Files modified:** `frontend/src/components/workspace/Menu.tsx`
**Commit:** `12c32dd3`
**Applied fix:** Added `event.preventDefault()` in `handlePointerDown`'s outside-click branch, before `onClose()`. Per the Pointer Events spec this suppresses the browser's compatibility `mousedown`/`click` dispatch (and their default focus-follows-click action) for mouse-originated pointer input, so `useModalBehavior`'s `restoreTarget.focus()` is the only focus mutation left standing for that interaction.

**Live verification:** Used the dev-browser skill against the running `wails dev` process (`:34115`) with real CDP-trusted mouse events (`page.click()` / `page.mouse.click()` — a first attempt using JS-level `dispatchEvent` synthetic events gave a false negative on both pre- and post-fix code, because untrusted synthetic events don't trigger the browser's native focus-follows-click default action; switching to CDP-level input reproduced the defect correctly).
- Pre-fix code (round-tripped via `git stash`): opened the `⋯` menu, clicked outside → `document.activeElement` was `BODY` — reproduced the regression exactly as `WINDOWS.md` entry 13 described.
- Post-fix code: same steps → `document.activeElement.className === 'ws-details-overflow'` (the trigger button), across repeated open/close cycles.

**WINDOWS.md entry 13** updated with this evidence and marked `fixed` via `gsd-tools windows fixed 13` (commit `ed2a8d1a`, bundled with the WR-02 fix since both touch `WINDOWS.md`). Not confirmed in the actual shipped app's WKWebView engine on macOS (Chromium/WebKit can differ on native focus-follows-click timing) — host-OS GUI automation remains prohibited by this project's standing constraint, so that gap stays open as a residual, documented risk rather than being papered over.

### WR-01: `internal/watch`'s error-path callback invocation contradicts its own documented threading contract

**Files modified:** `internal/watch/watcher.go`
**Commit:** `993bd266`
**Applied fix:** Changed the `Errors` branch in `loop()` from `w.c.fireNow()` to `go w.c.fireNow()`. `fireNow()` itself is documented and tested (`TestCoalescer_FiresNowIsImmediate`) to invoke `fn()` synchronously on its caller — that immediacy is a locked-in contract for its other caller pattern, so `fireNow()`'s own implementation was left untouched. Dispatching the call via `go` at this one call site keeps the callback off the `loop()` goroutine, matching `trigger()`'s existing off-loop delivery and closing the contradiction with the package doc comment's stated threading contract.
**Verification:** `go build`, `go vet`, and `go test ./internal/watch/... -race -count=1` all pass, including the existing `TestCoalescer_FiresNowIsImmediate` (unaffected — `fireNow()`'s own contract is unchanged) and the watcher-level debounce/burst tests.

### WR-02: `WriteFileAtomic`'s best-effort directory-sync failure will log on every single write on Windows

**Files modified:** `internal/catalog/atomicwrite.go`, `.planning/WINDOWS.md`
**Commit:** `ed2a8d1a`
**Applied fix:** Added a package-level `sync.Once` (`logDirSyncFailureOnce`) guarding the `log.Printf` call in `WriteFileAtomic`, so the directory-sync failure is logged at most once per process lifetime instead of on every write — satisfying both goals simultaneously: the failure stays observable (log-and-continue, never silently discarded), and a platform where it is expected to fail on every call (Windows) no longer drowns the log.

Also corrected `WINDOWS.md` entry 11's stale wording ("the error is deliberately discarded") — the error was already being logged unconditionally before this fix, not discarded; the description now reflects the actual pre-fix and post-fix behavior.
**Verification:** `go build`, `go vet`, and `go test ./internal/catalog/... -race -count=1` pass, including the existing `TestWriteFileAtomic_DirSyncFailureIsNotFatal`.

### WR-03: `RenameCatalog`/`DuplicateCatalog`'s derived `.html` sibling bypasses `osutil.ContainsPath`

**Files modified:** `internal/catalog/rename.go`, `internal/catalog/duplicate.go`, `internal/catalog/rename_test.go`, `internal/catalog/duplicate_test.go`
**Commits:** `f143b229` (fix), `767d2dfe` (regression tests)
**Applied fix:** Added a shared `resolveContainedSibling(siblingPath, baseDir string) (string, error)` helper in `rename.go` that resolves symlinks and containment-checks the derived `.html` sibling against its own parent directory via `osutil.ContainsPath` — mirroring `osutil.TrashPaths`'s own belt-and-braces re-validation. `RenameCatalog` now reads/writes the resolved path (rejecting the operation if the sibling resolves outside its own directory); `DuplicateCatalog`'s read-only source-html lookup goes through the same gate. Deliberately did not add a `catalogDir` parameter to either function's public signature (which would have forced ~20 mechanical test-call-site updates for no behavioral gain) — the sibling's own parent directory is the correct, narrower containment root, since the derived path is constructed to live in exactly that directory by string convention. The duplicate's *destination* `.html` needed no equivalent check: `nextCopyRoot`/`isCandidateRootFree` already guarantee neither `<newRoot>.json` nor `<newRoot>.html` exists yet, so it cannot be a pre-planted symlink.

Added one regression test per function (`TestRenameCatalog_RejectsHTMLSymlinkEscapingCatalogDir`, `TestDuplicateCatalog_RejectsHTMLSymlinkEscapingCatalogDir`), each confirmed to fail against the pre-fix code (verified via a `git show`-based round-trip of the two source files) and pass against the fix.
**Verification:** `go build`, `go vet`, and `go test ./internal/catalog/... -race -count=1` pass (including the full pre-existing rename/duplicate suite plus the two new tests); `go build ./...` (whole repo) confirms no import cycle from `internal/catalog` importing `internal/osutil`.

## Skipped Issues

### WR-04: `useModalBehavior`'s focus-restore may also lose to `<body>` on `DialogShell`'s own close-button click

**File:** `frontend/src/hooks/useModalBehavior.ts:135-141`, `frontend/src/components/workspace/DialogShell.tsx:41-43`
**Reason:** Speculative per the review itself ("not confirmed live this session"). Verified live via dev-browser against `wails dev` (`:34115`) using real CDP-trusted mouse clicks, per the fix guidance's explicit instruction to verify before fixing: opened the Rename dialog, clicked the "×" close button — `document.activeElement.className === 'ws-details-overflow'` (the trigger), not `<body>`. Repeated with the "Keep original title" footer cancel button — same result. Neither close path reproduces the speculated failure mode; `useModalBehavior.ts` was left untouched (it is shared by every overlay in the app, so a speculative fix there would have been the higher-risk move for a defect that does not exist). Evidence recorded in `WINDOWS.md` entry 13's updated description (bundled with the CR-01 evidence, since both were tested in the same session).
**Original issue:** Speculated that a real `<button>`'s own click-driven focus, followed by its removal on `DialogShell` unmount, could fall back focus to `<body>` before `restoreTarget.focus()` runs in the cleanup effect — the same failure class as CR-01 reached via a different path.

---

_Fixed: 2026-08-16T14:30:00Z_
_Fixer: Claude (gsd-code-fixer)_
_Iteration: 1_
