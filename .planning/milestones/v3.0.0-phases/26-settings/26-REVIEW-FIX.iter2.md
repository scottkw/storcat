---
phase: 26-settings
fixed_at: 2026-08-15T00:00:00Z
review_path: .planning/phases/26-settings/26-REVIEW.md
iteration: 1
findings_in_scope: 3
fixed: 3
skipped: 0
status: all_fixed
---

# Phase 26: Code Review Fix Report

**Fixed at:** 2026-08-15T00:00:00Z
**Source review:** .planning/phases/26-settings/26-REVIEW.md
**Iteration:** 1

**Summary:**
- Findings in scope: 3 (1 critical, 2 warning)
- Fixed: 3
- Skipped: 0

**Verification environment:** `workflow.use_worktrees` is `false` for this project, so all edits, verification, and commits ran directly in the main checkout (no isolated worktree was created). `go build ./... && go test ./... -race -count=1 && (cd frontend && npx tsc --noEmit && npm run build)` was run after each fix and again after all three fixes landed; the numbers below are reproducible from this same checkout.

## Fixed Issues

### CR-01: `config.Manager.Get()` (and `GetWindowPersistence()`) panic on the app's own documented config-load-failure fallback

**Files modified:** `internal/config/config.go`, `internal/config/config_test.go`, `app.go`
**Commit:** 423792df
**Applied fix:** Added `NewDefaultManager()` to `internal/config/config.go`, constructing a `Manager` pre-loaded with `DefaultConfig()` (never a nil `config` field). Updated `NewApp()` in `app.go` to call `config.NewDefaultManager()` instead of the bare `&config.Manager{}` literal in the `config.NewManager()` error fallback. This restores the original "run in-memory with defaults" degrade-gracefully behavior for `Get()`, `GetWindowPersistence()`, and every `Set*` setter, which all previously nil-dereferenced `m.config` on this path. Added `TestNewDefaultManager_GetDoesNotPanic` as a regression pin (matches the reviewer's empirically-confirmed panic repro, now asserting it no longer panics and returns default values). Applied exactly as the review's fix suggestion — code context matched cleanly.

### WR-01: `SETTINGS_HYDRATED` wholesale-replaces `state.settings`, racing a user's own in-flight toggle

**Files modified:** `frontend/src/contexts/AppContext.tsx`
**Commit:** 766d1476
**Applied fix:** Changed the `SETTINGS_HYDRATED` reducer case from a wholesale `{ ...state, settings: action.payload }` replace to a field-aware merge: each key in the hydrated payload only overwrites `state.settings[key]` if that field is still sitting at its untouched `DEFAULT_APP_SETTINGS` value. This mirrors the same "don't clobber a user-changed value" guard `CreateSlideOver.tsx`'s own re-seed effect already uses for `root`/`writeHTML`/`copyToSecondary`, applied exactly as the review's fix suggestion (code context matched cleanly — no `WorkspaceShell.tsx` or `settingsStore.ts` changes were needed, since the race lives entirely in the reducer's merge strategy).

### WR-02: "Watch catalog directory" toggle's copy claims live behavior this phase does not implement

**Files modified:** `frontend/src/components/workspace/settings/CatalogSettingsSection.tsx`
**Commit:** 2637ff5e
**Applied fix:** Reworded the toggle's note from `"refresh the rail automatically"` (a functional claim the current build cannot keep — no fsnotify watcher exists until Phase 27) to `"applies once file watching ships"` (describes the setting as a stored preference for a future capability, consistent with the `WatchDirectory` field's own locked doc-comment constraint in `internal/config/config.go`). Label left unchanged per the review's own note that the label alone is not a false claim.

## Skipped Issues

None — all in-scope findings were fixed.

---

_Fixed: 2026-08-15T00:00:00Z_
_Fixer: Claude (gsd-code-fixer)_
_Iteration: 1_
