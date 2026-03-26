---
phase: 05-frontend-shim
plan: 01
subsystem: frontend-shim
tags: [typescript, wails, api-wrapper, discriminated-union]
dependency_graph:
  requires: []
  provides: [createCatalog-full-envelope, wailsAPI-shim-complete]
  affects: [CreateCatalogTab]
tech_stack:
  added: []
  patterns: [as-const-discriminated-union]
key_files:
  created: []
  modified:
    - frontend/src/services/wailsAPI.ts
decisions:
  - "createCatalog wrapper now uses 'as const' on both success and error branches for TypeScript discriminated union narrowing, matching the Phase 4 pattern"
  - "All six CreateCatalogResult fields spread into success envelope: jsonPath, htmlPath, fileCount, totalSize, copyJsonPath, copyHtmlPath"
metrics:
  duration: 40s
  completed: "2026-03-26"
  tasks_completed: 2
  files_changed: 1
requirements_satisfied:
  - API-01
  - API-02
  - API-03
---

# Phase 5 Plan 01: Frontend Shim — createCatalog Return Fix Summary

**One-liner:** Fixed `createCatalog` wrapper to capture and forward all six `CreateCatalogResult` fields (`jsonPath`, `htmlPath`, `fileCount`, `totalSize`, `copyJsonPath`, `copyHtmlPath`) with `as const` discriminated union narrowing.

## What Was Done

The `wailsAPI.createCatalog` wrapper was discarding the `CreateCatalog` binding's return value — calling it with `await` but not capturing the result, then returning only `{success: true}`. The frontend's `CreateCatalogTab` needs `jsonPath`, `htmlPath`, `fileCount`, `totalSize`, `copyJsonPath`, and `copyHtmlPath` to display post-creation metadata.

### Task 1: Confirm Wails bindings are current

Ran `wails generate module` to regenerate TypeScript bindings from Go structs. Confirmed bindings were already current — `git diff frontend/wailsjs/` produced no output. Verified:
- `App.d.ts` contains `CreateCatalog(arg1:string,arg2:string,arg3:string,arg4:string):Promise<models.CreateCatalogResult>`
- `models.ts` contains `export class CreateCatalogResult` with `jsonPath`, `htmlPath`, `fileCount`, `totalSize`, `copyJsonPath?`, `copyHtmlPath?`

No commit needed — bindings unchanged.

### Task 2: Fix createCatalog wrapper

Edited `frontend/src/services/wailsAPI.ts` to:
1. Capture the return value: `const result = await CreateCatalog(...)`
2. Spread all six fields into the success envelope
3. Add `success: true as const` and `success: false as const` for discriminated union narrowing

TypeScript (`npx tsc --noEmit`) compiles clean. `getCatalogHtmlPath` and `readHtmlFile` wrappers confirmed unchanged.

**Commit:** `0543bece` — `feat(05-01): fix createCatalog wrapper to return all CreateCatalogResult fields`

## Verification Results

1. `npx tsc --noEmit` — exits 0, no type errors
2. `grep -c "result.jsonPath" wailsAPI.ts` — returns 1
3. `grep -c "result.fileCount" wailsAPI.ts` — returns 1
4. `grep -c "as const" wailsAPI.ts` — returns 16 (existing + 2 new)
5. `getCatalogHtmlPath` and `readHtmlFile` wrappers confirmed unchanged

## Deviations from Plan

None — plan executed exactly as written. Task 1 found bindings already current (as expected per research), so no binding commit was needed.

## Known Stubs

None — `wailsAPI.createCatalog` now returns real data from the Go backend. No placeholder values remain in the shim layer.
