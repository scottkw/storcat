---
phase: 28-re-scan-diff
plan: 05
subsystem: catalog-scan
tags: [react, typescript, rescan, delete, unreadable-panel]

requires:
  - phase: 28-re-scan-diff
    plan: 01
    provides: App.RescanCatalog, wailsAPI.rescanCatalog, RescanDialog shell, state.rescan slice
  - phase: 28-re-scan-diff
    plan: 03
    provides: RescanDialog Step 3 Variant B (reduced, no-diff summary)
  - phase: 28-re-scan-diff
    plan: 04
    provides: ResolveRescan binding, Overwrite/Keep-both resolution footer
provides:
  - "UnreadableCatalogPanel's action trio -- Re-scan volume / Open the .html instead / Remove from library, all three routed to already-shipped surfaces"
  - "Live-verified end-to-end: RescanDialog Variant B reachable from a genuinely unreadable catalog, the existing Phase 27 DeleteConfirmDialog reused unchanged, handleOpenHtml logic reused verbatim"
affects: [28-06]

actuals:
  tokens: 1998
  tasks: 2
  commits: 1

tech-stack:
  added: []
  patterns:
    - "UnreadableCatalogPanel's three new useState hooks (rescanOpen/deleteOpen/openBusy/actionError) are declared unconditionally ahead of the component's pre-existing early return (rawError === '' -> null), preserving hook-call-order safety without restructuring the early-return itself"
    - "handleOpenHtml is a second, independent copy of DetailsPanel Footer's exact call pair (getCatalogHtmlPath + openExternal) rather than an extracted shared function -- the plan's own acceptance criteria required zero DetailsPanel.tsx diff, and the two call sites are already small enough that extraction would be a speculative abstraction for two call sites, not three"

key-files:
  created: []
  modified:
    - frontend/src/components/workspace/UnreadableCatalogPanel.tsx
    - frontend/src/workspace.css

key-decisions:
  - "The action-trio row (and the RescanDialog/DeleteConfirmDialog it opens) only renders when `catalog` resolves from state.catalogs -- the panel's pre-existing rawError computation already tolerates an unresolved catalog (falls back to state.tree's error message), but none of the three actions have a meaningful target without a resolved CatalogMetadata, so they're gated on it rather than passing partial/undefined data into RescanDialog's typed props."

requirements-completed: [STATE-03]

coverage:
  - id: D1
    description: "A user looking at an unreadable catalog can re-scan its source volume, open its .html instead, or remove it from the library -- three actions, from the panel itself"
    requirement: "STATE-03"
    verification:
      - kind: e2e
        ref: "dev-browser live session against wails dev :34115 -- selected corrupt-1.json (unparseable, .html present), DOM query on .ws-unreadable-actions returned exactly 3 buttons: Re-scan volume / Open the .html instead / Remove from library, all enabled"
        status: pass
    human_judgment: false
  - id: D2
    description: "Re-scanning an unreadable catalog is a FRESH scan (Variant B): no old tree, reduced 'Scan complete' summary, no stat grid, no diff list, overwrite/keep-both only"
    requirement: "STATE-03"
    verification:
      - kind: e2e
        ref: "dev-browser live session -- clicked Re-scan volume on corrupt-2.json, stubbed SelectDirectory, ran a real scan; diff step showed headerTitle 'Scan complete', stepLabel 'step 3 of 3', hasStatGrid:false, hasDiffList:false, subline '2 files - 12B - scanned .../source just now', footer buttons exactly ['Overwrite catalog','Keep both'] plus 'Discard scan and close', no similarity banner"
        status: pass
    human_judgment: false
  - id: D3
    description: "Keep both from the unreadable path leaves the broken original exactly as unreadable as it was"
    requirement: "STATE-03"
    verification:
      - kind: e2e
        ref: "dev-browser live session -- clicked Keep both on corrupt-2.json's Variant B diff step; corrupt-2.json remained byte-for-byte the same truncated content (still raises json.JSONDecodeError: Unterminated string) after the click, while corrupt-2-copy.json/.html were written and parse cleanly"
        status: pass
    human_judgment: false
  - id: D4
    description: "Overwrite from the unreadable path rebuilds the catalog in place so its JSON parses again"
    requirement: null
    verification:
      - kind: e2e
        ref: "dev-browser live session -- clicked Overwrite catalog on corrupt-3.json's Variant B diff step; corrupt-3.json parsed successfully afterward (previously raised the same JSONDecodeError)"
        status: pass
    human_judgment: false
  - id: D5
    description: "'Open the .html instead' reuses the details-panel footer's existing open-HTML logic verbatim and is omitted entirely (not greyed) when the catalog has no .html"
    requirement: "STATE-03"
    verification:
      - kind: e2e
        ref: "dev-browser live session -- wrapped GetCatalogHtmlPath/OpenExternal to record calls while still invoking the real bindings: clicking the button on corrupt-1.json resolved GetCatalogHtmlPath to corrupt-1.html and OpenExternal returned no error (BrowserOpenURL succeeded); on corrupt-nohtml.json (no .html companion) the .ws-unreadable-actions button list contained only ['Re-scan volume','Remove from library'] -- the HTML button absent from the DOM entirely, not disabled"
        status: pass
    human_judgment: false
  - id: D6
    description: "'Remove from library' opens the EXISTING Phase 27 delete confirmation and routes to the existing delete-to-Trash binding -- no bespoke dialog, no hidden/excluded state"
    requirement: "STATE-03"
    verification:
      - kind: e2e
        ref: "dev-browser live session -- clicked Remove from library on corrupt-4.json; the opened dialog's #ws-delete-title read 'Delete catalog' (DeleteConfirmDialog's own title, unchanged), listing both the .json and .html rows with 'Move both to Trash' as the primary button -- clicked it, corrupt-4.json/.html disappeared from the rail's row listing and were found at ~/.Trash/corrupt-4.json and ~/.Trash/corrupt-4.html (recoverable, then cleaned up as scratch test artifacts)"
        status: pass
    human_judgment: false
  - id: D7
    description: "The 'Re-scan volume' button dims with the locked already-scanning tooltip while a Create scan is running; buttons 2 and 3 are unaffected"
    requirement: null
    verification:
      - kind: e2e
        ref: "dev-browser live session -- started a real Create scan (3000-file bigsource fixture) via .ws-new-pill while corrupt-1.json's unreadable panel was the active selection; mid-scan DOM query showed button 1 at disabled:true/aria-disabled:'true'/title:'A scan is already running — open it from the status bar.', while buttons 2 and 3 stayed disabled:false/aria-disabled:null/title:null"
        status: pass
    human_judgment: false

