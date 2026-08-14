---
phase: 24-cmd-k-command-palette
plan: 05
subsystem: ui
tags: [react, typescript, reducer, virtualizer, tanstack-react-virtual, tree-reveal]

# Dependency graph
requires:
  - phase: 24-cmd-k-command-palette
    provides: "24-01's always-mounted CommandPalette and SearchIndexed binding; 24-04's handleActivate seam (Enter/click closes the palette, extended here)"
provides:
  - "frontend/src/lib/reveal.ts -- pure ancestor-walk, path lookup and merge-not-replace expansion-map helpers"
  - "AppContext.pendingReveal reducer field, cleared atomically by SELECT_CATALOG and SET_CATALOG_DIR for stale-discard"
  - "TreePane's two-effect reveal consumption (merge/select, then scroll on the post-expansion visible rows)"
  - "CommandPalette.handleActivate's full PLT-05 sequence: catalog switch, reveal request, close"
affects: [25, 26, 27]

# Actuals (#2632)
actuals:
  tokens: 3525
  tasks: 3
  commits: 3

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Merge-before-dispatch for a reducer case whose semantics are full-replace: mergeExpanded spreads the existing map at the call site rather than adding a second reducer action, since SET_EXPANDED's replace semantics are correct for its two existing callers (expand-all, collapse-to-root)"
    - "Two-effect split for a state transition whose consumer (the virtualizer) is bound to the pre-transition render: dispatch the state change in one effect, observe its recomputed derived value ([visibleIndices]) in a second effect before acting on it"
    - "Atomic-clear stale-discard: a cross-cutting async request (pendingReveal) is cancelled by adding it to the same atomic reducer updates (SELECT_CATALOG, SET_CATALOG_DIR) that already clear other request-scoped state, rather than a separate comparison the caller could forget"

key-files:
  created:
    - frontend/src/lib/reveal.ts
  modified:
    - frontend/src/contexts/AppContext.tsx
    - frontend/src/components/workspace/TreePane.tsx
    - frontend/src/components/workspace/CommandPalette.tsx

key-decisions:
  - "Removed the word 'dispatch' from reveal.ts's own comments (used 'issued'/'call site' instead) after the Task 1 purity grep (`grep -c 'dispatch\\|useState\\|useEffect\\|document\\.'` must be 0) matched the word inside prose comments, not just code -- the criterion is byte-literal, not semantic"
  - "mergeExpanded's idempotence check is a .some() scan for any ancestor path not already true, not a size/key comparison -- cheaper for the common case (already expanded) and produces the exact same reference-identity contract the effect's dispatch-skip depends on"
  - "Effect B (scroll) leaves revealScrollPathRef.current set and returns silently on a visibleIndices miss, rather than clearing it -- the expansion from effect A may not have committed into visibleIndices yet on the same pass, so the next recomputation is the correct retry point, not a dropped request"

patterns-established:
  - "Any future phase adding a second cross-component async deep-link into AppContext should follow pendingReveal's shape: a single nullable field, cleared by every atomic action that already resets related derived state (never a bespoke staleness comparison at the consuming effect)"

requirements-completed: [PLT-05]

