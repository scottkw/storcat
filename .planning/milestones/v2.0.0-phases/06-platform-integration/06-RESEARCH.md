# Phase 6: Platform Integration - Research

**Researched:** 2026-03-25
**Domain:** Wails v2 CSS drag regions, Go build-time version injection
**Confidence:** HIGH

## Summary

Phase 6 addresses two independent platform concerns. PLAT-01 requires the macOS header to be draggable as a window titlebar. PLAT-02 requires the version string displayed in the UI to come from the build, not a hardcoded constant. Both are well-understood problems with clear Wails v2 solutions that require only small, targeted changes.

For PLAT-01: Wails v2 has a first-class CSS drag mechanism — any element with the CSS custom property `--wails-draggable: drag` becomes a window drag handle. The current `Header.tsx` does not set this property, but the type stub `WebkitAppRegion?: string` is already present, confirming intent. The fix is a one-line CSS addition. No `Frameless: true` is required for the drag property to work; dragging via CSS works on framed windows too.

For PLAT-02: The version is currently hardcoded as `const APP_VERSION = '2.0.0'` in `CreateCatalogTab.tsx`. The canonical approach in Wails v2 is to inject a version variable at build time via `wails build -ldflags "-X main.Version=2.0.0"` on the Go side, then expose it as a bound method (`GetVersion() string`) so the TypeScript frontend can call it on mount. The version source of truth is `wails.json` `info.productVersion` — a single authoritative file that should be read at build time and injected.

**Primary recommendation:** Apply `--wails-draggable: drag` to the `AntHeader` style in `Header.tsx` (with `--wails-draggable: no-drag` on interactive children if needed); add a `GetVersion() string` bound method in Go backed by an ldflags-injected variable; update `wailsAPI.ts` and `CreateCatalogTab.tsx` to call it on mount.

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| PLAT-01 | macOS header is draggable as window titlebar using `--wails-draggable: drag` | CSS drag property confirmed as Wails v2 standard mechanism; apply to `Header.tsx` AntHeader element |
| PLAT-02 | App version is derived at build time (not hardcoded constant), displayed correctly in UI | ldflags `-X main.Version` injection confirmed working in Wails v2.10.2; expose via bound `GetVersion()` method |
</phase_requirements>

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| github.com/wailsapp/wails/v2 | v2.10.2 (already in go.mod) | CSS drag property support, bound methods | Already in use; drag via `--wails-draggable` is the documented Wails v2 mechanism |
| Go ldflags (-X flag) | stdlib | Build-time variable injection | Standard Go linker flag; no external dependency |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| Vite `define` | ^3.0.7 (already in use) | Inject build-time constants into frontend bundle | Alternative to Go bound method — inject `__APP_VERSION__` at frontend build time |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| `wails build -ldflags + GetVersion()` | Vite `define` in vite.config.ts | Vite define is simpler (no new Go method), but couples version to frontend build rather than Go build. Either works; ldflags + Go method is more authoritative (single source: wails.json parsed at build time). |
| `wails build -ldflags + GetVersion()` | `//go:embed wails.json` + parse | Embed + parse is more complex; ldflags injection is simpler and well-proven |
| CSS `--wails-draggable` on Header | `Frameless: true` + custom chrome | Frameless mode requires rebuilding all window chrome (traffic lights, resize); CSS drag on framed window is far simpler |

**Version verification:**
```bash
# Wails CLI installed and confirmed at project:
wails version  # → v2.10.2
go version     # Go 1.23 (from go.mod)
```

## Architecture Patterns

### PLAT-01: CSS Drag Region

**What:** Set `--wails-draggable: drag` on the header element. Wails' WebKit integration intercepts this CSS custom property and treats the element as a native drag region.

**When to use:** Any element that should move the window when clicked and dragged. Works on all platforms (macOS, Windows, Linux) using the same CSS property.

**Key constraint:** Interactive child elements (buttons, inputs) must have `--wails-draggable: no-drag` or they will intercept drag events incorrectly. In the current `Header.tsx`, there are no interactive children in the header — only an icon and title text — so no `no-drag` children are needed.

