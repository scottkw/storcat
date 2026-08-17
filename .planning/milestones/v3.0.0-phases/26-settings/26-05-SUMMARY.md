---
phase: 26-settings
plan: 05
subsystem: ui
tags: [wails, react, typescript, go, config, settings, toggles, compat]

# Dependency graph
requires:
  - phase: 26-settings
    provides: "26-03's AppSettings shape, hydrateSettings() migration, CatalogSettingsSection.tsx, settingsStore write-through pattern; 26-04's completed security hardening"
provides:
  - "config.Config.WriteHTML / CopyToSecondary / WatchDirectory fields + Manager setters + App bindings"
  - "OptionsToggles.ToggleRow exported as the one shared toggle-switch control (noteMono prop) for both the create form and Settings"
  - "CatalogSettingsSection's four locked-order toggles: write HTML, copy to secondary, watch directory, remember window"
  - "settingsStore.setWriteHtmlSetting / setCopyToSecondarySetting / setWatchDirectorySetting / setRememberWindowSetting"
  - "CreateSlideOver's writeHTML/copyToSecondary option defaults seeded from state.settings, with the always-mounted re-seed guard 26-03 established"
  - "26-VALIDATION.md's completed Per-Task Verification Map (all 12 tasks across the phase) and executed Manual-Only Verifications matrix"
affects: [27, 28]

# Actuals (#2632)
actuals:
  tokens: 11560
  tasks: 3
  commits: 4

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "ToggleRow promoted from a private OptionsToggles.tsx function to an exported, shared component with an optional noteMono prop (defaults true) -- Settings' three descriptive-sentence notes pass false, its one path-valued note (copy-to-secondary, once set) leaves the default, matching 26-UI-SPEC's resolution of the design handoff's own internal toggle-geometry inconsistency"
    - "CreateSlideOver's always-mounted-initializer race (26-03's root-field finding) generalizes to any settings-seeded default: options.writeHTML and options.copyToSecondary get the same isOpen-keyed re-seed effect root already had, guarded on 'still at the built-in module default' (a boolean has no natural blank sentinel, so the guard compares against DEFAULT_OPTIONS instead of an empty string)"
    - "COMPAT-05 continuity pattern: a Settings toggle can drive a pre-existing config field and binding with zero new Go surface -- settingsStore.setRememberWindowSetting is a one-line pass-through to the already-existing wailsAPI.setWindowPersistence, its first call site"
    - "Real dev-browser proof of window-geometry persistence via window.runtime.Quit()/WindowSetSize/WindowSetPosition/WindowGetSize/WindowGetPosition plus a real wails dev process restart, not a reasoning-only claim -- extends the window.runtime.Quit() precedent 25-07 established for CRT-13"

key-files:
  created: []
  modified:
    - internal/config/config.go
    - internal/config/config_test.go
    - app.go
    - frontend/wailsjs/go/main/App.d.ts
    - frontend/wailsjs/go/main/App.js
    - frontend/wailsjs/go/models.ts
    - frontend/src/services/wailsAPI.ts
    - frontend/src/settingsStore.ts
    - frontend/src/components/workspace/create/OptionsToggles.tsx
    - frontend/src/components/workspace/settings/CatalogSettingsSection.tsx
    - frontend/src/components/workspace/CreateSlideOver.tsx
    - frontend/src/workspace.css
    - .planning/phases/26-settings/26-VALIDATION.md

