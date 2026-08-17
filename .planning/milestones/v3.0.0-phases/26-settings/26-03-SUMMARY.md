---
phase: 26-settings
plan: 03
subsystem: ui
tags: [wails, react, typescript, go, config, settings, migration]

# Dependency graph
requires:
  - phase: 26-settings
    provides: "26-01's write-through settingsStore.ts pattern, config.Manager RWMutex/copy-returning Get(), SettingsDialog.tsx shell; 26-02's setThemeSetting/setRailSideSetting precedent"
provides:
  - "config.Config.CatalogDirectory / DefaultFilenameRoot / SecondaryDirectory fields + Manager setters + App bindings"
  - "settingsStore.AppSettings (six-field shape) + DEFAULT_APP_SETTINGS -- the one place plan 26-05's toggles section widens, not redeclares"
  - "settingsStore.hydrateSettings() -- deduped, marker-gated localStorage-to-config migration plus config-to-cache write-back"
  - "AppContext.state.settings + SETTINGS_HYDRATED/SET_SETTINGS actions; SET_CATALOG_DIR reducer bail-out on an unchanged value"
  - "CatalogSettingsSection.tsx -- Catalogs section, directory chip + Change link, default-filename-root input"
  - "CreateSlideOver re-seeds its filename-root field from state.settings.defaultFilenameRoot on every open, not just at mount"
affects: [26-04, 26-05]

# Actuals (#2632)
actuals:
  tokens: 9833
  tasks: 3
  commits: 4

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "settingsStore setters continue the two-write shape for cached fields (setCatalogDirectorySetting, setSecondaryDirectorySetting); setDefaultFilenameRootSetting is deliberately config-only -- no cache, since nothing reads it pre-paint"
    - "Reducer same-value bail-out (SET_CATALOG_DIR, SET_SETTINGS) returns the identical state object so React's own bail-out skips the re-render -- the mechanism, not a manual dispatch guard at each call site"
    - "hydrateSettings() dedup: a module-level in-flight promise shared by every caller in the same tick, so React 18 StrictMode's double-invoked mount effect issues exactly one getConfig() round trip"
    - "Always-mounted overlay components (CreateSlideOver) cannot rely on a useState initializer to reflect state that hydrates asynchronously after the component's first render -- an isOpen-keyed re-seed effect is the fix, not a second apply path"

key-files:
  created:
    - frontend/src/components/workspace/settings/CatalogSettingsSection.tsx
  modified:
    - internal/config/config.go
    - internal/config/config_test.go
    - app.go
    - frontend/wailsjs/go/main/App.d.ts
    - frontend/wailsjs/go/main/App.js
    - frontend/wailsjs/go/models.ts
    - frontend/src/services/wailsAPI.ts
    - frontend/src/settingsStore.ts
    - frontend/src/contexts/AppContext.tsx
    - frontend/src/components/workspace/CatalogRail.tsx
    - frontend/src/components/workspace/SettingsDialog.tsx
    - frontend/src/components/workspace/CreateSlideOver.tsx
    - frontend/src/components/workspace/WorkspaceShell.tsx
    - frontend/src/components/workspace/create/OptionsToggles.tsx
    - frontend/src/workspace.css

key-decisions:
  - "Batched all three new Go config fields/setters/bindings (CatalogDirectory, DefaultFilenameRoot, SecondaryDirectory) and one `wails generate module` regeneration into Task 1's commit, rather than three separate regenerations -- a pragmatic grouping call (all three are trivial, identically-shaped setters), not a behavior change. Each task's own Go test trio still lands in that task's own commit."
  - "Found and fixed a real bug during Task 2's live proof: CreateSlideOver is always-mounted, so its `root` useState initializer ran at first render, before any settings change (via Settings dialog or async hydration) could ever reach it -- a Settings-set filename root never actually arrived in the create form. Fixed with an isOpen-keyed re-seed effect that only fires when the field is still blank, so it never clobbers a value the user already typed for the current attempt, and deliberately does not list the setting itself as a dependency (E4 partial: a setting changed while the panel is already open must not retroactively rewrite it)."
  - "Reset the local dev config.json to an unmigrated state and forced a full re-migration live via the real Set* Wails bindings (rather than editing the on-disk file while wails dev was running, which is a no-op against the in-memory config.Manager singleton) -- confirmed no flash, correct migrated values, all five localStorage keys survived, and a second reload neither re-migrated nor let stale cache clobber a newer config value"
  - "hydrateSettings' dedup (module-level in-flight promise) is asserted by code inspection and by the migration producing clean, non-corrupted values under React 18 StrictMode's double-invoked mount effect, not by an isolated call-count probe -- dev-browser's page.reload() discards any window-level call-counter shim before the new page's JS runs, so a live call-count proof isn't practically reachable through this harness"

