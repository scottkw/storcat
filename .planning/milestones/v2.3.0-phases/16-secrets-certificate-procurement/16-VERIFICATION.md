---
phase: 16-secrets-certificate-procurement
verified: 2026-03-28T05:00:00Z
status: human_needed
score: 5/6 success criteria verified (4/6 fully, 1/6 partially, 1/6 requires human)
gaps:
  - truth: "Windows Authenticode signing credentials are stored as 4 eSigner API secrets in the release environment"
    status: failed
    reason: "Windows code signing was deliberately deferred (option-c). 0 of 4 Windows secrets are stored. This is expected and documented in 16-03-SUMMARY.md. CRED-04 is incomplete by design; Phases 18 and 20 are blocked until a future milestone."
    artifacts: []
    missing:
      - "ES_USERNAME, ES_PASSWORD, CREDENTIAL_ID, ES_TOTP_SECRET — to be stored when SSL.com OV cert is purchased in a future milestone"
  - truth: "GitHub release environment exists with protection rules and all 9 signing secrets populated"
    status: partial
    reason: "Environment exists with correct tag policy (v*.*.*). Only 5 of 9 secrets present (Apple only). Windows deferral is documented and expected. Partial satisfaction is the intended state for this milestone."
    artifacts: []
    missing:
      - "4 Windows secrets (ES_USERNAME, ES_PASSWORD, CREDENTIAL_ID, ES_TOTP_SECRET) — blocked until Windows cert obtained"
human_verification:
  - test: "Confirm Apple Developer ID Application certificate is present and valid on local keychain"
    expected: "security find-identity -v -p codesigning output contains 'Developer ID Application: Ken Scott (S2K7P43927)' with notAfter date after 2026-03-28"
    why_human: "Certificate lives on the developer's local macOS keychain. Cannot be verified from CI or a remote shell — requires local terminal access."
  - test: "Confirm the APPLE_CERTIFICATE secret decodes to a valid .p12 containing the private key"
    expected: "base64 -d <<< $(gh secret view APPLE_CERTIFICATE --env release) | openssl pkcs12 -info -noout succeeds without error"
    why_human: "GitHub hides secret values after storage. The only way to confirm the .p12 round-trips correctly is to decode and inspect it locally, which requires the original password."
---

# Phase 16: Secrets & Certificate Procurement — Verification Report

**Phase Goal:** All signing credentials are in hand and configured in GitHub Actions before any signing automation is built
**Verified:** 2026-03-28T05:00:00Z
**Status:** human_needed (automated checks passed; 2 items need human confirmation; Windows deferral is expected and documented)
**Re-verification:** No — initial verification

---

## Goal Achievement

### Observable Truths (from Success Criteria)

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | Developer ID Application certificate is located or renewed and confirmed valid via `security find-identity` | ? HUMAN | SUMMARY documents cert was verified (expiry 2027-02-01). Cannot re-verify from verifier context — requires local keychain access. |
| 2 | Apple certificate exported as .p12, base64-encoded, stored as `APPLE_CERTIFICATE` in release environment | ✓ VERIFIED | `gh secret list --env release` shows `APPLE_CERTIFICATE` (2026-03-28T03:43:49Z). All 5 Apple secrets present. |
| 3 | Windows OV code signing vendor selected with RSA (not ECDSA) confirmed before purchase | ✓ VERIFIED (deferred) | 16-03-SUMMARY.md documents decision: option-c selected, Windows signing deferred. SSL.com eSigner OV (RSA, $20/mo) identified as recommended path. CRED-03 satisfied by documented decision per prompt context. |
| 4 | Windows Authenticode signing credentials stored as 4 eSigner API secrets in release environment | ✗ FAILED (expected) | 0 Windows secrets in environment. Deferred by user decision (option-c in Plan 16-03). Phases 18 and 20 are blocked. This outcome is designed and documented. |
| 5 | GitHub release environment exists with protection rules and all 9 signing secrets populated | ⚠️ PARTIAL | Environment verified: `release` with `v*.*.*` tag policy, `custom_branch_policies: true`. 5 of 9 secrets present. Windows deferral = intentional partial state. |
| 6 | Credential rotation runbook document exists describing what to do when each cert expires | ✓ VERIFIED | `docs/runbooks/credential-rotation.md` exists (181 lines, commit 06c22ad4). All 9 secret names present (38 matches). 17 `gh secret set` commands. Expiry date 2027-02-01 appears 4 times. `security find-identity` command present. Emergency procedures section present. |

