# Phase 24: Cmd-K Command Palette - Pattern Map

**Mapped:** 2026-08-14
**Files analyzed:** 9 (new + modified)
**Analogs found:** 9 / 9

## File Classification

| New/Modified File | Role | Data Flow | Closest Analog | Match Quality |
|--------------------|------|-----------|-----------------|----------------|
| `internal/search/search_indexed.go` (new `SearchIndexed` method on `*search.Service`) | service | CRUD (capped read) | `internal/search/flatten.go` (`LoadCatalogFlat`) | exact — same "thin GUI-shaping wrapper beside a CLI-shared method" pattern |
| `internal/search/search_indexed_test.go` (new) | test | CRUD | `internal/search/flatten_test.go` + `internal/search/service_test.go` | exact — table-driven, `t.TempDir()` fixture convention |
| `app.go` — new `SearchIndexed` binding method | controller (Wails binding) | request-response | `app.go:112-119` (`LoadCatalogFlat` binding) and `app.go:83-91` (`SearchCatalogs` binding) | exact |
| `frontend/wailsjs/go/**` (regenerated) | config/generated | request-response | itself (regenerate via `wails generate module`, do not hand-edit) | exact — mechanical |
| `frontend/src/services/wailsAPI.ts` — new `searchIndexed` wrapper | service (IPC wrapper) | request-response | `wailsAPI.ts`'s `loadCatalogFlat` / `searchCatalogs` entries | exact |
| `frontend/src/hooks/useModalBehavior.ts` (new) | hook | event-driven | `frontend/src/hooks/useMediaQuery.ts` (file layout/JSDoc) + `WorkspaceShell.tsx`'s inline Escape/backdrop-close logic (behavior to extract) | role-match (hook shape) / exact (behavior source) |
| `frontend/src/components/workspace/CommandPalette.tsx` (+ optional `palette/` subcomponents) | component | event-driven + request-response | `frontend/src/components/workspace/TreePane.tsx` (multi-state render, virtualizer-adjacent patterns) and `CatalogRail.tsx` (filter input debounce-adjacent, list rendering) | role-match |
| `frontend/src/contexts/AppContext.tsx` — add `pendingReveal` state/action | store (reducer) | event-driven | existing `SELECT_CATALOG` / `SET_EXPANDED` / `TOGGLE_EXPAND` / `TREE_LOADED` cases in the same file | exact — same reducer, same file |
| `frontend/src/components/workspace/TreePane.tsx` — consume `pendingReveal` | component | event-driven | its own existing `useEffect`/`useLayoutEffect` pairs (catalog-load effect, scroll-reset effect) | exact — same file, same idiom |
| `frontend/src/components/workspace/Toolbar.tsx` — wire `.ws-search` onClick | component | event-driven | itself (existing button markup, just needs `onClick`) | exact |
| `frontend/src/workspace.css` — palette styles | config (styles) | n/a | existing z-index vars (`--z-overlay`), `.ws-chip`/`.ws-dir-chip` sibling-class pattern, `.ws-backdrop` | exact |

## Pattern Assignments

### `internal/search/search_indexed.go` (service, CRUD)

**Analog:** `internal/search/flatten.go` (`LoadCatalogFlat`), full file read.

**Doc-comment + wrapping pattern** (`flatten.go:14-20`):
```go
// LoadCatalogFlat reads and parses a catalog JSON file via the unmodified
// LoadCatalog (reusing its dual-format v1/v2 parse -- this function performs
// no JSON decoding of its own), then flattens the nested tree into a single
// render-ready slice. The root itself is excluded from the returned Nodes;
// the root's direct children get Depth 0 and ParentIdx -1.
func (s *Service) LoadCatalogFlat(filePath string) (*models.FlatCatalog, error) {
	root, err := s.LoadCatalog(filePath)
	if err != nil {
		return nil, fmt.Errorf("load catalog for flatten: %w", err)
	}
	...
```

**The method to wrap, unchanged** (`internal/search/service.go:64-92`):
```go
// SearchCatalogs searches all JSON catalogs in the specified directory for the search term
func (s *Service) SearchCatalogs(searchTerm, catalogDirectory string) ([]*models.SearchResult, error) {
	var results []*models.SearchResult

	entries, err := os.ReadDir(catalogDirectory)
	if err != nil {
		return nil, fmt.Errorf("failed to read catalog directory: %w", err)
	}

	searchTermLower := strings.ToLower(searchTerm)

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		filePath := filepath.Join(catalogDirectory, entry.Name())
		catalogName := strings.TrimSuffix(entry.Name(), ".json")

		matches, err := s.searchInCatalogFile(filePath, catalogName, searchTermLower)
		if err != nil {
			// Skip files that can't be read or parsed
			continue
		}

		results = append(results, matches...)
	}

	return results, nil
}
```

