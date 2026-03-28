---
phase: 16
slug: secrets-certificate-procurement
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-27
---

# Phase 16 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | gh CLI + security CLI (no test framework — this is an ops/config phase) |
| **Config file** | none — validation via CLI commands |
| **Quick run command** | `security find-identity -v -p codesigning` |
| **Full suite command** | `gh api repos/scottkw/storcat/environments/release && security find-identity -v -p codesigning` |
| **Estimated runtime** | ~5 seconds |

---

## Sampling Rate

- **After every task commit:** Run quick verification command
- **After every plan wave:** Run full suite command
- **Before `/gsd:verify-work`:** Full suite must confirm all secrets populated
- **Max feedback latency:** 5 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 16-01-01 | 01 | 1 | CRED-01 | cli | `security find-identity -v -p codesigning \| grep "Developer ID Application"` | ✅ | ⬜ pending |
| 16-01-02 | 01 | 1 | CRED-02 | cli | `gh secret list -e release --repo scottkw/storcat \| grep APPLE_CERTIFICATE` | ❌ W0 | ⬜ pending |
| 16-01-03 | 01 | 1 | CRED-03 | manual | N/A — vendor enrollment verification | N/A | ⬜ pending |
| 16-01-04 | 01 | 1 | CRED-04 | cli | `gh secret list -e release --repo scottkw/storcat \| grep -E "ES_USERNAME\|ES_PASSWORD\|CREDENTIAL_ID\|ES_TOTP_SECRET"` | ❌ W0 | ⬜ pending |
| 16-01-05 | 01 | 1 | CRED-05 | cli | `gh api repos/scottkw/storcat/environments/release --jq '.protection_rules'` | ❌ W0 | ⬜ pending |
| 16-01-06 | 01 | 1 | CRED-06 | file | `test -f docs/credential-rotation-runbook.md` | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] GitHub `release` environment must be created before secrets can be set
- [ ] Apple Developer ID cert confirmed valid before export
- [ ] Windows signing vendor selected and enrollment started

*This phase is primarily operational — Wave 0 is the entire phase.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Windows OV cert uses RSA not ECDSA | CRED-03 | Vendor enrollment is external | Confirm RSA key type in SSL.com dashboard after enrollment |
| Apple cert export to .p12 | CRED-02 | Keychain Access UI operation | Export from Keychain, verify base64 encoding works |

---

## Validation Sign-Off

- [ ] All tasks have automated verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 5s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
