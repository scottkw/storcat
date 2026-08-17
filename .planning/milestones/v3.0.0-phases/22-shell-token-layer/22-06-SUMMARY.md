---
phase: 22-shell-token-layer
plan: 06
subsystem: ui
tags: [react, typescript, wails, css-custom-properties, hooks, responsive]

requires:
  - phase: 22-shell-token-layer (22-01)
    provides: "AppContext reducer skeleton, workspace.css min-width ladder and shell classes, Density/RailSide types + readPersistedPrefs/applyTokens in themeTokens.ts, DevStateSwitcher"
provides:
  - "AppContext.tsx: pruned reducer (16 dead tab-era fields removed) plus density, railSide, detailOverlay state and SET_DENSITY/SET_RAIL_SIDE/SET_DETAIL_OVERLAY actions, seeded from readPersistedPrefs()"
  - "hooks/useMediaQuery.ts: the codebase's first hook, matchMedia-subscription based"
  - "workspace.css data-rail-side=\"Right\" swap (grid template, order, divider edge), scoped inside the 1280px block only"
  - "WorkspaceShell: single useMediaQuery('(min-width: 1280px)') call, data-rail-side attribute on .ws-root, density-reapply effect"
  - "DevStateSwitcher: Ctrl+Alt+R rail-side toggle, Ctrl+Alt+D now dispatches into the reducer instead of calling applyTokens directly"
affects: [22-07]

actuals:
  tokens: 3300
  tasks: 3
  commits: 3

tech-stack:
  added: []
  patterns:
    - "useMediaQuery(query) subscribes via matchMedia's 'change' event, never a resize listener -- fires only on threshold crossings"
    - "Reducer-owned preference (density) drives a downstream effect (WorkspaceShell) that re-invokes the token layer, rather than the dispatcher applying tokens itself -- keeps the reducer the single source of truth"
    - "Rail-side CSS swap scoped entirely inside the widest media-query block, so the base rules (rail=left, right-edge divider) apply unchanged below that tier regardless of the stored preference"

key-files:
  created:
    - frontend/src/hooks/useMediaQuery.ts
  modified:
    - frontend/src/contexts/AppContext.tsx
    - frontend/src/workspace.css
    - frontend/src/components/workspace/WorkspaceShell.tsx
    - frontend/src/components/dev/DevStateSwitcher.tsx

key-decisions:
  - "DevStateSwitcher keeps theme as local component state (Phase 26 still owns theme's eventual reducer/Settings home) but reads density/railSide straight from useAppContext().state instead of keeping a shadow copy -- avoids two sources of truth for the two fields this plan moved into the reducer"
  - "WorkspaceShell's density-reapply effect re-reads readPersistedPrefs().theme on every density change rather than threading theme through props/context -- matches the plan's explicit 'for now' instruction ahead of Phase 26's Settings-owned theme state"

patterns-established:
  - "hooks/ directory established as the codebase's home for custom hooks, useMediaQuery its first member"

requirements-completed: [SHELL-03, SHELL-04, SHELL-05, THEME-04, THEME-06]

