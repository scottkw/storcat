# Phase 28: Re-scan & Diff - Pattern Map

**Mapped:** 2026-08-16
**Files analyzed:** ~14 (backend: 4 new/modified, frontend: ~9 new/modified, CI: 1 operational task)
**Analogs found:** 12 / 14 (2 have no analog — diff algorithm, RescanDialog shell)

**Key framing:** this phase is dominated by "modify existing tested code," not "write new code." Every pattern
below points at the EXACT existing seam to extract from or extend, not a generic style example.

## File Classification

| New/Modified File | Role | Data Flow | Closest Analog | Match Quality |
|---|---|---|---|---|
| `internal/catalog/service.go` (extract `Walk`) | service | transform (in-memory tree build) | itself — `CreateCatalogWithContext`/`traverseDirectory` (`service.go:133-203`, `:263-414`) | exact (extraction of existing code) |
| `internal/catalog/service.go` (`Options.MarkUnreadableOnSkip`, skip branches) | service | transform | itself — `service.go:298-320`, `:375-394` (existing skip-and-continue branches) | exact (behavior-gated extension) |
| `internal/catalog/service.go` (mtime capture) | service | transform | itself — `service.go:286-294` (file branch), directory branch's `info` at `:268` | exact |
| `internal/catalog/diff.go` (NEW) | service | transform (pure, in-memory) | `internal/search/flatten.go`'s `LoadCatalogFlat` (flatten convention only — root excluded, recurse children) | no true analog — genuinely new algorithm; see "No Analog Found" |
| `pkg/models/catalog.go` (`CatalogItem.ModTime`, new `DiffResult`/`DiffEntry` types) | model | CRUD (struct/JSON schema) | itself — `CatalogItem`'s existing `Unreadable`/`ReadError`/`Title` `omitempty` fields (`catalog.go:26-38`) | exact (same omitempty precedent to copy) |
| `internal/catalog/duplicate.go` (reused wholesale for Keep-both) | service | file I/O | itself — `DuplicateCatalog`/`nextCopyRoot` (`duplicate.go:23-106`) | exact (verbatim reuse, no new code) |
| `app.go` — `RescanCatalog` binding (NEW) | controller (Wails binding) | request-response | `app.go`'s `StartScan`/`startScan` (`app.go:248-389`) — for shape/validation/scanMu ONLY, must diverge on `lastPartial` | role-match, with an explicit divergence documented (see Pattern below) |
| `app.go` — `ResolveRescan` binding (NEW, overwrite/keep-both/discard) | controller (Wails binding) | request-response | `app.go`'s `DeleteCatalog` (`app.go:993-1026`) for containment-check shape; `WriteCatalogFrom`/`DuplicateCatalog` for the actual write | role-match |
| `app.go` — STATE-03 "remove from library" (reuses `DeleteCatalog` verbatim) | controller | request-response | `app.go:993-1026` `DeleteCatalog` | exact (zero new Go code — frontend wires to the existing binding) |
| `frontend/src/components/workspace/rescan/RescanDialog.tsx` (NEW) | component | request-response + streaming (progress) | `frontend/src/components/workspace/create/CreateSlideOver.tsx` (step-machine shell, `state.scan` subscription) for orchestration; `27-UI-SPEC.md`'s `RenameDialog`/`DeleteConfirmDialog`/`SettingsDialog` for the "own class family, not `DialogShell`" precedent | role-match (own-shell precedent), no true 620px 3-step analog |
| `frontend/src/components/workspace/rescan/*` — Step 1 (pick volume) | component | request-response | `VolumePicker.tsx` reused **unmodified** | exact (verbatim reuse) |
| `frontend/src/components/workspace/rescan/*` — Step 2 (scanning) / Error step | component | streaming (progress events) | `ScanningBody.tsx` / `ErrorBody.tsx` reused with additive optional props | exact (verbatim reuse + additive props) |
| `frontend/src/components/workspace/rescan/DiffList.tsx` (NEW) | component | transform/render (flat list) | `frontend/src/components/workspace/TreePane.tsx`'s row rendering pattern (NOT its virtualizer/`useVisibleRows`) | partial — explicitly NOT reusing the virtualized-tree machinery, see below |
| `frontend/src/components/workspace/UnreadableCatalogPanel.tsx:111-113` (STATE-03 trio) | component | request-response | `DetailsPanel.tsx`'s `Footer`'s `handleOpenHtml` (`:246-268`) for action 2; `DeleteConfirmDialog` (Phase 27) for action 3 | exact (both reused wholesale) |
| `frontend/src/components/workspace/DetailsPanel.tsx` (`Footer`, `CatalogActions`) | component | request-response | itself — existing stub comment at `Footer` (`:201-212`) and `CatalogActions`' `Menu` items | exact (extension point already exists) |
| `frontend/src/contexts/AppContext.tsx` (`state.rescan` NEW slice) | store (reducer) | event-driven | itself — the existing `state.scan` slice (`AppContext.tsx:37-39`, `:147`, `:312-421`) | role-match — pattern to copy the shape from, NOT to extend |
| `.github/workflows/build.yml` (COMPAT-06 proof) | config/CI | batch (CI run) | itself — already exists, unmodified; the "file" being produced is a real `git push`/PR, not new YAML | exact (operational task, no code change) |

