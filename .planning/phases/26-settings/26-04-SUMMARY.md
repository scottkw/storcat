---
phase: 26-settings
plan: 04
subsystem: security

tags: [go, wails, react, typescript, security, containment, path-traversal]

# Dependency graph
requires:
  - phase: 26-settings
    provides: "26-03's config-backed catalogDirectory (Settings + rail shared value) -- the boundary this plan enforces"
provides:
  - "internal/osutil.ResolveContainedFileURL -- scheme/regular-file/extension/containment validator for App.OpenExternal"
  - "GetCatalogHtmlPath and OpenExternal both catalogDir-gated, matching RevealInFileManager's WR-02 treatment"
  - "DetailsPanel.handleOpenHtml fail-closed guard, same shape as handleReveal's"
  - "Deletion of frontend/src/components/CatalogModal.tsx -- unsanitized iframe srcDoc surface removed"
affects: [27, 28]

# Actuals (#2632)
actuals:
  tokens: 8591
  tasks: 3
  commits: 7

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "ResolveContainedFileURL follows RevealInFileManager's validate-then-act ordering exactly: derive path, require absolute, EvalSymlinks, Stat+IsRegular, extension allowlist, ContainsPath -- reusing ContainsPath and allowedRevealExtensions rather than a second implementation"
    - "OpenExternal now passes the validator's returned canonical resolved URL to runtime.BrowserOpenURL, never the caller's raw string -- closes the symlink-reintroduction gap a validate-then-pass-through-original design would leave open"
    - "Frontend fail-closed guard pattern (handleReveal/handleOpenHtml): check catalogDir truthiness client-side before calling a binding that would reject an empty catalogDir anyway, so the user sees the app's own message instead of a raw Go error string"

key-files:
  created:
    - internal/osutil/openexternal.go
    - internal/osutil/openexternal_test.go
  modified:
    - app.go
    - app_test.go
    - frontend/wailsjs/go/main/App.d.ts
    - frontend/wailsjs/go/main/App.js
    - frontend/src/services/wailsAPI.ts
    - frontend/src/components/workspace/DetailsPanel.tsx
    - frontend/src/App.tsx
    - frontend/src/themeTokens.ts
    - .planning/STATE.md
  deleted:
    - frontend/src/components/CatalogModal.tsx

key-decisions:
  - "OpenExternal restricted to file:// URLs and bare absolute paths resolving inside catalogDir, rejecting every other URL scheme via an allow-first (not a deny-list) design -- a new hostile scheme (vbscript:, etc.) can't slip through the way it could with a growing denylist"
  - "GetCatalogHtmlPath keeps returning the unresolved html path (not the symlink-resolved form) on success, since OpenExternal independently re-validates and re-resolves -- avoids churning the three pre-existing tests for no correctness gain"
  - "Reachability of openCatalogModal re-confirmed by grep immediately before deleting CatalogModal.tsx (only themeChange is dispatched anywhere in frontend/src), per the plan's explicit re-verification requirement rather than trusting the stale discuss-time analysis"
  - "Task 2's own tsc/build verify gate was satisfied only after Task 3 ran (CatalogModal.tsx's stale one-argument getCatalogHtmlPath call blocked it) -- documented as a deviation below rather than patching a file about to be deleted"

requirements-completed: [SET-04]

coverage:
  - id: D1
    description: "ResolveContainedFileURL validates raw OpenExternal input (file:// URL or bare absolute path) and rejects every non-file scheme, non-regular-file, disallowed-extension, or out-of-catalogDir path before anything opens"
    requirement: "SET-04"
    verification:
      - kind: unit
        ref: "internal/osutil/openexternal_test.go#TestResolveContainedFileURL (16 subtests: accept x4, reject x12)"
        status: pass
      - kind: automated_ui
        ref: "dev-browser session: window.go.main.App.OpenExternal called directly against :34115 -- a contained .html path succeeded, /etc/passwd (outside catalogDir) rejected with 'unsupported extension \"\"', http://example.com rejected with 'unsupported URL scheme \"http\"'"
        status: pass
    human_judgment: false
  - id: D2
    description: "GetCatalogHtmlPath and OpenExternal both require catalogDir and reject a path resolving outside it, reusing osutil.ContainsPath -- no second containment implementation"
    requirement: "SET-04"
    verification:
      - kind: unit
        ref: "app_test.go#TestGetCatalogHtmlPath_RejectsEmptyCatalogDir, TestGetCatalogHtmlPath_RejectsPathOutsideCatalogDir, TestOpenExternal_RejectsBeforeTouchingRuntime"
        status: pass
    human_judgment: false
  - id: D3
    description: "DetailsPanel's 'Open HTML catalog' button still opens a catalog's HTML companion after the containment gate lands, with the identical fail-closed catalogDir guard handleReveal already has"
    requirement: "SET-04"
    verification:
      - kind: automated_ui
        ref: "dev-browser session against :34115: seeded a temp catalog dir with testcat.json/testcat.html, pointed the app at it via SetCatalogDirectory + localStorage cache write, reloaded, selected the rail row, clicked 'Open HTML catalog' -- no error span rendered, repeated after Task 3's deletion with the same result"
        status: pass
    human_judgment: false
  - id: D4
    description: "The unreachable CatalogModal.tsx and its App.tsx wiring are deleted; antd's ConfigProvider and the themeChange listener survive untouched; the app still boots cleanly"
    requirement: "SET-04"
    verification:
      - kind: unit
        ref: "grep -rn 'openCatalogModal' frontend/src (zero matches); grep -q 'ConfigProvider'/'themeChange' frontend/src/App.tsx (both present); npx tsc --noEmit && npm run build (both green)"
        status: pass
      - kind: automated_ui
        ref: "dev-browser session: page.reload() against :34115 with a console/pageerror listener attached -- zero errors, document.documentElement's data-theme rendered 'gruvbox-dark' (a non-default theme)"
        status: pass
    human_judgment: false

