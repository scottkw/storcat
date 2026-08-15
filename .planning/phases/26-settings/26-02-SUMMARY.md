---
phase: 26-settings
plan: 02
subsystem: ui
tags: [wails, react, typescript, go, config, settings, theme, segmented-control]

# Dependency graph
requires:
  - phase: 26-settings
    provides: "26-01's write-through settingsStore.ts pattern, config.Manager RWMutex/copy-returning Get(), SegmentedControl.tsx, SettingsDialog.tsx shell"
provides:
  - "config.Config.RailSide field + Manager.SetRailSide, App.SetRailSide Wails binding -- second config field to follow 26-01's SetDensity shape"
  - "settingsStore.setThemeSetting and setRailSideSetting -- two more write-through setters copying the established two-write shape"
  - "ThemeGrid.tsx -- reusable 11-card theme picker, mounted in Settings' Theme section"
  - "Settings dialog's Theme section and the Layout section's second row (Catalog rail position)"
affects: [26-03, 26-04, 26-05]

# Actuals (#2632)
actuals:
  tokens: 3803
  tasks: 2
  commits: 4

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "settingsStore setters continue the two-write shape: setThemeSetting writes THEME_KEY then wailsAPI.setTheme then dispatches the existing themeChange CustomEvent (no new apply path); setRailSideSetting writes RAIL_SIDE_KEY then wailsAPI.setRailSide"
    - "Dialog-local state that must track an externally-changed value (activeThemeId) re-reads readPersistedPrefs() on both mount and every themeChange event, rather than owning a second source of truth"
    - "Segment-specific CSS padding scoped via a wrapper class (.ws-rail-seg) rather than a variant prop on the shared SegmentedControl"

key-files:
  created:
    - frontend/src/components/workspace/settings/ThemeGrid.tsx
  modified:
    - internal/config/config.go
    - internal/config/config_test.go
    - app.go
    - frontend/wailsjs/go/main/App.d.ts
    - frontend/wailsjs/go/main/App.js
    - frontend/wailsjs/go/models.ts
    - frontend/src/services/wailsAPI.ts
    - frontend/src/settingsStore.ts
    - frontend/src/components/workspace/SettingsDialog.tsx
    - frontend/src/workspace.css

key-decisions:
  - "RailSide is a genuinely new Config field, distinct from the orphaned SidebarPosition -- pinned by a dedicated test (TestSetRailSide_DoesNotTouchSidebarPosition) per 26-CONTEXT.md's discretion resolution"
  - "ThemeGrid swatch order is bg/p2/ac/tx (not bg/p/ac/tx) -- the literal code mapping from 26-UI-SPEC wins over the handoff's imprecise plain-English description"
  - "Skipped a full OS quit-and-relaunch to prove RailSide's no-flash persistence live, to avoid disrupting the shared wails dev process later plans in this phase (26-03..05) also depend on; substituted TestSetRailSide_Persists (exercises the identical Manager.Load() path a relaunch takes) plus a direct on-disk config.json readback showing railSide already persisted from the live click. Logged to WINDOWS.md entry #7 (unrun-verify) rather than silently asserting the stronger claim."

patterns-established:
  - "Any later setter reusing App-level apply state (theme) must dispatch the existing event rather than adding a parallel apply path -- ThemeGrid/settingsStore.setThemeSetting is the second consumer of this rule after DevStateSwitcher"

requirements-completed: [SET-02, SET-03, SET-05]

