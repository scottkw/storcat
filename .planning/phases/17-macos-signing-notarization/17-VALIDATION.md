---
phase: 17
slug: macos-signing-notarization
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-27
---

# Phase 17 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | GitHub Actions (CI workflow validation) |
| **Config file** | `.github/workflows/release.yml` |
| **Quick run command** | `act -j build-macos --dryrun` (local) or push tag to trigger CI |
| **Full suite command** | Push a release tag and verify workflow completes |
| **Estimated runtime** | ~5-10 minutes (notarization polling) |

---

## Sampling Rate

- **After every task commit:** Validate YAML syntax with `yq eval '.jobs.build-macos' .github/workflows/release.yml`
- **After every plan wave:** Dry-run workflow validation
- **Before `/gsd:verify-work`:** Full CI run with signed artifact
- **Max feedback latency:** 600 seconds (notarization)

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 17-01-01 | 01 | 1 | SIGN-01 | integration | `yq '.jobs.build-macos.steps[] | select(.uses == "apple-actions/import-codesign-certs*")' .github/workflows/release.yml` | ❌ W0 | ⬜ pending |
| 17-01-02 | 01 | 1 | SIGN-02 | integration | `grep 'codesign.*--options runtime' .github/workflows/release.yml` | ❌ W0 | ⬜ pending |
| 17-02-01 | 02 | 2 | SIGN-03 | integration | `grep 'notarytool submit' .github/workflows/release.yml` | ❌ W0 | ⬜ pending |
| 17-02-02 | 02 | 2 | SIGN-04 | integration | `grep 'stapler staple' .github/workflows/release.yml` | ❌ W0 | ⬜ pending |
| 17-02-03 | 02 | 2 | SIGN-05 | integration | `grep 'spctl --assess' .github/workflows/release.yml` | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `build/darwin/entitlements.plist` — entitlements file for hardened runtime
- [ ] APPLE_ID secret or hardcoded email — required for notarytool credentials

*Existing CI infrastructure (release.yml) covers workflow scaffolding.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Gatekeeper acceptance on macOS 15+ | SIGN-06 | Requires real macOS machine download | Download DMG from release, open on macOS 15+, verify no Gatekeeper prompt |
| Notarization "Accepted" status | SIGN-03 | Requires Apple notary service | Check `xcrun notarytool log` output in CI logs |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 600s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
