# Phase 25: Create Slide-over + Progress/Cancellation/Partial-Catalog - Research

**Researched:** 2026-08-14
**Domain:** Go/Wails backend rework (context cancellation, error-tolerant walk, atomic writes, per-OS volume enumeration) + React create-flow UI
**Confidence:** HIGH on all codebase-grounded claims (read directly, cited file:line); MEDIUM on Wails-idiom claims cross-checked against the vendored v2.10.2 source; LOW/explicitly-flagged on anything requiring a Windows or Linux machine, neither of which was available this session

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

**Backend Surface & CLI Compatibility**
- Cancellation is threaded as `ctx context.Context` through `traverseDirectory`, reached via a new `CreateCatalogWithContext(ctx, …, opts)` method. The existing `CreateCatalog(title, directoryPath, outputRoot, copyToDirectory, onProgress)` remains as a thin wrapper calling it with `context.Background()` and default options. `cli/create.go` is left **literally untouched**.
- `ProgressCallback` changes to carry a struct (path, files seen, bytes seen, read-error count — exact shape at Claude's discretion) instead of `func(path string)`.
- Percentage/ETA computed against the selected volume's total size (Statfs). For a plain folder (no volume total), run a fast count-only pre-pass and show indeterminate progress until it completes. A two-pass count for every scan was rejected (doubles I/O at 40k-node scale). Dropping percentage was rejected (design rule: "No spinners — progress is always a real number").
- Progress reaches the frontend via Wails `runtime.EventsEmit`, emitted from `app.go` **only** — never from `internal/catalog` (COMPAT-04).

**Cancellation, Partial Writes & the Unreadable-Subtree Marker**
- A user cancel (CRT-09) and a window close mid-scan (CRT-13) write **nothing**. Only the volume-disappeared error state (CRT-10) offers a partial write (CRT-11).
- A partial catalog marks the unreadable subtree on the affected directory nodes only — an explicit, documented divergence from COMPAT-02 **scoped to partial catalogs**. A complete catalog remains byte-for-byte the v2.3.0 shape. Exact field name/shape deferred to this research pass. Silently omitting unreadable subtrees rejected (invisible data loss); a sidecar file rejected (splits the artifact from what it describes).
- Crash-safe writes (temp file, then `rename`) are implemented in this phase, reused by Phase 27's ACT-09.
- Forced close is wired through the existing `beforeClose` hook (`app.go:192`) — no renderer-side `beforeunload`.

**Volume Detection & the Slide-over**
- Volume enumeration is stdlib-only, in per-OS build-tagged files: macOS `/Volumes`, Linux `/media` + `/mnt` (cross-checked against `/proc/mounts`), Windows drive letters. No new Go dependency this phase.
- Size/free space via `syscall.Statfs` (unix) and `GetDiskFreeSpaceEx` (Windows), in the same build-tagged files. Windows path cannot be runtime-verified here — pre-log to `.planning/WINDOWS.md`.
- A volume card's `read errors` status comes from probing the volume root only (`os.ReadDir` on the mount point). Deeper failures surface during the scan (CRT-10).
- The slide-over's 340ms-in/260ms-out animation extends Phase 24's `useModalBehavior` hook rather than adding bespoke animation state.

### Claude's Discretion
- The exact `ProgressUpdate` struct shape, field names, and emit cadence.
- The new method's name and option-struct shape (`CreateCatalogWithContext` indicative, not binding).
- Volume-card component decomposition and the "WILL WRITE" preview's exact rendering.
- The newest-first log's retention limit and whether it is virtualized.
- How "Run in background" hands off state between the slide-over and the status bar.
- Temp-file naming and placement for the atomic write.

### Deferred Ideas (OUT OF SCOPE)
- Re-scan and diff of an existing catalog — Phase 28 (ACT-06/07/08).
- Rename, duplicate, delete-to-Trash — Phase 27 (reuses this phase's atomic-write primitive, not built here).
- Watching the catalog directory for outside changes — Phase 27 (WATCH-01/02/03).
- CLI subcommands for new capabilities — GUI-only this milestone (FUT-03).
- Frontend unit tests — TEST-01, deferred at requirements definition.
- Resuming an interrupted scan — not a requirement; CRT-11 offers retry-from-scratch only.
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| CRT-01 | 560px slide-over, 340ms in / 260ms out, no early unmount | `useModalBehavior` reuse pattern confirmed (Architecture Patterns); Pitfalls 9-11 sourced from `.planning/research/PITFALLS.md` |
| CRT-02 | Detected volume cards: name, mount path, size, `mounted`/`read errors` | Volume Enumeration section — live-verified `/Volumes` behavior, Statfs field shapes per OS |
| CRT-03 | Choose any folder instead of a volume | `SelectDirectory` binding already exists (`app.go:219-224`), reused as-is |
| CRT-04 | Title/filename-root + live WILL WRITE preview | **Headline finding**: WILL WRITE destination is the app's configured catalog directory, not the scanned source — verified against the design handoff's own state logic |
| CRT-05 | Toggle write-HTML / copy-to-secondary / include-hidden | `Options` struct design (Code Examples); explicit pitfall on `WriteHTML` default |
| CRT-06 | Start scan via Create / ⌘↵ | Wails per-call-goroutine dispatch confirmed — safe to fire a long call from a click handler |
| CRT-07 | Live progress: %, files, bytes, ETA, path, newest-first log | `ProgressUpdate` struct proposal; EventsEmit throttling pattern; ETA smoothing |
| CRT-08 | "Run in background" + status-bar `● scanning <name> · N%` | Per-call-goroutine dispatch means the scan survives panel close; EventsOn/Off lifecycle notes |
| CRT-09 | Cancel actually stops the walk | `context.Context` threading design; App-held single cancel-func field (one-scan-at-a-time) |
| CRT-10 | Distinct error state when volume disappears mid-scan | **Headline finding**: terminal-vs-single-entry error classification heuristic |
| CRT-11 | Write partial catalog / retry / cancel from error state | Partial-tree retention design (App-held last-scan state) |
| CRT-12 | Done state lists every written file; open in workspace | Atomic-write helper (reused from `internal/config/counts_cache.go` precedent) |
| CRT-13 | Window close mid-scan cancels and writes nothing | `beforeClose` / `OnBeforeClose(prevent bool)` / `runtime.Quit` mechanism, verified signatures |
| COMPAT-02 | v3.0.0 catalogs byte-for-byte identical JSON shape to v2.3.0 | `omitempty` marker fields verified against Go's documented `encoding/json` semantics; byte-diff test recommended |
| COMPAT-03 | CLI subcommands behave exactly as v2.3.0 | Wrapper-preserves-behavior design traced through every changed parameter, including the `WriteHTML` default pitfall |
| COMPAT-04 | `internal/catalog` usable from CLI without a Wails runtime | `EventsEmit` confined to `app.go`; `internal/catalog` gains no Wails import |
</phase_requirements>

## Summary

This phase's two ROADMAP-flagged unknowns both resolved to concrete, code-grounded answers this session, and a third, unflagged architecture gap surfaced that materially changes the new method's signature.

**Forced-close/cancel "writes nothing"** is structurally easy to guarantee: `CreateCatalog` today builds the entire tree in memory via `traverseDirectory` *before* either `writeJSONFile` or `writeHTMLFile` is ever called `[VERIFIED: internal/catalog/service.go:28-45]`. As long as `CreateCatalogWithContext` checks `ctx.Err()` after the walk and returns immediately on cancellation — without ever reaching the write step — no partial JSON/HTML can land on disk from a cancel or a forced close. The harder problem is *timing* the close itself: Wails runs each bound-method call in its own goroutine `[VERIFIED: wails v2.10.2, internal/frontend/desktop/darwin/frontend.go:389]`, so `beforeClose` does not automatically block until an in-flight scan's goroutine notices cancellation. The safe pattern is `beforeClose` returning `prevent: true` `[VERIFIED: pkg/options/options.go:64]` on the first close request while a scan is active, cancelling it, waiting (bounded) for the goroutine to actually stop, then calling `runtime.Quit(ctx)` `[VERIFIED: pkg/runtime/runtime.go:64]` — this exact mechanism is my own design synthesis from verified primitives, not itself found in Wails docs, so treat it as a recommendation needing live verification during execution.

**The unreadable-subtree marker** can be two new `omitempty` fields on `CatalogItem` — `Unreadable bool` and `ReadError string` — which are absent entirely from any clean-scan JSON (guaranteeing COMPAT-02) and silently ignored by every existing reader, since no code path in this repo sets `json.Decoder.DisallowUnknownFields` `[VERIFIED: grep, zero results]`. The harder problem, again, is not the schema but the *walk logic*: today's `traverseDirectory` already swallows every read failure completely silently — a directory whose `os.ReadDir` fails returns as an empty, error-free node `[VERIFIED: internal/catalog/service.go:108-117, quoted below]`, and a child whose `os.Stat` fails is dropped from `contents` with a bare `continue` `[VERIFIED: internal/catalog/service.go:140-144]`. Neither path can currently distinguish "one bad file" from "the volume is gone," which is exactly Pitfall 17 from the milestone's own prior research. This phase needs a real classification: probe whether the *scan root* is still stat-able whenever any read error occurs; if not, treat it as terminal (stop descending, preserve everything walked so far) rather than skip-and-continue.

**The unflagged finding**: the design handoff's own mocked state logic proves the WILL WRITE preview's primary destination is the app's *configured catalog directory*, not the scanned source volume — `willWrite: '~/dev/sd-catalogs/' + root + '.json' + …` `[VERIFIED: design_handoff_storcat_ui/designs/StorCat 1a Demo.dc.html:1148]`. Today's `CreateCatalog(title, directoryPath, outputRoot, copyToDirectory, onProgress)` writes `outputRoot+".json"` *inside* `directoryPath` (the thing being scanned) `[VERIFIED: internal/catalog/service.go:36]` — `directoryPath` serves double duty as both walk-source and write-destination. For the GUI flow this is wrong: cataloguing an SD card must not drop `sd48.json` onto the SD card itself, it must land in the user's persisted catalog directory so `BrowseCatalogs`/the rail can see it. `CreateCatalogWithContext` therefore needs a genuinely new parameter separating "what gets walked" from "where the primary JSON/HTML land," with the CLI wrapper defaulting the new parameter to `directoryPath` (byte-identical old behavior).

**Primary recommendation:** Add `CreateCatalogWithContext(ctx, title, sourcePath, outputDir, outputRoot, copyToDirectory string, opts Options, onProgress ProgressCallback)` to `internal/catalog`, where `CreateCatalog` wraps it with `outputDir = directoryPath` and `Options{WriteHTML: true}` (explicit, not the zero value). Classify walk errors by re-probing the scan root on any failure; mark only the directory node where a terminal failure was detected with `Unreadable`/`ReadError`, `omitempty`. Keep `EventsEmit`/throttling entirely in `app.go`, hold a single mutex-guarded `context.CancelFunc` on `App` (the product is one-scan-at-a-time by design, no map needed), and retain the last partial scan's in-memory tree on `App` until it is written, retried, or discarded, since the bound method that produced it has already returned.

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Directory walk / error classification | Backend (`internal/catalog`) | — | Owns the only tested, CLI-shared code path this phase touches |
| Progress throttling + event emission | Backend (`app.go`) | — | COMPAT-04 forbids Wails coupling inside `internal/catalog`; `app.go` is the sole adapter |
| Volume enumeration + size/free | Backend (new per-OS files) | — | Requires OS-specific syscall packages (`golang.org/x/sys/unix` vs `/windows`), not available to the frontend |
| Cancellation handle (CancelFunc) | Backend (`App` struct) | — | Wails cannot bind a `context.Context` parameter from JS; the handle must live server-side |
| Scan state machine (form→scanning→error→done) | Frontend (`CreateSlideOver`, lifted to `AppContext`) | Backend (last-partial-scan retention) | UI-SPEC requires background-reopen to resume mid-scan; state must survive slide-over unmount |
| Atomic write (temp+rename) | Backend (new shared helper) | — | Must be reused by Phase 27's rename/duplicate/delete; belongs beside the write functions it protects |
| WILL WRITE preview derivation | Frontend (pure render) | Backend (supplies `outputDir` = persisted catalog directory) | Frontend already holds `storcat-catalog-directory` in `localStorage`/`AppContext`; no new backend round trip needed |
| "Run in background" handoff | Frontend (`AppContext` + `StatusBar`) | — | The Go-side scan is unaffected by panel close (per-call goroutine dispatch); only UI state needs lifting |

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `github.com/wailsapp/wails/v2` | v2.10.2 (unchanged) | `EventsEmit`/`EventsOn`, `OnBeforeClose`, `runtime.Quit` | Already the project's runtime; every API used this phase verified against the vendored source `[VERIFIED: go.mod:10, github.com/wailsapp/wails/v2@v2.10.2]` |
| `golang.org/x/sys` | v0.30.0, promote `indirect`→direct in `go.mod` | `unix.Statfs`, `windows.GetDiskFreeSpaceEx` | Already resolved transitively via Wails; zero new dependency cost `[VERIFIED: go.mod:45, `golang.org/x/sys v0.30.0 // indirect`]` |
| Go stdlib `context` | go1.23 (project's pinned version) | Cancellation threading | First use in this codebase — no other file currently imports `context` outside `app.go`/Wails-generated code `[VERIFIED: grep for "context.Context" across `.go` files, only `app.go` and vendored Wails code]` |
| Go stdlib `os`/`path/filepath` | go1.23 | Atomic temp+rename write | Exact pattern already proven in this repo, see Code Examples |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| Go stdlib `os/exec` | go1.23 | None needed for volume enumeration on macOS/Linux (pure `os.ReadDir` + `unix.Statfs`) | Not needed this phase — `internal/osutil/reveal.go` already shows the project's `exec.Command` convention if a future phase needs it |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Hand-rolled per-OS volume enumeration | `github.com/shirou/gopsutil/v4` (already researched in `.planning/research/SUMMARY.md`, MEDIUM confidence) | Rejected by `25-CONTEXT.md` — no new dependency budget this phase; stdlib + `x/sys` covers everything needed for `/Volumes`, `/media`+`/mnt`, drive letters, `Statfs`/`GetDiskFreeSpaceEx` |
| `filepath.WalkDir`-based rewrite of the walk | Keep the existing hand-rolled recursive `traverseDirectory` | `WalkDir`'s callback-based skip/stop semantics don't map cleanly onto "preserve partial results but stop descending" without extra state; the existing recursive function already returns a tree node per call, which is what the terminal-classification design needs |

**Installation:**
```bash
go get golang.org/x/sys@v0.30.0   # promotes an already-resolved indirect dependency to direct in go.mod
```
No `npm install` — `25-CONTEXT.md` locks zero new frontend packages this phase; the toggle switch, round badges, and progress bar are hand-built CSS/HTML per `25-UI-SPEC.md`'s Registry Safety section.

**Version verification:** `golang.org/x/sys@v0.30.0` is already pinned as an indirect dependency in `go.mod` `[VERIFIED: go.mod:45]`; promoting it to direct requires no version change, only removing the `// indirect` comment (or running `go mod tidy` after the new import appears). Wails `v2.10.2` is unchanged.

## Package Legitimacy Audit

**No new external packages this phase.** `25-CONTEXT.md` explicitly locks "No new Go dependency — the milestone's one-new-dependency budget is already committed to `Bios-Marcel/wastebasket/v2` in Phase 27" and the frontend "adds zero new packages." `golang.org/x/sys` is promoted from indirect to direct — it is already present in `go.sum`, already vetted by the milestone's own prior research (`.planning/research/SUMMARY.md`), and is a first-party Go team package, so the legitimacy gate does not apply to a version already resolved and verified in this repository's dependency graph.

| Package | Registry | Age | Downloads | Source Repo | Verdict | Disposition |
|---------|----------|-----|-----------|-------------|---------|-------------|
| `golang.org/x/sys` | Go module proxy | Go team package, maintained since 2009 | N/A (stdlib-adjacent) | `github.com/golang/sys` | OK | Approved — already in `go.sum` as indirect, promoted to direct only |

**Packages removed due to [SLOP] verdict:** none.
**Packages flagged as suspicious [SUS]:** none.

## Architecture Patterns

### System Architecture Diagram

```
Frontend (React)                    Wails Bridge                    Backend (Go)
─────────────────                   ────────────                    ─────────────

CreateSlideOver                                                     App struct
  │ user picks volume/folder                                          │ mu sync.Mutex
  │ user sets title/root/toggles                                      │ activeScanCancel context.CancelFunc
  │ clicks "Create catalog" ────────► StartScan(source, outputDir, ───► a.StartScan(...)
  │                                    opts) [one bound call,            │ ctx, cancel := context.WithCancel(context.Background())
  │                                    returns a Promise]                │ a.activeScanCancel = cancel  (mutex-guarded)
  │                                                                      │ go func() {  ← per-call goroutine (Wails default)
  │ ◄── EventsOn("scan:progress") ◄── runtime.EventsEmit(a.ctx, ...) ◄──┤    catalogService.CreateCatalogWithContext(ctx, ...)
  │     live % / files / bytes /       [throttled ~200ms in app.go,     │      internal/catalog: traverseDirectory(ctx, walkState)
  │     path / log line                 never inside internal/catalog]  │        - checks ctx.Err() at each directory boundary
  │                                                                      │        - classifies read errors: single-entry vs terminal
  │ clicks Escape/×/scrim                                               │        - on terminal: stop descending, keep partial tree
  │ (scanning state) ──────────────► CancelScan() ─────────────────────►│    a.activeScanCancel()  (user-cancel path)
  │                                                                      │    on ctx.Err(): return, write NOTHING
  │                                                                      │    on terminal volume-vanished: retain tree on App,
  │                                                                      │      return distinguishable error (no write)
  │ ◄── Promise resolves/rejects ◄──────────────────────────────────────┤
  │                                                                      │
  │ (error state) clicks                                                │
  │ "Write partial catalog" ───────► WritePartialCatalog() ────────────►│ a.lastPartialScan → atomic temp+rename write
  │                                                                      │   internal/catalog: writeJSONFile/writeHTMLFile
  │                                                                      │     (both routed through the new atomic helper)
  │                                                                      │
  window close (OS event) ─────────────────────────────────────────────►│ beforeClose(ctx) bool
                                                                          │   if activeScanCancel != nil: cancel(), wait (bounded),
                                                                          │   then runtime.Quit(ctx)
```

### Recommended Project Structure
```
internal/
├── catalog/
│   ├── service.go          # CreateCatalog (wrapper) + CreateCatalogWithContext + traverseDirectory (extended)
│   ├── options.go          # new: Options{WriteHTML, IncludeHidden bool}
│   ├── errors.go           # new: ErrScanCancelled, ErrVolumeVanished sentinel errors, ReadErrorEntry
│   └── atomicwrite.go      # new: writeFileAtomic(dir, name string, data []byte) error — shared by Phase 27
├── volumes/                # new package
│   ├── volumes.go          # exported Volume{Name, MountPath, TotalBytes, FreeBytes, Readable bool}, ListVolumes()
│   ├── volumes_darwin.go   # //go:build darwin — /Volumes enumeration + unix.Statfs
│   ├── volumes_linux.go    # //go:build linux — /media+/mnt + /proc/mounts cross-check + unix.Statfs
│   └── volumes_windows.go  # //go:build windows — drive letters + windows.GetDiskFreeSpaceEx
app.go                      # new bound methods: StartScan, CancelScan, WritePartialCatalog, ListVolumes
                             # beforeClose extended to cancel in-flight scans
```

### Pattern 1: Wails cancellation handle lives server-side, never crosses the bridge
**What:** A Wails-bound method cannot take a `context.Context` parameter the frontend supplies — the binding layer's argument parser has no special case for it `[VERIFIED: internal/frontend/dispatcher/calls.go, and grep for "context.Context" across internal/binding/*.go returned zero matches]`. Cancellation must be a second, separate bound call (`CancelScan()`) that looks up a `context.CancelFunc` the *first* call already stored.
**When to use:** Any long-running bound method that needs mid-flight cancellation.
**Example:**
```go
// Source: synthesized from verified Wails v2.10.2 dispatch behavior
// (internal/frontend/desktop/darwin/frontend.go:389 — each bound call
// runs in its own goroutine, so a concurrent CancelScan() call is never
// blocked behind an in-flight StartScan()).
type App struct {
    ctx               context.Context // Wails runtime context, for EventsEmit only
    mu                sync.Mutex
    activeScanCancel  context.CancelFunc // nil when no scan is running
    lastPartialScan   *partialScanResult // retained after a volume-vanished stop
}

func (a *App) StartScan(sourcePath, outputDir string, opts ScanOptions) error {
    ctx, cancel := context.WithCancel(context.Background())
    a.mu.Lock()
    a.activeScanCancel = cancel
    a.mu.Unlock()
    defer func() {
        a.mu.Lock()
        a.activeScanCancel = nil
        a.mu.Unlock()
    }()

    result, err := a.catalogService.CreateCatalogWithContext(ctx, ..., onProgressThrottled)
    // ... classify err: cancelled (write nothing) vs volume-vanished (retain
    // partial tree on a.lastPartialScan, return a distinguishable error) vs
    // success (already written).
    return err
}

func (a *App) CancelScan() {
    a.mu.Lock()
    defer a.mu.Unlock()
    if a.activeScanCancel != nil {
        a.activeScanCancel()
    }
}
```
A map keyed by scan-id is *not* needed: the UI-SPEC's Background Handoff Contract disables every "+New" entry point while a scan is active ("one scan at a time... 'Run in background' already covers the real workflow" — `REQUIREMENTS.md`'s own Out-of-Scope table), so a single field suffices.

### Pattern 2: Build-tagged per-OS files are required here, unlike the existing `reveal.go` precedent
**What:** `internal/osutil/reveal.go` deliberately avoided Go build tags, dispatching on `runtime.GOOS` as a parameter within one file, specifically because it only calls `exec.Command` with OS-specific argv shapes — no OS-specific *package* is imported `[VERIFIED: internal/osutil/reveal.go:45-65, comment: "That is a deliberate deviation from 23-RESEARCH.md's three-build-tagged-files sketch... all three shapes are exercised by the same test binary"]`. Volume enumeration cannot follow that precedent: `golang.org/x/sys/windows` is only buildable under `GOOS=windows` (its own source files use the Go toolchain's automatic `_windows.go` filename suffix constraint `[VERIFIED: golang.org/x/sys@v0.30.0/windows/syscall_windows.go header]`), and `golang.org/x/sys/unix.Statfs`'s `Statfs_t` struct has genuinely different field types per OS — darwin's `Bsize` is `uint32`, Linux amd64's `Bsize` is `int64` `[VERIFIED: golang.org/x/sys@v0.30.0/unix/ztypes_darwin_arm64.go:87 vs ztypes_linux_amd64.go:444]`. A single non-build-tagged file cannot reference `windows.GetDiskFreeSpaceEx` on a non-Windows build at all — this is a hard compiler constraint, confirming `25-CONTEXT.md`'s build-tag decision is not a style preference but a genuine necessity here.
**When to use:** Any OS-specific syscall package import (as opposed to OS-specific `exec.Command` argv shapes, which can stay parameter-dispatched per the `reveal.go` precedent).

### Pattern 3: Atomic write — temp file in the same directory, then rename
**What:** Already implemented once in this codebase for the counts-cache sidecar file, and directly reusable as the model for catalog JSON/HTML writes.
**When to use:** Every write that produces `outputRoot+".json"`/`.html"` in this phase, and (per Phase 27's later reuse) rename/duplicate/delete.
**Example:**
```go
// Source: internal/config/counts_cache.go:107-135 (verified, existing code in this repo)
func (c *CountsCache) save() error {
	data, err := json.MarshalIndent(c.entries, "", "  ")
	if err != nil {
		return err
	}
	dir := filepath.Dir(c.path)
	tmp, err := os.CreateTemp(dir, "counts-cache-*.tmp")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpPath)
		return err
	}
	if err := os.Rename(tmpPath, c.path); err != nil {
		os.Remove(tmpPath)
		return err
	}
	return nil
}
```
Generalize this into `internal/catalog/atomicwrite.go`'s `writeFileAtomic(dir, filename string, data []byte) error`, called by both `writeJSONFile` and `writeHTMLFile` (renamed internally to build the bytes, then delegate to the shared helper) instead of today's direct `os.WriteFile(path, jsonBytes, 0644)` `[VERIFIED: internal/catalog/service.go:172, current non-atomic write]`. `os.CreateTemp(dir, ...)` in the *same* directory as the final path is what makes `os.Rename` atomic (same filesystem) — using `os.TempDir()` would risk a cross-device rename failing outright on removable media.

### Anti-Patterns to Avoid
- **Emitting `runtime.EventsEmit` inside `internal/catalog`:** breaks COMPAT-04 outright — the CLI has no Wails runtime context, and this call would panic or hang if reached from `cli/create.go`'s call path. Keep the throttling and emission entirely in `app.go`'s progress callback closure.
- **Zero-value `Options{}` in the `CreateCatalog` wrapper:** `WriteHTML` must be explicitly `true` (see Common Pitfalls) or CLI's HTML output silently disappears — a COMPAT-03 regression with no compile-time signal.
- **A scan-id-keyed map for cancel handles:** over-engineered for a product that is one-scan-at-a-time by explicit requirement; a single mutex-guarded field is the simplest correct implementation (ponytail-aligned).

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Cross-platform disk free/total space | A custom `/proc`-parsing or `df`-shelling helper | `golang.org/x/sys/unix.Statfs` (darwin/linux) + `golang.org/x/sys/windows.GetDiskFreeSpaceEx` | Already resolved as a transitive dependency; both signatures verified this session against the exact pinned version |
| Atomic file replacement | A custom fsync-then-swap sequence | `os.CreateTemp` (same dir) + `os.Rename` | Already implemented once in this repo (`internal/config/counts_cache.go`) — copy the proven pattern, don't reinvent it |
| Focus trap / Escape / scroll lock for the slide-over | A new hook | `frontend/src/hooks/useModalBehavior.ts`, unchanged | Explicitly written for this consumer — its doc comment names Phase 25's animated exit as the reason its effect is keyed on `[isOpen]` alone `[VERIFIED: frontend/src/hooks/useModalBehavior.ts:6-14]` |
| Wails error message extraction | A new per-call try/catch pattern | `extractErrorMessage()`/`wailsError()` in `frontend/src/services/wailsAPI.ts` | Every existing binding call routes through this; Wails rejects with a plain string, not an `Error` |

**Key insight:** every "don't hand-roll" item above already has a working, in-repo precedent from Phases 22-24. This phase's job is extension, not invention, for everything except the walk-error classification logic (genuinely new) and the volume-enumeration per-OS files (genuinely new, but thin wrappers over two verified stdlib-adjacent functions).

## Common Pitfalls

### Pitfall 1: `Options{}` zero value silently breaks CLI's HTML output
**What goes wrong:** `CreateCatalog`'s wrapper constructs `Options{}` to pass to `CreateCatalogWithContext`. Go's zero value for `WriteHTML bool` is `false`. Today, `CreateCatalog` *always* writes both JSON and HTML unconditionally `[VERIFIED: internal/catalog/service.go:41-45]` — there is no toggle at all currently. If the wrapper doesn't explicitly set `Options{WriteHTML: true}`, `storcat create` stops producing `.html` files with zero compile-time or type-check signal.
**Why it happens:** The struct-literal shorthand `Options{}` reads as "use the defaults," and it's easy to assume the zero value is the safe default without checking what CLI's *current* unconditional behavior actually is.
**How to avoid:** The wrapper must read `Options{WriteHTML: true}` explicitly, with a comment citing this exact pitfall. Add (or extend) a CLI regression test asserting `result.HtmlPath` is non-empty and the file exists after `storcat create`.
**Warning signs:** `cli/*_test.go` not re-run after this phase's changes; any code review that doesn't diff `Options{}` construction sites against "what does zero-value mean here."

### Pitfall 2: Terminal-vs-single-entry misclassification silently produces a catalog "missing everything after the SD card was pulled," with no error surfaced
**What goes wrong:** Today, a directory whose `os.ReadDir` fails returns an empty node with `nil` error `[VERIFIED: internal/catalog/service.go:108-117, quoted below]`; a child whose `os.Stat` fails is dropped via `continue` and the loop moves to the next sibling `[VERIFIED: internal/catalog/service.go:140-144]`. If a volume disappears mid-walk, *every remaining entry* in the directory being processed — and every directory after it — will independently hit this same silent-skip path, one syscall at a time, with no signal ever reaching the caller that anything went wrong. The scan would appear to "complete successfully" with a catalog silently missing a large chunk of the tree.

Verbatim current code:
```go
// internal/catalog/service.go:106-117
if info.IsDir() {
    entries, err := os.ReadDir(dirPath)
    if err != nil {
        // Return empty directory if we can't read it
        return &models.CatalogItem{
            Type:     "directory",
            Name:     displayPath,
            Size:     0,
            Contents: []*models.CatalogItem{},
        }, nil
    }
```
```go
// internal/catalog/service.go:139-148
childPath := filepath.Join(dirPath, entry.Name())
childItem, err := s.traverseDirectory(childPath, basePath, onProgress)
if err != nil {
    // Skip items we can't access
    continue
}
contents = append(contents, childItem)
totalSize += childItem.Size
```
**Why it happens:** This is the exact convention CLAUDE.md's own project conventions call for on an *ordinary* single-file permission error ("Silent error swallowing for file access errors — skip inaccessible files/dirs") — the code correctly implements that convention for the common case, but a volume-vanished event is not that case, and nothing in the current code distinguishes them.
**How to avoid:** On any `os.ReadDir`/`os.Stat` error inside the walk, re-probe the *scan root* (the top-level path passed into `CreateCatalogWithContext`, not the failing subdirectory) with a cheap `os.Stat`. If the root itself now fails too, classify as terminal: stop the current directory's loop immediately (`break`, not `continue`), do not descend further anywhere, and propagate a sentinel error (e.g. `ErrVolumeVanished`) up through every recursion level so `CreateCatalogWithContext` can retain the partial tree and route to the error state — distinct from a lone bad file, which keeps today's skip-and-continue behavior with an added `ReadErrors` counter increment.
**Warning signs:** No code path that re-checks the scan root's own reachability; a "volume disappeared" manual test (unplug an SD card mid-scan) that produces a catalog with no size/progress discontinuity and no error UI.

### Pitfall 3: `context.Context` cancellation does not interrupt an in-flight blocking syscall
**What goes wrong:** `ctx.Err()` checks between directory boundaries stop the walk from *starting new* work promptly, but Go's standard library gives `os.ReadDir`/`os.Stat` no context-awareness — if a single syscall physically hangs (a common failure mode for a device that's disconnecting rather than already gone, e.g. tens of seconds of I/O wait before the OS marks the mount as dead), no amount of `ctx.Err()` checking elsewhere in the loop will interrupt that one blocked call.
**Why it happens:** This is a documented, longstanding Go runtime limitation, not a bug in this codebase: "the standard library doesn't provide context-aware versions of these functions, and the runtime doesn't have a built-in mechanism to interrupt such syscalls" `[CITED: golang/go issue #41054, cross-checked via WebSearch this session]`.
**How to avoid:** Document this as a known, accepted limitation (matches the project's own existing "skip inaccessible" tolerance for edge-case I/O) rather than attempting to hand-roll a goroutine+timeout wrapper around every syscall, which would add real complexity for a rare, cosmetic worst-case (the UI would show "Cancelling…" a few extra seconds rather than closing instantly). Check `ctx.Err()` at the top of `traverseDirectory` (before each `os.Stat`/`os.ReadDir` call) so cancellation *is* prompt for every syscall that hasn't started yet — only an already-in-flight blocked syscall is unaffected.
**Warning signs:** A manual cancel-test performed only against a healthy, responsive local disk (which will always cancel promptly) rather than against slow/removable media, which is the actual worst case this pitfall describes.

### Pitfall 4: `/Volumes` enumeration on macOS includes non-volume noise that would corrupt the volume-card list
**What goes wrong:** A naive `os.ReadDir("/Volumes")` on a real macOS machine returns entries that are not user-relevant external volumes at all. Live-verified on this session's own macOS 26.6.1 (Darwin 25.6.0) machine:
```
$ ls -la /Volumes
drwxr-xr-x   5 root  wheel  160 Jul  4 15:43 .timemachine
drwxr-xr-x   3 root  wheel   96 Feb 12  2026 com.apple.TimeMachine.localsnapshots
lrwxr-xr-x   1 root  wheel    1 Aug 13 09:52 Macintosh HD -> /
d--x--x--x   2 ken   wheel   64 Aug 13 09:54 pi-downloader
d--x--x--x   2 ken   wheel   64 Aug 13 09:54 software
```
`[VERIFIED: live `ls -la /Volumes` command, this session, macOS 26.6.1]`. `Macintosh HD` is a symlink to `/` (the boot volume — cataloguing it as a "volume card" with the entire boot disk's size would be wrong and enormous); `.timemachine` is a hidden dotfile; `com.apple.TimeMachine.localsnapshots` is a real (non-symlink) directory that is not a user-mountable volume at all. `pi-downloader` and `software` are real directories with `d--x--x--x` permissions — execute-only, no read — meaning `os.ReadDir` on either would fail with a permission error, which is a live, concrete example of exactly the "read errors" status tag CRT-02 needs to display.
**Why it happens:** `/Volumes` is macOS's general-purpose mount namespace, not an "external drives only" list; Apple does not document a formal API distinction between a real external volume and these internal artifacts (confirmed by targeted search — no authoritative macOS API exists for this filter; the standard approach in comparable tools is heuristic).
**How to avoid:** Filter out dotfile-prefixed entries (matches the existing hidden-file-skip convention already used elsewhere in this codebase) and entries whose resolved symlink target is `/` (the boot volume). Treat `com.apple.TimeMachine.localsnapshots`-style Apple-reserved names as a lower-confidence heuristic (e.g., skip anything matching `com.apple.*`) — flag this specific filter as `[ASSUMED]`, since it's grounded in one live observation plus general web search, not an Apple-documented convention, and should be spot-checked again on whatever macOS version CI actually builds against.
**Warning signs:** A volume-card list showing a card named "Macintosh HD" with the boot disk's full multi-hundred-GB size, or a card named "com.apple.TimeMachine.localsnapshots".

### Pitfall 5: A per-file `EventsEmit` floods the bridge at 40k-node scale
**What goes wrong:** `traverseDirectory`'s `onProgress` callback currently fires on *every single file and directory visited* `[VERIFIED: internal/catalog/service.go:92-95]`. If `app.go`'s new progress callback calls `runtime.EventsEmit` unconditionally inside that callback, a fast SSD/SD-card walk emits thousands of IPC messages per second — Phase 23 measured 5.83MB of JSON for a single 42,551-node catalog `[VERIFIED: internal/search/search_indexed.go:8, comment: "Phase 23 measured 5.83MB of JSON for a single 42,551-node catalog"]`, giving a concrete sense of scale for this milestone's target catalogs.
**How to avoid:** Throttle inside `app.go`'s closure — e.g., track `lastEmit time.Time`, only call `EventsEmit` if `time.Since(lastEmit) > 200*time.Millisecond` (matching the design handoff's own mocked cadence of 220ms, `[VERIFIED: design_handoff_storcat_ui/designs/StorCat 1a Demo.dc.html:849]`, `}, 220);`), always emitting the *latest* counters/path rather than every intermediate one. `internal/catalog` itself stays untouched — it keeps calling `onProgress` on every file, exactly as it already does; only the Wails-aware wrapper in `app.go` decides how often to forward that to the frontend.
**Warning signs:** A stuttering percentage counter or visibly high CPU during a large-volume scan; no `time.Since`/ticker-style gate anywhere in the new `app.go` progress closure.

## Code Examples

### `ProgressUpdate` struct and throttled emission (Claude's Discretion, resolved)
```go
// internal/catalog — new type, replaces `func(path string)`
type ProgressUpdate struct {
    Path       string
    FilesSeen  int
    BytesSeen  int64
    ReadErrors int
}
type ProgressCallback func(ProgressUpdate)
```
```go
// app.go — throttling lives here, never in internal/catalog (COMPAT-04)
func (a *App) throttledProgress() catalog.ProgressCallback {
    var lastEmit time.Time
    return func(u catalog.ProgressUpdate) {
        if time.Since(lastEmit) < 200*time.Millisecond {
            return
        }
        lastEmit = time.Now()
        runtime.EventsEmit(a.ctx, "scan:progress", u)
    }
}
```
Frontend subscription, matching the existing `EventsOn` signature `[VERIFIED: frontend/wailsjs/runtime/runtime.d.ts:41, `export function EventsOn(eventName: string, callback: (...data: any) => void): () => void;`]`:
```typescript
useEffect(() => {
  if (!isOpen) return;
  const unsubscribe = EventsOn('scan:progress', (update) => {
    // update the lifted AppContext scan state
  });
  return unsubscribe; // Pitfall 13 in PITFALLS.md: StrictMode double-invoke leak if omitted
}, [isOpen]);
```

### `CreateCatalogWithContext` wrapper preserving CLI compatibility exactly
```go
// internal/catalog/service.go — CreateCatalog becomes a thin wrapper
func (s *Service) CreateCatalog(title, directoryPath, outputRoot, copyToDirectory string, onProgress ProgressCallback) (*models.CreateCatalogResult, error) {
    return s.CreateCatalogWithContext(
        context.Background(),
        title,
        directoryPath, // sourcePath: walked
        directoryPath, // outputDir: written -- SAME as source, preserving today's exact behavior
        outputRoot,
        copyToDirectory,
        Options{WriteHTML: true}, // NOT the zero value -- see Pitfall 1
        onProgress,
    )
}
```
`cli/create.go:81`'s call, `svc.CreateCatalog(*title, dir, *name, outputDir, nil)` `[VERIFIED: cli/create.go:81]`, needs zero edits — `nil` type-checks against any `ProgressCallback` function-type parameter regardless of the struct it now carries.

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|---------------|--------|
| `CreateCatalog` writes into the scanned directory itself | `CreateCatalogWithContext` writes into a separate, app-configured `outputDir` | This phase | The scanned volume is treated as read-only source; the catalog directory (already tracked in `frontend`'s `AppContext`/`localStorage`) becomes the real write destination — verified against the design handoff's own mocked state (`~/dev/sd-catalogs/`) |
| `ProgressCallback func(path string)` | `ProgressCallback func(ProgressUpdate)` | This phase | Only internal callers affected (CLI passes `nil`); safe, additive |
| Read errors silently swallowed with no signal | Read errors collected + classified (single-entry vs terminal) | This phase | New capability required by CRT-10; must not change the *existing* single-entry skip behavior CLI relies on |
| Direct `os.WriteFile` for catalog JSON/HTML | Temp-file-same-dir + `os.Rename` | This phase | First time catalog writes (not just the sidecar cache) become crash-safe; Phase 27 reuses this helper |

**Deprecated/outdated:** None — this is the first phase to touch `traverseDirectory` since the Go/Wails rewrite; there is no prior "old way" beyond what's described above.

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | `com.apple.*`-prefixed `/Volumes` entries (beyond the known `com.apple.TimeMachine.localsnapshots` case) should be filtered from the volume-card list | Common Pitfalls #4 | A stray Apple-internal mount could appear as a bogus volume card on a macOS version/config not tested this session; low severity (cosmetic), easy to patch later |
| A2 | The `beforeClose` → cancel → bounded-wait → `runtime.Quit(ctx)` sequence is safe to call *from within* `beforeClose` itself (i.e., `runtime.Quit` re-triggers close without infinite-looping back into `beforeClose`) | Summary; Pattern 1 area | If wrong, the app could hang on quit during an active scan, or `beforeClose` could re-enter recursively. This needs live verification during execution (force-quit-mid-scan test), not just static analysis — flagged as MEDIUM confidence, not HIGH, since I could not find this exact sequence documented in the vendored Wails source, only its constituent primitives |
| A3 | Windows drive-letter enumeration + `GetDiskFreeSpaceEx` will work as designed | Volume Enumeration / Standard Stack | No Windows machine was available this session (consistent with `WINDOWS.md` entries #1 and #2). The signature is verified against the real `golang.org/x/sys@v0.30.0/windows` source, but runtime behavior is unverified. Must be pre-logged to `.planning/WINDOWS.md` as entry #3 per `25-CONTEXT.md`'s own instruction |
| A4 | Linux `/media` + `/mnt` + `/proc/mounts` cross-check behaves as `25-CONTEXT.md` describes | Volume Enumeration | No Linux machine was available this session either (not currently tracked in `WINDOWS.md`, which is Windows-specific — the planner should decide whether to log this gap somewhere, since it is a genuine, symmetrical unverified-platform risk) |

**If this table is empty:** N/A — see rows above.

## Open Questions

1. **Does "Open in workspace" (CRT-12) need to add the new catalog to `state.catalogs` directly, or trigger a fresh `BrowseCatalogs` call?**
   - What we know: `CatalogRail.tsx` already has a `loadCatalogsForDirectory` helper that calls `wailsAPI.browseCatalogs(dir)` and dispatches `SET_CATALOGS` `[VERIFIED: frontend/src/components/workspace/CatalogRail.tsx:20-32]`. Since the new catalog now genuinely lands in the configured catalog directory (per this research's headline finding), a fresh `BrowseCatalogs` call after a successful scan is sufficient and simplest — no special "prepend to rail" logic needed.
   - What's unclear: Whether the plan should optimistically prepend the known result (avoiding a round-trip) or just re-fetch. Given catalog counts are typically small (tens, not thousands), re-fetch is simpler and matches the existing pattern.
   - Recommendation: re-fetch via the existing `loadCatalogsForDirectory`-equivalent path; do not build new optimistic-update logic for this phase.

2. **Exact wire shape of the "terminal volume-vanished" signal returned from `CreateCatalogWithContext` to `app.go`.**
   - What we know: it must be distinguishable from a plain user-cancel (`ctx.Err() == context.Canceled`) so `app.go` can decide whether to offer "Write partial catalog." A typed sentinel error (`var ErrVolumeVanished = errors.New(...)`, checked with `errors.Is`) plus the retained partial tree is the cleanest shape given Go idioms.
   - What's unclear: whether the partial tree and the read-error log (needed for the error state's "read error: {path}" lines, per `25-UI-SPEC.md`) should be bundled into a single returned struct or attached to the `App` struct out-of-band. This research recommends bundling into a `partialScanResult` struct returned alongside the sentinel error, then the caller (`app.go`) is responsible for retaining it — keeps `internal/catalog` stateless and CLI-safe.
   - Recommendation: plan-phase should finalize this exact struct shape as part of task breakdown; the design principle (typed sentinel + returned struct, not global mutable state inside `internal/catalog`) is the load-bearing part.

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| macOS (Darwin) | Live verification of `/Volumes` enumeration, `unix.Statfs` on darwin | ✓ | Darwin 25.6.0 / macOS 26.6.1 | — |
| Windows machine/VM | Runtime verification of `GetDiskFreeSpaceEx`, drive-letter enumeration | ✗ | — | Static signature verification only (done this session); log to `.planning/WINDOWS.md`, ship unverified per project's established precedent (entries #1, #2) |
| Linux machine/VM | Runtime verification of `/media`+`/mnt`+`/proc/mounts` cross-check | ✗ | — | Static reasoning only; `unix.Statfs` signature is shared with darwin (same `golang.org/x/sys/unix` package) so the syscall shape is verified, but the enumeration heuristic itself is unverified |
| `golang.org/x/sys@v0.30.0` module cache | `unix.Statfs`, `windows.GetDiskFreeSpaceEx` signatures | ✓ (already in local module cache, used for verification this session) | v0.30.0 | — |
| Wails v2.10.2 module cache | `EventsEmit`/`EventsOn`/`OnBeforeClose`/`runtime.Quit` signatures | ✓ (already in local module cache, used for verification this session) | v2.10.2 | — |

**Missing dependencies with no fallback:** none — every missing environment (Windows, Linux) has an accepted, project-precedented fallback (ship unverified, log to the Windows ledger, verify at a later sweep).

**Missing dependencies with fallback:** Windows and Linux runtime verification, per above.

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go stdlib `testing` (table-driven, `*_test.go` beside source) — no framework for Go; frontend has none by design (TEST-01 deferred) |
| Config file | none — plain `go test ./...` |
| Quick run command | `go test ./internal/catalog/... ./internal/search/... ./cli/...` |
| Full suite command | `go test ./... && cd frontend && npx tsc --noEmit && npx vite build` |

### Phase Requirements → Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| CRT-09 | Cancel actually stops the walk, writes nothing | unit | `go test ./internal/catalog/... -run TestCreateCatalogWithContext_Cancel -v` | ❌ Wave 0 |
| CRT-10 | Terminal vs single-entry error classification | unit | `go test ./internal/catalog/... -run TestTraverseDirectory_TerminalError -v` | ❌ Wave 0 |
| CRT-11 | Partial catalog write produces correct marker shape | unit | `go test ./internal/catalog/... -run TestWritePartialCatalog_Marker -v` | ❌ Wave 0 |
| COMPAT-02 | Clean-scan JSON byte-identical to pre-milestone shape | unit | `go test ./internal/catalog/... -run TestCreateCatalog_JSONShapeUnchanged -v` | ❌ Wave 0 |
| COMPAT-03 | CLI `create` behaves identically (incl. `WriteHTML` default) | unit + smoke | `go test ./cli/... -v` then `go run . create <tmpdir> --json` | ✅ existing `cli/*_test.go`, extend |
| COMPAT-04 | `internal/catalog` has no Wails import | static check | `go list -deps ./internal/catalog/... \| grep wailsapp` (expect empty) | ❌ Wave 0 (add as a CI-style assertion, not a Go test) |
| — | Atomic write survives a simulated crash (kill mid-write) | manual | Kill process during a large-volume scan, verify no truncated JSON, prior catalog (if overwriting — N/A this phase) unaffected | manual, documented in plan |
| CRT-02, Pitfall 4 | `/Volumes` filtering excludes boot volume + Apple-internal mounts | manual (dev-browser + live machine) | Run `wails dev`, open Create slide-over, inspect volume cards against live `ls /Volumes` | manual |

### Sampling Rate
- **Per task commit:** `go test ./internal/catalog/... ./internal/search/...`
- **Per wave merge:** `go test ./... && cd frontend && npx tsc --noEmit && npx vite build`
- **Phase gate:** Full suite green, plus the manual volume-disappearance and force-quit-mid-scan checks (cannot be automated without a real removable device), before `/gsd-verify-work`.

### Wave 0 Gaps
- [ ] `internal/catalog/service_test.go` extensions — cancellation, terminal-error classification, partial-marker `omitempty` shape, JSON byte-diff against a pre-milestone fixture
- [ ] `internal/volumes/volumes_test.go` — new package, needs its own test file (darwin-buildable assertions at minimum; windows/linux behind build tags, structurally testable only where the toolchain can build them)
- [ ] No new frontend test framework (TEST-01 stays deferred) — frontend proof remains `tsc --noEmit` + `vite build` + live `dev-browser` verification against `wails dev` on port `34115` (never Vite's `5173`, which exposes no `window.go`)

## Security Domain

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V2 Authentication | no | Single-user desktop app, no auth surface |
| V3 Session Management | no | No sessions |
| V4 Access Control | partial | `outputDir`/`copyToDirectory`/`sourcePath` are all user-chosen via native OS dialogs (`SelectDirectory`), not attacker-influenceable strings from an external source — same trust model as existing `RevealInFileManager`'s `catalogDir` containment check `[VERIFIED: internal/osutil/reveal.go:110-131]` |
| V5 Input Validation | yes | Filename root / title are free-text; must not be used to build a shell command (this phase adds no `exec.Command` calls at all — pure `os`/`path/filepath`); `filepath.Join` for path construction, never string concatenation |
| V6 Cryptography | no | No cryptographic operations this phase |
| V12 File Handling | yes | Atomic write (temp+rename) prevents partial/corrupt files on crash; temp file must be created in the *same* directory as the target (not a shared world-writable temp dir) to avoid both cross-device rename failures and TOCTOU risk from a predictable shared temp path |

### Known Threat Patterns for this stack

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| Path traversal via a crafted filename-root (e.g. `../../etc/passwd`) | Tampering | `filepath.Join(outputDir, outputRoot+".json")` still permits `..` segments in `outputRoot` to escape `outputDir` — since `outputRoot` is user-typed free text (CRT-04's "Filename root" field), the write path should be validated with the same `containsPath`-style containment check `internal/osutil/reveal.go:91-108` already implements for reads, applied here to the *write* path before any `os.CreateTemp`/`os.Rename` call |
| Symlink-swap between validation and write (TOCTOU) | Tampering | Lower severity here than in `RevealInFileManager` (which reveals attacker-influenceable catalog files) since `outputDir` is always an OS-native-dialog-selected directory the user just chose, not an arbitrary external input — accept the same residual risk `internal/osutil/reveal.go:141-147`'s comment already accepts for reads, unless the plan-phase decides otherwise |
| Predictable/racy temp filename | Information Disclosure / Tampering | `os.CreateTemp` (not a hand-rolled `fmt.Sprintf` temp name) generates a random suffix and opens the file exclusively — reuse this stdlib function exactly as `counts_cache.go:114` already does, never construct a temp path by string formatting alone |
| Race between two "Retry scan" clicks reusing a stale cancel handle | Tampering (internal, low severity) | The single mutex-guarded `activeScanCancel` field must be nilled out only by the goroutine that owns the *current* scan, not by a subsequent scan's setup — guard with the same struct-held mutex for both read and write, as sketched in Pattern 1 |

## Sources

### Primary (HIGH confidence — read directly this session)
- `internal/catalog/service.go` — full file, current `CreateCatalog`/`traverseDirectory`/write functions
- `cli/create.go`, `cli/show.go` — CLI call sites and shared `search.Service.LoadCatalog` consumer
- `app.go` — full file, `App` struct, `CreateCatalog`, `beforeClose`, `SelectDirectory`
- `pkg/models/catalog.go` — `CatalogItem`, `FlatNode`, `FlatCatalog`, `CreateCatalogResult`
- `internal/search/service.go`, `internal/search/flatten.go`, `internal/search/search_indexed.go` — `LoadCatalog` dual-format parse, `LoadCatalogFlat`, the Phase 23/24 "new method beside the CLI-shared one" precedent
- `internal/osutil/reveal.go` — per-OS dispatch-by-parameter precedent and its explicit rationale for *not* using build tags; `containsPath` containment pattern
- `internal/config/config.go`, `internal/config/counts_cache.go` — config persistence pattern, and the exact atomic temp+rename write already proven in this repo
- `internal/catalog/service_test.go` — existing test conventions to extend
- `internal/fixture/fixture.go` — synthetic large-catalog fixture generators available for scale testing
- `frontend/src/hooks/useModalBehavior.ts` — shared overlay-behavior hook, explicitly written with Phase 25 in mind
- `frontend/src/components/workspace/CommandPalette.tsx`, `StatusBar.tsx`, `CatalogRail.tsx` — always-mounted overlay pattern, existing status-bar segment structure, `localStorage` persistence convention (`safeGetItem`/`safeSetItem`, `storcat-*` key naming)
- `frontend/src/services/wailsAPI.ts` — `extractErrorMessage`/`wailsError` pattern
- `frontend/wailsjs/runtime/runtime.d.ts` — `EventsOn`/`EventsEmit`/`EventsOff` frontend signatures
- `design_handoff_storcat_ui/designs/StorCat 1a Demo.dc.html` — lines 780-850 (`closeCreate`, `startScan`, 220ms mock cadence), line 1148 (`willWrite` literal, the headline finding on output-directory semantics)
- `go.mod` — exact pinned versions (`golang.org/x/sys v0.30.0 // indirect`, `wailsapp/wails/v2 v2.10.2`)
- `golang.org/x/sys@v0.30.0` vendored source (local module cache) — `unix.Statfs` signature and `Statfs_t` field types on darwin/linux, `windows.GetDiskFreeSpaceEx` signature and its `UTF16PtrFromString` usage pattern from the package's own test file
- `github.com/wailsapp/wails/v2@v2.10.2` vendored source (local module cache) — `pkg/runtime/events.go` (`EventsEmit`/`EventsOn`), `pkg/options/options.go:64` (`OnBeforeClose func(ctx) (prevent bool)`), `pkg/runtime/runtime.go:64` (`Quit`), `internal/frontend/dispatcher/calls.go` + `internal/frontend/desktop/darwin/frontend.go:389` (per-call goroutine dispatch), `internal/binding/*.go` (no `context.Context` special-casing in bound-method argument parsing)
- Live shell commands this session: `ls -la /Volumes`, `df -k /`, `mount`, `sw_vers` on this machine (macOS 26.6.1, Darwin 25.6.0)
- `.planning/REQUIREMENTS.md`, `.planning/STATE.md`, `25-CONTEXT.md`, `25-UI-SPEC.md` — locked decisions and design contract

### Secondary (MEDIUM confidence — web-sourced, cross-checked)
- `golang/go` issue #41054 — Go runtime's inability to interrupt blocked syscalls via context cancellation, cross-checked against the stated limitation being consistent with `os.ReadDir`/`os.Stat` having no context-aware variants in the standard library
- General web search on macOS `/Volumes` Time Machine local-snapshot behavior — confirms `com.apple.TimeMachine.localsnapshots` is a real, non-volume mount artifact, but no authoritative Apple documentation of a formal filter API was found

### Tertiary (LOW confidence, flagged as assumptions)
- Windows drive-letter enumeration and `GetDiskFreeSpaceEx` runtime behavior — signature verified, runtime unverified (no Windows machine)
- Linux `/media`+`/mnt`+`/proc/mounts` cross-check runtime behavior — no Linux machine available this session
- `com.apple.*`-prefix filtering as a general rule beyond the one observed case

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — every library/API cited was verified against the exact pinned version's vendored source in the local module cache, not training-data recall
- Architecture: HIGH for the two headline ROADMAP-flagged questions (both resolved against directly-read code and the design handoff's own source-of-truth markup); MEDIUM for the `beforeClose`→`Quit` sequencing recommendation (verified primitives, unverified end-to-end sequence)
- Pitfalls: HIGH for codebase-grounded pitfalls (1, 2, 5); MEDIUM for pitfall 3 (web-sourced Go runtime limitation, well-established); MEDIUM for pitfall 4 (one live machine's observation, generalized)

**Research date:** 2026-08-14
**Valid until:** 30 days (stable Go/Wails APIs; the one fast-moving risk is macOS `/Volumes` behavior across OS updates, worth re-spot-checking if this phase's execution slips past a macOS point release)