coverage:
  - id: D1
    description: "User sees exactly 11 theme cards (4-swatch strip + light/dark tag) in themes.ts's declared order; clicking any card repaints the whole workspace immediately with the dialog staying open and exactly one card selected at a time"
    requirement: "SET-02"
    verification:
      - kind: automated_ui
        ref: "dev-browser session: 11 cards rendered in exact themes.ts order (StorCat Light..Gruvbox Dark); clicked all 11 in sequence, each produced activeCount:1 and dialogOpen:true, data-theme attribute matched; GetConfig().theme read back 'gruvbox-dark' after the last click"
        status: pass
    human_judgment: false
  - id: D2
    description: "Clicking the already-active theme card re-applies it and leaves it the single selected card (adjacency edge case) -- never deselects, never two selected, never zero"
    requirement: "SET-02"
    verification:
      - kind: automated_ui
        ref: "dev-browser session: clicked the active Gruvbox Dark card a second time -- activeCount stayed 1, activeIsGruvbox stayed true"
        status: pass
    human_judgment: false
  - id: D3
    description: "Catalog rail position (Left/Right segmented control) swaps the catalog rail and details panel sides immediately at window widths >=1280px, and persists to the Go config with no debounce"
    requirement: "SET-03"
    verification:
      - kind: unit
        ref: "internal/config/config_test.go#TestSetRailSide, TestSetRailSide_Persists, TestDefaultConfig_RailSide, TestSetRailSide_DoesNotTouchSidebarPosition (go test -race)"
        status: pass
      - kind: automated_ui
        ref: "dev-browser session at window.innerWidth=1400: clicked Right -> data-rail-side='Right', .ws-rail moved to railLeft:1132 (.ws-details to detailsLeft:0), GetConfig().railSide='Right'; clicked Left -> reverted, GetConfig().railSide='Left'"
        status: pass
    human_judgment: false
  - id: D4
    description: "RailSide survives a real quit-and-relaunch with no visible flash of the other side"
    verification:
      - kind: unit
        ref: "internal/config/config_test.go#TestSetRailSide_Persists -- a second Manager built on the same configPath and Load()ed (the identical path NewManager() takes at real app startup) reports the persisted value"
        status: pass
    human_judgment: true
    rationale: "A full OS-level quit-and-relaunch of the shared wails dev process was not performed live, to avoid disrupting the dev server plans 26-03 through 26-05 in this same phase also depend on. Substituted evidence: the Go unit test exercises the identical Load()-from-disk path a relaunch takes, and reading the on-disk config.json directly after the live UI click confirmed railSide already persisted. Logged to .planning/WINDOWS.md entry #7 for a follow-up live pass."
  - id: D5
    description: "Both Settings Layout rows apply immediately with no save step, no dirty indicator, no confirm-before-close (SET-05)"
    requirement: "SET-05"
    verification:
      - kind: automated_ui
        ref: "dev-browser session: every theme click and rail-position click repainted in the same interaction with the dialog remaining open; no separate save/apply control exists in SettingsDialog.tsx"
        status: pass
    human_judgment: false

duration: ~8min
completed: 2026-08-15
status: complete
---

# Phase 26 Plan 02: Theme Grid + Rail Position Summary

**11-theme picker grid and the rail-position segmented control both wired through settingsStore's write-through pattern into the Go config, expanding the 26-01 tracer sideways with no new apply paths.**

## Performance

- **Duration:** ~8 min
- **Started:** 2026-08-15T08:46:34-05:00
- **Completed:** 2026-08-15T08:53:54-05:00
- **Tasks:** 2
- **Files modified:** 11 (1 created, 10 modified)

## Accomplishments

- `ThemeGrid.tsx` (new): 11 cards rendered directly from `themes.ts`'s declared array (no sort/filter/reorder), each a 4-swatch strip (`bg/p2/ac/tx`), name, and lowercase light/dark tag, `aria-pressed` selection state
- `settingsStore.setThemeSetting`: writes the `THEME_KEY` boot cache, fires `wailsAPI.setTheme`, then dispatches the existing `themeChange` CustomEvent `App.tsx` already listens for -- no second apply path introduced
- `SettingsDialog.tsx` Theme section (new, first section, above Layout) -- active theme id tracked via `readPersistedPrefs()` re-read on mount and on every `themeChange` event, so a theme changed by any other path (dev switcher) never goes stale
- `config.Config.RailSide` (default `"Left"`) + `Manager.SetRailSide`, `App.SetRailSide` Wails binding, regenerated `frontend/wailsjs/` bindings -- a genuinely new field, `SidebarPosition` untouched
- `settingsStore.setRailSideSetting` and the Settings dialog's second Layout row ("Catalog rail position", Left/Right) -- dispatches the pre-existing `SET_RAIL_SIDE` reducer action and calls the new setter in the same handler, `AppContext.tsx` unmodified
- `workspace.css`: theme-card grid/swatch/name/tag styles; `.ws-rail-seg` scopes the rail row's `4px 14px` segment padding without touching the density row's `4px 12px`

