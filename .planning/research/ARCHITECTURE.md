# Architecture Research: StorCat v3.0.0 Workspace Redesign

**Domain:** Desktop app (Go/Wails backend, React/TS frontend) — integrating a full frontend replacement plus new backend capabilities into an existing, working codebase
**Researched:** 2026-08-13
**Confidence:** HIGH for structural/code-level findings (read directly from the repo); MEDIUM for the two externally-verified claims (WebView2/WebKitGTK `color-mix()` support — see §6 and Sources)

This file answers "how do the v3.0 features integrate with the existing architecture" against the actual code in `app.go`, `internal/catalog`, `internal/search`, `internal/config`, `pkg/models`, `frontend/src/contexts/AppContext.tsx`, `frontend/src/services/wailsAPI.ts`, `frontend/src/themes.ts`, and `cli/*.go`, read in full before writing this document.

---

## 1. Frontend state architecture

**Current state:** `AppContext.tsx` is one `useReducer` + one `AppContext.Provider`. Every consumer that calls `useAppContext()` re-renders on **any** dispatch, because the provider passes a single `{ state, dispatch }` object whose reference changes every action. `React.memo` on a child does nothing to prevent this — memo only stops re-renders caused by a parent re-rendering with the same props; it doesn't stop a re-render triggered by the component's *own* `useContext()` call returning a new value.

**Does it scale to the v3 state shape?** Not as-is. The handoff's state list mixes three categories with wildly different change frequencies:

| Category | Fields | Changes on |
|---|---|---|
| Rare / global | `themeId`, `density`, `side` | Settings clicks (~never) |
| Catalog data | `catalogs`, `curId`, and the loaded tree | Catalog switch/create/rescan (occasional) |
| **Hot / per-keystroke** | `railFilter`, `query`, `expanded`, `selected`, `pct/seen/log` (scan) | Every keystroke, every scroll-driven expand, every ~220ms scan tick |

A single context conflates all three. Typing in the rail filter would re-render everything that calls `useAppContext()`, including whatever component owns the 40k-row tree — even if that component's *rendered output* doesn't change, React still has to run its function body and diff.

**Laziest structure that avoids re-rendering the tree on every keystroke:**

