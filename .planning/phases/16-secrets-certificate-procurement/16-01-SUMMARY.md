---
phase: 16-secrets-certificate-procurement
plan: "01"
subsystem: infrastructure
tags: [github-actions, code-signing, credentials, secrets]
dependency_graph:
  requires: []
  provides: [github-release-environment, credential-rotation-runbook]
  affects: [phase-17-macos-signing, phase-18-windows-signing]
tech_stack:
  added: []
  patterns: [github-environments, deployment-branch-policies]
key_files:
  created:
    - docs/runbooks/credential-rotation.md
  modified: []
decisions:
  - "No required reviewers on release environment — tag pattern restriction (v*.*.*) is sufficient for solo developer; reviewer requirement would block automated CI on tag push (Pitfall 5)"
  - "Windows secrets documented for SSL.com eSigner (4 secrets) not traditional PFX — post-June 2023 CA/Browser Forum requirement makes OV cert export impossible"
metrics:
  duration_seconds: 94
  completed_date: "2026-03-28"
  tasks_completed: 2
  tasks_total: 2
  files_modified: 1
requirements_satisfied:
  - CRED-05
  - CRED-06
---

# Phase 16 Plan 01: GitHub Release Environment and Credential Rotation Runbook Summary

**One-liner:** GitHub `release` environment created with `v*.*.*` tag deployment policy; 9-secret credential rotation runbook committed covering Apple Developer ID and SSL.com eSigner credentials.

## What Was Built

### Task 1: GitHub Release Environment (GitHub API)

Created the `release` environment in `scottkw/storcat` via the GitHub REST API:

- Environment: `release` with `custom_branch_policies: true`
- Tag deployment policy: `v*.*.*` pattern (semantic version tags only)
- No required reviewers — solo developer project where reviewer gates would block automated CI

**Verification:**
- `gh api repos/scottkw/storcat/environments/release --jq '.name'` returns `release`
- `gh api repos/scottkw/storcat/environments/release/deployment-branch-policies --jq '.branch_policies[0].name'` returns `v*.*.*`
- `gh api repos/scottkw/storcat/environments/release --jq '.deployment_branch_policy.custom_branch_policies'` returns `true`

This environment is a prerequisite for Plans 02 and 03 (secret storage).

### Task 2: Credential Rotation Runbook

Created `docs/runbooks/credential-rotation.md` documenting all 9 signing secrets:

**Apple (5 secrets):**
- `APPLE_CERTIFICATE` — base64 .p12, expires 2027-02-01
- `APPLE_CERTIFICATE_PASSWORD` — .p12 export password
- `APPLE_CERTIFICATE_NAME` — `Developer ID Application: Ken Scott (S2K7P43927)`
- `APPLE_NOTARIZATION_PASSWORD` — app-specific password (generate at appleid.apple.com)
- `APPLE_TEAM_ID` — `S2K7P43927`

**Windows — SSL.com eSigner (4 secrets):**
- `ES_USERNAME` — SSL.com account email
- `ES_PASSWORD` — SSL.com account password
- `CREDENTIAL_ID` — signing credential ID from SSL.com dashboard
- `ES_TOTP_SECRET` — TOTP seed (shown once during enrollment — save immediately)

Runbook includes: expiry dates, step-by-step renewal instructions, `gh secret set` commands, verification commands, emergency procedures for compromise scenarios, and future improvement notes (SIGN-07, WSIGN-05, SIGN-09).

## Commits

| Task | Commit | Description |
|------|--------|-------------|
| Task 1 | (GitHub API - no file commit) | release environment created via GitHub REST API |
| Task 2 | `06c22ad4` | feat(16-01): add credential rotation runbook for all 9 signing secrets |

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] GitHub API field encoding for environment creation**
- **Found during:** Task 1
- **Issue:** Plan provided `--field` parameters for `reviewers=[]` and `deployment_branch_policy={...}` which the `gh` CLI encodes as strings, causing a 422 error ("not of type `array`"/"not of type `object`")
- **Fix:** Used `--input -` with a JSON heredoc to pass the request body correctly, which properly serializes arrays and objects
- **Files modified:** None (API call only)
- **Commit:** n/a

None other — plan executed with one auto-fixed API encoding issue.

## Known Stubs

None. The runbook is complete documentation covering all 9 secrets. No data or values are stubbed — actual secret names, expiry dates, and procedures are all present.

## Self-Check: PASSED

- [x] `docs/runbooks/credential-rotation.md` exists: FOUND
- [x] GitHub `release` environment exists: VERIFIED via API
- [x] Tag policy `v*.*.*` exists: VERIFIED via API
- [x] Commit `06c22ad4` exists: FOUND in git log
- [x] All 9 secret names present in runbook: VERIFIED (38 matches for `ES_|APPLE_|CREDENTIAL_ID`)
