# Phase 27: Catalog Actions + Watch - Research

**Researched:** 2026-08-15
**Domain:** Go/Wails backend file operations (OS Trash, filesystem watching, crash-safe atomic writes) + React/TS actions menu and dialogs
**Confidence:** HIGH (both new Go dependencies verified against their actual source on GitHub + `go list -m` against the module proxy; frontend/backend integration points verified by reading the live code this session)

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

**Catalog identity & rename semantics**
- A `title` field is added to the catalog JSON root and becomes authoritative. Backward compatible: absent → existing HTML-then-filename fallback applies unchanged.
- Rename is allowed on a catalog with no `.html` — writes the JSON title field only, no blocked/greyed state.
- The read-side escaping bug is fixed this phase: `BrowseCatalogs`'s raw `strings.Index` `<title>` extraction gains `html.UnescapeString`.
- Duplicate suffixes the filename root `-copy`, then `-copy-2`, `-copy-3` on collision.

**Destructive actions & crash safety**
- Trash uses `github.com/Bios-Marcel/wastebasket/v2` — the ROADMAP's one pre-approved new Go dependency. Do not hand-roll per-platform trash.
- A failed trash operation surfaces the real error and stops. No permanent-deletion fallback, no escape hatch, this phase.
- "Also delete the matching `.html`" is checked by default when `.html` exists, hidden entirely when it does not.
- `WriteFileAtomic` gains `File.Sync()` before close+rename, and gets a real SIGKILL-mid-write verification, closing `WINDOWS.md` entry #6.

**Actions menu & confirmation UI**
- A new minimal `Menu` component built on `useModalBehavior` — anchored popover, `role="menu"`/`role="menuitem"`, arrow-key nav, Escape/click-outside close, focus restored to `⋯`. No dropdown dependency; no second overlay implementation.
- The menu skips scroll-lock via `scrollLockSelector` pointed at a functional no-op target — configuration, not a fork.
- Delete confirmation is a new centred dialog on `useModalBehavior`, matching Phase 26's `SettingsDialog` shell — names both file paths verbatim, carries the HTML checkbox, destructive-styled confirm button (project's first destructive color token).
- Rename uses a text field in that same dialog shell, pre-filled with current title, Enter commits. Not an inline edit.
- The `⋯` button already exists in `DetailsPanel.tsx:43-68`, inert, ARIA withheld. This phase wires it, does not add it.

**File watching**
- Watching uses `fsnotify/fsnotify` — a second new Go dependency, explicit user decision (rejects polling).
- Bursts coalesced in Go with a ~300ms trailing debounce before emitting.
- The event is a bare `catalogs:changed` signal with no payload — rail re-lists via the existing idempotent `browseCatalogs`.
- Watcher lifecycle follows the `WatchDirectory` setting and the current catalog directory — stop/restart on directory change, genuinely release on toggle-off and app quit (not merely ignore events).
- `runtime.EventsEmit` is called from `app.go` only — `internal/catalog` (and any new watcher package) must stay usable from the CLI with no Wails runtime attached.

### Claude's Discretion
- Menu item ordering, labels, and icons within the actions menu.
- The exact JSON `title` field name and its position in the struct.
- Debounce implementation shape (timer reset vs. channel-based coalescing) and whether 300ms is a named constant or configurable.
- Whether the destructive color token is a new theme token per-theme or a single derived value.
- Package layout for the watcher (`internal/watch` vs. a file beside an existing package).

### Deferred Ideas (OUT OF SCOPE)
- Re-scan and diff (ACT-06/07/08, STATE-03) — Phase 28.
- Renaming a catalog's *filenames* — ACT-02 explicitly leaves filenames unchanged.
- Per-file incremental rail patching from watch events — rejected in favour of idempotent full re-list.
- Undo for delete-to-Trash beyond what the OS Trash itself provides.
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| ACT-01 | User can open a catalog actions menu from the `⋯` button in the details panel | Architecture Patterns → Menu component (built on `useModalBehavior`); Code Examples → Menu wiring |
| ACT-02 | User can rename a catalog's title, rewriting the `.html` `<title>` and leaving filenames unchanged | Common Pitfalls → "The HTML has two title occurrences, not one"; Code Examples → surgical title rewrite; `pkg/models/catalog.go` title field addition |
| ACT-03 | User can duplicate a catalog, copying `.json`/`.html` with a suffixed filename root | Architecture Patterns → reuse of `copyFile`/`WriteFileAtomic`; Don't Hand-Roll → collision-suffix loop |
| ACT-04 | User can delete a catalog to the OS Trash after a confirmation naming both paths, with an option to also delete `.html` | Standard Stack → `wastebasket/v2`; Common Pitfalls → macOS AppleScript injection surface, partial multi-path failure |
| ACT-05 | A failed trash operation surfaces as an error and never silently falls back to permanent deletion | Standard Stack → `wastebasket/v2` source audit (confirms no internal fallback on any of the 3 platform backends) |
| ACT-09 | No catalog write can corrupt an existing catalog file if the app crashes mid-write | Common Pitfalls → fsync-before-rename, parent-directory durability; Code Examples → SIGKILL-mid-write test harness design |
| WATCH-01 | User sees `● watching <catalog directory>` in the status bar when watching is enabled | Architecture Patterns → StatusBar segment (frontend already read this session) |
| WATCH-02 | User sees the rail update when catalogs are added/removed/modified outside the app | Standard Stack → `fsnotify` event semantics; Architecture Patterns → debounce → `catalogs:changed` → `CatalogRail` re-list |
| WATCH-03 | User can turn watching off in Settings, and the watcher is released | Common Pitfalls → `Watcher.Close()` semantics, lifecycle wiring in `app.go` |
</phase_requirements>

## Summary

This phase has two genuinely new pieces of external surface (`wastebasket/v2` for OS Trash, `fsnotify` for directory watching) plus a set of backend write-path hardening and frontend interaction-pattern work that all compose from primitives already built in Phases 22–26. Both new dependencies were read directly from their canonical GitHub source this session (not just documentation) — this matters because the most consequential finding of this research is buried in `wastebasket`'s own macOS implementation, not in its README.

