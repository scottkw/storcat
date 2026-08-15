---
phase: 27-catalog-actions-watch
plan: 04
subsystem: ui
tags: [react, typescript, wails, menu, dialog, workspace-css]

# Dependency graph
requires:
  - phase: 27-01
    provides: "wailsAPI.renameCatalog wrapper and the containment-gated App.RenameCatalog Wails binding this plan calls"
  - phase: 26-settings
    provides: "SettingsDialog's always-mounted shell shape, generalized here into the shared DialogShell"
  - phase: 24-command-palette
    provides: "useModalBehavior -- the one focus-trap/Escape/scroll-lock/focus-restore implementation, reused unchanged by both Menu.tsx and DialogShell.tsx"
provides:
  - "Menu.tsx -- the app's first anchored popover primitive: conditionally mounted, role=menu, roving tabIndex with ArrowUp/ArrowDown wraparound, Enter/Space activation, its own click-outside pointerdown listener, position measured once from the trigger's getBoundingClientRect, no viewport-flip logic"
  - "DialogShell.tsx -- the one shared 440px centred-dialog shell both this phase's dialogs use (rename here, delete-confirmation in 27-05)"
  - "RenameDialog.tsx -- ACT-02's rename surface: pre-filled and text-selected on open, Enter commits, disabled on empty trimmed value, real error banner on failure with the edited value preserved"
  - "DetailsPanel.tsx's inert '...' button now opens a real menu via a new CatalogActions component, used at both existing call sites"
  - "workspace.css's full Phase 27 CSS surface: --danger/--ondanger/--z-menu tokens plus every class the menu, both dialogs, and the status-bar watching segment need -- 27-05 and 27-07 add only TSX, never reopen this file"
affects: [27-05, 27-07]

# Actuals (#2632)
actuals:
  tokens: 7923
  tasks: 3
  commits: 3

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Conditionally-mounted overlay (Menu.tsx) vs always-mounted overlay (DialogShell.tsx) -- two deliberately distinct useModalBehavior consumer shapes in the same phase, not unified into one mounting convention"
    - "scrollLockSelector pointed at a non-scrolling trigger button (.ws-details-overflow) as a functional no-op configuration of the shared hook, not a fork of it"
    - "Imperative DOM .value sync immediately before .select() on a controlled input, so React's own reconciliation (which skips re-assigning value when it already matches) doesn't collapse the selection to a trailing caret"

key-files:
  created:
    - frontend/src/components/workspace/Menu.tsx
    - frontend/src/components/workspace/DialogShell.tsx
    - frontend/src/components/workspace/RenameDialog.tsx
  modified:
    - frontend/src/workspace.css
    - frontend/src/components/workspace/DetailsPanel.tsx

key-decisions:
  - "Menu.tsx's arrow-key/Enter/Space handling is a React onKeyDown prop on the menu container (not a document-level addEventListener), so it never collides with the shared hook's own window keydown listener that owns Escape"
  - "Menu's first-item ref doubles as useModalBehavior's initialFocusRef and as itemRefs[0] for roving-focus math -- one stable ref object, not two refs racing each other"
  - "CatalogActions holds its own useAppContext() call (dispatch for the optimistic SET_CATALOGS update) rather than threading dispatch down as a prop from DetailsPanel"
  - "Footer's error state hoisted to DetailsPanel (actionError/setActionError, declared before the early returns) so CatalogActions' menu-item placeholders and Footer's two existing actions share one error surface, per 27-UI-SPEC.md"
  - "Duplicate and Delete menu items report an explicit 'not yet connected' message through onError rather than being disabled, greyed, or silently no-oping -- 27-05 replaces both handlers in the same file"

requirements-completed: [ACT-01, ACT-02]

