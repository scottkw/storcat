---
phase: 2
slug: search-service-browse-metadata
status: approved
nyquist_compliant: true
wave_0_complete: true
created: 2026-03-25
audited: 2026-03-26
---

# Phase 2 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go `testing` (stdlib) |
| **Config file** | none (Go convention: `*_test.go` alongside source) |
| **Quick run command** | `go test ./internal/search/...` |
| **Full suite command** | `go test ./...` |
| **Estimated runtime** | ~5 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./internal/search/...`
- **After every plan wave:** Run `go test ./...`
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** 5 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 02-01-01 | 01 | 1 | CATL-01 | unit | `go test ./internal/search/... -run TestLoadCatalog` | ✅ | ✅ green |
| 02-01-02 | 01 | 1 | CATL-01 | unit | `go test ./internal/search/... -run TestLoadCatalogArrayFormat` | ✅ | ✅ green |
| 02-01-03 | 01 | 1 | DATA-03 | unit | `go test ./internal/search/... -run TestBrowseCatalogsSize` | ✅ | ✅ green |
| 02-01-04 | 01 | 1 | DATA-04 | unit | `go test ./internal/search/... -run TestBrowseCatalogsModified` | ✅ | ✅ green |
| 02-01-05 | 01 | 1 | DATA-05 | unit | `go test ./internal/search/... -run TestBrowseCatalogsCreated` | ✅ | ✅ green |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [x] `internal/search/service_test.go` — add `TestLoadCatalog` stub for CATL-01 bare-object format
- [x] `internal/search/service_test.go` — add `TestLoadCatalogArrayFormat` stub for CATL-01 array-wrapped v1 format

*Existing infrastructure: `internal/search/service_test.go` exists with 7 passing tests. Wave 0 items completed during phase execution.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Browse tab displays size column with byte values | DATA-03 | UI rendering requires visual/browser confirmation | Open Browse tab, load a catalog directory, verify size column is visible with formatted byte values |

---

## Validation Sign-Off

- [x] All tasks have `<automated>` verify or Wave 0 dependencies
- [x] Sampling continuity: no 3 consecutive tasks without automated verify
- [x] Wave 0 covers all MISSING references
- [x] No watch-mode flags
- [x] Feedback latency < 5s
- [x] `nyquist_compliant: true` set in frontmatter

**Approval:** approved 2026-03-26

---

## Validation Audit 2026-03-26

| Metric | Count |
|--------|-------|
| Gaps found | 0 |
| Resolved | 0 |
| Escalated | 0 |

All 5 requirements (CATL-01, DATA-03, DATA-04, DATA-05) have automated verification via 7 passing Go tests in `internal/search/service_test.go`. Wave 0 items (TestLoadCatalog, TestLoadCatalogArrayFormat) were completed during phase execution and are now green.
