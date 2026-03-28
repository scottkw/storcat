---
phase: 18-windows-authenticode-signing
verified: 2026-03-28T06:00:00Z
status: human_needed
score: 4/4 must-haves verified
re_verification: false
human_verification:
  - test: "Trigger a release tag push and confirm the Sign portable exe and Sign NSIS installer steps complete without error in the build-windows CI run"
    expected: "Both eSigner steps exit 0; the Verify Authenticode signatures step passes (signtool exits 0 for both files)"
    why_human: "Requires SSL.com eSigner secrets (ES_USERNAME, ES_PASSWORD, CREDENTIAL_ID, ES_TOTP_SECRET) to be stored in the GitHub release environment — Task 1 of the plan is a blocking human-action checkpoint. Cannot test actual signing without a live CI run with valid credentials."
  - test: "After a successful signed release run, download StorCat-*-windows-amd64.exe and StorCat-*-windows-amd64-installer.exe and run signtool verify /pa /v on each"
    expected: "Both files report 'Successfully verified' with the SSL.com certificate chain"
    why_human: "Requires a Windows machine with signtool installed and signed binaries from a real release run."
  - test: "Confirm the 10 secrets are present in the GitHub release environment (6 Apple + 4 Windows)"
    expected: "gh secret list --env release --repo scottkw/storcat returns 10 entries including ES_USERNAME, ES_PASSWORD, CREDENTIAL_ID, ES_TOTP_SECRET"
    why_human: "Secrets are user-provisioned (Task 1 checkpoint); cannot verify their presence from local filesystem."
---

# Phase 18: Windows Authenticode Signing Verification Report

**Phase Goal:** Every Windows NSIS installer and portable .exe produced by CI is signed with Authenticode before upload, suppressing or reducing SmartScreen blocking
**Verified:** 2026-03-28T06:00:00Z
**Status:** human_needed
**Re-verification:** No — initial verification

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | Windows portable .exe is signed with Authenticode before upload | VERIFIED | `Sign portable exe` step (line 146) uses `SSLcom/esigner-codesign@cf5f6c1d` on `build\bin\StorCat.exe`; step appears at line 146, before `Rename Windows binary` at line 176 and `Upload Windows artifact` at line 184 |
| 2 | Windows NSIS installer is signed with Authenticode before upload | VERIFIED | `Sign NSIS installer` step (line 158) uses `SSLcom/esigner-codesign@cf5f6c1d` on `build\bin\StorCat-amd64-installer.exe`; appears before rename (line 180) and upload (line 184) |
| 3 | signtool verify confirms valid signatures as a CI gate step | VERIFIED | `Verify Authenticode signatures` step (lines 170-174) runs `signtool verify /pa /v` on both files using `shell: cmd`; positioned after both sign steps and before rename — will exit non-zero on invalid signature, blocking the job |
| 4 | WinGet SHA256 hashes reference signed binaries (signing before upload) | VERIFIED | Step ordering confirmed: Build (143) -> Sign portable (146) -> Sign NSIS (158) -> Verify (170) -> Rename portable (176) -> Rename NSIS (180) -> Upload (184). Rename happens after signing, so the uploaded artifacts are signed binaries. |

**Score:** 4/4 truths verified

### Required Artifacts

| Artifact | Expected | Status | Details |
|----------|----------|--------|---------|
| `.github/workflows/release.yml` | Windows Authenticode signing in build-windows job | VERIFIED | File exists, contains both eSigner sign steps and signtool gate; YAML parses without error |

### Key Link Verification

| From | To | Via | Status | Details |
|------|----|-----|--------|---------|
| `.github/workflows/release.yml (build-windows)` | GitHub release environment secrets | `environment: release` | VERIFIED | `environment: release` present at line 121, between `runs-on: windows-2022` (line 120) and `steps:` (line 122) |
| `.github/workflows/release.yml (sign steps)` | upload-artifact step | sequential step ordering (sign -> verify -> rename -> upload) | VERIFIED | Grep confirms exact ordering: Build(143) Sign-portable(146) Sign-NSIS(158) Verify(170) Rename(176) Rename(180) Upload(184) |

### Data-Flow Trace (Level 4)

Not applicable. This phase modifies a CI workflow file, not a component that renders dynamic data. No data-flow trace required.

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
|----------|---------|--------|--------|
| YAML is syntactically valid | `python3 -c "import yaml; yaml.safe_load(open('.github/workflows/release.yml'))"` | YAML_VALID | PASS |
| Exactly 2 esigner-codesign uses present | `grep -c 'esigner-codesign' release.yml` | 2 | PASS |
| Both eSigner uses reference pinned SHA | `grep -n 'esigner-codesign@' release.yml` | Both lines reference `@cf5f6c1d38ad10f47e3ed9aca873f429b1a8d85b` | PASS |
| signtool verify gates both files | `grep -n 'signtool verify' release.yml` | Lines 173-174 verify both .exe files | PASS |
| shell: cmd used for signtool step | `grep -c 'shell: cmd' release.yml` | 1 | PASS |
| malware_block: false in both sign steps | `grep -c 'malware_block: false' release.yml` | 2 | PASS |
| override: true in both sign steps | `grep -c 'override: true' release.yml` | 2 | PASS |
| Commit 13bc04a2 exists in git history | `git log --oneline \| grep 13bc04a2` | `13bc04a2 feat(18-01): add Windows Authenticode signing to build-windows job` | PASS |
| No other jobs modified (linux/release have no eSigner refs) | `grep -n 'esigner\|ES_USERNAME' release.yml` | All 6 matches are within build-windows job (lines 147-165) | PASS |
| Anti-patterns absent (@develop, dlemstra, malware_block: true) | grep for forbidden patterns | NONE FOUND | PASS |

