# Phase 24: Cmd-K Command Palette - Context

**Gathered:** 2026-08-14
**Status:** Ready for planning
**Mode:** Smart discuss (autonomous) — all three grey areas accepted as recommended

<domain>
## Phase Boundary

Users find any file across every catalog instantly from a ⌘K palette without leaving the workspace.

**In scope:** the palette overlay itself (open/close, focus, scrim, geometry), cross-catalog search execution behind it, the 50-result cap with a truncation notice, keyboard and click navigation of results, deep-navigation from a hit into the Phase 23 tree (switch catalog → expand ancestors → select → scroll into view), and the shared modal-behavior hook (focus trap + Escape + scroll lock) that Phases 25–27's overlays must reuse rather than reimplement.

**Out of scope (later phases):** create slide-over (Phase 25), Settings surface including the catalog-directory picker's real UI (Phase 26), catalog actions and fsnotify watch (Phase 27), re-scan and diff (Phase 28). This palette searches files inside catalogs only — it is not a general command runner, and it does not gain non-search "commands" in this phase.

</domain>

<decisions>
## Implementation Decisions

### Search Execution & Backend Surface
- **A new thin GUI-only Go binding** wraps the existing `search.Service` walk rather than modifying `SearchCatalogs`. `cli/search.go` calls `Service.SearchCatalogs` and the milestone's locked decision is that new capabilities are GUI-only — so the existing method and its CLI caller stay byte-identical in behavior.
- **The 50-result cap and the `N` total are both computed in Go.** The service already walks every node in every catalog, so producing `total` is free, and only 50 results ever marshal across the Wails bridge. This is deliberate: Phase 23 measured **5.83 MB of JSON for a single 42,551-node catalog**; an uncapped cross-library query would be that figure multiplied by the catalog count, on every keystroke. Frontend `.slice(0, 50)` was rejected for exactly that reason. Return shape carries results, the total hit count, and enough information to render "Showing the first 50 of N hits" (PLT-03).
- **Live search, debounced ~200 ms, with a stale-response guard.** Each query carries a request id (or equivalent monotonic token); a response whose id is not the latest is dropped rather than rendered. Phase 23's rail filter used bare `useDeferredValue` with no debounce because it filters an already-loaded in-memory array — this path is disk I/O plus a full JSON parse per catalog per keystroke, so that precedent does not transfer.
- **Minimum query length is 2 characters.** Below 2, the palette shows a "Type to search…" hint — this is a distinct state from the empty-result copy, so PLT-06's exact string ("No file in any catalog matches that.") is never shown for a query that was never run.

### Palette UI & Result Presentation
- **Centered dialog anchored ~15vh from the top, max-width 640px, with a scrim behind it**, at the Phase 22 locked z-index scale position for the palette: **200** (`details drawer 100 → create slide-over / ⌘K palette 200 → dialogs / Settings 300`). Do not invent a new z-index.
- **Each result row renders: basename (with the match highlighted) + dimmed full path + a catalog-name chip + size.** Reuse `frontend/src/lib/format.ts` for the size and the existing `.ws-chip` token/class for the catalog chip rather than new styling.
- **No grouping — a flat list in backend order** (catalog, then path). The per-row catalog chip already identifies each hit's origin, and sticky group headers would spend vertical space inside a hard 50-row cap.
- **Highlight the matched substring in the basename only.** The path stays plain-dimmed. Matching is case-insensitive on both name and path (PLT-02 searches both), but only the basename gets visual highlight treatment.

### Navigation to a Hit & the Shared Modal Hook
- **A hit reaches the tree through new `pendingReveal: string | null` state in `AppContext`**, holding the target node's path. The palette dispatches the catalog switch plus the reveal request; `TreePane` consumes `pendingReveal` after `TREE_LOADED` lands and clears it. This is chosen over an imperative ref or a custom DOM event specifically because the catalog load is asynchronous — the reveal must survive the gap between "switch catalog" and "flat nodes arrived", which a synchronous imperative call cannot.
- **Reveal scrolls with `virtualizer.scrollToIndex(visibleIdx, { align: 'center' })`**, where `visibleIdx` is the index within the `useVisibleRows` output (not the raw flat-node index) — ancestors must be expanded first so the target is actually present in the visible slice.
- **Directory hits expand every ancestor and select the directory, but do not auto-expand the target itself.** PLT-05 specifies "expand every ancestor"; selection is not expansion, and auto-expanding a large directory on arrival would dump thousands of rows the user did not ask for.
- **PLT-07's shared modal behavior is a hand-rolled `useModalBehavior({ isOpen, onClose })` hook** in `frontend/src/hooks/`, covering focus trap, Escape-to-close, page scroll lock, and focus restore to the trigger element on close. No new dependency (`focus-trap-react` / `react-focus-lock` rejected — roughly 60 lines of hook against a package the project would carry for one overlay family). **Phases 25, 26, and 27 import this hook; they must not reimplement any of these four behaviors.** Its API therefore has to be general enough for a slide-over and a dialog, not palette-shaped.

