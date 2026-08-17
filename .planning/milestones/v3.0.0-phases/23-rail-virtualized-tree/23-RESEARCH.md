# Phase 23: Rail + Virtualized Tree - Research

**Researched:** 2026-08-13
**Domain:** Go/Wails bound-method design (tree flattening, OS process spawning, sidecar caching) + React virtualization over a flat array
**Confidence:** HIGH for all in-repo structural claims (every file below was read in full this session, several claims Go-verified by running code); MEDIUM for the two externally-sourced claims (`@tanstack/react-virtual` API shape, Windows `explorer /select,` argv quirk — no official doc fetch tool was available this session, only WebSearch)

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

**Tree Flattening & Virtualization**
- Flattening happens in **Go**, via a new `LoadCatalogFlat(path)` returning a flat slice of nodes carrying `name`, `type`, `size`, `depth`, parent index, and `hasChildren`. Milestone research is explicit: flatten in Go and leave `LoadCatalog` untouched, because `cli/show.go` depends on its current shape.
- Virtualization uses **`@tanstack/react-virtual`** — headless, ~4KB, and its fixed-size mode matches the locked `--rh` row-height contract exactly. Not hand-rolled (a solved problem), not `react-window`.
- Go returns the **full flat array once** per catalog; TypeScript derives the visible slice from an `expanded` Set. No IPC round-trip per expand — this is what makes "expand every directory" instant rather than 40,000 calls.
- Row height is **fixed, read from `--rh`** (27px Compact / 34px Comfortable). No dynamic measurement.

**Rail Data, Counts & Filtering**
- Per-catalog file count and total bytes come from a **sidecar cache keyed on path + mtime**, stored beside the config file. Catalog JSON on disk stays byte-identical.
- Filtering runs **in TypeScript over the already-loaded rail array**. Match is case-insensitive on both title and filename.
- The filter string lives in **local `useState` in the rail component, read through `useDeferredValue`** — it never enters `AppContext`.
- Parse failures are detected by **`BrowseCatalogs` gaining a `parseError string` field per catalog**.

**Backend Surface & Compatibility**
- This phase adds **exactly three backend surfaces**: `LoadCatalogFlat`, `RevealInFileManager`, and three new fields on `CatalogMetadata` (`fileCount`, `totalBytes`, `parseError`). The six existing CLI subcommands are untouched.
- `LoadCatalogFlat` **reuses `LoadCatalog`'s existing dual-format parse** (v1 array wrapper + v2 bare object) and then flattens.
- `RevealInFileManager(path)` is a **new Go binding with per-OS behavior**: `open -R` on macOS, `explorer /select,` on Windows, `xdg-open` on the parent directory for Linux.
- The 40,000-node fixture is **generated on demand** — a Go test helper plus a small committed generator script writing into a temp dir. Not committed as a JSON blob.

**Selection, State & Error Surfaces**
- Selection and expansion live in the **existing `AppContext` reducer**: `currentCatalogId`, `expanded: Record<string, boolean>`, `selected: string | null`.
- Switching catalogs fires **one atomic reducer action** that clears `expanded`, clears `selected`, and resets scroll to 0 (TREE-06).
- The unreadable-catalog state (STATE-02) renders **inline in the tree pane** — filename, byte offset, reason, and the raw parser error in a `--ch` code box.
- Status bar counts (SHELL-06) are **derived from the rail array**.

### Claude's Discretion
- Flat-node struct field names and the exact `CatalogMetadata` JSON keys.
- Sidecar cache file name, location beside the config, and its serialization format.
- Component decomposition within the tree pane (header, breadcrumb, rows).
- Breadcrumb interaction details beyond the specified collapse-to-root.
- Whether the rail's "＋ New" pill and directory chip stay inert this phase or get a minimal wiring.

### Deferred Ideas (OUT OF SCOPE)
- ⌘K search palette across catalogs — Phase 24.
- Create slide-over behind the "＋ New" pill — Phase 25.
- Real catalog-directory picker behind the rail's directory chip, and the Settings surface generally — Phase 26.
- Catalog rename/duplicate/delete and fsnotify watch keeping the rail current — Phase 27.
- Re-scan and diff — Phase 28.
- Frontend unit tests for the virtualizer (TEST-01) — deferred to a separate testing milestone.
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| SHELL-06 | Status bar reports catalog count, indexed file count, total bytes | §"Status bar" in Architecture Patterns — derived client-side from the rail array; no new binding |
| RAIL-01 | Rail lists every catalog with title, JSON size, filename, file count | `BrowseCatalogs` extension (§6) + existing `CatalogMetadata` fields already cover title/size/filename; `fileCount` is new |
| RAIL-02 | Filter without re-rendering the tree | `useDeferredValue` + local `useState`, confirmed against the actual (not-yet-extended) `AppContext.tsx` shape read this session |
| RAIL-03 | Select a catalog, loads tree, clears previous selection | One atomic reducer action — requires *adding* `currentCatalogId`/`expanded`/`selected` to `AppContext`, which do not exist yet (see Pitfall N1) |
| RAIL-04 | Red status dot on parse failure | `CatalogMetadata.ParseError` field, populated by `BrowseCatalogs` (§6) |
| RAIL-05 | Directory chip shows/changes catalog directory | Existing `SelectDirectory` binding, already present in `app.go`; no new backend work |
| TREE-01 | 40,000+-node smooth scroll, no freeze | `@tanstack/react-virtual` fixed-size mode (§2) + measured wire-size (§1) + dev-browser DOM-row-count validation (§ Validation Architecture) |
| TREE-02 | Expand/collapse; dir click toggles+selects, file click selects | Index-computation pattern over flat array + `expanded` Set (§3) |
| TREE-03 | Expand-all / collapse-to-root from breadcrumb | Same flat-array + Set pattern — O(n) Set population, no array rebuild (§3) |
| TREE-04 | Catalog header: title, chips, metadata line | Data already available from `CatalogMetadata` + `FlatCatalog` root node — no new binding |
| TREE-05 | Breadcrumb with accent ancestor segments | Pure frontend — per-segment `<span>`, derived from `selected` node's `ParentIdx` chain |
| TREE-06 | No scroll/expansion leak across catalog switch | One atomic reducer action + virtualizer keyed/reset on catalog id (§3, Pitfall N2) |
| TREE-07 | Details panel follows selection | Reads `AppContext`'s `selected` + the in-memory `FlatCatalog`/`CatalogMetadata` — no new binding |
| TREE-08 | Open HTML + reveal JSON from details panel | Existing `GetCatalogHtmlPath`/`OpenExternal`/`ReadHtmlFile` (Open) + new `RevealInFileManager` (§5) |
| STATE-01 | Empty-library state | Pure frontend, gated on `rail.length === 0` or no directory configured — no backend change |
| STATE-02 | Unreadable-catalog diagnostic panel | `BrowseCatalogs`'s new `ParseError` (byte offset + reason) sourced from `*json.SyntaxError` (§6) |
| COMPAT-01 | v1.x/v2.x catalogs open via the flat path | `LoadCatalogFlat` calls the *unmodified* `search.Service.LoadCatalog`, which already handles both formats (verified §1) |
</phase_requirements>

## Summary

This phase is almost entirely new construction: Phase 22 shipped five workspace components (`CatalogRail`, `TreePane`, `DetailsPanel`, `StatusBar`, `WorkspaceShell`) that are **unconditional-empty skeletons with zero props and zero state** — confirmed by reading every one of them this session. `AppContext.tsx` currently holds only `density`, `railSide`, and `detailOverlay`; it does **not** yet have `currentCatalogId`, `expanded`, or `selected` — CONTEXT.md's phrase "existing AppContext reducer" refers to the reducer *pattern*, not these specific fields, which this phase must add. Planners should not read CONTEXT.md's wording as "wire existing state" — it is "extend the reducer with new state," a materially different (if not large) task.

