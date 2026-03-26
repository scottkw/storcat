---
phase: 08-cli-foundation-and-platform-compatibility
plan: 02
subsystem: cli
tags: [go, cli, dispatch, macos, windows, psn-filter, install-script]

# Dependency graph
requires:
  - "08-01: cli/ package with Run(args []string, version string) int"
provides:
  - "main.go CLI dispatch wiring — cli.Run() called before wails.Run()"
  - "filterMacOSArgs() — strips -psn_* Gatekeeper args on all platforms"
  - "runGUI() extraction — current main() body isolated for clean dispatch"
  - "scripts/install-cli.sh — macOS /usr/local/bin/storcat symlink creator"
  - "scripts/build-windows.sh updated with -windowsconsole for visible CLI output"
  - "main_test.go — 6 table-driven tests for filterMacOSArgs"
affects: [phase-09, phase-10]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Dispatch-before-GUI: os.Args switch in main() routes known subcommands to cli.Run(), falls through to runGUI() for all other cases"
    - "filterMacOSArgs: strip -psn_ prefixed args (macOS Gatekeeper injection) before any inspection"
    - "-windowsconsole wails build flag: console subsystem for Windows CLI output visibility"

key-files:
  created:
    - scripts/install-cli.sh
    - main_test.go
  modified:
    - main.go
    - scripts/build-windows.sh

key-decisions:
  - "Use -windowsconsole build flag (not AttachConsole) — simpler, no runtime complexity, acceptable console flash for technical CLI users"
  - "filterMacOSArgs applied before dispatch on all platforms — one function, no-op on Windows/Linux"
  - "Unknown args fall through to runGUI() (not error/exit) — preserves wails dev hot-reload compatibility"

# Metrics
duration: 2min
completed: 2026-03-26
---

# Phase 8 Plan 02: Platform Compatibility and CLI Wiring Summary

**main.go CLI dispatch wiring with -psn_* filtering, runGUI() extraction, macOS install script, and Windows console build flag**

## Performance

- **Duration:** 2 min
- **Started:** 2026-03-26T14:54:42Z
- **Completed:** 2026-03-26T14:56:00Z
- **Tasks:** 3
- **Files modified:** 4 (main.go, scripts/build-windows.sh, scripts/install-cli.sh created, main_test.go created)

## Accomplishments

- Modified `main.go` to dispatch known CLI subcommands to `cli.Run()` before `wails.Run()` is called
- Extracted entire GUI startup into `runGUI()` for clean separation of concerns
- Added `filterMacOSArgs()` to strip `-psn_*` Gatekeeper-injected args before dispatch
- Created `scripts/install-cli.sh` for creating `/usr/local/bin/storcat` symlink after macOS .app install
- Updated `scripts/build-windows.sh` with `-windowsconsole` flag so CLI stdout/stderr is visible in Windows terminals
- Added 6 table-driven tests for `filterMacOSArgs` covering psn removal, edge cases, and double-dash non-match

## Task Commits

1. **Task 1 — main.go dispatch** - `6584e723` (feat(08-02): add CLI dispatch, psn filtering, and runGUI extraction in main.go)
2. **Task 2 — scripts** - `33f1070e` (feat(08-02): add macOS install-cli.sh and update Windows build for console output)
3. **Task 3 — tests** - `5e0b1a0e` (test(08-02): add filterMacOSArgs tests for psn filtering edge cases)

## Files Created/Modified

- `main.go` — New main() with dispatch, filterMacOSArgs(), runGUI() extraction, cli package import
- `scripts/install-cli.sh` — macOS symlink installer for /usr/local/bin/storcat (executable)
- `scripts/build-windows.sh` — Added -windowsconsole flag and trade-off comment
- `main_test.go` — 6 table-driven filterMacOSArgs tests in package main

## Decisions Made

- `-windowsconsole` over `AttachConsole`: console flag is simpler (no runtime code), no platform-specific syscall needed, acceptable trade-off (brief console flash when double-clicking GUI on Windows)
- `filterMacOSArgs` applied universally: one function on all platforms, no-op on Windows/Linux, no build tags needed
- Unrecognized args fall through to GUI: preserves `wails dev` hot-reload compatibility (wails injects its own flags)

## Deviations from Plan

None — plan executed exactly as written.

## Known Stubs

None — this plan has no UI stubs or placeholder data.

## Self-Check: PASSED

- FOUND: main.go (with cli.Run dispatch, filterMacOSArgs, runGUI)
- FOUND: scripts/install-cli.sh (executable, contains ln -sf)
- FOUND: scripts/build-windows.sh (contains -windowsconsole)
- FOUND: main_test.go (6 tests, all passing)
- FOUND commit: 6584e723
- FOUND commit: 33f1070e
- FOUND commit: 5e0b1a0e

---
*Phase: 08-cli-foundation-and-platform-compatibility*
*Completed: 2026-03-26*