**Recommended shape** (Claude's discretion, per RESEARCH.md's own sketch, adjusted to live on `*Service` for testability, per RESEARCH's Open Question 2 recommendation):
```go
// SearchIndexed wraps SearchCatalogs with a GUI-only cap: it performs the
// identical full walk, then returns at most 50 results plus the true total
// count. cli/search.go is untouched -- it calls SearchCatalogs directly
// (cli/search.go:61-62), not this method.
func (s *Service) SearchIndexed(searchTerm, catalogDirectory string) (*models.SearchIndexResult, error) {
	all, err := s.SearchCatalogs(searchTerm, catalogDirectory)
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

**Error handling pattern:** identical to every other Service method — return `(nil, fmt.Errorf(...))` on hard failure, never a partial/zero-value result silently swallowed. `SearchCatalogs` itself already tolerates per-file parse/read errors by `continue`-skipping (`service.go:80-84`) — `SearchIndexed` does not need to duplicate that, since it wraps the whole call.

---

### `app.go` — new `SearchIndexed` binding

**Analog:** `app.go:112-119` (`LoadCatalogFlat` binding, most recent addition) and `app.go:83-91` (`SearchCatalogs` binding).

**Exact pattern to copy** (`app.go:112-119`):
```go
// LoadCatalogFlat reads and parses a catalog JSON file, returning it as a
// single flattened node array ready for the virtualized tree pane.
func (a *App) LoadCatalogFlat(filePath string) (*models.FlatCatalog, error) {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, err
	}
	return a.searchService.LoadCatalogFlat(absPath)
}
```

New binding follows the identical shape:
```go
// SearchIndexed returns a capped (max 50) cross-catalog search result set
// plus the true total hit count, for the GUI palette only.
func (a *App) SearchIndexed(searchTerm string, catalogDir string) (*models.SearchIndexResult, error) {
	absPath, err := filepath.Abs(catalogDir)
	if err != nil {
		return nil, err
	}
	return a.searchService.SearchIndexed(searchTerm, absPath)
}
```

**Struct location:** add `SearchIndexResult` to `pkg/models/catalog.go`, sibling to `SearchResult` (per RESEARCH.md's `Recommended Project Structure`).

**Regeneration:** run `wails generate module` after adding the binding — this mechanically produces `frontend/wailsjs/go/main/App.d.ts`/`App.js` and updates `frontend/wailsjs/go/models.ts`. Do not hand-edit generated output.

---

### `internal/search/search_indexed_test.go` (test)

**Analog:** `internal/search/flatten_test.go` (fixture-writing helper + table walk) and `internal/search/service_test.go` (temp-dir catalog fixture pattern).

**Fixture-writing helper pattern** (`flatten_test.go:15-21`):
```go
func writeFlattenTestCatalog(t *testing.T, dir, name, content string) string {
	t.Helper()
	filePath := filepath.Join(dir, name)
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test catalog: %v", err)
	}
	return filePath
}
```

**Temp-dir single-catalog fixture pattern** (`service_test.go:12-31`):
```go
func writeTestCatalog(t *testing.T) (dir string, filePath string, fileSize int64) {
	t.Helper()
	dir = t.TempDir()
	content := []byte(`{"type":"directory","name":"./","size":0,"contents":[]}`)
	filePath = filepath.Join(dir, "test-catalog.json")
	if err := os.WriteFile(filePath, content, 0644); err != nil {
		t.Fatalf("failed to write test catalog: %v", err)
	}
	info, err := os.Stat(filePath)
	...
}
```

**Coverage needed (per RESEARCH.md Wave 0 Gaps):** boundary cases — exactly 50 matches (no truncation, `total == 50`), 51 matches (`total == 51`, `len(Results) == 50`), 0 matches (`total == 0`, `Results` empty/nil) — plus a regression test asserting `cli/search.go`'s direct `SearchCatalogs` call is unaffected (extend `cli/search_test.go`, don't replace).

---

### `frontend/src/services/wailsAPI.ts` — new `searchIndexed` wrapper

**Analog:** `loadCatalogFlat` and `searchCatalogs` entries in the same file.

**Exact pattern** (`wailsAPI.ts:87-93`):
```typescript
loadCatalogFlat: async (filePath: string) => {
  try {
    const flat = await LoadCatalogFlat(filePath);
    return { success: true as const, flat };
  } catch (error: any) {
    return wailsError(error);
  }
},
```

New wrapper follows identically:
```typescript
searchIndexed: async (searchTerm: string, catalogDir: string) => {
  try {
    const indexed = await SearchIndexed(searchTerm, catalogDir);
    return { success: true as const, indexed };
  } catch (error: any) {
    return wailsError(error);
  }
},
```

**Import block to extend** (`wailsAPI.ts:1-19`) — add `SearchIndexed` to the destructured import from `'../../wailsjs/go/main/App'`, alongside the existing `SearchCatalogs`, `LoadCatalogFlat`, etc. Every binding call routes through `wailsError()`/`extractErrorMessage()` (`wailsAPI.ts:26-39`) — Wails rejects with a plain string, not an `Error`, so `error?.message` alone is always `undefined`; this is the mandatory shared pattern for the new call.

---

### `frontend/src/hooks/useModalBehavior.ts` (hook, event-driven)

**Analog (file layout/JSDoc density):** `frontend/src/hooks/useMediaQuery.ts`, full file:
```typescript
import { useEffect, useState } from 'react';

