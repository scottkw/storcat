---
phase: 24-cmd-k-command-palette
plan: 04
subsystem: ui
tags: [react, typescript, aria, listbox, command-palette, dev-browser]

# Dependency graph
requires:
  - phase: 24-cmd-k-command-palette
    provides: "24-01's SearchIndexed binding and always-mounted CommandPalette; 24-02's live-proven capped search path; 24-03's useModalBehavior hook"
provides:
  - "PaletteResultRow -- one listbox option rendering shape, first-occurrence-highlighted basename, dimmed unhighlighted path, static catalog chip, right-aligned size, entirely from JSX text children"
  - "PaletteResultList -- the flat unvirtualized listbox, PLT-03's exact truncation line sourced from the Go-computed total, and scroll-into-view on the active option"
  - "CommandPalette's four mutually exclusive body states (Hint/Searching/Results/No-matches) with PLT-06's exact empty-result copy, plus the roving-activeIndex keyboard handler (Down/Up/Home/End/PageUp/PageDown/Enter) and full combobox/listbox ARIA wiring"
affects: [24-05, 25, 26, 27]

# Actuals (#2632)
actuals:
  tokens: 3996
  tasks: 3
  commits: 3

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Highlight spans are built from String.prototype.slice into three JSX text children (before/match/after), never an HTML string -- the only way to guarantee React's automatic escaping applies to attacker-influenceable catalog filenames (T-24-11)"
    - "PageUp/PageDown step is measured live (scroll container clientHeight / first rendered option's offsetHeight), not hardcoded, with a documented module-constant fallback for the pre-layout case"
    - "Row identity across catalogs uses a runtime-built separator (String.fromCharCode(0)) rather than a template-literal escape sequence in source, after the Write tool round-tripped an authored \\u0000-style escape into a literal control byte -- the runtime-built value achieves the same 'cannot occur in a filesystem path' guarantee without a raw byte sitting in the .tsx source"

key-files:
  created:
    - frontend/src/components/workspace/palette/PaletteResultRow.tsx
    - frontend/src/components/workspace/palette/PaletteResultList.tsx
  modified:
    - frontend/src/components/workspace/CommandPalette.tsx
    - frontend/src/workspace.css

key-decisions:
  - "PaletteResultList.scrollRef and PaletteResultList's own internal typing use React.RefObject<HTMLDivElement> (not | null) -- same fix 24-03 already made for ModalBehavior.containerRef: React 18's RefObject<T> already nests null in `current`, so unioning the generic with null breaks JSX ref assignment."
  - "Added an optional `id` prop to PaletteResultList (not in the plan's task-3 file list) so CommandPalette's input can point aria-controls at the listbox's actual DOM id -- necessary to complete the ARIA wiring task 3 requires; documented as a Rule 3 deviation below."
  - "Chose '000001' (matches exactly FILE_000001.BIN, 1 of 400) for the single-highlight-span live check, and 'CANON' (matches directory basenames like '100CANON' but not file basenames like 'IMG_0001.JPG', whose containing path does match) for the path-only-match live check -- both computed from the actual fixture JSON rather than guessed."
  - "The T-24-11 injection fixture uses a basename with a literal `<b>` and no closing tag (`evil<b>name.txt`) rather than `evil<b>bold</b>name.txt` -- a closing tag's `/` is a path separator, and Go's path-splitting treated an initial test fixture's embedded `/` as a directory boundary, silently truncating the basename to `b>name.txt`. That was a test-fixture bug, not a product defect: no real filesystem allows a literal `/` inside a single path component, so the corrected fixture is the representative attacker-controlled case."

patterns-established:
  - "Row-level highlight computation (case-insensitive first-occurrence indexOf against the basename only, JSX text children only) is the template for any future phase that needs to render attacker-influenceable filenames with a highlighted match."

requirements-completed: [PLT-03, PLT-04, PLT-06]

