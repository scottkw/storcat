---
phase: 27-catalog-actions-watch
plan: 07
subsystem: ui
tags: [react, typescript, go, wails, fsnotify, status-bar, broken-windows-ledger]

# Dependency graph
requires:
  - phase: 27-catalog-actions-watch
    provides: "27-04's .ws-status-right/.ws-status-watching* CSS surface; 27-06's catalogs:changed event and App.applyWatchState/shutdown lifecycle; 27-02's real SIGKILL evidence for WINDOWS.md #6"
provides:
  - "StatusBar.tsx's watching segment -- WATCH-01's ● watching <dir> in var(--fn), omitted entirely when off or unset"
  - "CatalogRail.tsx's catalogs:changed subscription re-triggering loadCatalogsForDirectory -- WATCH-02's frontend half, one refresh path"
  - "CatalogSettingsSection.tsx's corrected watch-directory note, replacing Phase 26's 'applies once it ships' placeholder"
  - "WINDOWS.md disposed: entry #6 fixed on 27-02's real SIGKILL evidence; 6 new Phase 27 entries (5 planned platform gaps + 1 live-verification finding)"
  - "The completed 28-row Phase 27 verification matrix with real observed results"
affects: []

# Actuals (#2632)
actuals:
  tokens: 5129
  tasks: 3
  commits: 3

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "gsd-tools windows append/fixed for all ledger mutations -- table, JSON array and frontmatter counters stay in sync, never hand-edited"
    - "Live fd-count/lsof evidence (lsof -p <pid>) as a stronger substitute for a handle-release proof when the literal OnShutdown path can't be exercised without killing a shared wails dev session -- the toggle-off path calls the identical Watcher.Close() the quit path calls"

key-files:
  created: []
  modified:
    - frontend/src/components/workspace/StatusBar.tsx
    - frontend/src/components/workspace/CatalogRail.tsx
    - frontend/src/components/workspace/settings/CatalogSettingsSection.tsx
    - .planning/WINDOWS.md

key-decisions:
  - "Entry #6 marked fixed via `gsd-tools windows fixed 6` on 27-02-SUMMARY.md's real evidence (42+ SIGKILL iterations, 100% mid-write landing, byte-identical survivors) -- not hand-waved, not left open out of caution once the evidence was read"
  - "Added a 6th Phase 27 WINDOWS.md entry beyond the plan's own planned five, for a genuine discrepancy found live during Task 3's matrix (row 5): Menu.tsx's click-outside focus-restore loses focus to <body> in this Chromium-based test session, contradicting 27-04-SUMMARY.md's own claimed-passing coverage. Not fixed here (Menu.tsx is outside this plan's files_modified) -- recorded honestly per the ledger's own purpose rather than silently re-affirming a claim this session's re-test contradicted"
  - "Row 27 (OS-quit handle release) verified via the toggle-off path's real lsof fd-count evidence (101 fds -> 74, watched-directory DIR fd disappears) rather than an actual application quit -- shutdown() calls the identical Close() on the identical watcher field applyWatchState's toggle-off path calls, so this is mechanistically the same release code path, gathered without risking the shared wails dev session this plan's own live verification depended on"
  - "Performed the Task 3 checkpoint's live verification directly via dev-browser rather than pausing for a human to click through it, per this project's CLAUDE.md directive to run GSD browser-based UAT with dev-browser rather than asking the user to manually verify"

requirements-completed: [WATCH-01, WATCH-02, WATCH-03]

