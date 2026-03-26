---
phase: 02-search-service-browse-metadata
plan: 01
subsystem: search-service, app-layer, browse-tab
tags: [go, tdd, wails-bindings, react, typescript, catalog]
dependency_graph:
  requires: [01-01-PLAN, 01-02-PLAN]
  provides: [LoadCatalog-method, loadCatalog-wailsAPI, browse-size-column]
  affects: [frontend/src/services/wailsAPI.ts, frontend/src/components/tabs/BrowseCatalogsTab.tsx]
tech_stack:
  added: []
  patterns: [TDD red-green cycle, dual-format JSON parsing, Wails binding regeneration]
key_files:
  created: []
  modified:
    - internal/search/service.go
    - internal/search/service_test.go
    - app.go
    - frontend/wailsjs/go/main/App.d.ts
    - frontend/wailsjs/go/main/App.js
    - frontend/src/services/wailsAPI.ts
    - frontend/src/components/tabs/BrowseCatalogsTab.tsx
decisions:
  - "LoadCatalog tries array format first (v1 compat) then bare object (v2), matching the existing searchInCatalogFile dual-format pattern"
  - "Size column uses 1024-based units (B/KB/MB/GB) with toFixed(1) formatting; value 0 renders as '0 B'"
metrics:
  duration_minutes: 8
  completed: "2026-03-25"
  tasks_completed: 3
  files_modified: 7
---

# Phase 2 Plan 1: LoadCatalog + Browse Size Column Summary

**One-liner:** LoadCatalog Go method with TDD, wired through Wails to wailsAPI.loadCatalog, plus a human-readable Size column in the Browse tab.

## What Was Built

### Task 1: LoadCatalog method on search.Service (TDD)

Added `func (s *Service) LoadCatalog(filePath string) (*models.CatalogItem, error)` to `/Users/ken/dev/storcat/internal/search/service.go`.

The method:
- Reads the file from disk (returns error on read failure)
- Tries array-wrapped format first `[{...}]` for v1.0 bash script compatibility
- Falls back to bare object format `{...}` for v2.0.0
- Returns error on invalid JSON that matches neither format

Tests added to `/Users/ken/dev/storcat/internal/search/service_test.go`:
- `TestLoadCatalog` — bare object format (uses existing `writeTestCatalog` helper)
- `TestLoadCatalogArrayFormat` — v1 array-wrapped format
- `TestLoadCatalogNotFound` — nonexistent file returns error
- `TestLoadCatalogInvalidJSON` — malformed JSON returns error

All 7 tests pass (3 existing + 4 new).

### Task 2: App.LoadCatalog wrapper, Wails binding regeneration, wailsAPI wrapper

Added `func (a *App) LoadCatalog(filePath string) (*models.CatalogItem, error)` to `/Users/ken/dev/storcat/app.go` — delegates to `a.searchService.LoadCatalog(absPath)` after resolving absolute path.

Ran `wails generate module` to regenerate bindings. `/Users/ken/dev/storcat/frontend/wailsjs/go/main/App.d.ts` now exports:
```typescript
export function LoadCatalog(arg1:string):Promise<models.CatalogItem>;
```

Added `LoadCatalog` to the import block and `loadCatalog` wrapper to `/Users/ken/dev/storcat/frontend/src/services/wailsAPI.ts`:
```typescript
loadCatalog: async (filePath: string) => {
    try {
        const catalog = await LoadCatalog(filePath);
        return { success: true, catalog };
    } catch (error: any) {
        return { success: false, error: error.message || 'Unknown error' };
    }
},
```
This matches the Electron v1 `load-catalog` IPC return shape exactly.

### Task 3: Size column in Browse tab

Added size column to `/Users/ken/dev/storcat/frontend/src/components/tabs/BrowseCatalogsTab.tsx` between Catalog File and Modified columns:

```typescript
{
  key: 'size',
  title: 'Size',
  width: 120,
  minWidth: 80,
  render: (value: number) => {
    if (value === 0) return '0 B';
    const units = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(value) / Math.log(1024));
    return `${(value / Math.pow(1024, i)).toFixed(1)} ${units[i]}`;
  },
},
```

Updated testData objects to include `size` field (1536 and 524288 bytes).

Browse tab now has 4 columns: Title, Catalog File, Size, Modified.

## Commits

| Task | Commit | Description |
|------|--------|-------------|
| Task 1 | f6d664f7 | feat(02-01): add LoadCatalog method to search.Service with TDD |
| Task 2 | cacf79cd | feat(02-01): wire App.LoadCatalog, regenerate bindings, add wailsAPI wrapper |
| Task 3 | 37f4385f | feat(02-01): add Size column to Browse Catalogs tab |

## Verification

All verification checks passed:

1. `go test ./... -v` — all 7 tests pass including 4 new LoadCatalog tests
2. `go build ./...` — exits 0, no build errors
3. `grep "LoadCatalog" frontend/wailsjs/go/main/App.d.ts` — binding present
4. `grep "loadCatalog:" frontend/src/services/wailsAPI.ts` — wrapper present
5. `grep "key: 'size'" frontend/src/components/tabs/BrowseCatalogsTab.tsx` — column present
6. 4 `key:` entries in columns array confirmed (title, name, size, modified)

## Deviations from Plan

None — plan executed exactly as written.

## Known Stubs

None. All data is wired end-to-end. The browse size column reads from real `CatalogMetadata.Size` populated by `info.Size()` in `BrowseCatalogs` (implemented in Phase 1 Plan 2). The `loadCatalog` wrapper calls the real Go method.

## Self-Check: PASSED

Files exist:
- FOUND: /Users/ken/dev/storcat/internal/search/service.go
- FOUND: /Users/ken/dev/storcat/internal/search/service_test.go
- FOUND: /Users/ken/dev/storcat/app.go
- FOUND: /Users/ken/dev/storcat/frontend/wailsjs/go/main/App.d.ts
- FOUND: /Users/ken/dev/storcat/frontend/src/services/wailsAPI.ts
- FOUND: /Users/ken/dev/storcat/frontend/src/components/tabs/BrowseCatalogsTab.tsx

Commits verified:
- FOUND: f6d664f7
- FOUND: cacf79cd
- FOUND: 37f4385f
