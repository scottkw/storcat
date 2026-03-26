# Architecture

**Analysis Date:** 2026-03-24

## Pattern Overview

**Overall:** Electron desktop application with a classic Main Process / Renderer Process split. The main process (Node.js/CommonJS) handles filesystem operations and native dialogs. The renderer process (React/TypeScript, bundled by Vite) provides the UI.

**Key Characteristics:**
- Context isolation enforced -- renderer communicates with main only through a preload bridge (`window.electronAPI`)
- All filesystem and catalog logic lives in the main process; the renderer is purely presentational
- State management via React `useReducer` + Context, supplemented by `localStorage` for persistence
- No router -- single-window app with tab-based navigation (Create, Search, Browse)
- Theming via CSS custom properties set at runtime from a theme registry

## Layers

**Main Process (Electron / Node.js / CommonJS):**
- Purpose: Application lifecycle, native OS integration, filesystem catalog operations
- Location: `src/main.js`
- Contains: Window creation, IPC handler registration, window state persistence, preferences management
- Depends on: `src/catalog-service.js`, `src/preload.js`, Electron APIs, Node `fs`/`path`
- Used by: Renderer process (indirectly, via IPC)

**Preload Bridge:**
- Purpose: Expose a safe, typed API surface from main process to renderer
- Location: `src/preload.js`
- Contains: `contextBridge.exposeInMainWorld('electronAPI', ...)` mapping 12 IPC channels
- Depends on: Electron `contextBridge`, `ipcRenderer`
- Used by: Every renderer component that touches filesystem or native dialogs

**Catalog Service:**
- Purpose: Core business logic -- directory traversal, catalog JSON/HTML generation, search
- Location: `src/catalog-service.js`
- Contains: `createCatalog()`, `searchCatalogs()`, `traverseDirectory()`, `generateTreeHTML()`, helpers
- Depends on: Node `fs.promises`, `path`
- Used by: `src/main.js` (IPC handlers for `create-catalog` and `search-catalogs`)

**Renderer (React / TypeScript):**
- Purpose: All UI rendering, user interaction, local state
- Location: `src/renderer/`
- Contains: React components, context/state, theme definitions, CSS, type declarations
- Depends on: React 18, Ant Design 5, `window.electronAPI` bridge
- Used by: Electron BrowserWindow (loaded via Vite dev server or built `dist/index.html`)

## Data Flow

**Catalog Creation:**
1. User fills sidebar form in `CreateCatalogTab.Sidebar` (title, directory, output root, optional copy dir)
2. Component calls `window.electronAPI.createCatalog(options)` -- IPC invoke to main
3. Main process handler in `src/main.js` delegates to `catalog-service.js` `createCatalog()`
4. `traverseDirectory()` recursively walks the filesystem building a tree of `{type, name, size, contents}`
5. JSON catalog written to `<directoryPath>/<outputRoot>.json`
6. HTML catalog generated via `generateTreeHTML()` and written to `<directoryPath>/<outputRoot>.html`
7. Optionally, both files copied to secondary directory
8. Result object returned through IPC to renderer; `message.success()` shown

**Catalog Search:**
1. User enters search term + selects catalog directory in `SearchCatalogsTab.Sidebar`
2. Component calls `window.electronAPI.searchCatalogs(term, directory)` -- IPC invoke
3. Main process reads all `.json` files in the directory, parses each, runs `searchInCatalog()` recursively
4. Matching items (by case-insensitive name substring) returned as flat array with catalog metadata
5. Results stored in `AppContext` state (`searchResults`), rendered in `ModernTable`

**Catalog Browsing:**
1. User selects directory in `BrowseCatalogsTab.Sidebar`
2. Component calls `window.electronAPI.getCatalogFiles(directory)` -- IPC invoke
3. Main process lists `.json` files, reads each for metadata, checks for companion `.html` files
4. Catalog list stored in `AppContext` state (`browseCatalogs`), rendered in `ModernTable`

**HTML Preview (Modal):**
1. User clicks catalog link in search results or browse table
2. Component dispatches `window.dispatchEvent(new CustomEvent('openCatalogModal', ...))`
3. `App.tsx` listens for this event, sets `catalogModalVisible` + `catalogModalPath`
4. `CatalogModal` loads HTML content via IPC (`getCatalogHtmlPath` then `readHtmlFile`)
5. HTML injected into an `<iframe>` via `srcDoc`, with dark theme CSS injected if needed

