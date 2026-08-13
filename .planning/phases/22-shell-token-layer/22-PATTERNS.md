# Phase 22: Shell + Token Layer - Pattern Map

**Mapped:** 2026-08-13
**Files analyzed:** 15 (new/modified/deleted)
**Analogs found:** 15 / 15 (deletions use the deleted file itself as the "analog" for conventions being carried forward)

## File Classification

| New/Modified File | Role | Data Flow | Closest Analog | Match Quality |
|---|---|---|---|---|
| `frontend/src/App.tsx` (rewritten) | component (shell root) | request-response (localStorage read, no IPC) | `frontend/src/App.tsx` (current) | exact — same file, same responsibilities minus tabs |
| `frontend/src/main.tsx` (modified) | provider/bootstrap | event-driven (sync init) | `frontend/src/main.tsx` (current) | exact |
| `frontend/src/themes.ts` (extended) | config/model | transform | `frontend/src/themes.ts` (current) | exact |
| `frontend/src/themeTokens.ts` (new) | utility | transform | `App.tsx`'s `applyTheme()` (current, lines 67-89) | role-match (closest existing "compute tokens, write to :root" logic) |
| `frontend/src/hooks/useMediaQuery.ts` (new) | hook | event-driven | none in codebase (first hook file) | no analog — see below |
| `frontend/src/workspace.css` (new) | config/styling | — | `frontend/src/index.css` + `frontend/src/style.css` (current) | role-match (token/`:root` pattern + `@font-face` pattern) |
| `frontend/src/components/workspace/Toolbar.tsx` (new) | component | event-driven (drag region, click) | `frontend/src/components/Header.tsx` (deleted this phase) | exact — same role (top chrome bar), only existing drag-region precedent |
| `frontend/src/components/workspace/CatalogRail.tsx` (new) | component (skeleton) | static-content | `frontend/src/components/tabs/BrowseCatalogsTab.tsx` (deleted this phase, `.Sidebar`) | role-match (rail/sidebar-shaped pane) |
| `frontend/src/components/workspace/TreePane.tsx` (new) | component (skeleton) | static-content | `frontend/src/components/MainContent.tsx` (deleted this phase) | role-match (main content area, empty-state pattern) |
| `frontend/src/components/workspace/DetailsPanel.tsx` (new) | component (skeleton, pane+drawer) | static-content | `frontend/src/components/tabs/SearchCatalogsTab.tsx` (deleted this phase, `.Sidebar`/detail-adjacent) + z-index/backdrop pattern is new | role-match, partial |
| `frontend/src/components/workspace/StatusBar.tsx` (new) | component (skeleton) | static-content | `frontend/src/components/Header.tsx` (deleted, as a fixed-height chrome bar) | partial (fixed-height bar convention only) |
| `frontend/src/contexts/AppContext.tsx` (extended + pruned) | store (Context + reducer) | event-driven | `frontend/src/contexts/AppContext.tsx` (current) | exact — same file, add fields, remove 15 dead ones |
| `frontend/src/assets/fonts/*.woff2` + `IBM-Plex-OFL.txt` (new) | static asset | file-I/O (build-time) | `frontend/src/assets/fonts/nunito-v16-latin-regular.woff2` + `OFL.txt` (current) | exact — identical vendoring pattern |
| `main.go` (modified — `options.App{}` literal) | config | — | `main.go`'s existing `options.App{}` literal (current, lines 61-75) | exact — same struct literal, additive field |
| `frontend/src/App.tsx` — `<CatalogModal>` render + `openCatalogModal` listener | component (unchanged, carried over) | event-driven | `frontend/src/App.tsx` (current, lines 53-64, 113-117) | exact — copy verbatim into new shell |

## Pattern Assignments

### `frontend/src/App.tsx` (component, shell root)

**Analog:** `frontend/src/App.tsx` (current file — being rewritten, not a different file)

**What to keep verbatim** (current `App.tsx` lines 1-8, 106-119):
```typescript
import { AppProvider } from './contexts/AppContext';
import { getThemeById, getDefaultTheme, Theme } from './themes';
import CatalogModal from './components/CatalogModal';
import './services/wailsAPI';  // Initialize Wails API wrapper
```
Keep the `ConfigProvider` (antd) wrapping at minimum `CatalogModal`, the `catalogModalVisible`/`catalogModalPath` local `useState`, the `openCatalogModal` window-event listener (current lines 53-64), and the `<CatalogModal ... />` render call (current lines 113-117) — CONTEXT.md/RESEARCH.md require `CatalogModal` to keep working even though it's unreachable this phase.

