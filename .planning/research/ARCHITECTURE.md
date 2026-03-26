# Architecture Patterns: Go/Wails v2 Backend

**Domain:** Go/Wails v2 desktop application (StorCat catalog manager)
**Researched:** 2026-03-24
**Confidence:** HIGH — based on direct source code analysis + verified Wails v2 docs

---

## Current Architecture (As-Is)

StorCat v2 maps directly onto the standard Wails v2 single-bound-struct pattern. The entire Go backend is exposed through one struct (`App`) bound to the frontend. Wails auto-generates TypeScript bindings from that struct's public methods.

```
main.go
  └── wails.Run(Bind: []*App)
        └── App (app.go)
              ├── catalogService  → internal/catalog/service.go
              ├── searchService   → internal/search/service.go
              └── configManager   → internal/config/config.go

frontend/wailsjs/go/main/App.js   (auto-generated, DO NOT EDIT)
frontend/wailsjs/go/models.ts     (auto-generated, DO NOT EDIT)
frontend/src/services/wailsAPI.ts (hand-written compatibility shim)
```

### Component Boundaries

| Component | Package | Responsibility | Communicates With |
|-----------|---------|---------------|-------------------|
| `App` | `main` | API surface — all public methods exposed to frontend via Wails | frontend (via generated bindings), all three services |
| `catalog.Service` | `internal/catalog` | Directory traversal, JSON + HTML catalog generation, file copy | `pkg/models`, `os`/`filepath` |
| `search.Service` | `internal/search` | JSON catalog reading, recursive search, browse metadata | `pkg/models`, `os`/`filepath` |
| `config.Manager` | `internal/config` | Config JSON read/write at `~/.config/storcat/config.json` | `os`/`filepath` only |
| `models` | `pkg/models` | Shared data types (`CatalogItem`, `SearchResult`, `CatalogMetadata`) | Used by catalog, search, App |
| Wails runtime | external | Window management, OS dialogs, browser URL open | Called from `App` via `runtime.*` functions |
| Generated bindings | `frontend/wailsjs/` | TypeScript facades for all `App` methods | Called by `wailsAPI.ts` |
| `wailsAPI.ts` | `frontend/src/services/` | Compatibility shim — wraps generated bindings to match `window.electronAPI` interface | Consumed by all React components |

### Isolation Rules

- `internal/*` packages must NOT import each other. They communicate only through `App` (the coordinator).
- `internal/*` packages must NOT import the Wails runtime. Only `App` uses `runtime.*`.
- `pkg/models` is the only cross-cutting package; it has no logic — data shapes only.
- The generated `wailsjs/` directory is owned by Wails tooling. Never hand-edit it.
- `wailsAPI.ts` is the single integration point on the frontend side. Components import from this file, never from `wailsjs/` directly.

---

## Data Flow

### Catalog Creation

```
React CreateCatalogTab
  → wailsAPI.createCatalog(title, dir, outputName, copyDir)
    → wailsjs/go/main/App.CreateCatalog(...)   [Wails IPC over WebSocket]
      → app.go: App.CreateCatalog(...)
        → catalog.Service.CreateCatalog(...)
          → traverseDirectory() → *models.CatalogItem tree
          → writeJSONFile()     → <dir>/<name>.json
          → writeHTMLFile()     → <dir>/<name>.html
          → copyFile() x2       → optional copy destination
        return: (error)
      → App wraps: return error (Go)
    → Wails runtime: error → Promise.reject(message)
  → wailsAPI catches, returns { success: false, error: string }
→ React component receives { success, error? }
```

**Current gap:** `CreateCatalog` returns `error` only. The Electron version returned `{ fileCount, totalSize, paths }`. The wailsAPI wrapper artificially injects `{ success: true }` on success — losing metadata the frontend may display.

### Catalog Search

```
React SearchCatalogsTab
  → wailsAPI.searchCatalogs(term, dir)
    → App.SearchCatalogs(term, dir) [IPC]
      → search.Service.SearchCatalogs(term, absDir)
        → os.ReadDir() for .json files
        → searchInCatalogFile() per file
          → json.Unmarshal (array or object format)
          → searchInCatalog() recursive
        return: []*models.SearchResult
      return: ([]*models.SearchResult, error)
    → Wails: slice → JSON → TypeScript array
  → wailsAPI returns { success: true, results: SearchResult[] }
→ AppContext stores results, ModernTable renders
```

### Catalog Browse

