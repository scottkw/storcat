---
phase: 26
slug: settings
# status lifecycle: draft (seeded by plan-phase) → validated (set by validate-phase §6)
status: validated
nyquist_compliant: true
wave_0_complete: true
created: 2026-08-15
validated: 2026-08-15
---

# Phase 26 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.
> Seeded by plan-phase from `26-RESEARCH.md`'s `## Validation Architecture`. The Per-Task
> Verification Map is filled in once PLAN.md task IDs exist.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go stdlib `testing`, table-driven, `*_test.go` beside source (existing: `internal/config/config_test.go`, `internal/osutil/reveal_test.go`). Frontend: **none by design** — TEST-01 (Vitest + Testing Library) is an explicitly deferred milestone item; do not add one. |
| **Config file** | none — plain `go test ./...`; `frontend/tsconfig.json`, `frontend/vite.config.ts` unchanged |
| **Quick run command** | `go test ./internal/config/... ./internal/osutil/... -race -count=1` |
| **Full suite command** | `go build ./... && go test ./... -race -count=1 && (cd frontend && npx tsc --noEmit && npm run build)` |
| **Estimated runtime** | ~60–90 seconds |

---

## Sampling Rate

- **After every task commit:** `go build ./... && go test ./internal/config/... ./internal/osutil/... -race -count=1` for Go tasks; `npx tsc --noEmit` for frontend tasks
- **After every plan wave:** full suite command above, plus a live dev-browser pass against `:34115`
- **Before `/gsd-verify-work`:** full suite green plus the manual-only entry-point and visual checks below
- **Max feedback latency:** ~90 seconds

