---
phase: 23-rail-virtualized-tree
plan: 02
subsystem: rail-data-backend
tags: [go, sidecar-cache, concurrency, json-parsing, wails-bindings]

requires: ["internal/search/flatten.go (23-01) — LoadCatalogFlat's own FileCount/TotalBytes walk, extended here with an opportunistic cache fill"]
provides:
  - "config.CountsCache — mutex-guarded, atomically-written sidecar cache keyed on path+mtime+size"
  - "models.CatalogMetadata.FileCount/.TotalBytes/.ParseError — the three fields the rail row, red dot and status bar all read"
  - "search.Service.SetCountsCache — nil-safe cache wiring, CLI's NewService() untouched"
affects: [23-04, 23-05]

actuals:
  tokens: 21000
  tasks: 3
  commits: 3

tech-stack:
  added: []
  patterns:
    - "Sidecar cache modeled on config.Manager's directory-resolution and load/save shape, but with a sync.Mutex config.Manager deliberately lacks — the one place copying the nearest analog verbatim would have been a real bug"
    - "json.Valid() fast path before a full Unmarshal, so a healthy catalog directory pays one read + one linear scan, not a struct-allocating parse, on every rail load"
    - "Cache-miss-never-blocks: FileCount/TotalBytes are *int/*int64 so a miss renders as nil (not known yet), never a fabricated zero"

key-files:
  created:
    - internal/config/counts_cache.go
    - internal/config/counts_cache_test.go
  modified:
    - internal/config/config.go
    - internal/search/service.go
    - internal/search/service_test.go
    - internal/search/flatten.go
    - internal/search/flatten_test.go
    - pkg/models/catalog.go
    - app.go
    - frontend/wailsjs/go/models.ts

key-decisions:
  - "Cache JSON serializes the map[string]CountEntry directly (no wrapping {\"entries\": ...} struct) — one less layer than the RESEARCH.md sketch, same information"
  - "detectParseError attempts array-then-object Unmarshal to mirror LoadCatalog's own order even though json.Valid()==false guarantees both attempts fail identically (a syntax error is format-agnostic) — kept for code-structure fidelity to the locked pattern, not because it changes behavior"
  - "storcatConfigDir() extracted from config.NewManager as a pure, behavior-unchanged refactor so the counts cache resolves the exact same directory rather than growing a second copy that could drift"
  - "Exported config.NewCountsCacheAt(path) (not an unexported test-only constructor) because internal/search's tests need to build a cache pointed at a temp path from outside the config package"

patterns-established:
  - "A local JSON sidecar file written from more than one call site gets a sync.Mutex and temp-file+rename atomic writes as a baseline, not an optimization — config.Manager's unmutexed pattern is the wrong analog once a background/opportunistic fill exists"

requirements-completed: [RAIL-01, RAIL-04, STATE-02, SHELL-06]

coverage:
  - id: E7
    description: "Every load-path failure mode surfaces its real cause: file missing, permission denied, truncated JSON, malformed JSON distinguishable from each other"
    requirement: "STATE-02"
    verification:
      - kind: unit
        ref: "internal/search/service_test.go#TestBrowseCatalogs_ParseError_Truncated,TestBrowseCatalogs_ParseError_Malformed,TestBrowseCatalogs_ParseError_UnreadableFile; internal/search/flatten_test.go#TestLoadCatalogFlat_MissingFile,TestLoadCatalogFlat_PermissionDenied"
        status: pass
    human_judgment: false
  - id: RAIL-01-ordering
    description: "BrowseCatalogs returns catalogs in os.ReadDir's documented filename-sorted order, not incidental filesystem/creation order"
    verification:
      - kind: unit
        ref: "internal/search/service_test.go#TestBrowseCatalogs_ReturnsFilenameOrder"
        status: pass
    human_judgment: false
  - id: RAIL-01-concurrency
    description: "The sidecar cache survives concurrent load-mutate-save cycles under the race detector"
    verification:
      - kind: unit
        ref: "internal/config/counts_cache_test.go#TestCountsCache_ConcurrentPutGet_NoRace, run via `go test -race`"
        status: pass
    human_judgment: false
  - id: N3-allocation
    description: "detectParseError's json.Valid fast path allocates strictly less for a valid document than an invalid one"
    verification:
      - kind: unit
        ref: "internal/search/service_test.go#TestDetectParseError_AllocsValidLessThanInvalid, testing.AllocsPerRun"
        status: pass
    human_judgment: false

duration: 24min
completed: 2026-08-14
status: complete
---

# Phase 23 Plan 02: Sidecar Counts Cache + BrowseCatalogs Parse Status Summary

**A mutex-guarded, atomically-written sidecar cache (proven race-clean under `-race`) backs the rail's file-count/byte-total columns, and `json.Valid()`-gated parse-status detection gives every catalog a red-dot-worthy `parseError` at the cost of one read plus one linear scan per catalog**

## Performance

- **Duration:** 24 min
- **Started:** 2026-08-14T01:07:00Z (approx, from STATE.md session continuity)
- **Completed:** 2026-08-14
- **Tasks:** 3
- **Files modified:** 9 (2 created, 7 modified)

## Accomplishments

