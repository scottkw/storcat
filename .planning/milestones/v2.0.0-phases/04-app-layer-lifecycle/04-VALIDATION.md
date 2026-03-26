---
phase: 4
slug: app-layer-lifecycle
status: complete
nyquist_compliant: true
wave_0_complete: true
created: 2026-03-25
updated: 2026-03-26
---

# Phase 4 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go `testing` package (app_test.go) + bash script (scripts/verify-envelopes.sh) |
| **Config file** | None (Go standard tooling) |
| **Quick run command** | `go test -v -run "TestGetCatalogHtmlPath\|TestReadHtmlFile" ./...` |
| **Full suite command** | `go test ./... && scripts/verify-envelopes.sh` |
| **Estimated runtime** | ~2 seconds (automated) |

---

## Sampling Rate

- **After every task commit:** Run `go test ./...` and `scripts/verify-envelopes.sh`
- **After every plan wave:** Full automated suite + manual smoke for WIN-04 (window restore)
- **Before `/gsd:verify-work`:** All automated checks green, WIN-04 confirmed manually
- **Max feedback latency:** ~2 seconds (automated); ~120 seconds including WIN-04 manual

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File | Status |
|---------|------|------|-------------|-----------|-------------------|------|--------|
| 04-01-01 | 01 | 1 | API-01 | unit | `go test -v -run TestGetCatalogHtmlPath_ReturnsHtmlPathWhenFileExists ./...` | app_test.go | green |
| 04-01-02 | 01 | 1 | API-01 | unit | `go test -v -run TestGetCatalogHtmlPath_ReturnsErrorWhenHtmlFileMissing ./...` | app_test.go | green |
| 04-01-03 | 01 | 1 | API-01 | unit | `go test -v -run TestGetCatalogHtmlPath_NonJsonInputAppendsHtmlExtension ./...` | app_test.go | green |
| 04-01-04 | 01 | 1 | API-02 | unit | `go test -v -run TestReadHtmlFile_ReturnsContentForValidFile ./...` | app_test.go | green |
| 04-01-05 | 01 | 1 | API-02 | unit | `go test -v -run TestReadHtmlFile_ReturnsErrorForNonexistentFile ./...` | app_test.go | green |
| 04-01-06 | 01 | 1 | API-03 | smoke | `scripts/verify-envelopes.sh` | scripts/verify-envelopes.sh | green |
| 04-01-07 | 01 | 1 | WIN-04 | manual | Manual — resize, quit, relaunch, verify restored size | N/A | manual-only |

*Status: green · red · flaky · manual-only*

---

## Wave 0 Requirements

No pre-existing automated infrastructure was needed. Tests created by Nyquist auditor on 2026-03-26 cover API-01, API-02, and API-03 directly. WIN-04 remains manual-only (requires full app restart cycle that cannot be automated without running the Wails desktop binary).

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Window size restores on relaunch | WIN-04 | Requires app restart cycle with running Wails desktop binary | Resize window, quit app, relaunch with `wails dev`, verify restored size |

---

## Validation Sign-Off

- [x] All tasks have automated verify or documented manual-only rationale
- [x] Sampling continuity: no 3 consecutive tasks without automated verify
- [x] WIN-04 manual-only rationale documented (requires running desktop binary)
- [x] No watch-mode flags
- [x] Feedback latency < 5s for automated checks
- [x] `nyquist_compliant: true` set in frontmatter

**Approval:** 2026-03-26 — Nyquist auditor gap fill complete (API-01, API-02, API-03 green)
