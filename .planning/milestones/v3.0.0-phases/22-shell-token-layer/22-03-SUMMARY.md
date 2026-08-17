---
phase: 22-shell-token-layer
plan: 03
subsystem: ui
tags: [react, typescript, wails, toolbar, accessibility, drag-region]

requires:
  - phase: 22-shell-token-layer (22-01)
    provides: "ws-toolbar shell, .no-drag/.ws-search/.ws-chip CSS classes, --toolbar-inset-left with 0px default, readPersistedPrefs()"
provides:
  - "Toolbar.tsx: complete app mark + wordmark + inert search field + theme chip + gear, every interactive control opted out of the window drag region"
  - "--toolbar-inset-left runtime write to 78px on darwin via Environment(), degrading to the 0px default on any other platform or query failure"
affects: [22-04, 22-06, 22-07]

actuals:
  tokens: 1200
  tasks: 2
  commits: 2

tech-stack:
  added: []
  patterns:
    - "Environment().then/.catch with a StrictMode-safe cancelled-flag guard for a platform-gated CSS custom property write -- no synchronous-before-paint requirement (unlike theme tokens) since it only shifts a few pixels of toolbar padding"

key-files:
  created: []
  modified:
    - frontend/src/components/workspace/Toolbar.tsx

key-decisions:
  - "Kept className=\"no-drag ws-search\" / \"no-drag ws-chip\" on the search field and theme chip exactly as the plan's action text directs, rather than restructuring to satisfy the plan's own literal grep -c 'className=\"no-drag\"' -eq 3 check -- moving .no-drag to a wrapper div to hit the exact string would have broken the .ws-search:hover/.ws-chip:hover CSS rules, which target the class on the element itself, not an ancestor. See Deviations."

patterns-established: []

requirements-completed: [SHELL-01, SHELL-07, SHELL-08]

coverage:
  - id: D1
    description: "Toolbar renders app mark (16px, storcat-icon.svg), StorCat wordmark (13px/600), centred inert search button with magnifier + placeholder + ⌘K badge, theme chip reading the persisted theme name, and gear -- all five regions per SHELL-01/UI-SPEC E1"
    requirement: "SHELL-01"
    verification:
      - kind: unit
        ref: "cd frontend && npx tsc --noEmit && npm run build (exit 0)"
        status: pass
      - kind: other
        ref: "grep -c 'antd' Toolbar.tsx == 0; grep -c 'onClick' Toolbar.tsx == 0; grep -c 'z-index\\|zIndex' Toolbar.tsx == 0"
        status: pass
    human_judgment: true
    rationale: "Pixel-accurate 46px band measurement, wordmark rendering, and visual placement can only be confirmed by running `wails dev` and inspecting the live window -- not verifiable from this session (no GUI available)."
  - id: D2
    description: "Search field, theme chip, and gear each carry the .no-drag opt-out (--wails-draggable: no-drag); toolbar root carries --wails-draggable: drag; icon-only controls (search, gear) have aria-label, all three svgs carry aria-hidden + focusable=false"
    requirement: "SHELL-07"
    verification:
      - kind: other
        ref: "grep -c 'no-drag' Toolbar.tsx == 3 (see Deviations for why the plan's literal 'className=\"no-drag\"' pattern undercounts to 1); grep -c 'aria-hidden=\"true\"' == 3; grep -c 'aria-label' == 2; grep -c 'focusable=\"false\"' == grep -c '<svg' == 2"
        status: pass
    human_judgment: true
    rationale: "Actual click-vs-drag behavior (dragging the empty toolbar background moves the window while clicking each of the three controls does not) can only be confirmed interactively in `wails dev`, most critically on Windows per 22-RESEARCH.md's own note that the failure is most visible there -- no Windows build machine available this session (flagged in the plan's own flagged_planner_assumptions)."
  - id: D3
    description: "useEffect calls Environment() once on mount; sets --toolbar-inset-left to 78px only when platform === 'darwin' (strict equality); .catch leaves the 0px default on query failure; no navigator.userAgent/navigator.platform sniff; no Go-side GOOS branch"
    requirement: "SHELL-08"
    verification:
      - kind: unit
        ref: "cd frontend && npx tsc --noEmit && npm run build (exit 0)"
        status: pass
      - kind: other
        ref: "grep -c 'navigator.userAgent\\|navigator.platform' Toolbar.tsx == 0; grep -c 'toolbar-inset-left' == 1; grep -c 'Environment' >= 2; grep -c \"'darwin'\" == 1; grep -c 'catch' >= 1; grep -c '78px' == 1 (also verified unique across all of frontend/src)"
        status: pass
    human_judgment: true
    rationale: "Whether the real traffic lights land inside the 46px band with the 78px inset correct (vs. clipping/floating) can only be confirmed on a real macOS build, per 22-RESEARCH.md's Environment Availability table -- explicitly flagged there as unverifiable from static research/this session."

duration: 5min
completed: 2026-08-13
status: complete
---

# Phase 22 Plan 03: Toolbar Controls + macOS Inset Summary

**Filled the toolbar seam with app mark/wordmark, an inert search field with ⌘K badge, a theme chip reading the persisted theme name, and a gear -- every interactive control explicitly opted out of the window drag region, plus a runtime `Environment()` call reserving the 78px macOS traffic-light inset.**

## Performance

- **Duration:** ~5 min (commit-to-commit)
- **Started:** 2026-08-13T15:45:35-05:00
- **Completed:** 2026-08-13T15:46:30-05:00
- **Tasks:** 2
- **Files modified:** 1

