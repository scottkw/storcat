---
phase: 27-catalog-actions-watch
plan: 06
subsystem: catalog-actions
tags: [go, wails, fsnotify, watch, debounce, lifecycle]

# Dependency graph
requires:
  - phase: 27-catalog-actions-watch
    provides: "27-01's containment-gated binding shape; app.go's throttledProgress a.ctx==nil-guarded emit pattern and beforeClose's scan-cancellation precedent (27-03 SUMMARY confirmed the go directive is 1.23.4)"
provides:
  - "internal/watch.New(dir, debounce, onChange) (*Watcher, error) -- Wails-runtime-free fsnotify wrapper with a trailing-debounce coalescer, drained Errors channel, and sync.Once-guarded idempotent Close"
  - "App.applyWatchState() -- the single start/stop point for the watcher, wired to startup, SetWatchDirectory, SetCatalogDirectory"
  - "App.emitCatalogsChanged() -- the second and last runtime.EventsEmit call site in the repo, a.ctx==nil-guarded"
  - "App.shutdown(ctx) -- new main.go OnShutdown hook, releases the watcher unconditionally on every quit path"
  - "github.com/fsnotify/fsnotify v1.10.1 -- the phase's second new Go dependency"
affects: [27-07]

# Actuals (#2632)
actuals:
  tokens: 5638
  tasks: 2
  commits: 3

# Tech tracking
tech-stack:
  added:
    - "github.com/fsnotify/fsnotify v1.10.1 -- cross-platform directory watching (inotify/kqueue/ReadDirectoryChangesW), non-recursive by design"
  patterns:
    - "Trailing-debounce coalescer (mutex + resettable time.AfterFunc + once-guarded stop) as a small unexported type, the first debounce-timer mechanic in this repo beyond throttledProgress's simpler last-emit-time throttle"
    - "Wails-runtime-free backend package + single-emitter discipline: internal/watch never imports wailsapp/wails, app.go alone bridges its callback to runtime.EventsEmit, mirroring throttledProgress's a.ctx==nil guard"

key-files:
  created:
    - internal/watch/watcher.go
    - internal/watch/watcher_test.go
  modified:
    - go.mod
    - go.sum
    - app.go
    - main.go

key-decisions:
  - "OnShutdown: app.shutdown added to main.go's options.App, per this plan's own recorded_decision id=shutdown-hook -- beforeClose can return true to prevent the close (already does, for an active scan) so it is not a reliable unconditional cleanup hook; OnShutdown fires once, unconditionally, after the close is confirmed. beforeClose's scan-cancellation logic is untouched."
  - "shutdown() closes the watcher directly rather than routing through applyWatchState -- applyWatchState's job is start-or-restart against the current config, which is the wrong behavior on quit (nothing should restart). grep -c 'a.applyWatchState()' app.go is 3 (startup, SetWatchDirectory, SetCatalogDirectory), matching the plan's own documented fallback shape for this choice."
  - "SetWatchDirectory and SetCatalogDirectory both persist through configManager FIRST, then call applyWatchState() -- reversing the order would start/restart the watcher against stale config, since applyWatchState reads cfg fresh from configManager.Get()."
  - "A watcher that fails to start degrades silently (a.watcher stays nil, no error surfaced, no new event) -- 27-UI-SPEC.md's E4 resolution locks this: the status-bar indicator is optimistic, tied to the setting, not a confirmed running watcher. No watching:started/watching:failed event was added, matching 27-RESEARCH.md Pitfall 9's explicit scope-creep warning."
  - "Errors branch of the fsnotify select loop calls fireNow() (immediate emit) rather than logging and dropping -- an error including fsnotify.ErrEventOverflow gets the same recovery as a real change, since the idempotent full re-list has no delta to reconcile either way."

requirements-completed: [WATCH-02, WATCH-03]

