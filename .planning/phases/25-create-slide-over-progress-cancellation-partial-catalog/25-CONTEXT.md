# Phase 25: Create Slide-over + Progress/Cancellation/Partial-Catalog - Context

**Gathered:** 2026-08-14
**Status:** Ready for planning
**Mode:** Smart discuss (autonomous) — all three grey areas accepted as recommended

<domain>
## Phase Boundary

Users create a new catalog from a detected volume or folder, watch it scan live, and can cancel it or recover from a volume that disappears mid-scan — all without risking data loss or breaking the CLI.

**In scope:** the 560px right slide-over and its five close paths; volume detection and the volume/folder picker; catalog title, filename root and the live "WILL WRITE" preview; the three creation toggles (write HTML, copy to secondary, include hidden); live scan progress (percentage, files, bytes, ETA, current path, newest-first log); "Run in background" handoff to the status bar; real cancellation that stops the underlying walk; the volume-disappeared error state with partial-write / retry / cancel; the done state; crash-safe (atomic) catalog writes; and the backend work all of that requires — `context.Context` threading, a richer progress callback, and per-OS volume enumeration.

**Out of scope (later phases):** Settings surface (Phase 26), catalog actions and fsnotify watch (Phase 27), re-scan and diff (Phase 28). Rename/duplicate/delete are Phase 27 even though they also write catalogs — this phase only establishes the atomic-write primitive they will reuse.

**This is the milestone's highest-regression-risk phase.** It is the only phase that modifies existing, tested Go code (`internal/catalog`'s `traverseDirectory`, its error-return contract, and the `ProgressCallback` signature). ROADMAP carries an explicit research flag: plan-phase must run a dedicated research pass before planning, to resolve the forced-close partial-write policy and the on-disk "unreadable subtree" marker shape.

</domain>

<decisions>
## Implementation Decisions

