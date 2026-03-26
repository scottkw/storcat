---
phase: 6
slug: platform-integration
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-25
---

# Phase 6 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go `testing` package |
| **Config file** | none (standard `go test`) |
| **Quick run command** | `go test ./... -run TestVersion` |
| **Full suite command** | `go test ./...` |
| **Estimated runtime** | ~5 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./... -run TestVersion`
- **After every plan wave:** Run `go test ./...`
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** 5 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 06-01-01 | 01 | 0 | PLAT-02 | unit | `go test . -run TestGetVersion` | ❌ W0 | ⬜ pending |
| 06-01-02 | 01 | 1 | PLAT-01 | manual | Manual: run `wails dev` and drag header | N/A | ⬜ pending |
| 06-01-03 | 01 | 1 | PLAT-02 | unit | `go test . -run TestGetVersion` | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `version_test.go` — unit test for `GetVersion()` returning the package-level `Version` variable — covers PLAT-02

*Existing infrastructure covers PLAT-01 (manual-only — CSS drag requires native WebKit runtime).*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| macOS header drag region is functional | PLAT-01 | CSS `--wails-draggable: drag` requires WebKit runtime to interpret; cannot be unit tested | 1. Run `wails dev` 2. Click and hold the app header 3. Drag the window 4. Verify window moves |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 5s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
