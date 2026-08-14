---
phase: 24-cmd-k-command-palette
plan: 02
subsystem: ui
tags: [wails, go, react, typescript, search, command-palette, dev-browser]

# Dependency graph
requires:
  - phase: 24-cmd-k-command-palette
    provides: "24-01's SearchIndexed binding, always-mounted CommandPalette overlay, both open paths"
provides:
  - "Live, observed proof (not assumed) of PLT-02/PLT-03 against a real multi-catalog directory at :34115: the 50-cap, the true total, cross-catalog non-dedup, the 2-char floor, the 200ms debounce, and the stale guard"
  - "localStorage['storcat-catalog-directory'] + reload technique for pointing the app at a fixture directory in dev-browser verification without driving the native OS folder picker"
  - "The 'curl liveness is not binding freshness' lesson: a running wails dev binary can predate a just-landed Go binding while still answering curl -sf 200; the real precondition probe is page.evaluate(() => Object.keys(window.go.main.App))"
affects: [24-03, 24-04, 24-05, 25, 26, 27, 28]

# Actuals (#2632)
actuals:
  tokens: 0
  tasks: 1
  commits: 1

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "dev-browser directory-pointing without the native picker: page.evaluate(() => localStorage.setItem('storcat-catalog-directory', dir)) then page.reload() -- CatalogRail's mount effect reads the persisted key and calls browseCatalogs itself, the same code path a real relaunch uses"
    - "dev-browser call-counting shim: wrap window.go.main.App.<Method> in page.evaluate before any interaction, increment a page-global counter, never write the shim to a repo file"
    - "Binding-freshness probe for a running wails dev instance: page.evaluate(() => Object.keys(window.go.main.App)) must list the specific new binding -- curl -sf only proves the HTTP server answers, not that the compiled Go binary includes recent changes"

key-files:
  created: []
  modified: []

key-decisions:
  - "No code changes were required in this plan -- all seven of Task 1's live checks passed against the tracer's existing implementation with no defect found, so files_modified's four listed files (CommandPalette.tsx, WorkspaceShell.tsx, Toolbar.tsx, search_indexed.go) remain byte-identical to 24-01"
  - "Chose 'FILE' (matches all 400 nodes of the -shape flat fixture) for the >50-cap check and '00' (matches 39 of the -shape dcim fixture's 48 nodes plus all 400 flat nodes) for the cross-catalog check, because dcim's filename sorts alphabetically before flat's so the capped top-50 response is guaranteed to include rows from both catalogs"
  - "Restarted a stale wails dev instance mid-plan (coordinator-approved, since killing processes required permission this executor's sandbox does not have): the instance serving :34115 at task start had been built one minute before the commit that added App.SearchIndexed, so window.go.main.App lacked the very binding this plan verifies, despite curl -sf returning 200 the whole time"

patterns-established:
  - "SearchIndexed live-evidence checks are keyed to generator-printed nodes= counts, never guessed -- 'FILE' against a -shape flat -files 400 fixture predicts exactly total=400; verified with jq/grep against the raw fixture JSON before trusting the UI's rendered total"

requirements-completed: []
# PLT-02 and PLT-03 are proven live by Task 1 below. PLT-01 is NOT marked complete --
# its keyboard-open path (⌘K reaching a window keydown listener inside real macOS
# WKWebView, RESEARCH Open Question #1) is Task 2, a blocking human checkpoint not
# yet run. Do not mark PLT-01 complete until Task 2 returns an observed pass.

