---
phase: 23-rail-virtualized-tree
plan: 01
subsystem: workspace-tree
tags: [go, wails, react, tanstack-react-virtual, json, virtualization]

requires: []
provides:
  - "internal/fixture generators (WriteDCIMCatalog/WriteFlatCatalog/WriteDeepCatalog) every downstream 23-0x plan measures against"
  - "search.Service.LoadCatalogFlat — the single flat-array binding the rest of the phase's tree/rail/details work reads"
  - "AppContext's tree/expanded/selected/currentCatalogId reducer state and SELECT_CATALOG/TREE_LOADED/TREE_FAILED actions"
  - "useVisibleRows O(n) visible-index derivation, and TreePane's @tanstack/react-virtual wiring keyed on density rowHeight"
affects: [23-02, 23-03, 23-04, 23-05, 23-06]

actuals:
  tokens: 11463
  tasks: 3
  commits: 3

tech-stack:
  added: ["@tanstack/react-virtual@3.14.9"]
  patterns:
    - "Go flattens nested catalogs into a single render-ready array; TypeScript never re-walks the nested structure"
    - "TreeState discriminated union (idle/loading/ready/error) instead of separate loading/error/data fields"
    - "Load results carry the id they were issued for; the reducer discards a superseded load by comparing against state.currentCatalogId"

key-files:
  created:
    - internal/fixture/fixture.go
    - internal/fixture/fixture_test.go
    - scripts/gen-fixture-catalog/main.go
    - internal/search/flatten.go
    - internal/search/flatten_test.go
    - internal/search/flatten_bench_test.go
    - frontend/src/hooks/useVisibleRows.ts
  modified:
    - pkg/models/catalog.go
    - app.go
    - frontend/wailsjs/go/main/App.d.ts
    - frontend/wailsjs/go/main/App.js
    - frontend/wailsjs/go/models.ts
    - frontend/package.json
    - frontend/package-lock.json
    - frontend/src/services/wailsAPI.ts
    - frontend/src/contexts/AppContext.tsx
    - frontend/src/components/workspace/CatalogRail.tsx
    - frontend/src/components/workspace/TreePane.tsx

key-decisions:
  - "LoadCatalogFlat excludes the root from the flat array and numbers the root's direct children Depth 0/ParentIdx -1 (FPA-23-01-D) — the only reading that satisfies both an empty-catalog zero-length array and a single-top-level-file one-row render"
  - "useVisibleRows requires a node's parent to be BOTH visible AND expanded (not just visible) — the sketch in 23-RESEARCH.md Pattern 3 checks only parent-visibility and would make every node always visible"
  - "rowHeight is read from AppState.density (27/34px), never getComputedStyle, so the virtualizer updates on the same render a density toggle fires"
  - "@tanstack/react-virtual@3.14.9 installed per 23-CONTEXT.md's standing pre-approval; the package-legitimacy SUS flag was not re-raised as a checkpoint"

patterns-established:
  - "Tracer task pattern: build the thinnest complete path through every layer (fixture -> Go parse/flatten -> Wails bridge -> reducer -> virtualizer -> DOM) production-quality, verify it end-to-end via dev-browser, then let 23-02..23-06 expand each layer"

requirements-completed: [TREE-01, TREE-06, RAIL-03, COMPAT-01]

coverage:
  - id: D1
    description: "Fixture generator producing 40,000+-node synthetic catalogs on demand (DCIM/flat/deep shapes), never committed to the repo"
    verification:
      - kind: unit
        ref: "internal/fixture/fixture_test.go#TestWriteDCIMCatalog_DefaultShape,TestWriteFlatCatalog_ExactFileCount,TestWriteDeepCatalog_ReParses"
        status: pass
    human_judgment: false
  - id: D2
    description: "LoadCatalogFlat: DFS flatten reusing LoadCatalog's dual-format parse, Name/Path split, HasChildren, FileCount/TotalBytes, 512-level depth guard, not-a-catalog rejection"
    requirement: "COMPAT-01"
    verification:
      - kind: unit
        ref: "internal/search/flatten_test.go#TestLoadCatalogFlat_Structure,TestLoadCatalogFlat_DualFormat,TestLoadCatalogFlat_EmptyRoot,TestLoadCatalogFlat_NotACatalog,TestLoadCatalogFlat_DepthCap"
        status: pass
    human_judgment: false
  - id: D3
    description: "End-to-end tracer: selecting a rail catalog loads its real contents through one LoadCatalogFlat call and renders virtualized .ws-tree-row rows; switching catalogs swaps rows with no leftovers"
    requirement: "RAIL-03"
    verification:
      - kind: automated_ui
        ref: "dev-browser at :34115 — clicked catalog-a/catalog-b/catalog-small rail rows in one directory, screenshot t23-tree-a.png; row counts 2/2/1 matched each fixture's own top-level dir count"
        status: pass
    human_judgment: false
  - id: D4
    description: "TREE-01 performance gate: 42,550-node fixture measured, not assumed — Go benchmark plus browser time-to-first-row and mounted-row-count at top/middle/bottom scroll"
    requirement: "TREE-01"
    verification:
      - kind: unit
        ref: "internal/search/flatten_bench_test.go#BenchmarkLoadCatalogFlat40k -bench BenchmarkLoadCatalogFlat40k -benchtime=3x"
        status: pass
      - kind: automated_ui
        ref: "dev-browser at :34115 — 42,550-file flat fixture, row counts 33/43/33 at top/middle/bottom, all under 60"
        status: pass
    human_judgment: false

