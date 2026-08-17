---
status: issues
phase: 27-catalog-actions-watch
source: [27-VERIFICATION.md]
started: 2026-08-16
updated: 2026-08-16
---

## Current Test

number: 7
name: Windows / Linux platform behavior (WINDOWS.md 8-11)
awaiting: none — all scenarios resolved (one blocked, one failed, five passed)

## Tests

### 1. Watcher released on a real app quit (WATCH-03)
expected: With watching enabled, quit the built app and relaunch. No leaked watch handle, no crash, no hang. A fresh `lsof -p <pid>` shows no dangling watch descriptor from the prior session.
why_human: 27-07's live matrix (row 27) deliberately did not exercise a real quit — doing so would have killed the shared `wails dev` session the other 27 rows depended on. Substitute evidence (`lsof` fd count 101→74 on the mechanistically identical `Close()` from the toggle-off path) is reasoned, not a direct observation of the `OnShutdown` path.
result: [pass] Live-tested against the real built app process (`/Users/ken/dev/storcat/build/bin/StorCat.app/Contents/MacOS/StorCat`, PID 45451, a genuine child process `wails dev` spawns and dev-browser's :34115 session binds to — not a browser tab). Enabled watching on a real fixture directory; `lsof -p 45451` confirmed a real open `DIR` fd (fd 40) on the watched directory before quitting. Issued `window.runtime.Quit()` (the same real quit path the app's own Cmd+Q / window-close ultimately invoke; the project's no-host-OS-GUI-automation rule forbids driving that via keystroke injection, so the runtime call is the closest available live exercise of `main.go`'s `OnShutdown: app.shutdown` hook). Result: PID 45451 (and its parent `wails dev` 45029 and `npm run dev` 45297) all fully exited within 2 seconds — `ps` confirmed no processes remained, and the dev-server log recorded a clean `"Development mode exited"` line with no panic or stack trace. Relaunched `wails dev` from scratch: it started cleanly with no port-34115 conflict, produced a genuinely new app process (PID 54254), and — because `watchDirectory: true` was still the persisted config from before the quit — the fresh process re-established watching entirely on its own at startup with its own new fd (fd 17), independent of and unrelated to the old process's now-fully-reclaimed fd 40. No dangling watch descriptor is possible once a process has fully exited, and the exit was graceful (not a timeout/force-kill), which is the direct evidence the `OnShutdown` cleanup path ran rather than the OS merely reclaiming resources after a crash.

### 2. Concurrent DeleteCatalog calls for the same catalog (backstop)
expected: Two concurrent deletes of the same catalog cannot both report success while leaving a file behind.
why_human: Declared `verification: backstop` by the plans themselves — reasoned, not tested. Racing two real Wails binding calls needs a harness the phase did not build.
result: [pass] Live-tested with genuine concurrency: `Promise.allSettled([App.DeleteCatalog(...), App.DeleteCatalog(...)])` fired both calls for the same catalog's real `.json`/`.html` paths at the same microtask tick, dispatched to the real Go binding (not mocked). Both calls resolved `fulfilled` (no error from either), and the filesystem ended up with exactly one outcome: both files present in `~/.Trash`, none left behind in the catalog directory. Matches `TrashPaths`'s documented not-exist-skip behavior extended to genuine concurrent callers, not just the two already-unit-tested sequential cases.

### 3. Menu does not overflow the viewport (backstop)
expected: The actions menu stays within the viewport when the `⋯` trigger is near a window edge.
why_human: Declared `verification: backstop`. Needs a resized/edge-positioned window, which the dev-browser pass did not cover.
result: [pass] Live-tested at two window sizes/layout variants, both smaller than any prior live check in this phase: (1) 1280×700 (the minimum wide-layout width at a short height) using the `pane` `DetailsPanel` variant — menu bounding box `top:88 bottom:225` against `innerHeight:700`, no overflow on any edge. (2) 1040×700 (narrow-layout, `drawer` `DetailsPanel` variant opened via the toolbar "Details" chip) — menu bounding box `bottom:225` against `innerHeight:700`, again no overflow. Both real `getBoundingClientRect()` measurements taken against the actually-rendered `.ws-menu` element, not computed/estimated. Confirms the backstop's "no realistic path to overflowing" reasoning holds with a wide safety margin (menu height ~137px against 700px available) at both supported layout variants.

### 4. Delete dialog busy-state visual treatment (backstop)
expected: While a delete is in flight, both footer buttons are visibly disabled and a second submit cannot fire.
why_human: Declared `verification: backstop`. The `busy` guard and `disabled={busy}` are code-verified; the visual treatment during the real in-flight window was not observed.
result: [pass] Live-tested by stubbing `window.go.main.App.DeleteCatalog` in-page to delay 2.5s before delegating to the real binding (the same "stub the binding, not the app logic" technique authorized elsewhere in this task, used here only to widen an otherwise sub-second window enough to sample it — the delete itself still genuinely executes and genuinely trashes the file). Opened the real delete confirmation dialog, clicked the real "Move to Trash" button, and sampled DOM state ~400ms into the (real) in-flight call: both `.ws-dialog-btn` (Keep catalog) and `.ws-dialog-btn-danger` (Move to Trash) had `disabled: true`, and the danger button's computed `opacity` was exactly `"0.7"`. Dialog remained open throughout, then closed cleanly once the real delete completed (confirmed the catalog landed in `~/.Trash`). Native `disabled` attribute is what would prevent a second click's `onClick` from firing — confirmed present during the entire sampled window, not just asserted from reading the component.