**The macOS Trash backend shells out to `osascript` with a hand-built AppleScript string** (`tell app "Finder" to delete POSIX file "<path>"`), escaping only literal double-quote characters in the path before interpolation. This is meaningfully different from every other OS-integration this codebase has built (`RevealInFileManager`, `OpenExternal`) — those always pass the path as its own `exec.Command` argv element, never string-interpolated into an intermediate command language. `wastebasket` cannot be changed (it's a third-party dependency), so this phase's job is to *not make it worse*: the existing `ContainsPath` containment check must gate every path this phase ever hands to `wastebasket.Trash()`, exactly as it already gates `RevealInFileManager` and `OpenExternal`, and the plan should record this AppleScript-interpolation behavior as an accepted, upstream-owned risk rather than something local code can fully close.

On the good-news side: `wastebasket.Trash(paths ...string) error` never falls back to permanent deletion on any of its three platform backends (verified by reading `wastebasket.go`, `wastebasket_darwin.go`, `wastebasket_windows.go`, `wastebasket_nix.go` in full) — every failure path returns an error, satisfying ACT-05 directly. It also silently no-ops on an already-missing path (`os.IsNotExist` → `continue`), which is exactly the property the UI-SPEC's retry button ("Try moving to Trash again") needs: a retry after a partial multi-path failure re-attempts only what's still there.

`fsnotify` is confirmed non-recursive (must `Add()` the one catalog directory only — catalogs are files directly inside it, not nested, so this is sufficient), and confirmed to auto-remove its own watch when the watched path is deleted or renamed — except on Windows, where the watch is *not* auto-removed on a rename of the watched directory itself, a real platform divergence worth a `WINDOWS.md` entry rather than a silent assumption.

`WriteFileAtomic`'s locked `File.Sync()` addition closes the write-then-crash gap for the file's own contents, but full crash durability for a *new* file also requires an `fsync` on the parent directory after `os.Rename` (the directory-entry write is a separate durability domain from the file's own data on POSIX filesystems) — this is not explicitly locked by CONTEXT.md and should be raised as a discretion item for the plan, not assumed silently.

**Primary recommendation:** Route every trash/rename/duplicate path through the existing `osutil.ContainsPath` containment gate before it reaches `wastebasket` or `WriteFileAtomic`; add `File.Sync()` + parent-directory `fsync` to `WriteFileAtomic`; build the watcher as a small `internal/watch` package with a debounce timer and a plain Go channel/callback (no Wails import), wired to `runtime.EventsEmit` only from `app.go`.

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Catalog actions menu (open/position/keyboard nav) | Frontend (React) | — | Pure UI state; no data dependency beyond the already-loaded `state.catalogs` |
| Rename (title write) | API/Backend (Go, `internal/catalog`) | Frontend (dialog + optimistic state update) | The write must be atomic and crash-safe; frontend only renders the result and updates local state optimistically |
| Duplicate | API/Backend (Go, `internal/catalog`) | — | Filesystem copy + collision-suffix logic belongs beside `copyFile`/`WriteFileAtomic`, the existing primitives |
| Delete to Trash | API/Backend (Go, `internal/osutil` new trash helper) | Frontend (confirmation dialog, error display) | OS Trash integration is platform-specific system code; frontend only confirms and displays outcome |
| Directory watching (detect changes) | API/Backend (new `internal/watch` package) | app.go (event emission only) | `fsnotify` requires OS filehandle/kqueue/inotify access; must stay Wails-runtime-free per COMPAT-04, so `app.go` is the sole bridge to the frontend |
| Watching indicator (status bar) | Frontend (React) | — | Purely reflects `state.settings.watchDirectory` + `state.catalogDir`; no live watcher-health channel exists (by design, see Common Pitfalls) |
| Rail refresh on external change | Frontend (React, reuses `CatalogRail.tsx`'s existing directory effect) | API/Backend (`BrowseCatalogs`, unchanged) | The debounced `catalogs:changed` signal triggers a re-list through the *same* code path the directory-change effect already uses — no new read path |

## Standard Stack

### Core

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `github.com/Bios-Marcel/wastebasket/v2` | v2.0.3 [VERIFIED: `go list -m -versions github.com/Bios-Marcel/wastebasket/v2` against the Go module proxy — result: `v2.0.0 v2.0.1 v2.0.2 v2.0.3`] | Cross-platform OS Trash | Pre-approved in `27-CONTEXT.md`/ROADMAP; no viable stdlib alternative exists (Go has no cross-platform trash primitive) |
| `github.com/fsnotify/fsnotify` | v1.10.1 [VERIFIED: `go list -m -versions github.com/fsnotify/fsnotify` — highest listed version `v1.10.1`, dated 2026-05-04 per the project's own CHANGELOG.md, fetched this session] | Cross-platform directory watching (inotify/kqueue/ReadDirectoryChangesW) | Pre-approved as a USER decision in `27-CONTEXT.md`; the de-facto standard Go filesystem-watch library (10,758 GitHub stars, 13,659 importing packages per pkg.go.dev [VERIFIED: `curl` against `api.github.com`/`pkg.go.dev` this session]) |

**Installation:**
```bash
go get github.com/Bios-Marcel/wastebasket/v2@v2.0.3
go get github.com/fsnotify/fsnotify@v1.10.1
```

Both modules require Go ≥1.23 [VERIFIED: `go.mod` fetched from each repo's `master`/`main` branch this session — `wastebasket/v2`'s reads `go 1.23.4`, `fsnotify`'s reads `go 1.23`]; the project's own `go.mod` already declares `go 1.23` and the local toolchain is `go1.26.6` [VERIFIED: `go version` run this session], so no toolchain bump is needed. `fsnotify` pulls `golang.org/x/sys` (the project already carries `golang.org/x/sys v0.30.0` as an indirect dependency [VERIFIED: `go.mod` read this session] — `go get` will promote it to direct, no version conflict since `fsnotify` only requires `≥v0.13.0`).

### Supporting

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| stdlib `html` | (stdlib) | `html.UnescapeString` for the `BrowseCatalogs` title-read fix | Already imported in `internal/catalog/service.go` for the write side; `internal/search/service.go` needs a new `"html"` import — it currently imports only `encoding/json, fmt, os, path/filepath, strings, time` [VERIFIED: `internal/search/service.go:1-14`, read this session] |
| stdlib `os`/`os/exec` (test-only) | (stdlib) | SIGKILL-mid-write crash test harness for ACT-09 | See Code Examples — no new dependency, uses the `exec.Command(os.Args[0], ...)`-style subprocess pattern Go's own `os/exec` test suite uses |

### Alternatives Considered

| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| `fsnotify` | Polling `browseCatalogs` on a timer | Explicitly rejected by `27-CONTEXT.md`: laggy, re-stats the whole directory every tick |
| `fsnotify` | `github.com/rjeczalik/notify` (recursive-capable alternative) | Not needed — catalogs are flat files directly in one directory, no recursion required; `fsnotify`'s non-recursive limitation is a non-issue here |
| `wastebasket/v2` | Hand-rolled per-platform trash (`osascript`, `SHFileOperationW`, freedesktop `.trashinfo` writer) | Explicitly rejected by `27-CONTEXT.md` ("do not hand-roll per-platform trash") — and for good reason: the freedesktop spec alone (mount-point detection, `.Trash` vs `.Trash-$uid`, sticky-bit/symlink checks) is exactly the kind of "don't hand-roll" problem this research flags below |

## Package Legitimacy Audit

> The `gsd_run query package-legitimacy check` seam only supports `npm`/`pypi`/`crates` ecosystems — Go is not covered. Manual verification substituted below using `go list -m -versions` (queries the authoritative Go module proxy) and GitHub/pkg.go.dev metadata.

| Package | Registry | Age | Stars / Importers | Source Repo | Verdict | Disposition |
|---------|----------|-----|--------------------|--------------|---------|-------------|
| `github.com/fsnotify/fsnotify` | Go module proxy | Created 2014-06-28, last push 2026-05-11 [VERIFIED: `api.github.com/repos/fsnotify/fsnotify`] | 10,758 stars, 13,659 importers [VERIFIED: `api.github.com` + `pkg.go.dev`, this session] | `github.com/fsnotify/fsnotify` | OK | Approved |
| `github.com/Bios-Marcel/wastebasket/v2` | Go module proxy | Created 2018-08-29, last push 2025-04-08 [VERIFIED: `api.github.com/repos/Bios-Marcel/wastebasket`] | 45 stars, 3 importers [VERIFIED: `api.github.com` + `pkg.go.dev`, this session] | `github.com/Bios-Marcel/wastebasket` | OK (low adoption, not suspicious) | Approved — pre-approved by ROADMAP; low importer count reflects a niche single-purpose utility, not an authenticity concern (full FreeDesktop-spec implementation with tests/CI verified by direct source read) |

**Packages removed due to [SLOP] verdict:** none.
**Packages flagged as suspicious [SUS]:** none — both packages' source was read in full this session (`wastebasket.go`, `wastebasket_darwin.go`, `wastebasket_windows.go`, `wastebasket_nix.go`, `fsnotify.go`), not merely their metadata. `wastebasket`'s last commit is ~16 months old relative to the project's current date; this is a maintenance-pace observation for the plan to note, not a legitimacy flag — the package is pre-approved by name in the ROADMAP and ships a complete, tested implementation.

## Architecture Patterns

### System Architecture Diagram

```
┌─────────────────────────── Frontend (React) ───────────────────────────┐
│                                                                          │
│  DetailsPanel "⋯" button ──click──▶ Menu (role=menu, useModalBehavior) │
│                                        │                                │
│                    ┌───────────────────┼───────────────────┐           │
│                    ▼                   ▼                   ▼           │
│              Rename dialog       Duplicate (no dialog)  Delete dialog   │
│              (text field)        wailsAPI.duplicate()   (2 path boxes, │
│                    │                   │                  checkbox)    │
│                    ▼                   │                   ▼           │
│         wailsAPI.renameCatalog()       │        wailsAPI.deleteCatalog()│
│                    │                   │                   │           │
└────────────────────┼───────────────────┼───────────────────┼───────────┘
                      ▼                   ▼                   ▼
┌───────────────────────────── app.go (Wails bindings) ───────────────────┐
│  RenameCatalog(path, newTitle)  DuplicateCatalog(path)  DeleteCatalog(  │
│       │                              │                   jsonPath,     │
│       │  containment check           │  containment      deleteHtml)  │
│       ▼  (osutil.ContainsPath)       ▼                       │         │
│  internal/catalog:                internal/catalog:          ▼         │
│   - read JSON, set Title           - copyFile(json)     internal/osutil:│
│   - WriteFileAtomic (json)         - copyFile(html?)      new trash.go │
│   - surgical <title>/<h1>          - suffix -copy[-N]      → wastebasket│
│     rewrite in .html (if exists)                             .Trash()  │
└───────────────────────────────────────────────────────────────────────┘

┌────────────────── Watching (independent of the above) ──────────────────┐
│                                                                          │
│  app.go: SetWatchDirectory(true) / catalog-dir change                  │
│       │                                                                 │
│       ▼                                                                 │
│  internal/watch.Watcher (no Wails import)                              │
│       │  fsnotify.NewWatcher() → Add(catalogDir)                       │
│       │  event loop: any Create/Write/Remove/Rename on *.json/*.html   │
│       │  → reset a 300ms trailing timer                                │
│       ▼  timer fires → callback(app.go-owned)                          │
│  app.go: runtime.EventsEmit(ctx, "catalogs:changed")                   │
│       │                                                                 │
│       ▼                                                                 │
│  Frontend: EventsOn("catalogs:changed") → reuses CatalogRail.tsx's     │
│            existing state.catalogDir effect → wailsAPI.browseCatalogs()│
│                                                                          │
│  WATCH-03: SetWatchDirectory(false) / app quit → watcher.Close()       │
│            (releases inotify fd / kqueue fd / Windows handle)          │
└──────────────────────────────────────────────────────────────────────┘
```

### Recommended Project Structure
```
internal/
├── catalog/
│   ├── atomicwrite.go       # gains File.Sync() before close+rename
│   ├── service.go           # gains RenameCatalog / DuplicateCatalog (or a new actions.go beside it)
│   └── actions.go           # (Claude's discretion) rename/duplicate logic, if split out
├── osutil/
│   ├── reveal.go            # existing — ContainsPath lives here, reused
│   └── trash.go             # new — thin wrapper around wastebasket.Trash(), containment-gated
├── watch/                   # (Claude's discretion — package layout) new
│   └── watcher.go           # fsnotify wrapper, debounce, no Wails import
pkg/models/
└── catalog.go                # CatalogItem gains Title field (position/name: discretion)
app.go                        # new bindings: RenameCatalog, DuplicateCatalog, DeleteCatalog;
                               # watcher start/stop wiring; catalogs:changed emission
frontend/src/components/workspace/
├── Menu.tsx                  # new — shared anchored popover
├── CatalogActionsMenu.tsx    # new — the 3-item menu content (discretion: may fold into Menu.tsx)
├── RenameDialog.tsx          # new
├── DeleteConfirmDialog.tsx   # new
├── DialogShell.tsx           # new — shared 440px shell both dialogs use (per UI-SPEC's "one shared component")
├── DetailsPanel.tsx          # wire the existing OverflowButton
├── StatusBar.tsx             # new .ws-status-right wrapper + watching segment
└── CatalogRail.tsx           # subscribe to catalogs:changed, reuse existing directory effect
```

### Pattern 1: Containment-gated destructive Go binding
**What:** Every new Wails binding that takes a filesystem path (`RenameCatalog`, `DuplicateCatalog`, `DeleteCatalog`) resolves the path with `filepath.Abs` + `filepath.EvalSymlinks`, then checks `osutil.ContainsPath(catalogDir, resolved)` before doing anything else — identical to the existing `GetCatalogHtmlPath`/`OpenExternal`/`RevealInFileManager` pattern.
**When to use:** Any binding reachable from renderer JS that touches a path outside a hardcoded location.
**Example:**
```go
// Source: app.go:753-789 (GetCatalogHtmlPath / OpenExternal), read this session —
// the exact shape to replicate for RenameCatalog/DuplicateCatalog/DeleteCatalog.
func (a *App) GetCatalogHtmlPath(catalogPath string, catalogDir string) (string, error) {
	// ... resolve htmlPath ...
	resolved, err := filepath.EvalSymlinks(htmlPath)
	if err != nil {
		return "", fmt.Errorf("get html path %s: %w", htmlPath, err)
	}
	ok, err := osutil.ContainsPath(catalogDir, resolved)
	if err != nil {
		return "", fmt.Errorf("get html path %s: resolve catalog directory: %w", htmlPath, err)
	}
	if !ok {
		return "", fmt.Errorf("get html path %s: outside configured catalog directory", htmlPath)
	}
	return htmlPath, nil
}
```
This is doubly important for the Trash binding specifically, since `wastebasket`'s macOS backend interpolates the path into an AppleScript string (see Common Pitfalls) — containment is the one thing standing between "worst case: reveals/trashes the wrong file inside `catalogDir`" and "worst case: an attacker-controlled string reaches a `tell app "Finder"` command with no scope restriction at all."

### Pattern 2: Watcher package stays Wails-runtime-free, `app.go` is the sole emitter
**What:** `internal/watch` (or wherever the watcher lands) exposes a constructor that takes a callback (`func()`) or a plain Go channel, never a `context.Context` tied to Wails and never imports `github.com/wailsapp/wails/v2/pkg/runtime`. `app.go` owns the debounced callback → `runtime.EventsEmit` translation, mirroring the existing `throttledProgress` pattern exactly.
**When to use:** Any new backend subsystem that needs to notify the frontend.
**Example:**
```go
// Source: app.go:173-189, throttledProgress — read this session. The existing,
// working precedent for "only app.go calls runtime.EventsEmit."
// internal/catalog must stay usable from the CLI with no Wails runtime attached
// (COMPAT-04) -- all throttling and emission live here. The a.ctx == nil guard
// makes the returned closure safe to call from a plain Go test with no Wails
// runtime attached.
func (a *App) throttledProgress(totalBytes int64) catalog.ProgressCallback {
	var lastEmit time.Time
	return func(u catalog.ProgressUpdate) {
		if a.ctx == nil {
			return
		}
		// ... emit ...
	}
}
```
A watcher package built the same way: `internal/watch.New(dir string, onChange func()) (*Watcher, error)` where `onChange` is supplied by `app.go` as a closure that debounces and then calls `runtime.EventsEmit(a.ctx, "catalogs:changed", nil)`, guarded by the same `a.ctx == nil` check.

### Pattern 3: One shared dialog shell, swap the body (not two dialogs)
**What:** Per `27-UI-SPEC.md`, Rename and Delete-confirmation reuse a single `DialogShell` component (header/body/footer slots) at the Settings-established 440px-narrower-than-660px geometry, rather than two near-duplicate 440px panel implementations.
**When to use:** Both of this phase's new modals.
**Example:** See `frontend/src/components/workspace/settings/` (Phase 26's `SettingsDialog`) for the exact scrim/fade/header/footer shape to extract into the shared shell — `27-UI-SPEC.md` lines 185-230 specify every value.

### Anti-Patterns to Avoid
- **String-concatenating a shell/AppleScript/osascript command from a user-controlled path:** `wastebasket`'s own macOS backend does this internally (unavoidable, third-party) — but nothing in *this codebase's own new code* should do the same. Every new binding must pass paths as their own `exec.Command`/library-call argument, never interpolated into a string later parsed by an interpreter.
- **A second overlay/focus-trap implementation:** `27-CONTEXT.md` is explicit — reuse `useModalBehavior` for both the menu and both dialogs; do not hand-roll a second focus-trap for "it's just a small menu."
- **Threading per-file watch deltas to the frontend:** rejected by `27-CONTEXT.md` in favor of the bare `catalogs:changed` signal + full re-list via the existing idempotent `browseCatalogs` — a second, delta-based code path could disagree with the source of truth.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Moving a file to the OS Trash on macOS/Windows/Linux | Per-platform `osascript`/`SHFileOperationW`/FreeDesktop-trash-spec code | `wastebasket.Trash(paths ...string) error` | The FreeDesktop spec alone (mount-point/topdir detection, `.Trash` vs `.Trash-$uid` with sticky-bit and symlink checks, collision-suffix naming for the trashed file, `.trashinfo` sidecar format) is ~250 lines of subtle, easy-to-get-wrong logic — verified by reading `wastebasket_nix.go` in full this session |
| Cross-platform filesystem change notification | Polling, or hand-rolled inotify/kqueue/ReadDirectoryChangesW syscalls | `fsnotify.NewWatcher()` | Three incompatible native APIs with different event semantics (see Common Pitfalls) — `fsnotify` normalizes them into one `Op` bitmask, at the cost of the platform quirks documented below, which is still far less risk than three custom syscall wrappers |
| Verifying a file survives a crash mid-write | Assuming `os.Rename` alone is sufficient | `File.Sync()` before close + rename, ideally + parent-directory `fsync` | `os.Rename`'s atomicity guarantees the destination is never observed in a half-written state *if the process survives to call Rename* — it says nothing about whether the file's own bytes, or the directory entry itself, are durable across a power loss. These are two separate POSIX durability domains |

**Key insight:** every "don't hand-roll" item in this phase maps to a real subtlety this session's source reads exposed directly (the AppleScript escaping in `wastebasket_darwin.go`, the Windows non-removal-on-rename note in `fsnotify.go`, the parent-directory fsync gap) — this is not generic caution, it's what the actual dependency code does.

## Common Pitfalls

### Pitfall 1: `wastebasket`'s macOS Trash implementation is a string-interpolated AppleScript command, not an argv-safe call
**What goes wrong:** Reading `wastebasket_darwin.go` directly (fetched and read in full this session) shows:
```go
// Source: github.com/Bios-Marcel/wastebasket wastebasket_darwin.go (master branch), fetched this session
path = strings.ReplaceAll(path, `"`, `\"`)
osascriptCommand := fmt.Sprintf(`tell app "Finder" to delete POSIX file "%s"`, path)
err = exec.Command("osascript", "-e", osascriptCommand).Run()
```
The path is passed through `exec.Command` as a single argv element (no shell is invoked, so classic shell-injection via `;`/`|`/backticks is not possible), but it *is* interpolated into an **AppleScript string literal** with only `"` escaped. Characters with special meaning inside AppleScript's own string/command grammar (or Unicode tricks around the escape) are not defused by this call.
**Why it happens:** This is upstream, third-party code — cannot be changed, and calling out to `osascript` is the only practical way to reach Finder's real Trash behavior on macOS from Go (Apple provides no public trash syscall).
**How to avoid:** This phase cannot fix `wastebasket` itself. What it *can* do — and must — is guarantee every path reaching `wastebasket.Trash()` has already passed `osutil.ContainsPath(catalogDir, resolved)`, so the AppleScript string can only ever contain a path the app itself already knows is a `.json`/`.html` file inside the configured catalog directory (filenames typically come from `os.ReadDir` during `BrowseCatalogs`, not raw user text — but a catalog *directory* configured to a hostile location, or a future feature that lets a user type an arbitrary filename, would need the same containment check). Record this as an accepted upstream-owned risk in the plan's threat model rather than treating it as fully closed.
**Warning signs:** Any future code path that lets renderer JS supply a raw string that reaches `Trash()` without going through the existing containment check.

### Pitfall 2: The HTML file has the title in *two* places, not one
**What goes wrong:** `internal/catalog/service.go`'s `writeHTMLFile` (read in full this session) writes `html.EscapeString(title)` into **both** the `<title>%s</title>` tag (line 451) and the `<h1>%s</h1>` visible page heading (line 474):
```go
// Source: internal/catalog/service.go:491, read this session
</html>`, html.EscapeString(title), html.EscapeString(title), treeStructure, totalSize, dirCount, fileCount)
```
ACT-02's success-criteria wording ("rewriting the `.html` `<title>` safely") names only the `<title>` tag. A rename that surgically patches only `<title>` and leaves `<h1>` unchanged will produce an HTML file whose browser-tab title and page heading visibly disagree.
**Why it happens:** The phase description's wording is a simplification of the actual generated markup; the generator's own source has two occurrences.
**How to avoid:** The rename operation's HTML rewrite must locate and replace **both** the `<title>...</title>` content and the `<h1>...</h1>` content, using the same `html.EscapeString` treatment the original writer used, or the two will drift. A simple, safe implementation: two independent `strings.Index`-delimited substring replacements (mirroring `BrowseCatalogs`'s own read-side pattern), each re-escaped with `html.EscapeString(newTitle)`.
**Warning signs:** A rename that "works" in the details panel and rail (JSON-sourced) but the `.html` file, when opened, still shows the old title as its big heading.

### Pitfall 3: `Trash(paths ...string)` can partially succeed across multiple paths — this is a feature, not a bug, for the retry UX
**What goes wrong (if not accounted for):** Calling `wastebasket.Trash(jsonPath, htmlPath)` together processes paths sequentially on the nix and darwin backends (windows batches all paths into one `SHFileOperationW` call). If `jsonPath` succeeds and `htmlPath` then fails, the function returns an error — but the JSON file is *already* in the Trash.
**Why it happens:** This is `wastebasket`'s own documented behavior (verified by reading all three platform `Trash` implementations this session), not a bug — a naive "on error, treat as if nothing happened" retry would attempt to trash `jsonPath` again.
**How to avoid:** This actually composes correctly with no extra code: every backend's `Trash` does `os.Stat` first and `continue`s past a path that no longer exists (`os.IsNotExist(err) { continue }` in `wastebasket_darwin.go`/`wastebasket_nix.go`; the Windows backend filters non-existent paths before the batched call too). So the UI-SPEC's "Try moving to Trash again" retry — which re-invokes with the same two paths — will correctly skip the already-trashed JSON and only re-attempt the HTML. **The plan does not need to build any "what already succeeded" tracking** — just re-call with the same path set on retry.
**Warning signs:** A retry implementation that tries to diff "what changed" instead of simply re-calling `Trash()` with the same arguments — unnecessary complexity `wastebasket`'s own idempotent-on-missing-file behavior already covers.

### Pitfall 4: `fsnotify` is non-recursive, and does not surface a "watched directory itself got deleted" event uniformly
**What goes wrong:** `fsnotify.go`'s own doc comment (read this session): *"All files in a directory are monitored, including new files that are created after the watcher is started. Subdirectories are not watched (i.e. it's non-recursive)."* This is a non-issue for StorCat's catalog directory (catalogs are flat `.json`/`.html` files, never nested) — but a naive implementation might try to recurse anyway.
Separately: *"A watch will be automatically removed if the watched path is deleted or renamed. The exception is the Windows backend, which doesn't remove the watcher on renames."* This means: if the user renames the catalog directory itself while StorCat has it open and is watching it, macOS/Linux will drop the watch cleanly (surfacing as silence — no more events, watcher still technically "open" but inert), while Windows will keep the (now-dangling) watch attached to the old handle.
**Why it happens:** Platform kernel primitives differ; `fsnotify` documents rather than papers over the difference.
**How to avoid:** Only ever `Add()` the single catalog directory, never its contents. For the directory-deleted/renamed case: since WATCH-01/02/03 give no explicit requirement about detecting the watched directory vanishing, the pragmatic behavior is: the watcher goes silent (no crash, no panic — `fsnotify`'s Events channel simply stops producing for that path), and the *next* explicit catalog-directory change (via Settings) tears down and rebuilds the watcher anyway, per the CONTEXT.md-locked lifecycle rule ("stop and restart on directory change"). Log a `WINDOWS.md` entry for the Windows non-removal-on-rename behavior, matching this project's established pattern of tracking unverified/divergent platform behavior rather than asserting parity.
**Warning signs:** Code that calls `watcher.Add()` on every subdirectory found via a walk — unnecessary and wrong for this phase's flat-directory case.

### Pitfall 5: Editor/tool atomic-save patterns show up as Rename+Create, not Write
**What goes wrong:** `fsnotify`'s own doc comment (read this session): *"many programs (especially editors) update files atomically: it will write to a temporary file which is then moved to destination... Watch the parent directory and use Event.Name to filter."* StorCat's own `WriteFileAtomic` does exactly this pattern (temp file + `os.Rename`) — meaning **StorCat's own create/rename/duplicate/delete operations will themselves generate `Rename`+`Create` event pairs on the watched directory**, not a single `Write`.
**Why it happens:** This is the same crash-safety technique ACT-09 requires, so the watcher will observe its own app's writes as multi-event bursts, indistinguishable at the event level from an external tool's atomic save.
**How to avoid:** This is precisely why `27-CONTEXT.md` locks a debounce (~300ms trailing) before emitting `catalogs:changed` — it's not just about absorbing a large external copy's event storm, it's also what prevents the app's own writes from needing to be filtered out by origin. Since the rail refresh is a full, idempotent re-list regardless of *why* it fired, there is no correctness requirement to distinguish "this app wrote it" from "something else wrote it" — only a performance one (avoid re-listing on every single event).
**Warning signs:** An attempt to filter out "our own writes" by tracking in-flight paths — unnecessary complexity; the debounce + idempotent re-list already handles this correctly and more simply.

### Pitfall 6: Event bursts can overflow the Events channel under sustained load
**What goes wrong:** `fsnotify.ErrEventOverflow` (documented in `fsnotify.go`, read this session): *"reported from the Errors channel when there are too many events... inotify: IN_Q_OVERFLOW... windows: buffer size too small."* A large external copy into the catalog directory (many files landing in a burst) could, in principle, overflow the OS-level event queue before this app's own 300ms debounce even gets a chance to coalesce.
**Why it happens:** The kernel-side event queue (inotify's `fs.inotify.max_queued_events`, Windows' buffer) is bounded independent of anything the debounce logic does downstream.
**How to avoid:** The watcher's Errors channel must be drained in the same select loop as Events (never left unread — an unread Errors channel can itself cause the watcher to stall), and an overflow error should trigger an immediate `catalogs:changed` emission (treat "we might have missed events" the same as "something changed" — the idempotent re-list is the correct recovery, there is no delta to reconcile). This is a low-probability case for StorCat's use pattern (catalog directories hold `.json`/`.html` pairs, not thousands of rapidly-changing files) but costs nothing to handle correctly.
**Warning signs:** A watcher goroutine that only selects on `w.Events`, never `w.Errors`.

### Pitfall 7: `File.Sync()` alone does not make a *new* file's existence durable — the parent directory needs its own fsync
**What goes wrong:** `27-CONTEXT.md` locks `File.Sync()` before close+rename on `WriteFileAtomic`. This makes the *temp file's contents* durable before the rename. But on POSIX filesystems, the directory entry created by `os.Rename` (the name-to-inode mapping) is a separate write that the filesystem may not have flushed to disk yet — a power loss immediately after a successful `os.Rename` can, on some filesystems/mount options, leave the destination path missing entirely even though the temp file's bytes were durable.
**Why it happens:** This is standard POSIX filesystem behavior, not a Go-specific gap — confirmed by web research this session: *"To achieve durability, you need to call fsync() twice: once on the file data and once on its parent directory... a power loss after file creation or rename can leave the filesystem in a state where the file does not exist, even though sync_all() was called on the file itself."* [CITED: general POSIX/filesystem durability guidance, corroborated by multiple sources including PostgreSQL's own durable-rename implementation notes, via WebSearch this session — treat as `MEDIUM` confidence, standard practice rather than a single authoritative spec citation]
**How to avoid:** After `os.Rename(tmpPath, path)` succeeds, open the *parent directory* (`os.Open(filepath.Dir(path))`) and call `.Sync()` on that directory handle, then close it. This works on Linux and macOS (both support directory `fsync`). **Windows does not support `fsync`-ing a directory handle the same way** — `os.Open` on a directory followed by `.Sync()` returns an error on Windows (directories aren't opened as syncable file handles via the Win32 API this way). The implementation should attempt the parent-directory sync and treat a platform "not supported"-shaped error as non-fatal (log/ignore) rather than failing the whole write — the file's own `File.Sync()` still gives the primary crash-safety guarantee on Windows, where `MoveFileEx`'s own semantics are different from POSIX rename durability.
**Warning signs:** A `WriteFileAtomic` implementation that assumes parent-directory fsync behaves identically across all three platforms — it does not, and this is exactly the kind of platform divergence `WINDOWS.md` exists to track. This is `27-CONTEXT.md`'s "real subtlety worth getting right" — flagged as Claude's Discretion since CONTEXT.md's locked decision only names `File.Sync()`, not the parent-directory fsync; the plan should decide explicitly whether to add it (recommended) rather than silently including or omitting it.

### Pitfall 8: A SIGKILL-mid-write test needs a real subprocess, not an in-process kill
**What goes wrong:** `t.Fatal`/`panic`/`os.Exit` from inside the same test process cannot simulate a SIGKILL, because Go's own deferred cleanup and the OS's own buffered-write flushing behave differently across a graceful-vs-violent process end. `WINDOWS.md` entry #6 exists precisely because "timing a SIGKILL inside `WriteFileAtomic`'s few-ms temp-then-rename window is not reliably schedulable" was the blocker in Phase 25.
**Why it happens:** The write-then-rename window really is only a few milliseconds for a small catalog JSON — too narrow to reliably interrupt from the *same* process (a `time.Sleep` + signal-to-self doesn't reproduce what an OS scheduler preemption + SIGKILL actually does to in-flight syscalls).
**How to avoid:** Build a **separate, tiny standalone Go program** (a new `internal/catalog/testdata/killtarget/main.go` or similar, built via `go build` in a `TestMain` or a `t.Helper()`, matching the pattern Go's own `os/exec` test suite uses — `exec.Command(testenv.Executable(t), ...)` dispatching to named subcommands) that: (1) accepts a destination path and payload size as args, (2) calls the *exact* `WriteFileAtomic` sequence with an injected artificial delay between `tmp.Write()` and `tmp.Close()` (e.g. write in two chunks with a `time.Sleep(50 * time.Millisecond)` between them, widening the real-world few-ms window to something reliably interruptible) — this delay must be a test-only build variant, not added to production `WriteFileAtomic`. The **parent test** then: launches the subprocess, polls the destination directory (`os.ReadDir` glob for `storcat-*.tmp`, matching `WriteFileAtomic`'s own temp-file naming pattern) until the temp file appears, immediately calls `cmd.Process.Kill()` (sends `SIGKILL` on Unix — this is the actual mechanism, not a graceful signal), then asserts: the original destination file (if one existed) is byte-identical to its pre-write content, and no partial/truncated file exists at the destination path. Repeat across many iterations (varying the injected delay slightly, or killing at multiple polled checkpoints) for statistical confidence, since a single kill timing proves only one point in the window.
**Warning signs:** A test that calls `WriteFileAtomic` and `os.Exit(1)` in the same goroutine — that's a clean-ish exit path (deferred closes still may or may not run depending on where interrupted), not a SIGKILL. [ASSUMED: this exact test-harness design is this session's own synthesis from Go's `os/exec` test-suite idiom (`helperCommand`/subprocess dispatch pattern, read from `golang.org/x/... /src/os/exec/exec_test.go` this session) applied to this specific crash-safety problem — it has not been run or verified working in this repo. The plan should treat the harness shape as a starting point, not a proven recipe.]

### Pitfall 9: The watching status-bar segment is optimistic, not a confirmed-alive signal — this is deliberate, not a gap
**What goes wrong (if "fixed"):** A tempting "improvement" is to have the watcher report back a `watching:started`/`watching:failed` event so the status bar only shows `● watching` once the `fsnotify.Add()` call actually succeeds.
**Why it's not needed here:** `27-UI-SPEC.md` (already checker-approved, 6/6 pillars) explicitly resolves this: *"This segment's visibility is optimistic, tied directly to the setting rather than a confirmed running watcher... A watcher that fails to start... degrades silently under the same 'non-critical background state, no toast, no error surface' precedent."* Building a watcher-health channel would be scope creep beyond what WATCH-01 requires and beyond what the approved UI contract calls for.
**Warning signs:** A plan task that adds a new Wails event for watcher start/failure — not required by any locked decision or the UI-SPEC.

## Code Examples

### Wiring the fsnotify watcher lifecycle (constructor/start/stop shape)
```go
// Synthesized from fsnotify.go's documented API (NewWatcher/Add/Remove/Close,
// read directly from github.com/fsnotify/fsnotify/fsnotify.go this session)
// combined with this codebase's throttledProgress precedent (app.go:173-189).
// [ASSUMED: illustrative shape, not copied from an official fsnotify example --
// fsnotify ships no directory-watcher-with-debounce example of its own.]
package watch

import (
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

type Watcher struct {
	fsw     *fsnotify.Watcher
	mu      sync.Mutex
	timer   *time.Timer
	onChange func()
	done    chan struct{}
}

// New starts watching dir (non-recursively -- catalogs are flat files, this
// is sufficient) and calls onChange, debounced 300ms trailing, whenever a
// Create/Write/Remove/Rename event fires. onChange must not import the Wails
// runtime -- the caller (app.go) supplies a closure that does the
// runtime.EventsEmit call, matching this repo's a.ctx==nil-guarded pattern.
func New(dir string, onChange func()) (*Watcher, error) {
	fsw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	if err := fsw.Add(dir); err != nil {
		fsw.Close()
		return nil, err
	}
	w := &Watcher{fsw: fsw, onChange: onChange, done: make(chan struct{})}
	go w.loop()
	return w, nil
}

func (w *Watcher) loop() {
	for {
		select {
		case event, ok := <-w.fsw.Events:
			if !ok {
				return
			}
			if event.Has(fsnotify.Create) || event.Has(fsnotify.Write) ||
				event.Has(fsnotify.Remove) || event.Has(fsnotify.Rename) {
				w.debounce()
			}
		case _, ok := <-w.fsw.Errors:
			// ErrEventOverflow and similar: treat as "something changed,
			// we may have missed events" -- fire immediately rather than
			// silently dropping, since the recovery (idempotent re-list)
			// is the same either way.
			if !ok {
				return
			}
			w.mu.Lock()
			w.onChange()
			w.mu.Unlock()
		case <-w.done:
			return
		}
	}
}

func (w *Watcher) debounce() {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.timer != nil {
		w.timer.Stop()
	}
	w.timer = time.AfterFunc(300*time.Millisecond, w.onChange)
}

// Close genuinely releases the underlying OS watch (inotify fd / kqueue fd /
// Windows handle) -- WATCH-03 requires release, not merely ignoring events.
func (w *Watcher) Close() error {
	close(w.done)
	return w.fsw.Close()
}
```

### app.go wiring (only place that calls runtime.EventsEmit for this feature)
```go
// [ASSUMED: illustrative -- shows how the existing throttledProgress
// (app.go:173-189, read this session) pattern extends to the watcher.]
func (a *App) startWatcherIfEnabled() {
	cfg := a.configManager.Get()
	if cfg == nil || !cfg.WatchDirectory || cfg.CatalogDirectory == "" {
		return
	}
	w, err := watch.New(cfg.CatalogDirectory, func() {
		if a.ctx == nil {
			return
		}
		runtime.EventsEmit(a.ctx, "catalogs:changed")
	})
	if err != nil {
		return // silent degrade -- matches 27-UI-SPEC.md's E4 resolution
	}
	a.watcher = w
}

func (a *App) SetWatchDirectory(enabled bool) error {
	if a.watcher != nil {
		a.watcher.Close()
		a.watcher = nil
	}
	if err := a.configManager.SetWatchDirectory(enabled); err != nil {
		return err
	}
	if enabled {
		a.startWatcherIfEnabled()
	}
	return nil
}
```
Note: `main.go` currently registers only `OnStartup`, `OnDomReady`, `OnBeforeClose` [VERIFIED: `main.go`, read this session — `wails.Run(&options.App{... OnStartup: app.startup, OnDomReady: app.domReady, OnBeforeClose: app.beforeClose, ...})`]. `OnBeforeClose` can return `true` to *prevent* close (as it already does for an active scan), so it is not a reliable unconditional-cleanup hook. Wails v2 also offers a distinct `OnShutdown` hook that fires unconditionally after close is confirmed [CITED: Wails documentation, via WebSearch this session — "OnShutdown: Executed during application shutdown, allowing cleanup... called just before the application shuts down"]. **Recommendation for the plan:** add `OnShutdown: app.shutdown` to `main.go`'s `options.App`, and call `a.watcher.Close()` there, rather than only inside `beforeClose` — this guarantees the watcher is released on every quit path, not only the ones that pass through `beforeClose`'s specific branches. This is a discretion item (`27-CONTEXT.md` names `internal/watch`-vs-`app.go` layout as discretion but doesn't name this specific lifecycle-hook choice) worth calling out explicitly in the plan rather than assuming `beforeClose` alone is sufficient.

### Frontend: subscribing to `catalogs:changed` (mirrors the existing `scan:progress` pattern)
```typescript
// Source: frontend/src/components/workspace/CreateSlideOver.tsx:193-206
// (EventsOn('scan:progress', ...) with unsubscribe-on-cleanup), read this
// session -- the exact shape to replicate for catalogs:changed.
useEffect(() => {
  const unsubscribe = EventsOn('catalogs:changed', () => {
    if (state.catalogDir) {
      loadCatalogsForDirectory(state.catalogDir);
    }
  });
  return () => unsubscribe();
}, [state.catalogDir, loadCatalogsForDirectory]);
```
`loadCatalogsForDirectory` is `CatalogRail.tsx`'s existing function (lines 19-31, read this session) — the exact same one its own `state.catalogDir` effect (lines 50-53) already calls. No new read path.

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | Parent-directory `fsync` after `os.Rename` should be added to `WriteFileAtomic` alongside the CONTEXT.md-locked `File.Sync()` | Common Pitfalls #7 | If omitted, a narrow durability gap remains (directory-entry write not confirmed flushed) on Linux/macOS; if added incorrectly on Windows (which doesn't support directory-handle fsync the same way), could introduce a spurious error on every write. Needs an explicit plan decision, not silent inclusion/omission. |
| A2 | Adding `OnShutdown: app.shutdown` to `main.go` is the correct place to guarantee watcher release on every quit path | Code Examples, "app.go wiring" | If `beforeClose` alone is used instead, a quit path that bypasses `beforeClose`'s specific branches (e.g., OS-forced termination that Wails still intercepts, or an already-known-different code path) could leave the watcher un-released. This is architecture guidance based on Wails' documented (not source-verified) hook semantics — `[CITED]`, not `[VERIFIED]` against the actual Wails v2 source. |
| A3 | The SIGKILL-mid-write test harness (separate subprocess + injected delay + `cmd.Process.Kill()` + polled temp-file detection) is a workable design for closing `WINDOWS.md` entry #6 | Common Pitfalls #8, Code Examples | This design has not been implemented or run in this repo this session — it's synthesized from Go's own `os/exec` test-suite idiom applied to this specific problem. If the injected-delay approach proves unreliable in practice (e.g., OS write buffering makes the temp file appear "complete" faster than expected), the plan may need a different synchronization mechanism (e.g., an explicit named pipe/signal file the helper writes after `tmp.Write()` but before `tmp.Close()`, which the parent watches instead of polling for the temp file's mere existence). |
| A4 | `wastebasket`'s AppleScript-interpolation risk (Pitfall 1) is best handled by relying on the existing `ContainsPath` containment gate rather than attempting to further sanitize the path string before it reaches `wastebasket.Trash()` | Common Pitfalls #1 | If a future feature ever lets a user type a raw filename that reaches the Trash binding without going through `catalogDir`-scoped file listing, containment alone would not fully close the AppleScript-string-interpolation surface (it restricts *where* the file lives, not what characters its name/path may contain). Out of scope for this phase's actual UI (paths always come from `BrowseCatalogs`'s own directory listing), but worth the plan recording explicitly as an accepted, scope-bounded risk. |

**If this table is empty:** N/A — see rows above.

## Open Questions (RESOLVED)

Both were resolved before planning; the resolutions are locked in `27-CONTEXT.md`'s "Post-research
resolutions" block and implemented in plans 27-01 and 27-03.

1. **RESOLVED** — resolution: **`omitempty`** (the researcher's own recommendation), preserving byte-parity
   for catalogs with no title override, which COMPAT-02 depends on. Implemented in plan 27-01.
   **Does the exact JSON `title` field position/name matter for COMPAT-02 (byte-for-byte JSON shape parity with v2.3.0)?**
   - What we know: `pkg/models/catalog.go`'s `CatalogItem` struct (read this session) already has a precedent for additive, `omitempty` fields (`Unreadable`, `ReadError`) that don't disturb the byte-for-byte guarantee for catalogs that don't use them — the doc comment states *"No reader in this repository rejects unrecognized JSON keys, so these two keys are silently ignored by every catalog reader that doesn't yet know about them."*
   - What's unclear: Whether `title` should be `omitempty` (so a catalog with no explicit rename stays byte-identical to today's output) or always-present (simpler mental model, matches `JsonPath`'s own non-`omitempty` convention noted in `CreateCatalogResult`). `27-CONTEXT.md` leaves the field name/position to Claude's discretion but doesn't address `omitempty`.
   - Recommendation: `omitempty`, consistent with `Unreadable`/`ReadError`'s existing precedent for "this key only appears when the feature that needs it was actually used" — a freshly-created catalog (never renamed) should stay byte-for-byte identical to v2.3.0's output, preserving COMPAT-02.

2. **RESOLVED** — resolution: **inherit the source title verbatim** (the researcher's own recommendation).
   Only the filename root gets the `-copy` suffix; ACT-03 speaks to the filename, not the title, and a user
   who wants a different title can rename afterwards. Implemented in plan 27-03.
   **Should `DuplicateCatalog`'s new JSON also get a fresh `title` set to something like "`<original> copy`", or should it inherit the original title verbatim?**
   - What we know: ACT-03 only specifies the *filename* gets suffixed (`-copy`, `-copy-2`); it says nothing about the title.
   - What's unclear: If title inherits verbatim, the rail will show two catalogs with the identical displayed title but different filenames — which is arguably correct (a duplicate genuinely is an identical copy) but could confuse a user scanning the rail.
   - Recommendation: Inherit the title verbatim (true duplicate semantics) — this matches the "duplicate" mental model better than inventing new copy-suffixed title text the user didn't ask for, and keeps this phase's scope minimal. Flag as Claude's Discretion for the plan to confirm.

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| Go toolchain ≥1.23 | Both new dependencies' `go.mod` minimums | ✓ | go1.26.6 (darwin/arm64) [VERIFIED: `go version`, this session] | — |
| `wastebasket/v2` macOS backend (`osascript`/Finder) | ACT-04/05 on this dev machine | ✓ (macOS is the dev platform) | — | — |
| `wastebasket/v2` Windows backend (`shell32.dll` `SHFileOperationW`) | ACT-04/05 on Windows | ✗ (no Windows machine/VM available this session, matching the project's existing `WINDOWS.md` pattern) | — | Compiles under `GOOS=windows` cross-build (not verified this session, but the code path is straightforward `syscall`/`golang.org/x/sys/windows` — same class of gap as existing `WINDOWS.md` entries #4/#5) |
| `wastebasket/v2` Linux/BSD backend (FreeDesktop trash spec) | ACT-04/05 on Linux | ✗ (no Linux machine/VM available this session) | — | Same as above — pure-Go, cross-builds cleanly, runtime behavior unverified |
| `fsnotify` kqueue backend (macOS) | WATCH-01/02/03 on this dev machine | ✓ | — | — |
| `fsnotify` inotify backend (Linux) | WATCH-01/02/03 on Linux | ✗ (no Linux machine/VM) | — | Same WINDOWS.md-pattern gap |
| `fsnotify` ReadDirectoryChangesW backend (Windows) | WATCH-01/02/03 on Windows | ✗ (no Windows machine/VM) | — | Same WINDOWS.md-pattern gap |

**Missing dependencies with no fallback:** none — every gap above is a *platform-runtime-verification* gap (code compiles for the target GOOS, behavior unverified), the same class this project already tracks in `WINDOWS.md` rather than a blocking missing dependency.

**Missing dependencies with fallback:** n/a — see above; the "fallback" in every row is simply "log a new `WINDOWS.md` entry and sweep before v3.0.0 ships," consistent with entries #1, #2, #4, #5, #7 already in the ledger.

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go stdlib `testing` (table-driven, `*_test.go` beside source) [VERIFIED: `internal/catalog/atomicwrite_test.go`, read in full this session] |
| Config file | none — no test framework config needed for stdlib `testing` |
| Quick run command | `go test ./internal/catalog/... ./internal/osutil/... ./internal/watch/... ./pkg/models/...` |
| Full suite command | `go test ./...` |

No frontend test framework exists in this project by design (TEST-01 deferred, confirmed in `STATE.md`). Frontend proof for this phase is `npx tsc --noEmit` + `npm run build` + live `dev-browser` verification against `:34115`, per this project's established convention.

### Phase Requirements → Test Map

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| ACT-01 | Menu opens/closes, keyboard nav, focus restore | manual/dev-browser (no frontend test framework) | live verification via dev-browser against `:34115` | n/a — TEST-01 deferred |
| ACT-02 | Rename writes JSON title, rewrites both HTML title occurrences, handles no-`.html` case, handles `&`/special chars round-trip | unit | `go test ./internal/catalog/... -run TestRenameCatalog -v` | ❌ Wave 0 — new test file |
| ACT-02 | `BrowseCatalogs` title-read fix (`html.UnescapeString`) | unit | `go test ./internal/search/... -run TestBrowseCatalogs_UnescapesTitle -v` | ❌ Wave 0 — new test |
| ACT-03 | Duplicate suffixes filename correctly across `-copy`/`-copy-2`/`-copy-3` collisions | unit | `go test ./internal/catalog/... -run TestDuplicateCatalog -v` | ❌ Wave 0 — new test file |
| ACT-04/05 | Trash binding never falls back to permanent delete on error; containment-gated | unit (mock/fake trash function via interface, not real `wastebasket` call in CI) | `go test ./internal/osutil/... -run TestTrash -v` | ❌ Wave 0 — new test file |
| ACT-09 | `WriteFileAtomic` calls `File.Sync()` before close+rename | unit | `go test ./internal/catalog/... -run TestWriteFileAtomic_Syncs -v` | ❌ Wave 0 — extends `atomicwrite_test.go` |
| ACT-09 | Real SIGKILL-mid-write leaves destination uncorrupted | manual/integration (subprocess-based, see Common Pitfalls #8) | `go test ./internal/catalog/... -run TestWriteFileAtomic_SurvivesKill -v -timeout 60s` | ❌ Wave 0 — new test file + helper subprocess |
| WATCH-01 | Status bar shows watching segment when setting+dir both set | manual/dev-browser | live verification | n/a — TEST-01 deferred |
| WATCH-02 | External file add/remove/modify triggers debounced `catalogs:changed` → rail refresh | unit (fake fsnotify events via a test double) + manual live verification | `go test ./internal/watch/... -v` | ❌ Wave 0 — new package + tests |
| WATCH-03 | `SetWatchDirectory(false)` and app quit both call `Watcher.Close()` | unit | `go test ./internal/watch/... -run TestWatcher_Close -v` | ❌ Wave 0 — new test |

### Sampling Rate
- **Per task commit:** `go test ./internal/catalog/... ./internal/osutil/... ./internal/watch/... ./internal/search/... ./pkg/models/...` (scoped to touched packages)
- **Per wave merge:** `go test ./...` + `npx tsc --noEmit` + `npm run build`
- **Phase gate:** Full suite green before `/gsd-verify-work`; live dev-browser pass for ACT-01/WATCH-01 (no frontend test framework to automate these)

### Wave 0 Gaps
- [ ] `internal/catalog/rename_test.go` (or similar) — covers ACT-02
- [ ] `internal/catalog/duplicate_test.go` — covers ACT-03
- [ ] `internal/osutil/trash_test.go` — covers ACT-04/ACT-05 (mock the trash call behind a small interface so CI doesn't actually touch a real OS Trash)
- [ ] `internal/watch/watcher_test.go` — covers WATCH-02/WATCH-03 (fsnotify itself is hard to unit-test directly against a real filesystem in CI sandboxes; consider testing the debounce logic in isolation with a fake event source, and reserve real `fsnotify` behavior for manual/live verification)
- [ ] `internal/catalog/atomicwrite_sigkill_test.go` + a small standalone helper `main.go` under `internal/catalog/testdata/` (or similar) — covers ACT-09's crash-safety claim with an actual kill, closing `WINDOWS.md` #6
- [ ] `internal/search/service_test.go` gains a title-unescape case if no such test file exists yet — verify before Wave 0 (not confirmed present or absent this session; grep at plan time)

## Security Domain

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V2 Authentication | no | Single-user desktop app, no auth surface added this phase |
| V3 Session Management | no | No session concept in this app |
| V4 Access Control | yes | Every new path-taking binding (`RenameCatalog`, `DuplicateCatalog`, `DeleteCatalog`) must reuse `osutil.ContainsPath(catalogDir, resolved)` exactly as `GetCatalogHtmlPath`/`OpenExternal`/`RevealInFileManager` already do — this is the project's established, working control for "renderer JS can only touch files inside the configured catalog directory" |
| V5 Input Validation | yes | Rename's title field is free-form user text; must round-trip through `html.EscapeString`/`html.UnescapeString` correctly (already the write-side pattern; this phase adds the missing read-side half) |
| V6 Cryptography | no | No cryptographic operations in this phase |
| V12 Files and Resources | yes | The Trash binding is this phase's highest-risk file-operation surface — see Threat Patterns below |

### Known Threat Patterns for this stack

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| Path traversal / containment escape via a renderer-supplied catalog path reaching Rename/Duplicate/Delete | Tampering, Elevation of Privilege | `osutil.ContainsPath` gate before any filesystem operation, exactly matching the existing `RevealInFileManager`/`OpenExternal`/`GetCatalogHtmlPath` precedent |
| AppleScript string interpolation in `wastebasket`'s macOS Trash backend (`osascript -e 'tell app "Finder" to delete POSIX file "<path>"'`, only `"` escaped) | Tampering (upstream, third-party, not directly fixable) | Containment-gate every path *before* it reaches `wastebasket.Trash()`, so the interpolated string can only ever be a path the app already resolved as inside `catalogDir` — see Common Pitfalls #1 and Assumption A4 |
| Stored markup injection via catalog title reaching `<title>`/`<h1>` in generated HTML | Tampering (low severity — HTML opens in the OS default browser, not inside the app's own webview since `CatalogModal`'s `srcDoc` render was deleted in Phase 26) | `html.EscapeString` on write (already correct, verified this session), `html.UnescapeString` on the `BrowseCatalogs` read path (this phase's locked fix) |
| A malicious/hostile catalog title breaking the HTML file's structure if the surgical `<title>`/`<h1>` rewrite doesn't re-escape correctly | Tampering | The rewrite must call `html.EscapeString(newTitle)` on write, exactly matching the original generator's own escaping — do not write the raw user-typed string into the HTML file |
| `fsnotify` watch on a symlinked catalog directory could watch an unintended location if the symlink target changes after `Add()` | Tampering (low severity, low likelihood for this app's use case) | Not flagged as requiring new code this phase — `ContainsPath` already resolves through `filepath.EvalSymlinks` for the read/write paths that matter; the watcher itself only ever produces a signal to re-list, never a direct filesystem write, bounding the impact of a symlink surprise to "the rail refreshes for the wrong reason," not data corruption |

## Sources

### Primary (HIGH confidence — direct source reads + tool verification this session)
- `github.com/Bios-Marcel/wastebasket` — `wastebasket.go`, `wastebasket_darwin.go`, `wastebasket_windows.go`, `wastebasket_nix.go`, `go.mod`, `README.md` fetched and read in full via `raw.githubusercontent.com` this session
- `github.com/fsnotify/fsnotify` — `fsnotify.go`, `go.mod`, `CHANGELOG.md`, `README.md` fetched and read via `raw.githubusercontent.com` this session
- `go list -m -versions github.com/Bios-Marcel/wastebasket/v2` and `go list -m -versions github.com/fsnotify/fsnotify` — run against the live Go module proxy this session
- `go version` — run this session (go1.26.6 darwin/arm64)
- `api.github.com/repos/{Bios-Marcel/wastebasket,fsnotify/fsnotify}` and `pkg.go.dev` — fetched this session for repo age/star/importer signals
- This repository's own source, read in full this session: `internal/catalog/atomicwrite.go`, `internal/catalog/service.go` (lines 380-622), `pkg/models/catalog.go`, `internal/search/service.go` (lines 1-25, 150-269), `internal/osutil/reveal.go`, `frontend/src/hooks/useModalBehavior.ts`, `frontend/src/components/workspace/DetailsPanel.tsx` (lines 1-80), `frontend/src/components/workspace/StatusBar.tsx`, `frontend/src/components/workspace/CatalogRail.tsx` (lines 1-70), `app.go` (multiple ranges), `main.go`, `internal/config/config.go` (lines 1-90, 290-324), `frontend/src/settingsStore.ts`, `frontend/src/contexts/AppContext.tsx` (grep + targeted reads), `go.mod`

### Secondary (MEDIUM confidence — WebSearch/WebFetch corroborated by, but not directly quoting, an authoritative primary source)
- fsync/parent-directory-durability guidance (Common Pitfalls #7) — WebSearch this session, general POSIX filesystem durability practice corroborated across multiple sources (blog posts, PostgreSQL mailing list discussion), not a single canonical spec citation
- Wails v2 `OnShutdown` vs `OnBeforeClose` lifecycle semantics — WebSearch this session; not independently verified against the Wails v2 Go source in this session (the claim rests on Wails' own documentation as summarized by search, not a direct source read the way the two new dependencies were)

### Tertiary (LOW confidence — flagged for validation)
- The SIGKILL-mid-write test harness design (Common Pitfalls #8, Assumption A3) — this session's own synthesis, not found as a documented pattern anywhere; treat as a starting design, not a proven recipe

## Metadata

**Confidence breakdown:**
- Standard stack (both new dependencies): HIGH — read actual source code + verified versions via the Go module proxy, not just documentation
- Architecture (containment pattern, watcher lifecycle, dialog shell reuse): HIGH — every pattern cited was read directly from this repo's own working code this session
- Pitfalls (AppleScript escaping, dual HTML title occurrence, fsnotify platform divergence, partial-Trash-failure semantics): HIGH for the wastebasket/fsnotify-sourced items (direct source reads); MEDIUM for the fsync/parent-directory-durability item (corroborated web research, not a single canonical citation); LOW for the SIGKILL test harness design (synthesized, unverified in this repo)

**Research date:** 2026-08-15
**Valid until:** 30 days (stable domain — both dependencies are mature, low-churn libraries; the codebase-specific findings are tied to this session's exact file contents and should be re-verified if significant refactoring lands before planning executes)
