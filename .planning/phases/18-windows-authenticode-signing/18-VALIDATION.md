---
phase: 18
slug: windows-authenticode-signing
status: audited
nyquist_compliant: true
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
| 18-01-01 | 01 | 1 | WSIGN-01 | integration | `grep -q 'esigner-codesign' .github/workflows/release.yml` | ✅ | ✅ green |
| 18-01-02 | 01 | 1 | WSIGN-02 | integration | `grep -q 'signtool verify' .github/workflows/release.yml` | ✅ | ✅ green |
| 18-01-03 | 01 | 1 | WSIGN-03 | integration | `grep -q 'environment: release' .github/workflows/release.yml` (in build-windows) | ✅ | ✅ green |
| 18-01-04 | 01 | 1 | WSIGN-04 | integration | Sign steps (L146,158) precede rename (L176) and upload (L184) | ✅ | ✅ green |

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

- [x] All tasks have `<automated>` verify or Wave 0 dependencies
- [x] Sampling continuity: no 3 consecutive tasks without automated verify
- [x] Wave 0 covers all MISSING references
- [x] No watch-mode flags
- [x] Feedback latency < 300s
- [x] `nyquist_compliant: true` set in frontmatter

**Approval:** approved

---

## Validation Audit 2026-03-28

| Metric | Count |
|--------|-------|
| Gaps found | 0 |
| Resolved | 0 |
| Escalated | 0 |

All 4 requirements (WSIGN-01 through WSIGN-04) have automated verification commands that pass against the current codebase. Wave 0 prerequisite (SSL.com secrets) remains a manual-only human-action checkpoint — this does not affect Nyquist compliance as the code changes are complete and verifiable.