coverage:
  - id: D1
    description: "Reducer prune: 16 tab-era AppState fields/actions/cases removed (selectedDirectory, selectedOutputDirectory, selectedSearchDirectory, selectedBrowseDirectory, isCreating, isSearching, isLoading, searchResults, sortColumn, sortDirection, browseCatalogs, browseSortColumn, browseSortDirection, sidebarCollapsed, sidebarPosition, activeTab), pre-existing dead catalog-modal fields left untouched, no storcat-last-* key touched anywhere in frontend/src/"
    requirement: "SHELL-03"
    verification:
      - kind: unit
        ref: "cd frontend && npx tsc --noEmit && npm run build (exit 0); grep -c for each removed identifier == 0; grep -c catalogModalOpen >= 1; grep -rc storcat-last- frontend/src/ == 0 everywhere"
        status: pass
    human_judgment: false
  - id: D2
    description: "density, railSide, detailOverlay added to AppState with SET_DENSITY/SET_RAIL_SIDE/SET_DETAIL_OVERLAY actions, Density/RailSide imported from themeTokens.ts (not redeclared), initialState seeded from readPersistedPrefs()"
    requirement: "THEME-04"
    verification:
      - kind: unit
        ref: "grep -c 'SET_DENSITY|SET_RAIL_SIDE|SET_DETAIL_OVERLAY' == 6 (one union member + one case each); grep -c readPersistedPrefs >= 1; grep -c inline union redeclaration == 0"
        status: pass
    human_judgment: false
  - id: D3
    description: "useMediaQuery(query) hook: matchMedia-backed, subscribes/unsubscribes via addEventListener('change')/removeEventListener, no resize listener; WorkspaceShell calls it once with the exact '(min-width: 1280px)' string workspace.css's widest breakpoint uses"
    requirement: "SHELL-03"
    verification:
      - kind: unit
        ref: "grep -c removeEventListener useMediaQuery.ts == 1; grep -c resize listener == 0; grep -c 'min-width: 1280px' in both workspace.css and WorkspaceShell.tsx == 1 each; no other file under components/workspace/ references matchMedia"
        status: pass
    human_judgment: false
  - id: D4
    description: "workspace.css min-width-only ladder (200px 1fr base, 236px 1fr at 1040px, 268px 1fr 288px at 1280px) confirmed intact with zero max-width or @container rules -- no gap or overlap at fractional widths"
    requirement: "SHELL-04"
    verification:
      - kind: unit
        ref: "grep -c max-width workspace.css == 0; grep -c @container workspace.css == 0; grep -c 'min-width: 1040px' == 1; grep -c 'min-width: 1280px' >= 1"
        status: pass
    human_judgment: true
    rationale: "The literal breakpoint values and rule shape are confirmed by grep, but that the rail visually measures 236px/200px/268px and the tree keeps 1fr at each tier while dragging the window through both boundaries can only be judged in a running wails dev window -- no GUI available this session."
  - id: D5
    description: "data-rail-side=\"Right\" swaps grid-template-columns to 288px 1fr 268px, rail to order 3 with its divider moved to the left edge, details to order 1 with its divider moved to the right edge -- all scoped inside the @media (min-width: 1280px) block only, so the rail snaps back to the left with its base right-edge divider below that tier regardless of the stored setting"
    requirement: "SHELL-05"
    verification:
      - kind: unit
        ref: "grep -c '288px 1fr 268px' workspace.css == 1; awk-scoped grep confirms every data-rail-side occurrence in workspace.css falls inside the 1280px block; grep -c data-rail-side WorkspaceShell.tsx == 1"
        status: pass
    human_judgment: true
    rationale: "That the rail visually moves to the right, keeps its 268px width, and the single 1px divider relocates correctly -- and that it snaps back below 1280px -- requires driving Ctrl+Alt+R in a running wails dev window and watching DevTools; not exercised by this session's automated checks. The 'user can move the rail' half of SHELL-05 is deliberately exercised only through this DEV-gated affordance this phase; the Settings control is Phase 26 (SET-03), per the plan's flagged_planner_assumptions."
  - id: D6
    description: "Density flows from the reducer through the token layer: WorkspaceShell's useEffect re-invokes applyTokens(theme, state.density) whenever state.density changes; DevStateSwitcher's Ctrl+Alt+D now dispatches SET_DENSITY (and writes DENSITY_KEY) instead of calling applyTokens itself, so the reducer is the single source of truth"
    requirement: "THEME-04"
    verification:
      - kind: unit
        ref: "grep -c SET_DENSITY DevStateSwitcher.tsx == 1; grep -c SET_RAIL_SIDE == 1; grep -c literal storcat-density/storcat-rail-side string keys == 0 (keys imported from themeTokens.ts); no pane component under components/workspace/ branches on Compact/Comfortable"
        status: pass
    human_judgment: true
    rationale: "That pressing Ctrl+Alt+D actually changes :root --rh between 27px/34px and --rp between the two density strings live in DevTools -- and that theme/density/rail-side all survive a full quit and relaunch of the built binary -- requires a running wails dev session and a built-binary restart, neither available in this environment."
  - id: D7
    description: "Rail side and density persist to storcat-rail-side / storcat-density via readPersistedPrefs()'s existing allowlist-and-rewrite logic (unchanged by this plan) and are restored into the reducer's initialState on module load"
    requirement: "THEME-06"
    verification:
      - kind: unit
        ref: "grep -c readPersistedPrefs AppContext.tsx >= 1; module-scope initialState seeded from persistedPrefs.density/persistedPrefs.railSide"
        status: pass
    human_judgment: true
    rationale: "Full-restart persistence of the built binary (not just a dev reload) is called out explicitly in the plan's own <human-check> and can only be confirmed by quitting and relaunching the packaged app -- not available in this environment."

duration: 3min
completed: 2026-08-13
status: complete
---

# Phase 22 Plan 06: Shell State + Responsive Breakpoints Summary

**Pruned 16 dead tab-era reducer fields, added density/railSide/detailOverlay to AppContext, wrote the codebase's first hook (`useMediaQuery`), and wired the 1280px/1040px width tiers plus the right-side rail swap into `workspace.css` scoped to the widest tier only.**

## Performance

- **Duration:** ~3 min (commit-to-commit)
- **Started:** 2026-08-13T16:05:53-05:00
- **Completed:** 2026-08-13T16:07:25-05:00
- **Tasks:** 3
- **Files modified:** 5 (1 created, 4 modified)

