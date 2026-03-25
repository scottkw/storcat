---
phase: 03-config-manager
plan: 01
subsystem: config
tags: [go, config, tdd, window-state]
dependency_graph:
  requires: []
  provides:
    - "Config.WindowX / WindowY fields (json: windowX/windowY)"
    - "Config.WindowPersistenceEnabled field (json: windowPersistenceEnabled)"
    - "Manager.SetWindowPosition(x, y int) error"
    - "Manager.SetWindowPersistence(enabled bool) error"
    - "Manager.GetWindowPersistence() bool"
  affects:
    - "internal/config/config.go"
    - "internal/config/config_test.go"
tech_stack:
  added: []
  patterns:
    - "TDD: RED (failing tests) → GREEN (minimal impl) → REFACTOR (none needed)"
    - "os.MkdirTemp for test isolation (each test gets own temp dir)"
    - "newTestManager helper creates Manager with temp configPath"
key_files:
  created:
    - "internal/config/config_test.go"
  modified:
    - "internal/config/config.go"
decisions:
  - "WindowPersistenceEnabled explicitly set to true in DefaultConfig() to override Go zero-value false"
  - "No refactor phase needed — changes are purely additive"
metrics:
  duration: "~10 minutes"
  completed: "2026-03-25"
  tasks_completed: 1
  files_changed: 2
requirements:
  - WIN-01
  - WIN-02
  - WIN-03
---

# Phase 03 Plan 01: Config Window State Fields and Methods Summary

**One-liner:** Extended Config struct with WindowX, WindowY, WindowPersistenceEnabled fields plus SetWindowPosition, SetWindowPersistence, GetWindowPersistence methods, with full TDD coverage using temp-dir isolation.

## What Was Built

Extended `internal/config/config.go` with three new fields and three new methods needed by the App layer (Plan 02) to wire window state persistence:

**Fields added to Config struct:**
- `WindowX int` — persisted window X coordinate (`json:"windowX"`)
- `WindowY int` — persisted window Y coordinate (`json:"windowY"`)
- `WindowPersistenceEnabled bool` — whether window state is saved/restored (`json:"windowPersistenceEnabled"`)

**DefaultConfig update:** All three fields initialized explicitly. `WindowPersistenceEnabled: true` was set deliberately — Go zero-value is `false`, which would disable persistence by default and break Electron parity.

**Methods added to Manager:**
- `SetWindowPosition(x, y int) error` — updates WindowX/WindowY and saves to disk
- `SetWindowPersistence(enabled bool) error` — updates WindowPersistenceEnabled and saves to disk
- `GetWindowPersistence() bool` — returns current WindowPersistenceEnabled value

**Test file created:** `internal/config/config_test.go` with 7 tests covering all new behavior.

## Tasks

| Task | Name | Commit | Files |
|------|------|--------|-------|
| RED | Failing tests for config window state fields | 62f2c809 | internal/config/config_test.go |
| GREEN | Window position and persistence fields and methods | 7169af60 | internal/config/config.go |

## Verification Results

```
=== RUN   TestDefaultConfig_WindowFields
--- PASS: TestDefaultConfig_WindowFields (0.00s)
=== RUN   TestSetWindowPosition
--- PASS: TestSetWindowPosition (0.01s)
=== RUN   TestSetWindowPosition_Persists
--- PASS: TestSetWindowPosition_Persists (0.02s)
=== RUN   TestSetWindowPersistence
--- PASS: TestSetWindowPersistence (0.01s)
=== RUN   TestSetWindowPersistence_Persists
--- PASS: TestSetWindowPersistence_Persists (0.00s)
=== RUN   TestGetWindowPersistence_Default
--- PASS: TestGetWindowPersistence_Default (0.01s)
=== RUN   TestWindowPosition_RoundTrip
--- PASS: TestWindowPosition_RoundTrip (0.01s)
PASS
ok  	storcat-wails/internal/config	0.842s
```

All 7 tests pass. No existing tests broken.

## Deviations from Plan

None — plan executed exactly as written. The additive nature of the changes (new fields, new methods, new test file) required no architectural decisions beyond what the plan specified.

## Known Stubs

None. This plan adds concrete implementation — no stubs, no placeholder values, no hardcoded returns.

## Self-Check: PASSED

- `internal/config/config.go` exists and contains all 3 fields and 3 methods
- `internal/config/config_test.go` exists with 7 test functions
- Commits 62f2c809 and 7169af60 verified in git log
- `go test ./internal/config/... -v -count=1` exits 0
