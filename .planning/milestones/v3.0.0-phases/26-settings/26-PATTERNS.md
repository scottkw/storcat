# Phase 26: Settings - Pattern Map

**Mapped:** 2026-08-15
**Files analyzed:** 14
**Analogs found:** 12 / 14

## File Classification

| New/Modified File | Role | Data Flow | Closest Analog | Match Quality |
|--------------------|------|-----------|-----------------|----------------|
| `internal/config/config.go` (modify — 8 new fields + Set*/Get*) | model/config | CRUD | itself (`SetTheme`/`SetSidebarPosition`/`SetWindowPersistence`) | exact (self-extension) |
| `internal/config/config_test.go` (modify — table-driven cases) | test | CRUD | itself (`TestSetWindowPosition`, `newTestManager`) | exact |
| `app.go` (modify — new bindings + containment on `GetCatalogHtmlPath`/`OpenExternal`) | controller/binding | request-response | itself (`SetTheme`, `RevealInFileManager`, `GetCatalogHtmlPath`) | exact |
| `internal/osutil` (reused, not modified) | utility | transform | `internal/osutil/reveal.go`'s `ContainsPath` | exact (reuse, no new file) |
| `frontend/src/components/workspace/SettingsDialog.tsx` (new) | component/dialog | request-response | `frontend/src/components/workspace/CommandPalette.tsx` | exact (dialog-shell mount pattern) |
| `frontend/src/components/workspace/settings/SegmentedControl.tsx` (new) | component | event-driven | `frontend/src/components/workspace/create/OptionsToggles.tsx` (`ToggleRow`) | role-match (new shared control, same "plain markup, ARIA-role, onClick+onKeyDown" shape) |
| `frontend/src/components/workspace/settings/ThemeGrid.tsx` (new, optional decomposition) | component | CRUD (render + select) | `frontend/src/themes.ts` (data) + `OptionsToggles.tsx` (control markup shape) | partial — no existing grid-of-cards analog; compose from data shape + control conventions |
| `frontend/src/components/workspace/settings/CatalogSettingsSection.tsx` (new, optional) | component | CRUD | `frontend/src/components/workspace/create/OptionsToggles.tsx` full file | exact (toggle rows) + `CatalogRail.tsx` (directory chip + picker) |
| `frontend/src/components/workspace/create/OptionsToggles.tsx` (modify — export `ToggleRow`) | component | event-driven | itself | exact |
| `frontend/src/components/workspace/Toolbar.tsx` (modify — gear/theme-chip `onClick`) | component | event-driven | itself (`onOpenSearch` prop wiring on the search button) | exact |
| `frontend/src/components/workspace/WorkspaceShell.tsx` (modify — mount `SettingsDialog`, ⌘, listener, overlay mutual exclusion) | component/shell | event-driven | itself (⌘K listener + `CommandPalette`/`CreateSlideOver` mount + mutual-exclusion effects) | exact |
| `frontend/src/App.tsx` (modify — remove `CatalogModal` import/state/listener) | component | event-driven | itself (current `openCatalogModal` wiring being deleted) | exact |
| `frontend/src/components/CatalogModal.tsx` (DELETE) | component | file-I/O | n/a — deletion, no replacement | n/a |
| `frontend/src/contexts/AppContext.tsx` (modify — density/railSide sourced from config) | store/reducer | event-driven | itself (`SET_DENSITY`/`SET_RAIL_SIDE` existing actions) | exact |
| `frontend/src/services/wailsAPI.ts` (modify — 8 new wrapper fns + updated `getCatalogHtmlPath`/`openExternal` signatures) | service | request-response | itself (`setTheme`, `setSidebarPosition`, `getCatalogHtmlPath` wrappers) | exact |
| `frontend/src/components/workspace/DetailsPanel.tsx` (modify — `handleOpenHtml` fail-closed guard + updated call args) | component | request-response | itself (`handleReveal`'s existing `catalogDir` guard, same file) | exact |
| `frontend/src/components/workspace/CatalogRail.tsx` (modify — fold `storcat-catalog-directory` into config-backed migration) | component | CRUD | itself (current `CATALOG_DIR_STORAGE_KEY` localStorage read/write) | exact |

## Pattern Assignments

### `internal/config/config.go` (model/config, CRUD)

**Analog:** itself — `internal/config/config.go:122-157` (four existing `Set*`/`Get*` methods)

**Core mechanical pattern** (verified, lines 122-157):
```go
// SetTheme updates theme setting
func (m *Manager) SetTheme(theme string) error {
	m.config.Theme = theme
	return m.Save()
}

// SetWindowPersistence updates the window state persistence toggle
func (m *Manager) SetWindowPersistence(enabled bool) error {
	m.config.WindowPersistenceEnabled = enabled
	return m.Save()
}

// GetWindowPersistence returns whether window state persistence is enabled
func (m *Manager) GetWindowPersistence() bool {
	return m.config.WindowPersistenceEnabled
}
```
Every one of the 8 new setters (`SetDensity`, `SetRailSide`, `SetCatalogDirectory`, `SetDefaultFilenameRoot`, `SetWriteHTML`, `SetCopyToSecondary`, `SetSecondaryDirectory`, `SetWatchDirectory`) is this identical 3-line shape: mutate field, `return m.Save()`. No debounce, no batching (CONTEXT.md locked decision).

**Struct + JSON tag pattern** (lines 9-18):
```go
type Config struct {
	Theme                   string `json:"theme"`
	SidebarPosition         string `json:"sidebarPosition"`
	WindowWidth             int    `json:"windowWidth"`
	...
	WindowPersistenceEnabled bool  `json:"windowPersistenceEnabled"`
}
```
Add the 8 new fields with this same tag style; also add each to `DefaultConfig()` (lines 21-31) with a sane zero/default value (e.g. `Density: "comfortable"`, `WriteHTML: true`).

**Leave `SidebarPosition`/`SetSidebarPosition` (lines 12, 129-132) untouched** — dead but out of scope per CONTEXT.md's discretion; add a one-line comment noting it predates and is distinct from the new `RailSide` field.

**Migration hook:** No existing analog for a "migrate from elsewhere" method — build as a new `Manager` method (e.g. `MigrateFromLocalStorage(values map[string]string) error` invoked once from a new `App` binding, or entirely on the frontend by reading `localStorage` and calling the new `Set*` bindings once, gated by a `migrated` marker field). Recommend the latter (frontend-driven) since `localStorage` is only readable from the frontend; add a boolean `SettingsMigrated bool` field to `Config` (same struct-field shape as above) as the marker.

---

### `internal/config/config_test.go` (test, CRUD)

**Analog:** itself — `internal/config/config_test.go:9-30` (`newTestManager` fixture), `TestSetWindowPosition`/`TestDefaultConfig_WindowFields` pattern.

```go
func newTestManager(t *testing.T) *Manager {
	t.Helper()
	dir, err := os.MkdirTemp("", "storcat-config-test-*")
	...
	m := &Manager{configPath: configPath, config: DefaultConfig()}
	if err := m.Save(); err != nil { t.Fatalf(...) }
	return m
}
```
Add one `TestSet<Field>`/`TestGet<Field>` pair per new field, following the exact table-driven or direct-assert style already used for `TestSetWindowPosition`.

---

### `app.go` (controller/binding, request-response)

**Analog:** itself — `app.go:549-595` (existing `Set*` bindings), `app.go:696-705` (`RevealInFileManager`), `app.go:674-694` (`GetCatalogHtmlPath`/`OpenExternal` — to be modified).

**Binding wrapper pattern** (lines 549-555):
```go
// SetTheme saves the theme preference
func (a *App) SetTheme(theme string) error {
	if a.configManager == nil {
		return nil
	}
	return a.configManager.SetTheme(theme)
}
```
Every new binding (`SetDensity`, `SetRailSide`, `SetCatalogDirectory`, `SetDefaultFilenameRoot`, `SetWriteHTML`, `SetCopyToSecondary`, `SetSecondaryDirectory`, `SetWatchDirectory`) follows this identical nil-guard + delegate shape.

**Containment-gate pattern to graft onto `GetCatalogHtmlPath`/`OpenExternal`** (current unguarded state, lines 674-694):
```go
// Current (unguarded):
func (a *App) GetCatalogHtmlPath(catalogPath string) (string, error) { ... }
func (a *App) OpenExternal(url string) { runtime.BrowserOpenURL(a.ctx, url) }
```
Target shape, mirroring `RevealInFileManager` (lines 696-705):
```go
// RevealInFileManager asks the operating system to reveal path ... catalogDir
// is the frontend's currently configured catalog directory; internal/osutil
// rejects any path that does not resolve inside it ...
func (a *App) RevealInFileManager(path string, catalogDir string) error {
	return osutil.RevealInFileManager(path, catalogDir)
}
```
Add a `catalogDir string` second parameter to both `GetCatalogHtmlPath` and `OpenExternal`, and call `osutil.ContainsPath(catalogDir, resolvedPath)` before doing anything with the path (Stat / `BrowserOpenURL`), returning/rejecting on `!ok || err != nil` exactly as `RevealInFileManager`'s underlying `osutil` implementation does.

---

### `internal/osutil.ContainsPath` (utility, transform) — reused verbatim

**Source:** `internal/osutil/reveal.go:93-110` (full function, verified):
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
Do not reimplement — import and call directly from `app.go`'s updated `GetCatalogHtmlPath`/`OpenExternal`. Its own doc comment (line 90-92) already states it was "Exported: it is now also the write-path containment gate ... not only the reveal read gate" — this phase is its next intended consumer.

---

### `frontend/src/components/workspace/SettingsDialog.tsx` (component/dialog, request-response)

**Analog:** `frontend/src/components/workspace/CommandPalette.tsx:1-48` + `frontend/src/hooks/useModalBehavior.ts` (full file).

**Imports pattern** (`CommandPalette.tsx:1-7`):
```typescript
import { useEffect, useRef, useState } from 'react';
import { useAppContext } from '../../contexts/AppContext';
import { wailsAPI } from '../../services/wailsAPI';
import { useModalBehavior } from '../../hooks/useModalBehavior';
```

**Always-mounted, isOpen-gated shape** (`CommandPalette.tsx:9-23`, doc comment is load-bearing):
```typescript
export interface CommandPaletteProps {
  isOpen: boolean;
  onClose: () => void;
}
// Always mounted by WorkspaceShell and returns null when closed -- it must
// not be conditionally mounted, because the shared useModalBehavior hook
// below observes the isOpen: true -> false transition ...
function CommandPalette({ isOpen, onClose }: CommandPaletteProps) {
```
`SettingsDialog` takes the identical `{ isOpen, onClose }` prop shape, is always mounted by `WorkspaceShell`, and returns `null` when `!isOpen` (UI-SPEC: no animated exit needed, simpler than `CreateSlideOver`'s `closing`-flag model — just the palette's plain boolean gate).

**Modal-behavior consumption** (`CommandPalette.tsx:44-48`):
```typescript
const { containerRef } = useModalBehavior({ isOpen, onClose, initialFocusRef: inputRef });
```
`SettingsDialog` calls `useModalBehavior` the same way — no bespoke focus-trap/Escape/scroll-lock implementation (PLT-07 standing constraint, confirmed in `useModalBehavior.ts`'s own doc comment: "Phases 25, 26 and 27 import this hook rather than reimplementing").

**Mount site pattern** (`WorkspaceShell.tsx:104-105`):
```tsx
<CommandPalette isOpen={paletteOpen} onClose={() => setPaletteOpen(false)} />
<CreateSlideOver isOpen={state.createOpen} onClose={() => dispatch({ type: 'SET_CREATE_OPEN', payload: false })} />
```
Add `<SettingsDialog isOpen={settingsOpen} onClose={() => setSettingsOpen(false)} />` alongside these two, same shape.

---

### `frontend/src/components/workspace/settings/SegmentedControl.tsx` (new shared control, event-driven)

**Analog:** `frontend/src/components/workspace/create/OptionsToggles.tsx:28-76` (`ToggleRow`) — closest existing "small interactive control with ARIA role + click/keydown" shape, even though the ARIA role differs (`switch` vs. `radio`/`radiogroup` per UI-SPEC).

**Core pattern to adapt** (verbatim, lines 41-58):
```tsx
function ToggleRow({ checked, label, note, noteClassName, disabled, onToggle, onNoteClick }: ToggleRowProps) {
  return (
    <div
      className={`ws-create-toggle-row${disabled ? ' ws-create-toggle-row-disabled' : ''}`}
      role="switch"
      aria-checked={checked}
      aria-disabled={disabled || undefined}
      tabIndex={disabled ? -1 : 0}
      onClick={() => { if (!disabled) onToggle(); }}
      onKeyDown={(event) => {
        if (disabled) return;
        if (event.key === 'Enter' || event.key === ' ') {
          event.preventDefault();
          onToggle();
        }
      }}
    >
      ...
    </div>
  );
}
```
`SegmentedControl` follows the same "plain hand-built markup, explicit ARIA attributes, onClick + onKeyDown handling its own key set" convention, adapted per UI-SPEC to `role="radiogroup"` on the container / `role="radio"` + `aria-checked` per segment, roving `tabIndex`, and `ArrowLeft`/`ArrowRight` (not `Enter`/`Space`) to move selection. No new library — matches this project's "first toggle-switch control ... plain markup ... no package added" precedent (`OptionsToggles.tsx:38-40` comment).

---

### `frontend/src/components/workspace/create/OptionsToggles.tsx` (modify — export `ToggleRow`)

**Analog:** itself, full file (151 lines, read this session).

Currently `ToggleRow` (lines 28-76) is a module-private function, not exported; `export default OptionsToggles` is the only export (line 150). Change: add `export` to the `ToggleRow` function declaration (or add a named export) so `CatalogSettingsSection.tsx` can import it directly, per UI-SPEC's explicit "build one `ToggleRow` component, shared between Create and Settings" resolution and CONTEXT.md's `frontend/src/components/workspace/create/OptionsToggles.tsx` reusable-asset callout. Use the Settings geometry values already locked in `25-UI-SPEC.md` (`30×17px` track / `13px` knob) — do not introduce a second toggle markup.

---

### `frontend/src/components/workspace/settings/CatalogSettingsSection.tsx` (new, optional decomposition, CRUD)

**Analog:** `OptionsToggles.tsx` (toggle rows + secondary-directory picker pattern, lines 93-117) and `CatalogRail.tsx` (directory chip + `SelectDirectory` picker pattern, `CATALOG_DIR_STORAGE_KEY` at line 7).

**Directory-picker pattern to reuse** (`OptionsToggles.tsx:93-108`, verbatim):
```typescript
async function handleToggleSecondary() {
  if (values.copyToSecondary) {
    onValuesChange({ ...values, copyToSecondary: false });
    return;
  }
  if (secondaryDir) {
    onValuesChange({ ...values, copyToSecondary: true }); // reuse, no dialog
    return;
  }
  const result = await wailsAPI.selectDirectory();
  if (!result.success || !result.path) return; // cancelled -- a declined choice, not an error
  safeSetItem(SECONDARY_DIR_STORAGE_KEY, result.path);
  onSecondaryDirChange(result.path);
  onValuesChange({ ...values, copyToSecondary: true });
}
```
The Settings "Catalog directory" row and "Copy to secondary" toggle both follow this identical shape: call `wailsAPI.selectDirectory()`, treat cancellation (`!result.path`) as a no-op (not an error), then write through both `safeSetItem` (localStorage boot cache) and the new config `Set*` binding in the same handler (per RESEARCH.md's Pitfall 1 write-through-cache resolution).

---

### `frontend/src/components/workspace/Toolbar.tsx` (modify, event-driven)

**Analog:** itself — current `onOpenSearch` prop wiring (lines 5, 13, 63).

Currently the gear button (lines 153-178) and theme chip (lines 136-151) render with no `onClick`. Wiring pattern to copy from the search button (line 59-63):
```tsx
<button
  type="button"
  className="no-drag ws-search"
  aria-label="Search every catalog"
  onClick={onOpenSearch}
  ...
```
Add `onOpenSettings: () => void` to `ToolbarProps` (alongside existing `onOpenSearch`), then `onClick={onOpenSettings}` on both the gear button and the theme chip button — same one-prop-per-trigger convention already established.

---

### `frontend/src/components/workspace/WorkspaceShell.tsx` (modify, event-driven)

**Analog:** itself — ⌘K listener (lines 61-73) and overlay mutual-exclusion effects (lines 78-80).

**⌘, listener to add, mirroring the ⌘K listener verbatim** (lines 61-73):
```typescript
useEffect(() => {
  const onKeyDown = (event: KeyboardEvent) => {
    if ((event.metaKey || event.ctrlKey) && event.key.toLowerCase() === 'k') {
      event.preventDefault();
      setPaletteOpen((open) => {
        if (!open) dispatch({ type: 'SET_CREATE_OPEN', payload: false });
        return open ? open : true;
      });
    }
  };
  window.addEventListener('keydown', onKeyDown);
  return () => window.removeEventListener('keydown', onKeyDown);
}, [dispatch]);
```
New `⌘,`/`Ctrl+,` listener follows this identical `event.preventDefault()`-unconditional, functional-state-update (second press is a no-op) shape, additionally closing the palette/create-slide-over per the UI-SPEC's overlay-mutual-exclusion rules (scanning state = no-op, matching the existing "disabled while a scan is running" precedent already used for `+New`).

**Mutual-exclusion effect pattern** (lines 78-80):
```typescript
useEffect(() => {
  if (state.createOpen) setPaletteOpen(false);
}, [state.createOpen]);
```
Add the equivalent for Settings: opening Settings closes the palette/create slide-over; opening either of those (if UI-SPEC requires) closes Settings.

**Density re-apply effect** (lines 30-32) is the existing analog for `SET_RAIL_SIDE`-driven behavior already wired — no change needed here beyond confirming Settings dispatches into the same `state.density`/`state.railSide` reducer fields.

---

### `frontend/src/App.tsx` (modify — delete `CatalogModal` wiring)

**Analog:** itself — current dead wiring to remove (verified via grep this session):
```
2:  import { ConfigProvider, theme as antdTheme } from 'antd';
6:  import CatalogModal from './components/CatalogModal';
13:  const [catalogModalVisible, setCatalogModalVisible] = useState(false);
14:  const [catalogModalPath, setCatalogModalPath] = useState<string | null>(null);
37:  window.addEventListener('openCatalogModal', handleOpenCatalog as EventListener);
45-47: handleCloseCatalogModal
62-66: <CatalogModal ... />
```
Remove lines 6, 13-14, the `openCatalogModal` listener registration/cleanup (~lines 33-37, 41), `handleCloseCatalogModal` (45-48), and the `<CatalogModal>` render (62-66). **Keep** the `ConfigProvider`/`antdTheme` import and usage (line 2, 51-58) — `antd` remains a live dependency for `ConfigProvider`; do not remove it (RESEARCH.md Pitfall 3). **Keep** the `themeChange` listener (line 29) — it is the live, load-bearing path, not part of this deletion.

---

### `frontend/src/contexts/AppContext.tsx` (modify, event-driven)

**Analog:** itself — existing `SET_DENSITY`/`SET_RAIL_SIDE` reducer actions (referenced in RESEARCH.md at `AppContext.tsx:49-50,127-130`; not re-quoted here as no further extraction was needed — the actions already exist and only need their *source* changed from localStorage-only to config-backed, not their shape).

No new action types are needed for density/railSide (already wired); this file's change is to the *read* path (boot-time hydration should eventually read from the migrated config value, staying synchronous via the `localStorage` write-through cache per RESEARCH.md Pitfall 1), not the reducer's action shape.

---

### `frontend/src/services/wailsAPI.ts` (modify, request-response)

**Analog:** itself — `wailsAPI.ts:209-216` (`setTheme`), `wailsAPI.ts:237-244` (`getCatalogHtmlPath`).

**Wrapper pattern to replicate for all 8 new fields** (verbatim, lines 209-216):
```typescript
setTheme: async (theme: string) => {
  try {
    await SetTheme(theme);
    return { success: true };
  } catch (error: any) {
    return wailsError(error);
  }
},
```
Every new wrapper (`setDensity`, `setRailSide`, `setCatalogDirectory`, `setDefaultFilenameRoot`, `setWriteHTML`, `setCopyToSecondary`, `setSecondaryDirectory`, `setWatchDirectory`) follows this identical 3-line try/catch/`wailsError` shape.

**Updated-signature pattern for `getCatalogHtmlPath`/`openExternal`** (current, lines 237-244):
```typescript
getCatalogHtmlPath: async (catalogPath: string) => {
  try {
    const htmlPath = await GetCatalogHtmlPath(catalogPath);
    return { success: true as const, htmlPath };
  } catch (error: any) {
    return wailsError(error);
  }
},
```
Add a `catalogDir: string` second parameter, threaded through to the generated `GetCatalogHtmlPath(catalogPath, catalogDir)` call — same pattern `revealInFileManager` already uses for its two-argument call (confirmed via `RevealInFileManager(path, catalogDir)` binding). `openExternal` gets the same second-argument treatment.

---

### `frontend/src/components/workspace/DetailsPanel.tsx` (modify, request-response)

**Analog:** itself — `handleReveal` (lines 130-147) already has the exact guard `handleOpenHtml` (lines 117-128) is missing.

**Guard pattern to copy verbatim into `handleOpenHtml`** (lines 130-147):
```typescript
async function handleReveal() {
  if (revealBusy) return;
  setRevealBusy(true);
  setError(null);
  if (!catalogDir) {
    setError('No catalog directory configured.');
    setRevealBusy(false);
    return;
  }
  const result = await wailsAPI.revealInFileManager(catalog.path, catalogDir);
  if (!result.success) {
    setError(result.error);
  }
  setRevealBusy(false);
}
```
Apply the identical `if (!catalogDir) { setError('No catalog directory configured.'); ...; return; }` guard to `handleOpenHtml` (currently lines 117-128, no guard), then update its two calls to `wailsAPI.getCatalogHtmlPath(catalog.path, catalogDir)` and `wailsAPI.openExternal(htmlPathResult.htmlPath, catalogDir)` to pass the new second argument.

---

### `frontend/src/components/workspace/CatalogRail.tsx` (modify, CRUD)

**Analog:** itself — current `CATALOG_DIR_STORAGE_KEY = 'storcat-catalog-directory'` (line 7) localStorage-only read/write.

This file's existing directory-persistence code is the literal thing being migrated into config (per CONTEXT.md's "Phase 25's `storcat-catalog-directory` ... localStorage keys migrate in the same pass"). Keep the localStorage read/write as the synchronous boot cache (per RESEARCH.md Pitfall 1) but add a parallel `wailsAPI.setCatalogDirectory(...)` call in the same handler that currently calls `safeSetItem(CATALOG_DIR_STORAGE_KEY, ...)`, using the same write-through-both-stores pattern established in `OptionsToggles.tsx`'s `handleToggleSecondary`.

## Shared Patterns

### Config Set*/Get* mechanical shape
**Source:** `internal/config/config.go:122-157`, `app.go:549-595`
**Apply to:** All 8 new config fields, both the `Manager` methods and their `App` binding wrappers.
```go
func (m *Manager) SetX(v T) error { m.config.X = v; return m.Save() }
func (a *App) SetX(v T) error {
	if a.configManager == nil { return nil }
	return a.configManager.SetX(v)
}
```

### wailsAPI try/catch wrapper
**Source:** `frontend/src/services/wailsAPI.ts:209-216`
**Apply to:** All 8 new frontend binding wrappers, and the two updated (`getCatalogHtmlPath`, `openExternal`).
```typescript
setX: async (v: T) => {
  try { await SetX(v); return { success: true }; }
  catch (error: any) { return wailsError(error); }
},
```

### Modal behavior (focus trap / Escape / scroll lock / focus restore)
**Source:** `frontend/src/hooks/useModalBehavior.ts` (full file, unchanged)
**Apply to:** `SettingsDialog.tsx` only — no bespoke reimplementation, per PLT-07 and the hook's own doc comment naming Phase 26 as an intended consumer.

### Fail-closed `catalogDir` guard before any containment-gated binding call
**Source:** `frontend/src/components/workspace/DetailsPanel.tsx:137-141` (`handleReveal`)
**Apply to:** `handleOpenHtml` (same file) and any new Settings call site that invokes `GetCatalogHtmlPath`/`OpenExternal`.

### Write-through boot cache (localStorage + config in the same synchronous tick)
**Source:** `frontend/src/components/workspace/create/OptionsToggles.tsx:93-108` (`safeSetItem` alongside state update)
**Apply to:** Every Settings control that has a pre-existing localStorage key (theme, density, railSide, catalogDirectory, secondaryDirectory) — call `safeSetItem` and the new `wailsAPI.set*` binding in the same event handler, per RESEARCH.md's Pitfall 1 resolution (do not delete the synchronous `main.tsx` boot read).

## No Analog Found

| File | Role | Data Flow | Reason |
|------|------|-----------|--------|
| `frontend/src/components/workspace/settings/ThemeGrid.tsx` | component | CRUD (render + select) | No existing "grid of selectable visual cards" component exists in the codebase; compose from `themes.ts`'s data shape (`THEMES` array, `tokens.{bg,p2,ac,tx}`) plus `ToggleRow`'s ARIA/click-handling conventions rather than a single direct analog. |
| Migration hook (`localStorage` → config, one-time) | utility/event-driven | No existing "read N localStorage keys once and write them into a new store, gated by a marker" pattern exists anywhere in this codebase (first data migration in the app's history) — build as a new, small function following this phase's own write-through-cache convention, not copied from an analog. |

## Metadata

**Analog search scope:** `frontend/src/components/workspace/`, `frontend/src/components/workspace/create/`, `frontend/src/hooks/`, `frontend/src/services/`, `frontend/src/contexts/`, `internal/config/`, `internal/osutil/`, `app.go`
**Files scanned:** 13 (full or targeted reads) + 2 grep sweeps (`App.tsx`, `CatalogRail.tsx`)
**Pattern extraction date:** 2026-08-15
