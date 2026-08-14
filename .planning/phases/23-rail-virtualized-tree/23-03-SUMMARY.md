---
phase: 23-rail-virtualized-tree
plan: 03
subsystem: workspace-tree
tags: [react, tanstack-react-virtual, css-grid, flexbox]

requires:
  - phase: 23-01
    provides: "FlatNode/FlatCatalog model, LoadCatalogFlat binding, useVisibleRows, the AppContext TreeState union and @tanstack/react-virtual wiring"
provides:
  - "TOGGLE_EXPAND/SET_EXPANDED/SET_SELECTED reducer actions on AppContext"
  - "frontend/src/lib/format.ts -- formatBytes/formatCount/formatGB/formatDate, the phase's only formatters"
  - "Real tree row rendering (carets, shapes, sizes, click semantics), scroll reset on catalog switch, and three mutually exclusive TreePane states (empty library, loading, empty catalog)"
  - "BreadcrumbBar.tsx -- per-segment coloured path, expand-all, collapse-to-root"
affects: [23-04, 23-05, 23-06]

actuals:
  tokens: 4706
  tasks: 3
  commits: 3

tech-stack:
  added: []
  patterns:
    - "Reducer actions stay pure state transitions -- SET_EXPANDED replaces the whole expansion map in one dispatch, TOGGLE_EXPAND flips a single key, neither ever iterates or rebuilds the node array"
    - "Interaction colors (hover, selected, per-segment) live in workspace.css keyed on data attributes, never inline -- so :hover/attribute rules can win over a static per-render style, per the Phase 22 defect precedent"
    - "Grid items that host unbounded nowrap text need their own min-width: 0, not just the innermost text span's -- a CSS Grid track's auto-sizing pulls the whole subtree's min-content regardless of a descendant flex item's min-width override"

key-files:
  created:
    - frontend/src/lib/format.ts
    - frontend/src/components/workspace/BreadcrumbBar.tsx
  modified:
    - frontend/src/contexts/AppContext.tsx
    - frontend/src/components/workspace/TreePane.tsx
    - frontend/src/workspace.css

key-decisions:
  - "TOGGLE_EXPAND gates on node.type === 'directory' only (not hasChildren) per the plan's literal 'a click on a directory row' wording -- toggling a childless directory is a harmless no-op in useVisibleRows since it has no children to reveal"
  - "The four TreePane states collapse 'no directory configured', 'directory has zero catalogs', and 'nothing selected yet' into the single empty-library branch -- the plan's action text enumerates only these two conditions plus loading/empty-catalog/rows, and the pre-selection gap is the same visual landing state the Phase 22 tracer already used"
  - "BreadcrumbBar is mounted only inside TreePane's rows-rendering branch (not during loading/empty-catalog/empty-library) -- TREE-04's catalog header slot it sits beneath doesn't exist until 23-05, and the prototype itself only shows the breadcrumb inside its hasTree block"

patterns-established:
  - "A CSS Grid item that will ever host nowrap/ellipsis text several layers down needs min-width: 0 declared on the grid item itself, in addition to any flex-item min-width: 0 further down the tree -- the grid track's auto-sizing algorithm reads the whole subtree's min-content and ignores a descendant's own shrink override"

requirements-completed: [TREE-02, TREE-03, TREE-05, TREE-06, STATE-01]