coverage:
  - id: D1
    description: "The ⋯ button opens a real three-item menu (Rename catalog… / Duplicate catalog / divider / Delete catalog…) with roving arrow-key focus that wraps at both ends, Enter/Space activation, Escape and click-outside close, and focus restore to the ⋯ button on every close path"
    requirement: "ACT-01"
    verification:
      - kind: e2e
        ref: "live dev-browser against :34115 -- Object.keys(window.go.main.App) confirmed fresh bindings, then ARIA/menu-open/ArrowDown/ArrowUp-wraparound/Escape-focus-restore/click-outside-close all exercised and asserted, see Task 3 evidence below"
        status: pass
    human_judgment: false
  - id: D2
    description: "Rename opens pre-filled with the catalog's current title, text selected so typing replaces it immediately, Enter commits identically to the primary button, the primary button disables exactly when the trimmed value is empty, and a successful rename updates the rail row and details panel in the same frame the dialog closes"
    requirement: "ACT-02"
    verification:
      - kind: e2e
        ref: "live dev-browser: real catalog created via CreateCatalog, renamed to 'Tom & Jerry Renamed' via the dialog's Enter-to-commit path, rail row and details panel both showed the new title immediately; separately confirmed selectionStart:0/selectionEnd:21 on open and primary-button disabled with an emptied field"
        status: pass
    human_judgment: false
  - id: D3
    description: "A failed rename shows the inline banner 'Couldn't rename this catalog: <system error message>.', the dialog stays open, the edited value is preserved, and the primary button re-enables for a retry with no retyping"
    requirement: "ACT-02"
    verification:
      - kind: e2e
        ref: "live dev-browser: RenameCatalog('/etc/hosts', catalogDir, 'x') called directly against the fresh binding, confirming the exact containment rejection string ('rename /etc/hosts: outside configured catalog directory') this dialog's catch branch surfaces verbatim"
        status: pass
    human_judgment: true
    rationale: "The underlying binding's error-string fidelity was proven directly against the live App binding (matching 27-01's own proof), but the RenameDialog's own catch branch -- banner render, dialog-stays-open, value-preserved, button-re-enabled -- was verified by code review against Footer's already-proven identical pattern, not exercised end-to-end through the dialog UI itself in this session. A human should click through one real failed rename (e.g. rename while catalogDir is briefly cleared) before this is fully closed."
  - id: D4
    description: "Rename and delete share one 440px dialog shell component; there is no second near-duplicate 440px panel implementation"
    verification:
      - kind: other
        ref: "acceptance-criteria greps: RenameDialog.tsx contains no ws-settings-panel/width:440 literal and renders exactly one <DialogShell; DialogShell.tsx is always-mounted-shaped (if (!isOpen) return null) and calls useModalBehavior exactly once"
        status: pass
    human_judgment: false

duration: ~20min
completed: 2026-08-15
status: complete
---

# Phase 27 Plan 04: Catalog Actions Menu + Rename Dialog Summary

**The details panel's `⋯` button now opens a real anchored, keyboard-navigable menu, and rename opens on a shared 440px dialog shell -- pre-filled, text-selected, Enter-to-commit, updating the rail optimistically -- while landing every CSS class this phase's four surfaces need in one pass.**

## Performance

- **Duration:** ~20 min
- **Started:** 2026-08-15T17:56:10Z
- **Completed:** 2026-08-15T18:10:05Z
- **Tasks:** 3
- **Files modified:** 5 (3 created, 2 modified)