key-decisions:
  - "WindowPersistenceEnabled deliberately left as the sole window-persistence field throughout -- COMPAT-05's toggle drives it directly via the pre-existing SetWindowPersistence binding, with grep-pinned acceptance criteria (internal/config/config.go and app.go contain no 'rememberWindow'/'RememberWindow' identifier) and a TestDefaultConfig_ToggleDefaults assertion that WindowPersistenceEnabled still defaults true"
  - "Added TestSetWindowPersistence_Idempotent (Rule 2) beyond the plan's literal Task 3 test list -- the plan's own must_haves name COMPAT-05 idempotency explicitly ('flipping the remember-window toggle to the value it already holds writes... byte-identical config.json content'), and the existing SetWindowPersistence tests didn't cover it the way TestSetDensity_Idempotent already covers density"
  - "COMPAT-05's both-direction restart proof used window.runtime.WindowSetSize/SetPosition (legitimate Wails JS runtime calls, not host-OS GUI automation) to set distinctive geometry, window.runtime.Quit() to trigger a real beforeClose-driven save/skip, and an actual wails dev process restart -- verified via window.runtime.WindowGetSize/GetPosition on the fresh process, not by reasoning about the code alone"
  - "The native OS folder-picker branch of the copy-to-secondary toggle (nothing stored -> enable -> dialog opens) was not driven through a real UI click, since SelectDirectory's runtime.OpenDirectoryDialog is a blocking, GUI-owned native call the standing no-host-OS-GUI-automation constraint forbids interacting with via CDP/keystrokes -- verified instead by tracing the exact code branch (mirrors OptionsToggles.handleToggleSecondary's already-proven logic) and exercising every other branch (already-stored reuse, disable, cancel path's absence of state mutation) live"
  - "CatalogSettingsSection's copy-to-secondary handlers (handleToggleCopyToSecondary, handleEditSecondaryPath) duplicate OptionsToggles.handleToggleSecondary/handleEditSecondaryPath's logic rather than extracting a shared hook -- the two components dispatch to different state trees (AppContext.state.settings vs. local useState props), so a shared hook would need an abstraction layer neither call site asked for; the plan's own action text specifies reproducing the logic, not extracting it"

requirements-completed: [SET-04, SET-05, COMPAT-05]

coverage:
  - id: D1
    description: "The four catalog toggles (write HTML, copy to secondary, watch directory, remember window) are all settable from Settings, in the locked order, each independently, through one shared ToggleRow component"
    requirement: "SET-04"
    verification:
      - kind: unit
        ref: "internal/config/config_test.go#TestSetWriteHTML, TestSetWriteHTML_Persists, TestSetWatchDirectory, TestSetWatchDirectory_Persists, TestSetCopyToSecondary, TestDefaultConfig_ToggleDefaults"
        status: pass
      - kind: automated_ui
        ref: "dev-browser session against :34115: opened Settings, confirmed all four toggle labels in locked order (Write HTML alongside JSON / Copy catalogs to a secondary location / Watch catalog directory for changes / Remember window size & position); clicked write-HTML and watch-directory toggles -- both flipped and GetConfig() reflected the change; grep confirms exactly one role=\"switch\" implementation in frontend/src"
        status: pass
    human_judgment: false
  - id: D2
    description: "The write-HTML and copy-to-secondary toggle values are the defaults the create slide-over opens with -- a Settings change is visible the next time a catalog is created"
    requirement: "SET-04"
    verification:
      - kind: automated_ui
        ref: "dev-browser session: turned Settings' write-HTML toggle off, opened the create slide-over -- its 'Also write HTML catalog' switch had no [checked] attribute (opened off); set copyToSecondary+secondaryDirectory via the real two-write shape then reloaded (fresh mount, avoiding the always-mounted-instance race) and opened the create panel -- its 'Copy both files to secondary location' switch was [checked] with the identical stored path"
        status: pass
    human_judgment: false
  - id: D3
    description: "Copy-to-secondary: turning the toggle off never clears the stored path; turning it on with a path already stored reuses it with no dialog; turning it on with nothing stored opens the native picker and a cancel leaves the toggle off with no error"
    requirement: "SET-04"
    verification:
      - kind: automated_ui
        ref: "dev-browser session: disabled the toggle with a path stored -- GetConfig() still reported the same secondaryDirectory, only copyToSecondary flipped to false; re-enabled with the same path stored -- returned immediately (no dialog hang), copyToSecondary true, path unchanged"
        status: pass
      - kind: other
        ref: "code-path trace of CatalogSettingsSection.handleToggleCopyToSecondary (mirrors the already-proven OptionsToggles.handleToggleSecondary): the empty-state enable branch calls wailsAPI.selectDirectory() and returns early on a falsy result with no dispatch and no error state -- not driven through a real click, since the underlying SelectDirectory binding opens a blocking native OS dialog the standing no-host-OS-GUI-automation constraint forbids interacting with"
        status: pass
    human_judgment: true
    rationale: "The native-picker-opens and cancel-leaves-it-off branches were verified by direct code inspection plus exercising every other reachable branch live, not by triggering the actual OS dialog (forbidden by the standing constraint after a prior-phase incident where automated GUI interaction hit an unintended window) -- a human should confirm the real dialog behaves as read, since this specific branch could not be driven end-to-end in this session."
  - id: D4
    description: "COMPAT-05: the 'Remember window size & position' toggle drives the pre-existing windowPersistenceEnabled field with zero new Go surface, and window geometry persistence works exactly as pre-milestone in both toggle states, proven by real quit-and-relaunch cycles"
    requirement: "COMPAT-05"
    verification:
      - kind: unit
        ref: "internal/config/config_test.go#TestSetWindowPersistence, TestSetWindowPersistence_Persists, TestSetWindowPersistence_Idempotent, TestDefaultConfig_ToggleDefaults; grep confirms zero 'rememberWindow'/'RememberWindow' identifiers in internal/config/config.go and app.go"
        status: pass
      - kind: automated_ui
        ref: "dev-browser session, two real quit-and-relaunch cycles via window.runtime.Quit() + a genuine wails dev process restart (confirmed via ps aux): ON -- window set to 900x650@(250,180), quit, config.json read 900/650/250/180, restarted, fresh process's WindowGetSize/GetPosition read back exactly 900x650@(250,180). OFF -- window set to a different 700x500@(60,60), quit, config.json still read the prior 900/650/250/180 (transient resize never saved), restarted, fresh process opened at the code's documented 1024x768 default at an OS-placed position -- neither transient value restored"
        status: pass
    human_judgment: false
  - id: D5
    description: "The full phase validation record (26-VALIDATION.md) is filled in with the real task IDs, requirements, threat refs and automated commands for all 12 tasks across all five plans, and every Manual-Only Verifications row has a recorded observed result"
    requirement: "SET-04"
    verification:
      - kind: other
        ref: ".planning/phases/26-settings/26-VALIDATION.md's Per-Task Verification Map (12 rows) and Manual-Only Verifications table (6 rows, each with an Observed Result column filled), Wave 0 and Validation Sign-Off checkboxes all ticked, nyquist_compliant: true / status: validated set in frontmatter"
        status: pass
    human_judgment: false

