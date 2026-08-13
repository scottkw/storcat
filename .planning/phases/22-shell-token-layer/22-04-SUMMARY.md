---
phase: 22-shell-token-layer
plan: 04
subsystem: ui
tags: [react, typescript, wails, css-custom-properties, theming]

requires:
  - phase: 22-shell-token-layer (22-01)
    provides: "ws-rail/ws-status shell classes, .pane-scroll/.mono utilities, .ws-new-pill/.ws-rail-row density-contract rules, --acs/--ac/--onac/--dm/--fn/--ch tokens from applyTokens"
provides:
  - "CatalogRail.tsx: header (CATALOGS 0 label, New pill, directory chip, filter input) + unconditional empty state, all copy verbatim from the Copywriting Contract"
  - "StatusBar.tsx: three literal mono zero-state segments (0 catalogs / 0 files indexed / 0.0 GB), no right-aligned watch segment"
affects: [23-catalog-data-wiring]

actuals:
  tokens: 1100
  tasks: 2
  commits: 2

tech-stack:
  added: []
  patterns:
    - "Rail/status regions built as plain function declarations with inline style objects reading var(--...) tokens, no Ant Design, consistent with the 22-01/22-03 skeleton convention"
    - "Inert interactive-looking controls (New pill, directory chip, filter input, empty-state link) render with zero onClick/onChange handlers -- their targets (Phase 25 create slide-over, Phase 26 settings, Phase 23 filter logic) don't exist yet, so no stub wiring or fabricated data is added"

key-files:
  created: []
  modified:
    - frontend/src/components/workspace/CatalogRail.tsx
    - frontend/src/components/workspace/StatusBar.tsx

key-decisions:
  - "Kept the rail's directory-chip and empty-state link as non-interactive elements (div/span, not button) with no cursor: pointer, per the plan's explicit prohibition against giving an affordance that does nothing when clicked"

patterns-established: []

requirements-completed: [SHELL-01, THEME-02, THEME-04]

coverage:
  - id: D1
    description: "Catalog rail renders header block (CATALOGS 0 label, New pill, directory chip, filter input) above a scrollable unconditional empty state, all copy transcribed verbatim from the Copywriting Contract"
    requirement: "SHELL-01"
    verification:
      - kind: unit
        ref: "cd frontend && npx tsc --noEmit && npm run build (exit 0)"
        status: pass
      - kind: other
        ref: "grep -F -c for each Copywriting Contract string (No catalogs here yet / body / Catalog a volume → / No catalog directory set / Filter catalogs… / New) == 1 in CatalogRail.tsx"
        status: pass
    human_judgment: true
    rationale: "Pixel-accurate 268px rail measurement, row spacing, and visual placement can only be confirmed by running `wails dev` and inspecting the live window -- not verifiable from this session (no GUI available)."
  - id: D2
    description: "New pill renders --ac text on --acs tint background; hover inverts to full --ac fill with --onac text via the .ws-new-pill CSS rule declared in 22-01, legible across all 11 themes"
    requirement: "THEME-02"
    verification:
      - kind: unit
        ref: "cd frontend && npx tsc --noEmit && npm run build (exit 0)"
        status: pass
      - kind: other
        ref: "CatalogRail.tsx button className=\"ws-new-pill\" with background: var(--acs), color: var(--ac); .ws-new-pill:hover rule in workspace.css sets background: var(--ac), color: var(--onac)"
        status: pass
    human_judgment: true
    rationale: "Legibility of the hover-inverted pill across all 11 themes (especially light-accent themes like Gruvbox orange and Monokai green) requires cycling themes via DevStateSwitcher in `wails dev` -- not exercised by this session's automated checks; carried forward as a phase-gate item alongside 22-01's D5."
  - id: D3
    description: "Status bar renders as a fixed 26px band with exactly three literal mono zero-state segments (0 catalogs, 0 files indexed, 0.0 GB) and no right-aligned watch segment"
    requirement: "SHELL-01"
    verification:
      - kind: unit
        ref: "cd frontend && npx tsc --noEmit && npm run build (exit 0)"
        status: pass
      - kind: other
        ref: "grep -c '0 catalogs' == 1, '0 files indexed' == 1, '0.0 GB' == 1, 'marginLeft' == 0, 'flexShrink' == 3 in StatusBar.tsx"
        status: pass
    human_judgment: false
  - id: D4
    description: "Rail row padding contract (var(--rp)) and rail scroll region (flex/overflow-y/min-height/scrollbar-gutter) are honored -- no hardcoded density pixel literal in CatalogRail.tsx"
    requirement: "THEME-04"
    verification:
      - kind: other
        ref: "grep -c '6px 8px\\|10px 10px' CatalogRail.tsx == 0; .pane-scroll class applied to the scroll region (declared in workspace.css by 22-01)"
        status: pass
    human_judgment: false