**Dev-server note:** browser verification runs against `wails dev` on **`:34115`**. Vite's `:5173` exposes no `window.go`, so every binding-dependent assertion passes vacuously there. Per STATE.md's standing note, `curl` liveness proves the server is up, not that bindings are fresh — verify `Object.keys(window.go.main.App)` includes the new/changed methods before recording any evidence.

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 26-01 Task 1: End-to-end "change Row density in Settings" tracer | 26-01 | 1 | SET-01, SET-03, SET-05 | T-26-01, T-26-02 | Density allowlist (`readPersistedPrefs`) rejects any non-`Compact`/`Comfortable` string before it reaches `applyTokens()`; `Manager`'s `RWMutex` makes concurrent `Set*` calls race-free | TDD unit (Go) + live dev-browser | `go build ./... && go test ./internal/config/... -race -count=1 -run 'TestSetDensity\|TestSetSettingsMigrated\|TestDefaultConfig\|TestManager_ConcurrentSetters\|TestSetWindow\|TestGetWindowPersistence\|TestWindowPosition' && grep -q 'export function SetDensity' frontend/wailsjs/go/main/App.d.ts && grep -q 'density' frontend/wailsjs/go/models.ts && (cd frontend && npx tsc --noEmit && npm run build)` | ✅ `internal/config/config_test.go` | ✅ green |
| 26-01 Task 2: Other two entry points, overlay coexistence, close paths, version string | 26-01 | 1 | SET-01, SET-05 | T-26-03, T-26-04, T-26-05 | `openSettings` no-ops during an active scan; config writes stay confined to `storcatConfigDir()/config.json`; `--z-dialog` scrim covers the full viewport | unit (Go) + live dev-browser | `go build ./... && go test ./... -race -count=1 -run TestGetVersion && grep -q '"productVersion": "3.0.0"' wails.json && grep -c 'onOpenSettings' frontend/src/components/workspace/Toolbar.tsx \| grep -qE '[3-9]\|[0-9]{2}' && (cd frontend && npx tsc --noEmit && npm run build)` | ✅ `app_test.go` | ✅ green |
| 26-02 Task 1: Theme section — 11 cards | 26-02 | 2 | SET-02 | T-26-06, T-26-07 | Theme id allowlisted through `getThemeById(id) ?? getDefaultTheme()`; swatch colors sourced only from the compile-time `themes` array, never from config/localStorage/binding return values | static assertion + live dev-browser | `test $(grep -c "id: '" frontend/src/themes.ts) -eq 11 && grep -q 'themes.map' frontend/src/components/workspace/settings/ThemeGrid.tsx && ! grep -nE '\.(sort\|reverse)\(' frontend/src/components/workspace/settings/ThemeGrid.tsx && (cd frontend && npx tsc --noEmit && npm run build)` | ✅ `frontend/src/themes.ts`, `ThemeGrid.tsx` | ✅ green |
| 26-02 Task 2: Catalog rail position | 26-02 | 2 | SET-03 | T-26-06, T-26-08 | `railSide` allowlisted through `readPersistedPrefs()`'s exact-string check; `setThemeSetting`'s single `themeChange` apply path has no fourth divergent path | TDD unit (Go) + live dev-browser | `go build ./... && go test ./internal/config/... -race -count=1 -run 'TestSetRailSide\|TestDefaultConfig' && grep -q 'export function SetRailSide' frontend/wailsjs/go/main/App.d.ts && grep -q 'railSide' frontend/wailsjs/go/models.ts && (cd frontend && npx tsc --noEmit && npm run build)` | ✅ `internal/config/config_test.go` | ✅ green |
| 26-03 Task 1: Catalog directory — one shared value | 26-03 | 3 | SET-04 | T-26-11 | Stored directory is a user-declared boundary only; every file binding that consumes it independently resolves symlinks and calls `osutil.ContainsPath` | TDD unit (Go) + live dev-browser | `go build ./... && go test ./internal/config/... -race -count=1 -run 'TestSetCatalogDirectory\|TestDefaultConfig' && grep -q 'export function SetCatalogDirectory' frontend/wailsjs/go/main/App.d.ts && test $(grep -rc "'storcat-catalog-directory'" frontend/src \| grep -v ':0$' \| wc -l) -eq 1 && (cd frontend && npx tsc --noEmit && npm run build)` | ✅ `internal/config/config_test.go` | ✅ green |
| 26-03 Task 2: Default filename root | 26-03 | 3 | SET-04 | — | Whitespace-stripped, empty-valid, no validation error surface | TDD unit (Go) + live dev-browser | `go build ./... && go test ./internal/config/... -race -count=1 -run 'TestSetDefaultFilenameRoot' && grep -q 'SETTINGS_HYDRATED' frontend/src/contexts/AppContext.tsx && grep -q 'settings.defaultFilenameRoot' frontend/src/components/workspace/CreateSlideOver.tsx && (cd frontend && npx tsc --noEmit && npm run build)` | ✅ `internal/config/config_test.go` | ✅ green |
| 26-03 Task 3: One-time localStorage-to-config migration and boot hydration | 26-03 | 3 | SET-05 | T-26-09, T-26-10, T-26-12, T-26-13 | Every migrated value passes the same allowlist the boot read applies before reaching a binding; migration deduped behind a module-level in-flight promise; no localStorage key ever deleted | unit (Go) + live dev-browser | `go build ./... && go test ./internal/config/... -race -count=1 && test $(grep -rl "'storcat-secondary-directory'" frontend/src \| wc -l) -eq 1 && grep -q 'hydrateSettings' frontend/src/components/workspace/WorkspaceShell.tsx && ! grep -nE 'localStorage.removeItem\|safeRemoveItem' frontend/src/settingsStore.ts && (cd frontend && npx tsc --noEmit && npm run build)` | ✅ `internal/config/config_test.go` | ✅ green |
| 26-04 Task 1: `ResolveContainedFileURL` — pure validator | 26-04 | 4 | SET-04 (FU-23-A) | T-26-14, T-26-15 | Scheme/regular-file/extension/containment validated before any open; discharges FU-23-A | TDD unit (Go), RED then GREEN | `go build ./... && go test ./internal/osutil/... -race -count=1 -v -run TestResolveContainedFileURL` | ✅ `internal/osutil/openexternal_test.go` | ✅ green |
| 26-04 Task 2: Thread `catalogDir` through both bindings | 26-04 | 4 | SET-04 (FU-23-A) | T-26-14, T-26-15, T-26-18 | `GetCatalogHtmlPath`/`OpenExternal` reject a path resolving outside `catalogDir`; a rejection surfaces through `setError`, never fails silently | TDD unit (Go), RED then GREEN + live dev-browser | `go build ./... && go test ./... -race -count=1 -run 'TestGetCatalogHtmlPath\|TestOpenExternal\|TestReadHtmlFile' && grep -q 'export function OpenExternal(arg1:string,arg2:string)' frontend/wailsjs/go/main/App.d.ts && grep -q 'export function GetCatalogHtmlPath(arg1:string,arg2:string)' frontend/wailsjs/go/main/App.d.ts && (cd frontend && npx tsc --noEmit && npm run build)` | ✅ `app_test.go` | ✅ green |
| 26-04 Task 3: Delete unreachable `CatalogModal.tsx` | 26-04 | 4 | SET-04 (T-22-05) | T-26-16 | Unsanitized `iframe srcDoc` surface removed outright, not sanitized; reachability re-confirmed by grep immediately before deletion; discharges T-22-05 | static assertion + live dev-browser | `test ! -f frontend/src/components/CatalogModal.tsx && ! grep -rn 'openCatalogModal' frontend/src && ! grep -rn "components/CatalogModal" frontend/src && grep -q "ConfigProvider" frontend/src/App.tsx && grep -q "'themeChange'" frontend/src/App.tsx && (cd frontend && npx tsc --noEmit && npm run build)` | ✅ (file absence itself is the assertion) | ✅ green |
| 26-05 Task 1: Write-HTML and watch-directory toggles, shared `ToggleRow` | 26-05 | 5 | SET-04 | T-26-19, T-26-21 | Both values are booleans with no path/string/privilege; watch-directory toggle surfaces no status-bar/rail claim that a watcher is active | TDD unit (Go) + live dev-browser | `go build ./... && go test ./internal/config/... -race -count=1 -run 'TestSetWriteHTML\|TestSetWatchDirectory\|TestDefaultConfig' && grep -q 'export function ToggleRow' frontend/src/components/workspace/create/OptionsToggles.tsx && test $(grep -c 'ws-create-toggle-track {' frontend/src/workspace.css) -eq 1 && (cd frontend && npx tsc --noEmit && npm run build)` | ✅ `internal/config/config_test.go` | ✅ green |
| 26-05 Task 2: Copy-to-secondary toggle, one stored location | 26-05 | 5 | SET-04 | T-26-20 | Stored secondary directory is validated at write time by `StartScan`'s existing `EvalSymlinks` + `ContainsPath` gate, unchanged by where the path is stored | static assertion + live dev-browser | `grep -q 'Choose a folder when enabled' frontend/src/components/workspace/settings/CatalogSettingsSection.tsx && grep -q 'ws-create-note-path' frontend/src/components/workspace/settings/CatalogSettingsSection.tsx && grep -q 'settings.copyToSecondary' frontend/src/components/workspace/CreateSlideOver.tsx && (cd frontend && npx tsc --noEmit && npm run build)` | ✅ `CatalogSettingsSection.tsx` | ✅ green |
| 26-05 Task 3: Remember-window toggle (COMPAT-05) + phase validation matrix | 26-05 | 5 | SET-04, SET-05, COMPAT-05 | T-26-22, T-26-23 | Toggle drives the pre-existing `WindowPersistenceEnabled` field with no duplicate field; `Manager`'s `RWMutex` serialises a toggle flip racing `beforeClose`'s geometry write | unit (Go, whole-suite) + live dev-browser (real quit-and-relaunch) | `go build ./... && go test ./... -race -count=1 && (cd frontend && npx tsc --noEmit && npm run build) && grep -q 'setWindowPersistence' frontend/src/settingsStore.ts && ! grep -nE 'rememberWindow\|RememberWindow' internal/config/config.go app.go` | ✅ `internal/config/config_test.go` | ✅ green |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