coverage:
  - id: D1
    description: "Watching status-bar segment: appears as '● watching <dir>' in var(--fn) (not var(--ac)) exactly when watchDirectory is on AND catalogDir is set; omitted entirely otherwise; sits left of the scan segment when both render; 160px truncation on the directory path"
    requirement: "WATCH-01"
    verification:
      - kind: e2e
        ref: "live dev-browser against :34115 -- matrix rows 20-22: DOM+computed-style proof segment color equals var(--fn) (rgb(119,113,97)) not var(--ac) (#fe8019); toggled off leaves .ws-status-right present but empty (no placeholder); live-polled DOM order ['ws-status-watching','ws-status-scan'] during a real concurrent scan"
        status: pass
    human_judgment: false
  - id: D2
    description: "CatalogRail subscribes to catalogs:changed and re-lists via the existing loadCatalogsForDirectory (no second read path); external add/remove reflected without a manual reload; a 10-file burst coalesces to one BrowseCatalogs call"
    requirement: "WATCH-02"
    verification:
      - kind: e2e
        ref: "live dev-browser -- matrix rows 23-25: external cp/rm reflected in the rail with watching on and no reload; instrumented window.go.main.App.BrowseCatalogs call-count = 1 for a real 10-file cp burst"
        status: pass
    human_judgment: false
  - id: D3
    description: "Toggling watching off stops delivery (no rail refresh on further external changes) and genuinely releases the OS watch handle, not just event delivery"
    requirement: "WATCH-03"
    verification:
      - kind: e2e
        ref: "live dev-browser -- matrix row 26: instrumented BrowseCatalogs call count 0 after toggle-off + external touch; lsof -p <StorCat pid> before/after toggle-off: 101 fds -> 74, watched-directory DIR fd (17r) present then absent"
        status: pass
      - kind: unit
        ref: "internal/watch/watcher_test.go#TestWatcher_Close, #TestWatcher_CloseIsIdempotent (unchanged from 27-06)"
        status: pass
    human_judgment: true
    rationale: "Matrix row 27 (the real application-quit OnShutdown path) was not exercised live -- doing so would have required quitting the shared wails dev process this plan's own verification session depended on, with no guarantee of a clean relaunch. Evidence instead: shutdown()'s body calls Close() on the identical watcher field the toggle-off path calls (27-06's own acceptance grep), and this plan's toggle-off path gathered fresh lsof fd-count evidence that Close() genuinely releases the OS handle, not just event delivery. A human with the built app should confirm one real quit-and-relaunch before treating this as fully closed."
  - id: D4
    description: "Settings' watch-directory toggle note describes what watching does, statically, independent of on/off state -- Phase 26's 'applies once file watching ships' placeholder is gone"
    requirement: "WATCH-01"
    verification:
      - kind: e2e
        ref: "live dev-browser -- matrix row 28: DOM query confirms note text 'Detects catalogs added, removed, or edited outside the app' present, old placeholder string absent anywhere in the page"
        status: pass
    human_judgment: false
  - id: D5
    description: ".planning/WINDOWS.md: entry #6 disposed on real evidence (fixed), 5 planned Phase 27 platform-gap entries recorded, and 1 additional live-verification finding recorded -- all three ledger representations (table/JSON/counters) stay in sync"
    verification:
      - kind: other
        ref: "node verify_ledger.js: table rows == JSON length == total_count == 13; open+waived+fixed == total; 6 Phase 27 entries found; every Phase 27 entry's file exists on disk and description matches /unverified|not removed|not supported|upstream|accepted/i"
        status: pass
    human_judgment: false
  - id: D6
    description: "Full Phase 27 verification matrix (28 rows spanning ACT-01..ACT-09, WATCH-01..03, SET-04) completed with real observed results, not reasoned expectations"
    verification:
      - kind: e2e
        ref: "live dev-browser against :34115, binding freshness confirmed via Object.keys(window.go.main.App) before recording evidence; see this SUMMARY's Verification Matrix section for the full per-row record"
        status: pass
    human_judgment: true
    rationale: "27 of 28 rows matched their expected result with direct live evidence. Row 5 (menu click-outside focus restore) did NOT match its expected result in this session's Chromium-based test environment -- recorded as a new WINDOWS.md finding rather than silently passed. A human should read the Verification Matrix section below (particularly rows 5 and 27) before signing off the phase as fully verified."

duration: 45min
completed: 2026-08-16
status: complete
---