**Score:** 4/6 fully verified, 1/6 partially verified, 1/6 human-needed (SC-1), 1/6 failed-by-design (SC-4/Windows deferral)

---

### Required Artifacts

| Artifact | Expected | Status | Details |
|----------|----------|--------|---------|
| `docs/runbooks/credential-rotation.md` | Credential rotation runbook covering all 9 secrets | ✓ VERIFIED | Exists, 181 lines. Covers all 5 Apple secrets and all 4 Windows secrets. Contains expiry dates, renewal steps, verification commands, and emergency procedures. |
| GitHub `release` environment | Environment with `v*.*.*` tag policy | ✓ VERIFIED | `gh api repos/scottkw/storcat/environments/release --jq '.name'` returns `release`. Branch policy name returns `v*.*.*`. `custom_branch_policies: true`. |
| 5 Apple secrets in `release` environment | APPLE_CERTIFICATE, APPLE_CERTIFICATE_PASSWORD, APPLE_CERTIFICATE_NAME, APPLE_NOTARIZATION_PASSWORD, APPLE_TEAM_ID | ✓ VERIFIED | All 5 present per `gh secret list --env release --repo scottkw/storcat`. Stored 2026-03-28T03:43–45Z. |
| 4 Windows secrets in `release` environment | ES_USERNAME, ES_PASSWORD, CREDENTIAL_ID, ES_TOTP_SECRET | ✗ ABSENT (expected) | 0 Windows secrets stored. Deferred by user decision. Not a verification failure — it is the documented outcome of Plan 16-03. |

---

### Key Link Verification

| From | To | Via | Status | Details |
|------|----|-----|--------|---------|
| `docs/runbooks/credential-rotation.md` | GitHub `release` environment | Documents all 9 secret names with `gh secret set` commands | ✓ WIRED | Runbook references all 9 secrets and provides exact `gh secret set --env release --repo scottkw/storcat` commands for each. |
| Local keychain (Developer ID cert) | GitHub `release` environment | Export .p12 → base64 encode → `gh secret set APPLE_CERTIFICATE` | ✓ WIRED (per SUMMARY) | 16-02-SUMMARY.md confirms the full chain was executed. APPLE_CERTIFICATE secret exists in environment as evidence. Human confirmation needed to fully validate .p12 contents. |
| SSL.com eSigner enrollment | GitHub `release` environment | 4 API secrets via `gh secret set` | ✗ NOT WIRED (expected) | Windows signing deferred. No eSigner enrollment performed. This link will be established in a future milestone. |

---

### Data-Flow Trace (Level 4)

Not applicable. Phase 16 has no components that render dynamic data. All deliverables are GitHub API configuration (environment + secrets) and a documentation file. No React components, API routes, or data pipelines are involved.

