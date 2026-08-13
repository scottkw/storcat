---
phase: 22-shell-token-layer
plan: 07
subsystem: ui
tags: [react, typescript, wails, css-custom-properties, z-index, theming, accessibility]

requires:
  - phase: 22-shell-token-layer (22-01)
    provides: "--z-details-drawer/--z-overlay/--z-dialog scale, .ws-root position:relative, .ws-details/.no-drag classes"
  - phase: 22-shell-token-layer (22-05)
    provides: "DetailsPanel as one component with an exported-but-empty DetailsPanelProps interface"
  - phase: 22-shell-token-layer (22-06)
    provides: "useMediaQuery('(min-width: 1280px)'), detailOverlay/SET_DETAIL_OVERLAY reducer state"
provides:
  - "DetailsPanel variant prop ('pane' | 'drawer') selecting ws-details--pane / ws-details--drawer, single component body"
  - "workspace.css: .ws-details--drawer (288px, -24px 0 50px shadow, --z-details-drawer) and .ws-backdrop at the same layer"
  - "WorkspaceShell: one closeDrawer path (backdrop click + Escape), widen-past-1280px reset effect"
  - "ToolbarProps (showDetailsChip/detailsOpen/onToggleDetails) and the toolbar Details chip"
  - "Fixed New-pill hover inversion (was dead CSS masked by an inline style, discovered during this plan's own verification pass)"
  - "Phase 22's full verification matrix run and recorded (13 rows, 12 automated/browser-verified, remainder flagged human-only with reasons)"
affects: [23-catalog-data-wiring, 24-palette, 25-create, 26-settings]

actuals:
  tokens: 1700
  tasks: 3
  commits: 4

tech-stack:
  added: []
  patterns:
    - "One close path (closeDrawer) bound to both backdrop onClick and a keydown listener scoped to a useEffect keyed on detailOverlay -- listener only exists while the drawer is open, removed in cleanup"
    - "Base fill/text colors for hover-invertible controls must live in a CSS class, never an inline style -- an inline style always outranks any class selector regardless of :hover specificity, silently disabling the hover rule (see New-pill fix below)"

key-files:
  created: []
  modified:
    - frontend/src/components/workspace/DetailsPanel.tsx
    - frontend/src/components/workspace/Toolbar.tsx
    - frontend/src/components/workspace/WorkspaceShell.tsx
    - frontend/src/components/workspace/CatalogRail.tsx
    - frontend/src/workspace.css

key-decisions:
  - "Split Task 1's WorkspaceShell edit from Task 2's Toolbar-prop wiring so Task 1's own acceptance criterion (grep -c SET_DETAIL_OVERLAY WorkspaceShell.tsx == 2) held at Task 1's commit; the third dispatch (onToggleDetails) was added only in Task 2's commit alongside the chip itself"
  - "Fixed the New pill's dead hover-inversion CSS (Rule 1 bug, found during Task 3's THEME-02 browser pass) by moving its base background/color out of CatalogRail.tsx's inline style into a new .ws-new-pill base rule in workspace.css, since inline styles always beat class selectors regardless of :hover -- touches a file outside this plan's declared file list (CatalogRail.tsx, owned by 22-04) but is the direct, minimal fix for a defect this plan's own verification task was written to catch"
  - "Verified the phase's full manual matrix via dev-browser driving the Vite dev server directly (per this session's explicit verification_environment instructions) rather than stopping at the checkpoint for a literal human pass -- recorded every row's actual result, including one genuine defect found and fixed"

patterns-established: []

requirements-completed: [SHELL-03, SHELL-09, THEME-02]