**What to remove:** `Header`/`MainContent` imports and renders, the `applyTheme()`-in-`useEffect` block (current lines 17-65) — token application moves to `main.tsx` module-init (see Pattern 1 below) to avoid FOUC. Do not delete the `storcat-theme` → `storcat-theme-id` migration logic; port it into the new `main.tsx`/`themeTokens.ts` init path.

**New structure:**
```typescript
function App() {
  return (
    <ConfigProvider theme={{ /* only affects CatalogModal now */ }}>
      <AppProvider>
        <WorkspaceShell />
        <CatalogModal visible={...} catalogPath={...} onClose={...} />
      </AppProvider>
    </ConfigProvider>
  );
}
```

---

### `frontend/src/main.tsx` (bootstrap, Pattern 1: theme applied before first paint)

**Analog:** `frontend/src/main.tsx` (current, all 15 lines) + `App.tsx`'s current `applyTheme()` (lines 67-89)

**Current pattern to extend** (main.tsx, full file):
```typescript
import React from 'react'
import {createRoot} from 'react-dom/client'
import './style.css'
import './index.css'
import App from './App'

const container = document.getElementById('root')
const root = createRoot(container!)
root.render(<React.StrictMode><App/></React.StrictMode>)
```
**Required change:** call the new `themeTokens.ts` init function (reads `storcat-theme-id`/`storcat-density`/`storcat-rail-side` from localStorage, resolves/validates against allowlists, computes all 14 tokens, writes to `document.documentElement.style`) **synchronously, before `root.render(...)`** — not inside a `useEffect`. Add `import './workspace.css'` alongside the existing `style.css`/`index.css` imports (both stay — `index.css`'s legacy `ThemeColors` block and `!important` antd overrides remain live per CONTEXT.md until Phase 26).

---

### `frontend/src/themes.ts` (extend with `tokens` block)

**Analog:** `frontend/src/themes.ts` (current, full file read — 349 lines)

**Current shape to preserve exactly** (`Theme` interface, lines 35-42, and per-theme object literal shape lines 44-70):
```typescript
export interface Theme {
  id: string;
  name: string;
  type: 'light' | 'dark';
  colors: ThemeColors;       // KEEP until Phase 26
  antdAlgorithm: 'default' | 'dark';
  antdPrimaryColor?: string;
}
```
**Add** a `tokens: { bg: string; p: string; p2: string; ch: string; l: string; tx: string; ac: string }` field per theme, sourced from RESEARCH.md's verified `THEMES` array (lines 604-616 of the prototype, quoted verbatim in RESEARCH.md's "Code Examples" section). The existing 11 `id`s already 1:1 match the prototype's 11 `id`s (verified in RESEARCH.md) — add `tokens` to each existing object literal, do not create new theme entries. Keep `getThemeById`/`getDefaultTheme` (lines 343-349) unchanged — `getDefaultTheme()` still returns `themes[0]` (`storcat-light`), matching the locked fresh-install default.

---

### `frontend/src/themeTokens.ts` (new — TS port of `varsFor`/`lum`/density branch)

**Analog:** `App.tsx`'s current `applyTheme()` (lines 67-89) for the "write CSS custom properties to `document.documentElement`" pattern; RESEARCH.md's verbatim-quoted prototype source for the actual algorithm.

**Structural pattern to copy from `applyTheme()`** (current `App.tsx:67-89`):
```typescript
const applyTheme = (theme: Theme) => {
  document.documentElement.setAttribute('data-theme', theme.id);
  const root = document.documentElement;
  root.style.setProperty('--app-bg', theme.colors.appBg);
  // ...one setProperty call per token
};
```
The new `applyTokens()` follows the identical `document.documentElement.style.setProperty(...)` mechanism, extended to all 14 tokens (`--bg --p --p2 --ch --l --tx --ac --l2 --dm --fn --sel --hov --acs --onac`) plus density vars (`--rh --rp --mp --hp --fs`). Also keep setting `data-theme` (used nowhere yet as a CSS selector, but harmless to preserve for future debugging/DevTools use, matching current behavior).

**Algorithm to port verbatim** — copy directly from RESEARCH.md's "Code Examples → Theme data" section (`lum()`, `varsFor()`, density branch) — do not re-derive. Per CONTEXT.md's locked decision, implement the OKLab-based mix (not RGB averaging) for `mixHex`/`mixAlpha`. Density fallback default is `'Comfortable'`, diverging intentionally from the prototype's own `'Compact'` default (RESEARCH.md flags this explicitly).