requirements-completed: [SET-04, SET-05]

coverage:
  - id: D1
    description: "Catalog directory is one shared value: Settings' directory chip and 'Change...' link write through the same setter/dispatch the rail uses; re-selecting the already-configured directory is a true no-op (open catalog, tree, expansion and selection all survive); choosing a different directory updates both surfaces and reaches disk"
    requirement: "SET-04"
    verification:
      - kind: unit
        ref: "internal/config/config_test.go#TestSetCatalogDirectory, TestSetCatalogDirectory_Persists, TestDefaultConfig_CatalogDirectoryEmpty"
        status: pass
      - kind: automated_ui
        ref: "dev-browser session: opened alpha-volume catalog, expanded its docs branch, opened Settings, re-selected the identical directory via a picker-return mock -- tree/expansion unchanged, dialogOpen stayed true, rowCount stayed 5; then selected a different directory -- rail chip, Settings chip, rail listing (now gamma-volume) and GetConfig().catalogDirectory all updated to the new path"
        status: pass
    human_judgment: false
  - id: D2
    description: "Default filename root is settable from Settings, whitespace-stripped on every keystroke, persisted with no save step, valid when empty, and pre-fills the create form's filename-root field"
    requirement: "SET-04"
    verification:
      - kind: unit
        ref: "internal/config/config_test.go#TestSetDefaultFilenameRoot, TestSetDefaultFilenameRoot_Persists, TestSetDefaultFilenameRoot_EmptyIsValid"
        status: pass
      - kind: automated_ui
        ref: "dev-browser session: typed 'my root' into the Settings input -- displayed and persisted as 'myroot' (GetConfig().defaultFilenameRoot); opened the create slide-over -- filename-root field arrived pre-filled with 'myroot' (only after the Task-2 re-seed-on-open fix); cleared the Settings field -- GetConfig() reported '' with no error and a fresh create form fell back to its source-derived placeholder"
        status: pass
    human_judgment: false
  - id: D3
    description: "The one-time localStorage-to-config migration runs exactly once, gated by the settingsMigrated marker (never a zero-value inference), carries theme/density/rail-side/catalog-directory/secondary-directory into the Go config through the same allowlist readPersistedPrefs uses, never deletes a localStorage key, and produces no launch flash"
    requirement: "SET-05"
    verification:
      - kind: unit
        ref: "internal/config/config_test.go#TestSetSecondaryDirectory, TestSetSecondaryDirectory_Persists (go test ./... -race -count=1, whole suite green)"
        status: pass
      - kind: automated_ui
        ref: "dev-browser session: reset config to unmigrated via real Set* bindings, seeded all five storcat-* keys to distinct non-default values, reloaded -- data-theme was the seeded value immediately post-reload (no flash), GetConfig() reported all five migrated values plus settingsMigrated:true, all five localStorage keys survived unchanged; changed density directly via a binding, reloaded again -- config value held (no re-migration) and the cache converged to match it rather than clobbering it"
        status: pass
    human_judgment: false
  - id: D4
    description: "hydrateSettings() is deduped behind a module-level in-flight promise, so React 18 StrictMode's double-invoked mount effect issues exactly one getConfig() round trip"
    verification: []
    human_judgment: true
    rationale: "Confirmed by code inspection (a standard module-level in-flight-promise idiom) and indirectly by the migration producing clean, non-corrupted values under StrictMode's double-invoke in the live proof above -- a direct call-count probe was attempted via a window-level shim, but dev-browser's page.reload() performs a real navigation that discards any such shim before the new page's JS executes, so an isolated count assertion wasn't reachable through this harness."

duration: ~19min
completed: 2026-08-15
status: complete
---

# Phase 26 Plan 03: Catalogs Section + localStorage-to-Config Migration Summary

**Catalog directory and default filename root become config-backed settings shared between Settings and the rail/create-form, and the one-time marker-gated migration carries every existing user's five localStorage settings into the Go config non-destructively.**

## Performance

