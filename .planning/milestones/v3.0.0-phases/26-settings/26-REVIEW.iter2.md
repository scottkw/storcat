---
phase: 26-settings
reviewed: 2026-08-15T00:00:00Z
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
  critical: 1
  warning: 2
  info: 0
  total: 3
status: issues_found
---

# Phase 26: Code Review Report

**Reviewed:** 2026-08-15T00:00:00Z
**Depth:** standard
**Files Reviewed:** 25
**Status:** issues_found

## Summary

Reviewed the Settings surface end to end: the new `internal/config` fields/setters and its new `RWMutex`, the `settingsStore.ts` write-through pattern, the marker-gated localStorage→Go-config migration, and the `catalogDir`-containment hardening on `GetCatalogHtmlPath`/`OpenExternal` (`internal/osutil/openexternal.go`, `ResolveContainedFileURL`).

The security-hardening work (`ResolveContainedFileURL`, the `catalogDir` threading into `GetCatalogHtmlPath`/`OpenExternal`) is solid: fail-closed on empty `catalogDir`, symlink-resolved before the containment check, extension-and-regular-file gated, argv-only process invocation elsewhere in the package, and the test suite (`openexternal_test.go`) actually exercises the sibling-directory-name-prefix and `../` escape cases that a naive `strings.HasPrefix` check would get wrong. No issues found there.

The one genuine blocker is in `internal/config/config.go`: this phase's `Get()` rewrite (returning a copy under `RLock` instead of the live pointer, for good reason — concurrency safety) unconditionally dereferences `m.config`, which is `nil` on the exact fallback path `app.go`'s `NewApp()` uses when `config.NewManager()` fails. That fallback previously degraded gracefully (every caller in `app.go` explicitly nil-checked `Get()`'s result); now it panics on the very first read, including the automatic `domReady` call at every app launch. See CR-01 below.

Two further issues, both narrower in blast radius: a settings-hydration race that can make the Settings dialog silently show a stale value (WR-01), and UI copy on the "Watch catalog directory" toggle that promises live behavior this phase explicitly does not implement (WR-02) — directly contradicting the constraint documented in `internal/config/config.go`'s own `WatchDirectory` comment.

## Critical Issues

### CR-01: `config.Manager.Get()` (and `GetWindowPersistence()`) panic on the app's own documented config-load-failure fallback

**File:** `internal/config/config.go:179-184` (also `internal/config/config.go:229-233`); triggered via `app.go:112-117`, `app.go:542-547`, `app.go:674-679`, `app.go:687-723`

**Issue:** `NewApp()` in `app.go` explicitly tolerates `config.NewManager()` failing (e.g. `os.UserConfigDir()`/`os.UserHomeDir()` both erroring, or `MkdirAll` failing on a read-only/sandboxed home) by falling back to a zero-value manager:

```go
configManager, err := config.NewManager()
if err != nil {
    // If config fails, just create one with defaults
    configManager = &config.Manager{}
}
```

Because `Manager.config` is unexported, `app.go` cannot populate it from outside the package — this zero-value `Manager` necessarily has `config == nil`. Before this phase, `Manager.Get()` simply returned `m.config` (i.e. `nil`) and every caller in `app.go` explicitly handled that:

```go
// pre-phase-26 domReady
cfg := a.configManager.Get()
if cfg == nil || !cfg.WindowPersistenceEnabled { return }
```

This phase's `Get()` rewrite (for the good reason of returning a copy so concurrent `Set*` calls can't race a reader — `T-26-02`) changed the implementation to:

```go
func (m *Manager) Get() *Config {
	m.mu.RLock()
	defer m.mu.RUnlock()
	cfg := *m.config   // panics if m.config == nil
	return &cfg
}
```

