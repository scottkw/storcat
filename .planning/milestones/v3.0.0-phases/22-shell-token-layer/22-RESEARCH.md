# Phase 22: Shell + Token Layer - Research

**Researched:** 2026-08-13
**Domain:** Full custom-UI replacement (React/TS shell, plain CSS token layer) inside an existing Go/Wails v2.10.2 desktop app — no backend changes
**Confidence:** HIGH — nearly every claim below is verified by reading source directly (`go doc`/module cache for Wails, `npm pack`/extraction for fonts, direct file reads for the prototype and existing frontend), not by web search.

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

**Shell Composition & Migration Scope**
- The three-tab UI is deleted in this phase — `App.tsx` renders only the new workspace; `Header.tsx`, `MainContent.tsx`, `WelcomeContent.tsx`, and `components/tabs/*` are removed rather than kept behind a flag or left unrendered.
- Ant Design is removed from every surface built in this phase, but the `antd` / `@ant-design/icons` dependencies stay installed until their last consumer (`CatalogModal.tsx`) is replaced in Phase 26. No big-bang rewrite in the first phase.
- Rail, tree, and details panes render as static skeletons in Phase 22 — correct dimensions, tokens, borders, and empty-state messaging, with no data wiring. Data wiring belongs to Phase 23.
- Styling is plain CSS with CSS custom properties in a dedicated `workspace.css`, matching the handoff's token model and the existing `index.css` pattern. No CSS-in-JS, no CSS Modules, no Tailwind (explicitly out of scope for the project).

