# Phase 5: Frontend Shim - Research

**Researched:** 2026-03-25
**Domain:** TypeScript/Wails bindings shim layer
**Confidence:** HIGH

## Summary

Phase 5 is a narrow, single-file patch to `frontend/src/services/wailsAPI.ts`. The Wails binding files (`wailsjs/go/main/App.d.ts`, `App.js`, `wailsjs/go/models.ts`) are already current — running `wails generate module` against the Phase 4 Go code produced no diff. The bindings already reflect `CreateCatalog` returning `Promise<models.CreateCatalogResult>` with `jsonPath`, `htmlPath`, `fileCount`, `totalSize`, `copyJsonPath`, and `copyHtmlPath` fields.

The sole code gap is `wailsAPI.createCatalog`: it calls `await CreateCatalog(...)` but discards the returned `CreateCatalogResult`, returning only `{ success: true }`. The Phase 5 fix is to capture and spread those fields into the success envelope. The `getCatalogHtmlPath` and `readHtmlFile` wrappers are already correct — the Go methods return `(string, error)`, Wails converts the `error` return to a Promise rejection, and the try/catch wrappers already convert that to `{success: false, error}` and the happy path to `{success: true as const, htmlPath/content}`.

TypeScript compiles clean right now (`npx tsc --noEmit` exits 0). Phase 5 must not introduce any new type errors.

**Primary recommendation:** Single task — update `wailsAPI.createCatalog` to capture the `CreateCatalogResult` and spread all fields (`jsonPath`, `htmlPath`, `fileCount`, `totalSize`, `copyJsonPath`, `copyHtmlPath`) into the `{ success: true }` envelope.

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| API-01 | `GetCatalogHtmlPath` verifies file existence and returns via wailsAPI wrapper as `{success, htmlPath}` | Wrapper already correct: Go throws on missing file, catch block returns `{success: false}`, happy path returns `{success: true as const, htmlPath}` |
| API-02 | `ReadHtmlFile` returns via wailsAPI wrapper as `{success, content}` | Wrapper already correct: same throw/catch pattern, happy path returns `{success: true as const, content}` |
| API-03 | All wailsAPI wrapper methods consistently return `{success, ...}` envelopes matching Electron's contract | `createCatalog` currently returns `{success: true}` without the result fields — this is the only remaining gap |
</phase_requirements>

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| Wails v2 | v2.10.2 | Go-to-TypeScript binding generator | Project runtime — generates `wailsjs/` files |
| TypeScript | ^5.2.2 | Static typing for the shim layer | Project language — strict mode, noUnusedLocals |
| React | ^18.2.0 | Consumer components | Project framework |

### No New Dependencies
Phase 5 requires no new packages. All tools are already installed.

**Binding regeneration command:**
```bash
wails generate module
```
Run from the project root (`/Users/ken/dev/storcat`). This updates `frontend/wailsjs/go/main/App.d.ts`, `App.js`, and `frontend/wailsjs/go/models.ts`.

**Current binding status:** Already current. Confirmed by running `wails generate module` — zero diff against HEAD.

## Architecture Patterns

### Wails Binding Contract

Wails v2 generates TypeScript bindings from Go bound methods:
- A Go method `func (a *App) Foo(arg string) (string, error)` becomes `export function Foo(arg1: string): Promise<string>`
- When the Go method returns an `error`, Wails converts it to a Promise rejection (the JS binding throws)
- The `wailsAPI.ts` shim wraps each binding in try/catch to produce `{success: boolean, ...fields}` envelopes

### Current createCatalog Gap

```typescript
// CURRENT (broken — discards result)
createCatalog: async (title, directoryPath, outputName, copyToDirectory) => {
  try {
    await CreateCatalog(title, directoryPath, outputName, copyToDirectory);
    return { success: true };   // <-- result discarded
  } catch (error: any) {
    return { success: false, error: error.message || 'Unknown error' };
  }
},
```

```typescript
// FIXED (captures and spreads result fields)
createCatalog: async (title, directoryPath, outputName, copyToDirectory) => {
  try {
    const result = await CreateCatalog(title, directoryPath, outputName, copyToDirectory);
    return {
      success: true as const,
      jsonPath: result.jsonPath,
      htmlPath: result.htmlPath,
      fileCount: result.fileCount,
      totalSize: result.totalSize,
      copyJsonPath: result.copyJsonPath,
      copyHtmlPath: result.copyHtmlPath,
    };
  } catch (error: any) {
    return { success: false as const, error: error.message || 'Unknown error' };
  }
},
```

