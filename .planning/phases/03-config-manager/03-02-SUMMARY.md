---
phase: 03-config-manager
plan: 02
subsystem: config
tags: [go, wails, window-persistence, typescript, frontend]

# Dependency graph
requires:
  - phase: 03-01
    provides: Config struct with WindowX/WindowY/WindowPersistenceEnabled fields and SetWindowPosition/SetWindowPersistence/GetWindowPersistence methods on Manager

provides:
  - App.GetWindowPersistence bound method callable from frontend
  - App.SetWindowPersistence bound method callable from frontend
  - App.SetWindowPosition bound method callable from frontend
  - App.domReady lifecycle hook restoring window size+position on launch
  - App.beforeClose lifecycle hook saving window size+position on close
  - main.go uses config-based initial window dimensions (no resize flash)
  - Frontend wailsAPI.ts calls real Go methods (no stubs)

affects: [phase-04, phase-05, phase-06, phase-07]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Wails OnDomReady for post-render window restoration (not OnStartup)"
    - "Wails OnBeforeClose returning false to save state then allow close"
    - "Read config before wails.Run to set initial window dimensions"
    - "Frontend wailsAPI error handling: console.warn for reads, console.error for writes"

key-files:
  created: []
  modified:
    - app.go
    - main.go
    - frontend/src/services/wailsAPI.ts
    - frontend/wailsjs/go/main/App.js
    - frontend/wailsjs/go/main/App.d.ts
    - frontend/wailsjs/go/models.ts

key-decisions:
  - "Use OnDomReady (not OnStartup) for window restoration — window not yet rendered at OnStartup"
  - "beforeClose returns false to allow close — true would cancel the close event"
  - "Skip position restore when both X and Y are 0 — let OS choose placement for fresh installs"
  - "config.windowX/windowY added to models.ts TypeScript class by wails generate module"

patterns-established:
  - "Lifecycle hooks: domReady (restore state) and beforeClose (save state) pattern for window persistence"
  - "Nil guard pattern on all bound methods: if a.configManager == nil { return safeDefault }"

requirements-completed: [WIN-01, WIN-02, WIN-03]

# Metrics
duration: 10min
completed: 2026-03-25
---

# Phase 03 Plan 02: Config Manager — App Layer Wiring Summary

**Window persistence wired end-to-end: Go lifecycle hooks save/restore size+position, Wails bindings regenerated, frontend stubs replaced with real Go calls**

## Performance

- **Duration:** ~10 min
- **Started:** 2026-03-25T16:22:00Z
- **Completed:** 2026-03-25T16:32:24Z
- **Tasks:** 2
- **Files modified:** 6

## Accomplishments

- Added GetWindowPersistence, SetWindowPersistence, SetWindowPosition bound methods to App (nil-guard pattern)
- Added domReady lifecycle hook that restores saved window size and position when persistence enabled
- Added beforeClose lifecycle hook that saves window size and position before close when persistence enabled
- Updated main.go to read config before wails.Run and use config-based dimensions (prevents resize flash)
- Regenerated Wails bindings (App.js, App.d.ts, models.ts) with new methods and config fields
- Replaced two frontend stubs in wailsAPI.ts with real Go method calls and proper error handling

## Task Commits

Each task was committed atomically:

1. **Task 1: Add App bound methods and lifecycle hooks** - `8f8b564c` (feat)
2. **Task 2: Regenerate Wails bindings and replace frontend stubs** - `f9f9d321` (feat)
3. **Task 2 (models update): Update wails models.ts with new config fields** - `b0a5550d` (chore)

**Plan metadata:** (docs commit to follow)

## Files Created/Modified

- `app.go` - Added GetWindowPersistence, SetWindowPersistence, SetWindowPosition bound methods; domReady and beforeClose lifecycle hooks
- `main.go` - Added config-based initial window size; wired OnDomReady and OnBeforeClose lifecycle hooks
- `frontend/src/services/wailsAPI.ts` - Added GetWindowPersistence/SetWindowPersistence imports; replaced stubs with real Go calls
- `frontend/wailsjs/go/main/App.js` - Regenerated: added GetWindowPersistence, SetWindowPersistence, SetWindowPosition exports
- `frontend/wailsjs/go/main/App.d.ts` - Regenerated: added typed declarations for three new methods
- `frontend/wailsjs/go/models.ts` - Regenerated: added windowX, windowY, windowPersistenceEnabled to config.Config class

## Decisions Made

- Used OnDomReady (not OnStartup) for window restoration — per RESEARCH.md Pitfall 1: window is not yet rendered at OnStartup
- beforeClose returns `false` to allow close — returning `true` would cancel/block the close event
- Position restore is skipped when both X and Y are 0 — lets OS choose placement for first-time users
- models.ts update committed separately from App.js/App.d.ts since it was a generated side-effect of wails generate module

## Deviations from Plan

**1. [Rule 2 - Missing Critical] Committed models.ts update separately**
- **Found during:** Task 2 (wails generate module)
- **Issue:** wails generate module also updated models.ts with new config fields (windowX, windowY, windowPersistenceEnabled) — not explicitly called out in task files list but was a necessary generated artifact
- **Fix:** Committed models.ts in a separate chore commit after the main Task 2 feat commit
- **Files modified:** frontend/wailsjs/go/models.ts
- **Verification:** TypeScript types now include all new config fields
- **Committed in:** b0a5550d

---

**Total deviations:** 1 (generated artifact committed separately)
**Impact on plan:** No scope creep. models.ts update is a required side-effect of the wails generate module command.

## Issues Encountered

None - plan executed cleanly. Go builds, all config tests pass, bindings regenerated correctly.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Window persistence is fully functional end-to-end: toggle in settings, save on close, restore on launch
- Phase 04 (or next active plan) can proceed — all WIN-01/02/03 requirements satisfied
- No blockers

---
*Phase: 03-config-manager*
*Completed: 2026-03-25*