**State Management:**
- **Global UI state:** `useReducer` in `src/renderer/contexts/AppContext.tsx` -- holds selected directories, loading flags, search results, browse catalogs, active tab, sidebar state, modal state
- **Theme state:** Managed in `App.tsx` via `useState`, persisted to `localStorage` (`storcat-theme-id`). CSS custom properties applied to `document.documentElement` at runtime.
- **User preferences persistence:** Split between `localStorage` (theme, sidebar position, last-used directories, last search term) and Electron `userData` JSON files (window position/size, window persistence toggle)
- **Cross-component communication:** Custom DOM events (`themeChange`, `openCatalogModal`) used to bridge `MainContent` settings modal to `App.tsx` and tab components to `App.tsx` modal

## Key Abstractions

**Tab Components (Sidebar + Content pattern):**
- Purpose: Each tab exports an object with `.Sidebar` and `.Content` sub-components
- Examples: `src/renderer/components/tabs/CreateCatalogTab.tsx`, `src/renderer/components/tabs/SearchCatalogsTab.tsx`, `src/renderer/components/tabs/BrowseCatalogsTab.tsx`
- Pattern: `MainContent` conditionally renders `<Tab.Sidebar />` in the sidebar and `<Tab.Content />` in the main area based on `state.activeTab`

**ModernTable:**
- Purpose: Reusable data table with sorting, column filtering, pagination, resizable columns
- Examples: `src/renderer/components/ModernTable.tsx`
- Pattern: Accepts `columns` (with optional `render`, `sortable`, `filterable`) and `data` props. Used by SearchCatalogsTab and BrowseCatalogsTab.

**Theme System:**
- Purpose: Multi-theme support (11 built-in themes) via CSS custom properties
- Examples: `src/renderer/themes.ts`
- Pattern: Each `Theme` object defines `id`, `name`, `type` (light/dark), `colors` (16 CSS variable values), `antdAlgorithm`, and optional `antdPrimaryColor`. Applied at runtime by setting CSS variables on `:root`.

**Catalog Data Format:**
- Purpose: Hierarchical directory tree stored as JSON
- Pattern: Recursive `{type: 'file'|'directory', name: string, size: number, contents?: CatalogItem[]}`
- Compatibility: Supports both own format (single root object) and legacy bash script format (array wrapping)

## Entry Points

**Electron Main Process:**
- Location: `src/main.js`
- Triggers: `electron .` or `npm start`
- Responsibilities: Creates `BrowserWindow`, registers all IPC handlers, manages window state/preferences

**Renderer Entry:**
- Location: `src/renderer/main.tsx`
- Triggers: Loaded by `src/renderer/index.html` via `<script type="module">`
- Responsibilities: Mounts React app (`App.tsx`) into `#root`

**Vite Dev Server:**
- Location: `vite.config.ts` (root: `src/renderer`)
- Triggers: `npm run dev:vite` (port 5173)
- Responsibilities: HMR for renderer development

**CLI Catalog Scripts (legacy):**
- Location: `sdcat/sdcat.sh`, `sdcat/sdcat_zsh.sh`
- Triggers: Manual shell execution
- Responsibilities: Original bash-based catalog generation (StorCat is the Electron replacement)

## Error Handling

**Strategy:** Defensive -- errors caught at IPC boundaries and returned as `{ success: false, error: message }` objects. Renderer shows `message.error()` toasts via Ant Design.

**Patterns:**
- Main process IPC handlers wrap all logic in try/catch, returning `{ success: false, error: error.message }` on failure
- Catalog traversal silently skips inaccessible files/directories (permissions, broken symlinks) with `console.warn`
- Window state load/save fails silently with fallback defaults
- Renderer components check `result.success` before processing IPC responses

## Cross-Cutting Concerns

**Logging:** `console.log` / `console.warn` / `console.error` throughout main process. No structured logging framework.

**Validation:** Minimal -- renderer checks for non-empty fields before IPC calls. No schema validation on catalog JSON.

**Authentication:** Not applicable -- local desktop application with no auth.

**Theming:** CSS custom properties set on `document.documentElement` by `App.tsx`. Ant Design theme algorithm (light/dark) configured via `ConfigProvider`. All components use `var(--*)` CSS variables. 11 themes defined in `src/renderer/themes.ts`.

**Persistence:** Hybrid -- `localStorage` for renderer preferences (theme, sidebar, directories), Electron `userData` JSON files for window state and persistence toggle.

**Platform Compatibility:** macOS-optimized (`titleBarStyle: 'hiddenInset'`, traffic light spacing), but builds for Windows and Linux. V8 `--jitless --no-opt` flags for ARM64 stability.

---

*Architecture analysis: 2026-03-24*