`*m.config` panics with a nil-pointer dereference before the caller's `cfg == nil` check ever runs. `GetWindowPersistence()` has the identical problem (`return m.config.WindowPersistenceEnabled`). Every code path that reaches this fallback Manager now crashes instead of degrading gracefully:
- `App.GetConfig()` (`app.go:542`) — checks `a.configManager == nil`, which is false here (it's a valid, non-nil pointer to an empty struct), so it proceeds to the panic.
- `App.domReady()` (`app.go:674`) — called automatically by Wails on every app launch, no nil guard on `configManager` at all.
- `App.beforeClose()` (`app.go:715`) — same.
- `App.GetWindowPersistence()` (`app.go:650`) — same `configManager == nil` false-negative as `GetConfig()`.

Confirmed empirically:
```
$ go test ./internal/config/ -run TestZeroValueManager_GetPanics -v
    zzz_nilcheck_test.go:10: confirmed panic: runtime error: invalid memory address or nil pointer dereference
--- PASS: TestZeroValueManager_GetPanics (0.00s)
```
(temporary test constructing `&Manager{}` and calling `.Get()`, removed after confirming — not part of the committed suite.)

This is a real regression: a condition the original code was explicitly written to tolerate ("If config fails, just create one with defaults") now crashes the app on the very first config read (`domReady`, unconditional on every launch) instead of running with in-memory defaults. It also affects every one of this phase's ~14 new setters identically — `SetTheme`, `SetDensity`, `SetRailSide`, `SetCatalogDirectory`, etc. all do `m.config.X = val` with no nil guard — but those are reachable only if a caller invokes a setter, whereas `Get()`/`GetWindowPersistence()` are hit unconditionally at startup via `domReady`.

**Fix:** Since `app.go` cannot construct a `Manager` with a populated (unexported) `config` field, fix this at the source — never let a `Manager` exist with a nil `config`. Add a constructor `internal/config` exports for exactly this fallback case, and make `NewApp()` use it instead of the bare struct literal:

```go
// internal/config/config.go
// NewDefaultManager returns a Manager pre-loaded with DefaultConfig() and
// no configPath -- for callers (app.go's NewApp) that need a working,
// non-persisting fallback when the disk-backed NewManager fails. Save()
// will fail gracefully (empty configPath) rather than every accessor
// panicking on a nil config.
func NewDefaultManager() *Manager {
	return &Manager{config: DefaultConfig()}
}
```

```go
// app.go
configManager, err := config.NewManager()
if err != nil {
	configManager = config.NewDefaultManager()
}
```

This one change fixes `Get()`, `GetWindowPersistence()`, and every `Set*` setter's identical latent nil-dereference at once, and preserves the original "run in-memory with defaults" intent instead of crashing. (Belt-and-suspenders alternative/addition: make `Get()`/`GetWindowPersistence()` nil-check `m.config` and fall back to `DefaultConfig()` internally, in case a future caller constructs a bare `Manager` again.)

## Warnings

### WR-01: `SETTINGS_HYDRATED` wholesale-replaces `state.settings`, racing a user's own in-flight toggle

**File:** `frontend/src/contexts/AppContext.tsx:386-387` (reducer case), `frontend/src/components/workspace/WorkspaceShell.tsx:45-60` (dispatch site), `frontend/src/settingsStore.ts:135-209` (`hydrateSettings`/`doHydrate`)

**Issue:** `WorkspaceShell` fires `hydrateSettings()` once on mount and, when it resolves, dispatches `SETTINGS_HYDRATED` with the full `AppSettings` snapshot read from Go config at that moment:

```ts
hydrateSettings().then((result) => {
  if (cancelled || !result) return;
  dispatch({ type: 'SETTINGS_HYDRATED', payload: result.settings });
  ...
});
```

The reducer case replaces the entire slice, not a merge:
```ts
case 'SETTINGS_HYDRATED':
  return { ...state, settings: action.payload };
```

`SettingsDialog` is reachable immediately at launch (⌘,/Ctrl+, is a global listener registered from the first render, `openSettings()` has no gate on hydration having completed). If a user opens Settings and toggles a row (e.g. "Write HTML alongside JSON" in `CatalogSettingsSection.tsx`) before the in-flight `hydrateSettings()` promise resolves, the sequence is:

1. `SET_SETTINGS{writeHtml: false}` updates `state.settings.writeHtml` to `false` and `setWriteHtmlSetting(false)` fires `wailsAPI.setWriteHTML(false)` (writes Go config immediately).
2. `hydrateSettings()`'s already-in-flight `getConfig()` call (issued at mount, before the toggle) resolves with the **pre-toggle** snapshot and dispatches `SETTINGS_HYDRATED`, which overwrites `state.settings` wholesale — silently reverting the toggle's on-screen value back to `true`, even though the underlying Go config file now correctly holds `false`.

The persisted value is not lost (the setter's own write already landed), but the UI now visibly disagrees with what was just persisted, with no error and no indication anything happened — exactly the kind of "map ≠ territory" divergence this codebase's own conventions call out elsewhere (e.g. `settingsStore.ts`'s top-of-file comment about the two stores never being allowed to diverge). The window is normally sub-second, but widens considerably during a first-run migration (`doHydrate`'s `if (!cfg.settingsMigrated)` branch performs up to 6 sequential awaited round-trips before the settings promise resolves).

**Fix:** Either gate the Settings entry points (⌘,/gear icon) until `hydrateSettings()` has resolved, or make the merge field-aware the same way `CreateSlideOver.tsx`'s own re-seed effect already handles this exact race for `root`/`writeHTML`/`copyToSecondary` (only overwrite a field if it is still at its untouched default):

```ts
case 'SETTINGS_HYDRATED': {
  const merged = { ...state.settings };
  (Object.keys(action.payload) as (keyof AppSettings)[]).forEach((key) => {
    if (state.settings[key] === DEFAULT_APP_SETTINGS[key]) {
      merged[key] = action.payload[key] as never;
    }
  });
  return { ...state, settings: merged };
}
```

### WR-02: "Watch catalog directory" toggle's copy claims live behavior this phase does not implement

**File:** `frontend/src/components/workspace/settings/CatalogSettingsSection.tsx:133-144`; contradicts the constraint documented at `internal/config/config.go:42-46`

**Issue:** `internal/config/config.go` is explicit that `WatchDirectory` is a persist-only field this phase, with a locked constraint on how it may be presented:

```go
// WatchDirectory persists only, this phase (26). Phase 27's
// WATCH-01..03 own the real fsnotify watcher; until then this field has
// no reader beyond the Settings toggle itself, and no surface (status
// bar, rail badge, or copy) may imply that watching is active.
WatchDirectory bool `json:"watchDirectory"`
```

The shipped toggle copy in `CatalogSettingsSection.tsx` violates that constraint directly:

```tsx
<ToggleRow
  checked={state.settings.watchDirectory}
  label="Watch catalog directory for changes"
  note="refresh the rail automatically"
  ...
```

"refresh the rail automatically" asserts the feature is live. Nothing in this phase reads `WatchDirectory` beyond the toggle itself (confirmed — `app.go`'s `SetWatchDirectory`/`config.go`'s `SetWatchDirectory` are the only reader/writer pair, no fsnotify, no rail-refresh subscriber exists yet). A user who enables this toggle is told their rail will auto-refresh on directory changes; it will not, until Phase 27 ships the actual watcher. This is a functional promise the current build cannot keep, not merely a stylistic nit — it directly contradicts the phase's own recorded constraint.

**Fix:** Use copy that describes the toggle as a preference rather than an active behavior, e.g. `"applies once file watching ships"` or simply drop the note entirely (the label alone, "Watch catalog directory for changes," is not itself a false claim — it's the note that oversells it).

---

_Reviewed: 2026-08-15T00:00:00Z_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_
