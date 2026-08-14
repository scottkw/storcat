# Phase 23: Rail + Virtualized Tree - Context

**Gathered:** 2026-08-13
**Status:** Ready for planning
**Mode:** Smart discuss (autonomous)

<domain>
## Phase Boundary

Wire real data into the shell Phase 22 built: the catalog rail lists and filters every catalog in the configured directory, selecting one loads its tree into a virtualized pane that stays smooth at 40,000+ nodes, and the details panel follows the selection with working open/reveal actions. The status bar goes live, and the empty-library and unreadable-catalog states appear with real diagnostic detail.

**In scope:** `LoadCatalogFlat` and the flat-node model, tree virtualization, expand/collapse and expand-all/collapse-to-root, catalog header and breadcrumb, rail listing/filtering/selection/red-dot, sidecar count cache, live status bar, details panel content and its two actions, reveal-in-file-manager, v1/v2 compatibility on the flat path, and the empty-library / unreadable-catalog states.

**Out of scope (later phases):** ⌘K palette (Phase 24), create slide-over (Phase 25), Settings — including the catalog-directory picker's real UI (Phase 26), catalog actions and fsnotify watch (Phase 27), re-scan and diff (Phase 28). The rail's "＋ New" pill and directory chip get wired to open their targets only when those targets exist; in this phase they may remain inert if the target is a later phase.

</domain>

<decisions>
## Implementation Decisions

### Tree Flattening & Virtualization
- Flattening happens in **Go**, via a new `LoadCatalogFlat(path)` returning a flat slice of nodes carrying `name`, `type`, `size`, `depth`, parent index, and `hasChildren`. Milestone research is explicit: flatten in Go and leave `LoadCatalog` untouched, because `cli/show.go` depends on its current shape.
- Virtualization uses **`@tanstack/react-virtual`** — headless, ~4KB, and its fixed-size mode matches the locked `--rh` row-height contract exactly. Not hand-rolled (a solved problem), not `react-window`.
- Go returns the **full flat array once** per catalog; TypeScript derives the visible slice from an `expanded` Set. No IPC round-trip per expand — this is what makes "expand every directory" instant rather than 40,000 calls.
- Row height is **fixed, read from `--rh`** (27px Compact / 34px Comfortable). No dynamic measurement: it is measurably slower at 40k rows and contradicts the Phase 22 token contract.

### Rail Data, Counts & Filtering
- Per-catalog file count and total bytes come from a **sidecar cache keyed on path + mtime** (already a locked milestone decision in PROJECT.md), stored beside the config file. Catalog JSON on disk stays byte-identical — writing counts into it would break COMPAT-01.
- Filtering runs **in TypeScript over the already-loaded rail array**. Catalogs number in the tens; a Go binding would be an IPC round-trip for a `.filter()`. Match is case-insensitive on both title and filename.
- The filter string lives in **local `useState` in the rail component, read through `useDeferredValue`** — it never enters `AppContext`, so the tree has nothing to re-render against on a keystroke (RAIL-02). No debounce timer to tune.
- Parse failures are detected by **`BrowseCatalogs` gaining a `parseError string` field per catalog**. It already stats every file, so this costs one pass — not a second validation binding, and not lazy detection on click (the red dot must be visible before the user clicks).

### Backend Surface & Compatibility
- This phase adds **exactly three backend surfaces**: `LoadCatalogFlat`, `RevealInFileManager`, and three new fields on `CatalogMetadata` (`fileCount`, `totalBytes`, `parseError`). The six existing CLI subcommands are untouched — new capabilities are GUI-only per the milestone decision.
- `LoadCatalogFlat` **reuses `LoadCatalog`'s existing dual-format parse** (v1 array wrapper + v2 bare object) and then flattens. One parser means one place for a compatibility bug to not live (COMPAT-01).
- `RevealInFileManager(path)` is a **new Go binding with per-OS behavior**: `open -R` on macOS, `explorer /select,` on Windows, `xdg-open` on the parent directory for Linux (which has no universal select-the-file mechanism).
- The 40,000-node fixture is **generated on demand** — a Go test helper plus a small committed generator script writing into a temp dir. A 40k-node JSON blob is not committed to the repo.

### Selection, State & Error Surfaces
- Selection and expansion live in the **existing `AppContext` reducer**: `currentCatalogId`, `expanded: Record<string, boolean>`, `selected: string | null` — matching the handoff's documented state shape. The details panel and breadcrumb both read it, so it cannot be component-local.
- Switching catalogs fires **one atomic reducer action** that clears `expanded`, clears `selected`, and resets scroll to 0 (TREE-06). Nothing can end up half-applied.
- The unreadable-catalog state (STATE-02) renders **inline in the tree pane** — filename, byte offset, reason, and the raw parser error in a `--ch` code box — with the rail keeping its red dot. Not a toast or modal: dismissible means losable, and the requirement is that the raw error stays inspectable.
- Status bar counts (SHELL-06) are **derived from the rail array**, which already carries per-catalog counts once the sidecar cache lands. Summing three numbers does not need its own binding.

