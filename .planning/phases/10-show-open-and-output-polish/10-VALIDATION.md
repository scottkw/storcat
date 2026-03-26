---
phase: 10
slug: show-open-and-output-polish
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-26
---

# Phase 10 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test |
| **Config file** | none — existing test infrastructure |
| **Quick run command** | `go test ./cli/...` |
| **Full suite command** | `go test ./...` |
| **Estimated runtime** | ~10 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./cli/...`
- **After every plan wave:** Run `go test ./...`
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** 10 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 10-01-01 | 01 | 1 | CLOF-04 | unit | `go test ./cli/ -run TestShowCommand` | ❌ W0 | ⬜ pending |
| 10-01-02 | 01 | 1 | CLOF-05 | unit | `go test ./cli/ -run TestShowDepth` | ❌ W0 | ⬜ pending |
| 10-01-03 | 01 | 1 | CLOF-04 | unit | `go test ./cli/ -run TestShowColor` | ❌ W0 | ⬜ pending |
| 10-01-04 | 01 | 1 | CLOF-02 | unit | `go test ./cli/ -run TestShowJSON` | ❌ W0 | ⬜ pending |
| 10-02-01 | 02 | 1 | CLCM-05 | unit | `go test ./cli/ -run TestOpenCommand` | ❌ W0 | ⬜ pending |
| 10-02-02 | 02 | 1 | CLPC-05 | unit | `go test ./cli/ -run TestNoColor` | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `internal/cli/show_test.go` — stubs for show command tests
- [ ] `internal/cli/open_test.go` — stubs for open command tests

*Existing test infrastructure (captureOutput helper, catalog fixtures) covers shared needs.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| `storcat open` launches browser | CLCM-05 | Requires GUI browser | Run `storcat open <catalog.json>` on macOS/Windows/Linux, verify browser opens |
| Color renders correctly in terminal | CLOF-04 | TTY rendering is visual | Run `storcat show <catalog.json>` in terminal, verify blue/bold directories |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 10s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