### Requirement → verification approach (from RESEARCH.md)

| Req ID | Behavior | Test Type | Automated Command |
|--------|----------|-----------|-------------------|
| SET-01 | ⌘,/gear/theme chip open Settings | manual-only (live dev-browser) | dev-browser click + keyboard event against `:34115` |
| SET-02 | 11 theme cards apply immediately | manual-only + Go persistence test | `go test ./internal/config/... -run TestSetTheme` |
| SET-03 | Density/rail-position segmented controls | manual-only + Go persistence test | `go test ./internal/config/... -run 'TestSetDensity\|TestSetRailSide'` |
| SET-04 | Catalog directory, filename root, 4 toggles | manual-only + Go persistence tests | `go test ./internal/config/... -run 'TestSetCatalog\|TestSetWriteHTML\|TestSetCopyToSecondary\|TestSetWatchDirectory'` |
| SET-05 | Save-as-you-change, no debounce | unit (Go): each `Set*` calls `Save()` synchronously | `go test ./internal/config/...` |
| COMPAT-05 | Window persistence toggle continuity | unit (Go, existing) + manual restart check | `go test ./internal/config/... -run TestWindowPersistence` |
| T-22-05 | `CatalogModal.tsx` deleted, no dead listener | static assertion | `grep -rn openCatalogModal frontend/src` returns nothing |
| FU-23-A | `GetCatalogHtmlPath`/`OpenExternal` containment | unit (Go), table-driven, mirroring `reveal_test.go` | `go test ./... -run 'TestGetCatalogHtmlPath\|TestOpenExternal'` |