### Resolved During Planning (from research findings)
- **Ship the single `LoadCatalogFlat` call; do not chunk.** Research measured a synthetic 42,551-node catalog at **5.83 MB** (~144 bytes/node) of marshaled JSON. The Wails bridge is in-process, not a network hop, and the UI-SPEC already specifies a loading state for exactly this. Chunking is real machinery for a problem not yet observed — YAGNI. **But the plan must actually measure it** (time-to-first-row against the generated 40k fixture) and record the number, so "it's fine" is evidence rather than assumption. If the measurement is bad, chunking becomes a follow-up, not a guess made now.
- **`CatalogItem.Name` holds a full relative path, not a basename** (`internal/catalog/service.go:86-90`) — this corrects the milestone `ARCHITECTURE.md` draft. The flat node carries BOTH: `Name` via `filepath.Base` and `Path` verbatim. Getting this wrong silently breaks the tree's row labels and the breadcrumb.
- **`AppContext` does not yet have `currentCatalogId` / `expanded` / `selected`.** Phase 22 added only `density` / `railSide` / `detailOverlay`. The CONTEXT wording above describes the target shape; this phase actually adds those three.
- **The sidecar cache must NOT copy `config.Manager` verbatim** — that manager has no mutex, and the cache is written from a background fill while the rail reads it. It needs its own concurrent-safe access.
- **`parseError` detection uses a `json.Valid()` fast path**, falling back to a full `Unmarshal` only for the rare catalog that fails. `BrowseCatalogs` currently only stats files and never opens them, so this is genuinely new I/O on every rail load and must not be naive.
- **`@tanstack/react-virtual` is pre-approved for install.** The package-legitimacy gate flags it `SUS`, but that heuristic fires on the latest version's publish date, not package age: it is the official TanStack package with ~21M weekly downloads. Install it; do not substitute or hand-roll. This authority replaces the checkpoint — do not re-raise it.

### Claude's Discretion
- Flat-node struct field names and the exact `CatalogMetadata` JSON keys.
- Sidecar cache file name, location beside the config, and its serialization format.
- Component decomposition within the tree pane (header, breadcrumb, rows).
- Breadcrumb interaction details beyond the specified collapse-to-root.
- Whether the rail's "＋ New" pill and directory chip stay inert this phase or get a minimal wiring, given their real targets are Phases 25 and 26.

</decisions>

<code_context>
## Existing Code Insights

### Reusable Assets
- `app.go` — existing bindings `BrowseCatalogs(catalogDir)`, `LoadCatalog(filePath)`, `GetCatalogHtmlPath(catalogPath)`, `ReadHtmlFile`, `OpenExternal`, `SelectDirectory`, `GetConfig`. `LoadCatalogFlat` and `RevealInFileManager` join these.
- `internal/catalog/service.go` — `countFiles()` and `countDirectories()` already exist and are the natural inputs to the sidecar count cache; `formatBytes`/`formatBytesForDisplay` already implement the display formatting the rail and status bar need.
- `pkg/models/catalog.go` — `CatalogItem` (recursive) and `CatalogMetadata` (rail row shape) are the structs to extend; `CatalogItem` itself must not change shape (CLI + on-disk compat).
- `internal/config/config.go` — where the sidecar cache should live alongside.
- `frontend/src/services/wailsAPI.ts` — the `{success, ...}` envelope wrapper every new binding must follow.
- Phase 22's shell: `CatalogRail.tsx`, `TreePane.tsx`, `DetailsPanel.tsx`, `StatusBar.tsx` are built as skeletons with the correct dimensions, tokens and empty states — this phase fills them, it does not rebuild them.
- `frontend/src/contexts/AppContext.tsx` — the reducer, already pruned and extended in Phase 22.

### Established Patterns
- Go bindings return `(*T, error)`; the frontend wrapper converts to a `{success, ...}` envelope. Never throw across the boundary.
- Directory traversal silently skips inaccessible entries with a `console.warn`-equivalent; catalog parsing supports both v1 (array-wrapped) and v2 (bare object) formats.
- Go: `go fmt`, `golangci-lint`, context-aware functions, table-driven tests in `*_test.go` beside the source.
- Frontend: function-declaration components, `PascalCase.tsx`, plain CSS + custom properties, `UPPER_SNAKE_CASE` reducer actions.

### Integration Points
- `app.go` — new binding registrations.
- `frontend/wailsjs/` — regenerated bindings (do not hand-edit).
- `frontend/package.json` — the one new runtime dependency this phase adds (`@tanstack/react-virtual`).
- Phase 22's `workspace.css` — the `--rh` / `--rp` / `--mp` density tokens and the z-index scale are already declared; consume them, do not redeclare.

### Reference Material
- `design_handoff_storcat_ui/README.md` §"Screens / views → 1. Workspace" (rail, tree, details), §"Interactions & behavior", §"State management".
- `design_handoff_storcat_ui/designs/StorCat 1a Demo.dc.html` — the working prototype's rail rows, tree rows, breadcrumb and details panel.
- `.planning/research/ARCHITECTURE.md` and `PITFALLS.md` — the `LoadCatalogFlat` recommendation and the 40k-node performance notes.

</code_context>

<specifics>
## Specific Ideas

- The 40,000-node requirement is a real gate, not a slogan: it must be exercised against a generated fixture of that size, with smooth scrolling and no freeze, before the phase can pass.
- Directory click toggles expansion **and** selects; file click selects only (TREE-02).
- Breadcrumb ancestor segments are accent-colored; the bar also carries expand-all and collapse-to-root (TREE-03, TREE-05).
- The catalog header shows title, `.json`/`.html` chips, and a metadata line of file count / JSON size / bytes catalogued / modified date (TREE-04).
- Every path, filename, size, count and timestamp renders in IBM Plex Mono per the Phase 22 type contract.
- The unreadable-catalog panel must surface byte offset and the raw parser error verbatim — enough to actually diagnose a broken JSON file, not just "failed to load".

</specifics>

<deferred>
## Deferred Ideas

- ⌘K search palette across catalogs — Phase 24.
- Create slide-over behind the "＋ New" pill — Phase 25.
- Real catalog-directory picker behind the rail's directory chip, and the Settings surface generally — Phase 26.
- Catalog rename/duplicate/delete and fsnotify watch keeping the rail current — Phase 27.
- Re-scan and diff — Phase 28.
- Frontend unit tests for the virtualizer (TEST-01) — deferred to a separate testing milestone.

</deferred>