On the Go side, the milestone's own `ARCHITECTURE.md` already designed `LoadCatalogFlat` in detail; this research confirms the design against the actual current code (not the pre-migration draft) and finds one important correction: `CatalogItem.Name` is **not a basename** — `internal/catalog/service.go`'s `traverseDirectory` sets it to a full relative display path (`"./"` for root, `"./sub/dir/file.txt"` for descendants). The flattener must derive `Name` (basename, via `filepath.Base`) and `Path` (the full string, verbatim) as two separate fields — mirroring exactly what `cli/show.go` and `generateTreeStructure` already do when rendering. A synthetic 42,551-node fixture, generated and measured in this session with Go's actual `encoding/json`, marshals to **5.83 MB** (~144 bytes/node) — large enough that the phase's real performance gate is the Wails Go→JS bridge transfer + parse time for that call, not the virtualizer (which is proven, off-the-shelf, and O(viewport) by construction).

The sidecar count cache and the `parseError` detection on `BrowseCatalogs` are two *different* costs that the locked decisions correctly keep separate: counts are cache-backed and degrade silently on a miss (never block the rail); `parseError` needs at least one read+validate pass per catalog *every* `BrowseCatalogs` call, because a catalog can go from valid to broken between launches. The efficient shape is `json.Valid(data)` first (fast, no struct allocation) and only fall through to a real `Unmarshal`/dual-format attempt — to capture the byte offset and reason — on the rare invalid case.

