---
phase: 20
slug: windows-cli-path-via-nsis
status: audited
nyquist_compliant: true
wave_0_complete: true
created: 2026-03-28
audited: 2026-03-28
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
| 20-01-01 | 01 | 1 | PKG-02 | structural | `grep -q 'EnVar::AddValue' build/windows/installer/project.nsi` | ✅ | ✅ green |
| 20-01-02 | 01 | 1 | PKG-02 | structural | `test -f build/windows/installer/Plugins/x86-unicode/EnVar.dll` | ✅ | ✅ green |
| 20-01-03 | 01 | 1 | PKG-02 | structural | `grep -q 'EnVar::DeleteValue' build/windows/installer/project.nsi` | ✅ | ✅ green |
| 20-01-04 | 01 | 1 | PKG-04 | manual | Human UAT: `winget install scottkw.StorCat` then `storcat version` | ✅ | ⬜ manual-only |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [x] No new Go test files needed — this phase modifies NSIS scripts and bundles a plugin DLL only
- [x] Structural verification via grep in CI (check `project.nsi` contains EnVar calls)

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

- [x] All tasks have `<automated>` verify or Wave 0 dependencies
- [x] Sampling continuity: no 3 consecutive tasks without automated verify
- [x] Wave 0 covers all MISSING references
- [x] No watch-mode flags
- [x] Feedback latency < 15s
- [x] `nyquist_compliant: true` set in frontmatter

**Approval:** approved 2026-03-28

---

## Validation Audit 2026-03-28

| Metric | Count |
|--------|-------|
| Gaps found | 0 |
| Resolved | 0 |
| Escalated | 0 |

All 3 automated structural checks pass. 1 manual-only verification (Windows E2E UAT) correctly categorized — cannot be automated from macOS dev environment.