duration: ~25min
completed: 2026-08-15
status: complete
---

# Phase 26 Plan 04: Security Hardening -- Catalog-Directory Containment + Unreachable Modal Deletion Summary

**Both remaining carried security obligations (T-22-05, FU-23-A) discharged: `GetCatalogHtmlPath`/`OpenExternal` now enforce the Settings-configured catalog directory via a new `osutil.ResolveContainedFileURL` validator, and the unreachable `CatalogModal.tsx` (unsanitized `iframe srcDoc`) is deleted outright with its `App.tsx` wiring.**

## Performance

- **Duration:** ~25 min
- **Started:** 2026-08-15T09:25:08-05:00
- **Completed:** 2026-08-15T09:32:52-05:00
- **Tasks:** 3
- **Files modified:** 12 (2 created, 9 modified, 1 deleted)

## Accomplishments

- `internal/osutil.ResolveContainedFileURL(raw, catalogDir string) (string, error)` -- a pure, Wails-free validator accepting only `file://` URLs or bare absolute paths that resolve (after `EvalSymlinks`) to a regular `.json`/`.html` file inside `catalogDir`; rejects every other scheme, shape, and location with a wrapped error naming the rejected input. Reuses `ContainsPath` and `allowedRevealExtensions` -- no second containment implementation.
- `App.GetCatalogHtmlPath(catalogPath, catalogDir string)` and `App.OpenExternal(rawURL, catalogDir string) error` both gained the required `catalogDir` parameter and reject a path resolving outside it; `OpenExternal` now returns an error (previously silent) and opens the validator's canonical resolved URL, never the caller's raw string.
- Wails bindings regenerated (`wails generate module`); `wailsAPI.ts`'s `getCatalogHtmlPath`/`openExternal` thread `catalogDir` through exactly as `revealInFileManager` already did.
- `DetailsPanel.tsx`'s `handleOpenHtml` gained the identical fail-closed `catalogDir` guard `handleReveal` already had, and now surfaces an `openExternal` rejection through the same `setError` path.
- `frontend/src/components/CatalogModal.tsx` deleted along with its `App.tsx` wiring (import, two `useState` hooks, `openCatalogModal` listener/cleanup, close handler, element) after re-confirming at execution time that nothing dispatches `openCatalogModal`. `ConfigProvider` and the `themeChange` listener were left untouched -- antd remains a live dependency.
- `.planning/STATE.md`'s Pending Todos corrected: both obligations marked discharged, `SearchIndexed`'s exclusion reason recorded, `WINDOWS.md` entry 3 noted as already fixed.

## Task Commits

Each task was committed atomically (Task 1 and the Go portion of Task 2 as TDD RED/GREEN pairs):

1. **Task 1: ResolveContainedFileURL -- RED** - `1533272d` (test)
2. **Task 1: ResolveContainedFileURL -- GREEN** - `f1f75f0a` (feat)
3. **Task 2: catalogDir-threaded bindings -- RED** - `9083341e` (test)
4. **Task 2: catalogDir-threaded bindings -- GREEN** - `588e23ff` (feat)
5. **Task 2: frontend wailsAPI/DetailsPanel wiring** - `58ac6c90` (feat)
6. **Task 3: delete CatalogModal.tsx** - `68301c1a` (feat)
7. **Task 3: App.tsx wiring removal + STATE.md correction** - `be2e1881` (feat)

**Plan metadata:** (this commit)

## Files Created/Modified