coverage:
  - id: D1
    description: "TOGGLE_EXPAND/SET_EXPANDED/SET_SELECTED reducer actions and the phase's single formatter module (formatBytes/formatCount/formatGB/formatDate), fixed for 23-04/23-05/23-06 to import unchanged"
    verification:
      - kind: unit
        ref: "node -e against an esbuild-transpiled build of format.ts: formatBytes(0)='0B', formatBytes(1023)='1023B', formatBytes(1024)='1K', formatBytes(1536)='1.5K', formatBytes(1048576)='1M', formatBytes(1099511627776)='1T', formatGB(1024^4 bytes)='1.0 TB', formatGB(1023.9*1024^3 bytes)='1023.9 GB', formatCount(999999)='999,999', formatCount(1000000)='1.0M' -- all match acceptance criteria exactly"
        status: pass
      - kind: other
        ref: "cd frontend && npx tsc --noEmit && npm run build; grep -c \"export function format\" src/lib/format.ts -eq 4"
        status: pass
    human_judgment: false
  - id: D2
    description: "Real tree rows (caret, shape, name, size), directory-toggles-and-selects vs. file-selects-only click semantics, scroll reset on catalog switch, and three mutually exclusive states (empty library, quiet loading line, distinct empty-catalog message)"
    requirement: "TREE-02, TREE-06, STATE-01"
    verification:
      - kind: automated_ui
        ref: "dev-browser at :34115 against a 42,550-node DCIM fixture: directory click toggled expansion (caret ▸->▾, mounted row count 33->32, child row appeared) and selected in one action; file click selected only (mounted count unchanged, prior directory selection cleared); 20 rapid clicks on one caret left it collapsed (even count, no lost update); scroll set to 800px + expansion + selection all reset to 0/empty/null on catalog switch, and re-selecting the original catalog showed the expansion map cleared (caret back to ▸); screenshots t2303-rows.png, t2303-selected.png"
        status: pass
      - kind: automated_ui
        ref: "dev-browser: a zero-node fixture (fixture-empty.json) rendered the distinct 'This catalog is empty' message, visually and structurally separate from the empty-library block (0 catalogs) -- screenshot t2303-empty-catalog.png"
        status: pass
      - kind: other
        ref: "cd frontend && npx tsc --noEmit && npm run build; grep -c ws-tree-row/getComputedStyle/27px|34px checks all pass"
        status: pass
    human_judgment: false
  - id: D3
    description: "BreadcrumbBar: per-segment ancestor/current coloring derived from the selected node's parentIdx chain, expand-all (one O(n) map, one dispatch) and collapse-to-root (one empty-map dispatch), both idempotent and inert on a no-directories catalog"
    requirement: "TREE-03, TREE-05"
    verification:
      - kind: automated_ui
        ref: "dev-browser at :34115: selecting IMG_0001.JPG four levels deep rendered 4 breadcrumb segments, ancestors computed color rgb(13,143,156) (--ac) vs. current rgb(33,37,41) (--tx); expand-all on the 42,550-node fixture completed in ~300ms with mounted rows staying at 32 and DFS row order preserved (VOL01, 100CANON, its 16 files, 101CANON...); a second expand-all click left mounted count identical (32) and total visible rows unchanged (42,550); collapse returned visible rows to exactly 50 (top-level count); on fixture-flat (no directories) expand-all/collapse both left the 5-row count unchanged with no console error -- screenshot t2303-breadcrumb.png"
        status: pass
      - kind: other
        ref: "cd frontend && npx tsc --noEmit && npm run build; grep -c SET_EXPANDED src/components/workspace/BreadcrumbBar.tsx -eq 2"
        status: pass
    human_judgment: false

duration: 17min
completed: 2026-08-13
status: complete
---

# Phase 23 Plan 03: Interactive Tree, Breadcrumb, and the Phase Formatter Module Summary

**Real tree rows with caret/shape/click semantics, atomic scroll-reset-on-switch, three mutually exclusive empty/loading states, and a per-segment-coloured breadcrumb with O(n) expand-all/O(1) collapse -- all verified live against a 42,550-node fixture with measured timing, not assumed**

## Performance

- **Duration:** 17 min
- **Started:** 2026-08-13T20:10:54-05:00
- **Completed:** 2026-08-13T20:27:31-05:00
- **Tasks:** 3
- **Files modified:** 5 (2 created, 3 modified)

## Accomplishments