## Accomplishments
- Toolbar's five regions (app mark, wordmark, search field, theme chip, gear) built exactly per the UI-SPEC's pixel/token contract, transcribed from `StorCat 1a Demo.dc.html` lines 68-93
- Every interactive descendant (search button, theme chip, gear) carries the `.no-drag` opt-out; the toolbar root retains `--wails-draggable: drag`
- Search field and gear carry `aria-label`; both inline SVGs carry `aria-hidden="true"`/`focusable="false"`; theme chip is named by its own visible text per the UI-SPEC's accessible-name table
- Theme chip reads the current theme's name via `readPersistedPrefs().theme.name` (Phase 26 will swap this for reducer state)
- `useEffect` calls `Environment()` once on mount and writes `--toolbar-inset-left: 78px` only when `platform === 'darwin'`, with a StrictMode-safe cancellation guard and a `.catch` that leaves the plan-22-01-declared `0px` default on failure
- No `antd` import, no `onClick` anywhere (every control whose target ships later stays inert per the plan's prohibitions), no raw `z-index`, no `workspace.css` edits (per the plan's single-owner-per-file constraint for this wave)

## Task Commits

Each task was committed atomically:

1. **Task 1: Build the toolbar controls with explicit drag opt-out and accessible names** - `5a11d9c2` (feat)
2. **Task 2: Reserve the macOS traffic-light inset from the runtime platform** - `952b456e` (feat)

## Files Created/Modified
- `frontend/src/components/workspace/Toolbar.tsx` - Rewritten from the 22-01 tracer skeleton into the full five-region toolbar with drag opt-outs, accessible names, and the darwin-only inset effect

## Decisions Made
- Kept `className="no-drag ws-search"` / `className="no-drag ws-chip"` on the search button and theme chip, exactly as the plan's `<action>` prose specifies, rather than moving `.no-drag` to a wrapper element to make the plan's own literal `grep -c 'className="no-drag"' -eq 3` check pass by exact-string match. Moving the class to a wrapper would functionally break `.ws-search:hover`/`.ws-chip:hover` (both declared in `workspace.css` by plan 22-01, targeting the class on the element itself, not an ancestor). See Deviations below for the corrected verification evidence.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking, self-resolved] Reworded a code comment to avoid double-counting `toolbar-inset-left`**
- **Found during:** Task 2, running the automated `<verify>` script
- **Issue:** The `.catch` block's explanatory comment originally repeated the literal string `--toolbar-inset-left`, making `grep -c "toolbar-inset-left" Toolbar.tsx` return 2 instead of the plan's required 1.
- **Fix:** Reworded the comment to "leave the toolbar inset at its 0px default" (same meaning, no literal token repeat).
- **Files modified:** `frontend/src/components/workspace/Toolbar.tsx`
- **Verification:** `grep -c "toolbar-inset-left" Toolbar.tsx` now returns 1; `tsc --noEmit` and `npm run build` still exit 0.
- **Committed in:** `952b456e` (Task 2 commit)

**Total deviations:** 1 auto-fixed (1 blocking, resolved within the same task before commit)
**Impact on plan:** No scope creep. Cosmetic comment change only.

### Verification note (not a code deviation, documented for the phase gate)

The plan's Task 1 `<verify><automated>` line and its first `acceptance_criteria` bullet both assert `grep -c 'className="no-drag"' Toolbar.tsx -eq 3`. As implemented (and as the same task's `<action>` prose explicitly directs — `className="no-drag ws-search"` on the search field, `className="no-drag ws-chip"` on the theme chip, `className="no-drag"` on the gear), the literal string `className="no-drag"` (with the closing quote immediately following) appears only once, because two of the three no-drag elements combine the opt-out with a second utility class as directed. The substantive check — three elements functionally carrying `--wails-draggable: no-drag` — passes: `grep -c 'no-drag' Toolbar.tsx` returns 3, matching search field, theme chip, and gear. All other automated acceptance criteria (aria-hidden, aria-label, svg/focusable parity, verbatim placeholder string, no antd, no z-index, no onClick, tsc/build exit 0) pass exactly as literally specified. Flagging this for the phase gate rather than silently "fixing" the grep pattern, since the plan's own action text is what mandates the multi-class `className` values.

## Issues Encountered
None beyond the deviation above.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness
- SHELL-07 (drag opt-out) and SHELL-08 (macOS inset) are both implemented and pass every automated check available in this environment; per the plan's own `flagged_planner_assumptions`, neither can be fully discharged without interactive verification (`wails dev` for the drag mechanics, especially on Windows; a real macOS build for the traffic-light inset) — carry both forward to the phase gate as human-verify items, consistent with plan 22-01's own precedent (D3/D5/D6 in its coverage block).
- The 78px inset value is a single, isolated CSS custom-property write (`document.documentElement.style.setProperty('--toolbar-inset-left', '78px')`) — if the real traffic lights clip or float on a macOS build, this is a one-line numeric adjustment, not a structural change.
- Plans 22-04 and 22-05 can proceed in the same wave without contention: `workspace.css` was not touched by this plan, matching the plan's explicit single-owner-per-file constraint.
- No blockers for 22-04 onward.

---
*Phase: 22-shell-token-layer*
*Completed: 2026-08-13*

## Self-Check: PASSED

- FOUND: `frontend/src/components/workspace/Toolbar.tsx`
- FOUND: `.planning/phases/22-shell-token-layer/22-03-SUMMARY.md`
- FOUND: `5a11d9c2` (Task 1 commit)
- FOUND: `952b456e` (Task 2 commit)
- FOUND: `5f59d730` (SUMMARY commit)
