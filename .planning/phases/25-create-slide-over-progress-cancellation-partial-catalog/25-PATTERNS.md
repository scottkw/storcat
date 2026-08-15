# Phase 25: Create Slide-over + Progress/Cancellation/Partial-Catalog - Pattern Map

**Mapped:** 2026-08-14
**Files analyzed:** 16
**Analogs found:** 14 / 16

## File Classification

| New/Modified File | Role | Data Flow | Closest Analog | Match Quality |
|---|---|---|---|---|
| `internal/catalog/service.go` (add `CreateCatalogWithContext`, extend `traverseDirectory`, `Options`) | service | event-driven (progress callback) + CRUD (file write) | `internal/search/flatten.go` (`LoadCatalogFlat`) / `internal/search/search_indexed.go` (`SearchIndexed`) — "new method beside CLI-shared one" precedent | role-match (precedent for the wrapping shape, not the mutation) |
| `internal/catalog/atomicwrite.go` (new: `writeFileAtomic`) | utility | file-I/O | `internal/config/counts_cache.go:107-135` (`save()`) | exact |
| `internal/catalog/errors.go` (new: sentinel errors) | utility | — | none in-repo (Go idiom, not project-precedented) | no analog |
| `internal/catalog/options.go` (new: `Options{WriteHTML, IncludeHidden}`) | model/config | — | `pkg/models/catalog.go` struct conventions | role-match |
| `internal/catalog/service_test.go` (extend) | test | CRUD | `internal/catalog/service_test.go` (existing, self) / `internal/search/search_indexed_test.go` | exact |
| `internal/volumes/volumes.go` + `volumes_darwin.go`/`_linux.go`/`_windows.go` (new package) | service | batch (enumerate + stat) | `internal/osutil/reveal.go` — per-OS dispatch precedent (parameter-dispatched, NOT build-tagged — see divergence note below) | role-match, deliberate divergence |
| `internal/volumes/volumes_test.go` (new) | test | batch | `internal/catalog/service_test.go` table-driven convention | role-match |
| `app.go` (`StartScan`/`CreateCatalogWithContext` binding, `CancelScan`, `WritePartialCatalog`, `ListVolumes`, `beforeClose` extension, `EventsEmit` throttling) | controller (Wails binding) | request-response + event-driven (progress emit) | `app.go` itself — `CreateCatalog` (63-81), `beforeClose` (204-215), `SelectDirectory` (219-224) | exact (self-extension) |
| `pkg/models/catalog.go` (`Unreadable`, `ReadError` omitempty fields on `CatalogItem`) | model | CRUD | `pkg/models/catalog.go` existing struct | exact |
| `cli/create.go` | UNTOUCHED | — | — | must remain byte-identical (compatibility anchor, not a pattern target) |
| `frontend/src/components/workspace/CreateSlideOver.tsx` (new, shell + state machine) | component | request-response + event-driven | `frontend/src/components/workspace/CommandPalette.tsx` | exact (overlay/multi-state structural precedent) |
| `frontend/src/components/workspace/create/*.tsx` (VolumePicker, OptionsToggles, ScanningBody, ErrorBody, DoneBody — decomposition at Claude's discretion) | component | request-response | `CommandPalette.tsx`'s `palette/PaletteResultList.tsx` sub-component split | role-match |
| `frontend/src/hooks/useModalBehavior.ts` (extend for closing/exit-animation consumer) | hook | event-driven | itself (already written with this phase in mind) | exact |
| `frontend/src/components/workspace/StatusBar.tsx` (add right-aligned scan segment) | component | request-response (derived render) | itself (existing `useMemo` + `<span>` segment pattern) | exact |
| `frontend/src/components/workspace/CatalogRail.tsx` (wire "＋ New" pill onClick) | component | event-driven | itself — `handleChooseDirectory` button (`onClick`) beside the inert pill | exact |
| `frontend/src/contexts/AppContext.tsx` (lift scan state, new actions) | store | CRUD (reducer) | itself — existing `AppAction`/`case` pattern | exact |
| `frontend/src/services/wailsAPI.ts` (add `startScan`, `cancelScan`, `writePartialCatalog`, `listVolumes` wrappers) | service | request-response | itself — `createCatalog`/`searchIndexed`/`selectDirectory` entries | exact |
| `frontend/src/workspace.css` (new keyframes, toggle/badge/progress-bar styles) | config/style | — | existing `ws-palette-scrim-in` keyframe block | role-match |
| `frontend/wailsjs/go/**` | REGENERATED | — | — | not hand-written; `wails generate module` output |

## Pattern Assignments

### `internal/catalog/service.go` (service, mixed)

**Analog:** `internal/search/flatten.go` (`LoadCatalogFlat`) and `internal/search/search_indexed.go` (`SearchIndexed`)

**"New method beside the shared one" pattern** (`internal/search/search_indexed.go:8-29`):
```go
// SearchIndexed is the GUI-only capped sibling of SearchCatalogs, used by
// the ⌘K command palette. It wraps the unmodified SearchCatalogs walk --
// the CLI calls SearchCatalogs directly at cli/search.go:61-62 and is
// untouched by this method's existence
func (s *Service) SearchIndexed(searchTerm, catalogDirectory string) (*models.SearchIndexResult, error) {
	all, err := s.SearchCatalogs(searchTerm, catalogDirectory)
	...
}
```
This phase's twist (flag in the plan): unlike `SearchIndexed`/`LoadCatalogFlat`, which wrap an **unmodified** shared function, `CreateCatalogWithContext` requires *mutating* `traverseDirectory` itself (ctx param, error classification) — `CreateCatalog` becomes the thin wrapper calling the new/changed method, not the reverse. RESEARCH.md's exact wrapper code (already resolved, copy verbatim):
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

**Current code being modified** (`internal/catalog/service.go:17, 74-159`):
```go
// ProgressCallback is called during directory traversal with the current path
type ProgressCallback func(path string)
```
```go
// traverseDirectory recursively builds catalog structure
func (s *Service) traverseDirectory(dirPath, basePath string, onProgress ProgressCallback) (*models.CatalogItem, error) {
	...
	if info.IsDir() {
		entries, err := os.ReadDir(dirPath)
		if err != nil {
			// Return empty directory if we can't read it
			return &models.CatalogItem{Type: "directory", Name: displayPath, Size: 0, Contents: []*models.CatalogItem{}}, nil
		}
		...
		for _, entry := range entries {
			if strings.HasPrefix(entry.Name(), ".") { continue }
			childPath := filepath.Join(dirPath, entry.Name())
			childItem, err := s.traverseDirectory(childPath, basePath, onProgress)
			if err != nil {
				// Skip items we can't access
				continue
			}
			...
		}
	}
}
```
These two silent-skip sites (`service.go:108-117` empty-dir return, `service.go:139-148` bare `continue`) are exactly what Pitfall 2 in RESEARCH.md requires re-probing the scan root to classify terminal-vs-single-entry. Do not touch the hidden-file skip (`strings.HasPrefix(entry.Name(), ".")`) without gating it behind the new `IncludeHidden` option — this is CRT-05's third toggle.

**Error handling pattern:** existing `fmt.Errorf("failed to X: %w", err)` wrapping throughout `CreateCatalog` (lines 30, 37, 43, 57, 60) — keep for the new method.

---

### `internal/catalog/atomicwrite.go` (utility, file-I/O) — NEW

**Analog:** `internal/config/counts_cache.go:107-135` (`(*CountsCache).save`)

**Full pattern to generalize** (already verified in RESEARCH.md's Code Examples):
```go
// internal/config/counts_cache.go:107-135
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
Generalize the signature to `writeFileAtomic(dir, filename string, data []byte) error`, called from both `writeJSONFile` (`internal/catalog/service.go:166-173`, currently `os.WriteFile(path, jsonBytes, 0644)` — direct, non-atomic) and `writeHTMLFile` (`service.go:176+`). `os.CreateTemp(dir, ...)` must use the **same directory** as the final path (not `os.TempDir()`), or rename can fail across filesystems on removable media.

---

### `internal/volumes/` (new package, per-OS build-tagged)

**Analog:** `internal/osutil/reveal.go` — for the *rationale of what NOT to copy*.

**Key divergence to flag in the plan:** `reveal.go` deliberately avoids build tags, dispatching per-OS via `runtime.GOOS` as a plain function parameter within one file:
```go
// internal/osutil/reveal.go:16-19
func revealArgvDarwin(path string) (string, []string) {
	return "open", []string{"-R", path}
}
```
`internal/volumes` **cannot** follow this — `golang.org/x/sys/windows` only compiles under `GOOS=windows`, and `unix.Statfs_t` field types differ per OS (darwin `Bsize uint32` vs linux amd64 `Bsize int64`). Build tags (`//go:build darwin`, etc.) are a hard compiler requirement here, not a style choice — RESEARCH.md's "Pattern 2" section covers this in full.

**Test convention to copy:** `internal/catalog/service_test.go` — table-driven `*_test.go` beside source (see Testing section below).

---

### `app.go` (controller / Wails binding)

**Analog:** self — extend the existing three functions in place.

**Current `CreateCatalog` binding** (`app.go:63-82`, to be superseded/extended):
```go
func (a *App) CreateCatalog(title string, directoryPath string, outputName string, copyToDirectory string) (*models.CreateCatalogResult, error) {
	absPath, err := filepath.Abs(directoryPath)
	if err != nil {
		return nil, err
	}
	// Progress callback (could be used to send progress to frontend in future)
	progressCallback := func(path string) {
		// For now, we don't send progress updates
		// In the future, we could use Wails events to send updates to frontend
	}
	result, err := a.catalogService.CreateCatalog(title, absPath, outputName, copyToDirectory, progressCallback)
	if err != nil {
		return nil, err
	}
	return result, nil
}
```
The literal comment "could be used to send progress to frontend in future" is this phase's target.

**Current `beforeClose`** (`app.go:203-215`) — extend, don't replace:
```go
func (a *App) beforeClose(ctx context.Context) bool {
	if a.configManager == nil {
		return false
	}
	cfg := a.configManager.Get()
	if cfg != nil && cfg.WindowPersistenceEnabled {
		w, h := runtime.WindowGetSize(ctx)
		x, y := runtime.WindowGetPosition(ctx)
		_ = a.configManager.SetWindowSize(w, h)
		_ = a.configManager.SetWindowPosition(x, y)
	}
	return false // false = allow close
}
```
CRT-13 needs a new branch checked *before* the existing window-persistence logic: if `a.activeScanCancel != nil`, cancel + bounded wait, return `true` (prevent) on the first call, then after the walk goroutine actually stops, call `runtime.Quit(ctx)` — RESEARCH.md flags this exact sequence as MEDIUM confidence / needing live verification (Assumption A2).

**Current `SelectDirectory`** (`app.go:219-224`) — reused as-is for CRT-03's folder picker; no changes needed.

**New cancellation-handle pattern to add** (RESEARCH.md Pattern 1, copy verbatim as the starting shape):
```go
type App struct {
    ctx               context.Context
    mu                sync.Mutex
    activeScanCancel  context.CancelFunc
    lastPartialScan   *partialScanResult
}

func (a *App) CancelScan() {
    a.mu.Lock()
    defer a.mu.Unlock()
    if a.activeScanCancel != nil {
        a.activeScanCancel()
    }
}
```

**Progress-throttling pattern** (new, RESEARCH.md Code Examples — the only place `EventsEmit` may appear, per COMPAT-04):
```go
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

**Error handling pattern:** bare `return nil, err` throughout `app.go`'s bound methods (no wrapping) — matches project convention of returning raw errors across the Wails boundary for `extractErrorMessage` to unwrap on the frontend side.

---

### `internal/catalog/service_test.go` (test, table-driven)

**Analog:** itself (existing conventions) — read structure, do not duplicate wholesale; extend with new `Test*` funcs per RESEARCH.md's Test Map (`TestCreateCatalogWithContext_Cancel`, `TestTraverseDirectory_TerminalError`, `TestWritePartialCatalog_Marker`, `TestCreateCatalog_JSONShapeUnchanged`).

**Secondary analog:** `internal/search/search_indexed_test.go` — fixture construction and boundary-case table style for the new package (`internal/volumes/volumes_test.go`).

---

### `frontend/src/components/workspace/CreateSlideOver.tsx` (component, overlay)

**Analog:** `frontend/src/components/workspace/CommandPalette.tsx`

**Always-mounted overlay pattern + doc comment to replicate** (`CommandPalette.tsx:19-23`):
```tsx
// Always mounted by WorkspaceShell and returns null when closed -- it must
// not be conditionally mounted, because the shared useModalBehavior hook
// below observes the isOpen: true -> false transition to release scroll
// lock and restore focus, and Phase 25's animated exit depends on the same
// contract.
function CommandPalette({ isOpen, onClose }: CommandPaletteProps) {
```
CreateSlideOver's version must add the local `closing` flag UI-SPEC.md's Shell contract requires (CommandPalette itself has no exit-animation state — this is the one place CreateSlideOver's shell diverges from its analog).

**Hook consumption pattern** (`CommandPalette.tsx:47`):
```tsx
const { containerRef } = useModalBehavior({ isOpen, onClose, initialFocusRef: inputRef });
```
UI-SPEC.md is explicit: call with the **real** `isOpen`, never `isOpen || closing`.

**Stale-response guard pattern** (`CommandPalette.tsx:36-40, 76-79`) — reusable for any Wails call issued from the slide-over that could be superseded (e.g. re-enumerating volumes on reopen):
```tsx
const requestIdRef = useRef(0);
...
const requestId = ++requestIdRef.current;
wailsAPI.searchIndexed(trimmed, catalogDir).then((result) => {
  if (requestId !== requestIdRef.current) return; // stale response, dropped
  ...
});
```

**Sub-component decomposition analog:** `frontend/src/components/workspace/palette/PaletteResultList.tsx` — CreateSlideOver's five body states (form/scanning/error/done, form's sub-sections) should follow the same "shell component + focused sub-components per concern" split rather than one large file.

---

### `frontend/src/hooks/useModalBehavior.ts` (hook, extend)

**Analog:** itself — already written with this exact consumer in mind (`useModalBehavior.ts:6-14`):
```ts
/**
 * Written for Phase 25's animated 260ms slide-over exit, not just this
 * phase's palette -- Phases 25, 26 and 27 import this hook rather than
 * reimplementing any of these four behaviors (24-CONTEXT.md). The single
 * decision that makes that work: the effect below is keyed on `[isOpen]`
 * alone, so its cleanup fires on the true->false transition while the
 * consumer is still mounted and still animating out, not only at unmount.
 */
```
No behavior change to the hook itself should be required — the "extension" is CreateSlideOver's consumption pattern (real `isOpen`, local `closing` flag), not new hook logic. If a genuine hook change is needed, keep the change additive (new optional param), never restructure the `[isOpen]`-keyed effect.

---

### `frontend/src/components/workspace/StatusBar.tsx` (component, extend)

**Analog:** itself — existing derived-render `useMemo` + flex segment pattern (`StatusBar.tsx:1-49`):
```tsx
function StatusBar() {
  const { state } = useAppContext();
  const { catalogCount, filesIndexed, totalBytes, partial } = useMemo(() => { ... }, [state.catalogs]);
  return (
    <div className="ws-status mono">
      <span style={{ flexShrink: 0 }}>{catalogCount} catalogs</span>
      ...
    </div>
  );
}
```
New fourth segment follows the same `<span style={{ flexShrink: 0 }}>` shape, conditionally rendered from lifted `AppContext` scan state (per UI-SPEC.md's Background Handoff Contract), added to a `.ws-status` container whose `justify-content` changes to `space-between` in `workspace.css`.

---

### `frontend/src/components/workspace/CatalogRail.tsx` ("＋ New" pill wiring)

**Analog:** itself — the adjacent `handleChooseDirectory` button in the same file (`CatalogRail.tsx:127-135`) shows the onClick convention to copy onto the currently-inert pill (`CatalogRail.tsx:94-122`):
```tsx
{/* Renders, hover-styled, and stays inert -- its target (the create
    slide-over) is Phase 25. Never attach a handler here (RAIL-06). */}
<button type="button" className="ws-new-pill" ...>
  ...
  New
</button>
```
This phase removes that comment and adds `onClick={onOpenCreate}` (prop threaded from `WorkspaceShell`, or a dispatched `AppContext` action opening the slide-over) plus the UI-SPEC's disabled-while-scanning state (`aria-disabled`, `title="A scan is already running — open it from the status bar."`).

---

### `frontend/src/contexts/AppContext.tsx` (store, extend)

**Analog:** itself — existing `AppAction` union + `case` reducer pattern (`AppContext.tsx:32-48, 71+`):
```ts
type AppAction =
  | { type: 'SET_DENSITY'; payload: Density }
  | { type: 'SET_CATALOG_DIR'; payload: string }
  | { type: 'SET_CATALOGS'; payload: models.CatalogMetadata[] }
  ...
```
New actions for lifted scan state (e.g. `SET_SCAN_STATE`, `SET_SCAN_PROGRESS`) follow the same `UPPER_SNAKE_CASE` string-type convention, each handled in its own `case` branch in the reducer switch.

---

### `frontend/src/services/wailsAPI.ts` (service, extend)

**Analog:** itself — every existing entry routes through the shared error helper (`wailsAPI.ts:28-44`):
```ts
function extractErrorMessage(error: any): string {
  if (typeof error === 'string') return error;
  return error?.message || 'Unknown error';
}
function wailsError(error: any): { success: false; error: string } {
  return { success: false, error: extractErrorMessage(error) };
}
export const wailsAPI = {
  createCatalog: async (title: string, directoryPath: string, outputName: string, copyToDirectory: string) => {
    try {
      ...
    } catch (error) {
      return wailsError(error);
    }
  },
  ...
};
```
New entries (`startScan`, `cancelScan`, `writePartialCatalog`, `listVolumes`) must follow this exact try/catch → `wailsError(error)` shape — never a bespoke catch block.

---

## Shared Patterns

### Atomic file write (temp + rename, same directory)
**Source:** `internal/config/counts_cache.go:107-135`
**Apply to:** `internal/catalog/atomicwrite.go`'s new `writeFileAtomic`, consumed by `writeJSONFile`/`writeHTMLFile`, and by the partial-catalog write path.

### "New method beside the CLI-shared one," CLI wrapper unchanged
**Source:** `internal/search/search_indexed.go`, `internal/search/flatten.go`
**Apply to:** `internal/catalog/service.go`'s `CreateCatalog`/`CreateCatalogWithContext` split. `cli/create.go` must not be edited at all.

### Wails error unwrapping
**Source:** `frontend/src/services/wailsAPI.ts` (`extractErrorMessage`/`wailsError`)
**Apply to:** every new `wailsAPI` entry this phase adds.

### Overlay always-mounted + shared modal-behavior hook
**Source:** `frontend/src/components/workspace/CommandPalette.tsx`, `frontend/src/hooks/useModalBehavior.ts`
**Apply to:** `CreateSlideOver.tsx`'s shell.

### Go table-driven tests beside source
**Source:** `internal/catalog/service_test.go`, `internal/search/search_indexed_test.go`
**Apply to:** all new/extended `*_test.go` files (`service_test.go`, `volumes_test.go`).

## No Analog Found

| File | Role | Data Flow | Reason |
|---|---|---|---|
| `internal/catalog/errors.go` (sentinel errors: `ErrScanCancelled`, `ErrVolumeVanished`) | utility | — | No prior sentinel-error file exists in this codebase; follow standard Go idiom (`var ErrX = errors.New(...)`, checked with `errors.Is`) rather than a project precedent — RESEARCH.md's Open Question 2 flags the exact shape as still to be finalized during planning. |
| `frontend/src/components/workspace/create/*.tsx` (new sub-component files: VolumePicker, ToggleRow, ScanningLog, etc.) | component | request-response | No prior toggle-switch or round-badge component exists in this codebase (`25-UI-SPEC.md` explicitly notes this is the first phase to need a toggle control) — build per the UI-SPEC's literal CSS spec, using `CommandPalette`'s sub-component split (`palette/PaletteResultList.tsx`) only for file-organization precedent, not visual pattern. |

## Metadata

**Analog search scope:** `internal/catalog/`, `internal/search/`, `internal/config/`, `internal/osutil/`, `app.go`, `frontend/src/components/workspace/`, `frontend/src/hooks/`, `frontend/src/services/`, `frontend/src/contexts/`
**Files scanned:** ~14 read directly (service.go, search_indexed.go, flatten.go, counts_cache.go, reveal.go, app.go, useModalBehavior.ts, CommandPalette.tsx, StatusBar.tsx, CatalogRail.tsx, wailsAPI.ts, AppContext.tsx, service_test.go conventions, PATTERNS drawn also from 25-RESEARCH.md's own verified Code Examples)
**Pattern extraction date:** 2026-08-14
