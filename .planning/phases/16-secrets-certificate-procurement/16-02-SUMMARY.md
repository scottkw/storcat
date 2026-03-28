---
phase: 16-secrets-certificate-procurement
plan: 02
subsystem: infra
tags: [apple, codesigning, p12, notarization, github-secrets]

requires:
  - phase: 16-01
    provides: GitHub release environment for secret storage
provides:
  - 5 Apple signing secrets in GitHub release environment
  - Verified Developer ID Application certificate (expiry 2027-02-01)
affects: [17-macos-signing-notarization, 19-release-workflow]

tech-stack:
  added: []
  patterns: [gh secret set for environment-scoped secrets]

key-files:
  created: []
  modified: []

key-decisions:
  - "Certificate identity confirmed as Developer ID Application: Ken Scott (S2K7P43927)"
  - "App-specific password named StorCat CI for notarization traceability"

patterns-established:
  - "Environment-scoped secrets via gh secret set --env release --repo scottkw/storcat"

requirements-completed: [CRED-01, CRED-02]

duration: 5min
completed: 2026-03-27
---

# Plan 16-02: Apple Signing Secrets Summary

**5 Apple signing secrets stored in GitHub release environment after verifying Developer ID Application certificate valid until 2027-02-01**

## Performance

- **Duration:** 5 min
- **Started:** 2026-03-27T20:40:00Z
- **Completed:** 2026-03-27T20:46:00Z
- **Tasks:** 3
- **Files modified:** 0 (all GitHub API operations)

## Accomplishments
- Developer ID Application certificate confirmed valid on local keychain (expiry Feb 1, 2027)
- .p12 exported with private key, base64-encoded, and stored as APPLE_CERTIFICATE
- All 5 Apple secrets stored: APPLE_CERTIFICATE, APPLE_CERTIFICATE_PASSWORD, APPLE_CERTIFICATE_NAME, APPLE_NOTARIZATION_PASSWORD, APPLE_TEAM_ID
- App-specific password generated for notarization (named "StorCat CI")

## Task Commits

1. **Task 1: Verify certificate** - No commit (verification only)
2. **Task 2: Export .p12 and generate app-specific password** - Human checkpoint (GUI/web actions)
3. **Task 3: Store 5 Apple secrets** - GitHub API operations (no file commits)

## Files Created/Modified
None — all operations were GitHub API secret storage and local keychain verification.

## Decisions Made
- Used exact secret names matching Phase 17 workflow expectations
- Certificate identity string taken directly from `security find-identity` output

## Deviations from Plan
None - plan executed exactly as written.

## Issues Encountered
- .p12 secure deletion blocked by permission policy — user advised to delete manually

## User Setup Required
User should manually delete `~/Documents/storcat-dev-id.p12` after confirming secrets are correctly stored.

## Next Phase Readiness
- All 5 Apple secrets ready for Phase 17 (macOS Signing and Notarization)
- Phase 17 can reference APPLE_CERTIFICATE, APPLE_CERTIFICATE_PASSWORD, APPLE_CERTIFICATE_NAME, APPLE_NOTARIZATION_PASSWORD, APPLE_TEAM_ID

---
*Phase: 16-secrets-certificate-procurement*
*Completed: 2026-03-27*