Step 7b: Live CI signing behavior cannot be tested — requires Windows runner with valid eSigner secrets.

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
|-------------|------------|-------------|--------|----------|
| WSIGN-01 | 18-01-PLAN.md | `build-windows` job signs NSIS installer with signtool.exe (or Azure Trusted Signing) | SATISFIED | `Sign NSIS installer` step (line 158) uses SSLcom/esigner-codesign to sign `build\bin\StorCat-amd64-installer.exe` — cloud HSM approach satisfies the requirement; signtool verify gate at line 173 confirms the signature |
| WSIGN-02 | 18-01-PLAN.md | `build-windows` job signs portable .exe with same certificate | SATISFIED | `Sign portable exe` step (line 146) uses same `SSLcom/esigner-codesign@cf5f6c1d` action with same credential inputs to sign `build\bin\StorCat.exe` |
| WSIGN-03 | 18-01-PLAN.md | Signing occurs before `upload-artifact` so WinGet SHA256 is computed from signed binary | SATISFIED | Step ordering verified: sign steps (146, 158) precede rename steps (176, 180) which precede upload step (184); WinGet SHA256 computed from renamed artifacts which are the signed binaries |
| WSIGN-04 | 18-01-PLAN.md | `signtool verify` step confirms valid signature before upload | SATISFIED | `Verify Authenticode signatures` step (line 170) runs `signtool verify /pa /v` on both files using `shell: cmd`; positioned before rename and upload; exits non-zero on failure, blocking the job |

No orphaned requirements found. All 4 WSIGN IDs mapped to Phase 18 in REQUIREMENTS.md are claimed in 18-01-PLAN.md and have corresponding verified implementations.

Note: WSIGN-05 (migrate to Azure Trusted Signing) is listed as a future/optional requirement in REQUIREMENTS.md and is NOT assigned to Phase 18. Not a gap.

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
|------|------|---------|----------|--------|
| None found | — | — | — | — |

No TODOs, FIXMEs, placeholder comments, empty returns, hardcoded empty data, or stub patterns found in the modified file. All signing steps contain real credentials references (not hardcoded values). The `malware_block: false` and `override: true` values are intentional configuration choices per plan decisions, not stubs.

### Human Verification Required

#### 1. Live CI Signing Run

**Test:** Push a release tag (e.g., `v2.3.0`) to the repository after storing the 4 eSigner secrets in the GitHub `release` environment. Observe the build-windows job in GitHub Actions.
**Expected:** `Sign portable exe` and `Sign NSIS installer` steps both exit 0. `Verify Authenticode signatures` step exits 0 for both files. The uploaded `StorCat-*-windows-amd64.exe` and `StorCat-*-windows-amd64-installer.exe` are Authenticode-signed.
**Why human:** Task 1 (store SSL.com eSigner secrets) is an unresolved blocking human-action checkpoint. The workflow code is correct and complete, but signing fails silently until secrets are stored. Cannot verify actual signing behavior from local filesystem inspection.

#### 2. Signed Binary Verification on Windows

**Test:** After a successful signed release run, download both Windows artifacts and run `signtool verify /pa /v StorCat-*-windows-amd64.exe` and `signtool verify /pa /v StorCat-*-windows-amd64-installer.exe` on a Windows machine.
**Expected:** Both commands output `Successfully verified: <filename>` with the SSL.com OV certificate chain.
**Why human:** Requires Windows machine with Windows SDK signtool, signed binaries from a real release, and visual inspection of the certificate details.

#### 3. GitHub Release Environment Secrets Verification

**Test:** Run `gh secret list --env release --repo scottkw/storcat` and confirm 10 secrets are present.
**Expected:** List includes `ES_USERNAME`, `ES_PASSWORD`, `CREDENTIAL_ID`, `ES_TOTP_SECRET` (4 Windows secrets) in addition to the 6 Apple signing secrets from Phase 16.
**Why human:** Secrets are user-provisioned. Cannot read secret values or count from local environment. Task 1 remains the only incomplete deliverable in this phase.

### Gaps Summary

No code gaps. All 4 must-have truths are verified in the codebase. The workflow file is complete, correctly ordered, YAML-valid, and free of anti-patterns.

The only outstanding item is **Task 1 (human-action checkpoint)**: the user must purchase the SSL.com OV code signing certificate, enroll in eSigner, extract the TOTP secret from the enrollment QR code, and store all 4 secrets in the GitHub `release` environment. This is a prerequisite for live CI signing to succeed — it is not a code gap but a user-provisioning dependency explicitly scoped as a checkpoint in the plan.

The phase goal is achievable and the mechanism is fully in place. Actual SmartScreen suppression will begin once the signing runs successfully in CI with real credentials.

---

_Verified: 2026-03-28T06:00:00Z_
_Verifier: Claude (gsd-verifier)_
