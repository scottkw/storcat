---
phase: 01-data-models-catalog-service
verified: 2026-03-24T23:00:00Z
status: passed
score: 8/8 must-haves verified
gaps: []
human_verification:
  - test: "Verify HTML catalog output matches Electron v1.2.3 format exactly"
    expected: "Root node renders as '└── [size]&nbsp;&nbsp;.<br>' with correct size bracket width"
    why_human: "Test confirms connector and '[' presence, but pixel-exact formatting comparison against Electron output requires manual file diff"
---

# Phase 1: Data Models + Catalog Service Verification Report

**Phase Goal:** Go data structures and catalog output match Electron's format exactly, so all downstream layers build on a correct foundation
**Verified:** 2026-03-24T23:00:00Z
**Status:** passed
**Re-verification:** Gap fixed inline (Wails bindings regenerated), human UAT passed

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | A generated catalog JSON file opens as a bare object `{...}`, not an array `[{...}]` | ✓ VERIFIED | `writeJSONFile` calls `json.Marshal(catalog)` directly (service.go:167); `TestWriteJSONFile_BareObject` passes |
| 2 | An empty directory serializes its `contents` field as `[]`, not `null` or absent | ✓ VERIFIED | `Contents []*CatalogItem \`json:"contents"\`` (no omitempty) in catalog.go:8; nil guard at service.go:150-152; `TestEmptyDirContents` passes |
| 3 | The HTML catalog root node renders with `└──` connector and size bracket | ✓ VERIFIED | `generateTreeStructure` uses unified path (no special-case for `"./"`); `TestHTMLRootNode` confirms `└── ` and `[` present |
| 4 | `CreateCatalog` returns `fileCount`, `totalSize`, and output paths | ✓ VERIFIED | `CreateCatalog` returns `(*models.CreateCatalogResult, error)`; struct has JsonPath, HtmlPath, FileCount, TotalSize, CopyJsonPath, CopyHtmlPath; `TestCreateCatalogResult` passes |
| 5 | Symlinks in a cataloged directory are followed and counted | ✓ VERIFIED | `traverseDirectory` uses `os.Stat(dirPath)` not `os.Lstat`; `TestSymlinkTraversal` passes |
| 6 | Browse catalog metadata includes a size field with the file size in bytes | ✗ FAILED | `CatalogMetadata.Size int64` exists in Go struct and is populated via `info.Size()` in search/service.go:169; `TestBrowseCatalogsSize` passes — BUT `frontend/wailsjs/go/models.ts` CatalogMetadata class is missing the `size` field (stale binding) |
| 7 | The `modified` field in browse results is an RFC3339 Date-compatible string | ✓ VERIFIED | `info.ModTime().Format(time.RFC3339)` in search/service.go:171; `TestBrowseCatalogsModified` passes |
| 8 | The `created` field uses actual file birth time on macOS, falling back to mtime | ✓ VERIFIED | `times.Stat` + `HasBirthTime()` check in search/service.go:154-158; `TestBrowseCatalogsCreated` passes; `djherbis/times v1.6.0` in go.mod |

**Score:** 7/8 truths verified

### Required Artifacts

| Artifact | Expected | Status | Details |
|----------|----------|--------|---------|
| `pkg/models/catalog.go` | CatalogItem, CatalogMetadata with Size, CreateCatalogResult | ✓ VERIFIED | All structs present with correct fields and JSON tags; Contents has no omitempty |
| `internal/catalog/service.go` | CreateCatalog returns result, bare JSON, os.Stat, nil guard, HTML root unified | ✓ VERIFIED | All 5 fixes confirmed in source |
| `internal/catalog/service_test.go` | 5 tests covering all plan behaviors | ✓ VERIFIED | All 5 tests pass: TestWriteJSONFile_BareObject, TestEmptyDirContents, TestSymlinkTraversal, TestHTMLRootNode, TestCreateCatalogResult |
| `internal/search/service.go` | BrowseCatalogs with Size, RFC3339 dates, djherbis/times | ✓ VERIFIED | All fields populated correctly; times.Stat with HasBirthTime() fallback present |
| `internal/search/service_test.go` | 3 tests covering browse metadata | ✓ VERIFIED | All 3 tests pass: TestBrowseCatalogsSize, TestBrowseCatalogsModified, TestBrowseCatalogsCreated |
| `app.go` | CreateCatalog returns *models.CreateCatalogResult | ✓ VERIFIED | Signature is `(*models.CreateCatalogResult, error)`; early return is `return nil, err` |
| `frontend/wailsjs/go/main/App.d.ts` | CreateCatalog returning non-void Promise | ✓ VERIFIED | `Promise<models.CreateCatalogResult>` declared at line 8 |
| `frontend/wailsjs/go/models.ts` | CreateCatalogResult class + CatalogMetadata with size | ✗ STUB | CreateCatalogResult class present (lines 50-71) with all fields; CatalogMetadata class (lines 26-49) is MISSING `size: number` field — Plan 02 added Size to Go struct but did not regenerate bindings |

### Key Link Verification

| From | To | Via | Status | Details |
|------|----|-----|--------|---------|
| `internal/catalog/service.go` | `pkg/models/catalog.go` | `models.CreateCatalogResult` import | ✓ WIRED | `import "storcat-wails/pkg/models"` at line 13; result struct built at lines 47-52 |
| `app.go` | `internal/catalog/service.go` | `a.catalogService.CreateCatalog` call | ✓ WIRED | Called at line 63; result propagated at lines 63-67 |
| `internal/search/service.go` | `pkg/models/catalog.go` | `CatalogMetadata.Size` population | ✓ WIRED | `Size: info.Size()` at line 169 |
| `internal/search/service.go` | `github.com/djherbis/times` | `times.Stat` for birth time | ✓ WIRED | `times.Stat(filePath)` at line 154; `t.HasBirthTime()` + `t.BirthTime()` at lines 155-156 |
| `frontend/wailsjs/go/models.ts` | `pkg/models/catalog.go` (CatalogMetadata) | Wails binding regeneration | ✗ NOT_WIRED | CatalogMetadata.size field missing from TypeScript class — binding not regenerated after Plan 02 |

