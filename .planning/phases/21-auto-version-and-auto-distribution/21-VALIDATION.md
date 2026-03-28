---
phase: 21
slug: auto-version-and-auto-distribution
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-28
---

# Phase 21 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | N/A — CI/CD workflow changes only |
| **Config file** | N/A |
| **Quick run command** | `yamllint .github/workflows/release-please.yml` |
| **Full suite command** | N/A (validation via actual GitHub Actions execution) |
| **Estimated runtime** | ~2 seconds (yamllint only) |

---

## Sampling Rate

- **After every task commit:** Verify YAML syntax with `yamllint` or manual review
- **After every plan wave:** Review workflow diffs for correctness
- **Before `/gsd:verify-work`:** Push test commit to main, verify release-please PR is created
- **Max feedback latency:** 2 seconds (local validation only)

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 21-01-01 | 01 | 1 | D-01, D-02 | file check | `test -f release-please-config.json && test -f .release-please-manifest.json` | N/A | ⬜ pending |
| 21-01-02 | 01 | 1 | D-03, D-04 | file check | `test -f .github/workflows/release-please.yml` | N/A | ⬜ pending |
| 21-01-03 | 01 | 1 | D-07 | grep | `grep -q 'jsonpath.*productVersion' release-please-config.json` | N/A | ⬜ pending |
| 21-02-01 | 02 | 2 | D-09, D-10 | grep | `grep -q 'draft: false' .github/workflows/release.yml` | N/A | ⬜ pending |
| 21-02-02 | 02 | 2 | D-05, D-11 | grep | `grep -q 'generate_release_notes' .github/workflows/release.yml; test $? -eq 1` | N/A | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] Verify recent commits use conventional commit format (`git log --oneline -20`)
- [ ] Check repo workflow permissions allow PR creation (Settings > Actions > General)

*Existing infrastructure covers CI/CD execution; no test framework install needed.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| release-please creates PR on push to main | D-04 | Requires actual GitHub Actions run | Push a `fix: test` commit to main, verify PR appears |
| Merging release PR creates tag + release | D-05 | Requires actual GitHub Actions run | Merge the release-please PR, check Releases page |
| release.yml uploads artifacts to existing release | D-09 | Requires full CI pipeline run | Verify release artifacts appear on the GitHub Release |
| distribute.yml fires on publish | D-11 | Requires full CI pipeline chain | Verify Homebrew tap and WinGet manifests are updated |
| Splash screen shows new version | D-06 | Requires build + visual check | Build locally after wails.json bump, check splash |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 2s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
