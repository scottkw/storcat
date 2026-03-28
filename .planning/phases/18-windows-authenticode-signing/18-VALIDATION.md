---
phase: 18
slug: windows-authenticode-signing
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-28
---

# Phase 18 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | GitHub Actions CI (workflow validation) |
| **Config file** | `.github/workflows/release.yml` |
| **Quick run command** | `act -j build-windows --dryrun` or manual workflow syntax check |
| **Full suite command** | Push tag to trigger release workflow |
| **Estimated runtime** | ~5 minutes (full CI run) |

---

## Sampling Rate

- **After every task commit:** Run `yamllint .github/workflows/release.yml` + syntax validation
- **After every plan wave:** Verify workflow YAML parses correctly
- **Before `/gsd:verify-work`:** Full CI run via tag push must complete successfully
- **Max feedback latency:** 300 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 18-01-01 | 01 | 1 | WSIGN-01 | integration | `grep -q 'esigner-codesign' .github/workflows/release.yml` | ✅ | ⬜ pending |
| 18-01-02 | 01 | 1 | WSIGN-02 | integration | `grep -q 'signtool verify' .github/workflows/release.yml` | ✅ | ⬜ pending |
| 18-01-03 | 01 | 1 | WSIGN-03 | integration | `grep -q 'environment: release' .github/workflows/release.yml` (in build-windows) | ✅ | ⬜ pending |
| 18-01-04 | 01 | 1 | WSIGN-04 | integration | Check distribute.yml SHA256 uses signed artifacts | ✅ | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] SSL.com eSigner credentials procured and stored as GitHub secrets (ES_USERNAME, ES_PASSWORD, CREDENTIAL_ID, ES_TOTP_SECRET)

*User-action prerequisite — cannot be automated.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| SmartScreen acceptance | WSIGN-01 | Requires downloading signed binary on fresh Windows | Download signed .exe from release, run on Windows, verify SmartScreen shows publisher name |
| eSigner cloud signing | WSIGN-01, WSIGN-02 | Requires real SSL.com credentials | Push a release tag and verify CI signing steps succeed |
| signtool verify output | WSIGN-02 | Requires signed binary artifacts | Check CI logs for `signtool verify /v /pa` success output |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 300s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
