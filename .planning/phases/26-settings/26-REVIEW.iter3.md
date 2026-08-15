---
phase: 26-settings
reviewed: 2026-08-15T10:30:00Z
depth: standard
files_reviewed: 25
files_reviewed_list:
  - app.go
  - app_test.go
  - internal/config/config.go
  - internal/config/config_test.go
  - internal/osutil/openexternal.go
  - internal/osutil/openexternal_test.go
  - wails.json
  - frontend/src/App.tsx
  - frontend/src/settingsStore.ts
  - frontend/src/themeTokens.ts
  - frontend/src/workspace.css
  - frontend/src/contexts/AppContext.tsx
  - frontend/src/services/wailsAPI.ts
  - frontend/src/components/workspace/SettingsDialog.tsx
  - frontend/src/components/workspace/Toolbar.tsx
  - frontend/src/components/workspace/WorkspaceShell.tsx
  - frontend/src/components/workspace/CatalogRail.tsx
  - frontend/src/components/workspace/CreateSlideOver.tsx
  - frontend/src/components/workspace/DetailsPanel.tsx
  - frontend/src/components/workspace/create/OptionsToggles.tsx
  - frontend/src/components/workspace/settings/CatalogSettingsSection.tsx
  - frontend/src/components/workspace/settings/SegmentedControl.tsx
  - frontend/src/components/workspace/settings/ThemeGrid.tsx
  - frontend/wailsjs/go/main/App.d.ts
  - frontend/wailsjs/go/main/App.js
  - frontend/wailsjs/go/models.ts
findings:
  critical: 0
  warning: 2
  info: 0
  total: 2
status: issues_found
---

# Phase 26: Code Review Report

**Reviewed:** 2026-08-15T10:30:00Z
**Depth:** standard
**Files Reviewed:** 25
**Status:** issues_found

## Summary

Re-review (iteration 2) of the Settings surface after the three fixes recorded in `26-REVIEW-FIX.md` landed (commits `423792df`, `766d1476`, `2637ff5e`). All three are verified fixed, correctly, with no regressions:

- **CR-01** (blocker): `internal/config/config.go` now exports `NewDefaultManager()`, which returns `&Manager{config: DefaultConfig()}` — never a nil `config`. `app.go:116` calls it in `NewApp()`'s `config.NewManager()` error fallback, replacing the old bare `&config.Manager{}` literal. Traced every construction path in the reviewed files: `NewManager()` itself now always leaves `m.config` populated on every one of its own error branches (the `os.IsNotExist` branch, the generic `Load()`-error branch via `NewManager`'s own catch, and the JSON-unmarshal-error branch), and the only other zero-value `&App{}` construction sites are in `app_test.go`, which never call `domReady`/`beforeClose` and only exercise methods that already nil-check `a.configManager` before touching it. `go test ./internal/config/... ./... -race -count=1` passes, including the added `TestNewDefaultManager_GetDoesNotPanic` regression pin. Confirmed clean — no new BLOCKER here.
- **WR-01**: `AppContext.tsx`'s `SETTINGS_HYDRATED` case (lines 386-402) now merges field-by-field instead of replacing wholesale, only folding in a hydrated value when the current in-memory field still equals `DEFAULT_APP_SETTINGS[key]`. Every field in `AppSettings` (`settingsStore.ts:60-67`) is a primitive (`string`/`boolean`), so the `===` comparison the merge relies on is well-defined for all six fields — no reference-equality trap from an object/array field. Confirmed correctly implemented for the case it targets. See WR-A below for a narrower residual edge this same heuristic introduces.
- **WR-02**: `CatalogSettingsSection.tsx:136`'s watch-toggle note now reads `"applies once file watching ships"`, no longer implying live behavior. Confirmed matches the `WatchDirectory` field's locked doc-comment constraint in `internal/config/config.go:42-46`.

Two new findings surfaced by this pass, both narrow-scope WARNINGs building on the same seams the prior fixes touched — no new BLOCKERs.

## Warnings

### WR-A: `domReady` is the one `App` method that still dereferences `a.configManager` with no nil guard

**File:** `app.go:673-684`

**Issue:** Every other `App` method that touches `a.configManager` — `GetConfig` (542), all ~14 setters (550-670), and `beforeClose` (712) — opens with `if a.configManager == nil { ... }` before calling into it. `domReady` is the sole exception:

```go
// domReady is called after the frontend DOM is ready
func (a *App) domReady(ctx context.Context) {
	cfg := a.configManager.Get()
	if cfg == nil || !cfg.WindowPersistenceEnabled {
		return
	}
	...
}
```

