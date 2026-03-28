---
gsd_state_version: 1.0
milestone: v2.3.0
milestone_name: Code Signing & Package Manager CLI
status: executing
stopped_at: Completed 16-01-PLAN.md
last_updated: "2026-03-28T03:36:27.342Z"
last_activity: 2026-03-28
progress:
  total_phases: 5
  completed_phases: 0
  total_plans: 3
  completed_plans: 1
---

# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-03-27)

**Core value:** Fast, lightweight directory catalog management — Go/Wails delivers 93% smaller binaries and 5x faster search, with full feature parity and CLI scriptability.
**Current focus:** Phase 16 — secrets-certificate-procurement

## Current Position

Phase: 16 (secrets-certificate-procurement) — EXECUTING
Plan: 2 of 3
Status: Ready to execute
Last activity: 2026-03-28

```
v2.3.0 Progress: [░░░░░░░░░░░░░░░░░░░░] 0% (0/5 phases)
Overall:         [████████████████████████████████░░░░░░░░] 75% (15/20 phases)
```

## Accumulated Context

### Decisions

Decisions are logged in PROJECT.md Key Decisions table.

Recent decisions affecting current work:

- Windows certificate type (Phase 16): Azure Trusted Signing vs. OV cert — must decide before Phase 18 begins. Azure: instant SmartScreen reputation, $9.99/month, US/Canada only. OV: ~$200-300/year, 2-8 week SmartScreen warning window, available everywhere.
- [Phase 16-secrets-certificate-procurement]: No required reviewers on release environment — tag pattern v*.*.* restriction is sufficient for solo developer; reviewer requirement would block automated CI
- [Phase 16-secrets-certificate-procurement]: Windows signing uses SSL.com eSigner (4 secrets: ES_USERNAME, ES_PASSWORD, CREDENTIAL_ID, ES_TOTP_SECRET) — post-June 2023 CA/Browser Forum requirement makes traditional OV cert PFX export impossible

### Key Research Findings

- Phase 16 is a hard prerequisite for all other v2.3.0 phases — no code changes until all 9 secrets are in GitHub
- Phase 17 (macOS) depends on Phase 16 only; Phase 18 (Windows) also depends on Phase 16 only — Phases 17 and 18 are independent of each other
- Phase 19 (Homebrew PATH) must follow Phase 17 — macOS 15+ Gatekeeper blocks symlinks from unsigned binaries
- Phase 20 (Windows PATH) must follow Phase 18 — signed installer should be in place first
- macOS signing requires 4 strict sequential steps: codesign → hdiutil → notarytool → stapler (cannot reorder)
- Windows signing must occur in the build job before `upload-artifact` to ensure WinGet SHA256 hashes match signed binaries
- `apple-actions/import-codesign-certs@v6` (SHA: b610f78) handles keychain ACL setup that prevents codesign hangs
- `dlemstra/code-sign-action` was archived October 2025 — do not use
- Existing `build/entitlements.mac.plist` from Electron era needs verification against Wails runtime before Phase 17

### Pending Todos

None.

### Blockers/Concerns

- Windows certificate type decision must be made during Phase 16 before Phase 18 begins
- Wails v2 custom NSIS script path (`build/windows/installer.nsi`) should be verified against v2.10.2 before Phase 20

## Session Continuity

Last session: 2026-03-28T03:36:27.340Z
Stopped at: Completed 16-01-PLAN.md
Resume file: None