- `internal/config/counts_cache.go`: `CountsCache` with `CountEntry`, `CountsKey` (path+mtime+size concatenation), mutex-guarded `Load`/`Get`/`Put`, and atomic temp-file+rename persistence — the single deliberate divergence from `config.Manager` (which has no mutex at all), because this cache is written from a background/opportunistic fill while `BrowseCatalogs` may be reading it concurrently
- `internal/config/config.go`: `storcatConfigDir()` extracted from `NewManager` as a pure, behavior-unchanged refactor (regression-gated by the pre-existing `config_test.go` suite) so the cache and the config resolve the identical directory
- `pkg/models/catalog.go`: `CatalogMetadata` gains exactly three fields — `FileCount *int`, `TotalBytes *int64` (pointers, so a cache miss renders as "not known yet" rather than a fabricated `0`), and `ParseError string`
- `internal/search/service.go`: `detectParseError` — `json.Valid()` first, falling back to a real `Unmarshal` (mirroring `LoadCatalog`'s array-then-object attempt order) only for a broken catalog, formatting a `*json.SyntaxError` as `byte {offset}: {reason}`; `BrowseCatalogs` extended additively to populate `ParseError` and cache-backed counts; `SetCountsCache` is nil-safe so `cli/search.go`/`cli/show.go`'s `search.NewService()` call sites are untouched
- `internal/search/flatten.go`: `LoadCatalogFlat` opportunistically fills the counts cache with the `FileCount`/`TotalBytes` its own DFS walk already computed, keyed identically to `BrowseCatalogs`' lookup — a cache miss on the rail self-heals the next time that catalog is opened
- `app.go`: `NewApp` constructs the counts cache after the config manager and wires it via `SetCountsCache`, following the exact same startup-tolerance pattern as a failed `configManager` — a cache-construction failure leaves the app fully usable
- `frontend/wailsjs/go/models.ts` regenerated via `wails generate module` — `CatalogMetadata` now carries `fileCount?`, `totalBytes?` (nullable numbers) and `parseError`

## Task Commits

Each task was committed atomically:

1. **Task 1: The sidecar counts cache** — `11228d8e` (feat)
2. **Task 2: BrowseCatalogs parse status and cache-backed counts** — `a61e5cb1` (feat)
3. **Task 3: Wire the cache into the app and regenerate the bridge** — `12352a10` (feat)

## Files Created/Modified

- `internal/config/counts_cache.go` — `CountsCache`, `CountEntry`, `CountsKey`, `NewCountsCache`/`NewCountsCacheAt`, mutex-guarded `Load`/`Get`/`Put`, atomic write
- `internal/config/counts_cache_test.go` — hit/miss-on-mtime/miss-on-size, corrupt-file and missing-file degrade to miss without error, fresh-instance persistence round-trip, atomic-write leaves no leftover temp file, 60-goroutine race-detector-clean concurrency test
- `internal/config/config.go` — `storcatConfigDir()` extracted, `NewManager` calls it
- `internal/search/service.go` — `detectParseError`, `SetCountsCache`, `BrowseCatalogs` extended with parse-status + cache-backed counts
- `internal/search/service_test.go` — parse-error coverage (well-formed v1/v2, truncated, malformed, unreadable), filename ordering, counts-cache hit/miss, `AllocsPerRun` allocation proof
- `internal/search/flatten.go` — opportunistic cache fill after a successful flatten
- `internal/search/flatten_test.go` — missing-file vs. permission-denied distinguishable errors, opportunistic-fill-matches-the-walk's-own-counts test
- `pkg/models/catalog.go` — `CatalogMetadata.FileCount`/`.TotalBytes`/`.ParseError`
- `app.go` — counts cache construction + wiring in `NewApp`
- `frontend/wailsjs/go/models.ts` — regenerated, not hand-edited

## Decisions Made

- Cache file serializes `map[string]CountEntry` directly rather than wrapping it in an `{"entries": ...}` struct — one less layer than the research sketch, same information, still `json.MarshalIndent`-formatted to match `config.json`'s on-disk style
- `detectParseError` performs the array-then-object mirrored attempt even though `json.Valid()==false` guarantees both fail identically (a syntax error is format-agnostic at that point) — kept for structural fidelity to the locked "mirror `LoadCatalog`'s attempt order" pattern, not because it changes the reported message
- `config.NewCountsCacheAt(path)` is exported (not a package-private test helper) because `internal/search`'s tests, in a different package, need to construct a cache pointed at a temp directory without ever touching the real user config dir

## Deviations from Plan

None — plan executed exactly as written. All three tasks, their file lists, and their acceptance criteria matched the plan.

## Issues Encountered

None. `wails dev` was already running in the background (per the environment brief) and its file watcher regenerated `frontend/wailsjs/go/models.ts` automatically after the Go struct changes; running `wails generate module` explicitly in Task 3 confirmed the regenerated output was already correct and produced no further diff beyond the expected three-field addition.

## Known Stubs

None. This plan's own goal — three new `CatalogMetadata` fields, a concurrency-safe sidecar cache, and cheap parse-status detection — is fully wired end-to-end: `BrowseCatalogs` populates all three fields today, and the frontend's regenerated types carry them. Consuming them in the rail UI (red dot rendering, status bar totals, the unreadable-catalog panel) is explicitly plan 23-04/23-05's work, not a stub left behind here.

## User Setup Required

None — no external service configuration required.

## Next Phase Readiness

- `CatalogMetadata.FileCount`/`.TotalBytes`/`.ParseError` and the regenerated Wails bridge are ready for 23-04 (rail row detail + filter, including the red dot) and 23-05 (details panel + unreadable-catalog panel) to consume directly.
- The sidecar cache self-heals: any catalog opened via `LoadCatalogFlat` backfills its own cache entry, so a cold rail load's `nil` counts fill in on the very next `BrowseCatalogs` call after the user opens that catalog once.
- No blockers.

## Self-Check: PASSED

All 9 files claimed as created/modified confirmed present on disk (verified via `git show` on each commit); all 3 task commit hashes (`11228d8e`, `a61e5cb1`, `12352a10`) confirmed in `git log`.
