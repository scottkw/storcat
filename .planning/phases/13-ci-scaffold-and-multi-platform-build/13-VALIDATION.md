---
phase: 13
slug: ci-scaffold-and-multi-platform-build
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-27
---

# Phase 13 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | No automated test framework for CI workflow YAML |
| **Config file** | None — workflow YAML is the artifact |
| **Quick run command** | `grep -cE "uses:.*@[0-9a-f]{40}" .github/workflows/release.yml` |
| **Full suite command** | Push a test tag: `git tag v0.0.0-test && git push origin v0.0.0-test` |
| **Estimated runtime** | ~2 seconds (grep); ~5 minutes (full workflow run) |

---

## Sampling Rate

- **After every task commit:** Run `grep -cE "uses:.*@[0-9a-f]{40}" .github/workflows/release.yml`
- **After every plan wave:** Validate YAML structure with `yq` or manual review
- **Before `/gsd:verify-work`:** Full suite must be green (test tag push)
- **Max feedback latency:** 2 seconds (local grep); 300 seconds (workflow run)

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 13-01-01 | 01 | 1 | CICD-01 | smoke | `test -f .github/workflows/release.yml` | ❌ W0 | ⬜ pending |
| 13-01-02 | 01 | 1 | CICD-01 | grep | `grep "push:" .github/workflows/release.yml && grep "v\\*\\.\\*\\.\\*" .github/workflows/release.yml` | ❌ W0 | ⬜ pending |
| 13-01-03 | 01 | 1 | CICD-06 | grep | `grep -cE "uses:.*@[0-9a-f]{40}" .github/workflows/release.yml` | ❌ W0 | ⬜ pending |
| 13-01-04 | 01 | 1 | CICD-02 | grep | `grep "darwin/universal" .github/workflows/release.yml` | ❌ W0 | ⬜ pending |
| 13-01-05 | 01 | 1 | CICD-03 | grep | `grep "windowsconsole" .github/workflows/release.yml` | ❌ W0 | ⬜ pending |
| 13-01-06 | 01 | 1 | CICD-04 | grep | `grep "ubuntu-22.04" .github/workflows/release.yml` | ❌ W0 | ⬜ pending |
| 13-01-07 | 01 | 1 | CICD-05 | grep | `grep "needs:" .github/workflows/release.yml` | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `.github/workflows/release.yml` — covers CICD-01 through CICD-06
- [ ] Update `.github/workflows/build.yml` — fix Windows runner, fix ubuntu-latest → ubuntu-22.04, SHA-pin all actions

*Note: No test files to create — validation is CI workflow execution and grep verification, not unit tests.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| macOS universal binary confirmed | CICD-02 | Requires CI runner to produce binary and `lipo -info` output | Push test tag, inspect `build-macos` job logs for `lipo -info` step |
| Windows -windowsconsole flag in build | CICD-03 | Requires CI runner to execute wails build | Push test tag, inspect `build-windows` job logs for `-windowsconsole` flag |
| Linux arm64 builds without WebKit errors | CICD-04 | Requires ubuntu-22.04-arm runner | Push test tag, inspect `build-linux-arm64` job for successful completion |
| Fan-in release job runs after all builds | CICD-05 | Requires all 4 build jobs to complete | Push test tag, verify `release` job starts only after all build jobs succeed |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 300s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