**Example — Header.tsx:**
```tsx
// Source: https://github.com/wailsapp/wails/blob/master/website/docs/guides/frameless.mdx
<AntHeader
  style={{
    // ... existing styles ...
    '--wails-draggable': 'drag',
  } as React.CSSProperties}
>
```

The TypeScript cast to `React.CSSProperties` does not include CSS custom properties by default. Use an intersection type or cast as `React.CSSProperties & { [key: string]: any }` — the existing code already has `as React.CSSProperties & { WebkitAppRegion?: string }` which can be extended.

**Wails options.App config — no change needed:** `CSSDragProperty` defaults to `"--wails-draggable"` and `CSSDragValue` defaults to `"drag"`. These are already the correct defaults; `main.go` does not need to set them explicitly.

### PLAT-02: Build-Time Version Injection

**Strategy:** ldflags `-X` flag injects version into a Go package-level variable at compile time. The variable is exposed via a bound method. The frontend calls it once on mount and stores it in state.

**Version source of truth:** `wails.json` `info.productVersion` is already `"2.0.0"`. The build script (or manual `wails build` command) reads this and passes it as an ldflags arg. This keeps one canonical location.

**Go side — add to app.go:**
```go
// Source: standard Go ldflags pattern; confirmed via wails build --help
// Declared at package level in main.go or a version.go file
var Version = "dev"  // overridden at build time via -ldflags "-X main.Version=x.y.z"

// Bound method on App struct in app.go
func (a *App) GetVersion() string {
    return Version
}
```

**Build command:**
```bash
# wails build -ldflags flag confirmed present in Wails v2.10.2
wails build -ldflags "-X main.Version=2.0.0"
```

**Frontend side — wailsAPI.ts:**
```typescript
// Import the auto-generated binding after wails generate
import { GetVersion } from '../../wailsjs/go/main/App';

getVersion: async () => {
  try {
    const version = await GetVersion();
    return { success: true as const, version };
  } catch (error: any) {
    return { success: false as const, version: '2.0.0' }; // fallback
  }
},
```

**Frontend side — CreateCatalogTab.tsx:**
```typescript
// Replace hardcoded const:
// const APP_VERSION = '2.0.0';  // REMOVE

// Use state instead:
const [appVersion, setAppVersion] = useState('...');
useEffect(() => {
  window.electronAPI?.getVersion?.().then(result => {
    if (result.success) setAppVersion(result.version);
  });
}, []);
```

### Anti-Patterns to Avoid
- **Hardcoded version in TypeScript constant:** The problem being fixed. `const APP_VERSION = '2.0.0'` in component file requires a code edit on every version bump and will drift.
- **Frameless: true for drag:** Enabling frameless mode just to get drag behavior is overkill and removes native window chrome (macOS traffic lights) requiring full custom re-implementation.
- **Vite define with no Go method:** Technically works but means version lives in two places (wails.json and vite.config.ts).
- **Embedding wails.json at runtime:** Requires JSON parse logic, error handling, and struct definition. ldflags injection is simpler.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Window drag region | Custom mouse event listeners, hit-test logic | `--wails-draggable: drag` CSS property | Wails WebKit integration handles platform-specific drag natively; manual approach fails on resize/move edge cases |
| Version sync between files | Script to grep/replace version strings | ldflags `-X main.Version` + single build command | ldflags inject at compile time with no file writes; version stays in wails.json only |

**Key insight:** Both problems have zero-dependency solutions built into the tools already in use. No new packages or build tooling needed.

## Common Pitfalls

### Pitfall 1: TypeScript type rejection of CSS custom properties
**What goes wrong:** `style={{ '--wails-draggable': 'drag' }}` causes a TypeScript error because `React.CSSProperties` does not include custom CSS properties.
**Why it happens:** React's CSSProperties type is generated from known CSS properties only.
**How to avoid:** Cast the style object with an intersection type: `as React.CSSProperties & Record<string, string>` or the project's existing pattern `as React.CSSProperties & { WebkitAppRegion?: string }` extended with the custom property.
**Warning signs:** TypeScript compile error `Object literal may only specify known properties`.