1. **Don't lift `railFilter` and palette `query` into global state at all.** Nothing outside the Rail component and the Palette component reads them. Keep both as local `useState` inside those two components. This is the single highest-leverage move — it removes the two hottest keys from any shared context entirely, for free, using nothing but React's own primitives (no library, no context-split ceremony needed for these two).
2. **Split what remains into 2–3 contexts, split by change-frequency, not by "feature area":**
   - `ThemeContext` — `themeId`, `density`, `side`. Changes are rare and are *supposed* to cause a full repaint (every color is derived from the theme), so no memoization concern here.
   - `WorkspaceContext` — `catalogs`, `curId`, the current catalog's flattened tree array, `expanded` (Set/Record), `selected`. This is read by Rail (for the selected-row highlight), Tree, and Details. Still changes less often than a keystroke (once per click/toggle), but it's the one place where memoization matters.
   - Modal/dialog flags (`paletteOpen`, `createOpen`, `settingsOpen`, `dialog`, `detailOverlay`) can be local state owned by whichever component renders the trigger + the overlay (e.g., a small `useState` in `App.tsx` toggled by toolbar/menu handlers), same pattern already used for `catalogModalOpen` today. A shared context is only needed if three or more unrelated components must open the same overlay; the current design has one opener and one owner each, so local state is simpler and there is nothing to memo.
   - Scan progress (`pct`, `seen`, `log[]`, `walkIdx`) is **its own** transient state, local to the create slide-over (it doesn't need to survive the slide-over unmounting except for the "run in background" case — see §3). Keep it out of any shared context; it would otherwise re-render Rail/Tree/Details on every progress tick for no reason.
3. **Memoize the expensive derivation, not the whole tree.** The flattened array is computed once per `LoadCatalog`/`LoadCatalogFlat` call (§2), not per render. The *visible* row list is `useMemo(() => computeVisible(flatArray, expandedSet), [flatArray, expandedSet])` — it must **not** depend on `railFilter` (the rail filter only filters the catalog list, not tree contents) or on `selected` (selection changes row *styling*, not row *visibility*). Each row component is `React.memo`'d and receives only `{row, isSelected}` as props from the virtualizer, so a selection change only re-renders the previously-selected and newly-selected row, not the array recomputation.
4. **The 40k-node concern applies only to the tree pane.** The rail list is bounded by catalog count (tens, not thousands) and the palette result list is capped at 50 by design — neither needs virtualization or context-split defensiveness; filtering them inline on every keystroke is free. Don't over-build memoization there.

Net effect: a keystroke in the rail filter touches only the Rail component's local state; a keystroke in ⌘K touches only the Palette component's local state; neither can reach the Tree pane's subscription because the Tree pane never subscribes to either value.

---

## 2. Tree flattening and virtualization boundary

**Recommendation: flatten in Go, in a *new* method — do not change `internal/search.Service.LoadCatalog`.**

Why not modify the existing `LoadCatalog`: `cli/show.go` calls `search.NewService().LoadCatalog(catalogPath)` directly and walks the **nested** `*models.CatalogItem` to pretty-print a tree with `--depth`/`--no-color` support (confirmed by reading `cli/show.go`). `internal/search` is shared, Wails-free, CLI-facing code — changing its return shape would break the CLI's `show` command and violate the milestone constraint that the CLI's 6 subcommands stay unchanged. `search.LoadCatalog` must keep returning the nested tree.

So: add a new bound method, e.g. `App.LoadCatalogFlat(filePath string) (*models.FlatCatalog, error)`, that calls the existing `searchService.LoadCatalog` (reuse, zero duplication of the dual-format v1/v2 parsing) and then flattens the result with a small new helper — a straightforward DFS mirroring the walk order `internal/catalog.traverseDirectory` already uses (dirs first, then files, alphabetical), producing:

```go
// pkg/models
type FlatNode struct {
    Depth     int    `json:"depth"`
    Name      string `json:"name"`      // display name (basename), not full path
    Path      string `json:"path"`      // full path — needed for search-hit lookup
    Kind      string `json:"kind"`      // "dir" | "file"
    Size      int64  `json:"size"`
    ParentIdx int    `json:"parentIdx"` // -1 for root
}
type FlatCatalog struct {
    Nodes []FlatNode `json:"nodes"`
}
```

Put the flattener in a small new file inside `internal/catalog` (it's a tree-shape transform, closest in spirit to the code already there) or a new `internal/tree` package if you'd rather keep `internal/catalog` scoped to "creation." Either is fine; `internal/catalog` avoids introducing a fourth `internal/*` package for ~30 lines.

**Why Go over TS, given the bridge-cost question:** Wails v2's Go↔JS binding call marshals the return value to JSON once, the same as any other bound method — for 40k nodes this is low hundreds of KB to a few MB either way (flat vs. nested carry comparable field counts per node), and JSON parse of that size is tens of milliseconds in both Go and V8, not a meaningful differentiator. The deciding factors are architectural, not bridge-throughput:
- `ParentIdx` falls out for free during Go's own DFS (it already knows the parent index while recursing) — no separate pass needed.
- It avoids writing and maintaining a second recursive tree-walker in TypeScript for a structure that can be tens of thousands of nodes deep in the worst case (stack-depth-safe iterative DFS in JS is more code than the equivalent in Go, where the existing `traverseDirectory` pattern is already recursive and works fine at this scale).
- The frontend then only ever consumes a flat, render-ready array — simpler mental model for the virtualizer and for ⌘K.

**⌘K ancestor expansion:** once a catalog's `FlatCatalog` is loaded and cached in `WorkspaceContext`, build a `Map<path, index>` client-side (one `O(n)` pass, done once per catalog load). On a palette hit (which carries `FullName`/`FullPath` from `SearchResult`, or the equivalent field on the new flat node), look up its index, then walk `parentIdx` up to `-1`, collecting every ancestor index into the `expanded` set, select the leaf, and scroll the virtualizer to that index. If the hit belongs to a catalog that isn't currently loaded, trigger `LoadCatalogFlat` for it first, then repeat.

**Virtualization:** with fixed row heights (27px compact / 34px comfortable, per the design tokens), a virtualizer is arithmetic, not a library: `startIndex = Math.floor(scrollTop / rowHeight)`, render a small overscan window, absolutely-position rows or use a spacer + transform. This avoids adding a dependency for something the fixed-height constraint makes trivial. If per-folder "show more" capping (the design's fallback affordance) is wanted in addition to full virtualization, it's a rendering-time cap on the *visible* array, not a data change. (`@tanstack/react-virtual` is a reasonable off-the-shelf alternative if the team would rather not hand-roll the scroll math — it's headless, zero-opinion, and handles overscan/resize edge cases already; react-window is lighter but effectively unmaintained. Either is optional, not required, given fixed row heights.)

---

## 3. Progress event stream

**`internal/catalog` must stay Wails-free — and today it almost already is.** `Service.CreateCatalog` already accepts an `onProgress ProgressCallback` parameter (`type ProgressCallback func(path string)`); `app.go`'s `CreateCatalog` currently passes a callback that does nothing (`// For now, we don't send progress updates`), and `cli/create.go` passes `nil`. This is the existing seam — it just needs to be filled in on both ends, not invented.

**Proposed interface change** (in `internal/catalog`, no Wails import added):

```go
type ProgressEvent struct {
    Path      string
    FilesSeen int
    BytesSeen int64
}
type ProgressCallback func(ctx context.Context, evt ProgressEvent) error // error => caller wants to stop
```

Returning an `error` from the callback (checked after each call inside `traverseDirectory`) is the cancellation hook — it lets the App-layer adapter signal "stop" without `internal/catalog` importing `context` cancellation semantics from Wails; a plain `context.Context` passed into `CreateCatalog`/`traverseDirectory` and checked via `ctx.Err()` on each entry (or every N entries, to avoid `Err()` overhead on tiny files) achieves the same thing more idiomatically. Either way, `context.Context` is stdlib, not Wails — safe for the CLI to use too (it can pass `context.Background()`).

- **App-layer adapter** (`app.go`): `progressCallback := func(ctx context.Context, evt catalog.ProgressEvent) error { runtime.EventsEmit(a.ctx, "scan:progress", evt); return nil }` unless a cancel flag is set, in which case return an error to unwind the walk. This is the only place `runtime.EventsEmit` appears — `internal/catalog` never sees it.
- **CLI**: `cli/create.go`'s existing `nil` argument still type-checks against the new signature (a `nil` func value is valid for any func type) — **zero CLI changes required**, confirmed by reading `cli/create.go`. If CLI progress output is later wanted, it's an additive `--verbose` flag passing a callback that writes to stderr.
- **Cancellation, two triggers:**
  - **Explicit "Cancel" in the scanning UI** — frontend calls a new `App.CancelScan()` (or the create call itself accepts a request ID and a paired cancel method); App-layer holds the in-flight `context.CancelFunc` in a field, `CancelScan` invokes it.
  - **Window close mid-scan** — `beforeClose` (already implemented for window-size persistence) additionally checks if a scan is in-flight and calls the same cancel func before returning. Recommend: **cancel, do not silently auto-write a partial catalog on forced close** — the design's "Write partial catalog" is an explicit, informed user action (it shows counts and a reason first); auto-writing on an abrupt close could silently produce a catalog the user never asked for and never saw. Flag this as a product decision for requirements, not something to default silently in either direction.
- **"Write partial catalog" path — smaller change than it looks.** Reading `traverseDirectory` closely: it already degrades gracefully for *most* read errors today — a child directory/file that fails to stat or read is silently `continue`d (skipped) by the parent's loop, and an unreadable directory returns an empty `Contents: []` rather than erroring. The only place a hard error propagates today is `os.Stat` failing on the **root** path passed to `CreateCatalog`, or `filepath.Rel` failing (essentially never). To support "volume disappeared mid-scan," change `CreateCatalog`/`traverseDirectory` to **always return the best-effort partial tree even when returning a non-nil error** (currently it returns `(nil, err)` on failure — change to `(partialItem, err)`), and add an `[]WalkError{Path, Reason}` accumulator (via the same progress callback, or a second parameter) so the UI's error-state log box has real "read error: … input/output error" lines to show, sourced from actual accumulated errors rather than mocked ones.

---

## 4. New backend surface

Every new method below returns the existing `(result, error)` Go shape; `frontend/src/services/wailsAPI.ts` wraps each in the same `try { ... } catch { return {success:false,...} }` pattern already used for all 17 current methods — no change to that convention.

| New method (on `App`) | Home package | New or existing package | Notes |
|---|---|---|---|
| `ListVolumes()` | new `internal/volumes` | **new** | OS-specific: `/Volumes` (macOS), `GetLogicalDrives`/WMI (Windows, likely via `golang.org/x/sys/windows`), `/media/$USER` + `/run/media` (Linux). Needs `volumes_darwin.go` / `volumes_windows.go` / `volumes_linux.go` build-tag split — the existing `internal/*` packages have no precedent for this, so it's the one genuinely new architectural pattern in this milestone. |
| `LoadCatalogFlat(path)` | `internal/catalog` (new file) | existing package, new file | See §2. Calls `search.Service.LoadCatalog` internally — depends on `internal/search`, so if placed in `internal/catalog` watch for an import cycle (`catalog` importing `search`); if `internal/search` doesn't already import `internal/catalog`, this is safe. Confirmed clean: `search/service.go` only imports `encoding/json`, `os`, `path/filepath`, `strings`, `time`, `github.com/djherbis/times`, `pkg/models` — no cycle risk. |
| `RenameCatalog(jsonPath, newTitle)` | `internal/catalog` (new file) | existing package | Rewrites the `<title>` tag in the `.html` file in place (string slice between `<title>` / `</title>`, mirroring how `search.BrowseCatalogs` already *reads* that tag). JSON/filenames untouched, per the frozen-format constraint. |
| `DuplicateCatalog(jsonPath, newRoot)` | `internal/catalog` | existing package | Reuses the already-private `copyFile` helper (promote to exported or add a package-level wrapper) for `.json` (+`.html` if present) with a suffixed filename root. |
| `DeleteCatalog(jsonPath, alsoDeleteHtml)` | new `internal/trash` | **new** | OS-trash, not `os.Remove` — deliberately a separate package from `internal/catalog` since it's an OS-integration concern, not a catalog-format concern. See Sources for library candidates (`github.com/hymkor/trash-go` uses `SHFileOperationW` on Windows and the FreeDesktop.org Trash spec elsewhere; vet before adopting — this is the one place the milestone adds a genuinely new third-party dependency). |
| `RescanAndDiff(catalogPath, volumePath)` | `internal/catalog` (new file, e.g. `diff.go`) | existing package | Re-walks `volumePath` (reuses the existing `traverseDirectory`), loads the existing catalog via `search.LoadCatalog`, diffs by path key (`CatalogItem.Name` is already the stable relative path used today — reuse it as the diff key, no new ID scheme needed). Correctly flagged in the handoff as the biggest new backend piece — it's the only feature needing both a fresh walk *and* a load *and* a comparison. |
| `RevealInFileManager(path)` | small, in `app.go` or new `internal/osutil` | new tiny package (optional) | `exec.Command("open", "-R", path)` macOS / `explorer /select,` Windows / `xdg-open`/`nautilus --select` Linux fallback. Parallels the existing `OpenExternal` in shape; a tiny `internal/osutil` package keeps `app.go` thin and testable the way `internal/config` already does, but inlining it in `app.go` next to `OpenExternal` is equally defensible for something this small — team preference, not an architectural requirement. |
| `WatchCatalogDirectory(dir)` / stop | new `internal/watch` | **new** | See §7. |
| `SetDensity`, `SetDefaultFilenameRoot`, `SetCatalogDirectory`, `SetWriteHtmlDefault`, `SetCopySecondaryDefault`, `SetWatchEnabled` | `internal/config` | existing package | Each is a new `Config` field + a `Set*`/`Save()` method, exactly mirroring the existing `SetTheme`/`SetSidebarPosition` pattern — purely additive, lowest-risk category in this table. |
| `SearchCatalogsCapped(term, dir, limit)` | `internal/search` (new method) | existing package, new method | **Do not modify** `SearchCatalogs` — `cli/search.go` depends on its current uncapped signature/behavior for `--json` output. Add a sibling method returning `(results []*SearchResult, total int, err error)`. |
| Parse-error surfacing | `internal/search` (extend `BrowseCatalogs`) | existing package | `searchInCatalogFile`/`LoadCatalog` currently swallow JSON errors via `continue`/plain `error`. Add `models.ParseError{Path string; ByteOffset int64; Reason string}` (extracted from `*json.SyntaxError.Offset` via type assertion), and extend `CatalogMetadata` with an optional `ParseError *models.ParseError` field (additive — doesn't touch the on-disk catalog format, only the metadata struct returned over the bridge). Populating it means `BrowseCatalogs` must attempt a parse (or at minimum `json.Valid`) per catalog file, not just stat + read the HTML `<title>` as it does today — acceptable given catalog *counts* are small even though individual catalogs are large. |

**Why `internal/volumes` and `internal/trash` are new packages, not folded into `internal/catalog`:** both are OS-integration concerns orthogonal to "what a catalog is" — `internal/catalog` should stay about walking directories and writing JSON/HTML, not about enumerating mount points or talking to Explorer/Finder's trash can. Keeping them separate also keeps `internal/catalog`'s build-tag surface (currently zero) from growing — the volume enumeration is the only piece of this milestone that needs per-OS files, and isolating it means the rest of `internal/catalog` stays a single cross-platform file.

---

## 5. Sidecar count cache

**Where it lives:** a new JSON file alongside the existing config, e.g. `os.UserConfigDir()/storcat/counts-cache.json`, following the exact load/save pattern `internal/config.Manager` already uses (read-whole-file, unmarshal, mutate, write-whole-file) — no new persistence mechanism needed; entry count equals catalog count (tens), so a single JSON map is not a performance concern the way it might be for thousands of entries.

**Key:** `path + "|" + mtimeRFC3339 + "|" + sizeBytes` (or a hash of the three) → `{FileCount int, TotalBytes int64}`. Using stat fields rather than a content hash means invalidation is a cheap `os.Stat` comparison, not a re-read of a potentially multi-GB `.json` catalog file.

**Population without blocking first paint:**
1. `BrowseCatalogs` (unchanged signature, still returns fast — it only stats files and reads the HTML `<title>`) does **not** compute counts. The rail renders immediately using only the JSON file's own byte size (already available today via `CatalogMetadata.Size`) plus, if a cache entry exists for that path+mtime+size, the cached `fileCount`/`totalBytes`.
2. For catalogs with no cache hit, the rail row simply omits the count line (or shows a dash) until it's known — no blocking, no spinner (consistent with the design's "no spinners" rule, which the handoff states explicitly for scan progress and should extend here).
3. Two ways to fill the gap, and they're complementary, not exclusive:
   - **Opportunistic:** `LoadCatalogFlat` (§2) already walks every node to build the flat array — summing count and bytes during that same pass is free. Write the result to the cache the moment any catalog is opened, even if nothing else has warmed it.
   - **Background fill:** after `BrowseCatalogs` returns, the App layer can walk the returned list, compute counts for any still-uncached catalogs one at a time in a background goroutine, and emit a `counts:updated` event per completed catalog so the frontend patches individual rail rows as they arrive — same event-adapter pattern as scan progress (§3), reusing `runtime.EventsEmit`.
4. **Degradation when the cache file is absent, corrupt, or unwritable:** treat it exactly like a full cache miss for every entry — no fileCount/totalBytes shown until a catalog is opened (which recomputes and re-populates it). Never let a cache read/write failure block `BrowseCatalogs` or `LoadCatalogFlat` — this is unindexed convenience data, not part of the catalog contract, so failures there should degrade silently to "unknown count," not propagate as an error to the rail.

---

## 6. Theme token layer

**Recommendation: compute derived tokens at theme-apply time in TypeScript, not via CSS `color-mix()`.**

The design table expresses several tokens as pure mixes (`--l2: mix(--l 55%, --p)`, `--dm: mix(--tx 66%, --bg)`, `--fn`, `--acs`, `--sel`, `--hov`) which map cleanly onto the CSS `color-mix()` function, and one (`--onac`) that needs relative-luminance branching, which `color-mix()` cannot express (no conditional logic in a plain custom-property value).

What each of Wails' three webview backends actually supports:
- **WebView2 (Windows):** Chromium-based; `color-mix()` shipped in Chrome/Chromium 111 (March 2023). WebView2's Evergreen distribution auto-updates in the field, so in practice current-generation Windows machines are well past this. Confidence: MEDIUM — no Wails-specific test found; extrapolated from Chromium version history, which is a reasonable proxy since WebView2 tracks Chromium closely.
- **WKWebView (macOS):** WebKit added `color-mix()` support around the Safari 16.4 timeframe (2023). Wails uses the OS-provided WKWebView, which updates with macOS itself.
- **WebKitGTK (Linux):** this is the risk. Unlike WebView2, **Wails does not bundle or control the WebKitGTK version on Linux — it's whatever the user's distro ships**, and Linux LTS distros (older Ubuntu/Debian point releases) can lag years behind upstream WebKit. A `webkit2gtk` build older than the version that picked up `color-mix()` support would silently fail to parse the custom property, and the derived token would fall back to invalid/inherited (likely transparent or black) rather than erroring visibly — a hard-to-diagnose theming bug on a subset of Linux installs.

Given that the Linux webview version is the one piece of this stack genuinely outside the app's control, the safer default is to **not** depend on `color-mix()` at all: generalize the small luminance helper the handoff already asks to be ported for `--onac` into a `mixHex()` utility used for *all* derived tokens (`l2`, `dm`, `fn`, `acs`, `sel`, `hov`), and set every token — primitive and derived — as a plain hex/rgba string on `:root` at theme-apply time, exactly like today's `applyTheme()` in `App.tsx` already does for the current 16-field `ThemeColors`. This:
- Costs nothing extra at runtime (recomputing ~7 hex values happens once per theme click, not per render — negligible).
- Removes the Linux-version unknown entirely rather than feature-detecting around it.
- Extends, rather than replaces, the exact mechanism already in the codebase — `themes.ts` gains the new primitive fields (`p2`, `ch`, explicit `accent`) called for in the handoff, and `applyTheme()` gains the mix/luminance math as additional computed properties before setting them all as CSS custom properties.

If a future audit confirms the minimum supported WebKitGTK across the actual install base, revisiting this in favor of native `color-mix()` (simpler code, cascade "for free" on theme switch) is a reasonable follow-up — but that's an optimization to make once the platform floor is known, not a default to build the theming system around now.

---

## 7. Directory watching lifecycle

The design's "watch" feature is scoped to the **catalog storage directory** (the flat folder holding `*.json`/`*.html` catalog files, driving `BrowseCatalogs` + the status-bar indicator) — **not** the potentially enormous volumes being scanned. This matters: `fsnotify` is non-recursive by default (confirmed — subdirectories are not watched automatically, and the standard recursive pattern requires walking the tree and adding a watch per subdirectory), but a single non-recursive watch on one flat directory is exactly fsnotify's default, no-frills use case. No recursive-watch complexity is needed here at all.

- **New `internal/watch` package**, Wails-free (mirrors `internal/catalog`/`internal/search` discipline — the CLI never needs this, so it's only ever imported by `app.go`): wraps `fsnotify.NewWatcher()`, exposes something like `Watch(dir string, onChange func()) (stop func(), error)`.
- **Debounce inside the package**, not the caller: a single catalog write typically fires `Create` + `Write` + possibly `Chmod`/`Rename` in quick succession; without a short debounce (e.g. 300ms after the last event before invoking `onChange`), the frontend would re-run `BrowseCatalogs` several times per actual change.
- **Lifecycle, owned by `App`:** start on `startup`/`domReady` if `config.WatchEnabled && config.CatalogDirectory != ""`; restart (stop old, start new) whenever the user changes the catalog directory or toggles the watch setting in Settings; stop in `beforeClose` (the `stop func()` returned by `Watch` is held as an `App` field alongside the scan cancel func from §3).
- **Push to frontend:** the `onChange` closure App provides calls `runtime.EventsEmit(a.ctx, "catalogs:changed")` — same adapter pattern as scan progress and count-cache updates, all three following the same "Go package stays plain, `app.go` is the only place that touches `runtime.EventsEmit`" rule. The frontend's single `EventsOn('catalogs:changed', ...)` listener (registered once at app root) re-invokes `BrowseCatalogs`. The status bar's "● watching …" text is driven directly by the `WatchEnabled` config flag (is watching *on*), not by the event stream (which only signals *something changed*) — those are two different pieces of state and conflating them would make the indicator flicker or lie during the debounce window.

---

## 8. Suggested build order — validated, with dependency call-outs

The handoff's order (shell → rail+tree → palette → create slide-over → settings → catalog actions) is **sound and worth keeping**, with the backend prerequisites made explicit — the handoff already gestures at some of these in "Mocked functionality" but doesn't sequence them against the frontend steps.

1. **Shell** (toolbar, 3-pane grid, status bar shell, token layer, theme switching) — no backend dependency beyond what exists (`GetConfig`/`SetTheme` are already there). **Independently shippable.** The theme-token extension (§6) belongs here, and it's pure `themes.ts`/`App.tsx` TypeScript work — zero Go involvement.
2. **Rail + tree, virtualized** — **backend must precede this step**, not run alongside it: `LoadCatalogFlat` (§2) is a small, isolated, low-risk Go addition (new method, new file, no changes to existing signatures) that the virtualized tree cannot be meaningfully built against otherwise (building it against nested `CatalogItem` first and refactoring later is strictly more work than doing the flatten once, up front). Recommend landing `LoadCatalogFlat` as its own tiny backend phase immediately before this step. The sidecar count cache (§5) is **not** a hard blocker here — the rail degrades fine without counts (§5's point 2) and can be backfilled in a later phase. Once shipped, this step is **independently shippable** — the handoff calls this out and the code review confirms nothing else in the milestone gates it.
3. **⌘K palette** — the existing uncapped `SearchCatalogs` is enough for an MVP palette (compute "N hits" from `results.length` client-side, skip the true "50 of N" cap-at-source line initially). `SearchCatalogsCapped` (§4) is a nice-to-have optimization, not a hard blocker — it can land alongside or slightly after this step without holding it up. **Independently shippable** once step 2's flat-array/ancestor-expansion plumbing exists (palette hits need the ⌘K ancestor-expansion logic from §2, which itself depends on step 2 being done).
4. **Create slide-over** — this step has the **largest, riskiest backend surface** in the milestone and deserves its own dedicated backend phase *before* the scanning/error UI is built, not interleaved with it:
   - **Progress events + cancellation + partial-catalog** (§3) changes the *existing, tested* `ProgressCallback` signature and `traverseDirectory`'s error-return contract — this is the one place the milestone touches working code rather than purely adding to it, so it's the highest-regression-risk backend work here and should be done, reviewed, and its existing `service_test.go` extended, before any frontend scanning-step UI consumes it.
   - **`ListVolumes`** (§4) is purely additive and can be built in parallel with the progress-event work — the step-1 form's volume cards need it, but it has no interaction with the progress/cancellation changes.
   - The form-only part of step 1 (title/filename inputs, options toggles, "WILL WRITE" preview) has no backend dependency at all and can be built immediately, before either backend piece lands, since it only needs the existing `CreateCatalog` call for a non-live "click and wait" version.
5. **Settings modal** — mostly additive `internal/config` fields/setters (§4), mirroring the existing `SetTheme` pattern exactly — low risk, and can be built **in parallel with steps 2–4**, gated only on the theme-token work from step 1 landing first (so the theme cards have real per-theme values to preview). Density and rail-position toggles need no new backend at all beyond what's already scaffolded (`SetSidebarPosition` already exists; density is a pure frontend concern once persisted the same way).
6. **`⋯` catalog actions** — correctly ordered last, and internally ordered correctly too:
   - **Rename, Duplicate** are small, additive, low-risk `internal/catalog` methods (§4) — safe to build as soon as the shell + tree exist to hang the `⋯` menu off of; don't need to wait for the rest of step 6.
   - **Delete-to-Trash** introduces the milestone's one new third-party dependency (`internal/trash`) — worth vetting the library choice early (even before this step, in parallel with earlier steps) so a bad pick doesn't surface late.
   - **Re-scan & diff** is, as the handoff says, the single biggest new backend piece (§4) — it's correctly last both because of its size and because representing diff rows sensibly against the tree pane benefits from the flat-array/`parentIdx` infrastructure already built in step 2.

**Cross-cutting dependency summary:**

```
LoadCatalogFlat (Go, §2)  ─────────────▶ Step 2 (Rail + Tree)
                                              │
                                              ▼
Theme token extension (TS, §6) ──▶ Step 1 ──▶ Step 5 theme cards
                                              │
Progress+cancel+partial (Go, §3) ──────────▶ Step 4 scanning/error UI
ListVolumes (Go, §4)             ──────────▶ Step 4 volume cards
                                              │
Sidecar count cache (Go, §5)  ── optional ──▶ Step 2 rail counts (degrades cleanly if deferred)
Watch (Go, §7)                ── optional ──▶ Status bar "watching" indicator, Settings toggle
```

Everything left of an arrow is Go work; nothing on the right can be meaningfully finished without it, except the two marked "optional," which the frontend can ship against a degraded/absent state and backfill later without rework.

---

## New vs. Modified vs. Deleted Inventory

### React components (`frontend/src/`)

| File | Disposition |
|---|---|
| `components/tabs/CreateCatalogTab.tsx` | **Deleted** — replaced by the create slide-over |
| `components/tabs/SearchCatalogsTab.tsx` | **Deleted** — replaced by the ⌘K palette |
| `components/tabs/BrowseCatalogsTab.tsx` | **Deleted** — replaced by rail + tree |
| `components/ModernTable.tsx` | **Deleted** — no tables in the workspace design; tree and palette results are custom rows per the handoff ("the tables and the tree are custom in this design and should be custom in code too") |
| `components/CatalogModal.tsx` | **Deleted** — tree browsing moves from a modal into the docked tree pane |
| `components/WelcomeContent.tsx` | **Deleted** — replaced by the empty-library state in the tree pane |
| `components/MainContent.tsx` | **Deleted** — the three-tab shell it implements no longer exists |
| `components/Header.tsx` | **Modified or deleted** — its responsibilities (theme switching, settings entry) move into the new 46px toolbar; whether any of its logic is reused depends on implementation, but the three-tab header bar itself is gone |
| `contexts/AppContext.tsx` | **Modified/split** — see §1: split into `ThemeContext` + `WorkspaceContext`, with hot per-keystroke state moved to local component state |
| `services/wailsAPI.ts` | **Modified (extended)** — new wrapper functions added for every method in §4's table; existing wrappers largely unchanged (envelope contract preserved) |
| `themes.ts` | **Modified (extended)** — new primitive fields (`p2`, `ch`, explicit `accent`) plus the derived-token computation from §6; existing `ThemeColors`/`Theme` types grow, don't shrink (old fields can be dropped once nothing references them, but the 11 themes' base identities are preserved) |
| New: toolbar, rail, tree pane, details panel, status bar, ⌘K palette, create slide-over, settings modal, actions menu components | **New** — one component (or a small family) per screen in the handoff's "Screens / views" section |

### Go packages/files

| Package/file | Disposition |
|---|---|
| `internal/catalog/service.go` | **Modified** — `ProgressCallback` signature change (§3), `traverseDirectory`/`CreateCatalog` error-return contract change to support partial catalogs (§3); both changes are backward-compatible with the CLI's `nil` callback argument (confirmed by reading `cli/create.go`) |
| `internal/catalog` (new files) | **New files in existing package** — flatten (§2), rename (§4), duplicate (§4), diff (§4) |
| `internal/search/service.go` | **Modified (additive)** — new `SearchCatalogsCapped` method, `ParseError` surfacing in `BrowseCatalogs`; existing `SearchCatalogs`/`LoadCatalog` signatures **unchanged** (CLI depends on both, confirmed by reading `cli/search.go`/`cli/show.go`) |
| `internal/config/config.go` | **Modified (additive)** — new `Config` fields + `Set*` methods for density, rail position, catalog directory, defaults, watch toggle (§4), mirroring the existing pattern exactly |
| `internal/volumes` | **New package** |
| `internal/trash` | **New package** |
| `internal/watch` | **New package** |
| `internal/osutil` (optional) | **New package**, or reveal-in-file-manager inlined in `app.go` — team preference |
| `pkg/models/catalog.go` | **Modified (additive)** — `FlatNode`, `FlatCatalog`, `ParseError`, `Volume`, diff-result types added; `CatalogItem`/`SearchResult`/`CatalogMetadata`/`CreateCatalogResult` **unchanged** in shape (on-disk format is frozen; `CatalogMetadata` only gains an optional field) |
| `app.go` | **Modified (additive)** — new bound methods per §4's table, new `runtime.EventsEmit` adapters for progress/watch/count-cache-fill (§3, §5, §7), new fields for in-flight scan cancel func and watch stop func |
| `cli/*.go` | **Unchanged** — confirmed no call sites need modification for any of the above |

---

## Sources

- Direct code reading: `app.go`, `internal/catalog/service.go`, `internal/search/service.go`, `internal/config/config.go`, `pkg/models/catalog.go`, `frontend/src/contexts/AppContext.tsx`, `frontend/src/services/wailsAPI.ts`, `frontend/src/themes.ts`, `cli/create.go`, `cli/show.go`, `design_handoff_storcat_ui/README.md`, `.planning/PROJECT.md` (all read in full for this research) — HIGH confidence, primary source.
- [Events System | wailsapp/wails | DeepWiki](https://deepwiki.com/wailsapp/wails/5.4-events-system) — `EventsEmit`/`EventsOn` pattern — MEDIUM confidence (third-party summary, not official docs, but consistent with Wails' known API shape).
- [CSS color-mix() | Chrome for Developers](https://developer.chrome.com/docs/css-ui/css-color-mix) and [color-mix() | Can I use](https://caniuse.com/mdn-css_types_color_color-mix) — Chromium 111 shipped `color-mix()` (March 2023); WebView2 tracks Chromium via Evergreen auto-update — MEDIUM confidence (general Chromium/WebView2 version correlation, not a Wails-specific test).
- [Web Inspector: Add initial support for color-mix CSS values (WebKit commit)](https://github.com/WebKit/WebKit/commit/f6173b46292f4fda92346060b35d6a78bf7eb650) — WebKit added `color-mix()` support around the Safari 16.4 timeframe — MEDIUM confidence; no direct evidence found for the minimum WebKitGTK version in the field on Linux, which is the actual risk case discussed in §6.
- [fsnotify/fsnotify — recursive watching discussion](https://github.com/fsnotify/fsnotify/issues/18) and general fsnotify documentation — confirms fsnotify is non-recursive by default, recursive watching requires manually adding a watch per subdirectory — HIGH confidence (matches fsnotify's documented, stable API design); not relevant to this milestone's scope since the watch target is a single flat directory (§7).
- [hymkor/trash-go](https://github.com/hymkor/trash-go) and [laurent22/go-trash](https://github.com/laurent22/go-trash) — candidate cross-platform Go trash libraries for `internal/trash` — MEDIUM confidence (community libraries, not evaluated in depth; flagged in §4 as needing vetting before adoption, the milestone's only new third-party dependency).
- [TanStack Virtual vs react-window vs react-virtuoso 2026 — PkgPulse Guides](https://www.pkgpulse.com/guides/tanstack-virtual-vs-react-window-vs-react-virtuoso-2026) — react-window is stable but unmaintained, `@tanstack/react-virtual` is actively maintained and headless — LOW/MEDIUM confidence (aggregator summary); noted as optional in §2 since fixed row heights make a hand-rolled virtualizer viable without any dependency.

---
*Architecture research for: StorCat v3.0.0 Workspace Redesign*
*Researched: 2026-08-13*
