<!-- GSD:project-start source:PROJECT.md -->
## Project

**StorCat v2.0.0 — Go/Wails Migration**

StorCat is a cross-platform desktop application for creating, browsing, and searching directory catalogs. It generates JSON and HTML representations of directory trees. v2.0.0 is a complete backend rewrite from Electron/Node.js to Go/Wails, keeping the same React/TypeScript/Ant Design frontend.

**Core Value:** The Go/Wails v2.0.0 release must have exact feature parity with the Electron v1.2.3 release — no regressions, no missing functionality. Users switching from v1 to v2 should notice only improvements (speed, size), never missing features.

### Constraints

- **Backward compatibility**: v1.0 catalog JSON/HTML files must remain readable
- **API surface**: `window.electronAPI` interface must be maintained — frontend components depend on it
- **Branch**: Work on `feature/go-refactor-2.0.0-clean`, merge to `main` when complete
- **No Electron dependencies**: All fixes must use Go/Wails patterns, not Electron APIs
<!-- GSD:project-end -->

<!-- GSD:stack-start source:codebase/STACK.md -->
## Technology Stack

## Languages
- JavaScript (ES2020+) - Electron main process (`src/main.js`, `src/preload.js`, `src/catalog-service.js`)
- TypeScript (^5.2.2) - Renderer/React UI (`src/renderer/**/*.tsx`, `src/renderer/**/*.ts`)
- Bash - Build scripts (`build-all.sh`, `build-complete.sh`, `build-fast.sh`, `build-simple.sh`) and legacy CLI tool (`sdcat/sdcat.sh`, `sdcat/sdcat_zsh.sh`)
## Runtime
- Electron ^35.0.0 (Chromium + Node.js)
- Node.js (required by Electron; no `.nvmrc` or `.node-version` detected)
- V8 flags applied at startup: `--jitless --no-opt` (ARM64 memory workaround in `src/main.js` line 4)
- npm (no `pnpm-lock.yaml` or `yarn.lock` detected)
- Lockfile: `package-lock.json` present
## Frameworks
- Electron ^35.0.0 - Desktop application shell, IPC between main/renderer
- React ^18.2.0 - Renderer UI framework
- Ant Design (antd) ^5.12.8 - UI component library (Layout, ConfigProvider, theme, Tabs, Table, Modal, etc.)
- Not detected - No test framework, test config, or test files found
- Vite ^5.0.8 - Bundler for renderer process (`vite.config.ts`)
- @vitejs/plugin-react ^4.2.1 - React Fast Refresh for Vite
- electron-builder ^24.9.1 - Application packaging and distribution
- concurrently ^8.2.2 - Run Vite dev server and Electron simultaneously
- wait-on ^7.2.0 - Wait for Vite dev server before launching Electron
## Key Dependencies
- `react` ^18.2.0 - UI rendering
- `react-dom` ^18.2.0 - DOM binding for React
- `antd` ^5.12.8 - Complete UI component library (tables, modals, forms, tabs, layout)
- `@ant-design/icons` ^5.2.6 - Icon set for Ant Design components
- `electron` ^35.0.0 - Runtime (dev dependency because electron-builder bundles it)
- `electron-builder` ^24.9.1 - Cross-platform packaging (DMG, MSI, AppImage, deb, rpm, snap)
- `typescript` ^5.2.2 - Type checking for renderer code
- `vite` ^5.0.8 - Dev server and bundler
- `@vitejs/plugin-react` ^4.2.1 - React plugin for Vite
- `concurrently` ^8.2.2 - Parallel process runner
- `wait-on` ^7.2.0 - HTTP wait utility
- `@types/react` ^18.2.45 - React type definitions
- `@types/react-dom` ^18.2.18 - ReactDOM type definitions
- `@typescript-eslint/eslint-plugin` ^6.14.0 - ESLint rules for TypeScript
- `@typescript-eslint/parser` ^6.14.0 - TypeScript parser for ESLint
- `eslint` ^8.55.0 - Linting
- `eslint-plugin-react-hooks` ^4.6.0 - React hooks linting
- `eslint-plugin-react-refresh` ^0.4.5 - React Refresh linting
## Configuration
- `tsconfig.json` - Renderer code only (`include: ["src/renderer"]`), target ES2020, strict mode, react-jsx, bundler module resolution
- `tsconfig.node.json` - Vite config only (`include: ["vite.config.ts"]`)
- `vite.config.ts` - Root set to `src/renderer`, builds to `../../dist`, dev server on port 5173
- Configured inline in `package.json` under `"build"` key
- App ID: `com.kenscott.storcat`
- macOS: Universal DMG, hardened runtime, code signing (identity `S2K7P43927`), notarization enabled
- Windows: Portable + MSI for x64 and arm64
- Linux: AppImage, deb, rpm, snap for x64 and arm64
- macOS entitlements: `build/entitlements.mac.plist` (allow-unsigned-executable-memory, disable-library-validation, allow-jit)
- `build/` - Icons and macOS entitlements
- `build/icons/` - App icons for all platforms
- ESLint ^8.55.0 with TypeScript and React plugins installed, but no `.eslintrc*` config file detected at project root
## Platform Requirements
- Node.js (version not pinned)
- npm
- macOS recommended (build scripts assume macOS for cross-platform builds)
- macOS: Universal binary (x64 + arm64) as DMG
- Windows: x64 and arm64 as Portable (.exe) and MSI
- Linux: x64 and arm64 as AppImage, deb, rpm, snap
## NPM Scripts
| Command | Purpose |
|---------|---------|
| `npm start` | Launch Electron directly (production mode) |
| `npm run dev` | Start Vite dev server + Electron with hot reload |
| `npm run build` | Build renderer + package with electron-builder |
| `npm run build:mac` | Build macOS package only |
| `npm run build:win` | Build Windows package only |
| `npm run build:linux` | Build Linux package only |
| `npm run build:all` | Build all platforms via electron-builder |
| `npm run build:complete` | Full build via `build-complete.sh` |
| `npm run version:patch-build` | Bump patch version + complete build |
<!-- GSD:stack-end -->