- **Duration:** ~19 min
- **Started:** 2026-08-15T09:00:56-05:00
- **Completed:** 2026-08-15T09:19:37-05:00
- **Tasks:** 3
- **Files modified:** 15 (1 created, 14 modified)

## Accomplishments

- `config.Config` gained `CatalogDirectory`, `DefaultFilenameRoot` and `SecondaryDirectory` (all default `""`), each with a lock-guarded `Manager` setter and `App` binding, following 26-01's established shape
- `AppContext`'s `SET_CATALOG_DIR` case now returns the identical state object when the payload equals the current value -- the SET-04 adjacency edge enforced structurally at the reducer, not per call site
- `CatalogRail.tsx` restructured: the mount effect only dispatches the persisted directory; a new effect keyed on `state.catalogDir` owns the listing call, so a Settings-driven directory change refreshes the rail with no duplicate listing
- `CatalogSettingsSection.tsx` (new): the Catalogs section -- directory chip + "Change..." link (both routed through `setCatalogDirectorySetting`/dispatch) and a default-filename-root input (whitespace-stripped on every keystroke, no validation/disabled/max-length)
- `settingsStore.AppSettings` (six fields) + `DEFAULT_APP_SETTINGS`, `AppContext.state.settings`, `SETTINGS_HYDRATED`/`SET_SETTINGS` actions -- the one place plan 26-05's toggles section will widen, not redeclare
- Fixed a real bug found live: `CreateSlideOver` is always-mounted, so its filename-root field never actually reflected a Settings-set default until an isOpen-keyed re-seed effect was added
- `settingsStore.hydrateSettings()`: deduped, marker-gated migration of the five `storcat-*` cache keys into the Go config (each validated through the same allowlist `readPersistedPrefs()` uses), plus a config-to-cache write-back -- never deletes a localStorage key. Wired into `WorkspaceShell`'s mount effect.
- `OptionsToggles.tsx` re-points `SECONDARY_DIR_STORAGE_KEY` at `settingsStore.SECONDARY_DIR_KEY` and both its handlers now call `setSecondaryDirectorySetting`

## Task Commits

Each task was committed atomically (plus one live-proof-driven fix):

1. **Task 1: Catalog directory -- one shared value across Settings and the rail** - `5a435c85` (feat)
2. **Task 2: Default filename root -- settings state slice and create-form pre-fill** - `12be8927` (feat)
3. **Live-proof fix: re-seed create form's filename root on each open** - `fa488649` (fix)
4. **Task 3: One-time localStorage-to-config migration and boot hydration** - `cd93e6cc` (feat)

**Plan metadata:** (this commit)

## Files Created/Modified

- `internal/config/config.go` - `CatalogDirectory`/`DefaultFilenameRoot`/`SecondaryDirectory` fields + three setters
- `internal/config/config_test.go` - 8 new tests across the three tasks (set/persist/default/empty-valid per field, plus the secondary-directory pair)
- `app.go` - three new Wails bindings
- `frontend/wailsjs/go/main/App.d.ts`, `App.js`, `frontend/wailsjs/go/models.ts` - regenerated via `wails generate module`
- `frontend/src/services/wailsAPI.ts` - `setCatalogDirectory`/`setDefaultFilenameRoot`/`setSecondaryDirectory` wrappers
- `frontend/src/settingsStore.ts` - `CATALOG_DIR_KEY`/`setCatalogDirectorySetting`, `AppSettings`/`DEFAULT_APP_SETTINGS`, `setDefaultFilenameRootSetting`, `SECONDARY_DIR_KEY`/`setSecondaryDirectorySetting`, `hydrateSettings()`
- `frontend/src/contexts/AppContext.tsx` - `SET_CATALOG_DIR` bail-out guard, `state.settings`, `SETTINGS_HYDRATED`/`SET_SETTINGS`
- `frontend/src/components/workspace/CatalogRail.tsx` - imports `CATALOG_DIR_KEY`, restructured mount/listing effects, `setCatalogDirectorySetting`
- `frontend/src/components/workspace/settings/CatalogSettingsSection.tsx` - Catalogs section (new)
- `frontend/src/components/workspace/SettingsDialog.tsx` - renders `<CatalogSettingsSection />`
- `frontend/src/components/workspace/CreateSlideOver.tsx` - seeds/re-seeds `root` from `state.settings.defaultFilenameRoot`
- `frontend/src/components/workspace/WorkspaceShell.tsx` - mount effect calling `hydrateSettings()`
- `frontend/src/components/workspace/create/OptionsToggles.tsx` - re-points the secondary-directory key, calls the shared setter
- `frontend/src/workspace.css` - `ws-settings-value-col`, `ws-settings-dir-chip`, `ws-settings-change-link`, `ws-settings-root-input`