- `AppContext.tsx` gained `TOGGLE_EXPAND` (synchronous per-path flip), `SET_EXPANDED` (whole-map replace, the shape expand-all/collapse both use), and `SET_SELECTED` -- no reducer case ever iterates the node array
- `frontend/src/lib/format.ts`: `formatBytes` (exact port of Go's `formatBytes`, value for value at every boundary), `formatCount` (locale-grouped below 1M, `N.NM` above), `formatGB` (`gb`-helper port, `>=1024 -> N.N TB` else `N.N GB`), `formatDate` -- the phase's only formatters, fixed for 23-04/23-05/23-06
- `TreePane.tsx`: full row contract (caret reserved-width glyph, accent-square/dim-circle shape, ellipsized name, right-aligned formatted size), directory-click-toggles-and-selects vs. file-click-selects-only, scroll reset to 0 in the same effect that reacts to `currentCatalogId`, and three states beyond rows (empty library, quiet loading line naming the file being read, distinct empty-catalog message)
- `BreadcrumbBar.tsx`: per-segment `<span>` path derived by walking `parentIdx` upward from the selected node and reversing, ancestors in `--ac`, current segment in `--tx`; "Expand all" builds one `hasChildren`-keyed map and dispatches once; "Collapse" dispatches an empty map
- Found and fixed a real CSS bug during Task 3 verification: a deeply-selected breadcrumb path (60 levels) let its unbounded nowrap text widen the whole `.ws-tree` CSS Grid track instead of truncating, because only the innermost text span had `min-width: 0` -- the grid item itself needed it too

## Task Commits

Each task was committed atomically:

1. **Task 1: Expansion and selection actions, and the phase's one formatter module** - `adbf4e58` (feat)
2. **Task 2: Real tree rows -- carets, shapes, sizes, click semantics, scroll reset and the three empty states** - `1d09cf40` (feat)
3. **Task 3: The breadcrumb bar -- per-segment colour, expand-all and collapse-to-root** - `99a15bd4` (feat, includes the min-width Grid fix)

## Files Created/Modified

- `frontend/src/lib/format.ts` - `formatBytes`, `formatCount`, `formatGB`, `formatDate`
- `frontend/src/contexts/AppContext.tsx` - `TOGGLE_EXPAND`/`SET_EXPANDED`/`SET_SELECTED` action types and reducer cases
- `frontend/src/components/workspace/TreePane.tsx` - full row rendering, click handler, scroll-reset effect, four-state render, mounts `BreadcrumbBar`
- `frontend/src/components/workspace/BreadcrumbBar.tsx` - per-segment path, expand-all/collapse links
- `frontend/src/workspace.css` - `.ws-tree-row` hover/selected rules, `.ws-tree-caret`/`-shape`/`-name`/`-size`, `.ws-crumb`/`-current`/`-collapse`, and `.ws-tree { min-width: 0 }`

## Decisions Made

- `TOGGLE_EXPAND` gates on `node.type === 'directory'` (not `hasChildren`) per the plan's literal wording -- toggling a childless directory is a harmless no-op since `useVisibleRows` finds no children to reveal
- The empty-library branch in `TreePane` covers three conditions at once (no directory configured, directory has zero catalogs, nothing selected yet) -- the plan's action text names only the first two plus loading/empty-catalog/rows; the pre-selection gap reuses the same landing block the Phase 22 tracer already showed, rather than inventing a fifth state
- `BreadcrumbBar` mounts only inside `TreePane`'s rows-rendering branch, not during loading/empty-catalog/empty-library -- it sits beneath a catalog-header slot that doesn't exist until 23-05, and the prototype itself only renders the breadcrumb inside its `hasTree` block

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] `.ws-tree` CSS Grid item missing `min-width: 0`, letting a long breadcrumb path blow out the whole app's layout instead of truncating**
- **Found during:** Task 3, human-check verification against a 60-level-deep fixture
- **Issue:** Selecting a deeply-nested node produced a breadcrumb path with ~60 unwrapped segments. The path `<span>` already had `flex: 1 1 auto; min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap`, but its ancestor `.ws-tree` (a CSS Grid item in `.ws-grid`'s `1fr` track) had no `min-width` override, so the grid's auto-sizing algorithm computed the track's minimum from the subtree's min-content -- which for nowrap text equals its full unwrapped width -- and grew the whole app to ~3600px wide (`document.body.scrollWidth` measured at 4162px against a 1395px viewport) instead of clipping the text. This directly violated the plan's own must-have ("the bar's height stays 34px with no wrap") and acceptance criterion (long path truncates, both links stay visible).
- **Fix:** Added `min-width: 0;` to `.ws-tree` in `workspace.css`, alongside its existing `min-height: 0`. This is the standard fix for the "unbounded nowrap content bubbles a flex/grid item's min-content into an ancestor's auto-sizing" gotcha -- the grid-track floor and the flex-shrink floor each need their own override, not just the innermost one.
- **Files modified:** `frontend/src/workspace.css`
- **Verification:** Re-ran the same 60-level-deep fixture via dev-browser after the fix: `document.body.scrollWidth` back to 1395px (matching viewport), breadcrumb bar stayed 34px tall, both action links stayed fully visible (`getBoundingClientRect().right` within the tree pane's own width), path span showed `text-overflow: ellipsis` actually clipping instead of growing. Screenshot `t2303-longpath-fixed.png`.
- **Committed in:** `99a15bd4` (part of the Task 3 commit)

---

**Total deviations:** 1 auto-fixed (1 bug)
**Impact on plan:** Necessary for correctness -- the plan's TREE-05/breadcrumb-overflow must-haves would otherwise fail on any catalog with a moderately deep selected path. No scope creep; the fix is a single CSS property.

## Issues Encountered

None beyond the deviation above. `wails dev` was already running at the start of this plan; dev-browser's session persisted across all three tasks' verification.

## Known Stubs

None. This plan's own scope (expansion/selection state, tree row rendering, breadcrumb) is fully wired end to end and verified live. `TREE-04`'s catalog header and `STATE-02`'s unreadable-catalog panel remain explicitly plan 23-05's work, as already recorded in 23-01's summary and this plan's own frontmatter -- not an orphaned gap.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- `TOGGLE_EXPAND`/`SET_EXPANDED`/`SET_SELECTED` and `frontend/src/lib/format.ts` are ready for 23-04 (rail row detail + filter) and 23-05 (catalog header + details panel + unreadable-catalog panel) to consume directly, unchanged.
- `BreadcrumbBar` is positioned to receive the catalog header above it once 23-05 builds `TREE-04` -- no restructuring needed, it already renders as the first child inside `TreePane`'s rows branch.
- The `.ws-tree { min-width: 0 }` fix is a general-purpose safeguard that also protects 23-05's catalog header title (which the UI-SPEC also requires to ellipsize on one line) from the same class of bug.
- No blockers.

## Self-Check: PASSED

All 5 files claimed as created/modified confirmed present on disk; all 3 task commit hashes (`adbf4e58`, `1d09cf40`, `99a15bd4`) confirmed in `git log`.

---
*Phase: 23-rail-virtualized-tree*
*Completed: 2026-08-13*
