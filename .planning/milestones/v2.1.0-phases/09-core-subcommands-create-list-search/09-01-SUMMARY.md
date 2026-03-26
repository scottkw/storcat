---
phase: 09-core-subcommands-create-list-search
plan: 01
subsystem: cli
tags: [go, cli, tablewriter, storcat-list, json-output]

# Dependency graph
requires:
  - phase: 08-cli-foundation-and-platform-compatibility
    provides: cli.Run() dispatch, captureOutput test helper, runVersion pattern

provides:
  - storcat list command with table and JSON output modes
  - cli/output.go with printJSON() and formatBytes() shared helpers
  - tablewriter v1.1.4 dependency for table rendering
  - FormatBytes() exported from internal/catalog/service.go
  - 7 TestRunList_* tests covering all acceptance criteria

affects:
  - 09-02 (search and create commands reuse output.go helpers)

# Tech tracking
tech-stack:
  added:
    - github.com/olekukonko/tablewriter v1.1.4
  patterns:
    - Interspersed flag parsing: pre-separate positional args from flags before fs.Parse to support `cmd <arg> --flag` ordering
    - cli/ package independence: formatBytes duplicated in cli/output.go to avoid importing internal/ packages
    - printJSON via json.NewEncoder with SetIndent("", "  ") for consistent JSON output

key-files:
  created:
    - cli/output.go
    - cli/list.go
    - cli/list_test.go
  modified:
    - cli/stubs.go (removed runList stub)
    - cli/cli_test.go (replaced stub test with implemented behavior test)
    - internal/catalog/service.go (exported FormatBytes)
    - go.mod / go.sum (added tablewriter and transitive deps)

key-decisions:
  - "Use tablewriter v1.1.4 builder API (Header/Append/Render) not old v0.x SetHeader/SetBorder/SetColumnSeparator"
  - "Pre-separate positional args from flags to support storcat list <dir> --json ordering"
  - "cli/output.go duplicates formatBytes instead of importing internal/catalog to keep cli/ package independent"

patterns-established:
  - "Interspersed flag pattern: scan args for '-' prefix, route to flagArgs vs positional separately"
  - "printJSON: json.NewEncoder(os.Stdout) + SetIndent(empty array guard for nil slices)"

requirements-completed: [CLCM-03, CLOF-01]

# Metrics
duration: 4min
completed: 2026-03-26
---

# Phase 09 Plan 01: Core Subcommands Create/List/Search — List Command Summary

**storcat list command with table/JSON output using tablewriter v1.1.4, shared printJSON/formatBytes helpers in cli/output.go**

## Performance

- **Duration:** 4 min
- **Started:** 2026-03-26T15:32:06Z
- **Completed:** 2026-03-26T15:36:34Z
- **Tasks:** 3
- **Files modified:** 7

## Accomplishments
- Added tablewriter v1.1.4 dependency and created cli/output.go with printJSON() and formatBytes() helpers for reuse by Plan 02
- Implemented storcat list command with table output (Name, Title, Size, Modified columns) and --json flag
- Added 7 tests covering cwd default, nonexistent dir error, empty dir, table output, JSON output, empty JSON array, --help

## Task Commits

1. **Task 1: Add tablewriter dep, create output.go helpers, export FormatBytes** - `9c73e973` (feat)
2. **Task 2: Implement storcat list command** - `581fcf92` (feat)
3. **Task 3: Add list command tests and fix interspersed flag parsing** - `5d727584` (test)

## Files Created/Modified
- `cli/output.go` - printJSON() and formatBytes() helpers shared across all CLI commands
- `cli/list.go` - runList() and printListTable() using search.BrowseCatalogs
- `cli/list_test.go` - 7 TestRunList_* tests covering all acceptance criteria
- `cli/stubs.go` - removed runList stub (now in list.go)
- `cli/cli_test.go` - replaced TestRun_StubList with TestRun_ListDefaultsCwd
- `internal/catalog/service.go` - added FormatBytes() exported wrapper
- `go.mod` / `go.sum` - tablewriter v1.1.4 and transitive deps

## Decisions Made
- tablewriter v1.1.4 uses completely different API from v0.x (builder pattern, `Header()` not `SetHeader()`). Used correct v1.1.4 API with `tw.Rendition` for border configuration.
- Flags after positional args (`storcat list <dir> --json`) require pre-separating args since stdlib flag.FlagSet stops at first non-flag. Pre-scan approach keeps stdlib flag while supporting natural CLI ordering.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] tablewriter v1.1.4 API incompatible with plan's v0.x method calls**
- **Found during:** Task 1 (output.go creation) and Task 2 (list.go implementation)
- **Issue:** Plan specified `table.SetHeader()`, `table.SetBorder()`, `table.SetColumnSeparator()` which are v0.x API methods. tablewriter v1.1.4 uses `table.Header()`, `table.Append()`, `table.Render()` with `tw.Rendition` for border config.
- **Fix:** Used correct v1.1.4 API throughout list.go; removed `newTable` helper from output.go that used old API
- **Files modified:** cli/output.go, cli/list.go
- **Verification:** `go build ./cli/...` exits 0; all tests pass
- **Committed in:** 581fcf92 (Task 2 commit)

**2. [Rule 1 - Bug] Flags after positional arg not parsed (storcat list <dir> --json)**
- **Found during:** Task 3 (list_test.go TestRunList_JSON)
- **Issue:** stdlib flag.FlagSet stops parsing at first non-flag argument, so `list <dir> --json` treated `--json` as unparsed positional. TestRunList_JSON and TestRunList_JSON_EmptyDir failed.
- **Fix:** Pre-separate args into positional (no `-` prefix) and flag args, parse only flagArgs with fs.Parse
- **Files modified:** cli/list.go
- **Verification:** `go test ./cli/... -run TestRunList -v -count=1` — all 7 tests pass
- **Committed in:** 5d727584 (Task 3 commit)

---

**Total deviations:** 2 auto-fixed (2 Rule 1 bugs)
**Impact on plan:** Both necessary for correct functionality. No scope creep. Plan API references used v0.x tablewriter methods but v1.1.4 was specified in the version pin — used the correct versioned API.

## Issues Encountered
- First `go get tablewriter` ran before output.go existed, so tablewriter wasn't added to go.mod. Re-ran after creating the file with the import.

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- cli/output.go is ready with printJSON() and formatBytes() for Plan 02 (search and create)
- tablewriter dependency is in go.mod for Plan 02 table output
- Interspersed flag parsing pattern established for search and create commands to adopt
- All existing tests pass, no regressions

---
*Phase: 09-core-subcommands-create-list-search*
*Completed: 2026-03-26*

## Self-Check: PASSED

- FOUND: cli/output.go
- FOUND: cli/list.go
- FOUND: cli/list_test.go
- FOUND: .planning/phases/09-core-subcommands-create-list-search/09-01-SUMMARY.md
- FOUND: commit 9c73e973 (Task 1)
- FOUND: commit 581fcf92 (Task 2)
- FOUND: commit 5d727584 (Task 3)