coverage:
  - id: D1
    description: "Toolbar .ws-search click opens the always-mounted CommandPalette with focus already on its input"
    requirement: "PLT-01"
    verification:
      - kind: automated_ui
        ref: "dev-browser :34115 -- document.activeElement.className === 'ws-palette-input' after click, observed"
        status: pass
    human_judgment: false
  - id: D2
    description: "A 1-character query issues zero SearchIndexed binding calls (2-char floor)"
    requirement: "PLT-01"
    verification:
      - kind: automated_ui
        ref: "dev-browser :34115 -- call-counting shim over window.go.main.App.SearchIndexed read 0 after typing 'F' and waiting 400ms, observed"
        status: pass
    human_judgment: false
  - id: D3
    description: "A >50-match term ('FILE' against the 400-node flat fixture) renders exactly 50 rows with total=400, matching the fixture generator's printed nodes=400"
    requirement: "PLT-03"
    verification:
      - kind: automated_ui
        ref: "dev-browser :34115 -- .ws-palette-row count=50, readout text '50 of 400', observed; screenshot t2402-over50-results.png"
        status: pass
    human_judgment: false
  - id: D4
    description: "A term present in both fixture catalogs ('00') produces rows attributed to both catalog names -- no cross-catalog dedup"
    requirement: "PLT-02"
    verification:
      - kind: automated_ui
        ref: "dev-browser :34115 -- distinct catalog values observed in rendered rows: ['fixture-dcim','fixture-flat'], readout '50 of 439', observed"
        status: pass
    human_judgment: false
  - id: D5
    description: "Debounce + stale-response guard: an 8-character rapid burst ('FILE_000', 15ms/char, ~120ms total) settles on the final query with far fewer than 8 binding calls and no intermediate result set left rendered"
    requirement: "PLT-01"
    verification:
      - kind: automated_ui
        ref: "dev-browser :34115 -- exactly 1 SearchIndexed call during the burst (observed delta), 50 rendered rows all containing 'file_000' case-insensitively, observed"
        status: pass
    human_judgment: false
  - id: D6
    description: "A term matching nothing ('zzznotfound') settles with zero rendered rows and no console error"
    requirement: "PLT-03"
    verification:
      - kind: automated_ui
        ref: "dev-browser :34115 -- .ws-palette-row count=0, consoleErrorsDuringRun=[], observed; screenshot t2402-zero-results.png"
        status: pass
    human_judgment: false
  - id: D7
    description: "Escape closes the palette (scrim removed from DOM)"
    requirement: "PLT-01"
    verification:
      - kind: automated_ui
        ref: "dev-browser :34115 -- .ws-palette-scrim count=0 after Escape keypress, observed"
        status: pass
    human_judgment: false
  - id: D8
    description: "⌘K reaches a window keydown listener inside the real macOS WKWebView packaged app (RESEARCH Open Question #1), plus the second-press no-op and rail-filter-focused override behaviors"
    requirement: "PLT-01"
    verification: []
    human_judgment: true
    rationale: "Cannot be observed from a Chrome/Playwright session against :34115 -- Chrome is not WKWebView and this is the phase's one load-bearing unknown per 24-RESEARCH.md Open Question #1. This is Task 2, a blocking checkpoint:human-verify not yet run at the time this SUMMARY was written."

# Metrics
duration: ~20min (Task 1 only; Task 2 not started)
completed: 2026-08-14
status: halted
---

# Phase 24 Plan 02: Live Multi-Catalog Search Proof (Task 1 of 2) Summary

**PLT-02/PLT-03 proven live against a real 448-node two-catalog fixture directory at wails dev's :34115 -- the 50-cap, the true total, cross-catalog non-dedup, the 2-char floor, the 200ms debounce, and the stale-response guard are all observed numbers, not inferred from the Go unit tests. Task 2 (⌘K in the real macOS window) has NOT run.**

**PLAN STATUS: HALTED, NOT COMPLETE.** Task 1 (this record) is done. Task 2 -- a blocking `checkpoint:human-verify` that only a human at the real native StorCat.app window can answer -- is pending. Do not treat PLT-01 as proven until Task 2 returns.

## Performance