coverage:
  - id: D1
    description: "A .json/.html create, write, remove or rename inside the watched directory produces exactly one catalogs:changed emission after the ~300ms trailing debounce; a burst (large copy) coalesces to one emission; the Errors channel is drained in the same select loop as Events and an error (including ErrEventOverflow) still triggers an emission"
    requirement: "WATCH-02"
    verification:
      - kind: unit
        ref: "internal/watch/watcher_test.go#TestWatcher_FiresOnJSONCreate"
        status: pass
      - kind: unit
        ref: "internal/watch/watcher_test.go#TestWatcher_FiresOnRemoveAndRename"
        status: pass
      - kind: unit
        ref: "internal/watch/watcher_test.go#TestWatcher_IgnoresUnrelatedExtension"
        status: pass
      - kind: unit
        ref: "internal/watch/watcher_test.go#TestWatcher_CoalescesAtomicWritePair"
        status: pass
      - kind: unit
        ref: "internal/watch/watcher_test.go#TestCoalescer_CollapsesBurst"
        status: pass
      - kind: unit
        ref: "internal/watch/watcher_test.go#TestCoalescer_ResetsOnEachTrigger"
        status: pass
      - kind: unit
        ref: "internal/watch/watcher_test.go#TestCoalescer_FiresNowIsImmediate"
        status: pass
      - kind: unit
        ref: "internal/watch/watcher_test.go#TestCoalescer_ConcurrentTriggers"
        status: pass
      - kind: live
        ref: "dev-browser against wails dev :34115 -- SetWatchDirectory(true), single external touch/rm each fired one catalogs:changed event (spaced >300ms apart, so two separate debounce windows, both correctly single-fire); a real 8-file cp burst fired exactly one event"
        status: pass
    human_judgment: false
  - id: D2
    description: "Turning watching off calls Watcher.Close() (genuine release, not merely ignoring events); changing the catalog directory stops the existing watcher and starts a new one; quitting the app releases the watcher via main.go's OnShutdown hook on every quit path; Close() is idempotent and safe on a watcher that never started"
    requirement: "WATCH-03"
    verification:
      - kind: unit
        ref: "internal/watch/watcher_test.go#TestWatcher_Close"
        status: pass
      - kind: unit
        ref: "internal/watch/watcher_test.go#TestWatcher_CloseIsIdempotent"
        status: pass
      - kind: unit
        ref: "internal/watch/watcher_test.go#TestCoalescer_StopCancelsPending"
        status: pass
      - kind: unit
        ref: "acceptance grep: awk over func (a *App) shutdown( body contains Close() -- app.go"
        status: pass
      - kind: live
        ref: "dev-browser: SetWatchDirectory(false) followed by an external touch+rm produced zero catalogs:changed events, confirming delivery genuinely stops"
        status: pass
      - kind: unit
        human_judgment: true
    human_judgment: true
    rationale: "The real OS-quit path (main.go's OnShutdown firing during an actual application quit) was not exercised live -- doing so would have killed the shared wails dev process this plan's own live-verification session depended on. Coverage rests on: (1) TestWatcher_Close/TestWatcher_CloseIsIdempotent proving Close()'s own semantics unit-level, (2) an acceptance grep proving shutdown()'s body actually calls Close() on the watcher field, and (3) main.go's OnShutdown: app.shutdown registration. The Wails-documented (not source-verified) semantics of OnShutdown itself is this plan's recorded_decision id=shutdown-hook's own residual risk, carried forward per that decision's text for 27-07's live quit check."
duration: ~12min
completed: 2026-08-15
status: complete
---

# Phase 27 Plan 06: File watching backend (WATCH-02, WATCH-03) Summary

**A Wails-runtime-free `internal/watch` package wraps `fsnotify` with a trailing-debounce coalescer and a drained error channel; `app.go`'s `applyWatchState` is the single start/stop point wired to launch, the watch-directory toggle, and catalog-directory changes, with a new `main.go` `OnShutdown` hook guaranteeing the watcher releases its OS handle on every quit path.**

## Performance

- **Duration:** ~12 min
- **Tasks:** 2
- **Files modified:** 6 (2 created, 4 modified)

## Accomplishments