```
React BrowseCatalogsTab
  → wailsAPI.browseCatalogs(dir)  OR  wailsAPI.getCatalogFiles(dir)
    → App.BrowseCatalogs(dir) [IPC]
      → search.Service.BrowseCatalogs(dir)
        → os.ReadDir() for .json files
        → HTML title extraction per catalog
        → entry.Info() for timestamps
        return: []*models.CatalogMetadata
    → Wails: slice → JSON → TypeScript array
  → wailsAPI returns { success: true, catalogs: CatalogMetadata[] }
→ ModernTable renders
```

**Current gap:** `CatalogMetadata` is missing `size` field. `Created` is using `ModTime` (a misnomer — birthtime is not available cross-platform in Go's `os` package without platform-specific calls).

### HTML Modal Preview

```
React component dispatches CustomEvent('openCatalogModal')
  → App.tsx listens, opens CatalogModal
    → wailsAPI.getCatalogHtmlPath(jsonPath)
      → App.GetCatalogHtmlPath(path) [IPC]
        → string path manipulation only (no file existence check)
        return: (string, error)
    → wailsAPI.readHtmlFile(htmlPath)
      → App.ReadHtmlFile(path) [IPC]
        → os.ReadFile()
        return: (string, error)
  → HTML injected into <iframe srcDoc>
```

**Current gap:** `GetCatalogHtmlPath` does not verify the .html file exists before returning. `ReadHtmlFile` returns raw `string` not `{ success, content }` envelope. The wailsAPI wrapper returns `null` on error, which the frontend must handle.

### Window State Persistence

```
App.startup(ctx) → saves ctx (cannot use runtime.* here safely)
  ↓
App.domReady(ctx) [NOT CURRENTLY WIRED]
  → runtime.WindowSetSize(ctx, w, h)
  → runtime.WindowSetPosition(ctx, x, y)

User resizes window
  → frontend calls wailsAPI.setWindowSize(w, h)
    → App.SetWindowSize(w, h)
      → config.Manager.SetWindowSize(w, h)
        → writes ~/.config/storcat/config.json
```

**Current gap:** `OnDomReady` is not registered in `main.go`. The Wails runtime requires `OnDomReady` (not `OnStartup`) for reliable window positioning because the window is initialised on a separate thread. Config stores `width`/`height` but not `x`/`y` position. Window position (`WindowGetPosition`/`WindowSetPosition`) is available in the Wails runtime but not yet used.

---

## Error Handling Architecture

### Wails v2 Error Contract

When a Go method has signature `func (a *App) Foo() (T, error)`:
- On success: Wails serialises `T` to JSON, resolves the Promise with the value
- On error: Wails rejects the Promise with the error message string

When a Go method has signature `func (a *App) Foo() error`:
- On success: Promise resolves with `undefined`
- On error: Promise rejects with the error message string

**Confidence: HIGH** — Verified against Wails GitHub issues #2242 and #3301.

### The Envelope Pattern vs. Native Errors

The Electron version used explicit `{ success: bool, error?: string }` envelopes at the IPC boundary. Wails provides this natively through Promise rejection. The `wailsAPI.ts` shim reconstructs the envelope pattern in the compatibility layer by wrapping every call in try/catch.

**Consequence:** The Go layer does not need to return envelopes. Go methods should return `(T, error)` or just `error`. The shim translates `throw` → `{ success: false, error }`.

### Current Inconsistencies to Fix

| Method | Current Return | Required Return | Gap |
|--------|---------------|-----------------|-----|
| `CreateCatalog` | `error` | should be `(CatalogResult, error)` | No metadata returned |
| `GetCatalogHtmlPath` | `(string, error)` | `(string, error)` with existence check | Missing file check |
| `ReadHtmlFile` | `(string, error)` | `(string, error)` | wailsAPI.ts wrapper returns `null` not `{ success, content }` |
| `LoadCatalog` | missing | `(*models.CatalogItem, error)` | Method does not exist |

### Recommended Error Handling Pattern for App methods

```go
// Go: return typed value + error — let Wails handle the envelope
func (a *App) CreateCatalog(...) (*models.CatalogResult, error) {
    result, err := a.catalogService.CreateCatalog(...)
    if err != nil {
        return nil, fmt.Errorf("catalog creation failed: %w", err)  // descriptive
    }
    return result, nil
}
```

```typescript
// TypeScript: wailsAPI.ts shim handles the try/catch
createCatalog: async (...) => {
  try {
    const result = await CreateCatalog(...);
    return { success: true, ...result };
  } catch (error: any) {
    return { success: false, error: error.message || 'Unknown error' };
  }
}
```

---

## Suggested Build Order

The active fixes in PROJECT.md have dependencies. This is the correct build sequence:

### 1. Data Model Layer (no dependencies)
Fix `pkg/models/catalog.go` first because every other component depends on it.
- Add `Size int64` to `CatalogMetadata`
- Decide and document the `Created` vs `Modified` field semantics
- Add `CatalogResult` struct (fileCount, totalSize, paths) for `CreateCatalog` return

### 2. Catalog Service Fixes (depends on: models)
- Fix `traverseDirectory` to use `os.Stat` (follows symlinks) instead of `os.Lstat`
- Fix `writeJSONFile` to write bare object `{...}` not array `[{...}]`
- Fix root node name rendering to match Electron's tree format
- Fix empty directory `Contents` to use `[]` not `nil` (omitempty causes field absence)
- Return `CatalogResult` with file count, total size, output paths

### 3. Search Service Fixes (depends on: models)
- Add `size` field population to `BrowseCatalogs` (read JSON to get total size)
- Fix `Modified` to use proper RFC3339 format consistently
- Clarify `Created` semantics (use ModTime, document that birthtime is unavailable cross-platform)

### 4. Config Manager Additions (independent of other services)
- Add `WindowX`, `WindowY` int fields to `Config` struct
- Add `SetWindowPosition(x, y int) error` method
- Add `WindowPersistenceEnabled bool` field
- Add `SetWindowPersistence(enabled bool) error` method

### 5. App Layer — API Surface Fixes (depends on: all services above)
- Add `LoadCatalog(filePath string) (*models.CatalogItem, error)` method
- Fix `CreateCatalog` signature: return `(*models.CatalogResult, error)` instead of `error`
- Fix `GetCatalogHtmlPath`: add `os.Stat` check, return `(string, error)` with real error if missing
- Add `GetWindowPosition() (int, int, error)` bound method for frontend to read position
- Add `SetWindowPosition(x, y int) error` bound method
- Add `GetWindowPersistence() bool` and `SetWindowPersistence(enabled bool) error`
- Fix version source: read from `config/version.json` or embed at build time

### 6. Wails Lifecycle Wiring (depends on: config additions)
- Register `OnDomReady: app.domReady` in `main.go`
- Implement `domReady(ctx context.Context)` in `app.go`:
  - Load saved window size + position from config
  - Call `runtime.WindowSetSize` + `runtime.WindowSetPosition`
- Wire `OnBeforeClose` or frontend `beforeunload` to save final window state

### 7. Frontend Shim Updates (depends on: App layer fixes)
- Update `wailsAPI.ts` to forward `CatalogResult` fields in `createCatalog` response
- Fix `getCatalogHtmlPath` wrapper to return `{ success, htmlPath }` envelope
- Fix `readHtmlFile` wrapper to return `{ success, content }` envelope
- Replace `getWindowPersistence`/`setWindowPersistence` stubs with real calls

### 8. macOS-Specific Fixes (independent, platform CSS only)
- Add `WebkitAppRegion: 'drag'` to header element in React
- No Go changes needed for this

### 9. Main Entry Point (depends on: config additions)
- Read saved `WindowWidth`/`WindowHeight` from config at startup
- Pass to `options.App{Width: ..., Height: ...}` so window opens at correct size immediately
- This prevents the double-resize jump that occurs when `OnDomReady` sets the size

---

## Patterns to Follow

### Pattern: Internal Services Are Stateless

`catalog.Service` and `search.Service` are empty structs — no shared state. This is correct for this domain. Each call is independent (read a directory, write files). Keep them stateless.

**Why:** Stateless services are trivially safe across goroutines and easy to test in isolation.

### Pattern: App as Thin Coordinator

`App` should be thin: resolve paths, call service methods, handle errors. Business logic belongs in `internal/*` services, not in `App` methods.

**Why:** Wails auto-generates bindings from `App`'s public methods. Keeping `App` thin keeps the API surface clean and makes it easy to see what the frontend can call.

### Pattern: Wails Runtime Only in App

`runtime.OpenDirectoryDialog`, `runtime.BrowserOpenURL`, `runtime.WindowSetSize`, etc. must only be called from `App` methods or lifecycle hooks (`startup`, `domReady`). Services must never import the Wails runtime.

**Why:** Services are reusable library code. Importing Wails runtime ties them to the desktop context, making them untestable in isolation and coupling business logic to the GUI framework.

### Pattern: Error Wrapping at Service Boundaries

Services should wrap errors with context using `fmt.Errorf("operation: %w", err)`. `App` methods propagate these wrapped errors directly to Wails (which sends the message to the frontend). Do not double-wrap.

```go
// In service — add context
return fmt.Errorf("traverse directory %s: %w", dirPath, err)

// In App — propagate as-is
result, err := a.catalogService.CreateCatalog(...)
if err != nil {
    return nil, err  // NOT: fmt.Errorf("create catalog: %w", err) — redundant
}
```

### Pattern: Omitempty Only on Optional Collections

`CatalogItem.Contents` uses `omitempty` which makes the field absent (not `[]`) when nil. Electron always emits `"contents": []` for directories. Use explicit empty slice initialization for directories.

```go
// Correct: initialize to empty slice, not nil
Contents: make([]*CatalogItem, 0),
// Then append to it
```

---

## Anti-Patterns to Avoid

### Anti-Pattern: Envelope Returns from Go

Returning `struct{ Success bool; Error string }` from Go methods is unnecessary in Wails v2. Go's native `error` return type maps directly to Promise rejection. The shim in `wailsAPI.ts` handles envelope construction for frontend compatibility.

**Why bad:** Double-wrapping errors loses the ability to use Go's standard error handling idioms and creates redundant struct types.

### Anti-Pattern: Runtime Calls in OnStartup

```go
// WRONG — runtime may not work here
func (a *App) startup(ctx context.Context) {
    a.ctx = ctx
    runtime.WindowSetSize(ctx, 1200, 800)  // Wails window thread not ready
}

// CORRECT — use OnDomReady
func (a *App) domReady(ctx context.Context) {
    runtime.WindowSetSize(ctx, a.configManager.Get().WindowWidth, a.configManager.Get().WindowHeight)
}
```

**Confidence: HIGH** — Documented in Wails official docs: "there is no guarantee the runtime will work in OnStartup as the window is initialising in a different thread."

### Anti-Pattern: Editing Generated Bindings

`frontend/wailsjs/go/main/App.js` and `App.d.ts` are regenerated by `wails generate module` or `wails dev`. Any edits are overwritten. Always update Go source and let Wails regenerate.

### Anti-Pattern: Config Manager Silent Failures

Current `NewApp()` has:
```go
configManager, err := config.NewManager()
if err != nil {
    configManager = &config.Manager{}  // zero-value Manager — no path, no config loaded
}
```
A zero-value `Manager` has an empty `configPath` and nil `config`. Calls to `configManager.SetTheme()` will attempt `json.MarshalIndent(nil, ...)` and `os.WriteFile("", ...)`. Both will silently corrupt or fail. Either propagate the error (crash on startup) or use `DefaultConfig()` and a valid-but-non-persisting manager.

---

## Scalability Considerations (Desktop Context)

| Concern | Current State | Threshold | Action |
|---------|--------------|-----------|--------|
| Catalog size | Recursive in-memory traversal | ~100K files before noticeable lag | Add goroutine worker pool in `traverseDirectory` |
| Search speed | Linear scan of all JSON files | ~50 catalogs before noticeable wait | Add index or async streaming results via Wails events |
| Config I/O | Sync file write per setting change | Not a concern at desktop scale | No action needed |
| Window state | Saved on each resize event | Debounce if called on every pixel | Debounce in frontend before calling `setWindowSize` |

---

## Sources

- Wails v2 official docs — how bindings work: https://wails.io/docs/howdoesitwork/
- Wails v2 runtime window API: https://wails.io/docs/reference/runtime/window/
- Wails v2 application lifecycle (OnStartup/OnDomReady): https://wails.io/docs/guides/application-development/
- Wails GitHub issue #2242 — error return type → Promise rejection: https://github.com/wailsapp/wails/issues/2242
- Wails GitHub issue #3301 — error handling discussion: https://github.com/wailsapp/wails/issues/3301
- Wails GitHub issue #1702 — window position restoration challenge: https://github.com/wailsapp/wails/issues/1702
- Direct source analysis: `/Users/ken/dev/storcat/app.go`, `main.go`, `internal/`, `pkg/`, `frontend/src/services/wailsAPI.ts`
- Electron architecture comparison: `/Users/ken/dev/storcat/.planning/codebase/ARCHITECTURE.md`

---

*Architecture research: 2026-03-24*
