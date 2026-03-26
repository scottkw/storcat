# Phase 4: App Layer + Lifecycle - Research

**Researched:** 2026-03-25
**Domain:** Go app.go method fixes + TypeScript wailsAPI.ts envelope corrections
**Confidence:** HIGH

---

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| API-01 | `GetCatalogHtmlPath` verifies file existence and returns via wailsAPI wrapper as `{success, htmlPath}` | Go method does path manipulation only (no `os.Stat` check); wailsAPI wrapper returns raw string or `null`, not `{success, htmlPath}` |
| API-02 | `ReadHtmlFile` returns via wailsAPI wrapper as `{success, content}` | Go method signature is correct `(string, error)`; wailsAPI wrapper returns raw string or `null`, not `{success, content}` |
| API-03 | All wailsAPI wrapper methods consistently return `{success, ...}` envelopes matching Electron's contract | 5 wrappers return raw values: `selectDirectory`, `selectSearchDirectory`, `selectOutputDirectory`, `getConfig`, `getCatalogHtmlPath`, `readHtmlFile`, `getWindowPersistence` |
| WIN-04 | Window state restores via `OnDomReady` hook (not `OnStartup`) | Already implemented in Phase 3: `app.go` has `domReady(ctx)` method and `main.go` wires `OnDomReady: app.domReady` — this requirement is ALREADY MET |
</phase_requirements>

---

## Summary

Phase 4 has two independent tracks of work.

**Track 1: Go layer fix for API-01.** `GetCatalogHtmlPath` in `app.go` currently does a string manipulation only (strips `.json`, appends `.html`) with no file existence check. The success criterion requires it to verify the `.html` file exists via `os.Stat` before returning. If the file is missing, it must return an error (not a raw path). The Go signature stays `(string, error)` — no envelope needed at the Go layer.

**Track 2: TypeScript wailsAPI wrapper fixes for API-01, API-02, API-03.** The wailsAPI shim in `frontend/src/services/wailsAPI.ts` is the adapter between Wails' Promise-based Go bindings and the `{success, ...}` envelope contract the React components expect. Currently, `getCatalogHtmlPath`, `readHtmlFile`, `selectDirectory` (three aliases), `getConfig`, and `getWindowPersistence` return raw values (`null` or the raw data) rather than envelopes. Every wrapper must be updated to match the established pattern used by `createCatalog`, `searchCatalogs`, etc.

**WIN-04 is already complete.** Phase 3 implemented `domReady` in `app.go` and wired `OnDomReady: app.domReady` in `main.go`. The requirement says "restores via `OnDomReady`" — this is verified true. No changes needed.

The phase is small and surgical: one Go method edit and multiple TypeScript wrapper edits. No new imports, no new Go methods, no binding regeneration needed (the Go method signatures do not change, only their implementations and the TS wrapper behavior changes).

**Primary recommendation:** Fix Go `GetCatalogHtmlPath` first (one method), then fix all non-conformant TypeScript wrappers in a single pass. Verify each API-03 wrapper against the conformance list below.

---

## Standard Stack

### Core

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `os` (stdlib) | stdlib | `os.Stat` for file existence check in `GetCatalogHtmlPath` | Already imported in `app.go`; no new import needed |
| `fmt` (stdlib) | stdlib | Wrap errors with context (`fmt.Errorf`) | Already imported in `app.go` pattern |

### Supporting

None. No new dependencies. No binding regeneration needed.

**Installation:** No new packages. All changes are within existing imports.

---

## Architecture Patterns

### Recommended Project Structure (no layout changes)

```
app.go
  GetCatalogHtmlPath()  ← EDIT: add os.Stat existence check before returning

frontend/src/services/wailsAPI.ts
  getCatalogHtmlPath    ← EDIT: return {success, htmlPath}
  readHtmlFile          ← EDIT: return {success, content}
  selectDirectory       ← EDIT: return {success, path} or {success: false}
  selectSearchDirectory ← EDIT: return {success, path} or {success: false}
  selectOutputDirectory ← EDIT: return {success, path} or {success: false}
  getConfig             ← EDIT: return {success, config} or {success: false}
  getWindowPersistence  ← EDIT: return {success, enabled} or {success: false}
```