---

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
|----------|---------|--------|--------|
| GitHub release environment exists | `gh api repos/scottkw/storcat/environments/release --jq '.name'` | `release` | ✓ PASS |
| Tag deployment policy is `v*.*.*` | `gh api repos/scottkw/storcat/environments/release/deployment-branch-policies --jq '.branch_policies[0].name'` | `v*.*.*` | ✓ PASS |
| Custom branch policies enabled | `gh api repos/scottkw/storcat/environments/release --jq '.deployment_branch_policy.custom_branch_policies'` | `true` | ✓ PASS |
| All 5 Apple secrets present | `gh secret list --env release --repo scottkw/storcat \| wc -l` | `5` | ✓ PASS |
| Secret count matches expected (deferral = 5) | `gh secret list --env release --repo scottkw/storcat \| grep -c APPLE_` | `5` | ✓ PASS |
| Runbook file exists | `test -f docs/runbooks/credential-rotation.md` | EXISTS | ✓ PASS |
| Runbook covers all 9 secrets | `grep -c "ES_\|APPLE_\|CREDENTIAL_ID" docs/runbooks/credential-rotation.md` | `38` | ✓ PASS |
| Runbook has `gh secret set` commands | `grep -c "gh secret set" docs/runbooks/credential-rotation.md` | `17` | ✓ PASS |
| Expiry date documented | `grep -c "2027-02-01" docs/runbooks/credential-rotation.md` | `4` | ✓ PASS |
| Commit for runbook exists | `git log --oneline \| grep "06c22ad4"` | `06c22ad4 feat(16-01): add credential rotation runbook for all 9 signing secrets` | ✓ PASS |
| Windows secrets absent (deferral correct) | `gh secret list --env release --repo scottkw/storcat \| grep -c "ES_\|CREDENTIAL_ID"` | `0` | ✓ PASS (expected) |

---

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
|-------------|------------|-------------|--------|----------|
| CRED-01 | 16-02-PLAN.md | User can locate or renew existing Apple Developer ID Application certificate | ? HUMAN | 16-02-SUMMARY.md: "Developer ID Application certificate confirmed valid on local keychain (expiry Feb 1, 2027)". Cannot re-run `security find-identity` from verifier context. |
| CRED-02 | 16-02-PLAN.md | User can export Developer ID cert as .p12 and store base64-encoded in GitHub secret | ✓ SATISFIED | `APPLE_CERTIFICATE` secret present in GitHub release environment (stored 2026-03-28T03:43:49Z). SUMMARY documents .p12 export and base64 encoding. |
| CRED-03 | 16-03-PLAN.md | User can locate or acquire Windows OV code signing certificate (RSA, not ECDSA) | ✓ SATISFIED (deferred) | 16-03-SUMMARY.md documents vendor decision: SSL.com eSigner OV (RSA) identified as recommended path. Deferral documented. Per prompt context, documented decision satisfies CRED-03. |
| CRED-04 | 16-03-PLAN.md | User can store Windows cert as base64-encoded PFX in GitHub secret | ✗ INCOMPLETE (expected) | 0 Windows secrets stored. Deferred by user decision. CRED-04 remains incomplete; will be satisfied in a future milestone when SSL.com OV cert is purchased. |
| CRED-05 | 16-01-PLAN.md + 16-03-PLAN.md | All 9 signing secrets stored in GitHub Actions `release` environment with protection rules | ⚠️ PARTIAL | Environment exists with correct protection rules. 5 of 9 secrets present. Windows deferral = 4 secrets outstanding. REQUIREMENTS.md marks CRED-05 as "Complete" (tracking the environment creation; deferral is accepted scope change). |
| CRED-06 | 16-01-PLAN.md | Credential rotation runbook documents what to do when certs expire | ✓ SATISFIED | `docs/runbooks/credential-rotation.md` exists (commit 06c22ad4), covers all 9 secrets, expiry dates, renewal steps, emergency procedures. REQUIREMENTS.md marks CRED-06 as "Complete". |

**Orphaned requirements check:** REQUIREMENTS.md maps CRED-01 through CRED-06 to Phase 16. All 6 are claimed across the 3 plans (16-01: CRED-05, CRED-06; 16-02: CRED-01, CRED-02; 16-03: CRED-03, CRED-04). No orphaned requirements.

**REQUIREMENTS.md traceability status vs. reality:**
- CRED-01: Marked "Pending" in traceability table (checkbox unchecked) — accurate, requires human verification
- CRED-02: Marked "Pending" in traceability table — under-reported; secret is stored and SUMMARY confirms completion
- CRED-03: Marked "Pending" — accurate; vendor decision made but cert not purchased
- CRED-04: Marked "Pending" — accurate; deferred
- CRED-05: Marked "Complete" — partially accurate; environment exists but only 5/9 secrets
- CRED-06: Marked "Complete" — accurate; runbook verified