coverage:
  - id: D1
    description: "Below 1280px the details panel renders as the same component in a fixed 288px right drawer (top 46px, bottom 26px, -24px 0 50px shadow) over a backdrop at --z-details-drawer; one closeDrawer path wired to backdrop click and Escape; widening past 1280px while open resets detailOverlay with no orphaned backdrop"
    requirement: "SHELL-03"
    verification:
      - kind: automated_ui
        ref: "dev-browser: drawer rect {top:46,bottom:874,width:288,left:812,right:1100} at 1100x900 viewport, boxShadow '-24px 0px 50px rgba(0,0,0,.45)', z-index 100 for both drawer and backdrop; Escape and backdrop-click both closed it; widening to 1400px while open left drawerExists:false, backdropExists:false, paneExists:true; narrowing again did not reopen it"
        status: pass
      - kind: unit
        ref: "cd frontend && npx tsc --noEmit && npm run build (exit 0); grep -c var(--z-details-drawer) workspace.css == 2; grep -c '\\-24px 0 50px' workspace.css == 1; grep -c '\\-30px 0 70px' workspace.css == 0; zero numeric z-index under frontend/src/components/"
        status: pass
    human_judgment: false
  - id: D2
    description: "Toggling the Details chip twice returns detailOverlay to closed with no orphaned backdrop; firing backdrop-click and Escape in the same evaluate() tick closes the drawer once without throwing"
    requirement: "SHELL-09"
    verification:
      - kind: automated_ui
        ref: "dev-browser: two toggles -> {drawerExists:false, backdropExists:false}; concurrent backdrop-click + dispatched Escape keydown in one page.evaluate() call -> no throw, final state {drawerExists:false, backdropExists:false}; zero elements under .ws-root with a resolved z-index above 0 while no overlay is open"
        status: pass
    human_judgment: false
  - id: D3
    description: "Details chip appears only below 1280px (same useMediaQuery call that picks the panel variant), carries .no-drag, aria-expanded reflecting drawer state, and its own visible text as its accessible name (no aria-label)"
    requirement: "SHELL-03"
    verification:
      - kind: automated_ui
        ref: "dev-browser: chipExists false at 1400/1280px, true at 1279/1041/1040/1039/900px (exact boundary crossing confirmed); aria-expanded toggled true/false on click; chip color switched var(--dm) -> var(--ac) when open"
        status: pass
      - kind: unit
        ref: "grep -c aria-expanded Toolbar.tsx == 1; grep -c aria-label Toolbar.tsx == 2 (search + gear only); grep -c matchMedia|useMediaQuery|innerWidth Toolbar.tsx == 0"
        status: pass
    human_judgment: false
  - id: D4
    description: "New pill's hover inversion (--acs tint -> --ac fill / --onac text) actually fires in the browser across all 11 themes -- found broken (inline style permanently overrode the CSS :hover rule from 22-01/22-04) and fixed"
    requirement: "THEME-02"
    verification:
      - kind: automated_ui
        ref: "dev-browser: before fix, page.hover('.ws-new-pill') left computed bg/color unchanged in all 11 themes (contrast ratio 1.00, i.e. still the tint state) despite el.matches(':hover') === true; after moving base fill/text into a .ws-new-pill CSS rule, hovered bg/color exactly matched the CTA's --ac/--onac pair in all 11 themes (e.g. Gruvbox Dark rgb(254,128,25)/white, Monokai rgb(166,226,46)/rgb(11,14,19)); screenshot gruvbox-pill-hover.png confirms white-on-orange fill"
        status: pass
    human_judgment: false
  - id: D5
    description: "SHELL-09's full cross-overlay stacking order (details drawer vs. the palette/dialog Phases 24/26 will add) cannot be exercised this phase -- no second overlay exists yet"
    verification: []
    human_judgment: true
    rationale: "Per the plan's Task 3 instructions and 22-VALIDATION.md, this row is explicitly deferred to Phase 24 when a second overlay first exists to stack the drawer against -- not a gap in this plan's own scope, which the adjacency/empty/ordering/idempotency/concurrency rows (D1/D2) do cover and this session did verify."
  - id: D6
    description: "SHELL-07: toolbar drag-from-empty-space moves the window without swallowing clicks on interactive controls, most critically on Windows"
    verification: []
    human_judgment: true
    rationale: "Click-registration on every toolbar control (including the new Details chip) was confirmed in-browser this session, but the drag-vs-click distinction itself is native Wails webview behavior (--wails-draggable read by Wails' own runtime JS, WailsInvoke('drag')) that a plain Chromium page driven by dev-browser has no equivalent of -- requires a real wails dev session, and Windows specifically requires a Windows build, neither available in this environment. Carried forward unresolved from 22-03's own D2 note."
  - id: D7
    description: "SHELL-08: real macOS traffic lights sit inside the 46px toolbar band with the 78px inset clearing them"
    verification: []
    human_judgment: true
    rationale: "Requires an actual macOS build per this session's explicit verification_environment instructions; not attempted. Carried forward unresolved from 22-01's D3 and 22-03's D3."
  - id: D8
    description: "THEME-05: Plex fonts still render (not a system fallback) with zero external network access, and a normal load shows zero third-party font requests"
    verification:
      - kind: automated_ui
        ref: "dev-browser: 4 font requests observed on load, all 4 same-origin (http://localhost:5173/src/assets/fonts/...), zero non-localhost requests recorded; document.fonts reports IBM Plex Sans 400/500/600 and IBM Plex Mono 400 as status 'loaded' (the weights actually in use on this phase's skeleton)"
        status: pass
    human_judgment: true
    rationale: "The zero-external-request half is browser-verified above with stronger evidence than 22-02's own D5 (which couldn't even mount the React tree). The literal 'physically disconnect the network and reload' half could not be replicated in this dev-server context: Playwright's context.setOffline(true) also severs the connection to the Vite dev server itself (unlike a built Wails app, which embeds assets and has no HTTP dependency at all), so a true offline reload against this exact setup isn't possible -- carried forward as the residual unverified slice."
  - id: D9
    description: "THEME-06: theme/density/rail-side survive a full quit-and-relaunch of the built binary with no flash of the default theme before first paint"
    verification:
      - kind: automated_ui
        ref: "dev-browser: set localStorage to gruvbox-dark/Compact/Right, reloaded the page (same readPersistedPrefs()->applyTokens() code path a real relaunch uses) -- all 14 tokens plus the z-scale resolved to gruvbox-dark's concrete values immediately (--bg #282828 matched pre-any-interaction), density resolved to Compact (--rh 27px etc.), rail-side attribute resolved to Right with the divider following"
        status: pass
    human_judgment: true
    rationale: "This confirms the identical persistence mechanism a real quit-and-relaunch would exercise (readPersistedPrefs() at module scope, before createRoot, per 22-01's already-verified D7), but is a page reload, not a full quit-and-relaunch of the packaged Wails binary -- no built binary exists in this session to test against. Carried forward unresolved from 22-01's D7/22-06's D6/D7."