/**
 * Subscribes to a media query via matchMedia's change event -- fires only on
 * threshold crossings, never on every pixel of a resize like a window
 * 'resize' listener would.
 */
export function useMediaQuery(query: string): boolean {
  const [matches, setMatches] = useState(() => window.matchMedia(query).matches);

  useEffect(() => {
    const mql = window.matchMedia(query);
    setMatches(mql.matches);
    const handleChange = (event: MediaQueryListEvent) => {
      setMatches(event.matches);
    };
    mql.addEventListener('change', handleChange);
    return () => {
      mql.removeEventListener('change', handleChange);
    };
  }, [query]);

  return matches;
}
```
Pattern to copy: named export, one exported function, a top-of-file JSDoc block explaining *why* (not *what*), `useEffect` cleanup returning a matching `removeEventListener`.

**Analog (Escape/backdrop-close behavior to extract into the hook):** `frontend/src/components/workspace/WorkspaceShell.tsx:30-41` and `:66`, VERBATIM — this is the exact bespoke logic that needs promoting into a hook:
```typescript
// One close path for both Escape and backdrop click.
const closeDrawer = () => dispatch({ type: 'SET_DETAIL_OVERLAY', payload: false });

// Escape closes the drawer; listener only lives while the drawer is open.
useEffect(() => {
  if (!state.detailOverlay) return;
  const onKeyDown = (e: KeyboardEvent) => {
    if (e.key === 'Escape') closeDrawer();
  };
  window.addEventListener('keydown', onKeyDown);
  return () => window.removeEventListener('keydown', onKeyDown);
}, [state.detailOverlay]);
...
<div className="ws-backdrop" onClick={closeDrawer} />
```
This confirms `window`-level `keydown` listeners with a guard on the open flag do work in this Wails webview (per RESEARCH.md's Open Question 1 reasoning) — copy the guard-while-open / cleanup-on-close idiom directly. The hook additionally needs focus-trap, scroll-lock, and focus-restore, none of which have an in-repo precedent (RESEARCH.md: "the current details drawer close-on-Escape/backdrop-click ... is bespoke, not a hook — this phase is the first to extract the pattern") — no analog exists for those three; UI-SPEC's "Shared Modal Behavior Hook Contract" section is the source of truth for their exact required semantics (release scroll lock and restore focus on the `isOpen: false` transition, not on unmount — needed for Phase 25's animated exit).

---

### `frontend/src/components/workspace/CommandPalette.tsx` (+ optional palette/ subcomponents)

**Analog (multi-branch state rendering, five mutually-exclusive states):** `TreePane.tsx`, full file. Copy its "one function, N early `return`s for mutually exclusive states" shape — TreePane has 5 states (empty-library, unreadable, loading, empty-catalog, rows); the palette needs 4 (hint, searching, results, no-matches). Excerpt (`TreePane.tsx:190-224`):
```tsx
if (state.tree.status === 'loading') {
  return (
    <div className="ws-tree">
      <div className="pane-scroll" style={{ display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
        <span className="mono" style={{ fontSize: 11.5, color: 'var(--dm)' }}>
          Reading {selectedCatalog?.filename ?? 'catalog'}…
        </span>
      </div>
    </div>
  );
}
```
This is the direct model for the palette's "Searching…" state (same quiet-text-no-spinner idiom, project-wide rule).

**Analog (virtualizer usage — palette result list is NOT virtualized, but the reveal's scroll call reuses TreePane's own instance):** `TreePane.tsx:26-31`:
```tsx
const virtualizer = useVirtualizer({
  count: visibleIndices.length,
  getScrollElement: () => scrollRef.current,
  estimateSize: () => rowHeight,
  overscan: 10,
});
```
The palette itself renders a plain unvirtualized `<div>`/`.map()` for its ≤50 rows (per RESEARCH's "Don't Hand-Roll" table); this excerpt is relevant only because `TreePane`'s existing `virtualizer` object is the one the reveal effect calls `.scrollToIndex()` on — the palette component itself never constructs a virtualizer.

**Analog (filter input + local-state-only query, not AppContext):** `CatalogRail.tsx:9-19`:
```tsx
// The filter string lives only here, never in AppContext, read through
// useDeferredValue for the matching pass. This isolation -- not a
// debounce timer -- is the entire mechanism by which the tree survives
// typing untouched (RAIL-02, locked in 23-CONTEXT.md).
const [filterInput, setFilterInput] = useState('');
const deferredFilter = useDeferredValue(filterInput);
const trimmedFilter = deferredFilter.trim();
```
Note: CONTEXT.md explicitly rejects `useDeferredValue` for the palette's query (RESEARCH.md: "that precedent does not transfer" — this path is disk I/O, needs a real ~200ms debounce + request-id guard, not just render deferral) — copy only the *local-state, not-in-AppContext* placement of the query string from this file, not its `useDeferredValue` mechanism.

**Component decomposition precedent** (RESEARCH.md's own citation): `TreePane.tsx:7-9`'s composition of `BreadcrumbBar` + `TreeHeader` + `UnreadableCatalogPanel` as separate files, each one rendered concern — same convention applies if splitting into `PaletteInput.tsx` / `PaletteResultList.tsx` / `PaletteResultRow.tsx`.

---

### `frontend/src/contexts/AppContext.tsx` — `pendingReveal` state + action

**Analog:** the file's own existing reducer cases.

**State/action shape to extend** (`AppContext.tsx:17-25` and `28-40`):
```typescript
export interface AppState {
  ...
  expanded: Record<string, boolean>;
  selected: string | null;
  // ADD: pendingReveal: string | null;
}

type AppAction =
  | ...
  | { type: 'TOGGLE_EXPAND'; payload: string }
  | { type: 'SET_EXPANDED'; payload: Record<string, boolean> }
  | { type: 'SET_SELECTED'; payload: string | null };
  // ADD: | { type: 'SET_PENDING_REVEAL'; payload: string | null }
```

**CRITICAL — `SET_EXPANDED` replace semantics, VERBATIM current code** (`AppContext.tsx:123-126`):
```typescript
case 'SET_EXPANDED':
  // Replaces the whole map in one state update -- expand-all and
  // collapse-to-root both use this, never a per-node dispatch loop.
  return { ...state, expanded: action.payload };
```
The reveal's dispatch site (in `TreePane`, not the reducer) MUST merge before dispatching: `dispatch({ type: 'SET_EXPANDED', payload: { ...state.expanded, ...ancestorMap } })`. Do not add a reducer-side merge case unless a second caller needs it — CONTEXT.md/RESEARCH.md both prefer the smaller, caller-side-merge diff.

**Stale-discard pattern to mirror for `pendingReveal` consumption** (`AppContext.tsx:95-99`):
```typescript
case 'TREE_LOADED':
  // A load that resolves after the user has already selected a
  // different catalog is discarded -- only applied when the id it was
  // issued for still matches current state (RAIL-03, TREE-06).
  if (action.payload.catalogId !== state.currentCatalogId) return state;
  ...
```
`pendingReveal` consumption in `TreePane` must apply the same discipline (RESEARCH.md Pattern 2): a reveal targeting catalog A must not fire if the user has since switched to catalog B before A's `TREE_LOADED` lands.

**Atomic multi-field dispatch precedent** (`AppContext.tsx:88-93`, `SELECT_CATALOG`):
```typescript
case 'SELECT_CATALOG':
  // Atomic: sets currentCatalogId, starts the tree load, and clears
  // expanded/selected together -- no intermediate state where one is
  // cleared and the other is not (TREE-06).
  return {
    ...state,
    currentCatalogId: action.payload,
    tree: { status: 'loading' },
    expanded: {},
    selected: null,
  };
```
Relevant if the palette dispatches catalog-switch + reveal-request together — model the multi-field update the same atomic way, not as two separate dispatches racing each other.

---

### `frontend/src/components/workspace/TreePane.tsx` — consume `pendingReveal`

**Analog:** its own existing `useEffect` pair (catalog-load effect + scroll-reset `useLayoutEffect`), full file already read above.

**Effect-pairing pattern to copy** (`TreePane.tsx:38-58` load effect, `:60-72` dependent scroll-reset effect):
```tsx
useEffect(() => {
  const catalogId = state.currentCatalogId;
  if (!catalogId) return;
  wailsAPI.loadCatalogFlat(catalogId).then((result) => {
    if (result.success) {
      dispatch({ type: 'TREE_LOADED', payload: { catalogId, nodes: result.flat.nodes, ... } });
    } else {
      dispatch({ type: 'TREE_FAILED', payload: { catalogId, message: result.error } });
    }
  });
}, [state.currentCatalogId]);

useLayoutEffect(() => {
  if (state.tree.status !== 'ready') return;
  if (scrollRef.current) scrollRef.current.scrollTop = 0;
  virtualizer.scrollToOffset(0);
}, [state.currentCatalogId, state.tree.status]);
```
Model the reveal as the SAME two-effect split RESEARCH.md's Pitfall 1 mandates: one effect (keyed on `[pendingReveal, state.tree.status]`) does the ancestor-merge-expand + select + clear-request; a SECOND effect (keyed on `[pendingReveal-was-just-cleared-ish, visibleIndices]`, per RESEARCH's Code Examples section) calls `virtualizer.scrollToIndex`. Do not call `scrollToIndex` synchronously inside the same effect/handler that dispatches `SET_EXPANDED` — the `virtualizer` object in that render is still bound to the pre-expansion `count`.

**Row click handler pattern** (`TreePane.tsx:78-84`) — model for `SET_SELECTED` dispatch inside the reveal effect:
```tsx
const handleRowClick = (node: models.FlatNode) => {
  if (node.type === 'directory') {
    dispatch({ type: 'TOGGLE_EXPAND', payload: node.path });
  }
  dispatch({ type: 'SET_SELECTED', payload: node.path });
};
```

---

### `frontend/src/components/workspace/Toolbar.tsx` — wire `.ws-search` onClick

**Analog:** itself, existing button markup, no analog needed elsewhere.

**Current inert button** (`Toolbar.tsx:57-101`, excerpt):
```tsx
<button
  type="button"
  className="no-drag ws-search"
  aria-label="Search every catalog"
  style={{ ... }}
>
  <svg ...>...</svg>
  <span style={{ fontSize: 12.5, color: 'var(--dm)' }}>Search every catalog…</span>
  <span className="mono" style={{ marginLeft: 'auto', ... }}>⌘K</span>
</button>
```
Add `onClick={...open palette...}` only — UI-SPEC is explicit: "No visual change to the button itself." `ToolbarProps` already threads callback props in from `WorkspaceShell` (`onToggleDetails` pattern, `Toolbar.tsx:5-9`) — follow the same prop-drilling convention for an `onOpenSearch` (or equivalent) callback rather than reaching into context from inside `Toolbar`.

---

### `frontend/src/workspace.css` — palette styles

**Analog:** existing z-index scale + chip-family convention.

**Z-index tokens already declared** (`workspace.css:29-31`):
```css
--z-details-drawer: 100;
--z-overlay: 200;
--z-dialog: 300;
```
Palette scrim uses `z-index: var(--z-overlay)` — do not add a new var.

**Chip sibling-class convention** (per UI-SPEC's Color section resolution — `.ws-chip` is hover-only/interactive, `.ws-dir-chip` is the existing precedent for a differently-styled sibling chip). Grep result confirms both classes exist at `workspace.css:66-80`. New `.ws-palette-chip` should be declared as a third sibling in the same family, static/filled, no `:hover` rule (mirrors `.ws-dir-chip`'s existence as proof this "add a sibling chip class" pattern is already established twice).

**Backdrop/overlay precedent** (`WorkspaceShell.tsx:66`, `workspace.css:386-394` `.ws-backdrop`):
```tsx
<div className="ws-backdrop" onClick={closeDrawer} />
```
The palette's scrim-click-to-close follows this exact click-handler-on-the-backdrop-div idiom, with its own distinct `rgba(4,6,9,.62)` value per UI-SPEC (not `.ws-backdrop`'s `rgba(0,0,0,.35)` — declare as a new literal on the palette's own scrim class, not a shared token, per UI-SPEC's explicit reasoning).

---

## Shared Patterns

### Wails binding call → `wailsError`/`extractErrorMessage`
**Source:** `frontend/src/services/wailsAPI.ts:22-39`
**Apply to:** `searchIndexed` wrapper (the only new binding call this phase)
```typescript
function extractErrorMessage(error: any): string {
  if (typeof error === 'string') return error;
  return error?.message || 'Unknown error';
}
function wailsError(error: any): { success: false; error: string } {
  return { success: false, error: extractErrorMessage(error) };
}
```

### Thin GUI-only Go wrapper over an unmodified CLI-shared method
**Source:** `internal/search/flatten.go` (`LoadCatalogFlat`) + `app.go`'s binding for it
**Apply to:** `SearchIndexed` (Service method + App binding)
Never modify `SearchCatalogs` or `cli/search.go` — the new capability wraps, it does not touch, the shared walk.

### `window`-level keydown listener, guarded by an open/active flag, with cleanup
**Source:** `WorkspaceShell.tsx:33-40`
**Apply to:** both the global ⌘K listener (new, top-level) and anything inside `useModalBehavior`'s Escape handling
```typescript
useEffect(() => {
  if (!state.detailOverlay) return;
  const onKeyDown = (e: KeyboardEvent) => {
    if (e.key === 'Escape') closeDrawer();
  };
  window.addEventListener('keydown', onKeyDown);
  return () => window.removeEventListener('keydown', onKeyDown);
}, [state.detailOverlay]);
```

### Reducer discard-if-stale discipline
**Source:** `AppContext.tsx:95-99` (`TREE_LOADED`)
**Apply to:** `pendingReveal` consumption in `TreePane` — must not act on a reveal whose target catalog is no longer `currentCatalogId`.

### Merge-before-dispatch for `SET_EXPANDED`
**Source:** `AppContext.tsx:123-126` (reducer does a full replace, by design, for its two existing callers)
**Apply to:** the reveal's ancestor-expansion dispatch site in `TreePane` — must pass `{ ...state.expanded, ...ancestorMap }`, never the ancestor map alone.

## No Analog Found

| File | Role | Data Flow | Reason |
|------|------|-----------|--------|
| Focus-trap / scroll-lock / focus-restore logic inside `useModalBehavior.ts` | hook | event-driven | No prior modal/overlay in this codebase implements these three behaviors — `WorkspaceShell.tsx`'s drawer only has Escape+backdrop-click, no focus trap or scroll lock. UI-SPEC's "Shared Modal Behavior Hook Contract" section is the spec of record; there is no in-repo code to copy for this part. |
| Debounce + request-id/stale-response guard for live search | component logic | event-driven | No existing component in this codebase debounces an async binding call with a monotonic request-id guard — `CatalogRail.tsx`'s filter uses `useDeferredValue` on an in-memory array, which CONTEXT.md explicitly says does not transfer to this disk-I/O case. Implement per RESEARCH.md's Code Examples / Pitfall 5 guidance (a `useRef`-held counter), not from an in-repo analog. |

## Metadata

**Analog search scope:** `internal/search/`, `app.go`, `pkg/models/`, `frontend/src/hooks/`, `frontend/src/contexts/`, `frontend/src/components/workspace/`, `frontend/src/services/`, `frontend/src/workspace.css`
**Files scanned:** `app.go`, `internal/search/flatten.go`, `internal/search/flatten_test.go`, `internal/search/service.go`, `internal/search/service_test.go`, `frontend/src/hooks/useVisibleRows.ts`, `frontend/src/hooks/useMediaQuery.ts`, `frontend/src/services/wailsAPI.ts`, `frontend/src/components/workspace/TreePane.tsx`, `frontend/src/contexts/AppContext.tsx`, `frontend/src/components/workspace/Toolbar.tsx`, `frontend/src/components/workspace/WorkspaceShell.tsx`, `frontend/src/components/workspace/CatalogRail.tsx`, `frontend/src/workspace.css` (grep for z-index/chip tokens)
**Pattern extraction date:** 2026-08-14
