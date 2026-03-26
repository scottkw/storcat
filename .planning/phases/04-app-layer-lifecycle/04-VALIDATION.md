---
phase: 4
slug: app-layer-lifecycle
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-25
---

# Phase 4 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | None — no automated test framework in project |
| **Config file** | None |
| **Quick run command** | `wails dev` — manual test in running app |
| **Full suite command** | Manual: exercise all IPC wrappers, verify envelope shapes in browser console, test catalog modal HTML flow |
| **Estimated runtime** | ~120 seconds (manual) |

---

## Sampling Rate

- **After every task commit:** Run `wails dev`, exercise the affected wrapper manually, verify envelope shape in browser console
- **After every plan wave:** Full modal flow: browse -> click catalog -> verify HTML preview; plus window resize -> close -> reopen -> verify size
- **Before `/gsd:verify-work`:** All 6 smoke behaviors confirmed
- **Max feedback latency:** 120 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 04-01-01 | 01 | 1 | API-01 | smoke | Manual — browse catalogs, click catalog with HTML, verify modal opens | N/A | ⬜ pending |
| 04-01-02 | 01 | 1 | API-01 | smoke | Manual — delete `.html` file, click catalog, verify error message | N/A | ⬜ pending |
| 04-01-03 | 01 | 1 | API-02 | smoke | Manual — verify HTML content renders in modal iframe | N/A | ⬜ pending |
| 04-01-04 | 01 | 1 | API-02 | smoke | Manual — pass nonexistent path, verify error shown not crash | N/A | ⬜ pending |
| 04-01-05 | 01 | 1 | API-03 | smoke | Manual — exercise all 7 wrapper paths in running app | N/A | ⬜ pending |
| 04-01-06 | 01 | 1 | WIN-04 | smoke | Manual — resize, quit, relaunch, verify restored size | N/A | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

Existing infrastructure covers all phase requirements. No test framework exists and none is expected (TEST-01 through TEST-03 deferred per REQUIREMENTS.md). Manual verification protocol covers all requirements.

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| `getCatalogHtmlPath` returns `{success, htmlPath}` for existing file | API-01 | No test framework; desktop app requires running Wails runtime | Browse catalogs, click one with HTML, verify modal opens with content |
| `getCatalogHtmlPath` returns `{success: false}` for missing file | API-01 | Desktop app UI interaction required | Delete an `.html` file, click the catalog, verify error message shown |
| `readHtmlFile` returns `{success, content}` for valid file | API-02 | Requires Wails runtime + browser rendering | Verify HTML content renders in modal iframe |
| `readHtmlFile` returns `{success: false}` for missing file | API-02 | Desktop app error path verification | Pass nonexistent path, verify error shown not crash |
| All 7 wrappers return `{success, ...}` envelopes | API-03 | Requires exercising each wrapper through UI | Call each wrapper, check console for envelope shape |
| Window size restores on relaunch | WIN-04 | Requires app restart cycle | Resize, quit, relaunch, verify restored size |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 120s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