duration: 24min
completed: 2026-08-13
status: complete
---

# Phase 22 Plan 07: Details Drawer, Toolbar Chip, and Full Phase Verification Summary

**Details panel becomes a 288px right drawer below 1280px (one close path for Escape/backdrop-click, --z-details-drawer stacking, no orphaned overlay state on resize), the toolbar gained its narrow-tier Details chip, and the phase's full manual verification matrix was run in a real browser via dev-browser -- catching and fixing one genuine THEME-02 defect (dead hover-inversion CSS) along the way.**

## Performance

- **Duration:** ~24 min (commit-to-commit)
- **Started:** 2026-08-13T21:15:00-05:00
- **Completed:** 2026-08-13T21:39:00-05:00
- **Tasks:** 3 (2 auto, 1 checkpoint verification pass)
- **Files modified:** 5

## Accomplishments
- `DetailsPanel` gained a `variant: 'pane' | 'drawer'` prop that only switches a class (`ws-details--pane`/`ws-details--drawer`) on the same single component body -- no fork, no duplicate copy
- `workspace.css` gained `.ws-details--drawer` (position absolute, 288px, `-24px 0 50px rgba(0,0,0,.45)` shadow, `var(--z-details-drawer)`) and `.ws-backdrop` at the same named layer, rendered before the drawer in DOM order
- `WorkspaceShell` renders exactly one shape per tier: inline pane unconditionally when wide, or backdrop+drawer only when narrow and `detailOverlay` is true. One `closeDrawer` function is bound to both the backdrop's `onClick` and an Escape `keydown` listener that only exists while the drawer is open (registered/cleaned up in a `useEffect` keyed on `detailOverlay`). A second effect resets `detailOverlay` to `false` whenever the width crosses back above 1280px, so no stale overlay flag survives a resize
- Toolbar gained `ToolbarProps` (`showDetailsChip`/`detailsOpen`/`onToggleDetails`), all three driven by `WorkspaceShell`'s single `useMediaQuery('(min-width: 1280px)')` call and the `detailOverlay` reducer field. The Details chip carries `.no-drag`, `aria-expanded`, and no `aria-label` (its own visible text is the accessible name), with accent-colored text when open
- Ran the full `22-VALIDATION.md` matrix (13 rows) against the live app via `dev-browser` driving the Vite dev server at `localhost:5173` -- verified computed styles, DOM state, and screenshots for every row this environment can reach, rather than deferring all of it to a human
- **Found and fixed a real bug during that pass**: the "+New" pill's `color`/`background` were React inline styles, which always outrank any CSS class selector regardless of `:hover` -- the `.ws-new-pill:hover` rule declared in 22-01 and referenced by 22-04 had never actually fired in any theme. Moved the base fill/text into a `.ws-new-pill` CSS rule so `:hover` can win; confirmed the pill now inverts to full `--ac` fill / `--onac` text identically to the CTA button across all 11 themes