- **Duration:** ~20 min for Task 1 (fixture generation through full-suite verification)
- **Tasks:** 1 of 2 (Task 2 pending)
- **Files modified:** 0 (no defect found in the tracer's files; nothing needed repair)

## Accomplishments (Task 1 only)

All seven of Task 1's numbered checks, run against `wails dev` on `:34115` via the `dev-browser` skill, with **observed** values:

1. **Toolbar click opens palette, focused input.** Clicked `.ws-search`; `document.activeElement.className` = `"ws-palette-input"`. PASS.
2. **1-char query issues no binding call.** Typed `"F"`; call-counting shim on `window.go.main.App.SearchIndexed` read **0** after a 400ms wait (> the 200ms debounce). PASS.
3. **>50-match term caps at exactly 50, total exceeds 50.** Typed `"FILE"` against the `-shape flat -files 400` fixture (generator printed `nodes=400`). Observed: `.ws-palette-row` count = **50**, readout text = **"50 of 400"**. 400 matches the generator's printed node count exactly (independently confirmed via `jq`/`grep` over the raw fixture JSON: 400/400 flat node names contain "file" case-insensitively). Screenshot: `t2402-over50-results.png`. PASS.
4. **Cross-catalog term is not deduplicated.** Typed `"00"` (matches 39 of the dcim fixture's 48 nodes -- confirmed via `jq`/`grep` -- plus all 400 flat nodes = 439 true total). Observed: 50 rendered rows, readout **"50 of 439"**, distinct catalog values in the rendered rows = **`["fixture-dcim", "fixture-flat"]`** (both present, because `fixture-dcim.json` sorts alphabetically before `fixture-flat.json` in `os.ReadDir`, so the capped top-50 always spans both). PASS.
5. **Debounce + stale guard on a rapid burst.** Cleared the input, then typed `"FILE_000"` (8 characters) via `pressSequentially` at 15ms/char (~120ms total, under the 200ms debounce window), then waited 500ms. Observed: exactly **1** `SearchIndexed` call fired during the whole burst (delta measured against the call counter immediately before typing began), 50 rendered rows, and every rendered row's basename contained `"file_000"` case-insensitively -- no intermediate/stale result set was left on screen. PASS.
6. **Zero-match term.** Typed `"zzznotfound"`; observed **0** rendered rows, **0** console errors captured during the run. Screenshot: `t2402-zero-results.png`. PASS.
7. **Escape closes the palette.** Pressed Escape; observed `.ws-palette-scrim` count = **0** (component unmounted, per its `isOpen ? ... : null` contract). PASS.

Cumulative `SearchIndexed` call count across the whole session: **4** (one each for "FILE", "00", the "FILE_000" burst, and "zzznotfound" -- the "F" 1-char query correctly contributed 0).

**No defect was found in any of the tracer's own files** (`CommandPalette.tsx`, `WorkspaceShell.tsx`, `Toolbar.tsx`, `search_indexed.go`). All seven checks passed against the existing 24-01 implementation with zero edits required.

Full automated suite, run after the live checks: `go build ./... && go test ./... -count=1` green (7 packages ok, 2 no-test-files); `cd frontend && npx tsc --noEmit && npm run build` green (pre-existing "Module level directives" bundler notices are unrelated antd noise, not from any file this plan touches).

`git status --porcelain` at the end of Task 1: clean. No fixture directory, no instrumentation file, and (before this SUMMARY commit) no other changes were left in the working tree.

## Screenshots

Both screenshots were captured successfully via `dev-browser`'s `saveScreenshot`, but **they are not committed to this repository** (consistent with Phase 23's `t23-tree-a.png` convention of naming a screenshot without committing it) and are **not currently referenced by any other file**, so record their actual disk location here as the evidence trail:

- `t2402-over50-results.png` -- `/Users/ken/.dev-browser/tmp/t2402-over50-results.png` (93,115 bytes) -- the `.ws-palette-panel` at the "FILE" query's 50-of-400 state.
- `t2402-zero-results.png` -- `/Users/ken/.dev-browser/tmp/t2402-zero-results.png` (14,540 bytes) -- the `.ws-palette-panel` at the "zzznotfound" query's zero-row state.

Both exist on disk as of this writing (confirmed with `ls -la`). They live under `dev-browser`'s own tmp directory, outside the repository and outside `.planning/`, and will not survive an unrelated `dev-browser` tmp cleanup -- if a later verifier needs them and they're gone, re-run Task 1's checks 3 and 6 to regenerate.

## Environment Incident: Stale `wails dev` Binary (Record For Future Phases)

At the start of Task 1, the `wails dev` instance already running and answering `curl -sf http://localhost:34115/` with 200 turned out to be **stale**: its compiled Go binary (`build/bin/StorCat.app/Contents/MacOS/StorCat`, built at 09:26) predated the commit (`2f226ef2`, landed at 09:27:01) that added `App.SearchIndexed` to `app.go` and regenerated the Wails bridge. `window.go.main.App` at runtime was missing `SearchIndexed` entirely, even though the generated frontend files on disk (`frontend/wailsjs/go/main/App.js`/`.d.ts`) already had it.

**Root cause:** two duplicate `wails dev` processes were running simultaneously (both started the prior evening), and neither's file-watcher picked up the Go-side change after 10+ minutes and a manual `touch app.go` nudge -- most likely the duplicate instances were contending over the same build output.

**This executor could not self-remediate:** restarting `wails dev` requires killing the stale processes, and the harness's auto-mode permission classifier denied both `kill <pids>` and `pkill -f "wails dev"` as out-of-scope actions. Work paused at a `checkpoint:human-action`; the coordinator obtained user approval, killed both stale instances and their child app windows, and relaunched a single clean `wails dev` instance (new binary built 09:42:29, after `2f226ef2`).

**Lesson for Phases 25-28 (recorded here so it isn't rediscovered):** `curl -sf` against a Wails dev server's HTTP port proves **liveness**, not **freshness**. It will return 200 from a binary that is minutes or hours stale relative to the current Go source. The real precondition probe for "does this running instance have the binding I need" is:

```js
await page.evaluate(() => Object.keys(window.go.main.App));
```

and asserting the specific new binding's name is in that list. Re-run this probe after any coordinator-initiated `wails dev` restart, even if `curl -sf` already passed -- do not treat `curl` alone as sufficient before recording live evidence.

## Verification Technique: Pointing the App at a Fixture Directory Without the Native Picker

`CatalogRail`'s `.ws-dir-chip` click calls `wailsAPI.selectDirectory()`, which opens the OS-native folder picker -- outside the webview's DOM and therefore outside what browser automation (`dev-browser`/Playwright) can drive directly.

The verified working substitute, usable by any future phase needing dev-browser to point the app at a specific directory:

```js
await page.evaluate((dir) => {
  localStorage.setItem('storcat-catalog-directory', dir);
}, fixtureDir);
await page.reload({ waitUntil: 'networkidle' });
```

This works because `CatalogRail`'s mount effect reads the `storcat-catalog-directory` localStorage key and calls `wailsAPI.browseCatalogs(dir)` itself -- the exact same code path a real app relaunch with a previously-chosen directory uses. Confirmed against the two-fixture directory: after reload, the rail showed `CATALOGS 2` with both `fixture-dcim` and `fixture-flat` listed. **Phases 25-28 will need this same technique** for any live verification requiring a specific catalog directory.

## Fixture Directory (for reuse/regeneration)

`/private/tmp/claude-501/-Users-ken-dev-storcat/f278c24f-ba59-4af9-907e-602d0eec8b01/scratchpad/t2402-fixtures/` (scratchpad, never committed):
- `fixture-flat.json` -- generated via `go run ./scripts/gen-fixture-catalog -out <dir> -shape flat -files 400`; generator printed `nodes=400 bytes=29663`.
- `fixture-dcim.json` -- generated via `go run ./scripts/gen-fixture-catalog -out <dir> -shape dcim -dirs 3 -subdirs 3 -files 4`; generator printed `nodes=48 bytes=4043`.

This directory is outside the repository and outside the scratchpad's guaranteed lifetime -- if a continuation needs it and it's gone, regenerate with the two commands above (same flags reproduce the same `nodes=` totals since the generator is deterministic).

## Task Commits

Task 1 required no code changes (no defect found), so there is no `feat`/`fix` commit for it. This SUMMARY's own commit is the record of Task 1's completion.

## Files Created/Modified

None -- Task 1 found no defects in `CommandPalette.tsx`, `WorkspaceShell.tsx`, `Toolbar.tsx`, or `search_indexed.go` and made no edits.

## Decisions Made

See `key-decisions` in frontmatter: search-term selection rationale (`FILE` / `00`), no-code-change outcome, and the coordinator-approved `wails dev` restart.

## Deviations from Plan

### Auto-fixed Issues

None -- Rules 1-3 were never triggered because every check passed against the existing implementation with no bug, missing functionality, or blocking issue found in the tracer's files.

### Process Deviation (not a Rule 1-3/4 code deviation)

**Stale `wails dev` binary blocked all live evidence at task start.** See "Environment Incident" section above for full detail. This was an environment/process staleness issue, not a defect in any plan file, and was resolved via a `checkpoint:human-action` (process restart requires permissions this executor's sandbox denies) rather than any of the four deviation rules -- no plan file was edited to work around it.

---

**Total deviations:** 0 code deviations. 1 environment/process incident (stale dev binary), resolved via coordinator-approved restart, documented above for future-phase reuse.
**Impact on plan:** None on scope. Task 1's evidence is fully live and fresh (captured against the rebuilt, verified-fresh `wails dev` instance).

## Issues Encountered

The stale-binary incident above. No other issues.

## User Setup Required

None for Task 1. Task 2 requires a human physically at the real macOS StorCat.app window (see checkpoint below) -- that is the plan's designed manual step, not an external-service setup.

## Next Phase Readiness

- PLT-02 and PLT-03 are proven live and can be relied on by 24-03/24-04/24-05 without re-verification.
- PLT-01's click-open, second-press no-op (implied by the always-mounted contract already exercised), Escape-close, debounce, and stale-guard behaviors are proven. **Only the ⌘K keyboard-open path inside a real WKWebView remains unverified** -- Task 2, next.
- The `localStorage` directory-pointing technique and the `curl`-is-not-freshness lesson are recorded above for Phases 25-28's own dev-browser verification work.
- **This plan is not complete.** Do not advance `.planning/STATE.md`'s plan counter or mark `requirements-completed` for PLT-01 until Task 2 returns an observed result.

---
*Phase: 24-cmd-k-command-palette*
*Completed: 2026-08-14 (Task 1 only; plan halted pending Task 2)*