duration: 4min
completed: 2026-08-13
status: complete
---

# Phase 22 Plan 04: Catalog Rail + Status Bar Summary

**Filled the catalog rail's header (CATALOGS 0 label, accent New pill, directory chip, filter input) and unconditional empty state, plus the status bar's three literal mono zero-state segments — every string transcribed verbatim from the Copywriting Contract, nothing wired to a capability that doesn't exist yet.**

## Performance

- **Duration:** ~4 min (commit-to-commit)
- **Started:** 2026-08-13T15:50:00-05:00
- **Completed:** 2026-08-13T15:50:59-05:00
- **Tasks:** 2
- **Files modified:** 2

## Accomplishments
- `CatalogRail.tsx` rebuilt as a plain function declaration: header block (uppercase `CATALOGS 0` label with the count in `--fn`, accent New pill with a 10px plus glyph, mono directory chip with a folder glyph, inert filter input with `aria-label="Filter catalogs"`), scroll region (`.pane-scroll`), and an unconditionally-rendered empty state (heading, two-sentence body, inert `Catalog a volume →` link)
- `StatusBar.tsx` rebuilt with three sibling mono spans reading `0 catalogs`, `0 files indexed`, `0.0 GB`; no right-aligned watch segment, no `marginLeft: auto`
- New pill's hover inversion (`--acs` tint → `--ac` fill / `--onac` text) sourced from the `.ws-new-pill:hover` rule 22-01 already declared in `workspace.css` — no new CSS added, only the class applied
- Zero interactive wiring: no `onClick`/`onChange` anywhere in either file; the directory chip, filter input, and empty-state link render but do nothing, matching the plan's explicit prohibition against fabricated affordances
- No Ant Design import, no raw `z-index`, no Wails binding call (`BrowseCatalogs`/`LoadCatalog`) in either file — live data wiring is Phase 23's SHELL-06/RAIL-01/RAIL-02

## Task Commits

Each task was committed atomically:

1. **Task 1: Build the catalog rail header, chips, filter input and empty state** - `daa3b355` (feat)
2. **Task 2: Build the 26px status bar with literal zero-state segments** - `7d5a7efa` (feat)

## Files Created/Modified
- `frontend/src/components/workspace/CatalogRail.tsx` - Header (label/New pill/directory chip/filter input), scroll region, unconditional empty state
- `frontend/src/components/workspace/StatusBar.tsx` - Three mono zero-state segments, `mono` class added to the root

## Decisions Made
- None beyond the plan's own explicit instructions — action text specified every style value, string, and class name precisely enough that no open decisions arose.

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
None.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness
- Both rail and status-bar seams the 22-01 tracer created are now filled at spec dimensions and tokens; Phase 23 can wire `BrowseCatalogs` data into the rail's rows and the status bar's live counts without touching this plan's structural/token work.
- **Manual verification still outstanding** (flagged in `coverage` above, `human_judgment: true`): full 11-theme New-pill hover legibility check via `DevStateSwitcher` in `wails dev`, and live-window pixel measurement of the rail (268/236/200px tiers) and status bar (26px, no wrap below 1040px) — carried forward to the phase gate alongside 22-01's D5/D6 and 22-03's D1/D2/D3, consistent with that plan's own precedent.
- `workspace.css` was not touched by this plan (single-owner-per-file constraint for the 22-04/22-05 wave respected).
- No blockers for 22-05 onward.

---
*Phase: 22-shell-token-layer*
*Completed: 2026-08-13*

## Self-Check: PASSED

- FOUND: `frontend/src/components/workspace/CatalogRail.tsx`
- FOUND: `frontend/src/components/workspace/StatusBar.tsx`
- FOUND: `.planning/phases/22-shell-token-layer/22-04-SUMMARY.md`
- FOUND: `daa3b355` (Task 1 commit)
- FOUND: `7d5a7efa` (Task 2 commit)