coverage:
  - id: D1
    description: "Each result row shows the basename with the matched substring highlighted (first occurrence only, JSX text children, never dangerouslySetInnerHTML), the dimmed unhighlighted full path, the catalog as a static filled chip, and the size -- directory and file hits distinguishable by shape"
    requirement: "PLT-04"
    verification:
      - kind: other
        ref: "cd frontend && npx tsc --noEmit && npm run build"
        status: pass
      - kind: other
        ref: "grep -rc 'dangerouslySetInnerHTML' frontend/src/components/workspace/palette/ == 0 for every file"
        status: pass
      - kind: automated_ui
        ref: "dev-browser :34115 -- query '000001' (1 exact basename match): 1 row, 1 accent span in .ws-palette-name, text 'FILE_000001.BIN'. Query 'CANON' (path-only match on file rows): file row 'IMG_0001.JPG' rendered with 0 spans, directory row '100CANON' (basename match) rendered with 1 span. Fixture with basename 'evil<b>name.txt': .ws-palette-name textContent exactly 'evil<b>name.txt', innerHTML shows '&lt;b&gt;' (escaped), document.querySelectorAll('.ws-palette-name b').length === 0 -- all observed"
        status: pass
    human_judgment: false
  - id: D2
    description: "When more than 50 nodes matched, 'Showing the first 50 of N hits' appears with the real Go-computed total substituted for N; when 50 or fewer matched, the line is absent entirely"
    requirement: "PLT-03"
    verification:
      - kind: other
        ref: "grep -q 'total > results.length' PaletteResultList.tsx; grep -c 'results.length} hits' PaletteResultList.tsx == 0"
        status: pass
      - kind: automated_ui
        ref: "dev-browser :34115 -- query 'FILE' against the 400-node flat fixture: .ws-palette-truncation textContent exactly 'Showing the first 50 of 400 hits', [role=option] count exactly 50. Query 'VOL01' against the dcim fixture (16 real matches): .ws-palette-truncation count 0, readout '16 hits' -- all observed"
        status: pass
    human_judgment: false
  - id: D3
    description: "The four body states (Hint/Searching/Results/No-matches) are mutually exclusive; PLT-06's exact empty-result copy renders only for a settled zero-result query and is unreachable for a query under 2 characters or still in flight"
    requirement: "PLT-06"
    verification:
      - kind: other
        ref: "grep -Fq 'No file in any catalog matches that.' CommandPalette.tsx; grep -Fq 'Type to search…' CommandPalette.tsx; grep -Fq 'Searching…' CommandPalette.tsx"
        status: pass
      - kind: automated_ui
        ref: "dev-browser :34115 -- 1-character query 'F': body shows 'Type to search…', never the no-match string. Query 'zzznotfound' (zero real matches): body shows 'No file in any catalog matches that.', document.querySelectorAll('.ws-palette-state').length === 1. During a settled-results session, typing an additional character kept rows rendered continuously -- polled row count never dropped below 50 across the debounce/fetch window -- observed"
        status: pass
    human_judgment: false
  - id: D4
    description: "Down/Up/Home/End/PageUp/PageDown move a single active row and clamp at both ends without wrapping; Enter activates; the active row is always scrolled into view; mouse hover retargets the same active index"
    requirement: "PLT-04"
    verification:
      - kind: other
        ref: "grep for all six key literals plus aria-activedescendant/role=\"combobox\" in CommandPalette.tsx; grep -c \"=== 'Escape'\" == 0; grep -c \"addEventListener('keydown'\" == 0 (Escape stays owned by the shared hook)"
        status: pass
      - kind: automated_ui
        ref: "dev-browser :34115, 50-row 'FILE' list -- 60x ArrowDown clamped aria-activedescendant at ws-palette-option-49 (no wrap); Home -> option-0; ArrowUp from option-0 stayed at option-0; End -> option-49; Home then PageDown moved to option-13 (>1, <50) -- all observed"
        status: pass
    human_judgment: false

# Metrics
duration: ~50min
completed: 2026-08-14
status: complete
---

# Phase 24 Plan 04: Result Row, Listbox, and Body States Summary

