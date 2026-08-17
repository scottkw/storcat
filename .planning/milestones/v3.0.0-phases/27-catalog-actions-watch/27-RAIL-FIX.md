---
phase: 27-catalog-actions-watch
fixed_at: 2026-08-16T15:13:33Z
verification_path: .planning/phases/27-catalog-actions-watch/27-VERIFICATION.md
commit: 5c7e460a
status: fixed
verified_live: true
verification_env: main checkout (workflow.use_worktrees=false), wails dev on :34115
---

# Phase 27: Rail Freshness Fix Report

**Fixed at:** 2026-08-16T15:13:33Z
**Source:** `27-VERIFICATION.md` Human Verification item #6 ("Rail freshness after in-app Delete/Duplicate with watching OFF")
**Commit:** `5c7e460a`

## The Defect

With directory watching disabled (`config.WatchDirectory` defaults to `false`), deleting or
duplicating a catalog through the UI did not update the catalog rail at all. `DeleteConfirmDialog`'s
`onDeleted` only cleared `currentCatalogId`/`selected` when the deleted catalog was the current
selection -- it never removed the row from `state.catalogs`. `duplicateCatalogAction` never dispatched
a new catalog into `state.catalogs` either. Both relied exclusively on the `catalogs:changed`
event-driven refresh, which only fires when the fsnotify watcher is running -- i.e. never, for most
users, since watching defaults to off. `RenameCatalog`'s UI path already updated the rail directly
(optimistic `SET_CATALOGS`), so the app was self-inconsistent: rename refreshed, delete and duplicate
did not.

## The Fix

Reused the existing single refresh path (`CatalogRail.tsx`'s `loadCatalogsForDirectory`, which calls
`wailsAPI.browseCatalogs` and dispatches `SET_CATALOGS`) rather than introducing a second way to
compute the rail's contents:

- **`frontend/src/contexts/AppContext.tsx`** -- added a `railRefreshToken: number` field to
  `AppState` (init `0`) and a `REQUEST_RAIL_REFRESH` action that increments it.
- **`frontend/src/components/workspace/CatalogRail.tsx`** -- added `state.railRefreshToken` to the
  dependency array of the existing `catalogDir`-driven effect. A token bump re-runs the effect and
  calls the exact same `loadCatalogsForDirectory` the mount path, a directory change, and the
  `catalogs:changed` handler all already call.
- **`frontend/src/components/workspace/DetailsPanel.tsx`** -- `duplicateCatalogAction` dispatches
  `REQUEST_RAIL_REFRESH` after `wailsAPI.duplicateCatalog` reports success (never on failure).
  `DeleteConfirmDialog`'s `onDeleted` callback (only invoked by the dialog after
  `wailsAPI.deleteCatalog` reports success) dispatches `REQUEST_RAIL_REFRESH` alongside its existing
  `CLEAR_CURRENT_CATALOG` dispatch.

No local array splicing, no optimistic row removal/insertion, no per-file patching -- the fix
re-invokes the one authoritative `browseCatalogs` listing on a new trigger, honoring 27-CONTEXT.md's
locked "single refresh path" decision (that decision rejected giving *watch events* a per-file-delta
payload; this doesn't touch that).

**Failed-delete guard:** `DeleteConfirmDialog.submit()` only calls `onDeleted()` inside the
`if (result.success)` branch -- a failed delete never reaches the dispatch, never clears the
selection, and never triggers a refresh. Same shape for duplicate (`onError`/return on failure, no
dispatch reached).

**Double-refresh with watching on:** When watching is enabled, the local dispatch fires the refresh
immediately and the debounced `catalogs:changed` event fires a second, idempotent re-list roughly
~300ms later (the watcher's own debounce). This is an acceptable redundant re-list per the fix
guidance, not a duplicate-fetch storm -- confirmed live below, no misbehavior observed.

## Verification

**Static:**
- `go build ./... && go vet ./... && go test ./... -race -count=1` -- all 9 Go packages pass (this
  fix touched no Go files; ran to confirm the baseline stayed green).
- `cd frontend && npx tsc --noEmit && npm run build` -- both clean.

**Live (dev-browser against `wails dev` on `:34115`, real CDP-trusted mouse clicks, run in the main
checkout since `workflow.use_worktrees=false`):**

1. Confirmed watching OFF in Settings (had to toggle it off -- the running dev session had it on from
   earlier phase testing; toggled off and reloaded before the real test).
2. Selected `burst-1.json`, opened the `⋯` menu, clicked **Duplicate catalog**. After the
   `browseCatalogs` round trip resolved (confirmed via temporary debug logging, then removed), the
   rail count went 19 → 21 catalogs (two duplicate clicks across the session) and
   `burst-1-copy.json`/`burst-1-copy-2.json`/`burst-1-copy-3.json` all appeared in the rail with no
   navigation and no page reload.
3. Selected a duplicate, opened `⋯`, clicked **Delete catalog…**, confirmed via the dialog's **Move
   to Trash**. Rail count dropped 21 → 20, the deleted row disappeared immediately, `burst-1.json`
   (the still-current, non-deleted catalog) correctly remained.
4. **One false start worth recording:** my first two live attempts (with a ~300-800ms wait after the
   click) appeared to show no rail update even though the file was confirmed gone/created on disk.
   Instrumented with temporary `console.log` statements and confirmed the refresh effect *was* firing
   correctly on every dispatch, and `browseCatalogs` *was* resolving `success: true` with the updated
   catalog count -- the apparent failure was purely insufficient wait time: this directory's listing
   includes a 3,000-file catalog, and a full re-list measurably takes over a second. Re-ran with a
   ~2s wait and the UI updated correctly every time. Debug logging was removed before committing;
   `git diff --stat` on the commit shows only the intended 3-file, 35-line changes.
5. Re-confirmed watching ON does not misbehave: no errors, no torn listing, no duplicate rows -- the
   local dispatch's immediate re-list and the watcher's later debounced re-list both resolve to the
   same idempotent listing.

Test artifacts (`burst-1-copy*.json`/`.html`) created in the scratch fixture directory during live
verification were deleted afterward; the pre-existing `test-catalog-one-copy.json`/`.html` fixture
from earlier phase testing was left untouched.

## Files Modified

- `frontend/src/contexts/AppContext.tsx`
- `frontend/src/components/workspace/CatalogRail.tsx`
- `frontend/src/components/workspace/DetailsPanel.tsx`

## Commit

`5c7e460a` -- `fix(27): rail reflects local Delete/Duplicate with watching off`

---

*Fixed: 2026-08-16T15:13:33Z*
*Fixer: Claude (gsd-code-fixer)*