`a.configManager.Get()` is a pointer-receiver method (`func (m *Manager) Get()`); if `a.configManager` itself is `nil` (not `m.config`, but the `*Manager` field on `App`), the call panics immediately on `m.mu.RLock()` — accessing the `mu` field requires dereferencing the nil receiver before the method body's own nil-check on `cfg` ever runs. This is exactly the failure shape CR-01 just fixed one layer down (a nil `Config` inside a non-nil `Manager`), left unfixed one layer up (a nil `Manager` itself) at this one call site.

Today this is latent, not reachable: `NewApp()` is the only production path that constructs `App`, and it now guarantees `configManager` is never nil (either `config.NewManager()` succeeds, or the CR-01 fix's `config.NewDefaultManager()` fallback runs). `app_test.go` constructs `&App{}` directly (zero-value, nil `configManager`) in ~28 places but never calls `domReady` in any of them, so the gap has no current test coverage exercising it either way. But it is a real inconsistency with the defensive pattern the rest of this same file just re-affirmed in the CR-01 fix, `domReady` is called unconditionally by Wails on every app launch, and nothing enforces that `NewApp()` remains the only construction path going forward (a future test, or a CLI-mode entry point, constructing `App` directly would panic here with no coverage to catch it).

**Fix:** Add the same guard every sibling method already uses:

```go
func (a *App) domReady(ctx context.Context) {
	if a.configManager == nil {
		return
	}
	cfg := a.configManager.Get()
	if cfg == nil || !cfg.WindowPersistenceEnabled {
		return
	}
	...
}
```

### WR-B: `SETTINGS_HYDRATED`'s field-aware merge can silently discard a deliberate user write that happens to equal the default value

**File:** `frontend/src/contexts/AppContext.tsx:386-402`

**Issue:** The WR-01 fix's merge heuristic treats "current value equals `DEFAULT_APP_SETTINGS[key]`" as a proxy for "the user hasn't touched this field yet, safe to overwrite with the hydrated value." That proxy is accurate for the case WR-01 targeted (a user changing a field away from its default before hydration resolves), but it is the wrong signal in the mirror case: a user whose persisted config already holds a non-default value for some field, who then explicitly sets that field back to its default value during the same hydration race window.

Concretely, with `writeHtml` (default `true`):
1. A previous session persisted `writeHtml: false` (the user turned it off before).
2. On this launch, `hydrateSettings()`'s `getConfig()` call is issued at mount (in flight); `state.settings.writeHtml` is still sitting at the initial-state default, `true`.
3. Before that promise resolves, the user opens Settings and explicitly re-enables the toggle back to `true` (matching the default). `SET_SETTINGS{writeHtml: true}` fires and `setWriteHtmlSetting(true)` persists it — the on-disk config now correctly holds `true`.
4. The in-flight `getConfig()` resolves with the pre-toggle snapshot (`writeHtml: false`) and dispatches `SETTINGS_HYDRATED`. Because `state.settings.writeHtml` (`true`) still equals `DEFAULT_APP_SETTINGS.writeHtml` (`true`) at merge time, the heuristic treats the field as "untouched" and folds in the stale hydrated `false` — silently reverting the just-persisted, deliberate `true` back to `false` on screen (and the config file itself; nothing re-writes it after the merge).

This is narrower than the bug WR-01 fixed (it additionally requires the persisted value to already differ from the default, which is not the common case on a fresh install), and it self-corrects on the next Settings-dialog open or app restart. But it is a real, silent divergence between what the user just did and what the UI (and the on-disk config, since nothing re-persists after the merge) end up showing, in the exact race window the codebase's own `settingsStore.ts` header comment says the two stores "cannot diverge under normal operation." The underlying race — `hydrateSettings()` is not awaited before Settings becomes reachable (`WorkspaceShell.tsx`'s `settingsOpen` state has no gate on the hydration effect) — is unchanged by either fix.

**Fix:** The two options from the original WR-01 finding remain the real fix: gate the Settings entry points (⌘,/gear icon in `Toolbar.tsx`) until `hydrateSettings()` has resolved, so no `SET_SETTINGS` can race the initial `SETTINGS_HYDRATED`. Given the field-aware merge is staying as belt-and-suspenders, tracking "touched" explicitly (e.g. a `Set<keyof AppSettings>` of dispatched keys, seeded empty and added to on every `SET_SETTINGS`) rather than inferring it from value-equals-default would close this remaining gap without reintroducing the default-value ambiguity:

```ts
case 'SET_SETTINGS':
  ...
  return { ...state, settings: { ...state.settings, ...action.payload }, touchedSettings: new Set([...state.touchedSettings, ...Object.keys(action.payload)]) };

case 'SETTINGS_HYDRATED': {
  const merged = { ...state.settings };
  (Object.keys(action.payload) as (keyof AppSettings)[]).forEach((key) => {
    if (!state.touchedSettings.has(key)) merged[key] = action.payload[key] as never;
  });
  return { ...state, settings: merged };
}
```

---

_Reviewed: 2026-08-15T10:30:00Z_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_
