---
phase: 08-cli-foundation-and-platform-compatibility
plan: 01
subsystem: cli
tags: [go, cli, flag, dispatch, subcommands, version]

# Dependency graph
requires: []
provides:
  - "cli/ package with Run(args []string, version string) int entry point"
  - "Switch dispatch for: version, create, search, list, show, open, help, --help, -h"
  - "runVersion() with stdout output and 0 exit code"
  - "Stub handlers for create/search/list/show/open (return 1, 'not yet implemented')"
  - "--help on all commands returns 0 (flag.FlagSet ContinueOnError pattern)"
  - "Exit code contract: 0=success, 1=runtime error, 2=usage error"
  - "Stdout/stderr routing: results to stdout, errors to stderr"
affects: [08-02, phase-09, phase-10]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "flag.FlagSet with ContinueOnError per subcommand (not Cobra)"
    - "Run(args []string, version string) int — version passed as parameter (main package not importable)"
    - "captureOutput test helper using os.Pipe() for stdout/stderr capture"
    - "TDD red/green: test files committed before implementation"

key-files:
  created:
    - cli/cli.go
    - cli/version.go
    - cli/stubs.go
    - cli/cli_test.go
    - cli/version_test.go
  modified: []

key-decisions:
  - "stdlib flag.FlagSet (not Cobra) — locked project decision, binary size constraint"
  - "version passed as parameter to cli.Run() — package main is not importable by cli/"
  - "TestRun_NoCobra skips _test.go files to avoid false positives from assertion strings"

patterns-established:
  - "Pattern: Each subcommand uses flag.NewFlagSet(name, flag.ContinueOnError) with custom Usage func"
  - "Pattern: Check err == flag.ErrHelp and return 0 (not 2) for --help handling"
  - "Pattern: captureOutput() test helper replaces os.Stdout/os.Stderr via os.Pipe() for testing"

requirements-completed: [CLIP-02, CLIP-03, CLIP-04, CLIP-05, CLIP-06, CLCM-06]

# Metrics
duration: 2min
completed: 2026-03-26
---

# Phase 8 Plan 01: CLI Package Foundation Summary

**stdlib flag.FlagSet CLI dispatch package with Run() entry point, version command, and 5 stub handlers — zero external dependencies, full --help/exit-code/stdout-stderr contract**

## Performance

- **Duration:** 2 min
- **Started:** 2026-03-26T14:48:51Z
- **Completed:** 2026-03-26T14:51:00Z
- **Tasks:** 1 (TDD: RED + GREEN + inline fix)
- **Files modified:** 5 created

## Accomplishments

- Created `cli/` package with `Run(args []string, version string) int` as the sole public API
- Implemented `storcat version` command printing "storcat X.Y.Z" to stdout, returning exit code 0
- Stub handlers for create/search/list/show/open with proper `--help` support (return 0) and "not yet implemented" on stderr (return 1)
- All 21 tests pass: dispatch, version, help, exit codes, stdout/stderr separation, and import safety
- Zero new dependencies: stdlib `flag`/`fmt`/`os` only

## Task Commits

TDD execution had 3 commits:

1. **RED — Failing tests** - `5cf4e4e9` (test: add failing tests for cli/ dispatch, version, stubs)
2. **GREEN — Implementation** - `0468d426` (feat: implement cli/ package)
3. **Fix — TestRun_NoCobra** - `5a0b44d1` (fix: skip test files in import scan)

## Files Created/Modified

- `cli/cli.go` - Run() entry point, printUsage(), switch dispatch for all 8 cases
- `cli/version.go` - runVersion() using flag.FlagSet ContinueOnError, prints to stdout
- `cli/stubs.go` - Five stub handlers (create/search/list/show/open) with per-command help text
- `cli/cli_test.go` - 17 test functions: dispatch, help, exit codes, stdout/stderr separation, no-cobra check
- `cli/version_test.go` - 5 version tests: table-driven, help to stderr check

## Decisions Made

- `stdlib flag.FlagSet` (not Cobra): locked project decision from STATE.md; ~2MB Cobra overhead unacceptable for size-sensitive binary
- `version string` as parameter: `version.go` uses `//go:embed wails.json` in `package main` which is not importable; parameter passing is the idiomatic solution
- `TestRun_NoCobra` skips `_test.go` files: test assertions contain the string "cobra" as expected failure message; only non-test files need the import check

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] TestRun_NoCobra false positive on _test.go files**
- **Found during:** Task 1 (GREEN phase test run)
- **Issue:** NoCobra test scanned all `.go` files including test files; test assertions contain the string "cobra" causing the test to fail on itself
- **Fix:** Updated test to skip files ending in `_test.go` when checking for forbidden imports; use quoted import string form `"github.com/spf13/cobra"` instead of bare substring
- **Files modified:** `cli/cli_test.go`
- **Verification:** `go test ./cli/... -v -count=1` — all 21 tests pass
- **Committed in:** `5a0b44d1`

---

**Total deviations:** 1 auto-fixed (Rule 1 - test correctness bug)
**Impact on plan:** Minor fix to test correctness. No scope creep. Implementation is exactly as specified.

## Issues Encountered

- `golangci-lint` not installed in this environment — skipped linting step. `go vet ./cli/...` passed with no issues.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- `cli/` package is ready for Phase 8 Plan 02 (main.go dispatch wiring, -psn_ filter, Windows console, install script)
- Phase 9 can import `cli/` and add real implementations for create/search/list/show/open by replacing stubs
- No blockers

## Self-Check: PASSED

- FOUND: cli/cli.go
- FOUND: cli/version.go
- FOUND: cli/stubs.go
- FOUND: cli/cli_test.go
- FOUND: cli/version_test.go
- FOUND commit: 5cf4e4e9
- FOUND commit: 0468d426
- FOUND commit: 5a0b44d1

---
*Phase: 08-cli-foundation-and-platform-compatibility*
*Completed: 2026-03-26*