**The ⌘K palette now renders a real listbox: highlighted-basename result rows built from JSX text children only (no injection path for a hostile catalog filename), a flat scroll-region with the exact "Showing the first 50 of N hits" truncation line sourced from Go's own total, and four mutually exclusive body states driven by a roving keyboard-navigable active index.**

## Performance

- **Duration:** ~50 min
- **Started:** 2026-08-14T15:34:00Z
- **Completed:** 2026-08-14T16:24:00Z
- **Tasks:** 3
- **Files modified:** 4 (2 created, 2 modified)

## Accomplishments
- `PaletteResultRow.tsx` renders shape / highlighted basename / dimmed path / static chip / size entirely as JSX text children -- a case-insensitive first-occurrence `indexOf` against the basename only, never the path, and never an HTML string. Live-verified against a fixture filename containing a literal `<b>`: it renders as visible escaped text, with zero real `<b>` elements in the DOM (T-24-11's machine-checked acceptance criterion).
- `PaletteResultList.tsx` is a flat, unvirtualized `role="listbox"` in backend order. The PLT-03 truncation line reads the `total` prop -- never `results.length` -- and is absent whenever the real total is 50 or fewer. The active option scrolls into view with `block: 'nearest'` on every `activeIndex` change.
- `CommandPalette.tsx` renders exactly one of Hint / Searching / Results / No-matches at a time, with every Copywriting Contract string quoted verbatim. Down/Up/Home/End/PageUp/PageDown/Enter drive a roving `activeIndex` that clamps at both ends without wrapping; the page step is measured live from the rendered viewport with a documented fallback. Full `combobox`/`listbox` ARIA wiring (`aria-expanded`, `aria-controls`, `aria-activedescendant`) was added without ever touching Escape, which stays owned by 24-03's shared `useModalBehavior` hook.
- All three tasks' automated acceptance criteria (grep + `tsc`/`build`) and all live `wails dev` :34115 behavior assertions were run and passed with observed values (see `coverage` above) -- including a re-confirmation that waves 2 and 3's proven behaviors (toolbar/⌘K open with focus, second ⌘K no-op preserving the query, Escape close with focus restore, scroll lock engage/release, 5-tab focus trap) remain unbroken.

## Task Commits

Each task was committed atomically:

1. **Task 1: The result row -- shape, highlighted basename, dimmed path, catalog chip, size** - `8796713b` (feat)
2. **Task 2: The listbox -- roving active index, scroll-into-view, and the PLT-03 truncation line** - `d7da84ec` (feat)
3. **Task 3: The four body states, the exact copy, and the keyboard handler** - `f8733283` (feat)

**Plan metadata:** commit follows this summary.

## Files Created/Modified
- `frontend/src/components/workspace/palette/PaletteResultRow.tsx` - new: one listbox option, safe highlight, T-24-11 mitigation
- `frontend/src/components/workspace/palette/PaletteResultList.tsx` - new: flat listbox, truncation line, scroll-into-view, optional `id` prop for ARIA wiring
- `frontend/src/components/workspace/CommandPalette.tsx` - four body states, keyboard handler, page-step measurement, ARIA wiring, footer
- `frontend/src/workspace.css` - row-anatomy classes, `.ws-palette-truncation`, `.ws-palette-state`, `.ws-palette-footer`, `.ws-palette-row[data-active]`

## Decisions Made
See `key-decisions` in frontmatter: the `RefObject<HTMLDivElement>` typing precedent reused from 24-03; the added `id` prop on `PaletteResultList`; the fixture terms chosen for live verification; and the injection-fixture correction (no embedded `/` inside a single path component).

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Added an optional `id` prop to `PaletteResultList`**
- **Found during:** Task 3, wiring `aria-controls` on the combobox input
- **Issue:** Task 3's ARIA requirement ("aria-controls pointing at the listbox element's id") needs the listbox to actually expose an id, but Task 2's `PaletteResultList` (declared in Task 2's own `<files>`, not Task 3's) had no id prop at all.
- **Fix:** Added an optional `id?: string` prop, applied to the root `div`, defaulting to `undefined` when omitted. `CommandPalette.tsx` passes a shared `PALETTE_LISTBOX_ID` constant to both the input's `aria-controls` and the list's `id`.
- **Files modified:** `frontend/src/components/workspace/palette/PaletteResultList.tsx` (already committed in Task 2, amended in Task 3's commit)
- **Verification:** `tsc --noEmit` and `npm run build` both pass; live check confirmed `aria-activedescendant` correctly targets rendered option ids.
- **Committed in:** `f8733283` (Task 3's commit, since it surfaced while wiring that task's ARIA requirement)

**2. [Rule 1 - Bug] `PaletteResultList.scrollRef`'s type broke `tsc` when wired to a real JSX ref**
- **Found during:** Task 2, running `npx tsc --noEmit` after passing `listScrollRef` (typed via `useRef<HTMLDivElement>(null)`) into the prop
- **Issue:** `scrollRef: React.RefObject<HTMLDivElement | null>` produced the same `TS2322` mismatch 24-03 already diagnosed for `ModalBehavior.containerRef` -- React 18's own `RefObject<T>` already nests `null` in `current`, so the extra union creates a non-assignable type rather than a redundant one.
- **Fix:** Changed the prop type to `React.RefObject<HTMLDivElement>`, matching `useRef<HTMLDivElement>(null)`'s actual return type exactly.
- **Files modified:** `frontend/src/components/workspace/palette/PaletteResultList.tsx`
- **Verification:** `tsc --noEmit` exits 0; `npm run build` succeeds
- **Committed in:** `d7da84ec` (Task 2's commit)

---

**Total deviations:** 2 auto-fixed (1x Rule 3 -- missing ARIA wiring surface, 1x Rule 1 -- recurrence of 24-03's known type-level fix). Neither changes runtime behavior beyond what the task already specified; both were required to make the task's own stated acceptance criteria achievable.
**Impact on plan:** None on scope.

## Issues Encountered
- The `Write` tool round-tripped an authored ` `-style escape-sequence text into a literal raw NUL byte in the `.tsx` source (confirmed via `od -c`/`xxd` byte inspection), which the plan explicitly prohibits ("written as an escape and never as a literal control byte"). Resolved by building the separator at runtime with `String.fromCharCode(0)` instead of embedding any control-character text in source -- functionally identical (still a NUL, still unrepresentable in a filesystem path), with no raw byte in the file. Documented here rather than silently worked around, since it's a tooling behavior worth knowing about for any future plan that specifies a literal control-character escape in source text.
- The first injection-safety test fixture (`evil<b>bold</b>name.txt`) exposed that a closing HTML tag's `/` is itself a valid path separator on POSIX, causing Go's path-splitting to treat it as a directory boundary and truncate the observed basename to `b>name.txt` -- a test-fixture artifact, not a product defect (no real filesystem allows a literal `/` inside one path component). Corrected to `evil<b>name.txt` (opening tag only, no `/`), which is what the plan's acceptance criterion actually specifies ("a fixture catalog whose node name contains a literal `<b>`").

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- PLT-03, PLT-04, and PLT-06 are all live-verified against the two-catalog fixture directory at `wails dev` :34115, with observed values recorded in `coverage` above.
- `handleActivate(result)` in `CommandPalette.tsx` is the intentional activation seam this plan leaves for 24-05: it currently only calls `onClose()`, and 24-05 extends the same function with the catalog switch and reveal request PLT-05 specifies, without replacing it.
- `PaletteResultList`'s new `id` prop and `PALETTE_LISTBOX_ID` constant are available for 24-05 or later phases if any additional ARIA relationship needs to reference the listbox.
- Zero new dependencies: `git diff --stat -- frontend/package.json frontend/package-lock.json` produces no output.
- The T-24-11 mitigation (JSX-text-only highlight rendering) is the template other phases should follow whenever attacker-influenceable filesystem text reaches this webview's DOM.

---
*Phase: 24-cmd-k-command-palette*
*Completed: 2026-08-14*

## Self-Check: PASSED

All 5 claimed files verified present on disk; all 4 claimed commit hashes (`8796713b`, `d7da84ec`, `f8733283`, `88e51428`) verified present in git log.
