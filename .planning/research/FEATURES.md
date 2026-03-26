# Feature Landscape

**Domain:** Electron-to-Wails migration, desktop catalog manager
**Researched:** 2026-03-24
**Milestone context:** Feature parity fixes for Go/Wails v2.0.0 vs Electron v1.2.3

---

## Table Stakes

Features users expect. Missing = product feels incomplete or broken vs v1.2.3.

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| `loadCatalog` Go method | Electron exposes `load-catalog` IPC channel; frontend depends on reading a single catalog JSON | Low | Missing entirely from Go backend; straightforward `os.ReadFile` + JSON unmarshal |
| `createCatalog` returns metadata | Electron returns `{fileCount, totalSize, paths}`; renderer displays this after creation | Medium | Go returns `error` only; need to accumulate counters during traversal and return a result struct |
| JSON output format: bare object `{...}` | Existing catalogs (and the legacy `sdcat.sh` tool) expect `{"type":"directory",...}`; array format `[{...}]` breaks backward compat and external tools | Low | Go service wraps root in an array; remove that wrapping |
| Symlink traversal (follow, not skip) | Electron uses `fs.stat` which follows symlinks; skipping them silently drops symlinked directories from catalogs | Medium | Go uses `os.Lstat`; change to `os.Stat`, handle cycles carefully |
| Window state persistence (size + position) | v1.2.3 saves size/position on close and restores on next launch; users expect windows to reopen where they left them | Medium | Config struct already has `WindowWidth`/`WindowHeight`; needs `WindowX`/`WindowY` fields, `OnBeforeClose` hook to save via `runtime.WindowGetSize` + `runtime.WindowGetPosition`, restore in `OnDomReady`. Known macOS coordinate drift issue was fixed in Wails v2 (PR #3479 — `visibleFrame` vs `frame` fix) |
| Window persistence toggle (enable/disable) | Electron exposes `set-window-persistence` / `get-window-persistence` IPC; frontend has a settings toggle for this | Low | Currently stubbed as `return true` / `return {success:true}`; needs real persistence in config |
| `getCatalogHtmlPath` existence check + envelope | Electron returns `{success, htmlPath}` after verifying the `.html` file exists via `fs.access`; Go returns a string without checking | Low | Add `os.Stat` existence check; wrap return in `{success bool, htmlPath string}` struct |
| `readHtmlFile` response envelope | Electron returns `{success, content}`; Go returns raw `(string, error)` which the wailsAPI wrapper maps to `null` on error | Low | Either return a struct from Go (`{Success bool, Content string}`) or ensure the TypeScript wrapper surfaces the error correctly — current wrapper silently swallows errors |
| Header drag region (macOS titlebar) | Electron uses `titleBarStyle: 'hiddenInset'` so the traffic lights appear inside the app chrome; users drag the header to move the window | Low | Wails supports this via `--wails-draggable:drag` CSS property OR `WebkitAppRegion: drag`; the Header component style block already has a comment placeholder but no actual drag property set. Also requires `Mac: &mac.Options{TitleBar: &mac.TitleBar{TitlebarAppearsTransparent: true, InvisibleTitleBarHeight: 40}}` (or equivalent) in `main.go` options |
| HTML root node rendering — bare object | Electron writes catalog root as `{type, name, size, contents:[...]}` (no wrapper); Go currently outputs array-wrapped root | Low | Tied to the JSON format fix above; same fix resolves both |
| Empty directory `contents` field — `[]` not `null` | Electron always serializes `contents: []` for directories; Go omitempty on nil slice produces no `contents` field, breaking frontend checks like `item.contents.length` | Low | Remove `omitempty` from `Contents` field in `CatalogItem`, or initialize empty slice instead of nil |
| Browse metadata fields: `size`, `modified` type, `created` semantics | `CatalogMetadata` is missing `size` field; `modified` is `string` but should be a timestamp type; `created` actually reflects file creation time not catalog creation time | Medium | Add `Size int64` to `CatalogMetadata`; fix `Modified` to be consistent with Electron (`mtime`); align `Created` semantics with v1 |
| Version source from config/build | Electron reads version from `package.json` at runtime; Go hardcodes a constant | Low | Embed `wails.json` or a version file at build time using `//go:embed`; expose via `GetVersion()` method |
| API return envelopes match Electron contract | `window.electronAPI` interface expected by all frontend components returns `{success, data}` or `{success, error}` objects; mismatches cause silent failures | Low-Med | The TypeScript `wailsAPI.ts` wrapper already normalizes most calls; the remaining gaps are `getCatalogHtmlPath` and `readHtmlFile` which return raw values instead of envelopes |

---

## Differentiators

Features that Wails/Go enables vs Electron — improvements the v2.0.0 migration delivers without extra work, or with minor additions.

| Feature | Value Proposition | Complexity | Notes |
|---------|-------------------|------------|-------|
| Native Go config file (`~/.config/storcat/config.json`) | Electron stored window state in Electron's `userData` directory, separate from preferences; Go consolidates both into one config file at a predictable OS-standard location | None (already done) | Config manager already uses `os.UserConfigDir()` |
| Typed Go IPC — no stringly-typed channel names | Wails generates TypeScript bindings from Go method signatures; channel name typos are a compile error, not a runtime bug | None (already done) | Wails auto-generates `wailsjs/go/main/App.js` + `.d.ts` |
| Error propagation as Go `error` (not `{success,error}` strings) | Go's native `(T, error)` return pairs give structured errors; the TypeScript wrapper translates them into the Electron-compatible envelope at the boundary | None (already done for most methods) | Pattern already established in `wailsAPI.ts`; remaining gaps are the two file methods |
| No V8 JIT workaround | Electron v1 had to disable V8 JIT globally (`--jitless --no-opt`) due to ARM64 crashes; Go binary is native code | None (structural advantage) | Noted in `CONCERNS.md` as HIGH severity technical debt in Electron version |
| `ResolvesAliases: true` in file dialogs | Wails `OpenDirectoryDialog` accepts `ResolvesAliases` option — macOS Finder aliases are resolved transparently | Low | Add to `SelectDirectory` call in `app.go`; no frontend change needed |
| Window position saved on close, not on resize | Wails `OnBeforeClose` is the correct hook; Electron polled on `resize`/`move` events — the Go version can do a single save at exit, less I/O | Low | Design choice when implementing persistence |

---

## Anti-Features

Electron patterns to explicitly NOT replicate in the Go/Wails version.

| Anti-Feature | Why Avoid | What to Do Instead |
|--------------|-----------|-------------------|
| Separate `window-state.json` + `preferences.json` files | Electron split state across two files in `userData`; confusing, requires two reads on startup | One `config.json` in `~/.config/storcat/` (already the Go approach) |
| `{success, data}` envelope from Go methods directly | Embedding envelope logic in Go methods leaks frontend concerns into the backend; Go's `(T, error)` is the correct abstraction | Keep Go methods returning `(T, error)`; envelope only in the TypeScript `wailsAPI.ts` wrapper at the boundary |
| V8 flags or Electron-specific app lifecycle hacks | Electron required `app.commandLine.appendSwitch`, `app.setPath`, etc.; none of these translate to Wails | Use Wails application options (`options.App{}`) and Go stdlib for all OS integration |
| String-based IPC channel names | Electron's `ipcRenderer.invoke('channel-name')` has no type safety; a typo is a silent runtime failure | Wails binding generation provides compile-time-checked method calls; do not reproduce untyped string channels |
| Polling for window state during resize/move events | Electron apps commonly attach `win.on('resize')` + `win.on('move')` listeners to save state continuously | Use `OnBeforeClose` for a single save at shutdown; avoids thrashing the config file on every drag |
| Global JIT disable for ARM64 stability | Electron workaround that penalized all platforms | Not applicable to Go; do not add any analogous performance-disabling flags |

---

## Feature Dependencies

```
loadCatalog Go method
  (none — self-contained)

createCatalog returns metadata
  → requires traversal accumulator in catalog/service.go
  → requires new CreateCatalogResult struct in pkg/models/

JSON format bare object
  → required by loadCatalog (must read own output)
  → required by backward compatibility with v1 catalogs
  → fixes HTML root node rendering (same root serialization)

Empty directory contents []
  → depends on CatalogItem model change (remove omitempty or init slice)
  → must not break JSON format fix

Symlink traversal
  → depends on catalog/service.go traverseDirectory function
  → independent of model changes

Window state persistence (full)
  → depends on config.Config adding WindowX/WindowY int fields
  → depends on OnBeforeClose in main.go calling runtime.WindowGetSize/GetPosition
  → depends on OnDomReady restoring via runtime.WindowSetSize/SetPosition
  → window persistence toggle depends on config.Config adding WindowPersistenceEnabled bool field

Header drag region (macOS)
  → depends on main.go options adding Mac TitleBar options
  → depends on Header.tsx adding --wails-draggable:drag or WebkitAppRegion style
  → independent of all Go backend changes

getCatalogHtmlPath envelope + existence check
  → return struct needs Success bool + HtmlPath string
  → TypeScript wailsAPI wrapper must be updated to match

readHtmlFile envelope
  → Go method can stay (string, error) if wrapper is fixed
  → OR go method returns {Success bool, Content string} — either way, wrapper must surface errors

Browse metadata fields
  → CatalogMetadata struct in pkg/models/catalog.go
  → BrowseCatalogs in internal/search/service.go must populate new fields
  → independent of catalog creation path

Version from config
  → requires //go:embed of wails.json or dedicated version file
  → requires GetVersion() method on App
  → requires wailsAPI.ts to expose it
```

---

## MVP Recommendation

All items in Table Stakes are MVP — this milestone's explicit goal is full feature parity. There are no "defer" items in scope; every listed fix is required before merging to main.

Priority order (lowest-risk to highest-risk, avoiding dependencies blocking other work):

1. JSON format + empty contents fix — foundational; other fixes build on correct output
2. `loadCatalog` Go method — isolated, unblocks frontend usage immediately
3. `createCatalog` metadata return — isolated to service layer
4. Browse metadata fields — isolated model + service change
5. Version source — simple embed, no structural risk
6. Symlink traversal — targeted change to traversal logic, moderate test surface
7. `getCatalogHtmlPath` + `readHtmlFile` envelopes — small Go + TS changes, pair together
8. Window persistence (full) — touches main.go options + config + OnBeforeClose/OnDomReady; highest cross-cutting surface
9. Header drag region + macOS titlebar — UI + main.go options; validate visually on macOS

---

## Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| Table stakes list | HIGH | Derived directly from PROJECT.md active requirements and INTEGRATIONS.md IPC channel audit |
| Wails window state API | MEDIUM | `runtime.WindowGetSize`, `WindowGetPosition`, `WindowSetSize`, `WindowSetPosition`, `OnBeforeClose`, `OnDomReady` confirmed via official Wails docs search; macOS coordinate drift fix confirmed (PR #3479) |
| macOS titlebar drag region | MEDIUM | `--wails-draggable:drag` CSS property and `Mac.TitleBar.TitlebarAppearsTransparent` + `InvisibleTitleBarHeight` confirmed via official Wails options docs |
| IPC envelope pattern | HIGH | Derived directly from existing `wailsAPI.ts` code and Electron ARCHITECTURE.md analysis |
| Anti-features | HIGH | Directly derived from known Electron technical debt in CONCERNS.md |

---

## Sources

- Wails v2 Window Runtime API: https://wails.io/docs/reference/runtime/window/
- Wails v2 Options Reference (Mac TitleBar): https://wails.io/docs/reference/options/
- Wails v2 Frameless Applications Guide: https://wails.io/docs/guides/frameless/
- Wails v2 Dialog Runtime API: https://wails.io/docs/reference/runtime/dialog/
- Wails PR #3479 — Fix macOS window position coordinate drift: https://github.com/wailsapp/wails/pull/3479
- Wails Issue #2739 — WindowSetPosition relative to launching monitor: https://github.com/wailsapp/wails/issues/2739
- Wails Discussion #977 — OnBeforeClose window close event: https://github.com/wailsapp/wails/discussions/977
- StorCat Electron INTEGRATIONS.md — IPC channel audit: .planning/codebase/INTEGRATIONS.md
- StorCat Electron ARCHITECTURE.md — error handling + persistence patterns: .planning/codebase/ARCHITECTURE.md
- StorCat Electron CONCERNS.md — known technical debt: .planning/codebase/CONCERNS.md
