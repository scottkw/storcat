---
phase: 01-data-models-catalog-service
plan: 01
subsystem: catalog-service
tags: [go, json, catalog, models, tdd, wails]
dependency_graph:
  requires: []
  provides: [CreateCatalogResult, bare-json-output, symlink-traversal, html-root-node, wails-bindings]
  affects: [app.go, frontend/wailsjs]
tech_stack:
  added: []
  patterns: [TDD red-green, struct-tag-fix, service-return-type-change, wails-generate-module]
key_files:
  created:
    - internal/catalog/service_test.go
  modified:
    - pkg/models/catalog.go
    - internal/catalog/service.go
    - app.go
    - frontend/wailsjs/go/main/App.d.ts
    - frontend/wailsjs/go/models.ts
    - frontend/wailsjs/runtime/package.json
    - frontend/wailsjs/runtime/runtime.d.ts
    - frontend/wailsjs/runtime/runtime.js
decisions:
  - "JSON format uses bare object via json.Marshal(catalog) — array wrapping removed for DATA-01 requirement"
  - "CatalogItem.Contents tag changed from json:\"contents,omitempty\" to json:\"contents\" plus nil guard for DATA-02"
  - "CreateCatalog returns (*models.CreateCatalogResult, error) carrying JsonPath, HtmlPath, FileCount, TotalSize"
  - "frontend/dist placeholder created temporarily to unblock wails generate module (gitignored, not committed)"
metrics:
  duration: "~12 minutes"
  completed: "2026-03-24"
  tasks_completed: 2
  files_modified: 8
  files_created: 1
---

# Phase 1 Plan 1: Fix Catalog Models, JSON Output, Traversal, HTML Root, and Wails Bindings Summary

**One-liner:** Fixed 5 Go catalog bugs (bare-object JSON, empty-dir contents, symlink traversal, HTML root connector, CreateCatalog result metadata) and regenerated Wails TypeScript bindings with CreateCatalogResult type.

## What Was Built

This plan addressed the foundation layer of the Go/Wails v2.0.0 migration. Five concrete divergences from Electron v1.2.3 were fixed in the Go backend using TDD:

1. **DATA-01 (JSON format):** `writeJSONFile` now calls `json.Marshal(catalog)` directly instead of wrapping in `[]*models.CatalogItem{catalog}`. Output is `{...}` not `[{...}]`.

2. **DATA-02 (empty dir contents):** `CatalogItem.Contents` tag changed from `json:"contents,omitempty"` to `json:"contents"`, plus a nil guard added after the traversal loop to ensure `contents = []*models.CatalogItem{}` when all entries are hidden (skipped).

3. **CATL-03 (symlink traversal):** `traverseDirectory` changed from `os.Lstat(dirPath)` to `os.Stat(dirPath)` so symlink targets are followed, matching Electron's `fs.stat` behavior.

4. **CATL-04 (HTML root node):** Removed the special-case `if item.Name == "./"` block that emitted `./<br>\n` without a connector or size bracket. Root node now falls through to the standard path: `└── [size]&nbsp;&nbsp;.<br>`.

5. **CATL-02 (CreateCatalog return type):** `CreateCatalog` changed from `error` to `(*models.CreateCatalogResult, error)`. New `CreateCatalogResult` struct added to `pkg/models/catalog.go` with fields: `JsonPath`, `HtmlPath`, `FileCount`, `TotalSize`, `CopyJsonPath`, `CopyHtmlPath`.

`App.CreateCatalog` in `app.go` was updated to propagate the new return type. Wails TypeScript bindings were regenerated, so `App.d.ts` now declares `CreateCatalog` as `Promise<models.CreateCatalogResult>` and `models.ts` contains the `CreateCatalogResult` class.

## Tasks Completed

| Task | Description | Commit |
|------|-------------|--------|
| 1 | Fix catalog models, JSON output, traversal, and HTML root node (TDD) | 2fd4fcfb |
| 2 | Update App.CreateCatalog signature and regenerate Wails bindings | 74572c3c |

## Test Results

All 5 unit tests pass:
- `TestWriteJSONFile_BareObject` — JSON output starts with `{` not `[`
- `TestEmptyDirContents` — empty directory serializes as `"contents":[]`
- `TestSymlinkTraversal` — symlink target included with correct size
- `TestHTMLRootNode` — root node HTML contains `└── ` connector and `[` size bracket
- `TestCreateCatalogResult` — CreateCatalog returns non-nil result with populated paths and counts

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] `frontend/dist` missing prevented `wails generate module`**

- **Found during:** Task 2
- **Issue:** `wails generate module` and `go build ./...` both fail with `pattern all:frontend/dist: no matching files found` because `frontend/dist` is the Wails embed target and it requires a build from npm/vite first.
- **Fix:** Created `frontend/dist/index.html` as a minimal placeholder to allow `wails generate module` to parse Go types. The placeholder is gitignored (`dist/` in `.gitignore`) and not committed.
- **Files modified:** `frontend/dist/index.html` (ephemeral, gitignored)
- **Commit:** N/A (gitignored)

**2. [Rule 1 - TDD] Test helper `min()` added**

- **Found during:** Task 1 RED phase
- **Issue:** Go 1.26.1 stdlib `min` is available as a builtin, but using it directly in test error formatting for string slicing required a local helper to avoid any ambiguity with older Go behavior.
- **Fix:** Added a local `min(a, b int) int` helper in `service_test.go`.
- **Files modified:** `internal/catalog/service_test.go`
- **Commit:** 2fd4fcfb

## Known Stubs

None — all 5 behaviors are fully implemented and verified by passing tests.

## Requirements Addressed

| Requirement | Status |
|-------------|--------|
| DATA-01 | Done — bare object JSON output |
| DATA-02 | Done — empty dir contents serialize as `[]` |
| CATL-02 | Done — CreateCatalog returns result metadata |
| CATL-03 | Done — symlinks followed via os.Stat |
| CATL-04 | Done — HTML root node has connector+size |

## Self-Check: PASSED

All created files verified present. Both task commits confirmed in git log.
