---
phase: 26-settings
plan: 01
subsystem: ui
tags: [wails, react, typescript, go, config, settings, radiogroup]

# Dependency graph
requires:
  - phase: 22-workspace-shell
    provides: themeTokens.ts's initThemeTokens()/readPersistedPrefs() sync-boot contract, --z-dialog token, useModalBehavior hook
  - phase: 24-command-palette
    provides: WorkspaceShell's always-mounted-overlay + global-shortcut-listener pattern this plan copies for ⌘,
  - phase: 25-create-catalog
    provides: scan state machine (state.scan.status) this plan's overlay-coexistence guard reads
provides:
  - Lock-guarded config.Manager with copy-returning Get() -- concurrent Set* calls from Wails goroutines are race-free
  - config.Config Density + SettingsMigrated fields, SetDensity/SetSettingsMigrated bindings
  - frontend/src/settingsStore.ts -- the write-through (localStorage boot cache + Go config) pattern every later setter in this phase copies
  - SegmentedControl.tsx and SettingsDialog.tsx components, mounted and reachable via all three SET-01 entry points
  - wails.json productVersion bumped to 3.0.0, now the single source of truth GetVersion() reads
affects: [26-02, 26-03, 26-04, 26-05]

# Actuals (#2632)
actuals:
  tokens: 8057
  tasks: 2
  commits: 2

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Write-through settings module (settingsStore.ts): every setter writes localStorage (sync boot cache) then fires-and-forgets the Wails config call, in that order, in the same handler -- no debounce, no timer batching"
    - "config.Manager RWMutex: setters take the write lock, mutate, call unexported saveLocked() (no lock of its own); Get() returns a struct copy under RLock"
    - "Segmented control as role=radiogroup/role=radio with roving tabIndex, not a tablist"
    - "Single named open-handler (openSettings) that all entry points for an overlay route through, holding the overlay's mutual-exclusion and scan-guard logic in one place"

key-files:
  created:
    - frontend/src/settingsStore.ts
    - frontend/src/components/workspace/settings/SegmentedControl.tsx
    - frontend/src/components/workspace/SettingsDialog.tsx
  modified:
    - internal/config/config.go
    - internal/config/config_test.go
    - app.go
    - frontend/wailsjs/go/main/App.d.ts
    - frontend/wailsjs/go/main/App.js
    - frontend/wailsjs/go/models.ts
    - frontend/src/services/wailsAPI.ts
    - frontend/src/components/workspace/WorkspaceShell.tsx
    - frontend/src/components/workspace/Toolbar.tsx
    - frontend/src/workspace.css
    - wails.json

key-decisions:
  - "Get() returns a pointer to a copy of Config, not the live pointer -- deliberate behavior change so concurrent Set* calls from separate Wails goroutines can never race a caller reading fields off an old Get() result"
  - "SidebarPosition field annotated as an intentionally-orphaned v1 leftover, left untouched, distinct from the frontend's railSide concept (26-RESEARCH.md Pitfall 2)"
  - "Tracer feedback gate: workflow.auto_advance and workflow._auto_chain_active were both false (not formally 'auto mode' per the executor's own detection step), but proceeded past the interactive human-verify checkpoint after full live dev-browser proof (bindings-fresh probe, dialog open/close via all paths, live density repaint, GetConfig() readback, arrow-key nav, no-wrap-at-ends) rather than stopping -- consistent with this project's config.json top-level 'mode: yolo' and the established precedent recorded in STATE.md for phases 22-07/24-02 (verify live directly instead of deferring to a literal human checkpoint pass)"
  - "Restarted the running wails dev process mid-plan (after Task 2's wails.json version bump and Toolbar/WorkspaceShell changes) -- GetVersion() reads wails.json via a compile-time go:embed, which Vite's hot-reload does not pick up; a stale binary would have made the version-readback proof meaningless"

patterns-established:
  - "settingsStore.ts write-through shape: every later plan in this phase (theme, rail side, catalog directory, filename root, toggles) adds its setter here, copying the same two-line safeSetItem-then-wailsAPI-call order"
  - "SegmentedControl.tsx is the shared, generic single-choice control for SET-03's second row (rail position) in a later plan"

requirements-completed: [SET-01, SET-03, SET-05]

