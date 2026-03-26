# Technical Concerns & Debt

## HIGH Severity

### 1. V8 JIT Disabled Globally
**Location:** `src/main.js:4`
```js
app.commandLine.appendSwitch('--js-flags', '--jitless --no-opt');
```
- Applied as a workaround for ARM64 memory allocation crashes
- Disables JIT compilation on **all platforms**, not just ARM64
- Significant performance penalty for x64 users who don't need this fix
- Should be conditionally applied based on `process.arch === 'arm64'`

### 2. Zero Test Coverage
- No test framework installed
- No test files exist
- No CI/CD test pipeline
- `catalog-service.js` contains pure functions ideal for unit testing
- AppContext reducer is a pure function that should be tested
- See TESTING.md for detailed analysis

## MEDIUM Severity

### 3. Dead Legacy Code (~1,300 lines)
**Location:** `src/app.js`, `src/index.html`
- Original vanilla JS UI still in the repo
- Not loaded by the React app but included in electron-builder `files` glob
- Increases bundle size unnecessarily
- Creates confusion about which code is active

### 4. Main Process in Plain JS, Renderer in TypeScript
**Location:** `src/main.js`, `src/preload.js`, `src/catalog-service.js`
- No type safety in the main process
- IPC channel names are strings with no shared type contract
- Easy to introduce mismatches between main and renderer IPC calls

### 5. Pervasive `any` Types in Renderer
**Location:** `src/renderer/contexts/AppContext.tsx`
- `searchResults: any[]`, `browseCatalogs: any[]`
- Defeats the purpose of TypeScript
- Should have proper interfaces for catalog data structures

### 6. Unrestricted File System Reads via IPC
**Location:** `src/main.js:263-346`
- `load-catalog`, `read-html-file`, `get-catalog-files` accept arbitrary file paths from renderer
- No path validation or sandboxing
- Renderer could theoretically read any file on disk
- Should validate paths are within expected directories

### 7. Theme State via CustomEvents Instead of React Context
**Location:** `src/renderer/App.tsx:42-48`
- Theme changes communicated via `window.addEventListener('themeChange', ...)`
- Breaks React's unidirectional data flow
- Should use React context or the existing AppContext reducer

### 8. electron-builder v24 with Electron v35
**Location:** `package.json:54`
- electron-builder `^24.9.1` may not fully support Electron `^35.0.0`
- Could cause build issues on newer platforms
- Latest electron-builder is v25+

### 9. Trial-and-Error index.html Resolution
**Location:** `src/main.js:118-145`
- Production mode tries 5 different paths to find `index.html`
- Indicates build output location isn't well-defined
- Should have a single, deterministic path

## LOW Severity

### 10. Duplicated Utility Functions
- `formatBytes()` likely duplicated between `src/app.js` (legacy) and renderer components
- Should be extracted to a shared utility if legacy code is kept

### 11. open-external IPC Without URL Allowlist
**Location:** `src/main.js:400-407`
- `shell.openExternal(url)` accepts any URL from renderer
- Should validate against an allowlist of expected domains

### 12. Large Files in Git
- `storcat-project.json` (680KB) and `storcat-project.html` (550KB) in repo
- These appear to be generated/exported project metadata
- Should be in `.gitignore` or stored elsewhere

### 13. No Linting Configuration Active
- ESLint dependencies installed but no `.eslintrc` or `eslint.config.*` found
- TypeScript and ESLint plugins are dev dependencies but unconfigured

### 14. Hardcoded Development Artifacts
- `demo.js` and `test-compatibility.js` are manual test scripts in the repo root
- Should be moved to a `scripts/` directory or documented

### 15. Missing Loading/Progress Indicators
- Long operations (catalog creation, search) set `isCreating`/`isSearching` state
- No cancel mechanism for long-running catalog operations