**Note on `as const`:** The Phase 4 pattern (commits `437313a1`) added `as const` discriminated union narrowing to `selectDirectory` and `getConfig`. Apply the same pattern to `createCatalog` for consistency — `success: true as const` and `success: false as const` enable TypeScript to narrow the type at call sites.

### getCatalogHtmlPath — Already Correct

```typescript
// Source: frontend/src/services/wailsAPI.ts (current)
getCatalogHtmlPath: async (catalogPath: string) => {
  try {
    const htmlPath = await GetCatalogHtmlPath(catalogPath);
    return { success: true as const, htmlPath };
  } catch (error: any) {
    return { success: false as const, error: error.message || 'Unknown error' };
  }
},
```

The Go method throws `fmt.Errorf("HTML file not found: %s", htmlPath)` when the file does not exist (Phase 4 decision). Wails converts the `error` return to a Promise rejection. The catch returns `{success: false}`. No change needed.

### readHtmlFile — Already Correct

```typescript
// Source: frontend/src/services/wailsAPI.ts (current)
readHtmlFile: async (filePath: string) => {
  try {
    const content = await ReadHtmlFile(filePath);
    return { success: true as const, content };
  } catch (error: any) {
    return { success: false as const, error: error.message || 'Unknown error' };
  }
},
```

`os.ReadFile` errors on missing file; Go returns `"", err`; Wails throws; catch returns `{success: false}`. No change needed.

### Anti-Patterns to Avoid
- **Discarding Wails binding return values:** Always capture the resolved value with `const result = await GoMethod(...)` before building the envelope.
- **Missing `as const` on discriminated unions:** `success: true` without `as const` gives type `boolean`, breaking TypeScript's ability to narrow the union. Maintain the pattern from Phase 4.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Binding types | Manual type declarations | `wails generate module` output | Auto-generated from Go structs; manual drift causes type mismatches |
| Error serialization | Custom error shape | Wails throw→catch pattern | Wails Go runtime serializes errors to JS exceptions automatically |

## Common Pitfalls

### Pitfall 1: Regenerating Bindings When Not Needed
**What goes wrong:** Running `wails generate module` during a hot dev session can overwrite bindings mid-edit.
**Why it happens:** Phase 5 success criterion 4 says "regenerate before TypeScript edits." But bindings are already current.
**How to avoid:** Run `wails generate module`, confirm zero diff, then proceed. Document in the plan that regeneration produced no changes (this eliminates ambiguity for verifiers).
**Warning signs:** `git diff frontend/wailsjs/` shows any output after the generate step.

### Pitfall 2: Forgetting Optional Fields in Envelope
**What goes wrong:** `copyJsonPath` and `copyHtmlPath` are optional in `CreateCatalogResult` (defined as `copyJsonPath?: string`). Spreading them without handling `undefined` is fine in TypeScript (the fields will be `undefined` on the returned object), but a consumer checking `result.copyJsonPath` when the field is absent gets `undefined`, not a missing key error.
**How to avoid:** Keep the spread as-is. The `?` fields propagate correctly.

### Pitfall 3: TypeScript `noUnusedLocals` Rejection
**What goes wrong:** Current tsconfig has `"noUnusedLocals": true`. If a variable is declared but not used, `tsc` fails.
**Why it happens:** The `result` variable in the fixed `createCatalog` must be used (all fields read).
**How to avoid:** Spread or destructure all fields into the return object. Run `npx tsc --noEmit` to verify clean compilation after the edit.

### Pitfall 4: Consumer Does Not Use Result Fields
**What goes wrong:** `CreateCatalogTab.tsx` only checks `result.success` and shows a generic message. It does not display `fileCount` or `totalSize`. This is expected and acceptable for Phase 5 — the success criteria require the shim to *forward* the fields, not the consumer to *display* them.
**How to avoid:** Do not modify `CreateCatalogTab.tsx` in Phase 5. Leave the display logic for a future enhancement (ENH-03). The phase is complete when the shim wrapper returns the correct shape.

## Code Examples

### CreateCatalogResult Type (from models.ts)
```typescript
// Source: frontend/wailsjs/go/models.ts (current, auto-generated)
export class CreateCatalogResult {
    jsonPath: string;
    htmlPath: string;
    fileCount: number;
    totalSize: number;
    copyJsonPath?: string;
    copyHtmlPath?: string;
    // ...constructor omitted
}
```

### App.d.ts Binding (current — no change needed)
```typescript
// Source: frontend/wailsjs/go/main/App.d.ts
export function CreateCatalog(arg1:string,arg2:string,arg3:string,arg4:string):Promise<models.CreateCatalogResult>;
```

