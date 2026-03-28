---
phase: 16-secrets-certificate-procurement
plan: 03
subsystem: infra
tags: [windows, codesigning, authenticode, ssl-com, esigner, deferred]

requires:
  - phase: 16-01
    provides: GitHub release environment for secret storage
provides:
  - Decision to defer Windows code signing to a future milestone
affects: [18-windows-authenticode-signing, 20-winget-cli-path]

tech-stack:
  added: []
  patterns: []

key-files:
  created: []
  modified: []

key-decisions:
  - "Windows code signing deferred — SSL.com eSigner OV was recommended but user chose to defer"
  - "Phase 18 (Windows Authenticode) and Phase 20 (WinGet CLI PATH) are blocked until Windows cert is obtained"
  - "macOS signing (Phases 17, 19) can proceed independently"

patterns-established: []

requirements-completed: [CRED-03]

duration: 2min
completed: 2026-03-27
---

# Plan 16-03: Windows Signing Secrets Summary

**Windows code signing deferred to future milestone — 0 Windows secrets stored, Phases 18 and 20 blocked**

## Performance

- **Duration:** 2 min
- **Started:** 2026-03-27T20:48:00Z
- **Completed:** 2026-03-27T20:50:00Z
- **Tasks:** 3 (Task 1 decision checkpoint, Tasks 2-3 skipped per deferral)
- **Files modified:** 0

## Accomplishments
- Windows code signing vendor decision made: defer to future milestone (option-c)
- Documented that SSL.com eSigner OV ($20/mo) is the recommended path when ready
- Confirmed 5 Apple secrets are present in release environment (partial phase gate)

## Task Commits

1. **Task 1: Confirm vendor** - Decision checkpoint: user selected option-c (defer)
2. **Task 2: Enroll in SSL.com** - Skipped (deferral)
3. **Task 3: Store Windows secrets** - Skipped (deferral)

## Files Created/Modified
None — decision checkpoint only.

## Decisions Made
- **Deferred Windows signing**: User chose option-c over SSL.com eSigner OV ($20/mo) and Azure Artifact Signing (ineligible). Windows users will continue seeing SmartScreen warnings until a future milestone addresses this.
- CRED-03 (vendor selection) is satisfied by the documented decision. CRED-04 (secret storage) and partial CRED-05 (Windows portion of 9-secret gate) remain incomplete.

## Deviations from Plan
None - deferral path was explicitly designed into the plan.

## Issues Encountered
None.

## User Setup Required
None for now. When ready to proceed with Windows signing:
1. Purchase SSL.com OV code signing certificate (~$20/mo)
2. Enroll in eSigner
3. Store 4 secrets: ES_USERNAME, ES_PASSWORD, CREDENTIAL_ID, ES_TOTP_SECRET
4. See `docs/runbooks/credential-rotation.md` for full details

## Next Phase Readiness
- Phase 17 (macOS Signing and Notarization): READY — all 5 Apple secrets present
- Phase 18 (Windows Authenticode Signing): BLOCKED — no Windows cert
- Phase 19 (Release Workflow): PARTIALLY READY — macOS signing available, Windows signing skipped
- Phase 20 (WinGet CLI PATH): BLOCKED — requires signed Windows binaries

---
*Phase: 16-secrets-certificate-procurement*
*Completed: 2026-03-27*