# Phase 27 Plan 07: Watching Status-Bar Segment, Rail Refresh, Ledger Disposition, and Phase Verification Summary

**The status bar's `● watching <dir>` segment and `CatalogRail`'s `catalogs:changed` subscription land WATCH-01/WATCH-02's frontend half; `.planning/WINDOWS.md` disposes entry #6 on 27-02's real SIGKILL evidence and records six Phase 27 gaps (five planned platform-unverified entries plus one live-verification finding); and the full 28-row Phase 27 verification matrix was run live against `wails dev :34115` via dev-browser, with 27 rows confirmed and one (menu click-outside focus restore) recorded as a genuine, unfixed discrepancy.**

## Performance

- **Duration:** ~45 min (includes the full 28-row live verification pass)
- **Started:** 2026-08-15T19:00:00-05:00 (approx.)
- **Completed:** 2026-08-16T00:24:00Z
- **Tasks:** 3 (2 auto tasks + 1 checkpoint:human-verify performed live via dev-browser per CLAUDE.md's UAT directive)
- **Files modified:** 4 (3 frontend, 1 planning ledger)

## Accomplishments

- `StatusBar.tsx` gained a `.ws-status-right` wrapper holding the new `.ws-status-watching` segment (`● watching <dir>`, a `<span>` not a `<button>`) ahead of the pre-existing scan segment. Visibility is `state.settings.watchDirectory && !!state.catalogDir` -- both already resolved synchronously in `AppContext`, so there's no loading branch. Confirmed live: color equals `var(--fn)` (`rgb(119,113,97)`), never `var(--ac)` (`#fe8019`); the segment is fully absent (not a placeholder) when either input is false; 160px truncation with ellipsis on the directory path.
- `CatalogRail.tsx` subscribes to `catalogs:changed` via `EventsOn`, re-triggering the exact same `loadCatalogsForDirectory` its existing `state.catalogDir` effect already calls -- confirmed only one `wailsAPI.browseCatalogs` call site exists in the file. The effect returns `EventsOn`'s own unsubscribe function (StrictMode double-invoke safe). Confirmed live: external `cp`/`rm` reflected in the rail with no reload; a real 10-file burst produced exactly one `BrowseCatalogs` call (instrumented live), not ten.
- `CatalogSettingsSection.tsx`'s watch-directory toggle note now reads "Detects catalogs added, removed, or edited outside the app," replacing Phase 26's "applies once file watching ships" placeholder -- confirmed absent anywhere in the page via a live DOM string search.
- `.planning/WINDOWS.md` entry #6 (atomicwrite SIGKILL crash-safety) marked **fixed** via `gsd-tools windows fixed 6`, citing `27-02-SUMMARY.md`'s real evidence: 42+ genuine `SIGKILL`s across two tests, 100% mid-write landing rate (confirmed by `.tmp` residue left behind every time), byte-identical (SHA-256) survivors, re-run 3+ times with identical results.
- Five planned Phase 27 deviation entries recorded (fsnotify Windows rename-non-removal; fsnotify Windows/Linux backends unverified; wastebasket Windows/Linux backends unverified; `WriteFileAtomic`'s Windows directory-fsync discard path unverified; wastebasket's macOS AppleScript interpolation as an accepted upstream residual risk) -- all via `gsd-tools windows append`, none hand-edited.
- A sixth, unplanned Phase 27 entry was added after Task 3's own live re-test contradicted `27-04-SUMMARY.md`'s claimed-passing coverage for Menu click-outside focus restore (see Deviations and the Verification Matrix's row 5 below).
- The full 28-row Phase 27 verification matrix was run live against a freshly-confirmed `wails dev :34115` session (bindings checked via `Object.keys(window.go.main.App)` before recording any evidence) using dev-browser -- no host-OS GUI automation, all filesystem perturbation via Bash (`cp`/`rm`/`touch`/`chflags`). 27 rows matched their expected result; row 5 did not and is recorded honestly rather than silently passed.

## Task Commits

1. **Task 1: The watching status-bar segment, the rail's catalogs:changed subscription, and the Settings note correction** - `393c3ce0` (feat)
2. **Task 2: Record every platform-gated gap in the broken-windows ledger** - `15b3eebd` (docs)
3. **Task 3: Phase 27 verification matrix -- live-verified via dev-browser; the one genuine discrepancy found (row 5) recorded as a new ledger entry** - `ba3c9d96` (docs)

**Plan metadata:** pending (this SUMMARY's own commit)

## Files Created/Modified
- `frontend/src/components/workspace/StatusBar.tsx` - `.ws-status-right` wrapper, `.ws-status-watching` segment
- `frontend/src/components/workspace/CatalogRail.tsx` - `catalogs:changed` `EventsOn` subscription with unsubscribe cleanup
- `frontend/src/components/workspace/settings/CatalogSettingsSection.tsx` - watch-directory toggle note corrected
- `.planning/WINDOWS.md` - entry #6 disposed (fixed); 6 new Phase 27 entries (ids 8-13)

## Decisions Made
- Entry #6 disposed as **fixed** on 27-02's real evidence, not left open out of caution once the evidence was read -- a false "still open" would understate real, already-proven crash safety just as much as a false "fixed" would overstate unproven safety.
- A 6th, unplanned WINDOWS.md entry (id 13) was added for the live-verification finding in matrix row 5 -- the ledger's own stated purpose (an honest cross-phase defect register) required recording a genuine discrepancy this session's own re-test surfaced, even though it wasn't one of the plan's five pre-scoped entries and even though fixing it is out of this plan's file scope.
- Row 27 (real OS-quit handle release) was verified via the mechanistically identical `Close()` call on the toggle-off path (fresh `lsof` fd-count evidence: 101 -> 74 fds, watched-directory `DIR` fd released) rather than by actually quitting the shared `wails dev` session, matching the task's own explicitly sanctioned fallback and avoiding risk to the session other rows' verification still depended on.
- Task 3's checkpoint was executed directly (live dev-browser verification performed by the executor) rather than paused for the user, per this project's `CLAUDE.md`: "When performing GSD verifications that involve browser-based UAT, always use dev-browser to test in the browser automatically. Do not ask the user to manually verify browser behavior."

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Comment text on StatusBar.tsx's new segment tripped the plan's own aria-live/title acceptance grep**
- **Found during:** Task 1 verification
- **Issue:** The explanatory comment above the watching segment used the literal words "aria-live" and "title" in prose, which the acceptance grep (`! grep -Eq 'aria-live|title=' StatusBar.tsx`) also matched.
- **Fix:** Reworded the comment ("no live-region announcement," "no hover tooltip," "hover-tooltip affordance") to convey the same intent without the literal forbidden substrings. No functional code change.
- **Files modified:** `frontend/src/components/workspace/StatusBar.tsx`
- **Verification:** `! grep -Eq 'aria-live|title=' StatusBar.tsx` succeeds; `tsc --noEmit && npm run build` green.
- **Committed in:** `393c3ce0` (Task 1 commit)

### Live-verification finding (not fixed -- outside this plan's file scope; recorded honestly)

**2. [Verification finding] Menu click-outside dismissal loses focus restore to `<body>`, contradicting 27-04's own claimed-passing coverage**
- **Found during:** Task 3, matrix row 5 ("Reopen, click elsewhere in the app -> Menu closes, focus returns to the trigger")
- **Issue:** `27-04-SUMMARY.md`'s coverage D1 claimed this was verified live and passing ("Escape and click-outside close, and focus restore to the ⋯ button on every close path"). Re-testing this phase in the same dev-browser (Chromium) environment against `wails dev :34115` shows Escape-driven close restores focus reliably every time, but a click on an empty, non-focusable area of the app does **not**: instrumentation of `HTMLElement.prototype.focus` confirmed `useModalBehavior`'s cleanup effect DOES call `.focus()` on the trigger button, but the click gesture's own native focus-follows-click default action (landing on `document.body`, since the click target isn't focusable) fires after that restore and wins, leaving `document.activeElement` as `<body>`.
- **Action NOT taken:** `Menu.tsx`/`useModalBehavior.ts` are outside this plan's `files_modified` scope (StatusBar.tsx, CatalogRail.tsx, CatalogSettingsSection.tsx, WINDOWS.md). Per the deviation rules' scope boundary, an out-of-scope pre-existing behavior found during verification is recorded, not silently fixed inline.
- **Recorded:** `.planning/WINDOWS.md` entry 13 (`kind: unmet-truth`, `file: frontend/src/components/workspace/Menu.tsx`), including the explicit caveat that this was observed in Chromium via dev-browser, not the actual shipped app's WKWebView engine on macOS, since host-OS GUI automation is prohibited and the two engines can differ on native focus-follows-click timing.
- **Verification:** Reproduced twice with different empty-click coordinates and wait durations (300ms and 800ms settle); `focusin`/`focusout` event log confirms the sequence `focusout:ws-menu-item -> focusin:ws-details-overflow -> focusout:ws-details-overflow`.

---

**Total deviations:** 2 (1 auto-fixed cosmetic wording mismatch, 1 unfixed live-verification finding recorded to the ledger).
**Impact on plan:** None on this plan's own deliverables (WATCH-01/WATCH-02 frontend, Settings note, ledger discipline) -- all verified working as specified. The unfixed finding is a pre-existing Phase 27 (27-04) behavior this plan's verification pass happened to catch; it is now tracked rather than silently missed.

## Verification Matrix (28 rows, live against `wails dev :34115`)

Bindings confirmed fresh before any evidence was recorded: `Object.keys(window.go.main.App)` included `RenameCatalog`, `DuplicateCatalog`, `DeleteCatalog`. All filesystem perturbation via Bash (`cp`/`rm`/`touch`/`chflags`); no host-OS GUI automation.

| # | Req | Check | Expected | Observed |
|---|-----|-------|----------|----------|
| 1 | ACT-01 | Click `⋯` in the details panel | Menu opens with all 3 items | **Match.** `["Rename catalog…","Duplicate catalog","Delete catalog…"]` |
| 2 | ACT-01 | Click `⋯` again while open | Menu closes | **Match.** Menu absent from DOM after second click. |
| 3 | ACT-01 | ArrowDown past last, ArrowUp past first | Focus wraps at both ends | **Match.** Live focus sequence: Rename→Duplicate→Delete→(wrap)Rename; Up-wrap→Delete. |
| 4 | ACT-01 | Press Escape | Menu closes, focus returns to `⋯` | **Match.** (Required widening viewport to ≥1280px so the panel renders as the inline pane, not the drawer variant, whose own Escape handler otherwise also closes the whole panel -- an app-width interaction, not a menu bug.) |
| 5 | ACT-01 | Reopen, click elsewhere | Menu closes, focus returns to trigger | **Did NOT match.** Menu closes correctly; focus does not survive -- lands on `<body>`. Recorded as WINDOWS.md entry 13 (see Deviations). |
| 6 | ACT-01 | Open menu during a real running scan | All 3 items enabled | **Match.** Live 32k-file scan in flight; all 3 items `disabled: false`. |
| 7 | ACT-02 | Rename to `Tom & Jerry <2024>` | Rail + details update immediately, no `&amp;` | **Match.** Rail title text `Tom & Jerry <2024>` (raw, unescaped) in both places. |
| 8 | ACT-02 | `grep -c` the `.html` | Returns 2 | **Match.** `grep -c 'Tom &amp; Jerry &lt;2024&gt;' test-catalog-one.html` = 2. |
| 9 | ACT-02 | Compare filenames before/after | Unchanged | **Match.** `test-catalog-one.{json,html}` unchanged. |
| 10 | ACT-02 | Rename a catalog with no `.html` | Succeeds, no `.html` created | **Match.** Only `no-html-catalog.json` present after rename. |
| 11 | ACT-03 | Duplicate twice | `-copy` then `-copy-2` | **Match.** `test-catalog-one-copy.{json,html}`, `test-catalog-one-copy-2.{json,html}` both created. |
| 12 | ACT-03 | Compare duplicate `.json` bytes | Byte-identical | **Match.** `diff` reports identical for both copies. |
| 13 | ACT-04 | Delete on a catalog with `.html` | 2 path boxes, checkbox checked, `Move both to Trash` | **Match.** Screenshot confirms both, verbatim paths, checked box. |
| 14 | ACT-04 | Uncheck the checkbox | Label becomes `Move to Trash` | **Match.** |
| 15 | ACT-04 | Delete on a catalog with no `.html` | 1 path box, no checkbox, `Move to Trash` | **Match.** `boxCount:1, checkboxPresent:false, primary:"Move to Trash"`. |
| 16 | ACT-04 | Confirm delete on a disposable catalog | Files appear in OS Trash | **Match.** File gone from catalog dir, present in `~/.Trash`. |
| 17 | ACT-05 | Induce a real trash failure (`chflags uchg`) | Error shows real system message | **Match.** "Couldn't move Tom & Jerry <2024> to the Trash: delete .../test-catalog-one-copy-2.json: trash: exit status 1." |
| 18 | ACT-05 | Inspect every control/string in error state | Only `Keep catalog` / `Try moving to Trash again`, no permanence vocab | **Match.** Buttons exactly `["×","Keep catalog","Try moving to Trash again"]`; permanence-vocab regex found nothing. |
| 19 | ACT-09 | `go test -run TestWriteFileAtomic -count=1 -v` | Green, including kill cases | **Match.** All 10 tests pass, including 21+21 real-SIGKILL iterations. |
| 20 | WATCH-01 | Enable watching with a catalog dir set | `● watching <dir>` in `--fn` | **Match.** DOM+computed-style confirmed `rgb(119,113,97)` == `--fn`; screenshot (dev-only theme-switcher overlay hidden for the capture). |
| 21 | WATCH-01 | Disable watching | Segment disappears entirely | **Match.** `.ws-status-right` present but empty; `.ws-status-watching` absent from DOM. |
| 22 | WATCH-01 | Start a scan while watching is on | Both segments render, watching left of scan | **Match.** Live-polled DOM order `["ws-status-watching","ws-status-scan"]` during a real in-flight scan. |
| 23 | WATCH-02 | External `cp` a `.json` | Rail shows it within ~1s | **Match.** New row present by first poll after the `cp`. |
| 24 | WATCH-02 | External `rm` a `.json` | Rail drops it within ~1s | **Match.** Row absent by first poll after the `rm`. |
| 25 | WATCH-02 | `cp` 10 files at once | Rail refreshes ~once, not 10x | **Match.** Instrumented `BrowseCatalogs` call count = 1 for the 10-file burst. |
| 26 | WATCH-03 | Toggle off, then `touch` a `.json` | Rail does not refresh | **Match.** 0 `BrowseCatalogs` calls; new file absent from rail. |
| 27 | WATCH-03 | Quit the app entirely while watching was on | No error; relaunch works normally | **Not exercised as a real quit** (would have required killing the shared `wails dev` session). Substitute evidence gathered instead: `lsof -p <StorCat pid>` before/after `SetWatchDirectory(false)` (which calls the identical `Close()` `shutdown()` calls) showed 101 fds -> 74 fds, with the watched directory's `DIR` fd (`17r`) present then genuinely absent -- real OS-level handle release, not just event-delivery stopping. `TestWatcher_Close`/`TestWatcher_CloseIsIdempotent` and `main.go`'s `OnShutdown: app.shutdown` registration remain unit/grep-verified per 27-06. Recorded as `human_judgment: true` in this SUMMARY's coverage (D3). |
| 28 | SET-04 | Open Settings, read the watch-directory note | Describes detection, no placeholder | **Match.** Note reads "Detects catalogs added, removed, or edited outside the app"; old placeholder string absent from the entire page. |

**Result: 27/28 matched their expected result with direct live evidence. Row 5 recorded as a genuine discrepancy (WINDOWS.md entry 13). Row 27 recorded with substitute evidence and flagged for human judgment.**

## Issues Encountered
- **Row 4/5 required widening the browser viewport to ≥1280px.** At the default (narrower) viewport, `WorkspaceShell.tsx`'s details panel renders as an overlay drawer, and a global `window`-level Escape listener (`state.detailOverlay`-gated, unrelated to the menu) also closes the whole drawer on any Escape keypress that bubbles up -- confounding the menu's own Escape-close test. Widening to the app's real default window width (1470x923, matching `config.json`'s persisted `windowWidth`/`windowHeight`) switches the panel to its inline-pane variant, matching normal desktop usage and eliminating the confound.
- **Row 22/6 needed genuinely slow scans to catch the in-flight `'counting'`/`'scanning'` window**, since Go's flat-file scan of a few thousand tiny local files completes in well under `throttledProgress`'s 200ms emit interval, meaning `SCAN_PROGRESS` may never dispatch at all for a fast scan (and `state.scan.status` only flips via `SCAN_STARTED`, which is dispatched by `CreateSlideOver`'s real submit handler, not by calling the raw `StartScan` Go binding directly). Resolved by driving the real UI submit flow (clicking `Catalog a volume` -> `Create catalog`, which does dispatch `SCAN_STARTED`) against a 32,000-file scratch tree, and by polling in a tight loop rather than a single fixed wait.
- **Row 20's evidence was initially screenshot-obscured** by `#storcat-dev-switcher`, a dev-only (`wails dev`-only, not shipped) debug overlay fixed at the same bottom-right corner as the status bar's right segment. DOM/computed-style evidence (unaffected by paint order) already proved the feature; the overlay was hidden via `element.style.display='none'` purely for a clean screenshot, not a source change.

## User Setup Required
None - no external service configuration required.

## Known Stubs
None.

## Threat Flags
None - no new network endpoints, auth paths, or trust-boundary schema changes introduced by this plan's files.

## Next Phase Readiness
- **Phase 27 is functionally complete.** All three plan-owned deliverables (WATCH-01/WATCH-02 frontend, Settings note, ledger discipline) are live-verified. The phase's full 9-requirement set (ACT-01..05, ACT-09, WATCH-01..03) has a completed 28-row verification matrix.
- **Two items carry forward for a human to close before full sign-off:**
  1. WINDOWS.md entry 13 (Menu.tsx click-outside focus restore) -- needs either a real fix in `Menu.tsx`/`useModalBehavior.ts` (out of this plan's scope) or a manual click-through in the actual built macOS app to confirm/deny it reproduces outside this session's Chromium test harness.
  2. Matrix row 27 -- a human with the built app should perform one real quit-and-relaunch to directly confirm `OnShutdown`'s watcher release, since this session substituted the mechanistically-identical toggle-off path's fd evidence rather than risking the shared verification session.
- **WINDOWS.md now carries 13 entries total (11 open, 2 fixed, 0 waived)** -- unchanged phases 23-26 entries plus Phase 27's 6. `open_count > 0` will still block `/gsd-ship` under `workflow.windows_enforce`; most open entries are cross-platform-runtime-unverified items explicitly slated for a pre-v3.0.0 sweep, not phase-27-specific blockers.
- No new dependencies, no schema changes, no new endpoints from this plan.

## Self-Check: PASSED

All 4 modified files verified present on disk (`frontend/src/components/workspace/StatusBar.tsx`, `frontend/src/components/workspace/CatalogRail.tsx`, `frontend/src/components/workspace/settings/CatalogSettingsSection.tsx`, `.planning/WINDOWS.md`); all 3 commits (`393c3ce0`, `15b3eebd`, `ba3c9d96`) verified present in `git log`.

---
*Phase: 27-catalog-actions-watch*
*Completed: 2026-08-16*