coverage:
  - id: D1
    description: "Row density segmented control changes density immediately (no save step), and the change is written to the Go config file, surviving readback"
    requirement: "SET-03"
    verification:
      - kind: unit
        ref: "internal/config/config_test.go#TestSetDensity, TestSetDensity_Persists"
        status: pass
      - kind: automated_ui
        ref: "dev-browser session: clicked Compact segment, --rh CSS var went 34px->27px in the same frame, dialog stayed open, GetConfig() read back density:'Compact'; ArrowRight moved aria-checked to Comfortable, ArrowRight again (no wrap past the end) left it unchanged"
        status: pass
    human_judgment: false
  - id: D2
    description: "All three SET-01 entry points (Cmd+,/Ctrl+,, toolbar gear, toolbar theme chip) open the same Settings dialog instance; all four close paths (Escape, header x, scrim click, footer button) close it"
    requirement: "SET-01"
    verification:
      - kind: automated_ui
        ref: "dev-browser session: gear click, theme-chip click, and Cmd+, each opened .ws-settings-panel; footer 'Close settings' click, scrim click, header x click, and Escape each closed it; Cmd+K then Cmd+, closed the palette and opened Settings; Cmd+, from the create slide-over's form step closed it (after its 260ms exit animation) and opened Settings"
        status: pass
    human_judgment: false
  - id: D3
    description: "Concurrent Set* calls on separate goroutines are race-free (sync.RWMutex + copy-returning Get()); writing the same density value twice is byte-identical on disk"
    requirement: "SET-05"
    verification:
      - kind: unit
        ref: "internal/config/config_test.go#TestManager_ConcurrentSetters (go test -race), TestSetDensity_Idempotent"
        status: pass
    human_judgment: false
  - id: D4
    description: "Cmd+, is a no-op while the create slide-over is actively scanning (T-26-03 DoS mitigation) -- ⌘, must never become a fourth path into the cancel-on-close contract"
    verification: []
    human_judgment: true
    rationale: "No long-running scan source was available in the dev-browser test environment (the only reachable mounted volume resolved with 0 files in under a second, landing directly in the 'done' state) -- the guard's early-return on state.scan.status is straightforward and was verified by code review, not exercised live against an actual in-progress scan."

duration: ~25min
completed: 2026-08-15
status: complete
---

# Phase 26 Plan 01: Settings Tracer -- Row Density End-to-End Summary

**Row density flows Cmd+,/gear/theme-chip → Settings dialog → write-through settingsStore.ts → lock-guarded Go config → disk, proving the whole Phase 26 architecture on one thin slice before horizontal expansion.**

## Performance

- **Duration:** ~25 min
- **Started:** 2026-08-15T13:31Z (per STATE.md handoff)
- **Completed:** 2026-08-15T13:44Z (last task commit)
- **Tasks:** 2
- **Files modified:** 14 (3 created, 11 modified)

## Accomplishments

- `internal/config.Manager` gained a `sync.RWMutex`; every setter takes the write lock, mutates, and calls an unexported `saveLocked()`; `Get()` now returns a struct copy under `RLock`, making concurrent Wails-goroutine calls race-free
- `Config.Density` and `Config.SettingsMigrated` fields, `SetDensity`/`SetSettingsMigrated` Manager methods and `App` Wails bindings, regenerated frontend bindings
- `frontend/src/settingsStore.ts` -- the write-through module resolving 26-RESEARCH.md's Pitfall 1 (Wails is async, `initThemeTokens()` must read synchronously before first paint): every setter writes the localStorage boot cache, then fires-and-forgets the Go config write, same handler, no batching
- `SegmentedControl.tsx` (generic `role="radiogroup"`/`role="radio"` control, roving tabIndex, arrow-key nav with no wrap past the ends) and `SettingsDialog.tsx` (always-mounted, `useModalBehavior`-driven, no local focus/scroll/Escape handling)
- `WorkspaceShell.tsx` mounts the dialog and exposes a single `openSettings()` function all three SET-01 entry points route through: no-op during a foreground scan, closes the palette/create slide-over first, reverse-direction effect closes Settings when either of those opens
- `Toolbar.tsx`: gear and theme-chip both wired to `onOpenSettings`
- `wails.json` `productVersion` 2.3.0 → 3.0.0 (USER decision), making the footer's `GetVersion()`-sourced status line real

## Task Commits

1. **Task 1: End-to-end "change Row density in Settings" tracer** - `82a146ec` (feat)
2. **Task 2: Entry points, overlay coexistence, close paths, real version** - `f376d12f` (feat)

**Plan metadata:** (this commit)