- `internal/watch.New(dir, debounce, onChange) (*Watcher, error)` watches exactly one flat directory (never a recursive walk of its contents — catalogs are flat `.json`/`.html` files) and calls `onChange`, debounced by a trailing `~300ms` window (`DefaultDebounce`, a named constant), on any `Create`/`Write`/`Remove`/`Rename` event touching a `.json` or `.html` file. The `Errors` channel is drained in the same `select` as `Events`, and any error — including `fsnotify.ErrEventOverflow` — fires the callback immediately rather than being silently dropped or left unread (which could stall the watcher outright).
- **`internal/watch` imports nothing from `github.com/wailsapp/wails`** — verified by an acceptance grep (`! grep -rEq 'wailsapp/wails' internal/watch/`) — so it stays usable from the CLI with no Wails runtime attached. `app.go` remains the sole place in the repository that calls `runtime.EventsEmit` for this feature (`emitCatalogsChanged`, guarded by the same `a.ctx == nil` check `throttledProgress` already established), verified by `grep -rl 'runtime.EventsEmit' --include='*.go' .` returning only `app.go` for actual call sites (the string also appears once as prose inside `internal/watch/watcher.go`'s package doc comment, explaining the discipline it inherits — not an actual call).
- `Watcher.Close()` is `sync.Once`-guarded — idempotent and safe to call twice, or on a watcher whose constructor already failed — pinned by `TestWatcher_Close` and `TestWatcher_CloseIsIdempotent`, and reused directly by both the toggle-off path (`applyWatchState`) and the app-quit path (`shutdown`).
- `App.applyWatchState()` is the single start/stop point: it always `Close()`s any existing watcher first, unconditionally, before deciding whether to start a new one — so it is idempotent regardless of which of its three call sites (`startup`, `SetWatchDirectory`, `SetCatalogDirectory`) triggered it. A watcher that fails to construct (e.g. the configured directory no longer exists) degrades silently, leaving `a.watcher` nil with no new event surfaced — `27-UI-SPEC.md`'s E4 resolution locks this as deliberate, and `27-RESEARCH.md` Pitfall 9 explicitly names a `watching:started`/`watching:failed` event as scope creep this plan must not add.
- `main.go` gained `OnShutdown: app.shutdown` alongside the existing `OnStartup`/`OnDomReady`/`OnBeforeClose` — `app.shutdown` closes any live watcher under `watchMu` directly (not via `applyWatchState`, which would try to restart it — wrong on quit). `beforeClose` itself is untouched: `git diff app.go | grep -c '^-.*beforeClose'` is `0`, matching this plan's `<recorded_decision id="shutdown-hook">`, which explains why `beforeClose` (which can return `true` to *prevent* the close, as it already does for an active scan) is not a reliable place for an unconditional release.
- `github.com/fsnotify/fsnotify v1.10.1` added as an exact-pinned direct dependency — **the phase's second new Go dependency**, beyond `wastebasket` (27-03). `golang.org/x/sys` was already present at `v0.30.0` (indirect) and was promoted to direct at the same version with no bump; `go.mod`'s `go 1.23.4` directive (already bumped by 27-03 for `wastebasket`) needed no further change.

## Task Commits

1. **Task 1: internal/watch — RED** — `7d7bea9a` (test) — `internal/watch/watcher_test.go`, 13 cases (5 coalescer, 8 watcher)
2. **Task 1: internal/watch — GREEN** — `2ed48569` (feat) — `internal/watch/watcher.go`, `go.mod`, `go.sum`
3. **Task 2: app.go lifecycle + main.go OnShutdown** — `9475af0b` (feat) — `app.go`, `main.go`

**Plan metadata:** pending (this SUMMARY's own commit)

## Files Created/Modified

- `internal/watch/watcher.go` — `DefaultDebounce`, `Watcher`, `New`, `(*Watcher).Close`; unexported `coalescer` (`newCoalescer`/`trigger`/`fireNow`/`stop`)
- `internal/watch/watcher_test.go` — `TestCoalescer_*` (5 cases, always-on, fake trigger source), `TestWatcher_*` (8 cases, skippable under `-short`, real temp-directory fsnotify)
- `go.mod`, `go.sum` — `github.com/fsnotify/fsnotify v1.10.1` (direct, exact-pinned); `golang.org/x/sys v0.30.0` promoted indirect→direct, version unchanged
- `app.go` — `App.watchMu`/`App.watcher` fields; `emitCatalogsChanged`, `applyWatchState`, `shutdown` methods; `startup`, `SetWatchDirectory`, `SetCatalogDirectory` extended to call `applyWatchState()`
- `main.go` — `OnShutdown: app.shutdown` added to `options.App`

## What Was Actually Exercised vs. Reasoned About

**Exercised live** (dev-browser against a freshly rebuilt `wails dev` on `:34115`, binding freshness confirmed via `Object.keys(window.go.main.App)` before recording any evidence — `wails dev` was restarted for this plan since the process running at task start predated all of this plan's Go changes):
- `SetWatchDirectory(true)` genuinely starts a live watcher against the real configured catalog directory.
- An external `touch`+`rm` (via Bash, no host-OS GUI automation) each produced exactly one `catalogs:changed` event, subscribed via `window.runtime.EventsOn`.
- A real 8-file `cp` burst produced exactly **one** debounced event, not eight.
- `SetWatchDirectory(false)` stops delivery: a subsequent external `touch`+`rm` produced **zero** events.

**Reasoned about / unit-tested only, not exercised live:**
- The real OS-quit path (`main.go`'s `OnShutdown` actually firing during a genuine application quit, and the watcher's OS handle actually being released at the syscall level). Exercising this would have required quitting the shared `wails dev` process this plan's own live-verification session depended on. Coverage instead rests on `TestWatcher_Close`/`TestWatcher_CloseIsIdempotent` (unit-level `Close()` semantics) plus an acceptance grep confirming `shutdown()`'s body calls `Close()` on the watcher field and that `OnShutdown: app.shutdown` is registered in `main.go`'s `options.App` literal. The Wails-documented (not source-code-verified) semantics of `OnShutdown` itself remain this plan's own `<recorded_decision id="shutdown-hook">`'s stated residual risk.
- Catalog-directory change mid-watch (stop old, start new) — covered by `applyWatchState`'s unconditional-Close-then-maybe-start structure and its call from `SetCatalogDirectory`, but not exercised live in this session (no second catalog directory was switched to during the live pass above).

## What 27-07 Needs to Record in `.planning/WINDOWS.md`

This plan does **not** write `.planning/WINDOWS.md` entries itself — per this plan's own `<critical_constraints>`, plan 27-07 owns that file. Flagging what it needs to add:

1. **fsnotify's Windows non-removal-on-rename divergence.** `27-RESEARCH.md`'s Pitfall 4 (confirmed from `fsnotify.go`'s own doc comment, read during research): a watch is auto-removed when the watched path itself is deleted or renamed on macOS/Linux, but the Windows backend does **not** remove the watcher on a rename of the watched directory. This repo has no Windows machine to verify the resulting behavior (a dangling watch attached to a stale handle) — it is a genuine, unverified platform divergence, not something local code in `internal/watch` closes. `applyWatchState`'s existing "stop and restart on directory change" rule (the *user*-triggered path, via Settings) is unaffected; the gap is specifically the *watched directory itself* vanishing out from under an active watch with no explicit directory-change action.
2. **The real OS-quit `OnShutdown` release path is unverified live in this session** (see "What Was Actually Exercised" above) — this plan's own `<recorded_decision id="shutdown-hook">` names this as a residual risk to be "verified behaviorally by the live quit check in plan 27-07 rather than assumed." 27-07 is the natural place to perform an actual application-quit-and-confirm-no-leaked-handle pass, if its own scope touches a live `wails dev` session again.

Both items are genuine, honestly-scoped gaps — not overclaimed as done, per this project's established `WINDOWS.md` convention (entries #1, #4, #5 are the same class of "compiles/reasoned-about, unverified on the actual platform" gap).

## Decisions Made

- **`OnShutdown: app.shutdown` added to `main.go`**, exactly as this plan's own `<recorded_decision id="shutdown-hook">` specifies — `beforeClose` is left completely untouched (verified: `git diff app.go | grep -c '^-.*beforeClose'` returns `0`).
- **`shutdown()` closes the watcher directly rather than calling `applyWatchState()`** — `applyWatchState` decides whether to *restart* based on the current config, which is wrong behavior on quit (nothing should restart after the app is closing). This is the plan's documented fallback shape: `grep -c 'a.applyWatchState()' app.go` returns `3` (startup, `SetWatchDirectory`, `SetCatalogDirectory`), and `shutdown`'s own body contains its own `Close()` call, matching the acceptance criteria's explicit alternative branch for this choice.
- **`SetWatchDirectory`/`SetCatalogDirectory` persist through `configManager` before calling `applyWatchState()`**, in that order — `applyWatchState` reads `cfg := a.configManager.Get()` fresh, so persisting first is what makes the subsequent read see the new value; reversing the order would start/restart the watcher against stale config.
- **`wails dev` was restarted mid-plan** rather than trusted from a prior plan's session, per this project's own standing constraint ("`curl` liveness is not binding freshness"). The process running at task start predated this plan's `app.go`/`main.go` changes entirely; freshness was reconfirmed via `Object.keys(window.go.main.App)` after the restart, before any live evidence was recorded.

## Deviations from Plan

None — plan executed as written. All 13 test cases passed on the first implementation attempt (no RED→fix→GREEN iteration needed beyond the initial RED-by-non-compilation step, which is the expected RED shape for a from-scratch package).

## Issues Encountered

None.

## Next Phase Readiness

- `catalogs:changed` is live and emitting correctly from the backend; plan 27-07 (frontend subscription in `CatalogRail.tsx`, reusing its existing `loadCatalogsForDirectory` per `27-RESEARCH.md`'s Code Examples) can subscribe with confidence the event fires exactly once per debounced burst.
- The two `.planning/WINDOWS.md` items above are ready for 27-07 to record.
- `.planning/WINDOWS.md` entry #6 (atomicwrite SIGKILL crash-safety) was separately proven dischargeable by 27-02's real subprocess harness (42+ real kill iterations, per that plan's SUMMARY) — also queued for 27-07 to mark `fixed`, per this plan's own `<critical_constraints>` note; not touched by this plan since it is out of this plan's `files_modified` scope.

## Self-Check: PASSED

All 2 created files verified present on disk (`internal/watch/watcher.go`, `internal/watch/watcher_test.go`); all 3 commits (`7d7bea9a`, `2ed48569`, `9475af0b`) verified present in `git log`.

---
*Phase: 27-catalog-actions-watch*
*Completed: 2026-08-15*