duration: ~40min
completed: 2026-08-16
status: complete
---

# Phase 28 Plan 05: Unreadable-Catalog Action Trio Summary

**`UnreadableCatalogPanel`'s stub comment replaced with three buttons -- Re-scan volume (RescanDialog Variant B), Open the .html instead (reused open-HTML logic), Remove from library (the existing Phase 27 delete-to-Trash dialog) -- all live-verified end-to-end against real corrupted catalog files via dev-browser**

## Performance

- **Duration:** ~40 min (implementation + full live checkpoint verification, run inline by the executor per this repo's CLAUDE.md dev-browser mandate)
- **Started:** 2026-08-16T22:00:00Z (approx)
- **Completed:** 2026-08-16T22:30:00Z (approx)
- **Tasks:** 2 (Task 1 auto, Task 2 a blocking checkpoint resolved live in this session)
- **Files modified:** 2

## Accomplishments

- `UnreadableCatalogPanel.tsx`'s stub comment (lines 111-113) replaced with a `.ws-unreadable-actions` row of three buttons:
  - **"Re-scan volume"** (no ellipsis, locked) -- opens `RescanDialog` at the pick-volume step with `oldTreeAvailable={false}`, so step 3 renders the reduced Variant B (no old tree to diff against). Dims with the shared `aria-disabled`/locked-tooltip guard while a Create scan is running elsewhere; buttons 2 and 3 are unaffected.
  - **"Open the .html instead"** -- a second call site of the exact `getCatalogHtmlPath` + `openExternal` pair `DetailsPanel`'s `Footer.handleOpenHtml` already uses. Omitted entirely (not greyed) when `catalog.hasHtml` is false.
  - **"Remove from library"** (no ellipsis, locked, text `--danger`) -- opens the existing, unmodified Phase 27 `DeleteConfirmDialog` passing this catalog directly, routing to the existing delete-to-Trash binding. No new dialog, no new binding, no hidden/excluded/archived membership concept.
- `frontend/src/workspace.css` gained `.ws-unreadable-actions`/`.ws-unreadable-action`/`.ws-unreadable-action-primary`/`.ws-unreadable-action-danger` -- no new CSS custom properties, matching the geometry and colors `28-UI-SPEC.md`'s Entry Points table 3 specifies exactly.
- **Zero Go code, zero new bindings.** `git diff HEAD -- app.go` and `git diff HEAD -- frontend/src/components/workspace/DeleteConfirmDialog.tsx` are both empty.
- **Live-verified end to end** (see below): all three buttons render correctly for a genuinely-unreadable catalog, the HTML button's absence is a real DOM omission (not a hidden/disabled element) on a catalog with no `.html`, Re-scan reaches the real Variant B UI and both its resolutions (Keep both / Overwrite) behave exactly per the UI-SPEC's guarantees, Remove from library opens the real, unmodified Phase 27 dialog and moves files to the real OS Trash, and the concurrent-scan guard dims only button 1.

## Task Commits

1. **Task 1: The unreadable-catalog action trio** - `241a0710` (feat)

_Task 2 (the blocking checkpoint) required no code changes -- it was a live verification pass, run directly against `wails dev` via dev-browser per this repo's CLAUDE.md mandate to test browser-based UAT automatically rather than asking the user to verify manually._

## Files Created/Modified

- `frontend/src/components/workspace/UnreadableCatalogPanel.tsx` - action trio, `handleOpenHtml`, `RescanDialog`/`DeleteConfirmDialog` wiring
- `frontend/src/workspace.css` - `.ws-unreadable-action*` class family

## Decisions Made

- **The action trio (and the two dialogs it opens) only render when `catalog` resolves from `state.catalogs`.** The panel's pre-existing `rawError` computation tolerates an unresolved `catalog` (falling back to `state.tree`'s error message for the rare case a catalog isn't in the rail's cache yet), but none of the three actions have a well-typed target without a resolved `CatalogMetadata` -- `RescanDialog`/`DeleteConfirmDialog` both require it as a required prop. Gating the whole row on `catalog &&` avoids passing partial data into typed props, at the cost of the trio not rendering in that rare edge case (which was already true implicitly, since `catalog?.filename` in the meta rows already degrades to `'—'` there).
- **`handleOpenHtml` is a second, independent copy of `Footer`'s exact call pair, not an extracted shared function.** The plan's own acceptance criteria required `DetailsPanel.tsx` to stay byte-for-byte unmodified, and extracting a shared helper for two four-line call sites would be a speculative abstraction ahead of a third real need.