### Claude's Discretion
- The exact name, signature, and return-struct shape of the new search binding, and the name of its Go-side result type.
- Component decomposition inside the palette (input, result list, result row, empty/hint/truncation states).
- The precise debounce constant, and the mechanism used for the stale-response guard.
- Keyboard specifics beyond the requirement: Up/Down/Enter/Escape are mandatory; Home/End/PageUp/PageDown wrap-around behavior is open.
- How the match-highlight substring is computed and rendered.
- Whether `pendingReveal` is one reducer action or a small pair, and how it is cleared.

</decisions>

<code_context>
## Existing Code Insights

### Reusable Assets
- `internal/search/service.go` — `Service.SearchCatalogs(searchTerm, catalogDirectory)` already performs the full cross-catalog walk (`searchInCatalogFile` → recursive `searchInCatalog`), matching case-insensitively against `item.Name`, which holds a **full relative path, not a basename** (corrected during Phase 23). `pkg/models.SearchResult` already carries `Catalog`, `CatalogFilePath`, `Basename`, `FullPath`, `FullName`, `Type`, `Size`.
- `app.go:84` — `SearchCatalogs` is already Wails-bound; the frontend already has a generated binding and `frontend/src/services/wailsAPI.ts` already wraps calls with `extractErrorMessage()` (Wails rejects with a plain string, not an `Error` — Phase 23 fix, applies to any new binding too).
- `frontend/src/components/workspace/Toolbar.tsx` — the search field is already a `<button type="button" className="no-drag ws-search" aria-label="Search every catalog">` rendering the placeholder text and a `⌘K` hint chip. It has **no `onClick`**; wiring it is the whole of PLT-01's click path.
- `frontend/src/hooks/useVisibleRows.ts` — pure `useMemo` deriving DFS-ordered visible indices from `nodes` + `expanded`. The reveal path needs its output to map a node path to a visible row index.
- `frontend/src/lib/format.ts` — size/number formatting already used by the rail and details panel.
- `frontend/src/themeTokens.ts` — exports `safeGetItem` / `safeSetItem` (made non-private in Phase 23) if the palette needs any persistence.

### Established Patterns
- Global state is a `useReducer` + Context in `frontend/src/contexts/AppContext.tsx`; `AppState` currently holds `density`, `railSide`, `detailOverlay`, `catalogDir`, `catalogs`, `currentCatalogId`, `tree`, `expanded`, `selected`. Actions are `UPPER_SNAKE_CASE`. Catalog switching is deliberately **one atomic action** that clears `expanded`/`selected`/scroll (Phase 23, TREE-06) — the reveal must not reintroduce a half-applied state.
- `TreePane.tsx` uses `@tanstack/react-virtual`'s `useVirtualizer` with a **fixed row height read from the `--rh` token** (27px Compact / 34px Comfortable). It already calls `virtualizer.scrollToOffset(0)` on catalog switch.
- Styling is CSS custom properties in `workspace.css` plus inline style objects; Ant Design was removed this milestone. Phase 22 declared the z-index scale as CSS vars in `workspace.css`.
- Transient/local UI state stays in component `useState` and out of `AppContext` when the tree has nothing to re-render against (Phase 23's rail filter precedent) — the palette's query string follows this; only `pendingReveal` crosses into global state.

### Integration Points
- **Toolbar search button** → palette open (PLT-01, click path).
- **A global keydown listener** for ⌘K / Ctrl+K → palette open (PLT-01, keyboard path). Must not fire while typing in the rail filter or another input where the shortcut would be swallowed.
- **`AppContext` reducer** → new `pendingReveal` state plus its action(s).
- **`TreePane`** → consumes `pendingReveal` post-`TREE_LOADED`, expands ancestors, dispatches selection, calls `scrollToIndex`, clears the request.
- **`app.go`** → registers the new search binding; `wails generate module` output under `frontend/wailsjs/` must be regenerated.
- **`workspace.css`** → palette styles slot into the existing token/z-index scale.

</code_context>

<specifics>
## Specific Ideas

- PLT-06's empty-state copy is exact: **"No file in any catalog matches that."**
- PLT-03's truncation copy follows the milestone handoff's wording: **"Showing the first 50 of N hits"**, shown only when the total exceeds 50.
- `PROJECT.md` states the milestone intent plainly: "⌘K search palette across all catalogs, capped at 50 with 'first 50 of N'" — the cap is a product decision, not a performance workaround, so it holds even if the payload were free.
- The shared modal hook is a **contract with three later phases**, not a convenience. Phase 25 (slide-over with five distinct close paths and an animated exit), Phase 26 (Settings), and Phase 27 (confirmation dialogs) all depend on it. Design its API for those consumers now; a palette-only hook will have to be rewritten in Phase 25.

</specifics>

<deferred>
## Deferred Ideas

- Non-search commands in the palette (run an action, jump to a setting) — the requirement set is file-search only; a general command registry belongs in its own phase if ever wanted.
- Fuzzy/subsequence matching and result ranking — PLT-02 specifies substring matching over names and paths; ranking is not a requirement and would need its own decisions.
- Search history / recent queries persistence.
- Frontend unit tests for palette keyboard navigation — already an explicitly deferred milestone item (TEST-01, deferred at v3.0.0 requirements definition).

</deferred>