## Pattern Assignments

### `internal/catalog/service.go` — Extract `Walk` (the walk/write split)

**Analog:** itself — `CreateCatalogWithContext` (`service.go:127-203`) and `traverseDirectory` (`service.go:263-414`)

**The exact seam to extract (everything before the final `WriteCatalogFrom` call, `service.go:140-200`):**
```go
// service.go:140-200 — the walk + three-way error classification that
// becomes Walk's body. Copy verbatim; do not re-derive.
tree, err := s.traverseDirectory(ctx, sourcePath, sourcePath, st)
if err != nil {
	var srcErr *SourceUnavailableError
	if errors.As(err, &srcErr) {
		srcErr.Partial = &PartialScan{
			Tree:       tree,
			FilesSeen:  st.filesSeen,
			BytesSeen:  st.bytesSeen,
			ReadErrors: st.readErrorEntries,
		}
		return nil, srcErr
	}
	if ctx.Err() == nil && st.classify() {
		return nil, &SourceUnavailableError{
			SourcePath: sourcePath,
			Partial: &PartialScan{
				Tree: &models.CatalogItem{
					Type: "directory", Name: "./", Size: 0,
					Contents: []*models.CatalogItem{}, Unreadable: true, ReadError: err.Error(),
				},
				FilesSeen:  st.filesSeen,
				BytesSeen:  st.bytesSeen,
				ReadErrors: st.readErrorEntries,
			},
		}
	}
	return nil, fmt.Errorf("failed to traverse directory: %w", err)
}
if err := ctx.Err(); err != nil {
	return nil, err
}
return tree, nil
```
`CreateCatalogWithContext` then reduces to `tree, err := s.Walk(ctx, sourcePath, opts, onProgress); ... return s.WriteCatalogFrom(tree, ...)`. Regression backstop: `TestCreateCatalog_JSONShapeUnchanged` (`service_test.go:373`), source-loss tests (`service_test.go:562,623`) must pass **unmodified**.

**When to use:** re-scan calls `Walk` directly, never `CreateCatalogWithContext`/`WriteCatalogFrom` until a resolution is picked.

### `internal/catalog/service.go` — `MarkUnreadableOnSkip` (the fourth diff state's enabling change)

**Analog:** itself — the two existing skip-and-continue branches

**Branch 1** (`service.go:301-321`, directory-level `os.ReadDir` failure):
```go
entries, err := os.ReadDir(dirPath)
if err != nil {
	st.recordReadError(dirPath, err)
	node := &models.CatalogItem{Type: "directory", Name: displayPath, Size: 0, Contents: []*models.CatalogItem{}}
	if st.classify() {
		node.Unreadable = true
		node.ReadError = err.Error()
		return node, &SourceUnavailableError{SourcePath: st.scanRoot}
	}
	// Root still reachable: today's skip-and-continue behavior, unchanged --
	// an empty, error-free directory node.
	return node, nil   // <-- NEW: gate this "return node, nil" with
	                    //     `if st.opts.MarkUnreadableOnSkip { node.Unreadable = true; node.ReadError = err.Error() }`
	                    //     BEFORE this return, still return nil error (no abort).
}
```
**Branch 2** (`service.go:375-394`, per-child recursive failure):
```go
// Plain single-entry failure (e.g. os.Stat failed for this child). Record
// it and classify against the scan root.
st.recordReadError(childPath, err)
if st.classify() {
	... // (unchanged — root-loss path)
}
// Skip items we can't access -- unchanged byte-for-byte behavior when the
// root is still reachable.
continue   // <-- NEW: the same gate must mark the *child's would-be node*
           //     before it is dropped. Since this branch never builds a
           //     node for the failed child (it just `continue`s past it),
           //     MarkUnreadableOnSkip here requires synthesizing a minimal
           //     unreadable child node and appending it to `contents`
           //     instead of a bare `continue` — this is the one place the
           //     "gate an existing return" framing needs an actual new
           //     node construction, not just a flag check on an existing one.
```
**Where the marker IS already set today (side-by-side reference, unchanged):** `service.go:310-316` (directory root-loss) and `:382-390` (child root-loss) — both require `st.classify() == true`, which the flagged new behavior deliberately does NOT require.