- `internal/osutil/openexternal.go` - `ResolveContainedFileURL`, `derivePath` -- the FU-23-A validator for `OpenExternal`
- `internal/osutil/openexternal_test.go` - 16-subtest table covering every accept/reject case in the plan's behavior block
- `app.go` - `GetCatalogHtmlPath`/`OpenExternal` gain `catalogDir`, `OpenExternal` returns an error
- `app_test.go` - three existing tests updated to pass `catalogDir`; three new tests added
- `frontend/wailsjs/go/main/App.d.ts`, `App.js` - regenerated for both two-argument signatures
- `frontend/src/services/wailsAPI.ts` - `getCatalogHtmlPath`/`openExternal` take `catalogDir`
- `frontend/src/components/workspace/DetailsPanel.tsx` - `handleOpenHtml` fail-closed guard + error surfacing
- `frontend/src/App.tsx` - `CatalogModal` import/state/listener/handler/element removed
- `frontend/src/themeTokens.ts` - `applyTokens` doc comment no longer names the deleted file
- `.planning/STATE.md` - both carried obligations marked discharged
- `frontend/src/components/CatalogModal.tsx` - deleted

## Decisions Made

- `OpenExternal`'s scheme check is allow-first (`file` only), not a growing denylist -- closes off future hostile schemes without maintenance
- `GetCatalogHtmlPath` keeps returning the unresolved html path on success rather than switching to the resolved form, since `OpenExternal` independently re-validates and re-resolves; changing the return value would only churn the three pre-existing tests
- Re-ran the `dispatchEvent` reachability grep immediately before deleting `CatalogModal.tsx` per the plan's explicit instruction, rather than trusting the discuss-time analysis as still valid
- Restored the wails-dev process's live `catalogDirectory` setting to its pre-session value after live verification, so the running dev instance was left exactly as found

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking, resolved by sequencing not a code fix] Task 2's own tsc/build verify could not pass in isolation**
- **Found during:** Task 2's verify step
- **Issue:** `frontend/src/components/CatalogModal.tsx` (not in Task 2's file list) still called `window.electronAPI.getCatalogHtmlPath(catalogPath)` with one argument. Once `wailsAPI.getCatalogHtmlPath` required a second `catalogDir` argument, `npx tsc --noEmit` failed with `Expected 2 arguments, but got 1` -- a compile error entirely caused by a file Task 3 (later in this same plan) deletes.
- **Fix:** Did not patch the doomed call site. Proceeded with Task 2's commit (Go tests fully green, bindings regenerated, all grep-based acceptance criteria passing) and deferred the `tsc`/`build` confirmation to immediately after Task 3's deletion, where it passed clean. Both tasks landed in the same continuous plan execution, so no window existed where the codebase was left in a broken, unverified state across separate agent runs.
- **Files modified:** None beyond what Task 2 already touched.
- **Verification:** `npx tsc --noEmit` and `npm run build` both green after Task 3's commit.
- **Committed in:** `58ac6c90` (Task 2 frontend commit, with the deferral noted in the commit message)

---

**Total deviations:** 1 (a sequencing note, not a code fix)
**Impact on plan:** None on the delivered code -- the plan's own Task 3 was already the fix for the tsc blocker Task 2 hit; this just documents why Task 2's verify gate wasn't literally green in isolation.

## Issues Encountered

- The live-verification catalog directory (`.../f5f2d8cb.../scratchpad/catalogs`) had no `.html` companion files, so a temporary `testcat.json`/`testcat.html` pair was created under this session's own scratchpad, pointed at via `SetCatalogDirectory` + a matching `localStorage` cache write (mirroring `setCatalogDirectorySetting`'s two-write shape), exercised, then the original `catalogDirectory` was restored via the same two writes so the shared `wails dev` process (relied on by plans 26-03..05) was left in its pre-session state.
- `dev-browser`'s `page.evaluate` only accepts a single argument; multi-value probes were wrapped in one object parameter.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Both carried security obligations (T-22-05, FU-23-A) are fully discharged; `.planning/STATE.md`'s Pending Todos section no longer names either as open.
- `SearchIndexed` remains correctly excluded from the FU-23-A sweep, with its exclusion reason now recorded in STATE.md rather than only in `26-CONTEXT.md`.
- `.planning/WINDOWS.md` entry 3 (Phase 25 CRT-13) is noted as already `fixed`; no ledger action was needed this plan.
- Plan 26-05 (the remaining Settings plan) can proceed against a `wails dev` instance already exercising the new two-argument bindings.

---
*Phase: 26-settings*
*Completed: 2026-08-15*

## Self-Check: PASSED

All created/modified files verified present on disk (or absent, for the deleted CatalogModal.tsx); all seven task commit hashes verified present in `git log`.
