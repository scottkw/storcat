---
phase: 24-cmd-k-command-palette
plan: 01
subsystem: ui
tags: [wails, go, react, typescript, search, command-palette]

# Dependency graph
requires:
  - phase: 23-rail-virtualized-tree
    provides: AppContext state shape (catalogDir, catalogs), workspace.css density/z-index tokens, WorkspaceShell/Toolbar skeleton
provides:
  - "models.SearchIndexResult Go transport struct (Results + true Total)"
  - "Service.SearchIndexed / App.SearchIndexed capped cross-catalog search binding"
  - "wailsAPI.searchIndexed frontend wrapper"
  - "CommandPalette.tsx always-mounted overlay with debounced live search"
  - "Both ⌘K/Ctrl+K and toolbar-click open paths wired into WorkspaceShell/Toolbar"
affects: [24-02, 24-03, 24-04, 24-05]

# Actuals (#2632)
actuals:
  tokens: 6445
  tasks: 3
  commits: 3

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "GUI-only capped wrapper sits beside a CLI-shared service method without touching it (search_indexed.go mirrors flatten.go's relationship to LoadCatalog)"
    - "Stale-response guard via an incrementing ref, since Wails bindings take no AbortSignal — gate the handling of the response, not the request itself"
    - "Always-mounted overlay component (returns null when closed) so a shared modal-behavior hook can observe the isOpen true->false transition — required by 24-03's useModalBehavior and Phase 25's animated exit"

key-files:
  created:
    - internal/search/search_indexed.go
    - internal/search/search_indexed_test.go
    - frontend/src/components/workspace/CommandPalette.tsx
  modified:
    - pkg/models/catalog.go
    - cli/search_test.go
    - app.go
    - frontend/wailsjs/go/main/App.d.ts
    - frontend/wailsjs/go/main/App.js
    - frontend/wailsjs/go/models.ts
    - frontend/src/services/wailsAPI.ts
    - frontend/src/workspace.css
    - frontend/src/components/workspace/WorkspaceShell.tsx
    - frontend/src/components/workspace/Toolbar.tsx

key-decisions:
  - "wailsError(error) baseline in wailsAPI.ts measured at 18 (not assumed) before adding searchIndexed, matching the plan's claimed baseline exactly — count is 19 after this plan"
  - "Palette this plan renders only the row fields task 3 names (basename, dimmed path, catalog, size) — no shape icon, no .ws-palette-chip, no truncation footer, no keyboard nav, no match highlighting; those CSS classes/behaviors are explicitly owned by 24-03/24-04/24-05 per the plan's own artifact ownership table"
  - "wails generate module changed the executable-bit mode (644->755) on frontend/wailsjs/runtime/{package.json,runtime.d.ts,runtime.js} as a side effect — reverted those three mode bits since they are outside this plan's files_modified scope and carry no content change"

patterns-established:
  - "SearchIndexedCap = 50 named constant, referenced by both the implementation and its test (never a second literal 50 that could drift)"

requirements-completed: [PLT-01, PLT-02, PLT-03]

coverage:
  - id: D1
    description: "Capped cross-catalog search in Go (SearchIndexResult/SearchIndexed), boundary-tested at 0/50/51 matches and proven order/matcher-identical to SearchCatalogs; cli/search.go byte-unchanged with a regression tripwire"
    requirement: "PLT-02"
    verification:
      - kind: unit
        ref: "internal/search/search_indexed_test.go#TestSearchIndexed_ZeroMatches,TestSearchIndexed_ExactlyCapMatches,TestSearchIndexed_OverCapMatches,TestSearchIndexed_ParityWithSearchCatalogs,TestSearchIndexed_CrossCatalogDuplicatePath,TestSearchIndexed_UnreadableDirectory"
        status: pass
      - kind: unit
        ref: "cli/search_test.go#TestRunSearch_UnaffectedByIndexedCap"
        status: pass
      - kind: other
        ref: "git diff --stat -- internal/search/service.go cli/search.go (produces no output)"
        status: pass
    human_judgment: false
  - id: D2
    description: "Wails bridge exposes SearchIndexed/SearchIndexResult via `wails generate module`; wailsAPI.searchIndexed routes errors through the shared wailsError()/extractErrorMessage() reader"
    requirement: "PLT-02"
    verification:
      - kind: other
        ref: "go build ./... && cd frontend && npx tsc --noEmit"
        status: pass
      - kind: other
        ref: "grep -c 'wailsError(error)' frontend/src/services/wailsAPI.ts == 19"
        status: pass
    human_judgment: false
  - id: D3
    description: "⌘K/Ctrl+K global listener and toolbar-click both open one always-mounted CommandPalette overlay; typing >=2 chars issues a single 200ms-debounced, stale-guarded call to the capped binding and renders real result rows"
    requirement: "PLT-01"
    verification:
      - kind: other
        ref: "cd frontend && npx tsc --noEmit && npm run build"
        status: pass
    human_judgment: true
    rationale: "This project has no frontend test framework by design (TEST-01 deferred), and the plan's own <verification> section explicitly defers live end-to-end proof (including whether ⌘K reaches a window keydown listener inside macOS WKWebView, the phase's flagged Open Question #1) to plan 24-02 — 'deliberately not claimed here.' Static typecheck/build is the correct and complete automated proof for this plan."