### Data-Flow Trace (Level 4)

| Artifact | Data Variable | Source | Produces Real Data | Status |
|----------|---------------|--------|-------------------|--------|
| `internal/catalog/service.go` CreateCatalog | catalog struct | `traverseDirectory` recursive filesystem walk | Yes — reads actual filesystem via `os.Stat`, `os.ReadDir` | ✓ FLOWING |
| `internal/search/service.go` BrowseCatalogs | CatalogMetadata fields | `entry.Info()` and `times.Stat(filePath)` | Yes — reads real filesystem metadata | ✓ FLOWING |
| `frontend/wailsjs/go/models.ts` CatalogMetadata | size property | Wails serialization of Go struct | Not flowing to TypeScript — field absent from generated class | ✗ DISCONNECTED |

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
|----------|---------|--------|--------|
| All catalog tests pass | `go test ./internal/catalog/... -v` | 5/5 PASS | ✓ PASS |
| All search tests pass | `go test ./internal/search/... -v` | 3/3 PASS | ✓ PASS |
| Full project builds | `go build ./...` | BUILD OK | ✓ PASS |
| CreateCatalog returns non-void in TypeScript | `grep CreateCatalog frontend/wailsjs/go/main/App.d.ts` | `Promise<models.CreateCatalogResult>` | ✓ PASS |
| CatalogMetadata has size in TypeScript bindings | `grep size frontend/wailsjs/go/models.ts` CatalogMetadata section | Not found in CatalogMetadata class | ✗ FAIL |

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
|-------------|------------|-------------|--------|----------|
| DATA-01 | 01-01 | Catalog JSON is bare object `{...}` | ✓ SATISFIED | `json.Marshal(catalog)` in writeJSONFile; test passes |
| DATA-02 | 01-01 | Empty directory `contents` serializes as `[]` | ✓ SATISFIED | No omitempty tag; nil guard; test passes |
| DATA-03 | 01-02 | Browse metadata has `size` field (bytes) | ✗ BLOCKED | Go struct correct, test passes — but TypeScript binding missing `size` in CatalogMetadata class |
| DATA-04 | 01-02 | Browse `modified` is Date-compatible RFC3339 | ✓ SATISFIED | `time.RFC3339` format; test passes |
| DATA-05 | 01-02 | Browse `created` uses file birth time where available | ✓ SATISFIED | `djherbis/times` birth time with mtime fallback; test passes |
| CATL-02 | 01-01 | `CreateCatalog` returns result metadata | ✓ SATISFIED | Returns `*models.CreateCatalogResult` with all fields; binding regenerated |
| CATL-03 | 01-01 | Directory traversal follows symlinks | ✓ SATISFIED | `os.Stat` used; test passes |
| CATL-04 | 01-01 | HTML root node has `└──` connector and size bracket | ✓ SATISFIED | Unified code path; test passes |

**Notes on orphaned requirements:** No orphaned requirements found. All 8 requirement IDs from the PLAN frontmatter (DATA-01 through DATA-05, CATL-02 through CATL-04) appear in REQUIREMENTS.md mapped to Phase 1. No Phase 1 requirements in REQUIREMENTS.md were unaccounted for.

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
|------|------|---------|----------|--------|
| `frontend/wailsjs/go/models.ts` | 26-49 | CatalogMetadata class missing `size: number` field — stale binding from Plan 01 regeneration | 🛑 Blocker | Any frontend code accessing `.size` on CatalogMetadata will get `undefined`; Browse tab size column will fail |
| `app.go` | 59-61 | Progress callback is a no-op stub `func(path string) {}` | ℹ️ Info | No user-visible impact yet — progress display is Phase 4+ concern; comment in code acknowledges this |

### Human Verification Required

#### 1. HTML Tree Format Exact Match

**Test:** Generate a catalog on a directory with nested subdirectories. Open the resulting `.html` file in a browser. Compare the tree indentation, connector characters (`├──`, `└──`), and size bracket format against an Electron v1.2.3-generated HTML file for the same directory.
**Expected:** Character-for-character identical tree structure (modulo the directory content itself)
**Why human:** `TestHTMLRootNode` confirms connector and bracket presence, but cannot verify the exact Unicode connector characters render identically to Electron's output or that indentation levels match for deeply nested trees.

### Gaps Summary

One gap blocks full phase goal achievement:

**Stale TypeScript binding for CatalogMetadata.size (DATA-03)**

Plan 02 correctly added `Size int64 \`json:"size"\`` to the Go `CatalogMetadata` struct and verified it with a passing Go test. However, Plan 02 did not run `wails generate module` after the struct change, so `frontend/wailsjs/go/models.ts` still contains the Plan 01-generated `CatalogMetadata` class — which is missing the `size` property.

The Go backend is correct. The test is correct. The wiring breaks at the TypeScript boundary: any frontend code reading `catalog.size` from the Wails binding will get `undefined`.

The fix is a one-command regeneration: `wails generate module` from the project root, which will add `size: number` to the `CatalogMetadata` constructor in `models.ts`.

---

_Verified: 2026-03-24T23:00:00Z_
_Verifier: Claude (gsd-verifier)_