### Pattern 1: Go File Existence Check (API-01)

**What:** Use `os.Stat` to verify the `.html` file exists before returning the path.
**When to use:** Any Go method that returns a derived path that may not exist on disk.

```go
// Source: app.go — replace current GetCatalogHtmlPath implementation
// Current (wrong): only string manipulation, no existence check
// Fixed:
func (a *App) GetCatalogHtmlPath(catalogPath string) (string, error) {
    var htmlPath string
    if filepath.Ext(catalogPath) == ".json" {
        htmlPath = catalogPath[:len(catalogPath)-5] + ".html"
    } else {
        htmlPath = catalogPath + ".html"
    }
    if _, err := os.Stat(htmlPath); err != nil {
        if os.IsNotExist(err) {
            return "", fmt.Errorf("HTML file not found: %s", htmlPath)
        }
        return "", err
    }
    return htmlPath, nil
}
```

### Pattern 2: wailsAPI Envelope Convention (API-01, API-02, API-03)

**What:** Every wailsAPI wrapper catches Promise rejections and returns `{success: bool, ...payload}`.
**When to use:** Every wrapper in `wailsAPI.ts` — no exceptions.

```typescript
// Source: frontend/src/services/wailsAPI.ts
// Established pattern from createCatalog, searchCatalogs, browseCatalogs:
someMethod: async (...args) => {
  try {
    const result = await GoMethod(...args);
    return { success: true, fieldName: result };
  } catch (error: any) {
    return { success: false, error: error.message || 'Unknown error' };
  }
},
```

**API-01 fix — getCatalogHtmlPath:**
```typescript
getCatalogHtmlPath: async (catalogPath: string) => {
  try {
    const htmlPath = await GetCatalogHtmlPath(catalogPath);
    return { success: true, htmlPath };
  } catch (error: any) {
    return { success: false, error: error.message || 'Unknown error' };
  }
},
```

**API-02 fix — readHtmlFile:**
```typescript
readHtmlFile: async (filePath: string) => {
  try {
    const content = await ReadHtmlFile(filePath);
    return { success: true, content };
  } catch (error: any) {
    return { success: false, error: error.message || 'Unknown error' };
  }
},
```

**API-03 fixes — remaining non-conformant wrappers:**

```typescript
// selectDirectory (and its two aliases)
selectDirectory: async () => {
  try {
    const path = await SelectDirectory();
    return { success: true, path };
  } catch (error: any) {
    return { success: false, error: error.message || 'Unknown error' };
  }
},

// getConfig
getConfig: async () => {
  try {
    const config = await GetConfig();
    return { success: true, config };
  } catch (error: any) {
    return { success: false, error: error.message || 'Unknown error' };
  }
},

// getWindowPersistence
getWindowPersistence: async () => {
  try {
    const enabled = await GetWindowPersistence();
    return { success: true, enabled };
  } catch (error: any) {
    return { success: false, enabled: true }; // default enabled on error
  }
},
```

### Pattern 3: Frontend Consumer Updates Required After Wrapper Changes

**What:** When wrappers change return shape, their callers must be updated.
**When to use:** After every wrapper shape change.

The following callers must be updated to handle the new shapes:

| Wrapper Changed | Caller Location | Current Usage | Required Update |
|-----------------|-----------------|---------------|-----------------|
| `getCatalogHtmlPath` | `CatalogModal.tsx:26-28` | `const htmlPath = await ...` then `if (!htmlPath)` | `const result = await ...; if (!result.success)` |
| `readHtmlFile` | `CatalogModal.tsx:33-35` | `const htmlContent = await ...` then `if (htmlContent)` | `const result = await ...; if (result.success)` use `result.content` |
| `selectDirectory` | `CreateCatalogTab.tsx` (all three directory select buttons) | `const directory = await ...` then `if (directory)` | `const result = await ...; if (result.success)` use `result.path` |
| `getConfig` | `App.tsx` or context init | `const config = await ...` then direct use | `const result = await ...; if (result.success)` use `result.config` |
| `getWindowPersistence` | Settings modal / `MainContent.tsx` | `const enabled = await ...` direct use | `const result = await ...; if (result.success)` use `result.enabled` |