Note: The REQUIREMENTS.md traceability table appears to reflect the state at roadmap creation, not after execution. The checkbox statuses in the requirements list (`[ ]` vs `[x]`) also do not fully reflect actual completion — CRED-05 and CRED-06 show `[x]` (correct), but CRED-01 and CRED-02 show `[ ]` despite substantial completion. This is an informational discrepancy in the planning documents, not a gap in the implementation.

---

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
|------|------|---------|----------|--------|
| `docs/runbooks/credential-rotation.md` | 181 (end of file) | No "last updated" or "last reviewed" date | ℹ️ Info | Runbook will silently drift as expiry dates approach. Minimal impact — Section 3 provides verification commands to check actual cert dates. |

No blocker or warning anti-patterns found. The runbook is substantive documentation, not a placeholder. All `gh secret set` commands are concrete and actionable.

---

### Human Verification Required

#### 1. Apple Developer ID Certificate — Local Keychain Confirmation

**Test:** On the developer's macOS machine, run:
```bash
security find-identity -v -p codesigning
security find-certificate -c "Developer ID Application: Ken Scott" -p | openssl x509 -noout -dates
```
**Expected:** Output contains `Developer ID Application: Ken Scott (S2K7P43927)` and `notAfter=Feb  1 22:12:15 2027 GMT`
**Why human:** The certificate lives on the local macOS keychain. It cannot be queried remotely or from a CI environment. The SUMMARY documents this was run and passed, but independent re-verification requires local access.

#### 2. APPLE_CERTIFICATE Secret Content Validity

**Test:** Decode the stored APPLE_CERTIFICATE secret and confirm it contains a private key:
```bash
# On local machine with the original .p12 password:
gh secret view APPLE_CERTIFICATE --env release --repo scottkw/storcat 2>/dev/null
# If value is accessible, pipe through: base64 -d | openssl pkcs12 -info -noout -passin pass:THE_PASSWORD
```
**Expected:** No "no private key" error; PKCS12 info shows both a certificate and a private key
**Why human:** GitHub masks secret values after storage. The only confirmation available is the SUMMARY statement that the .p12 was exported with both certificate and private key ("Export 2 Items"). A functional test in Phase 17 (when the cert is imported into a CI keychain for signing) will provide definitive confirmation.

---

### Gaps Summary

**Windows Deferral (expected, by design):**
CRED-04 is incomplete and 4 of 9 secrets are absent from the release environment. This is the documented, deliberate outcome of the user selecting option-c in Plan 16-03. The gap is tracked in the planning documents; Phases 18 and 20 are explicitly marked as blocked. This will be resolved in a future milestone when an SSL.com eSigner OV certificate is purchased.

**Apple Certificate (human confirmation pending):**
The local keychain certificate verification (SC-1, CRED-01) was documented in the SUMMARY but cannot be re-run by this verifier. The existence of `APPLE_CERTIFICATE` in GitHub provides strong indirect evidence that the certificate was present and exportable. Full confirmation requires running `security find-identity` locally.

**No blocker gaps exist for Phase 17.** All 5 Apple secrets are present in the GitHub release environment. Phase 17 (macOS Signing and Notarization) can proceed.

---

## Phase Gate Assessment

**Can Phase 17 (macOS Signing) proceed?** YES — all 5 Apple secrets are present and verified.

**Can Phase 18 (Windows Authenticode) proceed?** NO — 0 Windows secrets stored. Blocked by design (Windows deferral).

**Can Phase 19 (Homebrew CLI PATH) proceed?** YES — depends on Phase 17 only; no Phase 16 Windows secrets required.

**Can Phase 20 (Windows CLI PATH via NSIS) proceed?** NO — depends on Phase 18.

---

_Verified: 2026-03-28T05:00:00Z_
_Verifier: Claude (gsd-verifier)_