## Deviations from Plan

None functionally -- the trio, its wiring, and every acceptance criterion in the plan were met exactly as specified. One documentation note, not a deviation:

**Acceptance-criterion grep collision (pre-existing, not introduced by this plan).** The plan's acceptance criteria include `grep -ciE 'hidden|excluded|archived|libraryMembership' UnreadableCatalogPanel.tsx` returning 0. This file already contained `aria-hidden="true"` on its pre-existing `!` badge `<span>` (added in an earlier phase, present in `git show HEAD~1` before this plan touched the file at all) -- the substring `hidden` inside `aria-hidden` collides with the grep's intent to catch an invented "hidden" library-membership state. Verified with a tighter check (`grep -inE '\b(excluded|archived|libraryMembership)\b'` returns nothing, and the sole `hidden`-matching line is the accessibility attribute, unrelated to membership) that no second membership concept was actually introduced. The `aria-hidden="true"` attribute was left untouched -- removing it would break screen-reader behavior for a decorative icon, which is out of this plan's scope and would violate Chesterton's Fence for a pre-existing, correctly-functioning accessibility attribute.

## Task 2 Live Verification (dev-browser, `wails dev` on :34115)

Ran directly in this session per this repo's CLAUDE.md: *"When performing GSD verifications that involve browser-based UAT, always use the dev-browser skill to test in the browser automatically. Do not ask the user to manually verify browser behavior."*

**Setup:** No stale process was found squatting on :34115 (`lsof`/`ps` both confirmed clear before starting). Started a fresh `wails dev`; `Object.keys(window.go.main.App)` was probed immediately after connecting and confirmed fresh bindings (`RescanCatalog`, `ResolveRescan`, `DeleteCatalog`, `GetCatalogHtmlPath` all present). No `osascript`/System Events/host-OS GUI automation/keystroke injection was used at any point. All fixtures were staged with plain filesystem commands in the scratchpad, never in a real catalog directory; `window.go.main.App.SelectDirectory` was stubbed in-page (pure browser-sandbox JS) to bypass the native folder picker, the same technique 28-03/28-04 established.

**Fixture setup:** a real 2-file catalog (`rescan-fixture.json`/`.html`) created via a real `StartScan` call, then five corrupted scratch copies made by truncating the JSON mid-object with Python (confirmed each raises `json.JSONDecodeError` before use): `corrupt-1` (with `.html`, used for the button-render + open-HTML checks), `corrupt-2` (with `.html`, used for the Variant B + Keep-both check), `corrupt-3` (with `.html`, used for the Overwrite check), `corrupt-4` (with `.html`, used for the Remove-from-library check), `corrupt-nohtml` (no `.html` companion, used for the button-omission check).

