---
phase: 27-catalog-actions-watch
verified: 2026-08-16T15:10:00Z
status: human_needed
score: 8/9 requirement-level truths verified
behavior_unverified: 6 # WATCH-03's real-quit sub-claim + 5 backstop must_haves truths (concurrent-delete dedup x2, menu no-viewport-flip, delete busy-treatment reuse, catalogs:changed-vs-in-flight-fetch race)
overrides_applied: 0
re_verification: null # no prior VERIFICATION.md for this phase existed
behavior_unverified_items:
  - truth: "Quitting the app releases the watcher on every quit path, via main.go's OnShutdown hook, exercised as a real OS quit-and-relaunch (WATCH-03)."
    test: "With watching enabled, quit the built app (not `wails dev`) and relaunch it; confirm no leaked watch handle and no crash/hang."
    expected: "App relaunches cleanly; a fresh `lsof -p <pid>` (or platform equivalent) shows no dangling watch descriptor from the prior session."
    why_human: "27-07's own live-verification matrix (row 27) explicitly did not exercise a real quit — doing so would have killed the shared `wails dev` session other rows depended on. Substitute evidence (`lsof` fd-count 101→74 on the mechanistically identical `Close()` called by the toggle-off path) is reasoned, not a direct observation of the `OnShutdown` path itself."
  - truth: "Two concurrent DeleteCatalog calls for the same catalog cannot both report success while leaving a file behind (27-01/27-03 must_haves, verification: backstop)."
    test: "Fire two DeleteCatalog calls for the same catalog's paths at effectively the same time (e.g. two goroutines) and confirm only one trash-seam invocation reaches a real path, with the second no-op'ing cleanly through the not-exist skip."
    expected: "No error, no double-trash-attempt failure, no file left behind."
    why_human: "Declared `verification: backstop` in the PLAN frontmatter — reasoned from `TrashPaths`'s already-tested not-exist skip (`TestTrashPaths_SkipsMissingPath`, `TestTrashPaths_AllMissingReturnsNil`), not exercised under genuine concurrency."
  - truth: "Menu.tsx never needs viewport-flip logic — three short items opening downward from a trigger near the top of a bounded-height panel cannot overflow the window at any supported size (27-04 must_haves, verification: backstop)."
    test: "Open the catalog-actions menu at the smallest supported window height/width combinations (e.g. 1040×700, narrow-tier drawer) and confirm the menu never clips off-screen."
    expected: "Menu always fully visible."
    why_human: "Declared backstop — reasoned from the app's layout geometry, not proven by an automated boundary sweep across every supported window size."
  - truth: "DeleteConfirmDialog's in-flight busy treatment (opacity 0.7, disabled buttons) matches DetailsPanel's existing Footer pattern (27-05 must_haves, verification: backstop)."
    test: "Trigger a real (slow) delete and observe the busy state live."
    expected: "Both footer buttons disabled, primary at opacity 0.7, no spinner, for the duration of the async call."
    why_human: "Declared backstop — reused by reading Footer's implementation, not independently re-verified by a live click against this dialog's own async Trash call."
  - truth: "A catalogs:changed emission landing while a browseCatalogs call is already in flight cannot produce a torn rail listing (27-07 must_haves, verification: backstop)."
    test: "Force two overlapping BrowseCatalogs calls (e.g. rapid external file churn while a directory switch is in flight) and confirm the rail never shows a mixed/partial listing."
    expected: "Last-completed listing wins cleanly; no interleaved or truncated row set."
    why_human: "Declared backstop — reasoned from the reducer's replace-wholesale SET_CATALOGS semantics; this project has no frontend test framework (TEST-01 deferred), so the race is not proven mechanically."
  - truth: "With watching disabled (the shipped default), the rail visibly reflects an in-app Delete or Duplicate action without requiring the user to navigate away and back."
    test: "With Settings → watch directory OFF (the default), delete a catalog via the actions menu, then separately duplicate one. Observe the rail immediately after each action."
    expected: "Deleted catalog's row disappears from the rail; duplicated catalog's new row appears in the rail — both without needing an external watch event."
    why_human: "Not observed to happen: DetailsPanel's onDeleted only clears `currentCatalogId`/`selected` when the deleted catalog was selected — it never removes the row from `state.catalogs`. duplicateCatalogAction never dispatches a new catalog into `state.catalogs` either. Both rely exclusively on the `catalogs:changed`-driven re-list the file watcher establishes (explicitly locked in 27-CONTEXT.md/27-UI-SPEC.md as the one refresh path), which only fires when watching is on. Watching defaults to `false` (`internal/config/config.go` `WatchDirectory: false`). This is a deliberate, documented design decision — not an oversight — but its practical effect is that the majority of users (default settings) will see a just-deleted catalog remain listed, and a just-duplicated catalog stay invisible, until they switch directories or otherwise force a re-list. Flagging for a human product decision on whether this meets the phase goal's implicit expectation that the rail reflects the user's own actions, distinct from the explicitly-scoped WATCH-02 external-change clause."
