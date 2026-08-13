---
phase: 14
slug: platform-packaging
status: approved
nyquist_compliant: true
wave_0_complete: true
created: 2026-03-27
---

# Phase 14 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | None — CI/CD configuration phase (no runtime code) |
| **Config file** | N/A |
| **Quick run command** | `python3 -c "import yaml; yaml.safe_load(open('.github/workflows/release.yml'))" && echo VALID` |
| **Full suite command** | See Per-Task Verification Map below |
| **Estimated runtime** | ~1 second |

---

## Sampling Rate

- **After every task commit:** Run YAML validation + pattern grep checks
- **After every plan wave:** Run full verification suite (all 4 pattern checks + YAML + file existence)
- **Before `/gsd:verify-work`:** All grep patterns confirmed present
- **Max feedback latency:** 1 second

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 14-01-01 | 01 | 1 | — | file-check | `test -f build/linux/storcat.desktop && grep -q "Name=StorCat" build/linux/storcat.desktop && grep -q "Terminal=false" build/linux/storcat.desktop` | N/A (config) | ✅ green |
| 14-01-02 | 01 | 1 | PKG-01 | pattern-grep | `grep -q "create-dmg" .github/workflows/release.yml` | N/A (CI config) | ✅ green |
| 14-01-02 | 01 | 1 | PKG-02 | pattern-grep | `grep -q "windows-2022" .github/workflows/release.yml && grep -q "\-nsis" .github/workflows/release.yml && ! grep -q "windows-latest" .github/workflows/release.yml` | N/A (CI config) | ✅ green |
| 14-01-02 | 01 | 1 | PKG-03 | pattern-grep | `grep -q "linuxdeploy" .github/workflows/release.yml` | N/A (CI config) | ✅ green |
| 14-01-02 | 01 | 1 | PKG-04 | pattern-grep | `grep -c "dpkg-deb" .github/workflows/release.yml | grep -q 2` | N/A (CI config) | ✅ green |
| 14-01-02 | 01 | 1 | ALL | syntax | `python3 -c "import yaml; yaml.safe_load(open('.github/workflows/release.yml'))"` | N/A | ✅ green |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

Existing infrastructure covers all phase requirements. This is a CI/CD configuration phase — verification is pattern-matching on YAML files and file existence checks, all executable locally.

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| macOS DMG creation on macos-14 runner | PKG-01 | Requires real macOS CI runner with create-dmg and wails build output | Push a tag; verify `StorCat-*-darwin-universal.dmg` in release assets |
| Windows NSIS installer on windows-2022 | PKG-02 | Requires NSIS 3.10 on Windows runner and wails build with -nsis flag | Push a tag; verify `StorCat-*-windows-amd64-installer.exe` in release assets |
| Linux AppImage execution on x86_64 | PKG-03 | Requires x86_64 Linux with libfuse2 and libwebkit2gtk-4.0-37 | Download AppImage from release; `chmod +x` and execute on Ubuntu 22.04 |
| Linux .deb installation on arm64 | PKG-04 | Requires arm64 Ubuntu system for dpkg installation | Download arm64 .deb; `dpkg -i` on Ubuntu 22.04 ARM; launch StorCat |

---

## Validation Sign-Off

- [x] All tasks have `<automated>` verify or Wave 0 dependencies
- [x] Sampling continuity: no 3 consecutive tasks without automated verify
- [x] Wave 0 covers all MISSING references
- [x] No watch-mode flags
- [x] Feedback latency < 1s
- [x] `nyquist_compliant: true` set in frontmatter

**Approval:** approved 2026-03-27

---

## Validation Audit 2026-03-27

| Metric | Count |
|--------|-------|
| Gaps found | 0 |
| Resolved | 0 |
| Escalated | 0 |

All 4 requirements (PKG-01 through PKG-04) have automated pattern-grep verification that passes locally. 4 manual-only items documented for CI runner validation (inherently cannot be automated locally).
