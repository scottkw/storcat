---
phase: 26-settings
fixed_at: 2026-08-15T10:45:00Z
review_path: .planning/phases/26-settings/26-REVIEW.md
iteration: 2
findings_in_scope: 2
fixed: 2
skipped: 0
status: all_fixed
---

# Phase 26: Code Review Fix Report

**Fixed at:** 2026-08-15T10:45:00Z
**Source review:** .planning/phases/26-settings/26-REVIEW.md
**Iteration:** 2

**Summary (this iteration):**
- Findings in scope: 2 (0 critical, 2 warning)
- Fixed: 2
- Skipped: 0

**Cumulative picture (iterations 1-2):**
- Iteration 1 fixed CR-01, WR-01, WR-02 (commits `423792df`, `766d1476`, `2637ff5e`) — see below for details, carried forward from `26-REVIEW-FIX.md` iteration 1.
- Iteration 2 (this run) fixed WR-A and WR-B, the two narrow-scope residuals the iteration-1 fixes surfaced on re-review.
- All 5 findings raised across both review passes are now fixed. No open findings remain in scope.

**Verification environment:** `workflow.use_worktrees` is `false` for this project, so all edits, verification, and commits ran directly in the main checkout (no isolated worktree was created). `go build ./... && go test ./... -race -count=1 && (cd frontend && npx tsc --noEmit && npm run build)` was run after each fix and again after both fixes landed; the numbers below are reproducible from this same checkout.

## Fixed Issues (this iteration)

### WR-A: `domReady` is the one `App` method that still dereferences `a.configManager` with no nil guard

**Files modified:** `app.go`
**Commit:** 090dc418
**Applied fix:** Added the same `if a.configManager == nil { return }` guard every sibling method (`GetConfig`, the ~14 setters, `beforeClose`) already opens with, ahead of the existing `cfg := a.configManager.Get()` call. Applied exactly as the review's fix suggestion — code context matched cleanly, no restructuring of `App`'s construction or an accessor layer was needed (per fix guidance, proportionate to a one-line defensive-consistency gap). `go build ./... && go test ./... -race -count=1` green before and after.

### WR-B: `SETTINGS_HYDRATED`'s field-aware merge can silently discard a deliberate user write that happens to equal the default value

**Files modified:** `frontend/src/contexts/AppContext.tsx`
**Commit:** 2354bce5
**Applied fix:** Added a `touchedSettings: Set<keyof AppSettings>` field to `AppState`, seeded empty. `SET_SETTINGS` now records every key it writes into this set — including when the incoming value happens to already match the current one (the exact WR-B scenario: a user's first explicit write landing back on the current/default value must still be recorded, so the reducer's pre-existing "unchanged" bail-out was refined to require both "value unchanged" *and* "already touched" before returning the identical state object). `SETTINGS_HYDRATED` now defers to `touchedSettings.has(key)` instead of the WR-01 heuristic's `state.settings[key] === DEFAULT_APP_SETTINGS[key]` comparison, closing the gap where a deliberate re-set-to-default during the hydration race window looked identical to "never touched" and got silently reverted by the stale in-flight `getConfig()` response.

This is the real fix, not the documented-residual fallback: the review's own suggested code (a `Set<keyof AppSettings>` of dispatched keys) turned out to be a genuinely small, self-contained mechanism once the `SET_SETTINGS` bail-out was adjusted to still record the touch on a same-value write — no gate on Settings-dialog entry points was needed, and no `hydrateSettings()` await/blocking change to `WorkspaceShell.tsx` was needed. `npx tsc --noEmit && npm run build` green before and after; `go build ./... && go test ./... -race -count=1` unaffected (frontend-only change) and re-confirmed green.

## Fixed Issues (carried forward from iteration 1)

### CR-01: `config.Manager.Get()` (and `GetWindowPersistence()`) panic on the app's own documented config-load-failure fallback

**Files modified:** `internal/config/config.go`, `internal/config/config_test.go`, `app.go`
**Commit:** 423792df
**Applied fix:** Added `NewDefaultManager()` to `internal/config/config.go`, constructing a `Manager` pre-loaded with `DefaultConfig()` (never a nil `config` field). Updated `NewApp()` in `app.go` to call `config.NewDefaultManager()` instead of the bare `&config.Manager{}` literal in the `config.NewManager()` error fallback.

### WR-01: `SETTINGS_HYDRATED` wholesale-replaces `state.settings`, racing a user's own in-flight toggle

**Files modified:** `frontend/src/contexts/AppContext.tsx`
**Commit:** 766d1476
**Applied fix:** Changed the `SETTINGS_HYDRATED` reducer case from a wholesale replace to a field-aware merge based on value-equals-default. (Iteration 2's WR-B fix above supersedes the merge condition with the `touchedSettings` set, closing the residual this heuristic left open — the underlying field-aware-merge structure from this fix is retained.)

### WR-02: "Watch catalog directory" toggle's copy claims live behavior this phase does not implement

**Files modified:** `frontend/src/components/workspace/settings/CatalogSettingsSection.tsx`
**Commit:** 2637ff5e
**Applied fix:** Reworded the toggle's note from `"refresh the rail automatically"` to `"applies once file watching ships"`.

## Skipped Issues

None — all in-scope findings across both iterations were fixed.

---

_Fixed: 2026-08-15T10:45:00Z_
_Fixer: Claude (gsd-code-fixer)_
_Iteration: 2_