### Backend Surface & CLI Compatibility
- **Cancellation is threaded as `ctx context.Context` through `traverseDirectory`, and reached via a new `CreateCatalogWithContext(ctx, …, opts)` method. The existing `CreateCatalog(title, directoryPath, outputRoot, copyToDirectory, onProgress)` remains as a thin wrapper that calls it with `context.Background()` and default options.** The point of the wrapper is that **`cli/create.go` is left literally untouched** — its call at `cli/create.go:81` still compiles and still behaves identically. COMPAT-03 and COMPAT-04 then hold *by construction* rather than by test coverage, which is the strongest available guarantee for the one phase that modifies shared, CLI-critical code.
- **`ProgressCallback` changes to carry a struct** (path, files seen, bytes seen, read-error count — exact shape at Claude's discretion) instead of `func(path string)`. This is safe to change rather than duplicate: the type lives in `internal/`, so nothing outside this module can import it, and its only existing caller (`cli/create.go:81`) passes `nil`.
- **Percentage and ETA are computed against the selected volume's total size**, which CRT-02's volume cards must compute anyway. For a plain folder chosen via CRT-03 (no volume total available), run a **fast count-only pre-pass** and show indeterminate progress until it completes. A two-pass count for every scan was rejected — it doubles I/O on exactly the 40,000-node volumes this milestone targets. Dropping percentage entirely was rejected: CRT-07 requires it and the design handoff's rule is "No spinners — progress is always a real number."
- **Progress reaches the frontend via Wails `runtime.EventsEmit`, emitted from `app.go` only — never from `internal/catalog`.** COMPAT-04 requires `internal/catalog` stay usable from the CLI with no Wails runtime context; an emit inside the catalog package would couple it to a runtime the CLI does not have. The catalog package raises progress through the callback; the app layer decides to emit.

### Cancellation, Partial Writes & the Unreadable-Subtree Marker
- **What gets written on stop depends on the cause, and the split is deliberate:** a user cancel (CRT-09) and a window close mid-scan (CRT-13) write **nothing**. Only the **volume-disappeared error state** (CRT-10) offers a partial write (CRT-11). This mirrors the requirements exactly — a cancel is a decision not to have the catalog, whereas a vanished volume is a loss the user may want to salvage.
- **A partial catalog marks the unreadable subtree on the affected directory nodes only**, and this is an explicit, documented divergence from COMPAT-02 **scoped to partial catalogs**. A complete catalog remains byte-for-byte the shape v2.3.0 wrote. The exact field name and shape are **deferred to the research pass** — this is one of the two items ROADMAP's research flag names. Silently omitting unreadable subtrees was rejected (invisible data loss); a sidecar file was rejected (splits the artifact from the catalog it describes).
- **Crash-safe writes (write to temp, then `rename`) are implemented in this phase**, not deferred. This phase already owns the write path, and Phase 27's ACT-09 ("no catalog write can corrupt an existing catalog if the app crashes mid-write") then *reuses* the primitive across rename/duplicate/delete rather than retrofitting crash-safety onto four more call sites later.
- **Forced close is wired through the existing `beforeClose` hook (`app.go:192`)** so CRT-13 is enforced at the one place the OS actually notifies the app. A renderer-side `beforeunload` handler was rejected — it is unreliable-to-absent in a webview and fires too late to cancel a Go-side walk.

### Volume Detection & the Slide-over
- **Volume enumeration is stdlib-only, in per-OS build-tagged files:** macOS `/Volumes`, Linux `/media` + `/mnt` (cross-checked against `/proc/mounts`), Windows drive letters. No new Go dependency — the milestone's one-new-dependency budget is already committed to `Bios-Marcel/wastebasket/v2` in Phase 27.
- **Size and free space come from `syscall.Statfs` on unix and `GetDiskFreeSpaceEx` on Windows**, in those same build-tagged files. **The Windows path cannot be runtime-verified here** (no Windows machine available, consistent with WINDOWS.md entries #1 and #2) — pre-log it to `.planning/WINDOWS.md` for the pre-v3.0.0 sweep rather than claiming it works.
- **A volume card's `read errors` status comes from probing the volume root only** (`os.ReadDir` on the mount point). Deeper failures are not pre-discovered — they surface during the scan and drive CRT-10's error state. Pre-walking every volume to populate a picker was rejected as far too slow.
- **The slide-over's 340ms-in / 260ms-out animation extends Phase 24's `useModalBehavior` hook** rather than adding bespoke animation state. That hook was explicitly designed for this consumer — its cleanup already keys on the `isOpen` transition rather than unmount, precisely so an exit animation can run without the page staying scroll-locked. PLT-07 forbids Phases 25–27 reimplementing these behaviors.

### Claude's Discretion
- The exact `ProgressUpdate` struct shape, field names, and emit cadence (throttling a per-file callback into UI-rate events).
- The new method's name and option-struct shape (`CreateCatalogWithContext` is indicative, not binding).
- Volume-card component decomposition and the "WILL WRITE" preview's exact rendering.
- The newest-first log's retention limit and whether it is virtualized.
- How "Run in background" hands off state between the slide-over and the status bar.
- Temp-file naming and placement for the atomic write.

</decisions>

<code_context>
## Existing Code Insights

### Reusable Assets
- `internal/catalog/service.go` — `CreateCatalog` (line 28), `traverseDirectory` (74), `writeJSONFile` (166), `writeHTMLFile` (176), `copyFile` (339), plus `countFiles`/`countDirectories`/`formatBytes` helpers. This is the code the phase modifies.
- `cli/create.go:80-81` — the CLI's only construction of the service and its single `CreateCatalog(…, nil)` call. **The compatibility anchor.**
- `app.go:64-82` — `App.CreateCatalog`, which today builds a no-op progress callback with a literal comment saying progress could be emitted "in the future". This phase is that future.
- `app.go:192` — the existing `beforeClose(ctx)` hook, already wired by Wails, where CRT-13's cancel-on-close belongs.
- `frontend/src/hooks/useModalBehavior.ts` — Phase 24's shared hook (focus trap, Escape, scroll lock, focus restore), cleanup keyed on the `isOpen` transition specifically so an animated exit works.
- `frontend/src/components/workspace/StatusBar.tsx` — where CRT-08's `● scanning <name> · N%` lands.
- `frontend/src/lib/format.ts` — existing byte/count formatting.
- `internal/osutil` — has `containsPath`; relevant to path-safety on the output directory.
- `internal/fixture` + `scripts/gen-fixture-catalog` — fixture generation, useful for scan-progress and partial-catalog testing.

### Established Patterns
- New GUI-only Go capability goes in a **new method beside the CLI-shared one**, with the shared one left byte-unchanged — the `LoadCatalogFlat` (Phase 23) and `SearchIndexed` (Phase 24) precedent, now applied a third time.
- Every Wails binding call from the frontend routes through `frontend/src/services/wailsAPI.ts`'s `extractErrorMessage()` — Wails rejects with a plain string, not an `Error`.
- Go tests are table-driven `*_test.go` beside the source. There is **no frontend test framework** (TEST-01 explicitly deferred); frontend proof is `tsc --noEmit` + `vite build` + live `dev-browser` verification against `wails dev` on **`:34115`** (never Vite's `:5173`, which exposes no `window.go`).
- Styling is CSS custom properties in `workspace.css` + inline styles. Phase 22's z-index scale puts the **create slide-over at 200**, alongside the palette.

### Integration Points
- `internal/catalog/service.go` → `ctx` threading, richer callback, include-hidden option, partial-write path, atomic write.
- `app.go` → new context-aware binding, `runtime.EventsEmit` for progress, cancel handle, `beforeClose` wiring, new volume-enumeration binding.
- New per-OS Go files for volume enumeration + `Statfs`/`GetDiskFreeSpaceEx`.
- `frontend/wailsjs/**` → must be regenerated (`wails generate module`) after the new bindings and structs land.
- `WorkspaceShell.tsx` → mounts the slide-over; the rail's "＋ New" pill and `CatalogRail`'s directory chip get their real targets here.
- `StatusBar.tsx` → background-scan indicator.

</code_context>

<specifics>
## Specific Ideas

- CRT-01's timings are exact: **340ms in, 260ms out**, and the component must **not unmount early** — all five close paths (Escape, ×, Cancel, scrim, "Open in workspace") run the same exit.
- CRT-08's status-bar copy is specified: `● scanning <name> · N%`.
- The design handoff's rule, already carried through Phases 22–24: **"No spinners — progress is always a real number."** This is why CRT-07's percentage is a requirement rather than a nicety, and why the folder-scan pre-pass exists.
- COMPAT-02's "byte-for-byte the same JSON shape as v2.3.0" applies to **complete** catalogs. The partial-catalog marker is the one scoped, documented exception, and the research pass must confirm no field ordering or formatting drift creeps into the complete path.
- `.planning/WINDOWS.md` already carries entries #1 (Phase 23 `explorer /select,`) and #2 (Phase 24 Ctrl+K). The Windows `GetDiskFreeSpaceEx` path becomes #3 — pre-logged, not assumed.

</specifics>

<deferred>
## Deferred Ideas

- Re-scan and diff of an existing catalog — Phase 28 (ACT-06/07/08).
- Rename, duplicate, delete-to-Trash — Phase 27. They reuse this phase's atomic-write primitive but are not built here.
- Watching the catalog directory for outside changes — Phase 27 (WATCH-01/02/03).
- CLI subcommands for the new capabilities (volume listing, cancellable create) — the milestone decision is that new capabilities are GUI-only; FUT-03 already defers CLI parity.
- Frontend unit tests for the slide-over and progress rendering — TEST-01, deferred at v3.0.0 requirements definition.
- Resuming an interrupted scan from where it stopped — not a requirement; CRT-11 offers retry-from-scratch, not resume.

</deferred>