**Options struct extension point:** wherever `Options{WriteHTML, IncludeHidden, HaltOnSourceLoss}` is declared (grep `type Options struct` in `service.go`) — add `MarkUnreadableOnSkip bool`, default false. Re-scan's `app.go` binding sets it true; Create's `CreateCatalog`/`startScan` call sites do not touch it, preserving `false` by construction.

### `internal/catalog/service.go` — mtime capture

**Analog:** itself — `service.go:286-294` (file branch; `info` already in scope from `os.Stat` at `:268`)
```go
if info.Mode().IsRegular() {
	st.filesSeen++
	st.bytesSeen += info.Size()
	st.report(displayPath)
	return &models.CatalogItem{
		Type:    "file",
		Name:    displayPath,
		Size:    info.Size(),
		ModTime: info.ModTime().Unix(), // NEW — zero extra syscalls
	}, nil
}
```
Same `info` is in scope for the directory branch's own node construction (`service.go:405-410`) — add `ModTime: info.ModTime().Unix()` there too.

### `pkg/models/catalog.go` — `ModTime` field + new diff types

**Analog:** itself — the existing `Unreadable`/`ReadError`/`Title` `omitempty` precedent (`catalog.go:26-38`)
```go
// Copy this exact commenting/precedent style for the new field:
// ModTime is the entry's modification time as Unix seconds, captured at
// scan time. omitempty so a catalog written before this field existed
// stays byte-for-byte v2.3.0-shaped; the diff algorithm must treat
// ModTime == 0 as "unknown" (size-only fallback), never as the epoch.
ModTime int64 `json:"modTime,omitempty"`
```
New `DiffResult`/`DiffEntry` types are genuinely new — no existing type to extend; follow the same `omitempty`-on-derived-fields, doc-comment-heavy style as `CatalogItem`/`CreateCatalogResult` (`catalog.go:106-128`) for consistency.

### `internal/catalog/duplicate.go` — Keep-both reuse

**Analog:** itself — `DuplicateCatalog`/`nextCopyRoot`/`isCandidateRootFree` (`duplicate.go:23-120`), used **wholesale, unmodified**

Re-scan's Keep-both resolution calls `nextCopyRoot(dir, root)` directly to resolve the target filename, then `WriteCatalogFrom(newTree, ..., outputRoot: newRoot, ...)` — it does NOT call `DuplicateCatalog` itself (that copies bytes from an existing JSON; re-scan has a freshly-walked tree to write instead). Only the collision-loop half is reused; the copy-bytes half is not applicable.

### `app.go` — `RescanCatalog` binding (walk + diff orchestration)

**Analog:** `startScan` (`app.go:261-389`) for shape — `filepath.Abs`/`EvalSymlinks`/`ContainsPath` validation (`:262-315`), the `scanMu`-guarded one-scan-at-a-time gate (`:317-335`), the deferred cleanup (`:337-343`), and `throttledProgress`/`resolveScanTotal` reuse (`:352-363`).

**Copy this shape:**
```go
// Validation pattern to copy verbatim (app.go:262-296 style):
absSource, err := filepath.Abs(sourcePath)
// ... EvalSymlinks, ContainsPath as startScan already does for its own paths

a.scanMu.Lock()
if a.activeScanCancel != nil {
	a.scanMu.Unlock()
	return nil, fmt.Errorf("a scan is already running")
}
ctx, cancel := context.WithCancel(context.Background())
a.activeScanCancel = cancel
a.scanDone = make(chan struct{})
// NOTE: deliberately do NOT clear/set a.lastPartial/a.lastPartialResult/
// a.lastScanReq here — re-scan must never touch them (see Divergence below).
a.scanMu.Unlock()
```