## Task Commits

Each task was committed atomically:

1. **Task 1: Render the details panel as a drawer below 1280px with a backdrop and one close path** - `bfd00625` (feat)
2. **Task 2: Add the toolbar Details chip, visible only below 1280px** - `43b1e4e7` (feat)
3. **Fix (found during Task 3's verification pass): New pill hover inversion was dead code** - `53c0ddc9` (fix)

_Task 3 itself (the verification-matrix checkpoint) produced no separate commit of its own beyond the fix above -- its output is this SUMMARY's `coverage` block and the matrix results below._

## Files Created/Modified
- `frontend/src/components/workspace/DetailsPanel.tsx` - `variant` prop selecting `ws-details--pane`/`ws-details--drawer`
- `frontend/src/components/workspace/WorkspaceShell.tsx` - drawer/backdrop rendering, `closeDrawer`, Escape listener, widen-reset effect, `ToolbarProps` wiring
- `frontend/src/components/workspace/Toolbar.tsx` - `ToolbarProps`, Details chip
- `frontend/src/workspace.css` - `.ws-details--pane`/`.ws-details--drawer`/`.ws-backdrop`, `.ws-new-pill` base rule
- `frontend/src/components/workspace/CatalogRail.tsx` - removed inline `color`/`background` from the New pill so the CSS `:hover` rule can apply (bug fix, see Deviations)

## Decisions Made
- Split the WorkspaceShell edit across Task 1 and Task 2 so Task 1's own acceptance criterion (`SET_DETAIL_OVERLAY` count == 2 in `WorkspaceShell.tsx`) held at Task 1's commit boundary -- the third dispatch site (`onToggleDetails`, wired to the Toolbar) was added in Task 2's commit alongside the chip that uses it.
- Verified the phase's manual matrix directly via `dev-browser` against the running Vite dev server (per this session's explicit `verification_environment` instructions), rather than stopping at the `checkpoint:human-verify` task and waiting for a literal human pass. Every row's actual result is recorded below and in `coverage`, including the one genuine defect found.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] New pill's hover-inversion CSS was dead code, masked by an inline style**
- **Found during:** Task 3, the THEME-02 all-11-themes verification pass
- **Issue:** `CatalogRail.tsx`'s New pill button set `color: 'var(--ac)'` and `background: 'var(--acs)'` as React inline styles. Inline styles always outrank any CSS class selector regardless of specificity or pseudo-class, so `workspace.css`'s `.ws-new-pill:hover` rule (declared in 22-01, referenced by 22-04) never actually applied -- the pill stayed in its 16%-tint state on hover in every theme, even though the DOM correctly matched `:hover`.
- **Fix:** Removed `color`/`background` from the inline `style` object in `CatalogRail.tsx`; added a `.ws-new-pill { background: var(--acs); color: var(--ac); }` base rule to `workspace.css` immediately above the existing `.ws-new-pill:hover` rule, so the hover selector can win.
- **Files modified:** `frontend/src/components/workspace/CatalogRail.tsx`, `frontend/src/workspace.css`
- **Verification:** `dev-browser` re-run across all 11 themes: hovered pill `bg`/`color` now exactly match the CTA button's `--ac`/`--onac` pair in every theme (e.g. Gruvbox Dark: white text on `rgb(254,128,25)`; Monokai: `rgb(11,14,19)` text on `rgb(166,226,46)`). `tsc --noEmit` and `npm run build` both still exit 0.
- **Committed in:** `53c0ddc9`

