# Coding Conventions

**Analysis Date:** 2026-03-24

## Naming Patterns

**Files:**
- React components: `PascalCase.tsx` (e.g., `MainContent.tsx`, `CatalogModal.tsx`, `ModernTable.tsx`)
- Tab components: `PascalCase.tsx` inside `tabs/` directory (e.g., `CreateCatalogTab.tsx`, `SearchCatalogsTab.tsx`)
- Main process files: `kebab-case.js` (e.g., `catalog-service.js`, `main.js`, `preload.js`)
- Type definition files: `kebab-case.d.ts` (e.g., `electron.d.ts`)
- Non-component TS modules: `camelCase.ts` (e.g., `themes.ts`)
- CSS files: `kebab-case.css` (e.g., `index.css`)

**Functions:**
- Use `camelCase` for all functions: `createCatalog`, `handleTabChange`, `selectDirectory`
- React components: `PascalCase` function declarations: `function MainContent()`, `function CatalogModal()`
- Event handlers: prefix with `handle` (e.g., `handleThemeChange`, `handleSort`, `handleClose`)
- Async operations: prefix with action verb (e.g., `performSearch`, `loadCatalogs`, `loadHtmlContent`)
- IPC handlers in main process: use descriptive kebab-case channel names (e.g., `'select-directory'`, `'create-catalog'`)

**Variables:**
- Use `camelCase` for all variables: `catalogTitle`, `searchTerm`, `currentPage`
- State variables: descriptive nouns (e.g., `selectedDirectory`, `catalogModalVisible`)
- Boolean state: prefix with `is`/`has`/`can` (e.g., `isCreating`, `isSearching`, `canSearch`, `hasHtml`)
- Constants: `UPPER_SNAKE_CASE` for object constants (e.g., `DEFAULT_WINDOW_STATE`)

**Types/Interfaces:**
- Use `PascalCase` for interfaces and types: `AppState`, `AppAction`, `Theme`, `ThemeColors`, `ElectronAPI`
- Component prop interfaces: `ComponentNameProps` (e.g., `CatalogModalProps`, `ModernTableProps`)
- No `I` prefix on interfaces

## Code Style

**Formatting:**
- No Prettier or formatting tool configured
- 2-space indentation in TypeScript/React files
- 2-space indentation in JavaScript main process files
- Single quotes for string literals in both TS and JS
- Semicolons at end of statements
- Trailing commas in multi-line arrays, objects, and function parameters

**Linting:**
- ESLint configured via devDependencies (`@typescript-eslint/eslint-plugin`, `@typescript-eslint/parser`, `eslint-plugin-react-hooks`, `eslint-plugin-react-refresh`)
- No `.eslintrc` config file found in the project root (may rely on defaults or an unlisted config)

**TypeScript Configuration:**
- Strict mode enabled (`"strict": true` in `tsconfig.json`)
- `noUnusedLocals: true` and `noUnusedParameters: true`
- Target: ES2020 with React JSX transform
- TypeScript only applies to `src/renderer/` (main process files are plain JS)

## Import Organization

**Order (observed pattern):**
1. React imports: `import React, { useState, useEffect } from 'react';`
2. Third-party UI library imports: `import { Layout, Tabs, Button } from 'antd';`
3. Third-party icon imports: `import { FolderOutlined, SearchOutlined } from '@ant-design/icons';`
4. Local asset imports: `import storcatIcon from '../../storcat-icon.svg';`
5. Local module imports: `import { useAppContext } from '../../contexts/AppContext';`
6. Local component imports: `import ModernTable from '../ModernTable';`

**Path Aliases:**
- No path aliases configured. All imports use relative paths (e.g., `../../contexts/AppContext`, `../ModernTable`)

**Ant Design Destructuring:**
- Destructure Ant Design sub-components at module level:
  ```typescript
  const { Sider, Content } = Layout;
  const { Title, Text } = Typography;
  ```

**Main Process (CommonJS):**
- Use `require()` for all imports in main process files
- Destructure at import: `const { app, BrowserWindow, ipcMain } = require('electron');`
- Use `module.exports` for exports

## Component Patterns

**Component Declaration:**
- Use `function` declarations (not arrow functions) for components:
  ```typescript
  function MainContent() { ... }
  export default MainContent;
  ```
- Default export at bottom of file

**Tab Component Pattern (compound component):**
- Tab components export an object with `Sidebar` and `Content` sub-components:
  ```typescript
  function CreateCatalogSidebar() { ... }
  function CreateCatalogContent() { ... }
  const CreateCatalogTab = {
    Sidebar: CreateCatalogSidebar,
    Content: CreateCatalogContent,
  };
  export default CreateCatalogTab;
  ```
- Usage: `<CreateCatalogTab.Sidebar />` and `<CreateCatalogTab.Content />`

