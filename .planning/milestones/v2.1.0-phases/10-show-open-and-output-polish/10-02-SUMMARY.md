---
phase: 10-show-open-and-output-polish
plan: "02"
subsystem: cli
tags: [cli, open, browser, html, cross-platform]
dependency_graph:
  requires: [cli/open.go, internal/search/service.go, pkg/models/catalog.go, github.com/pkg/browser]
  provides: [cli/open.go - runOpen command with cross-platform browser launch]
  affects: [cli/stubs.go - deleted (runOpen now in open.go)]
tech_stack:
  added: [github.com/pkg/browser v0.0.0-20240102092130-5ac0b6a4141c (direct)]
  patterns: [interspersed flag parsing, HTML path derivation from JSON path, TDD red-green]
key_files:
  created:
    - cli/open.go
    - cli/open_test.go
  modified:
    - cli/cli_test.go
    - go.mod
    - go.sum
  deleted:
    - cli/stubs.go
decisions:
  - "pkg/browser promoted from indirect to direct dependency via go mod tidy after open.go uses it"
  - "fatih/color promoted from indirect to direct (already used by show.go; tidy picks it up)"
  - "stubs.go deleted entirely — all stub commands now fully implemented across Plan 01 and 02"
metrics:
  duration_minutes: 5
  completed_date: "2026-03-26"
  tasks_completed: 2
  tasks_total: 2
  files_changed: 5
requirements:
  - CLCM-05
  - CLPC-05
---

# Phase 10 Plan 02: Open Command Summary

**One-liner:** `storcat open` validates a .json catalog path, derives the sibling .html path, and opens it in the system default browser via `github.com/pkg/browser` (macOS: `open`, Linux: `xdg-open`, Windows: `cmd /c start`).

## What Was Built

Implemented the `storcat open <catalog.json>` CLI command replacing the stub in `cli/stubs.go`. The command:

- Validates: no arg exits 2 with "catalog file argument required"
- Validates: non-.json file exits 2 with "expected a .json catalog file"
- Validates: catalog is readable via `search.NewService().LoadCatalog()` (exits 1 on error)
- Derives HTML path: `strings.TrimSuffix(catalogPath, ".json") + ".html"`
- Checks HTML file exists via `os.Stat` (exits 1 with "HTML file not found: <path>")
- Opens HTML in system browser via `browser.OpenFile(htmlPath)` (exits 1 on launch error)
- `--help` exits 0

Also deleted `cli/stubs.go` — all 6 CLI subcommands are now fully implemented.
Updated `TestRun_StubOpen` in `cli_test.go` to match implemented behavior (exit 2, "catalog file argument required" instead of exit 1, "not yet implemented").

## Tasks

| Task | Name | Commit | Files |
|------|------|--------|-------|
| 1 (RED) | Add failing open tests | 5785bac7 | cli/open_test.go (+99 lines) |
| 1 (GREEN) | Implement open command, delete stubs.go | 44f74dcf | cli/open.go, go.mod, go.sum, stubs.go (deleted) |
| 2 | Update stub tests in cli_test.go | 10e59fb0 | cli/cli_test.go |

## Verification

```
go test ./cli/... -run TestRunOpen -v   → 6/6 tests pass
go test ./cli/... -run TestRun_Stub -v  → 5/5 tests pass (including updated StubOpen)
go test ./cli/... -v                    → all tests pass (no regressions)
go vet ./cli/...                        → no issues
ls cli/stubs.go                         → no such file (confirmed deleted)
go mod tidy                             → no changes (deps already tidy)
```

## Deviations from Plan

None - plan executed exactly as written.

## Known Stubs

None — all 6 CLI subcommands are now fully implemented:
- `storcat version` (Phase 8)
- `storcat list` (Phase 9)
- `storcat search` (Phase 9)
- `storcat create` (Phase 9)
- `storcat show` (Phase 10, Plan 01)
- `storcat open` (Phase 10, Plan 02)

## Self-Check: PASSED

- [x] cli/open.go exists and contains `func runOpen`, `browser.OpenFile(htmlPath)`, `strings.TrimSuffix(catalogPath, ".json") + ".html"`, `strings.HasSuffix`, `svc.LoadCatalog(`
- [x] cli/open_test.go contains `TestRunOpen_NoArg`, `TestRunOpen_MissingHtml`, `TestRunOpen_Help`
- [x] cli/stubs.go does NOT exist (deleted)
- [x] cli_test.go TestRun_StubOpen asserts exit code 2 + "catalog file argument required"
- [x] cli_test.go does NOT contain "not yet implemented" in any assertion
- [x] Commits 5785bac7 (test RED), 44f74dcf (feat GREEN), 10e59fb0 (feat task 2) exist
- [x] `go test ./cli/... -run TestRunOpen -v` passes (6/6)
- [x] `go test ./cli/... -v` passes (no regressions)