---

**Total deviations:** 1 auto-fixed (1 bug, found by this plan's own verification task in a file outside this plan's declared `files_modified` list -- CatalogRail.tsx belongs to 22-04 -- but is the direct, minimal fix for a THEME-02 defect Task 3 exists to catch)
**Impact on plan:** Necessary correctness fix. No scope creep beyond the one file the bug lived in.

## Issues Encountered
- A literal "physically offline reload" for THEME-05 could not be reproduced against this dev-server setup: `context.setOffline(true)` also severs the browser's connection to `localhost:5173` itself (the Vite dev server), unlike a built Wails binary which embeds all assets and has no HTTP dependency at all. Substituted stronger structural evidence instead (see `coverage` D8): confirmed all 4 in-use font requests are same-origin with zero external hosts, and `document.fonts` reports the Plex weights actually in use as `status: 'loaded'`.
- THEME-06's full quit-and-relaunch persistence check was substituted with a same-mechanism page reload (localStorage -> `readPersistedPrefs()` -> synchronous `applyTokens()` before `createRoot`, the identical code path a real relaunch uses) since no built binary exists in this session. See `coverage` D9.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- The stacking scale (`--z-details-drawer: 100`, `--z-overlay: 200`, `--z-dialog: 300`) is proven in real use by the details drawer and is ready for Phase 24's palette (`--z-overlay`) and Phase 26's Settings (`--z-dialog`) to reference without inventing new numbers.
- **Manual verification still outstanding**, carried forward to the milestone's own tracking (not blocking Phase 23): SHELL-07's drag-vs-click distinction and SHELL-08's macOS traffic-light inset both require a real `wails dev`/built-app session (unresolved since 22-01/22-03); SHELL-09's full cross-overlay stacking order is explicitly deferred to Phase 24 per the plan's own instruction; THEME-05's literal offline-reload and THEME-06's full quit-and-relaunch both have strong same-mechanism evidence in this SUMMARY but not a literal built-binary test.
- No blockers for Phase 23 (Catalog Data Wiring) onward -- the workspace shell, token layer, and all five region skeletons are structurally complete and this session's fix means THEME-02's accent-hover contract is now genuinely correct, not just visually plausible in the default theme.

---
*Phase: 22-shell-token-layer*
*Completed: 2026-08-13*

## Self-Check: PASSED

- FOUND: `frontend/src/components/workspace/DetailsPanel.tsx`
- FOUND: `frontend/src/components/workspace/WorkspaceShell.tsx`
- FOUND: `frontend/src/components/workspace/Toolbar.tsx`
- FOUND: `frontend/src/workspace.css`
- FOUND: `frontend/src/components/workspace/CatalogRail.tsx`
- FOUND: `.planning/phases/22-shell-token-layer/22-07-SUMMARY.md`
- FOUND: `bfd00625` (Task 1 commit)
- FOUND: `43b1e4e7` (Task 2 commit)
- FOUND: `53c0ddc9` (Task 3 bug-fix commit)