**State Management:**
- Global state: React Context + `useReducer` pattern in `src/renderer/contexts/AppContext.tsx`
- Actions use `UPPER_SNAKE_CASE` string types (e.g., `'SET_SELECTED_DIRECTORY'`, `'SET_IS_CREATING'`)
- Local component state: `useState` hooks for UI-specific state
- Custom hook: `useAppContext()` wraps `useContext` with undefined check

**Props Pattern:**
- Define props interface directly above component:
  ```typescript
  interface CatalogModalProps {
    visible: boolean;
    catalogPath: string | null;
    onClose: () => void;
  }
  function CatalogModal({ visible, catalogPath, onClose }: CatalogModalProps) { ... }
  ```

**Functional Components Only:**
- No class components anywhere in the codebase
- Hooks used throughout: `useState`, `useEffect`, `useMemo`, `useRef`, `useCallback`, `useReducer`, `useContext`

## Styling Approach

**Inline Styles:**
- Primary styling method is inline `style` objects on JSX elements
- CSS custom properties (variables) used extensively via `var(--property-name)`:
  ```typescript
  style={{ background: 'var(--card-bg)', color: 'var(--app-text)' }}
  ```

**CSS Custom Properties (theming):**
- Defined in `src/renderer/index.css` with `:root` defaults
- Overridden at runtime via JavaScript in `src/renderer/App.tsx` `applyTheme()` function
- Key variables: `--app-bg`, `--app-text`, `--card-bg`, `--border-color`, `--header-bg`, `--header-text`, `--sidebar-bg`, `--table-stripe`, `--table-hover`, `--modal-bg`, `--shadow-color`, `--input-bg`, `--code-bg`, `--icon-filter`, `--link-color`, `--link-hover`
- Layout variables: `--header-height: 45px`, `--tab-nav-height: 50px`

**Global CSS:**
- `src/renderer/index.css` contains Ant Design overrides using `!important` to enforce theme colors
- CSS class `.path-display` for directory path display
- CSS class `.catalog-link` for clickable catalog links
- Custom scrollbar styling via `::-webkit-scrollbar` pseudo-elements

**No CSS Modules or CSS-in-JS libraries.**

## Error Handling

**Renderer (React) Pattern:**
- Use try/catch around `window.electronAPI` calls
- Display errors via Ant Design `message.error()`:
  ```typescript
  try {
    const result = await window.electronAPI.createCatalog(options);
    if (result.success) {
      message.success('Catalog created successfully!');
    } else {
      message.error(`Failed: ${result.error}`);
    }
  } catch (error) {
    message.error('Failed to create catalog');
  }
  ```

**Main Process Pattern:**
- IPC handlers return `{ success: boolean, error?: string }` result objects
- Never throw errors across IPC boundary; always catch and return error message:
  ```javascript
  ipcMain.handle('create-catalog', async (event, options) => {
    try {
      const result = await createCatalog(options);
      return { success: true, ...result };
    } catch (error) {
      return { success: false, error: error.message };
    }
  });
  ```

**catalog-service.js Pattern:**
- Silent error swallowing for file access errors (skip inaccessible files/dirs)
- `console.warn()` for non-critical failures
- `throw new Error()` with descriptive messages for critical failures

## Inter-Process Communication

**Pattern:**
- All IPC uses `ipcMain.handle` / `ipcRenderer.invoke` (async request-response)
- No `ipcMain.on` / `send` (fire-and-forget) pattern used
- Preload script (`src/preload.js`) exposes `window.electronAPI` via `contextBridge`
- TypeScript types for the API defined in `src/renderer/types/electron.d.ts`

**Custom Events (renderer-internal):**
- `CustomEvent` dispatched on `window` for cross-component communication:
  - `'themeChange'` event for theme updates
  - `'openCatalogModal'` event for opening catalog previews

## Persistence

**localStorage keys (prefixed with `storcat-`):**
- `storcat-theme-id` - selected theme ID
- `storcat-sidebar-position` - left or right
- `storcat-last-create-directory` - last directory selected for catalog creation
- `storcat-last-output-directory` - last output directory
- `storcat-last-catalog-directory` - last catalog directory for search/browse
- `storcat-last-search-term` - last search term
- `storcat-last-search-results` - cached search results (JSON)
- `storcat-last-browse-catalogs` - cached browse catalogs (JSON)

**Electron userData files:**
- `window-state.json` - window size and position
- `preferences.json` - app preferences (window persistence toggle)

## Comments

**When to Comment:**
- Comments used sparingly, primarily for section headers within components (e.g., `{/* Sidebar Header with Toggle and Settings */}`)
- JSX comments for UI section delineation
- Code comments for non-obvious logic (e.g., `// macOS traffic light button spacing`)
- No JSDoc/TSDoc usage on functions

## Logging

**Framework:** `console` (native)

**Patterns:**
- Main process: `console.log()` for startup/lifecycle events, `console.warn()` for recoverable errors, `console.error()` for failures
- Renderer: `console.warn()` and `console.error()` for error conditions only
- No structured logging library

---

*Convention analysis: 2026-03-24*