**CRITICAL:** Verify every call site before updating wrappers. An incomplete caller update creates a silent `undefined` access bug — the wrapper returns `{success, path}` but the caller reads `.path` on `undefined` because it still uses the old raw shape.

### Anti-Patterns to Avoid

- **Returning raw values from wailsAPI wrappers:** `return await GetCatalogHtmlPath(catalogPath)` is wrong. If the Go call throws, it becomes an unhandled rejection. Always wrap in try/catch.
- **Swallowing errors with `return null`:** `return null` on catch hides error details. Return `{ success: false, error: ... }` so the component can display the actual error message.
- **Checking `if (result)` instead of `if (result.success)`:** After migrating to envelope pattern, callers must use `result.success` as the discriminant, not truthiness of the result object itself (which is always truthy as an object).
- **Changing Go method signatures for envelope compatibility:** Go methods should NOT return structs with `Success bool` fields. Wails translates Go `error` → Promise rejection, and the TypeScript shim handles envelope construction. This was established in prior research (ARCHITECTURE.md line 164).

---

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| File existence check | Custom path existence logic | `os.Stat` with `os.IsNotExist` check | Standard Go pattern; handles permission errors vs. not-found correctly |
| Error envelope in Go | Return `struct{ Success bool; Error string }` from Go | Return `(T, error)` — wailsAPI.ts constructs envelope | Wails v2 contract: Go errors become Promise rejections; shim wraps them |
| Multiple try/catch blocks | Per-wrapper boilerplate | Single consistent pattern across all wrappers | Consistency is the requirement; one standard template |

**Key insight:** The entire gap between Electron's explicit IPC envelopes and Wails' native Promise rejection is bridged by `wailsAPI.ts`. The Go layer stays idiomatic Go. Only the TypeScript shim needs envelope construction.

---

## Runtime State Inventory

Step 2.5: SKIPPED — Phase 4 is not a rename/refactor/migration phase. It is a bug fix and API conformance phase.

---

## Environment Availability

Step 2.6: SKIPPED — Phase 4 is purely Go source and TypeScript changes. No new external tools required. `go`, `wails`, and `npm`/`node` verified in prior phases.

---

## Common Pitfalls