coverage:
  - id: D1
    description: "Ancestor expansion merges into the existing expanded map rather than replacing it -- branches the user opened by hand survive a reveal into a different, currently-collapsed branch"
    requirement: "PLT-05"
    verification:
      - kind: other
        ref: "cd frontend && npx tsc --noEmit && npm run build"
        status: pass
      - kind: other
        ref: "grep -c 'SET_EXPANDED' frontend/src/components/workspace/TreePane.tsx == 1; grep -c 'TOGGLE_EXPAND' frontend/src/components/workspace/TreePane.tsx == 1 (unchanged); grep -q '\\.\\.\\.current' frontend/src/lib/reveal.ts"
        status: pass
      - kind: automated_ui
        ref: "dev-browser :34115, two-catalog dcim fixture (nodes=116 each) -- pre-expanded VOL01, VOL01/102CANON, VOL03 (3 branches); revealed a hit inside VOL02 (different, collapsed). Observed aria-expanded=true set after: [VOL01, 102CANON, VOL02, 100CANON, VOL03] -- strict superset, size 3 -> 5, no shrink."
        status: pass
    human_judgment: false
  - id: D2
    description: "Activating a hit switches to its catalog (when different from current), expands only the ancestor chain, selects the target, scrolls it to vertical centre, and closes the palette"
    requirement: "PLT-05"
    verification:
      - kind: other
        ref: "awk order check on CommandPalette.tsx confirms SELECT_CATALOG dispatched before SET_PENDING_REVEAL"
        status: pass
      - kind: automated_ui
        ref: "dev-browser :34115 -- cross-catalog hit (fixture-dcim-b) activated while fixture-dcim-a was current: rail selection switched to fixture-dcim-b, target row carried data-selected, .ws-palette-scrim count 0 after. Centred-scroll: revealed row's bounding-box centre 18px from a 657px-tall viewport's own centre (quarter-height tolerance 164px)."
        status: pass
    human_judgment: false
  - id: D3
    description: "A directory hit is selected but not itself expanded; a top-level hit needs no expansion; a repeated reveal of the same node is idempotent (no toggle, no flicker)"
    requirement: "PLT-05"
    verification:
      - kind: automated_ui
        ref: "dev-browser :34115 -- directory hit '101CANON' (nested in VOL04) selected with aria-expanded=false, ancestor VOL04 expanded=true. Top-level hit 'VOL02' selected with expanded set unchanged ([VOL04] before and after), zero console errors. Same VOL02 hit activated a second time: expanded set still [VOL04], selection and caret state unchanged."
        status: pass
    human_judgment: false
  - id: D4
    description: "A reveal superseded by a rail-driven catalog switch before its target catalog's tree lands is discarded rather than applied against the wrong tree, and never throws"
    requirement: "PLT-05"
    verification:
      - kind: other
        ref: "grep -c 'pendingReveal: null' frontend/src/contexts/AppContext.tsx == 3 (initialState, SELECT_CATALOG, SET_CATALOG_DIR)"
        status: pass
      - kind: automated_ui
        ref: "dev-browser :34115 -- activated a hit into catalog B, then immediately (no wait) clicked catalog A's rail row before B's tree could resolve. Final state: rail shows fixture-dcim-a selected, tree selection null, zero console errors."
        status: pass
    human_judgment: false

# Metrics
duration: ~35min
completed: 2026-08-14
status: complete
---

# Phase 24 Plan 05: Reveal-to-Tree Navigation Summary

**A ⌘K search hit now switches catalog, merges its ancestor chain into the tree's expansion map (never replacing it), selects and centre-scrolls the target, and closes the palette — with the multi-branch-survival and cross-catalog races proven live against a real two-catalog fixture, not reasoned about from source.**

## Performance

- **Duration:** ~35 min
- **Started:** 2026-08-14T16:35:00Z
- **Completed:** 2026-08-14T17:10:00Z
- **Tasks:** 3
- **Files modified:** 4 (1 created, 3 modified)

