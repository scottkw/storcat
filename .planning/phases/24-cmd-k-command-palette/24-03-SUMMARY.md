---
phase: 24-cmd-k-command-palette
plan: 03
subsystem: ui
tags: [react, typescript, hooks, accessibility, command-palette]

# Dependency graph
requires:
  - phase: 24-cmd-k-command-palette
    provides: "24-01's always-mounted CommandPalette overlay (isOpen contract); 24-02's live-verified ⌘K/Ctrl+K and toolbar open paths"
provides:
  - "frontend/src/hooks/useModalBehavior.ts -- the single implementation of focus trap, Escape-to-close, scroll lock, and focus restore for every overlay in the app"
  - "CommandPalette.tsx wired to the shared hook with its tracer-era inline Escape listener and autoFocus removed"
affects: [25, 26, 27]

# Actuals (#2632)
actuals:
  tokens: 2118
  tasks: 2
  commits: 2

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "One useEffect keyed on exactly [isOpen] (not []): React fires its cleanup on the true->false transition while the consumer is still mounted, which is what an animated 260ms exit (Phase 25) requires -- an empty-array effect only cleans up at unmount and would leave the page scroll-locked for the whole exit animation"
    - "onClose held in a ref, kept current by a separate tiny effect -- keeps the main effect's dependency array at exactly [isOpen] so inline arrow-function props from the consumer don't thrash the scroll lock / re-capture the restore target on every parent render"
    - "Restore-target liveness check (isConnected) before refocusing in cleanup -- a trigger removed from the DOM while the overlay was open must not throw and must not steal focus to a detached node"

key-files:
  created:
    - frontend/src/hooks/useModalBehavior.ts
  modified:
    - frontend/src/components/workspace/CommandPalette.tsx

key-decisions:
  - "ModalBehavior.containerRef typed as React.RefObject<HTMLDivElement> (not | null) -- caught by tsc while wiring Task 2: React 18's RefObject<T> already nests null in `current`, so unioning the generic with null broke assignment to the JSX ref attribute. Fixed as part of Task 2's commit since it surfaced during that task's tsc check, not Task 1's acceptance criteria."
  - "Focusable-descendant selector matches anchors-with-href, non-disabled buttons, input/select/textarea, and any [tabindex] attribute, then filters programmatically on el.tabIndex !== -1 -- the plan's 'tabindex that is not negative one' requirement needed a runtime filter, not something expressible in the CSS selector alone"

patterns-established:
  - "Shared modal-behavior hook at frontend/src/hooks/useModalBehavior.ts -- Phases 25 (slide-over), 26 (Settings modal), and 27 (confirmation dialogs) import this hook and must not reimplement any of its four behaviors, per 24-CONTEXT.md's locked decision"

requirements-completed: [PLT-07]

coverage:
  - id: D1
    description: "useModalBehavior hook implements all four PLT-07 behaviors (focus trap, Escape-to-close, scroll lock, focus restore) with cleanup keyed on the isOpen transition, not unmount, and no new dependency"
    requirement: "PLT-07"
    verification:
      - kind: other
        ref: "cd frontend && npx tsc --noEmit && npm run build"
        status: pass
      - kind: other
        ref: "grep -c '}, \\[isOpen\\]);' frontend/src/hooks/useModalBehavior.ts == 1; grep -c '}, \\[\\]);' frontend/src/hooks/useModalBehavior.ts == 0"
        status: pass
      - kind: other
        ref: "git diff --stat -- frontend/package.json frontend/package-lock.json (produces no output)"
        status: pass
    human_judgment: false
  - id: D2
    description: "CommandPalette.tsx is the hook's first consumer: containerRef on .ws-palette-panel, initialFocusRef on the query input, tracer-era inline keydown listener and autoFocus both deleted -- zero bespoke modal behavior left in the component"
    requirement: "PLT-07"
    verification:
      - kind: other
        ref: "grep -c \"addEventListener('keydown'\" frontend/src/components/workspace/CommandPalette.tsx == 0; grep -ci autoFocus frontend/src/components/workspace/CommandPalette.tsx == 0"
        status: pass
      - kind: automated_ui
        ref: "dev-browser :34115 -- Tab cycled 8x never left .ws-palette-panel; .ws-root overflow 'hidden' while open, restored across 5 open/close cycles; Escape closed with caret mid-typing ('hello' preserved, scrim removed); focus restored to .ws-search after toolbar-opened close and to the rail filter after ⌘K-opened close; second ⌘K while open stayed a no-op with the in-progress query ('abc') intact"
        status: pass
    human_judgment: false

# Metrics
duration: ~12min
completed: 2026-08-14
status: complete
---

# Phase 24 Plan 03: Shared Modal Behavior Hook Summary

**One `useModalBehavior` hook now owns focus trap, Escape-to-close, scroll lock, and focus restore for every overlay in the app; `CommandPalette.tsx` is its first consumer with its tracer-era inline Escape listener and `autoFocus` deleted.**

## Performance

- **Duration:** ~12 min
- **Started:** 2026-08-14T15:19:00Z
- **Completed:** 2026-08-14T15:31:00Z
- **Tasks:** 2
- **Files modified:** 2 (1 created, 1 modified)