duration: 11min
completed: 2026-08-13
status: complete
---

# Phase 23 Plan 01: Fixture Generator + End-to-End Tracer + Performance Measurement Summary

**LoadCatalogFlat proven end-to-end at 42,550 nodes: 5.641 MB wire payload, 107.7ms Go-side flatten, 932.9ms browser time-to-first-row, virtualizer holds row count at 33-43 regardless of scroll position**

## Performance

- **Duration:** 11 min
- **Started:** 2026-08-13T19:46:37-05:00
- **Completed:** 2026-08-13T19:57:23-05:00
- **Tasks:** 3
- **Files modified:** 18

## Accomplishments
- `internal/fixture` generators (`WriteDCIMCatalog`/`WriteFlatCatalog`/`WriteDeepCatalog`) plus a committed CLI wrapper, producing 40,000+-node synthetic catalogs on demand without committing a blob
- `LoadCatalogFlat` (Go): DFS-flattens a catalog into a single render-ready array by calling the unmodified `LoadCatalog` (zero `json.Unmarshal` calls of its own — the dual-format v1/v2 parse has exactly one home), splitting `Name`/`Path`, computing `FileCount`/`TotalBytes` in the same walk, and guarding against catalogs nested past 512 levels
- The whole path proven end-to-end: a catalog on disk reaches the screen as real rows through the fixture generator, the flattener, the regenerated Wails bridge, the extended `AppContext` reducer, `useVisibleRows`, and `@tanstack/react-virtual` — verified live via dev-browser, not asserted
- The phase's core performance assumption (23-RESEARCH.md Assumption A3, the highest-risk unverified claim in the document) is now a recorded measurement: **107,664,306 ns/op (~107.7ms), 42,550 nodes, 5.641 MB marshalled, 932.9ms time-to-first-row in the browser, and mounted `.ws-tree-row` counts of 33 (top) / 43 (middle) / 33 (bottom)** — all three well under the 60-row gate at 40,000+ loaded nodes

## Measured Numbers (recorded, not assumed)

| Metric | Value |
|---|---|
| Benchmark ns/op (`BenchmarkLoadCatalogFlat40k`, `-benchtime=3x`) | 107,664,306 ns/op (~107.7ms) |
| Node count | 42,550 |
| Marshalled size | 5.641 MB |
| Time-to-first-row (browser, `.ws-tree-row` first paint after click) | 932.9ms |
| Mounted `.ws-tree-row` count — scroll top | 33 |
| Mounted `.ws-tree-row` count — scroll middle (`scrollHeight/2`) | 43 |
| Mounted `.ws-tree-row` count — scroll bottom (`scrollHeight`) | 33 |

Time-to-first-row (932.9ms) is under the 2-second threshold 23-RESEARCH.md flags for a follow-up — no chunking/pagination follow-up is raised. Measured against `WriteFlatCatalog(42550)` (nothing to expand, so all 42,550 nodes are in the visible slice) — the harder case than the DCIM shape used for the Go benchmark.

## Task Commits

Each task was committed atomically:

1. **Task 1: Wave 0 — the synthetic fixture generator** - `ace2a2cd` (feat)
2. **Task 2: End-to-end "select a catalog, see its rows" tracer** - `4b009a88` (feat)
3. **Task 3: Measure the tracer at 42,000 nodes** - `751c2c75` (test)

## Files Created/Modified

