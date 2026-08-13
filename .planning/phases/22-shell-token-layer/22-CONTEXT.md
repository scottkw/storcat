# Phase 22: Shell + Token Layer - Context

**Gathered:** 2026-08-13
**Status:** Ready for planning
**Mode:** Smart discuss (autonomous)

<domain>
## Phase Boundary

Replace the three-tab Ant Design interface with the 1a Workspace shell: a 46px toolbar, catalog rail, tree pane, details panel, and 26px status bar in a single view — responsive across three width tiers, with the handoff's extended design-token set driving all 11 themes, row density, and vendored IBM Plex typography.

**In scope:** shell layout and grid, responsive tiers and the details drawer, rail-side swap, window chrome / drag regions, the token layer (14 CSS custom properties + density vars), theme application across all 11 themes, `--onac` luminance contrast helper, IBM Plex font vendoring, and persistence of theme / density / rail position.

**Out of scope (later phases):** catalog rail data and filtering, virtualized tree, details panel content, status bar counts (SHELL-06 → Phase 23), ⌘K palette (Phase 24), create slide-over (Phase 25), Settings surface (Phase 26). Panes render as structurally correct skeletons in this phase.

</domain>

<decisions>
## Implementation Decisions

### Shell Composition & Migration Scope
- The three-tab UI is deleted in this phase — `App.tsx` renders only the new workspace; `Header.tsx`, `MainContent.tsx`, `WelcomeContent.tsx`, and `components/tabs/*` are removed rather than kept behind a flag or left unrendered.
- Ant Design is removed from every surface built in this phase, but the `antd` / `@ant-design/icons` dependencies stay installed until their last consumer (`CatalogModal.tsx`) is replaced in Phase 26. No big-bang rewrite in the first phase.
- Rail, tree, and details panes render as static skeletons in Phase 22 — correct dimensions, tokens, borders, and empty-state messaging, with no data wiring. Data wiring belongs to Phase 23.
- Styling is plain CSS with CSS custom properties in a dedicated `workspace.css`, matching the handoff's token model and the existing `index.css` pattern. No CSS-in-JS, no CSS Modules, no Tailwind (explicitly out of scope for the project).