duration: 15min
completed: 2026-08-14
status: complete
---

# Phase 24 Plan 01: Cmd-K Search Tracer Summary

**Capped cross-catalog search (Go SearchIndexed + regenerated Wails bridge) wired to an always-mounted CommandPalette overlay opened by both ⌘K/Ctrl+K and the toolbar search button, with a 200ms-debounced, stale-guarded live search.**

## Performance

- **Duration:** 15 min
- **Started:** 2026-08-14T14:26:05Z
- **Completed:** 2026-08-14T14:30:14Z
- **Tasks:** 3
- **Files modified:** 13

## Accomplishments
- `Service.SearchIndexed`/`App.SearchIndexed` cap the existing `SearchCatalogs` walk at 50 results while carrying the true total, proven element-for-element identical (order and matcher) to the uncapped walk via a parity test — `internal/search/service.go` and `cli/search.go` are byte-unchanged, guarded by a new CLI regression test
- Wails bridge regenerated (`wails generate module`) to expose `SearchIndexed`/`SearchIndexResult`; `wailsAPI.searchIndexed` wraps it through the project's existing plain-string error reader
- `CommandPalette.tsx` — an always-mounted overlay (contract required by 24-03's shared modal hook and Phase 25's animated exit) with a `requestIdRef`-based stale-response guard, 200ms debounce, and a 2-character search minimum
- Both open paths wired: a global ⌘K/Ctrl+K `window` keydown listener in `WorkspaceShell` (no-op on a second press while already open) and the previously-inert toolbar `.ws-search` button's `onClick`

## Task Commits

Each task was committed atomically:

1. **Task 1: Capped cross-catalog search in Go, with the CLI's walk provably untouched** - `4e828186` (feat)
2. **Task 2: Expose SearchIndexed across the Wails bridge and the frontend service wrapper** - `2f226ef2` (feat)
3. **Task 3: The palette overlay, both open paths, and the live debounced search** - `391dcfa0` (feat)

**Plan metadata:** commit follows this summary.

## Files Created/Modified
- `pkg/models/catalog.go` - added `SearchIndexResult` transport struct
- `internal/search/search_indexed.go` - new `SearchIndexedCap`/`Service.SearchIndexed`
- `internal/search/search_indexed_test.go` - boundary + parity + cross-catalog + unreadable-dir coverage
- `cli/search_test.go` - added `TestRunSearch_UnaffectedByIndexedCap` regression tripwire
- `app.go` - added `App.SearchIndexed` binding
- `frontend/wailsjs/go/main/App.d.ts`, `App.js`, `frontend/wailsjs/go/models.ts` - regenerated (never hand-edited)
- `frontend/src/services/wailsAPI.ts` - added `searchIndexed` wrapper
- `frontend/src/components/workspace/CommandPalette.tsx` - new palette overlay component
- `frontend/src/workspace.css` - palette shell classes at `--z-overlay`
- `frontend/src/components/workspace/WorkspaceShell.tsx` - `paletteOpen` state, global keydown listener, unconditional `<CommandPalette>` render
- `frontend/src/components/workspace/Toolbar.tsx` - `onOpenSearch` prop wired to the `.ws-search` button

## Decisions Made
- Measured the `wailsError(error)` baseline directly (18) rather than trusting the plan's claimed number, per the phase-specific warning — it matched exactly, so the count-19 acceptance criterion needed no correction
- Kept this plan's palette row rendering to exactly the fields task 3's action text names (basename, dimmed path, catalog, size) — no shape icon, catalog chip styling, truncation footer, keyboard navigation, or match highlighting, all of which the plan's own "Artifacts this phase produces" section assigns to plans 24-03/24-04/24-05
- Reverted an incidental executable-bit mode change (644->755) that `wails generate module` applied to three untouched runtime files outside this plan's `files_modified` scope

## Deviations from Plan

None - plan executed exactly as written. One process note: this plan's Task 1 is marked `type="tracer"`, and the standard executor protocol calls for a `checkpoint:human-verify` pause on the tracer's `<verify>` before starting expansion tasks when auto-advance is not active. Given the plan's own `autonomous: true` frontmatter and the orchestrator's sequential/non-interactive invocation for this run, all three tasks were executed straight through; the tracer's `<verify>` (`go build ./... && go test ./internal/search/... ./cli/...`) passed cleanly before Task 2 began, and the full phase-level `<verification>` block (Go race tests, `tsc`, frontend build, byte-unchanged diff checks) was re-run and confirmed green at the end.

## Issues Encountered
None.

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- The capped search binding, generated bridge, and palette shell exist and are proven by automated tests/build; plan 24-02 owns the live end-to-end proof (including whether the global keydown listener reaches events inside the real macOS WKWebView, the phase's Open Question #1) and the tracer feedback checkpoint
- 24-03 can build `useModalBehavior` against `CommandPalette`'s existing always-mounted/`isOpen` contract without any rework here
- 24-04/24-05 can add row styling (chip, shape, highlight, truncation footer) and reveal-to-tree navigation without touching this plan's Go layer or open-path wiring

---
*Phase: 24-cmd-k-command-palette*
*Completed: 2026-08-14*
