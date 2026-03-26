---
phase: 3
slug: config-manager
status: complete
nyquist_compliant: true
wave_0_complete: true
created: 2026-03-25
updated: 2026-03-26
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
| 03-01-01 | 01 | 1 | WIN-01 | unit | `go test ./internal/config/... -run TestWindowPosition` | ✅ | ✅ green |
| 03-01-02 | 01 | 1 | WIN-02 | unit | `go test ./internal/config/... -run TestSetWindowPosition` | ✅ | ✅ green |
| 03-01-03 | 01 | 1 | WIN-03 | unit | `go test ./internal/config/... -run TestWindowPersistence` | ✅ | ✅ green |
| 03-02-01 | 02 | 2 | WIN-01, WIN-02 | build | `go build ./...` | ✅ | ✅ green |
| 03-02-02 | 02 | 2 | WIN-03 | build | `grep -q 'GetWindowPersistence' frontend/wailsjs/go/main/App.js` | ✅ | ✅ green |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

Existing infrastructure covers all phase requirements. `internal/config/config_test.go` created as part of Plan 01 TDD cycle.

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Window size restores on app restart | WIN-01 | Requires full Wails app lifecycle (runtime.WindowSetSize in domReady) | Launch app, resize window, restart, verify size restored |
| Window position restores on app restart | WIN-02 | Requires full Wails app lifecycle (runtime.WindowSetPosition in domReady) | Launch app, move window, restart, verify position restored |
| Settings toggle persists across sessions | WIN-03 | Requires UI interaction + app restart cycle | Toggle persistence in settings, restart, verify toggle state |
| beforeClose saves state | WIN-01, WIN-02 | Requires Wails runtime (runtime.WindowGetSize/GetPosition) | Close app, verify config.json updated with current window state |
| domReady skips position for 0,0 | WIN-02 | Requires Wails runtime | Fresh install, verify OS default placement (not 0,0 forced) |

---

## Test Coverage Summary

| Test Function | Requirement | What It Verifies |
|---------------|-------------|------------------|
| TestDefaultConfig_WindowFields | WIN-01, WIN-02, WIN-03 | DefaultConfig returns WindowX=0, WindowY=0, WindowPersistenceEnabled=true |
| TestSetWindowPosition | WIN-02 | SetWindowPosition(100,200) updates config fields |
| TestSetWindowPosition_Persists | WIN-02 | SetWindowPosition round-trips through disk |
| TestSetWindowPersistence | WIN-03 | SetWindowPersistence(false) updates config field |
| TestSetWindowPersistence_Persists | WIN-03 | SetWindowPersistence round-trips through disk |
| TestGetWindowPersistence_Default | WIN-03 | Fresh config returns true (not Go zero-value false) |
| TestWindowPosition_RoundTrip | WIN-01, WIN-02 | SetWindowPosition then Get() returns correct values |

**7 automated tests** cover all config-layer data persistence behavior.
**5 manual verifications** cover app-layer runtime behavior requiring Wails lifecycle.

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
| Gaps found | 2 (Plan 02 tasks missing from map) |
| Resolved | 2 (added to per-task map + manual-only) |
| Escalated | 0 |

**Notes:** Original VALIDATION.md only tracked Plan 01 tasks. Added Plan 02 tasks (03-02-01, 03-02-02) to per-task map. Updated all statuses from pending to green based on test run results. Added Test Coverage Summary section mapping all 7 tests to requirements. All config-layer behavior has automated verification; app-layer runtime behavior documented as manual-only (requires Wails runtime).