## Accomplishments
- `workspace.css` carries the phase's complete new surface in one pass: `--danger`/`--ondanger`/`--z-menu` tokens (the `--z-menu` tier sits correctly between `--z-details-drawer` and `--z-overlay`), `.ws-menu`/`.ws-menu-item`/`.ws-menu-item-danger`/`.ws-menu-divider`, the full `.ws-dialog-*` shared shell, `.ws-rename-input` (Sans, not mono), `.ws-delete-*` (for `27-05`), and `.ws-status-right`/`.ws-status-watching`/`.ws-status-watching-dir` (for `27-07`) -- the three pre-existing `#e5534b` literals were left untouched, formalized only via the new `--danger` declaration
- `Menu.tsx` is the app's first anchored popover: conditionally mounted (no internal `isOpen` gate, the opposite of `SettingsDialog`'s always-mounted shape), `role="menu"` with roving `tabIndex` that wraps at both ends on ArrowUp/ArrowDown, Enter/Space activation, its own `pointerdown` click-outside listener (not provided by the shared hook), and position computed once from the trigger's `getBoundingClientRect()` with no viewport-flip logic
- `DialogShell.tsx` generalizes `SettingsDialog`'s scrim/panel/header/body/footer shape into slots at 440px -- the one shared shell both `RenameDialog` (this plan) and the delete-confirmation dialog (`27-05`) use
- `RenameDialog.tsx` wires `wailsAPI.renameCatalog`: pre-filled and fully text-selected on open, Enter commits, primary button disabled exactly when the trimmed value is empty, real error string surfaced verbatim on failure with the field's edited value intact
- `DetailsPanel.tsx`'s previously-inert `⋯` button (`OverflowButton`) is replaced at both call sites by a new `CatalogActions` component: menu ARIA (`aria-haspopup`/`aria-expanded`/`aria-controls`) wired correctly, `Rename catalog…` opens the real dialog, `Duplicate catalog`/`Delete catalog…` report an explicit "not yet connected" message (not a disabled item, not a silent no-op) pending `27-05`
- `Footer`'s error state hoisted into `DetailsPanel` (`actionError`/`setActionError`) so the menu's placeholder messages and the footer's two existing actions share one error surface
- Full round trip proven live against `wails dev` on `:34115` after confirming fresh bindings (`Object.keys(window.go.main.App)` included `RenameCatalog`): opened the menu, walked all three items with ArrowDown/ArrowUp confirming wraparound, closed via Escape (focus restored to `⋯`) and via click-outside, reopened, renamed a real catalog to `"Tom & Jerry Renamed"` via Enter, and confirmed the rail row and details panel both updated in the same frame

## Task Commits

1. **Task 1: The phase's full CSS surface** - `0a5b7165` (feat) -- `--danger`/`--ondanger`/`--z-menu` tokens plus every class the menu, both dialogs, and the status-bar watching segment need
2. **Task 2: Menu.tsx -- the anchored popover primitive** - `4df8a19a` (feat)
3. **Task 3: DialogShell + RenameDialog, and the ⋯ button finally opens something** - `9c6d7e2f` (feat)

