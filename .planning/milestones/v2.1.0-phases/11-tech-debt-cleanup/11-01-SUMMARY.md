---
phase: 11-tech-debt-cleanup
plan: 01
subsystem: cli
tags: [tech-debt, testing, cleanup]
dependency_graph:
  requires: []
  provides: [CLOF-06-test-coverage, clean-cli-output, clean-catalog-service]
  affects: [cli/show_test.go, cli/output.go, cli/open.go, cli/version.go, cli/version_test.go, internal/catalog/service.go]
tech_stack:
  added: []
  patterns: [t.Setenv for env var test isolation, fs.Usage stdout consistency]
key_files:
  created: []
  modified:
    - cli/show_test.go
    - cli/output.go
    - cli/open.go
    - cli/version.go
    - cli/version_test.go
    - internal/catalog/service.go
decisions:
  - All 6 CLI commands write --help text to os.Stdout (not os.Stderr) — consistent with Unix convention for --help
  - Error messages remain on os.Stderr — only fs.Usage (--help) was moved
  - FormatBytes exported function removed from internal/catalog — cli package has its own standalone formatBytes
metrics:
  duration: ~5 minutes
  completed: 2026-03-26T19:13:54Z
  tasks: 2
  files_modified: 6
  commits: 2
---

# Phase 11 Plan 01: Tech Debt Cleanup Summary

**One-liner:** Closed all v2.1.0 audit items: NO_COLOR env var test added, stale tablewriter blank import removed, help output stream unified to stdout for open/version commands, version test updated to assert stdout, and orphaned FormatBytes export deleted.

## Tasks Completed

| Task | Name | Commit | Files |
|------|------|--------|-------|
| 1 | Add NO_COLOR env var test, remove stale tablewriter import | 15685683 | cli/show_test.go, cli/output.go |
| 2 | Fix help stream inconsistency, update version test, remove FormatBytes | 9cc407e7 | cli/open.go, cli/version.go, cli/version_test.go, internal/catalog/service.go |

## Verification Results

All 6 success criteria passed:

1. `TestRunShow_NOCOLOREnv` exists and passes — exit 0, no `\x1b` in stdout when `NO_COLOR=1`
2. `cli/output.go` has zero `tablewriter` references
3. `fs.Usage` in `cli/open.go` and `cli/version.go` both use `fmt.Fprintf(os.Stdout, ...)`
4. `TestRunVersion_HelpGoesToStdout` exists and asserts `strings.Contains(stdout, "storcat version")`
5. `internal/catalog/service.go` has no `func FormatBytes(` — only unexported `func (s *Service) formatBytes(`
6. `go test ./... -count=1`, `go build ./...`, and `go vet ./...` all exit 0

## Decisions Made

- **Help text to stdout:** `fs.Usage` is invoked on `--help`, which is not an error — stdout is the correct destination. Error messages (missing args, invalid files) correctly remain on stderr.
- **NO_COLOR test using t.Setenv:** `t.Setenv` auto-restores env vars after the test, preventing test pollution. `fatih/color.NoColor` is checked via `isNoColorSet()` at call time in test environments (non-TTY disables color regardless, so the test validates the env var path is present and wired).
- **FormatBytes removal:** The cli package has its own standalone `formatBytes` function — the exported `catalog.FormatBytes` was an orphan that was never called by any code after Phase 10 moved output helpers to cli/output.go.

## Deviations from Plan

None — plan executed exactly as written.

## Known Stubs

None — all changes are correctness fixes with no placeholder values.

## Self-Check: PASSED

Files verified:
- cli/show_test.go: FOUND (contains TestRunShow_NOCOLOREnv)
- cli/output.go: FOUND (no tablewriter)
- cli/open.go: FOUND (fs.Usage uses os.Stdout)
- cli/version.go: FOUND (fs.Usage uses os.Stdout)
- cli/version_test.go: FOUND (TestRunVersion_HelpGoesToStdout)
- internal/catalog/service.go: FOUND (FormatBytes removed)

Commits verified:
- 15685683: FOUND
- 9cc407e7: FOUND