## Accomplishments
- `frontend/src/hooks/useModalBehavior.ts` -- a single `useEffect` keyed on `[isOpen]` implements all four PLT-07 behaviors: captures the pre-open `document.activeElement` as the restore target, moves initial focus synchronously in the same frame the panel mounts, locks scroll on `.ws-root` (or a caller-supplied selector) saving/restoring the prior inline `overflow`, and registers one `window` keydown listener handling Escape (checked first, fires from anywhere inside the overlay) and Tab (cycles only within the container's fresh-queried focusable descendants, reversing direction with Shift)
- `CommandPalette.tsx` wired to the hook via `useModalBehavior({ isOpen, onClose, initialFocusRef: inputRef })`, `containerRef` attached to `.ws-palette-panel` (the trap boundary, not the click-to-close scrim); the tracer's own `window` keydown listener and the input's `autoFocus` attribute are both deleted
- Zero new dependencies: `git diff --stat -- frontend/package.json frontend/package-lock.json` is empty, confirming the `focus-trap-react`/`react-focus-lock` rejection from `24-CONTEXT.md` held

## Task Commits

Each task was committed atomically:

1. **Task 1: The shared useModalBehavior hook, written for Phase 25's animated exit** - `af077820` (feat)
2. **Task 2: Wire the palette to the shared hook and delete its tracer-era inline handling** - `3e7deed3` (feat)

**Plan metadata:** commit follows this summary.

## Files Created/Modified
- `frontend/src/hooks/useModalBehavior.ts` - new shared hook: `ModalBehaviorOptions`, `ModalBehavior`, `FOCUSABLE_SELECTOR`, `useModalBehavior`
- `frontend/src/components/workspace/CommandPalette.tsx` - consumes the hook; inline Escape listener and `autoFocus` removed, `containerRef`/`initialFocusRef` wired

## Decisions Made
- `ModalBehavior.containerRef` typed as `React.RefObject<HTMLDivElement>` rather than `React.RefObject<HTMLDivElement | null>` -- the latter failed `tsc` when assigned to the JSX `ref` attribute, because React 18's `RefObject<T>` already permits `current: T | null` without the extra union. Caught and fixed while wiring Task 2.
- The focusable-descendant selector combines a CSS selector (`a[href], button:not([disabled]), input, select, textarea, [tabindex]`) with a runtime `el.tabIndex !== -1` filter, since "any tabindex that is not negative one" isn't expressible as a single CSS attribute selector.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] `ModalBehavior.containerRef`'s type broke `tsc` when wired to a real JSX ref**
- **Found during:** Task 2, running `npx tsc --noEmit` after attaching `ref={containerRef}` to `.ws-palette-panel`
- **Issue:** `containerRef: React.RefObject<HTMLDivElement | null>` in the `ModalBehavior` interface produced `TS2322: Type 'RefObject<HTMLDivElement | null>' is not assignable to type 'LegacyRef<HTMLDivElement>'` -- React 18's own `RefObject<T>` type already has `current: T | null`, so the extra `| null` on the generic parameter created a mismatched, non-assignable type rather than a redundant one.
- **Fix:** Changed the interface field to `React.RefObject<HTMLDivElement>` (dropping the redundant union); `useRef<HTMLDivElement>(null)`'s actual return type already matches this exactly.
- **Files modified:** `frontend/src/hooks/useModalBehavior.ts`
- **Verification:** `npx tsc --noEmit` exits 0; `npm run build` succeeds
- **Committed in:** `3e7deed3` (part of Task 2's commit, since it surfaced during that task's typecheck)

---

**Total deviations:** 1 auto-fixed (1x Rule 1 -- type-level bug, zero runtime behavior change).
**Impact on plan:** None on scope. The fix is a pure type annotation correction; the hook's runtime behavior, all Task 1 acceptance-criteria greps, and all Task 2 live checks are unaffected.

## Issues Encountered
None beyond the type-level fix documented above.

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- PLT-07 is fully satisfied and live-verified: `frontend/src/hooks/useModalBehavior.ts` is ready for Phase 25 (slide-over, animated exit), Phase 26 (Settings modal), and Phase 27 (confirmation dialogs) to import unchanged, per `24-CONTEXT.md`'s locked contract.
- `CommandPalette.tsx` now implements none of the four modal behaviors itself -- confirmed by grep (zero bespoke `keydown` listener, zero `autoFocus`) and by five consecutive live open/close cycles at `wails dev` :34115 leaving `.ws-root`'s inline overflow back at its original empty string every time.
- 24-04 (result-row polish) and 24-05 (reveal-to-tree) can build on this plan's `CommandPalette.tsx` without any further modal-behavior rework.
- Phase 25's slide-over will be the first real test of the hook's animated-exit contract (`isOpen`-keyed cleanup while still mounted) -- this plan proves the contract in the palette's synchronous unmount case; Phase 25 should specifically verify the 260ms exit does not leave `.ws-root` locked.

---
*Phase: 24-cmd-k-command-palette*
*Completed: 2026-08-14*

## Self-Check: PASSED

All 3 claimed files verified present on disk; all 3 claimed commit hashes (`af077820`, `3e7deed3`, `21ad9a89`) verified present in git log.