### Pitfall 2: wails bindings not regenerated after adding GetVersion()
**What goes wrong:** `GetVersion` is not importable from `../../wailsjs/go/main/App` because the auto-generated binding file hasn't been updated.
**Why it happens:** Wails generates TypeScript bindings only when `wails dev` or `wails generate module` is run after Go method changes.
**How to avoid:** Run `wails generate module` (or `wails dev` briefly) after adding `GetVersion()` to `app.go` before editing `wailsAPI.ts`.
**Warning signs:** TypeScript import error on `GetVersion` from wailsjs path.

### Pitfall 3: Version variable in wrong package scope
**What goes wrong:** `var Version = "dev"` declared in `main` package but referenced from app.go methods that are also in `main` — should work, but placement matters.
**Why it happens:** Both `main.go` and `app.go` are in `package main`. A `version.go` file in `package main` is the cleanest home.
**How to avoid:** Create a dedicated `version.go` file in the root package with just `var Version = "dev"`. The `-X main.Version` ldflags arg then resolves unambiguously.
**Warning signs:** Build error `undefined: Version`.

### Pitfall 4: Drag region covers interactive elements
**What goes wrong:** Clicking a button in the header drags the window instead of activating the button.
**Why it happens:** `--wails-draggable: drag` cascades to all children.
**How to avoid:** Add `--wails-draggable: no-drag` style to any interactive elements inside the drag region. The current `Header.tsx` has no interactive children (only icon + title text), so this is not an issue for this project currently.
**Warning signs:** Buttons/links in header are not clickable.

### Pitfall 5: ldflags value not passed to wails build correctly
**What goes wrong:** `wails build -ldflags "-X main.Version=2.0.0"` — the inner quotes may need escaping depending on shell.
**Why it happens:** Shell quoting interacts with the ldflags string parsing.
**How to avoid:** Test with `wails build -dryrun -ldflags "-X main.Version=2.0.0"` first to see the compiled command. On macOS zsh this form works without extra escaping.
**Warning signs:** Binary still shows `"dev"` version after build.

## Code Examples

### Complete Header.tsx drag style addition
```tsx
// Source: Wails v2 frameless docs (github.com/wailsapp/wails/blob/master/website/docs/guides/frameless.mdx)
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

### version.go (new file)
```go
// Source: standard Go ldflags pattern — https://pkg.go.dev/cmd/link
package main

// Version is injected at build time via:
//   wails build -ldflags "-X main.Version=2.0.0"
// Falls back to "dev" in development mode.
var Version = "dev"
```

### GetVersion bound method (app.go addition)
```go
// GetVersion returns the application version injected at build time
func (a *App) GetVersion() string {
    return Version
}
```

### wails build command with version
```bash
# Pass version from wails.json productVersion at build time
VERSION=$(node -e "console.log(require('./wails.json').info.productVersion)")
wails build -ldflags "-X main.Version=${VERSION}"
```

### wailsAPI.ts addition
```typescript
import {
  // ... existing imports ...
  GetVersion,
} from '../../wailsjs/go/main/App';

// Inside wailsAPI object:
getVersion: async () => {
  try {
    const version = await GetVersion();
    return { success: true as const, version };
  } catch (error: any) {
    return { success: false as const, version: 'dev' };
  }
},
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| `WebkitAppRegion: drag` (Electron CSS) | `--wails-draggable: drag` (Wails CSS custom property) | Wails v2 initial release | Different property name; Wails does not use WebkitAppRegion |
| Hardcoded version constant in component | ldflags injection + bound method | Wails v2 idiomatic practice | Version stays in one place (wails.json); no file edits per release |

**Deprecated/outdated:**
- `WebkitAppRegion` in Wails: This is an Electron-only CSS property. The existing type stub in `Header.tsx` (`WebkitAppRegion?: string`) was likely copied from Electron patterns but has no effect in Wails. It can be removed or left harmlessly.

## Open Questions

1. **Should drag work in dev mode?**
   - What we know: `--wails-draggable` works in both dev and production builds.
   - What's unclear: During `wails dev`, the drag region may behave differently because the window has a standard frame in dev mode on some platforms. User testing confirms this is not a blocker — dev mode with standard titlebar already provides drag.
   - Recommendation: Implement drag CSS regardless; it will activate in production builds as expected.