**Critical divergence from `startScan` (do not copy this part):** `startScan`'s failure branch (`app.go:368-387`) populates `a.lastPartial`/`a.lastScanReq` on a `SourceUnavailableError`. `RescanCatalog` must call `a.catalogService.Walk` directly (never `a.startScan` or `CreateCatalogWithContext`) and must **omit** that assignment block entirely on failure — a failed re-scan just returns the error; the dialog offers Retry/Close only, per `28-UI-SPEC.md`'s Error Step. This is the single highest-risk regression this phase can introduce (see RESEARCH.md Pitfall 2) — flag as its own reviewable diff hunk, not folded into "reuse startScan's shape."

**Cancellation:** `CancelScan`/`cancelActiveScan` (`app.go:391-420`) reused unchanged — no new cancel binding needed.

**Old-tree loading:** `internal/search.Service.LoadCatalog` — already exists, already handles v1/v2 duality, called before/around the walk to build the diff's `oldTree` side. Skipped entirely when `oldTreeAvailable == false` (STATE-03 path).

### `app.go` — `ResolveRescan` binding (overwrite / keep-both / discard)

**Analog:** `DeleteCatalog` (`app.go:993-1026`) for the containment-check shape (`filepath.Abs` → `EvalSymlinks` → `osutil.ContainsPath`, `:998-1012`) — apply the identical sequence to re-validate the *original* catalog's own on-disk path before Overwrite writes to it (RESEARCH.md's Tampering threat: "re-derive the write target from the catalog's own on-disk path, don't trust a renderer-supplied path uncritically").

**Overwrite:** calls `WriteCatalogFrom(newTree, title, outputDir, outputRoot, copyToDirectory, opts)` (`service.go:211`) — the exact same signature Create already uses, with `outputDir`/`outputRoot` derived from the existing catalog's own path the same way `DuplicateCatalog` derives `dir`/`root` from `jsonPath` (`duplicate.go:33-34`).

**Keep-both:** `nextCopyRoot(dir, root)` (`duplicate.go:90-106`) for the target filename, then `WriteCatalogFrom` with that resolved root.

**Discard:** no Go call at all — dialog closes, nothing written (matches `28-UI-SPEC.md`'s "four close paths ... write nothing").

### `app.go` — STATE-03 "Remove from library"

**Analog:** `DeleteCatalog` (`app.go:993-1026`) — **zero new Go code**. The frontend's `UnreadableCatalogPanel` trio calls the existing binding directly, same as `27-UI-SPEC.md`'s `DeleteConfirmDialog` already does elsewhere. Copy the exact call site pattern from wherever `DeleteConfirmDialog` currently invokes `wailsAPI.deleteCatalog(...)`.

### `frontend/src/components/workspace/rescan/RescanDialog.tsx` — shell

**Analog:** `27-UI-SPEC.md`'s established "own class family, not `DialogShell`" precedent (`SettingsDialog`'s `.ws-settings-*` at 660px) — `RescanDialog` follows the identical reasoning at 620px with its own `.ws-rescan-*` class family. No single file is a structural analog for the 3-step flow; closest orchestration analog is `CreateSlideOver.tsx`'s step-machine + `state.scan` subscription pattern (the `EventsOn('scan:progress', ...)` listener — do not add a second listener, subscribe through the same one).

**Reused verbatim, additive optional props (copy this exact extension pattern):**
```typescript
// ScanningBody.tsx:7-9 (existing) — onRunInBackground becomes optional;
// when omitted, its button/helper text simply doesn't render.
export interface ScanningBodyProps {
  scan: ...;
  onRunInBackground?: () => void; // was required; make optional for re-scan's call site
}

// ErrorBody.tsx:5-8 (existing) — writingPartial/onWritePartial become
// optional TOGETHER; when both omitted, "Write partial catalog" doesn't
// render and "Retry scan" is promoted to the primary slot.
export interface ErrorBodyProps {
  scan: ...;
  writingPartial?: boolean;
  onWritePartial?: () => void;
  onRetry: () => void;
  onCloseWithoutWriting: () => void;
  explanation?: string; // NEW prop per 28-UI-SPEC.md — defaults to Create's
                         // existing literal so CreateSlideOver's call site
                         // needs no change
}
```

### `frontend/src/components/workspace/rescan/DiffList.tsx` (NEW, no true analog)

