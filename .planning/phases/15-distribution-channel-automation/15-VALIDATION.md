---
phase: 15
slug: distribution-channel-automation
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-27
---

# Phase 15 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | GitHub Actions workflow YAML validation + shell script linting |
| **Config file** | `.github/workflows/release.yml` |
| **Quick run command** | `python3 -c "import yaml; yaml.safe_load(open('.github/workflows/release.yml'))" && echo "VALID"` |
| **Full suite command** | `python3 -c "import yaml; yaml.safe_load(open('.github/workflows/release.yml'))" && bash -n packaging/homebrew/update-cask.sh 2>&1 && echo "ALL VALID"` |
| **Estimated runtime** | ~2 seconds |

---

## Sampling Rate

- **After every task commit:** Run quick YAML validation
- **After every plan wave:** Run full suite (YAML + shell syntax)
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** 2 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 15-01-01 | 01 | 1 | DIST-01 | integration | `grep -q "update-homebrew" .github/workflows/release.yml` | ❌ W0 | ⬜ pending |
| 15-01-02 | 01 | 1 | DIST-02 | integration | `grep -q "winget-releaser\|winget" .github/workflows/release.yml` | ❌ W0 | ⬜ pending |
| 15-01-03 | 01 | 1 | DIST-03 | integration | `grep -q "packaging/winget" .github/workflows/release.yml` | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- Existing infrastructure covers all phase requirements (workflow YAML validation via python3 yaml module, shell syntax via bash -n)

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Homebrew cask actually installs | DIST-01 | Requires macOS with Homebrew + published release | Push tag, wait for workflow, run `brew upgrade storcat` |
| WinGet PR submitted | DIST-02 | Requires published release + WinGet moderation | Push tag, wait for workflow, check microsoft/winget-pkgs PRs |
| Homebrew SHA256 matches DMG | DIST-01 | Requires actual built artifact | Verify SHA256 in cask matches `shasum -a 256` of downloaded DMG |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 2s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
