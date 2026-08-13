# Stack Research: StorCat v3.0.0 Workspace Redesign

**Domain:** Go/Wails desktop app — frontend replacement (React virtualization, custom design system) + backend capability additions (progress events, volume enumeration, OS trash, filesystem watching, font vendoring)
**Researched:** 2026-08-13
**Confidence:** HIGH (Go stdlib/`golang.org/x/sys`/Wails runtime/npm registry verified directly; MEDIUM on third-party Go library popularity/maintenance signals; see per-item notes)

This file covers only the **new** capabilities for v3.0.0. Existing validated stack (Go 1.23, Wails v2.10.2, React 18, TypeScript, Vite, djherbis/times, fatih/color, olekukonko/tablewriter, pkg/browser) is unchanged and not re-evaluated.

---

## Recommended Stack Additions

### Core Technologies

| Technology | Version | Purpose | Why Recommended |
|------------|---------|---------|-----------------|
| Hand-rolled fixed-row virtualization (no library) | n/a — ~100 lines in `frontend/src` | Windowed rendering of the flattened 40k+-node catalog tree | Rows are a hard-coded constant (27px / 34px, per design). Fixed-height windowing is index arithmetic + one spacer div — the exact case a virtualization library adds the least value for the most bundle weight. Matches the design doc's own call-out: "fixed 27/34px rows make windowing trivial." |
| `github.com/Bios-Marcel/wastebasket/v2` v2.0.3 | Go module | Move catalog files to OS trash/recycle bin (Finder Trash, Windows Recycle Bin, freedesktop `~/.local/share/Trash`) on delete-catalog action | Correctly implements 3 non-trivial platform integrations (Windows `SHFileOperationW` undo semantics, freedesktop's two-file `files/`+`info/*.trashinfo` spec, macOS Finder-integrated trash) with **zero cgo** anywhere. This is the one place in this milestone where writing it by hand is the wrong call — see §4 below. |
| `golang.org/x/sys` (already resolved at v0.30.0, indirect via Wails) | promote to **direct** dependency | Volume/mount enumeration + capacity (Windows `GetLogicalDrives`/`GetDiskFreeSpaceEx`, Unix `Statfs`) | Already sitting in `go.sum` at the version Wails pulls — using it directly for volume enumeration adds **zero new modules** to the dependency graph. See §3. |
| `github.com/fsnotify/fsnotify` v1.10.1 | Go module | Watch the catalog directory for changes, drive the "● watching …" status bar indicator | The de facto standard, wraps inotify/kqueue/FSEvents-via-kqueue/ReadDirectoryChangesW behind one API. No serious competing option — see §5. |

### Supporting Libraries (frontend)

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| *(none — plain React)* | n/a | Modal / Select / Input / Tooltip primitives replacing `antd` | See §7 — scope is 4 fully pixel-spec'd primitives; a headless library buys little here. |
| IBM Plex Sans + IBM Plex Mono (static woff2, self-hosted) | Sans 400/500/600, Mono 400/500 | UI + monospace type per design tokens | Vendor via the same pattern already in the repo for Nunito — see §6. |

### What Was Actively Rejected (do not add)

See the "What NOT to Add" table below — `antd` replacement libraries, virtualization libraries as the default choice, `gopsutil`, and headless-UI kits are all deliberately excluded with reasons.

---

## 1. React list virtualization — hand-roll it

**Recommendation: hand-rolled fixed-height windowing, no library.**

### Why hand-rolling is correct here

The design spec fixes row height to exactly two values (`--rh`: 27px compact / 34px comfortable) and already prescribes the state shape: flatten the `CatalogItem` tree once into `{depth, name, kind, size, parentIdx}[]`, derive visible rows from an `expanded: Record<string, boolean>` set, and virtualize. With a **constant** row height, windowing is:

```
startIndex = Math.max(0, Math.floor(scrollTop / rowHeight) - overscan)
endIndex   = Math.min(items.length, Math.ceil((scrollTop + viewportHeight) / rowHeight) + overscan)
totalHeight = items.length * rowHeight
// render items[startIndex..endIndex] absolutely/translateY-positioned inside a
// `height: totalHeight` spacer, listen to onScroll (passive), throttle via rAF
```

That's the entire algorithm — no measurement pass, no `ResizeObserver`, no dynamic-height cache invalidation, none of the complexity that virtualization libraries exist to hide. Density toggle (27↔34) is just swapping the constant. This is genuinely ~80–120 lines including TypeScript types, a `useVirtualRows` hook, and keyboard-nav/scroll-to-selected support (needed for "Expand all" and search-hit-select-and-scroll-into-view). It costs **zero new npm dependencies**, matches the project's demonstrated dependency discipline (stdlib-first Go stack, zero-dep CLI), and avoids fighting a library's API for behavior (per-catalog `expanded` state, `parentIdx`-based ancestor expansion for search-hit reveal) that's already bespoke.

### Library options, if a library is wanted instead

Verified live from npm (2026-08-13):

| Package | Version | React peer range | Bundle (min/gzip) | Maintenance |
|---|---|---|---|---|
| `@tanstack/react-virtual` | 3.14.9 | React 16–19 | ~17KB unpacked-class-lightweight, headless (no UI, just index math + measurement helpers) | Active, TanStack-backed |
| `react-window` | **2.3.0** | React **18–19 only** | 13.1KB min / 4.5KB gzip (bundlephobia, verified) | Actively maintained — v2 is a 2026 full rewrite by the original author (bvaughn) unifying fixed/variable-size lists into one `List` component; last published 16 days ago as of this research. (Older "unmaintained" commentary found in blog posts is stale — do not trust it.) |
| `react-virtuoso` | 4.18.11 | React ≥16 | ~15–17KB gzip, heaviest of the three | Active, richest feature set (grouping, sticky headers, infinite scroll) — overkill for a flat fixed-height list |

**If a library is later justified** (e.g., the team decides hand-rolled scroll-restoration edge cases aren't worth owning), pick `@tanstack/react-virtual` 3.14.9: it's headless (fits "the tree is custom, keep it custom" from the design handoff), has the widest React peer range, and is the standard pick for exactly this pattern (flat array + fixed/variable row heights, imperative `scrollToIndex`). Reject `react-virtuoso` for this use case — its value (grouping, sticky sections, infinite loading) isn't in the spec. `react-window` v2 is a legitimate second choice (smallest verified bundle) but its React 18/19-only peer range and brand-new v2 API (published 2026) is less battle-tested than tanstack's v3 line.

**Do not** reach for `react-window` v1's `FixedSizeList` — that API is gone in the v2 rewrite; any tutorial referencing `FixedSizeList`/`VariableSizeList` imports is stale.

---

## 2. Wails v2.10.2 runtime events

**Recommendation: `runtime.EventsEmit` / `EventsOn`, wired through the `ProgressCallback` hook that already exists in `internal/catalog/service.go`.**

### The API (verified against the v2.10.2 tag source, not docs paraphrase)

Go side (`github.com/wailsapp/wails/v2/pkg/runtime`, from `events.go` at tag `v2.10.2`):

```go
func EventsEmit(ctx context.Context, eventName string, optionalData ...interface{})
func EventsOn(ctx context.Context, eventName string, callback func(optionalData ...interface{})) func()
func EventsOnce(ctx context.Context, eventName string, callback func(optionalData ...interface{})) func()
func EventsOnMultiple(ctx context.Context, eventName string, callback func(optionalData ...interface{}), counter int) func()
func EventsOff(ctx context.Context, eventName string, additionalEventNames ...string)
func EventsOffAll(ctx context.Context)
```

TypeScript side (`wailsjs/runtime/runtime.d.ts`, generated at build time — verified against the v2.10.2 wrapper source, **not** the v3 docs, which describe a different `WailsEvent`/`Emit(): Promise<boolean>` shape that does not exist in v2):

```ts
export function EventsEmit(eventName: string, ...data: any): void;
export function EventsOn(eventName: string, callback: (...data: any) => void): () => void;
export function EventsOnMultiple(eventName: string, callback: (...data: any) => void, maxCallbacks: number): () => void;
export function EventsOnce(eventName: string, callback: (...data: any) => void): () => void;
export function EventsOff(eventName: string, ...additionalEventNames: string[]): void;
export function EventsOffAll(): void;
```

### How the bound-method context is obtained

Already solved in this codebase — `app.go` line 48 stores the context Wails hands to `OnStartup(ctx context.Context)` as `a.ctx`, and every existing `runtime.*` call (`OpenDirectoryDialog`, `BrowserOpenURL`) already threads it through. `EventsEmit` needs the same `a.ctx`, called from an `App` method — no new plumbing required.

### The integration point already exists and is a stub

`internal/catalog/service.go:17` already defines `type ProgressCallback func(path string)` and `traverseDirectory` already calls it during the walk. `app.go:58-62` wires a **no-op** callback with a comment: *"For now, we don't send progress updates... In the future, we could use Wails events to send updates to frontend."* This milestone is that future. The change is localized to `app.go`'s `CreateCatalog` wrapper:

```go
progressCallback := func(path string) {
    runtime.EventsEmit(a.ctx, "scan:progress", path)
}
```

Extend `ProgressCallback` (or add a second callback) to also carry counts (files seen, bytes) so the frontend's percentage/counter fields in the design have real data, since the current signature only passes `path`.

### Gotchas

- **Historical data race** (GitHub issue #2448, reported against v2.3.1 in 2023, fixed via PR #2453): concurrent `EventsOn` registration racing `EventsEmit` on an unsynchronized map. This predates v2.10.2 by many releases and is not a live concern, but it's a good reason to register all `EventsOn` listeners once on mount (React `useEffect` with the returned unsubscribe function as cleanup) rather than churning listeners during a scan.
- **No built-in throttling.** `EventsEmit` fires on every call — on a fast local SSD, `traverseDirectory` can visit thousands of files/second, which would flood the frontend with re-renders if you emit per-file. The design's own scan animation cadence (~220ms ticks) is the right signal: throttle emission in Go with a simple `time.Since(lastEmit) > 100*time.Millisecond` gate (always force-emit the final 100% event), not per-file. Do this in Go, not by adding a frontend debounce library — it's a 3-line `if`.
- **v2 has no typed event payload wrapper.** Data arrives as `...interface{}`/`...any` — define a small shared TS type for the progress payload shape (`{path, filesScanned, bytesScanned}`) and cast at the `EventsOn` callback boundary; there's no schema validation built in.

---

## 3. Cross-platform volume/mount enumeration — stdlib + `golang.org/x/sys`, not gopsutil

**Recommendation: pure stdlib for enumeration, `golang.org/x/sys` (already resolved indirectly at v0.30.0) for capacity — promote it to a direct dependency. Do not add `gopsutil`.**

### Why not gopsutil

`shirou/gopsutil/v4` (latest v4.26.7, verified via GitHub releases) is a full system-monitoring library — CPU, memory, host, process, disk — built for tools that need all of that. Pulling it in for "list 3 kinds of mount points and their free/total bytes" adds a dependency whose surface area is 10x the need, on a project (`go.mod`: djherbis/times, fatih/color, olekukonko/tablewriter, pkg/browser — all narrow, single-purpose) that has consistently chosen the smallest tool for the job (e.g., stdlib `flag.FlagSet` over Cobra for the CLI, explicitly documented as CLIP-03).

### The stdlib + x/sys approach, per platform

`golang.org/x/sys v0.30.0` is **already in `go.sum`** as an indirect dependency of Wails itself — using it directly costs nothing new in the module graph, just a promotion from indirect to direct.

| Platform | Enumeration | Capacity (total/free bytes) |
|---|---|---|
| macOS | `os.ReadDir("/Volumes")` | `golang.org/x/sys/unix.Statfs(path, &stat)` → `stat.Blocks * uint64(stat.Bsize)` (total), `stat.Bavail * uint64(stat.Bsize)` (free) |
| Windows | `golang.org/x/sys/windows.GetLogicalDrives()` bitmask, or glob `A:` through `Z:` + `windows.GetVolumeInformation` for the label | `golang.org/x/sys/windows.GetDiskFreeSpaceEx(path, &free, &total, &totalFree)` |
| Linux | `os.ReadDir("/media/" + os.Getenv("USER"))` **and** `os.ReadDir("/run/media/" + os.Getenv("USER"))` (both are used depending on distro/desktop; GVFS-based mounts commonly land in the latter) | `golang.org/x/sys/unix.Statfs(path, &stat)` (same call as macOS — both are "unix" build-tag targets) |

`golang.org/x/sys/windows` is already exercised transitively by Wails on the Windows build (WebView2 interop), so this isn't introducing an unfamiliar platform surface either.

**"Read errors" status** (per the design's volume card): a cheap `os.Stat`/`os.ReadDir` on the mount root that returns `EIO`/permission-denied is sufficient signal — no library needed for that either.

---

## 4. OS trash / recycle bin — use a library here, specifically `wastebasket/v2`

**Recommendation: `github.com/Bios-Marcel/wastebasket/v2` v2.0.3. This is the one deliberate exception to "stdlib/hand-roll first" in this milestone.**

### Why not hand-roll this one

Unlike volume enumeration (3 independent one-line stat calls), correct OS trash behavior is genuinely fiddly per platform:
- **Windows**: `SHFileOperationW` from `shell32.dll` needs a double-null-terminated UTF-16 path buffer and the `FOF_ALLOWUNDO` flag to get real Recycle Bin (undo-able) semantics rather than a permanent delete — easy to get subtly wrong.
- **Linux (freedesktop.org Trash spec)**: requires writing **two** files (`$trash/files/<name>` + a matching `$trash/info/<name>.trashinfo` with the original path and ISO-8601 deletion timestamp), collision-safe renaming, and choosing between `~/.local/share/Trash` and a per-mount-point `.Trash-$uid` when trashing across filesystems.
- **macOS**: genuine Finder-integrated trash (that supports "Put Back") isn't just `os.Rename` into `~/.Trash` — it needs Finder to know about the move.

### Why this library specifically

Verified against the source (not just the README) at the `v2.0.3` tag:
- **Zero cgo, on every platform.** macOS trashes via `os/exec` calling `osascript -e 'tell app "Finder" to delete POSIX file "…"'` (shells out, no cgo). Windows uses `golang.org/x/sys/windows.NewLazyDLL("shell32.dll")` + `syscall` — a lazy DLL load, not cgo. Linux is a pure-Go freedesktop-spec implementation. This matters directly for this project: cgo would complicate the macOS **universal** (arm64+x86_64 combined) build and Windows/Linux **cross-compilation** from the CI matrix, both of which currently work because the project is cgo-free.
- **Small transitive footprint when imported correctly.** The repo's `cmd/` package (the library's own CLI) depends on `spf13/cobra`, but the library root package (`wastebasket.go` + per-OS files) does not import `cmd/` — `go get github.com/Bios-Marcel/wastebasket/v2` and `import "github.com/Bios-Marcel/wastebasket/v2"` alone pulls only `golang.org/x/sys` (already present), `golang.org/x/text` (small, encoding-only), and `gobwas/glob` (small pattern matcher used internally). Cobra is never compiled into the StorCat binary as long as only the root package is imported.
- **License**: MPL-2.0 (file-level weak copyleft) — safe to use as an unmodified imported dependency in any binary, open-source or not.
- **Maintenance signal**: MEDIUM confidence — single maintainer, 45 GitHub stars, last pushed 2025-04-08. Small but purpose-fit and API-stable; the alternative below is worse on the metric that actually matters for this project.

### Why not `laurent22/go-trash`

The commonly cited alternative uses a `recycle.c` file compiled via **cgo** on Windows — a direct conflict with this project's cgo-free cross-compilation setup. Reject it on that basis alone, independent of maintenance status (27 stars, pushed 2025-03-04, comparable to wastebasket).

Other options surfaced (`trashbox`, `trash-go`, `go-recyclebin`) are smaller/newer projects with less verifiable track record; not recommended over `wastebasket` without a specific reason to prefer them.

---

## 5. fsnotify — watch the catalog directory

**Recommendation: `github.com/fsnotify/fsnotify` v1.10.1 (latest, verified via GitHub releases API, published 2026-05-04).**

There is no serious competing option for this — it's the standard cross-platform fs-event abstraction for Go (inotify on Linux, FSEvents-via-kqueue on macOS/BSD, `ReadDirectoryChangesW` on Windows), and this milestone's need (watch one directory, non-recursive, for add/remove/rename to re-run `BrowseCatalogs`) is exactly its core use case.

### Platform caveats

- **macOS**: fsnotify's `kqueue` backend opens a file descriptor **per watched file**, not just per directory — watching a directory with N catalog files consumes N+1 file descriptors. For a catalog directory (dozens of `.json`/`.html` files, not tens of thousands), this is a non-issue, but don't extend the watch to also cover catalog **contents** at scale. Also note: Spotlight indexing can generate duplicate/spurious events on macOS — debounce regardless (see below).
- **Linux (inotify)**: bounded by `/proc/sys/fs/inotify/max_user_watches` (commonly 8192–524288 depending on distro) — irrelevant here since only the catalog directory itself is watched (1 watch), not a recursive tree.
- **Windows**: `ReadDirectoryChangesW`-based, generally reliable for a single non-recursive directory watch; no known caveats for this project's scope.

### Debouncing guidance

fsnotify emits one event per underlying OS notification, and a single logical change (e.g., writing a catalog JSON file) commonly fires multiple raw events (Create, then Write, then a metadata Chmod). Standard pattern: collect events in a small buffered channel loop and debounce with a ~200–500ms idle timer before triggering `BrowseCatalogs` — e.g., reset a `time.Timer` on every incoming event and act only when it fires. This is a ~20-line goroutine; no debounce library needed (the project already has `github.com/bep/debounce` as an **indirect** dependency via Wails/labstack tooling — check before adding a second debounce mechanism; if reused directly it costs nothing new, but a hand-rolled `time.Timer` reset is simpler for this single call site and avoids depending on an internal-tooling package).

---

## 6. Vendoring IBM Plex Sans + Mono — follow the pattern already in the repo

**Recommendation: self-hosted static woff2, obtained the same way the existing Nunito font already is — via google-webfonts-helper, dropped into `frontend/src/assets/fonts/`, referenced by relative `url()` in CSS.**

### The pattern already exists — extend it, don't invent a new one

`frontend/src/assets/fonts/` already contains `nunito-v16-latin-regular.woff2` + a sibling `OFL.txt`, referenced from `frontend/src/style.css` via `@font-face { src: url("assets/fonts/nunito-v16-latin-regular.woff2") format("woff2"); }`. That's the Wails default-template pattern for offline-safe fonts, and it's already wired end-to-end: Vite (`vite.config.ts`) processes the `url()` reference, copies+hashes the font into `dist/assets/` at build time, and `main.go`'s existing `//go:embed all:frontend/dist` embeds it into the binary automatically. **No new Wails or Vite configuration is needed** — this is purely "add files, add `@font-face` rules."

### Sourcing

- **License**: IBM Plex is licensed under the SIL Open Font License 1.1 (OFL) — the same license family as the existing Nunito vendor (its `OFL.txt` is already SIL OFL). OFL explicitly permits bundling, self-hosting, and subsetting.
- **Source**: pull static woff2 files via `gwfh.mranftl.com/fonts/ibm-plex-sans` and `.../ibm-plex-mono` (the same "Google Webfonts Helper" tool that produced the existing Nunito file — file naming convention `ibm-plex-sans-v{N}-latin-{weight}.woff2` matches the existing `nunito-v16-latin-regular.woff2` pattern exactly). Weights needed per the design tokens: **Sans 400/500/600**, **Mono 400/500** — 5 files total, each with its accompanying `OFL.txt` copied alongside (one copy is sufficient since both Plex families share the license).
- **Subsetting**: gwfh's "latin" subset (the same one already used for Nunito) is sufficient — StorCat's UI text and paths are Latin-script only, and gwfh's latin-subset woff2 files are already small (typically 15–30KB per weight for Plex). Do **not** add a Python `fonttools`/`pyftsubset` toolchain for finer-grained subsetting — it would be a new build-time dependency (this project has no Python tooling today) for marginal gains over gwfh's existing subset, which already matches the established project pattern.

---

## 7. Removing Ant Design — plain React, no headless library

**Recommendation: plain React + a small hand-rolled focus-trap hook for the Modal shell. Do not add Radix, Ariakit, or react-aria.**

### What `antd`/`@ant-design/icons` currently provide (verified by grepping actual imports across `frontend/src`)

`Layout`, `ConfigProvider`+`theme` (algorithm-based theming), `Button`, `Input`, `Typography`, `Tabs`, `Modal`, `Select`, `Switch`, `Radio`, `Card`, `message` (toast notifications), `Space`, `Divider`, `Checkbox`, and icons `Left/Right/SearchOutlined`, `Menu/MenuFold/SettingOutlined`, `Folder/Search/InfoCircleOutlined`, `UnorderedListOutlined`, `FolderAdd/GlobalOutlined`.

Mapping to the redesign, per the design handoff (all geometry/color is already pixel-spec'd, which is most of what a component library buys you):

| Antd today | v3.0.0 replacement |
|---|---|
| `Layout`, `ConfigProvider`/`theme` algorithm | CSS Grid (`grid-template-columns: 268px 1fr 288px`) + the existing CSS-custom-property theme system in `themes.ts`, extended per the design's token table |
| `Tabs` | Removed entirely — no tabs in the workspace design |
| `Button`, `Input`, `Typography`, `Card`, `Space`, `Divider` | Plain elements + inline styles/CSS, exactly as the design's pixel specs (radii, padding, font sizes) already dictate |
| `Switch`, `Radio` | Custom toggle (30×17px track) and segmented control, both fully specified in the design (exact geometry given) |
| `Modal` | Custom `Modal` shell (see below) reused for Settings, Rename, Delete, Re-scan&diff |
| `Select` | Custom segmented control / theme-card grid (per design, these aren't dropdown selects — they're button groups) |
| `message` | **Gap**: no toast is specified in the design handoff. Needed for ambient IPC-failure feedback (`message.error()` call sites today). Build a minimal custom toast (~40–60 lines: array-of-toasts state + auto-dismiss timer + fixed-position stack) — do not add `react-hot-toast`/`sonner`/similar for this. |
| `@ant-design/icons` | The design explicitly specifies exactly 5 inline SVGs (magnifier, folder/card, plus, gear, caret), 10–15px, `stroke: currentColor`/`--dm` — build these directly, no icon package |

### Why not a headless UI library (Radix / Ariakit / react-aria)

The real value of a headless library is encapsulating *behavior* (focus trap, roving `tabindex`, ARIA roles/listbox semantics, portal/layering) while you own the styling. Here:
- The surface is narrow: one `Modal` shell (reused 5×) + 2 segmented controls + inputs + a nearly-decorative tooltip. That's not a growing design system where reuse compounds — it's 4 primitives, fully geometry-specified already.
- The stacking-order requirements are explicit and unusual (`details panel z-index:3`, `create slide-over`/`search palette z-index:6`, `dialogs/Settings z-index:7`, exit-animation-must-stay-mounted-260ms-for-the-slide-over) — these are easier to get right with plain positioned `div`s under direct control than by fighting a library's own portal/layering assumptions for a handful of simultaneously-open surfaces.
- Focus trapping for a modal is a well-known, small recipe (~30–40 lines: query focusable descendants, wrap `Tab`/`Shift+Tab`, restore focus to the trigger element on close, `Escape` to close) — reusable as one hook across all `Modal` instances. This is squarely in "stdlib/small-recipe before dependency" territory, matching the project's Go-side discipline.
- Adding a dependency here works against the project's demonstrated bias (down to CLIP-03's explicit "6 subcommands don't justify ~2MB" reasoning for the Cobra-vs-stdlib CLI decision) for a benefit (behavior-only, no styling help) that's marginal at this scope.

**If scope grows later** (e.g., a real combobox/multiselect appears in a future milestone — not in this design), reconsider `react-aria`'s individual hook packages (`@react-aria/listbox`, etc. — tree-shakeable per-primitive, not a monolith) over Radix's full component set, since react-aria's hooks-only model composes better with an already-fully-custom-styled system than Radix's styled-slot components do.

**Optional, worth a mention, not a recommendation**: the native `<dialog>` element gives a modal free focus-trap + `Escape` handling + a `::backdrop` scrim via `showModal()`, with no library at all. It's a legitimate zero-dependency alternative to the hand-rolled hook for the `Modal` shell specifically. It's not the primary recommendation because the design's slide-over exit animation (mount-for-260ms-after-close) and multi-surface z-index choreography are simpler to reason about with directly-controlled `div`s than with the browser's own top-layer stacking — but if the team wants even less hand-rolled code, evaluate it for the Settings/Rename/Delete/Re-scan modals (not the slide-over, which isn't really dialog-shaped) before ruling it out.

---

## Installation

```bash
# Go (from repo root)
go get github.com/Bios-Marcel/wastebasket/v2@v2.0.3
go get github.com/fsnotify/fsnotify@v1.10.1
# golang.org/x/sys is already resolved (indirect, v0.30.0) — no `go get` needed,
# it becomes a direct dependency automatically once imported and `go mod tidy` is run.

# Frontend — remove
cd frontend
npm uninstall antd @ant-design/icons

# Frontend — add
# (none required if hand-rolling virtualization; if a library is chosen instead:)
npm install @tanstack/react-virtual@3.14.9
```

No new frontend dependency is strictly required for this milestone if the virtualization and UI-primitive recommendations above are followed — the only frontend `package.json` change may be a net **removal** (`antd`, `@ant-design/icons`).

---

## Alternatives Considered

| Recommended | Alternative | When to Use Alternative |
|---|---|---|
| Hand-rolled fixed-row virtualization | `@tanstack/react-virtual` 3.14.9 | Team decides not to own scroll-restoration/keyboard-nav edge cases; headless fits the "custom tree" spirit better than `react-window`/`react-virtuoso` |
| `golang.org/x/sys` + stdlib for volume enumeration | `shirou/gopsutil/v4` v4.26.7 | Only if a future milestone needs broader system stats (CPU/mem/host) beyond disk — not justified for this milestone alone |
| `wastebasket/v2` for OS trash | Hand-rolled per-OS trash | Only if the maintenance/stars risk (single maintainer, 45★) becomes a real blocker later — but hand-rolling correctly (esp. the freedesktop two-file spec and Windows undo flags) is meaningfully more code and more ways to get it subtly wrong |
| Plain React + hand-rolled focus-trap hook | `react-aria` individual hooks | Scope grows to include a real combobox/multiselect not in the current design |
| Plain React + hand-rolled focus-trap hook | Native `<dialog>` element | Team wants even less owned code for the 4 non-slide-over modals specifically; accept the top-layer stacking tradeoff |

## What NOT to Add

| Avoid | Why | Use Instead |
|---|---|---|
| `react-window`/`react-virtuoso`/`@tanstack/react-virtual` as the *default* choice | Fixed 27/34px rows make windowing trivial; a library adds 4.5–17KB+ of bundle and an API surface for a problem that's ~100 lines of index math | Hand-rolled `useVirtualRows` hook (§1) |
| `shirou/gopsutil` | Full system-monitoring library for a 3-line-per-platform stat call; disproportionate to the need and to this project's dependency discipline | `golang.org/x/sys` (already resolved) + stdlib `os.ReadDir` (§3) |
| `laurent22/go-trash` | Uses `recycle.c` — **cgo** on Windows, in direct conflict with this project's cgo-free universal macOS build + cross-compiled Windows/Linux CI matrix | `Bios-Marcel/wastebasket/v2` (§4) |
| `radix-ui`, `@ariakit/react`, `react-aria-components` (full kits) | Narrow scope (4 already-pixel-specced primitives), marginal behavior-only value, works against the project's demonstrated zero/few-dependency posture | Plain React + one small hand-rolled focus-trap hook (§7) |
| `react-hot-toast` / `sonner` (toast libraries) | The one UI gap (`message.error()` replacement) is a ~40-line custom component, not a dependency-worthy problem | Minimal custom toast (§7) |
| Python `fonttools`/`pyftsubset` toolchain | No Python build tooling exists in this project today; gwfh's existing latin-subset woff2 (already used for Nunito) is small enough | google-webfonts-helper, same pattern as the existing Nunito vendor (§6) |
| `spf13/cobra` transitively via `wastebasket`'s `cmd/` package | Only import the `wastebasket/v2` **root** package; importing `cmd/` would pull Cobra into the binary for no reason | `import "github.com/Bios-Marcel/wastebasket/v2"` only |

## Version Compatibility

| Package A | Compatible With | Notes |
|---|---|---|
| `react-window@2.3.0` | React 18.x / 19.x only (peer dep) | Project is on React 18.2 — compatible, but note this is a narrower peer range than `@tanstack/react-virtual`'s 16–19 |
| `@tanstack/react-virtual@3.14.9` | React 16.8–19.x (peer dep) | Compatible with React 18.2 |
| `wastebasket/v2@v2.0.3` | `go.mod` requires Go 1.23.4 | Project is on Go 1.23 toolchain directive (`go 1.23` in `go.mod`) — verify local Go patch version is ≥1.23.4 when adding, or pin to a `wastebasket/v2` commit compatible with exactly 1.23.0 if the CI runner is pinned lower |
| `fsnotify@v1.10.1` | `go 1.17` minimum (per its `go.mod`) | Comfortably compatible with this project's Go 1.23 |
| `golang.org/x/sys@v0.30.0` | Already the version Wails v2.10.2 resolves | No version bump needed — promote indirect → direct |

## Sources

- `pkg.go.dev` + raw GitHub source at tag `v2.10.2` for `github.com/wailsapp/wails/v2/pkg/runtime` (`events.go`) and the frontend wrapper (`v2/internal/frontend/runtime/wrapper/runtime.d.ts`) — HIGH confidence, primary source read directly, not paraphrased docs
- `npm view` (live registry) for `@tanstack/react-virtual`, `react-window`, `react-virtuoso` versions and peer-dependency ranges — HIGH confidence
- `bundlephobia.com/api/size` for `react-window@2.3.0` verified bundle size (13,092B min / 4,528B gzip) — HIGH confidence (live API); tanstack/react-virtual and react-virtuoso sizes are from secondary sources (rate-limited on direct verification) — MEDIUM confidence, approximate
- GitHub Releases API for `fsnotify/fsnotify` (latest `v1.10.1`, published 2026-05-04), `shirou/gopsutil` (latest `v4.26.7`), `Bios-Marcel/wastebasket` (latest `v2.0.3`, pushed 2025-04-08) — HIGH confidence, live API
- Raw GitHub source inspection of `Bios-Marcel/wastebasket` (`wastebasket_darwin.go`, `wastebasket_windows.go`, repo tree, `go.mod`, `LICENSE`) confirming no-cgo implementation and MPL-2.0 license — HIGH confidence, primary source
- Raw GitHub source inspection of `fsnotify/fsnotify` `go.mod` (Go 1.17 minimum) — HIGH confidence
- Local repo inspection: `app.go`, `internal/catalog/service.go`, `internal/config/config.go`, `frontend/src/style.css`, `frontend/src/assets/fonts/` — HIGH confidence, ground truth for integration points
- WebSearch (general web, not primary-source-verified) for headless UI library landscape (Radix/Ariakit/react-aria positioning), IBM Plex OFL licensing, google-webfonts-helper tooling — MEDIUM confidence, cross-checked against multiple results and the project's own existing Nunito-vendoring precedent

---
*Stack research for: StorCat v3.0.0 Workspace Redesign*
*Researched: 2026-08-13*