<!-- GSD:conventions-start source:CONVENTIONS.md -->
## Conventions

## Naming Patterns
- React components: `PascalCase.tsx` (e.g., `MainContent.tsx`, `CatalogModal.tsx`, `ModernTable.tsx`)
- Tab components: `PascalCase.tsx` inside `tabs/` directory (e.g., `CreateCatalogTab.tsx`, `SearchCatalogsTab.tsx`)
- Main process files: `kebab-case.js` (e.g., `catalog-service.js`, `main.js`, `preload.js`)
- Type definition files: `kebab-case.d.ts` (e.g., `electron.d.ts`)
- Non-component TS modules: `camelCase.ts` (e.g., `themes.ts`)
- CSS files: `kebab-case.css` (e.g., `index.css`)
- Use `camelCase` for all functions: `createCatalog`, `handleTabChange`, `selectDirectory`
- React components: `PascalCase` function declarations: `function MainContent()`, `function CatalogModal()`
- Event handlers: prefix with `handle` (e.g., `handleThemeChange`, `handleSort`, `handleClose`)
- Async operations: prefix with action verb (e.g., `performSearch`, `loadCatalogs`, `loadHtmlContent`)
- IPC handlers in main process: use descriptive kebab-case channel names (e.g., `'select-directory'`, `'create-catalog'`)
- Use `camelCase` for all variables: `catalogTitle`, `searchTerm`, `currentPage`
- State variables: descriptive nouns (e.g., `selectedDirectory`, `catalogModalVisible`)
- Boolean state: prefix with `is`/`has`/`can` (e.g., `isCreating`, `isSearching`, `canSearch`, `hasHtml`)
- Constants: `UPPER_SNAKE_CASE` for object constants (e.g., `DEFAULT_WINDOW_STATE`)
- Use `PascalCase` for interfaces and types: `AppState`, `AppAction`, `Theme`, `ThemeColors`, `ElectronAPI`
- Component prop interfaces: `ComponentNameProps` (e.g., `CatalogModalProps`, `ModernTableProps`)
- No `I` prefix on interfaces
## Code Style
- No Prettier or formatting tool configured
- 2-space indentation in TypeScript/React files
- 2-space indentation in JavaScript main process files
- Single quotes for string literals in both TS and JS
- Semicolons at end of statements
- Trailing commas in multi-line arrays, objects, and function parameters
- ESLint configured via devDependencies (`@typescript-eslint/eslint-plugin`, `@typescript-eslint/parser`, `eslint-plugin-react-hooks`, `eslint-plugin-react-refresh`)
- No `.eslintrc` config file found in the project root (may rely on defaults or an unlisted config)
- Strict mode enabled (`"strict": true` in `tsconfig.json`)
- `noUnusedLocals: true` and `noUnusedParameters: true`
- Target: ES2020 with React JSX transform
- TypeScript only applies to `src/renderer/` (main process files are plain JS)
## Import Organization
- No path aliases configured. All imports use relative paths (e.g., `../../contexts/AppContext`, `../ModernTable`)
- Destructure Ant Design sub-components at module level:
- Use `require()` for all imports in main process files
- Destructure at import: `const { app, BrowserWindow, ipcMain } = require('electron');`
- Use `module.exports` for exports
## Component Patterns
- Use `function` declarations (not arrow functions) for components:
- Default export at bottom of file
- Tab components export an object with `Sidebar` and `Content` sub-components:
- Usage: `<CreateCatalogTab.Sidebar />` and `<CreateCatalogTab.Content />`
- Global state: React Context + `useReducer` pattern in `src/renderer/contexts/AppContext.tsx`
- Actions use `UPPER_SNAKE_CASE` string types (e.g., `'SET_SELECTED_DIRECTORY'`, `'SET_IS_CREATING'`)
- Local component state: `useState` hooks for UI-specific state
- Custom hook: `useAppContext()` wraps `useContext` with undefined check
- Define props interface directly above component:
- No class components anywhere in the codebase
- Hooks used throughout: `useState`, `useEffect`, `useMemo`, `useRef`, `useCallback`, `useReducer`, `useContext`
## Styling Approach
- Primary styling method is inline `style` objects on JSX elements
- CSS custom properties (variables) used extensively via `var(--property-name)`:
- Defined in `src/renderer/index.css` with `:root` defaults
- Overridden at runtime via JavaScript in `src/renderer/App.tsx` `applyTheme()` function
- Key variables: `--app-bg`, `--app-text`, `--card-bg`, `--border-color`, `--header-bg`, `--header-text`, `--sidebar-bg`, `--table-stripe`, `--table-hover`, `--modal-bg`, `--shadow-color`, `--input-bg`, `--code-bg`, `--icon-filter`, `--link-color`, `--link-hover`
- Layout variables: `--header-height: 45px`, `--tab-nav-height: 50px`
- `src/renderer/index.css` contains Ant Design overrides using `!important` to enforce theme colors
- CSS class `.path-display` for directory path display
- CSS class `.catalog-link` for clickable catalog links
- Custom scrollbar styling via `::-webkit-scrollbar` pseudo-elements
## Error Handling
- Use try/catch around `window.electronAPI` calls
- Display errors via Ant Design `message.error()`:
- IPC handlers return `{ success: boolean, error?: string }` result objects
- Never throw errors across IPC boundary; always catch and return error message:
- Silent error swallowing for file access errors (skip inaccessible files/dirs)
- `console.warn()` for non-critical failures
- `throw new Error()` with descriptive messages for critical failures
## Inter-Process Communication
- All IPC uses `ipcMain.handle` / `ipcRenderer.invoke` (async request-response)
- No `ipcMain.on` / `send` (fire-and-forget) pattern used
- Preload script (`src/preload.js`) exposes `window.electronAPI` via `contextBridge`
- TypeScript types for the API defined in `src/renderer/types/electron.d.ts`
- `CustomEvent` dispatched on `window` for cross-component communication:
## Persistence
- `storcat-theme-id` - selected theme ID
- `storcat-sidebar-position` - left or right
- `storcat-last-create-directory` - last directory selected for catalog creation
- `storcat-last-output-directory` - last output directory
- `storcat-last-catalog-directory` - last catalog directory for search/browse
- `storcat-last-search-term` - last search term
- `storcat-last-search-results` - cached search results (JSON)
- `storcat-last-browse-catalogs` - cached browse catalogs (JSON)
- `window-state.json` - window size and position
- `preferences.json` - app preferences (window persistence toggle)
## Comments
- Comments used sparingly, primarily for section headers within components (e.g., `{/* Sidebar Header with Toggle and Settings */}`)
- JSX comments for UI section delineation
- Code comments for non-obvious logic (e.g., `// macOS traffic light button spacing`)
- No JSDoc/TSDoc usage on functions
## Logging
- Main process: `console.log()` for startup/lifecycle events, `console.warn()` for recoverable errors, `console.error()` for failures
- Renderer: `console.warn()` and `console.error()` for error conditions only
- No structured logging library
<!-- GSD:conventions-end -->