### 5. `catalogs:changed` arriving during an in-flight rail fetch (backstop)
expected: A watch event landing mid-`browseCatalogs` does not produce a torn or stale listing.
why_human: Declared `verification: backstop`. Requires deterministic control over the event/fetch interleaving.
result: [fail] Live-tested by stubbing `window.go.main.App.BrowseCatalogs` to distinguish and control resolution order of two overlapping calls, then driving the exact real interleaving the scenario describes: (1) triggered a real directory-switch fetch through the rail's own "Change directory" chip (call #1, stubbed to return one distinguishable catalog and resolve after 900ms — the "old" fetch), then (2) ~250ms later, while call #1 was still in flight, emitted a real `window.runtime.EventsEmit('catalogs:changed')` — the identical event `app.go`'s real watcher emits, received by `CatalogRail.tsx`'s actual `EventsOn('catalogs:changed', ...)` subscription (call #2, stubbed to return two different distinguishable catalogs and resolve after 100ms — the "new," fresher fetch). Call #2 genuinely resolved first (~350ms) and dispatched `SET_CATALOGS` with the fresh data; call #1 resolved second (~900ms) and its `SET_CATALOGS` dispatch **overwrote** it. Final rendered rail showed only the **stale** result (`"OLD-1"`), discarding the fresher watch-triggered listing entirely. Root cause, confirmed by reading the code: `CatalogRail.tsx`'s `loadCatalogsForDirectory` (lines 20-32) has no request-ID/generation guard on its `wailsAPI.browseCatalogs(dir).then(...)` — unlike this same codebase's own `CommandPalette` and `VolumePicker`, which both carry an explicit `requestIdRef`-based stale-response guard for exactly this class of race. The listing is never **torn** (a mixed/interleaved row set) since `SET_CATALOGS` always fully replaces the array — but it genuinely **is stale**, contradicting the scenario's expected text verbatim ("does not produce a torn **or stale** listing"). This is a real, reproducible defect, not a false alarm from an unrealistic test: the exact trigger (a watch event firing while a slower directory-switch fetch is still in flight) is precisely the scenario the backstop was declared against. Reported here per this task's explicit instruction to record a genuine failure plainly and leave the code unchanged — **no fix was applied.**

### 6. RenameDialog failure sub-state, end to end
expected: A rename that fails (e.g. an unwritable `.html`) keeps the dialog open, shows the real error inline, and allows a retry.
why_human: Verified by direct binding rejection plus code-review parity with `Footer`'s proven pattern, not clicked through end-to-end in the dialog UI. Surfaced by the phase 27 UI review (priority fix 3).
result: [pass] Live-tested end-to-end through the actual dialog UI (not a direct binding call). A `chmod 444` on the sibling `.html` did **not** reproduce a failure — `RenameCatalog`'s atomic-write pattern (temp file + `rename()`) only needs directory write permission, not target-file write permission, so this was corrected to `chflags uchg` (the same real-failure-induction technique 27-07's own live matrix used for the ACT-05 Trash-failure row), which blocks the final rename regardless of permission bits. With the flag set: opened the rename dialog on a real catalog (pre-filled/selected title confirmed), typed a new title, clicked "Rename catalog" — the dialog stayed open, and a real, detailed inline error appeared (`.ws-dialog-error`, the genuine chained error string from `rename.go`, including "title updated in uat-beta.json but failed to update uat-beta.html: ... operation not permitted"). Retried with the same immutable flag still set: identical failure reproduced idempotently, dialog still open, submit button still enabled (retry genuinely possible, not stuck disabled). Cleared the flag and retried a third time: dialog closed, and both the JSON and HTML titles now matched on disk (confirmed via direct file read), completing the flow. One incidental observation, not a defect: this failure path leaves the JSON title updated while the HTML title is not (a genuine partial-write state during the failure window) — but this is `rename.go`'s own explicitly documented, accepted residual risk ("the pair is not a single atomic transaction," see the code comment above the HTML write), exactly the class of OS-level write failure a pre-validation readability check cannot catch, not a new or hidden gap. The dialog's own behavior — stay open, show the real error, allow retry — matches the scenario's expected text exactly.

### 7. Windows / Linux platform behavior (WINDOWS.md 8-11)
expected: fsnotify's Windows rename-release behavior, both new dependencies' Windows and Linux backends, and the parent-directory fsync's Windows discard path all behave as reasoned.
why_human: No Windows or Linux machine was available this session. Tracked as ledger entries 8-11 for the pre-v3.0.0 sweep, not claimed as covered.
result: [blocked] Cannot be exercised — this session ran entirely on macOS, with no Windows or Linux machine or VM available. `.planning/WINDOWS.md` entries #8-11 remain open exactly for this reason and are not claimed as covered here. Not fabricated; genuinely untested on those platforms.

## Summary

total: 7
passed: 5
issues: 1
pending: 0
skipped: 0
blocked: 1

## Gaps

- **Scenario 5 FAILED** — `catalogs:changed` arriving mid-`browseCatalogs` produces a stale (not torn, but stale) rail listing. `CatalogRail.tsx`'s `loadCatalogsForDirectory` (lines 20-32) has no stale-response/generation guard on its `browseCatalogs(dir).then(...)` chain, unlike `CommandPalette.tsx` and `VolumePicker.tsx` elsewhere in this same codebase, which both already carry a `requestIdRef`-based guard for this exact class of race. Reproduced live with genuine event/fetch interleaving (a real directory-switch fetch in flight, then a real `catalogs:changed` event landing before it resolves) — the earlier-issued, later-resolving fetch overwrites the fresher one. No fix applied per this task's instructions; left for a human decision on remediation (the fix shape is well-precedented in this same codebase — mirror the existing `requestIdRef` pattern onto `loadCatalogsForDirectory`).
- **Scenario 7 BLOCKED** — no Windows or Linux machine available this session. Tracked in `.planning/WINDOWS.md` entries #8-11, unresolved by design pending a genuine cross-platform sweep.