duration: ~27min
completed: 2026-08-15
status: complete
---

# Phase 26 Plan 05: Catalog Toggles, COMPAT-05, and Phase Validation Summary

**Four catalog toggles (write HTML, copy to secondary, watch directory, remember window) go live in Settings through one shared `ToggleRow` component; the remember-window toggle drives the pre-existing `windowPersistenceEnabled` field with zero new Go surface; and the phase's entire manual verification matrix — including COMPAT-05's persistence proven in both directions by real quit-and-relaunch cycles — is executed live and recorded.**

## Performance

- **Duration:** ~27 min
- **Started:** 2026-08-15T09:38:00-05:00 (approx.)
- **Completed:** 2026-08-15T10:01:25-05:00
- **Tasks:** 3
- **Files modified:** 13

## Accomplishments

- `config.Config` gained `WriteHTML` (default `true`), `CopyToSecondary` (default `false`) and `WatchDirectory` (default `false`), each with a lock-guarded `Manager` setter and `App` binding, regenerated via `wails generate module`. `WindowPersistenceEnabled` deliberately untouched throughout the plan.
- `OptionsToggles.ToggleRow` promoted from a private function to the phase's one shared, exported toggle-switch control (`ToggleRowProps`, new optional `noteMono` prop) — `CatalogSettingsSection` imports it rather than declaring a second toggle implementation; confirmed by grep there is exactly one `role="switch"` markup site in the whole frontend.
- `CatalogSettingsSection.tsx` renders all four toggles in the locked order (write HTML, copy to secondary, watch directory, remember window), the copy-to-secondary row reproducing `OptionsToggles.handleToggleSecondary`'s exact bootstrap-on-first-enable logic against the same stored `secondaryDirectory` value the create form's own toggle reads and writes.
- `CreateSlideOver` seeds its `writeHTML`/`copyToSecondary` option defaults from `state.settings`, with an isOpen-keyed re-seed effect (guarded on "still at the built-in module default", the boolean analogue of 26-03's blank-string guard for the always-mounted-initializer race) so a Settings change reaches the next create attempt even mid-session.
- `settingsStore.ts` gained `setWriteHtmlSetting`/`setCopyToSecondarySetting`/`setWatchDirectorySetting` (config-only, no boot cache — nothing reads them pre-paint) and `setRememberWindowSetting` (a one-line pass-through to the pre-existing `wailsAPI.setWindowPersistence`, its first call site). `hydrateSettings()` now maps all six `AppSettings` fields from real config values instead of `DEFAULT_APP_SETTINGS` fallbacks.
- `.planning/phases/26-settings/26-VALIDATION.md` fully populated: a 12-row Per-Task Verification Map covering every task across all five plans in the phase, Wave 0 requirements ticked, and every Manual-Only Verifications row executed live with a recorded observed result. `status: validated`, `nyquist_compliant: true`.
- Added `TestSetWindowPersistence_Idempotent` beyond the plan's literal test list (Rule 2) — the plan's own `must_haves` name COMPAT-05 idempotency explicitly, and no existing test covered it the way `TestSetDensity_Idempotent` already covers density.

## Task Commits

Each task was committed atomically, plus one deviation commit:

1. **Task 1: Write-HTML and watch-directory toggles, on the one shared toggle control** - `039e6250` (feat)
2. **Task 2: Copy-to-secondary toggle — one stored location shared with the create form** - `66027750` (feat)
3. **Task 3: Remember-window toggle (COMPAT-05) and the phase's live verification matrix** - `b9540419` (feat)
4. **[Rule 2] Pin COMPAT-05 idempotency for SetWindowPersistence** - `89d818cd` (test)

**Plan metadata:** (this commit)

## Files Created/Modified

- `internal/config/config.go` - `WriteHTML`/`CopyToSecondary`/`WatchDirectory` fields + three setters
- `internal/config/config_test.go` - 10 new tests (write-HTML/watch-directory set/persist, copy-to-secondary set, toggle defaults, window-persistence idempotency)
- `app.go` - three new Wails bindings (`SetWriteHTML`, `SetCopyToSecondary`, `SetWatchDirectory`)
- `frontend/wailsjs/go/main/App.d.ts`, `App.js`, `frontend/wailsjs/go/models.ts` - regenerated via `wails generate module`
- `frontend/src/services/wailsAPI.ts` - `setWriteHTML`/`setCopyToSecondary`/`setWatchDirectory` wrappers
- `frontend/src/settingsStore.ts` - four new setter functions, `hydrateSettings()` now reads all six real config fields
- `frontend/src/components/workspace/create/OptionsToggles.tsx` - `ToggleRow`/`ToggleRowProps` exported with `noteMono`
- `frontend/src/components/workspace/settings/CatalogSettingsSection.tsx` - all four toggle rows
- `frontend/src/components/workspace/CreateSlideOver.tsx` - `writeHTML`/`copyToSecondary` option defaults seeded from settings, with the isOpen-keyed re-seed guard
- `frontend/src/workspace.css` - `ws-settings-toggle-list` (row spacing only; every other rule reused verbatim from the create form's toggle block)
- `.planning/phases/26-settings/26-VALIDATION.md` - Per-Task Verification Map, Wave 0, Manual-Only Verifications with observed results, Validation Sign-Off

## Decisions Made

- `WindowPersistenceEnabled` remains the sole window-persistence field for the whole plan — pinned by grep-based acceptance criteria and a `TestDefaultConfig_ToggleDefaults` assertion that it still defaults `true`
- Added `TestSetWindowPersistence_Idempotent` (Rule 2) to close a must_have truth the plan's literal Task 3 test list didn't explicitly enumerate
- COMPAT-05's both-direction restart proof used `window.runtime.WindowSetSize/SetPosition/Quit` (legitimate JS-side Wails runtime calls, the same code path a real user resize/quit takes) plus an actual `wails dev` process restart — not a reasoning-only claim
- The native OS folder-picker branch of copy-to-secondary was verified by code-path tracing rather than a real click, since triggering it would open a blocking native dialog the standing no-host-OS-GUI-automation constraint forbids interacting with; every other branch (reuse-with-no-dialog, disable-preserves-path, and — via a faithful two-write simulation — the "opens with the persisted path" cross-surface read) was proven live
- `CatalogSettingsSection`'s copy-to-secondary handlers duplicate rather than share `OptionsToggles`' logic, since the two components dispatch to different state trees (`AppContext.state.settings` vs. local `useState` props) and the plan's action text specifies reproducing the logic, not extracting a hook

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 2 - Missing coverage for a locked must_have] Added `TestSetWindowPersistence_Idempotent`**
- **Found during:** Task 3, reviewing `must_haves.truths` against the actual test list before writing the SUMMARY
- **Issue:** The plan's `must_haves.truths` names "COMPAT-05 idempotency: flipping the remember-window toggle to the value it already holds writes the same value and produces byte-identical config.json content" as a truth to verify, but Task 3's own `<behavior>`/action text didn't enumerate a dedicated test for it, and no pre-existing test covered `SetWindowPersistence` the way `TestSetDensity_Idempotent` already covers `SetDensity`.
- **Fix:** Added `TestSetWindowPersistence_Idempotent`, identical shape to `TestSetDensity_Idempotent` — two consecutive `SetWindowPersistence(true)` calls, asserting the on-disk file is byte-identical.
- **Files modified:** `internal/config/config_test.go`
- **Verification:** `go test ./internal/config/... -race -count=1 -run TestSetWindowPersistence` — 3/3 pass; whole-repo `go test ./... -race -count=1` green.
- **Committed in:** `89d818cd` (separate commit, own reason line)

---

**Total deviations:** 1 auto-fixed (Rule 2 — missing test coverage for a locked must_have truth)
**Impact on plan:** No scope creep; closes a gap between the plan's own stated truths and its literal task instructions.

## Issues Encountered

- My first live-proof attempt at the copy-to-secondary cross-surface check used the raw `SetSecondaryDirectory` Go binding alone (bypassing `setSecondaryDirectorySetting`'s two-write shape), which left the `storcat-secondary-directory` localStorage boot cache stale from an earlier session. Since `CreateSlideOver` is always-mounted and only reads that cache once at its own first mount (per 26-03's precedent, deliberately unchanged by this plan), the create form's own toggle showed a stale path from a prior test run rather than the one I'd just set. Root cause was my test method, not the shipped code: redid the check mirroring the real two-write shape (config + localStorage) followed by a fresh page reload, which produced the correct cross-surface result.
- `dev-browser`'s globally-installed shim at `/opt/homebrew/lib/node_modules/dev-browser` was missing its native binary; the working install lived under the active nvm Node version's global `node_modules`. Invoked via the full nvm-relative path for the whole session.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- SET-04, SET-05 and COMPAT-05 are all closed. Phase 26 (Settings) is complete: all five plans (theme/density tracer, theme grid + rail position, catalog directory + migration, security hardening, catalog toggles + COMPAT-05) have landed.
- Phase 27 (fsnotify watch, catalog actions) can now wire the watch-directory toggle's persisted `WatchDirectory` value to a real watcher — this plan intentionally added no watcher, no status-bar indicator and no "watching" copy anywhere, per the plan's own prohibition and T-26-21's mitigation.
- `.planning/WINDOWS.md` gains no new entry from this plan — every row in the Manual-Only Verifications matrix, including COMPAT-05, was executed live with a genuine real-restart proof rather than deferred. Entry #7 (26-02's RailSide restart proof, deferred to avoid disrupting the shared `wails dev` process other plans in this phase depended on) remains open and unaffected by this plan's scope; it can now be swept independently since this was the phase's last plan and the shared-process constraint no longer applies.
- The local dev `wails dev` process and its `config.json` were both restored to sane values (theme `gruvbox-dark`, density `Comfortable`, railSide `Right`, writeHtml `true`, watchDirectory `false`, copyToSecondary `false`, window persistence `true` at the original 1470x923@(120,31) geometry) after the COMPAT-05 restart proofs, leaving the dev environment as found.

---
*Phase: 26-settings*
*Completed: 2026-08-15*

## Self-Check: PASSED

All created/modified files verified present on disk; all four task commit hashes (`039e6250`, `66027750`, `b9540419`, `89d818cd`) verified present in `git log`.