<!-- GSD:architecture-start source:ARCHITECTURE.md -->
## Architecture

## Pattern Overview
- Context isolation enforced -- renderer communicates with main only through a preload bridge (`window.electronAPI`)
- All filesystem and catalog logic lives in the main process; the renderer is purely presentational
- State management via React `useReducer` + Context, supplemented by `localStorage` for persistence
- No router -- single-window app with tab-based navigation (Create, Search, Browse)
- Theming via CSS custom properties set at runtime from a theme registry
## Layers
- Purpose: Application lifecycle, native OS integration, filesystem catalog operations
- Location: `src/main.js`
- Contains: Window creation, IPC handler registration, window state persistence, preferences management
- Depends on: `src/catalog-service.js`, `src/preload.js`, Electron APIs, Node `fs`/`path`
- Used by: Renderer process (indirectly, via IPC)
- Purpose: Expose a safe, typed API surface from main process to renderer
- Location: `src/preload.js`
- Contains: `contextBridge.exposeInMainWorld('electronAPI', ...)` mapping 12 IPC channels
- Depends on: Electron `contextBridge`, `ipcRenderer`
- Used by: Every renderer component that touches filesystem or native dialogs
- Purpose: Core business logic -- directory traversal, catalog JSON/HTML generation, search
- Location: `src/catalog-service.js`
- Contains: `createCatalog()`, `searchCatalogs()`, `traverseDirectory()`, `generateTreeHTML()`, helpers
- Depends on: Node `fs.promises`, `path`
- Used by: `src/main.js` (IPC handlers for `create-catalog` and `search-catalogs`)
- Purpose: All UI rendering, user interaction, local state
- Location: `src/renderer/`
- Contains: React components, context/state, theme definitions, CSS, type declarations
- Depends on: React 18, Ant Design 5, `window.electronAPI` bridge
- Used by: Electron BrowserWindow (loaded via Vite dev server or built `dist/index.html`)
## Data Flow
- **Global UI state:** `useReducer` in `src/renderer/contexts/AppContext.tsx` -- holds selected directories, loading flags, search results, browse catalogs, active tab, sidebar state, modal state
- **Theme state:** Managed in `App.tsx` via `useState`, persisted to `localStorage` (`storcat-theme-id`). CSS custom properties applied to `document.documentElement` at runtime.
- **User preferences persistence:** Split between `localStorage` (theme, sidebar position, last-used directories, last search term) and Electron `userData` JSON files (window position/size, window persistence toggle)
- **Cross-component communication:** Custom DOM events (`themeChange`, `openCatalogModal`) used to bridge `MainContent` settings modal to `App.tsx` and tab components to `App.tsx` modal
## Key Abstractions
- Purpose: Each tab exports an object with `.Sidebar` and `.Content` sub-components
- Examples: `src/renderer/components/tabs/CreateCatalogTab.tsx`, `src/renderer/components/tabs/SearchCatalogsTab.tsx`, `src/renderer/components/tabs/BrowseCatalogsTab.tsx`
- Pattern: `MainContent` conditionally renders `<Tab.Sidebar />` in the sidebar and `<Tab.Content />` in the main area based on `state.activeTab`
- Purpose: Reusable data table with sorting, column filtering, pagination, resizable columns
- Examples: `src/renderer/components/ModernTable.tsx`
- Pattern: Accepts `columns` (with optional `render`, `sortable`, `filterable`) and `data` props. Used by SearchCatalogsTab and BrowseCatalogsTab.
- Purpose: Multi-theme support (11 built-in themes) via CSS custom properties
- Examples: `src/renderer/themes.ts`
- Pattern: Each `Theme` object defines `id`, `name`, `type` (light/dark), `colors` (16 CSS variable values), `antdAlgorithm`, and optional `antdPrimaryColor`. Applied at runtime by setting CSS variables on `:root`.
- Purpose: Hierarchical directory tree stored as JSON
- Pattern: Recursive `{type: 'file'|'directory', name: string, size: number, contents?: CatalogItem[]}`
- Compatibility: Supports both own format (single root object) and legacy bash script format (array wrapping)
## Entry Points
- Location: `src/main.js`
- Triggers: `electron .` or `npm start`
- Responsibilities: Creates `BrowserWindow`, registers all IPC handlers, manages window state/preferences
- Location: `src/renderer/main.tsx`
- Triggers: Loaded by `src/renderer/index.html` via `<script type="module">`
- Responsibilities: Mounts React app (`App.tsx`) into `#root`
- Location: `vite.config.ts` (root: `src/renderer`)
- Triggers: `npm run dev:vite` (port 5173)
- Responsibilities: HMR for renderer development
- Location: `sdcat/sdcat.sh`, `sdcat/sdcat_zsh.sh`
- Triggers: Manual shell execution
- Responsibilities: Original bash-based catalog generation (StorCat is the Electron replacement)
## Error Handling
- Main process IPC handlers wrap all logic in try/catch, returning `{ success: false, error: error.message }` on failure
- Catalog traversal silently skips inaccessible files/directories (permissions, broken symlinks) with `console.warn`
- Window state load/save fails silently with fallback defaults
- Renderer components check `result.success` before processing IPC responses
## Cross-Cutting Concerns
<!-- GSD:architecture-end -->

<!-- GSD:workflow-start source:GSD defaults -->
## GSD Workflow Enforcement

Before using Edit, Write, or other file-changing tools, start work through a GSD command so planning artifacts and execution context stay in sync.

Use these entry points:
- `/gsd:quick` for small fixes, doc updates, and ad-hoc tasks
- `/gsd:debug` for investigation and bug fixing
- `/gsd:execute-phase` for planned phase work

Do not make direct repo edits outside a GSD workflow unless the user explicitly asks to bypass it.
<!-- GSD:workflow-end -->



<!-- GSD:profile-start -->
## Developer Profile

> Profile not yet configured. Run `/gsd:profile-user` to generate your developer profile.
> This section is managed by `generate-claude-profile` -- do not edit manually.
<!-- GSD:profile-end -->