**Primary recommendation:** Build `LoadCatalogFlat` as a thin wrapper around the *unmodified* `search.Service.LoadCatalog` (DFS + basename/path split + parent-index bookkeeping, ~40 lines), extend `AppContext` with the three new reducer fields CONTEXT.md specifies (they don't exist yet), wire `@tanstack/react-virtual` in fixed-size mode keyed on `--rh`, and validate TREE-01 empirically against the measured 5.83 MB / 42.5k-node fixture via dev-browser rather than trusting either the milestone research's optimism or a hand-wavy "should be fine."

## Architectural Responsibility Map

> Tier names adapted for a Wails desktop app: "Browser/Client" = React renderer, "API/Backend" = Go `App` bound methods + `internal/*` packages, "Database/Storage" = local JSON files (config dir), "CDN/Static" = not applicable to this app shape.

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Tree flattening (`LoadCatalogFlat`) | API/Backend (Go) | — | Locked decision; also the only place `ParentIdx` falls out for free during the existing recursive walk |
| Tree virtualization / windowing | Browser/Client | — | `@tanstack/react-virtual` is a rendering-layer concern; must never round-trip to Go per expand/collapse |
| Expand/collapse state (`expanded` Set) | Browser/Client | — | Lives in `AppContext`; derived-visible-slice computation must stay client-side and non-recursive |
| Rail filtering | Browser/Client | — | Locked: TS `.filter()` over already-loaded array, local component state, never `AppContext` |
| Per-catalog file count / total bytes | Database/Storage (sidecar cache file) | API/Backend (computes + persists it) | Cache lives beside `config.json`; Go reads/writes it, never the frontend |
| Parse-error detection (`parseError`) | API/Backend | — | Requires reading/validating the catalog JSON — cannot be done client-side without shipping the whole file first |
| Reveal-in-file-manager | API/Backend (Go spawns an OS process) | — | Only Go can `exec.Command`; this is new OS-integration surface, isolated from `internal/catalog`'s file-format concerns |
| Details panel content | Browser/Client | API/Backend (source data) | Panel renders from already-loaded `FlatCatalog`/`CatalogMetadata` in memory — no new binding per node click |
| Status bar counts | Browser/Client | — | Locked: summed from the rail array client-side, not its own binding |
| Selection/breadcrumb state | Browser/Client | — | `AppContext`; breadcrumb segments derived from `ParentIdx` chain, no backend round-trip |

## Standard Stack

### Core

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `@tanstack/react-virtual` | 3.14.9 (published 2026-07-28) `[VERIFIED: npm registry, but see Package Legitimacy Audit — package name itself is `[ASSUMED]` per provenance rule, discovered via CONTEXT.md/training knowledge]` | Headless windowing for the tree pane | Locked decision; peer deps `react ^16.8.0 \|\| ^17 \|\| ^18 \|\| ^19` confirmed via `npm view` — compatible with this repo's React 18.2/18.3 |
| Go stdlib `encoding/json` | Go 1.26.6 toolchain `[VERIFIED: go version, this session]` | `LoadCatalogFlat` reuses `search.Service.LoadCatalog`'s existing `json.Unmarshal` calls; `*json.SyntaxError.Offset` powers STATE-02's byte offset | Already the only JSON library in the codebase — no new dependency |
| Go stdlib `os/exec` | stdlib | `RevealInFileManager`'s per-OS process spawn | `runtime.BrowserOpenURL` (existing `OpenExternal`) doesn't cover "select this file in its parent folder"; `os/exec` with an argv slice (never a shell string) avoids injection entirely |

### Supporting

No new supporting libraries. The phase's "exactly three backend surfaces" constraint (CONTEXT.md) rules out adding an OS-trash library, a volumes library, or an fsnotify dependency — those belong to Phases 26–28.

### Alternatives Considered

| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| `@tanstack/react-virtual` | Hand-rolled `Math.floor(scrollTop/rowHeight)` windowing | Milestone `ARCHITECTURE.md` correctly notes fixed-height rows make this arithmetic-simple, but CONTEXT.md has already locked the library — do not re-litigate |
| `@tanstack/react-virtual` | `react-window` | Cited in milestone research as "stable but effectively unmaintained" `[CITED: pkgpulse.com aggregator, MEDIUM confidence, not re-verified this session]`; not the locked choice |
| `os/exec` argv slice | `exec.Command("cmd", "/c", "explorer /select," + path)` shell-string form | **Never do this** — see Security Domain; string-concatenating into a shell command reintroduces exactly the injection class argv-slice `exec.Command` avoids |

**Installation:**
```bash
cd frontend && npm install @tanstack/react-virtual@3.14.9
```

**Version verification:** `npm view @tanstack/react-virtual version` → `3.14.9`, published 2026-07-28T20:28:53.150Z `[VERIFIED: npm registry, this session]`. Peer deps confirmed compatible with installed React (18.3.1, deduped across the tree per `npm ls`).

## Package Legitimacy Audit

| Package | Registry | Age | Downloads | Source Repo | Verdict | Disposition |
|---------|----------|-----|-----------|-------------|---------|-------------|
| `@tanstack/react-virtual` | npm | Package itself is long-established (TanStack org); **latest version** published 2026-07-28 (~2 weeks before this research) | 21,143,075/week `[VERIFIED: npm registry, this session]` | `github.com/TanStack/virtual` `[VERIFIED: npm registry, this session]` | **SUS** (gate flags "too-new" — see note) | Approved, flagged |

**Note on the SUS verdict:** the `package-legitimacy check` seam flagged this `too-new` because the *latest published version* is recent, not because the package is new or suspicious — 21M weekly downloads and the official `TanStack/virtual` GitHub org are strong legitimacy signals a slopsquatted or hallucinated package would not have. No `postinstall` script is present `[VERIFIED: npm registry, this session]`. Per protocol this is still kept, not removed, and the planner **must** add a `checkpoint:human-verify` task before the `npm install` step even though this researcher's read is that the flag is a heuristic false-positive.

**Packages removed due to `[SLOP]` verdict:** none.
**Packages flagged as suspicious `[SUS]`:** `@tanstack/react-virtual` — planner inserts `checkpoint:human-verify` before install per the note above.

## Architecture Patterns

### System Architecture Diagram

```
┌─────────────────────────── React Renderer (Browser/Client tier) ───────────────────────────┐
│                                                                                                │
│  CatalogRail                     TreePane                          DetailsPanel               │
│  ┌──────────────┐   click row    ┌─────────────────────┐  selected  ┌───────────────────┐    │
│  │ rail[] (from │──dispatch────▶ │ FlatCatalog.Nodes[]  │───(via───▶ │ node/catalog meta  │    │
│  │ BrowseCatalogs)│  SELECT_     │  (in-memory, loaded   │  AppContext│  + 2 action buttons│    │
│  │ filtered      │  CATALOG      │  once per catalog)    │  .selected)│                     │    │
│  │ (local state, │               │        │               │            └───────────────────┘    │
│  │ useDeferredValue)│            │        ▼               │                                     │
│  └──────────────┘               │  computeVisible(        │                                     │
│         ▲                        │   nodes, expandedSet)  │                                     │
│         │                        │        │               │                                     │
│         │                        │        ▼               │                                     │
│         │                        │  @tanstack/react-virtual│                                     │
│         │                        │  useVirtualizer()       │                                     │
│         │                        │  (fixed rowHeight=--rh) │                                     │
│         │                        └─────────────────────────┘                                     │
│         │                                                                                          │
│  AppContext (extend this phase): currentCatalogId, expanded: Record<string,bool>, selected        │
└──────────┬───────────────────────────────────┬───────────────────────────────────────────────────┘
           │ BrowseCatalogs(dir)                │ LoadCatalogFlat(path)      │ RevealInFileManager(path)
           ▼                                     ▼                             ▼
┌──────────────────────────────────── Go `App` bound methods (API/Backend tier) ─────────────────┐
│                                                                                                    │
│  BrowseCatalogs                    LoadCatalogFlat (NEW)                RevealInFileManager (NEW) │
│  ┌────────────────────┐            ┌─────────────────────────┐         ┌───────────────────────┐ │
│  │ stat + html <title> │            │ search.Service.LoadCatalog│        │ exec.Command per-OS:  │ │
│  │ (existing, unchanged)│           │ (UNCHANGED — dual-format  │        │  darwin: open -R       │ │
│  │ + json.Valid() pass  │           │  v1-array/v2-object parse)│        │  windows: explorer /sel│ │
│  │ (NEW: parseError)    │           │        │                  │        │  linux: xdg-open(parent)│ │
│  │ + sidecar cache read │            │        ▼                  │        └───────────────────────┘ │
│  │ (NEW: fileCount/     │            │ DFS flatten: Name=Base(),│                                    │
│  │  totalBytes, cache-  │            │ Path=item.Name (verbatim)│                                    │
│  │  miss = omit, never  │            │ ParentIdx bookkeeping    │                                    │
│  │  block)              │            └─────────────────────────┘                                    │
│  └────────────────────┘                                                                              │
│           │                                                                                          │
│           ▼                                                                                          │
│  Sidecar cache (Database/Storage tier): os.UserConfigDir()/storcat/counts-cache.json                │
│  keyed on path+mtime+size → {FileCount, TotalBytes}                                                 │
└────────────────────────────────────────────────────────────────────────────────────────────────────┘
```

### Recommended Project Structure

```
pkg/models/
├── catalog.go              # add FlatNode, FlatCatalog structs (additive); extend CatalogMetadata
internal/search/
├── service.go               # add BrowseCatalogs' ParseError population; LoadCatalog UNTOUCHED
├── flatten.go                # NEW: LoadCatalogFlat + the DFS flattener (calls the unmodified LoadCatalog)
internal/config/
├── counts_cache.go           # NEW: sidecar cache Manager, mirrors config.Manager's load/mutate/save shape but WITH a mutex (see Pitfall N4)
internal/osutil/  (or inline in app.go — team preference, per milestone ARCHITECTURE.md)
├── reveal_darwin.go          # exec.Command("open", "-R", path)
├── reveal_windows.go         # exec.Command("explorer", "/select,"+path)  -- verify empirically, see Pitfall N5
├── reveal_linux.go           # exec.Command("xdg-open", filepath.Dir(path))
app.go                        # register LoadCatalogFlat, RevealInFileManager bindings; extend BrowseCatalogs call site (unchanged signature)
frontend/src/contexts/
├── AppContext.tsx             # ADD currentCatalogId, expanded, selected + SELECT_CATALOG (atomic), TOGGLE_EXPAND, SET_SELECTED actions
frontend/src/components/workspace/
├── CatalogRail.tsx             # fill: rows, filter (useDeferredValue, local state), directory chip wiring
├── TreePane.tsx                # fill: header, breadcrumb, virtualized rows, STATE-02 panel
├── DetailsPanel.tsx            # fill: catalog-level + node-level views, 2 footer buttons
├── StatusBar.tsx                # fill: live counts, derived from rail array
frontend/src/hooks/ (optional)
├── useVisibleRows.ts            # computeVisible(nodes, expandedSet) memoized derivation
```

### Pattern 1: `LoadCatalogFlat` — flatten without touching `LoadCatalog`

**What:** A new method on `search.Service` (co-located with `LoadCatalog` since it directly calls it) that parses once via the existing dual-format logic, then performs a single DFS producing a flat, render-ready array.

**Why `cli/show.go` constrains this:** `cli/show.go` calls `search.NewService().LoadCatalog(catalogPath)` directly `[VERIFIED: cli/show.go:69-70, this session — "svc := search.NewService(); root, err := svc.LoadCatalog(catalogPath)"]` and walks the **nested** `*models.CatalogItem` via its own recursive `printTree`, using `filepath.Base(item.Name)` for display `[VERIFIED: cli/show.go:107 — "name := filepath.Base(item.Name)"]`. Changing `LoadCatalog`'s return shape breaks this CLI command and violates COMPAT-03 (CLI subcommands unchanged). `LoadCatalogFlat` must be a *new* method that calls `LoadCatalog` internally, never a modification of it.

**Critical correction to the milestone's `ARCHITECTURE.md` draft:** that document's example `FlatNode` struct comments `Name string // display name (basename)`. Reading `internal/catalog/service.go`'s `traverseDirectory` directly shows this is wrong for the *existing* field it derives from:

```go
// internal/catalog/service.go:86-90, verified this session
displayPath := "./" + filepath.ToSlash(relPath)
if relPath == filepath.Base(basePath) {
    displayPath = "./"
}
// ... (line 101, file case) Name: displayPath,
// ... (line 156, directory case) Name: displayPath,
```

`CatalogItem.Name` is the **full relative display path** (`"./"` for root, `"./sub/dir/file.txt"` for descendants), not a basename. Every existing consumer that wants a basename computes it explicitly: `cli/show.go:107` (`filepath.Base(item.Name)`) and `internal/catalog/service.go:244` (`itemName := filepath.Base(item.Name)` inside `generateTreeStructure`). `LoadCatalogFlat` must do the same — split the single existing field into two `FlatNode` fields:

```go
// pkg/models/catalog.go — additive, does not touch CatalogItem
type FlatNode struct {
    Name        string `json:"name"`        // basename via filepath.Base(item.Name), e.g. "IMG_0001.JPG"
    Path        string `json:"path"`        // item.Name verbatim, e.g. "./Volume01/100CANON/IMG_0001.JPG"
    Type        string `json:"type"`        // "file" | "directory" -- copied verbatim from CatalogItem.Type
    Size        int64  `json:"size"`
    Depth       int    `json:"depth"`        // 0 = root
    ParentIdx   int    `json:"parentIdx"`    // -1 for root, else index into the same Nodes slice
    HasChildren bool   `json:"hasChildren"`  // true only for directories with len(Contents) > 0
}

type FlatCatalog struct {
    Nodes []FlatNode `json:"nodes"`
}
```

```go
// internal/search/flatten.go (new file)
func (s *Service) LoadCatalogFlat(filePath string) (*models.FlatCatalog, error) {
    root, err := s.LoadCatalog(filePath) // UNCHANGED call — dual-format parse reused verbatim
    if err != nil {
        return nil, err
    }
    var nodes []models.FlatNode
    var walk func(item *models.CatalogItem, depth, parentIdx int)
    walk = func(item *models.CatalogItem, depth, parentIdx int) {
        idx := len(nodes)
        nodes = append(nodes, models.FlatNode{
            Name:        filepath.Base(item.Name),
            Path:        item.Name,
            Type:        item.Type,
            Size:        item.Size,
            Depth:       depth,
            ParentIdx:   parentIdx,
            HasChildren: item.Type == "directory" && len(item.Contents) > 0,
        })
        for _, child := range item.Contents {
            walk(child, depth+1, idx)
        }
    }
    walk(root, 0, -1)
    return &models.FlatCatalog{Nodes: nodes}, nil
}
```

Node ordering falls directly out of `CatalogItem.Contents`'s on-disk array order — for catalogs this app created, that's already dirs-first-alphabetical (`internal/catalog/service.go:119-128`, unmodified by this phase); for v1 bash-script catalogs, whatever order is in the file is preserved as-is (no re-sort, matching the "reuse, don't duplicate" decision).

**Wire size — measured, not assumed:** a synthetic 42,551-node `FlatCatalog` (50 top-level dirs × 50 subdirs × 16 files, DCIM-shaped) marshaled with Go's `encoding/json` this session produces **6,114,397 bytes (5.83 MB)**, averaging 143.7 bytes/node `[VERIFIED: measured this session via `go run` against the exact struct above, see Sources]`. Deeper hierarchies or longer filenames will scale roughly linearly with average path-string length — a catalog with e.g. twice the average path depth could plausibly reach 10-12 MB. No official Wails v2 payload-size limit was found in this session's WebSearch pass `[CITED: github.com/wailsapp/wails, wails.io — general bound-method docs, no size-limit discussion found, MEDIUM/LOW confidence — absence of a documented limit is not proof none exists]`. **Recommend this be the phase's core empirical Nyquist check** (§ Validation Architecture) rather than trusted from either this research or the milestone's prior document.

### Pattern 2: `@tanstack/react-virtual` fixed-size rows, keyed on `--rh`

**What:** `useVirtualizer` with a constant `estimateSize` sourced from the CSS custom property `--rh` (27px Compact / 34px Comfortable), which is set via `root.style.setProperty('--rh', ...)` at density-change time — not a CSS class toggle `[VERIFIED: frontend/src/themeTokens.ts:117-118 — "Compact: { '--rh': '27px', ... }, Comfortable: { '--rh': '34px', ... }" and line 164 "root.style.setProperty(name, value);"]`.

```typescript
// TreePane.tsx — sketch, not a literal snippet from any fetched doc
const rowHeight = state.density === 'Compact' ? 27 : 34; // read from AppState.density, not re-derived from getComputedStyle
const parentRef = useRef<HTMLDivElement>(null);
const virtualizer = useVirtualizer({
  count: visibleRows.length,
  getScrollElement: () => parentRef.current,
  estimateSize: () => rowHeight,
  overscan: 10,
});
```

`[CITED: tanstack.com/virtual — general API shape (count/getScrollElement/estimateSize/scrollToIndex/getVirtualItems/getTotalSize) confirmed via WebSearch summary of tanstack.com/virtual/latest/docs/framework/react/react-virtual, MEDIUM confidence — not fetched via an authoritative-docs MCP tool this session]`. Because `state.density` is already a discriminated union (`'Compact' | 'Comfortable'`) in `AppState` `[VERIFIED: frontend/src/contexts/AppContext.tsx:6-7]`, deriving `rowHeight` from `state.density` directly (not from `getComputedStyle`) avoids a DOM read and guarantees the virtualizer's `estimateSize` updates synchronously on the same render the density reducer action fires — no separate remeasure/reset call needed. Verify this against real density-toggle behavior at review time (untested by any source this session — see Assumptions Log A2).

**React 18 gotcha:** no StrictMode-specific issue was found for `useVirtualizer` itself in this session's search (unlike Wails' `EventsOn`, flagged in the milestone's `PITFALLS.md` Pitfall 13, which does not apply here since this phase adds no new `EventsOn` listener). The one real gotcha is Pitfall 6/7 already documented in the milestone's `PITFALLS.md` (scroll/virtualizer state leaking across catalog switches, dynamic-measuring reintroduced by habit) — this research does not repeat those, they still apply verbatim to this phase.

**Reset scroll on catalog change (TREE-06):** key the virtualizer's scroll container (or call `virtualizer.scrollToOffset(0)`) inside the same effect that handles the atomic `SELECT_CATALOG` reducer action's side effects — do not rely on component remount alone, since `TreePane` itself is not remounted on catalog switch (it's the same mounted component receiving new `FlatCatalog` data).

### Pattern 3: Visible-slice computation from a flat array + `expanded` Set — O(n), no rebuild

**What:** `computeVisible(nodes: FlatNode[], expanded: Set<string>): FlatNode[]` — a single linear pass that includes a node only if every ancestor (walked via `ParentIdx`) is present in `expanded`, or more efficiently, a pass that tracks "is my direct parent currently visible AND expanded" using the fact that `nodes` is already in DFS pre-order (a node's parent always appears earlier in the array):

```typescript
function computeVisible(nodes: FlatNode[], expanded: Set<number>): FlatNode[] {
  const visible: FlatNode[] = [];
  const parentVisible = new Map<number, boolean>(); // nodeIdx -> is this node itself visible
  nodes.forEach((node, idx) => {
    const isVisible = node.parentIdx === -1 || (parentVisible.get(node.parentIdx) === true);
    parentVisible.set(idx, isVisible);
    if (isVisible) visible.push(node);
  });
  return visible;
}
```

This is O(n) per computation (not per expand — recomputed once via `useMemo` keyed on `[nodes, expanded]`), never re-walks `CatalogItem.Contents`, and never rebuilds `nodes` itself. This is the exact pattern the milestone `PITFALLS.md` Pitfall 5 warns must be avoided in its *naive* form (recursive re-walk of the nested structure) — the flat-array-plus-Set approach above only ever does a single linear scan.

**"Expand all" / "collapse to root" (TREE-03):** "Expand all" = `setExpanded(new Set(nodes.filter(n => n.hasChildren).map((_, i) => i)))` (or filter by index directly) — O(n) Set construction, no array rebuild, no recursion. "Collapse to root" = `setExpanded(new Set([0]))` (only the root index stays expanded) — O(1). Both are pure state updates; `computeVisible`'s `useMemo` re-runs once as a consequence, still O(n), not O(n²) or recursive.

**Row key stability (reinforces milestone `PITFALLS.md` Pitfall 4):** the flat array's own index is stable across expand/collapse (the *underlying* `nodes` array from `LoadCatalogFlat` never changes after load) — use that original index (or `node.path`) as the React key, never the position within the *visible* slice, which changes shape on every expand/collapse.

### Pattern 4: Sidecar count cache

**Location:** beside `config.json`. `internal/config.Manager.NewManager()` resolves its directory via `os.UserConfigDir()` with a home-dir fallback, then creates `storcatConfigDir := filepath.Join(configDir, "storcat")` and writes `config.json` there `[VERIFIED: internal/config/config.go:41-58 — "storcatConfigDir := filepath.Join(configDir, \"storcat\")" / "configPath := filepath.Join(storcatConfigDir, \"config.json\")"]`. The sidecar cache should live at `filepath.Join(storcatConfigDir, "counts-cache.json")` — same directory, sibling file, reusing the directory-creation the config manager already performs (no new `MkdirAll` call needed if the cache manager is constructed after or alongside the config manager).

**Schema:** `map[string]CountEntry` keyed on `path + "|" + mtimeRFC3339 + "|" + sizeBytes` (string concatenation, not a hash — cheaper to construct, still collision-safe for this cardinality of tens of catalogs):

```go
type CountEntry struct {
    FileCount  int   `json:"fileCount"`
    TotalBytes int64 `json:"totalBytes"`
}
type CountsCache struct {
    Entries map[string]CountEntry `json:"entries"`
}
```

**Concurrent-access safety — do NOT copy `config.Manager`'s pattern verbatim:** `internal/config/config.go` has **no mutex anywhere** — no `sync` import, no lock field `[VERIFIED: internal/config/config.go:1-8, full import block read this session — "encoding/json", "os", "path/filepath" only]`. That's tolerable for `config.Manager` today because its `Set*` methods are only ever called from discrete, user-triggered UI actions (theme click, window resize) that in practice serialize naturally. The sidecar cache does **not** have that guarantee: the milestone `ARCHITECTURE.md`'s own design (§5, "Background fill") proposes a goroutine that fills cache entries for multiple catalogs after `BrowseCatalogs` returns, running concurrently with the opportunistic fill that happens whenever `LoadCatalogFlat` is called for a catalog the user clicks mid-background-fill. Wails also does not guarantee bound methods are invoked serially from the frontend — two rapid catalog clicks can dispatch two concurrent Go calls. **Guard the cache manager's load-mutate-save cycle with a `sync.Mutex`**, unlike the config manager it's modeled after.

**Cache-miss behavior (must not block the rail):** `BrowseCatalogs` checks the cache for each catalog's key; on a hit, populates `FileCount`/`TotalBytes`; on a miss (new catalog, changed mtime, cache file absent/corrupt), **omit** — never trigger a synchronous full parse inline in the `BrowseCatalogs` call. The count backfills the next time that specific catalog is opened (`LoadCatalogFlat` already walks every node — summing count/bytes during that same DFS is free) or via an optional background-fill goroutine. Corrupt/unreadable cache file degrades to "every entry is a miss," never an error surfaced to the rail.

### Pattern 5: `RevealInFileManager` — per-OS, argv-only, never a shell string

```go
// internal/osutil/reveal_darwin.go
func RevealInFileManager(path string) error {
    absPath, err := filepath.Abs(path)
    if err != nil {
        return fmt.Errorf("resolve path: %w", err)
    }
    if _, err := os.Stat(absPath); err != nil {
        return fmt.Errorf("path not accessible: %w", err)
    }
    return exec.Command("open", "-R", absPath).Run()
}
```

```go
// internal/osutil/reveal_windows.go
func RevealInFileManager(path string) error {
    absPath, err := filepath.Abs(path)
    if err != nil {
        return fmt.Errorf("resolve path: %w", err)
    }
    if _, err := os.Stat(absPath); err != nil {
        return fmt.Errorf("path not accessible: %w", err)
    }
    // NOTE: Windows Explorer's /select switch does not follow standard argv
    // conventions. Community precedent (VS Code, various Go CLI tools) commonly
    // passes "/select,"+path as a SINGLE argv element (comma directly attached,
    // no space). This is [ASSUMED] -- verify empirically on an actual Windows
    // build; see Pitfall N5.
    return exec.Command("explorer", "/select,"+absPath).Run()
}
```

```go
// internal/osutil/reveal_linux.go
func RevealInFileManager(path string) error {
    absPath, err := filepath.Abs(path)
    if err != nil {
        return fmt.Errorf("resolve path: %w", err)
    }
    if _, err := os.Stat(absPath); err != nil {
        return fmt.Errorf("path not accessible: %w", err)
    }
    // No universal "select this file" mechanism on Linux (nautilus/dolphin/caja
    // each need their own --select flag; xdg-open has none). Per CONTEXT.md's
    // locked decision, open the PARENT directory via xdg-open rather than
    // attempting file-manager detection.
    return exec.Command("xdg-open", filepath.Dir(absPath)).Run()
}
```

`[CITED: WebSearch summary of general Go exec.Command argv-safety guidance + pypi.org/project/show-in-file-manager docs describing the "no universal select mechanism" problem on Linux, MEDIUM confidence]`. `[ASSUMED: the exact `"/select,"+path` single-argv-element form for Windows — training-knowledge pattern, not confirmed against an authoritative Microsoft or Wails source this session]`.

### Pattern 6: `BrowseCatalogs` parse-error detection — fast path first

```go
// internal/search/service.go, inside the existing BrowseCatalogs loop, additive
data, err := os.ReadFile(filePath) // already reading this file's bytes is new cost -- see Pitfall N6
var parseErr string
if err != nil {
    parseErr = err.Error()
} else if !json.Valid(data) {
    // Only now pay for a real Unmarshal attempt, to extract offset+reason.
    // Mirrors LoadCatalog's own dual-format attempt so "Parser" in STATE-02
    // can report which format was tried.
    var arr []*models.CatalogItem
    if uerr := json.Unmarshal(data, &arr); uerr != nil {
        if syn, ok := uerr.(*json.SyntaxError); ok {
            parseErr = fmt.Sprintf("byte %d: %s", syn.Offset, syn.Error())
        } else {
            parseErr = uerr.Error()
        }
    }
}
```

`json.Valid(data)` scans the byte stream without allocating the target struct tree, so the common (valid) case pays roughly the cost of one `os.ReadFile` + one linear scan — the expensive full-structure `Unmarshal` only runs on the rare broken-catalog path `[CITED: WebSearch summary of Go `encoding/json` `SyntaxError.Offset` semantics — "the error occurred after reading Offset bytes," MEDIUM confidence]`. This still means `BrowseCatalogs` now reads every catalog's full file content on every call (it previously only called `entry.Info()` for size, never opened the JSON file) — flag this cost explicitly for large catalog directories (see Pitfall N6).

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| List windowing at 40k+ rows | A custom scroll-math windowing calculator | `@tanstack/react-virtual` (locked) | Handles overscan, resize (`ResizeObserver`), and the visible-range math correctly; hand-rolling risks the exact off-by-one/flicker bugs the milestone `PITFALLS.md` catalogs (Pitfalls 4, 6, 7) |
| OS "reveal in file manager" cross-platform abstraction | A generic cross-platform reveal library | Three tiny per-OS `exec.Command` calls (Pattern 5) | CONTEXT.md's "exactly three backend surfaces" constraint rules out adding a dependency for something this small; the milestone's own `internal/trash` precedent (a genuinely complex cross-platform OS operation) is NOT the same complexity class as "run `open -R`" |
| JSON syntax-error line/column reporting | A custom line-scanner over the raw bytes | `*json.SyntaxError.Offset` (byte position) directly, as STATE-02 already specifies "byte offset" not "line:column" | The UI-SPEC's own copy contract asks for `Failed at: byte {offset}` — matches stdlib's native output exactly, no extra parsing needed |

**Key insight:** this phase's "don't hand-roll" list is short because CONTEXT.md already pre-empted the two biggest temptations (a custom windowing calculator, a custom flatten-on-every-render) with locked decisions. The remaining hand-roll risk is entirely in the *new* Go surfaces (reveal, sidecar cache) where the natural shortcut is a shell-string `exec.Command` or a config.Manager-style unmutexed cache — both are covered as pitfalls below, not "don't hand-roll" library gaps.

## Common Pitfalls

> Builds on `.planning/research/PITFALLS.md` Pitfalls 4-7, 27 (React key stability, "expand all" freeze, scroll leak, fixed-height virtualizer, sidecar cache schema leak) — those are not repeated here; they apply verbatim to this phase's Phase 2 work. The following are net-new findings from reading the actual current code this session.

### Pitfall N1: CONTEXT.md's "existing AppContext reducer" state does not exist yet

**What goes wrong:** A plan or executor reads CONTEXT.md's "Selection and expansion live in the existing `AppContext` reducer: `currentCatalogId`, `expanded`, `selected`" and assumes these fields are already present (since the reducer *pattern* is indeed established from Phase 22) — then either fails when they're missing, or silently invents a *different* shape than what TREE-06/RAIL-03's atomic-action requirement calls for.

**Why it happens:** `AppContext.tsx` today only has `density: Density`, `railSide: RailSide`, `detailOverlay: boolean` `[VERIFIED: frontend/src/contexts/AppContext.tsx:5-9, full interface read this session]` — confirmed also by Phase 22's own summary: "AppContext.tsx: pruned reducer (16 dead tab-era fields removed) plus density, railSide, detailOverlay state" `[VERIFIED: .planning/phases/22-shell-token-layer/22-06-SUMMARY.md, "provides" list, read this session]`. CONTEXT.md's wording describes the *target* state shape from the design handoff, not the current file.

**How to avoid:** The plan must include an explicit task to extend `AppState`/`AppAction`/`appReducer` with `currentCatalogId: string | null`, `expanded: Record<string, boolean>`, `selected: string | null`, and a single `SELECT_CATALOG` action type that atomically sets `currentCatalogId`, clears `expanded`/`selected` — this is real, first-time work, not "wire existing state."

**Warning signs:** A plan task that says "use `state.currentCatalogId`" without a preceding task that adds it to the reducer.

**Phase to address:** This phase (23), first task in the frontend wave.

### Pitfall N2: `TreePane`/`CatalogRail`/`DetailsPanel`/`StatusBar` currently take zero props

**What goes wrong:** Building data-wiring on the assumption that these components already accept e.g. a `catalogs` or `flatCatalog` prop.

**Why it happens:** All four are literal `function ComponentName() { return (...) }` with no destructured parameter at all `[VERIFIED: frontend/src/components/workspace/CatalogRail.tsx:1, TreePane.tsx:1, StatusBar.tsx:1 — no props parameter in any; DetailsPanel.tsx:5 has `{ variant = 'pane' }: DetailsPanelProps` only]`, and `WorkspaceShell.tsx` renders `<CatalogRail />` / `<TreePane />` with no props passed `[VERIFIED: frontend/src/components/workspace/WorkspaceShell.tsx:60-61]`.

**How to avoid:** Either read directly from `useAppContext()` inside each component (simplest, matches the existing `WorkspaceShell` pattern of calling `useAppContext()` itself), or add explicit prop interfaces — the plan should pick one convention and state it, since Phase 22 established no precedent either way for these specific components (only `DetailsPanelProps` exists, and it's empty on purpose, reserved for this phase).

**Phase to address:** This phase (23).

### Pitfall N3: `BrowseCatalogs` reading full file content on every call is new, unbudgeted cost

**What goes wrong:** The current `BrowseCatalogs` never opens a catalog's `.json` file — it only calls `entry.Info()` (a stat, from the directory listing) for size, and reads the paired `.html` file (small) for the title `[VERIFIED: internal/search/service.go:169-198, full `BrowseCatalogs` loop read this session — no `os.ReadFile` on the `.json` path anywhere]`. Adding `parseError` detection (Pattern 6) means `BrowseCatalogs` must now `os.ReadFile` every catalog's full JSON content, every time the rail loads — for a directory with several 5+ MB catalogs, this is a meaningfully larger I/O and CPU cost than today's version, and it runs on every app launch and every directory change, not just once.

**How to avoid:** Use `json.Valid()` (not a full `Unmarshal`) for the common case (Pattern 6) to at least avoid struct-allocation cost; if this proves too slow in practice for large catalog directories, the fallback is to cache "last known good" parse-validity alongside the sidecar count cache (keyed the same way, on path+mtime) so an unchanged catalog isn't re-validated on every launch — this optimization is not required by CONTEXT.md's locked decisions but is a reasonable escape hatch if TREE-01/RAIL-01 performance testing surfaces it as a bottleneck.

**Phase to address:** This phase (23) — flag for the planner's task-level performance budget; the fixture-based validation (§ Validation Architecture) should include a rail-load timing check with several large catalogs present, not just the one 40k-node tree fixture.

### Pitfall N4: Sidecar cache copied verbatim from `config.Manager`'s (unmutexed) pattern

Covered in full under Pattern 4 above — restated here as a pitfall because it's the single highest-risk "looks done but isn't" item in this phase's new Go surface: a cache manager built by literally copying `config.Manager`'s code will compile, pass a single-threaded test, and then corrupt or lose entries under the concurrent background-fill + opportunistic-fill pattern the milestone's own `ARCHITECTURE.md` recommends.

**Phase to address:** This phase (23).

### Pitfall N5: Windows `explorer /select,` argv shape is unverified

The exact argv shape for Windows Explorer's "select this file" invocation was not confirmed against an authoritative source this session (WebSearch results were generic Go `exec.Command` guidance, not Explorer-specific). This project's own development happens on macOS (`CLAUDE.md`: "macOS recommended... build scripts assume macOS"), and the milestone `PITFALLS.md` already flags (Pitfall 26) that Windows-specific behavior tends to go untested until an actual Windows build. **Do not consider TREE-08's Windows behavior verified by code review alone — an actual Windows build/VM test is required**, consistent with that existing pitfall's own recommendation.

**Phase to address:** This phase (23) for the code; explicit Windows-build verification should be flagged as a `checkpoint:human-verify` (or deferred CI-artifact check) rather than asserted as done from a macOS-only review.

### Pitfall N6: `LoadCatalogFlat`'s wire size is measured for one synthetic shape only

The 5.83 MB figure (Pattern 1) is real but specific to a DCIM-like fixture with fairly short, uniform names. A catalog with much longer average path strings (deeply nested project directories, long descriptive filenames) will be proportionally larger. Treat TREE-01's "smooth, no freeze" gate as needing to be re-validated against whatever shape the phase's actual generated 40k-node fixture ends up using — do not assume the number in this document transfers exactly.

**Phase to address:** This phase (23) — the fixture generator (locked decision) should log its own node count and marshaled byte size as part of its output, so this number is re-derived, not assumed, at execution time.

## Code Examples

See Architecture Patterns above — every code example in this document (`LoadCatalogFlat`, `FlatNode`, `computeVisible`, the three `RevealInFileManager` variants, the `BrowseCatalogs` parse-error snippet, the `useVirtualizer` sketch) is a code example produced for this research, cross-referenced against the actual repo files read this session. None are copy-pasted from an official docs page verbatim (no authoritative-docs MCP tool was available) — treat them as design sketches to adapt during planning, not drop-in final code.

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| n/a — this is new construction, not a migration off a prior tree-rendering approach | — | — | — |

Not applicable. This phase builds new capability (`LoadCatalogFlat`, virtualized tree) rather than replacing an existing outdated pattern; the only "old approach" in this codebase is the deleted `ModernTable`/`CatalogModal` Electron-era components, already fully removed per the milestone `ARCHITECTURE.md`'s "New vs. Modified vs. Deleted Inventory," which is not repeated here.

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | Windows `explorer /select,"+path"` single-argv-element form is correct | Pattern 5 | Reveal-in-Explorer silently opens the wrong window or fails on Windows; low severity (TREE-08 has a working macOS/Linux fallback, and this is empirically testable in minutes on a Windows build) |
| A2 | `state.density`-derived `rowHeight` updates the virtualizer's `estimateSize` synchronously with no separate remeasure call needed | Pattern 2 | If wrong, a density toggle could show stale row heights for one frame or require an explicit `virtualizer.measure()` call the plan didn't anticipate; low severity, easy to catch visually |
| A3 | No Wails v2 bound-method payload size limit exists that would reject a ~6-12 MB `LoadCatalogFlat` response | Pattern 1 | If a real limit exists below that range, `LoadCatalogFlat` would need chunking/pagination — a materially larger redesign; **this is the highest-risk assumption in this document** and should be the first thing validated empirically once any Go code exists, ideally before frontend virtualization work begins |
| A4 | 40k-node fixture shape (50×50×16, DCIM-like) is a reasonable target for the generator | § Validation Architecture | If real catalogs skew toward much deeper nesting or much longer paths, the measured 5.83 MB figure understates the real worst case — low severity since the fixture is a test tool, not shipped code, and can be adjusted |
| A5 | `@tanstack/react-virtual`'s general API shape (count/getScrollElement/estimateSize/scrollToIndex) as summarized by WebSearch matches the installed 3.14.9 release exactly | Pattern 2, Standard Stack | If the API has moved since the summarized docs, the sketch code needs adjustment during implementation — low severity, this library's core API has been stable across the 3.x line |

## Open Questions

1. **Does a ~6-12 MB single `LoadCatalogFlat` call actually feel instant over the Wails bridge, or does it need chunking/streaming?**
   - What we know: no documented Wails v2 payload limit found; JSON marshal/unmarshal of a few MB is tens of milliseconds in both Go and V8 per the milestone `ARCHITECTURE.md`'s own (unverified this session) claim.
   - What's unclear: actual end-to-end latency (Go marshal → Wails bridge transfer → JS parse → first virtualized paint) for the phase's real 40k+-node fixture, on a representative machine.
   - Recommendation: make this the first empirical checkpoint once `LoadCatalogFlat` and the fixture generator both exist — time the full round trip via dev-browser before investing further in frontend polish. If it's meaningfully slow (multi-second), the escape hatch is NOT re-architecting mid-phase but flagging it for a follow-up (streaming/paginated load) rather than silently accepting a bad TREE-01 experience.

2. **Should `parseError` detection reuse a "last known good" cache to avoid re-reading every catalog's full content on every `BrowseCatalogs` call (Pitfall N3)?**
   - What we know: CONTEXT.md locks the `parseError`-on-`BrowseCatalogs` approach and explicitly frames it as "costs one pass — not a second validation binding."
   - What's unclear: whether "one pass" was evaluated against catalog directories containing several large (multi-MB) catalogs, or only against the common case of a handful of modest ones.
   - Recommendation: implement the locked approach as specified first; treat the cache-based optimization as a fallback only if the phase's own performance testing surfaces a real problem, not a preemptive addition.

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| Go toolchain | `LoadCatalogFlat`, `RevealInFileManager`, sidecar cache, `go test ./...` | ✓ | go1.26.6 darwin/arm64 `[VERIFIED, this session]` | — |
| `go test ./...` (existing suite) | Regression safety for `internal/search`, `internal/catalog`, `internal/config`, `cli/*` | ✓ | All 6 packages currently pass (`ok`) `[VERIFIED, this session — ran the full suite]` | — |
| Node.js | Frontend build/dev | ✓ | v24.14.1 `[VERIFIED, this session]` | — |
| npm | Package install (`@tanstack/react-virtual`) | ✓ | 11.18.0 `[VERIFIED, this session]` | — |
| TypeScript | Frontend typecheck (`npm run build` = `tsc && vite build`) | ✓ | 4.9.5 installed (package.json declares `^4.6.4`) `[VERIFIED, this session]` | — |
| Wails CLI | Rebinding generated JS wrappers after adding `LoadCatalogFlat`/`RevealInFileManager` to `app.go` | ✓ | v2.10.2, matches `go.mod`'s `wailsapp/wails/v2 v2.10.2` `[VERIFIED, this session]` | — |
| A real Windows machine/VM | Verifying `RevealInFileManager`'s `explorer /select,` argv shape (Pitfall N5, Assumption A1) | ✗ (this research session ran on macOS) | — | Code-review-only verification is insufficient per Pitfall N5; flag as `checkpoint:human-verify` or defer to a CI Windows artifact test |
| Vite dev server + dev-browser | Empirical TREE-01 validation (DOM row count, scroll behavior) | ✓ (per task brief — "a Vite dev server + dev-browser workflow is available and was used successfully in Phase 22") | — | — |

**Missing dependencies with no fallback:** none blocking Go/TS development; only Windows-specific runtime verification (Pitfall N5) has no local fallback and must be flagged for later verification.

## Validation Architecture

### Test Framework

| Property | Value |
|----------|-------|
| Framework | Go stdlib `testing`, table-driven, `t.TempDir()`-based fixtures — the established pattern in `internal/search/service_test.go` `[VERIFIED: internal/search/service_test.go:10-30, "writeTestCatalog" helper read this session]` |
| Config file | none — `go test` needs no config |
| Quick run command | `go test ./internal/search/... ./internal/config/... -run TestLoadCatalogFlat` (or the specific new test names once written) |
| Full suite command | `go test ./...` — confirmed passing on all 6 packages before this phase's changes `[VERIFIED, this session]` |
| Frontend framework | **None** — TEST-01 (Vitest + Testing Library for the virtualizer) is explicitly deferred (v2 requirement, `.planning/REQUIREMENTS.md` line 130). No frontend automated test framework exists or should be added this phase. |

### Phase Requirements → Test Map

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| COMPAT-01 | `LoadCatalogFlat` correctly flattens both v1 array-wrapped and v2 bare-object catalogs | unit (Go) | `go test ./internal/search/... -run TestLoadCatalogFlat_DualFormat` | ❌ Wave 0 |
| — (flatten correctness, underpins TREE-02/03/06) | `ParentIdx`/`Depth`/`HasChildren` correctness on a small known nested fixture | unit (Go) | `go test ./internal/search/... -run TestLoadCatalogFlat_Structure` | ❌ Wave 0 |
| TREE-01 | `LoadCatalogFlat` wall-clock time on the 40k+-node fixture (Go-side half of the performance gate) | benchmark (Go) | `go test ./internal/search/... -bench BenchmarkLoadCatalogFlat40k -benchtime=3x` | ❌ Wave 0 |
| TREE-01 | Virtualizer renders a bounded DOM row count regardless of total node count (browser-side half of the performance gate) | browser (dev-browser) | Load app against the 40k fixture via dev-browser, count `.ws-tree-row` DOM elements, assert count stays proportional to viewport (e.g. < 60) while `FlatCatalog.Nodes.length` > 40,000 | ❌ Wave 0 (no such fixture-loading flow exists yet in the dev harness) |
| TREE-01 | End-to-end round trip (click catalog → first rendered row) does not visibly freeze | browser (dev-browser) + manual eyeball | dev-browser scroll-and-screenshot sequence against the fixture catalog; genuinely needs a human/AI-visual judgment call on "smooth," not just a DOM assertion | Neither — this is the one criterion in this phase that is *not* fully automatable; the DOM-row-count check above is the automatable proxy for it |
| RAIL-04, STATE-02 | Broken catalog produces `parseError` with correct byte offset and reason | unit (Go) | `go test ./internal/search/... -run TestBrowseCatalogs_ParseError` | ❌ Wave 0 |
| RAIL-01 | Sidecar cache hit/miss/corrupt-file degradation | unit (Go) | `go test ./internal/config/... -run TestCountsCache` | ❌ Wave 0 |
| RAIL-01 | Sidecar cache concurrent read/write safety (Pitfall N4) | unit (Go, `-race`) | `go test ./internal/config/... -run TestCountsCache_Concurrent -race` | ❌ Wave 0 |
| TREE-08 | `RevealInFileManager` builds the correct argv per OS (build-tag-gated, testable per-platform only on that platform) | unit (Go, build-tag scoped) | `go test ./internal/osutil/... -run TestRevealInFileManager` (macOS/Linux runnable in this session's environment; Windows requires a Windows runner) | ❌ Wave 0 |
| RAIL-02, RAIL-03, TREE-02, TREE-05, TREE-06, TREE-07 | All frontend interaction behavior | browser (dev-browser) only, manual/AI-visual | dev-browser click/scroll/screenshot sequences against the running Vite dev server — no automated assertion framework exists for these (TEST-01 deferred) | n/a — no test *file* concept applies; these are dev-browser session steps, not committed test code |
| COMPAT-03 | CLI subcommands unaffected by any `internal/search`/`internal/catalog` change this phase makes | integration (Go) | `go test ./cli/...` — re-run after every change to `internal/search`, per milestone `PITFALLS.md` Pitfall 29 | ✓ (existing, `cli/show_test.go` et al.) |

### Sampling Rate

- **Per task commit:** `go test ./internal/search/... ./internal/config/...` (fast, scoped to the packages touched) plus `go test ./cli/...` whenever `internal/search` changes (Pitfall 29 discipline, carried from milestone `PITFALLS.md`).
- **Per wave merge:** `go test ./...` (full suite, ~0.1s per the current baseline) + a dev-browser pass against the 40k fixture for any wave that touches `TreePane`/virtualization.
- **Phase gate:** Full `go test ./...` green, plus the dev-browser DOM-row-count check for TREE-01, plus explicit acknowledgment that Windows `RevealInFileManager` behavior (Pitfall N5) is unverified on this development machine and needs a separate check before `/gsd-ship`.

### Wave 0 Gaps

- [ ] `internal/search/flatten_test.go` — covers `LoadCatalogFlat` structure/dual-format/benchmark tests above
- [ ] `internal/search/service_test.go` extension — covers `BrowseCatalogs` `parseError` cases
- [ ] `internal/config/counts_cache_test.go` — new file, covers sidecar cache hit/miss/corrupt/concurrent
- [ ] `internal/osutil/reveal_test.go` (or wherever `RevealInFileManager` lands) — build-tag-scoped per-OS argv construction tests
- [ ] A committed 40k-node fixture *generator* (per CONTEXT.md's locked decision) — this is itself a Wave 0 deliverable other tests depend on, not an optional extra
- [ ] No frontend test framework gap to fill — TEST-01 is deferred; dev-browser is the only frontend validation surface this phase uses

## Security Domain

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V2 Authentication | No | Single-user local desktop app, no auth surface in this phase |
| V3 Session Management | No | Not applicable |
| V4 Access Control | No | Not applicable — filesystem permissions are the OS's own concern |
| V5 Input Validation | **Yes** | `RevealInFileManager(path)` and any path passed into `os/exec` must be validated (exists, resolved to absolute) before spawning a process — see Known Threat Patterns below |
| V6 Cryptography | No | Not applicable to this phase |
| V12 File Handling | **Yes** | `BrowseCatalogs`'s new full-content JSON reads (Pitfall N3) and the sidecar cache file both read/write local files; both must fail closed (never silently corrupt or leak partial state) per this project's own "Silent Fallbacks" convention |

### Known Threat Patterns for this stack

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| OS command injection via `RevealInFileManager`'s path argument | Tampering / Elevation of Privilege | **Never** build the command as a shell string (`exec.Command("sh", "-c", "open -R "+path)` or Windows `cmd /c` string concatenation) — always pass the path as a distinct `exec.Command` argv element (Pattern 5). Because Go's `exec.Command` bypasses the shell entirely when called this way, no shell metacharacter in a filesystem path (`;`, `|`, `&&`, backticks) can escape into a second command — this is the standard, sufficient mitigation for this exact risk class `[CITED: WebSearch summary of general Go command-injection guidance from stackhawk.com/snyk.io, MEDIUM confidence]`. The realistic exposure here is low (paths originate from the local filesystem catalog, not remote/untrusted network input) but the mitigation costs nothing and removes the risk class entirely. |
| Windows `explorer /select,` argument-shape uncertainty (Pitfall N5/Assumption A1) causing an unexpected argv split | Tampering (low severity) | If the "single argv element" form (A1) turns out to be wrong and a naive two-element form is used instead, worst case is Explorer opening an unexpected window — not a security escalation, since argv-based `exec.Command` still never invokes a shell. Treat as a correctness bug to verify, not a security gap. |
| Raw parser error text (STATE-02) dumped into the UI without bounding | Denial of usability (not a security vulnerability per se) | Milestone `PITFALLS.md`'s Security Mistakes table already flags this (truncate/summarize a pathologically large raw JSON parse error before rendering) — carried forward unchanged into this phase's STATE-02 implementation, not re-derived here. |
| Sidecar cache file (`counts-cache.json`) corrupted or partially written by a concurrent writer | Tampering (accidental, not adversarial) | Same atomic-write discipline the milestone `PITFALLS.md` Pitfall 18 already establishes for catalog writes (temp-file-in-same-dir + `os.Rename`) applies equally here — the cache is local convenience data, but a half-written JSON file should degrade to "cache miss," never crash `BrowseCatalogs`. |

## Sources

### Primary (HIGH confidence)
- Direct code reading, this session, in full: `app.go`, `pkg/models/catalog.go`, `internal/search/service.go`, `internal/catalog/service.go`, `internal/config/config.go`, `cli/show.go`, `frontend/src/contexts/AppContext.tsx`, `frontend/src/services/wailsAPI.ts`, `frontend/src/components/workspace/{CatalogRail,TreePane,DetailsPanel,StatusBar,WorkspaceShell}.tsx`, `frontend/src/workspace.css` (partial, targeted grep + read), `frontend/src/themeTokens.ts` (partial), `internal/search/service_test.go` (partial), all seven `22-0*-SUMMARY.md` files, `23-CONTEXT.md`, `23-UI-SPEC.md`, `.planning/REQUIREMENTS.md`, `.planning/STATE.md`, `.planning/research/ARCHITECTURE.md`, `.planning/research/PITFALLS.md`.
- Measured, this session: 42,551-node synthetic `FlatCatalog` marshaled via Go's actual `encoding/json` → 6,114,397 bytes (5.83 MB). Script: `/private/tmp/.../scratchpad/wiresize/main.go` (throwaway, not committed).
- Measured, this session: `go test ./...` (all 6 packages pass), `go version` (go1.26.6), `node --version` (v24.14.1), `npm --version` (11.18.0), `npx tsc --version` (4.9.5), `wails version` (v2.10.2), `npm view @tanstack/react-virtual version/peerDependencies/repository.url/scripts.postinstall`.
- `gsd-tools query package-legitimacy check` — `@tanstack/react-virtual` verdict SUS (too-new heuristic on latest-version publish date), signals: 21,143,075 weekly downloads, repo `github.com/TanStack/virtual`, no postinstall script.

### Secondary (MEDIUM confidence)
- WebSearch summary of `tanstack.com/virtual/latest/docs/framework/react/react-virtual` — `useVirtualizer` API shape (`count`, `getScrollElement`, `estimateSize`, `scrollToIndex`, `getVirtualItems`, `getTotalSize`).
- WebSearch summary of Go `encoding/json` `SyntaxError.Offset` semantics (multiple sources: pkg.go.dev, adrianhesketh.com, alexedwards.net).
- WebSearch summary of general Go `exec.Command` argv-safety guidance (stackhawk.com, snyk.io, golang-nuts mailing list) — supports the "never build a shell string" mitigation.
- WebSearch summary of `pypi.org/project/show-in-file-manager` documentation describing the Linux "no universal select mechanism" problem (`nautilus --select`, `dolphin --select`, `caja --select` per-file-manager fragmentation).
- Milestone research `[CITED: .planning/research/ARCHITECTURE.md]` for the original `LoadCatalogFlat`/sidecar-cache design intent — corrected in one respect this session (the `Name`-is-basename assumption, see Pattern 1).

### Tertiary (LOW confidence)
- `[ASSUMED]` Windows `explorer /select,"+path"` single-argv-element form (Pitfall N5, Assumption A1) — training-knowledge pattern, not confirmed against an authoritative source this session; flagged for empirical Windows verification.
- `[ASSUMED]` No Wails v2 bound-method payload size limit exists (Assumption A3) — absence of a documented limit found via WebSearch is not proof none exists; flagged as the highest-priority empirical validation for this phase.

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — `@tanstack/react-virtual` version/peer-deps/downloads verified via `npm view` this session; the package name itself was carried from CONTEXT.md's locked decision (not independently discovered), so tagged per the package-name provenance rule.
- Architecture (Go-side flatten/cache/reveal design): HIGH — every claim about existing code shape is read-and-quoted this session; the two genuinely new designs (sidecar cache mutex, parse-error fast-path) are reasoned recommendations, not verified against an external authority, and are labeled as such.
- Architecture (frontend virtualization/state): MEDIUM — the `@tanstack/react-virtual` API shape and the `AppContext` extension pattern are sound but not verified against the library's official docs (no authoritative-docs fetch tool was available this session) or against a working implementation.
- Pitfalls: HIGH for the six new pitfalls (N1-N6), all grounded in code read this session; the milestone's own `PITFALLS.md` (referenced, not repeated) is independently HIGH per its own research pass.
- Validation architecture: HIGH for the Go-testable half (existing test infrastructure confirmed working); MEDIUM for the browser-verifiable half (dev-browser workflow confirmed available per the task brief, but this session did not execute a dev-browser session against this phase's not-yet-built UI).

**Research date:** 2026-08-13
**Valid until:** ~30 days for the Go-side findings (stable stdlib/repo structure); ~14 days for the `@tanstack/react-virtual` version pin, given it published a new release two weeks before this research.