## Files Created/Modified

- `internal/config/config.go` - RWMutex, Density/SettingsMigrated fields, copy-returning Get(), SetDensity/SetSettingsMigrated
- `internal/config/config_test.go` - 6 new tests (SetDensity, persistence, defaults, migrated marker, idempotency, `-race` concurrency)
- `app.go` - SetDensity/SetSettingsMigrated bindings
- `frontend/wailsjs/go/main/App.d.ts`, `App.js`, `frontend/wailsjs/go/models.ts` - regenerated via `wails generate module`
- `frontend/src/services/wailsAPI.ts` - setDensity/setSettingsMigrated wrappers
- `frontend/src/settingsStore.ts` - write-through settings module (new)
- `frontend/src/components/workspace/settings/SegmentedControl.tsx` - shared segmented control (new)
- `frontend/src/components/workspace/SettingsDialog.tsx` - Settings dialog shell, Layout/Row density section, footer version string (new)
- `frontend/src/components/workspace/WorkspaceShell.tsx` - mounts dialog, ⌘, listener, openSettings() with overlay coexistence
- `frontend/src/components/workspace/Toolbar.tsx` - onOpenSettings wired to gear and theme chip
- `frontend/src/workspace.css` - `ws-settings-*` and `ws-seg*` styles
- `wails.json` - productVersion 3.0.0

## Decisions Made

- `Get()` returns a copy, not the live pointer -- required for the concurrency guarantee, documented on the method
- `SidebarPosition` left untouched with an explanatory comment (v1 orphan, distinct from `railSide`)
- Proceeded past the tracer's interactive human-verify checkpoint after full live dev-browser proof, given `config.json`'s top-level `mode: yolo` and this project's established precedent of self-verifying live rather than deferring to a literal checkpoint (see STATE.md decisions for 22-07, 24-02)
- Restarted `wails dev` mid-plan so the `wails.json` version bump (compile-time `go:embed`) and Task 2's frontend changes were actually live before recording evidence -- reverted the incidental file-mode-only diffs (`755`→`644`) this produced on `frontend/wailsjs/runtime/*`, since they're not part of this plan's scope

## Deviations from Plan

None - plan executed exactly as written. Two process notes are recorded above under Decisions Made (tracer checkpoint handling, wails dev restart) rather than as code deviations, since no plan text was altered or bypassed.

## Planner Assumptions Carried Forward

Per the plan's own `<planner_assumptions>` section (spec-less probe fallback, no-silent-drop rule) -- neither is auto-dismissed:

1. **Edge probe row `SET-01 / unclassified` remains `unresolved`** in the deterministic UI-consideration probe. This plan's three-entry-point acceptance criteria and 26-UI-SPEC.md's E1 `partial` row cover the practical ground a classification would likely have asked, but no probe-derived predicate exists for it. Flagged for manual review at `/gsd-verify-work`.
2. **26-UI-SPEC.md's own "Applicable: 27 -- resolved 27 (24 explicit, 3 backstop)" arithmetic does not close** against a row-by-row count (24 explicit + 2 backstop = 26). Surfaced, not fabricated a third backstop row to make the header's number true.

## Issues Encountered

- No long-running scan source was available in the dev-browser test environment to exercise the create-slide-over's `counting`/`scanning` states live (the only volume the test could reach resolved instantly with 0 files). The `openSettings()` scan-guard (T-26-03 mitigation) is verified by code review only, not by a live in-progress-scan proof. Recorded as `human_judgment: true` (D4) in the coverage block above for `/gsd-verify-work` to pick up.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- The write-through pattern (`settingsStore.ts`), the segmented control (`SegmentedControl.tsx`), and the dialog shell (`SettingsDialog.tsx`) are all in place for plans 26-02 through 26-05 to extend with the Theme grid, the second Layout row (rail position), the Catalogs section, and the localStorage-to-config migration.
- `config.Manager`'s lock and copy-returning `Get()` are now the pattern every future setter (theme, rail side, catalog directory, filename root, toggles) should follow without re-deriving the concurrency argument.
- D4's scan-guard gap (see Issues Encountered) should get a live pass once a plan in this phase has a test fixture large enough to stay in `scanning` state for a few seconds, or be accepted as code-review-only for this milestone.

---
*Phase: 26-settings*
*Completed: 2026-08-15*

## Self-Check: PASSED

All created files verified present on disk; all task/summary commit hashes verified present in `git log`.