### Token & Theme Layer
- All 14 tokens are applied on `:root` via the successor to the existing `applyTheme()` function — one code path, and existing `var(--…)` consumers keep working through the migration. Not a wrapper `<div>` (only needed if two themes coexist, which they don't).
- Derived tokens (`--l2 --dm --fn --acs --sel --hov --onac`) are computed in TypeScript as concrete `rgb()` values at theme-apply time. CSS `color-mix()` is explicitly rejected — Wails does not control WebKitGTK's version on Linux (see `.planning/research/PITFALLS.md`).
- `--onac` is computed from relative luminance (`> 0.45 → dark text`), ported from the handoff helper. This is what keeps Gruvbox orange / Monokai green (light accents) and GitHub blue (dark accent) all legible (THEME-02).
- `themes.ts` gains a `tokens` block per theme (`bg p p2 ch l tx ac`) sourced from the handoff's authoritative `THEMES` array; the legacy `colors` block stays until the last antd surface is retired in Phase 26, then gets dropped.
- IBM Plex Sans (400/500/600) and IBM Plex Mono (400/500) are vendored as self-hosted **latin-subset** woff2 in `frontend/src/assets/fonts/`, declared via `@font-face`, bundled by Vite into the embedded binary. No CDN, no network access (THEME-05). Latin subset chosen to control binary size.

### Responsive Layout & Window Chrome
- Breakpoints are driven by CSS media queries swapping the grid template (`268px 1fr 288px` at ≥1280px → `236px 1fr` at 1040–1279px → `200px 1fr` below 1040px). A small `useMediaQuery` hook is used only where React genuinely needs to know the tier (details rendered as pane vs drawer). No pure-JS resize listener; no container queries (same WebKitGTK version risk as `color-mix()`).
- Below 1280px the details panel is the *same component* rendered into a fixed-position right drawer with a backdrop, closable via Esc and backdrop click, toggled by the toolbar "Details" chip. No duplicate mobile-only component.
- Overlay stacking (SHELL-09) uses one documented z-index scale declared as CSS vars in `workspace.css`: details drawer 100 → create slide-over / ⌘K palette 200 → dialogs / Settings 300. Later phases slot into this scale rather than inventing numbers.
- macOS uses `TitleBarHiddenInset` in the Go app options (darwin only) so the real traffic lights sit inside the 46px toolbar, with ~78px reserved left inset on macOS; Windows and Linux keep the native title bar above the toolbar. Window drag uses `--wails-draggable` on the toolbar background, with the search field, theme chip, and gear explicitly opted out so clicks are not swallowed (SHELL-07). Frameless Windows/Linux chrome is FUT-02 — deferred, not in this milestone.

### Preferences & State
- Theme, density, and rail position persist to localStorage (THEME-06): reuse the existing `storcat-theme-id`, add `storcat-density` and `storcat-rail-side`. Go `preferences.json` remains for window state only — no IPC round-trip for a UI-only concern.
- Workspace state (density, rail side, `detailOverlay`, overlay flags) extends the existing `AppContext` reducer. No new context, no new state-management dependency.
- The old tab UI's localStorage keys (`storcat-last-*`, sidebar position) are left untouched in this phase — they are inert once the tabs are gone. They get swept in Phase 26 when the Settings surface is built.
- Fresh-install defaults: theme `storcat-light` (unchanged from today — new users are not silently flipped to dark), density `Comfortable`, rail side `Left`.

### Claude's Discretion
- Exact component file names and decomposition within the shell.
- Precise spacing/radius/shadow values within the handoff's documented scales.
- Skeleton/empty-state copy for the not-yet-wired panes.
- Structure of the TypeScript color-mix helper module.

</decisions>

<code_context>
## Existing Code Insights

### Reusable Assets
- `frontend/src/themes.ts` — 11 `Theme` entries with `id`, `name`, `type`, `colors`; extend with a `tokens` block rather than replace.
- `frontend/src/App.tsx` `applyTheme()` — already writes CSS custom properties to `document.documentElement`; this is the hook point for the new token layer, including its existing `storcat-theme` → `storcat-theme-id` migration path.
- `frontend/src/contexts/AppContext.tsx` — `useReducer` + Context with `UPPER_SNAKE_CASE` action types; extend for workspace state.
- `frontend/src/services/wailsAPI.ts` — the `window.electronAPI`-shaped bridge; untouched by this phase but consumed from Phase 23 on.
- `frontend/src/storcat-icon.svg` — the app mark, to be dropped into the toolbar at 16–18px (the 200px hero logo is gone).
- `frontend/src/index.css` — existing `:root` custom-property defaults and `::-webkit-scrollbar` styling; the antd `!important` overrides in it become dead as antd surfaces are removed.

### Established Patterns
- Function-declaration React components, `PascalCase.tsx`, default export at the bottom.
- 2-space indent, single quotes, semicolons, trailing commas; TypeScript strict with `noUnusedLocals` / `noUnusedParameters`.
- Theming through CSS custom properties set at runtime on `:root` — not a theme provider.
- Cross-component signalling via `CustomEvent` on `window` (`themeChange`, `openCatalogModal`) — the tab-era mechanism; workspace state moves into the reducer instead.
- localStorage for UI preferences; Go-side JSON files for window state and app preferences.

### Integration Points
- `frontend/src/App.tsx` — replaced wholesale as the workspace shell root.
- Go app options (`main.go` / Wails options) — `TitleBarHiddenInset` for darwin.
- `vite.config.ts` — font asset handling for the vendored woff2 files.
- `frontend/package.json` — antd stays for now; fonts are vendored files, not a dependency.

### Reference Material
- `design_handoff_storcat_ui/README.md` — §"Screens / views → 1. Workspace", §"Interactions & behavior", §"State management", §"Design tokens", §"Assets". The authoritative spec for this phase.
- `design_handoff_storcat_ui/designs/StorCat 1a Demo.dc.html` — the working prototype, including the `THEMES` array and the `--onac` luminance helper to port.

</code_context>

<specifics>
## Specific Ideas

- Match the handoff exactly on dimensions: 46px toolbar, 26px status bar, `268px 1fr 288px` grid, 1px dividers, tree row height 27px (Compact) / 34px (Comfortable).
- Port the `--onac` luminance helper verbatim in spirit — the handoff calls it out as ~6 lines that prevent the light-accent contrast bug.
- Type scale in use: 10.5, 11, 11.5, 12, 12.5, 13, 14, 15, 17, 26px. Titles carry `letter-spacing: -0.01em`; uppercase section labels 11–12px/600 with `+0.04–0.05em`.
- Radii: rows 6, buttons 7–8, panels/modals 12. Shadows: menu `0 18px 40px rgba(0,0,0,.5)`, modals `0 30px 70px rgba(0,0,0,.6)`, drawers `-30px 0 70px rgba(0,0,0,.5)`.
- Status colors are theme-independent: error `#e5534b`, warning `#f0b429`, success = `--ac`.
- Icons are inline SVG (magnifier, folder/card, plus, gear, caret) at 10–15px, `stroke-width: 1.4–1.8`, `stroke: currentColor`. No icon font, no image assets.

</specifics>

<deferred>
## Deferred Ideas

- Dropping the legacy `ThemeColors` block and uninstalling `antd` / `@ant-design/icons` — Phase 26, once `CatalogModal` is replaced.
- Sweeping the obsolete `storcat-last-*` and sidebar-position localStorage keys — Phase 26 Settings.
- Rail-as-drawer below ~820px (FUT-01), frameless Windows/Linux chrome (FUT-02) — deferred out of v3.0.0 at requirements definition.
- Frontend unit tests for the shell (TEST-01) — deferred to a separate testing milestone.

</deferred>