**Token & Theme Layer**
- All 14 tokens are applied on `:root` via the successor to the existing `applyTheme()` function — one code path, and existing `var(--…)` consumers keep working through the migration. Not a wrapper `<div>` (only needed if two themes coexist, which they don't).
- Derived tokens (`--l2 --dm --fn --acs --sel --hov --onac`) are computed in TypeScript as concrete `rgb()` values at theme-apply time. CSS `color-mix()` is explicitly rejected — Wails does not control WebKitGTK's version on Linux (see `.planning/research/PITFALLS.md`).
- `--onac` is computed from relative luminance (`> 0.45 → dark text`), ported from the handoff helper. This is what keeps Gruvbox orange / Monokai green (light accents) and GitHub blue (dark accent) all legible (THEME-02).
- `themes.ts` gains a `tokens` block per theme (`bg p p2 ch l tx ac`) sourced from the handoff's authoritative `THEMES` array; the legacy `colors` block stays until the last antd surface is retired in Phase 26, then gets dropped.
- IBM Plex Sans (400/500/600) and IBM Plex Mono (400/500) are vendored as self-hosted **latin-subset** woff2 in `frontend/src/assets/fonts/`, declared via `@font-face`, bundled by Vite into the embedded binary. No CDN, no network access (THEME-05). Latin subset chosen to control binary size.

**Responsive Layout & Window Chrome**
- Breakpoints are driven by CSS media queries swapping the grid template (`268px 1fr 288px` at ≥1280px → `236px 1fr` at 1040–1279px → `200px 1fr` below 1040px). A small `useMediaQuery` hook is used only where React genuinely needs to know the tier (details rendered as pane vs drawer). No pure-JS resize listener; no container queries (same WebKitGTK version risk as `color-mix()`).
- Below 1280px the details panel is the *same component* rendered into a fixed-position right drawer with a backdrop, closable via Esc and backdrop click, toggled by the toolbar "Details" chip. No duplicate mobile-only component.
- Overlay stacking (SHELL-09) uses one documented z-index scale declared as CSS vars in `workspace.css`: details drawer 100 → create slide-over / ⌘K palette 200 → dialogs / Settings 300. Later phases slot into this scale rather than inventing numbers.
- macOS uses `TitleBarHiddenInset` in the Go app options (darwin only) so the real traffic lights sit inside the 46px toolbar, with ~78px reserved left inset on macOS; Windows and Linux keep the native title bar above the toolbar. Window drag uses `--wails-draggable` on the toolbar background, with the search field, theme chip, and gear explicitly opted out so clicks are not swallowed (SHELL-07). Frameless Windows/Linux chrome is FUT-02 — deferred, not in this milestone.

**Preferences & State**
- Theme, density, and rail position persist to localStorage (THEME-06): reuse the existing `storcat-theme-id`, add `storcat-density` and `storcat-rail-side`. Go `preferences.json` remains for window state only — no IPC round-trip for a UI-only concern.
- Workspace state (density, rail side, `detailOverlay`, overlay flags) extends the existing `AppContext` reducer. No new context, no new state-management dependency.
- The old tab UI's localStorage keys (`storcat-last-*`, sidebar position) are left untouched in this phase — they are inert once the tabs are gone. They get swept in Phase 26 when the Settings surface is built.
- Fresh-install defaults: theme `storcat-light` (unchanged from today — new users are not silently flipped to dark), density `Comfortable`, rail side `Left`.

### Claude's Discretion
- Exact component file names and decomposition within the shell.
- Precise spacing/radius/shadow values within the handoff's documented scales.
- Skeleton/empty-state copy for the not-yet-wired panes.
- Structure of the TypeScript color-mix helper module.

### Deferred Ideas (OUT OF SCOPE)
- Dropping the legacy `ThemeColors` block and uninstalling `antd` / `@ant-design/icons` — Phase 26, once `CatalogModal` is replaced.
- Sweeping the obsolete `storcat-last-*` and sidebar-position localStorage keys — Phase 26 Settings.
- Rail-as-drawer below ~820px (FUT-01), frameless Windows/Linux chrome (FUT-02) — deferred out of v3.0.0 at requirements definition.
- Frontend unit tests for the shell (TEST-01) — deferred to a separate testing milestone.
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| SHELL-01 | Single workspace view — no tabs — 46px toolbar, rail, tree, details, 26px status bar | §Removal Surface (exact files/imports to delete), §Architecture Patterns (project structure, layout skeleton) |
| SHELL-02 | Three panes laid out `268px 1fr 288px` at ≥1280px | §Code Examples (grid CSS ported verbatim from prototype's `varsFor`/`cols` logic, read at `StorCat 1a Demo.dc.html:1001`) |
| SHELL-03 | Details becomes a drawer, "Details" chip toggles it, 1040–1279px (rail 236px) | §Code Examples (responsive tier table, `detailDisp`/`detailPos`/`detailW`/`detailShadow` logic read at lines 1006–1009) |
| SHELL-04 | Tree keeps priority below 1040px (rail 200px, details stays drawer) | Same as SHELL-03 — one CSS media-query ladder covers both |
| SHELL-05 | Rail movable to right side, divider follows | §Code Examples (`railOrder`/`detailOrder`/`railBorderR`/`railBorderL` logic read at lines 1002–1005) |
| SHELL-07 | Window draggable from toolbar without swallowing clicks on search/theme/gear | §Window Chrome & Drag Regions (verified `--wails-draggable` mechanism from Wails source, not docs) |
| SHELL-08 | macOS real traffic lights inside toolbar (`TitleBarHiddenInset`); native title bar above on Windows/Linux | §Window Chrome & Drag Regions (verified `mac.TitleBar` struct + no-build-tag-needed finding) |
| SHELL-09 | Overlays stack correctly at every tier | §Z-Index Scale (locked 100/200/300, supersedes handoff's raw 3/6/7 — reconciled with PITFALLS.md #1) |
| THEME-01 | All 11 themes switch with immediate full repaint | §Theme Data (THEMES array verified verbatim), §Applying Theme Before First Paint |
| THEME-02 | Legible text on accent-filled buttons/badges across all 11 themes | §Theme Data (`--onac` luminance helper verified verbatim) |
| THEME-03 | Extended token set (14 tokens) driving all themes | §Theme Data, §Code Examples (TS port of `varsFor`) |
| THEME-04 | Density toggle (Compact/Comfortable) changes row height/padding/font-size | §Theme Data (density branch verified verbatim, lines 855–862) |
| THEME-05 | IBM Plex Sans/Mono, no network access | §IBM Plex Font Vendoring (concrete acquisition path, verified file sizes, license obligations) |
| THEME-06 | Theme/density/rail-position survive restart | §Applying Theme Before First Paint, §Runtime State Inventory (localStorage keys) |
</phase_requirements>

## Summary

Phase 22 is a frontend-only replacement: delete the three-tab Ant Design shell and build a single-view workspace shell in plain CSS + CSS custom properties, matching a pixel-final prototype exactly. **No Go code changes are required for this phase** — confirmed by reading `app.go`/`main.go` (theme/density/rail-side already persist to localStorage per the locked decision, and window chrome (`TitleBarHiddenInset`) is a static field set once in `main.go`'s existing `options.App{}` literal, not a new bound method). The two areas with real technical risk are (1) getting the Wails macOS title-bar inset and drag-region mechanics exactly right — verified directly from the Wails v2.10.2 Go module source below — and (2) precisely porting the prototype's `THEMES`/`lum`/`varsFor` JS logic to TypeScript without drifting from the verbatim values, since `color-mix()` is banned for the same WebKitGTK-version reason documented in the milestone's own `PITFALLS.md`.

The prototype (`StorCat 1a Demo.dc.html`) is not just a visual reference — its JS state machine (`renderVals()`, `varsFor`, the `cols`/`railOrder`/`detailDisp` computed properties) is effectively the executable spec for the responsive grid and rail-swap logic. Every pixel value and every conditional in the "Layout & Responsive Contract" section of the UI-SPEC traces back to a specific line in that file, which this document cites directly rather than re-deriving.

The font vendoring question (THEME-05) has a concrete, low-risk answer: `@fontsource/ibm-plex-sans`/`@fontsource/ibm-plex-mono` npm packages already ship pre-split latin-subset woff2 files at the exact weights needed (400/500/600 Sans, 400/500 Mono) — extracting them via `npm pack` (not installing as a dependency) and committing the ~98KB of resulting files mirrors an *existing, working precedent already in this exact repo*: `frontend/src/assets/fonts/nunito-v16-latin-regular.woff2` + `OFL.txt`, vendored the same way for the Wails scaffold's default font.

**Primary recommendation:** Build the shell as new files (no reuse of `Header.tsx`/`MainContent.tsx`/`tabs/*`, all deleted outright), port `THEMES`/`lum`/`varsFor`/density-branch verbatim from the prototype into `themes.ts` and a new `themeTokens.ts` module, apply the darwin-only `mac.TitleBarHiddenInset()` unconditionally in `main.go`'s existing single `options.App{}` literal (no build tags needed — Wails' own internal package already branches by OS), and gate the ~78px inset padding in CSS behind a runtime check via `window.runtime.Environment()` (already generated in `frontend/wailsjs/runtime/runtime.d.ts`).

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Workspace grid layout, responsive tiers | Frontend (Browser/Client) | — | Pure CSS grid + media queries; no server involvement |
| Theme token computation (`--onac`, mixes) | Frontend (Browser/Client) | — | Computed in TS at theme-apply time; no Go round-trip (locked decision) |
| Theme/density/rail-side persistence | Frontend (Browser/Client, localStorage) | — | Explicitly not IPC — "no IPC round-trip for a UI-only concern" (CONTEXT.md) |
| Window chrome (traffic lights, title bar) | Native shell (Wails/Go `options.App`) | Frontend (Browser/Client, CSS inset + `no-drag` regions) | `mac.TitleBar` is a native Cocoa-level option; drag-region opt-out is CSS evaluated by Wails' native webview host |
| Window drag region | Native shell (Wails webview host) | Frontend (CSS `--wails-draggable`) | Wails' desktop runtime JS (`dragTest`) reads the CSS custom property via `getComputedStyle`, then calls into the native host (`WailsInvoke("drag")`) |
| Font asset delivery | CDN/Static (Vite build → embedded binary) | — | `//go:embed all:frontend/dist` bakes the woff2 files into the Go binary; there is no CDN/static-server tier at runtime, but the *build-time* role is asset-bundling, distinct from app logic |
| Skeleton/empty-state content | Frontend (Browser/Client) | — | No backend calls this phase (`BrowseCatalogs`/`LoadCatalog` not invoked until Phase 23) |

## Standard Stack

### Core

No new runtime dependencies are required for this phase.

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|---------------|
| React | ^18.2.0 (existing) | Component tree for the shell | Already in use; no change |
| TypeScript | existing (`frontend/tsconfig.json`, strict mode) | Type-checked token/theme logic | Already in use |
| Vite | existing (`frontend/vite.config.ts`) | Bundles CSS/assets (incl. vendored fonts) into `dist/`, embedded by Go's `//go:embed all:frontend/dist` | Already in use; confirmed no `root`/`assetsInclude` overrides exist in the current minimal `vite.config.ts` — default asset handling (any imported/referenced file under `src/` gets fingerprinted into `dist/assets/`) is sufficient for woff2 files referenced via `url()` in CSS |
| Wails v2 | v2.10.2 (existing, `go.mod`) | Native window chrome (`TitleBarHiddenInset`), drag-region CSS contract, `Environment()` platform detection | Already in use; **no version bump needed** — `TitleBarHiddenInset`, the CSS drag mechanism, and `Environment()` all exist unchanged in v2.10.2 (verified via `go doc`/source read against the exact installed module, not release notes) |

### Supporting

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| none | — | Hand-rolled `useMediaQuery` hook (a `matchMedia` listener, ~15 lines) | The three breakpoints (1280px, 1040px) are the *only* two thresholds needed, and only for one React decision point (details pane-vs-drawer + "Details" chip visibility) — a library (e.g. `usehooks-ts`) would be one more dependency for something CONTEXT.md already scoped as "a small hook" |
| none | — | Hand-rolled relative-luminance + alpha-blend TS module (`themeTokens.ts`) | Verbatim port of ~15 lines of prototype math (`lum`, and an `mixHex`/`mixAlpha` equivalent to the prototype's `color-mix(in oklab, …)` calls) — no npm package computes this exact formula, and pulling one in would be overkill for 2 small pure functions |

### Alternatives Considered

| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Hand-rolled `useMediaQuery` | `usehooks-ts`, `react-responsive` | Adds a dependency for ~15 lines of code covering exactly 2 breakpoints; rejected per CONTEXT.md's own framing ("a small `useMediaQuery` hook") |
| Vendoring via `npm pack` extraction | Adding `@fontsource/ibm-plex-sans`/`@fontsource/ibm-plex-mono` as real `package.json` dependencies | CONTEXT.md explicitly locks "fonts are vendored files, not a dependency" (Integration Points) — a real dependency would ship *all* subsets/weights at install/build time and needs `assetsInclude`/import-path wiring; extracting the 5 needed woff2 files once and committing them (like the existing Nunito precedent) keeps `package.json` unchanged |
| TS-computed derived tokens | CSS `color-mix()` | Explicitly rejected in CONTEXT.md/PITFALLS.md #24 — Wails does not control WebKitGTK's version on Linux, and the prototype's own `mix()` helper literally shells out to `color-mix(in oklab, …)`, confirming this concern applies to the *exact* code being ported, not a hypothetical |

**Installation:**
```bash
# No new package.json dependencies this phase.
# Font files are extracted once (not installed) — see "IBM Plex Font Vendoring" below.
```

**Version verification:** No new packages are added to `package.json` this phase — see Package Legitimacy Audit below for the one adjacent check performed (the font-source packages used transiently to extract files).

## Package Legitimacy Audit

**No new packages are installed as `package.json` dependencies in this phase.** The only ecosystem interaction is a one-time, non-persistent `npm pack` extraction (see IBM Plex Font Vendoring) used to source static `.woff2` binary assets — the packages themselves are never added to `dependencies`/`devDependencies`.

For completeness, the source packages used for extraction were checked against the legitimacy gate:

| Package | Registry | Age (this version) | Downloads | Source Repo | Verdict | Disposition |
|---------|----------|---------------------|-----------|--------------|---------|-------------|
| `@fontsource/ibm-plex-sans` | npm | Published 2026-07-19 (routine version bump; package itself is long-established) | 459,623/week | `github.com/fontsource/font-files` | SUS (`too-new` — flags the *version's* publish date, not the package) | Not installed — used only to extract static files via `npm pack`; not a `package.json` dependency |
| `@fontsource/ibm-plex-mono` | npm | Published 2026-07-19 (same) | 924,326/week | `github.com/fontsource/font-files` | SUS (`too-new`) | Not installed — same as above |

**Packages removed due to `[SLOP]` verdict:** none.
**Packages flagged as suspicious `[SUS]`:** `@fontsource/ibm-plex-sans`, `@fontsource/ibm-plex-mono` — flagged only because of a recent routine version publish on an otherwise high-download, actively-maintained package; not a hallucination signal. Since neither package is added to `package.json`, no `checkpoint:human-verify` gate is required before an *install* — but the planner should add a `checkpoint:human-verify` before the one-time `npm pack` extraction step regardless, since it still executes a registry fetch. The specific package names (`@fontsource/ibm-plex-sans`, `@fontsource/ibm-plex-mono`) were recalled from training knowledge, not read from an official IBM Plex or Fontsource doc this session — tagged `[ASSUMED]` per the package-name-provenance rule despite `npm view`/`npm pack` succeeding against the real registry.

## Removal Surface (SHELL-01 migration scope)

Read directly (`Read` + `grep`, this session) to establish exactly what breaks/goes dead when `Header.tsx`, `MainContent.tsx`, `WelcomeContent.tsx`, and `components/tabs/*` are deleted:

| File | Importers (verified by grep across `frontend/src/`) | Disposition |
|------|-------------------------------------------------------|-------------|
| `components/Header.tsx` | Only `App.tsx:5,108` | Delete; remove both lines from `App.tsx`. Uses antd `Layout`/`Typography` and an inline `'--wails-draggable': 'drag'` style (`Header.tsx:19`) — this is the **only existing precedent for the drag-region CSS property in the codebase**; the new toolbar should follow the same inline-style-with-cast pattern (`as React.CSSProperties & { '--wails-draggable'?: string }`) since TS's `CSSProperties` type doesn't recognize custom properties. |
| `components/MainContent.tsx` | Only `App.tsx:6,110` | Delete; remove both lines from `App.tsx`. **Also removes the current Settings modal** (theme picker, window-persistence toggle, sidebar-position radio) — this logic lives *inside* `MainContent.tsx`, not a separate file. This is consistent with the locked scope: the toolbar's theme chip opens Settings with "a no-op or a stub" this phase (UI-SPEC Copywriting Contract), because the real Settings modal doesn't exist until Phase 26 — deleting `MainContent.tsx` is what actually removes the only current way to change themes/window-persistence via UI. Not a regression to flag to the user; it is the explicit, locked scope. |
| `components/WelcomeContent.tsx` | **None** — grepped across all of `frontend/src/`; not imported by `MainContent.tsx` or anywhere else | Already-dead code prior to this phase. Delete with zero import updates needed elsewhere. |
| `components/tabs/CreateCatalogTab.tsx` | Only `MainContent.tsx` | Delete along with `MainContent.tsx`. |
| `components/tabs/SearchCatalogsTab.tsx` | Only `MainContent.tsx` | Delete. **Also removes the only code that dispatches `window.dispatchEvent(new CustomEvent('openCatalogModal', ...))`** (`SearchCatalogsTab.tsx:206`) — see CatalogModal note below. |
| `components/tabs/BrowseCatalogsTab.tsx` | Only `MainContent.tsx` | Delete. Also dispatches `openCatalogModal` (`BrowseCatalogsTab.tsx:169`) — same note. |
| `components/ModernTable.tsx` | Only `BrowseCatalogsTab.tsx` and `SearchCatalogsTab.tsx` | Delete along with both tabs (no orphaned import survives). |

### `CatalogModal.tsx` must keep working — verified dependency chain

`CatalogModal.tsx` imports only `Modal, message` from `antd` (2 symbols) — a small footprint that justifies CONTEXT.md's decision to keep `antd` installed through this phase rather than hand-rolling a modal here. It is:
1. Rendered unconditionally in `App.tsx` (`<CatalogModal visible={...} .../>`, `App.tsx:113-117`) — **this render call must stay in the new `App.tsx`**, along with the `catalogModalVisible`/`catalogModalPath` local `useState` and the `openCatalogModal`/`handleOpenCatalog` window-event listener (`App.tsx:53-64`).
2. Its *only* trigger in the current codebase is the `openCatalogModal` `CustomEvent`, dispatched exclusively from `BrowseCatalogsTab.tsx` and `SearchCatalogsTab.tsx` — **both deleted this phase**. This is not a bug to fix in Phase 22: the UI-SPEC explicitly marks the details-panel "Open HTML catalog" footer button as inert this phase (no `onClick` wiring; opens in Phase 23+). `CatalogModal` therefore becomes *reachable by no UI path* in Phase 22 while remaining functionally intact — exactly the "KEEP working" requirement (it still renders correctly if invoked; nothing currently invokes it). Wire it back up when Phase 23 gives the tree/details panel a real "open HTML" action.
3. `ConfigProvider` (antd theme algorithm/token wrapper) currently wraps the *entire* app in `App.tsx:97-119`, including `Header`/`MainContent`/`CatalogModal`. Since `CatalogModal` is the only antd consumer left after this phase's deletions, **`ConfigProvider` must stay in the new `App.tsx`, scoped around at minimum `CatalogModal`** (wrapping the whole app, as today, is simplest and lowest-risk — antd's `ConfigProvider` with no antd components rendered elsewhere has no visible effect on the plain-CSS shell).

### AppContext fields that become dead

Reading `AppContext.tsx` directly: after the above deletions, these `AppState` fields have **zero remaining readers/writers** in the codebase: `selectedDirectory`, `selectedOutputDirectory`, `selectedSearchDirectory`, `selectedBrowseDirectory`, `isCreating`, `isSearching`, `isLoading`, `searchResults`, `sortColumn`, `sortDirection`, `browseCatalogs`, `browseSortColumn`, `browseSortDirection`, `sidebarCollapsed`, `sidebarPosition`, `activeTab` — all were written/read exclusively by `MainContent.tsx`/`tabs/*`. CONTEXT.md scopes this phase to *extend* `AppContext` (add density/rail-side/`detailOverlay`), not necessarily prune it — but leaving 15 dead fields in the same reducer the planner is extending is worth an explicit call: **removing the dead fields in the same edit that adds the new ones is lower-risk than doing it in a later phase** (no other code references them, confirmed by the grep above), and keeps the reducer legible for the new workspace fields being added alongside them.

One pre-existing anomaly, **not introduced by this phase**: `catalogModalOpen`/`catalogModalTitle`/`catalogModalHtmlPath` and the `OPEN_CATALOG_MODAL`/`CLOSE_CATALOG_MODAL` action types already exist in `AppContext.tsx` but are **never read or dispatched anywhere** — `App.tsx` uses its own local `useState` for the modal, not these reducer fields. This dead code predates Phase 22 and is out of this phase's scope to fix, but the planner should not confuse it with newly-dead fields.

### `index.css` antd `!important` overrides

`.ant-layout`, `.ant-layout-header`, `.ant-layout-sider`, `.ant-layout-content`, `.ant-table*`, `.ant-modal*` rules in `index.css` (lines 42–150) become dead once `Header`/`MainContent`/`ModernTable` are gone — **except** the `.ant-modal-*` rules (lines 129–150), which still apply to `CatalogModal`'s `Modal`. Do not delete the `.ant-modal-*` block; the `.ant-layout-*`/`.ant-table-*` blocks are safe to delete or leave (dead but harmless — CSS classes that never match any rendered element cost nothing at runtime). The `--header-height`/`--tab-nav-height` custom properties (`index.css:1-4`) and the 16-field `ThemeColors` custom-property block (`index.css:6-22`) stay per CONTEXT.md ("legacy `colors` block stays until... Phase 26").

## Architecture Patterns

### System Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│ main.tsx (module init, synchronous, before ReactDOM.render)      │
│   1. read localStorage: storcat-theme-id, storcat-density,       │
│      storcat-rail-side (fall back to locked defaults if absent/  │
│      invalid, rewrite valid value — THEME-06 error case)         │
│   2. compute all 14 tokens (TS port of lum/varsFor) for the      │
│      resolved theme + density                                    │
│   3. write tokens to document.documentElement.style              │
│      ── this MUST complete before first paint (no FOUC) ──       │
└───────────────────────────┬────────────────────────────────────┘
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│ App.tsx (React tree, replaces old Header+MainContent+tabs)       │
│  ┌───────────────────────────────────────────────────────────┐ │
│  │ WorkspaceShell (46px toolbar, wails-draggable region)      │ │
│  │  toolbar: traffic-light space · app mark · search (no-drag)│ │
│  │           · theme chip (no-drag) · gear (no-drag)          │ │
│  │           · "Details" chip (no-drag, tier<1280 only)       │ │
│  └───────────────────────────────────────────────────────────┘ │
│  ┌───────────────┬─────────────────────────┬──────────────────┐ │
│  │ Rail (skeleton)│ Tree pane (skeleton,    │ Details (skeleton │ │
│  │ empty-state    │  centered empty state)  │  pane ≥1280px OR  │ │
│  │ order:1/3,     │  order: 2 (fixed)       │  fixed-pos drawer │ │
│  │ useMediaQuery  │                          │  <1280px, closed  │ │
│  │ decides side   │                          │  by toolbar chip) │ │
│  └───────────────┴─────────────────────────┴──────────────────┘ │
│  ┌───────────────────────────────────────────────────────────┐ │
│  │ Status bar (26px, literal zero-state text)                  │ │
│  └───────────────────────────────────────────────────────────┘ │
│  <CatalogModal ...>  (unchanged, unreachable this phase — antd) │
└─────────────────────────────────────────────────────────────────┘
                             │
                             ▼ (no calls this phase)
                    window.electronAPI / wailsjs bindings
                    (BrowseCatalogs/LoadCatalog first used Phase 23)
```

### Recommended Project Structure

```
frontend/src/
├── App.tsx                    # replaced wholesale: renders WorkspaceShell + CatalogModal only
├── main.tsx                   # gains synchronous pre-render theme application
├── workspace.css              # NEW — token layer, layout grid, z-index scale, density vars
├── themes.ts                  # extended: THEMES-sourced `tokens` block added per theme
├── themeTokens.ts             # NEW — lum(), mixHex()/mixAlpha(), applyTokens() (TS port of varsFor)
├── hooks/
│   └── useMediaQuery.ts       # NEW — small matchMedia hook, 2 breakpoints
├── components/
│   ├── CatalogModal.tsx       # UNCHANGED — kept working, currently unreachable
│   ├── workspace/             # NEW — one file/small family per handoff screen region
│   │   ├── Toolbar.tsx
│   │   ├── CatalogRail.tsx    # skeleton only this phase
│   │   ├── TreePane.tsx       # skeleton only this phase
│   │   ├── DetailsPanel.tsx   # skeleton only this phase; same component, pane or drawer
│   │   └── StatusBar.tsx      # literal zero-state text
│   ├── Header.tsx             # DELETED
│   ├── MainContent.tsx        # DELETED
│   ├── WelcomeContent.tsx     # DELETED (already dead)
│   ├── ModernTable.tsx        # DELETED
│   └── tabs/                  # DELETED (entire directory)
├── contexts/
│   └── AppContext.tsx         # extended: density/side/detailOverlay added, 15 dead fields removed
└── assets/
    └── fonts/
        ├── OFL.txt            # existing — Nunito's; add IBM Plex's OFL text (see below)
        ├── nunito-v16-latin-regular.woff2   # existing, unrelated, left alone
        ├── ibm-plex-sans-latin-400-normal.woff2   # NEW
        ├── ibm-plex-sans-latin-500-normal.woff2   # NEW
        ├── ibm-plex-sans-latin-600-normal.woff2   # NEW
        ├── ibm-plex-mono-latin-400-normal.woff2   # NEW
        └── ibm-plex-mono-latin-500-normal.woff2   # NEW
```

### Pattern 1: Theme tokens applied synchronously before first paint

**What:** Read persisted theme/density/rail-side and write all 14 CSS custom properties to `:root` in a code path that executes before React's first commit — not in a post-mount `useEffect`.
**When to use:** App startup (`main.tsx`, before `createRoot().render()`), and again on any theme/density change from the (stubbed, Phase 26) Settings surface.
**Why this specific approach (not `useLayoutEffect`):** The existing codebase's Vite setup has no SSR and no inline critical CSS in `index.html` (confirmed by reading `frontend/index.html` — it's a bare `<div id="root">` + one `<script type="module">`), so the actual risk is React's first commit painting with `index.css`'s hardcoded light-theme `:root` defaults, then a `useEffect` (which fires *after* paint) swapping them — a visible flash. Calling the token-application function synchronously at module scope in `main.tsx`, before `root.render(...)` is invoked, guarantees the correct tokens are already on `document.documentElement` before React ever renders a frame, which is strictly earlier and simpler than a `useLayoutEffect` (which only helps for post-mount re-renders, not the very first paint of an empty document). This satisfies UI-SPEC E6's "Load-bearing... synchronously at module init or in a `useLayoutEffect`, never a post-mount `useEffect`" requirement using the module-init option.
**Example (TS port — values verified verbatim against the prototype, see "Theme Data" below):**
```typescript
// Source: ported from design_handoff_storcat_ui/designs/StorCat 1a Demo.dc.html lines 618-636
// (lum + varsFor), verified by direct Read this session.

function lum(hex: string): number {
  const h = hex.replace('#', '');
  const v = [0, 2, 4]
    .map(i => parseInt(h.slice(i, i + 2), 16) / 255)
    .map(x => (x <= 0.03928 ? x / 12.92 : Math.pow((x + 0.055) / 1.055, 2.4)));
  return 0.2126 * v[0] + 0.7152 * v[1] + 0.0722 * v[2];
}

// TS equivalent of the prototype's mix(a, pct, b) => color-mix(in oklab, a pct%, b)
// Computed as a concrete rgb()/rgba() string instead of shipping color-mix() at runtime
// (banned per PITFALLS.md #24 — WebKitGTK version risk on Linux).
function mixHex(a: string, pct: number, b: string): string {
  const pa = hexToRgb(a);
  const pb = hexToRgb(b);
  const t = pct / 100;
  const r = Math.round(pa.r * t + pb.r * (1 - t));
  const g = Math.round(pa.g * t + pb.g * (1 - t));
  const bl = Math.round(pa.b * t + pb.b * (1 - t));
  return `rgb(${r}, ${g}, ${bl})`;
}

function mixAlpha(hex: string, pct: number): string {
  const { r, g, b } = hexToRgb(hex);
  return `rgba(${r}, ${g}, ${b}, ${pct / 100})`;
}
```
**Note on fidelity:** the prototype's `mix()` uses `color-mix(in oklab, ...)` — perceptual OKLab blending, not simple sRGB channel averaging. A byte-for-byte-identical port would need an sRGB↔OKLab conversion (a few more lines: linearize sRGB, matrix-multiply to OKLab, lerp, matrix back, re-encode). Simple RGB-channel averaging (as sketched above) is visually close but **not identical** to the prototype's OKLab blend, especially for `--sel`/`--hov`/`--acs` (14%/8%/16% alpha-style blends against very different base/accent hues per theme). Flag this as a decision point for the planner: either (a) implement the full OKLab conversion for exact fidelity, or (b) accept the simpler RGB-average approximation as "close enough" for a token that's mostly a subtle background tint. This is **not** a locked decision in CONTEXT.md — treat as an open question (see below).

### Pattern 2: Wails macOS title bar — no build tag needed

**What:** `main.go`'s single, cross-platform `options.App{}` literal (in `runGUI()`, currently lines 61-75) gains a `Mac: &mac.Options{TitleBar: mac.TitleBarHiddenInset()}` field, set unconditionally — **not** behind a `//go:build darwin` tag.
**Why this works cross-platform (verified via `go doc` + reading the installed module source, not docs):**
```go
// Source: go doc -all github.com/wailsapp/wails/v2/pkg/options/mac (v2.10.2, module cache read directly)
type Options struct {
    TitleBar             *TitleBar
    Appearance           AppearanceType
    WebviewIsTransparent bool
    WindowIsTranslucent  bool
    Preferences          *Preferences
    DisableZoom          bool
    About      *AboutInfo
    OnFileOpen func(filePath string)
    OnUrlOpen  func(filePath string)
}

// Source: pkg/options/mac/titlebar.go (v2.10.2, read directly — verbatim)
func TitleBarHiddenInset() *TitleBar {
    return &TitleBar{
        TitlebarAppearsTransparent: true,
        HideTitle:                  true,
        HideTitleBar:               false,
        FullSizeContent:            true,
        UseToolbar:                 true,
        HideToolbarSeparator:       true,
    }
}
```
`mac.Options`/`mac.TitleBar` are plain, OS-agnostic Go structs (no `//go:build` constraint on the file itself — `pkg/options/mac/*.go` has no build tags, confirmed by directory listing) — they compile into every platform's binary. What actually branches by OS is **inside Wails itself**: `internal/frontend/desktop/desktop_darwin.go` / `desktop_windows.go` / `desktop_linux.go` (three separate build-tagged files, confirmed present in the module) — only the darwin variant reads and applies `options.App.Mac.TitleBar` to the native window; Windows/Linux builds simply never look at that field. **Consequence for `main.go`:** add `Mac: &mac.Options{TitleBar: mac.TitleBarHiddenInset()}` to the existing `options.App{}` literal, in the shared, non-build-tagged `main.go`, with a plain unconditional import of `github.com/wailsapp/wails/v2/pkg/options/mac` — no new file, no build tag, no `runtime.GOOS` check in Go.
**The ~78px inset padding is a CSS/frontend concern, not a Go one** — it needs a *runtime* platform check in the renderer (Go's `GOOS` is a compile-time constant per-binary, but StorCat ships one binary per platform, so this distinction matters only in that the padding must be conditional CSS, applied via a JS check, not a Go `if runtime.GOOS == "darwin"`). Detect it via the already-generated Wails runtime binding:
```typescript
// Source: frontend/wailsjs/runtime/runtime.d.ts (read directly, already generated in this repo)
// export interface EnvironmentInfo { buildType: string; platform: string; arch: string; }
// export function Environment(): Promise<EnvironmentInfo>;
import { Environment } from '../wailsjs/runtime/runtime';

const env = await Environment();
if (env.platform === 'darwin') {
  document.documentElement.style.setProperty('--toolbar-inset-left', '78px');
}
```
`platform` is populated Go-side as `goruntime.GOOS` (verified by reading `pkg/runtime/runtime.go:91-104` in the module — `result.Platform = goruntime.GOOS`), so the string values are exactly Go's `GOOS` constants: `"darwin"`, `"windows"`, `"linux"`.

### Pattern 3: `--wails-draggable` — exact mechanism, not folklore

**What:** Verified by reading Wails' own bundled runtime JS (`internal/frontend/runtime/runtime/desktop/main.js`, not docs) rather than relying on the CONTEXT.md/PITFALLS.md description alone.
**Mechanism (verbatim from source):**
```javascript
// Source: internal/frontend/runtime/desktop/main.js (v2.10.2, read directly)
let dragTest = function (e) {
    var val = window.getComputedStyle(e.target).getPropertyValue(window.wails.flags.cssDragProperty);
    if (val) { val = val.trim(); }
    if (val !== window.wails.flags.cssDragValue) {
        return false;   // not exactly "drag" -> do not start a window drag
    }
    if (e.buttons !== 1) { return false; }   // only primary button
    if (e.detail !== 1) { return false; }    // not a double-click
    return true;
};
```
Default `cssDragProperty` = `"--wails-draggable"`, default `cssDragValue` = `"drag"` (verified in `pkg/options/options.go:153-157`, applied by `MergeDefaults` — StorCat's `main.go` sets neither `CSSDragProperty` nor `CSSDragValue` today, confirmed by reading `main.go`, so the defaults apply).
**The exact opt-out semantics:** `getComputedStyle(e.target)` reads the **inherited, cascaded** value of the CSS custom property on the precise element the mousedown landed on. Because CSS custom properties inherit down the DOM tree by default, setting `--wails-draggable: drag` on the toolbar `<div>` makes every descendant compute that same value **unless a descendant sets its own value for that property** — any value other than the literal string `"drag"` (conventionally `"no-drag"`, but literally anything else — even an empty string) makes `dragTest` return `false` for that element and everything under it (its own descendants inherit the overriding value in turn). There is no special "no-drag" keyword recognized by Wails — it is pure CSS custom-property shadowing.
**Recommendation:** define `.no-drag { --wails-draggable: no-drag; }` as a shared utility class in `workspace.css` (matches CONTEXT.md's wording, "explicit `.no-drag`... established as a project convention here") and apply it to the search field, theme chip, gear, and "Details" chip. This exact class doesn't exist yet anywhere in the codebase — `Header.tsx`'s current drag region (`Header.tsx:19`) has no interactive children needing opt-out, so there's no prior `.no-drag` precedent to reuse, only the raw inline-style pattern for the *positive* case.

### Pattern 4: Responsive layout — CSS ported from the prototype's own computed properties

The prototype's `renderVals()` (read directly, lines 999–1012) computes these values from a `width`/`side`/`detailOverlay` state — this is the literal logic to re-express as CSS media queries + a `useMediaQuery`-driven boolean, not something to redesign:
```javascript
// Source: StorCat 1a Demo.dc.html:999-1012 (read directly, verbatim)
frameW: s.width + 'px',
narrow: s.width < 1280,
cols: s.width >= 1280 ? (s.side === 'Left' ? '268px 1fr 288px' : '288px 1fr 268px')
                       : (s.width >= 1040 ? '236px 1fr' : '200px 1fr'),
railOrder: s.side === 'Left' || s.width < 1280 ? 1 : 3,
detailOrder: s.side === 'Left' ? 3 : 1,
railBorderR: s.side === 'Left' || s.width < 1280 ? '1px solid var(--l)' : 'none',
railBorderL: s.side === 'Left' || s.width < 1280 ? 'none' : '1px solid var(--l)',
detailDisp: s.width >= 1280 || s.detailOverlay ? 'flex' : 'none',
detailPos: s.width >= 1280 ? 'static' : 'absolute',
detailW: s.width >= 1280 ? 'auto' : '288px',
detailShadow: s.width >= 1280 ? 'none' : '-24px 0 50px rgba(0,0,0,.45)',
```
Key subtlety worth calling out explicitly for the planner: **`railOrder` forces the rail to `order: 1` at any width below 1280px regardless of the `side` setting** — i.e. even if the user set rail-side to "Right," below 1280px the rail visually snaps back to the left and only the details drawer's position/shadow changes. This matches the UI-SPEC's own note ("<1280px forces rail to `order: 1` regardless of side setting") — verified here against the actual conditional, not just the prose description of it.

### Anti-Patterns to Avoid
- **Building a real focus-trap/scroll-lock hook this phase:** PITFALLS.md #2 correctly scopes the shared `useModalBehavior` hook to Phase 3 (first *interactive* overlay — the ⌘K palette). Phase 22's only overlay is the details drawer at 1040–1279px, which per CONTEXT.md needs only Esc + backdrop-click to close — implement that inline/small, but write it so Phase 24 can lift the pattern into the shared hook rather than duplicating divergent logic (see PITFALLS.md #10's "single close path" lesson, applicable in miniature here).
- **Reaching for `color-mix()` "just for this one token" as a shortcut:** even a single raw `color-mix()` call reintroduces the exact Linux risk the rest of the token layer avoids — PITFALLS.md's technical-debt table calls this out by name as "never acceptable."
- **Adding `@fontsource/*` to `package.json`:** locked out by CONTEXT.md ("fonts are vendored files, not a dependency") — use the extraction approach only.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|--------------|-----|
| Detecting macOS at runtime for the toolbar inset | A `navigator.userAgent`/`navigator.platform` sniff | `Environment()` from `frontend/wailsjs/runtime/runtime.ts` (already generated, already wraps `window.runtime.Environment()`) | It's already generated and wired into this exact repo; UA sniffing is fragile and Wails already exposes the real Go-side `GOOS` |
| Latin-subset IBM Plex woff2 files | Manually converting the full IBM Plex TTF/OTF release (from `IBM/plex` on GitHub) into a latin-only woff2 with a subsetting tool (`glyphhanger`, `fonttools subset`) | Extract the pre-subsetted files fontsource already publishes (`npm pack @fontsource/ibm-plex-sans@5.3.0`, pull `files/ibm-plex-sans-latin-{400,500,600}-normal.woff2` and mono equivalents) | Fontsource's files are already correctly subsetted, already verified to exist at the exact needed weights, and the total size is known (~98KB) — hand-subsetting risks subtly wrong glyph coverage or a heavier/lighter file than expected |
| The relative-luminance contrast helper | Any generic "get readable text color" npm package (e.g. `get-contrast-color`) | The exact 6-line formula from the prototype, ported verbatim | THEME-02's acceptance criteria is "matches the handoff," not "any reasonable contrast heuristic" — a different formula could produce a different verdict for a borderline accent color and silently diverge from the design |

**Key insight:** almost everything in this phase has a byte-identical, already-written reference implementation (the prototype's JS) — the risk is not "what algorithm to use" but "faithfully porting the one that's already specified," so "don't hand-roll" here mostly means "don't re-derive from first principles when a verbatim port is available and required."

## Runtime State Inventory

This phase deletes existing UI surface (`Header.tsx`, `MainContent.tsx`, `WelcomeContent.tsx`, `components/tabs/*`) rather than renaming it, but the same "what points at the deleted thing" discipline applies:

| Category | Items Found | Action Required |
|----------|-------------|-------------------|
| Stored data (localStorage) | `storcat-theme-id` (kept, reused), `storcat-last-create-directory`, `storcat-last-output-directory`, `storcat-last-catalog-directory`, `storcat-last-search-term`, `storcat-last-search-results`, `storcat-last-browse-catalogs` (all written only by the deleted tab components) | None this phase — CONTEXT.md explicitly leaves these "untouched... inert," swept in Phase 26. Do not add code to read/write them from the new shell. |
| Live service config | None found | N/A — no n8n/Datadog/Tailscale-equivalent external services in this project |
| OS-registered state | None found | N/A — no Task Scheduler/pm2/launchd/systemd registrations in this project |
| Secrets/env vars | None found | N/A — Phase 22 has no secrets; `wails.json` and `go.mod` contain no phase-relevant env-var coupling |
| Build artifacts | `frontend/dist/` (Vite build output, embedded via `//go:embed all:frontend/dist`) will contain new hashed font asset filenames after this phase's build — no stale-artifact risk since Vite content-hashes filenames and Go re-embeds on every build | None — this is normal build behavior, not a migration concern |
| Dead custom events | `openCatalogModal` `CustomEvent` — dispatched only by code deleted this phase (`BrowseCatalogsTab.tsx`, `SearchCatalogsTab.tsx`); listener stays alive in `App.tsx` | Keep the listener (harmless, and `CatalogModal` needs it later); do not add a new dispatch site this phase — that's Phase 23's "Open HTML catalog" wiring |
| Dead AppContext fields | See "Removal Surface" above — 15 fields | Recommend removing in the same edit that extends `AppContext` for workspace state (see rationale above) |

**Nothing found in:** live service config, OS-registered state, secrets/env vars — verified by inspecting `main.go`, `app.go`, `wails.json`, `go.mod`, and `frontend/package.json` directly; this is a UI-only phase with no OS integration surface.

## Common Pitfalls

Cross-referenced against `.planning/research/PITFALLS.md` (milestone-level, already researched) — only phase-22-relevant pitfalls are re-summarized here with this phase's specific verification note; **do not duplicate the full write-ups**, see PITFALLS.md for complete detail.

### Pitfall: Z-index scale reconciliation (builds on PITFALLS.md #1)
**What's different from PITFALLS.md #1:** that pitfall describes the handoff's *raw* z-index values (`3`/`6`/`7`). CONTEXT.md has **already locked a different, wider scale** (`100`/`200`/`300`) specifically to leave room for future insertions — this is not still an open risk to "avoid," it's a decision already made that Phase 22 must *implement* (declare `--z-details-drawer: 100`, `--z-overlay: 200`, `--z-dialog: 300` in `workspace.css`) and that Phases 24–26 must *honor* (reference the CSS vars, never a raw number). The residual risk PITFALLS.md #1 describes — someone hardcoding a raw number in a later phase — still applies to the *new* scale, just with different numbers.

### Pitfall: Applying theme after mount causes FOUC (THEME-01/THEME-06)
**What goes wrong:** The *existing* `applyTheme()` in `App.tsx` runs inside a `useEffect` (confirmed by reading `App.tsx:18-65` — the whole theme-load block, including `applyTheme(themeToLoad)`, is inside a `useEffect(() => {...}, [])`). If the successor function is wired the same way, every launch briefly shows `index.css`'s hardcoded light-theme `:root` defaults before the effect fires and repaints — worse than today because the new shell has far more surface area painted in theme colors (whole-window chrome vs. a header bar).
**How to avoid:** See Pattern 1 above — move the initial application to synchronous module-init code in `main.tsx`, before `root.render(...)`.
**Warning signs:** A visible one-frame flash of light-theme colors on launch when the persisted theme is dark, especially noticeable on `storcat-dark`/`nord`/`dracula`/etc.

### Pitfall: `--wails-draggable` on the "Details" chip specifically
PITFALLS.md #12 already flags this generally ("the responsive-tier Details chip is added well after Phase 1... potentially by a different plan/session") — but in this phase's actual scope, the Details chip **is being built in Phase 22 itself** (it's part of THEME-04/SHELL-03's toolbar, per UI-SPEC's icon/aria-label table), so the "later phase forgets it" framing doesn't apply here — the risk in Phase 22 specifically is forgetting it needs `.no-drag` **on first implementation**, not regressing it later. Verify all four toolbar interactive elements (search, theme chip, gear, Details chip) carry `.no-drag` in the same pass, not just the three visible at ≥1280px.

### Pitfall: `scrollbar-gutter: stable` — real but non-fatal on old WebKitGTK
**[CITED: web search, MEDIUM confidence]** `scrollbar-gutter: stable` shipped in Chromium 94 (2021)/Firefox 97 (2022); Safari/WebKit support arrived later (~Safari 18.2, Dec 2024), so — mirroring the `color-mix()` risk in PITFALLS.md #24 — an older Linux distro's system WebKitGTK may not support it. Unlike `color-mix()`, an unrecognized `scrollbar-gutter` declaration is simply ignored by the parser (graceful degradation, not an invalid-value break), so this is safe to declare unconditionally on the rail/tree/details scroll regions per PITFALLS.md #26's rule — no JS feature-detection or fallback needed, worst case is the pre-existing (already-shipped) Windows-vs-macOS scrollbar-width inconsistency PITFALLS.md #26 describes, on old-Linux only.

### Pitfall: OFL license text must accompany the vendored fonts
**[CITED: web search, MEDIUM confidence]** The OFL-1.1 requires the copyright statement, license notice, and license text to travel with a bundled/embedded font. The existing Nunito precedent already does this correctly (`frontend/src/assets/fonts/OFL.txt`) — but that file is Nunito's copyright holder text, not IBM Plex's. **Do not assume one `OFL.txt` covers both fonts** — IBM Plex's OFL copyright block (IBM Corp.) differs from Nunito's (The Nunito Project Authors). Either add a second `IBM-Plex-OFL.txt` (extracted from the `@fontsource` package's `LICENSE` file, confirmed present at `extract/package/LICENSE` during this session's `npm pack` verification) or a combined `THIRD-PARTY-FONTS.md` covering both — but the IBM Plex copyright text must be present somewhere in the repo, not silently omitted because "we already have a fonts OFL.txt."

## Code Examples

### Theme data — verified verbatim (THEME-01, THEME-02, THEME-03)

Read directly from `design_handoff_storcat_ui/designs/StorCat 1a Demo.dc.html:604-636` this session — quoted verbatim (not paraphrased) so the planner can diff against it directly:

```javascript
// Source: StorCat 1a Demo.dc.html:604-616 — THEMES array, verbatim
const THEMES = [
  { id: 'storcat-dark', name: 'StorCat Dark', type: 'dark', c: { bg: '#0b0e13', p: '#0f1319', p2: '#12161d', ch: '#151a22', l: '#232a35', tx: '#e6ebf2', ac: '#4fd6e0' } },
  { id: 'storcat-light', name: 'StorCat Light', type: 'light', c: { bg: '#f4f5f6', p: '#ffffff', p2: '#fafbfc', ch: '#f1f3f5', l: '#dee2e6', tx: '#212529', ac: '#0d8f9c' } },
  { id: 'github-dark', name: 'GitHub Dark', type: 'dark', c: { bg: '#0d1117', p: '#161b22', p2: '#12171e', ch: '#21262d', l: '#30363d', tx: '#c9d1d9', ac: '#58a6ff' } },
  { id: 'github-light', name: 'GitHub Light', type: 'light', c: { bg: '#ffffff', p: '#f6f8fa', p2: '#fbfcfd', ch: '#eff2f5', l: '#d0d7de', tx: '#24292f', ac: '#0969da' } },
  { id: 'nord', name: 'Nord', type: 'dark', c: { bg: '#2e3440', p: '#3b4252', p2: '#353c4a', ch: '#434c5e', l: '#434c5e', tx: '#d8dee9', ac: '#88c0d0' } },
  { id: 'dracula', name: 'Dracula', type: 'dark', c: { bg: '#282a36', p: '#2f313d', p2: '#31333f', ch: '#44475a', l: '#44475a', tx: '#f8f8f2', ac: '#bd93f9' } },
  { id: 'one-dark', name: 'One Dark', type: 'dark', c: { bg: '#21252b', p: '#282c34', p2: '#242830', ch: '#2c313c', l: '#3e4451', tx: '#abb2bf', ac: '#61afef' } },
  { id: 'monokai', name: 'Monokai', type: 'dark', c: { bg: '#272822', p: '#31322a', p2: '#2c2d26', ch: '#3e3d32', l: '#49483e', tx: '#f8f8f2', ac: '#a6e22e' } },
  { id: 'gruvbox-dark', name: 'Gruvbox Dark', type: 'dark', c: { bg: '#282828', p: '#32302f', p2: '#2d2b29', ch: '#3c3836', l: '#504945', tx: '#ebdbb2', ac: '#fe8019' } },
  { id: 'solarized-dark', name: 'Solarized Dark', type: 'dark', c: { bg: '#002b36', p: '#073642', p2: '#04303c', ch: '#0a3d49', l: '#12495a', tx: '#93a1a1', ac: '#2aa198' } },
  { id: 'solarized-light', name: 'Solarized Light', type: 'light', c: { bg: '#fdf6e3', p: '#fffcf2', p2: '#f7f0dd', ch: '#eee8d5', l: '#ded6bd', tx: '#586e75', ac: '#268bd2' } }
];

// Source: StorCat 1a Demo.dc.html:618-636 — lum() + varsFor(), verbatim
const lum = (hex) => {
  const h = hex.replace('#', '');
  const v = [0, 2, 4].map(i => parseInt(h.slice(i, i + 2), 16) / 255)
    .map(x => x <= 0.03928 ? x / 12.92 : Math.pow((x + 0.055) / 1.055, 2.4));
  return 0.2126 * v[0] + 0.7152 * v[1] + 0.0722 * v[2];
};
const mix = (a, pct, b) => 'color-mix(in oklab, ' + a + ' ' + pct + '%, ' + b + ')';
const varsFor = (t, rowH) => {
  const c = t.c;
  return {
    '--bg': c.bg, '--p': c.p, '--p2': c.p2, '--ch': c.ch, '--l': c.l,
    '--l2': mix(c.l, 55, c.p), '--tx': c.tx,
    '--dm': mix(c.tx, 66, c.bg), '--fn': mix(c.tx, 44, c.bg),
    '--sel': mix(c.ac, 14, 'transparent'), '--hov': mix(c.tx, 8, c.bg),
    '--ac': c.ac, '--acs': mix(c.ac, 16, 'transparent'),
    '--onac': lum(c.ac) > 0.45 ? '#0b0e13' : '#ffffff',
    '--rh': rowH
  };
};
```

**Verified mapping against `frontend/src/themes.ts`'s existing 11 `id`s** (direct comparison, this session): `storcat-light`, `storcat-dark`, `dracula`, `solarized-dark`, `solarized-light`, `nord`, `one-dark`, `monokai`, `github-light`, `github-dark`, `gruvbox-dark` — **exact 1:1 set match** with `THEMES`' 11 ids above (order differs, set is identical). **No theme in `themes.ts` lacks a `THEMES` counterpart, and no `THEMES` entry lacks a `themes.ts` counterpart.**

### Density branch — verified verbatim (THEME-04)

```javascript
// Source: StorCat 1a Demo.dc.html:855-862, verbatim
const density = s.density ?? this.props.density ?? 'Compact';
const cmp = density === 'Compact';
const vars = varsFor(theme, cmp ? '27px' : '34px');
vars['--rp'] = cmp ? '6px 8px' : '10px 10px';
vars['--mp'] = cmp ? '6px' : '10px';
vars['--hp'] = cmp ? '7px 14px' : '11px 14px';
vars['--fs'] = cmp ? '12px' : '13px';
```
Note the prototype's own default is `'Compact'` when no density prop/state is set — but CONTEXT.md's locked fresh-install default for StorCat is **`Comfortable`**, explicitly diverging from the prototype's internal default. Do not copy the `?? 'Compact'` fallback verbatim; use `Comfortable` as the fallback per the locked decision.

### Font vendoring — verified file sizes and acquisition path (THEME-05)

Verified this session by actually running `npm pack @fontsource/ibm-plex-sans@5.3.0` / `@fontsource/ibm-plex-mono@5.3.0`, extracting the tarballs, and measuring the exact latin-subset files needed:

| File | Size (bytes, measured) | CSS declaration source (verified, `latin-400.css` etc. from the extracted package) |
|------|--------------------------|----------------------------------------------------------------------------------------|
| `ibm-plex-sans-latin-400-normal.woff2` | 22,588 | `font-weight: 400; font-style: normal;` — no `unicode-range` restriction (fontsource pre-splits by subset at the file level, so a shipped "latin" file needs no `unicode-range` — there's no other subset file present for the browser to choose between) |
| `ibm-plex-sans-latin-500-normal.woff2` | 24,184 | `font-weight: 500;` |
| `ibm-plex-sans-latin-600-normal.woff2` | 24,252 | `font-weight: 600;` |
| `ibm-plex-mono-latin-400-normal.woff2` | 14,708 | `font-weight: 400;` |
| `ibm-plex-mono-latin-500-normal.woff2` | 14,888 | `font-weight: 500;` |

**Total: ~100.6KB for all 5 files** — small relative to the app's existing binary size concerns (v3.0.0's core value proposition is a 93% smaller binary than the Electron predecessor).

**Acquisition steps (concrete, reproducible):**
```bash
# In a scratch directory, not the repo:
npm pack @fontsource/ibm-plex-sans@5.3.0 --silent
npm pack @fontsource/ibm-plex-mono@5.3.0 --silent
tar -xzf fontsource-ibm-plex-sans-5.3.0.tgz
tar -xzf fontsource-ibm-plex-mono-5.3.0.tgz
# Copy exactly these 5 files into frontend/src/assets/fonts/:
#   package/files/ibm-plex-sans-latin-400-normal.woff2
#   package/files/ibm-plex-sans-latin-500-normal.woff2
#   package/files/ibm-plex-sans-latin-600-normal.woff2
#   package/files/ibm-plex-mono-latin-400-normal.woff2
#   package/files/ibm-plex-mono-latin-500-normal.woff2
# Copy package/LICENSE (OFL-1.1, IBM Corp. copyright) into
#   frontend/src/assets/fonts/IBM-Plex-OFL.txt (do not overwrite the existing Nunito OFL.txt)
# Delete the .tgz/extracted scratch files — do not commit them, do not add
# @fontsource/* to package.json.
```

**`@font-face` declarations** (pattern matches the existing Nunito precedent in `frontend/src/style.css:15-21`, which is the only prior `@font-face` in this codebase — confirmed by `grep -rl "font-face"` across `frontend/src`):
```css
/* Source: pattern verified against frontend/src/style.css's existing Nunito @font-face,
   plus fontsource's latin-400.css/latin-500.css/latin-600.css (extracted this session) */
@font-face {
  font-family: 'IBM Plex Sans';
  font-style: normal;
  font-weight: 400;
  font-display: swap;
  src: url('./assets/fonts/ibm-plex-sans-latin-400-normal.woff2') format('woff2');
}
@font-face {
  font-family: 'IBM Plex Sans';
  font-style: normal;
  font-weight: 500;
  font-display: swap;
  src: url('./assets/fonts/ibm-plex-sans-latin-500-normal.woff2') format('woff2');
}
@font-face {
  font-family: 'IBM Plex Sans';
  font-style: normal;
  font-weight: 600;
  font-display: swap;
  src: url('./assets/fonts/ibm-plex-sans-latin-600-normal.woff2') format('woff2');
}
@font-face {
  font-family: 'IBM Plex Mono';
  font-style: normal;
  font-weight: 400;
  font-display: swap;
  src: url('./assets/fonts/ibm-plex-mono-latin-400-normal.woff2') format('woff2');
}
@font-face {
  font-family: 'IBM Plex Mono';
  font-style: normal;
  font-weight: 500;
  font-display: swap;
  src: url('./assets/fonts/ibm-plex-mono-latin-500-normal.woff2') format('woff2');
}
```
**Vite/build note:** No `vite.config.ts` changes are needed — the existing minimal config (`{ plugins: [react()] }`, confirmed by direct read, no `root`/`assetsInclude`/`publicDir` overrides) already handles `url()`-referenced assets under `src/` via Vite's default asset pipeline (content-hashed filename, copied into `dist/assets/`), exactly as it already does for the vendored Nunito woff2 referenced from `style.css`. Since `main.go` already does `//go:embed all:frontend/dist`, any file Vite places under `dist/` — including the new hashed font files — is automatically embedded in the Go binary with zero `main.go`/`app.go` changes. This is a **verified, not assumed**, zero-config path because it's the exact mechanism the existing Nunito font already uses successfully in this repo today.

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|---------------|-------------------|-----------------|--------|
| Runtime `color-mix()` for derived theme tokens (what the prototype itself does) | Precompute all derived tokens in TypeScript at theme-apply time | This decision, locked in CONTEXT.md this milestone | Removes the Linux WebKitGTK version floor as a shipping risk entirely, at the cost of a manual OKLab-vs-RGB-average fidelity tradeoff (see Pattern 1 above) |
| Google Fonts CDN `<link>` (what the prototype's own `<head>` does — `fonts.googleapis.com`) | Self-hosted, latin-subset, vendored woff2 | This decision, locked in CONTEXT.md this milestone (THEME-05: "no network access") | Zero runtime network dependency for typography; ~100KB added to the binary, offset against the milestone's binary-size goals |

**Deprecated/outdated:** None specific to this phase's tech surface — Wails v2.10.2 is the pinned version (`go.mod`), and the milestone's own requirements explicitly exclude a Wails v3 migration ("Still alpha, premature," per REQUIREMENTS.md's Out of Scope table).

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|-----------------|
| A1 | `@fontsource/ibm-plex-sans`/`@fontsource/ibm-plex-mono` are the correct, legitimate npm packages to source vendored files from (package names recalled from training knowledge, not read from an official doc this session, despite `npm view`/`npm pack` succeeding against the real registry) | Package Legitimacy Audit, IBM Plex Font Vendoring | Low — even if a better/more-official source exists (e.g. IBM's own `IBM/plex` GitHub release, or Google Fonts' served subset), the *files themselves* were verified this session (measured byte-for-byte, license confirmed OFL-1.1, weights confirmed match THEME-05's exact spec) and are functionally interchangeable with any other correctly-subsetted IBM Plex source; worst case is a slightly different vendoring workflow, not wrong output |
| A2 | The ~78px macOS toolbar left-inset value (from CONTEXT.md) is correct/sufficient to clear the real traffic-light cluster under `TitleBarHiddenInset` | Pattern 2 (Window Chrome) | Medium — this number is a CONTEXT.md locked decision, not independently re-derivable from Wails' Go API (traffic-light geometry is native macOS/Cocoa rendering, not exposed as a Wails option or documented pixel value); if visually wrong on a real macOS build, it needs empirical adjustment against an actual running app, not a spec re-read |
| A3 | Simple RGB-channel-average blending for `mixHex()`/`mixAlpha()` is an acceptable substitute for the prototype's `color-mix(in oklab, ...)` for the derived tokens (`--l2 --dm --fn --sel --hov --acs`) | Pattern 1 (Code Examples) | Medium — a visibly "off" subtle tint (selected-row fill, hover fill) relative to the prototype on some of the 11 themes if OKLab and RGB-average diverge enough at those specific blend percentages; THEME-01/02's "matches the handoff" bar is pixel-level, so this should be explicitly resolved (verify against the prototype screenshot per-theme, or implement the OKLab math) rather than assumed acceptable |
| A4 | `scrollbar-gutter: stable` degrades gracefully (silently ignored) rather than erroring on unsupported WebKitGTK, based on general CSS-parser behavior for unrecognized property values, not a WebKitGTK-specific test | Common Pitfalls | Low — standard CSS parsing behavior (unrecognized declarations are dropped, not fatal) is well-established across all CSS engines; if wrong, worst case is the pre-existing PITFALLS.md #26 scrollbar-width inconsistency persisting on old-Linux, not a new breakage |

## Open Questions

1. **OKLab-accurate vs. RGB-average color mixing for derived tokens**
   - What we know: the prototype uses `color-mix(in oklab, ...)`, banned for Linux WebKitGTK-version reasons; a TS port is required.
   - What's unclear: whether the planner should implement full OKLab conversion (byte-for-byte visual fidelity, more code) or accept simple RGB averaging (simpler, small color divergence possible) — CONTEXT.md doesn't specify which, only that computation must happen in TS.
   - Recommendation: implement OKLab conversion if THEME-01/02's "matches the handoff" verification is done by pixel-diffing against the prototype screenshots (likely, given this is a pixel-final design contract); a small, well-tested `oklab.ts` (linearize sRGB → OKLab matrix → lerp → inverse matrix → re-encode sRGB, ~30 lines, no dependency needed) is not large enough to justify the RGB-average shortcut given the fidelity bar this milestone has set everywhere else.

2. **Exact darwin-only inset padding mechanism: CSS media query vs. JS-set custom property**
   - What we know: `Environment().platform === 'darwin'` is the verified detection mechanism; CONTEXT.md specifies "~78px reserved left inset on macOS."
   - What's unclear: whether to gate this via a one-time async `Environment()` call setting a CSS custom property (as sketched in Pattern 2), or via a `data-platform="darwin"` attribute set the same way and matched with a CSS attribute selector — both work, this is an implementation-detail choice within Claude's Discretion (CONTEXT.md: "exact component file names and decomposition") that doesn't need to be pre-decided here.
   - Recommendation: either is fine; the async nature of `Environment()` means there will be a few-millisecond window where the inset isn't applied yet on macOS — acceptable for a launch-time layout detail (unlike the theme-token FOUC concern, this doesn't repaint already-visible colored content, just a few pixels of toolbar padding), so it does not need the same synchronous-before-first-paint treatment as Pattern 1.

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|--------------|-----------|---------|----------|
| Go toolchain | Build (`main.go`/`app.go` change for `mac.TitleBarHiddenInset`) | ✓ | 1.23 (per `go.mod`), module cache confirmed present locally | — |
| Wails v2 CLI/module | Build | ✓ | v2.10.2 (`go.mod`, module cache verified) | — |
| Node/npm | Frontend build, font extraction | ✓ | npm confirmed functional this session (`npm view`, `npm pack` succeeded) | — |
| Vite | Frontend bundling | ✓ | existing `frontend/package.json` devDependency, config read directly | — |
| Network access (one-time, dev-machine only) | `npm pack` font extraction | ✓ (confirmed — registry calls succeeded this session) | — | If offline: fall back to manually downloading the same woff2 files via `fonts.googleapis.com`'s CSS response (inspect the `url()` values in the response the prototype's own `<link>` tag would fetch) — same end files, more manual steps |
| macOS build target | Verifying `TitleBarHiddenInset`/traffic-light inset visually | Not verified this session (research is code/doc-based) | — | Verification against a real macOS build is required before considering SHELL-08 done — flag as a manual/human-verify step in the plan, this cannot be confirmed by static research |

**Missing dependencies with no fallback:** none.
**Missing dependencies with fallback:** network access for font extraction (fallback: manual Google Fonts CSS inspection, same output files).

## Validation Architecture

### Test Framework

| Property | Value |
|----------|-------|
| Framework | **None detected** — confirmed by reading `frontend/package.json` (no `vitest`/`jest`/`@testing-library/*` in dependencies or devDependencies) and by the milestone's own requirements (TEST-01, "Frontend unit tests... Vitest + Testing Library," is a v2/deferred requirement, not in scope for v1) |
| Config file | none — see Wave 0 |
| Quick run command | none available |
| Full suite command | none available |

Given TEST-01 is explicitly deferred and no test framework exists, **most of this phase's verification is human-visual and manual-interaction**, not automated. This is an honest constraint, not a gap to silently work around — do not introduce Vitest just for this phase (out of scope; TEST-01 is its own future milestone item) and do not claim automated coverage that doesn't exist.

### Phase Requirements → Test Map

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|----------------------|---------------|
| SHELL-01 | Single view renders, no tabs | manual-only (visual) | — | ❌ no test infra |
| SHELL-02 | `268px 1fr 288px` grid at ≥1280px | manual-only (resize + inspect computed style, or DevTools) | — | ❌ |
| SHELL-03/04 | Responsive tier swaps at 1279px/1040px boundaries | manual-only (resize through both breakpoints) | — | ❌ |
| SHELL-05 | Rail-side swap moves divider | manual-only (toggle rail-side, e.g. via a temporary devtools localStorage edit since no Settings UI exists yet) | — | ❌ |
| SHELL-07 | Toolbar drag works; search/theme/gear/Details clicks register | manual-only (click each control; drag from empty toolbar space; **must be tested on an actual Windows build per PITFALLS.md #12's own warning** — click-swallowing is "most visibly on Windows") | — | ❌ |
| SHELL-08 | macOS traffic lights inset correctly; Windows/Linux native title bar | manual-only, **requires a real macOS build** — cannot be verified from source reading alone (see Environment Availability) | — | ❌ |
| SHELL-09 | Overlay z-index ordering (only the details drawer exists this phase — full stacking-order check is for Phase 24+) | manual-only, partial this phase (no palette/dialog yet to stack against) | — | ❌ |
| THEME-01 | All 11 themes repaint immediately | manual-only — cycle all 11 via a temporary trigger (Settings doesn't exist yet; likely a devtools console call or temporary dev-only theme-switcher) | — | ❌ |
| THEME-02 | `--onac` legible on every theme's accent-filled elements ("+ New" pill hover, tree-empty CTA) | manual-only — **must visually check all 11 themes**, not just the default, per PITFALLS.md #23's own explicit warning ("the bug only appears when cycling through all 11 themes") | — | ❌ |
| THEME-03 | 14 tokens present on `:root` per theme | manual-only (DevTools "Computed" styles inspection) — or a throwaway `console.log(getComputedStyle(document.documentElement))` check during dev, not a committed test | — | ❌ |
| THEME-04 | Density changes row height/padding/font-size | manual-only — set `state.density` via reducer/devtools per UI-SPEC's own note (no toggle UI ships this phase) | — | ❌ |
| THEME-05 | IBM Plex renders with no network access | manual-only — **disconnect network and reload** (or DevTools "offline" mode) and visually confirm fonts still render correctly (not falling back to a system font) | — | ❌ |
| THEME-06 | Theme/density/rail-side survive restart | manual-only — set values, fully quit and relaunch the app (not just a page reload — verify against the real Wails binary, not `vite dev`) | — | ❌ |

### Sampling Rate
- **Per task commit:** manual visual check against the running `wails dev` build (no automated quick-run exists)
- **Per wave merge:** full manual pass through the requirements table above, plus a side-by-side comparison against `StorCat 1a Demo.dc.html` open in a browser (per PITFALLS.md #25's recommendation)
- **Phase gate:** all manual checks above pass; SHELL-07/08 specifically flagged as needing a real Windows build and a real macOS build respectively before being considered done — a macOS-only dev machine cannot self-verify Windows drag-swallowing or confirm the Linux WebKitGTK degradation is graceful

### Wave 0 Gaps
- No test framework exists and none is being added this phase (TEST-01 deferred) — this is a known, accepted gap, not an oversight to fix in Wave 0.
- Recommend the plan include an explicit **temporary dev-only affordance** for exercising untriggerable states (all 11 themes without a working theme picker, density without a working toggle, rail-side without a working setting) — e.g., a throwaway keyboard shortcut or a small dev-panel gated behind `import.meta.env.DEV`, removed or superseded once Phase 26's real Settings ships. This is Claude's Discretion territory (not locked by CONTEXT.md) but flagged here because several of the requirements above are otherwise **unverifiable by a human** in this phase's built UI (no theme picker, no density toggle, no rail-side control ship this phase per the locked scope).

## Security Domain

`security_enforcement` is absent from `.planning/config.json` → treated as enabled per policy, but this phase's actual attack surface is minimal: a static, data-free UI shell with no auth, no user-generated content rendering, no network calls, and no new IPC surface (confirmed — no Go changes this phase).

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control |
|-----------------|---------|----------------------|
| V2 Authentication | No | Desktop app, no auth surface exists or is touched this phase |
| V3 Session Management | No | N/A — no sessions |
| V4 Access Control | No | N/A — single-user local desktop app |
| V5 Input Validation | Minimal | The rail-filter and search-field inputs are rendered but **functionally inert this phase** (no filtering/search logic wired until Phase 23/24 per the Copywriting Contract) — no validation surface exists yet to secure. When wired in later phases, standard React controlled-input handling (no `dangerouslySetInnerHTML`, no raw HTML injection from user-typed strings) applies. |
| V6 Cryptography | No | N/A — no crypto operations in this phase |
| V11 (Business Logic, in spirit) | Minimal | Theme/density/rail-side values read from localStorage must be validated against the known-enum set before being trusted (see THEME-06's error case below) — this is the one place user-controllable (or corrupted) local storage data feeds into rendering logic this phase. |

### Known Threat Patterns for this stack

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|-------------------------|
| Corrupt/tampered `localStorage` value for `storcat-theme-id`/`storcat-density`/`storcat-rail-side` causing a crash or unstyled render | Tampering (low severity — local-only, single-user) | UI-SPEC E6 already specifies this: "A missing, unknown, or corrupt... value falls back to the locked defaults... and rewrites a valid value. Never throws, never renders untokenized." Implement via a strict allowlist check (`themes.find(t => t.id === stored)`, `density === 'Compact' \|\| density === 'Comfortable'`, `side === 'Left' \|\| side === 'Right'`) rather than trusting the stored string directly. |
| `CatalogModal`'s `srcDoc={htmlContent}` iframe rendering raw HTML read from disk (`readHtmlFile`) | Tampering / injection (pre-existing, not introduced by this phase) | Out of scope for Phase 22 — `CatalogModal.tsx` is unchanged and unreachable this phase (see Removal Surface); flagging only because it's adjacent code that stays in the diff. The catalog HTML is StorCat's own generated output (trusted, first-party), not arbitrary user/network content, so this is a pre-existing, accepted design, not a new risk this phase introduces. |

## Sources

### Primary (HIGH confidence — read/executed directly this session)
- `go doc -all github.com/wailsapp/wails/v2/pkg/options/mac` and direct `Read` of `pkg/options/mac/titlebar.go`, `pkg/options/mac/mac.go`, `pkg/options/options.go`, `pkg/runtime/runtime.go` (module `github.com/wailsapp/wails/v2@v2.10.2`, resolved via `go mod download` + local module cache at `/Users/ken/go/pkg/mod`)
- Direct `Read`/`grep` of `internal/frontend/runtime/desktop/main.js` (same module) — verified `dragTest`/`cssDragProperty`/`cssDragValue` mechanism
- Direct `Read`/`grep` of `design_handoff_storcat_ui/designs/StorCat 1a Demo.dc.html` (lines 604-636 THEMES/lum/varsFor, 840-862 density branch, 999-1012 responsive-tier computed properties, plus layout/font/CSS reset lines) — all quoted verbatim above
- Direct `Read` of `frontend/src/App.tsx`, `themes.ts`, `contexts/AppContext.tsx`, `index.css`, `style.css`, `main.tsx`, `components/Header.tsx`, `components/MainContent.tsx` (partial), `components/CatalogModal.tsx`, `components/tabs/*.tsx` (import lines), `frontend/vite.config.ts`, `frontend/tsconfig.json`, `frontend/index.html`, `frontend/package.json`, `frontend/wailsjs/runtime/runtime.d.ts`, `frontend/wailsjs/go/main/App.d.ts`, `main.go`, `app.go` (partial), `wails.json`, `go.mod`
- `npm pack @fontsource/ibm-plex-sans@5.3.0` / `@fontsource/ibm-plex-mono@5.3.0` executed this session, tarballs extracted, exact file sizes measured with `ls -la`, `LICENSE`/`latin-*.css` contents read directly
- `node gsd-tools.cjs query package-legitimacy check` executed against both `@fontsource/*` packages

### Secondary (MEDIUM confidence — web-sourced, cross-checked)
- SIL OFL-1.1 bundling/embedding/subsetting obligations — [choosealicense.com/licenses/ofl-1.1](https://choosealicense.com/licenses/ofl-1.1/), [openfontlicense.org/how-to-use-ofl-fonts](https://openfontlicense.org/how-to-use-ofl-fonts/)
- `scrollbar-gutter: stable` browser support timeline — [caniuse.com/mdn-css_properties_scrollbar-gutter_stable](https://caniuse.com/mdn-css_properties_scrollbar-gutter_stable), [MDN scrollbar-gutter](https://developer.mozilla.org/en-US/docs/Web/CSS/Reference/Properties/scrollbar-gutter)
- `.planning/research/PITFALLS.md` and `.planning/research/ARCHITECTURE.md` — milestone-level research, read directly, cited by pitfall/section number rather than re-derived

### Tertiary (LOW confidence)
- The `@fontsource/ibm-plex-sans`/`@fontsource/ibm-plex-mono` package names themselves — recalled from training knowledge, not read from an official doc this session (registry existence/contents were independently verified via `npm view`/`npm pack`, but per the package-name-provenance rule this doesn't upgrade the *name recall* to VERIFIED — see Assumptions Log A1)

## Metadata

**Confidence breakdown:**
- Standard stack / no-new-dependencies: HIGH — verified by reading `package.json` and confirming zero new runtime deps are needed
- Window chrome (Wails Mac/drag mechanics): HIGH — verified against the actual installed Go module source, not documentation summaries
- Theme data (THEMES/lum/varsFor/density): HIGH — verified verbatim against the prototype file, with exact line numbers
- Font vendoring: HIGH for file sizes/acquisition mechanics (measured directly), MEDIUM for OFL legal interpretation (web-sourced)
- Removal surface / dead-code mapping: HIGH — verified by direct grep/read of every file in the deletion set and its importers
- Color-mix TS-port fidelity (OKLab vs RGB-average): MEDIUM — flagged as an open question, not resolved by this research
- macOS traffic-light inset (~78px) exact correctness: LOW/MEDIUM — inherited as a locked CONTEXT.md value, not independently derivable from Wails' API surface; needs real-device visual verification

**Research date:** 2026-08-13
**Valid until:** Wails v2.10.2 pin and the prototype file are both stable, unversioned project assets — this research does not go stale on a calendar basis the way a fast-moving web framework would; re-verify only if `go.mod`'s Wails version changes or the design handoff is revised.
