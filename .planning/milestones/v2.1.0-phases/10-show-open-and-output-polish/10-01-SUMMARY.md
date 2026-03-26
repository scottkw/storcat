---
phase: 10-show-open-and-output-polish
plan: "01"
subsystem: cli
tags: [cli, show, tree, color, depth]
dependency_graph:
  requires: [cli/output.go, internal/search/service.go, pkg/models/catalog.go]
  provides: [cli/show.go - runShow command with tree rendering]
  affects: [cli/stubs.go - runShow stub removed]
tech_stack:
  added: [github.com/fatih/color v1.18.0 (direct)]
  patterns: [interspersed flag parsing with value-pair grouping, recursive tree walker, TDD red-green]
key_files:
  created:
    - cli/show.go
    - cli/show_test.go
  modified:
    - cli/stubs.go
    - cli/cli_test.go
    - go.mod
    - go.sum
decisions:
  - "Depth semantics: maxDepth=0 shows root only (direct children skipped); maxDepth=1 shows root+children; -1 unlimited"
  - "Interspersed flag parsing extended to handle --depth N value pairs (value args not starting with - consumed with their flag)"
  - "dirColor var at package level (created once, not per-call) per plan spec"
metrics:
  duration_minutes: 20
  completed_date: "2026-03-26"
  tasks_completed: 1
  tasks_total: 1
  files_changed: 5
requirements:
  - CLCM-04
  - CLOF-02
  - CLOF-04
  - CLOF-05
  - CLOF-06
---

# Phase 10 Plan 01: Show Command Summary

**One-liner:** `storcat show` displays catalog directory trees with Unicode connectors, bold-blue directory coloring via fatih/color, and `--depth N` / `--json` / `--no-color` flags.

## What Was Built

Implemented the `storcat show <catalog.json>` CLI command replacing the stub in `cli/stubs.go`. The command:

- Prints the root directory name on the first line (colorized blue+bold on TTY)
- Recursively prints the full directory tree with `├──` / `└──` connectors and `│   ` / `    ` prefix continuation
- `--depth N`: limits output depth (0 = root only, 1 = root + immediate children, -1 = unlimited)
- `--json`: outputs indented JSON of the full catalog (depth flag ignored)
- `--no-color`: disables ANSI escape sequences (also auto-disabled when `NO_COLOR` env var is set via fatih/color init)
- Input validation: missing arg exits 2, non-.json extension exits 2, unreadable file exits 1

## Tasks

| Task | Name | Commit | Files |
|------|------|--------|-------|
| 1 (RED) | Add failing show tests | 133ddd5c | cli/show_test.go (+249 lines) |
| 1 (GREEN) | Implement show command | ca0ae28e | cli/show.go, cli/stubs.go, cli/cli_test.go, go.mod |

## Verification

```
go test ./cli/... -run TestRunShow -v   → 11/11 tests pass
go test ./cli/... -v                    → 55/55 tests pass (no regressions)
go vet ./cli/...                        → no issues
go build ./cli/...                      → compiles cleanly
```

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Interspersed flag parsing didn't handle --depth N value pairs**
- **Found during:** Task 1 GREEN — TestRunShow_Depth0 and TestRunShow_Depth1 failed
- **Issue:** The list.go-style flag splitter categorized all non-`-` args as positionals. With `--depth 1`, the `1` was classified as positional (not a flag value), leaving `--depth` without its argument, causing the FlagSet to emit a usage error.
- **Fix:** Extended the splitter to track value-taking flags (`--depth`, `-depth`) and consume the next non-flag arg as the flag's value.
- **Files modified:** cli/show.go (interspersed flag splitter logic)
- **Commit:** ca0ae28e

**2. [Rule 1 - Bug] Depth semantics: items at depth boundary were printed instead of skipped**
- **Found during:** Task 1 GREEN — TestRunShow_Depth0 showed children; TestRunShow_Depth1 showed grandchildren
- **Issue:** Depth check was placed after printing (only prevented recursion), so items at the boundary were still rendered.
- **Fix:** Moved `maxDepth >= 0 && currentDepth >= maxDepth` check to the top of `printTree` to skip both printing and recursion when depth limit is reached.
- **Files modified:** cli/show.go (printTree function)
- **Commit:** ca0ae28e

**3. [Rule 2 - Update] TestRun_StubShow updated to match implemented behavior**
- **Found during:** Task 1 GREEN — existing test expected exit 1 + "not yet implemented"
- **Issue:** The stub test would fail once `runShow` was implemented.
- **Fix:** Updated `TestRun_StubShow` to expect exit 2 + "catalog file argument required" (matching the implemented validation).
- **Files modified:** cli/cli_test.go
- **Commit:** ca0ae28e

## Known Stubs

None — `show` is fully implemented. `runOpen` remains a stub in `cli/stubs.go` and will be implemented in Plan 02.

## Self-Check: PASSED

- [x] cli/show.go exists and contains `func runShow`, `func printTree`, `var dirColor`, `color.NoColor = true`, `svc.LoadCatalog(`, `printJSON(root)`, `strings.HasSuffix`
- [x] cli/show_test.go contains all 11 required test functions
- [x] stubs.go does NOT contain `func runShow`
- [x] Commits 133ddd5c (test RED) and ca0ae28e (feat GREEN) exist in git log
- [x] `go test ./cli/... -run TestRunShow -v` passes (11/11)
- [x] `go test ./cli/... -v` passes (55/55, no regressions)
