---
phase: 8
slug: cli-foundation-and-platform-compatibility
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-26
---

# Phase 8 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test (stdlib) |
| **Config file** | none — existing go test infrastructure |
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
| 8-01-01 | 01 | 1 | CLIP-01 | unit | `go test ./cli/ -run TestDispatch` | ❌ W0 | ⬜ pending |
| 8-01-02 | 01 | 1 | CLIP-02 | unit | `go test ./cli/ -run TestHelp` | ❌ W0 | ⬜ pending |
| 8-01-03 | 01 | 1 | CLIP-03 | unit | `go test ./cli/ -run TestExitCode` | ❌ W0 | ⬜ pending |
| 8-01-04 | 01 | 1 | CLIP-04 | unit | `go test ./cli/ -run TestStdout` | ❌ W0 | ⬜ pending |
| 8-01-05 | 01 | 1 | CLIP-05 | unit | `go test ./cli/ -run TestStderr` | ❌ W0 | ⬜ pending |
| 8-01-06 | 01 | 1 | CLIP-06 | unit | `go test ./cli/ -run TestVersion` | ❌ W0 | ⬜ pending |
| 8-02-01 | 02 | 1 | CLPC-01 | unit | `go test ./cli/ -run TestMacOSPsn` | ❌ W0 | ⬜ pending |
| 8-02-02 | 02 | 1 | CLPC-02 | manual | N/A (Windows console) | N/A | ⬜ pending |
| 8-02-03 | 02 | 1 | CLPC-03 | unit | `go test ./cli/ -run TestWailsDev` | ❌ W0 | ⬜ pending |
| 8-02-04 | 02 | 1 | CLPC-04 | integration | `go test ./cli/ -run TestInstall` | ❌ W0 | ⬜ pending |
| 8-03-01 | 03 | 1 | CLCM-06 | unit | `go test ./cli/ -run TestVersionCmd` | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `cli/cli_test.go` — dispatch tests, help text tests, exit code tests
- [ ] `cli/platform_test.go` — macOS psn filtering, Wails dev safety tests

*Existing `go test ./...` infrastructure is in place. Only new test files needed.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Windows console output visible | CLPC-02 | Requires Windows hardware with GUI subsystem | Build with `-windowsconsole`, run `storcat version` in cmd.exe, verify output appears |
| macOS Finder launch | CLPC-01 | Requires macOS Finder double-click | Double-click StorCat.app, verify GUI launches without crash |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 5s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
