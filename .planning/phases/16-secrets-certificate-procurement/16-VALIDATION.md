---
phase: 16
slug: secrets-certificate-procurement
status: validated
nyquist_compliant: false
wave_0_complete: true
created: 2026-03-27
last_audited: 2026-03-28T2
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
| 16-01-01 | 01 | 1 | CRED-05 | cli | `gh api repos/scottkw/storcat/environments/release --jq '.name'` | N/A (API) | ✅ green |
| 16-01-02 | 01 | 1 | CRED-06 | file | `test -f docs/runbooks/credential-rotation.md` | ✅ | ✅ green |
| 16-02-01 | 02 | 1 | CRED-01 | manual | `security find-identity -v -p codesigning \| grep "Developer ID Application"` | N/A (local) | ⚠️ manual-only |
| 16-02-02 | 02 | 1 | CRED-02 | cli | `gh secret list -e release --repo scottkw/storcat \| grep APPLE_CERTIFICATE` | N/A (API) | ✅ green |
| 16-03-01 | 03 | 1 | CRED-03 | manual | N/A — vendor decision documented in 16-03-SUMMARY.md | N/A | ⚠️ manual-only |
| 16-03-02 | 03 | 1 | CRED-04 | cli | `gh secret list -e release --repo scottkw/storcat \| grep -E "ES_USERNAME\|ES_PASSWORD\|CREDENTIAL_ID\|ES_TOTP_SECRET"` | N/A (API) | ❌ deferred |

*Status: ⬜ pending · ✅ green · ❌ red/deferred · ⚠️ manual-only*

---

## Wave 0 Requirements

- [x] GitHub `release` environment must be created before secrets can be set
- [x] Apple Developer ID cert confirmed valid before export
- [x] Windows signing vendor selected (deferred by user decision)

*This phase is primarily operational — Wave 0 is the entire phase.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Apple Developer ID cert present and valid on local keychain | CRED-01 | Certificate lives on developer's local macOS keychain | `security find-identity -v -p codesigning` must show `Developer ID Application: Ken Scott (S2K7P43927)` with expiry after 2027-02-01 |
| Windows OV cert vendor selected with RSA not ECDSA | CRED-03 | Vendor enrollment is external; deferral documented | Check 16-03-SUMMARY.md — SSL.com eSigner OV (RSA) identified as recommended path |
| Windows secrets stored in release environment | CRED-04 | Deferred by user decision (option-c) | When ready: purchase SSL.com OV cert, enroll in eSigner, store 4 secrets per `docs/runbooks/credential-rotation.md` |

---

## Validation Audit 2026-03-28

| Metric | Count |
|--------|-------|
| Requirements audited | 6 |
| Gaps found | 3 (CRED-01 manual, CRED-03 manual, CRED-04 deferred) |
| Resolved (automated) | 3 (CRED-02, CRED-05, CRED-06) |
| Escalated to manual-only | 2 (CRED-01, CRED-03) |
| Deferred | 1 (CRED-04) |

**Notes:**
- Path corrected for CRED-06: `docs/credential-rotation-runbook.md` → `docs/runbooks/credential-rotation.md`
- Task IDs realigned to plan structure (16-01 → environment/runbook, 16-02 → Apple secrets, 16-03 → Windows decision)
- CRED-04 marked deferred (not failed) — Windows signing deliberately postponed

---

## Validation Sign-Off

- [x] All tasks have automated verify or manual-only justification
- [x] Sampling continuity: CLI commands available for all automatable requirements
- [x] Wave 0 covers all dependencies
- [ ] No watch-mode flags (N/A — ops phase)
- [x] Feedback latency < 5s
- [ ] `nyquist_compliant: true` set in frontmatter (blocked by 2 manual-only + 1 deferred)

**Approval:** partial — 3/6 requirements have automated verification; 2 are manual-only (local keychain, vendor decision); 1 is deferred (Windows secrets)

---

## Validation Audit 2026-03-28 (re-audit)

| Metric | Count |
|--------|-------|
| Requirements audited | 6 |
| Gaps found | 0 new |
| Previously resolved (automated) | 3 (CRED-02, CRED-05, CRED-06) |
| Manual-only (unchanged) | 2 (CRED-01, CRED-03) |
| Deferred (unchanged) | 1 (CRED-04) |

**Live verification results:**
- `gh api repos/scottkw/storcat/environments/release --jq '.name'` → `release` ✅
- `test -f docs/runbooks/credential-rotation.md` → exists ✅
- `gh secret list --env release` → 6 secrets present (5 original Apple + APPLE_ID added by Phase 17) ✅
- No automatable gaps remain — all non-automated items have valid justification (local keychain, external vendor decision, user deferral)

**Notes:**
- Extra secret `APPLE_ID` present (added 2026-03-28 by Phase 17 for notarization workflow) — not part of Phase 16 scope but does not conflict
- No state drift detected from prior audit
- Phase remains partial due to inherent manual/deferred items — this is correct and expected for an ops/config phase
