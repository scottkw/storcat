# Project Research Summary

**Project:** StorCat v2.0.0 — Go/Wails Migration
**Domain:** Electron-to-Wails desktop app migration (catalog manager)
**Researched:** 2026-03-24
**Confidence:** HIGH

## Executive Summary

StorCat v2.0.0 is a feature-parity milestone: replace the Electron v1.2.3 backend with a Go/Wails v2 backend while keeping the React/TypeScript/Ant Design frontend identical. The research confirms the current codebase is structurally sound — the right patterns (thin `App` coordinator, internal services, `wailsAPI.ts` compatibility shim) are already in place. The remaining work is a well-scoped set of correctness gaps between what the Electron backend returned and what the Go backend currently returns. None of these gaps require architectural changes; they all resolve within the existing layer boundaries.

The recommended build order flows bottom-up through the dependency graph: fix data models first, then catalog and search services, then config manager additions, then the `App` API surface, then Wails lifecycle wiring, and finally the TypeScript shim and macOS-specific UI fixes. This order ensures no phase is blocked by an upstream gap and prevents the need to revisit already-completed layers. The most cross-cutting change — window state persistence — touches four layers (config, app, main.go lifecycle, wailsAPI.ts) and is correctly sequenced last among the Go-side changes.

The primary risks are silent runtime failures: nil slices serializing as JSON `null`, IPC response envelope mismatches, and Wails runtime API calls placed in the wrong lifecycle hook. All three are well-documented patterns with clear mitigations. The macOS window position coordinate drift bug is fixed in the current Wails v2.10.2 release (PR #3479). Cross-platform webview differences (WKWebView/WebView2/WebKitGTK) require a final multi-platform testing pass before declaring parity complete.

---

## Key Findings

### Recommended Stack

The stack is already fixed — the migration is in progress and the core technology choices (Go 1.23+, Wails v2.10.2, React 18, TypeScript, Vite, Ant Design 5) are correct and verified from the existing `go.mod` and `package.json`. Do not migrate to Wails v3 (still alpha as of March 2026). Do not switch the frontend package manager from npm to pnpm — `wails.json` specifies npm and the toolchain depends on it.

The key toolchain note: `wails dev` regenerates `wailsjs/` bindings on every run. After any Go struct or method signature change, run `wails dev` (or `wails generate module`) before writing frontend code — stale bindings are a silent type safety failure.

**Core technologies:**
- **Go 1.23 / Wails v2.10.2:** Desktop runtime and IPC layer — single native binary, typed auto-generated TS bindings, no V8 ARM64 crash risk
- **React 18 / TypeScript 4.6 / Ant Design 5:** Frontend UI — unchanged from Electron version, zero migration cost
- **`wailsAPI.ts` shim:** Compatibility adapter — translates Wails' Promise-based IPC into the `window.electronAPI` envelope interface all React components expect
- **Go stdlib only:** `encoding/json`, `os`, `filepath`, `embed`, `context` — no third-party Go dependencies needed for this milestone
- **`os.UserConfigDir()` JSON config:** Platform-appropriate config at `~/Library/Application Support/storcat/config.json` (macOS) — already implemented

### Expected Features

All items are mandatory for v2.0.0 — this milestone's explicit goal is full feature parity. There are no defer items. See `.planning/research/FEATURES.md` for full detail and the dependency graph.

**Must have (table stakes — all required for v2.0.0 parity):**
- `loadCatalog` Go method — missing entirely; straightforward `os.ReadFile` + JSON unmarshal
- `createCatalog` returns `{fileCount, totalSize, paths}` metadata — currently returns `error` only
- JSON output format: bare root object `{...}` not array `[{...}]` — required for v1 backward compatibility
- Empty directory `contents` field: `[]` not `null` — Go nil slice marshals as `null`, breaking frontend `.map()` calls
- Symlink traversal: follow symlinks (use `os.Stat`), not skip (current `os.Lstat`) — silent data loss on symlink-heavy filesystems
- Window state persistence (size + position) — save on `OnBeforeClose`, restore in `OnDomReady`
- Window persistence toggle — currently stubbed as `return true`; needs real config persistence
- `getCatalogHtmlPath` existence check + `{success, htmlPath}` envelope — currently returns bare string with no file check
- `readHtmlFile` response envelope — wrapper silently returns `null` on error
- Header drag region (macOS titlebar) — use `--wails-draggable: drag` CSS property, not Electron's `WebkitAppRegion`
- Browse metadata `size` field and timestamp semantics — `CatalogMetadata` missing `size`; `Modified` type inconsistency
- Version sourced from embedded build artifact — currently hardcoded constant

**Should have (differentiators already delivered by migration):**
- Typed IPC via Wails binding generation (compile-time-checked method calls, no stringly-typed channels)
- Native Go config file at OS-standard location (replaces Electron's split `window-state.json` + `preferences.json`)
- No V8 JIT disable workaround — structural advantage of native Go binary
- `ResolvesAliases: true` in file dialogs — macOS Finder aliases resolved transparently

**Defer (v3+):**
- Wails v3 migration — wait for stable release
- Progress reporting via `runtime.EventsEmit` during catalog creation
- Goroutine worker pool for catalogs with 100K+ files

### Architecture Approach

StorCat uses the standard Wails v2 single-bound-struct pattern. One `App` struct in `main` is the sole IPC surface exposed to the frontend. Three stateless internal services (`catalog.Service`, `search.Service`, `config.Manager`) handle business logic. Shared data types live in `pkg/models`. The auto-generated `wailsjs/` directory is the mechanical IPC boundary; `wailsAPI.ts` is the hand-written compatibility layer that reconstructs the `window.electronAPI` envelope interface. This architecture is correct — do not restructure it.

**Major components:**
1. `App` (main) — thin coordinator: routes IPC calls to services, calls `runtime.*` APIs, wires Wails lifecycle hooks
2. `catalog.Service` (internal/catalog) — directory traversal, JSON + HTML generation, file copy; stateless
3. `search.Service` (internal/search) — catalog reading, recursive search, browse metadata; stateless
4. `config.Manager` (internal/config) — JSON config read/write at OS-standard path; stateless
5. `pkg/models` — shared data types only (`CatalogItem`, `SearchResult`, `CatalogMetadata`, `CatalogResult`); no logic
6. `wailsAPI.ts` — single frontend integration point; wraps all `wailsjs/` calls in `{success, data}` envelopes

**Hard isolation rules:**
- `internal/*` packages must NOT import each other and must NOT import the Wails runtime
- Only `App` calls `runtime.*` functions
- Components never import from `wailsjs/` directly — always through `wailsAPI.ts`

### Critical Pitfalls

1. **IPC envelope mismatch** — Wails resolves Promises with raw Go return values; frontend expects `{success, data}`. The `wailsAPI.ts` shim must construct envelopes in TypeScript, not Go. Two methods currently missing this: `getCatalogHtmlPath` and `readHtmlFile`.

2. **Nil slice → JSON `null`** — Go nil slices marshal as `null`; Electron always returned `[]`. Remove `omitempty` from `Contents` and initialize with `make([]*CatalogItem, 0)`. Frontend `.map()` calls will panic on `null`.

3. **`filepath.Walk` silently skips symlinks** — Go's walk functions use `os.Lstat` and do not follow symlinks by design. Replace with a custom recursive walker using `os.Stat` for symlink targets. Missing entries produce no error — only catalog comparison reveals the gap.

4. **Window runtime calls in `OnStartup`** — Wails runtime APIs (`WindowSetSize`, `WindowSetPosition`) are unreliable in `OnStartup` because the window is initializing in a separate goroutine. Use `OnDomReady` for all window state restoration. Causes panics on Windows; silent failures on macOS.

5. **Wrong CSS drag property** — Electron uses `WebkitAppRegion: drag`; Wails uses `--wails-draggable: drag`. The Electron property silently does nothing in a Wails WebView. Frameless window becomes undraggable on macOS with no console error.

---

## Implications for Roadmap

Based on the dependency graph in ARCHITECTURE.md and the priority ordering in FEATURES.md, the correct phase structure flows strictly bottom-up. Each phase fully completes its layer before the next phase begins, ensuring no rework.

### Phase 1: Data Model Correctness
**Rationale:** Every subsequent phase depends on `pkg/models` having the right shapes. Fix models first so all downstream changes are built on the correct foundation.
**Delivers:** Correct `CatalogItem`, `CatalogMetadata`, and new `CatalogResult` structs. Removes `omitempty` from `Contents`. Adds `Size int64` to `CatalogMetadata`. Documents `Created`/`Modified` field semantics.
**Addresses:** JSON null vs `[]` (Pitfall 2), browse metadata gaps, `createCatalog` return type
**Avoids:** Reworking service layer after fixing models

### Phase 2: Catalog Service Fixes
**Rationale:** Core data generation logic. Fixing JSON output format here propagates correctly through the rest of the system.
**Delivers:** Correct JSON output (bare root object, `[]` not `null` for empty dirs), symlink traversal, `CatalogResult` return with file count + total size + output paths
**Addresses:** JSON format backward compatibility (Pitfall 10), symlink traversal (Pitfall 3), `createCatalog` metadata return
**Avoids:** Stale bindings (regenerate after struct changes per Pitfall 6)

### Phase 3: Search Service + Browse Metadata
**Rationale:** Independent of catalog service; depends only on models. Completes the backend read-path.
**Delivers:** `BrowseCatalogs` with correct `size` field, consistent RFC3339 timestamps, documented `ctime` semantics
**Addresses:** Browse metadata gaps, `time.Time` serialization (Pitfall 7)
**Avoids:** Date format mismatch at the frontend rendering layer

### Phase 4: Config Manager Additions
**Rationale:** Window persistence and toggle features require config changes before `App` layer can wire them. Config is stateless and independent of the other services.
**Delivers:** `WindowX`, `WindowY`, `WindowPersistenceEnabled` fields in config; `SetWindowPosition` and `SetWindowPersistence` methods
**Addresses:** Window state persistence prerequisite
**Avoids:** Partial implementation of persistence that stores size but not position

### Phase 5: App Layer API Surface
**Rationale:** `App` is the IPC surface. With models, services, and config correct, complete the `App` method set.
**Delivers:** `LoadCatalog` method (new), corrected `CreateCatalog` signature returning `*CatalogResult`, `GetCatalogHtmlPath` with existence check, window position get/set methods, `GetWindowPersistence`/`SetWindowPersistence`, version from embedded source
**Addresses:** Missing `loadCatalog` (table stakes), `getCatalogHtmlPath` existence check
**Avoids:** IPC envelope mismatch (Pitfall 1) — Go methods return `(T, error)`, envelopes stay in TypeScript

### Phase 6: Wails Lifecycle Wiring
**Rationale:** Window state restoration requires `OnDomReady` to be registered. This is a `main.go` change that cannot be done until config additions (Phase 4) are complete.
**Delivers:** `OnDomReady` registered and implemented in `app.go`; window size/position restored on launch; `OnBeforeClose` saves final window state; initial window dimensions passed to `options.App` to prevent resize flash
**Addresses:** Window persistence (full), `OnStartup` timing trap (Pitfall 8), macOS coordinate drift (Pitfall 4 — fixed in Wails v2.10.2)
**Avoids:** Runtime panics on Windows from premature `runtime.*` calls

### Phase 7: Frontend Shim + TypeScript Updates
**Rationale:** After all Go-side changes are complete and bindings regenerated, update `wailsAPI.ts` to match new method signatures and construct missing envelopes.
**Delivers:** `createCatalog` response forwards `CatalogResult` fields; `getCatalogHtmlPath` wrapper returns `{success, htmlPath}`; `readHtmlFile` wrapper returns `{success, content}`; `getWindowPersistence`/`setWindowPersistence` stubs replaced with real calls
**Addresses:** IPC envelope mismatch (Pitfall 1) for remaining two methods, persistence toggle stub
**Avoids:** Stale type bindings (regenerate `wailsjs/` before writing shim)

### Phase 8: macOS Titlebar + Drag Region
**Rationale:** Independent of all Go changes — purely a CSS + `main.go` options change. Batched last to keep it isolated.
**Delivers:** Functional drag region on macOS frameless window; `--wails-draggable: drag` on header element; `Mac.TitleBar` options in `main.go`
**Addresses:** Drag region regression (Pitfall 5), macOS titlebar parity
**Avoids:** Wrong CSS property (`WebkitAppRegion` silently does nothing in Wails)

### Phase 9: Cross-Platform Verification
**Rationale:** Wails uses OS-native webviews (WKWebView/WebView2/WebKitGTK) which have rendering differences. Verification must happen on all three platforms before declaring parity complete.
**Delivers:** Confirmed parity on macOS, Windows, Linux; validated window state persistence behavior per platform; confirmed Ant Design rendering consistency
**Addresses:** Platform webview differences (Pitfall 11), dialog option differences (Pitfall 9)
**Avoids:** Shipping regressions only visible on non-dev platforms

### Phase Ordering Rationale

- Phases 1-4 are the dependency foundation. Each is independent of the others (models, catalog service, search service, config are isolated) but all must complete before Phase 5.
- Phase 5 (App layer) is the integration point. Do not start until Phases 1-4 are verified.
- Phase 6 (lifecycle wiring) depends on Phase 4 (config) but can overlap with Phase 5 if window persistence config fields are done first.
- Phase 7 (TypeScript shim) is explicitly last in the Go/backend sequence — always regenerate bindings before writing TypeScript.
- Phase 8 (drag region) is independent and can be done at any time but is lowest risk so batched at the end.
- Phase 9 (cross-platform) must be last — it validates the complete integrated system.

### Research Flags

Phases with standard, well-documented patterns (skip deeper research):
- **Phase 1 (models):** Standard Go struct modification, no unknowns
- **Phase 2 (catalog service):** Custom recursive walker pattern is well-documented; the specific fix (`os.Stat` vs `os.Lstat`) is confirmed
- **Phase 3 (search service):** Straightforward field population; date semantics documented
- **Phase 4 (config):** Additive config fields, no architectural risk
- **Phase 7 (TypeScript shim):** Wrapper pattern already established in `wailsAPI.ts`
- **Phase 8 (drag region):** Single CSS property change + `main.go` options, documented in Wails frameless guide

Phases that may need targeted investigation during execution:
- **Phase 5 (App layer) — version embedding:** `//go:embed` of a version file needs a build-time artifact strategy; the exact approach (embed `wails.json`, separate `version.json`, or ldflags) should be decided before implementing
- **Phase 6 (lifecycle wiring) — multi-monitor position restore:** Position clamping to visible screen bounds is recommended but the exact Wails API behavior for multi-monitor scenarios has MEDIUM confidence (Wails issue #2739 unresolved); recommend saving size only on first implementation and deferring position restore
- **Phase 9 (cross-platform) — Windows WebView2 version check:** Older WebView2 versions (pre-118.0.2088.76) have a crash-on-refocus bug; need to confirm whether the Wails build configuration handles the minimum version requirement

---

## Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| Stack | HIGH | Versions verified from existing `go.mod`, `package.json`, `wails.json`. All toolchain choices already in production use. |
| Features | HIGH | Table stakes list derived directly from PROJECT.md active requirements and INTEGRATIONS.md IPC channel audit. No guesswork. |
| Architecture | HIGH | Based on direct source code analysis of the actual codebase + verified Wails v2 official docs for lifecycle semantics. |
| Pitfalls | HIGH (critical), MEDIUM (window persistence) | Critical pitfalls (nil slice, symlinks, IPC envelopes, OnStartup timing) are well-documented. macOS position coordinate drift is fixed in v2.10.2 (PR #3479 confirmed). Multi-monitor position behavior is MEDIUM — Wails issue #2739 open. |

**Overall confidence:** HIGH

### Gaps to Address

- **Version embedding strategy:** The exact mechanism to embed a build-time version string (ldflags vs `//go:embed` vs `wails.json` parse) is not decided. Recommend deciding this in Phase 5 planning before implementation.
- **Window position restore on multi-monitor:** Wails `WindowGetPosition`/`WindowSetPosition` coordinates are monitor-relative with no monitor identity API. Safe approach: implement size-only persistence first; add position restore as a follow-on after cross-platform testing confirms behavior.
- **`Created` field semantics cross-platform:** Go's `os.Stat` does not expose birthtime portably. macOS has `birthtime`, Linux generally does not. `CatalogMetadata.Created` will reflect `ctime` (inode change time) on Linux. This is an acceptable semantic difference but should be documented in the struct definition.
- **WebView2 minimum version enforcement on Windows:** Confirm whether the current Wails build flags handle minimum WebView2 version or whether an explicit check is needed.

---

## Sources

### Primary (HIGH confidence)
- Wails v2 official docs: https://wails.io/docs/ — options, runtime window API, application lifecycle, frameless guide
- Wails v2 GitHub issues #2242, #3301 — error return type → Promise rejection behavior
- Wails v2 PR #3479 — macOS window position coordinate fix (visibleFrame vs frame)
- Direct source analysis: `app.go`, `main.go`, `internal/`, `pkg/`, `frontend/src/services/wailsAPI.ts`
- StorCat `.planning/codebase/INTEGRATIONS.md` — IPC channel audit
- StorCat `.planning/codebase/CONCERNS.md` — known technical debt

### Secondary (MEDIUM confidence)
- Wails GitHub issues #3478, #2739, #1702 — macOS position drift, multi-monitor behavior, window restoration
- Wails GitHub issue #1660 — OnStartup runtime panic
- Go issue #4759 — `filepath.Walk` symlink behavior (documented language behavior, not a bug)

### Tertiary (reference)
- Go nil slice JSON serialization: https://apoorvam.github.io/blog/2017/golang-json-marshal-slice-as-empty-array-not-null/
- Go `time.Time` JSON format: https://www.willem.dev/articles/change-time-format-json/
- Wails as Electron alternative overview: https://dev.to/kartik_patel/wails-as-electron-alternative-4dmn

---

*Research completed: 2026-03-24*
*Ready for roadmap: yes*
