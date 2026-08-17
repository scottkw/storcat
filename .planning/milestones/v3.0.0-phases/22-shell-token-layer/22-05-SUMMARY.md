---
phase: 22-shell-token-layer
plan: 05
subsystem: ui
tags: [react, typescript, wails, css-custom-properties, theming]

requires:
  - phase: 22-shell-token-layer (22-01)
    provides: "ws-tree/ws-details shell classes, .pane-scroll/.ws-tree-row/.ws-meta-row/.ws-meta-value density-contract rules, --ac/--onac/--dm/--l/--rh/--fs/--mp tokens from applyTokens"
provides:
  - "TreePane.tsx: centred empty state (dashed 46px placeholder, heading, 420px-capped body, accent primary CTA + outlined secondary CTA), both CTAs inert, copy verbatim from the Copywriting Contract"
  - "DetailsPanel.tsx: uppercase DETAILS label + centred no-selection placeholder, single-component (no tier branching); exported empty DetailsPanelProps interface for 22-07 to extend additively"
affects: [22-07, 23-catalog-data-wiring]

actuals:
  tokens: 1300
  tasks: 2
  commits: 2

tech-stack:
  added: []
  patterns:
    - "TreePane/DetailsPanel built as plain function declarations with inline style objects reading var(--...) tokens, no Ant Design, consistent with the 22-01/22-03/22-04 skeleton convention"
    - "DetailsPanelProps declared exported and empty now so plan 22-07's variant: 'pane' | 'drawer' addition is a pure interface extension, not a signature change"

key-files:
  created: []
  modified:
    - frontend/src/components/workspace/TreePane.tsx
    - frontend/src/components/workspace/DetailsPanel.tsx

key-decisions:
  - "DetailsPanel takes no destructured props parameter at all (rather than an unused `_props: DetailsPanelProps` parameter) -- the exported interface exists for 22-07 to extend, but the component itself has nothing to read from it yet, so there is no dead parameter to justify."

patterns-established: []

requirements-completed: [SHELL-01, THEME-02, THEME-04]

coverage:
  - id: D1
    description: "Tree pane renders a centred empty state (dashed 46px placeholder mark, 'Nothing catalogued yet' heading, 420px-capped body, accent-filled 'Catalog a volume' primary CTA and outlined 'Choose catalog folder…' secondary CTA), all copy transcribed verbatim from the Copywriting Contract, both CTAs inert"
    requirement: "SHELL-01"
    verification:
      - kind: unit
        ref: "cd frontend && npx tsc --noEmit && npm run build (exit 0)"
        status: pass
      - kind: other
        ref: "grep -F -c for each Copywriting Contract string ('Nothing catalogued yet' / body sentence / 'Catalog a volume' / 'Choose catalog folder…') == 1 in TreePane.tsx; grep -c 'onClick|antd|zIndex' == 0"
        status: pass
    human_judgment: true
    rationale: "Pixel-accurate centring, 420px wrap measurement, and confirming the accent CTA reads as the screen's single loudest element can only be judged by running `wails dev` and inspecting the live window -- not verifiable from this session (no GUI available)."
  - id: D2
    description: "The accent-filled 'Catalog a volume' button uses --ac fill with --onac label text so it stays legible on light accents (Gruvbox orange, Monokai green) and dark accents (GitHub blue); it is the workspace's only large accent fill this phase"
    requirement: "THEME-02"
    verification:
      - kind: unit
        ref: "cd frontend && npx tsc --noEmit && npm run build (exit 0)"
        status: pass
      - kind: other
        ref: "TreePane.tsx primary button: background: var(--ac), color: var(--onac); grep -c 'var(--onac)' TreePane.tsx == 1; no other accent-filled element in the file"
        status: pass
    human_judgment: true
    rationale: "Legibility of --onac on the accent fill across all 11 themes (especially light-accent themes) requires cycling themes via DevStateSwitcher in `wails dev` -- not exercised by this session's automated checks; carried forward as a phase-gate item alongside 22-01's D5 and 22-04's D2."
  - id: D3
    description: "Details panel renders an uppercase DETAILS label (CSS textTransform, mixed-case markup text) and a centred no-selection placeholder, with no menu button and no footer action buttons"
    requirement: "SHELL-01"
    verification:
      - kind: unit
        ref: "cd frontend && npx tsc --noEmit && npm run build (exit 0)"
        status: pass
      - kind: other
        ref: "grep -c 'Nothing selected. Pick a catalog in the rail, or catalog a volume to get started.' DetailsPanel.tsx == 1; grep -c 'textTransform' == 1; grep -c 'onClick' == 0"
        status: pass
    human_judgment: false
  - id: D4
    description: "The tree row height/font-size contract (--rh/--fs via .ws-tree-row) and the details key/value row padding contract (--mp via .ws-meta-row) are established in workspace.css by 22-01 and referenced (not hardcoded) by these two components; no density pixel literal in either file; details panel has no internal tier branching -- no matchMedia/useMediaQuery/innerWidth"
    requirement: "THEME-04"
    verification:
      - kind: other
        ref: "grep -c '27px|34px' TreePane.tsx == 0; grep -c 'matchMedia|useMediaQuery|innerWidth' DetailsPanel.tsx == 0; workspace.css untouched by this plan (git diff confirms only the two component files changed)"
        status: pass
    human_judgment: false