### Pitfall 1: Forgetting the Caller Updates After Wrapper Shape Change
**What goes wrong:** `getCatalogHtmlPath` now returns `{success, htmlPath}` but `CatalogModal.tsx` still reads the return value as a raw string. `const htmlPath = result` then `if (!htmlPath)` is always truthy (it's an object), so a missing-file case never shows the error and instead passes `{success: false}` as the path to `readHtmlFile`.
**Why it happens:** The wrapper change is local to `wailsAPI.ts`; the compiler does not catch shape mismatches because the return type was `any`.
**How to avoid:** After each wrapper fix, grep for every call site of that wrapper and update the consumer before moving on.
**Warning signs:** TypeScript errors on `result.path` where `result` is typed as `string`, OR missing-file errors silently fail with no user feedback.

### Pitfall 2: `os.Stat` Error Type — Not All Errors Are "Not Found"
**What goes wrong:** Checking `if err != nil { return "", err }` catches permission errors, I/O errors, etc. but the caller interprets all errors as "file not found."
**Why it happens:** Conflating `os.IsNotExist(err)` with any error from `os.Stat`.
**How to avoid:** Use `os.IsNotExist(err)` to distinguish "file does not exist" from other I/O errors. Both should return an error to the frontend; the message should differ.
**Warning signs:** Permission-denied errors on an existing file show "HTML file not found" to the user.

### Pitfall 3: WIN-04 Is Already Done — Do Not Re-Implement
**What goes wrong:** Researcher/planner sees WIN-04 as "pending" in REQUIREMENTS.md and implements `OnDomReady` wiring again, causing duplicate calls.
**Why it happens:** REQUIREMENTS.md still shows `[ ] WIN-04` because the requirements file was not updated after Phase 3.
**How to avoid:** Verify in `main.go` — `OnDomReady: app.domReady` is present. Verify in `app.go` — `func (a *App) domReady(ctx context.Context)` is implemented. This requirement is DONE. The plan should mark it complete, not re-implement it.
**Warning signs:** Duplicate window resize on startup; second `domReady` call fighting with the first.

### Pitfall 4: selectDirectory Alias Wrappers Must All Be Updated
**What goes wrong:** `selectDirectory` is updated to return `{success, path}` but `selectSearchDirectory` and `selectOutputDirectory` (which are identical implementations calling the same `SelectDirectory()`) are left as raw-return wrappers.
**Why it happens:** Three separate wrappers calling the same underlying function; easy to miss the aliases.
**How to avoid:** Update all three together. Confirm all three appear in the same consistent shape.
**Warning signs:** Directory picker works in Create tab (uses `selectDirectory`) but fails in Browse/Search tabs (use `selectSearchDirectory`/`selectOutputDirectory`).

### Pitfall 5: CatalogModal Dark Mode Branch After htmlPath Change
**What goes wrong:** After updating `getCatalogHtmlPath` to return `{success, htmlPath}`, the second call to `readHtmlFile` passes the old variable `htmlPath` which is now `{success, htmlPath}` (an object) instead of the string path.
**Why it happens:** The variable `const htmlPath = await window.electronAPI.getCatalogHtmlPath(catalogPath)` type changes; if the caller isn't updated, the wrong value is passed downstream.
**How to avoid:** Update `CatalogModal.tsx` in the same commit as the wrapper changes. Use TypeScript strictly — add explicit types to force the compiler to catch the mismatch.
**Warning signs:** `ReadHtmlFile` receives `[object Object]` as the path; returns "file not found" error for every catalog.

---

## Code Examples

Verified patterns from official sources and direct codebase inspection:

### GetCatalogHtmlPath — Fixed Go Implementation
```go
// Source: app.go (current lines 201-208 to replace)
func (a *App) GetCatalogHtmlPath(catalogPath string) (string, error) {
    var htmlPath string
    if filepath.Ext(catalogPath) == ".json" {
        htmlPath = catalogPath[:len(catalogPath)-5] + ".html"
    } else {
        htmlPath = catalogPath + ".html"
    }
    if _, err := os.Stat(htmlPath); err != nil {
        if os.IsNotExist(err) {
            return "", fmt.Errorf("HTML file not found: %s", htmlPath)
        }
        return "", fmt.Errorf("cannot access HTML file: %w", err)
    }
    return htmlPath, nil
}
```

### CatalogModal.tsx — Fixed Consumer
```typescript
// Source: frontend/src/components/CatalogModal.tsx (lines 20-46 to replace)
const loadHtmlContent = async () => {
  if (!catalogPath) return;

  setLoading(true);
  try {
    const pathResult = await window.electronAPI.getCatalogHtmlPath(catalogPath);
    if (!pathResult.success) {
      message.error(pathResult.error || 'HTML file not found for this catalog');
      return;
    }

    const readResult = await window.electronAPI.readHtmlFile(pathResult.htmlPath);
    if (!readResult.success) {
      message.error(readResult.error || 'Failed to read HTML file');
      return;
    }

    const modifiedHtml = modifyHtmlForTheme(readResult.content);
    setHtmlContent(modifiedHtml);
  } catch (error) {
    message.error('Failed to load catalog HTML');
  } finally {
    setLoading(false);
  }
};
```

### wailsAPI.ts — Non-Conformant Wrappers (API-03 Full List)

Current state of each non-conformant wrapper and its required fix:

```typescript
// BEFORE (lines 57-65, all three directory wrappers):
selectDirectory: async () => {
  try {
    const result = await SelectDirectory();
    return result;        // ← raw string, not envelope
  } catch (error) {
    return null;          // ← loses error message
  }
},

// AFTER:
selectDirectory: async () => {
  try {
    const path = await SelectDirectory();
    return { success: true, path };
  } catch (error: any) {
    return { success: false, error: error.message || 'Unknown error' };
  }
},
// (same pattern for selectSearchDirectory and selectOutputDirectory)

// BEFORE (lines 86-91):
getConfig: async () => {
  try {
    return await GetConfig();    // ← raw Config object
  } catch (error) {
    return null;
  }
},

// AFTER:
getConfig: async () => {
  try {
    const config = await GetConfig();
    return { success: true, config };
  } catch (error: any) {
    return { success: false, error: error.message || 'Unknown error' };
  }
},

// BEFORE (lines 158-164):
getWindowPersistence: async () => {
  try {
    return await GetWindowPersistence();    // ← raw boolean
  } catch (error) {
    return true;    // ← inconsistent fallback
  }
},

// AFTER:
getWindowPersistence: async () => {
  try {
    const enabled = await GetWindowPersistence();
    return { success: true, enabled };
  } catch (error: any) {
    return { success: false, enabled: true }; // default to enabled on error
  }
},
```

---

## API-03 Conformance Checklist

Complete inventory of all wailsAPI methods and their conformance status:

| Method | Current Return | Conformant? | Action |
|--------|---------------|-------------|--------|
| `createCatalog` | `{success, ...}` | YES | No change |
| `searchCatalogs` | `{success, results}` | YES | No change |
| `browseCatalogs` | `{success, catalogs}` | YES | No change |
| `loadCatalog` | `{success, catalog}` | YES | No change |
| `selectDirectory` | raw string or `null` | NO | Fix → `{success, path}` |
| `selectSearchDirectory` | raw string or `null` | NO | Fix → `{success, path}` |
| `selectOutputDirectory` | raw string or `null` | NO | Fix → `{success, path}` |
| `getConfig` | raw Config or `null` | NO | Fix → `{success, config}` |
| `setTheme` | `{success, ...}` | YES | No change |
| `setSidebarPosition` | `{success, ...}` | YES | No change |
| `setWindowSize` | `{success, ...}` | YES | No change |
| `getCatalogHtmlPath` | raw string or `null` | NO | Fix → `{success, htmlPath}` |
| `readHtmlFile` | raw string or `null` | NO | Fix → `{success, content}` |
| `openExternal` | `{success, ...}` | YES | No change |
| `getCatalogFiles` | `{success, catalogs}` | YES | No change |
| `getWindowPersistence` | raw boolean or `true` | NO | Fix → `{success, enabled}` |
| `setWindowPersistence` | `{success, ...}` | YES | No change |

**Total to fix: 7 wrappers** (selectDirectory×3, getConfig, getCatalogHtmlPath, readHtmlFile, getWindowPersistence)

---

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Electron IPC: explicit `{success, ...}` envelopes in preload | Wails: Go `(T, error)` → Promise; wailsAPI.ts shim wraps in envelope | Initial migration | Envelope contract preserved in shim layer, not Go layer |
| `GetCatalogHtmlPath`: string path derivation only | After fix: path derivation + `os.Stat` existence check | Phase 4 | API-01 satisfied; missing-HTML files handled gracefully |
| `readHtmlFile`: returns raw string or `null` | After fix: returns `{success, content}` | Phase 4 | API-02 satisfied; components use `.content` not raw value |
| 7 wrappers return raw values | All 17 wrappers return `{success, ...}` | Phase 4 | API-03 satisfied; uniform contract |
| WIN-04: `OnDomReady` hook | Already implemented in Phase 3 | Phase 3 | No Phase 4 work needed |

---

## Open Questions

1. **selectDirectory caller sites — how many exist?**
   - What we know: Three alias wrappers exist (`selectDirectory`, `selectSearchDirectory`, `selectOutputDirectory`). `CreateCatalogTab.tsx` uses them. Possible other callers in `SearchCatalogsTab.tsx`.
   - What's unclear: Exact line references in every tab component.
   - Recommendation: Grep for `selectDirectory\|selectSearchDirectory\|selectOutputDirectory` in `frontend/src/` before editing to find all call sites. Update all callers in the same wave as the wrapper fix.

2. **getConfig caller sites — shape dependency**
   - What we know: `getConfig` is called somewhere to initialize settings (theme, sidebar position). After the fix it returns `{success, config}` instead of raw config.
   - What's unclear: Whether callers access `.config.theme` or `.theme` directly.
   - Recommendation: Grep for `getConfig` usages in `frontend/src/` before editing. Update all callers.

3. **getWindowPersistence caller — settings modal**
   - What we know: The settings toggle exists in `MainContent.tsx` or a settings modal. It reads `getWindowPersistence()` to show current state.
   - What's unclear: The exact component and how it handles the current raw boolean return.
   - Recommendation: Grep for `getWindowPersistence` in `frontend/src/` to find the caller. The caller currently may use `if (await getWindowPersistence())` — after fix it must use `const r = await getWindowPersistence(); if (r.success && r.enabled)`.

---

## Validation Architecture

### Test Framework

| Property | Value |
|----------|-------|
| Framework | None — no automated test framework in project |
| Config file | None |
| Quick run command | `wails dev` — manual test in running app |
| Full suite command | Manual: open catalog modal for a known catalog, verify HTML loads; open for missing catalog, verify error message |

### Phase Requirements → Test Map

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| API-01 | `getCatalogHtmlPath` with existing HTML → `{success: true, htmlPath: "..."}` | smoke | Manual — browse catalogs, click a catalog with HTML, verify modal opens | ❌ No automated tests |
| API-01 | `getCatalogHtmlPath` with missing HTML → `{success: false, error: "..."}` | smoke | Manual — delete an `.html` file, click the catalog, verify error message shown | ❌ No automated tests |
| API-02 | `readHtmlFile` for valid file → `{success: true, content: "..."}` | smoke | Manual — verify HTML content renders in modal iframe | ❌ No automated tests |
| API-02 | `readHtmlFile` for missing file → `{success: false, error: "..."}` | smoke | Manual — pass nonexistent path, verify error shown not crash | ❌ No automated tests |
| API-03 | All 7 fixed wrappers return `{success, ...}` | smoke | Manual — exercise all 7 wrapper paths in running app | ❌ No automated tests |
| WIN-04 | Window size restores on relaunch | smoke | Manual — resize, quit, relaunch, verify restored size | ❌ No automated tests — ALREADY PASSING from Phase 3 |

### Sampling Rate

- **Per task commit:** Run `wails dev`, exercise the affected wrapper manually, verify envelope shape in browser console
- **Per wave merge:** Full modal flow: browse → click catalog → verify HTML preview; plus window resize → close → reopen → verify size
- **Phase gate:** All 6 smoke behaviors confirmed before `/gsd:verify-work`

### Wave 0 Gaps

None — no test infrastructure exists and none is expected (TEST-01 through TEST-03 deferred per REQUIREMENTS.md). Manual verification protocol covers all requirements.

---

## Sources

### Primary (HIGH confidence)

- Direct source analysis: `/Users/ken/dev/storcat/app.go` — `GetCatalogHtmlPath` lines 201-208, `ReadHtmlFile` lines 192-198, `domReady` lines 155-166, `main.go` `OnDomReady: app.domReady` line 37 (WIN-04 already wired)
- Direct source analysis: `/Users/ken/dev/storcat/frontend/src/services/wailsAPI.ts` — complete wrapper inventory, non-conformant returns identified at lines 61-63, 70-72, 79-81, 88-90, 124-126, 132-134, 160-163
- Direct source analysis: `/Users/ken/dev/storcat/frontend/src/components/CatalogModal.tsx` — consumer of `getCatalogHtmlPath` (line 26) and `readHtmlFile` (line 33); current raw-value consumption pattern documented
- Prior project research: `.planning/research/ARCHITECTURE.md` — error handling architecture section (lines 146-199), confirmed Wails envelope contract and wailsAPI.ts shim pattern
- Prior project research: `.planning/research/ARCHITECTURE.md` — "Current Inconsistencies to Fix" table (lines 168-173), build order §7 (lines 246-250)

### Secondary (MEDIUM confidence)

- Wails v2 error contract verified in prior research (ARCHITECTURE.md line 158): Wails GitHub issues #2242 and #3301

### Tertiary (LOW confidence)

- None

---

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — no new dependencies; all changes within existing stdlib and TypeScript
- Architecture: HIGH — exact file locations, line numbers, and patterns verified via direct source inspection
- Pitfalls: HIGH — derived from direct code reading and established patterns in codebase

**Research date:** 2026-03-25
**Valid until:** 2026-05-25 (stable patterns; no external dependency changes expected)
