# Phase 27: Catalog Actions + Watch - Pattern Map

**Mapped:** 2026-08-15
**Files analyzed:** 18 (11 new, 7 modified)
**Analogs found:** 16 / 18 (2 explicitly have no analog — see below)

## File Classification

| New/Modified File | Role | Data Flow | Closest Analog | Match Quality |
|---|---|---|---|---|
| `pkg/models/catalog.go` (add `Title` field) | model | transform | same file, `Unreadable`/`ReadError` fields (:14-20) | exact (self-precedent) |
| `internal/catalog/atomicwrite.go` (add `File.Sync()` + dir fsync) | utility | file-I/O | same file (current impl) | exact (self-modification) |
| `internal/catalog/rename.go` (new) | service | file-I/O / CRUD | `internal/catalog/service.go` `writeJSONFile`/`writeHTMLFile` (:417-497) | exact |
| `internal/catalog/duplicate.go` (new) | service | file-I/O / CRUD | `internal/catalog/service.go` `copyFile` (:606-623) | exact |
| `internal/search/service.go` (`BrowseCatalogs` title-read fix) | service | transform | same file, `<title>` extraction (:187-197) | exact (self-modification) |
| `internal/osutil/trash.go` (new) | utility | file-I/O | `internal/osutil/reveal.go` `ContainsPath` + `RevealInFileManager` (:16-64, :112+) | exact |
| `internal/watch/watcher.go` (new) | service | event-driven | `app.go` `throttledProgress` (:167-189) — closest emit-throttle precedent; no watcher precedent exists | role-match |
| `internal/catalog/atomicwrite_sigkill_test.go` + `testdata/killtarget/main.go` (new) | test | subprocess | none in repo | **no analog** |
| `app.go` (new bindings: `RenameCatalog`, `DuplicateCatalog`, `DeleteCatalog`; watcher wiring) | controller | request-response + event-driven | `app.go` `GetCatalogHtmlPath`/`OpenExternal` (:753-789) for bindings; `throttledProgress` (:167-189) for emit wiring | exact |
| `main.go` (add `OnShutdown` hook) | config | event-driven | same file, existing `OnStartup`/`OnDomReady`/`OnBeforeClose` registration | exact (self-precedent) |
| `frontend/src/components/workspace/Menu.tsx` (new) | component | event-driven | `frontend/src/hooks/useModalBehavior.ts` (whole file) + `CommandPalette.tsx` consumer shape | role-match (no menu precedent; hook + consumer pattern) |
| `frontend/src/components/workspace/DialogShell.tsx` (new) | component | request-response | `frontend/src/components/workspace/SettingsDialog.tsx` (:1-147) | exact |
| `frontend/src/components/workspace/RenameDialog.tsx` (new) | component | request-response | `SettingsDialog.tsx` shell + `CreateSlideOver.tsx` form-field pattern | exact |
| `frontend/src/components/workspace/DeleteConfirmDialog.tsx` (new) | component | request-response | `SettingsDialog.tsx` shell (confirm sub-state is new; no destructive-dialog precedent) | role-match |
| `frontend/src/components/workspace/DetailsPanel.tsx` (wire `⋯` button) | component | event-driven | same file, `OverflowButton` (:38-67) | exact (self-modification) |
| `frontend/src/components/workspace/StatusBar.tsx` (add watching segment) | component | transform | same file, scan segment (:44-77) | exact (self-precedent) |
| `frontend/src/components/workspace/CatalogRail.tsx` (subscribe `catalogs:changed`) | component | event-driven | `CreateSlideOver.tsx` `EventsOn('scan:progress', …)` (:193-198) | exact |
| `frontend/src/services/wailsAPI.ts` (add rename/duplicate/delete/watch bindings) | service | request-response | existing methods in same file (not read this pass; follow the file's own established wrapper shape around generated Wails bindings) | exact (self-precedent) |

## Pattern Assignments

### `internal/catalog/rename.go` (service, file-I/O/CRUD)

**Analog:** `internal/catalog/service.go` (`writeJSONFile` :417-429, `writeHTMLFile` :431-497)

**Core pattern — JSON write** (service.go :417-429):
```go
func (s *Service) writeJSONFile(catalog *models.CatalogItem, path string) (int64, error) {
	jsonBytes, err := json.Marshal(catalog)
	if err != nil {
		return 0, err
	}
	if err := WriteFileAtomic(path, jsonBytes, 0644); err != nil {
		return 0, err
	}
	return int64(len(jsonBytes)), nil
}
```
Rename reuses this exact shape: unmarshal the existing JSON, set `Title`, re-marshal, `WriteFileAtomic`.

**HTML has the title in TWO places — copy both, not one** (service.go :465, :474, :491):
```go
htmlContent := fmt.Sprintf(`...
 <title>%s</title>
...
	<h1>%s</h1><p>
...
</html>`, html.EscapeString(title), html.EscapeString(title), treeStructure, totalSize, dirCount, fileCount)
```
The rename operation's surgical rewrite must locate and replace **both** the `<title>...</title>` content and the `<h1>...</h1>` content, each re-escaped with `html.EscapeString(newTitle)` — mirroring `BrowseCatalogs`'s own `strings.Index`-delimited read pattern (see below), not a full HTML regeneration (tree/counts must stay byte-identical).

**Containment gate before any write** — reuse Pattern from `app.go:753-789` (`GetCatalogHtmlPath`), excerpted under `internal/osutil/trash.go` below; identical shape applies to `RenameCatalog`.

---

### `internal/catalog/duplicate.go` (service, file-I/O/CRUD)

**Analog:** `internal/catalog/service.go` `copyFile` (:606-623)

```go
func (s *Service) copyFile(src, dst string) (int64, error) {
	data, err := os.ReadFile(src)
	if err != nil {
		return 0, err
	}
	if err := WriteFileAtomic(dst, data, 0644); err != nil {
		return 0, err
	}
	return int64(len(data)), nil
}
```
Duplicate calls this verbatim for both `.json` and (if present) `.html`, with the destination filename root suffixed `-copy`, `-copy-2`, `-copy-3` on collision (new collision-loop logic, no existing analog — implement as a small helper that `os.Stat`s each candidate path until one doesn't exist).

---

### `internal/catalog/atomicwrite.go` (utility, file-I/O — self-modification)

**Current implementation (full file), the exact sequence to extend:**
```go
func WriteFileAtomic(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, "storcat-*.tmp")
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
	if err := os.Chmod(tmpPath, perm); err != nil {
		os.Remove(tmpPath)
		return err
	}
	if err := os.Rename(tmpPath, path); err != nil {
		os.Remove(tmpPath)
		return err
	}
	return nil
}
```
**ACT-09 additions (locked):** insert `tmp.Sync()` after the `tmp.Write` call and before `tmp.Close()` — same error-handling shape (`os.Remove(tmpPath); return err` on failure). Parent-directory fsync after the final `os.Rename` is Claude's Discretion per CONTEXT.md/RESEARCH.md Pitfall 7 — if added, open `filepath.Dir(path)` via `os.Open`, call `.Sync()`, close, and treat a platform "not supported" error as non-fatal (this is new code with no existing analog in the repo; follow the same fail-open-on-non-fatal-platform-divergence posture `internal/osutil/reveal.go`'s cross-platform functions use).

---

### `internal/search/service.go` `BrowseCatalogs` title-read fix (service, transform — self-modification)

**Current extraction (the bug)** (:187-197):
```go
if htmlData, err := os.ReadFile(htmlPath); err == nil {
	htmlContent := string(htmlData)
	if startIdx := strings.Index(htmlContent, "<title>"); startIdx != -1 {
		startIdx += 7 // len("<title>")
		if endIdx := strings.Index(htmlContent[startIdx:], "</title>"); endIdx != -1 {
			title = htmlContent[startIdx : startIdx+endIdx]
		}
	}
}
```
Fix: wrap the extracted substring in `html.UnescapeString(...)` before assigning to `title`. Also add a `"html"` import to this file's import block (currently `encoding/json, fmt, os, path/filepath, strings, time`, per RESEARCH.md Supporting Stack). **New precedence rule this phase adds:** when the catalog's own JSON has a non-empty `Title` field, that takes priority over the HTML-derived title before the existing HTML-then-filename fallback chain runs (per CONTEXT.md, "title field... becomes authoritative").

---

### `internal/osutil/trash.go` (utility, file-I/O)

**Analog:** `internal/osutil/reveal.go` — `ContainsPath` (:83-100) and `RevealInFileManager`'s argv-per-element pattern.

**Containment gate — copy exactly:**
```go
func ContainsPath(catalogDir, resolved string) (bool, error) {
	absCatalogDir, err := filepath.Abs(catalogDir)
	if err != nil {
		return false, err
	}
	resolvedCatalogDir, err := filepath.EvalSymlinks(absCatalogDir)
	if err != nil {
		return false, err
	}
	rel, err := filepath.Rel(resolvedCatalogDir, resolved)
	if err != nil {
		return false, err
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return false, nil
	}
	return true, nil
}
```
Every new Trash-bound path must resolve via `filepath.Abs` + `filepath.EvalSymlinks`, then pass this exact `ContainsPath(catalogDir, resolved)` check before reaching `wastebasket.Trash()` — identical shape to `app.go`'s `GetCatalogHtmlPath` (below), and the specific mitigation RESEARCH.md's Pitfall 1 calls out for `wastebasket`'s macOS AppleScript-interpolation risk.

**Argv-per-element invocation precedent** (reveal.go's darwin builder, :16-19):
```go
func revealArgvDarwin(path string) (string, []string) {
	return "open", []string{"-R", path}
}
```
`wastebasket.Trash(paths ...string) error` is itself already an argv/library call (never a shell string on the Go side); the trash helper's job is just the containment gate above plus a thin call-through, mirroring how `RevealInFileManager` composes `revealArgvFor` + `exec.Command(name, args...)`.

---

### `internal/watch/watcher.go` (service, event-driven)

**Analog:** `app.go` `throttledProgress` (:167-189) — the only existing "coalesce events, call back on a timer" precedent in the repo, plus its Wails-runtime-free discipline.

```go
func (a *App) throttledProgress(totalBytes int64) catalog.ProgressCallback {
	var lastEmit time.Time
	return func(u catalog.ProgressUpdate) {
		if a.ctx == nil {
			return
		}
		if time.Since(lastEmit) < 200*time.Millisecond {
			return
		}
		lastEmit = time.Now()
		runtime.EventsEmit(a.ctx, "scan:progress", ScanProgress{ /* ... */ })
	}
}
```
The watcher package must NOT import `runtime` itself (COMPAT-04 — `internal/catalog` and any new watcher package must stay usable from the CLI with no Wails runtime attached). `internal/watch.New(dir string, onChange func()) (*Watcher, error)` takes a plain callback; `app.go` supplies a closure that debounces (or the debounce lives inside `internal/watch` per RESEARCH.md's Code Examples — package layout is Claude's Discretion) and then calls `runtime.EventsEmit(a.ctx, "catalogs:changed")`, guarded by the same `if a.ctx == nil { return }` check as `throttledProgress`. See RESEARCH.md's "Code Examples" section (lines ~354-478) for a full illustrative `Watcher` struct/`loop()`/`debounce()`/`Close()` sketch — flagged there as `[ASSUMED: illustrative shape]`, not copied from an existing file, since no watcher precedent exists anywhere in this repo.

---

### `app.go` new bindings (controller, request-response + event-driven)

**Analog:** `app.go` `GetCatalogHtmlPath` (:753-789, per RESEARCH.md's own excerpt) — the containment-gated binding shape every new path-taking binding must replicate.

```go
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
`RenameCatalog(path, newTitle string) error`, `DuplicateCatalog(path string) (*models.CatalogMetadata, error)`, `DeleteCatalog(jsonPath string, deleteHtml bool) error` all resolve their path(s) the same way, run the same `osutil.ContainsPath` gate, and return a wrapped error on failure (never a panic, never a silent success) — the project's established `fmt.Errorf("<verb> %s: %w", path, err)` shape.

**Watcher lifecycle wiring** — new code, no direct analog; compose from `throttledProgress`'s `a.ctx == nil` guard (above) plus `SetWatchDirectory`'s existing shape in `app.go:632-638` (read `config.WatchDirectory`, call `configManager.SetWatchDirectory`) — extend it to also `Close()` any existing `a.watcher` before starting/stopping. `runtime.EventsEmit` for `catalogs:changed` may ONLY be called from `app.go`, per CONTEXT.md's hard constraint.

---

### `frontend/src/components/workspace/Menu.tsx` (component, event-driven — new primitive)

**Analog:** `frontend/src/hooks/useModalBehavior.ts` (full file, especially :1-50) + its three existing consumers' mount/gate shape (`CommandPalette.tsx`, `CreateSlideOver.tsx`, `SettingsDialog.tsx`).

**Hook contract to consume, unchanged:**
```typescript
export interface ModalBehaviorOptions {
  isOpen: boolean;
  onClose: () => void;
  initialFocusRef?: React.RefObject<HTMLElement | null>;
  scrollLockSelector?: string;
}
export interface ModalBehavior {
  containerRef: React.RefObject<HTMLDivElement>;
}
```
**Consumer wiring pattern (SettingsDialog.tsx :28):**
```typescript
const { containerRef } = useModalBehavior({ isOpen, onClose });
```
UI-SPEC's locked call for the menu specifically:
```typescript
useModalBehavior({
  isOpen,
  onClose,
  initialFocusRef: firstMenuItemRef,
  scrollLockSelector: '.ws-details-overflow',
})
```
**Mount gating — SettingsDialog's `if (!isOpen) return null` is the WRONG shape to copy for the menu.** SettingsDialog is always-mounted (its top comment: "Always mounted by WorkspaceShell... must not be conditionally mounted, because the shared useModalBehavior hook... observes the isOpen: true -> false transition"). The Menu is explicitly UI-SPEC-locked to conditional mounting instead ("rendered only while open... a menu has no exit animation to preserve") — so Menu.tsx should render `null` from its *parent* (not call the hook at all) when closed, rather than mounting always and gating internally like SettingsDialog does. Flagged explicitly so the executor doesn't default to the more common always-mounted analog shape.

**Click-outside is NOT provided by `useModalBehavior`** (confirmed reading the hook — no scrim, no pointer listener). Menu.tsx must add its own `document` `pointerdown` listener, registered only while open, closing on any target outside `containerRef` and the trigger button — genuinely new logic, no existing analog.

**Focus-restore precedent** — same hook, same file, used unchanged by all three; no excerpt needed beyond the interface above (Menu.tsx passes no new option for this).

---

### `frontend/src/components/workspace/DialogShell.tsx`, `RenameDialog.tsx`, `DeleteConfirmDialog.tsx` (component, request-response)

**Analog:** `frontend/src/components/workspace/SettingsDialog.tsx` (full file, 147 lines) — the shell to extract into a shared `DialogShell`.

**Mount + hook wiring** (:20-28):
```typescript
function SettingsDialog({ isOpen, onClose }: SettingsDialogProps) {
  const { state, dispatch } = useAppContext();
  const { containerRef } = useModalBehavior({ isOpen, onClose });
  // ...
  if (!isOpen) return null;
```

**Scrim/header/footer markup shape** (:64-138, structure to generalize into slots):
```tsx
<div className="ws-settings-scrim" onClick={onClose}>
  <div
    ref={containerRef}
    className="ws-settings-panel"
    role="dialog"
    aria-modal="true"
    aria-labelledby="ws-settings-title"
    onClick={(event) => event.stopPropagation()}
  >
    <div className="ws-settings-header">
      <span id="ws-settings-title" className="ws-settings-title">Settings</span>
      <span className="ws-settings-hint mono">⌘,</span>
      <button type="button" className="ws-settings-close-x" aria-label="Close" onClick={onClose}>×</button>
    </div>
    <div className="ws-settings-body">{/* ... */}</div>
    <div className="ws-settings-footer">
      <span className="ws-settings-status mono">{statusText}</span>
      <button type="button" className="ws-settings-close-btn" onClick={onClose}>Close settings</button>
    </div>
  </div>
</div>
```
`DialogShell` extracts this into header/body/footer slot props at 440px width (vs. Settings' 660px, per UI-SPEC), reusing the exact scrim/fade/`role="dialog"`/`aria-modal`/`stopPropagation`-on-panel-click shape verbatim. `RenameDialog` and `DeleteConfirmDialog` each supply their own body/footer content into this shared shell — UI-SPEC and CONTEXT.md both explicitly forbid a second near-duplicate 440px panel implementation.

**No destructive-dialog analog exists** — the delete confirmation's two-sub-state (confirm → error) body/footer swap is new; UI-SPEC's own text notes it follows "the same component, swap the body" pattern `25-UI-SPEC.md`'s create-flow states established (not read this pass — CreateSlideOver.tsx is the pointer if the executor needs that specific precedent).

---

### `frontend/src/components/workspace/CatalogRail.tsx` (component, event-driven — `catalogs:changed` subscription)

**Analog:** `frontend/src/components/workspace/CreateSlideOver.tsx` `EventsOn('scan:progress', …)` (:193-198):
```typescript
useEffect(() => {
  const unsubscribe = EventsOn('scan:progress', (payload: ScanProgress) => {
    dispatch({ type: 'SCAN_PROGRESS', payload });
  });
  return unsubscribe;
}, [dispatch]);
```
CatalogRail's own existing directory-change effect is the refresh target to re-trigger, not a new read path (:50-53):
```typescript
useEffect(() => {
  if (!state.catalogDir) return;
  loadCatalogsForDirectory(state.catalogDir);
}, [state.catalogDir, loadCatalogsForDirectory]);
```
New subscription combines both shapes:
```typescript
useEffect(() => {
  const unsubscribe = EventsOn('catalogs:changed', () => {
    if (state.catalogDir) {
      loadCatalogsForDirectory(state.catalogDir);
    }
  });
  return () => unsubscribe();
}, [state.catalogDir, loadCatalogsForDirectory]);
```
`loadCatalogsForDirectory` (CatalogRail.tsx :19-31) is the existing, unchanged function to call — no bespoke second refresh path per CONTEXT.md.

---

### `frontend/src/components/workspace/DetailsPanel.tsx` `OverflowButton` wiring (component, event-driven — self-modification)

**Current inert button** (:38-67):
```tsx
function OverflowButton() {
  return (
    <button
      type="button"
      className="ws-details-overflow"
      aria-label="Catalog actions"
      style={{ /* 22x22, border, transparent bg */ }}
    >
      <span aria-hidden="true">⋯</span>
    </button>
  );
}
```
This phase adds `aria-haspopup="menu"`, `aria-expanded={isOpen}`, `aria-controls="ws-catalog-actions-menu"` when open, and an `onClick` toggling the `Menu` component's open state — per UI-SPEC's Catalog Actions Menu section. The button's geometry/glyph/`aria-label` are unchanged (do not restyle).

---

### `frontend/src/components/workspace/StatusBar.tsx` watching segment (component, transform — self-precedent)

**Analog:** same file's existing scan segment, the pattern to structurally clone (:44-77):
```tsx
const showScanSegment = scan.status === 'counting' || scan.status === 'scanning';
// ...
{showScanSegment && (
  <button type="button" className="ws-status-scan" onClick={() => dispatch({ type: 'SET_CREATE_OPEN', payload: true })}>
    <span aria-hidden="true">●</span>
    <span className="ws-status-scan-name">{scan.title}</span>
    <span style={{ flexShrink: 0 }}>· {scanPct !== null ? `${scanPct}%` : 'counting…'}</span>
  </button>
)}
```
Conditional-render-vs-omit (`showScanSegment &&`), the `●`-prefixed span pattern, and `aria-hidden="true"` on the bullet are all the shapes to replicate. The watching segment differs per UI-SPEC: it's a `<span>` not a `<button>` (no click destination), and both segments now live inside a new `.ws-status-right` flex wrapper — this file's current `<div className="ws-status-left">`/scan-segment two-child structure (:57-77) is the exact precedent for that wrapper's sibling-grouping shape, generalized to three potential children (left group, watching, scan).

## Shared Patterns

### Containment gate (`osutil.ContainsPath`)
**Source:** `internal/osutil/reveal.go:83-100`
**Apply to:** `RenameCatalog`, `DuplicateCatalog`, `DeleteCatalog` bindings in `app.go`; the new `internal/osutil/trash.go` helper before any path reaches `wastebasket.Trash()`.
```go
func ContainsPath(catalogDir, resolved string) (bool, error) {
	absCatalogDir, err := filepath.Abs(catalogDir)
	// ... EvalSymlinks + filepath.Rel, ".." prefix check ...
}
```

### Wails-runtime-free backend package + single-emitter discipline
**Source:** `app.go:167-189` (`throttledProgress`)
**Apply to:** `internal/watch` (must not import `runtime`), `internal/catalog` rename/duplicate (already runtime-free); `app.go` remains the sole `runtime.EventsEmit` caller for `catalogs:changed`, guarded by `if a.ctx == nil { return }`.

### `WriteFileAtomic` as the one crash-safe write primitive
**Source:** `internal/catalog/atomicwrite.go` (full file)
**Apply to:** every new write path this phase introduces (rename's JSON/HTML rewrite, duplicate's copies) — no new ad hoc `os.WriteFile`/`os.Create` call should appear anywhere in this phase's diff; everything routes through this one function (or its extended, `Sync()`-added form).

### `useModalBehavior` for every new overlay
**Source:** `frontend/src/hooks/useModalBehavior.ts`
**Apply to:** `Menu.tsx`, `DialogShell.tsx` (shared by `RenameDialog`/`DeleteConfirmDialog`) — no second focus-trap/Escape/scroll-lock implementation anywhere in this phase.

### `EventsOn` subscribe-with-cleanup
**Source:** `frontend/src/components/workspace/CreateSlideOver.tsx:193-198`
**Apply to:** `CatalogRail.tsx`'s new `catalogs:changed` subscription — always return the unsubscribe function from the effect.

## No Analog Found

| File | Role | Data Flow | Reason |
|------|------|-----------|--------|
| `internal/catalog/atomicwrite_sigkill_test.go` + `internal/catalog/testdata/killtarget/main.go` | test | subprocess-based crash simulation | No subprocess-launch/SIGKILL test exists anywhere in this repo. RESEARCH.md's own Common Pitfalls #8 and Assumption A3 flag this design as `[LOW confidence, synthesized this session, not run]` — build from the shape sketched there (separate `go build`-able helper binary, injected artificial delay between `tmp.Write()` and `tmp.Close()` as a test-only variant, `cmd.Process.Kill()` from the parent test after polling for the temp file's appearance), not from an existing StorCat pattern. |
| `internal/watch/watcher.go` (fsnotify event loop + debounce internals) | service | event-driven | Closest thing in-repo (`throttledProgress`) only covers the "throttle + emit" half, not "own an OS filehandle, run a goroutine loop selecting on two channels, debounce with a resettable timer, and expose `Close()`." Treated as role-match above for the emit-discipline half; the loop/debounce/Close mechanics themselves have no in-repo precedent — follow RESEARCH.md's Code Examples sketch (explicitly marked `[ASSUMED: illustrative shape]`) as the starting point instead. |

## Metadata

**Analog search scope:** `internal/catalog/`, `internal/osutil/`, `internal/search/`, `pkg/models/`, `app.go`, `main.go`, `frontend/src/hooks/`, `frontend/src/components/workspace/` (top level + `settings/`)
**Files scanned:** `atomicwrite.go`, `service.go` (:1-50, :380-623), `reveal.go` (:1-120), `app.go` (:140-200, :753-789 per RESEARCH.md excerpt), `catalog.go` (:1-20), `search/service.go` (:160-262), `useModalBehavior.ts` (:1-50), `SettingsDialog.tsx` (full), `CreateSlideOver.tsx` (:1-10, :186-206, grep for `EventsOn`/`useModalBehavior`), `DetailsPanel.tsx` (:1-80), `StatusBar.tsx` (full), `CatalogRail.tsx` (:1-60)
**Pattern extraction date:** 2026-08-15