---

## Wave 0 Requirements

- [x] New table-driven cases in `internal/config/config_test.go` for the new `Set*`/`Get*` pairs (`density`, `railSide`, `catalogDirectory`, `defaultFilenameRoot`, `writeHTML`, `copyToSecondary`, `secondaryDirectory`, `watchDirectory`), following `newTestManager(t)`'s existing temp-dir fixture pattern (`config_test.go:11-30`)
- [x] Containment tests for `GetCatalogHtmlPath` / `OpenExternal` (`internal/osutil/openexternal_test.go`, `app_test.go`), following the existing `ContainsPath` table structure — covers `../` escape, symlink escape, and same-prefix sibling directory (26-04)
- [x] Migration test: localStorage-sourced values land in Go config on first run and are not re-migrated on subsequent runs (26-03 Task 3, proven live via a forced re-migration through the real `Set*` bindings)
- [x] No new frontend test file — none needed, consistent with TEST-01's deferral and every prior phase's precedent

---

## Manual-Only Verifications

All rows below were executed live via the dev-browser skill against a running `wails dev` on `:34115` (26-05 Task 3, 2026-08-15), with `Object.keys(window.go.main.App)` probed for binding freshness before each session before recording evidence.

| Behavior | Requirement | Why Manual | Test Instructions | Observed Result |
|----------|-------------|------------|-------------------|------------------|
| ⌘, / gear / theme chip all open Settings | SET-01 | No frontend test framework (TEST-01 deferred) | `wails dev` on `:34115`; dev-browser: press ⌘,, click gear, click theme chip — dialog opens from each | **PASSED.** All three triggers (gear button, theme chip, `Meta+,` keyboard shortcut) each opened the `dialog "Settings"` element, verified independently with an Escape close between each. |
| Theme applies immediately on card click across all 11 themes | SET-02 | Visual token application | Click each of the 11 cards; confirm tokens repaint with no reload and no flash | **PASSED.** Clicked all 11 theme cards (StorCat Light, StorCat Dark, Dracula, Solarized Dark, Solarized Light, Nord, One Dark, Monokai, GitHub Light, GitHub Dark, Gruvbox Dark) in sequence; `document.documentElement`'s `data-theme` attribute updated to the correct id after every single click, same tick, no reload. |
| Density and rail-position segmented controls take effect live | SET-03 | Visual/layout | Toggle each segment; confirm row height and rail side change immediately | **PASSED.** Clicking "Compact" changed the live `--rh` CSS custom property from `34px` to `27px` immediately; clicking "Left" changed `.ws-root`'s `data-rail-side` attribute from `Right` to `Left` immediately, and both new values were also confirmed via `GetConfig()`. |
| Settings survive an app restart (all fields) | SET-05 | Requires real process restart | Change every control, quit, relaunch, confirm every value persisted | **PASSED.** Set theme=`nord`, density=`Compact`, railSide=`Left`, writeHtml=`false`, watchDirectory=`true`, defaultFilenameRoot=`restarttest`, called `window.runtime.Quit()` (a real process exit, confirmed via `ps aux` showing the `wails dev` process gone), restarted `wails dev`, reconnected, and `GetConfig()` on the fresh process reported every one of those six values unchanged. |
| Window size/position persistence continues to work exactly as pre-milestone, gated by the toggle | COMPAT-05 | Requires real window manager + restart | Toggle on: resize/move, quit, relaunch → restored. Toggle off: resize/move, quit, relaunch → not restored | **PASSED, both directions, via real quit-and-relaunch cycles (not reasoning-only).** **ON:** persistence enabled, window set to 900×650 at (250,180) via `window.runtime.WindowSetSize/SetPosition`, `Quit()`'d (process confirmed exited), `config.json` on disk showed `windowWidth:900, windowHeight:650, windowX:250, windowY:180` immediately after exit, `wails dev` restarted, and the fresh process's live `window.runtime.WindowGetSize()`/`WindowGetPosition()` reported exactly `{w:900,h:650}` / `{x:250,y:180}` — restored. **OFF:** persistence disabled, window set to a *different* distinctive 700×500 at (60,60), `Quit()`'d — `config.json` after exit still read the **previous** ON-direction values (900×650 @ 250,180), proving the transient OFF-direction resize was never written and the previously stored geometry was not overwritten. `wails dev` restarted; the fresh process opened at the code's documented default (`1024×768`, from `main.go`'s `!cfg.WindowPersistenceEnabled` fallback) at an OS-placed position (`343,77`) — neither the transient 700×500/60,60 nor the persisted 900×650/250,180 — confirming no restoration occurred. Config/window state restored to sane defaults afterward. |
| Flash-free boot after the localStorage→config migration | (architecture risk, RESEARCH Pitfall 1) | Only observable as a first-paint artifact | Cold-launch on a non-default theme; confirm no default-theme flash before first paint | **PASSED.** Set `theme=dracula` in both the Go config and the `storcat-theme-id` localStorage boot cache (the real two-write shape `setThemeSetting` uses), navigated fresh to `:34115`, and read `document.documentElement`'s `data-theme` attribute **immediately** after navigation resolved (before any additional wait): `dracula`, not the default `light`/`storcat-light` — no flash observed at the earliest inspectable point, and it remained `dracula` after a 1s settle with zero console/page errors. |

---

## Validation Sign-Off

- [x] All tasks have `<automated>` verify or Wave 0 dependencies — all 12 tasks across the five plans have an `<automated>` command, listed in the Per-Task Verification Map above
- [x] Sampling continuity: no 3 consecutive tasks without automated verify — every task in the map has its own automated command; none are gapped
- [x] Wave 0 covers all MISSING references — config field tests, containment tests, and the migration test are all present and green (checked above)
- [x] No watch-mode flags — every `<automated>` command runs to completion and exits (`go test -count=1`, `tsc --noEmit`, `npm run build`); none invoke `--watch`
- [x] Feedback latency < 90s — the full suite command (`go build ./... && go test ./... -race -count=1 && (cd frontend && npx tsc --noEmit && npm run build)`) completes in well under 90s locally
- [x] `nyquist_compliant: true` set in frontmatter

**Approval:** granted — 2026-08-15, by the 26-05 executor after the full suite (`go build ./... && go test ./... -race -count=1`, `npx tsc --noEmit`, `npm run build`) ran green and every Manual-Only Verifications row above was executed live with a recorded observed result, including COMPAT-05's real quit-and-relaunch proof in both toggle directions.