**Plan metadata:** pending (this SUMMARY's own commit)

## Files Created/Modified
- `frontend/src/workspace.css` - `--danger`/`--ondanger`/`--z-menu` tokens; `.ws-menu*`, `.ws-dialog-*`, `.ws-rename-input`, `.ws-delete-*`, `.ws-status-right`/`.ws-status-watching*` classes
- `frontend/src/components/workspace/Menu.tsx` (new) - anchored popover menu primitive
- `frontend/src/components/workspace/DialogShell.tsx` (new) - shared 440px dialog shell
- `frontend/src/components/workspace/RenameDialog.tsx` (new) - ACT-02's rename surface
- `frontend/src/components/workspace/DetailsPanel.tsx` - `OverflowButton` replaced by `CatalogActions` at both call sites; `Footer`'s error state hoisted to `DetailsPanel`

## Decisions Made
- Menu's arrow-key/Enter/Space handling lives in a React `onKeyDown` prop on the menu container rather than a `document`-level listener, so it can never collide with the shared hook's own `window` keydown listener (which owns Escape).
- The menu's first item's DOM ref doubles as `useModalBehavior`'s `initialFocusRef` and as `itemRefs[0]` for the roving-focus array -- one stable ref object serving both purposes rather than two refs that could drift out of sync.
- `CatalogActions` calls `useAppContext()` itself for the optimistic `SET_CATALOGS` dispatch, rather than threading `dispatch` down as a prop from `DetailsPanel`.
- Fixed a real bug found live: the rename input's DOM `.value` is now synced imperatively immediately before `.select()` runs. Calling `.select()` against the input's stale pre-render DOM value (still empty on the very first paint after `isOpen` flips true) selected nothing; React's subsequent controlled re-render then collapsed the cursor to the end the instant it landed the real value. Pre-syncing the DOM value means React's reconciliation finds `node.value` already equal to the incoming value and skips touching it, so the selection this effect makes survives the commit.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Rename field's pre-fill selection collapsed to a trailing caret instead of selecting the whole title**
- **Found during:** Task 3 live verification
- **Issue:** `RenameDialog`'s seeding effect called `inputRef.current?.select()` directly, but the input's DOM value at that instant was still the previous render's stale value (empty, on first open) -- `select()` against an empty string selects nothing, and React's subsequent controlled-value commit then reset the cursor to the end rather than preserving any selection, since a genuine value change moves the caret.
- **Fix:** Set `input.value = catalog.title` imperatively immediately before calling `.focus()`/`.select()`, so React's reconciliation finds the DOM value already matches the incoming state and skips reassigning it -- leaving the selection this effect makes untouched.
- **Files modified:** `frontend/src/components/workspace/RenameDialog.tsx`
- **Verification:** Live dev-browser: `selectionStart:0, selectionEnd:21` (the full 21-character pre-filled title) confirmed on open, versus `selectionStart:21, selectionEnd:21` (collapsed) before the fix.
- **Committed in:** `9c6d7e2f` (Task 3 commit)

### Acceptance-criterion mismatch (not a bug, documented for the record — same class as 27-01's two documented mismatches)

**1. `grep -c 'useModalBehavior' Menu.tsx` returns `3`, not the literal `1` the plan's acceptance criteria state**
- **Found during:** Task 2 verification
- **Cause:** The literal count is unreachable for any file that both imports and calls the hook by name -- the `import { useModalBehavior } from ...` line alone contributes one match, and `SettingsDialog.tsx` (the plan's own named analog for this exact wiring) also matches 3 times for the identical reason. The functional intent -- "the call passes `scrollLockSelector: '.ws-details-overflow'`" -- is fully satisfied: there is exactly one `useModalBehavior({...})` call site in the file.
- **Action taken:** Implemented correctly (one hook call, correct `scrollLockSelector`); did not attempt to reduce the import/comment mentions to force the literal grep count, since doing so would mean either not importing the hook by its real name or omitting explanatory comments -- neither is a real fix.
- **Verification:** `grep -n 'useModalBehavior' Menu.tsx` shows exactly one call site (`const { containerRef } = useModalBehavior({...`); `npx tsc --noEmit && npm run build` both green.

---

**Total deviations:** 2 (1 auto-fixed real bug, 1 acceptance-criterion/checkpoint-wording mismatch against the plan's own analog file) -- no scope creep.
**Impact on plan:** The bug fix directly strengthens a locked `must_haves.truth` (pre-fill + full selection). The wording mismatch has no functional impact; documented per this repo's established precedent (27-01) rather than distorting the implementation to chase an unreachable literal grep count.

## Issues Encountered
- `wails dev` was already running from a prior session (`27-01`'s process); `Object.keys(window.go.main.App)` was checked before recording any binding-dependent live evidence per the standing "curl liveness is not binding freshness" constraint -- `RenameCatalog` was present, confirming the bridge was already fresh for this plan's frontend-only changes (Vite HMR had already picked up the new components).
- The configured catalog directory in the running app pointed at a prior scratchpad path with zero catalogs. Created a fresh scratch source directory (`source-27-04`) and output directory (`catalogs-27-04`), set it via the live `SetCatalogDirectory` binding, and created a real catalog via the live `CreateCatalog` binding rather than hand-crafting a fixture -- the same precedent `27-01` established.

## Next Phase Readiness
- `27-05` (delete-confirmation dialog) can consume `DialogShell` and the already-declared `.ws-delete-*` classes directly, and only needs to replace the two placeholder `onSelect` handlers (`Duplicate catalog`, `Delete catalog…`) inside `DetailsPanel.tsx`'s `CatalogActions` component -- no new menu wiring, no new dialog shell, no CSS.
- `27-07` (watching status-bar segment) can consume `.ws-status-right`/`.ws-status-watching`/`.ws-status-watching-dir` directly -- no CSS work remains for that plan either.
- One item flagged for human follow-up before full sign-off: `RenameDialog`'s failure sub-state (inline banner, dialog-stays-open, value-preserved, retry-enabled) was verified by direct binding rejection and code-review parity with `Footer`'s already-proven pattern, but not clicked through end-to-end in the dialog UI itself this session (see coverage `D3`).

## Self-Check: PASSED

All 5 created/modified files verified present on disk; all 3 task commits (`0a5b7165`, `4df8a19a`, `9c6d7e2f`) verified present in `git log`.

---
*Phase: 27-catalog-actions-watch*
*Completed: 2026-08-15*
