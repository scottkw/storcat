# Phase 24: Cmd-K Command Palette - Research

**Researched:** 2026-08-14
**Domain:** Go/Wails thin-binding design (capped cross-catalog search), React state architecture for async deep-navigation (reveal-in-tree), hand-rolled modal-behavior hook (focus trap / scroll lock / Escape / focus restore), `@tanstack/react-virtual` scroll-to-index timing
**Confidence:** HIGH for all in-repo structural claims (every file below was read in full this session, several claims are direct quotes with line numbers). LOW for every externally-sourced claim, including the `@tanstack/react-virtual` scroll-timing findings — this project's `classify-confidence` seam rates the `webfetch`/`websearch` providers LOW regardless of the target domain's authority (confirmed by running `gsd_run query classify-confidence --provider webfetch --verified`, which still returned `LOW`), since no Context7 or other MCP-mediated documentation tool was available this session. The TanStack Virtual API-shape claims are tagged `[CITED: tanstack.com]` in-text because they were read directly off the vendor's own docs page (not inferred), but per this project's confidence tiers that is LOW, not MEDIUM — treat every external-source claim below as needing empirical confirmation during execution, not as settled fact.

## Summary

This phase adds exactly one new Go binding (a thin, capped wrapper around the existing `search.Service.SearchCatalogs` walk), one new frontend hook (`useModalBehavior`), one new piece of `AppContext` state (`pendingReveal`), and one new component tree (the palette overlay). No new npm dependency. The riskiest technical unknowns are not "how do I write React" — they are two very specific, previously-mis-stated facts about this codebase's own state machine that will silently break the phase's two hardest requirements (PLT-04/05 reveal, PLT-07 hook reuse) if planned against the wrong assumption:

1. **`SET_EXPANDED` in `AppContext.tsx` REPLACES the whole `expanded` map — it does not merge.** The UI-SPEC's own prose (E5 "partial" row) says the reveal path should use "`SET_EXPANDED` (merging the ancestor chain into the existing map)", but the actual reducer code at `frontend/src/contexts/AppContext.tsx:123-126` does `return { ...state, expanded: action.payload }` — a full overwrite of the map, not a merge. If the reveal path dispatches `SET_EXPANDED` with only the target's ancestor chain, it will **collapse every other branch the user had open** — violating the UI-SPEC's own explicit constraint that "Ancestor expansion must therefore set each ancestor to `true`, never toggle it... Revealing the same node twice in a row must be idempotent, not an open/close flicker" (and implicitly: must not touch unrelated open branches). The plan must either (a) merge client-side before dispatching `SET_EXPANDED` (`{ ...state.expanded, ...ancestorMap }`), or (b) add a new merging action. Merging client-side before dispatch requires no reducer change and is the smaller diff.
2. **The catalog root is confirmed excluded from `LoadCatalogFlat`'s output** (`internal/search/flatten.go:19-20,63-67`) — direct quote: "The root itself is excluded from the returned Nodes; the root's direct children get Depth 0 and ParentIdx -1" and the walk loop starts at `for _, child := range root.Contents { walk(child, 0, -1) }`. This matches Phase 23's SUMMARY claim; the ancestor-walk from a search hit's path must stop at `ParentIdx === -1`, not attempt to resolve a synthetic root node.

Backend-side, the new binding is straightforward: `Service.SearchCatalogs` already walks every catalog in the directory and returns every match with no cap (`internal/search/service.go:64-92`); the new capped binding wraps it, slices to 50, and returns `{ results, total }`. `cli/search.go:61-62` calls `svc.SearchCatalogs` directly and needs zero changes. `wails generate module` (installed, `v2.10.2` confirmed via `wails version`) regenerates `frontend/wailsjs/go/main/App.d.ts`/`App.js` and `frontend/wailsjs/go/models.ts` from `app.go` — every existing binding in this repo already follows this generated-file pattern, so the new one is additive, not novel.

