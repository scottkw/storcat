---
phase: 15
slug: distribution-channel-automation
status: active
nyquist_compliant: true
wave_0_complete: true
created: 2026-03-27
---

# Phase 15 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | GitHub Actions workflow YAML validation + shell script linting |
| **Config file** | `.github/workflows/distribute.yml` |
| **Quick run command** | `python3 -c "import yaml; yaml.safe_load(open('.github/workflows/distribute.yml'))" && echo "VALID"` |
| **Full suite command** | `python3 -c "import yaml; yaml.safe_load(open('.github/workflows/distribute.yml'))" && echo "ALL VALID"` |
| **Estimated runtime** | ~2 seconds |

---

## Sampling Rate

- **After every task commit:** Run quick YAML validation
- **After every plan wave:** Run full suite (YAML validation)
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** 2 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 15-01-T1 | 01 | 1 | DIST-01, DIST-02 | integration | `grep -q "darwin-universal.dmg" packaging/homebrew/storcat.rb.template && grep -q "nullsoft" packaging/winget/manifests/s/scottkw/StorCat/2.1.0/scottkw.StorCat.installer.yaml` | yes | ⬜ pending |
| 15-01-T2 | 01 | 1 | DIST-01, DIST-02, DIST-03 | integration | `grep -q "update-homebrew" .github/workflows/distribute.yml && grep -q "winget-releaser" .github/workflows/distribute.yml && grep -q "update-winget-manifests" .github/workflows/distribute.yml` | ⬜ W0 | ⬜ pending |
| 15-02-T1 | 02 | 2 | DIST-01, DIST-02 | manual | `gh secret list --repo scottkw/storcat 2>/dev/null \| grep -q "HOMEBREW_TAP_TOKEN"` | n/a | ⬜ pending |
| 15-02-T2 | 02 | 2 | DIST-02 | manual | `gh api 'repos/microsoft/winget-pkgs/contents/manifests/s/scottkw/StorCat' 2>/dev/null && echo 'PASS - package exists' \|\| echo 'INFO - not yet submitted (non-blocking)'` | n/a | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- Existing infrastructure covers all phase requirements (workflow YAML validation via python3 yaml module)
- distribute.yml is created by Plan 01 Task 2 — no separate Wave 0 scaffold needed

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Homebrew cask actually installs | DIST-01 | Requires macOS with Homebrew + published release | Push tag, wait for workflow, run `brew upgrade storcat` |
| WinGet PR submitted | DIST-02 | Requires published release + WinGet moderation | Push tag, wait for workflow, check microsoft/winget-pkgs PRs |
| Homebrew SHA256 matches DMG | DIST-01 | Requires actual built artifact | Verify SHA256 in cask matches `shasum -a 256` of downloaded DMG |

---

## Validation Sign-Off

- [x] All tasks have `<automated>` verify or Wave 0 dependencies
- [x] Sampling continuity: no 3 consecutive tasks without automated verify
- [x] Wave 0 covers all MISSING references
- [x] No watch-mode flags
- [x] Feedback latency < 2s
- [x] `nyquist_compliant: true` set in frontmatter

**Approval:** approved