### Wails Error-to-Rejection Pattern
```typescript
// Go: func (a *App) GetCatalogHtmlPath(catalogPath string) (string, error)
// Go returns ("", error) → Wails throws in JS
// Go returns ("path.html", nil) → Wails resolves with "path.html"
// Shim:
const htmlPath = await GetCatalogHtmlPath(catalogPath); // throws on error
return { success: true as const, htmlPath };            // resolves here
// catch (error) → return { success: false as const, error: error.message }
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| `await CreateCatalog(...)` (discard result) | `const result = await CreateCatalog(...); return { success: true, ...result }` | Phase 5 | Exposes fileCount, totalSize, paths to consumers |
| `success: true` (boolean) | `success: true as const` (literal) | Phase 4 commit 437313a1 | Enables TypeScript discriminated union narrowing |

## Environment Availability

Step 2.6: No new external dependencies. All tools present.

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| wails CLI | Binding regeneration | ✓ | v2.10.2 | — |
| Node.js | Frontend TypeScript compilation | ✓ | v20.19.3 | — |
| TypeScript (npx tsc) | Type checking | ✓ | ^5.2.2 (frontend package) | — |

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | TypeScript compiler (tsc --noEmit) |
| Config file | `frontend/tsconfig.json` |
| Quick run command | `cd frontend && npx tsc --noEmit` |
| Full suite command | `cd frontend && npx tsc --noEmit` |

No runtime test framework exists in this project (TEST-01 through TEST-03 are deferred v2 requirements).

### Phase Requirements → Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| API-01 | `getCatalogHtmlPath` wrapper returns `{success, htmlPath}` / `{success: false}` | manual smoke | `wails dev` → open catalog | N/A (no test files) |
| API-02 | `readHtmlFile` wrapper returns `{success, content}` / `{success: false}` | manual smoke | `wails dev` → open catalog | N/A |
| API-03 | `createCatalog` wrapper returns `{success, jsonPath, htmlPath, fileCount, totalSize}` | type check | `cd frontend && npx tsc --noEmit` | ✅ |

### Sampling Rate
- **Per task commit:** `cd /Users/ken/dev/storcat/frontend && npx tsc --noEmit`
- **Per wave merge:** `cd /Users/ken/dev/storcat/frontend && npx tsc --noEmit`
- **Phase gate:** `tsc --noEmit` green before `/gsd:verify-work`

### Wave 0 Gaps
None — no test infrastructure gaps. TypeScript compiler is the validation mechanism. No new test files are required.

## Open Questions

1. **Should createCatalog consumer display fileCount/totalSize?**
   - What we know: `CreateCatalogTab.tsx` shows only "Catalog created successfully!" and does not use the result fields
   - What's unclear: Is this the desired UX for v2.0.0?
   - Recommendation: Leave the consumer unchanged in Phase 5. Display logic is ENH-03 (deferred). Phase 5 success criteria only require the shim to forward the fields — not the consumer to display them.

2. **Should the success envelope type be explicitly declared?**
   - What we know: The `window.electronAPI` type is inferred from `typeof wailsAPI` via the `declare global` block at the bottom of wailsAPI.ts
   - What's unclear: Whether adding an explicit return type annotation would improve maintainability
   - Recommendation: No explicit return type annotation needed. TypeScript infers the return type from the implementation. Adding annotations would require a separate interface definition and increase maintenance surface. Defer to a future refactor.

## Sources

### Primary (HIGH confidence)
- Direct code inspection: `frontend/src/services/wailsAPI.ts` — current implementation analyzed line by line
- Direct code inspection: `frontend/wailsjs/go/main/App.d.ts` — current binding types verified
- Direct code inspection: `frontend/wailsjs/go/models.ts` — `CreateCatalogResult` struct verified
- Direct code inspection: `app.go` — `GetCatalogHtmlPath` and `ReadHtmlFile` Go implementations verified
- `git diff` after `wails generate module` — confirmed bindings are already current (zero diff)
- `npx tsc --noEmit` — confirmed TypeScript compiles clean before Phase 5 changes

### Secondary (MEDIUM confidence)
- Phase 04 commit `437313a1` — established `as const` discriminated union pattern
- Phase 04 SUMMARY.md — confirmed API-01, API-02, API-03 wrappers and consumers updated

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — all libraries already in use; no new dependencies
- Architecture: HIGH — direct inspection of all relevant files; `wails generate module` confirmed no binding changes needed
- Pitfalls: HIGH — TypeScript strict mode behavior verified by running compiler

**Research date:** 2026-03-25
**Valid until:** Until Go method signatures change (stable — no Go edits planned for Phase 5)