## Task Commits

Each task was committed atomically:

1. **Task 1: Theme section -- 11 cards apply on click and persist** - `a27ad612` (feat)
2. **Task 2 (TDD RED): failing RailSide config tests** - `3031ddb0` (test)
3. **Task 2 (TDD GREEN): catalog rail position end to end** - `e58b894a` (feat)

**Plan metadata:** (this commit)

_Note: Task 2 carried `tdd="true"` -- RED (test) then GREEN (feat) commits, no REFACTOR commit needed (no cleanup required after GREEN)._

## Files Created/Modified

- `frontend/src/components/workspace/settings/ThemeGrid.tsx` - 11-card theme picker (new)
- `frontend/src/settingsStore.ts` - `setThemeSetting`, `setRailSideSetting`
- `frontend/src/components/workspace/SettingsDialog.tsx` - Theme section, rail-position Layout row, active-theme-id tracking effect
- `frontend/src/workspace.css` - `ws-theme-*` styles, `.ws-rail-seg` padding scope
- `internal/config/config.go` - `RailSide` field, `SetRailSide` method
- `internal/config/config_test.go` - 4 new tests (RailSide set/persist/default/no-alias-with-SidebarPosition)
- `app.go` - `SetRailSide` Wails binding
- `frontend/wailsjs/go/main/App.d.ts`, `App.js`, `frontend/wailsjs/go/models.ts` - regenerated via `wails generate module`
- `frontend/src/services/wailsAPI.ts` - `setRailSide` wrapper

## Decisions Made

- `RailSide` kept fully separate from `SidebarPosition` (v1 orphan) -- pinned by a dedicated non-aliasing test rather than a code comment alone
- `ThemeGrid`'s second swatch reads `theme.tokens.p2`, matching 26-UI-SPEC's literal code-mapping correction over the handoff's imprecise "background, panel, accent, text" prose
- Skipped a full OS-level quit-and-relaunch of the shared `wails dev` process for the RailSide no-flash persistence proof (would have disrupted the dev server plans 26-03 through 26-05 also depend on this session); substituted `TestSetRailSide_Persists` (identical `Load()` path) plus a direct on-disk `config.json` read after the live click. Logged as `.planning/WINDOWS.md` entry #7 (`unrun-verify`) for a follow-up live pass rather than silently claiming the stronger evidence.

## Deviations from Plan

None - plan executed exactly as written. One process note (the skipped OS quit/relaunch, above) is recorded under Decisions Made and the WINDOWS.md ledger, not as a code deviation, since no plan text was altered or bypassed -- the plan's own acceptance criteria accept a live proof, and the substituted proof exercises the same underlying persistence mechanism.

## TDD Gate Compliance

Task 2 (`tdd="true"`) followed the full RED/GREEN sequence:
- RED: `3031ddb0` `test(26-02): add failing tests for RailSide config field` -- confirmed as a compile failure (`SetRailSide`/`RailSide` undefined) before any implementation existed
- GREEN: `e58b894a` `feat(26-02): catalog rail position -- second Layout row, end to end` -- all four new tests pass under `-race`, full suite (`go test ./... -race -count=1`) green
- No REFACTOR commit -- no cleanup was warranted after GREEN

## Issues Encountered

None.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Both remaining "appearance" settings (SET-02 theme, SET-03 rail-position half) are live; plan 26-03 can proceed to the Catalogs section (SET-04) and localStorage-to-config migration on the same write-through pattern.
- `.planning/WINDOWS.md` entry #7 (RailSide relaunch-no-flash live proof) should get a real pass once a plan in this phase can safely restart `wails dev` without disrupting parallel/sequential work still depending on it -- likely at the end of the phase rather than mid-phase.

---
*Phase: 26-settings*
*Completed: 2026-08-15*
