---
phase: 3
slug: config-manager
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-25
---

# Phase 3 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test |
| **Config file** | internal/config/config_test.go |
| **Quick run command** | `go test ./internal/config/...` |
| **Full suite command** | `go test ./...` |
| **Estimated runtime** | ~5 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./internal/config/...`
- **After every plan wave:** Run `go test ./...`
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** 5 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 03-01-01 | 01 | 1 | WIN-01 | unit | `go test ./internal/config/... -run TestWindowPosition` | ❌ W0 | ⬜ pending |
| 03-01-02 | 01 | 1 | WIN-02 | unit | `go test ./internal/config/... -run TestWindowPersistence` | ❌ W0 | ⬜ pending |
| 03-01-03 | 01 | 1 | WIN-03 | unit | `go test ./internal/config/... -run TestWindowPersistenceToggle` | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `internal/config/config_test.go` — test stubs for WIN-01, WIN-02, WIN-03 (position fields, persistence toggle round-trip)
- [ ] Verify existing test infrastructure with `go test ./internal/config/...`

*If existing config_test.go already has test infrastructure, Wave 0 only adds new test stubs.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Window position restores on app restart | WIN-01 | Requires full Wails app lifecycle | Launch app, move window, restart, verify position |
| Settings toggle persists across sessions | WIN-03 | Requires UI interaction | Toggle persistence in settings, restart, verify toggle state |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 5s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