## Accomplishments
- `AppContext.tsx`'s reducer holds only live state now: the three new workspace fields (`density`, `railSide`, `detailOverlay`) plus the pre-existing (untouched, out-of-scope) modal fields — every field the deleted tab UI owned is gone, confirmed by `tsc --noEmit` under `noUnusedLocals`
- `density`/`railSide` seed from `readPersistedPrefs()` at module scope, so a relaunch restores the user's choices without a second, divergent validation path
- `frontend/src/hooks/useMediaQuery.ts` — the first hook file in the codebase — subscribes to a `MediaQueryList`'s `change` event (never a resize listener), so window dragging can't trigger a per-pixel render storm
- `WorkspaceShell` calls `useMediaQuery('(min-width: 1280px)')` once, using the exact threshold string `workspace.css`'s widest breakpoint uses; the min-width-only ladder (200px/236px/268px) it inherited from 22-01 was confirmed intact with zero drift
- `data-rail-side="Right"` swap rules (grid template, rail/details `order`, divider edge) added scoped entirely inside `@media (min-width: 1280px)`, so the rail snaps back to left with its base right-edge divider below that tier regardless of the stored setting — matching the prototype's own conditional
- `WorkspaceShell` re-invokes `applyTokens` whenever `state.density` changes, making the density reducer field exercisable rather than theoretical; `DevStateSwitcher` gained `Ctrl+Alt+R` for rail side and now dispatches `SET_DENSITY` into the reducer instead of calling `applyTokens` itself

## Task Commits

Each task was committed atomically:

1. **Task 1: Prune the reducer's dead tab-era fields and add the three workspace fields** - `c3ee5979` (feat)
2. **Task 2: Add useMediaQuery and the min-width breakpoint ladder** - `c748e698` (feat)
3. **Task 3: Swap the rail to the right side, move its divider, and exercise density and rail side from the dev affordance** - `90692729` (feat)

## Files Created/Modified
- `frontend/src/contexts/AppContext.tsx` - Removed 16 dead fields/actions/cases; added `density`/`railSide`/`detailOverlay` state, `SET_DENSITY`/`SET_RAIL_SIDE`/`SET_DETAIL_OVERLAY` actions, seeded from `readPersistedPrefs()`
- `frontend/src/hooks/useMediaQuery.ts` - New: `useMediaQuery(query): boolean`, matchMedia-change-event subscription
- `frontend/src/workspace.css` - `data-rail-side="Right"` swap rules, scoped inside the existing `@media (min-width: 1280px)` block
- `frontend/src/components/workspace/WorkspaceShell.tsx` - `useMediaQuery` call gating `DetailsPanel`'s render, `data-rail-side` attribute on `.ws-root`, density-reapply `useEffect`
- `frontend/src/components/dev/DevStateSwitcher.tsx` - `Ctrl+Alt+R` rail-side toggle; `Ctrl+Alt+D` now dispatches `SET_DENSITY` instead of calling `applyTokens` directly; reads density/railSide from `useAppContext()` instead of a shadow local copy

## Decisions Made
- Kept theme as `DevStateSwitcher`'s own local state (Phase 26 still owns theme's eventual reducer/Settings home) while reading `density`/`railSide` straight from `useAppContext().state` — avoids a second, divergent copy of the two fields this plan moved into the reducer.
- `WorkspaceShell`'s density-reapply effect re-reads `readPersistedPrefs().theme` on every density change rather than threading theme through props or context, per the plan's own "for now" instruction ahead of Phase 26's Settings-owned theme state.

## Deviations from Plan

None - plan executed exactly as written. All automated acceptance-criteria greps passed on first implementation; no auto-fixes were required.

## Issues Encountered

None.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness
- Plan 22-07 can now consume `WorkspaceShell`'s `isWide` boolean to choose `DetailsPanel`'s pane-or-drawer `variant`, replacing this plan's temporary "render only when wide" branch.
- **Manual verification still outstanding** (flagged in `coverage` above, `human_judgment: true`): live `wails dev` pixel/tier confirmation across both breakpoints, `Ctrl+Alt+R` rail-side visual swap and snap-back below 1280px, `Ctrl+Alt+D` `--rh`/`--rp` value confirmation in DevTools, and a full quit-and-relaunch of the built binary confirming theme/density/rail-side persistence — carried forward to the phase gate alongside every prior plan's own `human_judgment: true` items (22-01 D5/D6, 22-02 D5, 22-03 D1/D2/D3, 22-04 D1/D2, 22-05 D1/D2).
- No blockers for 22-07 onward.

---
*Phase: 22-shell-token-layer*
*Completed: 2026-08-13*