---

### `frontend/src/hooks/useMediaQuery.ts` (new — no analog)

**No existing hook file in the codebase** — this is the first file in a `hooks/` directory. Follow the project's general conventions (function declaration or arrow function exported as default/named, `camelCase.ts` filename per CLAUDE.md naming patterns for "non-component TS modules"). RESEARCH.md specifies the shape: a `matchMedia`-based listener for exactly two breakpoints (1280px, 1040px), no library.

---

### `frontend/src/workspace.css` (new — token layer + layout grid)

**Analog 1 — `:root` custom-property pattern:** `frontend/src/index.css` (lines 1-22):
```css
:root {
  --header-height: 45px;
  --tab-nav-height: 50px;
}
:root {
  --app-bg: #f8f9fa;
  /* ...legacy ThemeColors block, stays until Phase 26 */
}
```
Follow this same "custom properties on `:root`, JS overrides at runtime" convention for the new 14 tokens + density vars + z-index scale (`--z-details-drawer: 100; --z-overlay: 200; --z-dialog: 300;`) — declared as defaults in `workspace.css`, overwritten by `themeTokens.ts`'s `applyTokens()` at init.

**Analog 2 — `@font-face` pattern:** `frontend/src/style.css` (lines 15-21):
```css
@font-face {
    font-family: "Nunito";
    font-style: normal;
    font-weight: 400;
    src: local(""),
    url("assets/fonts/nunito-v16-latin-regular.woff2") format("woff2");
}
```
Copy this exact shape for the 5 new IBM Plex `@font-face` blocks (declarations already fully written out in RESEARCH.md's "Font vendoring" section — reuse verbatim, adjusting only the relative path if `workspace.css` lives at the same directory depth as `style.css`, i.e. `./assets/fonts/...`).

**Analog 3 — scrollbar/box-sizing global rules:** `frontend/src/index.css` (lines 24-39, `body`/`* { box-sizing: border-box; }`/`#root { height: 100vh; }`) — keep these in `index.css`, do not duplicate in `workspace.css`; `workspace.css` owns only the new grid/token/z-index/density rules.

---

### `frontend/src/components/workspace/Toolbar.tsx` (new — 46px toolbar, drag region)

**Analog:** `frontend/src/components/Header.tsx` (deleted this phase, but read for its pattern before deletion — full file, 57 lines)

**Drag-region pattern — the only existing precedent in the codebase** (`Header.tsx:10-21`):
```typescript
<AntHeader
    style={{
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'space-between',
      padding: '0 24px',
      height: 'var(--header-height)',
      background: 'var(--header-bg)',
      borderBottom: 'none',
      '--wails-draggable': 'drag',
    } as React.CSSProperties & { '--wails-draggable'?: string }}
  >
```
Copy this exact TypeScript cast pattern (`as React.CSSProperties & { '--wails-draggable'?: string }`) for the new Toolbar's root `<div>` — TS's `CSSProperties` type doesn't recognize custom properties natively, this is the established workaround. Apply `--wails-draggable: drag` to the toolbar root; apply the new `.no-drag` CSS class (defined in `workspace.css` per RESEARCH.md Pattern 3: `.no-drag { --wails-draggable: no-drag; }`) to the search field, theme chip, gear, and "Details" chip — this exact class doesn't exist yet, establish it here.

**Component/import pattern to follow** (plain function declaration, default export, `PascalCase.tsx`) — same shape as `Header.tsx:1,8,57` (`function Header() { ... } export default Header;`), but drop the `antd`/`Layout`/`Typography` imports entirely (Toolbar is plain-CSS per CONTEXT.md).

---

### `frontend/src/components/workspace/{CatalogRail,TreePane,DetailsPanel,StatusBar}.tsx` (new — skeleton panes)

**Analog:** `frontend/src/components/tabs/BrowseCatalogsTab.tsx` / `SearchCatalogsTab.tsx` (deleted this phase) for the `.Sidebar`/`.Content` sub-component export pattern used by the old tab system — **do not reuse this specific `.Sidebar`/`.Content` object-export pattern** (that was for tab-switching; the new shell renders all panes simultaneously, not conditionally per tab). Instead, each new pane is a single plain function-declaration component, default-exported, following the same base conventions:
- Function declarations, not arrow functions (CLAUDE.md convention, consistent across all deleted components)
- Props interface directly above the component, named `{ComponentName}Props`
- No class components

**Empty-state copy source:** All literal copy for these skeletons is transcribed verbatim from UI-SPEC's "Copywriting Contract" table — do not invent strings. `TreePane`'s CTA buttons are rendered but inert (no `onClick` wiring) per the locked scope.

**Scroll region pattern (all three panes):**
```css
flex: 1; overflow-y: auto; min-height: 0; scrollbar-gutter: stable;
```
This is a new convention (no prior scrollable pane exists in plain CSS in this codebase — the old `ModernTable.tsx` used antd's own scroll handling) — declare it once in `workspace.css` as a shared class (e.g. `.pane-scroll`) and apply to rail/tree/details.

---

### `frontend/src/contexts/AppContext.tsx` (extend + prune)

**Analog:** `frontend/src/contexts/AppContext.tsx` (current, full file, 144 lines) — same file, edited in place.

**Convention to preserve exactly** (current lines 26-44, action type union; lines 68-119, reducer switch):
```typescript
type AppAction =
  | { type: 'SET_SELECTED_DIRECTORY'; payload: string | null }
  // ...
function appReducer(state: AppState, action: AppAction): AppState {
  switch (action.type) {
    case 'SET_SELECTED_DIRECTORY':
      return { ...state, selectedDirectory: action.payload };
    // ...
  }
}
```
`UPPER_SNAKE_CASE` action type strings, one case per field, `{ ...state, field: action.payload }` spread pattern — replicate exactly for new fields: `density: 'Compact' | 'Comfortable'`, `railSide: 'Left' | 'Right'`, `detailOverlay: boolean`.

**Fields to remove** (confirmed zero remaining readers/writers by RESEARCH.md's grep): `selectedDirectory`, `selectedOutputDirectory`, `selectedSearchDirectory`, `selectedBrowseDirectory`, `isCreating`, `isSearching`, `isLoading`, `searchResults`, `sortColumn`, `sortDirection`, `browseCatalogs`, `browseSortColumn`, `browseSortDirection`, `sidebarCollapsed`, `sidebarPosition`, `activeTab` — remove their `AppState` fields, `AppAction` union members, `initialState` entries, and reducer cases in the same edit that adds the new workspace fields.

**Do NOT touch:** `catalogModalOpen`/`catalogModalTitle`/`catalogModalHtmlPath` and `OPEN_CATALOG_MODAL`/`CLOSE_CATALOG_MODAL` — RESEARCH.md flags these as pre-existing dead code (never read/dispatched; `App.tsx` uses local `useState` instead) that predates this phase and is out of scope to fix.

**`useAppContext()` hook (lines 138-144) — keep unchanged verbatim:**
```typescript
export function useAppContext() {
  const context = useContext(AppContext);
  if (context === undefined) {
    throw new Error('useAppContext must be used within an AppProvider');
  }
  return context;
}
```

---

### `frontend/src/assets/fonts/*` (new vendored files)

**Analog:** `frontend/src/assets/fonts/nunito-v16-latin-regular.woff2` + `frontend/src/assets/fonts/OFL.txt` (current, existing precedent)

Same directory, same "extract pre-subsetted woff2, commit binary + license text, declare via `@font-face`, no `package.json` dependency" pattern. Add a **second**, separate `IBM-Plex-OFL.txt` (do not overwrite/merge into the existing `OFL.txt`, which is Nunito's copyright holder text) — RESEARCH.md's Pitfalls section confirms IBM Plex's OFL copyright block (IBM Corp.) differs from Nunito's and must travel with the files independently.

---

### `main.go` (modified — `options.App{}` literal)

**Analog:** `main.go`'s own existing `options.App{}` literal (current, lines 61-75):
```go
err := wails.Run(&options.App{
    Title:  "StorCat",
    Width:  startWidth,
    Height: startHeight,
    AssetServer: &assetserver.Options{
        Assets: assets,
    },
    BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
    OnStartup:        app.startup,
    OnDomReady:       app.domReady,
    OnBeforeClose:    app.beforeClose,
    Bind: []interface{}{
        app,
    },
})
```
**Additive change only:** add `Mac: &mac.Options{TitleBar: mac.TitleBarHiddenInset()}` as one more field in this same literal, plus an unconditional `import "github.com/wailsapp/wails/v2/pkg/options/mac"` in the existing import block (current lines 3-13) — RESEARCH.md verifies (via `go doc` against the installed module) that no build tag is needed; `mac.Options`/`mac.TitleBar` are OS-agnostic Go structs, and Wails' own internal per-OS files decide whether to honor the field. Do not create a new file or add a `runtime.GOOS` check in Go — the darwin-only inset padding is a frontend/CSS concern (see `themeTokens.ts`/`Environment()` pattern in RESEARCH.md Pattern 2), not a Go one.

---

## Shared Patterns

### Theme application (single code path, no wrapper div)
**Source:** `App.tsx`'s current `applyTheme()` (lines 67-89), superseded by new `themeTokens.ts`
**Apply to:** `main.tsx` (init), any future Settings-triggered theme change (Phase 26, `themeChange` `CustomEvent` listener pattern stays — see below)
```typescript
window.addEventListener('themeChange', handleThemeChange as EventListener);
```
This `CustomEvent`-on-`window` mechanism (current `App.tsx:43-51`) is the established cross-component signalling convention (CLAUDE.md: "Cross-component signalling via `CustomEvent` on `window`") — keep it as the mechanism Phase 26's Settings surface will use to trigger a theme change; Phase 22 only needs the listener present in `App.tsx`/`main.tsx` (there is no theme picker UI yet, so this listener currently has no dispatcher — that's fine, matches the `openCatalogModal` precedent of "listener kept alive, no current dispatch site").

### Drag-region opt-out (`.no-drag`)
**Source:** New convention, first precedent is `Header.tsx`'s positive case (`'--wails-draggable': 'drag'`, `Header.tsx:19`) plus RESEARCH.md's verified Wails mechanism (CSS custom-property inheritance/shadowing, not a special keyword).
**Apply to:** `Toolbar.tsx` — every interactive descendant (search field, theme chip, gear, "Details" chip) needs the `.no-drag` class in the *same* implementation pass, not added later.
```css
.no-drag { --wails-draggable: no-drag; }
```

### localStorage read with allowlist validation (THEME-06 error case)
**Source:** `App.tsx`'s current theme-load block (lines 18-37) — established pattern of "read key, validate/fallback, rewrite valid value":
```typescript
const savedThemeId = localStorage.getItem('storcat-theme-id');
let themeToLoad = getDefaultTheme();
if (savedThemeId) {
  const savedTheme = getThemeById(savedThemeId);
  if (savedTheme) { themeToLoad = savedTheme; }
}
```
**Apply to:** `themeTokens.ts`'s init function for all three persisted keys (`storcat-theme-id`, new `storcat-density`, new `storcat-rail-side`) — extend this exact "get, validate against known set, fall back, rewrite" shape rather than trusting stored strings directly (RESEARCH.md's Security Domain section requires a strict allowlist check).

### Function-declaration components, PascalCase, default export at bottom
**Source:** Every existing component in `frontend/src/components/` (verified across `Header.tsx`, `MainContent.tsx`, `CatalogModal.tsx`, `tabs/*.tsx`)
**Apply to:** All new `components/workspace/*.tsx` files.

## No Analog Found

| File | Role | Data Flow | Reason |
|---|---|---|---|
| `frontend/src/hooks/useMediaQuery.ts` | hook | event-driven | First hook file in the codebase (no `hooks/` directory exists today) — build per RESEARCH.md's `matchMedia`-listener sketch, no in-repo precedent to copy structure from |
| `frontend/src/workspace.css`'s z-index scale declarations | config | — | No prior z-index scale exists anywhere in the codebase (no overlays/drawers existed before this phase) — values are locked in CONTEXT.md (100/200/300), not derived from an analog |
| OKLab color-mix TS module (`mixHex`/`mixAlpha` internals of `themeTokens.ts`) | utility | transform | No prior color-math utility exists in the codebase; port directly from the prototype's `color-mix(in oklab, ...)` semantics per RESEARCH.md's open-question recommendation (implement full OKLab conversion, not RGB averaging) |

## Metadata

**Analog search scope:** `frontend/src/` (all components, contexts, themes, styles, entry points), `main.go` (Wails app options)
**Files scanned:** `App.tsx`, `main.tsx`, `themes.ts`, `contexts/AppContext.tsx`, `index.css`, `style.css`, `components/Header.tsx`, `components/CatalogModal.tsx` (partial), `components/MainContent.tsx` / `tabs/*` (via RESEARCH.md's prior direct reads, cross-referenced), `assets/fonts/` directory listing, `main.go` (options.App literal + imports)
**Pattern extraction date:** 2026-08-13