---

# Phase 27: Catalog Actions + Watch Verification Report

**Phase Goal:** Users manage existing catalogs — rename, duplicate, delete to Trash — and see the rail stay current when catalogs change outside the app
**Verified:** 2026-08-16T15:10:00Z
**Status:** human_needed
**Re-verification:** No — initial verification

## Goal Achievement

### Observable Truths (by requirement)

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | ACT-01: User can open a catalog actions menu from `⋯` in the details panel, navigate with arrows, close via Escape/click-outside, focus restores to the trigger | ✓ VERIFIED | `Menu.tsx` implements roving tabIndex, wraparound, `role="menu"`, click-outside listener excluding the trigger; live matrix rows 1-6 in 27-07-SUMMARY.md all "Match" after the WR-01/CR-01 focus-restore fix (WINDOWS.md #13, status: fixed, re-verified live with CDP-trusted mouse events) |
| 2 | ACT-02: User can rename a catalog's title; JSON `title` field authoritative, both HTML `<title>`/`<h1>` sites rewritten with matching escaping, filenames unchanged | ✓ VERIFIED | `internal/catalog/rename.go` read directly — order-preserving JSON root rewrite (`setRootStringField`), dual HTML rewrite (`rewriteHTMLTitle`); `TestRenameCatalog_*` (12 subtests) all pass; live round-trip in 27-01-SUMMARY.md and 27-07-SUMMARY.md matrix rows 7-10; write-ordering bug (JSON written before HTML validated) found and fixed by code review (`TestRenameCatalog_RejectedHTMLStepLeavesJSONTitleUnchanged` passes, confirmed by me directly) |
| 3 | ACT-03: User can duplicate a catalog, copying `.json` and any `.html` with a `-copy`/`-copy-N` suffixed root, byte-identical | ✓ VERIFIED | `internal/catalog/duplicate.go` — `nextCopyRoot` checks both extensions before taking a candidate, both writes go through `WriteFileAtomic`; `TestDuplicateCatalog_*` (10 subtests) pass; live matrix rows 11-12 "Match" |
| 4 | ACT-04: User can delete a catalog to the OS Trash after a confirmation naming both file paths in full, with an optional `.html` checkbox | ✓ VERIFIED | `DeleteConfirmDialog.tsx` — full-path boxes, checkbox present only when `catalog.hasHtml`, correct primary-button label logic; live matrix rows 13-16 "Match", including a real file moved to `~/.Trash` and confirmed via `ls` |
| 5 | ACT-05: A failed Trash operation surfaces as an error and never silently falls back to permanent deletion | ✓ VERIFIED | `internal/osutil/trash.go` — no local removal call anywhere, `trashSeam` is the only deletion mechanism, error wrapped and returned verbatim; live matrix rows 17-18 — a real induced failure (`chflags uchg`) produced the verbatim OS error with only "Keep catalog"/"Try moving to Trash again" in the footer |
| 6 | ACT-09: No catalog write can corrupt an existing catalog file if the app crashes mid-write | ✓ VERIFIED | `internal/catalog/atomicwrite.go` — `tmp.Sync()` before close, best-effort `syncDir` after rename; `TestWriteFileAtomic_SurvivesKill`/`_NoPriorFile` — I ran these myself: 21+21 real `SIGKILL`ed subprocess iterations, all byte-identical post-write (SHA-256 verified) |
| 7 | WATCH-01: User sees `● watching <catalog directory>` in the status bar when watching is enabled and a directory is set | ✓ VERIFIED | `StatusBar.tsx`/`workspace.css` — span (not button), `var(--fn)`, 160px ellipsis, no `aria-live`/`title` (confirmed via grep — neither attribute present); live matrix rows 20-22 "Match" including computed-style color check |
| 8 | WATCH-02: User sees the rail update when catalogs are added/removed/modified outside the app | ✓ VERIFIED | `CatalogRail.tsx` subscribes to `catalogs:changed` via `EventsOn`, returns unsubscribe from the effect; `internal/watch/watcher.go`'s debounced coalescer + `TestWatcher_*`/`TestCoalescer_*` (13 tests) all pass; live matrix rows 23-25 "Match" — external `cp`/`rm` reflected within ~1s, a 10-file burst produced exactly one `BrowseCatalogs` call |
| 9 | WATCH-03: User can turn watching off in Settings, and the watcher is genuinely released | ⚠️ PRESENT_BEHAVIOR_UNVERIFIED | Toggle-off portion is proven: `TestWatcher_Close`/`_CloseIsIdempotent` pass (I ran these), plus live `lsof` fd-count evidence (101→74, watched DIR fd released) in 27-07-SUMMARY.md row 26. The **app-quit** portion (`main.go`'s `OnShutdown: app.shutdown`, confirmed wired at `main.go:76` and `app.go:823-830`) was **not exercised as a real quit** — matrix row 27 explicitly substituted the toggle-off path's evidence because a real quit would have killed the shared `wails dev` session. Code inspection confirms `shutdown()` calls the identical `Close()`, but this is reasoning, not a direct observation of that code path executing. Routed to Human Verification. |

**Score:** 8/9 requirement-level truths verified (1 present-behavior-unverified: WATCH-03's real-quit sub-claim)

Additionally, 5 must_haves truths across the 7 plans were explicitly declared `verification: backstop` (reasoned, not test-proven) at plan time — see `behavior_unverified_items` in the frontmatter and Human Verification below. These are not requirement-level truths and are not counted in the 8/9 score, but per the honest-verifier standard they cannot be silently marked VERIFIED either.

### Required Artifacts

| Artifact | Expected | Status | Details |
|----------|----------|--------|---------|
| `internal/catalog/rename.go` | RenameCatalog + dual HTML rewrite | ✓ VERIFIED | Exists, exports `RenameCatalog`; containment-checked sibling read fixed (WR-01/WR-03) |
| `internal/catalog/rename_test.go` | ACT-02 tests | ✓ VERIFIED | 12 subtests, all pass |
| `pkg/models/catalog.go` | `Title` omitempty field | ✓ VERIFIED | Present |
| `internal/search/service.go` | JSON-title precedence + `html.UnescapeString` | ✓ VERIFIED | `extractJSONTitle` present; `TestBrowseCatalogs_UnescapesHTMLTitle` etc. pass |
| `internal/catalog/atomicwrite.go` | fsync-hardened write | ✓ VERIFIED | `tmp.Sync()`, `syncDir`, `sync.Once`-guarded log (WR-02 fix) |
| `internal/catalog/atomicwrite_sigkill_test.go` | subprocess SIGKILL proof | ✓ VERIFIED | `TestWriteFileAtomic_SurvivesKill*` pass, run directly by me |
| `internal/catalog/testdata/killtarget/main.go` | standalone kill-target helper | ✓ VERIFIED | Present, no project imports (WINDOWS.md #14 records the by-convention-only coupling, correctly flagged as a residual, not a missing mitigation) |
| `internal/catalog/duplicate.go` | DuplicateCatalog | ✓ VERIFIED | Exports `DuplicateCatalog`, `-copy`/`-copy-N` collision loop |
| `internal/osutil/trash.go` | containment-gated Trash helper | ✓ VERIFIED | Exports `TrashPaths`, swappable `trashSeam`, no local removal |
| `internal/osutil/trash_test.go` | ACT-04/ACT-05 tests, no real Trash touched | ✓ VERIFIED | 11 subtests pass, seam mocked |
| `go.mod` | wastebasket v2.0.3 + fsnotify v1.10.1 pinned | ✓ VERIFIED | Both present |
| `internal/watch/watcher.go` | fsnotify wrapper, debounce, error-drain, idempotent Close | ✓ VERIFIED | Reviewed directly; matches every must_have truth |
| `internal/watch/watcher_test.go` | WATCH-02/03 tests | ✓ VERIFIED | 13 tests pass, incl. `TestWatcher_Close` run directly |
| `main.go` | OnShutdown hook | ✓ VERIFIED | `OnShutdown: app.shutdown` wired alongside existing hooks |
| `frontend/src/components/workspace/Menu.tsx` | anchored popover primitive | ✓ VERIFIED | Conditionally mounted, roving focus, click-outside, position: fixed via CSS |
| `frontend/src/components/workspace/DialogShell.tsx` | shared 440px dialog shell | ✓ VERIFIED | One implementation, used by both RenameDialog and DeleteConfirmDialog |
| `frontend/src/components/workspace/RenameDialog.tsx` | ACT-02 UI surface | ✓ VERIFIED | Pre-filled/selected, Enter commits, disabled-on-empty, error banner format matches spec |
| `frontend/src/components/workspace/DeleteConfirmDialog.tsx` | ACT-04/ACT-05 UI surface | ✓ VERIFIED | Confirm + error sub-states, no permanence vocabulary, full unellipsized paths |
| `frontend/src/workspace.css` | `--danger`/`--ondanger`/`--z-menu` + all new classes | ✓ VERIFIED | All tokens and classes present and correctly valued |
| `frontend/src/components/workspace/StatusBar.tsx` | WATCH-01 watching segment | ✓ VERIFIED | `ws-status-watching` present, correctly conditioned |
| `frontend/src/components/workspace/CatalogRail.tsx` | WATCH-02 subscription | ✓ VERIFIED | `catalogs:changed` subscribed with unsubscribe cleanup |
| `.planning/WINDOWS.md` | phase's ledger entries | ✓ VERIFIED | 7 entries added this phase (#8-14): 6 open (cross-platform sweeps + SIGKILL-harness residual), 1 fixed (#13, focus-restore) |

### Key Link Verification

| From | To | Via | Status | Details |
|------|----|----|--------|---------|
| `app.go` (`App.RenameCatalog`) | `internal/catalog.RenameCatalog` | containment gate then call | ✓ WIRED | `app.go:903-936` — `osutil.ContainsPath` before `catalog.RenameCatalog` |
| `app.go` (`App.DuplicateCatalog`) | `internal/catalog.DuplicateCatalog` | containment gate then call | ✓ WIRED | `app.go:941-980` |
| `app.go` (`App.DeleteCatalog`) | `internal/osutil.TrashPaths` | containment gate then call | ✓ WIRED | `app.go:993-1030` |
| `frontend/src/services/wailsAPI.ts` | generated `App.d.ts`/`App.js` bindings | `renameCatalog`/`duplicateCatalog`/`deleteCatalog` wrappers | ✓ WIRED | All three exported in `App.d.ts`, wrapped in `wailsAPI.ts` |
| `frontend/.../DetailsPanel.tsx` | `Menu.tsx` / `RenameDialog.tsx` / `DeleteConfirmDialog.tsx` | conditional mount + prop wiring | ✓ WIRED | `CatalogActions` wires all three, three-item menu in the correct order with `dividerBefore` on Delete |
| `app.go` (`applyWatchState`) | `internal/watch.New` | closure emitting `catalogs:changed` | ✓ WIRED | `app.go:637-659`, called from `startup`, `SetWatchDirectory`, `SetCatalogDirectory` |
| `main.go` | `app.go` (`app.shutdown`) | `OnShutdown` option | ✓ WIRED | `main.go:76`; `app.shutdown` closes the watcher unconditionally |
| `frontend/.../CatalogRail.tsx` | `app.go` (`emitCatalogsChanged`) | `EventsOn('catalogs:changed')` | ✓ WIRED | Confirmed both emit side (`app.go:618`) and subscribe side |
| `internal/watch` package | Wails runtime | **absence** of import | ✓ VERIFIED | `grep` for `wails/v2/pkg/runtime` in `internal/watch/*.go` returns nothing — package stays CLI-usable (COMPAT-04) |

### Data-Flow Trace (Level 4)

| Artifact | Data Variable | Source | Produces Real Data | Status |
|----------|---------------|--------|---------------------|--------|
| Rail rows after external filesystem change | `state.catalogs` | `catalogs:changed` event → `loadCatalogsForDirectory` → real `BrowseCatalogs` Wails call | Yes | ✓ FLOWING |
| Rail row title after rename | `state.catalogs[i].title` | Optimistic `dispatch(SET_CATALOGS, ...)` in `DetailsPanel.tsx` using the dialog's own submitted value, matching what was just durably written to disk | Yes | ✓ FLOWING |
| Status-bar watching segment | `state.catalogDir` | `AppContext` (resolved synchronously) | Yes | ✓ FLOWING |
| Rail rows after an in-app Delete (watching off, the default) | `state.catalogs` | **No dispatch removes the row** — relies solely on `catalogs:changed`, which does not fire when watching is off | No | ⚠️ STATIC — see behavior_unverified_items and Human Verification |
| Rail rows after an in-app Duplicate (watching off, the default) | `state.catalogs` | **No dispatch adds the row** — same gap | No | ⚠️ STATIC — see behavior_unverified_items and Human Verification |

### Behavioral Spot-Checks (run directly by me)

| Behavior | Command | Result | Status |
|----------|---------|--------|--------|
| Full workspace test suite | `go build ./... && go vet ./... && go test ./... -race -count=1` | All 9 Go packages pass, build/vet clean | ✓ PASS |
| Frontend typecheck + build | `cd frontend && npx tsc --noEmit && npm run build` | Both clean, `dist/` produced | ✓ PASS |
| Real SIGKILL crash-safety proof | `go test ./internal/catalog/... -run TestWriteFileAtomic_SurvivesKill -v -count=1 -timeout 120s` | `PASS` — 21+21 iterations, byte-identical | ✓ PASS |
| Watcher idempotent Close | `go test ./internal/watch/... -run TestWatcher_Close -v -count=1` | `PASS` | ✓ PASS |
| Named test enumeration for every claimed test | `go test ./internal/catalog/... -list '.*'`, same for `osutil`, `watch`, `search` | Every test name cited in the 7 SUMMARYs exists | ✓ PASS |
| Regression test for the write-ordering fix | `go test ./internal/catalog/... -run TestRenameCatalog_RejectedHTMLStepLeavesJSONTitleUnchanged -v -count=1` | `PASS` (2 subtests) | ✓ PASS |

### Probe Execution

No `scripts/*/tests/probe-*.sh` files or PLAN/SUMMARY probe references found in this phase — SKIPPED (no probes declared).

### Requirements Coverage

| Requirement | Source Plan(s) | Description | Status | Evidence |
|-------------|-----------------|-------------|--------|----------|
| ACT-01 | 27-04 | Actions menu from `⋯` | ✓ SATISFIED | Truth #1 above |
| ACT-02 | 27-01, 27-04 | Rename title | ✓ SATISFIED | Truth #2 above |
| ACT-03 | 27-03, 27-05 | Duplicate | ✓ SATISFIED | Truth #3 above |
| ACT-04 | 27-03, 27-05 | Delete to Trash with confirmation | ✓ SATISFIED | Truth #4 above |
| ACT-05 | 27-03, 27-05 | Never silently fall back to permanent delete | ✓ SATISFIED | Truth #5 above |
| ACT-09 | 27-02 | Crash-safe writes | ✓ SATISFIED | Truth #6 above |
| WATCH-01 | 27-07 | Status-bar watching indicator | ✓ SATISFIED | Truth #7 above |
| WATCH-02 | 27-06, 27-07 | Rail updates on external change | ✓ SATISFIED | Truth #8 above |
| WATCH-03 | 27-06, 27-07 | Watcher released on toggle-off and quit | ? NEEDS HUMAN | Truth #9 above — toggle-off proven, quit path reasoned only |

**No orphaned requirements.** Cross-referenced against `REQUIREMENTS.md`'s traceability table: ACT-01 through ACT-05, ACT-09, WATCH-01 through WATCH-03 are all marked "Phase 27 / Complete" there, and the union of every plan's `requirements:` frontmatter field across the 7 plans (`27-01`: ACT-02; `27-02`: ACT-09; `27-03`: ACT-03/04/05; `27-04`: ACT-01/02; `27-05`: ACT-03/04/05; `27-06`: WATCH-02/03; `27-07`: WATCH-01/02/03) exactly matches the 9 requirement IDs given for this verification, with no gaps and no extras. ACT-06/07/08 and STATE-03 remain correctly deferred to Phase 28 per `REQUIREMENTS.md`.

### Anti-Patterns Found

None. Scanned every file this phase created/modified for `TBD`/`FIXME`/`XXX`/`TODO`/`HACK`/`PLACEHOLDER`/empty-implementation patterns — zero unresolved debt markers. The one match (a comment in `StatusBar.tsx` containing the word "placeholder" in prose, describing what is *not* rendered) is documentation, not a stub.

### Gates Already Run (not re-litigated here, per verification context)

- **Code review:** 3 iterations + targeted regression pass, `status: all_fixed`. I independently confirmed the fixes are present in the current tree: `Menu.tsx`'s scoped `preventDefault()` (line 91-96), `RenameCatalog`'s sibling-before-write ordering (`rename.go:43-87`) and its distinguishing partial-failure error message, the watcher's `go w.c.fireNow()` off-loop dispatch, and the `sync.Once`-guarded directory-fsync log.
- **Security:** `threats_open: 0`, 26/26 closed (confirmed by reading `27-SECURITY.md` frontmatter directly).
- **Regression gate:** all 9 Go packages pass under `-race` — I ran this myself, not just trusted the SUMMARY.
- **WINDOWS.md ledger:** 7 entries opened this phase (#8-14), 6 open (all cross-platform-unverified or the SIGKILL-harness-by-convention residual, correctly classified as deviations, not defects), 1 fixed (#13, focus-restore, re-verified live post-fix).

## Human Verification Required

### 1. Real app-quit watcher release (WATCH-03)

**Test:** With watching enabled, quit the actual built StorCat app (not `wails dev`) and relaunch it.
**Expected:** Clean relaunch; no leaked watch handle from the prior session (verifiable via `lsof`/Activity Monitor on macOS, or platform equivalent).
**Why human:** 27-07's own live matrix explicitly substituted evidence (identical `Close()` call path, `lsof` fd-count before/after toggle-off) rather than risking the shared `wails dev` session by actually quitting it. `main.go`'s `OnShutdown: app.shutdown` wiring is confirmed present by code inspection, but the actual quit path has not been observed executing.

### 2. Concurrent DeleteCatalog dedup (backstop, 27-01/27-03)

**Test:** Fire two DeleteCatalog calls for the same catalog's paths at effectively the same time.
**Expected:** No error; the file is trashed exactly once; the second call no-ops cleanly.
**Why human:** Declared `verification: backstop` in the plan — reasoned from `TrashPaths`'s tested not-exist skip, not exercised under genuine concurrency.

### 3. Menu never overflows the viewport (backstop, 27-04)

**Test:** Open the catalog-actions menu at the app's smallest supported window sizes.
**Expected:** Menu always fully visible, never clipped off-screen.
**Why human:** Declared backstop — reasoned from layout geometry, no automated boundary sweep exists (no frontend test framework, by design).

### 4. Delete dialog's busy-state visual (backstop, 27-05)

**Test:** Trigger a real (slow) delete and observe the in-flight state.
**Expected:** Both footer buttons disabled, primary at opacity 0.7, no spinner.
**Why human:** Declared backstop — pattern reused by reading Footer's implementation, not independently exercised live against this dialog's own async call.

### 5. catalogs:changed racing an in-flight browseCatalogs call (backstop, 27-07)

**Test:** Force overlapping BrowseCatalogs calls (rapid external file churn during a directory switch).
**Expected:** No torn/interleaved rail listing.
**Why human:** Declared backstop — reasoned from reducer semantics, not proven mechanically (no frontend test framework).

### 6. Rail freshness after in-app Delete/Duplicate with watching OFF (the shipped default) — product judgment requested

**Test:** With Settings → watch directory OFF (the default), delete a catalog via the actions menu, then duplicate one.
**Expected (per current implementation):** Neither action updates the rail — the deleted row stays listed, the duplicated row doesn't appear — until the user navigates away and back or an external watch event fires.
**Why human:** This is not a bug in the sense of contradicting the plan's own must-haves — 27-CONTEXT.md and 27-UI-SPEC.md explicitly lock "reuse the catalogs:changed-driven refresh, no bespoke second path" as the design, and the phase's own Phase Boundary sentence scopes "rail stays current" specifically to "when catalogs change **outside** the app." But since watching defaults to off, most users performing the very actions this phase ships (delete, duplicate) will not see their own action reflected in the rail without extra navigation. Flagging for a human decision: is this acceptable for v3.0.0, or does it warrant a small follow-up (e.g., have Delete/Duplicate directly patch `state.catalogs` the way Rename already does optimistically)?

## Gaps Summary

No blocking gaps. Every artifact this phase's 7 plans declared exists, is substantive, and is wired; every key link resolves to real code; the full Go test suite (including two SIGKILL-based crash-safety tests I ran directly) passes under `-race`; frontend typecheck and build are clean; all 9 requirement IDs (ACT-01 through ACT-05, ACT-09, WATCH-01 through WATCH-03) are satisfied with no orphans against REQUIREMENTS.md. Code review (3 iterations) and security review (26/26 threats closed) already ran and I independently confirmed their fixes are present in the current tree rather than re-litigating them.

The phase routes to `human_needed` rather than `passed` because of six items that genuinely cannot be settled by static analysis: one requirement (WATCH-03) whose real-quit sub-claim rests on reasoned-equivalence rather than direct observation, four explicitly-declared `backstop` must-haves the plans themselves flagged as reasoned-not-tested, and one product-level judgment call about whether the rail should reflect a user's own Delete/Duplicate action even when watching is off (the shipped default) — a real, observable, and honestly-documented consequence of a deliberate design decision, not a hidden defect.

---

*Verified: 2026-08-16T15:10:00Z*
*Verifier: Claude (gsd-verifier)*