## Decisions Made

- Batched all three Go setters/bindings and one `wails generate module` pass into Task 1's commit (each task still lands its own Go test trio in its own commit) -- a grouping call for three identically-shaped, trivial setters, not a behavior change
- Fixed `CreateSlideOver`'s always-mounted-initializer race (see Deviations) with a minimal isOpen-keyed re-seed effect rather than restructuring the component's mount lifecycle
- Forced the live migration test through the real `Set*` bindings (not by editing `config.json` on disk while `wails dev` was running, which is a no-op against the already-loaded in-memory `config.Manager`) after discovering that editing the file directly doesn't affect a running process

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Create form's filename-root field never reflected a Settings-set default**
- **Found during:** Task 2's live proof (open Settings, type a root, open the create slide-over -- field was blank, not pre-filled)
- **Issue:** `CreateSlideOver` is always-mounted per its own top comment; its `root` useState initializer (`useState(() => state.settings.defaultFilenameRoot)`) therefore ran once, at the component's very first render -- before any Settings-dialog edit or async hydration effect could ever populate `state.settings`. The "Pre-filled for every new catalog" promise (SET-04) silently never held.
- **Fix:** Added an effect keyed on `isOpen` that re-seeds `root` from `state.settings.defaultFilenameRoot` whenever the panel opens on a still-blank field. Guarded so it never clobbers a root the user already typed, and deliberately does not depend on the setting value itself (E4 partial: a setting changed while the panel is already open must not retroactively rewrite it).
- **Files modified:** `frontend/src/components/workspace/CreateSlideOver.tsx`
- **Commit:** `fa488649`

Otherwise: plan executed as written.

## Known Stubs

None -- `writeHtml`, `copyToSecondary` and `watchDirectory` in `AppSettings`/`hydrateSettings()` intentionally fall back to `DEFAULT_APP_SETTINGS`' values rather than reading a Go config field, because plan 26-05 (not yet executed) is the plan that adds those three fields to `config.Config`. This is the shape the plan itself specifies ("plan 26-05 adds the Go fields for the three toggle values"), not an unplanned gap.

## Issues Encountered

- `hydrateSettings()`'s in-flight-promise dedup could not be proven via an isolated call-count probe: dev-browser's `page.reload()` performs a real navigation, which discards any `window`-level call-count shim before the newly loaded page's JS executes, so a monkey-patched `GetConfig` counter never survives the reload it needs to observe. Verified by code inspection (a standard module-level in-flight-promise idiom, structurally identical to the pattern) and indirectly by the migration producing clean, non-corrupted values under React 18 StrictMode's double-invoked mount effect in the full migration proof. Logged here rather than asserted as directly measured.
- Editing the on-disk `config.json` while `wails dev` was still running had no effect on the live process's in-memory `config.Manager` (expected, once noticed -- the Manager is a process-lifetime singleton loaded once at `NewManager()`), which initially produced a confusing partial-migration result. Resolved by resetting config state through the real `Set*` Wails bindings instead, which is also a more faithful test of the actual code path.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- The write-through pattern, `AppSettings` shape, and `hydrateSettings()` migration are all in place for plan 26-05 to add the three remaining toggle fields (`writeHtml`, `copyToSecondary`, `watchDirectory`) to `config.Config` and wire the Toggles section -- `hydrateSettings()` already has the exact spot (marked in a comment) where those three lines change from `DEFAULT_APP_SETTINGS` fallbacks to real config reads.
- SET-04 is now fully closed (catalog directory + default filename root); SET-05's no-save-step and idempotency contract is proven for both new rows.
- `.planning/WINDOWS.md` still carries entries #1, #2, #4, #5, #6 from earlier phases (unrelated to this plan) plus #7 (RailSide relaunch-no-flash, 26-02) -- unaffected by this plan's scope.

---
*Phase: 26-settings*
*Completed: 2026-08-15*

## Self-Check: PASSED

All created/modified files verified present on disk; all task commit hashes (`5a435c85`, `12be8927`, `fa488649`, `cd93e6cc`) verified present in `git log`.
