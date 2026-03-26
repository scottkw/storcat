---
phase: 8
slug: cli-foundation-and-platform-compatibility
status: green
nyquist_compliant: true
wave_0_complete: true
created: 2026-03-26
updated: 2026-03-26
---

# Phase 8 -- Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test (stdlib) |
| **Config file** | none -- existing go test infrastructure |
| **Quick run command** | `go test ./cli/...` |
| **Full suite command** | `go test ./...` |
| **Estimated runtime** | ~5 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./cli/...`
- **After every plan wave:** Run `go test ./...`
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** 5 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 8-01-01 | 01 | 1 | CLIP-01 | unit | `go test ./cli/ -run TestRun -v -count=1` | cli/cli_test.go | green |
| 8-01-02 | 01 | 1 | CLIP-02 | unit | `go test ./cli/ -run TestRunHelp -v -count=1` | cli/cli_test.go | green |
| 8-01-03 | 01 | 1 | CLIP-03 | unit | `go test ./cli/ -run TestRun_UnknownCommand -v -count=1` | cli/cli_test.go | green |
| 8-01-04 | 01 | 1 | CLIP-04 | unit | `go test ./cli/ -run TestRun_StdoutStderrSeparation -v -count=1` | cli/cli_test.go | green |
| 8-01-05 | 01 | 1 | CLIP-05 | unit | `go test ./cli/ -run TestRun_StdoutStderrSeparation -v -count=1` | cli/cli_test.go | green |
| 8-01-06 | 01 | 1 | CLIP-06 | unit | `go test ./cli/ -run TestRunVersion -v -count=1` | cli/version_test.go | green |
| 8-02-01 | 02 | 2 | CLPC-02 | unit | `go test -run TestFilterMacOSArgs -v -count=1` | main_test.go | green |
| 8-02-02 | 02 | 2 | CLPC-01 | manual | N/A (Windows console / macOS Finder launch) | N/A | pending |
| 8-02-03 | 02 | 2 | CLPC-03 | unit | `go test -run TestDispatch_UnrecognizedArgsFallThrough -v -count=1` | main_test.go | green |
| 8-02-04 | 02 | 2 | CLPC-04 | integration | `go test -run TestInstallCLIScript_Content -v -count=1` | main_test.go | green |
| 8-03-01 | 03 | 1 | CLCM-06 | unit | `go test ./cli/ -run TestRunVersion -v -count=1` | cli/version_test.go | green |

*Status: pending / green / red / flaky*

---

## Wave 0 Requirements

- [x] `cli/cli_test.go` -- dispatch tests, help text tests, exit code tests
- [x] `main_test.go` -- macOS psn filtering, Wails dev fall-through safety, install script content

*All wave 0 tests implemented and passing. Full suite: `go test ./... -count=1`*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Windows console output visible | CLPC-01 | Requires Windows hardware with GUI subsystem | Build with `-windowsconsole`, run `storcat version` in cmd.exe, verify output appears |
| macOS Finder launch | CLPC-01 | Requires macOS Finder double-click | Double-click StorCat.app, verify GUI launches without crash |

---

## Validation Sign-Off

- [x] All tasks have `<automated>` verify or Wave 0 dependencies
- [x] Sampling continuity: no 3 consecutive tasks without automated verify
- [x] Wave 0 covers all MISSING references
- [x] No watch-mode flags
- [x] Feedback latency < 5s
- [x] `nyquist_compliant: true` set in frontmatter

**Approval:** green — all automated tests passing, 2 gaps resolved

---

## Validation Audit 2026-03-26

| Metric | Count |
|--------|-------|
| Gaps found | 2 |
| Resolved | 2 |
| Escalated | 0 |

**Details:**
- Gap 1 (CLPC-03): Added `TestDispatch_UnrecognizedArgsFallThrough` — verifies wails dev flags fall through to GUI
- Gap 2 (CLPC-04): Added `TestInstallCLIScript_Content` — 8 subtests validating install script correctness
