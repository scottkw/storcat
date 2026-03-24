---
phase: 1
slug: data-models-catalog-service
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-24
---

# Phase 1 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go `testing` (stdlib) |
| **Config file** | none (Go convention: `*_test.go` alongside source) |
| **Quick run command** | `go test ./internal/catalog/... ./pkg/models/...` |
| **Full suite command** | `go test ./...` |
| **Estimated runtime** | ~5 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./internal/catalog/... ./pkg/models/...`
- **After every plan wave:** Run `go test ./...`
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** 5 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 01-01-01 | 01 | 1 | DATA-01 | unit | `go test ./internal/catalog/... -run TestWriteJSONFile` | ❌ W0 | ⬜ pending |
| 01-01-02 | 01 | 1 | DATA-02 | unit | `go test ./internal/catalog/... -run TestEmptyDirContents` | ❌ W0 | ⬜ pending |
| 01-01-03 | 01 | 1 | CATL-02 | unit | `go test ./internal/catalog/... -run TestCreateCatalogResult` | ❌ W0 | ⬜ pending |
| 01-01-04 | 01 | 1 | CATL-03 | unit | `go test ./internal/catalog/... -run TestSymlinkTraversal` | ❌ W0 | ⬜ pending |
| 01-01-05 | 01 | 1 | CATL-04 | unit | `go test ./internal/catalog/... -run TestHTMLRootNode` | ❌ W0 | ⬜ pending |
| 01-02-01 | 02 | 1 | DATA-03 | unit | `go test ./internal/search/... -run TestBrowseCatalogsSize` | ❌ W0 | ⬜ pending |
| 01-02-02 | 02 | 1 | DATA-04 | unit | `go test ./internal/search/... -run TestBrowseCatalogsModified` | ❌ W0 | ⬜ pending |
| 01-02-03 | 02 | 1 | DATA-05 | unit | `go test ./internal/search/... -run TestBrowseCatalogsCreated` | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `internal/catalog/service_test.go` — stubs for DATA-01, DATA-02, CATL-02, CATL-03, CATL-04
- [ ] `internal/search/service_test.go` — stubs for DATA-03, DATA-04, DATA-05
- [ ] Test fixtures: small scratch directory with a symlink for CATL-03

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| HTML visual tree appearance | CATL-04 | Visual comparison with Electron output | Open generated HTML in browser, compare root node rendering with v1 screenshot |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 5s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