**Explicitly NOT built on:** `frontend/src/components/workspace/TreePane.tsx`'s `useVisibleRows`/`@tanstack/react-virtual` virtualizer. Per `28-UI-SPEC.md`'s Diff List Contract: a diff has no hierarchy (flat `Name` paths, no expand/collapse) and is scale-bounded by what differs, not by catalog size — a plain `<div style={{maxHeight: 200, overflowY: 'auto'}}>` with native scroll is correct here. If row count ever needs virtualizing, `useVirtualizer` (flat mode) is the fallback template, not `useVisibleRows`.

**Row-rendering convention to copy** (glyph + mono path + right-aligned mono size, NOT `TreePane`'s indentation/expand logic): reuse the `+`/`−`/`~`/`!` glyph convention `UnreadableCatalogPanel`'s `!` badge and `ErrorBody`'s round badge already established (same literal glyph, same semantic weight).

### `frontend/src/components/workspace/UnreadableCatalogPanel.tsx:111-113` — STATE-03 trio

**Analog for action 2 ("Open the .html instead"):** `DetailsPanel.tsx`'s `Footer`'s `handleOpenHtml` (`:246-268`) — reuse verbatim (`getCatalogHtmlPath` + `openExternal`).

**Analog for action 3 ("Remove from library"):** the existing `DeleteConfirmDialog` (Phase 27, unchanged) — pass this catalog into it directly, do not build a bespoke library-removal dialog.

**Action 1 ("Re-scan volume"):** opens the new `RescanDialog` at the pick-volume step with `oldTreeAvailable: false` — same component as the other two entry points, different initial prop.

### `frontend/src/components/workspace/DetailsPanel.tsx` — entry points

**Menu item insertion** (`CatalogActions`'s `Menu` items): `Menu.tsx`'s roving-`tabIndex` nav is already order-agnostic over its `items` array (confirmed by `28-UI-SPEC.md`) — reordering to insert "Re-scan volume & diff…" second is a pure data change to the existing items array, not a `Menu.tsx` code change.

**Footer button insertion:** `Footer`'s existing stub comment (`DetailsPanel.tsx:201-212`, "the handoff's third footer action ... is omitted entirely") is the exact insertion point — append using the same `buttonBase` (30px, radius 7px, 12.5px) as the existing "Reveal JSON" button, but outlined + `--dm` text per `28-UI-SPEC.md`.

### `frontend/src/contexts/AppContext.tsx` — `state.rescan` (NEW slice)

**Analog:** the existing `state.scan` slice (`AppContext.tsx:37-39` type, `:147` initial state, `:312-421` reducer cases) — **copy the shape, do not extend the slice itself.**
```typescript
// AppContext.tsx:37-39 (existing, DO NOT MODIFY — copy its pattern instead)
scan: ScanState;

// AppContext.tsx:147 (existing initial state, for reference)
scan: { status: 'idle' },
```
`state.rescan` is a new, separate reducer slice. Live progress (`SCAN_STARTED`/`SCAN_PROGRESS`/`SCAN_FAILED`) continues to drive `state.scan` unchanged, shared with Create — `RescanDialog` subscribes to the same `scan:progress` `EventsOn` listener `CreateSlideOver.tsx` already owns, does not add a second one. Only the terminal outcome diverges: on success, dispatch a new action (e.g. `RESCAN_DIFFED`) that populates `state.rescan.diff` AND resets `state.scan` back to `{ status: 'idle' }` in the same transition — copy the existing terminal-reset pattern at `AppContext.tsx:421` (`return { ...state, scan: { status: 'idle' } }`) as the template for that reset half.

---

## Shared Patterns

### Path containment validation (V4/Tampering mitigation)
**Source:** `app.go:262-315` (`startScan`'s `filepath.Abs` → `filepath.EvalSymlinks` → `osutil.ContainsPath` sequence) and `app.go:998-1012` (`DeleteCatalog`'s identical sequence)
**Apply to:** `RescanCatalog`'s `sourcePath` param, `ResolveRescan`'s re-derivation of the original catalog's own path before Overwrite
```go
absSource, err := filepath.Abs(sourcePath)
resolvedOutput, err := filepath.EvalSymlinks(absOutput)
if ok, err := osutil.ContainsPath(resolvedOutput, dest); err != nil {
	return nil, err
} else if !ok {
	return nil, fmt.Errorf("... escapes the output directory")
}
```

### Crash-safe write (already fully hardened, reused not re-derived)
**Source:** `internal/catalog/atomicwrite.go`'s `WriteFileAtomic`, reached only via `WriteCatalogFrom` (`service.go:211`)
**Apply to:** ACT-07's Overwrite and Keep-both — both call `WriteCatalogFrom`, never a new write primitive.

### `omitempty` field precedent for new/derived data
**Source:** `pkg/models/catalog.go:26-38` (`Title`, `Unreadable`, `ReadError`)
**Apply to:** `CatalogItem.ModTime` — same "absent for the overwhelming majority, byte-parity preserved" doc-comment style.

### One-scan-at-a-time guard
**Source:** `app.go:317-343` (`scanMu`/`activeScanCancel`), `app.go:404-420` (`CancelScan`/`cancelActiveScan`)
**Apply to:** `RescanCatalog` reuses this guard unchanged — no new mutex, no new cancel binding.

### Scan progress event
**Source:** `app.go:188` `throttledProgress` (≤1 emit/200ms), `scan:progress` event name
**Apply to:** `RescanCatalog`'s `onProgress` callback — same throttle, same event, same frontend `EventsOn` listener (do not add a second subscription).

### Diff-glyph-and-color convention
**Source:** `UnreadableCatalogPanel`'s `!` badge, `ErrorBody`'s round badge (established glyph), `27-UI-SPEC.md`'s `--danger`/`--ondanger` tokens
**Apply to:** all four diff-row marks (`+`/`−`/`~`/`!`) and their tile colors, per `28-UI-SPEC.md`'s Color section — `unreadable` deliberately reuses `--danger` (same token as `removed`), differentiated by glyph and grouping, not hue.

---

## No Analog Found

| File | Role | Data Flow | Reason |
|---|---|---|---|
| `internal/catalog/diff.go` (`Compute`, `categorize`, `flatten`) | service | transform | Confirmed via grep this session: no diff/set-comparison code exists anywhere in `internal/`, `pkg/`, or `frontend/src`. Genuinely new — implement per RESEARCH.md's Pattern 3 sketch (path-keyed `map[string]*models.CatalogItem` set comparison), not from an existing analog. `internal/search/flatten.go`'s `LoadCatalogFlat` is a *loose* precedent for the "exclude root, recurse children" flatten convention only — its `FlatNode` output type is not reusable (carries `Depth`/`ParentIdx`/`HasChildren` for tree-pane rendering, drops `ModTime`/`Unreadable`/`ReadError`). |
| `frontend/src/components/workspace/rescan/RescanDialog.tsx` (the 620px 3-step shell itself, not its reused sub-bodies) | component | request-response + streaming | No existing dialog in this codebase is both multi-step AND 620px — `SettingsDialog` (660px) is single-screen tabs, not a step machine; `CreateSlideOver` is a step machine but is a full-height slide-over, not a centered dialog. `28-UI-SPEC.md` explicitly designs it as its own class family for this reason. Use `CreateSlideOver`'s step-machine *state management* pattern and `27-UI-SPEC.md`'s "own class family, not `DialogShell`" *styling* precedent — but there is no single file to copy the whole shell from. |

**Explicitly evaluated and rejected as an analog:** Phase 23's virtualized tree (`TreePane.tsx` + `useVisibleRows` + `@tanstack/react-virtual`). Per `28-UI-SPEC.md`'s own Diff List Contract, it is **not reusable** for the diff list — the diff has no hierarchy to virtualize against (flat paths, no expand/collapse), and diff-list scale is bounded by what differs (small), not by total catalog size (large, the reason the tree pane needs virtualization at all). Use a plain native-scroll `<div>` instead; `useVirtualizer` (flat mode, no `useVisibleRows`) is named as the fallback template only if a future pathological-scale diff proves it necessary.

## Metadata

**Analog search scope:** `internal/catalog/`, `internal/search/`, `pkg/models/`, `app.go`, `frontend/src/components/workspace/` (incl. `create/`), `frontend/src/contexts/AppContext.tsx`, `.github/workflows/`
**Files scanned:** `service.go`, `duplicate.go`, `errors.go`, `catalog.go` (models), `app.go` (full), `flatten.go`, `VolumePicker.tsx`, `ScanningBody.tsx`, `ErrorBody.tsx`, `UnreadableCatalogPanel.tsx`, `DetailsPanel.tsx`, `AppContext.tsx`
**Pattern extraction date:** 2026-08-16