**Findings, one per checkpoint step:**
1. Selected `corrupt-1.json` in the rail -- `.ws-unreadable-actions` rendered exactly 3 enabled buttons: "Re-scan volume" / "Open the .html instead" / "Remove from library".
2. Clicked "Open the .html instead" (with `GetCatalogHtmlPath`/`OpenExternal` wrapped to record calls while still invoking the real bindings) -- `GetCatalogHtmlPath` resolved to `corrupt-1.html`'s full path, `OpenExternal` returned no error (Wails' `BrowserOpenURL` succeeded). No error banner appeared in the panel.
3. Selected `corrupt-nohtml.json` (never had an `.html` companion) -- `.ws-unreadable-actions` contained exactly 2 buttons (`Re-scan volume`, `Remove from library`); the HTML button was absent from the DOM entirely, confirmed by array length, not a visual/disabled check.
4. Clicked "Re-scan volume" on `corrupt-2.json`, stubbed `SelectDirectory`, clicked "...or choose any folder", clicked "Start re-scan": reached step 3 with `headerTitle: "Scan complete"`, `stepLabel: "step 3 of 3"`, `hasStatGrid: false`, `hasDiffList: false`, subline `"2 files - 12B - scanned .../source just now"`, caption verbatim the Variant B copy naming `corrupt-2.json`. Footer offered exactly `["Overwrite catalog", "Keep both"]` plus "Discard scan and close", no similarity banner. Clicked "Keep both": `corrupt-2.json` stayed byte-for-byte the same truncated content (still raised the identical `JSONDecodeError` afterward) while `corrupt-2-copy.json`/`.html` were written and parsed cleanly.
5. Re-ran the same Re-scan flow on `corrupt-3.json` and clicked "Overwrite catalog" instead: `corrupt-3.json` parsed successfully afterward (previously raised the same decode error).
6. Selected `corrupt-4.json`, clicked "Remove from library": the dialog that opened had `#ws-delete-title` reading "Delete catalog" (the real, unmodified `DeleteConfirmDialog`), listing both the `.json` and `.html` rows with "Move both to Trash" as the primary action. Clicked it: `corrupt-4.json`/`.html` disappeared from the rail's row listing and were confirmed present (recoverable) at `~/.Trash/corrupt-4.json`/`.html`.
7. Re-selected `corrupt-1.json`, opened the Create panel (`.ws-new-pill`), stubbed `SelectDirectory` to a 3000-file fixture directory, clicked "Create catalog": mid-scan, the unreadable panel's button 1 read `disabled:true`, `aria-disabled:"true"`, `title:"A scan is already running — open it from the status bar."`, while buttons 2 and 3 stayed `disabled:false`/`aria-disabled:null`/`title:null`.
8. **Cleanup:** waited for the Create scan to finish (`"Catalog written"`), closed the Create panel, deleted the entire scratchpad fixture directory, killed `wails dev`/`StorCat`/`vite`/`esbuild` (confirmed via `lsof -i :34115`/`:5173` both returning empty and `ps` showing none of those processes), reverted `frontend/wailsjs/runtime/{package.json,runtime.d.ts,runtime.js}` file-mode noise via `git checkout --`, and removed the two Trash test artifacts (`corrupt-4.json`/`.html`) since they were scratch fixtures, not anything belonging to the user. Working tree confirmed clean before this SUMMARY's commit.

## Known Stubs

None. All three actions are fully wired to real, already-shipped surfaces and live-verified against real files on disk.

## Issues Encountered

- **Acceptance-criterion grep collision with a pre-existing, unrelated attribute** -- see Deviations above. Not a real issue with the implementation; documented for a future reader who re-runs the plan's literal `<acceptance_criteria>` greps and sees a nonzero count on the `hidden|excluded|archived` check.
- **`StartScan`'s real Wails binding signature differs from the naive positional guess** (it takes `(title, sourcePath, outputDir, outputRoot, opts ScanOptions)`, not `(sourcePath, outputDir, outputRoot, includeHidden, writeHTML)`) -- caught immediately by a 30s script timeout on the first attempt, fixed by reading `app.go`'s real signature before retrying. No product code was touched; this only affected fixture-setup tooling.

## User Setup Required

None -- no external service configuration required.

## Next Phase Readiness

- `UnreadableCatalogPanel`'s action trio is complete, live-verified, and closes STATE-03 -- an unreadable catalog is no longer a dead end.
- Plan 28-06 (the phase's final sweep/CI-proof plan) can proceed; this plan adds one new reachable call site into `ResolveRescan`/`DeleteCatalog` (unchanged bindings, new frontend call sites only) and records no new open items in `.planning/WINDOWS.md` beyond what 28-04 already logged for the same bindings.

---
*Phase: 28-re-scan-diff*
*Completed: 2026-08-16*

## Self-Check: PASSED

Files verified present on disk (`frontend/src/components/workspace/UnreadableCatalogPanel.tsx`, `frontend/src/workspace.css`, this SUMMARY). Commit hash `241a0710` verified present in `git log`. `go build ./... && go vet ./... && go test ./...` green. `cd frontend && npx tsc --noEmit && npm run build` green. Working tree clean, no process listening on :34115 or :5173, no scratch fixtures remaining, Trash test artifacts removed.