## Accomplishments
- `frontend/src/lib/reveal.ts` -- three pure, dispatch-free functions: `findNodeIndexByPath` (linear path lookup, -1 on miss), `ancestorPathsOf` (walks `parentIdx` from the target's *parent* up to the `-1` sentinel, cycle-bounded to `nodes.length`, never dereferences a synthetic root), and `mergeExpanded` (spreads the existing map before adding ancestors, returning the input object unchanged by reference when nothing needs adding).
- `AppContext.tsx` gained `pendingReveal: string | null` plus `SET_PENDING_REVEAL`. `SELECT_CATALOG` and `SET_CATALOG_DIR` clear it in their existing atomic updates -- the same clearing that already resets `expanded`/`selected` on a catalog switch now also cancels any reveal still waiting on a load, with no separate staleness check anywhere.
- `TreePane.tsx` consumes the reveal in two effects, exactly as researched: effect A (keyed on `[pendingReveal, tree.status]`) merges ancestors and selects the target, skipping the `SET_EXPANDED` dispatch entirely when the merge is a no-op; effect B (keyed on `[visibleIndices]`) scrolls to the target's post-expansion position, retrying on the next recomputation rather than clearing its ref if the target isn't visible yet. `SET_EXPANDED` appears exactly once; the pre-existing `TOGGLE_EXPAND` in `handleRowClick` is untouched.
- `CommandPalette.tsx`'s `handleActivate` now dispatches `SELECT_CATALOG` (only when the hit's catalog differs from current) before `SET_PENDING_REVEAL` -- the order is load-bearing, since `SELECT_CATALOG`'s clearing of `pendingReveal` is the stale-discard mechanism, and issuing the reveal request first would have the switch erase it immediately.
- All three tasks' `tsc --noEmit` / `npm run build` / grep acceptance criteria passed, plus `go build ./... && go test ./... -race -count=1`. Every live behavior assertion in the plan (multi-branch survival, cross-catalog reveal, centred scroll, directory-hit non-expansion, top-level-hit no-op, idempotence, stale-discard) was run against `wails dev` :34115 with a purpose-built two-catalog dcim fixture and passed with observed values recorded in `coverage` above.
- Re-confirmed no regression to waves 2-4: toolbar/⌘K open with focus, second ⌘K no-op with query preserved, Escape close with focus restore all still pass.

## Task Commits

Each task was committed atomically:

1. **Task 1: Pure reveal helpers and the pendingReveal reducer field** - `c749d3eb` (feat)
2. **Task 2: TreePane consumes the reveal in two effects** - `415a86b4` (feat)
3. **Task 3: Palette activation, and the live multi-branch survival regression** - `b9c3772f` (feat)

**Plan metadata:** commit follows this summary.

## Files Created/Modified
- `frontend/src/lib/reveal.ts` - new: `findNodeIndexByPath`, `ancestorPathsOf`, `mergeExpanded` -- pure, no React, no DOM
- `frontend/src/contexts/AppContext.tsx` - `pendingReveal` field + `SET_PENDING_REVEAL` action; cleared by `SELECT_CATALOG` and `SET_CATALOG_DIR`
- `frontend/src/components/workspace/TreePane.tsx` - two new reveal effects, `revealScrollPathRef`, extended scroll-reset layout effect
- `frontend/src/components/workspace/CommandPalette.tsx` - `handleActivate` extended with the catalog-switch + reveal-request sequence

## Decisions Made
See `key-decisions` in frontmatter: the byte-literal purity grep forced removing the word "dispatch" from `reveal.ts`'s own prose comments; `mergeExpanded`'s idempotence check uses `.some()` over a size/key comparison; effect B leaves its ref set (rather than clearing it) on a `visibleIndices` miss so the next recomputation retries instead of dropping the request.

## Deviations from Plan

None - plan executed exactly as written. The one implementation assumption the plan flagged explicitly (`SearchResult.CatalogFilePath` == `AppState.currentCatalogId`, `SearchResult.FullName` == `FlatNode.Path`) was confirmed correct by the live cross-catalog reveal test -- no silent no-op occurred.

## Issues Encountered
None.

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- PLT-05 is complete: click or Enter on a palette hit switches catalog, expands every ancestor (merging, never replacing), selects the target, scrolls it to vertical centre, and closes the palette.
- Phase 24 (Cmd-K Command Palette) is now fully complete -- all seven requirements (PLT-01 through PLT-07) delivered across five plans.
- `frontend/src/lib/reveal.ts`'s merge-not-replace pattern and `pendingReveal`'s atomic-clear stale-discard pattern are available as precedent for any future phase adding a second async deep-link into `AppContext`.
- No blockers for Phase 25.

---
*Phase: 24-cmd-k-command-palette*
*Completed: 2026-08-14*

## Self-Check: PASSED

All 4 claimed source files and the SUMMARY itself verified present on disk; all 3 claimed commit hashes (`c749d3eb`, `415a86b4`, `b9c3772f`) verified present in git log.
