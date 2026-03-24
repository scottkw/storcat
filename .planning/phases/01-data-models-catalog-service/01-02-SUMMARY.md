---
phase: 01-data-models-catalog-service
plan: 02
subsystem: search-service
tags: [data-models, browse, metadata, tdd, go]
dependency_graph:
  requires: [01-01]
  provides: [correct-browse-metadata, size-field, rfc3339-dates]
  affects: [frontend-browse-tab]
tech_stack:
  added: [github.com/djherbis/times v1.6.0]
  patterns: [birth-time-with-mtime-fallback, rfc3339-string-formatting]
key_files:
  created:
    - internal/search/service_test.go
  modified:
    - pkg/models/catalog.go
    - internal/search/service.go
    - go.mod
    - go.sum
decisions:
  - "CatalogMetadata.Size is int64 populated via info.Size() in BrowseCatalogs"
  - "Created field uses djherbis/times BirthTime() on macOS, falls back to ModTime() elsewhere"
  - "Both Created and Modified use time.RFC3339 constant for consistent JS Date-compatible strings"
metrics:
  duration: "~5 minutes"
  completed: "2026-03-24T21:58:48Z"
  tasks: 1
  files_changed: 5
---

# Phase 01 Plan 02: Browse Catalog Metadata Fix Summary

## One-liner

Browse catalog metadata now exposes file size in bytes, RFC3339 Created (using macOS birth time via djherbis/times with mtime fallback), and RFC3339 Modified fields.

## What Was Built

Fixed `BrowseCatalogs` in `internal/search/service.go` and extended `CatalogMetadata` in `pkg/models/catalog.go` to address three data model gaps (DATA-03, DATA-04, DATA-05):

1. **DATA-03 — Size field**: Added `Size int64 \`json:"size"\`` to `CatalogMetadata` struct and populated it with `info.Size()` in `BrowseCatalogs`.

2. **DATA-04 — Modified RFC3339**: Changed `Modified` format from `"2006-01-02T15:04:05Z07:00"` to `time.RFC3339` (functionally equivalent, now uses the named constant).

3. **DATA-05 — Created birth time**: Changed `Created` from the non-JS-parseable `"2006-01-02 15:04:05"` format to RFC3339. Uses `djherbis/times` to get actual file birth time on macOS (and Windows); falls back to `ModTime()` on platforms without birth time support.

Tests written first (TDD) in `internal/search/service_test.go`:
- `TestBrowseCatalogsSize` — verifies Size equals actual file bytes
- `TestBrowseCatalogsModified` — verifies Modified parses via `time.Parse(time.RFC3339, ...)`
- `TestBrowseCatalogsCreated` — verifies Created parses via `time.Parse(time.RFC3339, ...)` and contains `T` separator (confirming it is not the old non-RFC3339 format)

## Tasks Completed

| Task | Name | Commit | Files |
|------|------|--------|-------|
| 1 | Add Size, fix date formats, add birth time (TDD) | 084b98ea | pkg/models/catalog.go, internal/search/service.go, internal/search/service_test.go, go.mod, go.sum |

## Decisions Made

- `djherbis/times` v1.6.0 added as direct dependency (not indirect) — only Go community library with cross-platform `BirthTime()` support for macOS `Birthtimespec`, Windows `CreationTime`, and Linux fallback
- `time.RFC3339` constant used for both Created and Modified for consistency and clarity
- `Size int64` placed between `Filename` and `Created` in struct field order per plan spec

## Verification Results

```
=== RUN   TestBrowseCatalogsSize
--- PASS: TestBrowseCatalogsSize (0.00s)
=== RUN   TestBrowseCatalogsModified
--- PASS: TestBrowseCatalogsModified (0.00s)
=== RUN   TestBrowseCatalogsCreated
--- PASS: TestBrowseCatalogsCreated (0.00s)
PASS
ok  storcat-wails/internal/search  0.221s

ok  storcat-wails/internal/catalog  (cached - Plan 01 tests pass, no regressions)
build successful (go build ./...)
```

## Deviations from Plan

None - plan executed exactly as written.

## Known Stubs

None — all fields are populated from real filesystem data.

## Self-Check: PASSED

Files exist:
- FOUND: /Users/ken/dev/storcat/internal/search/service_test.go
- FOUND: /Users/ken/dev/storcat/pkg/models/catalog.go
- FOUND: /Users/ken/dev/storcat/internal/search/service.go

Commits exist:
- FOUND: 084b98ea
