---
phase: 20
slug: windows-cli-path-via-nsis
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-28
---

# Phase 20 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go `testing` package + NSIS structural verification (grep) |
| **Config file** | none — standard Go test discovery |
| **Quick run command** | `go test ./... -count=1` |
| **Full suite command** | `go test ./... -count=1 -race` |
| **Estimated runtime** | ~15 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./... -count=1`
- **After every plan wave:** Run `go test ./... -count=1 -race`
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** 15 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 20-01-01 | 01 | 1 | PKG-02 | structural | `grep -q 'EnVar::AddValue' build/windows/installer/project.nsi` | ❌ W0 | ⬜ pending |
| 20-01-02 | 01 | 1 | PKG-02 | structural | `test -f build/windows/installer/Plugins/x86-unicode/EnVar.dll` | ❌ W0 | ⬜ pending |
| 20-01-03 | 01 | 1 | PKG-02 | structural | `grep -q 'EnVar::DeleteValue' build/windows/installer/project.nsi` | ❌ W0 | ⬜ pending |
| 20-01-04 | 01 | 1 | PKG-04 | manual | Human UAT: `winget install scottkw.StorCat` then `storcat version` | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] No new Go test files needed — this phase modifies NSIS scripts and bundles a plugin DLL only
- [ ] Structural verification via grep in CI (check `project.nsi` contains EnVar calls)

*Existing infrastructure covers Go test requirements. NSIS validation is structural + manual UAT.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| PATH visible in System Environment Variables | PKG-02 | Windows GUI verification | Install StorCat → System Properties → Advanced → Environment Variables → System PATH contains install dir |
| `storcat version` works in new terminal | PKG-04 | Windows E2E only | Install via WinGet → open new cmd.exe → run `storcat version` → expect version output |
| PATH removed on uninstall | PKG-02 | Windows E2E only | Uninstall StorCat → verify PATH no longer contains install dir |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 15s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
