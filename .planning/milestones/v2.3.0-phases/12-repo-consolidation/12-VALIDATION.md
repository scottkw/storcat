---
phase: 12
slug: repo-consolidation
status: complete
nyquist_compliant: true
wave_0_complete: true
created: 2026-03-27
validated: 2026-03-27
---

# Phase 12 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | None (file migration + git operations — shell smoke tests only) |
| **Config file** | none |
| **Quick run command** | `ls packaging/winget/manifests/s/scottkw/StorCat/ && ls packaging/homebrew/` |
| **Full suite command** | `test -f packaging/winget/manifests/s/scottkw/StorCat/2.1.0/scottkw.StorCat.installer.yaml && test -f packaging/homebrew/storcat.rb.template && test -f packaging/homebrew/update-tap.sh` |
| **Estimated runtime** | ~2 seconds |

---

## Sampling Rate

- **After every task commit:** Run `ls packaging/winget/manifests/s/scottkw/StorCat/ && ls packaging/homebrew/`
- **After every plan wave:** Run full suite command
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** 2 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 12-01-01 | 01 | 1 | REPO-01 | smoke | `ls packaging/winget/manifests/s/scottkw/StorCat/` | ✅ | ✅ green |
| 12-01-02 | 01 | 1 | REPO-02 | smoke | `test -f packaging/homebrew/storcat.rb.template && test -f packaging/homebrew/update-tap.sh` | ✅ | ✅ green |
| 12-02-01 | 02 | 2 | REPO-03 | smoke | `gh repo view scottkw/winget-storcat --json isArchived` | ✅ | ✅ green |
| 12-02-02 | 02 | 2 | REPO-04 | smoke | `gh api repos/scottkw/homebrew-storcat/contents/README.md` | ✅ | ✅ green |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [x] `packaging/winget/manifests/s/scottkw/StorCat/` directory — covers REPO-01
- [x] `packaging/homebrew/storcat.rb.template` — covers REPO-02
- [x] `packaging/homebrew/update-tap.sh` — covers REPO-02

*No test framework install needed — all validation is shell smoke tests*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| `winget-storcat` repo is archived | REPO-03 | Irreversible GitHub operation | Run `gh repo view scottkw/winget-storcat --json isArchived` and confirm `"isArchived": true` |
| `homebrew-storcat` README updated | REPO-04 | Cross-repo change | Run `gh api repos/scottkw/homebrew-storcat/contents/README.md` and verify "auto-managed" or "main repo" text present |

---

## Validation Sign-Off

- [x] All tasks have `<automated>` verify or Wave 0 dependencies
- [x] Sampling continuity: no 3 consecutive tasks without automated verify
- [x] Wave 0 covers all MISSING references
- [x] No watch-mode flags
- [x] Feedback latency < 2s
- [x] `nyquist_compliant: true` set in frontmatter

**Approval:** approved

## Validation Audit 2026-03-27

| Metric | Count |
|--------|-------|
| Gaps found | 0 |
| Resolved | 0 |
| Escalated | 0 |

All 4 smoke tests pass. No test framework needed — phase is file migration and git operations only. All requirements verified via shell commands and `gh` CLI.