- `internal/fixture/fixture.go` - `WriteDCIMCatalog`/`WriteFlatCatalog`/`WriteDeepCatalog`, reproducing `traverseDirectory`'s `"./"`-prefixed display-path convention
- `internal/fixture/fixture_test.go` - node-count/re-parse/size assertions for all three shapes
- `scripts/gen-fixture-catalog/main.go` - CLI wrapper printing `path=/nodes=/bytes=`, defaults to `os.TempDir()`
- `pkg/models/catalog.go` - `FlatNode`/`FlatCatalog` structs, additive; `CatalogItem` untouched
- `internal/search/flatten.go` - `LoadCatalogFlat`
- `internal/search/flatten_test.go` - structure/dual-format/empty/not-a-catalog/depth-cap coverage
- `internal/search/flatten_bench_test.go` - `BenchmarkLoadCatalogFlat40k`
- `app.go` - `LoadCatalogFlat` binding
- `frontend/wailsjs/go/main/App.d.ts`, `App.js`, `frontend/wailsjs/go/models.ts` - regenerated via `wails generate module`, not hand-edited
- `frontend/package.json`, `frontend/package-lock.json` - `@tanstack/react-virtual@3.14.9`
- `frontend/src/services/wailsAPI.ts` - `loadCatalogFlat` wrapper, `{success, ...}` envelope
- `frontend/src/contexts/AppContext.tsx` - `catalogDir`, `catalogs`, `currentCatalogId`, `tree` (`TreeState` union), `expanded`, `selected`; `SET_CATALOG_DIR`, `SET_CATALOGS`, `SELECT_CATALOG` (atomic), `TREE_LOADED`/`TREE_FAILED` (id-guarded)
- `frontend/src/hooks/useVisibleRows.ts` - O(n) parent-visible-AND-expanded derivation
- `frontend/src/components/workspace/CatalogRail.tsx` - tracer-minimal fill: reads persisted directory, calls `browseCatalogs`, renders clickable `.ws-rail-row` per catalog
- `frontend/src/components/workspace/TreePane.tsx` - tracer-minimal fill: loads the selected catalog's flat tree, renders `.ws-tree-row` elements through `useVirtualizer`

## Decisions Made

- Excluded the catalog root from the flat array (FPA-23-01-D) so an empty catalog yields a zero-length slice and a single top-level file renders as one row at depth 1 — both required by the plan's `must_haves`, and only satisfiable together this way
- `useVisibleRows` checks parent visibility AND the parent's `expanded[path]` flag together, correcting 23-RESEARCH.md Pattern 3's sketch (which checks only visibility and would show every node regardless of expansion)
- `rowHeight` is derived from `state.density` in the reducer, never `getComputedStyle` — verified via `grep -c getComputedStyle` returning 0
- `@tanstack/react-virtual@3.14.9` installed without re-raising the package-legitimacy checkpoint, per 23-CONTEXT.md's standing pre-approval covering the SUS "too-new latest version" heuristic

## Deviations from Plan

None — plan executed exactly as written. Tasks, verification commands, and file list all matched the plan's `files_modified` and acceptance criteria.

## Issues Encountered

`dev-browser`'s native binary was missing on first invocation (`Native binary not found for darwin-arm64`) — its postinstall script had not run. Fixed by re-running `node scripts/postinstall.js` inside the global `dev-browser` package directory, then `dev-browser install` for the Playwright/Chromium dependency. Not a deviation from this plan's scope — a one-time local tooling gap, resolved before any browser verification was attempted.

## Known Stubs

Both intentional, both scoped to a named later plan in this same phase per the plan's own action text:

- `frontend/src/components/workspace/TreePane.tsx` — rows render only `node.name`. Carets, shape icons, right-aligned size, selection styling, expand/collapse, and the catalog header/breadcrumb are plan **23-03**'s work.
- `frontend/src/components/workspace/CatalogRail.tsx` — rows render only `catalog.title`. The full two-line row markup (`{filename} · {fileCount} files`), the filter input's wiring, the red status dot, and directory-chip interactivity are plan **23-04**'s work.

Neither blocks this plan's own goal (prove the load-to-render path end to end and measure it), and both are already named future plans in the phase's own `ROADMAP.md`/plan sequence, not orphaned debt — not appended to `.planning/WINDOWS.md`.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- `LoadCatalogFlat`, the extended `AppContext` reducer, `useVisibleRows`, and the virtualizer wiring are all in place for 23-02 (sidecar count cache + `BrowseCatalogs` parse-error field), 23-03 (tree row detail + header/breadcrumb), 23-04 (rail row detail + filter), 23-05 (details panel + unreadable-catalog panel), and 23-06 (reveal-in-file-manager) to build on directly.
- No blockers. The one open item is Pitfall N5 (Windows `explorer /select,` argv shape) — already flagged for 23-06, not this plan.

## Self-Check: PASSED

All 18 files claimed as created/modified confirmed present on disk; all 3 task commit hashes (`ace2a2cd`, `4b009a88`, `751c2c75`) confirmed in `git log`.