**Primary recommendation:** Wrap `SearchCatalogs`'s existing walk in a new `SearchIndexed(term, dir string) (*models.SearchIndexResult, error)` binding that caps to 50 in Go, add `pendingReveal` as a single reducer field carrying the target path plus enough hit metadata to distinguish file/directory, merge (not replace) the ancestor set client-side before dispatching `SET_EXPANDED`, and call `virtualizer.scrollToIndex` from a `useEffect` keyed on the reveal request (not synchronously in the click/Enter handler) so it observes the post-expansion `visibleIndices` — the existing `TreePane.tsx` fixed-row-height virtualizer (no `measureElement`/dynamic sizing) sidesteps the alignment bug that affects TanStack Virtual's *dynamic*-mode `scrollToIndex`, but the effect-timing requirement (new state must have committed before scrolling) still applies regardless of row-height mode.

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Cross-catalog search walk + 50-cap + total count | Backend (Go, `internal/search`) | — | Already the sole owner of catalog-file I/O and parsing; capping in Go avoids marshaling megabytes of JSON per keystroke (measured 5.83MB for one 42,551-node catalog, Phase 23) |
| Debounce + stale-response guard | Frontend (React, palette component) | — | Purely a UI-responsiveness concern; Go has no notion of "the user is still typing" |
| Palette open/close, focus trap, scroll lock, Escape | Frontend (`useModalBehavior` hook + palette component) | — | DOM-only concerns (focus, scroll, keyboard) with no backend counterpart |
| Ancestor expansion / selection / scroll-into-view | Frontend (`AppContext` reducer + `TreePane`) | — | Operates entirely on already-loaded `FlatNode[]` state; no new Go call needed once the target catalog's flat tree is loaded |
| Catalog switch (if hit is in a different catalog) | Frontend (`AppContext` `SELECT_CATALOG` → `TreePane`'s existing `useEffect`) | Backend (`LoadCatalogFlat`, already exists) | Reuses the exact Phase 23 catalog-load path; the reveal is a consumer of `TREE_LOADED`, not a new load mechanism |
| Global ⌘K keydown capture | Frontend (`window.addEventListener('keydown', ...)`) | — | No native OS-level menu accelerator exists in this app (`main.go` registers no `Menu`); the palette's own DOM listener is the only capture point, matching the existing Escape-listener precedent in `WorkspaceShell.tsx:34-41` |

## User Constraints

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

**Search Execution & Backend Surface**
- A new thin GUI-only Go binding wraps the existing `search.Service` walk rather than modifying `SearchCatalogs`. `cli/search.go` calls `Service.SearchCatalogs` and the milestone's locked decision is that new capabilities are GUI-only — so the existing method and its CLI caller stay byte-identical in behavior.
- The 50-result cap and the `N` total are both computed in Go. The service already walks every node in every catalog, so producing `total` is free, and only 50 results ever marshal across the Wails bridge. Return shape carries results, the total hit count, and enough information to render "Showing the first 50 of N hits" (PLT-03).
- Live search, debounced ~200 ms, with a stale-response guard. Each query carries a request id (or equivalent monotonic token); a response whose id is not the latest is dropped rather than rendered.
- Minimum query length is 2 characters. Below 2, the palette shows a "Type to search…" hint — a distinct state from the empty-result copy.

**Palette UI & Result Presentation**
- Centered dialog anchored ~15vh from the top, max-width 640px, with a scrim behind it, at z-index 200 (`details drawer 100 → create slide-over / ⌘K palette 200 → dialogs / Settings 300`). Do not invent a new z-index.
- Each result row renders: basename (with the match highlighted) + dimmed full path + a catalog-name chip + size. Reuse `frontend/src/lib/format.ts` for the size and the existing `.ws-chip` token/class convention for the catalog chip (UI-SPEC resolves this to a new sibling class, `.ws-palette-chip` — see below).
- No grouping — a flat list in backend order (catalog, then path).
- Highlight the matched substring in the basename only. The path stays plain-dimmed. Matching is case-insensitive on both name and path (PLT-02), but only the basename gets visual highlight treatment.

**Navigation to a Hit & the Shared Modal Hook**
- A hit reaches the tree through new `pendingReveal: string | null` state in `AppContext`, holding the target node's path. The palette dispatches the catalog switch plus the reveal request; `TreePane` consumes `pendingReveal` after `TREE_LOADED` lands and clears it.
- Reveal scrolls with `virtualizer.scrollToIndex(visibleIdx, { align: 'center' })`, where `visibleIdx` is the index within `useVisibleRows` output — ancestors must be expanded first so the target is actually present in the visible slice.
- Directory hits expand every ancestor and select the directory, but do not auto-expand the target itself.
- PLT-07's shared modal behavior is a hand-rolled `useModalBehavior({ isOpen, onClose })` hook in `frontend/src/hooks/`, covering focus trap, Escape-to-close, page scroll lock, and focus restore to the trigger element on close. No new dependency (`focus-trap-react` / `react-focus-lock` rejected). **Phases 25, 26, and 27 import this hook; they must not reimplement any of these four behaviors.**

### Claude's Discretion
- The exact name, signature, and return-struct shape of the new search binding, and the name of its Go-side result type.
- Component decomposition inside the palette (input, result list, result row, empty/hint/truncation states).
- The precise debounce constant, and the mechanism used for the stale-response guard.
- Keyboard specifics beyond the requirement: Up/Down/Enter/Escape are mandatory; Home/End/PageUp/PageDown wrap-around behavior is open (UI-SPEC resolves this to no-wrap, Home/End jump-to-edge, PageUp/PageDown by measured viewport row count).
- How the match-highlight substring is computed and rendered (UI-SPEC resolves this to case-insensitive `indexOf`, first occurrence only, no multi-span highlighting).
- Whether `pendingReveal` is one reducer action or a small pair, and how it is cleared.

### Deferred Ideas (OUT OF SCOPE)
- Non-search commands in the palette (run an action, jump to a setting) — file-search only; a general command registry belongs in its own phase if ever wanted.
- Fuzzy/subsequence matching and result ranking — PLT-02 specifies substring matching over names and paths; ranking is not a requirement.
- Search history / recent queries persistence.
- Frontend unit tests for palette keyboard navigation — already an explicitly deferred milestone item (TEST-01, deferred at v3.0.0 requirements definition).
</user_constraints>

## Phase Requirements

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| PLT-01 | User can open a search palette with ⌘K or by clicking the toolbar search field, with the input autofocused | Toolbar button already exists with no `onClick` (`Toolbar.tsx:58-61`); global keydown listener pattern already proven in this codebase for Escape (`WorkspaceShell.tsx:34-41`); autofocus is the shared hook's job per UI-SPEC |
| PLT-02 | User can search names and paths across every catalog in the directory | `Service.SearchCatalogs` already matches case-insensitively against `item.Name`, which is a full relative path (`service.go:72,121`) — no backend change needed, just capping |
| PLT-03 | User sees at most 50 results, with a "Showing the first 50 of N hits" notice when more matched | New Go binding caps server-side (see Package/Binding design below); avoids the 5.83MB/catalog problem measured in Phase 23 |
| PLT-04 | User can navigate results by keyboard and dismiss the palette with Escape | Standard combobox/listbox roving-active-index pattern (Claude's Discretion on wrap behavior, UI-SPEC resolves it); Escape handled by shared hook |
| PLT-05 | User can click a hit to switch to its catalog, expand every ancestor, select it, scroll it into view, and close the palette | Requires the `SET_EXPANDED`-replaces-not-merges fix (see Summary #1) and effect-timing fix for `scrollToIndex` (see Pitfalls) |
| PLT-06 | User sees "No file in any catalog matches that." when nothing matched | Exact string, verbatim — no backend involvement, pure frontend state branch |
| PLT-07 | Focus is trapped inside the palette while open, and page scroll is locked behind it, via a shared `useModalBehavior` hook Phases 25-27 reuse | No existing precedent in this codebase (the current details drawer close-on-Escape/backdrop-click in `WorkspaceShell.tsx:30-41` is bespoke, not a hook — this phase is the first to extract the pattern) |
</phase_requirements>

## Standard Stack

### Core
No new external dependencies this phase. Every capability is built on already-installed libraries:

| Library | Version (installed) | Purpose | Why Standard (in this repo) |
|---------|---------|---------|--------------|
| `@tanstack/react-virtual` | 3.14.9 [VERIFIED: frontend/package.json] | Reused, unmodified — `scrollToIndex` on the existing `TreePane` virtualizer instance | Already the project's chosen virtualization library (Phase 23); adding a second instance for a max-50-row palette list would be needless — the list is never large enough to need virtualization itself |
| Go stdlib (`sort`, `strings`) | Go 1.23 [VERIFIED: go.mod:3] | Slicing/capping the existing `[]*models.SearchResult` to 50 and counting the total | No sorting requirement locked (backend order = catalog-then-path, i.e. insertion order); `sort` is not actually needed unless a stable secondary order is desired — `strings`/slicing only |
| Wails v2 runtime | v2.10.2 [VERIFIED: `wails version` / go.mod:10] | Binding generation (`wails generate module`) for the new Go method + new result struct | Already the project's only IPC mechanism; every existing binding (`SearchCatalogs`, `LoadCatalogFlat`, etc.) follows this exact generate-then-import pattern |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| None | — | — | This phase deliberately adds zero npm packages, per `24-CONTEXT.md`'s explicit rejection of `focus-trap-react`/`react-focus-lock` and the UI-SPEC's Registry Safety section confirming "No new npm dependency is added this phase" |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Hand-rolled `useModalBehavior` hook | `focus-trap-react` or `react-focus-lock` | Rejected in `24-CONTEXT.md`: "roughly 60 lines of hook against a package the project would carry for one overlay family" — and per WebSearch, the native `<dialog>` element or the `inert` attribute are 2025-era alternatives to a hand-rolled trap, but neither was evaluated by the user and both are a bigger behavioral change (native `<dialog>` changes stacking/rendering semantics; `inert` requires marking the *rest* of the DOM tree inert, not just trapping the modal) than the locked "~60-line hook" scope — not recommended to introduce mid-research without a discuss-phase reopening |
| Frontend `.slice(0, 50)` | Backend cap | Explicitly rejected in `24-CONTEXT.md` for the multiplied-payload reason above — do not resurrect this |
| `AbortController` for stale-response guard | Request-id ref | Wails-generated bindings return plain `Promise`s with no `AbortSignal` parameter (confirmed: `frontend/wailsjs/go/main/App.d.ts` — every exported function takes only its declared Go args, no signal); an `AbortController` would abort the fetch-equivalent only in spirit (React would still await the promise to resolution) — a request-id/generation-counter ref is the only mechanism that actually works against this binding shape |

**Installation:** None — no `npm install` or `go get` needed this phase.

**Version verification:** `@tanstack/react-virtual` `^3.14.9` confirmed installed via `frontend/package.json` (already present from Phase 23, not reinstalled). Wails `v2.10.2` confirmed via both `go.mod:10` and `wails version` CLI output. Go `1.23` confirmed via `go.mod:3`.

## Package Legitimacy Audit

**Not applicable — this phase adds no new packages.** No `npm install` or `go get` command appears anywhere in this research or in `24-CONTEXT.md`/`24-UI-SPEC.md`. The Package Legitimacy Gate protocol is skipped per its own scope ("whenever this phase installs external packages").

## Architecture Patterns

### System Architecture Diagram

```
User keystroke (⌘K) or click .ws-search
        │
        ▼
Toolbar.tsx (existing button, gains onClick)  ──┐
        │                                        │  both open the same
Global window keydown listener (new)  ───────────┘  local `open` state
        │
        ▼
CommandPalette (new component, mounted in WorkspaceShell)
  useModalBehavior({ isOpen, onClose })  ← focus trap, scroll lock, Escape, focus-restore
        │
        ▼
User types (local useState query, NOT AppContext — Phase 23 rail-filter precedent)
        │
        ▼  (debounce ~200ms, requestId ref increments)
wailsAPI.searchIndexed(term, catalogDir)  ── new thin wrapper ──▶  App.SearchIndexed (Go)
        │                                                              │
        │                                                              ▼
        │                                                  search.Service.SearchCatalogs()
        │                                                  (UNCHANGED — same method CLI calls)
        │                                                              │
        │                                                    Go slices to 50, computes total
        │                                                              │
        ◀──────────────── { results: SearchResult[≤50], total: int } ──┘
        │
   requestId still current? ── no ──▶ drop response (stale-response guard)
        │ yes
        ▼
Render: hint | searching | results (≤50 rows + optional truncation line) | no-matches
        │
   user clicks / Enters a row
        │
        ▼
dispatch SELECT_CATALOG (if hit.catalog !== currentCatalogId)
dispatch SET_PENDING_REVEAL(path)  ── AppContext.pendingReveal
onClose()  (palette closes, focus restores to trigger)
        │
        ▼ (async — catalog load may already be in flight or may need to start)
TreePane's existing useEffect(loadCatalogFlat) fires on currentCatalogId change
        │
        ▼  TREE_LOADED lands
useEffect in TreePane, keyed on [pendingReveal, state.tree.status]:
  1. find node index by path in nodes[] (build/reuse a path→index Map)
  2. walk ParentIdx chain to root (stops at ParentIdx === -1) → ancestor path set
  3. merge ancestor set into state.expanded (NOT overwrite) → dispatch SET_EXPANDED
  4. dispatch SET_SELECTED(path)
  5. clear pendingReveal
        │
        ▼ (next render: visibleIndices now includes the target)
second useEffect, keyed on [pendingReveal-was-just-cleared, visibleIndices]:
  virtualizer.scrollToIndex(visibleIdx, { align: 'center' })
```

### Recommended Project Structure
```
frontend/src/
├── components/workspace/
│   ├── CommandPalette.tsx        # top-level overlay: scrim, panel, wires useModalBehavior
│   ├── palette/
│   │   ├── PaletteInput.tsx      # input row + magnifier + right-aligned hint/count text
│   │   ├── PaletteResultList.tsx # role="listbox", maps results, owns active-index state
│   │   └── PaletteResultRow.tsx  # role="option", shape/basename/highlight/path/chip/size
│   └── Toolbar.tsx                # MODIFIED: .ws-search gains onClick
├── hooks/
│   └── useModalBehavior.ts        # NEW: focus trap + Escape + scroll lock + focus restore
├── contexts/
│   └── AppContext.tsx             # MODIFIED: + pendingReveal state/action
├── services/
│   └── wailsAPI.ts                 # MODIFIED: + searchIndexed wrapper (extractErrorMessage pattern)
└── components/workspace/TreePane.tsx  # MODIFIED: + pendingReveal-consuming useEffect(s)

internal/search/
└── (SearchCatalogs UNCHANGED — new capping logic lives in app.go's new binding method,
     or a new small method on Service if capping logic is judged worth unit-testing in
     isolation from the Wails layer — Claude's Discretion per CONTEXT.md)

pkg/models/
└── catalog.go                      # MODIFIED: + new result struct (e.g. SearchIndexResult)
```
Component decomposition (`PaletteInput`/`PaletteResultList`/`PaletteResultRow` vs. one file) is Claude's Discretion per `24-CONTEXT.md`; the split above follows this codebase's existing convention of one component per rendered concern (compare `TreeHeader.tsx` + `BreadcrumbBar.tsx` + `UnreadableCatalogPanel.tsx` all composed inside `TreePane.tsx`'s render, `TreePane.tsx:7-9`).

### Pattern 1: Thin capping Go binding over an unmodified service method
**What:** `App.SearchIndexed` (or similar name) calls `a.searchService.SearchCatalogs(term, absPath)` unchanged, then caps and counts in the App-layer method — not inside `search.Service`.
**When to use:** Whenever a GUI-only view needs a *presentation* transform (cap, count) of data a CLI-shared service already produces in full, and the milestone's locked decision is "new capabilities are GUI-only."
**Example (sketch, not verified against a written implementation — this method does not exist yet):**
```go
// app.go — new method, sibling to the existing SearchCatalogs (app.go:84-91)
// SearchIndexed wraps searchService.SearchCatalogs with a GUI-only cap: it
// performs the identical full walk, then returns at most 50 results plus
// the true total count. cli/search.go is untouched -- it calls
// searchService.SearchCatalogs directly (cli/search.go:62), not this method.
func (a *App) SearchIndexed(searchTerm string, catalogDir string) (*models.SearchIndexResult, error) {
	absPath, err := filepath.Abs(catalogDir)
	if err != nil {
		return nil, err
	}
	all, err := a.searchService.SearchCatalogs(searchTerm, absPath)
	if err != nil {
		return nil, err
	}
	total := len(all)
	if total > 50 {
		all = all[:50]
	}
	return &models.SearchIndexResult{Results: all, Total: total}, nil
}
```
This exact function body is **not** present in the repo today (`app.go` currently has 260 lines, ending at `GetVersion`, `app.go:256-259`) — it is Claude's proposed sketch grounded in the confirmed shape of `SearchCatalogs` (`service.go:64-92`) and the existing `App.SearchCatalogs` binding (`app.go:83-91`), tagged `[ASSUMED]` for the exact naming/shape (locked as Claude's Discretion in `24-CONTEXT.md`).

### Pattern 2: `pendingReveal` as async-safe cross-component state
**What:** A reducer field (not a ref, not a DOM CustomEvent) carrying the pending target path, consumed by `TreePane` only after `TREE_LOADED` for the *matching* catalog has landed.
**When to use:** Any time one component (palette) needs to trigger a multi-step async sequence (switch catalog → wait for load → mutate a different component's derived state) where the intermediate async gap could span multiple renders/unmounts.
**Grounding:** `TreePane.tsx`'s existing `TREE_LOADED` dispatch already carries a `catalogId` and the reducer already discards a load that no longer matches `state.currentCatalogId` (`AppContext.tsx:95-99`, quoted: `if (action.payload.catalogId !== state.currentCatalogId) return state;`) — `pendingReveal` consumption must follow the identical discard-if-stale discipline: if the palette closes having switched to catalog A, then the user immediately switches to catalog B from the rail before A's `TREE_LOADED` lands, the pending reveal for A's target node must not fire against B's now-current tree. This is not yet an issue AppContext's current code handles for `pendingReveal` (the field does not exist yet) — the reducer/consuming effect must be designed to check `pendingReveal` state against `currentCatalogId`/`TREE_LOADED`'s own id the same way `TREE_LOADED` already discards stale loads.

### Pattern 3: Merge-not-replace ancestor expansion
**What:** Before dispatching `SET_EXPANDED`, compute `{ ...state.expanded, ...ancestorPathsToTrue }` in the consuming code (e.g., inside the `TreePane` effect that owns `dispatch`), since the reducer itself performs a full replace.
**When to use:** Every ancestor-reveal dispatch, per the UI-SPEC's own explicit non-negotiable ("Ancestor expansion must therefore set each ancestor to true, never toggle it... idempotent, not an open/close flicker" — UI-SPEC E5 partial row) combined with the actual reducer behavior confirmed by reading `AppContext.tsx:123-126`:
```typescript
// AppContext.tsx:123-126 (VERBATIM, current code)
case 'SET_EXPANDED':
  // Replaces the whole map in one state update -- expand-all and
  // collapse-to-root both use this, never a per-node dispatch loop.
  return { ...state, expanded: action.payload };
```
The comment itself documents the *intended* use (expand-all / collapse-to-root, both of which legitimately want a full replace). A reveal is a *different* use case that needs a merge, so the reveal's dispatch site must pass `{ ...state.expanded, ...ancestorMap }` as `action.payload`, not just `ancestorMap`.

### Anti-Patterns to Avoid
- **Dispatching `SET_EXPANDED` with only the ancestor chain:** Silently collapses every other open branch in the tree — a regression a user would experience as "I had three folders open, clicked a search result, and now only one path is expanded." Not covered by any existing test (frontend has no test framework, TEST-01 deferred) — must be caught by manual/dev-browser verification with a pre-existing expanded branch as setup.
- **Calling `virtualizer.scrollToIndex` synchronously in the same function that dispatches the expansion:** The `virtualizer` object read in a click handler is a closure over the *pre-dispatch* render's `count` (derived from `visibleIndices.length`, itself derived from the *pre-expansion* `state.expanded`) — see Common Pitfalls below.
- **Frontend-side `.slice(0, 50)` on an uncapped Go response:** Explicitly rejected in `24-CONTEXT.md` for measured-payload reasons (Phase 23: 5.83MB/42,551-node catalog) — this is not a style preference, it's a locked decision with a numeric justification.
- **A per-ancestor `TOGGLE_EXPAND` loop instead of one `SET_EXPANDED`:** `TOGGLE_EXPAND` flips per path (`AppContext.tsx:112-122`) — looping it over an ancestor chain would *collapse* any ancestor that happened to already be `true`, the exact bug the UI-SPEC calls out by name.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Cross-catalog file/path search | A new search algorithm or index | `Service.SearchCatalogs` (`service.go:64-92`), unchanged | Already correct, already tested (`internal/search/service_test.go`), already what the CLI uses — the entire point of this phase's backend work is capping its output, not reimplementing its logic |
| Size/count formatting in result rows | A new formatter | `frontend/src/lib/format.ts`'s `formatBytes`/`formatCount` | Locked in `24-CONTEXT.md`; already the single source of truth for size/count formatting across the rail, tree, and status bar |
| Virtualized scrolling for the ≤50-row result list | A new virtualizer instance for the palette list | Plain unvirtualized `<div>`/`map()` — 50 rows never needs virtualization | The 50-cap is the same product decision that makes virtualization for *this* list actively unnecessary — `@tanstack/react-virtual` stays reserved for the tree's 40k+ node case |
| Focus trap / scroll lock | `focus-trap-react`, `react-focus-lock`, or a fresh reimplementation of either | The locked hand-rolled `useModalBehavior` hook | `24-CONTEXT.md` explicitly rejected both libraries for a one-overlay-family use case; but *within* the hook, don't reinvent `document.activeElement` tracking from scratch each phase — this hook is a single shared implementation precisely so nobody hand-rolls it four times (this phase, 25, 26, 27) |

**Key insight:** This phase's actual engineering risk is not "build a search UI" (a solved, common pattern) — it is getting three pieces of *this specific codebase's* async state machine exactly right: the reducer's replace-vs-merge semantics, the catalog-switch race between `pendingReveal` and `TREE_LOADED`, and the virtualizer's read-after-write timing. All three are invisible until tested live, none are caught by TypeScript, and the frontend has no test framework to catch a regression automatically (TEST-01 deferred) — treat this phase's live dev-browser pass as the primary safety net, not typechecking.

## Common Pitfalls

### Pitfall 1: `scrollToIndex` fires against the pre-expansion virtualizer
**What goes wrong:** The reveal handler expands ancestors and calls `scrollToIndex` in the same tick/handler; the row is not yet in `visibleIndices` because the `expanded` state update that would put it there hasn't committed and re-rendered yet, so `scrollToIndex` either scrolls to the wrong row (whatever index N currently means) or silently no-ops if N is out of the current (smaller) range.
**Why it happens:** `useVirtualizer` is called fresh on every render of `TreePane` (`TreePane.tsx:26-31`) with `count: visibleIndices.length`; `visibleIndices` is itself a `useMemo` over `[nodes, state.expanded]` (`useVisibleRows.ts:16-31`). The `virtualizer` object obtained in any given render call is bound to that render's `count`. A synchronous call from inside a click handler executes against the handler-invocation-time closure, before React has re-rendered with the new `expanded` map.
**How to avoid:** Do the expansion dispatch and the `scrollToIndex` call in **two separate effects**, the second keyed on a value that only changes *after* the expansion has committed and `visibleIndices` has been recomputed (e.g., `useEffect(() => { ...scrollToIndex... }, [pendingReveal, visibleIndices])`, guarded so it only fires once per reveal). TanStack Virtual's own issue tracker documents a related but distinct failure mode for *dynamic-mode* (measured, not fixed-height) scrolling, where the maintainer's workaround is a `setTimeout(fn, 0)` deferral inside a `useEffect` [CITED: github.com/TanStack/virtual/issues/615] — `TreePane` uses fixed `estimateSize: () => rowHeight` (`TreePane.tsx:29`), not `measureElement`, so the *dynamic-mode* alignment bug itself should not apply, but the general principle (defer to a post-commit effect, don't call synchronously mid-handler) still governs this codebase's own two-effect ordering requirement.
**Warning signs:** The palette closes, the correct catalog loads, the correct row highlights as selected, but the viewport doesn't scroll (or scrolls to a visually wrong position) — especially reproducible when the reveal requires expanding 3+ levels of previously-collapsed ancestors (a bigger `visibleIndices` delta between the pre- and post-expansion renders).

### Pitfall 2: `SET_EXPANDED` replace-semantics silently destroys unrelated open branches
**What goes wrong:** Described in Summary #1 and Pattern 3 above — a reveal that dispatches `SET_EXPANDED(ancestorMapOnly)` instead of a merged map collapses everything else the user had open.
**Why it happens:** The reducer case is a one-line full-object replace (`AppContext.tsx:126`), originally written correctly for its two existing callers (expand-all: pass the full node-path set; collapse-to-root: pass `{}`) — a third caller with different semantics (merge) was never anticipated in that code.
**How to avoid:** The reveal's dispatch site computes the merged map itself: `dispatch({ type: 'SET_EXPANDED', payload: { ...state.expanded, ...ancestorPaths } })`.
**Warning signs:** Manually open 2-3 unrelated tree branches, then trigger a reveal via the palette to a node in a *different, currently-collapsed* branch — if the previously-open branches collapse, this pitfall has been hit.

### Pitfall 3: Ancestor walk resolves an off-by-one root, or infinite-loops on a self-referential `ParentIdx`
**What goes wrong:** Walking `ParentIdx` naively without an explicit `=== -1` sentinel check could either (a) attempt `nodes[-1]` (root is excluded from the array, `flatten.go:19-20`) or (b) loop forever if a malformed/future catalog ever produced a cyclic parent chain.
**Why it happens:** `ParentIdx` is `-1` specifically for depth-0 nodes (`flatten.go:63-67`, `walk(child, 0, -1)`), meaning the loop terminator is a sentinel value, not "reached the root object" — there is no root object in the array to reach.
**How to avoid:** `while (idx !== -1) { ancestorPaths.add(nodes[idx].path); idx = nodes[idx].parentIdx; }` — but exclude the target node's own path from the ancestor set being *expanded* (the UI-SPEC is explicit: directory hits select but do not auto-expand the *target* itself, only its ancestors) — start the walk from `node.parentIdx`, not `idx` itself, when building the expand-set; use `idx` itself (the target) only for `SET_SELECTED` and `scrollToIndex`.
**Warning signs:** A directory hit's own children flash open on reveal (target auto-expanded, violating the explicit "do not auto-expand the target itself" rule) — or, if the walk starts one level too high, the *target's parent* fails to expand and the row is invisible despite "success" being dispatched.

### Pitfall 4: Global ⌘K listener conflicts with typing in the rail filter or the palette's own input
**What goes wrong:** A naive `keydown` listener bound at `window` could double-fire, refocus-and-clear an already-open palette on a second ⌘K press (UI-SPEC explicitly forbids this: "a second ⌘K while open is a no-op"), or need special-casing to still work while the rail filter input has focus.
**Why it happens:** Browsers deliver `keydown` to the focused element first, then bubble; a listener on `window` with the `capture` phase or a bubble-phase listener that doesn't call `preventDefault()` risks the browser's own default handling (in a bare browser tab, though not necessarily inside a frameless WKWebView) or, more relevantly here, simply needs an explicit `if (paletteOpen) return;` early-out to satisfy the no-op-on-repeat requirement.
**How to avoid:** Model the listener after the codebase's own working precedent: `WorkspaceShell.tsx:34-41` already registers a `window.addEventListener('keydown', ...)` for Escape, conditionally (only while `state.detailOverlay` is true), and cleans it up — this is direct in-repo evidence that a DOM `keydown` listener *does* fire correctly inside this app's actual Wails webview (at least for one key), which is reassuring context for the ⌘K listener even though Escape and ⌘K are different keys with potentially different OS/webview-level reservations. The ⌘K listener should check `event.metaKey || event.ctrlKey` (Cmd on macOS, Ctrl on Windows/Linux) and `event.key === 'k'`, call `preventDefault()` unconditionally (so no browser/webview default ever engages, regardless of platform), and no-op if the palette is already open — it does **not** need an input-focus exclusion, since the UI-SPEC explicitly requires it to fire even while the rail filter is focused.
**Warning signs (untestable from repo alone — flag for early manual verification):** ⌘K does nothing, inserts a literal "k" character into whatever text field has focus, or (on macOS specifically) triggers some WKWebView/Safari-chrome-adjacent behavior. No repo evidence confirms or denies this for Cmd+K specifically — WebSearch found only generic, non-conclusive reports of *some* keyboard shortcuts being intercepted by WKWebView on macOS in Wails apps [LOW confidence, WebSearch only — see Open Questions].

### Pitfall 5: Debounce race between the 200ms timer and rapid catalog/search-term changes
**What goes wrong:** A `setTimeout`-based debounce that doesn't clear its previous timer on every keystroke fires multiple overlapping searches; without a request-id guard, a slow response to an early keystroke can overwrite a fast response to a later one, showing stale results for the wrong query.
**Why it happens:** Wails-generated bindings return plain `Promise`s (confirmed: no `AbortSignal` parameter on any generated function in `frontend/wailsjs/go/main/App.d.ts`) — there is no way to actually cancel an in-flight Go-side call from the frontend; only the frontend's *handling* of the response can be gated.
**How to avoid:** Increment a `useRef`-held counter on every new debounced fire, capture its value in the closure, and check it still matches the ref's current value when the response resolves before calling `setState` — the standard "request id ref" pattern [LOW confidence, WebSearch-sourced pattern description — this is a widely-documented React pattern, not project-specific, but no single authoritative source was fetched this session].
**Warning signs:** Typing quickly and then pausing shows a flicker of an earlier query's results before settling on the correct final results (visible without any special tooling) — or, more subtly, briefly shows "No file in any catalog matches that." for a query that does have matches, if a fast empty response to an intermediate keystroke lands after a slower non-empty response to the final one.

## Runtime State Inventory

Not applicable — this is a greenfield-additive phase (new binding, new hook, new component tree, new reducer field). No rename, refactor, or migration of existing stored/registered state is involved.

## Code Examples

### Ancestor-walk from a search hit's `FullName` to the flat-node index
```typescript
// Sketch — not present in the repo yet. Grounded in FlatNode's confirmed shape
// (pkg/models/catalog.go:48-56, VERBATIM: `Path string \`json:"path"\`` /
// `ParentIdx int \`json:"parentIdx"\``) and the confirmed root-exclusion
// (internal/search/flatten.go:19-20).
//
// SearchResult.FullName (models/catalog.go:17, VERBATIM: `FullName string
// \`json:"fullName"\``) is the same full relative path FlatNode.Path carries
// -- both trace back to CatalogItem.Name (models/catalog.go:6, VERBATIM:
// `Name string \`json:"name"\``), so a hit's fullName can be looked up
// directly against a path -> index map built once per loaded tree.

function findNodeIndexByPath(nodes: models.FlatNode[], path: string): number {
  // Build once (e.g. useMemo keyed on `nodes`), not per-reveal, if this is
  // called more than once per tree load.
  for (let i = 0; i < nodes.length; i++) {
    if (nodes[i].path === path) return i;
  }
  return -1;
}

function ancestorPathsOf(nodes: models.FlatNode[], targetIdx: number): Set<string> {
  const ancestors = new Set<string>();
  let idx = nodes[targetIdx].parentIdx; // start at the PARENT, not the target itself
  while (idx !== -1) {
    ancestors.add(nodes[idx].path);
    idx = nodes[idx].parentIdx;
  }
  return ancestors;
}
```

### Merged `SET_EXPANDED` dispatch for a reveal (vs. the existing expand-all/collapse-to-root callers)
```typescript
// Existing reducer case is a full replace (AppContext.tsx:123-126, quoted
// above in Pattern 3) -- the CALLER must merge, the reducer does not.
const ancestorPaths = ancestorPathsOf(nodes, targetIdx);
const merged = { ...state.expanded };
for (const p of ancestorPaths) merged[p] = true;
dispatch({ type: 'SET_EXPANDED', payload: merged });
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|---------------|--------|
| Hand-rolled focus-trap via manual `keydown` Tab-cycling + `aria-hidden` on siblings | The `inert` HTML attribute, reached Baseline "Widely available" across major browsers around October 2025 per WebSearch [LOW confidence — not verified against caniuse or MDN directly this session] | ~Oct 2025 (claimed) | Not adopted this phase — `24-CONTEXT.md` already locked "hand-rolled `useModalBehavior` hook" before this research ran, and `inert` would still need to be applied to *all* content behind the scrim (the rail, toolbar, tree, details panel), a bigger blast-radius change than the CONTEXT decision anticipated. Worth a footnote for Phase 25/26/27 discuss-phase sessions, not a reason to reopen this phase's locked decision. |

**Deprecated/outdated:** Nothing in this phase's actual scope is deprecated — `@tanstack/react-virtual` v3's `scrollToIndex` API is current and unchanged from Phase 23's usage.

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | The new Go binding's exact name (`SearchIndexed`), method placement (App-layer vs. new `Service` method), and result-struct name (`SearchIndexResult`) | Architecture Patterns, Pattern 1 | Low — explicitly Claude's Discretion per `24-CONTEXT.md`; any reasonable name works as long as it wraps `SearchCatalogs` unchanged and caps in Go |
| A2 | WKWebView on macOS does not swallow or reserve Cmd+K for its own chrome (Wails apps are frameless/chromeless, unlike Safari where Cmd+K is bound to the search field) | Common Pitfalls, Pitfall 4 | Medium — if wrong, PLT-01's keyboard-open path silently fails on macOS specifically; must be verified live via dev-browser or a real build early in execution (first task), not assumed from this research alone |
| A3 | The `inert` attribute's "Baseline Widely available ~October 2025" claim, and the general framing that native `<dialog>`/`inert` are now "modern best practice" over hand-rolled traps | State of the Art | Low — purely informational footnote, does not affect this phase's locked hand-rolled-hook decision; would only matter if a future phase's discuss-phase reopens the library-vs-hand-rolled question |
| A4 | The TanStack Virtual GitHub issue #615 workaround (`setTimeout(fn, 0)` in a `useEffect`) generalizes correctly to this codebase's fixed-row-height (non-dynamic) virtualizer — i.e., that the two-effect pattern recommended in Pitfall 1 is necessary even though the *specific bug* in #615 (dynamic-mode alignment) shouldn't apply here | Common Pitfalls, Pitfall 1 | Medium — if the fixed-height virtualizer actually *does* scroll correctly from a single synchronous effect (no second effect needed), the two-effect design is extra complexity, not a correctness bug; if the assumption is wrong in the other direction (single effect is NOT sufficient), the reveal silently fails to scroll. Verify live in the first palette-reveal manual test rather than trusting either assumption at plan time. |
| A5 | The request-id-ref debounce/stale-response pattern is idiomatic and sufficient — no more specific library or React-18-specific primitive (e.g. `useTransition`, `useDeferredValue`) is better suited here | Common Pitfalls, Pitfall 5 | Low — `useDeferredValue` was explicitly rejected in `24-CONTEXT.md` for this exact use case ("that precedent does not transfer" from the rail filter) on the grounds that this path involves real I/O latency, not just render-cost deferral; the request-id-ref pattern is well-established and matches the CONTEXT's own "request id (or equivalent monotonic token)" wording |

## Open Questions

1. **Does Cmd+K actually reach the frontend's `keydown` listener inside the built macOS app (not just `wails dev`), given WKWebView's own keyboard handling?**
   - What we know: This codebase already has one working precedent for a `window`-level `keydown` listener (Escape, `WorkspaceShell.tsx:34-41`), and Phase 23's `23-VALIDATION.md` records that precedent as live-verified (`23-VERIFICATION.md`). No Wails application menu (`options.App.Menu`) is currently registered (`main.go:62-79` has no `Menu` field), so there is no OS-level accelerator competing for the key combination.
   - What's unclear: Whether Cmd+K specifically (as opposed to Escape) has any WKWebView/macOS-chrome-level reservation. WebSearch found only generic, non-Cmd+K-specific reports of "some keyboard shortcuts not being properly intercepted by the application layer" in WKWebView apps, plus a real Wails GitHub issue (#2582) about a *different* shortcut (Cmd+Ctrl+F) not being interceptable — not proof either way for Cmd+K.
   - Recommendation: Treat this as a first-task manual verification checkpoint (dev-browser against `wails dev` at :34115, matching Phase 23's established verification port — `23-VALIDATION.md:114` notes `:5173` exposes no bindings) before building the rest of the palette on top of it. If it fails, the fallback is a Wails `options.App.Menu`-registered accelerator (a bigger, unplanned change — flag back to discuss-phase if this is needed).

2. **Should the new capping logic live as an `App`-layer method (in `app.go`) or a new `Service`-layer method (in `internal/search`)?**
   - What we know: `24-CONTEXT.md` locks "GUI-only" and leaves the exact placement to Claude's Discretion. All of `internal/search`'s existing methods (`SearchCatalogs`, `BrowseCatalogs`, `LoadCatalog`, `LoadCatalogFlat`) are unit-tested via `*_test.go` files beside them; `app.go` itself has no dedicated Go test file for its thin-wrapper methods (confirmed: `app_test.go` exists at the repo root but wasn't read this session to confirm its actual coverage scope).
   - What's unclear: Whether the plan should make the capping logic a testable `Service` method (consistent with the phase's Go-testable-surface precedent from `23-VALIDATION.md`) purely for unit-test coverage of the cap-at-50/count-total logic, vs. keeping it a one-off `App` method since it's a thin, low-risk transform.
   - Recommendation: Lean toward a small `Service`-layer method (e.g. `SearchIndexed` on `*search.Service`, called by a one-line `App.SearchIndexed` binding) purely because it makes the 50-cap/total-count logic directly `go test`-able without a Wails runtime — consistent with this project's established pattern (per `23-VALIDATION.md`: "Go **is** unit-testable here... a large share of this phase's risk lands in genuinely automated tests"). Final call belongs to the planner/executor.

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| Go toolchain | Backend binding + `wails generate module` | ✓ | go.mod declares 1.23 | — |
| Wails CLI (`wails`) | `wails generate module` regeneration step | ✓ | v2.10.2 [VERIFIED: `wails version` output this session] | — |
| Node/npm (frontend build) | `tsc`/`vite build` verification | Not directly probed this session — assumed available per `23-VALIDATION.md`'s established quick-run command (`npx tsc --noEmit`) which the prior phase's audit confirmed green | — [ASSUMED, not re-probed this session] | — |
| `@tanstack/react-virtual` | Reveal scroll-into-view | ✓ | 3.14.9, already installed [VERIFIED: frontend/package.json] | — |

**Missing dependencies with no fallback:** None identified.

**Missing dependencies with fallback:** None — no new dependency is needed this phase.

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go: `go test` (table-driven, `*_test.go` beside source — established project pattern, confirmed by the existing `internal/search/*_test.go` files). Frontend: none by design (TEST-01 deferred); `tsc --noEmit` + `vite build` + live browser verification via `dev-browser` against `wails dev`, matching `23-VALIDATION.md`'s precedent. |
| Config file | `go.mod`, `frontend/tsconfig.json`, `frontend/vite.config.ts` (unchanged this phase) |
| Quick run command | `go test ./internal/search/... && (cd frontend && npx tsc --noEmit)` |
| Full suite command | `cd frontend && npx tsc --noEmit && npm run build && cd .. && go build ./... && go test ./... -race -count=1` |

### Phase Requirements → Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| PLT-02, PLT-03 | New capping binding returns ≤50 results plus the true total, for both an under-50-match and an over-50-match fixture | Go unit | `go test ./internal/search/... -run TestSearchIndexed` (name illustrative — exact test name is executor's choice) | ❌ Wave 0 — no test file for this new surface exists yet |
| PLT-02 | `cli/search.go`'s existing behavior is provably unchanged (same output for the same fixture, before/after this phase's Go changes) | Go unit (regression) | `go test ./cli/... -run TestSearch` | Existing (`cli/search_test.go`) — extend, don't replace |
| PLT-01, PLT-04, PLT-05, PLT-06, PLT-07 | Palette open (click + ⌘K), keyboard nav, click-to-reveal (multi-branch-open scenario per Pitfall 2), Escape-close, focus-trap/scroll-lock, empty-state copy | Live browser (dev-browser) | Manual dev-browser session against `wails dev` at `:34115` (Phase 23's established port — `:5173` exposes no `window.go` bindings per `23-VALIDATION.md:114`) | N/A — no frontend test framework this project (TEST-01 deferred) |
| PLT-05 (regression-shaped) | Revealing a node in one branch does not collapse unrelated already-open branches (Pitfall 2's specific failure mode) | Live browser (dev-browser), scripted setup | Manual: pre-expand 2-3 unrelated branches, trigger reveal to a node in a different branch, assert the pre-expanded branches remain expanded (DOM inspection of `.ws-tree-row[aria-expanded]` or the `expanded` state via a debug hook) | N/A |

### Sampling Rate
- **Per task commit:** `go test ./internal/search/...` for the new Go binding/capping logic; `npx tsc --noEmit` for every frontend task
- **Per wave merge:** Full suite (`go build ./... && go test ./... -race -count=1 && npx tsc --noEmit && npm run build`)
- **Phase gate:** Full suite green, plus a live dev-browser pass covering PLT-01/04/05/06/07 and the Pitfall-2 multi-branch-open regression check, before `/gsd-verify-work`

### Wave 0 Gaps
- [ ] A Go test file for the new capping binding/method (e.g. `internal/search/search_indexed_test.go` or similar) — covers PLT-02/PLT-03's cap-at-50 and total-count behavior, including the boundary case (exactly 50 matches, 51 matches, 0 matches)
- [ ] No new frontend test infrastructure needed or expected (TEST-01 remains deferred; do not add Vitest/Testing Library this phase — that would be scope creep against a locked deferral)

*Everything else uses existing test infrastructure (`go test`, `tsc --noEmit`, `dev-browser`) already established in Phase 22/23.*

## Security Domain

`security_enforcement` is absent from `.planning/config.json`, so it is treated as enabled.

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V2 Authentication | No | Single-user local desktop app, no auth surface touched this phase |
| V3 Session Management | No | No session concept in this app |
| V4 Access Control | No | No new access-control boundary — the new binding reads the same already-configured `catalogDir` every other binding reads |
| V5 Input Validation | Yes, narrowly | The search term is user-typed free text passed to `strings.Contains`/`strings.ToLower` (`service.go:72,121`) — no shell/SQL/path interpolation risk (it's a pure in-memory substring match, never used to construct a file path or command). No new validation needed beyond what `SearchCatalogs` already does; do not add path-traversal logic to the new binding since it takes the identical `catalogDir` parameter every existing binding already takes and abs-resolves (`app.go:85-88` pattern) |
| V6 Cryptography | No | No cryptographic operation in this phase |

### Known Threat Patterns for this stack

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| Malicious/oversized search term causing excessive backend work (a very long or regex-metacharacter-laden query string against a large catalog set) | Denial of Service | `strings.Contains` is a linear substring scan, not a regex — no ReDoS surface. The existing `SearchCatalogs` walk is already bounded by the number of catalogs/nodes on disk, unchanged by this phase's capping (which only limits the *response* size, not the walk cost) — this is an accepted, pre-existing characteristic of `SearchCatalogs`, not a new risk introduced this phase |
| A reveal-to-tree deep-link (`pendingReveal`) racing a directory rename/delete between search-time and click-time | Tampering (data integrity, not security-boundary) | Already covered by the E5 UI-SPEC resolution: "the pending reveal is discarded rather than retried" if the target catalog fails to load or yields zero matching nodes by the time the reveal is processed — no new Go-side validation needed, this is a frontend state-discard concern already designed for |
| XSS via `dangerouslySetInnerHTML`-style rendering of the match-highlight substring | Tampering/Injection | The UI-SPEC specifies rendering the highlighted substring as a React `<span>` with inline `style`, built from `basename.slice()` operations on already-known-safe string data (catalog file/directory names from the local filesystem) — standard JSX text-child rendering auto-escapes; do **not** use `dangerouslySetInnerHTML` to inject the highlight span as raw HTML string concatenation, since a filename containing `<`/`>`/`&` would otherwise render as unescaped markup |

## Sources

### Primary (HIGH confidence — in-repo, read directly this session)
- `internal/search/service.go` — full file read; `SearchCatalogs`/`searchInCatalogFile`/`searchInCatalog` implementation
- `internal/search/flatten.go` — full file read; root-exclusion confirmed at lines 19-20, 63-67
- `pkg/models/catalog.go` — full file read; `SearchResult`, `FlatNode`, `FlatCatalog` struct shapes
- `app.go` — full file read; existing binding registration pattern, no `Menu` option registered
- `main.go` — full file read; `wails.Run` options, confirms no `options.App.Menu`
- `frontend/src/contexts/AppContext.tsx` — full file read; `SET_EXPANDED` replace-not-merge confirmed at lines 123-126, `TOGGLE_EXPAND` at 112-122, `TREE_LOADED` stale-discard pattern at 95-99
- `frontend/src/components/workspace/TreePane.tsx` — full file read; virtualizer instantiation (fixed `estimateSize`, no `measureElement`), existing `useEffect`/`useLayoutEffect` patterns
- `frontend/src/hooks/useVisibleRows.ts` — full file read
- `frontend/src/hooks/useMediaQuery.ts` — full file read (hooks-folder convention precedent)
- `frontend/src/components/workspace/Toolbar.tsx` — full file read; `.ws-search` button has no `onClick`
- `frontend/src/components/workspace/WorkspaceShell.tsx` — full file read; existing Escape-listener + backdrop-click precedent (not yet a shared hook)
- `frontend/src/components/workspace/StatusBar.tsx` — full file read; `totalIndexedFiles`/`catalogCount` derivation for the placeholder-text formula
- `frontend/src/services/wailsAPI.ts` — full file read; `extractErrorMessage` wrapper pattern every new binding call must follow
- `cli/search.go` — full file read; confirms direct, unwrapped call to `svc.SearchCatalogs`
- `frontend/src/lib/format.ts` — full file read; `formatBytes`/`formatCount`/`formatGB`/`formatDate`
- `frontend/src/workspace.css` — z-index custom properties (`--z-overlay: 200` at line 30) and `.ws-chip`/`.ws-dir-chip` hover-only rule confirmed at lines 71-82
- `frontend/wailsjs/go/main/App.d.ts` / `App.js` / `frontend/wailsjs/go/models.ts` — confirmed the generated-binding shape and pattern (no `AbortSignal` parameter on any binding)
- `go.mod` — Go 1.23, Wails v2.10.2 confirmed
- `wails version` (CLI command run this session) — v2.10.2 confirmed
- `.planning/phases/23-rail-virtualized-tree/23-VALIDATION.md` — full file read; test-infrastructure precedent (`go test`, no frontend framework, dev-browser at `:34115`)
- `.planning/phases/24-cmd-k-command-palette/24-CONTEXT.md`, `24-UI-SPEC.md`, `.planning/REQUIREMENTS.md`, `.planning/STATE.md` — full files read per task instructions

### Secondary (LOW confidence per this project's `classify-confidence` seam, but fetched directly off the vendor's own docs page via WebFetch — treat as more reliable than the Tertiary/WebSearch-only items below despite the identical tier label)
- [TanStack Virtual v3 Virtualizer API docs](https://tanstack.com/virtual/v3/docs/api/virtualizer) — `scrollToIndex` TypeScript signature (`index`, `{ align, behavior }`) confirmed
- [TanStack/virtual GitHub Issue #615](https://github.com/TanStack/virtual/issues/615) — dynamic-mode alignment bug and the maintainer-suggested `setTimeout(fn, 0)` deferral workaround

### Tertiary (LOW confidence — WebSearch only, not independently verified against an authoritative source this session)
- WKWebView/Wails Cmd+K keyboard-interception question — no conclusive source found; see Open Question 1
- `inert` attribute "Baseline Widely available ~October 2025" and native `<dialog>` framing as 2025-2026 "modern best practice" for modals — informational only, not adopted this phase
- Request-id-ref debounce/stale-response-guard pattern description — a widely-documented, low-risk React pattern; not project-specific, no single authoritative source fetched

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — no new dependencies, every library/version claim verified against `package.json`/`go.mod`/CLI output directly this session
- Architecture: HIGH for in-repo structural claims (reducer semantics, root exclusion, existing component wiring — all read and quoted verbatim with line numbers); LOW (per the `classify-confidence` seam) for the two externally-sourced virtualizer-timing claims, though they were read directly off TanStack's own docs/issue tracker rather than inferred; LOW for the WKWebView Cmd+K question, explicitly flagged as an Open Question requiring first-task manual verification rather than being planned around as settled fact
- Pitfalls: HIGH for Pitfalls 1-3 (grounded entirely in in-repo code read this session); LOW for Pitfalls 4-5 (WebSearch/WebFetch-sourced, explicitly flagged)

**Research date:** 2026-08-14
**Valid until:** 30 days (stable in-repo findings) / re-verify the WKWebView Cmd+K question and the TanStack Virtual dynamic-mode-vs-fixed-mode distinction empirically during the phase's first execution task, regardless of this date — those two items are marked LOW confidence precisely because they were not verified against running code this session
