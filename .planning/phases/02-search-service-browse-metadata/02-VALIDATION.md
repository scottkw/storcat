---
phase: 2
slug: search-service-browse-metadata
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-25
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
| 02-01-01 | 01 | 1 | CATL-01 | unit | `go test ./internal/search/... -run TestLoadCatalog` | ❌ W0 | ⬜ pending |
| 02-01-02 | 01 | 1 | CATL-01 | unit | `go test ./internal/search/... -run TestLoadCatalogArrayFormat` | ❌ W0 | ⬜ pending |
| 02-01-03 | 01 | 1 | DATA-03 | unit | `go test ./internal/search/... -run TestBrowseCatalogsSize` | ✅ | ⬜ pending |
| 02-01-04 | 01 | 1 | DATA-04 | unit | `go test ./internal/search/... -run TestBrowseCatalogsModified` | ✅ | ⬜ pending |
| 02-01-05 | 01 | 1 | DATA-05 | unit | `go test ./internal/search/... -run TestBrowseCatalogsCreated` | ✅ | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `internal/search/service_test.go` — add `TestLoadCatalog` stub for CATL-01 bare-object format
- [ ] `internal/search/service_test.go` — add `TestLoadCatalogArrayFormat` stub for CATL-01 array-wrapped v1 format

*Existing infrastructure: `internal/search/service_test.go` exists with 3 passing tests. Wave 0 adds 2 more test cases to the same file.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Browse tab displays size column with byte values | DATA-03 | UI rendering requires visual/browser confirmation | Open Browse tab, load a catalog directory, verify size column is visible with formatted byte values |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 5s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
