---
phase: 9
slug: core-subcommands-create-list-search
status: draft
nyquist_compliant: false
wave_0_complete: false
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
| 09-01-01 | 01 | 1 | CLCM-01 | integration | `go test ./cli/... -run TestRunCreate -v` | ❌ W0 | ⬜ pending |
| 09-01-02 | 01 | 1 | CLCM-01 | integration | `go test ./cli/... -run TestRunCreate_Flags -v` | ❌ W0 | ⬜ pending |
| 09-01-03 | 01 | 1 | CLCM-01 | unit | `go test ./cli/... -run TestRunCreate_MissingArg -v` | ❌ W0 | ⬜ pending |
| 09-02-01 | 02 | 1 | CLCM-02 | integration | `go test ./cli/... -run TestRunSearch -v` | ❌ W0 | ⬜ pending |
| 09-02-02 | 02 | 1 | CLCM-02 | unit | `go test ./cli/... -run TestRunSearch_MissingArgs -v` | ❌ W0 | ⬜ pending |
| 09-03-01 | 03 | 1 | CLCM-03 | integration | `go test ./cli/... -run TestRunList -v` | ❌ W0 | ⬜ pending |
| 09-03-02 | 03 | 1 | CLCM-03 | unit | `go test ./cli/... -run TestRunList_NoArg -v` | ❌ W0 | ⬜ pending |
| 09-04-01 | 01 | 1 | CLOF-01 | integration | `go test ./cli/... -run TestRunList_JSON -v` | ❌ W0 | ⬜ pending |
| 09-04-02 | 01 | 1 | CLOF-01 | integration | `go test ./cli/... -run TestRunSearch_JSON -v` | ❌ W0 | ⬜ pending |
| 09-05-01 | 01 | 1 | CLOF-03 | integration | `go test ./cli/... -run TestRunCreate_JSON -v` | ❌ W0 | ⬜ pending |
| 09-06-01 | all | 1 | All | unit | `go test ./cli/... -run TestRun.*Error -v` | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `cli/create_test.go` — stubs for CLCM-01, CLOF-03
- [ ] `cli/list_test.go` — stubs for CLCM-03, CLOF-01 (list)
- [ ] `cli/search_test.go` — stubs for CLCM-02, CLOF-01 (search)
- [ ] `go get github.com/olekukonko/tablewriter@v1.1.4 && go mod tidy` — must precede any import

*Existing `captureOutput` helper in `cli_test.go` is reusable across new test files — no new test infrastructure needed*

---

## Manual-Only Verifications

*All phase behaviors have automated verification.*

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 5s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
