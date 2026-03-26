---
phase: 9
slug: core-subcommands-create-list-search
status: approved
nyquist_compliant: true
wave_0_complete: true
created: 2026-03-26
---

# Phase 9 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go `testing` stdlib, v1.23 |
| **Config file** | none — `go test ./...` |
| **Quick run command** | `go test ./cli/... -run TestRun -v` |
| **Full suite command** | `go test ./...` |
| **Estimated runtime** | ~5 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./cli/... -v`
- **After every plan wave:** Run `go test ./...`
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** 5 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 09-01-01 | 01 | 1 | CLCM-01 | integration | `go test ./cli/... -run TestRunCreate_Success -v` | ✅ cli/create_test.go | ✅ green |
| 09-01-02 | 01 | 1 | CLCM-01 | integration | `go test ./cli/... -run TestRunCreate_WithFlags -v` | ✅ cli/create_test.go | ✅ green |
| 09-01-03 | 01 | 1 | CLCM-01 | unit | `go test ./cli/... -run TestRunCreate_MissingArg -v` | ✅ cli/create_test.go | ✅ green |
| 09-02-01 | 02 | 1 | CLCM-02 | integration | `go test ./cli/... -run TestRunSearch_WithResults -v` | ✅ cli/search_test.go | ✅ green |
| 09-02-02 | 02 | 1 | CLCM-02 | unit | `go test ./cli/... -run TestRunSearch_MissingArgs -v` | ✅ cli/search_test.go | ✅ green |
| 09-03-01 | 03 | 1 | CLCM-03 | integration | `go test ./cli/... -run TestRunList_WithCatalogs -v` | ✅ cli/list_test.go | ✅ green |
| 09-03-02 | 03 | 1 | CLCM-03 | unit | `go test ./cli/... -run TestRunList_MissingDir_DefaultsCwd -v` | ✅ cli/list_test.go | ✅ green |
| 09-04-01 | 01 | 1 | CLOF-01 | integration | `go test ./cli/... -run TestRunList_JSON -v` | ✅ cli/list_test.go | ✅ green |
| 09-04-02 | 01 | 1 | CLOF-01 | integration | `go test ./cli/... -run TestRunSearch_JSON -v` | ✅ cli/search_test.go | ✅ green |
| 09-05-01 | 01 | 1 | CLOF-03 | integration | `go test ./cli/... -run TestRunCreate_JSON -v` | ✅ cli/create_test.go | ✅ green |
| 09-06-01 | all | 1 | All | unit | `go test ./cli/... -run "TestRun.*NonExistent\|TestRun.*MissingArg" -v` | ✅ multiple | ✅ green |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [x] `cli/create_test.go` — 7 tests for CLCM-01, CLOF-03
- [x] `cli/list_test.go` — 7 tests for CLCM-03, CLOF-01 (list)
- [x] `cli/search_test.go` — 6 tests for CLCM-02, CLOF-01 (search)
- [x] `go get github.com/olekukonko/tablewriter@v1.1.4 && go mod tidy` — installed in Plan 01

*Existing `captureOutput` helper in `cli_test.go` is reusable across new test files — no new test infrastructure needed*

---

## Manual-Only Verifications

*All phase behaviors have automated verification.*

---

## Validation Sign-Off

- [x] All tasks have `<automated>` verify or Wave 0 dependencies
- [x] Sampling continuity: no 3 consecutive tasks without automated verify
- [x] Wave 0 covers all MISSING references
- [x] No watch-mode flags
- [x] Feedback latency < 5s
- [x] `nyquist_compliant: true` set in frontmatter

**Approval:** approved 2026-03-26

---

## Validation Audit 2026-03-26

| Metric | Count |
|--------|-------|
| Gaps found | 0 |
| Resolved | 0 |
| Escalated | 0 |

All 11 validation entries mapped to existing tests. 20 phase-specific tests across 3 test files (create_test.go, list_test.go, search_test.go) all passing. Full CLI suite: 65 tests green.