duration: 5min
completed: 2026-08-13
status: complete
---

# Phase 22 Plan 05: Tree Pane + Details Panel Summary

**Filled the tree pane's centred empty state (the workspace's single accent-filled visual focal point) and the details panel's no-selection placeholder, built as one component with an exported-but-empty props interface ready for plan 22-07's pane/drawer variant.**

## Performance

- **Duration:** ~5 min (commit-to-commit)
- **Started:** 2026-08-13T15:53:00-05:00
- **Completed:** 2026-08-13T15:54:07-05:00
- **Tasks:** 2
- **Files modified:** 2

## Accomplishments
- `TreePane.tsx` rebuilt as a plain function declaration: `.pane-scroll` scroll region wrapping an unconditional centred empty state -- dashed 46px placeholder mark, "Nothing catalogued yet" heading, a 420px-capped body sentence transcribed verbatim (including the em dash), and a two-button row
- The accent-filled "Catalog a volume" primary CTA (`--ac` fill / `--onac` label) is the only large accent fill in the file, positioned as the workspace's single visual focal point per the UI-SPEC's Layout & Responsive Contract; the outlined "Choose catalog folder…" secondary CTA sits beside it. Both render with `type="button"` and no `onClick` -- their targets (create slide-over, Settings) are Phase 25/26
- `DetailsPanel.tsx` rebuilt as a plain function declaration: uppercase `DETAILS` label (`textTransform: uppercase` on mixed-case markup text, `--dm` color, no adjacent menu button) above a `.pane-scroll` body centring a single no-selection placeholder sentence, transcribed verbatim
- Exported an empty `DetailsPanelProps` interface -- deliberately no members yet, so plan 22-07's `variant: 'pane' | 'drawer'` addition is an additive interface change rather than a new signature; no `matchMedia`/`useMediaQuery`/`innerWidth` branching lives in this component
- Zero interactive wiring in either file: no `onClick` anywhere; no Ant Design import; no raw `z-index`/`zIndex`; `workspace.css` untouched (single-owner-per-file constraint for the 22-04/22-05 wave respected)

## Task Commits

Each task was committed atomically:

1. **Task 1: Build the tree pane's centred empty state and two inert CTAs** - `d0b07382` (feat)
2. **Task 2: Build the details panel as one component with a no-selection placeholder** - `fc5487d7` (feat)

## Files Created/Modified
- `frontend/src/components/workspace/TreePane.tsx` - Centred empty state: dashed placeholder mark, heading, 420px-capped body, accent primary CTA + outlined secondary CTA, both inert
- `frontend/src/components/workspace/DetailsPanel.tsx` - `DETAILS` label + centred no-selection placeholder + exported empty `DetailsPanelProps` interface

## Decisions Made
- Gave `DetailsPanel` no props parameter at all (not even an unused, underscore-prefixed one) rather than declaring `_props: DetailsPanelProps` and never reading it -- the exported interface exists purely for plan 22-07 to extend; adding a parameter the function never touches would be dead code the plan's own "no unrequested abstractions" spirit argues against.

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
None.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness
- Both seams the 22-01 tracer created (`TreePane`, `DetailsPanel`) are now filled at spec dimensions and tokens; the details panel's single-component identity is the hard precondition plan 22-07 needs to add the `variant` prop and render it as either an inline pane or a fixed drawer without duplicating the component.
- **Manual verification still outstanding** (flagged in `coverage` above, `human_judgment: true`): full 11-theme `--onac` legibility check on the "Catalog a volume" CTA via `DevStateSwitcher` in `wails dev`, and live-window pixel measurement of the tree pane's centring/420px wrap and the details panel's 288px width/14px padding/16px gap -- carried forward to the phase gate alongside 22-01's D5/D6, 22-03's D1/D2/D3, and 22-04's D1/D2.
- `workspace.css` was not touched by this plan (single-owner-per-file constraint for the 22-04/22-05 wave respected).
- No blockers for 22-06/22-07 onward.

---
*Phase: 22-shell-token-layer*
*Completed: 2026-08-13*

## Self-Check: PASSED

- FOUND: `frontend/src/components/workspace/TreePane.tsx`
- FOUND: `frontend/src/components/workspace/DetailsPanel.tsx`
- FOUND: `.planning/phases/22-shell-token-layer/22-05-SUMMARY.md`
- FOUND: `d0b07382` (Task 1 commit)
- FOUND: `fc5487d7` (Task 2 commit)