2. **Should the version fallback be `"dev"` or `"2.0.0"`?**
   - What we know: In development (no ldflags), `Version = "dev"`. The frontend will show "dev" if `GetVersion()` is called in dev mode.
   - What's unclear: Whether the user wants "dev" displayed in development mode or the production version.
   - Recommendation: Default to `"dev"` in the Go variable. The wailsAPI wrapper can keep `"dev"` as fallback. The production build always injects the real version.

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| wails CLI | `wails build -ldflags` | Yes | v2.10.2 | — |
| Go | ldflags injection | Yes | 1.23 (go.mod) | — |
| Node.js | Extract version from wails.json in build script | Yes (implied by frontend build) | — | Hardcode version string in build command |

**Missing dependencies with no fallback:** None.

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go `testing` package |
| Config file | none (standard `go test`) |
| Quick run command | `go test ./... -run TestVersion` |
| Full suite command | `go test ./...` |

### Phase Requirements → Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| PLAT-01 | Header element has `--wails-draggable: drag` in rendered output | Manual (visual/runtime — drag requires native WebKit runtime) | Manual: run `wails dev` and drag header | N/A |
| PLAT-02 | `GetVersion()` returns injected version string | unit | `go test . -run TestGetVersion` | No — Wave 0 gap |

### Sampling Rate
- **Per task commit:** `go test ./...`
- **Per wave merge:** `go test ./...`
- **Phase gate:** Manual drag verification + `go test ./...` green before `/gsd:verify-work`

### Wave 0 Gaps
- [ ] `version_test.go` — unit test for `GetVersion()` returning the package-level `Version` variable — covers PLAT-02

*(PLAT-01 CSS drag region cannot be tested with automated unit tests — it requires the WebKit runtime to interpret CSS custom properties. Manual verification is the only viable approach.)*

## Project Constraints (from CLAUDE.md)

- **No Electron dependencies:** All fixes must use Go/Wails patterns. `WebkitAppRegion` is Electron-only; use `--wails-draggable` instead.
- **API surface:** `window.electronAPI` interface must be maintained. New `getVersion` method must be added to `wailsAPI.ts` and the `wailsAPI` object (which is aliased to `window.electronAPI`).
- **Branch:** Work on `feature/go-refactor-2.0.0-clean`.
- **TypeScript conventions:** `camelCase` methods (`getVersion`), `PascalCase` interfaces, strict mode.
- **Go conventions:** `go fmt`, context-aware where applicable (GetVersion needs no context).
- **Version source:** `wails.json` `info.productVersion = "2.0.0"` is the canonical version. ldflags must read from this file, not duplicate the value.
- **Wails v3:** Do NOT migrate to Wails v3 (still alpha as of March 2026).

## Sources

### Primary (HIGH confidence)
- GitHub wailsapp/wails — frameless.mdx — CSS drag property `--wails-draggable: drag` syntax and no-drag override
- `wails build --help` (local CLI v2.10.2) — confirmed `-ldflags string` flag is available
- `pkg.go.dev/github.com/wailsapp/wails/v2/pkg/options/mac` — TitleBar options, no InvisibleTitleBarHeight in v2.10.2
- `pkg.go.dev/github.com/mooijtech/wails/v2/cmd/wails/flags` — `LdFlags string` field in BuildCommon

### Secondary (MEDIUM confidence)
- github.com/wailsapp/wails/issues/2712 — maintainer recommended `//go:embed wails.json` or ldflags for version; confirmed no built-in runtime.GetVersion() in Wails v2
- WebSearch: `wails build -ldflags "-X config.Version=v1.1.1"` usage pattern — multiple sources confirm working

### Tertiary (LOW confidence)
- None — all critical claims verified against official sources or local CLI

## Metadata

**Confidence breakdown:**
- PLAT-01 (CSS drag): HIGH — documented Wails v2 feature, confirmed via official GitHub docs source
- PLAT-02 (version injection): HIGH — `wails build --help` confirms `-ldflags` flag; Go ldflags -X is standard stdlib behavior; maintainer confirmed no built-in runtime alternative
- TypeScript cast pattern: HIGH — project already uses this pattern in Header.tsx

**Research date:** 2026-03-25
**Valid until:** 2026-09-25 (stable Wails v2 — changes only if Wails v2 major update)
